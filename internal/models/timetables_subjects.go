package models

import (
	"strings"
	"time"
)

type SubjectArea string

const SubjectAreaExact SubjectArea = "Takyk"
const SubjectAreaNatural SubjectArea = "Tebigy"
const SubjectAreaHumanity SubjectArea = "Ynsanperwer"
const SubjectAreaSpeciality SubjectArea = "Hunar"

var DefaultSubjectAreas = []SubjectArea{
	SubjectAreaExact,
	SubjectAreaNatural,
	SubjectAreaHumanity,
}

// TODO: fill subject areas
var DefaultSubjects = [][]string{
	{"Algebra", "Algebra", string(SubjectAreaExact), "#3E5F8A"},
	{"Astronomiýa", "Astronomiýa", string(SubjectAreaNatural), "#C7B446"},
	{"Aýdym-saz", "Aýdym-saz", string(SubjectAreaHumanity), "#78ab68"},
	{"Bedenterbiýe", "Bedenterbiýe", string(SubjectAreaNatural), "#C7B446"},
	{"Biologiýa", "Biologiýa", string(SubjectAreaNatural), "#C7B446"},
	{"Çeper. zähmeti", "Çeperçilik zähmeti", string(SubjectAreaHumanity), "#DC9D00"},
	{"Döwr. Tehno. E.", "Döwrebap tehnologiýanyň esaslary", string(SubjectAreaExact), "#3E5F8A"},
	{"Dünýä med.", "Dünýä medeniýeti", string(SubjectAreaHumanity), "#C1876B"},
	{"Dünýä taryhy", "Dünýä taryhy", string(SubjectAreaHumanity), "#C1876B"},
	{"Daşary ýurt dili", "Daşary ýurt dili", string(SubjectAreaHumanity), "#78ab68"},
	{"Daşary ýurt dili (garyşyk)", "Daşary ýurt dili (garyşyk)", string(SubjectAreaHumanity), "#78ab68"},
	{"Edebiýat", "Edebiýat", string(SubjectAreaHumanity), "#78ab68"},
	{"Ekologiýa", "Ekologiýa", string(SubjectAreaNatural), "#C7B446"},
	{"Ählumumy taryh", "Ählumumy taryh", string(SubjectAreaHumanity), "#C1876B"},
	{"Fizika", "Fizika", string(SubjectAreaNatural), "#C7B446"},
	{"Geografiýa", "Geografiýa", string(SubjectAreaNatural), "#C7B446"},
	{"Geometriýa", "Geometriýa", string(SubjectAreaExact), "#3E5F8A"},
	{"Himiýa", "Himiýa", string(SubjectAreaNatural), "#C7B446"},
	{"Hor (aýdym-saz)", "Hor (aýdym-saz)", string(SubjectAreaHumanity), "#C7B446"},
	{"Informatika", "Informatikanyň esaslary", string(SubjectAreaExact), "#3E5F8A"},
	{"Informatika", "Informatika", string(SubjectAreaExact), "#3E5F8A"},
	{"IKIT", "Informasiýa-kommunikasiýa we innowasion tehnologiýalar", string(SubjectAreaExact), "#3E5F8A"},
	{"Iňlis dili", "Iňlis dili", string(SubjectAreaHumanity), "#C1876B"},
	{"Jemgyýet", "Jemgyýeti öwreniş", string(SubjectAreaHumanity), "#C1876B"},
	{"Matematika", "Matematika", string(SubjectAreaExact), "#3E5F8A"},
	{"Model we G.", "Modelirleme we grafika", string(SubjectAreaExact), "#3E5F8A"},
	{"Ene dili", "Ene dili", string(SubjectAreaHumanity), "#2D572C"},
	{"Okuw", "Okuw", string(SubjectAreaHumanity), "#2D572C"},
	{"Özüňi alyp B.M.", "Özüňi alyp barmagyň medeniýeti", string(SubjectAreaHumanity), "#2D572C"},
	{"Proýek. Esas.", "Proýektirlemegiň esaslary", string(SubjectAreaExact), "#3E5F8A"},
	{"Rus dili", "Rus dili", string(SubjectAreaHumanity), "#78ab68"},
	{"Rus edebiýaty", "Rus edebiýaty", string(SubjectAreaHumanity), "#C1876B"},
	{"Türkmen edebiýaty", "Türkmen edebiýaty", string(SubjectAreaHumanity), "#C1876B"},
	{"Şekil. Sun.", "Şekillendiriş sungaty", string(SubjectAreaHumanity), "#DC9D00"},
	{"Surat", "Surat", string(SubjectAreaHumanity), "#DC9D00"},
	{"Solfedžio (aýdym-saz)", "Solfedžio (aýdym-saz)", string(SubjectAreaHumanity), "#DC9D00"},
	{"Saz edebiýaty", "Saz edebiýaty", string(SubjectAreaHumanity), "#DC9D00"},
	{"Tebigat", "Tebigaty öwreniş", string(SubjectAreaNatural), "#C7B446"},
	{"Türkmen dili", "Türkmen dili", string(SubjectAreaHumanity), "#78ab68"},
	{"Hukuk esaslary", "Türkmenistanyň Döwlet we hukuk esaslary", string(SubjectAreaHumanity), "#2D572C"},
	{"Medeni miras", "Türkmenistanyň medeni mirasy", string(SubjectAreaHumanity), "#C1876B"},
	{"T-nyň taryhy", "Türkmenistanyň taryhy", string(SubjectAreaHumanity), "#C1876B"},
	{"Ykdysadyýet", "Ykdysadyýetiň esaslary", string(SubjectAreaHumanity), "#2D572C"},
	{"Ýaşaýyş D.E.", "Ýaşaýyş durmuş esaslary", string(SubjectAreaHumanity), "#2D572C"},
	{"Ýazuw", "Ýazuw", string(SubjectAreaHumanity), "#2D572C"},
	{"Zähmet", "Zähmet", string(SubjectAreaHumanity), "#DC9D00"},
	{"Durmuş zähmeti", "Durmuş zähmeti", string(SubjectAreaHumanity), "#DC9D00"},
	{"Çyzuw", "Çyzuw", string(SubjectAreaHumanity), "#DC9D00"},
	{"Synp sagady", "Synp sagady", string(SubjectAreaHumanity), "#2D572C"},
	{"Nemes dili", "Nemes dili", string(SubjectAreaHumanity), "#78ab68"},
	{"Fransuz dili", "Fransuz dili", string(SubjectAreaHumanity), "#78ab68"},
	{"Ýapon dili", "Ýapon dili", string(SubjectAreaHumanity), "#78ab68"},
	{"Hytaý dili", "Hytaý dili", string(SubjectAreaHumanity), "#78ab68"},
	{"Pars dili", "Pars dili", string(SubjectAreaHumanity), "#78ab68"},
	{"Arap dili", "Arap dili", string(SubjectAreaHumanity), "#78ab68"},
	{"Koreý dili", "Koreý dili", string(SubjectAreaHumanity), "#78ab68"},
	{"Italýan dili", "Italýan dili", string(SubjectAreaHumanity), "#78ab68"},
	{"Hünär ders", "Hünär ders", string(SubjectAreaHumanity), "#78ab68"},
	{"Ýörite ders 1", "Ýörite ders 1", string(SubjectAreaHumanity), "#78ab68"},
	{"Ýörite ders 2", "Ýörite ders 2", string(SubjectAreaHumanity), "#78ab68"},
	{"Ders amaly okuw", "Ders amaly okuw", string(SubjectAreaHumanity), "#78ab68"},
	{"Sözleýiş dilini ösdürmek", "Sözleýiş dilini ösdürmek", string(SubjectAreaHumanity), "#78ab68"},
	{"Matematika we konstruirleme", "Matematika we konstruirleme", string(SubjectAreaExact), "#3E5F8A"},
	{"Menin Watanym", "Menin Watanym", string(SubjectAreaHumanity), "#78ab68"},
	{"Aýdym we sazlaşykly hereket okuwlary", "Aýdym we sazlaşykly hereket okuwlary", string(SubjectAreaHumanity), "#78ab68"},
	{"Önümçilik zähmeti", "Önümçilik zähmeti", string(SubjectAreaHumanity), "#78ab68"},
	{"Durmuş we önümçilik zähmeti", "Durmuş we önümçilik zähmeti", string(SubjectAreaHumanity), "#78ab68"},
	{"Gurluşyk ugurly hünärler", "Gurluşyk ugurly hünärler", string(SubjectAreaSpeciality), "#57ec9a"},
	{"Agaç ussasy", "Agaç ussasy", string(SubjectAreaSpeciality), "#57ec9a"},
	{"Çyzgyçy", "Çyzgyçy", string(SubjectAreaSpeciality), "#57ec9a"},
	{"Suwagçy-reňkleýji", "Suwagçy-reňkleýji", string(SubjectAreaSpeciality), "#57ec9a"},
	{"Mebel ýygnaýjy", "Mebel ýygnaýjy", string(SubjectAreaSpeciality), "#57ec9a"},
	{"Plastmas önümleri kebşirleýji", "Plastmas önümleri kebşirleýji", string(SubjectAreaSpeciality), "#57ec9a"},
	{"Betonçy-kerpiç örüji", "Betonçy-kerpiç örüji", string(SubjectAreaSpeciality), "#57ec9a"},
	{"Biçimçi-tikinçi", "Biçimçi-tikinçi", string(SubjectAreaSpeciality), "#57ec9a"},
	{"Halyçy", "Halyçy", string(SubjectAreaSpeciality), "#57ec9a"},
	{"El örüm işleri", "El örüm işleri", string(SubjectAreaSpeciality), "#57ec9a"},
	{"Keşdeçi", "Keşdeçi", string(SubjectAreaSpeciality), "#57ec9a"},
	{"Biçimçi", "Biçimçi", string(SubjectAreaSpeciality), "#57ec9a"},
	{"Tikinçi", "Tikinçi", string(SubjectAreaSpeciality), "#57ec9a"},
	{"Keşdeçi (el we maşyn işleri)", "Keşdeçi (el we maşyn işleri)", string(SubjectAreaSpeciality), "#57ec9a"},
	{"Dokmaçy", "Dokmaçy", string(SubjectAreaSpeciality), "#57ec9a"},
	{"Aýakgap tikýän ussa", "Aýakgap tikýän ussa", string(SubjectAreaSpeciality), "#57ec9a"},
	{"Jorap önümlerini örüji", "Jorap önümlerini örüji", string(SubjectAreaSpeciality), "#57ec9a"},
	{"Aşpez", "Aşpez", string(SubjectAreaSpeciality), "#57ec9a"},
	{"Satyjy", "Satyjy", string(SubjectAreaSpeciality), "#57ec9a"},
	{"Aşpez-konditer", "Aşpez-konditer", string(SubjectAreaSpeciality), "#57ec9a"},
	{"Çörek-bulka önümlerini taýýarlaýjy", "Çörek-bulka önümlerini taýýarlaýjy", string(SubjectAreaSpeciality), "#57ec9a"},
	{"Konditer", "Konditer", string(SubjectAreaSpeciality), "#57ec9a"},
	{"Gülçi", "Gülçi", string(SubjectAreaSpeciality), "#57ec9a"},
	{"Bagban-nahal oturdyjy", "Bagban-nahal oturdyjy", string(SubjectAreaSpeciality), "#57ec9a"},
	{"Gülçi-bagban (nahal oturdyjy)", "Gülçi-bagban (nahal oturdyjy)", string(SubjectAreaSpeciality), "#57ec9a"},
	{"Dizaýner-gülçi", "Dizaýner-gülçi", string(SubjectAreaSpeciality), "#57ec9a"},
	{"Ekerançy", "Ekerançy", string(SubjectAreaSpeciality), "#57ec9a"},
	{"Mülkdar-maldar", "Mülkdar-maldar", string(SubjectAreaSpeciality), "#57ec9a"},
	{"Maldar", "Maldar", string(SubjectAreaSpeciality), "#57ec9a"},
	{"Gök ekerançy", "Gök ekerançy", string(SubjectAreaSpeciality), "#57ec9a"},
	{"Arybalçy", "Arybalçy", string(SubjectAreaSpeciality), "#57ec9a"},
	{"Mülkdar-ekerançy", "Mülkdar-ekerançy", string(SubjectAreaSpeciality), "#57ec9a"},
	{"Fermeçilik hojalygynyň operatory", "Fermeçilik hojalygynyň operatory", string(SubjectAreaSpeciality), "#57ec9a"},
	{"Oba hojalyk traktor-maşynlaryny abatlaýjy", "Oba hojalyk traktor-maşynlaryny abatlaýjy slesar", string(SubjectAreaSpeciality), "#57ec9a"},
	{"Buhgalter hasabynyň esaslary", "Buhgalter hasabynyň esaslary", string(SubjectAreaSpeciality), "#57ec9a"},
	{"Buhgalter hasaby we telekeçilik", "Buhgalter hasabynyň we telekeçiligiň esaslary", string(SubjectAreaSpeciality), "#57ec9a"},
	{"Menejer", "Menejer", string(SubjectAreaSpeciality), "#57ec9a"},
	{"Telekeçilik ulgamynyň menejmenti", "Telekeçilik ulgamynyň menejmenti", string(SubjectAreaSpeciality), "#57ec9a"},
	{"Telekeçilik işjeňligine giriş", "Telekeçilik işjeňligine giriş", string(SubjectAreaSpeciality), "#57ec9a"},
	{"Sanly bilimiň esaslary", "Sanly bilimiň esaslary", string(SubjectAreaSpeciality), "#57ec9a"},
	{"Sanly ykdysadyýetiň esaslary", "Sanly ykdysadyýetiň esaslary", string(SubjectAreaSpeciality), "#57ec9a"},
	{"Kätip-çap ediji", "Kätip-çap ediji", string(SubjectAreaSpeciality), "#57ec9a"},
	{"Kätip, kompýuteriň operatory", "Kätip, kompýuteriň operatory", string(SubjectAreaSpeciality), "#57ec9a"},
	{"Elektron resminama dolanyşygy", "Elektron resminama dolanyşygy", string(SubjectAreaSpeciality), "#57ec9a"},
	{"Suratkeş", "Suratkeş", string(SubjectAreaSpeciality), "#57ec9a"},
	{"Fotosuratçy", "Fotosuratçy", string(SubjectAreaSpeciality), "#57ec9a"},
	{"Bezegçi (oformitel)", "Bezegçi (oformitel)", string(SubjectAreaSpeciality), "#57ec9a"},
	{"Toý dabaralarynyň dizaýneri", "Toý dabaralarynyň dizaýneri", string(SubjectAreaSpeciality), "#57ec9a"},
	{"Gid-terjimeçi (syýahatçylyk)", "Gid-terjimeçi (syýahatçylyk pudagynda)", string(SubjectAreaSpeciality), "#57ec9a"},
	{"Teleradio apparaturalary bejeriji", "Teleradio apparaturalary bejeriji", string(SubjectAreaSpeciality), "#57ec9a"},
	{"Milli saz gurallaryny ýasaýjy", "Milli saz gurallaryny ýasaýan ussasy", string(SubjectAreaSpeciality), "#57ec9a"},
	{"Elektrik enjamlary abatlaýjy", "Elektrik enjamlary abatlaýjy slesar", string(SubjectAreaSpeciality), "#57ec9a"},
	{"Öý hojalyk abzallaryny abatlaýjy", "Öý hojalyk elektroabzallaryny abatlaýjy slesar", string(SubjectAreaSpeciality), "#57ec9a"},
	{"Kompýuter sowatlylygy", "Kompýuter sowatlylygynyň esaslary", string(SubjectAreaSpeciality), "#57ec9a"},
	{"Kompýuteriň operatory", "Kompýuteriň operatory", string(SubjectAreaSpeciality), "#57ec9a"},
	{"Kompýuter tehnikasyny bejeriji", "Kompýuter tehnikasyny bejeriji we hyzmat ediji ussa", string(SubjectAreaSpeciality), "#57ec9a"},
	{"Programmirlemegiň esaslary", "Programmirlemegiň esaslary", string(SubjectAreaSpeciality), "#57ec9a"},
	{"Kiberhowpsuzlygyň esaslary", "Kiberhowpsuzlygyň esaslary", string(SubjectAreaSpeciality), "#57ec9a"},
}

type Subject struct {
	ID               string         `json:"id"`
	SchoolId         string         `json:"school_id"`
	ClassroomId      string         `json:"classroom_id"`
	ClassroomType    *string        `json:"classroom_type"`
	ClassroomTypeKey *int           `json:"classroom_type_key"`
	BaseSubjectId    *string        `json:"base_subject_id"`
	ParentId         *string        `json:"parent_id"`
	TeacherId        *string        `json:"teacher_id"`
	SecondTeacherId  *string        `json:"second_teacher_id"`
	Name             *string        `json:"name"`
	FullName         *string        `json:"full_name"`
	WeekHours        *uint          `json:"week_hours"`
	UpdatedAt        *time.Time     `json:"updated_at"`
	CreatedAt        *time.Time     `json:"created_at"`
	School           *School        `json:"school"`
	Classroom        *Classroom     `json:"classroom"`
	Teacher          *User          `json:"teacher"`
	SecondTeacher    *User          `json:"second_teacher"`
	Exams            []*SubjectExam `json:"exams"`
	Parent           *Subject       `json:"parent"`
	Children         []*Subject     `json:"children"`
	BaseSubject      *BaseSubjects  `json:"base_subject"`
}

func (Subject) RelationFields() []string {
	return []string{"School", "Classroom", "Teacher", "SecondTeacher", "Exams", "Parent", "Children", "BaseSubject"}
}

func (s Subject) GetId() string {
	if s.ParentId != nil {
		return *s.ParentId
	}
	return s.ID
}

func (s Subject) IsTeacherEq(teacherId string) bool {
	return s.SecondTeacherId == nil && s.TeacherId != nil && *s.TeacherId == teacherId ||
		s.SecondTeacherId != nil && *s.SecondTeacherId == teacherId
}

func (s Subject) BelongsToTeacher(t *User) bool {
	return s.SecondTeacherId == nil && *s.TeacherId == t.ID ||
		s.SecondTeacherId != nil && *s.SecondTeacherId == t.ID
}

type SubjectFilterRequest struct {
	ID               *string   `form:"id"`
	NotIds           *[]string `form:"not_ids"`
	Ids              *[]string `form:"ids"`
	SchoolId         *string   `form:"school_id"`
	SchoolIds        []string  `form:"school_ids[]"`
	TeacherId        *string   `form:"teacher_id"`
	TeacherIds       []string  `form:"teacher_ids[]"`
	ClassroomId      *string   `form:"classroom_id"`
	ClassroomIds     []string  `form:"classroom_ids[]"`
	BaseSubjectId    *string   `json:"base_subject_id"`
	ClassroomTypeKey *int      `form:"classroom_type_key"`
	SubjectNames     []string  `form:"subject_names[]"`
	WeekHours        *int      `form:"week_hours"`
	WeekHoursRange   []int     `form:"week_hours[]"`
	IsSecondTeacher  *bool     `form:"is_second_teacher"`
	IsSubjectExam    *bool     `form:"is_subject_exam"`
	Search           *string   `form:"search"`
	Sort             *string   `form:"sort"`
	PaginationRequest
}

type SubjectResponse struct {
	ID               string                 `json:"id"`
	Name             *string                `json:"name"`
	FullName         *string                `json:"full_name"`
	WeekHours        *uint                  `json:"week_hours"`
	UpdatedAt        *time.Time             `json:"updated_at"`
	CreatedAt        *time.Time             `json:"created_at"`
	ClassroomType    *string                `json:"classroom_type"`
	ClassroomTypeKey *int                   `json:"classroom_type_key"`
	ParentId         *string                `json:"parent_id"`
	School           *SchoolResponse        `json:"school"`
	Classroom        *ClassroomResponse     `json:"classroom"`
	Teacher          *UserResponse          `json:"teacher"`
	SecondTeacher    *UserResponse          `json:"second_teacher"`
	Exams            []*SubjectExamResponse `json:"exams"`
	Parent           *SubjectResponse       `json:"parent"`
	Children         []*SubjectResponse     `json:"children"`
	BaseSubject      *BaseSubjectsResponse  `json:"base_subject"`
}

type SubjectRequest struct {
	ID               *string               `json:"id"`
	SchoolId         string                `json:"school_id"`
	ClassroomId      string                `json:"classroom_id"`
	ClassroomType    *string               `json:"classroom_type"`
	ClassroomTypeKey *int                  `json:"classroom_type_key"`
	TeacherId        *string               `json:"teacher_id"`
	SecondTeacherId  *string               `json:"second_teacher_id"`
	BaseSubjectId    *string               `json:"base_subject_id"`
	Name             *string               `json:"name"`
	FullName         *string               `json:"full_name"`
	WeekHours        *uint                 `json:"week_hours"`
	Exams            []*SubjectExamRequest `json:"exams"`
}

type CreateSubjectByNamesRequestDto struct {
	Name                string  `json:"name" form:"name" validate:"required"`
	FullName            string  ``
	ClassroomName       string  `json:"classroom_name" form:"classroom_name" validate:"required"`
	ClassroomTypeKeyStr *string `json:"classroom_type_key" form:"classroom_type_key"`
	ClassroomTypeKey    *uint   `json:"classroom_type_key" form:"classroom_type_key"`
	TeacherFullName     *string `json:"teacher_full_name" form:"teacher_full_name"`
	WeekHours           *uint   ``
	WeekHoursStr        *string `json:"week_hours" form:"week_hours"`
}

type CreateSubjectsByNamesRequestDto struct {
	Subjects []CreateSubjectByNamesRequestDto `json:"subjects" form:"subjects"`
}

type SubjectValueResponse struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func (r *SubjectResponse) ToValues() SubjectValueResponse {
	value := ""
	if r.Name != nil && r.Teacher != nil && r.Teacher.Username != nil && r.Classroom != nil && r.Classroom.Name != nil {
		firstNameInitial := ""
		if r.Teacher.FirstName != nil && len(*r.Teacher.FirstName) > 0 {
			firstNameInitial = string((*r.Teacher.FirstName)[0])
		}
		value += " " + *r.Classroom.Name + ", " + *r.Name + ", " + firstNameInitial + ". " + *r.Teacher.LastName
	}
	return SubjectValueResponse{
		Key:   r.ID,
		Value: strings.Trim(value, " "),
	}
}

func (r *SubjectResponse) FromModel(m *Subject) {
	r.ID = m.ID
	r.ParentId = m.ParentId
	r.Name = m.Name
	r.FullName = m.FullName
	r.WeekHours = m.WeekHours
	r.ClassroomType = m.ClassroomType
	r.ClassroomTypeKey = m.ClassroomTypeKey
	if m.School != nil {
		r.School = &SchoolResponse{}
		r.School.FromModel(m.School)
	}
	if m.Classroom != nil {
		r.Classroom = &ClassroomResponse{}
		r.Classroom.FromModel(m.Classroom)
	}
	if m.Teacher != nil {
		r.Teacher = &UserResponse{}
		r.Teacher.FromModel(m.Teacher)
	}
	if m.SecondTeacher != nil {
		r.SecondTeacher = &UserResponse{}
		r.SecondTeacher.FromModel(m.SecondTeacher)
	}
	if m.Parent != nil {
		r.Parent = &SubjectResponse{}
		r.Parent.FromModel(m.Parent)
	}
	if m.Exams != nil {
		r.Exams = []*SubjectExamResponse{}
		for _, v := range m.Exams {
			se := SubjectExamResponse{}
			se.FromModel(v)
			r.Exams = append(r.Exams, &se)
		}
	}
	r.Children = []*SubjectResponse{}
	for _, i := range m.Children {
		u := &SubjectResponse{}
		u.FromModel(i)
		r.Children = append(r.Children, u)
	}
	if m.BaseSubject != nil {
		r.BaseSubject = &BaseSubjectsResponse{}
		r.BaseSubject.FromModel(m.BaseSubject)
	}
	r.CreatedAt = m.CreatedAt
	r.UpdatedAt = m.UpdatedAt
}

func (r *SubjectRequest) ToModel(m *Subject) {
	if r.ID == nil {
		r.ID = new(string)
	}
	m.ID = *r.ID
	m.Name = r.Name
	m.FullName = r.FullName
	m.WeekHours = r.WeekHours
	m.SchoolId = r.SchoolId
	m.ClassroomId = r.ClassroomId
	m.ClassroomType = r.ClassroomType
	m.ClassroomTypeKey = r.ClassroomTypeKey
	m.TeacherId = r.TeacherId
	m.SecondTeacherId = r.SecondTeacherId
	if r.Exams != nil {
		m.Exams = []*SubjectExam{}
		for _, se := range r.Exams {
			exam := &SubjectExam{}
			se.ToModel(exam)
			m.Exams = append(m.Exams, exam)
		}
	}
	m.BaseSubjectId = r.BaseSubjectId
}
