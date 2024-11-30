package api

import (
	"github.com/gin-gonic/gin"
	"github.com/mekdep/server/internal/api/utils"
	"github.com/mekdep/server/internal/app"
	"github.com/mekdep/server/internal/app/app_validation"
	"github.com/mekdep/server/internal/models"
)

func ReportItemsRoutes(api *gin.RouterGroup) {
	reportItemsRoutes := api.Group("/report-forms/items")
	{
		reportItemsRoutes.POST(":id", ReportItemsCreate)
	}
}

func ReportItemsCreate(c *gin.Context) {
	ses := utils.InitSession(c)
	err := app.Ap().UserActionCheckWrite(&ses, app.PermToolReportForms, func(user *models.User) (err error) {
		r := models.ReportItemsRequest{}
		if err := BindAny(c, &r); err != nil {
			return err
		}
		// TODO: id
		id := c.Param("id")
		if id == "" {
			return app.ErrRequired.SetKey("id")
		}
		err = app_validation.ValidateReportItemsCreate(&ses, r)
		if err != nil {
			return err
		}
		r.ID = &id
		report_item, err := app.ReportItemsCreate(&ses, r)
		if err != nil {
			return err
		}
		userLog(models.UserLog{
			SchoolId:          ses.GetSchoolId(),
			SessionId:         ses.GetSessionId(),
			UserId:            user.ID,
			SubjectId:         &report_item.ID,
			Subject:           models.LogSubjectReportItems,
			SubjectAction:     models.LogActionCreate,
			SubjectProperties: r,
		})
		Success(c, gin.H{
			"report_item": report_item,
		})
		return nil
	})
	if err != nil {
		handleError(c, err)
		return
	}
}
