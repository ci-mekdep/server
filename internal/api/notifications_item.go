package api

import (
	"slices"

	"github.com/gin-gonic/gin"
	"github.com/mekdep/server/internal/api/utils"
	"github.com/mekdep/server/internal/app"
	"github.com/mekdep/server/internal/app/app_validation"
	"github.com/mekdep/server/internal/models"
)

func UserNotificationsRoutes(api *gin.RouterGroup) {
	userNotificationRoutes := api.Group("/users/notifications")
	{
		userNotificationRoutes.GET("", UserNotificationsList)
		userNotificationRoutes.PUT(":id", UserNotificationsUpdate)
	}
}

func UserNotificationsList(c *gin.Context) {
	ses := utils.InitSession(c)
	err := app.Ap().UserActionCheckRead(&ses, app.PermUser, func(user *models.User) error {
		r := models.UserNotificationFilterRequest{}
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
		r.UserId = &user.ID
		r.Role = (*string)(ses.GetRole())

		userNotifications, total, totalUnread, err := app.UserNotificationsList(&ses, r)

		if err != nil {
			return err
		}

		Success(c, gin.H{
			"items":        userNotifications,
			"total":        total,
			"total_unread": totalUnread,
		})
		return nil
	})
	if err != nil {
		handleError(c, err)
		return
	}
}

func UserNotificationsUpdate(c *gin.Context) {
	ses := utils.InitSession(c)
	err := app.Ap().UserActionCheckWrite(&ses, app.PermUser, func(user *models.User) (err error) {
		r := models.UserNotificationRequest{}
		if errMsg, errKey := BindAndValidate(c, &r); errMsg != "" || errKey != "" {
			return app.NewAppError(errMsg, errKey, "")
		}
		id := c.Param("id")
		r.ID = &id

		if id == "" {
			return app.ErrRequired.SetKey("id")
		}
		err = app_validation.ValidateNotificationsItemCreate(&ses, r)
		if err != nil {
			return err
		}

		// TODO refactor: new function for this??
		var model models.UserNotification
		if model.CommentFiles == nil {
			model.CommentFiles = new([]string)
		}

		paths := *model.CommentFiles
		if r.CommentFilesDelete != nil {
			if len(*r.CommentFilesDelete) < 1 || (*r.CommentFilesDelete)[0] == "" {
				*r.CommentFilesDelete = paths
			}
			for _, v := range *r.CommentFilesDelete {
				deleteFile(c, v, "notifications")
				k := slices.Index(paths, v)
				if k >= 0 {
					paths = slices.Delete(paths, k, k+1)
				}
			}
		}

		commentFiles, err := handleFilesUpload(c, "comment_files", "notifications")
		if err != nil {
			return app.NewAppError(err.Error(), "comment_files", "")
		}
		paths = append(paths, commentFiles...)

		r.CommentFiles = &paths

		if v, ok := c.Request.Form["comment_files_delete"]; ok {
			r.CommentFilesDelete = &v
		}

		userNotification, err := app.UserNotificationUpdate(&ses, r)
		if err != nil {
			return err
		}

		Success(c, gin.H{
			"item": userNotification,
		})
		return nil
	})
	if err != nil {
		handleError(c, err)
		return
	}
}
