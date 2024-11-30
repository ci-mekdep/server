package app_validation

import (
	"github.com/mekdep/server/internal/api/utils"
	"github.com/mekdep/server/internal/models"
)

func ValidateSchoolTransfersCreate(ses *utils.Session, dto models.SchoolTransferCreateDto) error {
	errs := AppErrorCollection{}
	errsE := ValidateStruct(dto)
	if errsE != nil {
		errs = errsE.(AppErrorCollection)
	}
	if dto.StudentId == nil || *dto.StudentId == "" {
		errs.Append(*ErrRequired.SetKey("student_id"))
		return errs
	}
	if dto.TargetSchoolId == nil || *dto.TargetSchoolId == "" {
		errs.Append(*ErrRequired.SetKey("target_school_id"))
		return errs
	}
	if dto.SourceClassroomId == nil || *dto.SourceClassroomId == "" {
		errs.Append(*ErrRequired.SetKey("source_classroom_id"))
		return errs
	}
	if dto.SourceSchoolId == nil || *dto.SourceSchoolId == "" {
		errs.Append(*ErrRequired.SetKey("source_school_id"))
	}
	if dto.SentBy == nil || *dto.SentBy == "" {
		errs.Append(*ErrRequired.SetKey("sent_by"))
	}
	if dto.Status == nil || *dto.Status == "" {
		errs.Append(*ErrRequired.SetKey("status"))
	}
	if !errs.HasError() {
		return nil
	}
	return errs
}

func ValidateSchoolTransfersInboxCreate(ses *utils.Session, dto models.SchoolTransferCreateDto) error {
	errs := AppErrorCollection{}
	errsE := ValidateStruct(dto)
	if errsE != nil {
		errs = errsE.(AppErrorCollection)
	}
	if dto.TargetClassroomId == nil || *dto.TargetClassroomId == "" {
		errs.Append(*ErrRequired.SetKey("target_classroom_id"))
		return errs
	}
	if dto.Status == nil || *dto.Status == "" {
		errs.Append(*ErrRequired.SetKey("status"))
		return errs
	}
	if dto.ReceivedBy == nil || *dto.ReceivedBy == "" {
		errs.Append(*ErrRequired.SetKey("received_by"))
		return errs
	}
	if !errs.HasError() {
		return nil
	}
	return errs
}
