package api

import (
	"os"

	"github.com/gin-gonic/gin"
	"github.com/mekdep/server/internal/app"
)

func ToolRoutes(api *gin.RouterGroup) {
	r := api.Group("/tools")
	{
		r.GET("reset", ToolResetPassword)
	}
}

func ToolResetPassword(c *gin.Context) {
	// ses := utils.InitSession(c)
	// err := app.Ap().UserActionCheckWrite(&ses, app.PermToolReset, func(user *models.User) error {
	r := app.ResetPasswordRequest{}
	if errMsg, errKey := BindAndValidate(c, &r); errMsg != "" || errKey != "" {
		handleError(c, app.NewAppError(errMsg, errKey, ""))
		return
	}

	path, err := app.ToolResetPassword()
	if err != nil {
		handleError(c, err)
		return
	}

	f, err := os.ReadFile(path)
	c.Writer.Header().Set("Content-Disposition", "attachment; filename=report.xlsx")
	c.Writer.Header().Set("Content-Type", "application/octet-stream")
	c.Writer.Write(f)
	// })
	if err != nil {
		handleError(c, err)
		return
	}

}
