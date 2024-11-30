package app_validation

import (
	"github.com/mekdep/server/internal/api/utils"
	"github.com/mekdep/server/internal/models"
)

func ValidatePaymentsCreate(ses *utils.Session, dto models.PaymentCheckoutRequest) error {
	errs := AppErrorCollection{}
	errsE := ValidateStruct(dto)
	if errsE != nil {
		errs = errsE.(AppErrorCollection)
	}

	if !errs.HasError() {
		return nil
	}
	return errs
}
