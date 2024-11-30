package models

type JournalResponse struct {
	Lessons      []JournalItemResponse `json:"lessons"`
	StudentNotes []StudentNoteResponse `json:"student_notes"`
	PeriodGrades []PeriodGradeResponse `json:"period_grades"`
	Students     []UserResponse        `json:"students"`
}

type JournalItemResponse struct {
	Lesson  LessonResponse   `json:"lesson"`
	Grades  []GradeResponse  `json:"grades"`
	Absents []AbsentResponse `json:"absents"`
}

type TeacherTimetableResponse struct {
	Shift     ShiftResponse         `json:"shift"`
	Timetable TeacherTimetableValue `json:"timetable"`
}

type TeacherTimetableValue [][]*TeacherTimetableItemResponse

type TeacherTimetableItemResponse struct {
	Date           string                    `json:"date"`
	ShiftTimes     []string                  `json:"shift_times"`
	Subject        *SubjectResponse          `json:"subject"`
	SubjectPercent *DashboardSubjectsPercent `json:"subject_percent"`
}

type JournalRequest struct {
	Lesson      LessonRequest       `json:"lesson"`
	Grade       *GradeRequest       `json:"grade"`
	Absent      *AbsentRequest      `json:"absent"`
	StudentNote *StudentNoteRequest `json:"student_note"`
}

type JournalRequestV2 struct {
	LessonId    string              `json:"lesson_id"`
	Lesson      LessonRequest       `json:"lesson"`
	Assignment  AssignmentRequest   `json:"assignment"`
	LessonPro   LessonProRequest    `json:"lesson_pro"`
	Grade       *GradeRequest       `json:"grade"`
	Absent      *AbsentRequest      `json:"absent"`
	StudentNote *StudentNoteRequest `json:"student_note"`
}
type JournalFormRequest struct {
	LessonId              string    `json:"lesson_id" form:"lesson_id"`
	AssignmentFilesDelete *[]string `json:"assignment_files_delete" validate:"omitempty"`
	AssignmentFiles       *[]string ``
	LessonProFilesDelete  *[]string `json:"lesson_pro_files_delete" validate:"omitempty"`
	LessonProFiles        *[]string ``
}

func (d *JournalRequestV2) SetKeys() {
	d.Lesson.ID = &d.LessonId
	d.Lesson.Assignment = d.Assignment
	d.Lesson.LessonPro = d.LessonPro
}

type ReportRequest struct {
}
