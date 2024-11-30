package app

import (
	"fmt"
	"slices"

	"github.com/mekdep/server/internal/api/utils"
	"github.com/mekdep/server/internal/models"
	"github.com/mekdep/server/internal/store"
	"go.elastic.co/apm/v2"
)

// TODO: VALIDATE REQUEST DTO AND DON'T TRUST THE FRONTEND
func GetMessageGroups(ses *utils.Session, dto models.GetMessageGroupsRequest) ([]*models.MessageGroup, int, error) {
	sp, ctx := apm.StartSpan(ses.Context(), "GetMessageGroups", "app")
	ses.SetContext(ctx)
	defer sp.End()

	dto.SchoolId = &ses.GetSchool().ID
	messageGroups := []*models.MessageGroup{}
	total := 0

	messageGroupsFromClassroom, totalFromClassroom, err := getOrInitMessageGroupByClassroom(ses, dto)
	if err != nil {
		return nil, 0, err
	}
	messageGroups = append(messageGroups, messageGroupsFromClassroom...)
	total += totalFromClassroom

	messageGroupsFromRole, totalFromRole, err := getOrInitMessageGroupByRole(ses, dto)
	if err != nil {
		return nil, 0, err
	}
	messageGroups = append(messageGroups, messageGroupsFromRole...)
	total += totalFromRole

	return messageGroups, total, err
}

func getOrInitMessageGroupByClassroom(ses *utils.Session, dto models.GetMessageGroupsRequest) ([]*models.MessageGroup, int, error) {
	allowedRoles := []models.Role{models.RoleParent, models.RoleTeacher}
	role := *(ses.GetRole())
	userId := ses.GetUser().ID
	messageGroupType := string(models.MessageGroupParentsType)
	dto.Type = &messageGroupType

	if dto.ClassroomId == nil {
		return []*models.MessageGroup{}, 0, nil
	}
	if !slices.Contains(allowedRoles, role) {
		return []*models.MessageGroup{}, 0, nil
	}

	if role == models.RoleParent {
		childrenDTO := models.UserFilterRequest{
			SchoolId:    dto.SchoolId,
			ClassroomId: dto.ClassroomId,
			ParentId:    &userId,
		}
		_, childrenTotal, err := store.Store().UsersFindBy(
			ses.Context(),
			childrenDTO,
		)
		if err != nil {
			return nil, 0, err
		}
		if childrenTotal == 0 {
			return nil, 0, ErrForbidden
		}
	}

	teacherClassroom := ses.GetUser().TeacherClassroom
	if teacherClassroom != nil && teacherClassroom.ID != *dto.ClassroomId {
		return nil, 0, ErrForbidden
	}

	dto.CurrentUserId = &userId
	messageGroups, total, err := store.Store().MessageGroupsFindBy(ses.Context(), dto)
	if err != nil {
		return nil, 0, err
	}

	if total == 0 {
		groupTitle, err := getGroupTitle(ses, models.MessageGroupParentsType, dto.ClassroomId)
		if err != nil {
			return nil, 0, err
		}

		model := models.MessageGroup{
			AdminId:     userId,
			SchoolId:    *dto.SchoolId,
			ClassroomId: dto.ClassroomId,
			Title:       groupTitle,
			Type:        *dto.Type,
		}
		model.SetDefaults()

		message_group, err := store.Store().CreateMessageGroupCommand(ses.Context(), model)
		if err != nil {
			return nil, 0, err
		}

		messageGroups = []*models.MessageGroup{&message_group}
	}

	return messageGroups, total, err
}

func getOrInitMessageGroupByRole(ses *utils.Session, dto models.GetMessageGroupsRequest) ([]*models.MessageGroup, int, error) {
	allowedRoles := []models.Role{models.RoleTeacher, models.RoleAdmin}
	userId := ses.GetUser().ID

	if !slices.Contains(allowedRoles, *(ses.GetRole())) {
		return []*models.MessageGroup{}, 0, nil
	}

	dto.Type = (*string)(ses.GetRole())
	dto.ClassroomId = nil

	dto.CurrentUserId = &userId
	messageGroups, total, err := store.Store().MessageGroupsFindBy(ses.Context(), dto)
	if err != nil {
		return nil, 0, err
	}

	if total == 0 {
		groupTitle, err := getGroupTitle(ses, models.MessageGroupType(*dto.Type), nil)
		if err != nil {
			return nil, 0, err
		}

		model := models.MessageGroup{
			AdminId:  ses.GetUser().ID,
			SchoolId: *dto.SchoolId,
			Title:    groupTitle,
			Type:     *dto.Type,
		}
		model.SetDefaults()

		message_group, err := store.Store().CreateMessageGroupCommand(ses.Context(), model)
		if err != nil {
			return nil, 0, err
		}

		messageGroups = []*models.MessageGroup{&message_group}
	}

	return messageGroups, total, err
}

func getGroupTitle(ses *utils.Session, groupType models.MessageGroupType, classroomId *string) (string, error) {
	var groupTitle string
	switch string(groupType) {
	case string(models.RoleParent):
		{
			classroom, err := store.Store().ClassroomsFindById(ses.Context(), *classroomId)
			if err != nil {
				return "", err
			}
			groupTitle = *classroom.Name + " Ata-eneler"
		}
	case string(models.RoleTeacher):
		{
			groupTitle = "Mugallymlar"
		}
	case string(models.RoleAdmin):
		{
			groupTitle = "Adminler"
		}
	}
	return groupTitle, nil
}

func GetMessageGroup(ses *utils.Session, id string) (*models.MessageGroup, error) {
	sp, ctx := apm.StartSpan(ses.Context(), "GetMessageGroup", "app")
	ses.SetContext(ctx)
	defer sp.End()

	messageGroup, err := store.Store().MessageGroupsFindById(ses.Context(), fmt.Sprint(id))
	if err != nil {
		return nil, err
	}

	if *ses.GetSchoolId() != messageGroup.SchoolId {
		return nil, ErrForbidden
	}

	return &messageGroup, nil
}

// TODO: VALIDATE REQUEST DTO AND DON'T TRUST THE FRONTEND
func GetMessageGroupMembers(ses *utils.Session, group models.MessageGroup, classroomId *string) ([]*models.User, int, error) {
	sp, ctx := apm.StartSpan(ses.Context(), "GetMessageGroup", "app")
	ses.SetContext(ctx)
	defer sp.End()

	var members []*models.User
	var total int
	var err error

	switch group.Type {
	case string(models.MessageGroupParentsType):
		members, total, err = getParentsGroupMembers(ses, classroomId)
	case string(models.MessageGroupTeachersType):
		members, total, err = getTeachersGroupMembers(ses)
	default:
		return nil, 0, ErrForbidden
	}

	return members, total, err
}

func getParentsGroupMembers(ses *utils.Session, classroomId *string) ([]*models.User, int, error) {
	allowedRoles := []models.Role{models.RoleParent, models.RoleTeacher}
	role := ses.GetRole()
	schooldId := ses.GetSchoolId()

	var members []*models.User
	var total int
	membersLimit := 500

	if !slices.Contains(allowedRoles, *role) {
		return nil, 0, ErrForbidden
	}
	if classroomId == nil || *classroomId == "" {
		return nil, 0, ErrForbidden
	}

	// get class teacher
	teachersDTO := models.UserFilterRequest{
		SchoolId:    schooldId,
		Roles:       &[]string{string(models.RoleTeacher)},
		ClassroomId: classroomId,
		PaginationRequest: models.PaginationRequest{
			Limit: &membersLimit,
		},
	}
	teachers, teachersTotal, err := store.Store().UsersFindBy(ses.Context(), teachersDTO)
	if err != nil {
		return nil, 0, err
	}
	members = append(members, teachers...)
	total += teachersTotal

	parentsDTO := models.UserFilterRequest{
		SchoolId:    schooldId,
		Roles:       &[]string{string(models.RoleParent)},
		ClassroomId: classroomId,
		PaginationRequest: models.PaginationRequest{
			Limit: &membersLimit,
		},
	}
	parents, parentsTotal, err := store.Store().UsersFindBy(ses.Context(), parentsDTO)
	if err != nil {
		return nil, 0, err
	}
	members = append(members, parents...)
	total += parentsTotal

	return members, total, err
}

func getTeachersGroupMembers(ses *utils.Session) ([]*models.User, int, error) {
	allowedRoles := []models.Role{models.RoleTeacher}
	role := ses.GetRole()
	schooldId := ses.GetSchoolId()
	membersLimit := 500

	if !slices.Contains(allowedRoles, *role) {
		return nil, 0, ErrForbidden
	}

	teachersDTO := models.UserFilterRequest{
		SchoolId: schooldId,
		Roles:    &[]string{string(models.RoleTeacher)},
		PaginationRequest: models.PaginationRequest{
			Limit: &membersLimit,
		},
	}
	teachers, total, err := store.Store().UsersFindBy(ses.Context(), teachersDTO)
	if err != nil {
		return nil, 0, err
	}

	return teachers, total, err
}
