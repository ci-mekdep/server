package app_validation

import (
	"github.com/mekdep/server/internal/api/utils"
	"github.com/mekdep/server/internal/models"
)

func ValidatePeriodsCreate(ses *utils.Session, dto models.PeriodRequest) error {
	errs := AppErrorCollection{}
	errsE := ValidateStruct(dto)
	if errsE != nil {
		errs = errsE.(AppErrorCollection)
	}
	if ses.GetSchoolId() != nil {
		dto.SchoolId = ses.GetSchoolId()
	} else if dto.SchoolId == nil || *dto.SchoolId == "" {
		errs.Append(*ErrRequired.SetKey("school_id"))
		return errs
	}
	if dto.Title == nil || *dto.Title == "" {
		errs.Append(*ErrRequired.SetKey("title"))
		return errs
	}
	if dto.Value == nil && *dto.Value == nil {
		errs.Append(*ErrRequired.SetKey("value"))
		return errs
	}
	if !errs.HasError() {
		return nil
	}
	return errs
}
