package api

import (
	"github.com/gin-gonic/gin"
	"github.com/mekdep/server/internal/api/utils"
	"github.com/mekdep/server/internal/app"
	"github.com/mekdep/server/internal/app/app_validation"
	"github.com/mekdep/server/internal/models"
)

func TeacherExcuseRoutes(api *gin.RouterGroup) {
	teacherExcuseRoutes := api.Group("/users")
	{
		teacherExcuseRoutes.GET("teacher-excuses", TeacherExcuseList)
		teacherExcuseRoutes.GET("teacher-excuses/:id", TeacherExcuseDetail)
		teacherExcuseRoutes.POST("teacher-excuses", TeacherExcuseCreate)
		teacherExcuseRoutes.PUT("teacher-excuses/:id", TeacherExcuseUpdate)
		teacherExcuseRoutes.DELETE("teacher-excuses", TeacherExcuseDelete)
	}
}

func teacherExcuseAvailableCheck(ses *utils.Session, data models.TeacherExcuseQueryDto) (bool, error) {
	t, err := app.TeacherExcuseList(ses, data)
	if err != nil {
		return false, app.ErrRequired.SetKey("id")
	}
	return t.Total > 0, nil
}

func TeacherExcuseList(c *gin.Context) {
	ses := utils.InitSession(c)
	err := app.Ap().UserActionCheckRead(&ses, app.PermAdminTeacherExcuses, func(user *models.User) (err error) {
		req := models.TeacherExcuseQueryDto{}
		if err = BindAny(c, &req); err != nil {
			return err
		}
		err = app_validation.ValidateTeacherExcuseQuery(&ses, req)
		if err != nil {
			return err
		}
		if req.SchoolId == nil && ses.GetSchoolId() != nil {
			req.SchoolId = ses.GetSchoolId()
		}
		response, err := app.TeacherExcuseList(&ses, req)
		if err != nil {
			return err
		}
		Success(c, gin.H{
			"teacher_excuses": response.TeacherExcuses,
			"total":           response.Total,
		})
		return
	})
	if err != nil {
		handleError(c, err)
		return
	}
}

func TeacherExcuseDetail(c *gin.Context) {
	ses := utils.InitSession(c)
	err := app.Ap().UserActionCheckRead(&ses, app.PermAdminTeacherExcuses, func(u *models.User) (err error) {
		req := models.TeacherExcuseQueryDto{}
		id := c.Param("id")
		if id == "" {
			return app.ErrNotfound
		}
		if ok, err := teacherExcuseAvailableCheck(&ses, models.TeacherExcuseQueryDto{ID: &id}); err != nil {
			return err
		} else if !ok {
			return app.ErrNotfound
		}
		req.ID = &id
		response, err := app.TeacherExcuseDetail(&ses, id)
		if err != nil {
			return err
		}
		Success(c, gin.H{
			"teacher_excuse": response,
		})
		return nil
	})
	if err != nil {
		handleError(c, err)
		return
	}
}

func TeacherExcuseCreate(c *gin.Context) {
	ses := utils.InitSession(c)
	err := app.Ap().UserActionCheckWrite(&ses, app.PermAdminTeacherExcuses, func(user *models.User) (err error) {
		// bind request
		req := models.TeacherExcuseCreateDto{}
		if err = BindAny(c, &req); err != nil {
			return err
		}
		req.SchoolId = ses.GetSchoolId()
		// validate
		if err = app_validation.ValidateTeacherExcuseCreate(&ses, req); err != nil {
			return err
		}

		// upload
		documentFiles, err := handleFilesUpload(c, "files", "teacher_excuses")
		if err != nil {
			return app.NewAppError(err.Error(), "files", "")
		}

		if documentFiles != nil {
			req.DocumentFiles = documentFiles
		}

		// handle
		teacherExcuse, err := app.TeacherExcuseCreate(&ses, req)
		if err != nil {
			return err
		}
		Success(c, gin.H{
			"teacher_excuse": teacherExcuse,
		})
		return
	})
	if err != nil {
		handleError(c, err)
		return
	}
}

func TeacherExcuseUpdate(c *gin.Context) {
	ses := utils.InitSession(c)
	err := app.Ap().UserActionCheckWrite(&ses, app.PermAdminTeacherExcuses, func(user *models.User) (err error) {
		// bind request
		req := models.TeacherExcuseCreateDto{}
		if err = BindAny(c, &req); err != nil {
			return err
		}
		id := c.Param("id")
		// validate
		if id == "" {
			return app.ErrRequired.SetKey("id")
		}
		req.SchoolId = ses.GetSchoolId()
		// validate
		if err = app_validation.ValidateTeacherExcuseCreate(&ses, req); err != nil {
			return err
		}

		// upload
		documentFiles, err := handleFilesUpload(c, "files", "teacher_excuses")
		if err != nil {
			return app.NewAppError(err.Error(), "galleries", "")
		}

		if documentFiles != nil {
			req.DocumentFiles = documentFiles
		}
		// handle
		resp, err := app.TeacherExcuseUpdate(&ses, id, req)
		if err != nil {
			return err
		}
		Success(c, gin.H{
			"teacher_excuse": resp,
		})
		return
	})
	if err != nil {
		handleError(c, err)
		return
	}
}

func TeacherExcuseDelete(c *gin.Context) {
	ses := utils.InitSession(c)
	err := app.Ap().UserActionCheckWrite(&ses, app.PermAdminTeacherExcuses, func(user *models.User) (err error) {
		// bind request
		ids := c.QueryArray("ids")
		// validate
		if len(ids) == 0 {
			return app.ErrRequired.SetKey("id")
		}
		if ok, err := teacherExcuseAvailableCheck(&ses, models.TeacherExcuseQueryDto{Ids: ids}); err != nil {
			return err
		} else if !ok {
			return app.ErrForbidden
		}
		// handle
		resp, err := app.TeacherExcuseDelete(&ses, ids)
		if err != nil {
			return err
		}
		Success(c, gin.H{
			"teacher_excuses": resp.TeacherExcuses,
		})
		return
	})
	if err != nil {
		handleError(c, err)
		return
	}
}
