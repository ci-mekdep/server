package api

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mekdep/server/internal/api/utils"
	"github.com/mekdep/server/internal/app"
	"github.com/mekdep/server/internal/models"
)

func DashboardNumbersV2(c *gin.Context) {
	ses := utils.InitSession(c)
	err := app.Ap().UserActionCheckRead(&ses, app.PermUser, func(user *models.User) error {
		r := app.DashboardNumbersRequest{}
		if errMsg, errKey := BindAndValidate(c, &r); errMsg != "" || errKey != "" {
			return app.NewAppError(errMsg, errKey, "")
		}
		d, err := app.Ap().DashboardNumbersV2Cached(&ses, r)
		if err != nil {
			return err
		}
		Success(c, gin.H{
			"dashboards": d,
		})
		return nil
	})
	if err != nil {
		handleError(c, err)
		return
	}
}

func DashboardDetailsV2(c *gin.Context) {
	ses := utils.InitSession(c)
	err := app.Ap().UserActionCheckRead(&ses, app.PermUser, func(user *models.User) error {
		now := time.Now()
		now = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		r := app.DashboardDetailsRequest{
			StartDate: &time.Time{},
			EndDate:   &time.Time{},
		}
		*r.StartDate = now.Truncate(time.Hour * 24)
		*r.EndDate = now.Truncate(time.Hour * 24)
		if errMsg, errKey := BindAndValidate(c, &r); errMsg != "" || errKey != "" {
			return app.NewAppError(errMsg, errKey, "")
		}
		d, err := app.Ap().DashboardDetailsV2(&ses, r)
		if err != nil {
			return err
		}
		Success(c, gin.H{
			"dashboards": d,
		})
		return nil
	})
	if err != nil {
		handleError(c, err)
		return
	}
}
