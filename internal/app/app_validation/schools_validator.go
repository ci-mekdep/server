package app_validation

import (
	"github.com/mekdep/server/internal/api/utils"
	"github.com/mekdep/server/internal/models"
)

func ValidateSchoolsCreate(ses *utils.Session, dto models.SchoolRequest) error {
	errs := AppErrorCollection{}
	errsE := ValidateStruct(dto)
	if errsE != nil {
		errs = errsE.(AppErrorCollection)
	}
	if dto.Code == nil || *dto.Code == "" {
		errs.Append(*ErrRequired.SetKey("code"))
		return errs
	}
	if dto.Name == nil {
		errs.Append(*ErrRequired.SetKey("name"))
		return errs
	}
	if dto.ParentUid == nil {
		errs.Append(*ErrRequired.SetKey("parent_id"))
	}
	if !errs.HasError() {
		return nil
	}
	return errs
}
