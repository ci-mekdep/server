package models

import (
	"encoding/json"
	"time"
)

type LogSubject string
type LogAction string

const LogSubjectUsers LogSubject = "users"
const LogSubjectSchools LogSubject = "schools"
const LogSubjectClassrooms LogSubject = "classrooms"
const LogSubjectLesson LogSubject = "lessons"
const LogSubjectGrade LogSubject = "grades"
const LogSubjectAbsent LogSubject = "absents"
const LogSubjectTimetable LogSubject = "timetables"
const LogSubjectSubject LogSubject = "subjects"
const LogSubjectShift LogSubject = "shifts"
const LogSubjectTopics LogSubject = "topics"
const LogSubjectPeriod LogSubject = "periods"
const LogSubjectPayment LogSubject = "payments"
const LogSubjectBook LogSubject = "books"
const LogSubjectBaseSubjects LogSubject = "base_subjects"
const LogSubjectReports LogSubject = "reports"
const LogSubjectReportItems LogSubject = "report_items"
const LogSubjectSchoolTransfers LogSubject = "school_transfers"

const LogActionCreate LogAction = "create"
const LogActionUpdate LogAction = "update"
const LogActionDelete LogAction = "delete"
const LogActionLogout LogAction = "logout"
const LogActionLogin LogAction = "login"
const LogActionUpdatePassword LogAction = "update_password"
const LogActionUpdateProfile LogAction = "update_profile"

type UserLog struct {
	ID                 string      `json:"id"`
	SchoolId           *string     `json:"school_id"`
	SessionId          *string     `json:"session_id"`
	UserId             string      `json:"user_id"`
	SubjectId          *string     `json:"subject_id"`
	Subject            LogSubject  `json:"subject"`
	SubjectAction      LogAction   `json:"subject_action"`
	SubjectDescription *string     `json:"subject_description"`
	SubjectProperties  interface{} `json:"subject_properties"`
	CreatedAt          time.Time   `json:"created_at"`
	School             *School
	User               *User
	SubjectModel       interface{}
	Session            *Session
}

func (UserLog) RelationFields() []string {
	return []string{"School", "User", "SubjectModel", "Session"}
}

type UserLogResponse struct {
	ID                 string                 `json:"id"`
	SchoolId           *string                `json:"school_id"`
	UserId             string                 `json:"user_id"`
	SessionId          *string                `json:"session_id"`
	SubjectId          *string                `json:"subject_id"`
	Subject            LogSubject             `json:"subject_name"`
	SubjectAction      LogAction              `json:"subject_action"`
	SubjectDescription *string                `json:"subject_description"`
	SubjectProperties  map[string]interface{} `json:"properties"`
	CreatedAt          time.Time              `json:"created_at"`
	School             *SchoolResponse        `json:"school"`
	User               *UserResponse          `json:"user"`
	Session            *SessionResponse       `json:"session"`
}

func (r *UserLogResponse) FromModel(m *UserLog) error {
	r.ID = m.ID
	r.SchoolId = m.SchoolId
	r.UserId = m.UserId
	r.SessionId = m.SessionId
	r.SubjectId = m.SubjectId
	r.Subject = m.Subject
	r.SubjectDescription = m.SubjectDescription
	r.SubjectAction = m.SubjectAction
	if m.SubjectProperties != nil {
		str, _ := json.Marshal(m.SubjectProperties)
		_ = json.Unmarshal(str, &r.SubjectProperties)
	}
	r.CreatedAt = m.CreatedAt

	if m.School != nil {
		r.School = &SchoolResponse{}
		r.School.FromModel(m.School)
	}
	if m.User != nil {
		r.User = &UserResponse{}
		r.User.FromModel(m.User)
	}
	if m.Session != nil {
		r.Session = &SessionResponse{}
		r.Session.FromModel(m.Session)
	}
	return nil
}

type UserLogFilterRequest struct {
	ID            *string
	SchoolId      *string    `form:"school_id"`
	SchoolIds     *[]string  `form:"school_id[]"`
	UserId        *string    `form:"user_id"`
	SessionId     *string    `form:"session_id"`
	SubjectId     *string    `form:"subject_id"`
	SubjectName   *string    `form:"subject_name"`
	SubjectAction *string    `form:"subject_action"`
	Ip            *string    `form:"ip"`
	StartDate     *time.Time `form:"start_date"`
	EndDate       *time.Time `form:"end_date"`
	Search        *string    `form:"search"`
	SearchType    *string    `form:"search_type"`
	PaginationRequest
}
