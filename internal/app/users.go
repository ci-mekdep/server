package app

import (
	"slices"

	"github.com/mekdep/server/internal/api/utils"
	"github.com/mekdep/server/internal/models"
	"github.com/mekdep/server/internal/store"
	"go.elastic.co/apm/v2"
	"golang.org/x/crypto/bcrypt"
)

func UsersList(ses *utils.Session, data models.UserFilterRequest) ([]*models.UserResponse, int, error) {
	sp, ctx := apm.StartSpan(ses.Context(), "UsersList", "app")
	ses.SetContext(ctx)
	defer sp.End()
	l, total, err := store.Store().UsersFindBy(ses.Context(), data)
	if err != nil {
		return nil, 0, err
	}
	if !data.IsValuesOnly {
		err = store.Store().UsersLoadRelations(ses.Context(), &l, false)
		if err != nil {
			return nil, 0, err
		}
	}
	res := []*models.UserResponse{}
	for _, m := range l {
		item := models.UserResponse{}
		item.FromModel(m)
		res = append(res, &item)
	}
	return res, total, nil
}

func UsersListValues(ses *utils.Session, data models.UserFilterRequest) ([]models.ModelValueResponse, int, error) {
	sp, ctx := apm.StartSpan(ses.Context(), "UsersListValues", "app")
	ses.SetContext(ctx)
	defer sp.End()
	data.IsValuesOnly = true
	res, total, err := UsersList(ses, data)
	if err != nil {
		return nil, 0, err
	}
	values := []models.ModelValueResponse{}
	for _, v := range res {
		values = append(values, v.ToValues())
	}
	return values, total, nil
}

func UsersDetail(ses *utils.Session, f models.UserFilterRequest) (*models.UserResponse, error) {
	sp, ctx := apm.StartSpan(ses.Context(), "UsersDetail", "app")
	ses.SetContext(ctx)
	defer sp.End()
	m, _, err := store.Store().UsersFindBy(ses.Context(), f)
	if err != nil {
		return nil, err
	}
	err = store.Store().UsersLoadRelations(ses.Context(), &m, true)
	if err != nil {
		return nil, err
	}

	mm := m[0]

	tmp := mm.Classrooms
	mm.Classrooms = []*models.UserClassroom{}
	for _, v := range tmp {
		if v.Type == nil || v.TypeKey == nil {
			mm.Classrooms = append(mm.Classrooms, v)
		}
	}

	if len(m) < 1 {
		return nil, ErrNotfound
	}
	res := &models.UserResponse{}
	res.FromModel(m[0])
	return res, nil
}

// response roles x2
// teacher_classroom_uid update not working
func UsersUpdate(ses *utils.Session, data *models.UserRequest) (*models.UserResponse, error) {
	sp, ctx := apm.StartSpan(ses.Context(), "UsersUpdate", "app")
	ses.SetContext(ctx)
	defer sp.End()
	existingUser, err := store.Store().UsersFindById(ses.Context(), *data.ID)
	if err != nil {
		return nil, err
	}
	err = store.Store().UsersLoadRelations(ses.Context(), &[]*models.User{existingUser}, true)
	if err != nil {
		return nil, err
	}

	dataModel := &models.User{}
	data.ToModel(dataModel)
	// ????
	dataModel.Schools = existingUser.Schools
	dataModel.SchoolId = ses.GetSchoolId()

	// add new roles
	newRoles := []*models.UserSchool{}
	if data.SchoolIds != nil {
		for _, v := range *data.SchoolIds {
			isExists := false
			for _, vv := range existingUser.Schools {
				if *v.RoleCode == vv.RoleCode && (v.SchoolUid == vv.SchoolUid ||
					v.SchoolUid != nil && vv.SchoolUid != nil && *v.SchoolUid == *vv.SchoolUid) {
					isExists = true
				}
			}
			if !isExists {
				newRoles = append(newRoles, &models.UserSchool{
					RoleCode:  *v.RoleCode,
					SchoolUid: v.SchoolUid,
				})
			}
		}
		dataModel.Schools = append(dataModel.Schools, newRoles...)
	}

	// delete roles
	filteredSchools := []*models.UserSchool{}
	if dataModel.Schools != nil {
		for _, vv := range dataModel.Schools {
			tobeDeleted := false
			if data.SchoolIds != nil {
				for _, v := range *data.SchoolIds {
					if v.IsDelete != nil && *v.IsDelete && *v.RoleCode == vv.RoleCode && (v.SchoolUid == vv.SchoolUid ||
						v.SchoolUid != nil && vv.SchoolUid != nil && *v.SchoolUid == *vv.SchoolUid) {
						tobeDeleted = true
						break
					}
				}
			}
			if !tobeDeleted {
				filteredSchools = append(filteredSchools, vv)
			}
		}
		dataModel.Schools = filteredSchools
	}

	// dataModel.Classrooms = existingUser.Classrooms
	err = encryptPassword(dataModel, data)
	if err != nil {
		return nil, err
	}

	model, err := store.Store().UserUpdate(ses.Context(), dataModel)
	if err != nil {
		return nil, err
	}
	model, err = store.Store().UserUpdateRelations(ses.Context(), dataModel)
	if err != nil {
		return nil, err
	}
	err = store.Store().UsersLoadRelations(ses.Context(), &[]*models.User{model}, true)
	if err != nil {
		return nil, err
	}
	UserCacheClear(ses, model.ID)
	for _, v := range model.Parents {
		UserCacheClear(ses, v.ID)
	}
	res := &models.UserResponse{}
	res.FromModel(model)
	return res, nil
}

func UsersUpdateFile(ses *utils.Session, data *models.UserRequest) (*models.UserResponse, error) {
	existingUser, err := store.Store().UsersFindById(ses.Context(), *data.ID)
	if err != nil {
		return nil, err
	}
	if data.Avatar != nil {
		existingUser.Avatar = data.Avatar
	}
	if data.DocumentFiles != nil {
		existingUser.DocumentFiles = data.DocumentFiles
	}
	model, err := store.Store().UserUpdate(ses.Context(), existingUser)
	if err != nil {
		return nil, err
	}
	res := models.UserResponse{}
	res.FromModel(model)
	return &res, nil
}

func UsersCreate(ses *utils.Session, data *models.UserRequest) (res *models.UserResponse, isCreated bool, err error) {
	sp, ctx := apm.StartSpan(ses.Context(), "UsersCreate", "app")
	ses.SetContext(ctx)
	defer sp.End()
	setUserDefaults(data)
	model := &models.User{}
	data.ToModel(model)

	// school_id is reqiored when
	if !slices.Contains([]models.Role{
		models.RoleAdmin,
		models.RoleOperator}, model.Schools[0].RoleCode) && model.Schools[0].SchoolUid == nil {
		return nil, isCreated, ErrRequired.SetKey("school_id")
	}

	err = encryptPassword(model, data)
	if err != nil {
		return nil, isCreated, err
	}

	isMerged, err := checkExistingAndMerge(ses, model)
	if err != nil {
		return nil, isCreated, err
	}
	isCreated = !isMerged

	// check if exists
	var m *models.User = model
	if isCreated {
		// else create user
		m, err = store.Store().UserCreate(ses.Context(), model)
		if err != nil {
			return nil, isCreated, err
		}
		m, err = store.Store().UserUpdateRelations(ses.Context(), model)
		if err != nil {
			return nil, isCreated, err
		}
		err = store.Store().UsersLoadRelations(ses.Context(), &[]*models.User{model}, true)
		if err != nil {
			return nil, isCreated, err
		}
	}
	res = &models.UserResponse{}
	res.FromModel(m)
	return res, isCreated, nil
}

func setUserDefaults(model *models.UserRequest) {
	// set required fields random strings
	if model.Password == nil {
		model.Password = new(string)
		*model.Password = RandStringBytes(8)
	}
	if model.Username == nil {
		model.Username = new(string)
		*model.Username = RandStringBytes(8)
	}
	if model.Status == nil {
		model.Status = new(string)
		*model.Status = string(models.StatusActive)
	}
}

func encryptPassword(model *models.User, data *models.UserRequest) error {
	// encrypt password
	if data.Password != nil {
		model.Password = new(string)
		tmp, err := bcrypt.GenerateFromPassword([]byte(*data.Password), 0)
		if err != nil {
			return err
		}
		*model.Password = string(tmp)
	}

	return nil
}

func checkExistingAndMerge(ses *utils.Session, model *models.User) (bool, error) {
	// format first_name
	origFirstName := model.FirstName
	origLastName := model.LastName

	if model.Birthday == nil || model.FirstName == nil || model.LastName == nil {
		return false, nil
		// return false, ErrNotSet.SetKey("first_name").SetComment("required fields to check: first_name, last_name, birthday")
	}
	var classroomID string
	// TODO: ClassroomName nirden almaly men???? SORAMALY
	if model.Classrooms != nil && len(model.Classrooms) > 0 {
		for _, v := range model.Classrooms {
			if v.Classroom != nil && v.Classroom.ID != "" {
				classroomID = v.Classroom.ID
			}
		}
	}
	var userFilter models.UserFilterRequest
	if classroomID != "" {
		userFilter = models.UserFilterRequest{
			LowFirstName: model.FirstName,
			LowLastName:  model.LastName,
			ClassroomId:  &classroomID,
		}
	} else if model.Birthday != nil && !model.Birthday.IsZero() {
		birthdayStr := model.Birthday.Format("2006-01-02")
		userFilter = models.UserFilterRequest{
			LowFirstName: model.FirstName,
			LowLastName:  model.LastName,
			Birthday:     &birthdayStr,
		}
	}
	sameUsers, _, err := store.Store().UsersFindBy(ses.Context(), userFilter)
	if err != nil {
		return false, err
	}
	// found same user
	if len(sameUsers) > 0 {
		u := sameUsers[0]
		err := store.Store().UsersLoadRelations(ses.Context(), &[]*models.User{u}, true)
		if err != nil {
			return false, err
		}
		// add new school-role and filter unique
		u.Schools = append(u.Schools, model.Schools...)
		dataSchools := []*models.UserSchool{}
		for k, v := range u.Schools {
			isContains := false
			for k2, v2 := range dataSchools {
				if k != k2 && v.SchoolUid != nil && v2.SchoolUid != nil && *v.SchoolUid == *v2.SchoolUid && v.RoleCode == v2.RoleCode {
					isContains = true
				}
			}
			if !isContains {
				dataSchools = append(dataSchools, v)
			}
		}
		u.Schools = dataSchools

		// add new children
		if model.Children != nil {
			if u.Children == nil {
				u.Children = []*models.User{}
			}
			u.Children = append(u.Children, model.Children...)
		}

		// add new parents
		if model.Parents != nil {
			if u.Parents == nil {
				u.Parents = []*models.User{}
			}
			u.Parents = append(u.Parents, model.Parents...)
		}

		// add new classrooms
		if model.Classrooms != nil {
			if u.Classrooms == nil {
				u.Classrooms = []*models.UserClassroom{}
			}
			u.Classrooms = append(u.Classrooms, model.Classrooms...)
		}

		// update relations
		u, err = store.Store().UserUpdateRelations(ses.Context(), u)
		if err != nil {
			return false, err
		}
		*model = *u
		return true, nil
	}
	model.FirstName = origFirstName
	model.LastName = origLastName

	return false, nil
}

type DeleteUserRoleQuery struct {
	UserId   string `form:"user" json:"user"`
	SchoolId string `form:"school" json:"school"`
	RoleCode string `form:"role" json:"role"`
}
type DeleteUserRoleQueryList struct {
	UserRoles []DeleteUserRoleQuery `form:"user_roles"`
}

func UsersDelete(ses *utils.Session, deleteRoles []DeleteUserRoleQuery) (int, error) {
	sp, ctx := apm.StartSpan(ses.Context(), "UsersDelete", "app")
	ses.SetContext(ctx)
	defer sp.End()
	var err error

	// userIds := []string{}
	// for _, v := range deleteRoles {
	// 	userIds = append(userIds, v.UserId)
	// }
	// deleteUsers, err := store.Store().UsersFindByIds(ses.Context(), userIds)
	// if err != nil {
	// 	return nil, err
	// }
	// err = store.Store().UsersLoadRelations(ses.Context(), &deleteUsers, false)
	// if err != nil {
	// 	return nil, err
	// }
	// if len(deleteUsers) < 1 {
	// 	return nil, errors.New("user not found: ")
	// }

	// err = store.Store().ClassroomsDeleteStudent(ses.Context(), userIds)
	// if err != nil {
	// 	return nil, err
	// }
	deleted := 0
	for _, v := range deleteRoles {
		var rowsAffected int
		rowsAffected, err = store.Store().UserDeleteSchoolRole(ses.Context(), []string{v.UserId}, []string{v.SchoolId}, []string{v.RoleCode})
		deleted += rowsAffected
	}
	if err != nil {
		return 0, err
	}
	return deleted, nil
}

func (a *App) UserMe(ses *utils.Session) (models.UserResponse, []string, error) {
	sp, ctx := apm.StartSpan(ses.Context(), "UserMe", "app")
	ses.SetContext(ctx)
	defer sp.End()
	// user, err := store.Store().UsersFindById(ses.Context(), ses.GetUser().ID)
	// if err != nil {
	// 	return models.UserResponse{}, nil, err
	// }
	// err = store.Store().UsersLoadRelations(ses.Context(), &[]*models.User{user}, true)
	// if err != nil {
	// 	return models.UserResponse{}, nil, err
	// }
	user := ses.GetUser()
	var err error
	user.Children, err = UserChildrenGet(ses, user)
	if err != nil {
		return models.UserResponse{}, nil, err
	}
	res := models.UserResponse{}
	res.FromModel(user)
	fields := []string{}

	return res, fields, nil
}

func UserMeUpdate(ses *utils.Session, data *models.UserProfileUpdateRequest) (*models.UserResponse, error) {
	sp, ctx := apm.StartSpan(ses.Context(), "UserMeUpdate", "app")
	ses.SetContext(ctx)
	defer sp.End()
	res := &models.UserResponse{}
	user, err := store.Store().UsersFindById(ses.Context(), ses.GetUser().ID)
	if err != nil {
		return nil, err
	}
	if *ses.GetRole() == models.RoleParent || *ses.GetRole() == models.RoleStudent {
		res.FromModel(user)
		return res, nil
	}
	user, err = store.Store().UserUpdate(ses.Context(), user)
	if err != nil {
		return nil, err
	}

	err = store.Store().UsersLoadRelations(ses.Context(), &[]*models.User{user}, true)
	if err != nil {
		return nil, err
	}

	sl := []*models.School{}
	for _, s := range user.Schools {
		if s.School != nil {
			sl = append(sl, s.School)
		}
	}
	err = store.Store().SchoolsLoadRelations(ses.Context(), &sl)
	if err != nil {
		return nil, err
	}

	res.FromModel(user)
	return res, err
}

func UserPasswordUpdate(ses *utils.Session, data models.UserPasswordUpdateRequest) (*models.UserResponse, error) {
	sp, ctx := apm.StartSpan(ses.Context(), "UserPasswordUpdate", "app")
	ses.SetContext(ctx)
	defer sp.End()
	res := &models.UserResponse{}
	user := ses.GetUser()
	var err error
	err = bcrypt.CompareHashAndPassword([]byte(*user.Password), []byte(*data.OldPassword))
	if err != nil {
		return nil, ErrInvalid.SetKey("old_password")
	}

	if data.NewPassword != nil {
		user.Password = new(string)
		tmp, err := bcrypt.GenerateFromPassword([]byte(*data.NewPassword), 0)
		if err != nil {
			return nil, err
		}
		*user.Password = string(tmp)
	}

	user, err = store.Store().UserUpdate(ses.Context(), user)
	if err != nil {
		return nil, err
	}

	err = store.Store().UsersLoadRelations(ses.Context(), &[]*models.User{user}, true)
	if err != nil {
		return nil, err
	}

	res.FromModel(user)
	return res, err
}

func (a App) UserSessions(ses *utils.Session) ([]models.SessionResponse, error) {
	sp, ctx := apm.StartSpan(ses.Context(), "UserSessions", "app")
	ses.SetContext(ctx)
	defer sp.End()
	res := []models.SessionResponse{}
	user := ses.GetUser()
	l, err := utils.SessionByUserId(user.ID)
	for _, v := range l {
		r := models.SessionResponse{}
		r.FromModel(&v)
		res = append(res, r)
	}
	return res, err
}

func (a App) UserSessionsDelete(ses *utils.Session, sesIds []string) error {
	sp, ctx := apm.StartSpan(ses.Context(), "UserSessionsDelete", "app")
	ses.SetContext(ctx)
	defer sp.End()
	user := ses.GetUser()
	l, err := utils.SessionByUserId(user.ID)
	if err != nil {
		return err
	}
	for _, sesId := range sesIds {
		for _, v := range l {
			if sesId == v.ID {
				utils.SessionDelete(v)
			}
		}
	}
	if len(sesIds) < 1 {
		s, _ := utils.SessionByToken(ses.GetToken())
		utils.SessionDelete(s)
	}
	return err
}
