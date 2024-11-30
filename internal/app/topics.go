package app

import (
	"errors"
	"strings"

	"github.com/mekdep/server/internal/api/utils"
	"github.com/mekdep/server/internal/models"
	"github.com/mekdep/server/internal/store"
	"go.elastic.co/apm/v2"
)

func TopicsList(ses *utils.Session, f models.TopicsFilterRequest) ([]*models.TopicsResponse, int, error) {
	sp, ctx := apm.StartSpan(ses.Context(), "TopicsList", "app")
	ses.SetContext(ctx)
	defer sp.End()
	topics, total, err := store.Store().TopicsFindBy(ses.Context(), f)
	if err != nil {
		return nil, 0, err
	}
	err = store.Store().TopicsLoadRelations(ses.Context(), &topics)
	if err != nil {
		return nil, 0, err
	}
	topicsResponse := []*models.TopicsResponse{}
	for _, shift := range topics {
		s := models.TopicsResponse{}
		s.FromModel(shift)
		topicsResponse = append(topicsResponse, &s)
	}
	return topicsResponse, total, err
}

func TopicsDetail(ses *utils.Session, id string) (*models.TopicsResponse, error) {
	sp, ctx := apm.StartSpan(ses.Context(), "TopicsDetail", "app")
	ses.SetContext(ctx)
	defer sp.End()
	m, err := store.Store().TopicsFindById(ses.Context(), id)
	if err != nil {
		return nil, err
	}
	err = store.Store().TopicsLoadRelations(ses.Context(), &[]*models.Topics{m})
	if err != nil {
		return nil, err
	}
	res := &models.TopicsResponse{}
	res.FromModel(m)
	return res, nil
}

func TopicsListValues(ses *utils.Session, data models.TopicsFilterRequest) ([]models.TopicsValueResponse, int, error) {
	sp, ctx := apm.StartSpan(ses.Context(), "TopicsListValues", "app")
	ses.SetContext(ctx)
	defer sp.End()
	res, total, err := TopicsList(ses, data)
	if err != nil {
		return nil, 0, err
	}
	values := []models.TopicsValueResponse{}
	for _, v := range res {
		values = append(values, v.ToValues())
	}
	return values, total, nil
}

func TopicsUpdate(ses *utils.Session, data models.TopicsRequest) (*models.TopicsResponse, error) {
	sp, ctx := apm.StartSpan(ses.Context(), "TopicsUpdate", "app")
	ses.SetContext(ctx)
	defer sp.End()
	model := &models.Topics{}
	data.ToModel(model)

	var err error
	model, err = store.Store().TopicsUpdate(ses.Context(), model)
	if err != nil {
		return nil, err
	}
	err = store.Store().TopicsLoadRelations(ses.Context(), &[]*models.Topics{model})
	if err != nil {
		return nil, err
	}
	res := &models.TopicsResponse{}
	res.FromModel(model)
	return res, nil
}

func TopicsCreate(ses *utils.Session, data models.TopicsRequest) (*models.TopicsResponse, error) {
	sp, ctx := apm.StartSpan(ses.Context(), "TopicsCreate", "app")
	ses.SetContext(ctx)
	defer sp.End()
	model := &models.Topics{}
	data.ToModel(model)
	res := &models.TopicsResponse{}
	var err error
	model, err = store.Store().TopicsCreate(ses.Context(), model)
	if err != nil {
		return nil, err
	}
	err = store.Store().TopicsLoadRelations(ses.Context(), &[]*models.Topics{model})
	if err != nil {
		return nil, err
	}
	res.FromModel(model)
	return res, nil
}

func TopicsMultipleCreate(ses *utils.Session, data models.TopicsMultipleRequest) ([]*models.TopicsResponse, error) {
	res := make([]*models.TopicsResponse, len(data.Topics))
	for _, item := range data.Topics {
		model := &models.Topics{}
		item.ToModel(model)
		var err error
		model, err = store.Store().TopicsCreate(ses.Context(), model)
		if err != nil {
			return nil, err
		}
	}
	return res, nil
}

func TopicsDelete(ses *utils.Session, ids []string) ([]*models.Topics, error) {
	sp, ctx := apm.StartSpan(ses.Context(), "TopicsDelete", "app")
	ses.SetContext(ctx)
	defer sp.End()
	topics, err := store.Store().TopicsFindByIds(ses.Context(), ids)
	if err != nil {
		return nil, err
	}
	if len(topics) < 1 {
		return nil, errors.New("model not found: " + strings.Join(ids, ","))
	}
	return store.Store().TopicsDelete(ses.Context(), topics)
}
