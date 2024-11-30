package models

import (
	"encoding/json"
	"time"
)

type Timetable struct {
	ID            string     `json:"id"`
	SchoolId      string     `json:"school_id"`
	ClassroomId   string     `json:"classroom_id"`
	ShiftId       *string    `json:"shift_id"`
	PeriodId      *string    `json:"period_id"`
	Value         *string    `json:"value"`
	UpdatedAt     *time.Time `json:"updated_at"`
	CreatedAt     *time.Time `json:"created_at"`
	UpdatedBy     *string    `json:"updated_by"`
	School        *School    `json:"school"`
	Classroom     *Classroom `json:"classroom"`
	Shift         *Shift     `json:"shift"`
	UpdatedByUser *User      `json:"updated_by_user"`
}

func (Timetable) RelationFields() []string {
	return []string{"School", "Classroom", "Shift", "UpdatedByUser"}
}

func (r *TimetableResponse) FromModel(m *Timetable) error {
	r.ID = m.ID
	r.CreatedAt = m.CreatedAt
	r.UpdatedAt = m.UpdatedAt
	if m.Classroom != nil {
		r.Classroom = &ClassroomResponse{}
		r.Classroom.FromModel(m.Classroom)
	}
	if m.School != nil {
		r.School = &SchoolResponse{}
		r.School.FromModel(m.School)
	}
	if m.Shift != nil {
		r.Shift = &ShiftResponse{}
		r.Shift.FromModel(m.Shift)
	}
	if m.UpdatedByUser != nil {
		r.UpdatedByUser = &UserResponse{}
		r.UpdatedByUser.FromModel(m.UpdatedByUser)
	}
	if m.Value != nil {
		err := json.Unmarshal([]byte(*m.Value), &r.Value)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *TimetableRequest) ToModel(m *Timetable) error {
	if r.ID == nil {
		r.ID = new(string)
	}
	m.ID = *r.ID
	m.SchoolId = r.SchoolId
	m.ClassroomId = r.ClassroomId
	m.ShiftId = r.ShiftId
	tmp, err := json.Marshal(r.Value)
	if err != nil {
		return err
	}
	tmp2 := string(tmp)
	m.Value = &tmp2
	m.UpdatedBy = &r.UpdatedBy
	return nil
}

type TimetableValueHour struct {
	Hour      int  `json:"hour" validate:"required"`
	SubjectId uint `json:"subject_id" validate:"required"`
}

type TimetableValueDay struct {
	Day   int                   `json:"day"`
	Hours []*TimetableValueHour `json:"hours"`
}

type TimetableFilterRequest struct {
	ID           *string   `form:"id"`
	Ids          *[]string `form:"ids"`
	ClassroomIds *[]string `form:"classroom_ids[]"`
	ShiftIds     *[]string `form:"shift_ids[]"`
	Search       *string   `form:"search"`
	Sort         *string   `form:"sort"`
	SchoolId     *string   `form:"school_id"`
	SchoolIds    *[]string `form:"school_ids[]"`
	PaginationRequest
}

type TimetableValue [][]string

type TimetableRequest struct {
	ID          *string        `json:"id"`
	Value       TimetableValue `json:"value"`
	SchoolId    string         `json:"school_id"`
	ClassroomId string         `json:"classroom_id"`
	ShiftId     *string        `json:"shift_id"`
	IsThisWeek  bool           `json:"this_week"`
	SchoolIds   *[]string      `json:"school_ids[]"`
	UpdatedBy   string
}

type TimetableResponse struct {
	ID            string             `json:"id"`
	Value         TimetableValue     `json:"value"`
	UpdatedAt     *time.Time         `json:"updated_at"`
	CreatedAt     *time.Time         `json:"created_at"`
	School        *SchoolResponse    `json:"school"`
	Classroom     *ClassroomResponse `json:"classroom"`
	Shift         *ShiftResponse     `json:"shift"`
	UpdatedByUser *UserResponse      `json:"updated_by"`
}
