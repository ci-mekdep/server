package app_validation

import (
	"github.com/mekdep/server/internal/api/utils"
	"github.com/mekdep/server/internal/models"
)

func ValidateReportsCreate(ses *utils.Session, dto models.ReportsRequest) error {
	errs := AppErrorCollection{}
	errsE := ValidateStruct(dto)
	if errsE != nil {
		errs = errsE.(AppErrorCollection)
	}
	if dto.Title == "" {
		errs.Append(*ErrRequired.SetKey("title"))
		return errs
	}
	if dto.ValueTypes == nil {
		errs.Append(*ErrRequired.SetKey("value_types"))
		return errs
	}
	if dto.SchoolIds == nil {
		errs.Append(*ErrRequired.SetKey("school_ids"))
		return errs
	}
	if !errs.HasError() {
		return nil
	}
	return errs
}
