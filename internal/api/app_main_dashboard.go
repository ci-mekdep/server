package api

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mekdep/server/internal/api/utils"
	"github.com/mekdep/server/internal/app"
)

func Dashboards(c *gin.Context) {
	ses := utils.InitSession(c)
	dateStr := c.QueryArray("date[]")
	startDate := time.Now()
	endDate := time.Now()
	var err error
	if len(dateStr) > 0 {
		if len(dateStr) > 1 {
			startDate, err = ParseDateUTC(dateStr[0])
			endDate, err = ParseDateUTC(dateStr[1])
			if err != nil {
				startDate = time.Now()
				endDate = time.Now()
			}
		} else {
			startDate, err = ParseDateUTC(dateStr[0])
			endDate, err = ParseDateUTC(dateStr[0])
			if err != nil {
				startDate = time.Now()
				endDate = time.Now()
			}
		}
	}
	d, err := app.Ap().Dashboards(&ses, startDate, endDate)
	if err != nil {
		handleError(c, err)
		return
	}
	Success(c, gin.H{
		"dashboards": d,
	})
}

func DashboardDetails(c *gin.Context) {
	// TODO refactor: everything is same with Dashboards???
	ses := utils.InitSession(c)
	dateStr := c.QueryArray("date[]")
	detailIdStr := c.Query("detail_id")
	var detailId *string

	if detailIdStr != "" {
		tmp := detailIdStr
		if tmp != "" {
			detailId = new(string)
			*detailId = tmp
		}
	}
	startDate := time.Now()
	endDate := time.Now()
	var err error
	if len(dateStr) > 0 {
		if len(dateStr) > 1 {
			startDate, err = ParseDateUTC(dateStr[0])
			endDate, err = ParseDateUTC(dateStr[1])
			if err != nil {
				startDate = time.Now()
				endDate = time.Now()
			}
		} else {
			startDate, err = ParseDateUTC(dateStr[0])
			endDate, err = ParseDateUTC(dateStr[0])
			if err != nil {
				startDate = time.Now()
				endDate = time.Now()
			}
		}
	}

	d, err := app.Ap().DashboardDetails(&ses, startDate, endDate, detailId)
	if err != nil {
		handleError(c, err)
		return
	}
	Success(c, gin.H{
		"dashboards": d,
	})
}
