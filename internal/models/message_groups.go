package models

type MessageGroupType string

const (
	MessageGroupParentsType  MessageGroupType = "parent"
	MessageGroupTeachersType MessageGroupType = "teacher"
)

type MessageGroup struct {
	ID             string     `json:"id"`
	AdminId        string     `json:"admin_id"`
	Title          string     `json:"title"`
	SchoolId       string     `json:"school_id"`
	Type           string     `json:"type"`
	Description    *string    `json:"description"`
	ClassroomId    *string    `json:"classroom_id"`
	UnreadMessages *int       `json:"unread_messages"`
	Admin          *User      `json:"admin"`
	School         *School    `json:"school"`
	Classroom      *Classroom `json:"classroom"`
}

func (MessageGroup) RelationFields() []string {
	return []string{"Admin", "School", "Classroom"}
}

func (model *MessageGroup) SetDefaults() {
	if model.Admin == nil {
		model.Admin = &User{}
	}
	if model.School == nil {
		model.School = &School{}
	}
	if model.Classroom == nil {
		model.Classroom = &Classroom{}
	}
}

type MessageGroupResponse struct {
	ID          string             `json:"id"`
	AdminId     *string            `json:"admin_id"`
	Title       *string            `json:"title"`
	Description *string            `json:"description"`
	SchoolId    *string            `json:"school_id"`
	ClassroomId *string            `json:"classroom_id"`
	Type        *string            `json:"type"`
	Admin       *UserResponse      `json:"admin"`
	School      *SchoolResponse    `json:"school"`
	Classroom   *ClassroomResponse `json:"classroom"`
	UnreadCount *int               `json:"unread_count"`
}

func (response *MessageGroupResponse) FromModel(model *MessageGroup) {
	if model.Description == nil {
		model.Description = new(string)
	}
	response.ID = model.ID
	response.AdminId = &model.AdminId
	response.Title = &model.Title
	response.SchoolId = &model.SchoolId
	response.ClassroomId = model.ClassroomId
	response.Type = &model.Type
	response.UnreadCount = model.UnreadMessages

	if model.Admin != nil {
		user_response := UserResponse{}
		user_response.FromModel(model.Admin)
		response.Admin = &user_response
	}
	if model.School != nil {
		school_response := SchoolResponse{}
		school_response.FromModel(model.School)
		response.School = &school_response
	}
	if model.Classroom != nil {
		classroom_response := ClassroomResponse{}
		classroom_response.FromModel(model.Classroom)
		response.Classroom = &classroom_response
	}
	if model.Classroom != nil {
		classroom_response := ClassroomResponse{}
		classroom_response.FromModel(model.Classroom)
		response.Classroom = &classroom_response
	}
}

type GetMessageGroupsRequest struct {
	ID            *string   `json:"id"`
	Ids           *[]string `json:"ids[]"`
	SchoolId      *string   `json:"school_id"`
	ClassroomId   *string   `json:"classroom_id" form:"classroom_id"`
	Type          *string   `json:"type"`
	Types         *[]string `json:"types"`
	CurrentUserId *string   `json:"current_user_id"`

	SchoolIds    *[]string `json:"school_ids[]"`
	AdminId      *string   `json:"admin_id"`
	AdminIds     *[]string `json:"admin_ids[]"`
	ClassroomIds *[]string `json:"classroom_ids[]"`
	Search       *string   `json:"search"`
	Sort         *string   `json:"sort"`
	PaginationRequest
}
