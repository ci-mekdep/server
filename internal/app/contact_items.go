package app

import (
	"errors"
	"strings"

	"github.com/mekdep/server/internal/api/utils"
	"github.com/mekdep/server/internal/models"
	"github.com/mekdep/server/internal/store"
	"go.elastic.co/apm/v2"
)

func ContactItemsList(ses *utils.Session, f models.ContactItemsFilterRequest) ([]*models.ContactItemsResponse, int, error) {
	sp, ctx := apm.StartSpan(ses.Context(), "ContactItemsList", "app")
	ses.SetContext(ctx)
	defer sp.End()
	contactItems, total, err := store.Store().ContactItemsFindBy(ses.Context(), f)
	if err != nil {
		return nil, 0, err
	}

	err = store.Store().ContactItemLoadRelations(ses.Context(), &contactItems, true)
	if err != nil {
		return nil, 0, err
	}
	if f.Status != nil {
		f.NotStatus = new(string)
		*f.NotStatus = string(models.ContactStatusRejected)
	}

	contactItemsResponse := []*models.ContactItemsResponse{}
	for _, contact := range contactItems {
		s := models.ContactItemsResponse{}
		s.FromModel(contact)
		contactItemsResponse = append(contactItemsResponse, &s)
	}
	return contactItemsResponse, total, err
}

func ContactItemsDetail(ses *utils.Session, id string) (*models.ContactItemsResponse, error) {
	sp, ctx := apm.StartSpan(ses.Context(), "ContactItemsDetail", "app")
	ses.SetContext(ctx)
	defer sp.End()
	m, err := store.Store().ContactItemsFindById(ses.Context(), id)
	if err != nil {
		return nil, err
	}
	err = store.Store().ContactItemLoadRelations(ses.Context(), &[]*models.ContactItems{m}, true)
	if err != nil {
		return nil, err
	}
	res := &models.ContactItemsResponse{}
	res.FromModel(m)
	return res, nil
}

func ContactItemsUpdate(ses *utils.Session, data models.ContactItemsRequest) (*models.ContactItemsResponse, error) {
	sp, ctx := apm.StartSpan(ses.Context(), "ContactItemsUpdate", "app")
	ses.SetContext(ctx)
	defer sp.End()
	var err error
	if data.RelatedChildrenIds != nil && len(*data.RelatedChildrenIds) > 0 {
		relatedChildren, err := store.Store().ContactItemsFindByIds(ses.Context(), *data.RelatedChildrenIds)
		if err != nil {
			return nil, err
		}
		if relatedChildren != nil && len(relatedChildren) > 0 {
			for _, child := range relatedChildren {
				child.RelatedId = data.ID
				_, err = store.Store().ContactItemUpdate(ses.Context(), child)
				if err != nil {
					return nil, err
				}
			}
		}
		return nil, nil
	}
	model, err := store.Store().ContactItemsFindById(ses.Context(), *data.ID)
	if data.Status != "" {
		model.Status = data.Status
	}
	if data.Note != nil {
		model.Note = data.Note
	}
	if data.RelatedId != nil {
		model.RelatedId = data.RelatedId
	}
	model.UpdatedBy = &ses.GetUser().ID

	model, err = store.Store().ContactItemUpdate(ses.Context(), model)
	if err != nil {
		return nil, err
	}
	err = store.Store().ContactItemLoadRelations(ses.Context(), &[]*models.ContactItems{model}, true)
	if err != nil {
		return nil, err
	}

	res := &models.ContactItemsResponse{}
	res.FromModel(model)
	return res, nil
}

func ContactItemsCreate(ses *utils.Session, data models.ContactItemsRequest) (*models.ContactItemsResponse, error) {
	sp, ctx := apm.StartSpan(ses.Context(), "ContactItemsCreate", "app")
	ses.SetContext(ctx)
	defer sp.End()
	if data.SchoolId == nil && ses.GetSchoolId() != nil {
		data.SchoolId = ses.GetSchoolId()
	}
	model := &models.ContactItems{}
	data.ToModel(model)
	res := &models.ContactItemsResponse{}
	var err error
	model, err = store.Store().ContactItemCreate(ses.Context(), model)
	if err != nil {
		return nil, err
	}
	err = store.Store().ContactItemLoadRelations(ses.Context(), &[]*models.ContactItems{model}, true)
	if err != nil {
		return nil, err
	}
	res.FromModel(model)
	return res, nil
}

func ContactItemsDelete(ses *utils.Session, ids []string) ([]*models.ContactItems, error) {
	sp, ctx := apm.StartSpan(ses.Context(), "ContactItemsDelete", "app")
	ses.SetContext(ctx)
	defer sp.End()
	contactItems, err := store.Store().ContactItemsFindByIds(ses.Context(), ids)
	if err != nil {
		return nil, err
	}
	if len(contactItems) < 1 {
		return nil, errors.New("model not found: " + strings.Join(ids, ","))
	}
	return store.Store().ContactItemsDelete(ses.Context(), contactItems)
}
