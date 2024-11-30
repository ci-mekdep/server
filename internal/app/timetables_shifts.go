package app

import (
	"errors"
	"strings"

	"github.com/mekdep/server/internal/api/utils"
	"github.com/mekdep/server/internal/models"
	"github.com/mekdep/server/internal/store"
	"go.elastic.co/apm/v2"
)

func ShiftList(ses *utils.Session, f models.ShiftFilterRequest) ([]*models.ShiftResponse, int, error) {
	sp, ctx := apm.StartSpan(ses.Context(), "ShiftList", "app")
	ses.SetContext(ctx)
	defer sp.End()
	shifts, total, err := store.Store().ShiftsFindBy(ses.Context(), f)
	if err != nil {
		return nil, 0, err
	}
	err = store.Store().ShiftLoadRelations(ses.Context(), &shifts)
	if err != nil {
		return nil, 0, err
	}
	shiftsResponse := []*models.ShiftResponse{}
	for _, shift := range shifts {
		s := models.ShiftResponse{}
		s.FromModel(shift)
		shiftsResponse = append(shiftsResponse, &s)
	}
	return shiftsResponse, total, err
}

func ShiftDetail(ses *utils.Session, id string) (*models.ShiftResponse, error) {
	sp, ctx := apm.StartSpan(ses.Context(), "ShiftDetail", "app")
	ses.SetContext(ctx)
	defer sp.End()
	m, err := store.Store().ShiftsFindById(ses.Context(), id)
	if err != nil {
		return nil, err
	}
	err = store.Store().ShiftLoadRelations(ses.Context(), &[]*models.Shift{m})
	if err != nil {
		return nil, err
	}
	res := &models.ShiftResponse{}
	res.FromModel(m)
	return res, nil
}

func UpdateShift(ses *utils.Session, data models.ShiftRequest) (*models.ShiftResponse, error) {
	sp, ctx := apm.StartSpan(ses.Context(), "UpdateShift", "app")
	ses.SetContext(ctx)
	defer sp.End()
	model := &models.Shift{}
	data.ToModel(model)

	var err error
	model, err = store.Store().UpdateShift(ses.Context(), model)
	if err != nil {
		return nil, err
	}
	err = store.Store().ShiftLoadRelations(ses.Context(), &[]*models.Shift{model})
	if err != nil {
		return nil, err
	}

	res := &models.ShiftResponse{}
	res.FromModel(model)
	return res, nil
}

func CreateShift(ses *utils.Session, data models.ShiftRequest) (*models.ShiftResponse, error) {
	sp, ctx := apm.StartSpan(ses.Context(), "CreateShift", "app")
	ses.SetContext(ctx)
	defer sp.End()
	model := &models.Shift{}
	data.ToModel(model)
	res := &models.ShiftResponse{}
	var err error
	model, err = store.Store().CreateShift(ses.Context(), model)
	if err != nil {
		return nil, err
	}
	err = store.Store().ShiftLoadRelations(ses.Context(), &[]*models.Shift{model})
	if err != nil {
		return nil, err
	}
	res.FromModel(model)
	return res, nil
}

func ShiftsDelete(ses *utils.Session, ids []string) ([]*models.Shift, error) {
	sp, ctx := apm.StartSpan(ses.Context(), "ShiftsDelete", "app")
	ses.SetContext(ctx)
	defer sp.End()
	shifts, err := store.Store().ShiftsFindByIds(ses.Context(), ids)
	if err != nil {
		return nil, err
	}
	if len(shifts) < 1 {
		return nil, errors.New("model not found: " + strings.Join(ids, ","))
	}
	return store.Store().DeleteShifts(ses.Context(), shifts)
}
