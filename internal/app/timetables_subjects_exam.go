package app

import (
	"errors"
	"strings"

	"github.com/mekdep/server/internal/api/utils"
	"github.com/mekdep/server/internal/models"
	"github.com/mekdep/server/internal/store"
	"go.elastic.co/apm/v2"
)

func SubjectExamList(ses *utils.Session, data *models.SubjectExamFilterRequest) ([]*models.SubjectExamResponse, int, error) {
	sp, ctx := apm.StartSpan(ses.Context(), "SubjectExamList", "app")
	ses.SetContext(ctx)
	defer sp.End()
	if data.ID == nil {
		data.ID = new(string)
	}
	l, total, err := store.Store().SubjectExamsFindBy(ses.Context(), data)
	if err != nil {
		return nil, 0, err
	}
	err = store.Store().SubjectExamLoadRelations(ses.Context(), &l)
	if err != nil {
		return nil, 0, err
	}
	res := []*models.SubjectExamResponse{}
	for _, m := range l {
		item := models.SubjectExamResponse{}
		item.FromModel(m)
		res = append(res, &item)
	}
	return res, total, nil
}

func SubjectExamCreate(ses *utils.Session, data *models.SubjectExamRequest) (*models.SubjectExamResponse, error) {
	sp, ctx := apm.StartSpan(ses.Context(), "SubjectExamCreate", "app")
	ses.SetContext(ctx)
	defer sp.End()

	// validate member_teacher_ids
	if data.MemberTeacherIds != nil {
		roles := []string{string(models.RoleTeacher), string(models.RolePrincipal)}
		filter := models.UserFilterRequest{
			Ids:      data.MemberTeacherIds,
			SchoolId: data.SchoolId,
			Roles:    &roles,
		}
		users, _, err := store.Store().UsersFindBy(ses.Context(), filter)
		if err != nil {
			return nil, err
		}
		if len(users) != len(*data.MemberTeacherIds) {
			return nil, errors.New("one or more member_teacher_ids do not exist")
		}
	}
	// validate subject_id
	subject, err := store.Store().SubjectsFindById(ses.Context(), *data.SubjectId)
	if err != nil {
		return nil, err
	}
	if subject == nil {
		return nil, ErrInvalid.SetKey("subject_id")
	}
	// validate exists
	_, examsCount, err := store.Store().SubjectExamsFindBy(ses.Context(), &models.SubjectExamFilterRequest{
		SubjectId: data.SubjectId,
	})
	if err != nil {
		return nil, err
	}
	if examsCount > 0 {
		return nil, ErrUnique.SetKey("subject_id").SetComment("Already exists")
	}

	// set default vars
	data.ClassroomId = &subject.ClassroomId
	data.SchoolId = &subject.SchoolId
	// create model
	m := &models.SubjectExam{}
	data.ToModel(m)
	m, err = store.Store().SubjectExamCreate(ses.Context(), m)
	if err != nil {
		return nil, err
	}
	res := &models.SubjectExamResponse{}
	err = store.Store().SubjectExamLoadRelations(ses.Context(), &[]*models.SubjectExam{m})
	if err != nil {
		return nil, err
	}
	res.FromModel(m)
	return res, nil
}

func SubjectExamUpdate(ses *utils.Session, data *models.SubjectExamRequest) (*models.SubjectExamResponse, error) {
	sp, ctx := apm.StartSpan(ses.Context(), "SubjetExamUpdate", "app")
	ses.SetContext(ctx)
	defer sp.End()
	if data.MemberTeacherIds != nil {
		roleTeacher := string(models.RoleTeacher)
		filter := models.UserFilterRequest{
			Ids:      data.MemberTeacherIds,
			SchoolId: data.SchoolId,
			Role:     &roleTeacher,
		}
		users, _, err := store.Store().UsersFindBy(ses.Context(), filter)
		if err != nil {
			return nil, err
		}
		if len(users) != len(*data.MemberTeacherIds) {
			return nil, errors.New("one or more member_teacher_ids do not exist")
		}
	}
	m := &models.SubjectExam{}
	var err error
	if m, err = store.Store().SubjectExamFindById(ses.Context(), *data.ID); err != nil {
		return nil, err
	}
	data.ToModel(m)
	m, err = store.Store().SubjectExamUpdate(ses.Context(), m)
	if err != nil {
		return nil, err
	}
	err = store.Store().SubjectExamLoadRelations(ses.Context(), &[]*models.SubjectExam{m})
	if err != nil {
		return nil, err
	}
	res := &models.SubjectExamResponse{}
	res.FromModel(m)
	return res, nil
}

func SubjectExamDelete(ses *utils.Session, ids []string) ([]*models.SubjectExam, error) {
	sp, ctx := apm.StartSpan(ses.Context(), "SubjectExamDelete", "app")
	ses.SetContext(ctx)
	defer sp.End()
	l, err := store.Store().SubjectExamFindByIds(ses.Context(), ids)
	if err != nil {
		return nil, err
	}
	if len(l) < 1 {
		return nil, errors.New("model not found: " + strings.Join(ids, ","))
	}
	return store.Store().SubjectExamDelete(ses.Context(), l)
}
