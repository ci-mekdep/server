package api

import (
	"slices"

	"github.com/gin-gonic/gin"
	"github.com/mekdep/server/internal/api/utils"
	"github.com/mekdep/server/internal/app"
	"github.com/mekdep/server/internal/app/app_validation"
	"github.com/mekdep/server/internal/models"
)

func PeriodRoutes(api *gin.RouterGroup) {
	periodRoutes := api.Group("/periods")
	{
		periodRoutes.DELETE("", PeriodDelete)
		periodRoutes.GET("", PeriodList)
		periodRoutes.PUT(":id", PeriodUpdate)
		periodRoutes.POST("", PeriodCreate)
		periodRoutes.GET(":id", PeriodDetail)
	}
}

func periodsListQuery(ses *utils.Session, data models.PeriodFilterRequest) ([]*models.PeriodResponse, int, error) {
	if data.SchoolIds == nil {
		data.SchoolIds = &[]string{}
	}
	if len(*data.SchoolIds) > 0 {
		sids := *data.SchoolIds
		data.SchoolIds = &[]string{}
		for _, sid := range sids {
			if slices.Contains(ses.GetSchoolIds(), sid) {
				*data.SchoolIds = append(*data.SchoolIds, sid)
			}
		}
	}
	if len(*data.SchoolIds) < 1 {
		*data.SchoolIds = ses.GetSchoolIds()
	}
	return app.PeriodsList(ses, data)
}

func periodsAvailableCheck(ses *utils.Session, data models.PeriodFilterRequest) (bool, error) {
	_, t, err := periodsListQuery(ses, data)
	if err != nil {
		return false, app.ErrRequired.SetKey("id")
	}
	return t > 0, nil
}

func PeriodList(c *gin.Context) {
	ses := utils.InitSession(c)
	err := app.Ap().UserActionCheckRead(&ses, app.PermAdminPeriods, func(user *models.User) error {
		r := models.PeriodFilterRequest{}
		if errMsg, errKey := BindAndValidate(c, &r); errMsg != "" || errKey != "" {
			return app.NewAppError(errMsg, errKey, "")
		}
		if r.Limit == nil {
			r.Limit = new(int)
			*r.Limit = 12
		}
		if r.Offset == nil {
			r.Offset = new(int)
			*r.Offset = 0
		}

		periods, total, err := periodsListQuery(&ses, r)
		if err != nil {
			return err
		}
		Success(c, gin.H{
			"total":   total,
			"periods": periods,
		})
		return nil
	})
	if err != nil {
		handleError(c, err)
		return
	}
}

func PeriodDetail(c *gin.Context) {
	ses := utils.InitSession(c)
	err := app.Ap().UserActionCheckRead(&ses, app.PermAdminPeriods, func(user *models.User) error {
		id := c.Param("id")
		if ok, err := periodsAvailableCheck(&ses, models.PeriodFilterRequest{ID: &id}); err != nil {
			return err
		} else if !ok {
			return app.ErrNotfound
		}
		m, err := app.PeriodDetail(&ses, id)
		if err != nil {
			return err
		}
		Success(c, gin.H{
			"period": m,
		})
		return nil
	})
	if err != nil {
		handleError(c, err)
		return
	}
}

func PeriodUpdate(c *gin.Context) {
	ses := utils.InitSession(c)
	err := app.Ap().UserActionCheckWrite(&ses, app.PermAdminPeriods, func(user *models.User) (err error) {
		r := models.PeriodRequest{}
		if err := BindAny(c, &r); err != nil {
			return err
		}
		id := c.Param("id")
		r.ID = &id
		if id == "" {
			return app.ErrRequired.SetKey("id")
		}
		err = app_validation.ValidatePeriodsCreate(&ses, r)
		if err != nil {
			return err
		}

		if ok, err := periodsAvailableCheck(&ses, models.PeriodFilterRequest{ID: r.ID}); err != nil {
			return err
		} else if !ok {
			return app.ErrForbidden
		}

		period, err := app.PeriodsUpdate(&ses, &r)
		if err != nil {
			return err
		}
		Success(c, gin.H{
			"period": period,
		})
		return nil
	})
	if err != nil {
		handleError(c, err)
		return
	}
}

func PeriodCreate(c *gin.Context) {
	ses := utils.InitSession(c)
	err := app.Ap().UserActionCheckWrite(&ses, app.PermAdminPeriods, func(user *models.User) (err error) {
		r := models.PeriodRequest{}
		if err := BindAny(c, &r); err != nil {
			return err
		}

		err = app_validation.ValidatePeriodsCreate(&ses, r)
		if err != nil {
			return err
		}
		period, err := app.PeriodsCreate(&ses, &r)
		if err != nil {
			return err
		}
		Success(c, gin.H{
			"period": period,
		})
		return nil
	})
	if err != nil {
		handleError(c, err)
		return
	}

}

func PeriodDelete(c *gin.Context) {
	ses := utils.InitSession(c)
	err := app.Ap().UserActionCheckWrite(&ses, app.PermAdminPeriods, func(user *models.User) error {
		var ids []string = c.QueryArray("ids")
		if len(ids) == 0 {
			return app.ErrRequired.SetKey("ids")
		}

		if ok, err := periodsAvailableCheck(&ses, models.PeriodFilterRequest{Ids: &ids}); err != nil {
			return err
		} else if !ok {
			return app.ErrForbidden
		}

		periods, err := app.PeriodsDelete(&ses, ids)
		if err != nil {
			return err
		}
		Success(c, gin.H{
			"periods": periods,
		})
		return nil
	})
	if err != nil {
		handleError(c, err)
		return
	}
}
