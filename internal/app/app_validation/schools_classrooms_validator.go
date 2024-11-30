package app_validation

import (
	"github.com/mekdep/server/internal/api/utils"
	"github.com/mekdep/server/internal/models"
)

func ValidateClassroomsCreate(ses *utils.Session, dto models.ClassroomRequest) error {
	errs := AppErrorCollection{}
	errsE := ValidateStruct(dto)
	if errsE != nil {
		errs = errsE.(AppErrorCollection)
	}
	if ses.GetSchoolId() != nil {
		schoolId := new(string)
		schoolId = ses.GetSchoolId()
		dto.SchoolId = schoolId
	} else if dto.SchoolId == nil {
		errs.Append(*ErrRequired.SetKey("school_id"))
		return errs
	}
	if dto.Name == nil || *dto.Name == "" {
		errs.Append(*ErrRequired.SetKey("name"))
		return errs
	}
	if !errs.HasError() {
		return nil
	}
	return errs
}
