package api

import (
	"slices"

	"github.com/gin-gonic/gin"
	"github.com/mekdep/server/internal/api/utils"
	"github.com/mekdep/server/internal/app"
	"github.com/mekdep/server/internal/models"
)

func MessageGroupRoutes(api *gin.RouterGroup) {
	messageGroupRoutes := api.Group("/message-groups")
	{
		messageGroupRoutes.GET("", GetMessageGroups)
	}
}

func GetMessageGroups(c *gin.Context) {
	ses := utils.InitSession(c)
	err := app.Ap().UserActionCheckRead(&ses, app.PermUser, func(user *models.User) error {
		dto := models.GetMessageGroupsRequest{}
		if errMsg, errKey := BindAndValidate(c, &dto); errMsg != "" || errKey != "" {
			return app.NewAppError(errMsg, errKey, "")
		}
		if dto.Limit == nil {
			dto.Limit = new(int)
			*dto.Limit = 12
		}
		if dto.Offset == nil {
			dto.Offset = new(int)
			*dto.Offset = 0
		}

		messageGroupsResponse := []*models.MessageGroupResponse{}
		allowedRoles := []models.Role{
			models.RoleParent,
			models.RoleTeacher,
			models.RoleAdmin,
		}

		role := ses.GetRole()
		if !slices.Contains(allowedRoles, *role) {
			return app.ErrForbidden
		}

		messageGroups, total, err := app.GetMessageGroups(&ses, dto)
		if err != nil {
			return err
		}

		for _, messageGroup := range messageGroups {
			response := models.MessageGroupResponse{}
			response.FromModel(messageGroup)
			messageGroupsResponse = append(messageGroupsResponse, &response)
		}

		Success(c, gin.H{
			"message_groups": messageGroupsResponse,
			"total":          total,
		})
		return nil
	})
	if err != nil {
		handleError(c, err)
		return
	}
}
