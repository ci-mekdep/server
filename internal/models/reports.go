package models

import (
	"time"
)

type ValueTypes struct {
	Header      string          `json:"header"`
	Key         *string         `json:"key"`
	Type        ReportValueType `json:"type"`
	TypeOptions []string        `json:"type_options"`
	Group       *string         `json:"group"`
}

type Reports struct {
	ID                   string         `json:"id"`
	Title                string         `json:"title"`
	Description          *string        `json:"description"`
	ValueTypes           []ValueTypes   `json:"value_types"`
	SchoolIds            []string       `json:"school_ids"`
	RegionIds            []string       `json:"region_ids"`
	IsPinned             *bool          `json:"is_pinned"`
	IsCenterRating       *bool          `json:"is_center_rating"`
	IsClassroomsIncluded *bool          `json:"is_classrooms_included"`
	CreatedAt            *time.Time     `json:"created_at"`
	UpdatedAt            *time.Time     `json:"updated_at"`
	ItemsCount           *int           `json:"items_count"`
	ItemsCountFilled     *int           `json:"items_filled_count"`
	ReportItem           *ReportItems   `json:"report_item"`
	ReportItems          []*ReportItems `json:"report_items"`
}

type ReportRating struct {
	Report           ReportsResponse    `json:"report"`
	ReportRatingList []ReportRatingItem `json:"rating"`
}
type ReportRatingItem struct {
	ReportItems *ReportItemsResponse `json:"report_item"`
	Value       int                  `json:"value"`
	Index       int                  `json:"index"`
}

func (Reports) RelationFields() []string {
	return []string{"ReportItems", "ReportItem"}
}

type ReportsRequest struct {
	ID                   *string      `json:"id"`
	Title                string       `json:"title"`
	Description          *string      `json:"description"`
	ValueTypes           []ValueTypes `json:"value_types"`
	SchoolIds            []string     `json:"school_ids"`
	IsPinned             *bool        `json:"is_pinned"`
	IsClassroomsIncluded *bool        `json:"is_classrooms_included"`
	IsCenterRating       *bool        `json:"is_center_rating"`
}

type ReportsResponse struct {
	ID                   string                `json:"id"`
	Title                string                `json:"title"`
	Description          *string               `json:"description"`
	ValueTypes           []ValueTypes          `json:"value_types"`
	SchoolIds            []string              `json:"school_ids"`
	RegionIds            []string              `json:"region_ids"`
	IsPinned             *bool                 `json:"is_pinned"`
	IsCenterRating       *bool                 `json:"is_center_rating"`
	CreatedAt            *time.Time            `json:"created_at"`
	UpdatedAt            *time.Time            `json:"updated_at"`
	ItemsCount           *int                  `json:"items_count"`
	ItemsCountFilled     *int                  `json:"items_filled_count"`
	IsClassroomsIncluded *bool                 `json:"is_classrooms_included"`
	ReportItem           *ReportItemsResponse  `json:"report_item"`
	ReportItems          []ReportItemsResponse `json:"report_items"`
}

type ReportsFilterRequest struct {
	ID  *string   `form:"id"`
	IDs *[]string `form:"ids"`
	// TODO: kop yerde gaytalanyany ucin ids bolsa id gerekdal
	SchoolIds      *[]string `form:"school_ids"`
	ClassroomUids  *[]string `json:"classroom_id"`
	IsPinned       *bool     `form:"is_pinned"`
	IsCenterRating *bool     `form:"is_center_rating"`
	Search         *string   `form:"search"`
	PaginationRequest
}

func (r *ReportsResponse) FromModel(m *Reports) error {
	r.ID = m.ID
	r.Title = m.Title
	if m.Description != nil {
		r.Description = m.Description
	}
	r.ValueTypes = m.ValueTypes
	r.SchoolIds = m.SchoolIds
	r.RegionIds = m.RegionIds
	if m.IsPinned != nil {
		r.IsPinned = m.IsPinned
	}
	if m.IsCenterRating != nil {
		r.IsCenterRating = m.IsCenterRating
	}
	if m.CreatedAt != nil {
		r.CreatedAt = m.CreatedAt
	}
	if m.UpdatedAt != nil {
		r.UpdatedAt = m.UpdatedAt
	}
	if m.ReportItem != nil {
		r.ReportItem = &ReportItemsResponse{}
		r.ReportItem.FromModel(m.ReportItem)
	}
	r.ItemsCount = m.ItemsCount
	r.ItemsCountFilled = m.ItemsCountFilled
	r.IsClassroomsIncluded = m.IsClassroomsIncluded
	return nil
}

func (r *ReportsRequest) ToModel(m *Reports) error {
	if r.ID == nil {
		r.ID = new(string)
	}
	m.ID = *r.ID
	m.Title = r.Title
	m.Description = r.Description
	m.ValueTypes = r.ValueTypes
	m.SchoolIds = r.SchoolIds
	if r.IsPinned != nil {
		m.IsPinned = r.IsPinned
	}
	if r.IsCenterRating != nil {
		m.IsCenterRating = r.IsCenterRating
	}
	if r.IsClassroomsIncluded != nil {
		m.IsClassroomsIncluded = r.IsClassroomsIncluded
	}
	return nil
}

// ------------> report_items <------------- //
type ReportItems struct {
	ID               string     `json:"id"`
	ReportId         *string    `json:"report_id"`
	SchoolId         *string    `json:"school_id"`
	PeriodId         *string    `json:"period_id"`
	ClassroomId      *string    `json:"classroom_id"`
	UpdatedBy        *string    `json:"updated_by"`
	Values           []*string  `json:"values"`
	IsEditedManually *bool      `json:"is_edited_manually"`
	CreatedAt        *time.Time `json:"created_at"`
	UpdatedAt        *time.Time `json:"updated_at"`
	Report           *Reports   `json:"report"`
	School           *School    `json:"school"`
	Region           *School    `json:"region"`
	Period           *Period    `json:"period"`
	Classroom        *Classroom `json:"classroom"`
	UpdatedByUser    *User      `json:"updated_by_user"`
}

func (ReportItems) RelationFields() []string {
	return []string{"Report", "School", "Region", "Period", "Classroom", "UpdatedByUser"}
}

type ReportItemsRequest struct {
	ID               *string   `json:"id"`
	ReportId         *string   `json:"report_id"`
	SchoolId         *string   `json:"school_id"`
	PeriodId         *string   `json:"period_id"`
	ClassroomId      *string   `json:"classroom_id"`
	UpdatedBy        *string   `json:"updated_by"`
	IsEditedManually *bool     `json:"is_edited_manually"`
	Values           []*string `json:"values"`
}

type ReportItemsResponse struct {
	ID               string             `json:"id"`
	Report           *ReportsResponse   `json:"report"`
	School           *SchoolResponse    `json:"school"`
	Period           *PeriodResponse    `json:"period"`
	Classroom        *ClassroomResponse `json:"classroom"`
	UpdatedByUser    *UserResponse      `json:"updated_by_user"`
	IsEditedManually *bool              `json:"is_edited_manually"`
	Values           []*string          `json:"values"`
	CreatedAt        *time.Time         `json:"created_at"`
	UpdatedAt        *time.Time         `json:"updated_at"`
}

type ReportItemsFilterRequest struct {
	ID            *string   `json:"id"`
	IDs           *[]string `json:"ids"`
	ReportId      *string   `json:"report_id"`
	SchoolId      *string   `json:"school_id"`
	SchoolIds     []string  `json:"school_ids[]"`
	ClassroomId   *string   `json:"classroom_id"`
	PeriodId      *string   `json:"period_id"`
	Sort          *string   `json:"sort"`
	OnlyClassroom *bool     `json:"only_clasroom"`
	PaginationRequest
}

func (r *ReportItemsResponse) FromModel(m *ReportItems) error {
	r.ID = m.ID
	if m.Report != nil {
		r.Report = &ReportsResponse{}
		r.Report.FromModel(m.Report)
	}
	if m.School != nil {
		r.School = &SchoolResponse{}
		r.School.FromModel(m.School)
	}
	if m.Period != nil {
		r.Period = &PeriodResponse{}
		r.Period.FromModel(m.Period)
	}
	if m.Classroom != nil {
		r.Classroom = &ClassroomResponse{}
		r.Classroom.FromModel(m.Classroom)
	}
	if m.UpdatedByUser != nil {
		r.UpdatedByUser = &UserResponse{}
		r.UpdatedByUser.FromModel(m.UpdatedByUser)
	}
	r.IsEditedManually = m.IsEditedManually
	r.Values = m.Values
	if m.CreatedAt != nil {
		r.CreatedAt = m.CreatedAt
	}
	if m.UpdatedAt != nil {
		r.UpdatedAt = m.UpdatedAt
	}
	return nil
}

func (r *ReportItemsRequest) ToModel(m *ReportItems) error {
	if r.ID == nil {
		r.ID = new(string)
	}
	m.ID = *r.ID
	m.ReportId = r.ReportId
	m.SchoolId = r.SchoolId
	m.PeriodId = r.PeriodId
	m.ClassroomId = r.ClassroomId
	m.IsEditedManually = r.IsEditedManually
	m.UpdatedBy = r.UpdatedBy
	m.Values = r.Values
	return nil
}

type ReportValueType string

const ReportValueTypeNumber ReportValueType = "number"
const ReportValueTypeText ReportValueType = "text"
const ReportValueTypeList ReportValueType = "list"
const ReportValueTypeSelect ReportValueType = "select"

type ReportKey string

const ReportKeySeasonStudents ReportKey = "season_students"
const ReportKeySeasonStudentsCompleted ReportKey = "season_students_completed"
const ReportKeyCourses ReportKey = "courses"
const ReportKeyGroups ReportKey = "groups"

type reportTypeKey struct {
	Title       string          `json:"header"`
	Group       string          `json:"group"`
	Description string          `json:"description"`
	Key         ReportKey       `json:"key"`
	Type        ReportValueType `json:"type"`
	CalcPoint   (func(value interface{}) int)
	Value       string `json:"value"`
}

var DefaultRatingReports = []reportTypeKey{
	{
		Title: "Hasabat döwri üçin ähli ugurlar boýunça okuwa kabul edilen diňleýjileriň sany (okuwy doly tamamlanan toparlar boýunça)",
		Key:   ReportKeySeasonStudents,
		Type:  ReportValueTypeNumber,
		// CalcPoint: func(val interface{}) int {
		// 	num, _ := strconv.Atoi(val.(string))
		// 	return num * 5
		// },
	},
	{
		Title: "Hasabat döwri üçin ähli ugurlar boýunça okuwa kabul edilen we üstünlikli tamamlanan diňleýjileriň sany (okuwy doly tamamlanan toparlar boýunça)",
		Key:   ReportKeySeasonStudentsCompleted,
		Type:  ReportValueTypeNumber,
		// CalcPoint: func(val interface{}) int {
		// 	num, _ := strconv.Atoi(val.(string))
		// 	return num * 5
		// },
	},

	{
		Title: "Hasabat döwri üçin pedagogik işgärleri tarapyndan okatmagyň usullary işlenip taýýarlanylmagy we tejribä ornaşdyrylmagy",
		Key:   "pedagogy",
		Type:  ReportValueTypeList,
		CalcPoint: func(val interface{}) int {
			return val.(int) * 5
		},
	},

	{ // TODO: ask duplicate?
		Title: "Hasabat döwri üçin okuwa kabul edilen diňleýjileriň sany (ähli şahamçalar boýunça)",
		Key:   "students2",
		Type:  ReportValueTypeNumber,
		CalcPoint: func(val interface{}) int {
			num := val.(int)
			if num < 100 {
				return 20
			} else if num < 280 {
				return 25
			} else if num < 400 {
				return 30
			} else if num < 880 {
				return 35
			} else if num < 1600 {
				return 40
			} else {
				return 50
			}
		},
	},

	{
		Title: "Hasabat döwri üçin açylan halkara synaglaryna taýýarlyk toparlary", // list
		Key:   "exams_international",
		Type:  ReportValueTypeList,
	},
	{
		Title: "Halkara synaglaryna taýýarlyk toparlaryna kabul edilen diňleýjileriň sany (ähli şahamçalar boýunça)",
		Key:   "exams_international_count",
		Type:  ReportValueTypeNumber,
		CalcPoint: func(val interface{}) int {
			num := val.(int)
			if num < 10 {
				return 25
			} else if num < 20 {
				return 30
			} else if num < 50 {
				return 35
			} else if num < 100 {
				return 40
			} else {
				return 50
			}
		},
	},

	{
		Title: "Hasabat döwri üçin Okuw merkezleriniň arasynda geçirilen döwlet bäsleşikleride orun gazanylanlary",
		Key:   "competitions",
		Type:  ReportValueTypeList,
	},
	{
		Title: "Hasabat döwri üçin Okuw merkezleriniň arasynda geçirilen döwlet bäsleşiklerinde Okuw merkeziniň diňleýjileriniň gazanan I orun sany",
		Key:   "competitions1_count",
		Type:  ReportValueTypeNumber,
		CalcPoint: func(val interface{}) int {
			num := val.(int)
			return num * 20
		},
	},
	{
		Title: "II orun sany",
		Key:   "competitions2_count",
		Type:  ReportValueTypeNumber,
		CalcPoint: func(val interface{}) int {
			num := val.(int)
			return num * 15
		},
	},
	{
		Title: "III orun sany",
		Key:   "competitions3_count",
		Type:  ReportValueTypeNumber,
		CalcPoint: func(val interface{}) int {
			num := val.(int)
			return num * 10
		},
	},

	{
		Title: "Hasabat döwri üçin Türkmenistanyň ýaşlarynyň arasynda geçirilýän bilim-taslama bäsleşikleriň orun gazanylanlary",
		Key:   "exhibitions",
		Type:  ReportValueTypeNumber,
		CalcPoint: func(val interface{}) int {
			num := val.(int)
			return num * 5
		},
	},
	{
		Title: "Hasabat döwri üçin ýaşlaryň arasynda geçirilýän döwlet we halkara bilim-taslama bäsleşiklerinde Okuw merkeziniň diňleýjileriniň gazanan I orun sany",
		Key:   "exhibitions1_count",
		Type:  ReportValueTypeNumber,
		CalcPoint: func(val interface{}) int {
			num := val.(int)
			return num * 20
		},
	},
	{
		Title: "II orun sany",
		Key:   "exhibitions2_count",
		Type:  ReportValueTypeNumber,
		CalcPoint: func(val interface{}) int {
			num := val.(int)
			return num * 15
		},
	},
	{
		Title: "II orun sany",
		Key:   "exhibitions3_count",
		Type:  ReportValueTypeNumber,
		CalcPoint: func(val interface{}) int {
			num := val.(int)
			return num * 10
		},
	},

	{
		Title: "Okuw merkezlerinde bilim alýan mümkinçiligi çäkli adamlaryň sany (A)",
		Key:   "disabilities_count",
		Type:  ReportValueTypeNumber,
		CalcPoint: func(val interface{}) int {
			num := val.(int)
			return num * 5
		},
	},

	{
		Title: "Hasabat döwri üçin okadylýan ugurlar boýunça I hünär derejeli pedagogik işgärleriň sanynyň (A) okadylýan ugurlar boýunça jemi sanyna (B) gatnaşygy",
		Key:   "teachers1_count",
		Type:  ReportValueTypeNumber,
		CalcPoint: func(val interface{}) int {
			num := val.(int)
			return num * 50
		},
	},
	{
		Title: "Hasabat döwri üçin okadylýan ugurlar boýunça II hünär derejeli pedagogik işgärleriň sanynyň (A) okadylýan ugurlar boýunça jemi sanyna (B) gatnaşygy",
		Key:   "teachers2_count",
		Type:  ReportValueTypeNumber,
		CalcPoint: func(val interface{}) int {
			num := val.(int)
			return num * 70
		},
	},
	{
		Title: "Hasabat döwri üçin okadylýan ugurlar boýunça III hünär derejeli pedagogik işgärleriň sanynyň (A) okadylýan ugurlar boýunça jemi sanyna (B) gatnaşygy",
		Key:   "teachers3_count",
		Type:  ReportValueTypeNumber,
		CalcPoint: func(val interface{}) int {
			num := val.(int)
			return num * 70
		},
	},
	{
		Title: "Hasabat döwri üçin okadylýan ugurlar boýunça Alymlyk hünär derejeli pedagogik işgärleriň sanynyň (A) okadylýan ugurlar boýunça jemi sanyna (B) gatnaşygy",
		Key:   "teachers4_count",
		Type:  ReportValueTypeNumber,
		CalcPoint: func(val interface{}) int {
			num := val.(int)
			return num * 100
		},
	},

	{
		Title: "Hasabat döwri üçin Okuw merkeziniň pedagogik işgärleriň halkara maslahatlara, seminarlara gatnaşanlarynyň sany (A)",
		Key:   "teachers_international",
		Type:  ReportValueTypeNumber,
		CalcPoint: func(val interface{}) int {
			num := val.(int)
			return num * 10
		},
	},

	{
		Title: "Hasabat döwri üçin Okuw merkeziniň pedagogik işgärleriň tejribelerini artdyrmak (tejribe alyşmak) maksady bilen, okuw merkezi tarapyndan daşary ýurtlara iş saparyna iberilen pedagogiki işgärleriň sany (A)",
		Key:   "teachers_international_trained",
		Type:  ReportValueTypeNumber,
		CalcPoint: func(val interface{}) int {
			num := val.(int)
			return num * 20
		},
	},

	{
		Title: "Halkara we döwlet derejesinde geçirilýän jemgyýetçilik we bilim çärelerine işjeň gatnaşandygy üçin sylaglandyrylan (Hormat haty, Minnetdarlyk haty we beýlekiler) işgärleriň hasabat döwrine degişli sylaglarynyň sany",
		Key:   "teachers_social",
		Type:  ReportValueTypeNumber,
		CalcPoint: func(val interface{}) int {
			num := val.(int)
			return num * 10
		},
	},
	{
		Title: "Arkadag we Aşgabat şäherleri hem-de welaýat derejesinde geçirilýän jemgyýetçilik we bilim çärelerine işjeň gatnaşandygy üçin sylaglandyrylan (Hormat haty, Minnetdarlyk haty we beýlekiler) işgärleriň hasabat döwrüne degişli sylaglarynyň sany",
		Key:   "teachers_social_capital",
		Type:  ReportValueTypeNumber,
		CalcPoint: func(val interface{}) int {
			num := val.(int)
			return num * 5
		},
	},
	{
		Title: "Hasabat döwri üçin Okuw merkeziniň Sanly bilim portalyna gündelik girýän ulanyjylaryň sanynyň (A) okuw merkeziniň mugallymlarynyň, diňleýjileriniň sanyna (B) bolan gatnaşygy (okuw ýyly boýunça ortaça bir güne düşýän sanyndan ugur alynýar)",
		Key:   "online_count",
		Type:  ReportValueTypeNumber,
		CalcPoint: func(val interface{}) int {
			num := val.(int)
			return num
		},
	},
}
