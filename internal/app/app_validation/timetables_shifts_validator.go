package app_validation

import (
	"github.com/mekdep/server/internal/api/utils"
	"github.com/mekdep/server/internal/models"
)

func ValidateShiftsCreate(ses *utils.Session, dto models.ShiftRequest) error {
	errs := AppErrorCollection{}
	errsE := ValidateStruct(dto)
	if errsE != nil {
		errs = errsE.(AppErrorCollection)
	}
	if ses.GetSchoolId() != nil && *ses.GetSchoolId() != "" {
		dto.SchoolId = *ses.GetSchoolId()
	} else if dto.SchoolId == "" {
		errs.Append(*ErrRequired.SetKey("school_id"))
		return errs
	}
	if dto.Name == nil || *dto.Name == "" {
		errs.Append(*ErrRequired.SetKey("name"))
		return errs
	}
	if dto.Value == nil {
		errs.Append(*ErrRequired.SetKey("value"))
		return errs
	}
	if !errs.HasError() {
		return nil
	}
	return errs
}
