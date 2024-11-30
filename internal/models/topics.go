package models

import "strings"

type Topics struct {
	ID          string    `json:"ids"`
	BookId      *string   `json:"book_id"`
	BookPage    *uint     `json:"book_page"`
	SubjectName *string   `json:"subject_name"`
	Classyear   *string   `json:"classyear"`
	Period      *string   `json:"period"`
	Level       *string   `json:"level"`
	Language    *string   `json:"language"`
	Tags        *[]string `json:"tags"`
	Title       *string   `json:"title"`
	Content     *string   `json:"content"`
	Files       *[]string `json:"files"`
	Book        *Book     `json:"book"`
}

func (Topics) RelationFields() []string {
	return []string{"Book"}
}

type TopicsValueResponse struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func (r *TopicsResponse) ToValues() TopicsValueResponse {
	value := ""
	if r.Title != nil {
		value += " " + *r.Title
	}
	return TopicsValueResponse{
		Key:   r.ID,
		Value: strings.Trim(value, " "),
	}
}

type TopicsRequest struct {
	ID          *string   `json:"id" form:"id"`
	BookId      *string   `json:"book_id" form:"book_id"`
	BookPage    *uint     `json:"book_page" form:"book_page"`
	SubjectName *string   `json:"subject" form:"subject"`
	Classyear   *string   `json:"classyear" form:"classyear"`
	Period      *string   `json:"period" form:"period"`
	Level       *string   `json:"level" form:"level"`
	Language    *string   `json:"language" form:"language"`
	Tags        *[]string `json:"tags" form:"tags"`
	Title       *string   `json:"title" form:"title"`
	Content     *string   `json:"content" form:"content"`
	Files       *[]string ``
	FilesDelete *[]string `json:"files_delete" form:"files_delete"`
}

type TopicsResponse struct {
	ID          string        `json:"id"`
	SubjectName *string       `json:"subject"`
	Classyear   *string       `json:"classyear"`
	Period      *string       `json:"period"`
	Level       *string       `json:"level"`
	Language    *string       `json:"language"`
	Tags        *[]string     `json:"tags"`
	Title       *string       `json:"title"`
	Content     *string       `json:"content"`
	BookPage    *uint         `json:"book_page"`
	Files       *[]string     `json:"files"`
	Book        *BookResponse `json:"book"`
}

type TopicsMultipleRequest struct {
	Topics []*TopicsRequest `json:"topics"`
}

func (r *TopicsResponse) FromModel(m *Topics) error {
	r.ID = m.ID
	if m.SubjectName != nil {
		r.SubjectName = m.SubjectName
	}
	if m.Classyear != nil {
		r.Classyear = m.Classyear
	}
	if m.Period != nil {
		r.Period = m.Period
	}
	r.Level = m.Level
	r.Language = m.Language
	r.Tags = m.Tags
	r.Title = m.Title
	r.Content = m.Content
	r.BookPage = m.BookPage
	if m.Files != nil {
		r.Files = &[]string{}
		for _, f := range *m.Files {
			*r.Files = append(*r.Files, fileUrl(&f))
		}
	}
	if m.Book != nil {
		r.Book = &BookResponse{}
		r.Book.FromModel(m.Book)
	}
	return nil
}

func (r *TopicsRequest) ToModel(m *Topics) error {
	if r.ID == nil {
		r.ID = new(string)
	}
	m.ID = *r.ID
	m.BookId = r.BookId
	m.BookPage = r.BookPage
	m.SubjectName = r.SubjectName
	m.Classyear = r.Classyear
	m.Period = r.Period
	m.Level = r.Level
	m.Language = r.Language
	m.Tags = r.Tags
	m.Title = r.Title
	m.Content = r.Content
	m.Files = r.Files
	return nil
}

type TopicsFilterRequest struct {
	ID           *string   `form:"id"`
	IDs          *[]string `form:"ids"`
	SubjectName  *string   `form:"subject"`
	Classyear    *string   `form:"classyear"`
	PeriodNumber *string   `form:"period_number"`
	Level        *string   `form:"level"`
	Language     *string   `form:"language"`
	Search       *string   `form:"search"`
	Sort         *string   `form:"sort"`
	PaginationRequest
}

var DefaultTopicSubjects = [][]string{
	{"Algebra", "Algebra", string(SubjectAreaExact)},
	{"Astronomiýa", "Astronomiýa", string(SubjectAreaNatural)},
	{"Aýdym-saz", "Aýdym-saz", string(SubjectAreaHumanity)},
	{"Bedenterbiýe", "Bedenterbiýe", string(SubjectAreaNatural)},
	{"Biologiýa", "Biologiýa", string(SubjectAreaNatural)},
	{"Çeper. zähmeti", "Çeperçilik zähmeti", string(SubjectAreaHumanity)},
	{"Döwr. Tehno. E.", "Döwrebap tehnologiýanyň esaslary", string(SubjectAreaExact)},
	{"Dünýä med.", "Dünýä medeniýeti"},
	{"Dünýä taryhy", "Dünýä taryhy"},
	{"Daşary ýurt dili", "Daşary ýurt dili"},
	{"Daşary ýurt dili (garyşyk)", "Daşary ýurt dili (garyşyk)"},
	{"Edebiýat", "Edebiýat"},
	{"Ekologiýa", "Ekologiýa", string(SubjectAreaNatural)},
	{"Ählumumy taryh", "Ählumumy taryh"},
	{"Fizika", "Fizika", string(SubjectAreaNatural)},
	{"Geografiýa", "Geografiýa", string(SubjectAreaNatural)},
	{"Geometriýa", "Geometriýa", string(SubjectAreaExact)},
	{"Himiýa", "Himiýa", string(SubjectAreaNatural)},
	{"Informatika", "Informatikanyň esaslary", string(SubjectAreaExact)},
	{"Informatika", "Informatika", string(SubjectAreaExact)},
	{"IKIT", "Informasiýa-kommunikasiýa we innowasion tehnologiýalar", string(SubjectAreaExact)},
	{"Iňlis dili", "Iňlis dili"},
	{"Jemgyýet", "Jemgyýeti öwreniş"},
	{"Matematika", "Matematika", string(SubjectAreaExact)},
	{"Model we G.", "Modelirleme we grafika", string(SubjectAreaExact)},
	{"Ene dili", "Ene dili"},
	{"Okuw", "Okuw"},
	{"Özüňi alyp B.M.", "Özüňi alyp barmagyň medeniýeti"},
	{"Proýek. Esas.", "Proýektirlemegiň esaslary", string(SubjectAreaExact)},
	{"Rus dili", "Rus dili"},
	{"Rus edebiýaty", "Rus edebiýaty"},
	{"Türkmen edebiýaty", "Türkmen edebiýaty"},
	{"Şekil. Sun.", "Şekillendiriş sungaty"},
	{"Surat", "Surat"},
	{"Tebigat", "Tebigaty öwreniş", string(SubjectAreaNatural)},
	{"Türkmen dili", "Türkmen dili"},
	{"Hukuk esaslary", "Türkmenistanyň Döwlet we hukuk esaslary"},
	{"Medeni miras", "Türkmenistanyň medeni mirasy"},
	{"T-nyň taryhy", "Türkmenistanyň taryhy"},
	{"Ykdysadyýet", "Ykdysadyýetiň esaslary"},
	{"Ýaşaýyş D.E.", "Ýaşaýyş durmuş esaslary"},
	{"Ýazuw", "Ýazuw", string(SubjectAreaExact)},
	{"Zähmet", "Zähmet"},
	{"Durmuş zähmeti", "Durmuş zähmeti"},
	{"Çyzuw", "Çyzuw"},
	{"Synp sagady", "Synp sagady"},
	{"Nemes dili", "Nemes dili", string(SubjectAreaHumanity), "#78ab68"},
	{"Fransuz dili", "Fransuz dili", string(SubjectAreaHumanity), "#78ab68"},
	{"Ýapon dili", "Ýapon dili", string(SubjectAreaHumanity), "#78ab68"},
	{"Hytaý dili", "Hytaý dili", string(SubjectAreaHumanity), "#78ab68"},
	{"Pars dili", "Pars dili", string(SubjectAreaHumanity), "#78ab68"},
	{"Arap dili", "Arap dili", string(SubjectAreaHumanity), "#78ab68"},
	{"Koreý dili", "Koreý dili", string(SubjectAreaHumanity), "#78ab68"},
	{"Italýan dili", "Italýan dili", string(SubjectAreaHumanity), "#78ab68"},
}

var TopicTags = []string{"Takyk", "Ynsanperwer", "Tebigy", "Hünär"}
