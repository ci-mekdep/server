package utils

import (
	"context"
	"errors"
	"log"
	"slices"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/mekdep/server/config"
	"github.com/mekdep/server/internal/models"
	"github.com/mekdep/server/internal/store"
)

var st []models.Session

func SessionStoreInit() error {
	st = []models.Session{}
	// load all sessions from DB
	var err error
	st, err = store.Store().SessionsSelect(context.Background(), models.SessionFilter{})
	if err != nil {
		return err
	}
	// load relation users
	userIds := []string{}
	for _, v := range st {
		userIds = append(userIds, v.UserId)
	}
	users, err := store.Store().UsersFindByIds(context.Background(), userIds)
	if err != nil {
		return err
	}
	err = store.Store().UsersLoadRelations(context.Background(), &users, false)
	if err != nil {
		return err
	}
	for k := range st {
		for _, u := range users {
			if u.ID == st[k].UserId {
				st[k].User = *u
				break
			}
		}
	}

	return err
}

func AddSession(c *gin.Context, claims jwt.MapClaims, userModel models.User, deviceToken *string) (models.Session, error) {
	appIsReadonly := config.Conf.AppIsReadonly
	config.Conf.AppIsReadonly = new(bool)
	*config.Conf.AppIsReadonly = false
	ses, err := store.Store().SessionsCreate(context.Background(), models.Session{
		DeviceToken: deviceToken,
		Token:       claims["token"].(string),
		UserId:      userModel.ID,
		Ip:          c.ClientIP(),
		Agent:       c.Request.UserAgent(),
		Exp:         time.Unix(claims["exp"].(int64), 0),
		Iat:         time.Now(),
		Lat:         time.Now(),
	})
	ses.User = userModel
	config.Conf.AppIsReadonly = appIsReadonly
	st = append(st, ses)
	return ses, err
}

func SessionDelete(ses models.Session) {
	rk := -1
	for k, v := range st {
		if v.ID == ses.ID {
			rk = k
		}
	}
	if rk >= 0 {
		st = append(st[:rk], st[rk+1:]...)
	}
	_ = store.Store().SessionsDelete(context.Background(), models.SessionFilter{
		ID: &ses.ID,
	})
}

func SessionDeleteByUserId(userId string) error {
	if st == nil {
		return errors.New("session store not set")
	}
	newSt := []models.Session{}
	for _, session := range st {
		if session.UserId != userId {
			newSt = append(newSt, session)
		}
	}
	st = newSt

	_ = store.Store().SessionsDelete(context.Background(), models.SessionFilter{
		UserId: &userId,
	})

	return nil
}

func GetLastSession(userId string) (*models.Session, error) {
	if st == nil {
		return nil, errors.New("session store not set")
	}

	var lastSession *models.Session
	for _, v := range st {
		if v.UserId == userId {
			lastSession = &v
		}
	}

	return lastSession, nil
}

func SessionByToken(token string) (ses models.Session, err error) {
	if st == nil {
		err = errors.New("session store not set")
		return
	}

	for _, v := range st {
		if v.Token == token {
			ses = v
			return
		}
	}
	err = errors.New("session not found")
	return
}

func SessionActByToken(token string, lat time.Time) (err error) {
	if st == nil {
		err = errors.New("session store not set")
		return
	}

	for k, v := range st {
		if v.Token == token {
			st[k].Lat = lat
			return
		}
	}
	return
}

func SessionOnlineCount(schoolId *uint, min int) int {
	c := 0
	now := time.Now().Add(time.Hour * (-5)).Add(time.Minute * time.Duration(-min))
	log.Println(now)
	for _, v := range st {
		if v.Iat.After(now) {
			c++
		}
	}
	return c
}

func SessionByUserId(userId string) ([]models.Session, error) {
	if st == nil {
		return nil, errors.New("session store not set")
	}
	s := []models.Session{}

	for _, v := range st {
		if v.UserId == userId {
			s = append(s, v)
		}
	}
	return s, nil
}

func SessionByUserIds(userIds []string) ([]models.Session, error) {
	if st == nil {
		return nil, errors.New("session store not set")
	}
	s := []models.Session{}

	for _, v := range st {
		if slices.Contains(userIds, v.UserId) {
			s = append(s, v)
		}
	}
	return s, nil
}
