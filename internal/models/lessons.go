package models

import (
	"time"
)

// TODO: add period Key
type Lesson struct {
	ID                string                  `json:"id"`
	SchoolId          string                  `json:"school_id"`
	SubjectId         string                  `json:"subject_id"`
	BookId            *string                 `json:"book_id"`
	BookPage          *uint                   `json:"book_page"`
	PeriodId          *string                 `json:"period_id"`
	PeriodKey         *int                    `json:"period_key"`
	Date              time.Time               `json:"date"`
	HourNumber        *int                    `json:"hour_number"`
	TypeTitle         *string                 `json:"type_title"`
	Title             *string                 `json:"title"`
	Content           *string                 `json:"content"`
	ProTitle          *string                 `json:"pro_title"`
	ProFiles          *[]string               `json:"pro_files"`
	AssignmentTitle   *string                 `json:"assignment_title"`
	AssignmentContent *string                 `json:"assignment_content"`
	AssignmentFiles   *[]string               `json:"assignment_files"`
	LessonAttributes  *map[string]interface{} `json:"lesson_attributes"`
	IsTeacherExcused  *bool                   `json:"is_teacher_excused"`
	UpdatedAt         *time.Time              `json:"updated_at"`
	CreatedAt         *time.Time              `json:"created_at"`
	School            *School                 `json:"school"`
	Subject           *Subject                `json:"subject"`
	Book              *Book                   `json:"book"`
	Period            *Period                 `json:"period"`
	Assignment        *Assignment             `json:"assignment"`
}

func (Lesson) RelationFields() []string {
	return []string{"School", "Subject", "Book", "Period", "Assignment"}
}

func (lesson Lesson) Label() string {
	label := ""
	if lesson.Subject != nil {
		if lesson.Subject.Classroom != nil {
			label = *lesson.Subject.Name + " " + *lesson.Subject.Classroom.Name + " " + lesson.Date.Format(time.DateOnly)
		} else {
			label = *lesson.Subject.Name + " ? " + lesson.Date.Format(time.DateOnly)
		}
	}
	return label
}

type LessonResponse struct {
	ID               string                  `json:"id"`
	SchoolId         string                  `json:"school_id"`
	SubjectId        string                  `json:"subject_id"`
	PeriodNumber     int                     `json:"period_number"`
	Date             string                  `json:"date"`
	HourNumber       int                     `json:"hour_number"`
	TypeTitle        string                  `json:"type_title"`
	Title            string                  `json:"title"`
	Content          *string                 `json:"content"`
	BookPage         *uint                   `json:"book_page"`
	LessonAttributes *map[string]interface{} `json:"lesson_attributes"`
	UpdatedAt        *time.Time              `json:"updated_at"`
	CreatedAt        *time.Time              `json:"created_at"`
	School           *SchoolResponse         `json:"school"`
	Subject          *SubjectResponse        `json:"subject"`
	Book             *BookResponse           `json:"book"`
	Assignment       *AssignmentResponse     `json:"assignment"`
	LessonPro        *LessonProResponse      `json:"lesson_pro"`
}

type LessonLike struct {
	ID       *string `json:"id"`
	UserId   *string `json:"user_id"`
	LessonId *string `json:"lesson_id"`
	User     *User   `json:"user"`
	Lesson   *Lesson `json:"lesson"`
}

type LessonLikeRequest struct {
	ID       *string `json:"id"`
	UserId   *string `json:"user_id"`
	LessonId *string `json:"lesson_id"`
}

type LessonProResponse struct {
	Title *string   `json:"title"`
	Files *[]string `json:"files"`
}

func (r *LessonResponse) FromModel(m *Lesson) {
	if m.PeriodId == nil {
		m.PeriodId = new(string)
	}
	if m.TypeTitle == nil {
		m.TypeTitle = new(string)
	}
	if m.Title == nil {
		m.Title = new(string)
	}
	r.ID = m.ID
	r.SchoolId = m.SchoolId
	r.SubjectId = m.SubjectId
	if m.PeriodKey != nil {
		r.PeriodNumber = *m.PeriodKey
	}
	if !m.Date.IsZero() {
		r.Date = m.Date.Format(time.DateOnly)
	}
	if m.HourNumber == nil {
		m.HourNumber = new(int)
	}
	r.HourNumber = *m.HourNumber
	r.TypeTitle = *m.TypeTitle
	r.Title = *m.Title
	r.Content = m.Content
	r.BookPage = m.BookPage
	if m.LessonAttributes != nil {
		r.LessonAttributes = m.LessonAttributes
	}
	r.UpdatedAt = m.UpdatedAt
	r.CreatedAt = m.CreatedAt
	if m.School != nil {
		t := SchoolResponse{}
		t.FromModel(m.School)
		r.School = &t
	}
	if m.Subject != nil {
		t := SubjectResponse{}
		t.FromModel(m.Subject)
		r.Subject = &t
	}
	if m.Book != nil {
		t := BookResponse{}
		t.FromModel(m.Book)
		r.Book = &t
	}
	r.BookPage = m.BookPage
	if m.AssignmentTitle != nil || m.AssignmentContent != nil || m.AssignmentFiles != nil && len(*m.AssignmentFiles) > 0 {
		r.Assignment = &AssignmentResponse{}
		if m.AssignmentTitle != nil {
			r.Assignment.Title = *m.AssignmentTitle
		}
		if m.AssignmentFiles != nil {
			r.Assignment.Files = &[]string{}
			for _, f := range *m.AssignmentFiles {
				*r.Assignment.Files = append(*r.Assignment.Files, fileUrl(&f))
			}
		}
		if m.AssignmentContent != nil {
			r.Assignment.Content = m.AssignmentContent
		}
	}
	if m.ProTitle != nil || m.ProFiles != nil && len(*m.ProFiles) > 0 {
		r.LessonPro = &LessonProResponse{}
		if m.ProTitle != nil {
			r.LessonPro.Title = m.ProTitle
		}
		if m.ProFiles != nil {
			r.LessonPro.Files = &[]string{}
			for _, f := range *m.ProFiles {
				*r.LessonPro.Files = append(*r.LessonPro.Files, fileUrl(&f))
			}
		}
	}
}

type LessonFilterRequest struct {
	ID               *string    `form:"id"`
	SchoolId         *string    `form:"school_id"`
	SchoolIds        *[]string  `form:"school_ids[]"`
	SubjectId        *string    `form:"subject_id"`
	SubjectIds       *[]string  `form:"subject_ids[]"`
	ClassroomId      *string    `form:"classroom_id"`
	PeriodId         *string    `form:"period_id"`
	PeriodNumber     *int       `form:"period_number"`
	TypeTitle        *string    `form:"type_title"`
	Date             *time.Time `form:"date"`
	DateRange        *[]string  `form:"date_range"`
	HourNumber       *int       `form:"hour_number"`
	HourNumberRange  *[]int     `form:"hour_number_range"`
	IsTeacherExcused *bool      `form:"is_teacher_excused"`
	Search           *string    `form:"search"`
	Sort             *string    `form:"sort"`
	PaginationRequest
}

type LessonProRequest struct {
	Title       *string   `json:"title" form:"Title"`
	Files       *[]string `json:"files" validate:"omitempty"`
	FilesDelete *[]string `json:"files_delete" validate:"omitempty"`
}

type LessonRequest struct {
	ID               *string                 `json:"id"`
	TypeTitle        *string                 `json:"type_title"`
	Title            *string                 `json:"title"`
	Content          *string                 `json:"content"`
	BookPage         *uint                   `json:"book_page"`
	Date             *string                 `json:"date" `
	HourNumber       int                     `json:"hour_number"`
	SubjectID        string                  `json:"subject_id"`
	BookId           *string                 `json:"book_id"`
	LessonAttributes *map[string]interface{} `json:"lesson_attributes"`
	Assignment       AssignmentRequest       `json:"assignment"`
	LessonPro        LessonProRequest        `json:"lesson_pro"`
}

func (m *Lesson) FromRequest(r *LessonRequest) error {
	if r.ID != nil {
		m.ID = *r.ID
	}
	if r.Date != nil {
		m.Date, _ = time.Parse(time.DateOnly, *r.Date)
	}
	if r.HourNumber != 0 {
		m.HourNumber = &r.HourNumber
	}
	m.TypeTitle = r.TypeTitle
	m.Title = r.Title
	m.Content = r.Content
	m.BookPage = r.BookPage
	m.SubjectId = r.SubjectID
	m.BookId = r.BookId
	if r.LessonAttributes != nil {
		m.LessonAttributes = r.LessonAttributes
	}
	m.Assignment = &Assignment{}
	if r.Assignment.Title != nil {
		m.AssignmentTitle = r.Assignment.Title
	}
	if r.Assignment.Files != nil {
		m.AssignmentFiles = r.Assignment.Files
	}
	if r.Assignment.Content != nil {
		m.AssignmentContent = r.Assignment.Content
	}
	if r.LessonPro.Title != nil {
		m.ProTitle = r.LessonPro.Title
	}
	if r.LessonPro.Files != nil {
		m.ProFiles = r.LessonPro.Files
	}
	return nil
}

type LessonFinalResponse struct {
	Student UserResponse                    `json:"student"`
	Periods map[string]*PeriodGradeResponse `json:"periods"`
	Exam    *PeriodGradeResponse            `json:"exam"`
	Final   *PeriodGradeResponse            `json:"final"`
}

type ExamWithGrade struct {
	Exam  *SubjectExamResponse `json:"exam"`
	Grade *PeriodGradeResponse `json:"grade"`
}

type LessonFinalBySubjectResponse struct {
	Students []LessonFinalBySubjectStudentResponse `json:"students"`
	Subjects []*SubjectResponse                    `json:"subjects"`
	Exams    []*SubjectExamResponse                `json:"exams"`
}
type LessonFinalBySubjectStudentResponse struct {
	Student  *UserResponse                         `json:"student"`
	Subjects []LessonFinalBySubjectSubjectResponse `json:"subjects"`
}

type LessonFinalBySubjectSubjectResponse struct {
	PeriodGrade *PeriodGradeResponse   `json:"period_grade"`
	ExamGrades  *[]PeriodGradeResponse `json:"exam_grades"`
	SubjectId   *string                `json:"subject_id"`
}

type LessonFinalResponseV2 struct {
	Student UserResponse                    `json:"student"`
	Periods map[string]*PeriodGradeResponse `json:"periods"`
	Exams   []*ExamWithGrade                `json:"exams"`
	Final   *PeriodGradeResponse            `json:"final"`
}

var DefaultLessonTypes = [][]string{
	{"new_topic", "Täze maglumatlary öwretmek sapagy",
		"Урок изучения нового материала",
		"Täze maglumatlary öwretmek sapagy"},
	{"skills", "Bilimleri, başarnyklary we endikleri berkitmek sapagy.",
		"Урок для закрепления знаний, умений и навыков",
		"Bilimleri, başarnyklary we endikleri berkitmek sapagy."},
	{"test", "Bilimleri barlamak we bahalandyrmak sapagy.",
		"Урок проверки и оценки знаний",
		"Bilimleri barlamak we bahalandyrmak sapagy."},
	{"repeat", "Gaýtalamak sapagy",
		"Урок повторения",
		"Gaýtalamak sapagy"},
	{"lecture", "Leksiýa sapagy",
		"Урок-лекция",
		"Leksiýa sapagy"},
	{"practice", "Amaly-tejribe sapagy",
		"Практический урок",
		"Amaly-tejribe sapagy"},
	{"mix", "Garyşyk sapagy",
		"Комбинированный урок",
		"Garyşyk sapagy"},
	{"exam", "Ýazuw-barlag işi sapagy",
		"Контрольная работа",
		"Ýazuw-barlag işi sapagy"},
	{"essay", "Düzme-Beýannama sapagy",
		"Сочинение-эссе",
		"Düzme-Beýannama sapagy"},
}
