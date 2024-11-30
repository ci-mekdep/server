package api

import (
	"github.com/gin-gonic/gin"
	"github.com/mekdep/server/internal/api/utils"
	"github.com/mekdep/server/internal/app"
	"github.com/mekdep/server/internal/models"
)

func LessonLikes(c *gin.Context) {
	ses := utils.InitSession(c)
	err := app.Ap().UserActionCheckWrite(&ses, app.PermUser, func(user *models.User) error {
		req := models.LessonLikesRequest{}
		if errMsg, errKey := BindAndValidate(c, &req); errMsg != "" || errKey != "" {
			return app.NewAppError(errMsg, errKey, "")
		}
		lessonIdStr := c.Param("id")
		req.UserId = &ses.GetUser().ID
		req.LessonId = &lessonIdStr
		lessonLike, err := app.LessonLikes(&ses, &req)
		if err != nil {
			return err
		}
		Success(c, gin.H{
			"lessonlike": lessonLike,
		})
		return nil
	})
	if err != nil {
		handleError(c, err)
		return
	}
}
