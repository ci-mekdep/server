package api

import (
	"slices"

	"github.com/gin-gonic/gin"
	"github.com/mekdep/server/internal/api/utils"
	"github.com/mekdep/server/internal/app"
	"github.com/mekdep/server/internal/app/app_validation"
	"github.com/mekdep/server/internal/models"
)

func ShiftRoutes(api *gin.RouterGroup) {
	shiftRoutes := api.Group("/shifts")
	{
		shiftRoutes.GET("", ShiftList)
		shiftRoutes.PUT(":id", ShiftUpdate)
		shiftRoutes.POST("", ShiftCreate)
		shiftRoutes.DELETE("", ShiftsDelete)
		shiftRoutes.GET(":id", ShiftDetail)
	}
}

func shiftListQuery(ses *utils.Session, data models.ShiftFilterRequest) ([]*models.ShiftResponse, int, error) {
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
	return app.ShiftList(ses, data)
}

func shiftAvailableCheck(ses *utils.Session, data models.ShiftFilterRequest) (bool, error) {
	_, t, err := shiftListQuery(ses, data)
	if err != nil {
		return false, app.ErrRequired.SetKey("id")
	}
	return t > 0, nil
}

func ShiftList(c *gin.Context) {
	ses := utils.InitSession(c)
	err := app.Ap().UserActionCheckRead(&ses, app.PermAdminShifts, func(user *models.User) error {
		r := models.ShiftFilterRequest{}
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

		shifts, total, err := shiftListQuery(&ses, r)
		if err != nil {
			return err
		}

		Success(c, gin.H{
			"shifts": shifts,
			"total":  total,
		})
		return nil
	})
	if err != nil {
		handleError(c, err)
		return
	}

}

func ShiftDetail(c *gin.Context) {
	ses := utils.InitSession(c)
	err := app.Ap().UserActionCheckRead(&ses, app.PermAdminShifts, func(u *models.User) error {
		id := c.Param("id")
		if ok, err := shiftAvailableCheck(&ses, models.ShiftFilterRequest{ID: &id}); err != nil {
			return err
		} else if !ok {
			return app.ErrNotfound
		}
		m, err := app.ShiftDetail(&ses, id)
		if err != nil {
			return err
		}
		Success(c, gin.H{
			"shift": m,
		})
		return nil
	})
	if err != nil {
		handleError(c, err)
		return
	}
}

func ShiftUpdate(c *gin.Context) {
	ses := utils.InitSession(c)
	err := app.Ap().UserActionCheckWrite(&ses, app.PermAdminShifts, func(user *models.User) (err error) {
		r := models.ShiftRequest{}
		if errMsg, errKey := BindAndValidate(c, &r); errMsg != "" || errKey != "" {
			return app.NewAppError(errMsg, errKey, "")
		}
		id := c.Param("id")
		r.Id = &id

		if id == "" {
			return app.ErrRequired.SetKey("id")
		}
		err = app_validation.ValidateShiftsCreate(&ses, r)
		if err != nil {
			return err
		}

		if ok, err := shiftAvailableCheck(&ses, models.ShiftFilterRequest{ID: r.Id}); err != nil {
			return err
		} else if !ok {
			return app.ErrForbidden
		}
		r.UpdatedBy = user.ID

		shift, err := app.UpdateShift(&ses, r)

		if err != nil {
			return err
		}
		userLog(models.UserLog{
			SchoolId:          ses.GetSchoolId(),
			SessionId:         ses.GetSessionId(),
			UserId:            user.ID,
			SubjectId:         r.Id,
			Subject:           models.LogSubjectShift,
			SubjectAction:     models.LogActionUpdate,
			SubjectProperties: r,
		})

		Success(c, gin.H{
			"shift": shift,
		})
		return nil
	})
	if err != nil {
		handleError(c, err)
		return
	}

}

func ShiftCreate(c *gin.Context) {
	ses := utils.InitSession(c)
	err := app.Ap().UserActionCheckWrite(&ses, app.PermAdminShifts, func(user *models.User) (err error) {
		r := models.ShiftRequest{}
		if err := BindAny(c, &r); err != nil {
			return err
		}
		r.UpdatedBy = user.ID
		err = app_validation.ValidateShiftsCreate(&ses, r)
		if err != nil {
			return err
		}
		shift, err := app.CreateShift(&ses, r)
		if err != nil {
			return err
		}
		userLog(models.UserLog{
			SchoolId:          ses.GetSchoolId(),
			SessionId:         ses.GetSessionId(),
			UserId:            user.ID,
			SubjectId:         &shift.Id,
			Subject:           models.LogSubjectShift,
			SubjectAction:     models.LogActionCreate,
			SubjectProperties: r,
		})
		Success(c, gin.H{
			"shift": shift,
		})
		return nil
	})
	if err != nil {
		handleError(c, err)
		return
	}
}

func ShiftsDelete(c *gin.Context) {
	ses := utils.InitSession(c)
	err := app.Ap().UserActionCheckWrite(&ses, app.PermAdminShifts, func(user *models.User) error {
		var ids []string = c.QueryArray("ids")
		if len(ids) == 0 {
			return app.ErrRequired.SetKey("ids")
		}

		if ok, err := shiftAvailableCheck(&ses, models.ShiftFilterRequest{IDs: &ids}); err != nil {
			return err
		} else if !ok {
			return app.ErrForbidden
		}

		shifts, err := app.ShiftsDelete(&ses, ids)
		if err != nil {
			return err
		}
		userLog(models.UserLog{
			SchoolId:          ses.GetSchoolId(),
			SessionId:         ses.GetSessionId(),
			UserId:            user.ID,
			Subject:           models.LogSubjectShift,
			SubjectAction:     models.LogActionDelete,
			SubjectProperties: ids,
		})
		Success(c, gin.H{
			"shifts": shifts,
		})
		return nil
	})
	if err != nil {
		handleError(c, err)
		return
	}
}
