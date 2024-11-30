package models

import (
	"time"
)

type DiaryResponse struct {
	LastReviewedAt *time.Time         `json:"reviewed_at"`
	Days           []DiaryDayResponse `json:"days"`
}

type StudentSubjectResponse struct {
	PeriodCounts []StudentPeriodCount         `json:"period_counts"`
	Items        []StudentSubjectResponseItem `json:"subjects"`
}

type StudentSubjectResponseItem struct {
	Subject      *SubjectResponse            `json:"subject"`
	PeriodGrades map[int]PeriodGradeResponse `json:"period_grades"`
	ExamGrade    *PeriodGradeResponse        `json:"exam_grade"`
	FinalGrade   *PeriodGradeResponse        `json:"final_grade"`
}

type StudentPeriodCount struct {
	PeriodKey       int `json:"period_key"`
	LessonsCount    int `json:"lessons_count"`
	AbsentsYesCount int `json:"absents_yes_count"`
	AbsentsIllCount int `json:"absents_ill_count"`
	AbsentsNoCount  int `json:"absents_no_count"`
}

type DiaryDayResponse struct {
	Holiday *string               `json:"holiday"`
	Date    string                `json:"date"`
	Hours   []DiaryLessonResponse `json:"hours"`
}

type DiarySubjectResponse struct {
	ID       string        `json:"id"`
	Name     string        `json:"name"`
	FullName string        `json:"full_name"`
	Teacher  *UserResponse `json:"teacher"`
}

type DiaryLessonResponse struct {
	ShiftTimes []string            `json:"shift_times"`
	Subject    *SubjectResponse    `json:"subject"`
	Lesson     *LessonResponse     `json:"lesson"`
	Absent     *AbsentResponse     `json:"absent"`
	Grade      *GradeResponse      `json:"grade"`
	Assignment *AssignmentResponse `json:"assignment"`
}

type ParentChildrenResponse struct {
	User      UserResponse      `json:"user"`
	Classroom ClassroomResponse `json:"classroom"`
	School    SchoolResponse    `json:"school"`
	Payment   struct {
		AccountStatus string    `json:"account_status"`
		LastPaid      time.Time `json:"last_paid"`
	} `json:"payment"`
}
