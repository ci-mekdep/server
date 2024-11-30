package models

import (
	"errors"
	"log"
	"strconv"
	"time"

	"github.com/mekdep/server/config"
)

type GradeDetail struct {
	Lesson *Lesson `json:"lesson"`
	Grade  *Grade  `json:"grade"`
	Absent *Absent `json:"absent"`
}

type RatingCenterSchool struct {
	RatingByGroup []RatingByGroup `json:"rating_by_group"`
	RatingByLevel []RatingByLevel `json:"rating_by_level"`
}

type RatingByGroup struct {
	User   User `json:"user"`
	Points int  `json:"points"`
	Index  int  `json:"index"`
}

type RatingByLevel struct {
	User   User `json:"user"`
	Points int  `json:"points"`
	Index  int  `json:"index"`
}

type Grade struct {
	ID               string     `json:"id"`
	LessonId         string     `json:"lesson_id"`
	StudentId        string     `json:"student_id"`
	Value            *int       `json:"value"`
	Values           *[]int     `json:"values"`
	Reason           *string    `json:"reason"`
	Comment          *string    `json:"comment"`
	ParentReviewedAt *time.Time `json:"parent_reviewed_at"`
	CreatedAt        *time.Time `json:"created_at"`
	UpdatedAt        *time.Time `json:"updated_at"`
	CreatedBy        *string    `json:"created_by"`
	UpdatedBy        *string    `json:"updated_by"`
	DeletedAt        *time.Time `json:"deleted_at"`
	Lesson           *Lesson    `json:"lesson"`
	Student          *User      `json:"student"`
	CreatedByUser    *User      `json:"created_by_user"`
	UpdatedByUser    *User      `json:"updated_by_user"`
}

func (Grade) RelationFields() []string {
	return []string{"Lesson", "Student", "CreatedByUser", "UpdatedByUser"}
}
func (g Grade) GetValue() *string {
	var value *string
	if g.Value != nil {
		value = new(string)
		*value = strconv.Itoa(*g.Value)
	}
	return value
}
func (m *GradeRequest) IsValueDelete() bool {
	return m.Value == nil && m.Values == nil || m.Value != nil && *m.Value == 0
}
func (m *Grade) IsCreateExpired(phone string) bool {
	now := time.Now()
	expirationDate := -14
	if now.Before(time.Date(2024, 10, 31, 23, 59, 0, 0, config.RequestLocation)) {
		expirationDate = -365
	}
	for _, devPhone := range config.Conf.DevPhones {
		if devPhone == phone {
			expirationDate = -365
			break
		}
	}
	min := now.AddDate(0, 0, expirationDate)
	return m.Lesson.Date.Before(min) || m.Lesson.Date.After(now)
}

func (m *Grade) IsUpdateExpired() bool {
	// TODO: .Add(time.Hour * 5) is timezone problem, pgx uses UTC+0, ours UTC+5
	now := time.Now()
	if m.CreatedAt.Location().String() != now.Location().String() {
		log.Println("DEBUG: " + errors.New("database timezone is incorrect: "+m.CreatedAt.Location().String()+"; machine timezone: "+now.Location().String()).Error())
		*m.CreatedAt = m.CreatedAt.Local()
	}
	return m.CreatedAt.IsZero() || m.CreatedAt.Add(time.Minute*60).Before(now)
}

type GradeResponse struct {
	ID               string          `json:"id"`
	LessonId         string          `json:"lesson_id"`
	StudentId        string          `json:"student_id"`
	Value            *int            `json:"value"`
	Values           *[]int          `json:"values"`
	Reason           *string         `json:"reason"`
	Comment          *string         `json:"comment"`
	ParentReviewedAt *time.Time      `json:"parent_reviewed_at"`
	UpdatedAt        *time.Time      `json:"updated_at"`
	CreatedAt        *time.Time      `json:"created_at"`
	DeletedAt        *time.Time      `json:"deleted_at"`
	UpdatedBy        *string         `json:"updated_by"`
	CreatedBy        *string         `json:"created_by"`
	Lesson           *LessonResponse `json:"lesson"`
	Student          *UserResponse   `json:"student"`
	UpdatedByUser    *UserResponse   `json:"updated_by_user"`
	CreatedByUser    *UserResponse   `json:"created_by_user"`
}

func (r *GradeResponse) FromModel(m *Grade) {
	if m.Value != nil {
		r.Value = m.Value
	}
	if m.Values != nil {
		r.Values = m.Values
	}
	if m.Reason != nil {
		r.Reason = m.Reason
	}
	if m.Comment != nil {
		r.Comment = m.Comment
	}
	r.ID = m.ID
	r.LessonId = m.LessonId
	r.StudentId = m.StudentId
	r.UpdatedAt = m.UpdatedAt
	r.CreatedAt = m.CreatedAt
	r.ParentReviewedAt = m.ParentReviewedAt
	if m.Lesson != nil {
		t := LessonResponse{}
		t.FromModel(m.Lesson)
		r.Lesson = &t
	}
	if m.Student != nil {
		t := UserResponse{}
		t.FromModel(m.Student)
		r.Student = &t
	}
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

type GradeRequest struct {
	StudentId  string   `json:"student_id"`
	StudentIds []string `json:"student_ids"`
	Value      *int     `json:"value"`
	Values     *[]int   `json:"values"`
	ExamId     *string  `json:"exam_id"`
	Reason     *string  `json:"reason"`
	Comment    *string  `json:"comment"`
	UpdatedBy  string
}

func (m *Grade) FromRequest(r *GradeRequest) {
	m.StudentId = r.StudentId
	m.Value = r.Value
	m.Values = r.Values
	m.Reason = r.Reason
	m.Comment = r.Comment
	m.UpdatedBy = &r.UpdatedBy
	m.CreatedBy = &r.UpdatedBy
}

func (m *Grade) ValueString() string {
	if m.Value != nil {
		return strconv.Itoa(*m.Value)
	}
	if m.Values != nil {
		return strconv.Itoa((*m.Values)[0]) + "/" + strconv.Itoa((*m.Values)[1])

	}
	return ""
}

type GradeFilterRequest struct {
	ID         *string   `json:"id"`
	IDs        *[]string `form:"ids[]" validate:"omitempty"`
	LessonId   *string   `json:"lesson_id"`
	LessonIds  *[]string `json:"lesson_ids[]"`
	StudentId  *string   `json:"student_id"`
	StudentIds *[]string `json:"student_ids[]"`
	SchoolId   *string   `form:"school_id" validate:"omitempty"`
	SchoolIds  *[]string `form:"school_ids[]"`
	PaginationRequest
}

// TODO: constant
var DefaultGradeReasons = [][]string{
	{"yes", "Sebäpli", "Уважительная причина", "Excused"},
	{"no", "Sebäpsiz", "Не уважительная причина", "Not excused"},
	{"ill", "Ýarawsyz", "Болеет", "Ill"},
}
