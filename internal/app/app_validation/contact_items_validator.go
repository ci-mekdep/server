package app_validation

import (
	"github.com/mekdep/server/internal/api/utils"
	"github.com/mekdep/server/internal/models"
)

func ValidateContactItemsCreate(ses *utils.Session, dto models.ContactItemsRequest) error {
	errs := AppErrorCollection{}
	errsE := ValidateStruct(dto)
	if errsE != nil {
		errs = errsE.(AppErrorCollection)
	}
	if dto.Children != nil && len(dto.Children) < 1 {
		errs.Append(*ErrRequired.SetKey("children").SetComment("children cannot be empty"))
		return errs
	}
	if !errs.HasError() {
		return nil
	}
	return errs
}
