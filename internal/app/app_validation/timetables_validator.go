package app_validation

import (
	"github.com/mekdep/server/internal/api/utils"
	"github.com/mekdep/server/internal/models"
)

func ValidateTimetablesCreate(ses *utils.Session, dto models.TimetableRequest) error {
	errs := AppErrorCollection{}
	errsE := ValidateStruct(dto)
	if errsE != nil {
		errs = errsE.(AppErrorCollection)
	}
	if dto.Value == nil {
		errs.Append(*ErrRequired.SetKey("value"))
		return errs
	}
	if ses.GetSchoolId() != nil {
		dto.SchoolId = *ses.GetSchoolId()
	} else if dto.SchoolId == "" {
		errs.Append(*ErrRequired.SetKey("school_id"))
		return errs
	}
	if dto.ClassroomId == "" {
		errs.Append(*ErrRequired.SetKey("classroom_id"))
		return errs
	}
	if dto.ShiftId == nil {
		errs.Append(*ErrRequired.SetKey("shift_id"))
		return errs
	}
	if !errs.HasError() {
		return nil
	}
	return errs
}
