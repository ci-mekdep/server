package models

import (
	"encoding/json"
	"time"
)

// first key = 0 = monday
type ShiftValue [][][]string

type Shift struct {
	Id              string     `json:"id"`
	Name            *string    `json:"name"`
	SchoolId        string     `json:"school_id"`
	Value           *string    `json:"value"`
	UpdatedBy       *string    `json:"updated_by"`
	CreatedAt       *time.Time `json:"created_at"`
	UpdatedAt       *time.Time `json:"updated_at"`
	TimetablesCount *int       `json:"timetables_count"`
	ClassroomsCount *int       `json:"classrooms_count"`
	UpdatedByUser   *User      `json:"updated_by_user"`
	School          *School    `json:"school"`
}

func (Shift) RelationFields() []string {
	return []string{"UpdatedByUser", "School"}
}

type ShiftFilterRequest struct {
	ID          *string   `form:"id"`
	IDs         *[]string `form:"ids[]"`
	TimetableId *string   `form:"timetable_id"`
	ClassroomId *string   `form:"classroom_id"`
	SchoolId    *string   `form:"school_id"`
	SchoolIds   *[]string `form:"school_ids[]"`
	Search      *string   `form:"search"`
	PaginationRequest
}

type ShiftRequest struct {
	Id        *string    `json:"id"`
	Name      *string    `json:"name"`
	SchoolId  string     `json:"school_id"`
	SchoolIds *[]string  `json:"school_ids[]"`
	Value     ShiftValue `json:"value"`
	UpdatedBy string
}

type ShiftResponse struct {
	Id              string          `json:"id"`
	Name            string          `json:"name"`
	School          *SchoolResponse `json:"school"`
	Value           ShiftValue      `json:"value"`
	UpdatedBy       *string         `json:"updated_by"`
	CreatedAt       *time.Time      `json:"created_at"`
	UpdatedAt       *time.Time      `json:"updated_at"`
	TimetablesCount *int            `json:"timetables_count"  binding:"required"`
	ClassroomsCount *int            `json:"classrooms_count"`
}

func (r *ShiftResponse) FromModel(m *Shift) error {
	r.Id = m.Id
	r.Name = *m.Name
	if m.School != nil {
		r.School = &SchoolResponse{}
		r.School.FromModel(m.School)
	}
	err := json.Unmarshal([]byte(*m.Value), &r.Value)
	if err != nil {
		return err
	}
	r.UpdatedBy = m.UpdatedBy
	r.CreatedAt = m.CreatedAt
	r.UpdatedAt = m.UpdatedAt
	r.TimetablesCount = m.TimetablesCount
	r.ClassroomsCount = m.ClassroomsCount
	return nil
}

func (r *ShiftRequest) ToModel(m *Shift) error {
	if r.Id == nil {
		r.Id = new(string)
	}
	m.Id = *r.Id
	m.Name = r.Name
	m.SchoolId = r.SchoolId
	tmp, err := json.Marshal(r.Value)
	if err != nil {
		return err
	}
	tmp2 := string(tmp)
	m.Value = &tmp2

	return nil
}
