package models

import (
	"strings"
	"time"
)

type Classroom struct {
	ID            string                    `json:"id"`
	SchoolId      string                    `json:"school_id"`
	ShiftId       *string                   `json:"shift_id"`
	Name          *string                   `json:"name"`
	NameCanonical *string                   `json:"name_canonical"`
	Avatar        *string                   `json:"avatar" `
	Description   *string                   `json:"description"`
	Language      *string                   `json:"language"`
	Level         *string                   `json:"level"`
	TeacherId     *string                   `json:"teacher_id"`
	StudentId     *string                   `json:"student_id"`
	ParentId      *string                   `json:"parent_id"`
	PeriodId      *string                   `json:"period_id"`
	UpdatedAt     *time.Time                `json:"updated_at"`
	CreatedAt     *time.Time                `json:"created_at"`
	ArchivedAt    *time.Time                `json:"archived_at"`
	StudentsCount *int                      `json:"students_count"`
	Parent        *Classroom                `json:"parent"`
	Period        *Period                   `json:"period"`
	Teacher       *User                     `json:"teacher"`
	Student       *User                     `json:"student"`
	School        *School                   `json:"school"`
	Shift         *Shift                    `json:"shift"`
	Students      []*User                   `json:"students"`
	SubGroups     []ClassroomStudentsByType `json:"sub_groups"`
	Subjects      []*Subject                `json:"subjects"`
}

func (Classroom) RelationFields() []string {
	return []string{"Parent", "Period", "Teacher", "Student", "School", "Shift", "Students", "StudentsByType", "Subjects"}
}

// todo: make clearer
type UserClassroom struct {
	ClassroomId string     `json:"classroom_id"`
	UserId      string     `json:"user_id"`
	Type        *string    `json:"type"`
	TypeKey     *int       `json:"type_key"`
	TariffEndAt *time.Time `json:"tariff_end_at"`
	TariffType  *string    `json:"tariff_type"`
	Classroom   *Classroom `json:"classroom"`
	User        *User      `json:"user"`
}

type ClassroomValueResponse struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func (r *ClassroomResponse) ToValues() ClassroomValueResponse {
	value := ""
	if r.Name != nil {
		value += " " + *r.Name
	}
	return ClassroomValueResponse{
		Key:   r.ID,
		Value: strings.Trim(value, " "),
	}
}

type ClassroomFilterRequest struct {
	ID                *string   `form:"id"`
	Ids               *[]string `form:"ids"`
	ParentId          *string   `form:"parent_id"`
	ShiftId           *string   `form:"shift_id"`
	PeriodId          *string   `form:"period_id"`
	TeacherId         *string   `form:"teacher_id"`
	LoadRelation      bool      `form:"load_relation"`
	Name              *string   `form:"name"`
	ExamsCountBetween *string   `form:"exams_count_between[]"`
	Search            *string   `form:"search"`
	Sort              *string   `form:"sort"`
	SchoolId          *string   `form:"school_id"`
	SchoolIds         *[]string `form:"school_ids[]"`
	PaginationRequest
}

type ClassroomResponse struct {
	ID            string                    `json:"id"`
	Name          *string                   `json:"name"`
	Avatar        *string                   `json:"avatar"`
	Description   *string                   `json:"description"`
	Language      *string                   `json:"language"`
	Level         *string                   `json:"level"`
	UpdatedAt     *time.Time                `json:"updated_at"`
	CreatedAt     *time.Time                `json:"created_at"`
	ArchivedAt    *time.Time                `json:"archived_at"`
	StudentsCount *int                      `json:"students_count"`
	SchoolId      *string                   `json:"school_id"`
	TariffEndAt   *time.Time                `json:"tariff_end_at"`
	TariffType    *string                   `json:"tariff_type"`
	Teacher       *UserResponse             `json:"teacher"`
	Student       *UserResponse             `json:"student"`
	School        *SchoolResponse           `json:"school"`
	Shift         *ShiftResponse            `json:"shift"`
	Parent        *ClassroomResponse        `json:"parent"`
	Period        *PeriodResponse           `json:"period"`
	Students      []*UserResponse           `json:"students"`
	Subjects      []*SubjectResponse        `json:"subjects"`
	SubGroups     []ClassroomStudentsByType `json:"sub_groups"`
}

type ClassroomRequest struct {
	ID            *string                   `json:"id"`
	Name          *string                   `json:"name"`
	NameCanonical *string                   `json:"name_canonical"`
	Description   *string                   `json:"description"`
	Language      *string                   `json:"language"`
	Level         *string                   `json:"level"`
	Avatar        *string                   `json:"avatar"`
	SchoolId      *string                   `json:"school_id"`
	ShiftId       *string                   `json:"shift_id"`
	TeacherId     *string                   `json:"teacher_id"`
	StudentId     *string                   `json:"student_id"`
	ParentId      *string                   `json:"parent_id"`
	PeriodId      *string                   `json:"period_id"`
	IsArchive     *bool                     `json:"is_archive"`
	WithRelation  *bool                     `json:"with_relation"`
	StudentIds    *[]string                 `json:"student_ids"`
	Subjects      []*SubjectRequest         `json:"subjects"`
	SubGroups     []ClassroomStudentsByType `json:"sub_groups"`
}

func (r ClassroomRequest) ToModel(m *Classroom) {
	if r.ID != nil {
		m.ID = *r.ID
	}
	if r.Name != nil {
		m.Name = r.Name
	}
	if r.NameCanonical != nil {
		m.NameCanonical = r.NameCanonical
	}
	if r.Description != nil {
		m.Description = r.Description
	}
	if r.Language != nil {
		m.Language = r.Language
	}
	if r.Level != nil {
		m.Level = r.Level
	}
	if r.Avatar != nil {
		m.Avatar = r.Avatar
	}
	if r.SchoolId != nil {
		m.SchoolId = *r.SchoolId
	}
	if r.ShiftId != nil {
		m.ShiftId = r.ShiftId
	}
	if r.TeacherId != nil {
		m.TeacherId = r.TeacherId
	}
	if r.StudentId != nil {
		m.StudentId = r.StudentId
	}
	if r.ParentId != nil {
		m.ParentId = r.ParentId
	}
	if r.PeriodId != nil {
		m.PeriodId = r.PeriodId
	}
	if r.IsArchive != nil && *r.IsArchive && m.ArchivedAt == nil {
		m.ArchivedAt = new(time.Time)
		*m.ArchivedAt = time.Now()
	}
	if r.StudentIds != nil {
		m.Students = []*User{}
		for _, v := range *r.StudentIds {
			m.Students = append(m.Students, &User{ID: v})
		}
	}
	if r.Subjects != nil {
		m.Subjects = []*Subject{}
		for _, sr := range r.Subjects {
			subject := &Subject{}
			sr.ToModel(subject)
			m.Subjects = append(m.Subjects, subject)
		}
	}
	if r.SubGroups != nil {
		m.SubGroups = r.SubGroups
	}
}

type ClassroomStudentsByType struct {
	Type       *string  `json:"type" form:"type"`
	TypeKey    *int     `json:"type_key" form:"type_key"`
	StudentIds []string `json:"student_ids" form:"student_ids"`
}
type ClassroomStudentsByTypeResponse struct {
	Type     *string         `json:"type" form:"type"`
	TypeKey  *int            `json:"type_key" form:"type_key"`
	Students []*UserResponse `json:"students"`
}

type UserClassroomRequest struct {
	ClassroomId *string `json:"classroom_id" form:"classroom_id"`
	Type        *string `json:"type" form:"type"`
	TypeKey     *int    `json:"type_key" form:"type_key"`
	UserId      *string `json:"user_id" validate:"omitempty"`
}

type UserClassroomResponse struct {
	Type        *string            `json:"type"`
	TypeKey     *int               `json:"type_key"`
	TariffEndAt *time.Time         `json:"tariff_end_at"`
	TariffType  *string            `json:"tariff_type"`
	TariffName  *string            `json:"tariff_name"`
	Classroom   *ClassroomResponse `json:"classroom"`
}

func (u *UserClassroomResponse) FromModel(m *UserClassroom) {
	u.Type = m.Type
	u.TypeKey = m.TypeKey
	u.TariffEndAt = m.TariffEndAt
	u.TariffType = m.TariffType
	u.TariffName = GetTariffName(m.TariffType)
	if u.TariffEndAt != nil && u.TariffEndAt.Before(time.Now()) {
		u.TariffEndAt = nil
		u.TariffType = nil
		u.TariffName = nil
	}
	u.Classroom = &ClassroomResponse{}
	u.Classroom.FromModel(m.Classroom)
}

func (r *ClassroomResponse) FromModel(m *Classroom) {
	r.ID = m.ID
	r.Name = m.Name
	r.Avatar = m.Avatar
	r.Description = m.Description
	r.Language = m.Language
	r.Level = m.Level
	r.StudentsCount = m.StudentsCount
	r.SchoolId = &m.SchoolId
	if m.School != nil {
		r.School = &SchoolResponse{}
		r.School.FromModel(m.School)
	}
	if m.Shift != nil {
		r.Shift = &ShiftResponse{}
		r.Shift.FromModel(m.Shift)
	}
	if m.Teacher != nil {
		r.Teacher = &UserResponse{}
		r.Teacher.FromModel(m.Teacher)
	}
	if m.Student != nil {
		r.Student = &UserResponse{}
		r.Student.FromModel(m.Student)
	}
	if m.Parent != nil {
		r.Parent = &ClassroomResponse{}
		r.Parent.FromModel(m.Parent)
	}
	if m.Period != nil {
		r.Period = &PeriodResponse{}
		r.Period.FromModel(m.Period)
	}
	if m.Students != nil {
		r.Students = []*UserResponse{}
		for _, v := range m.Students {
			ur := UserResponse{}
			ur.FromModel(v)
			r.Students = append(r.Students, &ur)
		}
	}
	if m.SubGroups != nil {
		r.SubGroups = m.SubGroups
	}
	if m.Subjects != nil {
		r.Subjects = []*SubjectResponse{}
		for _, v := range m.Subjects {
			ur := SubjectResponse{}
			ur.FromModel(v)
			r.Subjects = append(r.Subjects, &ur)
		}
	}
}

type GetClassroomIdByNameQueryDto struct {
	SchoolId uint
	Name     string
}

var ClassroomGroupKeys = []string{
	"lang4",
	"lang5",
	"labor1",
	"lang1",
	"lang2",
	"lang3",
	"other1",
	"informatics",
}
