package models

import (
	"time"
)

type SubjectExam struct {
	ID                string     `json:"id"`
	SubjectId         string     `json:"subject_id"`
	SchoolId          string     `json:"school_id"`
	ClassroomId       string     `json:"classroom_id"`
	TeacherId         string     `json:"teacher_id"`
	HeadTeacherId     *string    `json:"head_teacher_id"`
	MemberTeacherIds  *[]string  `json:"member_teacher_ids"`
	RoomNumber        *string    `json:"room_number"`
	StartTime         *time.Time `json:"start_time"`
	TimeLengthMin     *int       `json:"time_length_min"`
	ExamWeightPercent *uint      `json:"exam_weight_percent"`
	Name              *string    `json:"name"`
	IsRequired        *bool      `json:"is_required"`
	CreatedAt         *time.Time `json:"created_at"`
	UpdatedAt         *time.Time `json:"updated_at"`
	Subject           *Subject   `json:"subject"`
	School            *School    `json:"school"`
	Classroom         *Classroom `json:"classroom"`
	Teacher           *User      `json:"teacher"`
	HeadTeacher       *User      `json:"head_teacher"`
	MemberTeachers    []*User    `json:"member_teachers"`
}

func (SubjectExam) RelationFields() []string {
	return []string{"Subject", "School", "Classroom", "Teacher", "HeadTeacher", "MemberTeachers"}
}

type SubjectExamFilterRequest struct {
	ID             *string   `form:"id"`
	Ids            *[]string `form:"ids"`
	SchoolId       *string   `form:"school_id"`
	SchoolIds      []string  `form:"school_ids[]"`
	ClassroomId    *string   `form:"classroom_id"`
	SubjectId      *string   `form:"subject_id"`
	SubjectIds     []string  `json:"subject_ids"`
	TeacherId      *string   `form:"teacher_id"`
	ClassroomNames []string  `form:"classroom_names"`
	IsGraduate     *bool     `form:"is_graduate"`
	Search         *string   `form:"search"`
	Sort           *string   `form:"sort"`
	PaginationRequest
}

type SubjectExamRequest struct {
	ID                *string    `json:"id"`
	RoomNumber        string     `json:"room_number"`
	TimeLengthMin     *int       `json:"time_length_min"`
	TeacherId         *string    `json:"teacher_id"`
	SubjectId         *string    `json:"subject_id"`
	SchoolId          *string    `json:"school_id"`
	ClassroomId       *string    `json:"classroom_id"`
	HeadTeacherId     *string    `json:"head_teacher_id"`
	MemberTeacherIds  *[]string  `json:"member_teacher_ids"`
	StartTime         *time.Time `json:"start_time" time_format:"2006-01-02 15:04:05"`
	ExamWeightPercent *uint      `json:"exam_weight_percent"`
	Name              *string    `json:"name"`
	IsRequired        *bool      `json:"is_required"`
}

func (r *SubjectExamRequest) ToModel(m *SubjectExam) {
	if r.ID == nil {
		r.ID = new(string)
	}
	m.ID = *r.ID
	m.RoomNumber = &r.RoomNumber
	m.TimeLengthMin = r.TimeLengthMin
	if r.TeacherId != nil {
		m.TeacherId = *r.TeacherId
	}
	if r.SubjectId != nil {
		m.SubjectId = *r.SubjectId
	}
	if r.SchoolId != nil {
		m.SchoolId = *r.SchoolId
	}
	if r.ClassroomId != nil {
		m.ClassroomId = *r.ClassroomId
	}
	if r.MemberTeacherIds != nil {
		m.MemberTeachers = []*User{}
		for _, i := range *r.MemberTeacherIds {
			r := &User{ID: i}
			m.MemberTeachers = append(m.MemberTeachers, r)
		}
	}
	m.HeadTeacherId = r.HeadTeacherId
	m.MemberTeacherIds = r.MemberTeacherIds
	m.StartTime = r.StartTime
	m.ExamWeightPercent = r.ExamWeightPercent
	m.Name = r.Name
	m.IsRequired = r.IsRequired
}

type SubjectExamResponse struct {
	ID                string             `json:"id"`
	RoomNumber        *string            `json:"room_number"`
	TimeLengthMin     *int               `json:"time_length_min"`
	StartTime         *time.Time         `json:"start_time"`
	ExamWeightPercent *uint              `json:"exam_weight_percent"`
	Name              *string            `json:"name"`
	IsRequired        *bool              `json:"is_required"`
	CreatedAt         *time.Time         `json:"created_at"`
	UpdatedAt         *time.Time         `json:"updated_at"`
	Subject           *SubjectResponse   `json:"subject"`
	School            *SchoolResponse    `json:"school"`
	Classroom         *ClassroomResponse `json:"classroom"`
	Teacher           *UserResponse      `json:"teacher"`
	HeadTeacher       *UserResponse      `json:"head_teacher"`
	MemberTeachers    []*UserResponse    `json:"member_teachers"`
}

func (r *SubjectExamResponse) FromModel(m *SubjectExam) {
	r.ID = m.ID
	r.RoomNumber = m.RoomNumber
	r.TimeLengthMin = m.TimeLengthMin
	r.StartTime = m.StartTime
	r.ExamWeightPercent = m.ExamWeightPercent
	r.Name = m.Name
	r.IsRequired = m.IsRequired
	if m.Subject != nil {
		r.Subject = &SubjectResponse{}
		r.Subject.FromModel(m.Subject)
	}
	if m.School != nil {
		r.School = &SchoolResponse{}
		r.School.FromModel(m.School)
	}
	if m.Teacher != nil {
		r.Teacher = &UserResponse{}
		r.Teacher.FromModel(m.Teacher)
	}
	if m.HeadTeacher != nil {
		r.HeadTeacher = &UserResponse{}
		r.HeadTeacher.FromModel(m.HeadTeacher)
	}
	if m.Classroom != nil {
		r.Classroom = &ClassroomResponse{}
		r.Classroom.FromModel(m.Classroom)
	}
	r.MemberTeachers = []*UserResponse{}
	for _, i := range m.MemberTeachers {
		u := &UserResponse{}
		u.FromModel(i)
		r.MemberTeachers = append(r.MemberTeachers, u)
	}
	r.CreatedAt = m.CreatedAt
	r.UpdatedAt = m.UpdatedAt
}
