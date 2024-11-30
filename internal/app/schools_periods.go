package app

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/mekdep/server/internal/api/utils"
	"github.com/mekdep/server/internal/models"
	"github.com/mekdep/server/internal/store"
	"go.elastic.co/apm/v2"
)

func PeriodsList(ses *utils.Session, data models.PeriodFilterRequest) ([]*models.PeriodResponse, int, error) {
	sp, ctx := apm.StartSpan(ses.Context(), "PeriodsList", "app")
	ses.SetContext(ctx)
	defer sp.End()
	if data.Limit == nil {
		data.Limit = new(int)
		*data.Limit = 12
	}
	if data.Offset == nil {
		data.Offset = new(int)
	}
	l, total, err := store.Store().PeriodsListFilters(ses.Context(), data)
	if err != nil {
		return nil, 0, err
	}
	err = store.Store().PeriodsLoadRelations(ses.Context(), &l)
	if err != nil {
		return nil, 0, err
	}
	res := []*models.PeriodResponse{}
	now := time.Now()
	for _, m := range l {
		item := models.PeriodResponse{}
		item.FromModel(m)
		if i, err := periodGetKey(*m, now); err == nil && i > 0 {
			item.CurrentIndex = &i
		}
		res = append(res, &item)
	}
	return res, total, nil
}

func PeriodDetail(ses *utils.Session, id string) (*models.PeriodResponse, error) {
	sp, ctx := apm.StartSpan(ses.Context(), "PeriodDetail", "app")
	ses.SetContext(ctx)
	defer sp.End()
	m, err := store.Store().PeriodsFindById(ses.Context(), id)
	if err != nil {
		return nil, err
	}
	err = store.Store().PeriodsLoadRelations(ses.Context(), &[]*models.Period{m})
	if err != nil {
		return nil, err
	}
	res := &models.PeriodResponse{}
	res.FromModel(m)
	return res, nil
}

func PeriodsUpdate(ses *utils.Session, data *models.PeriodRequest) (*models.PeriodResponse, error) {
	sp, ctx := apm.StartSpan(ses.Context(), "PeriodsUpdate", "app")
	ses.SetContext(ctx)
	defer sp.End()
	// request to model
	m := models.Period{}
	data.ToModel(&m)
	// validate value dates
	for perK, per := range *data.Value {
		for dateK, date := range per {
			_, err := time.Parse(time.DateOnly, date)
			if err != nil {
				return nil, NewAppError("invalid", fmt.Sprintf("value.%d.%d", perK, dateK), err.Error())
			}
		}
	}
	// request relation to model
	// ...
	// update by store
	updatedM, err := store.Store().PeriodsUpdate(ses.Context(), &m)
	if err != nil {
		return nil, err
	}
	err = store.Store().PeriodsLoadRelations(ses.Context(), &[]*models.Period{updatedM})
	if err != nil {
		return nil, err
	}

	// to response
	res := &models.PeriodResponse{}
	res.FromModel(updatedM)
	return res, nil
}

func PeriodsCreate(ses *utils.Session, data *models.PeriodRequest) (*models.PeriodResponse, error) {
	sp, ctx := apm.StartSpan(ses.Context(), "PeriodsCreate", "app")
	ses.SetContext(ctx)
	defer sp.End()
	if data.Title == nil {
		return nil, ErrRequired.SetKey("title")
	}
	if data.SchoolId == nil {
		return nil, ErrRequired.SetKey("school_id")
	}
	m := models.Period{}
	data.ToModel(&m)
	// validate value dates
	for perK, per := range *data.Value {
		for dateK, date := range per {
			_, err := time.Parse(time.DateOnly, date)
			if err != nil {
				return nil, NewAppError("invalid", fmt.Sprintf("value.%d.%d", perK, dateK), err.Error())
			}
		}
	}
	updatedM, err := store.Store().PeriodsCreate(ses.Context(), &m)
	if err != nil {
		return nil, err
	}
	err = store.Store().PeriodsLoadRelations(ses.Context(), &[]*models.Period{updatedM})
	if err != nil {
		return nil, err
	}

	res := &models.PeriodResponse{}
	res.FromModel(updatedM)
	return res, nil
}

func PeriodsDelete(ses *utils.Session, ids []string) ([]*models.Period, error) {
	sp, ctx := apm.StartSpan(ses.Context(), "PeriodsDelete", "app")
	ses.SetContext(ctx)
	defer sp.End()
	l, err := store.Store().PeriodsFindByIds(ses.Context(), ids)
	if err != nil {
		return nil, err
	}
	if len(l) < 1 {
		return nil, errors.New("user not found: " + strings.Join(ids, ","))
	}
	return store.Store().PeriodsDelete(ses.Context(), l)
}

func periodGetKey(m models.Period, t time.Time) (index int, err error) {
	var prevT1, prevT2 time.Time
	// for using prev dates, for tmp adding something to end
	m.Value = append(m.Value, m.Value[0])
	for k, i := range m.Value {
		if len(i) < 2 || (i[0] == "" || i[1] == "") {
			continue
		}
		t1, err := time.Parse(time.DateOnly, i[0]) // period start
		if err != nil {
			return 0, err
		}
		t2, err := time.Parse(time.DateOnly, i[1]) // period end
		if err != nil {
			return 0, err
		}
		// if prev end date is more than start date, then use prev end date as start date
		if t1.Compare(prevT1) <= 0 {
			t1 = prevT2
		}
		// check with start date and next start date
		if !prevT1.IsZero() && !prevT2.IsZero() {
			if prevT1.Compare(t) <= 0 && t1.Compare(t) > 0 {
				// added +1 to key (start not 0, but 1), because  it is next key
				return k, nil
			}
		}
		// save as prev dates
		prevT1, prevT2 = t1, t2

	}
	return index, nil
}

func periodsGetByDate(ses *utils.Session, t time.Time, schoolId string) (*models.Period, int, error) {
	sp, ctx := apm.StartSpan(ses.Context(), "periodsGetByDate", "app")
	ses.SetContext(ctx)
	defer sp.End()
	// TODO: awgust girer yaly etmeli, (indiki mowsum basyna 15gun galanda)
	l, _, err := store.Store().PeriodsListFilters(ses.Context(), models.PeriodFilterRequest{
		SchoolId: &schoolId,
	})
	if err != nil {
		return nil, 0, err
	}
	for _, m := range l {
		if i, err := periodGetKey(*m, t); err == nil && i > 0 {
			return m, i, nil
		} else if err != nil {
			return nil, 0, err
		}
		if i, err := periodGetKey(*m, t.AddDate(0, 0, -15)); err == nil && i > 0 {
			return m, i, nil
		} else if err != nil {
			return nil, 0, err
		}
		if i, err := periodGetKey(*m, t.AddDate(0, 0, 15)); err == nil && i > 0 {
			return m, i, nil
		} else if err != nil {
			return nil, 0, err
		}
	}
	return nil, 0, nil
}

func periodsGetByKey(m models.Period, num int) (time.Time, time.Time, error) {
	if len(m.Value) >= num && num > 0 {
		dates := m.Value[num-1]
		if len(dates) > 1 {
			start, err := time.Parse(time.DateOnly, dates[0])
			if err != nil {
				return time.Time{}, time.Time{}, err
			}
			end, err := time.Parse(time.DateOnly, dates[1])
			if err != nil {
				return time.Time{}, time.Time{}, err
			}
			return start, end, nil
		}
	}
	return time.Time{}, time.Time{}, errors.New("no periods found")
}
