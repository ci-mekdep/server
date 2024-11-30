package app

import (
	"context"
	"encoding/json"
	"slices"
	"time"

	"github.com/gorilla/websocket"
	"github.com/mekdep/server/internal/api/utils"
	"github.com/mekdep/server/internal/models"
	"github.com/mekdep/server/internal/store"
	apputils "github.com/mekdep/server/internal/utils"
	"go.elastic.co/apm/v2"
)

func GetMessagesByMessageGroup(ses *utils.Session, dto models.MessageFilterRequest, group models.MessageGroup) ([]*models.Message, int, error) {
	sp, ctx := apm.StartSpan(ses.Context(), "GetMessagesByMessageGroup", "app")
	ses.SetContext(ctx)
	defer sp.End()

	var messages []*models.Message
	var total int
	var err error

	switch group.Type {
	case string(models.MessageGroupParentsType):
		messages, total, err = getParentsGroupMessages(ses, dto)
		if err != nil {
			return nil, 0, err
		}
	case string(models.MessageGroupTeachersType):
		messages, total, err = getTeachersGroupMessages(ses, dto)
		if err != nil {
			return nil, 0, err
		}
	default:
		return nil, 0, ErrForbidden
	}

	messages_reads := []models.MessageRead{}
	for _, message := range messages {
		message_read := models.MessageRead{
			UserId:    ses.GetUser().ID,
			SessionId: *ses.GetSessionId(),
			MessageId: message.ID,
		}
		messages_reads = append(messages_reads, message_read)
	}
	err = createMessageReads(ses, messages_reads)
	if err != nil {
		return nil, 0, err
	}

	return messages, total, err
}

func createMessageReads(ses *utils.Session, messages_reads []models.MessageRead) error {
	err := store.Store().CreateMessageReadsCommand(context.Background(), messages_reads)

	if err != nil {
		return err
	}

	return nil
}

func getParentsGroupMessages(ses *utils.Session, dto models.MessageFilterRequest) ([]*models.Message, int, error) {
	allowedRoles := []models.Role{models.RoleParent, models.RoleTeacher}
	role := ses.GetRole()
	user := ses.GetUser()

	if !slices.Contains(allowedRoles, *role) {
		return nil, 0, ErrForbidden
	}
	if dto.ClassroomId == nil || *(dto.ClassroomId) == "" {
		return nil, 0, ErrForbidden
	}

	if *role == models.RoleParent {
		childrenDTO := models.UserFilterRequest{
			ClassroomId: dto.ClassroomId,
			ParentId:    &(user.ID),
		}
		_, childrenTotal, err := store.Store().UsersFindBy(ses.Context(), childrenDTO)
		if err != nil {
			return nil, 0, err
		}
		if childrenTotal == 0 {
			return nil, 0, ErrForbidden
		}
	}

	messages, err := store.Store().GetMessagesQuery(
		ses.Context(),
		models.GetMessagesQueryDto{
			GroupId: dto.GroupId,
		},
	)
	if err != nil {
		return nil, 0, err
	}
	err = store.Store().LoadMessagesWithParents(ses.Context(), &messages)
	if err != nil {
		return nil, 0, err
	}

	return messages, 0, err
}

func getTeachersGroupMessages(ses *utils.Session, dto models.MessageFilterRequest) ([]*models.Message, int, error) {
	allowedRoles := []models.Role{models.RoleTeacher}
	role := ses.GetRole()

	if !slices.Contains(allowedRoles, *role) {
		return nil, 0, ErrForbidden
	}
	if dto.ClassroomId != nil {
		dto.ClassroomId = nil
	}

	messages, err := store.Store().GetMessagesQuery(
		ses.Context(),
		models.GetMessagesQueryDto{
			GroupId: dto.GroupId,
		},
	)
	if err != nil {
		return nil, 0, err
	}
	err = store.Store().LoadMessagesWithParents(ses.Context(), &messages)
	if err != nil {
		return nil, 0, err
	}
	return messages, 0, err
}

func CreateMessage(ses *utils.Session, dto *models.MessageRequest) (*models.Message, error) {
	sp, ctx := apm.StartSpan(ses.Context(), "CreateMessage", "app")
	ses.SetContext(ctx)
	defer sp.End()

	model := models.Message{}
	model.FromRequest(dto)
	res := &models.MessageResponse{}
	message, err := store.Store().CreateMessageCommand(ses.Context(), model)
	if err != nil {
		return nil, err
	}
	err = store.Store().LoadMessagesWithParents(ses.Context(), &[]*models.Message{&message})
	if err != nil {
		return nil, err
	}
	res.FromModel(&message)
	return &message, err
}

type Client struct {
	Conn      *websocket.Conn
	UserId    *string
	SessionId *string
}

var groupsWithClients = map[string]map[*Client]bool{}

func ConnectAndHandleMessages(ses *utils.Session, dto *models.MessageRequest, conn *websocket.Conn) error {
	client := &Client{
		Conn:      conn,
		UserId:    &ses.GetUser().ID,
		SessionId: ses.GetSessionId(),
	}
	if innerMap, ok := groupsWithClients[*dto.GroupId]; ok {
		innerMap[client] = true
	} else {
		clientMap := map[*Client]bool{
			client: true,
		}
		groupsWithClients[*dto.GroupId] = clientMap
	}

	defer func() {
		conn.Close()
		delete(groupsWithClients[*dto.GroupId], client)
		if len(groupsWithClients[*dto.GroupId]) == 0 {
			delete(groupsWithClients, *dto.GroupId)
		}
	}()

	dto.UserId = &ses.GetUser().ID
	dto.SessionId = ses.GetSessionId()
	err := listenMessage(ses, dto, conn)
	if err != nil {
		return err
	}

	return nil
}

const (
	stringType int = 1
	fileType   int = 2
)

func listenMessage(ses *utils.Session, dto *models.MessageRequest, conn *websocket.Conn) error {
	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			return err
		}

		if messageType == stringType {
			messageString := string(message)
			dto.Message = &messageString
		} else if messageType == fileType {
			// implement real file uploading here
			// file, err := handleFileUpload
			filePath := "web/uploads/messages/file" + string(message)[:20]
			dto.Message = &filePath
			// if err != nil {
			// 	return err
			// }
		}

		err = broadcastMessage(ses, dto)
		if err != nil {
			return err
		}
	}
}

func broadcastMessage(ses *utils.Session, dto *models.MessageRequest) error {

	// fill message model from request
	messageModel := models.Message{}
	messageModel.FromRequest(dto)

	// get user response data
	userDto := models.UserFilterRequest{}
	userDto.ID = dto.UserId
	userResponse, err := UsersDetail(ses, userDto)
	if err != nil {
		return err
	}

	// fill message response from model
	messageResponse := models.MessageResponse{}
	messageResponse.FromModel(&messageModel)
	now := time.Now()
	messageResponse.CreatedAt = &now
	messageResponse.User = userResponse

	// convert message response to json object
	messageResponseInJSON, err := json.Marshal(messageResponse)
	if err != nil {
		return err
	}

	// broadcast to all group's clients
	for client := range groupsWithClients[*dto.GroupId] {
		err := client.Conn.WriteMessage(websocket.TextMessage, messageResponseInJSON)
		if err != nil {
			return err
		}
	}

	// store message to db
	go func() {
		model := models.Message{}
		model.FromRequest(dto)

		message, err := store.Store().CreateMessageCommand(context.Background(), model)
		if err != nil {
			apputils.Logger.Error(err)
		}

		messages_reads := []models.MessageRead{}
		for client := range groupsWithClients[*dto.GroupId] {
			if client.SessionId == nil {
				client.SessionId = new(string)
			}
			message_read := models.MessageRead{
				UserId:    *client.UserId,
				SessionId: *client.SessionId,
				MessageId: message.ID,
			}
			messages_reads = append(messages_reads, message_read)
		}
		err = createMessageReads(ses, messages_reads)
		if err != nil {
			apputils.Logger.Error(err)
		}

		return
	}()

	return nil
}
