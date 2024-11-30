package models

import (
	"strings"
	"time"
)

type School struct {
	ID                string     `json:"id"`
	Code              *string    `json:"code"`
	Name              *string    `json:"name"`
	FullName          *string    `json:"full_name"`
	Description       *string    `json:"description"`
	Address           *string    `json:"address"`
	Avatar            *string    `json:"avatar"`
	Background        *string    `json:"background"`
	Phone             *string    `json:"phone"`
	Email             *string    `json:"email"`
	Level             *string    `json:"level"`
	Galleries         *[]string  `json:"galleries"`
	Latitude          *string    `json:"latitude"`
	Longitude         *string    `json:"longitude"`
	IsDigitalized     *bool      `json:"is_digitalized"`
	IsSecondarySchool *bool      `json:"is_secondary_school"`
	ParentUid         *string    `json:"parent_id"`
	AdminUid          *string    `json:"admin_id"`
	SpecialistUid     *string    `json:"specialist_id"`
	UpdatedAt         *time.Time `json:"updated_at"`
	CreatedAt         *time.Time `json:"created_at"`
	ArchivedAt        *time.Time `json:"archived_at"`
	TimetablesCount   *int       `json:"timetables_count"`
	ClassroomsCount   *int       `json:"classrooms_count"`
	Parent            *School    `json:"parent"`
	Admin             *User      `json:"admin"`
	Specialist        *User      `json:"specialist"`
}

func (School) RelationFields() []string {
	return []string{"Parent", "Admin", "Specialist"}
}

type UserSchool struct {
	SchoolUid *string `json:"school_id"`
	UserId    string  `json:"user_id"`
	RoleCode  Role    `json:"role_code"`
	School    *School `json:"school"`
	User      *User   `json:"user"`
}

type SchoolValueResponse struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func (r *SchoolResponse) ToValues() SchoolValueResponse {
	value := ""
	if r.Name != nil {
		value = *r.Name + value
	}
	if r.Parent != nil && r.Parent.Name != nil {
		value = *r.Parent.Name + ", " + value
	}
	return SchoolValueResponse{
		Key:   *r.ID,
		Value: strings.Trim(value, " "),
	}
}

type SchoolSetting struct {
	SchoolId  *string          `json:"-"`
	Key       SchoolSettingKey `json:"key"`
	Value     *string          `json:"value"`
	UpdatedAt *time.Time       `json:"updated_at"`
	School    *School          `json:"-"`
}

func (SchoolSetting) RelationFields() []string {
	return []string{"School"}
}

type SchoolSettingRequest struct {
	Key      SchoolSettingKey `json:"key" validate:"required"`
	Value    *string          `json:"value"`
	SchoolId *string          `json:"school_id"`
}
type SchoolSettingRequestForm struct {
	Settings []SchoolSettingRequest `json:"settings"`
}

type SchoolSettingModel struct {
	SchoolId  string            `json:"school_id"`
	Value     *string           `json:"value"`
	UpdatedAt *time.Time        `json:"updated_at"`
	Key       SchoolSettingKey  `json:"key"`
	Type      SchoolSettingType `json:"type"`
	Title     string            `json:"title"`
	Options   *[]string         `json:"options"`
}

func (ss *SchoolSettingModel) FromModel(s SchoolSetting) {
	ss.Value = s.Value
	ss.UpdatedAt = s.UpdatedAt
	if s.SchoolId != nil {
		ss.SchoolId = *s.SchoolId
	}
}

type SchoolSettingKey string
type SchoolSettingType string

const SchoolSettingTypeNumber SchoolSettingType = "number"
const SchoolSettingTypeText SchoolSettingType = "text"
const SchoolSettingTypeTextarea SchoolSettingType = "textarea"
const SchoolSettingTypeBoolean SchoolSettingType = "boolean"
const SchoolSettingTypeOptions SchoolSettingType = "options"

const SchoolSettingTeachersCount SchoolSettingKey = "teachers_count"
const SchoolSettingStudentsCount SchoolSettingKey = "students_count"
const SchoolSettingParentsCount SchoolSettingKey = "parents_count"
const SchoolSettingClassroomsCount SchoolSettingKey = "classrooms_count"
const SchoolSettingStudentsPlaceCount SchoolSettingKey = "students_place_count"
const SchoolSettingRoomsCount SchoolSettingKey = "rooms_count"
const SchoolSettingBoardsCount SchoolSettingKey = "boards_count"
const SchoolSettingComputersCount SchoolSettingKey = "computers_count"
const SchoolSettingInternetSpeed SchoolSettingKey = "internet_speed"
const SchoolSettingInternetArea SchoolSettingKey = "internet_area"
const SchoolSettingLocalNetwork SchoolSettingKey = "local_network"

type SchoolResponse struct {
	ID                *string         `json:"id"`
	Code              *string         `json:"code"`
	Name              *string         `json:"name"`
	FullName          *string         `json:"full_name"`
	Description       *string         `json:"description"`
	Address           *string         `json:"address"`
	Avatar            *string         `json:"avatar"`
	Background        *string         `json:"background"`
	Phone             *string         `json:"phone"`
	Email             *string         `json:"email"`
	Level             *string         `json:"level"`
	Galleries         *[]string       `json:"galleries"`
	Latitude          *string         `json:"latitude"`
	Longitude         *string         `json:"longitude"`
	IsDigitalized     *bool           `json:"is_digitalized"`
	IsSecondarySchool *bool           `json:"is_secondary_school"`
	UpdatedAt         *time.Time      `json:"updated_at"`
	CreatedAt         *time.Time      `json:"created_at"`
	ArchivedAt        *time.Time      `json:"archived_at"`
	TimetablesCount   *int            `json:"timetables_count"`
	ClassroomsCount   *int            `json:"classrooms_count"`
	Parent            *SchoolResponse `json:"parent"`
	Admin             *UserResponse   `json:"admin"`
	Specialist        *UserResponse   `json:"specialist"`
}

type SchoolFilterRequest struct {
	ID                *string   `form:"id"`
	NotIds            *[]string `form:"not_ids"`
	Code              *string   `form:"code"`
	Codes             *[]string `form:"codes"`
	Uids              *[]string `form:"ids"`
	IsParent          *bool     `form:"is_parent"`
	IsSecondarySchool *bool     `form:"is_secondary_school"`
	ParentUid         *string   `form:"parent_id"`
	ParentUids        *[]string `form:"parent_ids"`
	Regions           *bool     `form:"regions"`
	AdminId           *string   `form:"admin_id"`
	SpecialistId      *string   `form:"specialist_id"`
	Address           *string   `form:"address"`
	Search            *string   `form:"search"`
	Sort              *string   `form:"sort"`
	PaginationRequest
}

type SchoolRequest struct {
	ID                *string   `json:"id" form:"id"`
	Code              *string   `json:"code" form:"code"`
	Name              *string   `json:"name" form:"name"`
	FullName          *string   `json:"full_name" form:"full_name"`
	Description       *string   `json:"description" form:"description"`
	Address           *string   `json:"address" form:"address"`
	Avatar            *string   ``
	AvatarDelete      *bool     `json:"avatar_delete" form:"avatar_delete"`
	Background        *string   `json:"background" form:"backgrount"`
	Phone             *string   `json:"phone" form:"phone"`
	Email             *string   `json:"email" form:"email"`
	Level             *string   `json:"level" form:"level"`
	Galleries         *[]string ``
	GalleriesDelete   *[]string `json:"galleries_delete" form:"galleries_delete"`
	Latitude          *string   `json:"latitude" form:"latitude"`
	Longitude         *string   `json:"longitude" form:"longitude"`
	IsDigitalized     *bool     `json:"is_digitalized" form:"is_digitalized"`
	IsSecondarySchool *bool     `json:"is_secondary_school" form:"is_secondary_school"`
	IsArchive         *bool     `json:"is_archive" form:"is_archive"`
	ParentUid         *string   `json:"parent_id" form:"parent_id"`
	AdminUid          *string   `json:"admin_id" form:"admin_id"`
	SpecialistUId     *string   `json:"specialist_id" form:"specialist_id"`
}

type UserSchoolResource struct {
	RoleCode Role            `json:"role_code"`
	School   *SchoolResponse `json:"school"`
}

func (u *UserSchoolResource) FromModel(m *UserSchool) {
	u.RoleCode = m.RoleCode
	if m.School != nil {
		u.School = &SchoolResponse{}
		u.School.FromModel(m.School)
	}
}

func (r *SchoolResponse) FromModel(m *School) {
	r.ID = &m.ID
	r.Code = m.Code
	r.Name = m.Name
	r.FullName = m.FullName
	r.Description = m.Description
	r.Address = m.Address
	if m.Avatar != nil {
		a := fileUrl(m.Avatar)
		r.Avatar = &a
	}
	r.Background = m.Background
	r.Phone = m.Phone
	r.Email = m.Email
	r.Level = m.Level
	if m.Galleries != nil {
		r.Galleries = &[]string{}
		for _, f := range *m.Galleries {
			*r.Galleries = append(*r.Galleries, fileUrl(&f))
		}
	}
	r.Latitude = m.Latitude
	r.Longitude = m.Longitude
	r.IsDigitalized = m.IsDigitalized
	r.IsSecondarySchool = m.IsSecondarySchool
	r.UpdatedAt = m.UpdatedAt
	r.CreatedAt = m.CreatedAt
	r.ArchivedAt = m.ArchivedAt
	r.TimetablesCount = m.TimetablesCount
	r.ClassroomsCount = m.ClassroomsCount
	if m.Admin != nil {
		r.Admin = &UserResponse{}
		r.Admin.FromModel(m.Admin)
	}
	if m.Parent != nil {
		r.Parent = &SchoolResponse{}
		r.Parent.FromModel(m.Parent)
	}
	if m.Specialist != nil {
		r.Specialist = &UserResponse{}
		r.Specialist.FromModel(m.Specialist)
	}
}

func (r *SchoolRequest) ToModel(m *School) {
	if r.ID != nil {
		m.ID = *r.ID
	}
	m.Code = r.Code
	m.Name = r.Name
	m.FullName = r.FullName
	m.Description = r.Description
	m.Address = r.Address
	if r.Avatar != nil {
		m.Avatar = r.Avatar
	}
	m.Background = r.Background
	m.Phone = r.Phone
	m.Email = r.Email
	m.Level = r.Level
	m.Galleries = r.Galleries
	m.Latitude = r.Latitude
	m.Longitude = r.Longitude
	m.IsDigitalized = r.IsDigitalized
	if r.IsSecondarySchool != nil {
		m.IsSecondarySchool = r.IsSecondarySchool
	}
	m.ParentUid = r.ParentUid
	m.AdminUid = r.AdminUid
	m.SpecialistUid = r.SpecialistUId
	if m.ArchivedAt == nil && r.IsArchive != nil && *r.IsArchive {
		m.ArchivedAt = new(time.Time)
		*m.ArchivedAt = time.Now()
	}
}

var Regions map[string][]string = map[string][]string{
	"ark": []string{"ark"},
	"bm":  []string{"bm"},
	"ag":  []string{"brk", "bgt", "bzm", "kpt", "ag"},
	"ah":  []string{"tjn", "akb", "bhr", "gkd", "kka", "kkas", "srs", "tjns", "bbd"},
	"bn":  []string{"esn", "brkt", "etr", "hzr", "bnb", "tmbs", "tmb", "gmd", "mgt", "gzt"},
	"mr":  []string{"mrs", "byrs", "byr", "wkl", "ylt", "grg", "mre", "mrg", "skr", "tgt", "tmg"},
	"lb":  []string{"tmas", "tma", "drg", "dnw", "krk", "syt", "hlc", "hjm", "crj", "kyt"},
	"dz":  []string{"dzs", "bld", "kng", "akd", "stb", "gor", "rhb", "sbt"},
}

type StateType struct {
	Name string `json:"name"`
	Url  string `json:"url"`
	Code string `json:"code"`
}

var States []APIInstance = []APIInstance{
	{Name: "Aşgabat", Url: "https://mekdep.edu.tm", Code: "ag"},
	{Name: "Arkadag", Url: "https://mekdep.edu.tm", Code: "ark"},
	{Name: "Ahal welaýaty", Url: "https://ah.mekdep.edu.tm", Code: "ah"},
	{Name: "Mary welaýaty", Url: "https://mr.mekdep.edu.tm", Code: "mr"},
	{Name: "Lebap welaýaty", Url: "https://lb.mekdep.edu.tm", Code: "lb"},
	{Name: "Daşoguz welaýaty", Url: "https://dz.mekdep.edu.tm", Code: "dz"},
	{Name: "Balkan welaýaty", Url: "https://bn.mekdep.edu.tm", Code: "bn"},
	{Name: "Bilim merkezler", Url: "https://mekdep.edu.tm", Code: "bm"},
	// {Name: "Beta", Url: "https://beta.mekdep.org", Code: "ag"},
}

func GetStateLabel(code string) string {
	for _, v := range States {
		if v.Code == code {
			return v.Name
		}
	}
	return ""
}
