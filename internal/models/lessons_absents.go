package models

import (
	"errors"
	"log"
	"time"

	"github.com/mekdep/server/config"
)

type Absent struct {
	ID               string     `json:"id"`
	LessonId         string     `json:"lesson_id"`
	StudentId        string     `json:"student_id"`
	Reason           *string    `json:"reason"`
	Comment          *string    `json:"comment"`
	ParentReviewedAt *time.Time `json:"parent_reviewed_at"`
	UpdatedAt        *time.Time `json:"updated_at"`
	CreatedAt        *time.Time `json:"created_at"`
	UpdatedBy        *string    `json:"updated_by"`
	CreatedBy        *string    `json:"created_by"`
	DeletedAt        *time.Time `json:"deleted_at"`
	Lesson           *Lesson    `json:"lesson"`
	Student          *User      `json:"student"`
	UpdatedByUser    *User      `json:"updated_by_user"`
	CreatedByUser    *User      `json:"created_by_user"`
}

func (Absent) RelationFields() []string {
	return []string{"Lesson", "Student", "UpdatedByUser", "CreatedByUser"}
}

func (m *AbsentRequest) IsValueDelete() bool {
	return m.Reason == nil
}

func (m *Absent) IsUpdateExpired() bool {
	// TODO: .Add(time.Hour * 5) is timezone problem, pgx uses UTC+0, ours UTC+5
	now := time.Now()
	if m.CreatedAt.Location().String() != now.Location().String() {
		log.Println("DEBUG: " + errors.New("database timezone is incorrect: "+m.CreatedAt.Location().String()+"; machine timezone: "+now.Location().String()).Error())
		*m.CreatedAt = m.CreatedAt.Local()
	}
	return m.CreatedAt.IsZero() || m.CreatedAt.Add(time.Minute*60).Before(now)
}

func (m *Absent) IsCreateExpired() bool {
	// return false
	if m.Lesson == nil {
		return false
	}
	now := time.Now()
	expirationDate := -14
	if now.Before(time.Date(2024, 10, 31, 23, 59, 0, 0, config.RequestLocation)) {
		expirationDate = -365
	}
	// for _, devPhone := range config.Conf.DevPhones {
	// 	if devPhone == phone {
	// 		expirationDate = -365
	// 		break
	// 	}
	// }
	min := now.AddDate(0, 0, expirationDate) // TODO: make 14
	return m.Lesson.Date.Before(min) || m.Lesson.Date.After(now)
}

type AbsentRequest struct {
	StudentId  string   `json:"student_id"`
	StudentIds []string `json:"student_ids"`
	Reason     *string  `json:"reason"`
	Comment    *string  `json:"comment"`
	UpdatedBy  string
}

func (m *Absent) FromRequest(r *AbsentRequest) {
	m.StudentId = r.StudentId
	m.Reason = r.Reason
	m.Comment = r.Comment
	m.UpdatedBy = &r.UpdatedBy
	m.CreatedBy = &r.UpdatedBy
}

type AbsentResponse struct {
	ID               string          `json:"id"`
	LessonId         string          `json:"lesson_id"`
	StudentId        string          `json:"student_id"`
	Reason           *string         `json:"reason"`
	Comment          *string         `json:"comment"`
	ParentReviewedAt *time.Time      `json:"parent_reviewed_at"`
	UpdatedAt        *time.Time      `json:"updated_at"`
	CreatedAt        *time.Time      `json:"created_at"`
	DeletedAt        *time.Time      `json:"deleted_at"`
	Lesson           *LessonResponse `json:"lesson"`
	Student          *UserResponse   `json:"student"`
	UpdatedByUser    *UserResponse   `json:"updated_by_user"`
	CreatedByUser    *UserResponse   `json:"created_by_user"`
}

func (r *AbsentResponse) FromModel(m *Absent) {
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

type AbsentFilterRequest struct {
	ID         *string   `json:"id"`
	IDs        *[]string `form:"ids[]"`
	LessonId   *string   `json:"lesson_id"`
	LessonIds  *[]string `json:"lesson_ids[]"`
	StudentId  *string   `json:"student_id"`
	StudentIds *[]string `json:"student_ids[]"`
	SchoolId   *string   `form:"school_id"`
	SchoolIds  *[]string `form:"school_ids[]"`
	PaginationRequest
}
