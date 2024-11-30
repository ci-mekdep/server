package models

import (
	"math"
	"strconv"
	"time"
)

var (
	DefaultSuggestionsTeacher = []StringLocale{
		{
			Tm: "Bahalandyrmaklyga köp wagt gidýär, has tizleşdirmeli, aňsatlaşdyrmaly",
			Ru: "Bahalandyrmaklyga köp wagt gidýär, has tizleşdirmeli, aňsatlaşdyrmaly",
			En: "Bahalandyrmaklyga köp wagt gidýär, has tizleşdirmeli, aňsatlaşdyrmaly",
		},
		{
			Tm: "Tema ýazmak üçin ulanýarys, okuwçylar bilen maglumat paýlaşmagy giňeltmeli, aňsatlaşdyrmaly",
			Ru: "Tema ýazmak üçin ulanýarys, okuwçylar bilen maglumat paýlaşmagy giňeltmeli, aňsatlaşdyrmaly",
			En: "Tema ýazmak üçin ulanýarys, okuwçylar bilen maglumat paýlaşmagy giňeltmeli, aňsatlaşdyrmaly",
		},
		{
			Tm: "Ata-eneler bilen aragatnaşyk üçin peýdalanýarys, mümkinçilikleri köpeltmeli, has ýakynalşdyrmaly",
			Ru: "Ata-eneler bilen aragatnaşyk üçin peýdalanýarys, mümkinçilikleri köpeltmeli, has ýakynalşdyrmaly",
			En: "Ata-eneler bilen aragatnaşyk üçin peýdalanýarys, mümkinçilikleri köpeltmeli, has ýakynalşdyrmaly",
		},
		{
			Tm: "Temalary awtomat doldurýan etmeli",
			Ru: "Temalary awtomat doldurýan etmeli",
			En: "Temalary awtomat doldurýan etmeli",
		},
		{
			Tm: "Test almak mümkinçiligi goşmaly",
			Ru: "Test almak mümkinçiligi goşmaly",
			En: "Test almak mümkinçiligi goşmaly",
		},
		{
			Tm: "Öý işleri onlaýn barlamak mümkinçilik goşmaly",
			Ru: "Öý işleri onlaýn barlamak mümkinçilik goşmaly",
			En: "Öý işleri onlaýn barlamak mümkinçilik goşmaly",
		},
		{
			Tm: "Onlaýn okuwçylar bilen sorag-jogap mümkinçiligi goşmaly",
			Ru: "Onlaýn okuwçylar bilen sorag-jogap mümkinçiligi goşmaly",
			En: "Onlaýn okuwçylar bilen sorag-jogap mümkinçiligi goşmaly",
		},
	}

	DefaultSuggestionsParent = []StringLocale{
		{
			Tm: "Çagamyň gündeligini barlamak üçin ulanýaryn, okuw maglumatlaryň görnüşlerini has köpeltmeli",
			Ru: "Çagamyň gündeligini barlamak üçin ulanýaryn, okuw maglumatlaryň görnüşlerini has köpeltmeli",
			En: "Çagamyň gündeligini barlamak üçin ulanýaryn, okuw maglumatlaryň görnüşlerini has köpeltmeli",
		},
		{
			Tm: "Çagamyň analitikasyny barlamak üçin ulanýaryn, görkezijilerini, hasaplamalaryň görnüşlerini has köpeltmeli",
			Ru: "Çagamyň analitikasyny barlamak üçin ulanýaryn, görkezijilerini, hasaplamalaryň görnüşlerini has köpeltmeli",
			En: "Çagamyň analitikasyny barlamak üçin ulanýaryn, görkezijilerini, hasaplamalaryň görnüşlerini has köpeltmeli",
		},
		{
			Tm: "Mugallymlar bilen aragatnaşyk üçin peýdalanýarys, mümkinçilikleri köpeltmeli, has ýakynalşdyrmaly",
			Ru: "Mugallymlar bilen aragatnaşyk üçin peýdalanýarys, mümkinçilikleri köpeltmeli, has ýakynalşdyrmaly",
			En: "Mugallymlar bilen aragatnaşyk üçin peýdalanýarys, mümkinçilikleri köpeltmeli, has ýakynalşdyrmaly",
		},
		{
			Tm: "Analitikalary SMS görnüşde almagy goşmaly",
			Ru: "Analitikalary SMS görnüşde almagy goşmaly",
			En: "Analitikalary SMS görnüşde almagy goşmaly",
		},
		{
			Tm: "Okuwçydan test almak mümkinçiligi goşmaly",
			Ru: "Okuwçydan test almak mümkinçiligi goşmaly",
			En: "Okuwçydan test almak mümkinçiligi goşmaly",
		},
		{
			Tm: "Temalaryna degişli öwrenmek üçin widýo gollanmalary goşmaly",
			Ru: "Temalaryna degişli öwrenmek üçin widýo gollanmalary goşmaly",
			En: "Temalaryna degişli öwrenmek üçin widýo gollanmalary goşmaly",
		},
		{
			Tm: "Öwrenmek üçin kitaplary goşmaly",
			Ru: "Öwrenmek üçin kitaplary goşmaly",
			En: "Öwrenmek üçin kitaplary goşmaly",
		},
	}

	DefaultComplaintsParent = []StringLocale{
		{
			Tm: "Telefon belgim / maglumat nädogry ýa-da tapylanok",
			Ru: "Telefon belgim / maglumat nädogry ýa-da tapylanok",
			En: "Telefon belgim / maglumat nädogry ýa-da tapylanok",
		},
		{
			Tm: "SMS gelenok",
			Ru: "SMS gelenok",
			En: "SMS gelenok",
		},
		{
			Tm: "Ders rejäm nädogry görkezýär / ders tapylanok",
			Ru: "Ders rejäm nädogry görkezýär / ders tapylanok",
			En: "Ders rejäm nädogry görkezýär / ders tapylanok",
		},
		{
			Tm: "Žurnal senelerim nädogry görkezýär",
			Ru: "Žurnal senelerim nädogry görkezýär",
			En: "Žurnal senelerim nädogry görkezýär",
		},
		{
			Tm: "Žurnalda baha goýup bilemok, emma goýmak möhleti geçmedik",
			Ru: "Žurnalda baha goýup bilemok, emma goýmak möhleti geçmedik",
			En: "Žurnalda baha goýup bilemok, emma goýmak möhleti geçmedik",
		},
		{
			Tm: "Žurnallary doldurylan bolsada doldurmady diýýär",
			Ru: "Žurnallary doldurylan bolsada doldurmady diýýär",
			En: "Žurnallary doldurylan bolsada doldurmady diýýär",
		},
		{
			Tm: "Ders rejäm köne goýulan, täzelenmedik",
			Ru: "Ders rejäm köne goýulan, täzelenmedik",
			En: "Ders rejäm köne goýulan, täzelenmedik",
		},
		{
			Tm: "Žurnalda baha goýup bilemok, emma goýmak möhleti geçmedik",
			Ru: "Žurnalda baha goýup bilemok, emma goýmak möhleti geçmedik",
			En: "Žurnalda baha goýup bilemok, emma goýmak möhleti geçmedik",
		},
	}

	DefaultComplaintsTeacher = []StringLocale{
		{
			Tm: "Telefon belgim / maglumat nädogry ýa-da tapylanok",
			Ru: "Telefon belgim / maglumat nädogry ýa-da tapylanok",
			En: "Telefon belgim / maglumat nädogry ýa-da tapylanok",
		},
		{
			Tm: "SMS gelenok",
			Ru: "SMS gelenok",
			En: "SMS gelenok",
		},
		{
			Tm: "Çagalarym nädogry",
			Ru: "Çagalarym nädogry",
			En: "Çagalarym nädogry",
		},
		{
			Tm: "Çagalarym tapylanok",
			Ru: "Çagalarym tapylanok",
			En: "Çagalarym tapylanok",
		},
		{
			Tm: "Gündeliginde bahalary goýulanok",
			Ru: "Gündeliginde bahalary goýulanok",
			En: "Gündeliginde bahalary goýulanok",
		},
		{
			Tm: "Gündeliginde nädogry bahalar goýulýar",
			Ru: "Gündeliginde nädogry bahalar goýulýar",
			En: "Gündeliginde nädogry bahalar goýulýar",
		},
		{
			Tm: "Analitika nädogry hasaplaýar",
			Ru: "Analitika nädogry hasaplaýar",
			En: "Analitika nädogry hasaplaýar",
		},
	}
)

var StudentComments = []StudentComment{
	{
		Name: "Bilim",
		Types: []Type{
			{
				Type: "Gowy",
				Comments: []StringLocale{
					{
						Tm: "Çagaňyz okuwdan başarjaň! Tüweleme!",
						Ru: "Çagaňyz okuwdan başarjaň! Tüweleme!ru",
						En: "Çagaňyz okuwdan başarjaň! Tüweleme!en",
					},
					{
						Tm: "Çagaňyz okuwdan başarjaň! Tüweleme!",
						Ru: "Çagaňyz okuwda kyn mesele çözmegi başardy! Tüweleme!",
						En: "Çagaňyz okuwda kyn mesele çözmegi başardy! Tüweleme!",
					},
				},
			},
			{
				Type: "Erbet",
				Comments: []StringLocale{
					{
						Tm: "Çagaňyz okuwdan ýetişenok! Üns bermeli!",
						Ru: "Çagaňyz okuwdan ýetişenok! Üns bermeli!",
						En: "Çagaňyz okuwdan ýetişenok! Üns bermeli!",
					},
				},
			},
		},
	},
	{
		Name: "Terbiýe",
		Types: []Type{
			{
				Type: "Gowy",
				Comments: []StringLocale{
					{
						Tm: "Çagaňyz sapakda ýokary terbiýesini görkezdi, ata-ene sagbolsun!",
						Ru: "Çagaňyz sapakda ýokary terbiýesini görkezdi, ata-ene sagbolsun!",
						En: "Çagaňyz sapakda ýokary terbiýesini görkezdi, ata-ene sagbolsun!",
					},
				},
			},
			{
				Type: "Erbet",
				Comments: []StringLocale{
					{
						Tm: "Çagaňyz sapakda özüni terbiýesiz alyp barýar! Ata-ene çäre görüň!",
						Ru: "Çagaňyz sapakda özüni terbiýesiz alyp barýar! Ata-ene çäre görüň!",
						En: "Çagaňyz sapakda özüni terbiýesiz alyp barýar! Ata-ene çäre görüň!",
					},
				},
			},
		},
	},
	{
		Name: "Gatnaşygy",
		Types: []Type{
			{
				Type: "Gowy",
				Comments: []StringLocale{
					{
						Tm: "Çagaňyz sapakda köp soraglara jogap berýär, tüweleme!",
						Ru: "Çagaňyz sapakda köp soraglara jogap berýär, tüweleme!",
						En: "Çagaňyz sapakda köp soraglara jogap berýär, tüweleme!",
					},
				},
			},
			{
				Type: "Erbet",
				Comments: []StringLocale{
					{
						Tm: "Çagaňyz sapaklara höwesi ýok! Ata-ene çäre görüň!",
						Ru: "Çagaňyz sapaklara höwesi ýok! Ata-ene çäre görüň!",
						En: "Çagaňyz sapaklara höwesi ýok! Ata-ene çäre görüň!",
					},
				},
			},
		},
	},
}

var ContactVideosParent = []SettingVideos{
	{
		Title:    "1. eMekdep dolandyryş ulgamy",
		FileUrl:  "https://mekdep.edu.tm/uploads/videos/teacher/1.mp4",
		ImageUrl: "https://mekdep.edu.tm/uploads/videos/teacher/1.jpg",
	},
}
var ContactVideosAdmin = []SettingVideos{
	{
		Title:    "1. eMekdep dolandyryş ulgamy",
		FileUrl:  "https://mekdep.edu.tm/uploads/videos/teacher/1.mp4",
		ImageUrl: "https://mekdep.edu.tm/uploads/videos/teacher/1.jpg",
	},
}
var ContactVideosTeacher = []SettingVideos{
	{
		Title:    "1. eMekdep dolandyryş ulgamy",
		FileUrl:  "https://mekdep.edu.tm/uploads/videos/teacher/1.mp4",
		ImageUrl: "https://mekdep.edu.tm/uploads/videos/teacher/1.jpg",
	},
	{
		Title:    "2. Ulgama nädip girmeli?",
		FileUrl:  "https://mekdep.edu.tm/uploads/videos/teacher/1.mp4",
		ImageUrl: "https://mekdep.edu.tm/uploads/videos/teacher/1.jpg",
	},
	{
		Title:    "3. Mugallymyň iş sahypasy",
		FileUrl:  "https://mekdep.edu.tm/uploads/videos/teacher/2.mp4",
		ImageUrl: "https://mekdep.edu.tm/uploads/videos/teacher/2.jpg",
	},
	{
		Title:    "4. Elektron žurnaly barada we tertibi",
		FileUrl:  "https://mekdep.edu.tm/uploads/videos/teacher/3.mp4",
		ImageUrl: "https://mekdep.edu.tm/uploads/videos/teacher/3.jpg",
	},
	{
		Title:    "5. Beýlekiler menýuda goşmaça mümkinçilikleri",
		FileUrl:  "https://mekdep.edu.tm/uploads/videos/teacher/4.mp4",
		ImageUrl: "https://mekdep.edu.tm/uploads/videos/teacher/4.jpg",
	},
	{
		Title:    "6. Gurallar bölümi - mugallymyň kömekçisi",
		FileUrl:  "https://mekdep.edu.tm/uploads/videos/teacher/5.mp4",
		ImageUrl: "https://mekdep.edu.tm/uploads/videos/teacher/5.jpg",
	},
	{
		Title:    "7. Bildirişlerden habarly bolmak, resminamalary almak",
		FileUrl:  "https://mekdep.edu.tm/uploads/videos/teacher/6.mp4",
		ImageUrl: "https://mekdep.edu.tm/uploads/videos/teacher/6.jpg",
	},
}

var UserDocumentKeys = []string{"birth_certificate", "order_number", "passport"}

type Setting struct {
	General GeneralSetting `json:"general"`
	Lesson  LessonSetting  `json:"lesson"`
	Subject SubjectSetting `json:"subject"`
}
type SettingRequest struct {
	General GeneralSettingRequest `json:"general"`
}

type GeneralSetting struct {
	APIInstances               []APIInstance                  `json:"api_instances"`
	APIVersion                 *string                        `json:"api_version"`
	BankTypes                  []PaymentBank                  `json:"bank_types"`
	BookCategories             []string                       `json:"book_categories"`
	DefaultPeriod              *DefaultPeriod                 `json:"default_period"`
	Holidays                   []HolidaySetting               `json:"holidays"`
	MenuApps                   []MenuApps                     `json:"menu_apps"`
	ContactMessages            map[string][]map[string]string `json:"contact_messages"`
	Contact                    []SettingContactMessagesByRole `json:"contact"`
	MobileRequiredVersion      *int                           `json:"mobile_required_version"`
	AbsentUpdateMinutes        int                            `json:"absent_update_minutes"`
	AlertMessage               string                         `json:"alert_message"`
	LoginAlert                 *string                        `json:"login_alert"`
	DelayedGradeUpdateHours    int                            `json:"delayed_grade_update_hours"`
	GradeUpdateMinutes         int                            `json:"grade_update_minutes"`
	IsArchive                  bool                           `json:"is_archive"`
	TimetableUpdateCurrentWeek bool                           `json:"timetable_update_week_available"`
	ContactPhones              []ContactPhones                `json:"contact_phones"`
	IsForeignCountry           bool                           `json:"is_foreign_country"`
	UserDocumentKeys           []string                       `json:"user_document_keys"`
}

type GeneralSettingRequest struct {
	AbsentUpdateMinutes        *int    `json:"absent_update_minutes"`
	AlertMessage               *string `json:"alert_message"`
	DelayedGradeUpdateHours    *int    `json:"delayed_grade_update_hours"`
	GradeUpdateMinutes         *int    `json:"grade_update_minutes"`
	IsArchive                  *bool   `json:"is_archive"`
	TimetableUpdateCurrentWeek *bool   `json:"timetable_update_week_available"`
}

type LessonSetting struct {
	GradeReasons    []GradeReason    `json:"grade_reasons"`
	LessonTypes     []LessonType     `json:"lesson_types"`
	StudentComments []StudentComment `json:"student_comments"`
}

type SubjectSetting struct {
	BaseSubjectSetting []string         `json:"base_subjects"`
	ClassroomGroupKeys []string         `json:"classroom_group_keys"`
	SubjectSetting     []SubjectElement `json:"subjects"`
	TopicTags          []string         `json:"topic_tags"`
}

type APIInstance struct {
	Code string `json:"code"`
	Name string `json:"name"`
	Url  string `json:"url"`
}

type DefaultPeriod struct {
	CurrentNumber int        `json:"current_number"`
	Value         [][]string `json:"value"`
}

type FavouriteResources struct {
	Name StringLocale `json:"name"`
	Link string       `json:"link"`
	Icon string       `json:"icon"`
}

type HolidaySetting struct {
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
	Name      string    `json:"name"`
}

type MenuApps struct {
	Title StringLocale `json:"title"`
	Link  string       `json:"link"`
	Icon  string       `json:"icon"`
}

type GradeReason struct {
	Code string          `json:"code"`
	Name GradeReasonName `json:"name"`
}

type GradeReasonName struct {
	En string `json:"en"`
	Ru string `json:"ru"`
	Tm string `json:"tm"`
}

type LessonType struct {
	Code string         `json:"code"`
	Name LessonTypeName `json:"name"`
}

type ContactPhones struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type LessonTypeName struct {
	En string `json:"en"`
	Ru string `json:"ru"`
	Tm string `json:"tm"`
}

type StudentComment struct {
	Name  string `json:"name"`
	Types []Type `json:"types"`
}

type Type struct {
	Comments []StringLocale `json:"comments"`
	Type     string         `json:"type"`
}

type StringLocale struct {
	En string `json:"en"`
	Ru string `json:"ru"`
	Tm string `json:"tm"`
}

type SubjectElement struct {
	Code     string `json:"code"`
	FullName string `json:"full_name"`
	Color    string `json:"color"`
	Area     string `json:"area"`
	Name     string `json:"name"`
}

type SettingContactMessagesByRole struct {
	Role    Role                   `json:"role"`
	Contact SettingContactMessages `json:"contact"`
}
type SettingContactMessages struct {
	Complaints  []StringLocale  `json:"complaints"`
	Suggestions []StringLocale  `json:"suggestions"`
	Videos      []SettingVideos `json:"contact_videos"`
}

type SettingVideos struct {
	Title    string `json:"title"`
	FileUrl  string `json:"file_url"`
	ImageUrl string `json:"image_url"`
}

type Settings struct {
	ID        string     `json:"id"`
	Key       string     `json:"key"`
	Value     string     `json:"value"`
	UpdatedAt *time.Time `json:"updated_at"`
}

type PaymentSetting struct {
	MonthTypes  []int                   `json:"month_types"`
	TariffTypes []PaymentTariffResponse `json:"tariff_types"`
}

type Money string

func MoneyFromFloat(v float64) Money {
	m := strconv.Itoa(int(v * 100))
	return Money(m)
}

func (m Money) ToFloat64() (float64, error) {
	i, err := strconv.Atoi(string(m))
	if err != nil {
		return 0, err
	}
	return float64(i) / 100.0, nil // Convert cents to dollars
}

func (m Money) ToInt() int {
	i, err := strconv.Atoi(string(m))
	if err != nil {
		return 0
	}
	return i / 100.0
}

type HasRelationFields interface {
	RelationFields() []string
}

type DashboardSubjectsPercent struct {
	SubjectId        string     `json:"subject_id"`
	ClassroomId      string     `json:"classroom_id"`
	SubjectName      string     `json:"subject_name"`
	ClassroomName    string     `json:"classroom_name"`
	LessonDate       *time.Time `json:"lesson_date"`
	LessonTitle      *string    `json:"lesson_title"`
	AssignmentTitle  *string    `json:"assignment_title"`
	StudentsCount    int        `json:"students_count"`
	GradesCount      int        `json:"grades_count"`
	AbsentsCount     int        `json:"absents_count"`
	AbsentPercent    int        `json:"absent_percent"`
	GradeFullPercent int        `json:"grade_full_percent"`
	IsGradeFull      bool       `json:"is_grade_full"`
}

type DashboardSubjectsPercentBySchool struct {
	SchoolID         string  `json:"school_id"`
	LessonsCount     int     `json:"lessons_count"`
	TopicsCount      int     `json:"topics_count"`
	StudentsCount    int     `json:"students_count"`
	GradesCount      int     `json:"grades_count"`
	AbsentsCount     int     `json:"absents_count"`
	AbsentPercent    int     `json:"absent_percent"`
	GradeFullPercent float64 `json:"grade_full_percent"`
	IsGradeFull      bool    `json:"is_grade_full"`
	DaysCount        int     `json:"days_count"`
}

func (d *DashboardSubjectsPercent) SetOtherKeys() {
	if d.StudentsCount > 0 && d.GradesCount > 0 {
		d.GradeFullPercent = int(math.Floor(float64(d.GradesCount+d.AbsentsCount) * 100 / float64(d.StudentsCount)))
	}
	if d.StudentsCount > 0 && d.AbsentsCount > 0 {
		d.AbsentPercent = 100 - int(math.Floor(float64(d.AbsentsCount)*100/float64(d.StudentsCount)))
	} else {
		d.AbsentPercent = 100
	}
	if SubjectIsNoGrade(Subject{
		Name: &d.SubjectName,
		Classroom: &Classroom{
			Name: &d.ClassroomName,
		},
	}) {
		d.IsGradeFull = true
	} else {
		d.IsGradeFull = d.GradeFullPercent >= 10 && d.LessonTitle != nil && len(*d.LessonTitle) > 2
	}
}

func (d *DashboardSubjectsPercentBySchool) SetOtherKeys() {
	if d.StudentsCount > 0 && d.GradesCount > 0 {
		d.GradeFullPercent = float64(d.TopicsCount) * 100 / float64(d.LessonsCount)
		d.GradeFullPercent = math.Round(d.GradeFullPercent)
		if d.GradeFullPercent > 100 {
			d.GradeFullPercent = 100
		}
	}
	if d.StudentsCount > 0 && d.AbsentsCount > 0 && d.DaysCount > 0 {
		totalStudentsbyDays := d.StudentsCount * d.DaysCount
		d.AbsentPercent = 100 - int(math.Floor(float64(d.AbsentsCount)/float64(totalStudentsbyDays)*100))
	} else {
		d.AbsentPercent = 100
	}
	d.IsGradeFull = d.GradeFullPercent >= 10
}

type ContactItemsCount struct {
	SchoolCode         string `json:"school_code"`
	TotalCount         int    `json:"total_count"`
	ReviewCount        int    `json:"review_count"`
	ComplaintCount     int    `json:"complaint_count"`
	SuggestionCount    int    `json:"suggestion_count"`
	DataComplaintCount int    `json:"data_complaint_count"`
}

type PaymentTransactionsCount struct {
	SchoolCode       string `json:"school_code"`
	TotalCount       int    `json:"total_count"`
	HalkbankCount    int    `json:"halkbank_count"`
	SenagatbankCount int    `json:"senagatbank_count"`
	RysgalbankCount  int    `json:"rysgalbank_count"`
	TfebCount        int    `json:"tfeb_count"`
}

type DashboardUsersCount struct {
	SchoolId                 *string `json:"school_id"`
	SchoolsCount             int     `json:"schools_count"`
	StudentsCount            int     `json:"students_count"`
	ParentsCount             int     `json:"parents_count"`
	TeachersCount            int     `json:"teachers_count"`
	PrincipalsCount          int     `json:"principals_count"`
	OrganizationsCount       int     `json:"organizations_count"`
	ClassroomsCount          int     `json:"classrooms_count"`
	TimetablesCount          int     `json:"timetables_count"`
	UsersOnlineCount         int     `json:"users_online_count"`
	StudentsWithParentsCount int     `json:"students_with_parents_count"`
}

type DashboardUsersCountByClassroom struct {
	Classroom          Classroom `json:"classroom"`
	SchoolId           string    `json:"school_id"`
	StudentsCount      int       `json:"students_count"`
	ParentsCount       int       `json:"parents_count"`
	ParentsOnlineCount int       `json:"parents_online_count"`
	ParentsPaidCount   int       `json:"parents_paid_count"`
}

type SchoolStudentsCount struct {
	SchoolCode string  `json:"school_code"`
	Counts     [][]int `json:"counts"`
}

var DefaultHolidays = []HolidaySetting{
	{
		StartDate: time.Date(2000, 1, 1, 0, 0, 0, 0, time.Local),
		EndDate:   time.Date(2000, 1, 1, 0, 0, 0, 0, time.Local),
		Name:      "Täze ýyl baýramy",
	},
	{
		StartDate: time.Date(2000, 3, 8, 0, 0, 0, 0, time.Local),
		EndDate:   time.Date(2000, 3, 8, 0, 0, 0, 0, time.Local),
		Name:      "8-nji mart - Halkara zenanlar güni",
	},
	{
		StartDate: time.Date(2000, 3, 21, 0, 0, 0, 0, time.Local),
		EndDate:   time.Date(2000, 3, 22, 0, 0, 0, 0, time.Local),
		Name:      "Milli bahar baýramy",
	},
	{
		StartDate: time.Date(2000, 3, 23, 0, 0, 0, 0, time.Local),
		EndDate:   time.Date(2000, 3, 28, 0, 0, 0, 0, time.Local),
		Name:      "Dynç alyş möwsümi",
	},
	{
		StartDate: time.Date(2000, 4, 10, 0, 0, 0, 0, time.Local),
		EndDate:   time.Date(2000, 4, 10, 0, 0, 0, 0, time.Local),
		Name:      "Oraza baýramy",
	},
	{
		StartDate: time.Date(2000, 5, 18, 0, 0, 0, 0, time.Local),
		EndDate:   time.Date(2000, 5, 18, 0, 0, 0, 0, time.Local),
		Name:      "Türkmenistanyň Konstitusiýasynyň we Türkmenistanyň Döwlet baýdagynyň güni",
	},
	{
		StartDate: time.Date(2000, 9, 27, 0, 0, 0, 0, time.Local),
		EndDate:   time.Date(2000, 9, 27, 0, 0, 0, 0, time.Local),
		Name:      "Türkmenistanyň Garaşsyzlyk güni",
	},
	{
		StartDate: time.Date(2000, 12, 12, 0, 0, 0, 0, time.Local),
		EndDate:   time.Date(2000, 12, 12, 0, 0, 0, 0, time.Local),
		Name:      "Bitaraplyk baýramy",
	},
}
