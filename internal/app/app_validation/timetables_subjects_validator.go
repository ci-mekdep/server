package app_validation

import (
	"github.com/mekdep/server/internal/api/utils"
	"github.com/mekdep/server/internal/models"
)

func ValidateSubjectsCreate(ses *utils.Session, dto models.SubjectRequest) error {
	errs := AppErrorCollection{}
	errsE := ValidateStruct(dto)
	if errsE != nil {
		errs = errsE.(AppErrorCollection)
	}
	if ses.GetSchoolId() != nil {
		dto.SchoolId = *ses.GetSchoolId()
	} else if dto.SchoolId == "" {
		errs.Append(*ErrRequired.SetKey("school_id"))
		return errs
	}
	if dto.ClassroomId == "" {
		errs.Append(*ErrRequired.SetKey("classroom_id"))
		return errs
	}
	if dto.Name == nil {
		errs.Append(*ErrRequired.SetKey("name"))
		return errs
	}
	if dto.WeekHours != nil && *dto.WeekHours < 0 {
		errs.Append(*ErrInvalid.SetKey("week_hours").SetComment("Week hours cannot be a negative value"))
		return errs
	}
	if !errs.HasError() {
		return nil
	}
	return errs
}
