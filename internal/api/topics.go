package api

import (
	"fmt"
	"slices"

	"github.com/gin-gonic/gin"
	"github.com/mekdep/server/internal/api/utils"
	"github.com/mekdep/server/internal/app"
	"github.com/mekdep/server/internal/app/app_validation"
	"github.com/mekdep/server/internal/models"
)

func TopicRoutes(api *gin.RouterGroup) {
	topicRoutes := api.Group("/topics")
	{
		topicRoutes.GET("", TopicsList)
		topicRoutes.GET("/values", TopicsListValues)
		topicRoutes.PUT(":id", TopicsUpdate)
		topicRoutes.POST("", TopicsCreate)
		topicRoutes.DELETE("", TopicsDelete)
		topicRoutes.GET(":id", TopicsDetail)
		topicRoutes.POST("/batch", TopicsMultipleCreate)
	}
}

func topicsListQuery(ses *utils.Session, data models.TopicsFilterRequest, IsValue bool) ([]*models.TopicsResponse, []models.TopicsValueResponse, int, error) {
	if IsValue {
		res, total, err := app.TopicsListValues(ses, data)
		return nil, res, total, err
	} else {
		res, total, err := app.TopicsList(ses, data)
		return res, nil, total, err
	}
}

func topicsAvailableCheck(ses *utils.Session, data models.TopicsFilterRequest) (bool, error) {
	_, _, t, err := topicsListQuery(ses, data, false)
	if err != nil {
		return false, app.ErrRequired.SetKey("id")
	}
	return t > 0, nil
}

func TopicsList(c *gin.Context) {
	ses := utils.InitSession(c)
	r := models.TopicsFilterRequest{}
	if errMsg, errKey := BindAndValidate(c, &r); errMsg != "" || errKey != "" {
		app.NewAppError(errMsg, errKey, "")
		return
	}

	topics, _, total, err := topicsListQuery(&ses, r, false)
	if err != nil {
		handleError(c, err)
		return
	}

	Success(c, gin.H{
		"topics": topics,
		"total":  total,
	})
}

func TopicsListValues(c *gin.Context) {
	ses := utils.InitSession(c)
	r := models.TopicsFilterRequest{}
	if errMsg, errKey := BindAndValidate(c, &r); errMsg != "" || errKey != "" {
		handleError(c, app.NewAppError(errMsg, errKey, ""))
	}

	_, topics, total, err := topicsListQuery(&ses, r, true)
	if err != nil {
		handleError(c, err)
	}
	Success(c, gin.H{
		"total":  total,
		"topics": topics,
	})
	return
}

func TopicsDetail(c *gin.Context) {
	ses := utils.InitSession(c)
	err := app.Ap().UserActionCheckRead(&ses, app.PermAdminTopics, func(u *models.User) error {
		id := c.Param("id")
		if ok, err := topicsAvailableCheck(&ses, models.TopicsFilterRequest{ID: &id}); err != nil {
			return err
		} else if !ok {
			return app.ErrNotfound
		}
		m, err := app.TopicsDetail(&ses, id)
		if err != nil {
			return err
		}
		Success(c, gin.H{
			"topic": m,
		})
		return nil
	})
	if err != nil {
		handleError(c, err)
		return
	}
}

func TopicsUpdate(c *gin.Context) {
	ses := utils.InitSession(c)
	err := app.Ap().UserActionCheckWrite(&ses, app.PermAdminTopics, func(user *models.User) (err error) {
		r := models.TopicsRequest{}
		if errMsg, errKey := BindAndValidate(c, &r); errMsg != "" || errKey != "" {
			return app.NewAppError(errMsg, errKey, "")
		}
		id := c.Param("id")
		r.ID = &id

		if id == "" {
			return app.ErrRequired.SetKey("id")
		}

		var model models.Topics
		if model.Files == nil {
			model.Files = new([]string)
		}

		paths := *model.Files
		if r.Files != nil {
			if len(*r.FilesDelete) < 1 || (*r.FilesDelete)[0] == "" {
				*r.FilesDelete = paths
			}
			for _, v := range *r.FilesDelete {
				deleteFile(c, v, "topics")
				k := slices.Index(paths, v)
				if k >= 0 {
					paths = slices.Delete(paths, k, k+1)
				}
			}
		}

		files, err := handleFilesUpload(c, "files", "topics")
		if err != nil {
			return app.NewAppError(err.Error(), "files", "")
		}
		paths = append(paths, files...)
		r.Files = &paths

		if v, ok := c.Request.Form["files_delete"]; ok {
			r.FilesDelete = &v
		}

		if ok, err := topicsAvailableCheck(&ses, models.TopicsFilterRequest{ID: r.ID}); err != nil {
			return err
		} else if !ok {
			return app.ErrForbidden
		}
		topic, err := app.TopicsUpdate(&ses, r)

		if err != nil {
			return err
		}
		userLog(models.UserLog{
			SchoolId:          ses.GetSchoolId(),
			SessionId:         ses.GetSessionId(),
			UserId:            user.ID,
			SubjectId:         r.ID,
			Subject:           models.LogSubjectTopics,
			SubjectAction:     models.LogActionUpdate,
			SubjectProperties: r,
		})

		Success(c, gin.H{
			"topic": topic,
		})
		return nil
	})
	if err != nil {
		handleError(c, err)
		return
	}

}

func TopicsCreate(c *gin.Context) {
	ses := utils.InitSession(c)
	err := app.Ap().UserActionCheckWrite(&ses, app.PermAdminTopics, func(user *models.User) (err error) {
		r := models.TopicsRequest{}
		if err := BindAny(c, &r); err != nil {
			return err
		}

		err = app_validation.ValidateTopicsCreate(&ses, r)
		if err != nil {
			return err
		}

		var model models.Topics
		if model.Files == nil {
			model.Files = new([]string)
		}

		files, err := handleFilesUpload(c, "files", "topics")
		if err != nil {
			return app.NewAppError(err.Error(), "files", "")
		}

		if files != nil {
			r.Files = &files
		}

		topic, err := app.TopicsCreate(&ses, r)
		if err != nil {
			return err
		}
		userLog(models.UserLog{
			SchoolId:          ses.GetSchoolId(),
			SessionId:         ses.GetSessionId(),
			UserId:            user.ID,
			SubjectId:         &topic.ID,
			Subject:           models.LogSubjectTopics,
			SubjectAction:     models.LogActionCreate,
			SubjectProperties: r,
		})
		Success(c, gin.H{
			"topic": topic,
		})
		return nil
	})
	if err != nil {
		handleError(c, err)
		return
	}
}

func TopicsMultipleCreate(c *gin.Context) {
	ses := utils.InitSession(c)
	err := app.Ap().UserActionCheckWrite(&ses, app.PermAdminTopics, func(user *models.User) error {
		r := models.TopicsMultipleRequest{}
		if errMsg, errKey := BindAndValidate(c, &r); errMsg != "" || errKey != "" {
			return app.NewAppError(errMsg, errKey, "")
		}
		for idx, item := range r.Topics {
			idx += 1
			if item.SubjectName == nil || *item.SubjectName == "" {
				return app.NewAppError("required", fmt.Sprintf("items.%d.subject", idx), "subject cannot be empty")
			}
			if item.Language == nil || *item.Language == "" {
				return app.NewAppError("required", fmt.Sprintf("items.%d.language", idx), "language cannot be empty")
			}
			if item.Title == nil || *item.Title == "" {
				return app.NewAppError("required", fmt.Sprintf("items.%d.title", idx), "title cannot be empty")
			}
		}
		topics, err := app.TopicsMultipleCreate(&ses, r)
		if err != nil {
			return err
		}
		Success(c, gin.H{
			"topics": topics,
		})
		return nil
	})
	if err != nil {
		handleError(c, err)
		return
	}
}

func TopicsDelete(c *gin.Context) {
	ses := utils.InitSession(c)
	err := app.Ap().UserActionCheckWrite(&ses, app.PermAdminTopics, func(user *models.User) error {
		var ids []string = c.QueryArray("ids")
		if len(ids) == 0 {
			return app.ErrRequired.SetKey("ids")
		}

		if ok, err := topicsAvailableCheck(&ses, models.TopicsFilterRequest{IDs: &ids}); err != nil {
			return err
		} else if !ok {
			return app.ErrForbidden
		}

		topics, err := app.TopicsDelete(&ses, ids)
		if err != nil {
			return err
		}
		userLog(models.UserLog{
			SchoolId:          ses.GetSchoolId(),
			SessionId:         ses.GetSessionId(),
			UserId:            user.ID,
			Subject:           models.LogSubjectTopics,
			SubjectAction:     models.LogActionDelete,
			SubjectProperties: ids,
		})
		Success(c, gin.H{
			"topics": topics,
		})
		return nil
	})
	if err != nil {
		handleError(c, err)
		return
	}
}
