package models

import (
	"errors"
	"log"
	"math"
	"strconv"
	"time"
)

type PeriodGrade struct {
	ID             string     `json:"id"`
	PeriodId       *string    `json:"period_id"`
	PeriodKey      int        `json:"period_key"`
	SubjectId      *string    `json:"subject_id"`
	StudentId      *string    `json:"student_id"`
	ExamId         *string    `json:"exam_id"`
	LessonCount    int        `json:"lesson_count"`
	AbsentCount    int        `json:"absent_count"`
	GradeCount     int        `json:"grade_count"`
	GradeSum       int        `json:"grade_sum"`
	OldAbsentCount int        `json:"old_absent_count"`
	OldGradeCount  int        `json:"old_grade_count"`
	OldGradeSum    int        `json:"old_grade_sum"`
	PrevGradeCount int        `json:"prev_grade_count"`
	PrevGradeSum   int        `json:"prev_grade_sum"`
	UpdatedAt      *time.Time `json:"updated_at"`
	CreatedAt      *time.Time `json:"created_at"`
	Student        *User      `json:"student"`
}

func (PeriodGrade) RelationFields() []string {
	return []string{"Student"}
}

func (m *PeriodGrade) IsUpdateExpired() bool {
	// TODO: .Add(time.Hour * 5) is timezone problem, pgx uses UTC+0, ours UTC+5
	now := time.Now()
	if m.CreatedAt.Location().String() != now.Location().String() {
		log.Println("DEBUG: " + errors.New("database timezone is incorrect: "+m.CreatedAt.Location().String()+"; machine timezone: "+now.Location().String()).Error())
		*m.CreatedAt = m.CreatedAt.Local()
	}
	return m.CreatedAt.IsZero() || m.CreatedAt.Add(time.Minute*60).Before(now)
}

func (m *PeriodGrade) IsCreateExpired(exams []*SubjectExam) bool {
	if len(exams) == 0 {
		return false
	}
	now := time.Now()
	min := now.AddDate(0, 0, -15) // TODO: tmp need to change monday (17.03.2024)
	for _, exam := range exams {
		if exam.StartTime == nil {
			return true
		}
		if exam.StartTime.Before(min) || exam.StartTime.After(now) {
			return true
		}
	}
	return false
}

func (m *PeriodGrade) GradeValue() float64 {
	pr := 1
	// how exact (zero after ,)
	dec := math.Pow10(pr)
	var a float64
	count := m.GetGradeCount()
	sum := m.GetGradeSum()
	if count > 0 && sum > 0 {
		a = math.Round(float64(sum)/float64(count)*float64(dec)) / float64(dec)
	}
	return a
}

func (m PeriodGrade) GetAbsentCount() int {
	return m.AbsentCount + m.OldAbsentCount
}

func (m PeriodGrade) GetGradeCount() int {
	return m.GradeCount + m.OldGradeCount
}

func (m PeriodGrade) GetGradeSum() int {
	return m.GradeSum + m.OldGradeSum
}

func (m *PeriodGrade) AppendGrade(v *PeriodGrade) {
	if v.GradeIntValue() == 0 {
		return
	}
	m.LessonCount += v.LessonCount
	m.GradeCount += 1
	m.GradeSum += v.GradeIntValue()
	m.AbsentCount += v.GetAbsentCount()
}

func (m *PeriodGrade) AppendPowerGrade(v *PeriodGrade) {
	m.GradeCount += 1
	m.GradeSum += v.GradeIntValue()
}

func (m *PeriodGrade) GradeValuePrev() float64 {
	pr := 1
	dec := math.Pow10(pr)
	var a float64
	if m.PrevGradeCount > 0 && m.PrevGradeSum > 0 {
		a = math.Round(float64(m.PrevGradeSum)/float64(m.PrevGradeCount)*float64(dec)) / float64(dec)
	}
	return a
}

func (m *PeriodGrade) GradeIntValue() int {
	return int(math.Round(m.GradeValue()))
}

func (m *PeriodGrade) GradeIntValuePrev() int {
	return int(math.Round(m.GradeValuePrev()))
}

type PeriodGradeFilterRequest struct {
	PeriodId   *string   `json:"period_id"`
	PeriodKey  *int      `json:"period_key"`
	PeriodKeys *[]int    `json:"period_keys"`
	SubjectId  *string   `json:"subject_id"`
	ExamId     *string   `json:"exam_id"`
	ExamIds    []*string `json:"exam_ids"`
	StudentId  *string   `json:"student_id"`
	SubjectIds *[]string `json:"subject_ids"`
	StudentIds *[]string `json:"student_ids"`
}

const PeriodGradeExamKey = -1

type PeriodGradeResponse struct {
	PeriodId       *string       `json:"period_id"`
	PeriodKey      int           `json:"period_key"`
	StudentId      *string       `json:"student_id"`
	ExamId         *string       `json:"exam_id"`
	LessonCount    int           `json:"lesson_count"`
	AbsentCount    int           `json:"absent_count"`
	GradeCount     int           `json:"grade_count"`
	GradeValue     string        `json:"grade_value"`
	GradeValuePrev string        `json:"grade_value_prev"`
	OldGradeCount  int           `json:"old_grade_count"`
	OldGradeSum    int           `json:"old_grade_sum"`
	OldAbsentCount int           `json:"old_absent_count"`
	UpdatedAt      *time.Time    `json:"updated_at"`
	Student        *UserResponse `json:"student"`
}

func (r *PeriodGradeResponse) FromModel(m *PeriodGrade) {
	r.PeriodId = m.PeriodId
	r.PeriodKey = m.PeriodKey
	r.StudentId = m.StudentId
	r.ExamId = m.ExamId
	r.LessonCount = m.LessonCount
	r.AbsentCount = m.GetAbsentCount()
	r.GradeCount = m.GetGradeCount()
	r.UpdatedAt = m.UpdatedAt
	if m.Student != nil {
		r.Student = &UserResponse{}
		r.Student.FromModel(m.Student)
	}
	if m.GradeValuePrev() > 0 {
		r.GradeValuePrev = strconv.FormatFloat(float64(m.GradeIntValuePrev()), 'f', -1, 64)
	}
	if m.GradeValue() > 0 {
		r.GradeValue = strconv.FormatFloat(float64(m.GradeIntValue()), 'f', -1, 64)
	}
}

func (PeriodGrade) MinGradeCount() int {
	return 3
}

func (r *PeriodGrade) IsCompleted() bool {
	return r.GradeCount >= r.MinGradeCount()
}

func (r *PeriodGrade) IsNoGrade() bool {
	return r.LessonCount-r.GetAbsentCount() < PeriodGrade{}.MinGradeCount()
}

func (r *PeriodGradeResponse) IsCompleted() bool {
	return r.GradeCount >= PeriodGrade{}.MinGradeCount()
}

func (r *PeriodGradeResponse) IsNoGrade() bool {
	return r.LessonCount >= PeriodGrade{}.MinGradeCount() && r.LessonCount-r.AbsentCount < PeriodGrade{}.MinGradeCount()
}

func (r *PeriodGradeResponse) SetValueByRules() {
	if !r.IsCompleted() {
		r.GradeValue = ""
		r.GradeValuePrev = ""
	}
	if r.IsNoGrade() {
		r.GradeValue = "BA"
		r.GradeValuePrev = "BA"
	}
}
