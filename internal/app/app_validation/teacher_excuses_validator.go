package app_validation

import (
	"github.com/mekdep/server/internal/api/utils"
	"github.com/mekdep/server/internal/app"
	"github.com/mekdep/server/internal/models"
)

func ValidateTeacherExcuseCreate(ses *utils.Session, dto models.TeacherExcuseCreateDto) error {
	errs := AppErrorCollection{}
	errsE := ValidateStruct(dto)
	if errsE != nil {
		errs = errsE.(AppErrorCollection)
	}
	if dto.StartDate == nil {
		errs.Append(*ErrRequired.SetKey("start_date"))
		return errs
	}
	if dto.EndDate == nil {
		errs.Append(*ErrRequired.SetKey("end_date"))
		return errs
	}
	if dto.TeacherId == "" {
		errs.Append(*ErrRequired.SetKey("teacher_id"))
		return errs
	}
	if dto.Reason == "" {
		errs.Append(*ErrRequired.SetKey("reason"))
		return errs
	}
	if dto.SchoolId == nil {
		errs.Append(*ErrRequired.SetKey("school_id"))
		return errs
	}
	if (dto.EndDate.Unix()-dto.StartDate.Unix())/60/60/24 > 60 {
		return app.ErrExceeded.SetKey("start_date").SetComment("more than 60 days")
	}
	if !errs.HasError() {
		return nil
	}
	return errs
}

func ValidateTeacherExcuseQuery(ses *utils.Session, dto models.TeacherExcuseQueryDto) error {
	err := ValidateStruct(dto)

	if err != nil {
		return err
	}
	return nil
}
