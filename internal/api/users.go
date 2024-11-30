package api

import (
	"slices"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/mekdep/server/internal/api/utils"
	"github.com/mekdep/server/internal/app"
	"github.com/mekdep/server/internal/app/app_validation"
	"github.com/mekdep/server/internal/models"
	"github.com/mekdep/server/internal/store"
)

func UserRoutes(api *gin.RouterGroup) {
	userRoutes := api.Group("/users")
	{
		userRoutes.POST("login", UserLogin)
		userRoutes.DELETE("", UserDelete)
		userRoutes.GET("", UserList)
		userRoutes.GET("/values", UserListValues)
		userRoutes.GET(":id", UserDetail)
		userRoutes.PUT(":id", UserUpdate)
		userRoutes.POST("", UserCreate)
		userRoutes.POST("batch", UserCreateBatch)
		userRoutes.GET("me", UserMe)
		userRoutes.PATCH("me", UserMeUpdate)
		userRoutes.POST("me/change/school", UserMeUpdateSchool)
		userRoutes.GET("sessions", UserSessions)
		userRoutes.DELETE("sessions", UserSessionsDelete)
		userRoutes.POST("security", UserPasswordUpdate)
		userRoutes.POST(":id/promote", UserPromote)
	}
}

func userListQuery(ses *utils.Session, data models.UserFilterRequest, IsValue bool) ([]*models.UserResponse, []models.ModelValueResponse, int, error) {
	// access filter by school ids
	if data.SchoolIds != nil && len(*data.SchoolIds) > 0 {
		sids := *data.SchoolIds
		data.SchoolIds = &[]string{}
		for _, sid := range sids {
			if slices.Contains(ses.GetSchoolIds(), sid) {
				*data.SchoolIds = append(*data.SchoolIds, sid)
			}
		}
	}
	// comment
	data.SchoolIds = &[]string{}
	if data.SchoolIds == nil || data.SchoolIds != nil && len(*data.SchoolIds) < 1 {
		*data.SchoolIds = ses.GetSchoolIds()
	}
	if data.SchoolId == nil {
		data.SchoolId = ses.GetSchoolId()
	}

	// access filter by role
	if data.Roles != nil {
		for k, v := range *data.Roles {
			if !app_validation.CheckAvailableRole(ses.GetUser(), ses.GetRole(), models.Role(v)) {
				(*data.Roles)[k] = ""
				// data.Roles = &[]string{}
				// data.Ids = &[]uint{}
			}
		}
	}

	// if filtered by admin then no school filter
	if data.Roles != nil && (slices.Contains(*data.Roles, string(models.RoleAdmin)) ||
		slices.Contains(*data.Roles, string(models.RoleOrganization))) {
		data.SchoolIds = nil
	}
	if *ses.GetRole() == models.RoleTeacher && ses.GetUser().TeacherClassroom != nil {
		data.ClassroomId = &ses.GetUser().TeacherClassroom.ID
	} else if *ses.GetRole() == models.RoleTeacher && ses.GetUser().TeacherClassroom == nil {
		data.ClassroomId = nil
		return []*models.UserResponse{}, nil, 0, nil
	}
	if IsValue {
		res, total, err := app.UsersListValues(ses, data)
		return nil, res, total, err
	} else {
		res, total, err := app.UsersList(ses, data)
		return res, nil, total, err
	}
}

func userAvailableCheck(ses *utils.Session, data models.UserFilterRequest) (bool, error) {
	_, _, t, err := userListQuery(ses, data, false)
	if err != nil {
		return false, err
	}
	if t < 1 && data.Role == nil {
		data.Role = new(string)
		*data.Role = string(models.RoleParent)
		_, _, t, err = userListQuery(ses, data, false)
		if err != nil {
			return false, err
		}
	}
	return t > 0, nil
}

func UserListValues(c *gin.Context) {
	ses := utils.InitSession(c)
	err := app.Ap().UserActionCheckRead(&ses, app.PermAdminUsers, func(user *models.User) (err error) {
		r := models.UserFilterRequest{}
		if errMsg, errKey := BindAndValidate(c, &r); errMsg != "" || errKey != "" {
			return app.NewAppError(errMsg, errKey, "")
		}

		_, users, total, err := userListQuery(&ses, r, true)
		if err != nil {
			return err
		}
		Success(c, gin.H{
			"total": total,
			"users": users,
		})
		return
	})
	if err != nil {
		handleError(c, err)
		return
	}
}
func UserList(c *gin.Context) {
	ses := utils.InitSession(c)
	err := app.Ap().UserActionCheckRead(&ses, app.PermAdminUsers, func(user *models.User) (err error) {
		r := models.UserFilterRequest{}
		if errMsg, errKey := BindAndValidate(c, &r); errMsg != "" || errKey != "" {
			return app.NewAppError(errMsg, errKey, "")
		}
		users, _, total, err := userListQuery(&ses, r, false)
		if err != nil {
			return err
		}
		Success(c, gin.H{
			"total": total,
			"users": users,
		})
		return
	})
	if err != nil {
		handleError(c, err)
		return
	}
}

func UserPromote(c *gin.Context) {
	ses := utils.InitSession(c)
	err := app.Ap().UserActionCheckWrite(&ses, app.PermAdminUsers, func(user *models.User) (err error) {
		id := c.Param("id")
		if err != nil {
			return app.ErrRequired.SetKey("id")
		}
		if id == "" {
			return app.ErrRequired.SetKey("id")
		}

		if ok, err := userAvailableCheck(&ses, models.UserFilterRequest{ID: &id}); err != nil {
			return err
		} else if !ok {
			return app.ErrNotfound
		}
		uu, err := store.Store().UsersFindById(ses.Context(), id)
		if err != nil {
			return err
		}
		err = store.Store().UsersLoadRelations(ses.Context(), &[]*models.User{uu}, true)
		if err != nil {
			return err
		}

		err = app.UserTariffUpgrade(&ses, &models.PaymentTransaction{TariffType: models.PaymentTrial}, *uu, uu.Parents)
		if err != nil {
			return err
		}

		Success(c, gin.H{
			"message": "Successfully upgraded to plus",
		})

		return nil
	})
	if err != nil {
		handleError(c, err)
		return
	}
}

func UserUpdate(c *gin.Context) {
	ses := utils.InitSession(c)
	err := app.Ap().UserActionCheckWrite(&ses, app.PermAdminUsers, func(user *models.User) (err error) {
		if _, err := c.MultipartForm(); err == nil {
			UserUpdateHandleFiles(c)
		} else {
			r := models.UserRequest{}
			if err := BindAny(c, &r); err != nil {
				return err
			}

			id := c.Param("id")
			r.ID = &id
			if *r.ID == "" {
				return app.ErrRequired.SetKey("id")
			}
			if ok, err := userAvailableCheck(&ses, models.UserFilterRequest{ID: r.ID}); err != nil {
				return err
			} else if !ok {
				return app.ErrForbidden
			}

			// filter out by role access
			filteredSchoolIds := []models.UserSchoolRequest{}
			if r.SchoolIds == nil {
				r.SchoolIds = &[]models.UserSchoolRequest{}
			}
			for _, v := range *r.SchoolIds {
				if !app_validation.CheckAvailableRole(ses.GetUser(), ses.GetRole(), *v.RoleCode) {
					// errs.Append(*ErrInvalid.SetKey("role_code").SetComment("invalid role index: " + strconv.Itoa(k)))
				} else {
					filteredSchoolIds = append(filteredSchoolIds, v)
				}
			}
			r.SchoolIds = &filteredSchoolIds

			// validate
			err = app_validation.ValidateUser(&ses, r, true)
			if err != nil {
				return err
			}
			if r.Parents != nil {
				for i, parent := range *r.Parents {
					if err := app_validation.ValidateUser(&ses, parent, true); err != nil {
						appErr := err.(app_validation.AppErrorCollection)
						is := strconv.Itoa(i)
						for k, _ := range appErr.Errors {
							return appErr.Errors[k].SetKey("parents." + is + "." + appErr.Errors[k].Key())
						}

					}
				}
			}

			if r.Children != nil {
				for i, child := range *r.Children {
					if err := app_validation.ValidateUser(&ses, child, true); err != nil {
						appErr := err.(app_validation.AppErrorCollection)
						in := strconv.Itoa(i)
						for k, _ := range appErr.Errors {
							return appErr.Errors[k].SetKey("children" + "." + in + appErr.Errors[k].Key())
						}
					}
				}
			}

			// update
			res, err := app.UsersUpdate(&ses, &r)
			if err != nil {
				return err
			}
			if r.Parents != nil {
				for _, parent := range *r.Parents {
					// Create or Update parent user
					if parent.ID != nil {
						_, err := app.UsersUpdate(&ses, &parent)
						if err != nil {
							return err
						}
					} else {
						parent.ChildIds = &[]string{res.ID}
						_, _, err := app.UsersCreate(&ses, &parent)
						if err != nil {
							return err
						}
					}
				}
			}
			if r.Children != nil {
				for _, child := range *r.Children {
					// Create or Update child user
					if child.ID != nil {
						_, err := app.UsersUpdate(&ses, &child)
						if err != nil {
							return err
						}
					} else {
						child.ParentIds = &[]string{res.ID}
						_, _, err := app.UsersCreate(&ses, &child)
						if err != nil {
							return err
						}
					}
				}
			}

			userLog(models.UserLog{
				SchoolId:          ses.GetSchoolId(),
				SessionId:         ses.GetSessionId(),
				UserId:            user.ID,
				SubjectId:         r.ID,
				Subject:           models.LogSubjectUsers,
				SubjectAction:     models.LogActionUpdate,
				SubjectProperties: r,
			})
			Success(c, gin.H{
				"user": res,
			})
		}
		return
	})
	if err != nil {
		handleError(c, err)
		return
	}
}

func UserUpdateHandleFiles(c *gin.Context) {
	ses := utils.InitSession(c)
	err := app.Ap().UserActionCheckWrite(&ses, app.PermAdminUsers, func(user *models.User) (err error) {
		r := models.UserRequest{}
		if err := BindAny(c, &r); err != nil {
			return err
		}

		id := c.Param("id")
		if err != nil {
			return app.ErrRequired.SetKey("id")
		}
		r.ID = &id
		if *r.ID == "" {
			return app.ErrRequired.SetKey("id")
		}

		if ok, err := userAvailableCheck(&ses, models.UserFilterRequest{ID: r.ID}); err != nil {
			return err
		} else if !ok {
			return app.ErrForbidden
		}

		var model models.User
		if model.DocumentFiles == nil {
			model.DocumentFiles = new([]string)
		}

		documentFiles, err := handleFilesUpload(c, "document_files", "users")
		if err != nil {
			return app.NewAppError(err.Error(), "document_files", "")
		}
		r.DocumentFiles = &documentFiles

		avatar, _, err := handleFileUpload(c, "avatar", "users", true)
		if err != nil {
			return app.NewAppError(err.Error(), "avatar", "")
		}
		if avatar != "" {
			r.Avatar = &avatar
		}

		if r.AvatarDelete != nil && *r.AvatarDelete {
			r.Avatar = new(string)
		}

		// TODO: separate update files (because notnull fields get nulled, atomic update)
		res, err := app.UsersUpdateFile(&ses, &r)
		if err != nil {
			return err
		}
		userLog(models.UserLog{
			SchoolId:          ses.GetSchoolId(),
			SessionId:         ses.GetSessionId(),
			UserId:            user.ID,
			SubjectId:         r.ID,
			Subject:           models.LogSubjectUsers,
			SubjectAction:     models.LogActionUpdate,
			SubjectProperties: r,
		})
		Success(c, gin.H{
			"user": res,
		})
		return
	})
	if err != nil {
		handleError(c, err)
		return
	}
}

func UserDetail(c *gin.Context) {
	ses := utils.InitSession(c)
	err := app.Ap().UserActionCheckRead(&ses, app.PermAdminUsers, func(user *models.User) (err error) {
		id := c.Param("id")
		if err != nil {
			return app.ErrRequired.SetKey("id")
		}
		if id == "" {
			return app.ErrRequired.SetKey("id")
		}

		args := models.UserFilterRequest{
			ID: &id,
		}
		if ses.GetRole() != nil && *ses.GetRole() != models.RoleAdmin {
			args.SchoolId = ses.GetSchoolId()
		}
		if ok, err := userAvailableCheck(&ses, args); err != nil {
			return err
		} else if !ok {
			return app.ErrNotfound
		}

		res, err := app.UsersDetail(&ses, args)
		if err != nil {
			return err
		}
		Success(c, gin.H{
			"user": res,
		})
		return
	})
	if err != nil {
		handleError(c, err)
		return
	}
}

func UserCreate(c *gin.Context) {
	ses := utils.InitSession(c)
	err := app.Ap().UserActionCheckWrite(&ses, app.PermAdminUsers, func(user *models.User) (err error) {
		if _, err := c.MultipartForm(); err == nil {
			UserCreateHandleFiles(c)
		} else {
			r := models.UserRequest{}
			if err := BindAny(c, &r); err != nil {
				return err
			}

			// validate
			err = app_validation.ValidateUser(&ses, r, true)
			if err != nil {
				return err
			}

			if r.Parents != nil {
				for i, parent := range *r.Parents {
					if err := app_validation.ValidateUser(&ses, parent, true); err != nil {
						appErr := err.(app_validation.AppErrorCollection)
						is := strconv.Itoa(i)
						for k, _ := range appErr.Errors {
							return appErr.Errors[k].SetKey("parents." + is + "." + appErr.Errors[k].Key())
						}

					}
				}
			}

			if r.Children != nil {
				for i, child := range *r.Children {
					if err := app_validation.ValidateUser(&ses, child, true); err != nil {
						appErr := err.(app_validation.AppErrorCollection)
						in := strconv.Itoa(i)
						for k, _ := range appErr.Errors {
							return appErr.Errors[k].SetKey("children." + in + "." + appErr.Errors[k].Key())
						}
					}
				}
			}

			res, _, err := app.UsersCreate(&ses, &r)
			if err != nil {
				return err
			}
			// Validate and collect Parents
			if r.Parents != nil {
				for _, parent := range *r.Parents {
					parent.ChildIds = &[]string{res.ID}
					// Create or Update parent user
					if parent.ID != nil {
						_, err := app.UsersUpdate(&ses, &parent)
						if err != nil {
							return err
						}
					} else {
						_, _, err := app.UsersCreate(&ses, &parent)
						if err != nil {
							return err
						}
					}
				}
			}

			// Validate and collect Children
			if r.Children != nil {
				for _, child := range *r.Children {
					child.ParentIds = &[]string{res.ID}
					// Create or Update child user
					if child.ID != nil {
						_, err := app.UsersUpdate(&ses, &child)
						if err != nil {
							return err
						}
					} else {
						_, _, err := app.UsersCreate(&ses, &child)
						if err != nil {
							return err
						}
					}
				}
			}

			userLog(models.UserLog{
				SchoolId:          ses.GetSchoolId(),
				SessionId:         ses.GetSessionId(),
				UserId:            user.ID,
				SubjectId:         &res.ID,
				Subject:           models.LogSubjectUsers,
				SubjectAction:     models.LogActionCreate,
				SubjectProperties: r,
			})
			Success(c, gin.H{
				"user": res,
			})
		}
		return nil
	})
	if err != nil {
		handleError(c, err)
		return
	}
}

func UserCreateHandleFiles(c *gin.Context) {
	ses := utils.InitSession(c)
	err := app.Ap().UserActionCheckWrite(&ses, app.PermAdminUsers, func(user *models.User) (err error) {
		r := models.UserRequest{}
		if err := BindAny(c, &r); err != nil {
			return err
		}

		documentFiles, err := handleFilesUpload(c, "document_files", "users")
		if err != nil {
			return app.NewAppError(err.Error(), "documentFiles", "")
		}

		if documentFiles != nil {
			r.DocumentFiles = &documentFiles
		}
		// handle avatar
		avatar, _, err := handleFileUpload(c, "avatar", "users", true)
		if err != nil {
			return app.NewAppError(err.Error(), "avatar", "")
		}
		if avatar != "" {
			r.Avatar = &avatar
		}

		res, _, err := app.UsersCreate(&ses, &r)
		if err != nil {
			return err
		}
		userLog(models.UserLog{
			SchoolId:          ses.GetSchoolId(),
			SessionId:         ses.GetSessionId(),
			UserId:            user.ID,
			SubjectId:         &res.ID,
			Subject:           models.LogSubjectUsers,
			SubjectAction:     models.LogActionCreate,
			SubjectProperties: r,
		})
		Success(c, gin.H{
			"user": res,
		})
		return nil
	})
	if err != nil {
		handleError(c, err)
		return
	}
}

func UserCreateBatch(c *gin.Context) {
	ses := utils.InitSession(c)
	err := app.Ap().UserActionCheckWrite(&ses, app.PermAdminUsers, func(user *models.User) error {
		dto := models.UserCollectionRequest{}
		if err := BindAny(c, &dto); err != nil {
			return err
		}
		users := []*models.UserResponse{}

		errs := app_validation.AppErrorCollection{}

		totalCreated := 0
		for i, _ := range dto.Users {
			dto.Users[i].Format()
		}

		for i, userDto := range dto.Users {
			i = i + 1
			keyPrefix := "users." + strconv.Itoa(i) + "."
			if userDto.FirstName != nil && strings.Contains(*userDto.FirstName, " ") {
				errs.Append(*app_validation.NewAppError(keyPrefix+"first_name", "First name must be a single word", ""))
			}
			if userDto.LastName != nil && strings.Contains(*userDto.LastName, " ") {
				errs.Append(*app_validation.NewAppError(keyPrefix+"last_name", "Last name must be a single word without spaces", ""))
			}
			if userDto.MiddleName != nil && strings.Contains(*userDto.MiddleName, " ") {
				errs.Append(*app_validation.NewAppError(keyPrefix+"middle_name", "Middle name must be a single word without spaces", ""))
			}
			// validate
			err := app_validation.ValidateUser(&ses, userDto, false)
			if errA, ok := err.(app_validation.AppErrorCollection); err != nil && ok {
				for k, _ := range errA.Errors {
					errA.Errors[k].SetKey(keyPrefix + errA.Errors[k].Key())
				}
				errs.Merge(errA)
				continue
			} else if err != nil {
				return err
			}
			if userDto.Parents != nil {
				for ii, parentDto := range *userDto.Parents {
					ii = ii + 1
					subKeyPrefix := keyPrefix + "parents." + strconv.Itoa(ii) + "."
					err := app_validation.ValidateUser(&ses, parentDto, false)
					if errA, ok := err.(app_validation.AppErrorCollection); err != nil && ok {
						for k, _ := range errA.Errors {
							errA.Errors[k].SetKey(subKeyPrefix + errA.Errors[k].Key())
						}
						errs.Merge(errA)
						continue
					} else if err != nil {
						return err
					}
				}
			}
		}
		if errs.HasError() {
			return errs
		}

		for _, userDto := range dto.Users {
			if *(*userDto.SchoolIds)[0].RoleCode == models.RoleStudent {
				if userDto.ClassroomName == nil || *userDto.ClassroomName == "" {
					errs.Append(*app_validation.NewAppError("classroom_name", "", ""))
					continue
				}
				classrooms, _, err := store.Store().ClassroomsFindBy(ses.Context(), models.ClassroomFilterRequest{
					SchoolId: ses.GetSchoolId(),
					Name:     userDto.ClassroomName,
				})
				if err != nil {
					errs.Append(*app_validation.NewAppError("classroom_name", err.Error(), ""))
					continue
				}
				if len(classrooms) != 1 {
					errs.Append(*app_validation.ErrInvalid.SetKey("classroom_name").SetComment("Found " + strconv.Itoa(len(classrooms))))
					continue
				}
				userDto.ClassroomIds = &[]models.UserClassroomRequest{
					{
						ClassroomId: &classrooms[0].ID,
					},
				}
			}
			// create
			userResponse, isCreated, err := app.UsersCreate(&ses, &userDto)
			if err != nil {
				return err
			}
			users = append(users, userResponse)
			if isCreated {
				totalCreated++
			}

			// create parents
			if userDto.Parents != nil {
				for ii, parentDto := range *userDto.Parents {
					ii = ii + 1
					parentDto.ChildIds = &[]string{userResponse.ID}

					userResponse, isCreated, err := app.UsersCreate(&ses, &parentDto)
					if err != nil {
						return err
					}
					users = append(users, userResponse)
					if isCreated {
						totalCreated++
					}
				}
			}

			userLog(models.UserLog{
				SchoolId:          ses.GetSchoolId(),
				SessionId:         ses.GetSessionId(),
				UserId:            user.ID,
				SubjectId:         &userResponse.ID,
				Subject:           models.LogSubjectUsers,
				SubjectAction:     models.LogActionCreate,
				SubjectProperties: userDto,
			})
		}
		Success(c, gin.H{
			"users":         users,
			"total_created": totalCreated,
		})
		return nil
	})
	if err != nil {
		handleError(c, err)
		return
	}
}

func UserDelete(c *gin.Context) {
	ses := utils.InitSession(c)
	err := app.Ap().UserActionCheckWrite(&ses, app.PermAdminUsers, func(user *models.User) (err error) {
		// binds
		if *ses.GetRole() == models.RoleTeacher {
			return app.ErrForbidden
		}
		var ids []string
		var idroles app.DeleteUserRoleQueryList
		c.BindQuery(&idroles)
		if len(idroles.UserRoles) > 0 {
			ids = []string{}
			for _, v := range idroles.UserRoles {
				ids = append(ids, v.UserId)
			}
		} else if len(c.QueryArray("ids")) > 0 {
			ids = []string{}
			idroles = app.DeleteUserRoleQueryList{
				UserRoles: []app.DeleteUserRoleQuery{},
			}
			for _, id := range c.QueryArray("ids") {
				ids = append(ids, id)
				idroles.UserRoles = append(idroles.UserRoles, app.DeleteUserRoleQuery{
					UserId:   id,
					RoleCode: string(models.RoleTeacher),
				})
				idroles.UserRoles = append(idroles.UserRoles, app.DeleteUserRoleQuery{
					UserId:   id,
					RoleCode: string(models.RoleParent),
				})
				idroles.UserRoles = append(idroles.UserRoles, app.DeleteUserRoleQuery{
					UserId:   id,
					RoleCode: string(models.RoleStudent),
				})
				idroles.UserRoles = append(idroles.UserRoles, app.DeleteUserRoleQuery{
					UserId:   id,
					RoleCode: string(models.RolePrincipal),
				})
			}
		} else {
			return app.ErrRequired.SetKey("user_roles")
		}

		// check and school_id fill
		if ok, err := userAvailableCheck(&ses, models.UserFilterRequest{Ids: &ids}); err != nil {
			return err
		} else if !ok {
			return app.ErrForbidden.SetComment("not found")
		}
		for k, _ := range idroles.UserRoles {
			if idroles.UserRoles[k].SchoolId == "" && ses.GetSchoolId() != nil {
				idroles.UserRoles[k].SchoolId = *ses.GetSchoolId()
			}
			if !slices.Contains(ses.GetSchoolIds(), idroles.UserRoles[k].SchoolId) {
				return app.ErrForbidden
			}
		}

		// delete
		deletedUsers, err := app.UsersDelete(&ses, idroles.UserRoles)
		if err != nil {
			return err
		}
		for _, id := range ids {
			userLog(models.UserLog{
				SchoolId:          ses.GetSchoolId(),
				SessionId:         ses.GetSessionId(),
				UserId:            user.ID,
				SubjectId:         &id,
				Subject:           models.LogSubjectUsers,
				SubjectAction:     models.LogActionDelete,
				SubjectProperties: ids,
			})
		}
		Success(c, gin.H{
			"deleted": deletedUsers,
		})
		return
	})
	if err != nil {
		handleError(c, err)
		return
	}
}

func UserMe(c *gin.Context) {
	ses := utils.InitSession(c)
	err := app.Ap().UserActionCheckRead(&ses, app.PermUser, func(user *models.User) (err error) {
		res, fields, err := app.Ap().UserMe(&ses)
		if err != nil {
			return err
		}
		var schoolRes *models.SchoolResponse
		if ses.GetSchool() != nil {
			schoolRes = &models.SchoolResponse{}
			_ = store.Store().SchoolsLoadRelations(ses.Context(), &[]*models.School{ses.GetSchool()})
			schoolRes.FromModel(ses.GetSchool())
		}
		Success(c, gin.H{
			"user":                 res,
			"current_role":         res.Role,
			"current_school":       res.SchoolId,
			"current_school_model": schoolRes,
			"edit_fields":          fields,
		})
		return
	})
	if err != nil {
		handleError(c, err)
		return
	}
}

// TODO: refactor this to separate logic, split into "changeRole"
func UserMeUpdate(c *gin.Context) {
	ses := utils.InitSession(c)
	err := app.Ap().UserActionCheckRead(&ses, app.PermUser, func(user *models.User) (err error) {
		dto := models.UserProfileUpdateRequest{}
		if errMsg, errKey := BindAndValidate(c, &dto); errMsg != "" || errKey != "" {
			return app.NewAppError(errMsg, errKey, "")
		}

		avatar, _, err := handleFileUpload(c, "avatar", "users", true)
		if err != nil {
			return app.NewAppError(err.Error(), "avatar", "")
		}

		if avatar != "" {
			dto.Avatar = &avatar
		}
		if dto.AvatarDelete != nil && *dto.AvatarDelete {
			dto.Avatar = new(string)
		}
		// allowed to update
		res, err := app.UserMeUpdate(&ses, &dto)
		if err != nil {
			return err
		}
		Success(c, gin.H{
			"user":           res,
			"current_role":   res.Role,
			"current_school": res.SchoolId,
		})
		return
	})
	if err != nil {
		handleError(c, err)
		return
	}
}

func UserMeUpdateSchool(c *gin.Context) {
	ses := utils.InitSession(c)
	err := app.Ap().UserActionCheckRead(&ses, app.PermUser, func(user *models.User) (err error) {
		dto := models.UserMeUpdateSchoolRequest{}
		if errMsg, errKey := BindAndValidate(c, &dto); errMsg != "" || errKey != "" {
			return app.NewAppError(errMsg, errKey, "")
		}
		claims := jwt.MapClaims{}
		var sesId *string
		var userSchool *models.UserSchool
		if dto.RoleCode != nil {
			if dto.SchoolId != nil && *dto.SchoolId == "" {
				dto.SchoolId = nil
			}
			if dto.SchoolId == nil {
				dto.SchoolId = dto.RegionId
			}
			if *dto.SchoolId == "" {
				dto.SchoolId = nil
			}
			if userSchool = ses.GetAvailableRole(models.Role(*dto.RoleCode), dto.SchoolId); userSchool != nil {
				// delete current token
				if sId := ses.GetSessionId(); sId != nil {
					utils.SessionDelete(models.Session{
						ID: *sId,
					})
				}
				if dto.SchoolId == nil {
					userSchool.SchoolUid = dto.RegionId
				}
				// new token
				claims, sesId, err = utils.GenerateToken(
					c,
					*user,
					(*models.Role)(dto.RoleCode),
					userSchool.SchoolUid,
					dto.PeriodId,
					nil,
				)
				if err != nil {
					return err
				}
			} else {
				return app.ErrForbidden.SetKey("role_code")
			}
		}
		userLog(models.UserLog{
			SchoolId:          ses.GetSchoolId(),
			SessionId:         ses.GetSessionId(),
			UserId:            user.ID,
			Subject:           models.LogSubjectUsers,
			SubjectAction:     models.LogActionUpdateProfile,
			SubjectProperties: dto,
		})
		if userSchool != nil {
			var schoolRes *models.SchoolResponse
			if userSchool.School != nil {
				schoolRes = &models.SchoolResponse{}
				_ = store.Store().SchoolsLoadRelations(ses.Context(), &[]*models.School{userSchool.School})
				schoolRes.FromModel(userSchool.School)
			}
			var currentSchoolModel *models.SchoolResponse
			var currentRegionModel *models.SchoolResponse
			var currentPeriodModel *models.PeriodResponse
			if userSchool.School != nil {
				currentSchoolModel = new(models.SchoolResponse)
				currentSchoolModel.FromModel(userSchool.School)
			}
			if currentSchoolModel != nil && currentSchoolModel.Parent == nil {
				currentRegionModel = currentSchoolModel
				currentSchoolModel = nil
			}
			if dto.PeriodId != nil && *dto.PeriodId == "" {
				dto.PeriodId = nil
			}
			if dto.PeriodId != nil {
				period, err := store.Store().PeriodsFindById(ses.Context(), *dto.PeriodId)
				if err != nil {
					return err
				}
				if period != nil {
					currentPeriodModel = new(models.PeriodResponse)
					currentPeriodModel.FromModel(period)
				}
			}
			// TODO: refactor below
			userRes := models.UserResponse{}
			userRes.FromModel(user)
			resp := UserLoginResponse{
				User:               &userRes,
				CurrentRole:        &userSchool.RoleCode,
				CurrentSchoolId:    userSchool.SchoolUid,
				CurrentSchoolModel: currentSchoolModel,
				CurrentRegionModel: currentRegionModel,
				CurrentPeriodModel: currentPeriodModel,
				SessionId:          sesId,
				Token:              (claims["token"]).(string),
				ExpiresAt:          (claims["exp"]).(int64),
			}
			if resp.CurrentSchoolId != nil && *resp.CurrentSchoolId == "" {
				resp.CurrentSchoolId = nil
			}
			Success(c, gin.H{
				"user":                 resp.User,
				"current_role":         resp.CurrentRole,
				"current_school":       resp.CurrentSchoolId,
				"current_school_model": resp.CurrentSchoolModel,
				"current_region_model": resp.CurrentRegionModel,
				"current_period_model": resp.CurrentPeriodModel,
				"session_id":           resp.SessionId,
				"new_token":            resp.Token, // TODO: rename to token
				"expires_at":           resp.ExpiresAt,
			})
		} else {
			Success(c, gin.H{
				"new_token":  claims["token"],
				"session_id": sesId,
				"expires_at": claims["exp"],
			})
		}
		return
	})
	if err != nil {
		handleError(c, err)
		return
	}
}

func UserPasswordUpdate(c *gin.Context) {
	ses := utils.InitSession(c)
	err := app.Ap().UserActionCheckWrite(&ses, app.PermUser, func(user *models.User) (err error) {
		r := models.UserPasswordUpdateRequest{}
		if errMsg, errKey := BindAndValidate(c, &r); errMsg != "" || errKey != "" {
			return app.NewAppError(errMsg, errKey, "")
		}

		res, err := app.UserPasswordUpdate(&ses, r)
		if err != nil {
			return err
		}
		userLog(models.UserLog{
			SchoolId:          ses.GetSchoolId(),
			SessionId:         ses.GetSessionId(),
			UserId:            user.ID,
			Subject:           models.LogSubjectUsers,
			SubjectAction:     models.LogActionUpdatePassword,
			SubjectProperties: r,
		})
		Success(c, gin.H{
			"password": res,
		})
		return
	})
	if err != nil {
		handleError(c, err)
		return
	}
}

func UserSessions(c *gin.Context) {
	ses := utils.InitSession(c)
	err := app.Ap().UserActionCheckRead(&ses, app.PermUser, func(user *models.User) error {
		ss, err := app.Ap().UserSessions(&ses)
		if err != nil {
			return err
		}

		Success(c, gin.H{
			"session": ss,
		})
		return err
	})
	if err != nil {
		handleError(c, err)
		return
	}

}

func UserSessionsDelete(c *gin.Context) {
	ses := utils.InitSession(c)
	err := app.Ap().UserActionCheckRead(&ses, app.PermUser, func(user *models.User) error {
		sesIds := c.QueryArray("ids")
		err := app.Ap().UserSessionsDelete(&ses, sesIds)
		if err != nil {
			return err
		}
		userLog(models.UserLog{
			SchoolId:          ses.GetSchoolId(),
			SessionId:         ses.GetSessionId(),
			UserId:            user.ID,
			Subject:           models.LogSubjectUsers,
			SubjectAction:     models.LogActionLogout,
			SubjectProperties: sesIds,
		})
		Success(c, gin.H{})
		return err
	})
	if err != nil {
		handleError(c, err)
		return
	}
}
