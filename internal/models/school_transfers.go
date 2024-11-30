package models

import "time"

type StatusSchoolTransfer string

const (
	StatusAccepted StatusSchoolTransfer = "accepted"
	StatusDeclined StatusSchoolTransfer = "rejected"
	StatusWaiting  StatusSchoolTransfer = "waiting"
)

type SchoolTransfer struct {
	ID                string     `json:"id"`
	StudentId         *string    `json:"student_id"`
	TargetSchoolId    *string    `json:"target_school_id"`
	SourceSchoolId    *string    `json:"source_school_id"`
	TargetClassroomId *string    `json:"target_classroom_id"`
	SourceClassroomId *string    `json:"source_classroom_id"`
	SenderNote        *string    `json:"sender_note"`
	SenderFiles       []string   `json:"sender_files"`
	ReceiverNote      *string    `json:"receiver_note"`
	SentBy            *string    `json:"sent_by"`
	ReceivedBy        *string    `json:"received_by"`
	Status            *string    `json:"status"`
	UpdatedAt         *time.Time `json:"updated_at"`
	CreatedAt         *time.Time `json:"created_at"`
	Student           *User      `json:"student"`
	TargetSchool      *School    `json:"target_school"`
	SourceSchool      *School    `json:"source_school"`
	TargetClassroom   *Classroom `json:"target_classroom"`
	SourceClassroom   *Classroom `json:"source_classroom"`
	SentByUser        *User      `json:"sent_by_user"`
	ReceivedByUser    *User      `json:"received_by_user"`
}

type SchoolTransfers struct {
	SchoolTransfers []*SchoolTransfer
	Total           int
}

func (SchoolTransfer) RelationFields() []string {
	return []string{"Student", "TargetSchool", "SourceSchool", "TargetClassroom", "SourceClassroom", "SentByUser", "ReceivedByUser"}
}

type SchoolTransferCreateDto struct {
	ID                *string    `json:"id" form:"id"`
	StudentId         *string    `json:"student_id" form:"student_id"`
	TargetSchoolId    *string    `json:"target_school_id" form:"target_school_id"`
	SourceSchoolId    *string    `json:"source_school_id" form:"source_school_id"`
	TargetClassroomId *string    `json:"target_classroom_id" form:"target_classroom_id"`
	SourceClassroomId *string    `json:"source_classroom_id" form:"source_classroom_id"`
	SenderNote        *string    `json:"sender_note" form:"sender_note"`
	SenderFiles       []string   ``
	ReceiverNote      *string    `json:"receiver_note" form:"receiver_note"`
	SentBy            *string    `json:"sent_by" form:"sent_by"`
	ReceivedBy        *string    `json:"received_by" form:"received_by"`
	Status            *string    `json:"status" form:"status"`
	UpdatedAt         *time.Time `json:"updated_at" form:"updated_at"`
	CreatedAt         *time.Time `json:"created_at" form:"created_at"`
}

type SchoolTransferResponse struct {
	ID                string             `json:"id"`
	StudentId         *string            `json:"student_id"`
	TargetSchoolId    *string            `json:"target_school_id"`
	SourceSchoolId    *string            `json:"source_school_id"`
	TargetClassroomId *string            `json:"target_classroom_id"`
	SourceClassroomId *string            `json:"source_classroom_id"`
	SenderNote        *string            `json:"sender_note"`
	SenderFiles       []string           `json:"sender_file"`
	ReceiverNote      *string            `json:"receiver_note"`
	Status            *string            `json:"status"`
	UpdatedAt         *time.Time         `json:"updated_at"`
	CreatedAt         *time.Time         `json:"created_at"`
	Student           *UserResponse      `json:"student"`
	TargetSchool      *SchoolResponse    `json:"target_school"`
	SourceSchool      *SchoolResponse    `json:"source_school"`
	TargetClassroom   *ClassroomResponse `json:"target_classroom"`
	SourceClassroom   *ClassroomResponse `json:"source_classroom"`
	SentByUser        *UserResponse      `json:"sent_by"`
	RecievedByUser    *UserResponse      `json:"received_by"`
}

type SchoolTransfersResponse struct {
	SchoolTransfersResponse []SchoolTransferResponse `json:"school_transfer_response"`
	Total                   int                      `json:"total"`
}

type SchoolTransferQueryDto struct {
	ID             *string  `form:"id"`
	IDs            []string `form:"ids"`
	StudentId      *string  `form:"student_id"`
	TargetSchoolId *string  `form:"target_school_id"`
	SourceSchoolId *string  `form:"source_school_id"`
	Status         *string  `form:"status"`
	Limit          int      `form:"limit"`
	Offset         int      `form:"offset"`
}

func (response *SchoolTransferResponse) FromModel(model *SchoolTransfer) {
	response.ID = model.ID
	response.StudentId = model.StudentId
	response.TargetSchoolId = model.TargetSchoolId
	response.SourceSchoolId = model.SourceSchoolId
	response.TargetClassroomId = model.TargetClassroomId
	response.SourceClassroomId = model.SourceClassroomId
	response.SenderNote = model.SenderNote
	if model.SenderFiles != nil {
		response.SenderFiles = []string{}
		for _, v := range model.SenderFiles {
			response.SenderFiles = append(response.SenderFiles, fileUrl(&v))
		}
	}
	response.ReceiverNote = model.ReceiverNote
	response.Status = model.Status
	response.CreatedAt = model.CreatedAt
	response.UpdatedAt = model.UpdatedAt
	if model.Student != nil {
		t := UserResponse{}
		t.FromModel(model.Student)
		response.Student = &t
	}
	if model.TargetSchool != nil {
		t := SchoolResponse{}
		t.FromModel(model.TargetSchool)
		response.TargetSchool = &t
	}
	if model.SourceSchool != nil {
		t := SchoolResponse{}
		t.FromModel(model.SourceSchool)
		response.SourceSchool = &t
	}
	if model.TargetClassroom != nil {
		t := ClassroomResponse{}
		t.FromModel(model.TargetClassroom)
		response.TargetClassroom = &t
	}
	if model.SourceClassroom != nil {
		t := ClassroomResponse{}
		t.FromModel(model.SourceClassroom)
		response.SourceClassroom = &t
	}
	if model.SentByUser != nil {
		t := UserResponse{}
		t.FromModel(model.SentByUser)
		response.SentByUser = &t
	}
	if model.ReceivedByUser != nil {
		t := UserResponse{}
		t.FromModel(model.ReceivedByUser)
		response.RecievedByUser = &t
	}
}

func (req *SchoolTransferCreateDto) ToModel(model *SchoolTransfer) {
	if req.ID == nil {
		req.ID = new(string)
	}
	model.ID = *req.ID
	model.StudentId = req.StudentId
	model.TargetSchoolId = req.TargetSchoolId
	model.SourceSchoolId = req.SourceSchoolId
	model.TargetClassroomId = req.TargetClassroomId
	model.SourceClassroomId = req.SourceClassroomId
	if req.SenderNote != nil {
		model.SenderNote = req.SenderNote
	}
	if req.SenderFiles != nil {
		model.SenderFiles = req.SenderFiles
	}
	if req.ReceiverNote != nil {
		model.ReceiverNote = req.ReceiverNote
	}
	model.SentBy = req.SentBy
	model.ReceivedBy = req.ReceivedBy
	model.Status = req.Status
}

func ConvertSchoolTransferQueryToMap(query SchoolTransferQueryDto) (m map[string]interface{}) {
	m = map[string]interface{}{}

	if query.ID != nil && *query.ID != "" {
		m["id"] = *query.ID
	}
	if v := query.IDs; v != nil && len(v) >= 1 {
		m["ids"] = v
	}
	if query.StudentId != nil && *query.StudentId != "" {
		m["student_id"] = *query.StudentId
	}
	if query.TargetSchoolId != nil && *query.TargetSchoolId != "" {
		m["target_school_id"] = *query.TargetSchoolId
	}
	if query.SourceSchoolId != nil && *query.SourceSchoolId != "" {
		m["source_school_id"] = *query.SourceSchoolId
	}
	if v := query.Limit; v > 0 {
		m["limit"] = v
	} else {
		m["limit"] = 12
	}
	if v := query.Offset; v > 0 {
		m["offset"] = v
	} else {
		m["offset"] = 0
	}
	return
}
