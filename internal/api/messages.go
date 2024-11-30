package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/mekdep/server/internal/api/utils"
	"github.com/mekdep/server/internal/app"
	"github.com/mekdep/server/internal/models"
)

func MessageRoutes(api *gin.RouterGroup) {
	messageRoutes := api.Group("/messages")
	{
		messageRoutes.GET("", GetMessagesAndMembers)
		messageRoutes.GET("/connect", ConnectAndHandleMessages)
	}
}

func GetMessagesAndMembers(c *gin.Context) {
	ses := utils.InitSession(c)
	err := app.Ap().UserActionCheckRead(&ses, app.PermUser, func(user *models.User) error {
		dto := models.MessageFilterRequest{}
		if errMsg, errKey := BindAndValidate(c, &dto); errMsg != "" || errKey != "" {
			return app.NewAppError(errMsg, errKey, "")
		}
		if dto.Limit == nil {
			dto.Limit = new(int)
			*dto.Limit = 50
		}
		if dto.Offset == nil {
			dto.Offset = new(int)
			*dto.Offset = 0
		}
		var response = map[string]interface{}{}

		messageGroup, err := app.GetMessageGroup(&ses, *dto.GroupId)
		if err != nil {
			return err
		}
		messages, messagesTotal, err := app.GetMessagesByMessageGroup(&ses, dto, *messageGroup)
		if err != nil {
			return err
		}
		response["messages"] = models.SerializeMessages(messages)
		response["messages_total"] = messagesTotal

		if dto.GetUsers != nil && *dto.GetUsers {
			members, membersTotal, err := app.GetMessageGroupMembers(&ses, *messageGroup, dto.ClassroomId)
			if err != nil {
				return err
			}
			response["users"] = models.SerializeUsers(members)
			response["users_total"] = membersTotal
		}

		ginResponse := gin.H{}
		for key, value := range response {
			ginResponse[key] = value
		}
		Success(c, ginResponse)
		return nil
	})
	if err != nil {
		handleError(c, err)
		return
	}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func ConnectAndHandleMessages(c *gin.Context) {
	ses := utils.InitSession(c)
	err := app.Ap().UserActionCheckRead(&ses, app.PermUser, func(user *models.User) error {
		dto := models.MessageRequest{}
		if errMsg, errKey := BindAndValidate(c, &dto); errMsg != "" || errKey != "" {
			return app.NewAppError(errMsg, errKey, "")
		}
		upgrader.CheckOrigin = func(r *http.Request) bool { return true }
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			return err
		}

		err = app.ConnectAndHandleMessages(&ses, &dto, conn)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		handleError(c, err)
		return
	}
}
