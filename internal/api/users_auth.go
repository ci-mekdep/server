package api

import (
	"slices"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/mekdep/server/config"
	"github.com/mekdep/server/internal/api/utils"
	"github.com/mekdep/server/internal/app"
	"github.com/mekdep/server/internal/models"
)

type UserLoginResponse struct {
	User               *models.UserResponse    `json:"user"`
	CurrentRole        *models.Role            `json:"current_role"`
	CurrentSchoolId    *string                 `json:"current_school"`
	CurrentSchoolModel *models.SchoolResponse  `json:"current_school_model"`
	CurrentRegionModel *models.SchoolResponse  `json:"current_region_model"`
	CurrentPeriodModel *models.PeriodResponse  `json:"current_period_model"`
	Token              string                  `json:"new_token"`
	LastSession        *models.SessionResponse `json:"last_session"`
	SessionId          *string                 `json:"session_id"`
	ExpiresAt          int64                   `json:"expires_at"`
}

func UserLogin(c *gin.Context) {
	r := app.UserLoginRequest{}
	if errMsg, errKey := BindAndValidate(c, &r); errMsg != "" || errKey != "" {
		handleError(c, app.NewAppError(errMsg, errKey, ""))
		return
	}

	isMobile := r.SchoolID != nil && *r.SchoolID != ""
	model, err := app.Login(&r, !isMobile)
	if err != nil {
		handleError(c, err)
		return
	}
	if model == nil {
		Success(c, gin.H{})
		return
	}

	// select login role
	role := new(models.Role)
	var currentSchoolId *string
	var currentSchool *models.School
	var currentRegion *models.School
	var sesId *string
	if model == nil || len(model.Schools) < 0 {
		handleError(c, app.ErrPassword.SetComment("no school"))
		return
	}
	for _, v := range model.Schools {
		if role == nil {
			// init value for first iteration
			role = &v.RoleCode
			if v.School != nil {
				currentSchoolId = &v.School.ID
				currentSchool = v.School
			}
		}
		if r.SchoolID != nil && v.School != nil && v.School.ID != "" &&
			v.School.ID == *r.SchoolID ||
			v.School == nil || r.SchoolID == nil {
			role = &v.RoleCode
			if v.School != nil {
				currentSchoolId = &v.School.ID
				currentSchool = v.School
			}
			break
		}
		if currentSchool != nil && currentSchool.Parent == nil {
			currentSchool = nil
			currentSchoolId = nil
			currentRegion = currentSchool
		}
	}

	// admins can login only with one device
	var lastSession *models.Session
	var lastSessionRes *models.SessionResponse
	if !config.Conf.AdminMultiDevice {
		if slices.Contains([]models.Role{models.RoleAdmin, models.RoleOrganization, models.RolePrincipal}, *role) {
			lastSession, err = utils.GetLastSession(model.ID)
			if err != nil {
				handleError(c, err)
				return
			}
			err := utils.SessionDeleteByUserId(model.ID)
			if err != nil {
				handleError(c, err)
				return
			}
		}
		if lastSession != nil {
			lastSessionRes = &models.SessionResponse{}
			lastSessionRes.FromModel(lastSession)
		}
	}

	// token
	claims := jwt.MapClaims{}
	claims, sesId, err = utils.GenerateToken(c, *model, role, currentSchoolId, nil, r.DeviceToken)
	if err != nil {
		handleError(c, err)
		return
	}

	userLog(models.UserLog{
		SchoolId:          currentSchoolId,
		SessionId:         sesId,
		UserId:            model.ID,
		Subject:           models.LogSubjectUsers,
		SubjectAction:     models.LogActionLogin,
		SubjectProperties: r,
	})
	// response
	userRes := &models.UserResponse{}
	userRes.FromModel(model)
	currentSchoolRes := &models.SchoolResponse{}
	if currentSchool != nil {
		currentSchoolRes.FromModel(currentSchool)
	} else {
		currentSchoolRes = nil
	}
	currentRegionRes := &models.SchoolResponse{}
	if currentRegion != nil {
		currentRegionRes.FromModel(currentRegion)
	} else {
		currentRegionRes = nil
	}

	resp := UserLoginResponse{
		User:               userRes,
		CurrentRole:        role,
		CurrentSchoolId:    currentSchoolId,
		CurrentSchoolModel: currentSchoolRes,
		CurrentRegionModel: currentRegionRes,
		SessionId:          sesId,
		LastSession:        lastSessionRes,
		Token:              (claims["token"]).(string),
		ExpiresAt:          (claims["exp"]).(int64),
	}
	Success(c, gin.H{
		"user":                 resp.User,
		"current_role":         resp.CurrentRole,
		"current_school":       resp.CurrentSchoolId,
		"current_school_model": resp.CurrentSchoolModel,
		"current_region_model": resp.CurrentRegionModel,
		"current_period_model": nil,
		"session_id":           resp.SessionId,
		"last_session":         resp.LastSession,
		"token":                resp.Token,
		"expires_at":           resp.ExpiresAt,
	})
}
