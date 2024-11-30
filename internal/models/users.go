package models

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/mekdep/server/config"
)

type Role string

const (
	RoleStudent      Role = "student"
	RoleParent       Role = "parent"
	RoleTeacher      Role = "teacher"
	RolePrincipal    Role = "principal"
	RoleOrganization Role = "organization"
	RoleOperator     Role = "operator"
	RoleAdmin        Role = "admin"
)

type Status string

const (
	StatusActive  Status = "active"
	StatusWait    Status = "wait"
	StatusBlocked Status = "blocked"
)

type Gender int

const (
	GenderMale   Gender = 1
	GenderFemale Gender = 2
)

var DefaultRoles []Role = []Role{
	RoleStudent,
	RoleParent,
	RoleTeacher,
	RolePrincipal,
	RoleOperator,
	RoleOrganization,
	RoleAdmin,
}

var DefaultStatuses []Status = []Status{
	StatusActive,
	StatusWait,
	StatusBlocked,
}

type Documents struct {
	Key    *string `json:"key"`
	Number *string `json:"number"`
	Date   *string `json:"date"`
}

type User struct {
	ID              string     `json:"id"`
	FirstName       *string    `json:"first_name"`
	LastName        *string    `json:"last_name"`
	MiddleName      *string    `json:"middle_name"`
	Username        *string    `json:"username" `
	Password        *string    `json:"password"`
	Status          *string    `json:"status"`
	Phone           *string    `json:"phone"`
	PhoneVerifiedAt *string    `json:"phone_verified_at"`
	Email           *string    `json:"email"`
	EmailVerifiedAt *string    `json:"email_verified_at"`
	Birthday        *time.Time `json:"birthday"`
	Gender          *int       `json:"gender"`
	Address         *string    `json:"address"`
	Avatar          *string    `json:"avatar"`
	LastActiveAt    *time.Time `json:"last_active_at"`

	PassportNumber  *string `json:"passport_number"`
	BirthCertNumber *string `json:"birth_cert_number"`
	ApplyNumber     *string `json:"apply_number"`
	WorkTitle       *string `json:"work_title"`
	WorkPlace       *string `json:"work_place"`
	District        *string `json:"district"`
	Reference       *string `json:"reference"`
	NickName        *string `json:"nickname"`
	EducationTitle  *string `json:"education_title"`
	EducationPlace  *string `json:"education_place"`
	EducationGroup  *string `json:"education_group"`

	Documents         []Documents `json:"documents"`
	DocumentFiles     *[]string   `json:"document_files"`
	UpdatedAt         *time.Time  `json:"updated_at"`
	CreatedAt         *time.Time  `json:"created_at"`
	ArchivedAt        *time.Time  `json:"archived_at"`
	Role              *string     `json:"role"`
	SchoolName        *string     `json:"school_name"`
	SchoolId          *string     `json:"school_id"`
	SchoolParent      *string     `json:"school_parent"`
	ClassroomId       *string     `json:"classroom_id"`
	ClassroomName     *string     `json:"classroom_name"`
	ChildrenSchoolIds []string    `json:"children_school_ids"`
	// todo: load teacher_classroom from classrooms
	TeacherClassroom *Classroom       `json:"teacher_classroom"`
	Children         []*User          `json:"children"`
	Parents          []*User          `json:"parents"`
	Schools          []*UserSchool    `json:"schools"`
	Classrooms       []*UserClassroom `json:"classrooms"`
}

func (User) RelationFields() []string {
	return []string{"TeacherClassroom", "Children", "Parents", "Schools", "Classrooms"}
}

func (u *User) FormattedPhone() (string, error) {
	p := ""
	if u.Phone != nil {
		p = *u.Phone
		if len(p) < 8 {
			return "", errors.New("phone is invalid")
		}
	} else {
		return "", errors.New("phone is not set")
	}
	return "993" + p, nil
}

func (u *User) FullName() string {
	res := ""
	if u.LastName != nil {
		res += " " + *u.LastName
	}
	if u.FirstName != nil {
		res += " " + *u.FirstName
	}
	if u.MiddleName != nil {
		res += " " + *u.MiddleName
	}
	return strings.Trim(res, " ")
}

func (u *User) ShortName() string {
	res := ""
	if u.LastName != nil {
		res += *u.LastName
	}
	if u.FirstName != nil {
		res += " " + string((*u.FirstName)[0]) + "."
	}
	return res
}

type ParentChildren struct {
	ID       string `json:"id"`
	ParentID string `json:"parent_id"`
	ChildID  string `json:"child_id"`
}

type UserSetting struct {
	ID     string `json:"id"`
	UserID string `json:"user_id"`
	Key    string `json:"key"`
	Value  string `json:"value"`
}

type ModelValueResponse struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func (r *UserResponse) ToValues() ModelValueResponse {
	value := ""
	if r.LastName != nil {
		value += " " + *r.LastName
	}
	if r.FirstName != nil {
		value += " " + *r.FirstName
	}
	if r.MiddleName != nil {
		value += " " + *r.MiddleName
	}
	return ModelValueResponse{
		Key:   r.ID,
		Value: strings.Trim(value, " "),
	}
}

type UserRating struct {
	Value     int                `json:"value"`
	Rating    int                `json:"rating"`
	User      *UserResponse      `json:"user"`
	Classroom *ClassroomResponse `json:"classroom"`
	School    *SchoolResponse    `json:"school"`
}

type UserResponse struct {
	ID              string     `json:"id"`
	FirstName       *string    `json:"first_name"`
	LastName        *string    `json:"last_name"`
	MiddleName      *string    `json:"middle_name"`
	Username        *string    `json:"username" `
	Status          *string    `json:"status"`
	Phone           *string    `json:"phone"`
	PhoneVerifiedAt *string    `json:"phone_verified_at"`
	Email           *string    `json:"email"`
	EmailVerifiedAt *string    `json:"email_verified_at"`
	Birthday        *string    `json:"birthday"`
	Gender          *int       `json:"gender"`
	Address         *string    `json:"address"`
	Avatar          *string    `json:"avatar"`
	LastOnlineAt    *time.Time `json:"last_online_at"`
	IsActive        bool       `json:"is_active"`
	Role            *string    `json:"role"`
	SchoolName      *string    `json:"school_name"`
	SchoolId        *string    `json:"school_id"`
	SchoolParent    *string    `json:"school_parent"`
	ClassroomId     *string    `json:"classroom_id"`
	ClassroomName   *string    `json:"classroom_name"`

	PassportNumber  *string     `json:"passport_number"`
	BirthCertNumber *string     `json:"birth_cert_number"`
	ApplyNumber     *string     `json:"apply_number"`
	WorkTitle       *string     `json:"work_title"`
	WorkPlace       *string     `json:"work_place"`
	District        *string     `json:"district"`
	Reference       *string     `json:"reference"`
	NickName        *string     `json:"nickname"`
	EducationTitle  *string     `json:"education_title"`
	EducationPlace  *string     `json:"education_place"`
	EducationGroup  *string     `json:"education_group"`
	Documents       []Documents `json:"documents"`
	DocumentFiles   *[]string   `json:"document_files"`
	TariffEndAt     *time.Time  `json:"tariff_end_at"`
	TariffType      *string     `json:"tariff_type"`
	TariffName      *string     `json:"tariff_name"`
	CreatedAt       *time.Time  `json:"created_at"`
	UpdatedAt       *time.Time  `json:"updated_at"`

	Children         []*UserResponse       `json:"children"`
	Parents          []*UserResponse       `json:"parents"`
	Schools          []*UserSchoolResource `json:"schools"`
	Classrooms       []*ClassroomResponse  `json:"classrooms"`
	TeacherClassroom *ClassroomResponse    `json:"teacher_classroom"`
}

type UserRequest struct {
	ID              *string   `json:"id" form:"id" validate:"omitempty"`
	FirstName       *string   `json:"first_name" binding:"omitempty" form:"first_name"`
	LastName        *string   `json:"last_name" form:"last_name"`
	MiddleName      *string   `json:"middle_name" form:"middle_name"`
	Username        *string   `json:"username" form:"username"`
	Status          *string   `json:"status" form:"status"`
	Phone           *string   `json:"phone" form:"phone"`
	Email           *string   `json:"email" form:"email"`
	Birthday        *string   `json:"birthday" form:"birthday"`
	Gender          *int      `json:"gender" form:"gender"`
	Address         *string   `json:"address" form:"address,omitempty" binding:"omitempty"`
	Avatar          *string   ``
	AvatarDelete    *bool     `json:"avatar_delete"`
	Password        *string   `json:"password" form:"password"`
	ChildIds        *[]string `json:"child_ids" form:"child_ids"`
	ParentIds       *[]string `json:"parent_ids" form:"parent_ids"`
	PassportNumber  *string   `json:"passport_number" form:"passport_number"`
	BirthCertNumber *string   `json:"birth_cert_number" form:"birth_cert_number"`
	ApplyNumber     *string   `json:"apply_number" form:"apply_number,omitempty"`
	WorkTitle       *string   `json:"work_title" form:"work_title"`
	WorkPlace       *string   `json:"work_place" form:"work_place"`
	District        *string   `json:"district" form:"district"`
	Reference       *string   `json:"reference" form:"reference"`
	NickName        *string   `json:"nickname" form:"nickname"`
	EducationTitle  *string   `json:"education_title" form:"education_title"`
	EducationPlace  *string   `json:"education_place" form:"education_place"`
	EducationGroup  *string   `json:"education_group" form:"education_group"`

	Documents     *[]Documents            `json:"documents" form:"documents"`
	DocumentFiles *[]string               ``
	Parents       *[]UserRequest          `json:"parents" form:"parents"`
	Children      *[]UserRequest          `json:"children" form:"children"`
	ClassroomIds  *[]UserClassroomRequest `json:"classrooms" form:"classrooms"`
	ClassroomName *string                 `json:"classroom_name" form:"classroom_name"`
	// TODO: add teacher_classroom_id
	TeacherClassroomId *string              `json:"teacher_classroom_id" form:"teacher_classroom_id"`
	SchoolIds          *[]UserSchoolRequest `json:"schools" form:"schools"`
}

type UserProfileUpdateRequest struct {
	FirstName       *string `form:"first_name"`
	LastName        *string `form:"last_name"`
	MiddleName      *string `form:"middle_name"`
	Username        *string `form:"username"`
	Phone           *string `form:"phone"`
	Email           *string `form:"email"`
	Birthday        *string `form:"birthday"`
	Gender          *int    `form:"gender"`
	Address         *string `form:"address"`
	Avatar          *string ``
	AvatarDelete    *bool   `form:"avatar_delete"`
	CurrentSchoolId *string `form:"current_school"`
	CurrentRegionId string  `form:"current_region"`
	CurrentRoleCode *string `form:"current_role"`
	DeviceToken     *string `form:"device_token"`
}

type UserPasswordUpdateRequest struct {
	OldPassword *string `json:"old_password" validate:"required,min=8"`
	NewPassword *string `json:"new_password" validate:"required,min=8"`
}

type UserMeUpdateSchoolRequest struct {
	SchoolId *string `json:"school_id"`
	RegionId *string `json:"region_id"`
	RoleCode *string `json:"role_code"`
	PeriodId *string `json:"period_id"`
}

type UserSchoolRequest struct {
	SchoolUid *string `json:"school_id" form:"school_id"`
	RoleCode  *Role   `json:"role_code" form:"role_code"`
	IsDelete  *bool   `json:"is_delete" form:"is_delete"`
	UserId    *string `json:"user_id" validate:"omitempty"`
}

// TODO: update by docs
type UserFilterRequest struct {
	ID        *string   `form:"id"`
	NotID     *string   `form:"not_id"`
	Ids       *[]string `form:"ids"`
	SchoolId  *string   `form:"school_id"`
	SchoolIds *[]string `form:"school_ids[]"`
	Username  *string   `form:"username"`
	// TODO: remove
	ParentId               *string    `form:"parent_id"`
	ClassroomIdForBirthday *string    `form:"classroom_id_for_birthday"`
	ClassroomId            *string    `form:"classroom_id"`
	ClassroomType          *string    `form:"classroom_type"`
	ClassroomTypeKey       *int       `form:"classroom_type_key"`
	ClassroomName          *string    `form:"classroom_name"`
	Phone                  *string    `json:"phone"`
	NoClassroom            *bool      `form:"no_classroom"`
	IsActive               *bool      `form:"is_active"`
	Role                   *string    `form:"role"`
	TariffEndMin           *time.Time `form:"tariff_end_min"`
	Roles                  *[]string  `form:"roles[]"`
	Status                 *string    `form:"status"`
	Gender                 *int       `form:"gender"`
	Address                *string    `form:"address"`
	Search                 *string    `form:"search"`
	Sort                   *string    `form:"sort"`
	Birthday               *string    `form:"birthday"`
	BirthdayToday          *time.Time `form:"birthday_today"`
	ChildrenCount          *int       `form:"children_count"`
	ParentsCount           *int       `form:"parents_count"`
	IsValuesOnly           bool       `form:"is_values_only"`
	LessonHours            *[]int     `form:"lesson_hours[]"`
	LowFirstName           *string    `form:"low_first_name"`
	LowLastName            *string    `form:"low_last_name"`
	FirstName              *string    ``
	LastName               *string    ``
	NoParent               *bool      `form:"no_parent"`
	PaginationRequest
}

func (data UserFilterRequest) ToMap() map[string]interface{} {
	f := map[string]interface{}{}
	if data.ID != nil {
		f["id"] = *data.ID
	}
	if data.SchoolId != nil {
		f["school_id"] = *data.SchoolId
	}
	if data.ClassroomId != nil {
		f["classroom_id"] = *data.ClassroomId
	}
	if data.Role != nil {
		f["role_code"] = *data.Role
	}
	if data.Status != nil {
		f["status"] = *data.Status
	}
	if data.Gender != nil {
		f["gender"] = *data.Gender
	}
	if data.Birthday != nil {
		f["birthday"] = *data.Birthday
	}
	if data.Address != nil {
		f["address"] = *data.Address
	}
	if data.Search != nil {
		f["search"] = *data.Search
	}
	if data.Sort != nil {
		f["sort"] = *data.Sort
	}
	return f
}

// TODO move
func fileUrl(path *string) string {
	if path != nil {
		return config.Conf.AppUrl + "/uploads/" + *path
	}
	return ""
}

func (u *UserResponse) FromModel(m *User) {
	u.ID = m.ID
	u.FirstName = m.FirstName
	u.LastName = m.LastName
	u.MiddleName = m.MiddleName
	u.Username = m.Username
	u.Status = m.Status
	u.Phone = m.Phone
	u.PhoneVerifiedAt = m.PhoneVerifiedAt
	u.Email = m.Email
	u.EmailVerifiedAt = m.EmailVerifiedAt
	if m.Birthday != nil {
		u.Birthday = new(string)
		*u.Birthday = m.Birthday.Format(time.DateOnly)
	}
	u.Gender = m.Gender
	u.Address = m.Address
	u.LastOnlineAt = m.LastActiveAt
	u.Role = m.Role
	u.SchoolName = m.SchoolName
	u.SchoolId = m.SchoolId
	u.SchoolParent = m.SchoolParent
	u.ClassroomId = m.ClassroomId
	u.ClassroomName = m.ClassroomName
	u.CreatedAt = m.CreatedAt
	u.UpdatedAt = m.UpdatedAt
	if m.LastActiveAt != nil {
		u.IsActive = time.Now().Before((*m.LastActiveAt).Add(time.Minute * 10))
	}
	if m.Avatar != nil {
		a := fileUrl(m.Avatar)
		u.Avatar = &a
	}
	u.Children = []*UserResponse{}
	for _, i := range m.Children {
		r := &UserResponse{}
		r.FromModel(i)
		u.Children = append(u.Children, r)
	}
	u.Parents = []*UserResponse{}
	for _, i := range m.Parents {
		r := &UserResponse{}
		r.FromModel(i)
		u.Parents = append(u.Parents, r)
	}
	u.Schools = []*UserSchoolResource{}
	for _, i := range m.Schools {
		r := &UserSchoolResource{}
		r.FromModel(i)
		u.Schools = append(u.Schools, r)
	}
	u.Classrooms = []*ClassroomResponse{}
	for _, i := range m.Classrooms {
		if i.Classroom != nil {
			u.TariffEndAt = i.TariffEndAt
			u.TariffType = i.TariffType
			u.TariffName = GetTariffName(i.TariffType)
			if u.TariffEndAt != nil && u.TariffEndAt.Before(time.Now()) {
				u.TariffEndAt = nil
				u.TariffType = nil
				u.TariffName = nil
			}
			r := &ClassroomResponse{}
			r.FromModel(i.Classroom)
			u.Classrooms = append(u.Classrooms, r)
		}
	}
	if m.TeacherClassroom != nil {
		u.TeacherClassroom = &ClassroomResponse{}
		u.TeacherClassroom.FromModel(m.TeacherClassroom)
	}
	if m.PassportNumber != nil {
		u.PassportNumber = m.PassportNumber
	}
	if m.BirthCertNumber != nil {
		u.BirthCertNumber = m.BirthCertNumber
	}
	if m.ApplyNumber != nil {
		u.ApplyNumber = m.ApplyNumber
	}
	if m.WorkTitle != nil {
		u.WorkTitle = m.WorkTitle
	}
	if m.WorkPlace != nil {
		u.WorkPlace = m.WorkPlace
	}
	if m.District != nil {
		u.District = m.District
	}
	if m.Reference != nil {
		u.Reference = m.Reference
	}
	if m.NickName != nil {
		u.NickName = m.NickName
	}
	if m.EducationTitle != nil {
		u.EducationTitle = m.EducationTitle
	}
	if m.EducationPlace != nil {
		u.EducationPlace = m.EducationPlace
	}
	if m.EducationGroup != nil {
		u.EducationGroup = m.EducationGroup
	}
	if m.Documents != nil {
		u.Documents = m.Documents
	}
	if m.DocumentFiles != nil {
		u.DocumentFiles = &[]string{}
		for _, f := range *m.DocumentFiles {
			*u.DocumentFiles = append(*u.DocumentFiles, fileUrl(&f))
		}
	}
}

func (u *UserRequest) Format() {
	if u.Username != nil {
		*u.Username = strings.ToLower(*u.Username)
		*u.Username = strings.ReplaceAll(*u.Username, "ý", "y")
		*u.Username = strings.ReplaceAll(*u.Username, "ä", "a")
		*u.Username = strings.ReplaceAll(*u.Username, "ü", "u")
		*u.Username = strings.ReplaceAll(*u.Username, "ö", "o")
		*u.Username = strings.ReplaceAll(*u.Username, "ň", "n")
		*u.Username = strings.ReplaceAll(*u.Username, "ç", "c")
	}
	if u.FirstName != nil {
		*u.FirstName = strings.Trim(*u.FirstName, " ")
	}
	if u.LastName != nil {
		*u.LastName = strings.Trim(*u.LastName, " ")
	}
	if u.MiddleName != nil {
		*u.MiddleName = strings.Trim(*u.MiddleName, " ")
	}

	// if gender is not specified
	if u.Gender == nil && u.LastName != nil {
		lastName := *u.LastName
		if len(lastName) > 0 && strings.HasSuffix(lastName, "a") {
			gender := GenderFemale
			u.Gender = new(int)
			*u.Gender = int(gender)
		} else if strings.HasSuffix(lastName, "w") {
			gender := GenderMale
			u.Gender = new(int)
			*u.Gender = int(gender)
		}
	}
}

func (u *UserRequest) ToModel(m *User) {
	u.Format()
	if u.ID != nil {
		m.ID = *u.ID
	}
	m.FirstName = u.FirstName
	m.LastName = u.LastName
	m.MiddleName = u.MiddleName
	m.Username = u.Username
	m.Status = u.Status
	m.Password = u.Password
	m.Phone = u.Phone
	m.Email = u.Email
	if u.Birthday != nil {
		var err error
		m.Birthday, err = dateFormatter(*u.Birthday)
		if err != nil {
			m.Birthday = nil
		}
	}
	m.Gender = u.Gender
	m.Address = u.Address
	if u.Avatar != nil {
		m.Avatar = u.Avatar
	}
	if u.ChildIds != nil {
		m.Children = []*User{}
		for _, i := range *u.ChildIds {
			r := &User{ID: i}
			m.Children = append(m.Children, r)
		}
	}
	if u.ParentIds != nil {
		m.Parents = []*User{}
		for _, i := range *u.ParentIds {
			r := &User{ID: i}
			m.Parents = append(m.Parents, r)
		}
	}
	if u.SchoolIds != nil {
		m.Schools = []*UserSchool{}
		for _, i := range *u.SchoolIds {
			r := &UserSchool{SchoolUid: i.SchoolUid, RoleCode: *i.RoleCode}
			m.Schools = append(m.Schools, r)
		}
	}
	if u.Parents != nil {
		m.Parents = []*User{}
		for _, v := range *u.Parents {
			r := &User{}
			v.ToModel(r)
			m.Parents = append(m.Parents, r)
		}
	}
	if u.Children != nil {
		m.Children = []*User{}
		for _, v := range *u.Children {
			r := &User{}
			v.ToModel(r)
			m.Children = append(m.Children, r)
		}
	}
	if u.ClassroomIds != nil {
		m.Classrooms = []*UserClassroom{}
		for _, i := range *u.ClassroomIds {
			r := &UserClassroom{ClassroomId: *i.ClassroomId, Type: i.Type, TypeKey: i.TypeKey}
			m.Classrooms = append(m.Classrooms, r)
		}
	}
	if u.PassportNumber != nil {
		m.PassportNumber = u.PassportNumber
	}
	if u.BirthCertNumber != nil {
		m.BirthCertNumber = u.BirthCertNumber
	}
	if u.ApplyNumber != nil {
		m.ApplyNumber = u.ApplyNumber
	}
	if u.WorkTitle != nil {
		m.WorkTitle = u.WorkTitle
	}
	if u.WorkPlace != nil {
		m.WorkPlace = u.WorkPlace
	}
	if u.District != nil {
		m.District = u.District
	}
	if u.Reference != nil {
		m.Reference = u.Reference
	}
	if u.NickName != nil {
		m.NickName = u.NickName
	}
	if u.EducationTitle != nil {
		m.EducationTitle = u.EducationTitle
	}
	if u.EducationPlace != nil {
		m.EducationPlace = u.EducationPlace
	}
	if u.EducationGroup != nil {
		m.EducationGroup = u.EducationGroup
	}
	if u.Documents != nil {
		m.Documents = *u.Documents
	}
	if u.DocumentFiles != nil {
		m.DocumentFiles = u.DocumentFiles
	}
}

func dateFormatter(inputDate string) (date *time.Time, err error) {
	re := regexp.MustCompile("[0-9]+")
	numbers := re.FindAllString(inputDate, -1)
	isSlashed := strings.Contains(inputDate, "/")
	orderedDate := ""
	if len(numbers) == 3 {
		var year, month, day string
		if n, _ := strconv.Atoi(numbers[0]); n > 1900 {
			year = numbers[0]
			if len(year) == 2 {
				// Add "20" prefix if the year is <= 24
				if yearInt, _ := strconv.Atoi(year); yearInt <= 24 {
					year = "20" + year
				} else {
					year = "19" + year
				}
			}
			if isSlashed { // 1992/19/4
				month = fmt.Sprintf("%02s", numbers[2])
				day = fmt.Sprintf("%02s", numbers[1])
			} else { // 1992.04.19
				month = fmt.Sprintf("%02s", numbers[1])
				day = fmt.Sprintf("%02s", numbers[2])
			}
		} else {
			year = numbers[2]
			if len(year) == 2 {
				// Add "20" prefix if the year is <= 24
				if yearInt, _ := strconv.Atoi(year); yearInt <= 24 {
					year = "20" + year
				} else {
					year = "19" + year
				}
			}
			if isSlashed { // 4/19/92
				month = fmt.Sprintf("%02s", numbers[0])
				day = fmt.Sprintf("%02s", numbers[1])
			} else { // 04.09.1992
				month = fmt.Sprintf("%02s", numbers[1])
				day = fmt.Sprintf("%02s", numbers[0])
			}
		}
		orderedDate = fmt.Sprintf("%s-%s-%s", year, month, day)
	}
	date = &time.Time{}
	*date, err = time.Parse(time.DateOnly, orderedDate)
	if err != nil {
		return nil, err
	}
	return date, nil
}

func (u *UserProfileUpdateRequest) ToModel(m *User) {
	if u.FirstName != nil {
		m.FirstName = u.FirstName
	}
	if u.LastName != nil {
		m.LastName = u.LastName
	}
	if u.MiddleName != nil {
		m.MiddleName = u.MiddleName
	}
	if u.Username != nil {
		m.Username = u.Username
	}
	if u.Phone != nil {
		m.Phone = u.Phone
	}
	if u.Email != nil {
		m.Email = u.Email
	}
	if u.Birthday != nil {
		m.Birthday = &time.Time{}
		*m.Birthday, _ = time.Parse(time.DateOnly, *u.Birthday)
	}
	if u.Gender != nil {
		m.Gender = u.Gender
	}
	if u.Address != nil {
		m.Address = u.Address
	}
	if u.Avatar != nil {
		m.Avatar = u.Avatar
	}
}

func SerializeUsers(users []*User) []*UserResponse {
	usersResponse := []*UserResponse{}
	for _, user := range users {
		response := UserResponse{}
		response.FromModel(user)
		usersResponse = append(usersResponse, &response)
	}
	return usersResponse
}

type UserCollectionRequest struct {
	Users []UserRequest `json:"users" form:"users"`
}

type GetTeacherIdByNameQueryDto struct {
	SchoolId  string
	FirstName string
	LastName  *string
}
