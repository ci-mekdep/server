package api

import (
	"slices"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mekdep/server/internal/api/utils"
	"github.com/mekdep/server/internal/app"
	"github.com/mekdep/server/internal/models"
)

func ReportRoutes(api *gin.RouterGroup) {
	cRoutes := api.Group("/reports")
	{
		cRoutes.GET("data", StatisticsSchoolData)
		cRoutes.GET("online", StatisticsOnline)
		cRoutes.GET("students", StatisticsStudents)
		cRoutes.GET("exams", StatisticsSchoolExams)
		cRoutes.GET("period_finished", StatisticsPeriodFinished)
		cRoutes.GET("attendance", StatisticsAttendance)
		cRoutes.GET("parents", StatisticsParents)
		cRoutes.GET("journal", StatisticsJournal)
		cRoutes.GET("payments", StatisticsPayments)
		cRoutes.GET("contact-items", StatisticsContactItems)
	}
}

func StatisticsSchoolData(c *gin.Context) {
	ses := utils.InitSession(c)
	err := app.Ap().UserActionCheckRead(&ses, app.PermToolReports, func(u *models.User) error {
		startDate, endDate := getDateBetween(c)

		rep, err := app.Ap().StatisticsSchoolData(&ses, ses.GetSchoolsByAdminRoles(), startDate, endDate)
		if err != nil {
			return err
		}
		Success(c, gin.H{
			"report": rep,
		})
		return nil
	})
	if err != nil {
		handleError(c, err)
		return
	}
}

func StatisticsSchoolExams(c *gin.Context) {
	ses := utils.InitSession(c)
	err := app.Ap().UserActionCheckRead(&ses, app.PermToolReports, func(u *models.User) error {
		now := time.Now()
		now = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		r := app.ReportsRequestDto{
			SchoolId: ses.GetSchoolId(),
		}
		r.StartDate = new(time.Time)
		*r.StartDate = now
		r.EndDate = new(time.Time)
		*r.EndDate = now
		if errMsg, errKey := BindAndValidate(c, &r); errMsg != "" || errKey != "" {
			return app.NewAppError(errMsg, errKey, "")
		}
		r.SchoolIds = new([]string)
		*r.SchoolIds = ses.SchoolAllIds()
		isGraduate, _ := strconv.Atoi(c.Query("is_graduate"))
		isGraduateBool := isGraduate != 0
		rep, err := app.Ap().StatisticsExam(&ses, r, &isGraduateBool)
		if err != nil {
			return err
		}
		Success(c, gin.H{
			"report": rep,
		})
		return nil
	})
	if err != nil {
		handleError(c, err)
		return
	}
}

func StatisticsOnline(c *gin.Context) {
	ses := utils.InitSession(c)
	err := app.Ap().UserActionCheckRead(&ses, app.PermToolReports, func(u *models.User) error {
		startDate, endDate := getDateBetween(c)

		rep, err := app.Ap().StatisticsOnline(&ses, ses.SchoolAllIds(), startDate, endDate)
		if err != nil {
			return err
		}
		Success(c, gin.H{
			"report": rep,
		})
		return nil
	})
	if err != nil {
		handleError(c, err)
		return
	}
}

func StatisticsPayments(c *gin.Context) {
	ses := utils.InitSession(c)
	err := app.Ap().UserActionCheckRead(&ses, app.PermAdminPayments, func(u *models.User) error {
		r := models.PaymentTransactionFilterRequest{}
		if errMsg, errKey := BindAndValidate(c, &r); errMsg != "" {
			return app.NewAppError(errMsg, errKey, "")
		}
		r.SchoolIds = new([]string)
		*r.SchoolIds = ses.GetSchoolsByAdminRoles()

		response, err := app.StatisticsPayments(&ses, r)
		if err != nil {
			return err
		}
		Success(c, gin.H{
			"report_payments": response,
		})
		return nil
	})
	if err != nil {
		handleError(c, err)
		return
	}
}

func StatisticsContactItems(c *gin.Context) {
	ses := utils.InitSession(c)
	err := app.Ap().UserActionCheckRead(&ses, app.PermToolReports, func(u *models.User) error {
		r := models.ContactItemsFilterRequest{}
		if errMsg, errKey := BindAndValidate(c, &r); errMsg != "" {
			return app.NewAppError(errMsg, errKey, "")
		}
		r.SchoolIds = new([]string)
		*r.SchoolIds = ses.GetSchoolsByAdminRoles()

		response, err := app.StatisticsContactItems(&ses, r)
		if err != nil {
			return err
		}
		Success(c, gin.H{
			"report_contact_items": response,
		})
		return nil
	})
	if err != nil {
		handleError(c, err)
		return
	}
}

func StatisticsStudents(c *gin.Context) {
	ses := utils.InitSession(c)
	err := app.Ap().UserActionCheckRead(&ses, app.PermToolReports, func(u *models.User) error {
		rep, err := app.Ap().StatisticsStudents(&ses, ses.SchoolAllIds())
		if err != nil {
			return err
		}
		Success(c, gin.H{
			"report": rep,
		})
		return nil
	})
	if err != nil {
		handleError(c, err)
		return
	}
}

func StatisticsPeriodFinished(c *gin.Context) {
	ses := utils.InitSession(c)
	err := app.Ap().UserActionCheckRead(&ses, app.PermToolReports, func(u *models.User) error {
		periodNumber := 1
		schoolId := ""
		if str := c.Query("period_number"); str != "" {
			periodNumber, _ = strconv.Atoi(str)
			if periodNumber > 4 || periodNumber < 1 {
				periodNumber = 1
			}
		}
		if str := c.Query("school_id"); str != "" {
			schoolIdNum := str
			schoolId = schoolIdNum
		}
		if len(ses.GetSchoolIds()) == 1 {
			schoolId = ses.GetSchoolIds()[0]
		}

		var rep app.StatisticsResponse
		var isDetail bool
		var err error
		if schoolId != "" {
			rep, err = app.Ap().StatisticsPeriodFinishedByTeacher(&ses, schoolId, periodNumber)
			isDetail = false
			if err != nil {
				return err
			}
		} else {
			rep, err = app.Ap().StatisticsPeriodFinished(&ses, ses.SchoolAllIds(), periodNumber)
			isDetail = true
			if err != nil {
				return err
			}
		}
		rep.HasDetail = isDetail
		Success(c, gin.H{
			"report": rep,
		})
		return nil
	})
	if err != nil {
		handleError(c, err)
		return
	}
}

func StatisticsJournal(c *gin.Context) {
	ses := utils.InitSession(c)
	err := app.Ap().UserActionCheckRead(&ses, app.PermToolReports, func(u *models.User) error {
		now := time.Now()
		now = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		r := app.ReportsRequestDto{
			SchoolId: ses.GetSchoolId(),
		}
		r.StartDate = new(time.Time)
		*r.StartDate = now
		r.EndDate = new(time.Time)
		*r.EndDate = now
		if errMsg, errKey := BindAndValidate(c, &r); errMsg != "" || errKey != "" {
			return app.NewAppError(errMsg, errKey, "")
		}
		r.SchoolIds = new([]string)
		*r.SchoolIds = ses.SchoolAllIds()
		var rep *app.StatisticsResponse
		var err error
		if *ses.GetRole() == models.RolePrincipal || r.SchoolId != nil {
			if !slices.Contains(*r.SchoolIds, *r.SchoolId) {
				return app.ErrForbidden.SetKey("school_id")
			}
			if *ses.GetRole() == models.RoleTeacher || r.UserId != nil {
				rep, err = app.Ap().StatisticsJournalByLesson(&ses, r)
				if err != nil {
					return err
				}
			} else {
				rep, err = app.Ap().StatisticsJournalByTeacher(&ses, r)
				if err != nil {
					return err
				}
			}
		} else {
			rep, err = app.Ap().StatisticsJournal(&ses, r)
			if err != nil {
				return err
			}
		}
		Success(c, gin.H{
			"report": rep,
		})
		return nil
	})
	if err != nil {
		handleError(c, err)
		return
	}
}

func StatisticsAttendance(c *gin.Context) {
	ses := utils.InitSession(c)
	err := app.Ap().UserActionCheckRead(&ses, app.PermToolReports, func(u *models.User) error {
		now := time.Now()
		now = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		r := app.ReportsRequestDto{
			SchoolId: ses.GetSchoolId(),
		}
		r.StartDate = new(time.Time)
		*r.StartDate = now
		r.EndDate = new(time.Time)
		*r.EndDate = now
		if errMsg, errKey := BindAndValidate(c, &r); errMsg != "" || errKey != "" {
			return app.NewAppError(errMsg, errKey, "")
		}
		r.SchoolIds = new([]string)
		*r.SchoolIds = ses.SchoolAllIds()
		var rep app.StatisticsResponse
		var err error
		if *ses.GetRole() == models.RolePrincipal || r.SchoolId != nil {
			if !slices.Contains(*r.SchoolIds, *r.SchoolId) {
				return app.ErrForbidden.SetKey("school_id")
			}
			rep, err = app.Ap().StatisticsAttendanceByClassroom(&ses, r)
			if err != nil {
				return err
			}
		} else {
			rep, err = app.Ap().StatisticsAttendance(&ses, r)
			if err != nil {
				return err
			}
		}
		Success(c, gin.H{
			"report": rep,
		})
		return nil
	})
	if err != nil {
		handleError(c, err)
		return
	}
}

func StatisticsParents(c *gin.Context) {
	ses := utils.InitSession(c)
	err := app.Ap().UserActionCheckRead(&ses, app.PermToolReports, func(u *models.User) error {

		schoolId := ""
		if str := c.Query("school_id"); str != "" {
			schoolId = str
		}
		if len(ses.GetSchoolIds()) == 1 {
			schoolId = ses.GetSchoolIds()[0]
		}

		var rep app.StatisticsResponse
		var isDetail bool
		var err error
		if schoolId != "" {
			rep, err = app.Ap().StatisticsParentsBySchool(&ses, schoolId)
			isDetail = false
			if err != nil {
				return err
			}
		} else {
			rep, err = app.Ap().StatisticsParents(&ses, ses.SchoolAllIds())
			isDetail = true
			if err != nil {
				return err
			}
		}
		rep.HasDetail = isDetail
		Success(c, gin.H{
			"report": rep,
		})
		return nil
	})
	if err != nil {
		handleError(c, err)
		return
	}
}

func getDateBetween(c *gin.Context) (time.Time, time.Time) {
	startDate := time.Now()
	endDate := time.Now()
	if startStr := c.Query("start_date"); startStr != "" {
		item, err := ParseDate(startStr)
		if err == nil {
			startDate = item
		}
	} else if len(c.QueryArray("date[]")) > 0 {
		if startStr := c.QueryArray("date[]")[0]; startStr != "" {
			item, err := ParseDate(startStr)
			if err == nil {
				startDate = item
			}
		}
	}
	if endStr := c.Query("end_date"); endStr != "" {
		item, err := ParseDate(endStr)
		if err == nil {
			endDate = item
		}
	} else if len(c.QueryArray("date[]")) > 1 {
		if endStr := c.QueryArray("date[]")[1]; endStr != "" {
			item, err := ParseDate(endStr)
			if err == nil {
				endDate = item
			}
		}
	}
	return startDate, endDate

}
