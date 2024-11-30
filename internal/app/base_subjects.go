package app

import (
	"errors"
	"strings"

	"github.com/mekdep/server/internal/api/utils"
	"github.com/mekdep/server/internal/models"
	"github.com/mekdep/server/internal/store"
	"go.elastic.co/apm/v2"
)

func BaseSubjectsList(ses *utils.Session, f models.BaseSubjectsFilterRequest) ([]*models.BaseSubjectsResponse, int, error) {
	sp, ctx := apm.StartSpan(ses.Context(), "BaseSubjectsList", "app")
	ses.SetContext(ctx)
	defer sp.End()
	baseSubjects, total, err := store.Store().BaseSubjectsFindBy(ses.Context(), f)
	if err != nil {
		return nil, 0, err
	}
	err = store.Store().BaseSubjectsLoadRelations(ses.Context(), &baseSubjects)
	if err != nil {
		return nil, 0, err
	}
	baseSubjectsResponse := []*models.BaseSubjectsResponse{}
	for _, v := range baseSubjects {
		s := models.BaseSubjectsResponse{}
		s.FromModel(v)
		baseSubjectsResponse = append(baseSubjectsResponse, &s)
	}
	return baseSubjectsResponse, total, err
}

func BaseSubjectsDetail(ses *utils.Session, id string) (*models.BaseSubjectsResponse, error) {
	sp, ctx := apm.StartSpan(ses.Context(), "BaseSubjectsDetail", "app")
	ses.SetContext(ctx)
	defer sp.End()
	m, err := store.Store().BaseSubjectsFindById(ses.Context(), id)
	if err != nil {
		return nil, err
	}
	err = store.Store().BaseSubjectsLoadRelations(ses.Context(), &[]*models.BaseSubjects{m})
	if err != nil {
		return nil, err
	}
	res := &models.BaseSubjectsResponse{}
	res.FromModel(m)
	return res, nil
}

func BaseSubjectsUpdate(ses *utils.Session, data models.BaseSubjectsRequest) (*models.BaseSubjectsResponse, error) {
	sp, ctx := apm.StartSpan(ses.Context(), "BaseSubjectsUpdate", "app")
	ses.SetContext(ctx)
	defer sp.End()
	model := &models.BaseSubjects{}
	data.ToModel(model)

	var err error
	model, err = store.Store().BaseSubjectsUpdate(ses.Context(), model)
	if err != nil {
		return nil, err
	}
	err = store.Store().BaseSubjectsLoadRelations(ses.Context(), &[]*models.BaseSubjects{model})
	if err != nil {
		return nil, err
	}
	res := &models.BaseSubjectsResponse{}
	res.FromModel(model)
	return res, nil
}

func BaseSubjectsCreate(ses *utils.Session, data models.BaseSubjectsRequest) (*models.BaseSubjectsResponse, error) {
	sp, ctx := apm.StartSpan(ses.Context(), "BaseSubjectsCreate", "app")
	ses.SetContext(ctx)
	defer sp.End()
	model := &models.BaseSubjects{}
	if data.Price != nil && *data.Price < 0 {
		return nil, ErrInvalid.SetKey("price")
	}
	if data.ExamMinGrade != nil && *data.ExamMinGrade < 0 {
		return nil, ErrInvalid.SetKey("exam_min_grade")
	}
	data.ToModel(model)
	res := &models.BaseSubjectsResponse{}
	var err error
	model, err = store.Store().BaseSubjectsCreate(ses.Context(), model)
	if err != nil {
		return nil, err
	}
	err = store.Store().BaseSubjectsLoadRelations(ses.Context(), &[]*models.BaseSubjects{model})
	if err != nil {
		return nil, err
	}
	res.FromModel(model)
	return res, nil
}

func BaseSubjectsDelete(ses *utils.Session, ids []string) ([]*models.BaseSubjects, error) {
	sp, ctx := apm.StartSpan(ses.Context(), "BaseSubjectsDelete", "app")
	ses.SetContext(ctx)
	defer sp.End()
	books, err := store.Store().BaseSubjectsFindByIds(ses.Context(), ids)
	if err != nil {
		return nil, err
	}
	if len(books) < 1 {
		return nil, errors.New("model not found: " + strings.Join(ids, ","))
	}
	return store.Store().BaseSubjectsDelete(ses.Context(), books)
}
