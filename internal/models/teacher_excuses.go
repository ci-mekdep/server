package models

import "time"

type TeacherExcuse struct {
	ID            string     `json:"id"`
	TeacherId     string     `json:"teacher_id"`
	SchoolId      string     `json:"school_id"`
	StartDate     time.Time  `json:"start_date"`
	EndDate       time.Time  `json:"end_date"`
	Reason        string     `json:"reason"`
	Note          *string    `json:"note"`
	DocumentFiles []string   `json:"document_files"`
	CreatedAt     *time.Time `json:"created_at"`
	UpdatedAt     *time.Time `json:"updated_at"`
	Teacher       *User      `json:"teacher"`
	School        *School    `json:"school"`
}

type TeacherExcuses struct {
	TeacherExcuses []*TeacherExcuse
	Total          int
}

type TeacherExcuseCreateDto struct {
	ID            string     `json:"id" form:"id"`
	TeacherId     string     `json:"teacher_id" form:"teacher_id"`
	SchoolId      *string    `json:"school_id" form:"school_id"`
	StartDate     *time.Time `json:"start_date" form:"start_date" time_format:"2006-01-02"`
	EndDate       *time.Time `json:"end_date" form:"end_date" time_format:"2006-01-02"`
	Reason        string     `json:"reason" form:"reason"`
	Note          *string    `json:"note" form:"note"`
	DocumentFiles []string   ``
	CreatedAt     *time.Time `json:"created_at" form:"created_at"`
	UpdatedAt     *time.Time `json:"updated_at" form:"updated_at"`
}

type TeacherExcuseQueryDto struct {
	ID            *string    `form:"id"`
	Ids           []string   `form:"ids"`
	TeacherId     *string    `form:"teacher_id"`
	SchoolId      *string    `form:"school_id"`
	Date          *time.Time `form:"date" time_format:"2006-01-02"`
	Reason        *string    `form:"reason"`
	LoadRelations *bool      `form:"load_relations"`
	Search        *string    `form:"search"`
	Sort          *string    `form:"sort"`
	Limit         int        `form:"limit"`
	Offset        int        `form:"offset"`
}

type TeacherExcusesResponse struct {
	TeacherExcuses []TeacherExcuseResponse `json:"teacher_excuses"`
	Total          int                     `json:"total"`
}

type TeacherExcuseResponse struct {
	ID            string          `json:"id"`
	TeacherId     string          `json:"teacher_id"`
	StartDate     time.Time       `json:"start_date"`
	SchoolId      string          `json:"school_id"`
	SchoolName    string          `json:"school_name"`
	EndDate       time.Time       `json:"end_date"`
	Reason        string          `json:"reason"`
	Note          *string         `json:"note"`
	DocumentFiles []string        `json:"document_files"`
	CreatedAt     *time.Time      `json:"created_at"`
	UpdatedAt     *time.Time      `json:"updated_at"`
	Teacher       *UserResponse   `json:"teacher"`
	School        *SchoolResponse `json:"school"`
}

func ConvertTeacherExcuseToResponse(model TeacherExcuse) (response TeacherExcuseResponse) {
	response = TeacherExcuseResponse{
		ID:        model.ID,
		TeacherId: model.TeacherId,
		SchoolId:  model.SchoolId,
		StartDate: model.StartDate,
		EndDate:   model.EndDate,
		Reason:    model.Reason,
		Note:      model.Note,
		CreatedAt: model.CreatedAt,
		UpdatedAt: model.UpdatedAt,
	}
	if model.School != nil {
		response.SchoolName = *model.School.Name
	}
	if model.Teacher != nil {
		resTeacher := UserResponse{}
		resTeacher.FromModel(model.Teacher)
		response.Teacher = &resTeacher
	}
	if model.School != nil {
		resSchool := SchoolResponse{}
		resSchool.FromModel(model.School)
		response.School = &resSchool
	}
	if model.DocumentFiles != nil {
		response.DocumentFiles = []string{}
		for _, v := range model.DocumentFiles {
			response.DocumentFiles = append(response.DocumentFiles, fileUrl(&v))
		}
	}
	return response
}

func ConvertTeacherExcuseQueryToMap(query TeacherExcuseQueryDto) (m map[string]interface{}) {
	m = map[string]interface{}{}
	if v := query.ID; v != nil && *v != "" {
		m["id"] = v
	}
	if v := query.Ids; v != nil && len(v) >= 1 {
		m["ids"] = v
	}
	if v := query.TeacherId; v != nil && *v != "" {
		m["teacher_id"] = v
	}
	if v := query.SchoolId; v != nil && *v != "" {
		m["school_id"] = v
	}
	if v := query.Date; v != nil {
		m["date"] = v
	}
	if v := query.Reason; v != nil && *v != "" {
		m["reason"] = v
	}
	if v := query.LoadRelations; v != nil {
		m["load_relations"] = *v
	} else {
		m["load_relations"] = true
	}
	if v := query.Limit; v != 0 {
		m["limit"] = v
	} else {
		m["limit"] = 100
	}
	if v := query.Offset; v != 0 {
		m["offset"] = v
	} else {
		m["offset"] = 0
	}
	return
}
