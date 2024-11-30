package api

import (
	"github.com/gin-gonic/gin"
	"github.com/mekdep/server/internal/api/utils"
	"github.com/mekdep/server/internal/app"
	"github.com/mekdep/server/internal/app/app_validation"
	"github.com/mekdep/server/internal/models"
)

func ReportsRoutes(api *gin.RouterGroup) {
	reportsRoutes := api.Group("/report-forms")
	{
		reportsRoutes.GET("", ReportsList)
		reportsRoutes.GET("total", ReportsListTotal)
		reportsRoutes.POST("", ReportsCreate)
		reportsRoutes.PUT(":id", ReportsUpdate)
		reportsRoutes.DELETE("", ReportsDelete)
		reportsRoutes.GET(":id", ReportsDetail)
		reportsRoutes.GET("rating/:id/center", ReportsRatingCenter)
	}
}

func reportsListQuery(ses *utils.Session, req models.ReportsFilterRequest) ([]*models.ReportsResponse, int, int, error) {
	req.SchoolIds = &[]string{}
	if *ses.GetRole() != models.RoleAdmin {
		*req.SchoolIds = ses.GetSchoolIds()
	} else {
		req.SchoolIds = nil
	}
	// TODO: organization -> user.schools[]=organization -> user.schools.school_id
	if *ses.GetRole() == models.RoleOrganization {
		for _, schoolId := range ses.GetUser().Schools {
			if schoolId.RoleCode == models.RoleOrganization && schoolId.SchoolUid != nil {
				req.SchoolIds = &[]string{*schoolId.SchoolUid}
			}
		}
	}
	// TODO role teacher -> clasrooms.teacher_id=user_id -> classroom_ids
	if *ses.GetRole() == models.RoleTeacher {
		if ses.GetUser().TeacherClassroom != nil {
			req.ClassroomUids = &[]string{ses.GetUser().TeacherClassroom.ID}
		}
	}

	res, total, totalUnfilled, err := app.ReportsList(ses, req)
	return res, total, totalUnfilled, err
}

func reportsAvailableCheck(ses *utils.Session, data models.ReportsFilterRequest) (bool, error) {
	_, t, _, err := reportsListQuery(ses, data)
	if err != nil {
		return false, app.ErrRequired.SetKey("id")
	}
	return t > 0, nil
}

func ReportsList(c *gin.Context) {
	ses := utils.InitSession(c)
	err := app.Ap().UserActionCheckRead(&ses, app.PermUser, func(user *models.User) error {
		r := models.ReportsFilterRequest{}
		if errMsg, errKey := BindAndValidate(c, &r); errMsg != "" || errKey != "" {
			return app.NewAppError(errMsg, errKey, "")
		}

		reports, total, totalUnfilled, err := reportsListQuery(&ses, r)

		if err != nil {
			return err
		}

		Success(c, gin.H{
			"reports":        reports,
			"total":          total,
			"total_unfilled": totalUnfilled,
		})
		return nil
	})
	if err != nil {
		handleError(c, err)
		return
	}
}

func ReportsListTotal(c *gin.Context) {
	ses := utils.InitSession(c)
	err := app.Ap().UserActionCheckRead(&ses, app.PermUser, func(user *models.User) error {
		r := models.ReportsFilterRequest{}
		if errMsg, errKey := BindAndValidate(c, &r); errMsg != "" || errKey != "" {
			return app.NewAppError(errMsg, errKey, "")
		}

		_, total, totalUnfilled, err := reportsListQuery(&ses, r)

		if err != nil {
			return err
		}

		Success(c, gin.H{
			"reports":        nil,
			"total":          total,
			"total_unfilled": totalUnfilled,
		})
		return nil
	})
	if err != nil {
		handleError(c, err)
		return
	}
}

func ReportsDetail(c *gin.Context) {
	ses := utils.InitSession(c)
	err := app.Ap().UserActionCheckRead(&ses, app.PermToolReportForms, func(u *models.User) error {
		id := c.Param("id")
		if ok, err := reportsAvailableCheck(&ses, models.ReportsFilterRequest{ID: &id}); err != nil {
			return err
		} else if !ok {
			return app.ErrNotfound
		}
		m, err := app.ReportsDetail(&ses, id)
		if err != nil {
			return err
		}
		Success(c, gin.H{
			"report": m,
		})
		return nil
	})
	if err != nil {
		handleError(c, err)
		return
	}
}

func ReportsCreate(c *gin.Context) {
	ses := utils.InitSession(c)
	err := app.Ap().UserActionCheckWrite(&ses, app.PermToolReportForms, func(user *models.User) (err error) {
		r := models.ReportsRequest{}
		if err := BindAny(c, &r); err != nil {
			return err
		}
		err = app_validation.ValidateReportsCreate(&ses, r)
		if err != nil {
			return err
		}
		report, err := app.ReportsCreate(&ses, r)
		if err != nil {
			return err
		}
		userLog(models.UserLog{
			SchoolId:          ses.GetSchoolId(),
			SessionId:         ses.GetSessionId(),
			UserId:            user.ID,
			SubjectId:         &report.ID,
			Subject:           models.LogSubjectReports,
			SubjectAction:     models.LogActionCreate,
			SubjectProperties: r,
		})
		Success(c, gin.H{
			"report": report,
		})
		return nil
	})
	if err != nil {
		handleError(c, err)
		return
	}
}

func ReportsUpdate(c *gin.Context) {
	ses := utils.InitSession(c)
	err := app.Ap().UserActionCheckWrite(&ses, app.PermToolReportForms, func(user *models.User) (err error) {
		r := models.ReportsRequest{}
		if errMsg, errKey := BindAndValidate(c, &r); errMsg != "" || errKey != "" {
			return app.NewAppError(errMsg, errKey, "")
		}
		err = app_validation.ValidateReportsCreate(&ses, r)
		if err != nil {
			return err
		}
		report, err := app.ReportsUpdate(&ses, r)
		if err != nil {
			return err
		}
		userLog(models.UserLog{
			SchoolId:          ses.GetSchoolId(),
			SessionId:         ses.GetSessionId(),
			UserId:            user.ID,
			SubjectId:         &report.ID,
			Subject:           models.LogSubjectReports,
			SubjectAction:     models.LogActionUpdate,
			SubjectProperties: r,
		})
		Success(c, gin.H{
			"report": report,
		})
		return nil
	})
	if err != nil {
		handleError(c, err)
		return
	}
}

func ReportsRatingCenter(c *gin.Context) {
	ses := utils.InitSession(c)
	err := app.Ap().UserActionCheckRead(&ses, app.PermToolReportForms, func(u *models.User) error {
		id := c.Param("id")
		if ok, err := reportsAvailableCheck(&ses, models.ReportsFilterRequest{ID: &id}); err != nil {
			return err
		} else if !ok {
			return app.ErrNotfound
		}
		m, err := app.ReportsRatingCenter(&ses, id)
		if err != nil {
			return err
		}
		Success(c, gin.H{
			"report_rating": m,
		})
		return nil
	})
	if err != nil {
		handleError(c, err)
		return
	}
}

func ReportsDelete(c *gin.Context) {
	ses := utils.InitSession(c)
	err := app.Ap().UserActionCheckWrite(&ses, app.PermToolReportForms, func(user *models.User) error {
		var ids []string = c.QueryArray("ids")
		if len(ids) == 0 {
			return app.ErrRequired.SetKey("ids")
		}

		if ok, err := reportsAvailableCheck(&ses, models.ReportsFilterRequest{IDs: &ids}); err != nil {
			return err
		} else if !ok {
			return app.ErrForbidden
		}

		reports, err := app.ReportsDelete(&ses, ids)
		if err != nil {
			return err
		}
		userLog(models.UserLog{
			SchoolId:          ses.GetSchoolId(),
			SessionId:         ses.GetSessionId(),
			UserId:            user.ID,
			Subject:           models.LogSubjectReports,
			SubjectAction:     models.LogActionDelete,
			SubjectProperties: ids,
		})
		Success(c, gin.H{
			"reports": reports,
		})
		return nil
	})
	if err != nil {
		handleError(c, err)
		return
	}
}
