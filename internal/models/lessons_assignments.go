package models

import "time"

type Assignment struct {
	ID            string     `json:"id"`
	LessonId      string     `json:"lesson_id"`
	Title         string     `json:"title"`
	Files         *[]string  `json:"files"`
	UpdatedAt     *time.Time `json:"updated_at"`
	CreatedAt     *time.Time `json:"created_at"`
	UpdatedBy     *string    `json:"updated_by"`
	CreatedBy     *string    `json:"created_by"`
	Lesson        *Lesson    `json:"lesson"`
	UpdatedByUser *User      `json:"updated_by_user"`
	CreatedByUser *User      `json:"created_by_user"`
}

func (Assignment) RelationFields() []string {
	return []string{"Lesson", "UpdatedByUser", "CreatedByUser"}
}

type AssignmentResponse struct {
	Title         string        `json:"title"`
	Files         *[]string     `json:"files"`
	Content       *string       `json:"content"`
	UpdatedByUser *UserResponse `json:"updated_by_user"`
	CreatedByUser *UserResponse `json:"created_by_user"`
}

type AssignmentRequest struct {
	LessonID    *string   `json:"lesson_id" form:"lesson_id"`
	Title       *string   `json:"title" form:"title"`
	Content     *string   `json:"content" form:"content"`
	FilesDelete *[]string `json:"files_delete" validate:"omitempty"`
	Files       *[]string `json:"files" validate:"omitempty"`
	UpdatedBy   string
}

type AssignmentFilterRequest struct {
	ID        *string   `json:"id"`
	LessonID  *string   `json:"lesson_id"`
	LessonIDs *[]string `json:"lesson_ids"`
	Search    *string   `json:"search"`
	Sort      *string   `json:"sort"`
	PaginationRequest
}

func (m *Assignment) FromRequest(r *AssignmentRequest) {
	if r.LessonID != nil {
		m.LessonId = *r.LessonID
	}
	if r.Title != nil {
		m.Title = *r.Title
	}
	// m.Content = r.Content
	m.Files = r.Files
	m.UpdatedBy = &r.UpdatedBy
	m.CreatedBy = &r.UpdatedBy
}

func (r *AssignmentResponse) FromModel(m *Assignment) {
	if m.Files != nil {
		r.Files = &[]string{}

		for _, f := range *m.Files {
			*r.Files = append(*r.Files, fileUrl(&f))
		}
	}
	r.Title = m.Title
	if m.UpdatedByUser != nil {
		t := UserResponse{}
		t.FromModel(m.UpdatedByUser)
		r.UpdatedByUser = &t
	}
	if m.CreatedByUser != nil {
		t := UserResponse{}
		t.FromModel(m.CreatedByUser)
		r.CreatedByUser = &t
	}
}
