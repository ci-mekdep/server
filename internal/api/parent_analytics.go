package api

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mekdep/server/internal/api/utils"
	"github.com/mekdep/server/internal/app"
	"github.com/mekdep/server/internal/models"
)

func AnalyticsRoutes(api *gin.RouterGroup) {
	rs := api.Group("/parent")
	{
		rs.GET("/analytics/weekly", ParentAnalyticsWeekly)
		rs.GET("/analytics/summary", ParentAnalyticsSummary)
		rs.GET("/analytics/grades", ParentAnalyticsGrades)
		rs.GET("/analytics/exams", ParentAnalyticsExams)
	}
}

func ParentAnalyticsWeekly(c *gin.Context) {
	ses := utils.InitSession(c)
	p := app.PermPlus
	err := app.Ap().UserActionCheckRead(&ses, p, func(user *models.User) error {
		date, err := getDate(c, false)
		if err != nil {
			return err
		}

		student, err := getParentChild(c, &ses, user)
		if err != nil {
			return err
		}

		classroomId := c.Request.FormValue("classroom_id")
		if classroomId == "" {
			return app.ErrRequired.SetKey("classroom_id")
		}

		uu, err := app.Ap().ParentAnalyticsWeekly(&ses, student, classroomId, date, p)
		if err != nil {
			return err
		}

		Success(c, gin.H{
			"weekly": uu,
		})
		return nil
	})
	if err != nil {
		handleError(c, err)
		return
	}
}

func ParentAnalyticsSummary(c *gin.Context) {
	ses := utils.InitSession(c)
	p := app.PermPlus
	err := app.Ap().UserActionCheckRead(&ses, p, func(user *models.User) error {
		date, err := getDate(c, false)
		if err != nil {
			return err
		}
		student, err := getParentChild(c, &ses, user)
		if err != nil {
			return err
		}
		classroomId := c.Request.FormValue("classroom_id")
		if classroomId == "" {
			return app.ErrRequired.SetKey("classroom_id")
		}

		res, err := app.Ap().ParentAnalyticsSummary(&ses, student, classroomId, date, p)
		if err != nil {
			return err
		}
		Success(c, gin.H{"summary": res})
		return nil
	})
	if err != nil {
		handleError(c, err)
		return
	}
}

func ParentAnalyticsGrades(c *gin.Context) {
	ses := utils.InitSession(c)
	p := app.PermPlus
	err := app.Ap().UserActionCheckRead(&ses, p, func(user *models.User) error {
		subjectId := c.Request.FormValue("subject_id")
		if subjectId == "" {
			return app.ErrRequired.SetKey("subject_id")
		}

		classroomId := c.Request.FormValue("classroom_id")
		if classroomId == "" {
			return app.ErrRequired.SetKey("classroom_id")
		}
		// TODO: period number boyuncha period_id (modelini) tapmaly we bashlangyc ahyrky senesine gora filterlap lessonlary almaly
		var periodNumber *int
		if pn := c.Query("period_number"); pn != "" {
			pnInt, err := strconv.Atoi(pn)
			if err != nil {
				return err
			}
			periodNumber = &pnInt
		}

		var limit *int
		if lim := c.Query("limit"); lim != "" {
			limInt, err := strconv.Atoi(lim)
			if err != nil {
				return err
			}
			limit = &limInt
		}

		var endDate *time.Time
		if endStr := c.Query("end_date"); endStr != "" {
			parsedDate, err := ParseDate(endStr)
			if err == nil {
				endDate = &parsedDate
			}
		}

		showOnlyAbsent := c.Query("show_only_absent") == "true"
		showOnlyGrade := c.Query("show_only_grade") == "true"

		student, err := getParentChild(c, &ses, user)
		if err != nil {
			return err
		}

		grades, err := app.Ap().ParentAnalyticsGrades(&ses, student, subjectId, classroomId, periodNumber, limit, endDate, showOnlyAbsent, showOnlyGrade, p)
		if err != nil {
			return err
		}
		if grades == nil {
			grades = []models.GradeDetail{}
		}
		Success(c, gin.H{"grades": grades})
		return nil
	})
	if err != nil {
		handleError(c, err)
		return
	}
}

func ParentAnalyticsExams(c *gin.Context) {
	ses := utils.InitSession(c)
	p := app.PermPlus
	err := app.Ap().UserActionCheckRead(&ses, p, func(user *models.User) error {
		subjectId := c.Request.FormValue("subject_id")
		if subjectId == "" {
			return app.ErrRequired.SetKey("subject_id")
		}
		classroomId := c.Request.FormValue("classroom_id")
		if classroomId == "" {
			return app.ErrRequired.SetKey("classroom_id")
		}

		student, err := getParentChild(c, &ses, user)
		if err != nil {
			return err
		}

		grades, err := app.Ap().ParentAnalyticsExams(&ses, student, subjectId, classroomId, p)
		if err != nil {
			return err
		}
		Success(c, gin.H{"rating_center_school": grades})
		return nil
	})
	if err != nil {
		handleError(c, err)
		return
	}
}
