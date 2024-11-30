package models

import "time"

type BaseSubjects struct {
	ID           string     `json:"id"`
	SchoolId     *string    `json:"school_id"`
	Name         *string    `json:"name"`
	Category     *string    `json:"category"`
	Price        *uint      `json:"price"`
	ExamMinGrade *uint      `json:"exam_min_grade"`
	AgeCategory  *string    `json:"age_category"`
	IsAvailable  *bool      `json:"is_available"`
	UpdatedAt    *time.Time `json:"updated_at"`
	CreatedAt    *time.Time `json:"created_at"`
	School       *School    `json:"school"`
}

func (BaseSubjects) RelationFields() []string {
	return []string{"School"}
}

type BaseSubjectsRequest struct {
	ID           string  `json:"id"`
	SchoolId     *string `json:"school_id"`
	Name         *string `json:"name"`
	Category     *string `json:"category"`
	Price        *uint   `json:"price"`
	ExamMinGrade *uint   `json:"exam_min_grade"`
	AgeCategory  *string `json:"age_category"`
	IsAvailable  *bool   `json:"is_available"`
}

type BaseSubjectsResponse struct {
	ID           string          `json:"id"`
	Name         *string         `json:"name"`
	Category     *string         `json:"category"`
	Price        *uint           `json:"price"`
	ExamMinGrade *uint           `json:"exam_min_grade"`
	AgeCategory  *string         `json:"age_category"`
	IsAvailable  *bool           `json:"is_available"`
	School       *SchoolResponse `json:"school"`
	UpdatedAt    *time.Time      `json:"updated_at"`
	CreatedAt    *time.Time      `json:"created_at"`
}

func (r *BaseSubjectsResponse) FromModel(m *BaseSubjects) error {
	r.ID = m.ID
	r.Name = m.Name
	r.Category = m.Category
	r.Price = m.Price
	r.ExamMinGrade = m.ExamMinGrade
	r.AgeCategory = m.AgeCategory
	r.IsAvailable = m.IsAvailable
	if m.School != nil {
		r.School = &SchoolResponse{}
		r.School.FromModel(m.School)
	}
	r.UpdatedAt = m.UpdatedAt
	r.CreatedAt = m.CreatedAt
	return nil
}

func (r *BaseSubjectsRequest) ToModel(m *BaseSubjects) error {
	m.ID = r.ID
	m.SchoolId = r.SchoolId
	m.Name = r.Name
	m.Category = r.Category
	m.Price = r.Price
	m.ExamMinGrade = r.ExamMinGrade
	m.AgeCategory = r.AgeCategory
	m.IsAvailable = r.IsAvailable
	return nil
}

type BaseSubjectsFilterRequest struct {
	ID            string    `form:"id"`
	IDs           *[]string `form:"ids"`
	SchoolId      *string   `form:"school_id"`
	SchoolIds     *[]string `form:"school_ids"`
	Categories    *[]string `form:"categories"`
	AgeCategories *[]string `form:"age_categories"`
	Search        *string   `form:"search"`
	Sort          *string   `form:"sort"`
	PaginationRequest
}
