package api

import (
	"slices"

	"github.com/gin-gonic/gin"
	"github.com/mekdep/server/internal/api/utils"
	"github.com/mekdep/server/internal/app"
	"github.com/mekdep/server/internal/models"
)

func ReportFormRoutes(api *gin.RouterGroup) {
	cRoutes := api.Group("/reports")
	{
		cRoutes.GET("form", StatisticsSchoolInput)
		cRoutes.POST("form", StatisticsSchoolInputUpdate)
		cRoutes.GET("form/options", StatisticsSchoolInputOptions)
		cRoutes.GET("form/school", StatisticsSchoolFormGet)
		cRoutes.POST("form/school", StatisticsSchoolFormUpdate)
	}
}

func StatisticsSchoolFormGet(c *gin.Context) {
	ses := utils.InitSession(c)
	err := app.Ap().UserActionCheckWrite(&ses, app.PermToolReports, func(u *models.User) error {
		res, err := app.Ap().StatisticsSchoolForm(&ses, ses.SchoolAllIds())
		if err != nil {
			return err
		}
		Success(c, gin.H{
			"settings": res,
		})
		return nil
	})
	if err != nil {
		handleError(c, err)
		return
	}
}

func StatisticsSchoolFormUpdate(c *gin.Context) {
	ses := utils.InitSession(c)
	err := app.Ap().UserActionCheckWrite(&ses, app.PermToolReports, func(u *models.User) error {
		r := struct {
			Settings []app.StatisticsFormUpdateRequestDto `json:"settings"`
		}{}
		if errMsg, errKey := BindAndValidate(c, &r); errMsg != "" || errKey != "" {
			return app.NewAppError(errMsg, errKey, "")
		}

		for _, v := range r.Settings {
			if !slices.Contains(ses.SchoolAllIds(), v.SchoolId) {
				return app.ErrInvalid.SetKey("school_id")
			}
		}

		res, err := app.Ap().StatisticsSchoolFormUpdate(&ses, r.Settings)
		if err != nil {
			return err
		}
		Success(c, gin.H{
			"settings": res,
		})
		return nil
	})
	if err != nil {
		handleError(c, err)
		return
	}
}

func StatisticsSchoolInputUpdate(c *gin.Context) {
	ses := utils.InitSession(c)
	err := app.Ap().UserActionCheckWrite(&ses, app.PermToolReports, func(u *models.User) error {
		if ses.GetSchoolId() == nil {
			return app.ErrForbidden.SetComment("school_id is null")
		}
		r := models.SchoolSettingRequestForm{}
		if errMsg, errKey := BindAndValidate(c, &r); errMsg != "" || errKey != "" {
			return app.NewAppError(errMsg, errKey, "")
		}
		res, err := app.Ap().StatisticsSchoolInputUpdate(&ses, *ses.GetSchoolId(), r.Settings)
		if err != nil {
			return err
		}
		Success(c, gin.H{
			"settings": res,
		})
		return nil
	})
	if err != nil {
		handleError(c, err)
		return
	}
}

func StatisticsSchoolInput(c *gin.Context) {
	ses := utils.InitSession(c)
	err := app.Ap().UserActionCheckWrite(&ses, app.PermToolReports, func(u *models.User) error {
		if ses.GetSchoolId() == nil {
			return app.ErrForbidden.SetComment("school_id is null")
		}
		res, err := app.Ap().StatisticsSchoolInput(&ses, *ses.GetSchoolId())
		if err != nil {
			return err
		}
		Success(c, gin.H{
			"settings": res,
		})
		return nil
	})
	if err != nil {
		handleError(c, err)
		return
	}
}

func StatisticsSchoolInputOptions(c *gin.Context) {
	ses := utils.InitSession(c)
	err := app.Ap().UserActionCheckWrite(&ses, app.PermToolReports, func(u *models.User) error {
		if ses.GetSchoolId() == nil {
			return app.ErrForbidden.SetComment("school_id is null")
		}
		res, err := app.Ap().StatisticsSchoolInputOptions(&ses)
		if err != nil {
			return err
		}
		Success(c, gin.H{
			"options": res,
		})
		return nil
	})
	if err != nil {
		handleError(c, err)
		return
	}
}
