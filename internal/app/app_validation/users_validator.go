package app_validation

import (
	"slices"
	"strconv"

	"github.com/mekdep/server/internal/api/utils"
	"github.com/mekdep/server/internal/models"
	"github.com/mekdep/server/internal/store"
)

func ValidateUser(ses *utils.Session, dto models.UserRequest, checkUnique bool) error {
	dto.Format()
	errs := AppErrorCollection{}
	errsE := ValidateStruct(dto)
	if errsE != nil {
		errs = errsE.(AppErrorCollection)
	}
	if dto.FirstName == nil || dto.FirstName != nil && *dto.FirstName == "" {
		errs.Append(*ErrRequired.SetKey("first_name"))
	}
	if dto.LastName == nil || dto.LastName != nil && *dto.LastName == "" {
		errs.Append(*ErrRequired.SetKey("last_name"))
	}
	// check available schools
	for k, v := range *dto.SchoolIds {
		(*dto.SchoolIds)[k].SchoolUid = ses.GetSchoolIdByFilter(v.SchoolUid)
	}

	// school_id and role_code required
	if dto.SchoolIds == nil || len(*dto.SchoolIds) < 1 {
		errs.Append(*ErrRequired.SetKey("schools.0.role_code"))
		return errs
	}
	roleCodes := []models.Role{}
	for k, v := range *dto.SchoolIds {
		if v.RoleCode == nil || *v.RoleCode == "" {
			errs.Append(*ErrRequired.SetKey("schools." + strconv.Itoa(k) + ".role_code"))
		}
		if !slices.Contains([]models.Role{
			models.RoleAdmin, models.RoleOperator,
		}, *v.RoleCode) && (v.SchoolUid == nil || *v.SchoolUid == "") {
			errs.Append(*ErrRequired.SetKey("schools." + strconv.Itoa(k) + ".school_id"))
		}
		if models.RoleStudent == *v.RoleCode && (dto.ClassroomIds == nil ||
			len(*dto.ClassroomIds) == 0 ||
			(*dto.ClassroomIds)[0].ClassroomId == nil) && dto.ClassroomName == nil {
			errs.Append(*ErrRequired.SetKey("classrooms.0.classroom_id"))
		}
		roleCodes = append(roleCodes, *v.RoleCode)
	}

	// password required when
	if dto.ID == nil && slices.ContainsFunc([]models.Role{
		models.RoleAdmin,
		models.RoleOrganization,
		models.RolePrincipal}, func(i models.Role) bool {
		return slices.Contains(roleCodes, i)
	}) {
		if dto.Password == nil || dto.Password != nil && len(*dto.Password) < 8 {
			errs.Append(*ErrRequired.SetKey("password").SetComment("password is required and length minimum 8"))
		}
	}
	// phone required when
	if slices.ContainsFunc([]models.Role{
		models.RoleParent,
		models.RoleTeacher}, func(i models.Role) bool {
		return slices.Contains(roleCodes, i)
	}) {
		if dto.Phone == nil || len(*dto.Phone) != 8 {
			errs.Append(*ErrRequired.SetKey("phone").SetComment("phone is required and length must be 8"))
		}
	}

	// check available roles
	for k, v := range *dto.SchoolIds {
		if !CheckAvailableRole(ses.GetUser(), ses.GetRole(), *v.RoleCode) {
			errs.Append(*ErrInvalid.SetKey("role_code").SetComment("invalid role index: " + strconv.Itoa(k)))
		}
	}
	if checkUnique {
		if err := checkUserUniqueness(ses, dto); err != nil {
			errs.Append(*err)
		}
	}

	if !errs.HasError() {
		return nil
	}
	return errs
}

func checkUserUniqueness(ses *utils.Session, model models.UserRequest) *AppError {
	// Check for username uniqueness
	if model.Username != nil && *model.Username != "" {
		sameUsers, _, err := store.Store().UsersFindBy(ses.Context(), models.UserFilterRequest{
			Username: model.Username,
			NotID:    model.ID,
		})
		if err != nil {
			return ErrInvalid.SetComment(err.Error())
		}
		if len(sameUsers) > 0 {
			return ErrUnique.SetKey("username")
		}
	}
	schoolIds := []string{}
	for _, v := range *model.SchoolIds {
		if v.SchoolUid != nil {
			schoolIds = append(schoolIds, *v.SchoolUid)
		}
	}

	// Check for phone uniqueness
	if model.Phone != nil && *model.Phone != "" && model.SchoolIds != nil {
		sameUsers, _, err := store.Store().UsersFindBy(ses.Context(), models.UserFilterRequest{
			Phone:     model.Phone,
			SchoolIds: &schoolIds,
			NotID:     model.ID,
		})
		if err != nil {
			return ErrInvalid.SetComment(err.Error())
		}
		if len(sameUsers) > 0 {
			return ErrUnique.SetKey("phone")
		}
	}

	return nil
}

func CheckAvailableRole(user *models.User, r *models.Role, targetRole models.Role) bool {
	isMain := false
	if *r == models.RoleAdmin && user.Email != nil {
		isMain = slices.Contains([]string{"agoyliq@gmail.com"}, *user.Email)
	} else if *r == models.RolePrincipal && len(user.Schools) > 0 {
		for _, v := range user.Schools {
			if v.School != nil && v.School.AdminUid != nil {
				if *v.School.AdminUid == user.ID {
					isMain = true
					break
				}
			}
		}
	}

	if *r == models.RolePrincipal {
		if isMain {
			return slices.Contains([]models.Role{
				models.RolePrincipal,
				models.RoleTeacher,
				models.RoleStudent,
				models.RoleParent,
			}, targetRole)
		} else {
			return slices.Contains([]models.Role{
				models.RoleTeacher,
				models.RoleStudent,
				models.RoleParent,
			}, targetRole)
		}
	} else if *r == models.RoleOrganization {
		return slices.Contains([]models.Role{
			models.RolePrincipal,
			models.RoleTeacher,
			models.RoleStudent,
			models.RoleParent,
		}, targetRole)
	} else if *r == models.RoleOperator {
		return slices.Contains([]models.Role{
			models.RoleStudent,
			models.RoleParent,
		}, targetRole)
	} else if *r == models.RoleAdmin {
		if isMain {
			return slices.Contains([]models.Role{
				models.RoleAdmin,
				models.RoleOperator,
				models.RoleOrganization,
				models.RolePrincipal,
				models.RoleTeacher,
				models.RoleStudent,
				models.RoleParent,
			}, targetRole)
		} else {
			return slices.Contains([]models.Role{
				models.RoleOperator,
				models.RoleOrganization,
				models.RolePrincipal,
				models.RoleTeacher,
				models.RoleStudent,
				models.RoleParent,
			}, targetRole)
		}
	} else if *r == models.RoleTeacher {
		return slices.Contains([]models.Role{
			models.RoleStudent,
			models.RoleParent,
		}, targetRole)
	}
	return false
}
