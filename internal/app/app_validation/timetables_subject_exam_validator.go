package app_validation

import (
	"github.com/mekdep/server/internal/api/utils"
	"github.com/mekdep/server/internal/models"
)

func ValidateSubjectExamCreate(ses *utils.Session, dto models.SubjectExamRequest) error {
	errs := AppErrorCollection{}
	errsE := ValidateStruct(dto)
	if errsE != nil {
		errs = errsE.(AppErrorCollection)
	}
	if dto.TeacherId == nil {
		errs.Append(*ErrRequired.SetKey("teacher_id"))
		return errs
	}
	if !errs.HasError() {
		return nil
	}
	return errs
}
