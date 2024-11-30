package api

import (
	"slices"

	"github.com/gin-gonic/gin"
	"github.com/mekdep/server/internal/api/utils"
	"github.com/mekdep/server/internal/app"
	"github.com/mekdep/server/internal/app/app_validation"
	"github.com/mekdep/server/internal/models"
)

func TimetableRoutes(api *gin.RouterGroup) {
	r := api.Group("/timetables")
	{
		r.GET("", TimetableList)
		r.GET(":id", TimetableDetail)
		r.PUT(":id", TimetableUpdate)
		r.POST("", TimetableCreate)
		r.DELETE("", TimetableDelete)
	}
}

func timetableListQuery(ses *utils.Session, data models.TimetableFilterRequest) ([]*models.TimetableResponse, int, error) {
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
	return app.Ap().TimetablesList(ses, data)
}

func timetableAvailableCheck(ses *utils.Session, data models.TimetableFilterRequest) (bool, error) {
	_, t, err := timetableListQuery(ses, data)
	if err != nil {
		return false, err
	}
	return t > 0, nil
}

func TimetableList(c *gin.Context) {
	ses := utils.InitSession(c)
	err := app.Ap().UserActionCheckRead(&ses, app.PermAdminTimetables, func(user *models.User) error {
		r := models.TimetableFilterRequest{}
		if errMsg, errKey := BindAndValidate(c, &r); errMsg != "" || errKey != "" {
			return app.NewAppError(errMsg, errKey, "")
		}

		l, total, err := timetableListQuery(&ses, r)
		if err != nil {
			return err
		}
		Success(c, gin.H{
			"total":      total,
			"timetables": l,
		})
		return nil
	})
	if err != nil {
		handleError(c, err)
		return
	}
}

func TimetableDetail(c *gin.Context) {
	ses := utils.InitSession(c)
	err := app.Ap().UserActionCheckRead(&ses, app.PermAdminTimetables, func(user *models.User) error {
		id := c.Param("id")
		if ok, err := timetableAvailableCheck(&ses, models.TimetableFilterRequest{ID: &id}); err != nil {
			return err
		} else if !ok {
			return app.ErrNotfound
		}
		m, err := app.TimetableDetail(&ses, id)
		if err != nil {
			return err
		}
		Success(c, gin.H{
			"timetable": m,
		})
		return nil
	})
	if err != nil {
		handleError(c, err)
		return
	}
}

func TimetableUpdate(c *gin.Context) {
	ses := utils.InitSession(c)
	err := app.Ap().UserActionCheckWrite(&ses, app.PermAdminTimetables, func(user *models.User) (err error) {
		r := models.TimetableRequest{}
		if errMsg, errKey := BindAndValidate(c, &r); errMsg != "" || errKey != "" {
			return app.NewAppError(errMsg, errKey, "")
		}
		id := c.Param("id")
		r.ID = &id
		if id == "" {
			return app.ErrRequired.SetKey("id")
		}
		err = app_validation.ValidateTimetablesCreate(&ses, r)
		if err != nil {
			return err
		}

		if ok, err := timetableAvailableCheck(&ses, models.TimetableFilterRequest{ID: r.ID}); err != nil {
			return err
		} else if !ok {
			return app.ErrForbidden
		}

		r.UpdatedBy = user.ID
		m, err := app.Ap().TimetableUpdate(&ses, user, r)
		if err != nil {
			return err
		}
		userLog(models.UserLog{
			SchoolId:          ses.GetSchoolId(),
			SessionId:         ses.GetSessionId(),
			UserId:            user.ID,
			SubjectId:         &m.ID,
			Subject:           models.LogSubjectTimetable,
			SubjectAction:     models.LogActionUpdate,
			SubjectProperties: r,
		})
		Success(c, gin.H{
			"timetable": m,
		})
		return nil
	})
	if err != nil {
		handleError(c, err)
		return
	}
}

func TimetableCreate(c *gin.Context) {
	ses := utils.InitSession(c)
	err := app.Ap().UserActionCheckWrite(&ses, app.PermAdminTimetables, func(user *models.User) (err error) {
		r := models.TimetableRequest{}
		if err := BindAny(c, &r); err != nil {
			return err
		}
		err = app_validation.ValidateTimetablesCreate(&ses, r)
		if err != nil {
			return err
		}
		r.UpdatedBy = user.ID
		m, err := app.Ap().TimetableCreate(&ses, r)
		if err != nil {
			return err
		}
		userLog(models.UserLog{
			SchoolId:          ses.GetSchoolId(),
			SessionId:         ses.GetSessionId(),
			UserId:            user.ID,
			SubjectId:         &m.ID,
			Subject:           models.LogSubjectTimetable,
			SubjectAction:     models.LogActionCreate,
			SubjectProperties: r,
		})
		Success(c, gin.H{
			"timetable": m,
		})
		return nil
	})
	if err != nil {
		handleError(c, err)
		return
	}

}

func TimetableDelete(c *gin.Context) {
	ses := utils.InitSession(c)
	err := app.Ap().UserActionCheckWrite(&ses, app.PermAdminTimetables, func(user *models.User) error {
		var ids []string = c.QueryArray("ids")
		if len(ids) == 0 {
			return app.NewAppError("required", "ids", "")
		}

		if ok, err := timetableAvailableCheck(&ses, models.TimetableFilterRequest{Ids: &ids}); err != nil {
			return err
		} else if !ok {
			return app.ErrForbidden
		}

		l, err := app.Ap().TimetablesDelete(&ses, ids)
		if err != nil {
			return err
		}
		userLog(models.UserLog{
			SchoolId:          ses.GetSchoolId(),
			SessionId:         ses.GetSessionId(),
			UserId:            user.ID,
			Subject:           models.LogSubjectTimetable,
			SubjectAction:     models.LogActionDelete,
			SubjectProperties: ids,
		})
		Success(c, gin.H{
			"timetables": l,
		})
		return nil
	})
	if err != nil {
		handleError(c, err)
		return
	}
}
