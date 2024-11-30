package api

import (
	"github.com/gin-gonic/gin"
	"github.com/mekdep/server/internal/api/utils"
	"github.com/mekdep/server/internal/app"
	"github.com/mekdep/server/internal/models"
)

func UserLogsRoutes(api *gin.RouterGroup) {
	userLogRoutes := api.Group("/users/logs")
	{
		userLogRoutes.GET("", UserLogsList)
	}
}

func UserLogsList(c *gin.Context) {
	ses := utils.InitSession(c)
	err := app.Ap().UserActionCheckRead(&ses, app.PermUser, func(user *models.User) (err error) {
		r := models.UserLogFilterRequest{}
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

		r.SchoolIds = new([]string)
		if ses.GetRole() != nil && *ses.GetRole() == models.RoleAdmin {
			if ses.GetSchoolId() != nil {
				*r.SchoolIds = []string{*ses.GetSchoolId()}
			} else {
				*r.SchoolIds = ses.GetSchoolsByAdminRoles()
			}
		} else if ses.GetRole() != nil && *ses.GetRole() == models.RoleOrganization {
			if ses.GetSchoolId() != nil {
				*r.SchoolIds = []string{*ses.GetSchoolId()}
			} else {
				*r.SchoolIds = ses.GetSchoolsByAdminRoles()
			}
		} else if ses.GetRole() != nil && *ses.GetRole() == models.RolePrincipal {
			*r.SchoolIds = ses.GetSchoolsByAdminRoles()
		} else {
			r.UserId = &user.ID
		}

		userLogs, total, err := app.UserLogsList(&ses, r)

		if err != nil {
			return err
		}

		Success(c, gin.H{
			"items": userLogs,
			"total": total,
		})
		return nil
	})
	if err != nil {
		handleError(c, err)
		return
	}
}
