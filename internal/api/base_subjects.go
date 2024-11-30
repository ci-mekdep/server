package api

import (
	"slices"

	"github.com/gin-gonic/gin"
	"github.com/mekdep/server/internal/api/utils"
	"github.com/mekdep/server/internal/app"
	"github.com/mekdep/server/internal/app/app_validation"
	"github.com/mekdep/server/internal/models"
)

func BaseSubjectsRoutes(api *gin.RouterGroup) {
	baseSubjectsRoutes := api.Group("/base-subjects")
	{
		baseSubjectsRoutes.GET("", BaseSubjectsList)
		baseSubjectsRoutes.PUT(":id", BaseSubjectsUpdate)
		baseSubjectsRoutes.POST("", BaseSubjectsCreate)
		baseSubjectsRoutes.DELETE("", BaseSubjectsDelete)
		baseSubjectsRoutes.GET(":id", BaseSubjectsDetail)
	}
}

func baseSubjectsListQuery(ses *utils.Session, data models.BaseSubjectsFilterRequest) ([]*models.BaseSubjectsResponse, int, error) {
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
	res, total, err := app.BaseSubjectsList(ses, data)
	return res, total, err
}

func baseSubjectsAvailableCheck(ses *utils.Session, data models.BaseSubjectsFilterRequest) (bool, error) {
	_, t, err := baseSubjectsListQuery(ses, data)
	if err != nil {
		return false, app.ErrRequired.SetKey("id")
	}
	return t > 0, nil
}

func BaseSubjectsList(c *gin.Context) {
	ses := utils.InitSession(c)
	err := app.Ap().UserActionCheckRead(&ses, app.PermUser, func(user *models.User) error {
		r := models.BaseSubjectsFilterRequest{}
		if errMsg, errKey := BindAndValidate(c, &r); errMsg != "" || errKey != "" {
			return app.NewAppError(errMsg, errKey, "")
		}

		items, total, err := baseSubjectsListQuery(&ses, r)
		if err != nil {
			return err
		}

		Success(c, gin.H{
			"base_subjects": items,
			"total":         total,
		})
		return nil
	})
	if err != nil {
		handleError(c, err)
		return
	}
}

func BaseSubjectsDetail(c *gin.Context) {
	ses := utils.InitSession(c)
	err := app.Ap().UserActionCheckRead(&ses, app.PermAdminSubjects, func(u *models.User) error {
		id := c.Param("id")
		if ok, err := baseSubjectsAvailableCheck(&ses, models.BaseSubjectsFilterRequest{ID: id}); err != nil {
			return err
		} else if !ok {
			return app.ErrNotfound
		}
		m, err := app.BaseSubjectsDetail(&ses, id)
		if err != nil {
			return err
		}
		Success(c, gin.H{
			"base_subject": m,
		})
		return nil
	})
	if err != nil {
		handleError(c, err)
		return
	}
}

func BaseSubjectsUpdate(c *gin.Context) {
	ses := utils.InitSession(c)
	err := app.Ap().UserActionCheckWrite(&ses, app.PermAdminSubjects, func(user *models.User) (err error) {
		r := models.BaseSubjectsRequest{}
		if errMsg, errKey := BindAndValidate(c, &r); errMsg != "" || errKey != "" {
			return app.NewAppError(errMsg, errKey, "")
		}
		id := c.Param("id")
		r.ID = id

		if id == "" {
			return app.ErrRequired.SetKey("id")
		}
		if ok, err := baseSubjectsAvailableCheck(&ses, models.BaseSubjectsFilterRequest{ID: r.ID}); err != nil {
			return err
		} else if !ok {
			return app.ErrForbidden
		}
		book, err := app.BaseSubjectsUpdate(&ses, r)

		if err != nil {
			return err
		}
		userLog(models.UserLog{
			SchoolId:          ses.GetSchoolId(),
			SessionId:         ses.GetSessionId(),
			UserId:            user.ID,
			SubjectId:         &r.ID,
			Subject:           models.LogSubjectBaseSubjects,
			SubjectAction:     models.LogActionUpdate,
			SubjectProperties: r,
		})

		Success(c, gin.H{
			"base_subject": book,
		})
		return nil
	})
	if err != nil {
		handleError(c, err)
		return
	}

}

func BaseSubjectsCreate(c *gin.Context) {
	ses := utils.InitSession(c)
	err := app.Ap().UserActionCheckWrite(&ses, app.PermAdminSubjects, func(user *models.User) (err error) {
		r := models.BaseSubjectsRequest{}
		if err := BindAny(c, &r); err != nil {
			return err
		}
		err = app_validation.ValidateBaseSubjectsCreate(&ses, r)
		if err != nil {
			return err
		}
		book, err := app.BaseSubjectsCreate(&ses, r)
		if err != nil {
			return err
		}
		userLog(models.UserLog{
			SchoolId:          ses.GetSchoolId(),
			SessionId:         ses.GetSessionId(),
			UserId:            user.ID,
			SubjectId:         &book.ID,
			Subject:           models.LogSubjectBaseSubjects,
			SubjectAction:     models.LogActionCreate,
			SubjectProperties: r,
		})
		Success(c, gin.H{
			"base_subject": book,
		})
		return nil
	})
	if err != nil {
		handleError(c, err)
		return
	}
}

func BaseSubjectsDelete(c *gin.Context) {
	ses := utils.InitSession(c)
	err := app.Ap().UserActionCheckWrite(&ses, app.PermAdminSubjects, func(user *models.User) error {
		var ids []string = c.QueryArray("ids")
		if len(ids) == 0 {
			return app.ErrRequired.SetKey("ids")
		}

		if ok, err := baseSubjectsAvailableCheck(&ses, models.BaseSubjectsFilterRequest{IDs: &ids}); err != nil {
			return err
		} else if !ok {
			return app.ErrForbidden
		}

		items, err := app.BaseSubjectsDelete(&ses, ids)
		if err != nil {
			return err
		}
		userLog(models.UserLog{
			SchoolId:          ses.GetSchoolId(),
			SessionId:         ses.GetSessionId(),
			UserId:            user.ID,
			Subject:           models.LogSubjectBaseSubjects,
			SubjectAction:     models.LogActionDelete,
			SubjectProperties: ids,
		})
		Success(c, gin.H{
			"base_subjects": items,
		})
		return nil
	})
	if err != nil {
		handleError(c, err)
		return
	}
}
