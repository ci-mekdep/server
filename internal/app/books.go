package app

import (
	"errors"
	"strings"

	"github.com/mekdep/server/internal/api/utils"
	"github.com/mekdep/server/internal/models"
	"github.com/mekdep/server/internal/store"
	"go.elastic.co/apm/v2"
)

func BookList(ses *utils.Session, f models.BookFilterRequest) ([]*models.BookResponse, int, error) {
	sp, ctx := apm.StartSpan(ses.Context(), "BookList", "app")
	ses.SetContext(ctx)
	defer sp.End()
	books, total, err := store.Store().BookFindBy(ses.Context(), f)
	if err != nil {
		return nil, 0, err
	}
	booksResponse := []*models.BookResponse{}
	for _, shift := range books {
		s := models.BookResponse{}
		s.FromModel(shift)
		booksResponse = append(booksResponse, &s)
	}
	return booksResponse, total, err
}

func BookDetail(ses *utils.Session, id string) (*models.BookResponse, error) {
	sp, ctx := apm.StartSpan(ses.Context(), "BookDetail", "app")
	ses.SetContext(ctx)
	defer sp.End()
	m, err := store.Store().BookFindById(ses.Context(), id)
	if err != nil {
		return nil, err
	}
	res := &models.BookResponse{}
	res.FromModel(m)
	return res, nil
}

func BookGetAuthors(ses *utils.Session) ([]string, error) {
	sp, ctx := apm.StartSpan(ses.Context(), "BookGetAuthors", "app")
	ses.SetContext(ctx)
	defer sp.End()
	authors, err := store.Store().BookGetAuthors(ses.Context())
	if err != nil {
		return nil, err
	}
	return authors, nil
}

func BookUpdate(ses *utils.Session, data models.BookRequest) (*models.BookResponse, *models.Book, error) {
	sp, ctx := apm.StartSpan(ses.Context(), "BookUpdate", "app")
	ses.SetContext(ctx)
	defer sp.End()
	model := &models.Book{}
	data.ToModel(model)

	var err error
	model, err = store.Store().BookUpdate(ses.Context(), model)
	if err != nil {
		return nil, nil, err
	}

	res := &models.BookResponse{}
	res.FromModel(model)
	return res, model, nil
}

func BookCreate(ses *utils.Session, data models.BookRequest) (*models.BookResponse, *models.Book, error) {
	sp, ctx := apm.StartSpan(ses.Context(), "BookCreate", "app")
	ses.SetContext(ctx)
	defer sp.End()
	model := &models.Book{}
	data.ToModel(model)
	res := &models.BookResponse{}
	var err error
	model, err = store.Store().BookCreate(ses.Context(), model)
	if err != nil {
		return nil, nil, err
	}
	res.FromModel(model)
	return res, model, nil
}

func BookDelete(ses *utils.Session, ids []string) ([]*models.Book, error) {
	sp, ctx := apm.StartSpan(ses.Context(), "BookDelete", "app")
	ses.SetContext(ctx)
	defer sp.End()
	books, err := store.Store().BookFindByIds(ses.Context(), ids)
	if err != nil {
		return nil, err
	}
	if len(books) < 1 {
		return nil, errors.New("model not found: " + strings.Join(ids, ","))
	}
	return store.Store().BookDelete(ses.Context(), books)
}
