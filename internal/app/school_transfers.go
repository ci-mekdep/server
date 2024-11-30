package app

import (
	"context"
	"errors"
	"strings"

	"github.com/mekdep/server/internal/api/utils"
	"github.com/mekdep/server/internal/models"
	"github.com/mekdep/server/internal/store"
	"go.elastic.co/apm/v2"
)

func SchoolTransfersList(ses *utils.Session, data *models.SchoolTransferQueryDto) (resp *models.SchoolTransfersResponse, err error) {
	sp, ctx := apm.StartSpan(ses.Context(), "SchoolTransfersList", "app")
	ses.SetContext(ctx)
	defer sp.End()
	resp = &models.SchoolTransfersResponse{}

	l, err := store.Store().SchoolTransfersFindBy(context.Background(), models.ConvertSchoolTransferQueryToMap(*data))
	if err != nil {
		return nil, err
	}
	err = store.Store().SchoolTransfersLoadRelations(context.Background(), l)
	if err != nil {
		return nil, err
	}
	for _, v := range l.SchoolTransfers {
		res := &models.SchoolTransferResponse{}
		res.FromModel(v)
		resp.SchoolTransfersResponse = append(resp.SchoolTransfersResponse, *res)
	}
	resp.Total = l.Total
	return resp, nil
}

func SchoolTransferDetail(ses *utils.Session, id string) (response *models.SchoolTransferResponse, err error) {
	sp, ctx := apm.StartSpan(ses.Context(), "SchoolTransferDetail", "app")
	ses.SetContext(ctx)
	defer sp.End()
	schoolTransfer, err := store.Store().SchoolTransfersFindById(ses.Context(), id)
	if err != nil {
		return nil, err
	}
	err = store.Store().SchoolTransfersLoadRelations(context.Background(), &models.SchoolTransfers{SchoolTransfers: []*models.SchoolTransfer{schoolTransfer}})
	if err != nil {
		return nil, err
	}
	res := models.SchoolTransferResponse{}
	res.FromModel(schoolTransfer)
	return &res, nil
}

func SchoolTransfersCreate(ses *utils.Session, data *models.SchoolTransferCreateDto) (*models.SchoolTransferResponse, error) {
	sp, ctx := apm.StartSpan(ses.Context(), "SchoolTransfersCreate", "app")
	ses.SetContext(ctx)
	defer sp.End()
	m := &models.SchoolTransfer{}
	data.ToModel(m)
	var err error
	m, err = store.Store().SchoolTransfersInsert(context.Background(), m)
	if err != nil {
		return nil, err
	}
	schoolTransfers := &models.SchoolTransfers{
		SchoolTransfers: []*models.SchoolTransfer{m},
	}
	res := &models.SchoolTransferResponse{}
	err = store.Store().SchoolTransfersLoadRelations(context.Background(), schoolTransfers)
	if err != nil {
		return nil, err
	}
	res.FromModel(m)
	return res, nil
}

func SchoolTransfersUpdate(ses *utils.Session, data *models.SchoolTransferCreateDto) (*models.SchoolTransferResponse, error) {
	sp, ctx := apm.StartSpan(ses.Context(), "SchoolTransfersUpdate", "app")
	ses.SetContext(ctx)
	defer sp.End()

	m := &models.SchoolTransfer{}
	var err error
	if m, err = store.Store().SchoolTransfersFindById(context.Background(), *data.ID); err != nil {
		return nil, err
	}
	if m.Status != data.Status && data.Status != nil && *data.Status == string(models.StatusAccepted) {
		data.ToModel(m)
		if m, err = store.Store().SchoolTransfersUpdate(context.Background(), m); err != nil {
			return nil, err
		}
		schoolTranfers := &models.SchoolTransfers{
			SchoolTransfers: []*models.SchoolTransfer{m},
		}
		if err = store.Store().SchoolTransfersLoadRelations(context.Background(), schoolTranfers); err != nil {
			return nil, err
		}
		if m.Status != nil && *m.Status == string(models.StatusAccepted) {
			err = syncSchoolTransfers(ses, m)
			if err != nil {
				return nil, err
			}
		}
		res := &models.SchoolTransferResponse{}
		res.FromModel(m)
		return res, nil
	}
	return nil, nil
}

func SchoolTransfersDelete(ses *utils.Session, ids []string) (*models.SchoolTransfers, error) {
	sp, ctx := apm.StartSpan(ses.Context(), "SchoolTransfersDelete", "app")
	ses.SetContext(ctx)
	defer sp.End()

	req := &models.SchoolTransferQueryDto{
		IDs: ids,
	}
	req.Status = new(string)
	*req.Status = string(models.StatusWaiting)

	l, err := store.Store().SchoolTransfersFindBy(context.Background(), models.ConvertSchoolTransferQueryToMap(*req))
	if err != nil {
		return nil, err
	}
	if len(l.SchoolTransfers) < 1 {
		return nil, errors.New("model not found: " + strings.Join(ids, ","))
	}
	return store.Store().SchoolTransfersDelete(context.Background(), ids)
}

func syncSchoolTransfers(ses *utils.Session, m *models.SchoolTransfer) error {
	_, err := store.Store().UserDeleteSchoolRole(ses.Context(), []string{*m.StudentId}, []string{*m.SourceSchoolId}, []string{string(models.RoleStudent)})
	if err != nil {
		return err
	}
	err = store.Store().UserDeleteFromClassroom(ses.Context(), *m.StudentId, []string{*m.SourceClassroomId})
	if err != nil {
		return err
	}
	err = store.Store().UserChangeSchoolAndClassroom(ses.Context(), m.StudentId, m.TargetSchoolId, m.TargetClassroomId)
	if err != nil {
		return err
	}
	periodGrades, err := store.Store().PeriodGradeByStudent(ses.Context(), *m.StudentId)
	if err != nil {
		return err
	}
	oldSubjects, err := store.Store().SubjectsFindByClassroomId(ses.Context(), *m.SourceClassroomId)
	if err != nil {
		return err
	}
	newSubjects, err := store.Store().SubjectsFindByClassroomId(ses.Context(), *m.TargetClassroomId)
	if err != nil {
		return err
	}
	err = store.Store().MapOldSubjectsToNewSubjectsInPeriodGrade(ses.Context(), *m.StudentId, periodGrades, oldSubjects, newSubjects)
	if err != nil {
		return err
	}
	return nil
}
