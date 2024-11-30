package models

import "time"

type Book struct {
	ID             string     `json:"id"`
	Title          *string    `json:"title"`
	Categories     *[]string  `json:"categories"`
	Description    *string    `json:"description"`
	Year           *int       `json:"year"`
	Pages          *int       `json:"pages"`
	Authors        *[]string  `json:"authors"`
	File           *string    `json:"file"`
	FileSize       *int       `json:"file_size"`
	FilePreview    *string    `json:"file_preview"`
	IsDownloadable bool       `json:"is_downloadable"`
	CreatedAt      *time.Time `json:"created_at"`
	UpdatedAt      *time.Time `json:"updated_at"`
}

func (Book) RelationFields() []string {
	return []string{}
}

type BookRequest struct {
	ID             string    `json:"id" form:"id"`
	Title          *string   `json:"title" form:"title"`
	Categories     *[]string `json:"categories" form:"categories"`
	Description    *string   `json:"description" form:"description"`
	Year           *int      `json:"year" form:"year"`
	Pages          *int      `json:"pages" form:"pages"`
	Authors        *[]string `json:"authors" form:"authors"`
	File           *string   ``
	FileSize       *int      `json:"file_size" form:"file_size"`
	FilePreview    *string   ``
	IsDownloadable bool      `json:"is_downloadable" form:"is_downloadable"`
}

type BookResponse struct {
	ID             string     `json:"id"`
	Title          *string    `json:"title"`
	Categories     *[]string  `json:"categories"`
	Description    *string    `json:"description"`
	Year           *int       `json:"year"`
	Pages          *int       `json:"pages"`
	Authors        *[]string  `json:"authors"`
	File           *string    `json:"file"`
	FileSize       *int       `json:"file_size"`
	FilePreview    *string    `json:"file_preview"`
	IsDownloadable bool       `json:"is_downloadable"`
	CreatedAt      *time.Time `json:"created_at"`
	UpdatedAt      *time.Time `json:"updated_at"`
}

func (r *BookResponse) FromModel(m *Book) error {
	r.ID = m.ID
	if m.Title != nil {
		r.Title = m.Title
	}
	if m.Categories != nil {
		r.Categories = m.Categories
	}
	if m.Description != nil {
		r.Description = m.Description
	}
	if m.Year != nil {
		r.Year = m.Year
	}
	if m.Pages != nil {
		r.Pages = m.Pages
	}
	if m.Authors != nil {
		r.Authors = m.Authors
	}
	if m.File != nil {
		f := fileUrl(m.File)
		r.File = &f
	}
	if m.FileSize != nil {
		r.FileSize = m.FileSize
	}
	if m.FilePreview != nil {
		f := fileUrl(m.FilePreview)
		r.FilePreview = &f
	}
	r.IsDownloadable = m.IsDownloadable
	r.CreatedAt = m.CreatedAt
	r.UpdatedAt = m.UpdatedAt
	return nil
}

func (r *BookRequest) ToModel(m *Book) error {
	m.ID = r.ID
	m.Title = r.Title
	m.Categories = r.Categories
	m.Description = r.Description
	m.Year = r.Year
	m.Pages = r.Pages
	m.Authors = r.Authors
	m.File = r.File
	m.FileSize = r.FileSize
	m.FilePreview = r.FilePreview
	m.IsDownloadable = r.IsDownloadable
	return nil
}

type BookFilterRequest struct {
	ID             string    `form:"id"`
	IDs            *[]string `form:"ids"`
	Categories     *[]string `form:"categories"`
	Year           *int      `form:"year"`
	Authors        *[]string `form:"authors"`
	IsDownloadable bool      `form:"is_downloadable"`
	Search         *string   `form:"search"`
	Sort           *string   `form:"sort"`
	PaginationRequest
}

var BookCategories = []string{
	"1-nji synp",
	"2-nji synp",
	"3-nji synp",
	"4-nji synp",
	"5-nji synp",
	"6-nji synp",
	"7-nji synp",
	"8-nji synp",
	"9-nji synp",
	"10-nji synp",
	"11-nji synp",
	"12-nji synp",
	"Edebiýat",
	"Taryh",
	"Iňlis dili",
	"Matematika",
}

var SubjectCategories = []string{
	"Iňlis dili",
	"Rus dili",
	"Nemes dili",
	"Hytaý dili",
	"Ýapon dili",
	"Fransuz dili",
	"Pars dili",
	"Arap dili",
	"Italýan dili",
	"Koreý dili",
	"Matematika",
	"Informatika",
	"Taryh",
	"Biologiýa",
	"Himiýa",
}

var FavouriteResource = [][]string{
	{"Turkmen-English phrasebook", "https://play.google.com/store/apps/details?id=com.translator.en.tm.team.tm_entranslator", "https://mekdep.edu.tm/uploads/emekdepMenuAdd/turkmenEnglish.webp"},
	{"Mekdep kitaphana", "https://play.google.com/store/apps/details?id=mekdep.kitaphana.net", "https://mekdep.edu.tm/uploads/emekdepMenuAdd/mekdepKitaphana.webp"},
	{"Kitaphana", "https://play.google.com/store/apps/details?id=app.kitaphana.net", "https://mekdep.edu.tm/uploads/emekdepMenuAdd/kitaphana.webp"},
	{"Türkmen Nakyllary", "https://play.google.com/store/apps/details?id=com.fibogame.turkmenhalknakyllary", "https://mekdep.edu.tm/uploads/emekdepMenuAdd/turkmenNakyllary.webp"},
	{"Poeziýa:  Türkmen Şygyrlar", "https://play.google.com/store/apps/details?id=com.izigroup.poeziya", "https://mekdep.edu.tm/uploads/emekdepMenuAdd/poeziya.webp"},
	{"Hazynam: sesli goşgylar", "https://play.google.com/store/apps/details?id=com.zehinz.hazynam", "https://mekdep.edu.tm/uploads/emekdepMenuAdd/hazynam.webp"},
	{"Magtymguly's World", "https://play.google.com/store/apps/details?id=tm.edu.datmddi.magtymgulyapp", "https://mekdep.edu.tm/uploads/emekdepMenuAdd/magtymgulysWorld.webp"},
	{"Öwrenmeli: peýdaly we gyzykly", "https://play.google.com/store/apps/details?id=com.ts.owrenmeli", "https://mekdep.edu.tm/uploads/emekdepMenuAdd/owrenmeliPeydaly.webp"},
	{"Duolingo: Language Lessons", "https://play.google.com/store/apps/details?id=com.duolingo", "https://mekdep.edu.tm/uploads/emekdepMenuAdd/duolingo.webp"},
	{"Memrise: speak a new language", "https://play.google.com/store/apps/details?id=com.memrise.android.memrisecompanion", "https://mekdep.edu.tm/uploads/emekdepMenuAdd/memrise.webp"},
	{"Sololearn: Learn to code", "https://play.google.com/store/apps/details?id=com.sololearn", "https://mekdep.edu.tm/uploads/emekdepMenuAdd/sololearn.webp"},
	{"Learn to Read - Duolingo ABC", "https://play.google.com/store/apps/details?id=com.duolingo.literacy", "https://mekdep.edu.tm/uploads/emekdepMenuAdd/duolingoABC.webp"},
	{"Khan Academy Kids", "https://play.google.com/store/apps/details?id=org.khankids.android", "https://mekdep.edu.tm/uploads/emekdepMenuAdd/khanAcademyKids.webp"},
	{"Khan Academy", "https://play.google.com/store/apps/details?id=org.khanacademy.android", "https://mekdep.edu.tm/uploads/emekdepMenuAdd/khanAcademy.webp"},
	{"Coursera: Learn career skills", "https://play.google.com/store/apps/details?id=org.coursera.android", "https://mekdep.edu.tm/uploads/emekdepMenuAdd/Coursera.webp"},
	{"edX: Courses by Harvard & MIT", "https://play.google.com/store/apps/details?id=org.edx.mobile", "https://mekdep.edu.tm/uploads/emekdepMenuAdd/edX.jpg"},
	{"TED", "https://play.google.com/store/apps/details?id=com.ted.android", "https://mekdep.edu.tm/uploads/emekdepMenuAdd/TED.webp"},
	{"Udemy: Online Courses", "https://play.google.com/store/apps/details?id=com.udemy.android", "https://mekdep.edu.tm/uploads/emekdepMenuAdd/udemy.jpg"},
	{"Wikipedia", "https://play.google.com/store/apps/details?id=org.wikipedia", "https://mekdep.edu.tm/uploads/emekdepMenuAdd/wikipedia.png"},
}
