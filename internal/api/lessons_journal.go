package api

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/mekdep/server/internal/api/utils"
	"github.com/mekdep/server/internal/app"
	"github.com/mekdep/server/internal/models"
)

func JournalRoutes(api *gin.RouterGroup) {
	rs := api.Group("lessons")
	{
		// rs.Use(middleware.JwtTokenCheck)
		rs.GET("journal", LessonJournal)
		rs.GET("final/:subject_id", LessonFinal)
		rs.GET("final/v2", LessonFinalV2)
		rs.GET("final/subjects", LessonFinalBySubject)
		rs.GET("final", LessonFinalOld)
		rs.POST("final/:subject_id", LessonFinalMake)
		rs.POST("final/v2", LessonFinalMake)
		rs.POST("", LessonUpdate)
		rs.DELETE("", LessonUpdate)
		rs.POST("v2", LessonUpdateV2)
		rs.POST("like/:id", LessonLikes)
	}
}

func LessonJournal(c *gin.Context) {
	ses := utils.InitSession(c)
	err := app.Ap().UserActionCheckRead(&ses, app.PermJournal, func(user *models.User) error {
		r := app.LessonJournalRequest{}
		if errMsg, errKey := BindAndValidate(c, &r); errMsg != "" || errKey != "" {
			return app.NewAppError(errMsg, errKey, "")
		}
		if r.Date != nil && r.PeriodNumber != nil || r.Date == nil && r.PeriodNumber == nil {
			r.PeriodNumber = new(int)
			*r.PeriodNumber = 2
			// TODO: add checker
			// return app.ErrInvalid.SetKey("date").SetComment("date or period_number") TODO: remove
		}
		res, err := app.Ap().LessonJournal(&ses, r)
		if err != nil {
			return err
		}
		Success(c, gin.H{
			"data": res,
		})
		return nil
	})
	if err != nil {
		handleError(c, err)
		return
	}

}

func LessonUpdate(c *gin.Context) {
	ses := utils.InitSession(c)
	err := app.Ap().UserActionCheckWrite(&ses, app.PermJournal, func(user *models.User) error {
		if _, err := c.MultipartForm(); err == nil {
			LessonUpdateHandleFiles(c)
		} else {
			r := models.JournalRequest{}
			if errMsg, errKey := BindAndValidate(c, &r); errMsg != "" || errKey != "" {
				return app.NewAppError(errMsg, errKey, "")
			}
			r.Lesson.Assignment.UpdatedBy = user.ID
			if r.Grade != nil {
				r.Grade.UpdatedBy = user.ID
			}
			if r.Absent != nil {
				r.Absent.UpdatedBy = user.ID
			}
			res, resPG, err := app.Ap().LessonUpdate(&ses, &r)
			if err != nil {
				return err
			}
			Success(c, gin.H{
				"items":         res,
				"period_grades": resPG,
			})
		}
		return nil
	})
	if err != nil {
		handleError(c, err)
		return
	}

}

func LessonUpdateV2(c *gin.Context) {
	ses := utils.InitSession(c)
	err := app.Ap().UserActionCheckWrite(&ses, app.PermJournal, func(user *models.User) error {
		if _, err := c.MultipartForm(); err == nil {
			LessonUpdateFormData(c)
		} else {
			LessonUpdateJSON(c)
		}
		return nil
	})
	if err != nil {
		handleError(c, err)
		return
	}
}
func LessonUpdateJSON(c *gin.Context) {
	ses := utils.InitSession(c)
	err := app.Ap().UserActionCheckWrite(&ses, app.PermJournal, func(user *models.User) error {
		r := models.JournalRequestV2{}
		if errMsg, errKey := BindAndValidate(c, &r); errMsg != "" || errKey != "" {
			return app.NewAppError(errMsg, errKey, "")
		}
		r.Lesson.Assignment.UpdatedBy = user.ID
		if r.Grade != nil {
			r.Grade.UpdatedBy = user.ID
		}
		if r.Absent != nil {
			r.Absent.UpdatedBy = user.ID
		}
		res, resPG, err := app.Ap().LessonUpdateV2(&ses, &r)
		if err != nil {
			return err
		}
		Success(c, gin.H{
			"items":         res,
			"period_grades": resPG,
		})
		return nil
	})
	if err != nil {
		handleError(c, err)
		return
	}
}

func LessonUpdateFormData(c *gin.Context) {
	ses := utils.InitSession(c)
	err := app.Ap().UserActionCheckWrite(&ses, app.PermJournal, func(user *models.User) error {
		r := models.JournalFormRequest{}
		if errMsg, errKey := BindAndValidate(c, &r); errMsg != "" || errKey != "" {
			return app.NewAppError(errMsg, errKey, "")
		}
		if v, ok := c.Request.Form["assignment_files_delete"]; ok {
			r.AssignmentFilesDelete = &v
		}
		if v, ok := c.Request.Form["lesson_pro_files_delete"]; ok {
			r.LessonProFilesDelete = &v
		}
		ma, err := app.LessonUpdateFormLessonPro(c, &r)
		if err != nil {
			return err
		}
		ma, err = app.LessonUpdateFormAssignment(c, &r)
		if err != nil {
			return err
		}
		res := models.LessonResponse{}
		res.FromModel(ma)
		Success(c, gin.H{
			"lesson": res,
		})
		return nil
	})
	if err != nil {
		handleError(c, err)
		return
	}
}

func LessonUpdateHandleFiles(c *gin.Context) {
	ses := utils.InitSession(c)
	err := app.Ap().UserActionCheckWrite(&ses, app.PermJournal, func(user *models.User) error {
		r := models.AssignmentRequest{}
		if errMsg, errKey := BindAndValidate(c, &r); errMsg != "" || errKey != "" {
			return app.NewAppError(errMsg, errKey, "")
		}
		r.UpdatedBy = user.ID
		if v, ok := c.Request.Form["files_delete"]; ok {
			r.FilesDelete = &v
		}
		ma, err := app.LessonUpdateAssignment(c, &r)
		if err != nil {
			return err
		}
		res := models.AssignmentResponse{}
		res.FromModel(ma.Assignment)
		Success(c, gin.H{
			"assignment": res,
		})
		return nil
	})
	if err != nil {
		handleError(c, err)
		return
	}
}

func LessonFinal(c *gin.Context) {
	ses := utils.InitSession(c)
	err := app.Ap().UserActionCheckWrite(&ses, app.PermJournal, func(user *models.User) error {
		subjectId := c.Param("subject_id")
		if subjectId == "" {
			return app.ErrRequired.SetKey("subject_id")
		}
		subjectIdInt := subjectId

		res, err := app.LessonFinal(&ses, subjectIdInt)
		if err != nil {
			return err
		}
		Success(c, gin.H{
			"students": res,
		})
		return nil
	})
	if err != nil {
		handleError(c, err)
		return
	}

}

func LessonFinalV2(c *gin.Context) {
	ses := utils.InitSession(c)
	err := app.Ap().UserActionCheckWrite(&ses, app.PermUser, func(user *models.User) error {
		childId := c.Request.FormValue("child_id")
		// if childId == 0 {
		// 	return app.ErrRequired.SetKey("child_id")
		// }
		classroomId := c.Request.FormValue("classroom_id")
		// if classroomId == 0 {
		// 	return app.ErrRequired.SetKey("classroom_id")
		// }
		subjectId := c.Request.FormValue("subject_id")
		// if subjectId == 0 {
		// 	return app.ErrRequired.SetKey("subject_id")
		// }
		res, err := app.LessonFinalV2(&ses, subjectId, childId, classroomId)
		if err != nil {
			return err
		}
		Success(c, gin.H{
			"students": res,
		})
		return nil
	})
	if err != nil {
		handleError(c, err)
		return
	}
}

func LessonFinalBySubject(c *gin.Context) {
	ses := utils.InitSession(c)
	err := app.Ap().UserActionCheckWrite(&ses, app.PermUser, func(user *models.User) error {
		classroomId := c.Request.FormValue("classroom_id")
		periodNumStr := c.Request.FormValue("period_number")
		periodNum, _ := strconv.Atoi(periodNumStr)

		if periodNumStr == "" {
			return app.ErrRequired.SetKey("period_number")
		}

		res, err := app.LessonFinalBySubject(&ses, classroomId, periodNum)
		if err != nil {
			return err
		}
		Success(c, gin.H{
			"students": res.Students,
			"subjects": res.Subjects,
			"exams":    res.Exams,
		})
		return nil
	})
	if err != nil {
		handleError(c, err)
		return
	}

}

func LessonFinalOld(c *gin.Context) {
	ses := utils.InitSession(c)
	err := app.Ap().UserActionCheckWrite(&ses, app.PermJournal, func(user *models.User) error {
		subjectId := c.Request.FormValue("subject_id")
		if subjectId == "" {
			return app.ErrRequired.SetKey("subject_id")
		}

		res, err := app.LessonFinal(&ses, subjectId)
		if err != nil {
			return err
		}
		Success(c, gin.H{
			"students": res,
		})
		return nil
	})
	if err != nil {
		handleError(c, err)
		return
	}

}

func LessonFinalMake(c *gin.Context) {
	ses := utils.InitSession(c)
	err := app.Ap().UserActionCheckWrite(&ses, app.PermJournal, func(user *models.User) error {
		r := models.GradeRequest{}
		if errMsg, errKey := BindAndValidate(c, &r); errMsg != "" || errKey != "" {
			return app.NewAppError(errMsg, errKey, "")
		}
		subjectId := c.Param("subject_id")
		if subjectId == "" {
			subjectId = c.Query("subject_id")
			if subjectId == "" {
				return app.ErrRequired.SetKey("subject_id")
			}
		}
		r.UpdatedBy = user.ID
		res, err := app.LessonFinalMake(&ses, subjectId, &r)
		if err != nil {
			return err
		}
		Success(c, gin.H{
			"exam_grade": res,
		})
		return nil
	})
	if err != nil {
		handleError(c, err)
		return
	}
}
