package api

import (
	"github.com/gin-gonic/gin"
	"github.com/mekdep/server/internal/api/utils"
	"github.com/mekdep/server/internal/app"
	"github.com/mekdep/server/internal/app/app_validation"
	"github.com/mekdep/server/internal/models"
)

func ContactItemsRoutes(api *gin.RouterGroup) {
	contactItemsRoutes := api.Group("/contact-items")
	{
		contactItemsRoutes.GET("", ContactItemsList)
		contactItemsRoutes.PUT(":id", ContactItemsUpdate)
		contactItemsRoutes.POST("", ContactItemsCreate)
		contactItemsRoutes.DELETE("", ContactItemsDelete)
		contactItemsRoutes.GET(":id", ContactItemsDetail)
	}
}

func contactItemsListQuery(ses *utils.Session, data models.ContactItemsFilterRequest) ([]*models.ContactItemsResponse, int, error) {
	return app.ContactItemsList(ses, data)
}

func contactItemsAvailableCheck(ses *utils.Session, data models.ContactItemsFilterRequest) (bool, error) {
	_, t, err := contactItemsListQuery(ses, data)
	if err != nil {
		return false, app.ErrRequired.SetKey("id")
	}
	return t > 0, nil
}

func ContactItemsList(c *gin.Context) {
	ses := utils.InitSession(c)
	err := app.Ap().UserActionCheckRead(&ses, app.PermAdminContactItems, func(user *models.User) error {
		r := models.ContactItemsFilterRequest{}
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
		r.OnlyNotRelated = new(bool)
		*r.OnlyNotRelated = true

		contactItems, total, err := contactItemsListQuery(&ses, r)
		if err != nil {
			return err
		}

		Success(c, gin.H{
			"contact_items": contactItems,
			"total":         total,
		})
		return nil
	})
	if err != nil {
		handleError(c, err)
		return
	}

}

func ContactItemsDetail(c *gin.Context) {
	ses := utils.InitSession(c)
	err := app.Ap().UserActionCheckRead(&ses, app.PermAdminContactItems, func(u *models.User) error {
		id := c.Param("id")
		if ok, err := contactItemsAvailableCheck(&ses, models.ContactItemsFilterRequest{ID: &id}); err != nil {
			return err
		} else if !ok {
			return app.ErrNotfound
		}
		m, err := app.ContactItemsDetail(&ses, id)
		if err != nil {
			return err
		}
		Success(c, gin.H{
			"contact_item": m,
		})
		return nil
	})
	if err != nil {
		handleError(c, err)
		return
	}
}

func ContactItemsUpdate(c *gin.Context) {
	ses := utils.InitSession(c)
	err := app.Ap().UserActionCheckWrite(&ses, app.PermAdminContactItems, func(user *models.User) (err error) {
		r := models.ContactItemsRequest{}
		if errMsg, errKey := BindAndValidate(c, &r); errMsg != "" || errKey != "" {
			return app.NewAppError(errMsg, errKey, "")
		}
		id := c.Param("id")
		r.ID = &id

		if id == "" {
			return app.ErrRequired.SetKey("id")
		}

		if ok, err := contactItemsAvailableCheck(&ses, models.ContactItemsFilterRequest{ID: r.ID}); err != nil {
			return err
		} else if !ok {
			return app.ErrForbidden
		}
		contactItem, err := app.ContactItemsUpdate(&ses, r)

		if err != nil {
			return err
		}
		Success(c, gin.H{
			"contact_item": contactItem,
		})
		return nil
	})
	if err != nil {
		handleError(c, err)
		return
	}

}

func ContactItemsCreate(c *gin.Context) {
	ses := utils.InitSession(c)
	if _, err := c.MultipartForm(); err == nil {
		ContactItemsCreateHandleFiles(c)
	} else {
		r := models.ContactItemsRequest{}
		if err := BindAny(c, &r); err != nil {
			handleError(c, err)
			return
		}
		if user := ses.GetUser(); user != nil {
			r.UserId = new(string)
			*r.UserId = user.ID
			schoolId := c.GetString("school_id")
			r.SchoolId = &schoolId
		}
		r.Status = string(models.ContactStatusWaiting)
		r.RelatedId = nil

		err := app_validation.ValidateContactItemsCreate(&ses, r)
		if err != nil {
			handleError(c, err)
			return
		}
		contactItem, err := app.ContactItemsCreate(&ses, r)
		if err != nil {
			handleError(c, err)
			return
		}
		Success(c, gin.H{
			"contact_item": contactItem,
		})
	}
}

func ContactItemsCreateHandleFiles(c *gin.Context) {
	ses := utils.InitSession(c)
	r := models.ContactItemsRequest{}
	if err := BindAny(c, &r); err != nil {
		handleError(c, err)
		return
	}

	if user := ses.GetUser(); user != nil {
		r.UserId = new(string)
		*r.UserId = user.ID
		r.SchoolId = ses.GetSchoolId()
	}
	r.Status = string(models.ContactStatusWaiting)
	r.RelatedId = nil

	err := app_validation.ValidateContactItemsCreate(&ses, r)
	if err != nil {
		handleError(c, err)
		return
	}
	files, err := handleFilesUpload(c, "files", "contact_files")
	if err != nil {
		handleError(c, app.NewAppError(err.Error(), "files", ""))
		return
	}
	if files != nil {
		r.Files = &files
	}
	contactItem, err := app.ContactItemsCreate(&ses, r)
	if err != nil {
		handleError(c, err)
		return
	}
	Success(c, gin.H{
		"contact_item": contactItem,
	})
}

func ContactItemsDelete(c *gin.Context) {
	ses := utils.InitSession(c)
	err := app.Ap().UserActionCheckWrite(&ses, app.PermAdminContactItems, func(user *models.User) error {
		var ids []string = c.QueryArray("ids")
		if len(ids) == 0 {
			return app.ErrRequired.SetKey("ids")
		}

		if ok, err := contactItemsAvailableCheck(&ses, models.ContactItemsFilterRequest{IDs: &ids}); err != nil {
			return err
		} else if !ok {
			return app.ErrForbidden
		}

		contactItems, err := app.ContactItemsDelete(&ses, ids)
		if err != nil {
			return err
		}
		Success(c, gin.H{
			"contact_items": contactItems,
		})
		return nil
	})
	if err != nil {
		handleError(c, err)
		return
	}
}
