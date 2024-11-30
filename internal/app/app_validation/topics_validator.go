package app_validation

import (
	"github.com/mekdep/server/internal/api/utils"
	"github.com/mekdep/server/internal/models"
)

func ValidateTopicsCreate(ses *utils.Session, dto models.TopicsRequest) error {
	errs := AppErrorCollection{}
	errsE := ValidateStruct(dto)
	if errsE != nil {
		errs = errsE.(AppErrorCollection)
	}
	if dto.SubjectName == nil || *dto.SubjectName == "" {
		errs.Append(*ErrRequired.SetKey("subject"))
		return errs
	}
	if dto.Title == nil || *dto.Title == "" {
		errs.Append(*ErrRequired.SetKey("title"))
		return errs
	}
	if dto.Language == nil || *dto.Language == "" {
		errs.Append(*ErrRequired.SetKey("language"))
		return errs
	}

	if !errs.HasError() {
		return nil
	}
	return errs
}
