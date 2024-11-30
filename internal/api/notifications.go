package api

import (
	"slices"

	"github.com/gin-gonic/gin"
	"github.com/mekdep/server/internal/api/utils"
	"github.com/mekdep/server/internal/app"
	"github.com/mekdep/server/internal/app/app_validation"
	"github.com/mekdep/server/internal/models"
)

func NotificationsRoutes(api *gin.RouterGroup) {
	notificationRoutes := api.Group("/users/notifications/sender")
	{
		notificationRoutes.GET("", NotificationsList)
		notificationRoutes.GET(":id", NotificationsDetail)
		notificationRoutes.PUT(":id", NotificationsUpdate)
		notificationRoutes.POST("", NotificationsCreate)
		notificationRoutes.DELETE(":id", NotificationDelete)
	}
}

// TODO: add notificationsListQuery
func notificationsListQuery(ses *utils.Session, data models.NotificationsFilterRequest) ([]*models.NotificationsResponse, int, error) {
	data.AuthorId = &ses.GetUser().ID
	return app.NotificationList(ses, data)
}
func notificationsAvailableCheck(ses *utils.Session, data models.NotificationsFilterRequest) (bool, error) {
	_, t, err := notificationsListQuery(ses, data)
	if err != nil {
		return false, app.ErrRequired.SetKey("id")
	}
	return t > 0, nil
}

func NotificationsList(c *gin.Context) {
	ses := utils.InitSession(c)
	err := app.Ap().UserActionCheckRead(&ses, app.PermToolNotifier, func(user *models.User) error {
		r := models.NotificationsFilterRequest{}
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

		notifications, total, err := notificationsListQuery(&ses, r)

		if err != nil {
			return err
		}

		Success(c, gin.H{
			"items": notifications,
			"total": total,
		})
		return nil
	})
	if err != nil {
		handleError(c, err)
		return
	}
}

func NotificationsDetail(c *gin.Context) {
	ses := utils.InitSession(c)
	err := app.Ap().UserActionCheckRead(&ses, app.PermToolNotifier, func(user *models.User) error {
		id := c.Param("id")
		if id == "" {
			return app.ErrRequired.SetKey("id")
		}
		if ok, err := notificationsAvailableCheck(&ses, models.NotificationsFilterRequest{ID: &id}); err != nil {
			return err
		} else if !ok {
			return app.ErrNotfound
		}

		r := models.UserNotificationFilterRequest{NotificationId: &id}
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

		n, total, err := app.NotificationDetail(r)
		if err != nil {
			return err
		}

		Success(c, gin.H{
			"notification": n,
			"total":        total,
		})
		return nil
	})
	if err != nil {
		handleError(c, err)
		return
	}
}

func NotificationsUpdate(c *gin.Context) {
	ses := utils.InitSession(c)
	err := app.Ap().UserActionCheckWrite(&ses, app.PermToolNotifier, func(user *models.User) (err error) {
		r := models.NotificationsRequest{}
		if errMsg, errKey := BindAndValidate(c, &r); errMsg != "" || errKey != "" {
			return app.NewAppError(errMsg, errKey, "")
		}
		id := c.Param("id")
		r.ID = &id

		if id == "" {
			return app.ErrRequired.SetKey("id")
		}
		err = app_validation.ValidateNotificationsCreate(&ses, r)
		if err != nil {
			return err
		}
		var model models.Notifications
		if model.Files == nil {
			model.Files = new([]string)
		}

		paths := *model.Files
		if r.Files != nil {
			if len(*r.FilesDelete) < 1 || (*r.FilesDelete)[0] == "" {
				*r.FilesDelete = paths
			}
			for _, v := range *r.FilesDelete {
				deleteFile(c, v, "notifications")
				k := slices.Index(paths, v)
				if k >= 0 {
					paths = slices.Delete(paths, k, k+1)
				}
			}
		}

		files, err := handleFilesUpload(c, "files", "notifications")
		if err != nil {
			return app.NewAppError(err.Error(), "files", "")
		}
		paths = append(paths, files...)
		r.Files = &paths

		if v, ok := c.Request.Form["files_delete"]; ok {
			r.FilesDelete = &v
		}

		notification, err := app.NotificationUpdate(&ses, r)
		if err != nil {
			return err
		}

		Success(c, gin.H{
			"item": notification,
		})
		return nil
	})
	if err != nil {
		handleError(c, err)
		return
	}
}

func NotificationsCreate(c *gin.Context) {
	ses := utils.InitSession(c)
	err := app.Ap().UserActionCheckWrite(&ses, app.PermToolNotifier, func(user *models.User) (err error) {
		r := models.NotificationsRequest{}
		if err := BindAny(c, &r); err != nil {
			return err
		}
		err = app_validation.ValidateNotificationsCreate(&ses, r)
		if err != nil {
			return err
		}
		for _, v := range r.Roles {
			if !app_validation.CheckAvailableRole(ses.GetUser(), ses.GetRole(), models.Role(v)) {
				if !(*ses.GetRole() == models.RoleAdmin && models.Role(v) == models.RoleAdmin) {
					return app.ErrForbidden.SetKey("roles").SetComment(v + " not available for you")
				}
			}
		}

		files, err := handleFilesUpload(c, "files", "notifications")
		if err != nil {
			return app.NewAppError(err.Error(), "files", "")
		}

		if files != nil {
			r.Files = &files
		}
		notification, err := app.NotificationCreate(&ses, r)
		if err != nil {
			return err
		}
		Success(c, gin.H{
			"notification": notification,
		})
		return nil
	})
	if err != nil {
		handleError(c, err)
		return
	}
}

func NotificationDelete(c *gin.Context) {
	ses := utils.InitSession(c)
	err := app.Ap().UserActionCheckWrite(&ses, app.PermToolNotifier, func(user *models.User) error {
		id := c.Param("id")

		err := app.NotificationDelete(&ses, id)
		if err != nil {
			return err
		}

		Success(c, gin.H{
			"message": "Notification deleted successfully",
		})
		return nil
	})
	if err != nil {
		handleError(c, err)
		return
	}
}
