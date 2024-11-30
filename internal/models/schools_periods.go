package models

import (
	"errors"
	"path"
	"time"

	"github.com/mekdep/server/config"
)

type DataCounts struct {
	StudentsCount   *int `json:"students_count"`
	TeachersCount   *int `json:"teachers_count"`
	ClassroomsCount *int `json:"classrooms_count"`
	SubjectHours    *int `json:"subject_hours"`
	GradesCount     *int `json:"grades_count"`
	AbsentsCount    *int `json:"absents_count"`
	TimetablesCount *int `json:"timetables_count"`
}

type Period struct {
	ID              string      `json:"id"`
	Title           string      `json:"title"`
	Value           [][]string  `json:"value"`
	DataCounts      *DataCounts `json:"data_counts"`
	UpdatedAt       *time.Time  `json:"updated_at"`
	CreatedAt       *time.Time  `json:"created_at"`
	SchoolId        *string     `json:"school_id"`
	TimetablesCount *int        `json:"timetables_count"`
	ClassroomsCount *int        `json:"classrooms_count"`
	School          *School     `json:"school"`
}

func (Period) RelationFields() []string {
	return []string{"School"}
}

func (p Period) GetPeriodKeys() []int {
	pk := []int{}
	for k := range p.Value {
		pk = append(pk, k+1)
	}
	return pk
}

func (p Period) Dates() (time.Time, time.Time, error) {
	start := time.Time{}
	end := time.Time{}
	if len(p.Value) < 1 {
		return start, end, errors.New("period does not have dates")
	}
	stt := p.Value[0]
	if len(stt) < 2 {
		return start, end, errors.New("period start does not have dates")
	}
	ett := p.Value[len(p.Value)-1]
	if len(ett) < 2 {
		return start, end, errors.New("period end does not have dates")
	}
	var err error
	start, err = time.Parse(time.DateOnly, stt[0])
	if err != nil {
		return start, end, err
	}
	end, err = time.Parse(time.DateOnly, ett[1])
	if err != nil {
		return start, end, err
	}
	return start, end, nil
}

func (p Period) DatesByKey(key int) (time.Time, time.Time, error) {
	start := time.Time{}
	end := time.Time{}
	if len(p.Value) < key {
		return start, end, errors.New("period does not have dates")
	}
	tt := p.Value[key-1]
	if len(tt) < 2 {
		return start, end, errors.New("period end does not have dates")
	}
	var err error
	start, err = time.Parse(time.DateOnly, tt[0])
	if err != nil {
		return start, end, err
	}
	end, err = time.Parse(time.DateOnly, tt[1])
	if err != nil {
		return start, end, err
	}
	return start, end, nil
}

func (m Period) GetKey(t time.Time, isExact bool) (index int, err error) {
	var prevT1, prevT2 time.Time
	m.Value = append(m.Value, m.Value[0])
	for k, i := range m.Value {
		if len(i) < 2 || (i[0] == "" || i[1] == "") {
			continue
		}
		t1, err := time.Parse(time.DateOnly, i[0])
		if err != nil {
			return 0, err
		}
		t2, err := time.Parse(time.DateOnly, i[1])
		if err != nil {
			return 0, err
		}
		if t1.Compare(prevT1) <= 0 {
			t1 = prevT2
		}
		if !prevT1.IsZero() && !prevT2.IsZero() {
			if isExact {
				if prevT1.Compare(t) <= 0 && prevT2.Compare(t) >= 0 {
					// add +1 to key, because by prevT1
					return k, nil
				}
			} else {
				if prevT1.Compare(t) <= 0 && t1.Compare(t) >= 0 {
					// add +1 to key, because by prevT1
					return k, nil
				}
			}
		}
		prevT1, prevT2 = t1, t2

	}
	return index, nil
}

func (p Period) IsArchived() (bool, error) {
	_, end, err := p.Dates()
	if err != nil {
		return false, err
	}
	return end.Before(time.Now()), nil
}

type PeriodFilterRequest struct {
	ID        *string   `form:"id"`
	Ids       *[]string `form:"ids"`
	SchoolId  *string   `form:"school_id"`
	SchoolIds *[]string `form:"school_ids[]"`
	Search    *string   `form:"search"`
	Sort      *string   `form:"sort"`
	PaginationRequest
}

type PeriodResponse struct {
	ID              string          `json:"id"`
	Title           string          `json:"title"`
	Value           [][]string      `json:"value"`
	CurrentIndex    *int            `json:"current_index"`
	IsArchived      *bool           `json:"is_archived"`
	ArchiveLink     *string         `json:"archive_link"`
	DataCounts      *DataCounts     `json:"data_counts"`
	UpdatedAt       *time.Time      `json:"updated_at"`
	CreatedAt       *time.Time      `json:"created_at"`
	TimetablesCount *int            `json:"timetables_count"`
	ClassroomsCount *int            `json:"classrooms_count"`
	School          *SchoolResponse `json:"school"`
}

func (r *PeriodResponse) FromModel(m *Period) {
	r.ID = m.ID
	r.Title = m.Title
	r.Value = m.Value
	r.DataCounts = m.DataCounts
	if r.DataCounts == nil {
		r.DataCounts = &DataCounts{}
	}
	r.DataCounts.ClassroomsCount = m.ClassroomsCount
	r.DataCounts.TimetablesCount = m.TimetablesCount
	lastDateStr := r.Value[len(r.Value)-1][1]
	lastDate, _ := time.Parse(time.DateOnly, lastDateStr)
	if lastDate.Before(time.Now()) {
		r.IsArchived = new(bool)
		*r.IsArchived = true
		r.ArchiveLink = new(string)
		*r.ArchiveLink = path.Join(config.Conf.ArchiveLink, r.ID)
	}
	r.CreatedAt = m.CreatedAt
	r.UpdatedAt = m.UpdatedAt
	r.TimetablesCount = m.TimetablesCount
	r.ClassroomsCount = m.ClassroomsCount
	if m.School != nil {
		r.School = &SchoolResponse{}
		r.School.FromModel(m.School)
	}
}

type PeriodRequest struct {
	ID        *string     `json:"id"`
	SchoolId  *string     `json:"school_id"`
	Title     *string     `json:"title"`
	Value     *[][]string `json:"value"`
	SchoolIds *[]uint     `json:"school_ids"`
}

func (r *PeriodRequest) ToModel(m *Period) {
	if r.ID != nil {
		m.ID = *r.ID
	}
	if r.Title != nil {
		m.Title = *r.Title
	}
	if r.Value != nil {
		m.Value = *r.Value
	}
	m.SchoolId = r.SchoolId
}
