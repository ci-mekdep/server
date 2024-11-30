package api

import (
	"github.com/gin-gonic/gin"
	"github.com/mekdep/server/internal/api/utils"
	"github.com/mekdep/server/internal/app"
	"github.com/mekdep/server/internal/app/app_validation"
	"github.com/mekdep/server/internal/models"
)

func SubjectExamRoutes(api *gin.RouterGroup) {
	r := api.Group("/subjects/exams")
	{
		r.GET("", SubjectExamList)
		r.POST("", SubjectExamCreate)
		r.PUT(":id", SubjectExamUpdate)
		r.DELETE("", SubjectExamDelete)
	}
}

func subjectExamListQuery(ses *utils.Session, data models.SubjectExamFilterRequest) ([]*models.SubjectExamResponse, int, error) {
	res, total, err := app.SubjectExamList(ses, &data)
	return res, total, err
}

func subjectExamAvailableCheck(ses *utils.Session, data models.SubjectExamFilterRequest) (bool, error) {
	_, t, err := subjectExamListQuery(ses, data)
	if err != nil {
		return false, app.ErrRequired.SetKey("id")
	}
	return t > 0, nil
}

// TODO: exam list load relations only: school, subject, teacher; not all of them;
func SubjectExamList(c *gin.Context) {
	ses := utils.InitSession(c)
	err := app.Ap().UserActionCheckRead(&ses, app.PermAdminSubjectsExams, func(user *models.User) error {
		r := models.SubjectExamFilterRequest{}
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
		l, total, err := subjectExamListQuery(&ses, r)
		if err != nil {
			return err
		}
		Success(c, gin.H{
			"total":         total,
			"subject_exams": l,
		})
		return nil
	})
	if err != nil {
		handleError(c, err)
		return
	}
}

func SubjectExamUpdate(c *gin.Context) {
	ses := utils.InitSession(c)
	err := app.Ap().UserActionCheckWrite(&ses, app.PermAdminSubjectsExams, func(user *models.User) (err error) {
		r := models.SubjectExamRequest{}
		if errMsg, errKey := BindAndValidate(c, &r); errMsg != "" || errKey != "" {
			return app.NewAppError(errMsg, errKey, "")
		}
		id := c.Param("id")
		r.ID = &id
		if id == "" {
			return app.ErrRequired.SetKey("id")
		}
		err = app_validation.ValidateSubjectExamCreate(&ses, r)
		if err != nil {
			return err
		}
		if ok, err := subjectExamAvailableCheck(&ses, models.SubjectExamFilterRequest{ID: r.ID}); err != nil {
			return err
		} else if !ok {
			return app.ErrForbidden
		}

		m, err := app.SubjectExamUpdate(&ses, &r)
		if err != nil {
			return err
		}
		Success(c, gin.H{
			"subject_exams": m,
		})
		return nil
	})
	if err != nil {
		handleError(c, err)
		return
	}

}

func SubjectExamCreate(c *gin.Context) {
	ses := utils.InitSession(c)
	err := app.Ap().UserActionCheckWrite(&ses, app.PermAdminSubjectsExams, func(user *models.User) (err error) {
		r := models.SubjectExamRequest{}
		if err := BindAny(c, &r); err != nil {
			return err
		}

		err = app_validation.ValidateSubjectExamCreate(&ses, r)
		if err != nil {
			return err
		}
		m, err := app.SubjectExamCreate(&ses, &r)
		if err != nil {
			return err
		}
		Success(c, gin.H{
			"subject_exams": m,
		})
		return nil
	})
	if err != nil {
		handleError(c, err)
		return
	}
}

func SubjectExamDelete(c *gin.Context) {
	ses := utils.InitSession(c)
	err := app.Ap().UserActionCheckWrite(&ses, app.PermAdminSubjectsExams, func(user *models.User) error {
		var ids []string = c.QueryArray("ids")
		if len(ids) == 0 {
			return app.ErrRequired.SetKey("ids")
		}

		if ok, err := subjectExamAvailableCheck(&ses, models.SubjectExamFilterRequest{Ids: &ids}); err != nil {
			return err
		} else if !ok {
			return app.ErrNotfound
		}

		l, err := app.SubjectExamDelete(&ses, ids)
		if err != nil {
			return err
		}
		Success(c, gin.H{
			"subject_exams": l,
		})
		return nil
	})
	if err != nil {
		handleError(c, err)
		return
	}
}
