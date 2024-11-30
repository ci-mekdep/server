package models

import "time"

type Notifications struct {
	ID        string     `json:"id"`
	SchoolIds []string   `json:"school_ids"`
	Roles     []string   `json:"roles"`
	UserIds   []string   `json:"user_ids"`
	AuthorID  *string    `json:"author_id"`
	Title     *string    `json:"title"`
	Content   *string    `json:"content"`
	Files     *[]string  `json:"files"`
	UpdatedAt *time.Time `json:"updated_at"`
	CreatedAt *time.Time `json:"created_at"`
	Schools   *[]School  `json:"school"`
	Author    *User      `json:"author"`
}

func (Notifications) RelationFields() []string {
	return []string{"Schools", "Author"}
}

type NotificationsFilterRequest struct {
	ID       *string   `form:"id"`
	UserIds  *[]string `form:"user_ids[]"`
	AuthorId *string   `form:"author_id"`
	Search   *string   `form:"search"`
	Sort     *string   `form:"sort"`
	PaginationRequest
}

type NotificationsRequest struct {
	ID          *string   `json:"id" form:"id"`
	SchoolIds   []string  `json:"school_ids" form:"school_ids[]"`
	Roles       []string  `json:"roles" form:"roles[]"`
	UserIds     []string  `json:"user_ids" form:"user_ids[]"`
	AuthorId    string    `json:"author_id" form:"author_id"`
	Title       *string   `json:"title" form:"title"`
	Content     *string   `json:"content" form:"content"`
	Files       *[]string ``
	FilesDelete *[]string `json:"files_delete" form:"files_delete"`
}

type NotificationsResponse struct {
	ID        string                      `json:"id"`
	SchoolIds []string                    `json:"school_ids"`
	Roles     []string                    `json:"roles"`
	UserIds   []string                    `json:"user_ids"`
	AuthorId  *string                     `json:"author_id"`
	Title     *string                     `json:"title"`
	Content   *string                     `json:"content"`
	Files     *[]string                   `json:"files"`
	Items     *[]UserNotificationResponse `json:"items"`
	UpdatedAt *time.Time                  `json:"updated_at"`
	CreatedAt *time.Time                  `json:"created_at"`
	Author    *UserResponse               `json:"author"`
}

func (r *NotificationsResponse) FromModel(m *Notifications) error {
	r.ID = m.ID
	if m.SchoolIds != nil {
		r.SchoolIds = m.SchoolIds
	}
	if m.Roles != nil {
		r.Roles = m.Roles
	}
	if m.UserIds != nil {
		r.UserIds = m.UserIds
	}
	r.AuthorId = m.AuthorID
	r.Title = m.Title
	r.Content = m.Content
	r.UpdatedAt = m.UpdatedAt
	if m.CreatedAt != nil {
		localTime := m.CreatedAt.Local()
		r.CreatedAt = &localTime
	}
	if m.Files != nil {
		r.Files = &[]string{}
		for _, f := range *m.Files {
			*r.Files = append(*r.Files, fileUrl(&f))
		}
	}
	if m.Author != nil {
		r.Author = &UserResponse{}
		r.Author.FromModel(m.Author)
	}
	return nil
}

func (r *NotificationsRequest) ToModel(m *Notifications) error {
	if r.ID == nil {
		r.ID = new(string)
	}
	m.ID = *r.ID
	m.SchoolIds = r.SchoolIds
	m.Roles = r.Roles
	m.UserIds = r.UserIds
	m.AuthorID = &r.AuthorId
	m.Title = r.Title
	m.Content = r.Content
	m.Files = r.Files
	return nil
}

// ++++++++++++++++ USER NOTIFICATIONS +++++++++++++++++++++

type UserNotification struct {
	ID             string         `json:"id"`
	NotificationId string         `json:"notification_id"`
	UserId         string         `json:"user_id"`
	ReadAt         *time.Time     `json:"read_at"`
	Role           *string        `json:"role"`
	Comment        *string        `json:"comment"`
	CommentFiles   *[]string      `json:"comment_files"`
	Notifications  *Notifications `json:"notifications"`
	User           *User          `json:"user"`
}

func (UserNotification) RelationFields() []string {
	return []string{"Notifications", "User"}
}

type UserNotificationFilterRequest struct {
	NotificationId *string   `form:"notification_id"`
	IDs            *[]string `form:"id[]"`
	UserId         *string   `form:"user_id"`
	Role           *string   `form:"role"`
	IsRead         *bool     `form:"read"`
	Sort           *string   `form:"sort"`
	PaginationRequest
}

type UserNotificationRequest struct {
	ID                 *string   `json:"id" form:"id"`
	NotificationId     string    `json:"notification_id" form:"notification_id"`
	UserId             string    `json:"user_id" form:"user_id"`
	Role               *string   `json:"role" form:"role"`
	Comment            *string   `json:"comment" form:"comment"`
	CommentFiles       *[]string ``
	CommentFilesDelete *[]string `json:"comment_files_delete" form:"comment_files_delete"`
}

type UserNotificationResponse struct {
	ID           string                 `json:"id"`
	Notification *NotificationsResponse `json:"notification"`
	UserId       string                 `json:"user_id"`
	User         *UserResponse          `json:"user"`
	Role         *string                `json:"role"`
	ReadAt       *time.Time             `json:"read_at"`
	Comment      *string                `json:"comment"`
	CommentFiles *[]string              `json:"comment_files"`
}

func (r *UserNotificationResponse) FromModel(m *UserNotification) error {
	r.ID = m.ID
	if m.Notifications != nil {
		r.Notification = &NotificationsResponse{}
		r.Notification.FromModel(m.Notifications)
	}
	if m.User != nil {
		r.User = &UserResponse{}
		r.User.FromModel(m.User)
	}
	r.UserId = m.UserId
	r.Role = m.Role
	r.ReadAt = m.ReadAt
	r.Comment = m.Comment
	if m.CommentFiles != nil {
		r.CommentFiles = &[]string{}
		for _, f := range *m.CommentFiles {
			*r.CommentFiles = append(*r.CommentFiles, fileUrl(&f))
		}
	}
	return nil
}

func (r *UserNotificationRequest) ToModel(m *UserNotification) error {
	if r.ID != nil {
		m.ID = *r.ID
	}
	m.NotificationId = r.NotificationId
	m.UserId = r.UserId
	m.Role = r.Role
	m.Comment = r.Comment
	m.CommentFiles = r.CommentFiles
	return nil
}
