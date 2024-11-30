package app_validation

import (
	"github.com/mekdep/server/internal/api/utils"
	"github.com/mekdep/server/internal/models"
)

func ValidateNotificationsCreate(ses *utils.Session, dto models.NotificationsRequest) error {
	errs := AppErrorCollection{}
	errsE := ValidateStruct(dto)
	if errsE != nil {
		errs = errsE.(AppErrorCollection)
	}
	if dto.SchoolIds == nil {
		errs.Append(*ErrRequired.SetKey("school_ids"))
		return errs
	}
	if dto.Title == nil {
		errs.Append(*ErrRequired.SetKey("title"))
		return errs
	}
	if !errs.HasError() {
		return nil
	}
	return errs
}
