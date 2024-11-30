package pgx

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"

	"github.com/mekdep/server/internal/models"
	"github.com/mekdep/server/internal/utils"
)

const sqlPeriodFields = sqlPeriodFieldsNoRelation + `, null, null`
const sqlPeriodFieldsNoRelation = `p.uid, p.title, p.value, p.data_counts, p.updated_at, p.created_at, p.school_uid`
const sqlPeriodSelect = `select ` + sqlPeriodFields + ` from periods p where uid = ANY($1::uuid[])`
const sqlPeriodSelectMany = `select ` + sqlPeriodFieldsNoRelation + `, count(distinct tt.uid) as timetables_count, count(distinct c.uid) as classrooms_count, count(*) over() as total from periods p
	left join timetables tt on (tt.period_uid=p.uid)
	left join classrooms c on (c.period_uid=p.uid) where p.uid=p.uid limit $1 offset $2`
const sqlPeriodInsert = `insert into periods`
const sqlPeriodUpdate = `update periods set uid=uid`
const sqlPeriodDelete = `delete from periods where uid = ANY($1::uuid[])`

const sqlPeriodSchool = `select ` + sqlSchoolFields + `, p.uid from periods p 
	right join schools s on (s.uid=p.school_uid) where p.uid = ANY($1::uuid[])`

func scanPeriod(rows pgx.Row, sub *models.Period, addColumns ...interface{}) (err error) {
	err = rows.Scan(parseColumnsForScan(sub, addColumns...)...)
	return
}

func (d *PgxStore) PeriodsFindByIds(ctx context.Context, ids []string) ([]*models.Period, error) {
	l := []*models.Period{}
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		rows, err := tx.Query(ctx, sqlPeriodSelect, (ids))
		for rows.Next() {
			u := models.Period{}

			err := scanPeriod(rows, &u)
			if err != nil {
				return err
			}
			l = append(l, &u)
		}
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}

	return l, nil
}

func (d *PgxStore) PeriodsFindById(ctx context.Context, id string) (*models.Period, error) {
	l, err := d.PeriodsFindByIds(ctx, []string{id})
	return l[0], err
}

func (d *PgxStore) PeriodsListFilters(ctx context.Context, f models.PeriodFilterRequest) ([]*models.Period, int, error) {
	if f.Limit == nil {
		f.Limit = new(int)
		*f.Limit = 12
	}
	if f.Offset == nil {
		f.Offset = new(int)
		*f.Offset = 0
	}
	args := []interface{}{f.Limit, f.Offset}
	qs, args := PeriodsListBuildQuery(f, args)
	l := []*models.Period{}
	var total int

	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		rows, err := tx.Query(ctx, qs, args...)
		for rows.Next() {
			sub := models.Period{}
			err = scanPeriod(rows, &sub, &total)
			if err != nil {
				return err
			}
			l = append(l, &sub)
		}
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, 0, err
	}

	return l, total, nil
}

func (d *PgxStore) PeriodsCreate(ctx context.Context, data *models.Period) (*models.Period, error) {
	qs, args := PeriodsCreateQuery(data)
	qs += " RETURNING uid"
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		err = tx.QueryRow(ctx, qs, args...).Scan(&data.ID)
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}

	editModel, err := d.PeriodsFindById(ctx, data.ID)
	if err != nil {
		return nil, err
	}
	if editModel == nil {
		return nil, errors.New("model not found: " + string(data.ID))
	}

	d.PeriodsUpdateRelations(ctx, data, editModel)
	return editModel, nil
}

func (d *PgxStore) PeriodsUpdate(ctx context.Context, data *models.Period) (*models.Period, error) {
	// origModel := d.UsersFindById(strconv.Itoa(int(model.ID)))
	qs, args := PeriodsUpdateQuery(data)
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		_, err = tx.Query(ctx, qs, args...)
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}

	editModel, err := d.PeriodsFindById(ctx, data.ID)
	if err != nil {
		return nil, err
	}
	if editModel == nil {
		return nil, errors.New("model not found: " + string(data.ID))
	}
	d.PeriodsUpdateRelations(ctx, data, editModel)
	return editModel, nil
}

func (d *PgxStore) PeriodsDelete(ctx context.Context, l []*models.Period) ([]*models.Period, error) {
	ids := []string{}
	for _, i := range l {
		ids = append(ids, i.ID)
	}
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		_, err = tx.Query(ctx, sqlPeriodDelete, (ids))
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}
	return l, nil
}

func PeriodsCreateQuery(m *models.Period) (string, []interface{}) {
	args := []interface{}{}
	cols := ""
	vals := ""
	q := PeriodAtomicQuery(m, true)
	for k, v := range q {
		args = append(args, v)
		cols += ", " + k
		vals += ", $" + strconv.Itoa(len(args))
	}
	qs := sqlPeriodInsert + " (" + strings.Trim(cols, ", ") + ") VALUES (" + strings.Trim(vals, ", ") + ")"
	return qs, args
}

func PeriodsUpdateQuery(m *models.Period) (string, []interface{}) {
	args := []interface{}{}
	sets := ""
	q := PeriodAtomicQuery(m, false)
	for k, v := range q {
		args = append(args, v)
		sets += ", " + k + "=$" + strconv.Itoa(len(args))
	}
	args = append(args, m.ID)
	qs := strings.ReplaceAll(sqlPeriodUpdate, "set uid=uid", "set uid=uid "+sets+" ") + "where uid=$" + strconv.Itoa(len(args))
	return qs, args
}

func (d *PgxStore) PeriodsUpdateRelations(ctx context.Context, data *models.Period, model *models.Period) error {
	return nil
}

func PeriodAtomicQuery(m *models.Period, isCreate bool) map[string]interface{} {
	q := map[string]interface{}{}
	if m.Title != "" {
		q["title"] = m.Title
	}
	if m.Value != nil {
		q["value"] = m.Value
	}
	if m.SchoolId != nil {
		q["school_uid"] = m.SchoolId
	}
	if isCreate {
		q["created_at"] = time.Now()
	}
	q["updated_at"] = time.Now()
	return q
}

func PeriodsListBuildQuery(f models.PeriodFilterRequest, args []interface{}) (string, []interface{}) {
	var wheres string = ""

	if f.ID != nil {
		args = append(args, f.ID)
		wheres += " and p.uid=$" + strconv.Itoa(len(args))
	}
	if f.SchoolId != nil && *f.SchoolId != "" {
		args = append(args, *f.SchoolId)
		wheres += " and p.school_uid=$" + strconv.Itoa(len(args))
	}
	if f.SchoolIds != nil {
		args = append(args, *f.SchoolIds)
		wheres += " and p.school_uid = ANY($" + strconv.Itoa(len(args)) + "::uuid[])"
	}
	if f.Search != nil {
		*f.Search = strings.ToLower(*f.Search)
		args = append(args, f.Search)
		wheres += " and (lower(title) like '%' || $" + strconv.Itoa(len(args)) + "|| '%')"
	}
	wheres += " group by p.uid "
	sortBy := " order by p.created_at desc"
	if f.Sort != nil {
		dir := "desc"
		if strings.HasSuffix(*f.Sort, "~") {
			dir = "asc"
		}
		*f.Sort = strings.ReplaceAll(*f.Sort, "~", "")
		if strings.Contains(sqlPeriodFields, *f.Sort) {
			sortBy = " order by " + *f.Sort + " " + dir
		}
	}
	wheres += sortBy
	qs := sqlPeriodSelectMany
	qs = strings.ReplaceAll(qs, "p.uid=p.uid", "p.uid=p.uid "+wheres+" ")
	return qs, args
}

func (d *PgxStore) PeriodsLoadRelations(ctx context.Context, l *[]*models.Period) error {
	ids := []string{}
	for _, m := range *l {
		ids = append(ids, m.ID)
	}
	if len(ids) < 1 {
		return nil
	}

	if rs, err := d.PeriodsLoadSchool(ctx, ids); err != nil {
		return err
	} else {
		var schoolParents []*models.School
		for _, r := range rs {
			for _, m := range *l {
				if r.Id == m.ID {
					schoolParents = append(schoolParents, r.Relation)
					m.School = r.Relation
				}
			}
		}
		err = d.SchoolsLoadParents(ctx, &schoolParents)
		if err != nil {
			return err
		}
	}
	return nil
}

type PeriodsLoadSchoolItem struct {
	Id       string
	Relation *models.School
}

func (d *PgxStore) PeriodsLoadSchool(ctx context.Context, ids []string) ([]PeriodsLoadSchoolItem, error) {
	res := []PeriodsLoadSchoolItem{}
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		rows, err := tx.Query(ctx, sqlPeriodSchool, (ids))
		for rows.Next() {
			sub := models.School{}
			pid := ""
			err = scanSchool(rows, &sub, &pid)
			if err != nil {
				return nil
			}
			res = append(res, PeriodsLoadSchoolItem{Id: pid, Relation: &sub})
		}
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}

	return res, nil

}
