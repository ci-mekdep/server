package app_validation

import (
	"github.com/mekdep/server/internal/api/utils"
	"github.com/mekdep/server/internal/models"
)

func ValidateNotificationsItemCreate(ses *utils.Session, dto models.UserNotificationRequest) error {
	errs := AppErrorCollection{}
	errsE := ValidateStruct(dto)
	if errsE != nil {
		errs = errsE.(AppErrorCollection)
	}
	if dto.UserId == "" {
		errs.Append(*ErrRequired.SetKey("school_ids"))
		return errs
	}
	if dto.NotificationId == "" {
		errs.Append(*ErrRequired.SetKey("school_ids"))
		return errs
	}
	if !errs.HasError() {
		return nil
	}
	return errs
}
