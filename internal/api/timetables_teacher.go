package api

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mekdep/server/internal/api/utils"
	"github.com/mekdep/server/internal/app"
	"github.com/mekdep/server/internal/models"
)

func TeacherRoutes(api *gin.RouterGroup) {
	rs := api.Group("teacher")
	{
		rs.GET("timetable", TeacherTimetable)
	}
}

func TeacherTimetable(c *gin.Context) {
	ses := utils.InitSession(c)
	err := app.Ap().UserActionCheckRead(&ses, app.PermTeacher, func(user *models.User) (err error) {
		dateStr := c.Query("date")
		date := time.Now()
		if dateStr != "" {
			date, err = ParseDate(dateStr)
			if err != nil {
				err = nil
				date = time.Now()
			}
		}
		res, err := app.Ap().TeacherTimetable(&ses, date)
		if err != nil {
			return err
		}
		Success(c, gin.H{
			"week": res,
		})
		return
	})
	if err != nil {
		handleError(c, err)
		return
	}
}
