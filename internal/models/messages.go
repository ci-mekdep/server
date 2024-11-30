package models

import "time"

// TODO: messages refactor (created_at cant be null)
type Message struct {
	ID        string     `json:"id"`
	UserId    string     `json:"user_id"`
	SessionId *string    `json:"session_id"`
	GroupId   string     `json:"group_id"`
	ParentId  *string    `json:"parent_id"`
	Message   *string    `json:"message"`
	Files     *[]string  `json:"files"`
	CreatedAt *time.Time `json:"created_at"`

	User    *User         `json:"user"`
	Session *Session      `json:"session"`
	Group   *MessageGroup `json:"group"`
	Parent  *Message      `json:"parent"`
}

type Messages struct {
	Messages []*Message `json:"messages"`
	Total    int        `json:"total"`
}

// TODO: add below dto to domain.dtos layer

type GetMessagesQueryDto struct {
	GroupId *string
	Limit   *uint
	Offset  uint
}
type GetMessageReadsQueryDto struct {
	UserIds       *[]string
	IsRead        *bool
	IsNotified    *bool
	NotifiedAtMax *time.Time
	SetNotified   *bool
}

func (dto *GetMessagesQueryDto) SetDefaults() {
	if dto.Limit == nil {
		dto.Limit = new(uint)
		*dto.Limit = 100
	}
}

func (Message) RelationFields() []string {
	return []string{"User", "Session", "Group", "Parent"}
}

type MessageRead struct {
	ID        string     `json:"id"`
	UserId    string     `json:"user_id"`
	SessionId string     `json:"session_id"`
	MessageId string     `json:"message_id"`
	ReadAt    *time.Time `json:"read_at"`
}

type MessageRequest struct {
	ID        *string   `json:"id"`
	UserId    *string   `json:"user_id"`
	SessionId *string   `json:"session_id"`
	GroupId   *string   `json:"group_id" form:"group_id"`
	ParentId  *string   `json:"parent_id"`
	Message   *string   `json:"message"`
	Files     *[]string `json:"files"`
}

type MessageFilterRequest struct {
	ID          *string   `json:"id"`
	UserId      *string   `json:"user_id"`
	UserIds     *[]string `json:"user_ids[]"`
	SessionId   *string   `json:"session_id"`
	SessionIds  *[]string `json:"session_ids[]"`
	GroupId     *string   `json:"group_id" form:"group_id"`
	GroupIds    *[]string `json:"group_ids[]"`
	Files       *[]string ``
	GetUsers    *bool     `json:"get_users" form:"get_users"`
	ClassroomId *string   `json:"classroom_id" form:"classroom_id"`
	Search      *string   `json:"search"`
	Sort        *string   `json:"sort"`
	PaginationRequest
}

type MessageResponse struct {
	ID        string                `json:"id"`
	UserId    string                `json:"user_id"`
	SessionId *string               `json:"session_id"`
	GroupId   string                `json:"group_id"`
	Message   *string               `json:"message"`
	Files     *[]string             `json:"files"`
	CreatedAt *time.Time            `json:"created_at"`
	User      *UserResponse         `json:"user"`
	Session   *SessionResponse      `json:"session"`
	Group     *MessageGroupResponse `json:"group"`
	Parent    *MessageResponse      `json:"parent"`
}

func (response *MessageResponse) FromModel(model *Message) {
	if model.Message == nil {
		model.Message = new(string)
	}
	if model.Files == nil {
		model.Files = new([]string)
	}
	response.ID = model.ID
	response.UserId = model.UserId
	response.SessionId = model.SessionId
	response.GroupId = model.GroupId
	response.Message = model.Message
	response.CreatedAt = model.CreatedAt

	if model.User != nil {
		user_response := UserResponse{}
		user_response.FromModel(model.User)
		response.User = &user_response
	}
	if model.Session != nil {
		session_response := SessionResponse{}
		session_response.FromModel(model.Session)
		response.Session = &session_response
	}
	if model.Group != nil {
		group_response := MessageGroupResponse{}
		group_response.FromModel(model.Group)
		response.Group = &group_response
	}
	if model.Parent != nil {
		parent_response := MessageResponse{}
		parent_response.FromModel(model.Parent)
		response.Parent = &parent_response
	}
}

func (model *Message) FromRequest(request *MessageRequest) error {
	if request.ID != nil {
		model.ID = *request.ID
	}
	if request.Message != nil {
		model.Message = request.Message
	}
	if request.Files != nil {
		model.Files = request.Files
	}
	model.UserId = *request.UserId
	model.SessionId = request.SessionId
	model.GroupId = *request.GroupId
	model.ParentId = request.ParentId
	model.User = &User{}
	model.Session = &Session{}
	model.Group = &MessageGroup{}
	return nil
}

func SerializeMessages(messages []*Message) []*MessageResponse {
	messagesResponse := []*MessageResponse{}
	for _, message := range messages {
		response := MessageResponse{}
		response.FromModel(message)
		messagesResponse = append(messagesResponse, &response)
	}
	return messagesResponse
}
