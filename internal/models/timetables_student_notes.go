package models

import "time"

type StudentNote struct {
	ID        string     `json:"id"`
	SchoolId  string     `json:"school_id"`
	SubjectId *string    `json:"subject_id"`
	StudentId string     `json:"student_id"`
	TeacherId string     `json:"teacher_id"`
	Note      string     `json:"note"`
	UpdatedAt *time.Time `json:"updated_at"`
	CreatedAt *time.Time `json:"created_at"`
	School    *School    `json:"school"`
	Subject   *Subject   `json:"subject"`
	Student   *User      `json:"student"`
	Teacher   *User      `json:"teacher"`
}

func (StudentNote) RelationFields() []string {
	return []string{"School", "Subject", "Student", "Teacher"}
}

type StudentNoteFilterRequest struct {
	SchoolId   *string   `form:"school_id"`
	SubjectId  *string   `form:"subject_id"`
	StudentId  *string   `form:"student_id"`
	StudentIds *[]string `form:"student_ids"`
	TeacherId  *string   `form:"teacher_id"`
}

type StudentNoteRequest struct {
	SchoolId  string  `json:"school_id"`
	SubjectId *string `json:"subject_id"`
	StudentId string  `json:"student_id" validate:"required"`
	TeacherId string  `json:"teacher_id"` // force filled in api
	Note      string  `json:"note"`
}

func (m *StudentNote) FromRequest(r *StudentNoteRequest) {
	m.SchoolId = r.SchoolId
	m.SubjectId = r.SubjectId
	m.StudentId = r.StudentId
	m.TeacherId = r.TeacherId
	m.Note = r.Note
}

type StudentNoteResponse struct {
	StudentId string           `json:"student_id"`
	Note      string           `json:"note"`
	UpdatedAt *time.Time       `json:"updated_at"`
	CreatedAt *time.Time       `json:"created_at"`
	Subject   *SubjectResponse `json:"subject"`
}

func (m *StudentNoteResponse) FromModel(r *StudentNote) {
	m.StudentId = r.StudentId
	m.Note = r.Note
	m.UpdatedAt = r.UpdatedAt
	m.CreatedAt = r.CreatedAt
	if r.Subject != nil {
		m.Subject.FromModel(r.Subject)
	}
}
