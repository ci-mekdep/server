package app

import (
	"errors"
	"strings"

	"github.com/mekdep/server/internal/api/utils"
	"github.com/mekdep/server/internal/models"
	"github.com/mekdep/server/internal/store"
	"go.elastic.co/apm/v2"
)

func (a App) Schools(ses *utils.Session, data models.SchoolFilterRequest) ([]*models.SchoolResponse, int, error) {
	sp, ctx := apm.StartSpan(ses.Context(), "Schools", "app")
	ses.SetContext(ctx)
	defer sp.End()
	l, total, err := store.Store().SchoolsFindBy(ses.Context(), data)
	if err != nil {
		return nil, 0, err
	}
	err = store.Store().SchoolsLoadRelations(ses.Context(), &l)
	if err != nil {
		return nil, 0, err
	}

	res := []*models.SchoolResponse{}
	for _, m := range l {
		item := models.SchoolResponse{}
		item.FromModel(m)
		res = append(res, &item)
	}
	return res, total, nil
}

func SchoolsListValues(ses *utils.Session, data models.SchoolFilterRequest) ([]models.SchoolValueResponse, int, error) {
	sp, ctx := apm.StartSpan(ses.Context(), "SchoolsListValues", "app")
	ses.SetContext(ctx)
	defer sp.End()
	res, total, err := Ap().Schools(ses, data)
	if err != nil {
		return nil, 0, err
	}
	values := []models.SchoolValueResponse{}
	for _, v := range res {
		values = append(values, v.ToValues())
	}
	return values, total, nil
}

func SchoolsDetail(ses *utils.Session, id string) (*models.SchoolResponse, error) {
	sp, ctx := apm.StartSpan(ses.Context(), "SchoolsDetail", "app")
	ses.SetContext(ctx)
	defer sp.End()
	m, err := store.Store().SchoolsFindById(ses.Context(), id)
	if err != nil {
		return nil, err
	}
	err = store.Store().SchoolsLoadRelations(ses.Context(), &[]*models.School{m})
	if err != nil {
		return nil, err
	}
	res := &models.SchoolResponse{}
	res.FromModel(m)
	return res, nil
}

func SchoolsCreate(ses *utils.Session, data *models.SchoolRequest) (*models.SchoolResponse, error) {
	sp, ctx := apm.StartSpan(ses.Context(), "SchoolsCreate", "app")
	ses.SetContext(ctx)
	defer sp.End()
	dataModel := models.School{}
	data.ToModel(&dataModel)
	var err error
	schools, _, err := store.Store().SchoolsFindBy(ses.Context(), models.SchoolFilterRequest{})
	if err != nil {
		return nil, err
	}
	for _, v := range schools {
		if v.Code != nil && dataModel.Code != nil && *v.Code == *dataModel.Code {
			return nil, ErrUnique.SetKey("code")
		}
	}
	m, err := store.Store().SchoolCreate(ses.Context(), &dataModel)
	if err != nil {
		return nil, err
	}
	err = store.Store().SchoolUpdateRelations(ses.Context(), &dataModel, m)
	if err != nil {
		return nil, err
	}
	err = store.Store().SchoolsLoadRelations(ses.Context(), &[]*models.School{m})
	if err != nil {
		return nil, err
	}
	res := &models.SchoolResponse{}
	res.FromModel(m)
	return res, nil
}

func SchoolsUpdate(ses *utils.Session, data *models.SchoolRequest) (*models.SchoolResponse, error) {
	sp, ctx := apm.StartSpan(ses.Context(), "SchoolsUpdate", "app")
	ses.SetContext(ctx)
	defer sp.End()
	dataModel := models.School{}
	data.ToModel(&dataModel)
	var err error
	if err = SchoolIsExists(ses, &dataModel); err != nil {
		return nil, err
	}
	// check admin_id
	if dataModel.AdminUid != nil {
		adminIdStr := (*dataModel.AdminUid)
		admin, err := store.Store().UsersFindById(ses.Context(), adminIdStr)
		if err != nil || admin == nil {
			return nil, ErrInvalid.SetKey("admin_id")
		}
		err = store.Store().UsersLoadRelations(ses.Context(), &[]*models.User{admin}, false)
		if err != nil {
			return nil, err
		}
		// is admin has correct role and school
		isAdminOk := false
		for _, v := range admin.Schools {
			if v.RoleCode == models.RolePrincipal && v.SchoolUid != nil && *v.SchoolUid == dataModel.ID {
				isAdminOk = true
			}
		}
		if !isAdminOk {
			return nil, ErrInvalid.SetKey("admin_id")
		}
	}

	m, err := store.Store().SchoolUpdate(ses.Context(), &dataModel)
	if err != nil {
		return nil, err
	}
	err = store.Store().SchoolUpdateRelations(ses.Context(), &dataModel, m)
	if err != nil {
		return nil, err
	}
	err = store.Store().SchoolsLoadRelations(ses.Context(), &[]*models.School{m})
	if err != nil {
		return nil, err
	}
	res := &models.SchoolResponse{}
	res.FromModel(m)
	return res, nil
}

func SchoolIsExists(ses *utils.Session, data *models.School) error {
	_, t, err := store.Store().SchoolsFindBy(ses.Context(), models.SchoolFilterRequest{
		NotIds: &[]string{data.ID},
		Code:   data.Code,
	})
	if err != nil {
		return err
	}
	if t > 0 {
		return ErrUnique.SetKey("school")
	}
	return nil
}

func SchoolsDelete(ses *utils.Session, ids []string) ([]*models.School, error) {
	sp, ctx := apm.StartSpan(ses.Context(), "SchoolsDelete", "app")
	ses.SetContext(ctx)
	defer sp.End()
	l, err := store.Store().SchoolsFindByIds(ses.Context(), ids)
	if err != nil {
		return nil, err
	}
	if len(l) < 1 {
		return nil, errors.New("school not found: " + strings.Join(ids, ","))
	}
	return store.Store().SchoolDelete(ses.Context(), l)
}
