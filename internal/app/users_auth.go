package app

import (
	"context"
	"slices"
	"sort"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/mekdep/server/config"
	"github.com/mekdep/server/internal/models"
	"github.com/mekdep/server/internal/store"
	"golang.org/x/crypto/bcrypt"
)

type UserLoginRequest struct {
	Username      *string   `json:"login" validate:"required"`
	Password      *string   `json:"password" validate:"omitempty,min=8"`
	Otp           *string   `json:"otp" validate:"omitempty"`
	SchoolID      *string   `json:"school_id" validate:"omitempty"`
	DeviceToken   *string   `json:"device_token" validate:"omitempty"`
	RolesPriority *[]string `json:"roles_priority"`
}

var ErrPassword = ErrInvalid.SetKey("password").SetComment("invalid password or user")

func Login(req *UserLoginRequest, isAdmin bool) (*models.User, error) {
	// TODO: user response convert , refactor
	m, err := store.Store().UsersFindByUsername(context.Background(), *req.Username, req.SchoolID, false)
	if err != nil {
		m, err = store.Store().UsersFindByUsername(context.Background(), *req.Username, req.SchoolID, true)
	}
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrPassword
		}
		return nil, err
	}
	if m.ID == "" {
		return nil, ErrPassword
	}
	err = store.Store().UsersLoadRelations(context.Background(), &[]*models.User{&m}, false)
	if err != nil {
		return nil, err
	}

	// roles priority to select login role
	rolePriorityMap := make(map[models.Role]int)
	if req.RolesPriority != nil {
		for i, role := range *req.RolesPriority {
			rolePriorityMap[models.Role(role)] = i
		}
	} else {
		rolePriorityMap = map[models.Role]int{
			models.RolePrincipal:    10,
			models.RoleTeacher:      20,
			models.RoleParent:       30,
			models.RoleStudent:      40,
			models.RoleAdmin:        50,
			models.RoleOrganization: 60,
			"":                      99,
		}
		if isAdmin {
			rolePriorityMap[models.RoleAdmin] = 1
			rolePriorityMap[models.RoleOrganization] = 2
			rolePriorityMap[models.RolePrincipal] = 3
		}
	}
	sort.Slice(m.Schools, func(i, j int) bool {
		pI, iEx := rolePriorityMap[m.Schools[i].RoleCode]
		pJ, jEx := rolePriorityMap[m.Schools[j].RoleCode]
		if !iEx {
			pI = 99
		}
		if !jEx {
			pJ = 99
		}
		return pI < pJ
	})

	// load school relations
	sl := []*models.School{}
	for _, s := range m.Schools {
		if s.School != nil {
			sl = append(sl, s.School)
		}
	}
	err = store.Store().SchoolsLoadRelations(context.Background(), &sl)
	if err != nil {
		return nil, err
	}

	// when archived disabled logins for non admins
	if config.Conf.AppIsReadonly != nil && *config.Conf.AppIsReadonly {
		allowedRoles := []models.Role{models.RoleAdmin, models.RoleOrganization}
		isOk := false
		for _, v := range m.Schools {
			if slices.Contains(allowedRoles, v.RoleCode) {
				isOk = true
			}
		}
		if !isOk {
			return nil, ErrPassword.SetComment("archive server")
		}
	}
	// in dev only devs are permitted to login
	if !config.Conf.AppEnvIsProd {
		isOk := false
		if isPhoneDebug(&m, false) {
			isOk = true
		}
		if !isOk {
			return nil, ErrPassword.SetComment("dev server")
		}
	}

	// check password or otp
	if req.Password != nil {
		err = LoginByPassword(&m, req)
		if err != nil {
			if ok := rateLimit("check"+*req.Username, 5, 300); !ok {
				return nil, ErrExceeded.SetComment("limit exceed (login)")
			}
			return nil, err
		}
	} else {
		isOk := false
		isOk, err = LoginByOtp(&m, req)
		if err != nil {
			return nil, err
		}
		if !isOk {
			return nil, nil
		}
	}
	// login
	return &m, nil
}

func LoginByPassword(m *models.User, req *UserLoginRequest) error {
	err := bcrypt.CompareHashAndPassword([]byte(*m.Password), []byte(*req.Password))
	if err != nil {
		return ErrPassword
	}
	return nil
}

func LoginByOtp(m *models.User, req *UserLoginRequest) (bool, error) {
	phone, err := m.FormattedPhone()
	if err != nil {
		return false, err
	}
	// debug phone login, static otp, permit minimal roles
	if req.Otp != nil {
		// rate limit sending
		if ok := rateLimit("check"+phone, 5, 60); !ok {
			return false, ErrExceeded.SetComment("limit exceed (check)")
		}
		if isPhoneDebug(m, true) && *req.Otp == "13321" {
			// ok, no error
			return true, nil
		} else if _, err := store.Store().CheckConfirmCode(context.Background(), m, *req.Otp); err != nil {
			if err == pgx.ErrNoRows {
				return false, ErrInvalid.SetKey("otp")
			}
			return false, ErrInvalid.SetKey("otp")
		} else {
			return true, nil
		}
	} else {
		// rate limit sending
		if ok := rateLimit(phone, 3, 60); !ok {
			return false, ErrExceeded.SetComment("limit exceed")
		}
		code, err := store.Store().ConfirmCodeGenerate(context.Background(), m)
		if err != nil {
			return false, err
		}
		_ = SendSMS([]string{phone}, "emekdep code: "+code, models.SmsTypeOTP)
	}

	return false, nil
}

var UsedCount map[string]int = map[string]int{}
var UsedTimestamp map[string]int = map[string]int{}

func rateLimit(key string, limit int, blockSec int) bool {
	if UsedTimestamp[key] == 0 {
		UsedTimestamp[key] = int(time.Now().Unix()) + blockSec
	}
	if UsedTimestamp[key] < int(time.Now().Unix()) {
		resetLimit(key)
	}
	UsedCount[key]++
	return UsedCount[key] <= limit
}
func resetLimit(key string) {
	delete(UsedCount, key)
	delete(UsedTimestamp, key)
}

func isPhoneDebug(m *models.User, isStrictRole bool) bool {
	hasDebugRoles := true

	if isStrictRole {
		debugRoles := []models.Role{models.RoleStudent, models.RoleParent, models.RoleTeacher}
		for _, v := range m.Schools {
			if !slices.Contains(debugRoles, v.RoleCode) {
				hasDebugRoles = false
			}
		}
	}

	phone, _ := m.FormattedPhone()
	if hasDebugRoles {
		for _, debugPhone := range config.Conf.DevPhones {
			if debugPhone == phone {
				return true
			}
		}
	}
	return false
}
