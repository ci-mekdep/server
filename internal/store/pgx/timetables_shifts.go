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

const sqlShiftFields = sqlShiftFieldsNoRelation + `, null, null`
const sqlShiftFieldsNoRelation = `sh.uid, sh.name, sh.school_uid, sh.value, sh.updated_by_uid, sh.created_at, sh.updated_at`
const sqlShiftSelect = `SELECT ` + sqlShiftFields + ` FROM shifts sh WHERE sh.uid = ANY($1::uuid[])`
const sqlShiftSelectMany = `SELECT ` + sqlShiftFieldsNoRelation + `, count(tt.uid) as timetables_count, count(distinct c.uid) as classrooms_count, count(1) over() as total FROM shifts sh
	left join timetables tt on (tt.shift_uid=sh.uid) 
	left join classrooms c on (c.uid=tt.classroom_uid) where sh.uid=sh.uid limit $1 offset $2 `
const sqlShiftUpdate = `UPDATE shifts sh set uid=uid`
const sqlShiftInsert = `INSERT INTO shifts`
const sqlShiftDelete = `delete from shifts sh where uid = ANY($1::uuid[])`

const sqlShiftSchool = `select ` + sqlSchoolFields + `, sh.uid from shifts sh
	right join schools s on (s.uid=sh.school_uid) where sh.uid = ANY($1::uuid[])`

func scanShift(rows pgx.Row, m *models.Shift, addColumns ...interface{}) (err error) {
	err = rows.Scan(parseColumnsForScan(m, addColumns...)...)
	return
}

func (d *PgxStore) ShiftsFindBy(ctx context.Context, f models.ShiftFilterRequest) (shifts []*models.Shift, total int, err error) {
	args := []interface{}{f.Limit, f.Offset}
	qs, args := ShiftsListBuildQuery(f, args)
	err = d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		rows, err := tx.Query(ctx, qs, args...)
		for rows.Next() {
			shift := models.Shift{}
			err = scanShift(rows, &shift, &total)
			if err != nil {
				return err
			}
			shifts = append(shifts, &shift)
		}
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, 0, err
	}

	return shifts, total, nil
}

func (d *PgxStore) ShiftsFindById(ctx context.Context, Id string) (*models.Shift, error) {
	row, err := d.ShiftsFindByIds(ctx, []string{Id})
	if err != nil {
		return nil, err
	}
	if len(row) < 1 {
		return nil, errors.New("shift not found by uid: " + Id)
	}
	return row[0], nil
}

func (d *PgxStore) ShiftsFindByIds(ctx context.Context, Ids []string) ([]*models.Shift, error) {
	shifts := []*models.Shift{}
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		rows, err := tx.Query(ctx, sqlShiftSelect, (Ids))
		for rows.Next() {
			shift := models.Shift{}
			err := scanShift(rows, &shift)
			if err != nil {
				return err
			}
			shifts = append(shifts, &shift)
		}
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}

	return shifts, nil
}

func (d *PgxStore) UpdateShift(ctx context.Context, model *models.Shift) (*models.Shift, error) {
	qs, args := ShiftUpdateQuery(model)
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		_, err = tx.Query(ctx, qs, args...)
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}
	editModel, err := d.ShiftsFindById(ctx, model.Id)
	if err != nil {
		return nil, err
	}
	d.ShiftUpdateRelations(ctx, model, editModel)
	return editModel, nil
}

func (d *PgxStore) CreateShift(ctx context.Context, model *models.Shift) (*models.Shift, error) {
	qs, args := ShiftCreateQuery(model)
	qs += " RETURNING uid"
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		err = tx.QueryRow(ctx, qs, args...).Scan(&model.Id)
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}
	editModel, err := d.ShiftsFindById(ctx, model.Id)
	if err != nil {
		return nil, err
	}
	d.ShiftUpdateRelations(ctx, model, editModel)
	return editModel, nil
}

func (d *PgxStore) DeleteShifts(ctx context.Context, items []*models.Shift) ([]*models.Shift, error) {
	ids := []string{}
	for _, i := range items {
		ids = append(ids, i.Id)
	}
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		_, err = tx.Query(ctx, sqlShiftDelete, (ids))
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}
	return items, nil
}

func (d *PgxStore) ShiftUpdateRelations(ctx context.Context, data *models.Shift, model *models.Shift) {

}

func ShiftCreateQuery(m *models.Shift) (string, []interface{}) {
	args := []interface{}{}
	cols := ""
	vals := ""
	q := ShiftAtomicQuery(m, true)
	for k, v := range q {
		args = append(args, v)
		cols += ", " + k
		vals += ", $" + strconv.Itoa(len(args))
	}
	qs := sqlShiftInsert + " (" + strings.Trim(cols, ", ") + ") VALUES (" + strings.Trim(vals, ", ") + ")"
	return qs, args
}

func ShiftUpdateQuery(m *models.Shift) (string, []interface{}) {
	args := []interface{}{}
	sets := ""
	q := ShiftAtomicQuery(m, false)
	for k, v := range q {
		args = append(args, v)
		sets += ", " + k + "=$" + strconv.Itoa(len(args))
	}
	args = append(args, m.Id)
	qs := strings.ReplaceAll(sqlShiftUpdate, "set uid=uid", "set uid=uid "+sets+" ") + "where uid=$" + strconv.Itoa(len(args))
	return qs, args
}

func ShiftAtomicQuery(m *models.Shift, isCreate bool) map[string]interface{} {
	q := map[string]interface{}{}
	if m.Value != nil {
		q["value"] = m.Value
	}
	if m.Name != nil {
		q["name"] = m.Name
	}
	if m.SchoolId != "" {
		q["school_uid"] = m.SchoolId
	}
	if isCreate {
		q["created_at"] = time.Now()
	}
	q["updated_at"] = time.Now()
	q["updated_by_uid"] = m.UpdatedBy
	return q
}

func ShiftsListBuildQuery(f models.ShiftFilterRequest, args []interface{}) (string, []interface{}) {
	var wheres string = ""

	if f.ID != nil && *f.ID != "" {
		args = append(args, *f.ID)
		wheres += " and sh.uid=$" + strconv.Itoa(len(args))
	}
	if f.IDs != nil && len(*f.IDs) > 0 {
		args = append(args, *f.IDs)
		wheres += " and (sh.uid = ANY($" + strconv.Itoa(len(args)) + ") )"
	}
	if f.SchoolId != nil && *f.SchoolId != "" {
		args = append(args, *f.SchoolId)
		wheres += " and sh.school_uid=$" + strconv.Itoa(len(args))
	}
	if f.ClassroomId != nil && *f.ClassroomId != "" {
		args = append(args, *f.ClassroomId)
		wheres += " and tt.classroom_uid=$" + strconv.Itoa(len(args))
	}
	if f.TimetableId != nil && *f.TimetableId != "" {
		args = append(args, *f.TimetableId)
		wheres += " and tt.uid=$" + strconv.Itoa(len(args))
	}
	if f.SchoolIds != nil {
		args = append(args, *f.SchoolIds)
		wheres += " and sh.school_uid = ANY($" + strconv.Itoa(len(args)) + "::uuid[])"
	}
	if f.Search != nil && *f.Search != "" {
		*f.Search = strings.ToLower(*f.Search)
		args = append(args, *f.Search)
		wheres += " and (lower(sh.name) like '%' || $" + strconv.Itoa(len(args)) + "|| '%')"
	}
	wheres += " group by sh.uid "
	wheres += " order by sh.created_at desc"

	qs := sqlShiftSelectMany
	qs = strings.ReplaceAll(qs, "sh.uid=sh.uid", "sh.uid=sh.uid "+wheres+" ")

	return qs, args
}

func (d *PgxStore) ShiftLoadRelations(ctx context.Context, l *[]*models.Shift) error {
	ids := []string{}
	for _, m := range *l {
		ids = append(ids, m.Id)
	}
	if len(ids) < 1 {
		return nil
	}

	if rs, err := d.ShiftLoadSchool(ctx, ids); err != nil {
		return err
	} else {
		var schoolParents []*models.School
		for _, r := range rs {
			for _, m := range *l {
				if r.Id == m.Id {
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

type ShiftLoadSchoolItem struct {
	Id       string
	Relation *models.School
}

func (d *PgxStore) ShiftLoadSchool(ctx context.Context, ids []string) ([]ShiftLoadSchoolItem, error) {
	res := []ShiftLoadSchoolItem{}
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		rows, err := tx.Query(ctx, sqlShiftSchool, (ids))
		for rows.Next() {
			sub := models.School{}
			pid := ""
			err = scanSchool(rows, &sub, &pid)
			if err != nil {
				return err
			}
			res = append(res, ShiftLoadSchoolItem{Id: pid, Relation: &sub})
		}
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}

	return res, nil
}
