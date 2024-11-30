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

// base
const sqlTimetableFields = `tt.uid, tt.school_uid, tt.classroom_uid, tt.shift_uid, tt.period_uid, tt.value, tt.updated_at, tt.created_at, tt.updated_by_uid`
const sqlTimetableSelect = `select ` + sqlTimetableFields + ` from timetables tt where uid = ANY($1::uuid[])`
const sqlTimetableSelectMany = `select ` + sqlTimetableFields + `, count(*) over() as total from timetables tt 
	left join classrooms cl on cl.uid=tt.classroom_uid where tt.uid=tt.uid limit $1 offset $2`
const sqlTimetableInsert = `insert into timetables`
const sqlTimetableUpdate = `update timetables tt set uid=uid`
const sqlTimetableDelete = `delete from timetables tt where uid = ANY($1::uuid[])`

// relations
const sqlTimetableSchool = `select ` + sqlSchoolFields + `, tt.uid from timetables tt 
	right join schools s on (s.uid=tt.school_uid) where tt.uid = ANY($1::uuid[])`
const sqlTimetableClassroom = `select ` + sqlClassroomFields + `, tt.uid from timetables tt 
	right join classrooms c on (c.uid=tt.classroom_uid)  where tt.uid = ANY($1::uuid[])`
const sqlTimetableShift = `select ` + sqlShiftFields + `, tt.uid from timetables tt 
	right join shifts sh on (sh.uid=tt.shift_uid) where tt.uid = ANY($1::uuid[])`
const sqlTimetableTeacher = `select ` + sqlUserFields + `, tt.uid from timetables tt 
	right join users u on (u.uid=tt.teacher_uid) where tt.uid = ANY($1::uuid[])`
const sqlTimetableStudent = `select ` + sqlUserFields + `, tt.uid from timetables tt 
	right join users u on (u.uid=tt.student_uid) where tt.uid = ANY($1::uuid[])`

// many to many
const sqlTimetableStudents = `select ` + sqlUserFields + `, utt.type, utt.timetable_uid from user_timetables uc 
	left join users u on utt.user_uid=u.uid where utt.timetable_uid = ANY($1::uuid[])`
const sqlTimetableStudentsDelete = `delete from user_timetables where timetable_uid=$1`
const sqlTimetableStudentsInsert = `insert into user_timetables (timetable_uid, user_uid, type) values ($1, $2, $3)`

func scanTimetable(rows pgx.Row, m *models.Timetable, addColumns ...interface{}) (err error) {
	err = rows.Scan(parseColumnsForScan(m, addColumns...)...)
	return
}

func (d *PgxStore) TimetablesFindByIds(ctx context.Context, ids []string) ([]*models.Timetable, error) {
	l := []*models.Timetable{}
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		rows, err := tx.Query(ctx, sqlTimetableSelect, (ids))
		for rows.Next() {
			m := models.Timetable{}
			err := scanTimetable(rows, &m)
			if err != nil {
				return err
			}
			l = append(l, &m)
		}
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}

	return l, nil
}

func (d *PgxStore) TimetableFindById(ctx context.Context, id string) (*models.Timetable, error) {
	l, err := d.TimetablesFindByIds(ctx, []string{id})
	if err != nil {
		return nil, err
	}
	if len(l) < 1 {
		return nil, errors.New("timetable not found by uid: " + id)
	}
	return l[0], nil
}

func (d *PgxStore) TimetablesFindBy(ctx context.Context, f models.TimetableFilterRequest) ([]*models.Timetable, int, error) {
	if f.Limit == nil {
		f.Limit = new(int)
		*f.Limit = 12
	}
	if f.Offset == nil {
		f.Offset = new(int)
		*f.Offset = 0
	}
	args := []interface{}{f.Limit, f.Offset}
	qs, args := TimetablesListBuildQuery(f, args)

	l := []*models.Timetable{}
	var total int

	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		rows, err := tx.Query(ctx, qs, args...)
		for rows.Next() {
			sub := models.Timetable{}
			err = scanTimetable(rows, &sub, &total)
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

func (d *PgxStore) TimetableUpdate(ctx context.Context, data *models.Timetable) (*models.Timetable, error) {
	qs, args := TimetableUpdateQuery(data)
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		_, err = tx.Exec(ctx, qs, args...)
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}

	editModel, err := d.TimetableFindById(ctx, data.ID)
	if err != nil {
		return nil, err
	}
	return editModel, nil
}

func (d *PgxStore) TimetableCreate(ctx context.Context, m *models.Timetable) (*models.Timetable, error) {
	qs, args := TimetableCreateQuery(m)
	qs += " RETURNING uid"
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		err = tx.QueryRow(ctx, qs, args...).Scan(&m.ID)
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}

	editModel, err := d.TimetableFindById(ctx, m.ID)
	if err != nil {
		return nil, err
	}
	return editModel, nil
}

func (d *PgxStore) TimetablesDelete(ctx context.Context, l []*models.Timetable) ([]*models.Timetable, error) {
	ids := []string{}
	for _, i := range l {
		ids = append(ids, i.ID)
	}
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		_, err = tx.Query(ctx, sqlTimetableDelete, (ids))
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}
	return l, nil
}

func (d *PgxStore) TimetableUpdateRelations(ctx context.Context, data *models.Timetable, model *models.Timetable) {

}

func TimetableCreateQuery(m *models.Timetable) (string, []interface{}) {
	args := []interface{}{}
	cols := ""
	vals := ""
	q := TimetableAtomicQuery(m, true)
	for k, v := range q {
		args = append(args, v)
		cols += ", " + k
		vals += ", $" + strconv.Itoa(len(args))
	}
	qs := sqlTimetableInsert + " (" + strings.Trim(cols, ", ") + ") VALUES (" + strings.Trim(vals, ", ") + ")"
	return qs, args
}

func TimetableUpdateQuery(m *models.Timetable) (string, []interface{}) {
	args := []interface{}{}
	sets := ""
	q := TimetableAtomicQuery(m, false)
	for k, v := range q {
		args = append(args, v)
		sets += ", " + k + "=$" + strconv.Itoa(len(args))
	}
	args = append(args, m.ID)
	qs := strings.ReplaceAll(sqlTimetableUpdate, "set uid=uid", "set uid=uid "+sets+" ") + "where uid=$" + strconv.Itoa(len(args))
	return qs, args
}

func TimetableAtomicQuery(m *models.Timetable, isCreate bool) map[string]interface{} {
	q := map[string]interface{}{}
	if m.ClassroomId != "" {
		q["classroom_uid"] = m.ClassroomId
	}
	if m.ShiftId != nil {
		q["shift_uid"] = *m.ShiftId
	}
	if m.Value != nil {
		q["value"] = *m.Value
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

func TimetablesListBuildQuery(f models.TimetableFilterRequest, args []interface{}) (string, []interface{}) {
	var wheres string = ""

	if f.ID != nil && *f.ID != "" {
		args = append(args, *f.ID)
		wheres += " and tt.uid=$" + strconv.Itoa(len(args))
	}
	if f.SchoolId != nil && *f.SchoolId != "" {
		args = append(args, *f.SchoolId)
		wheres += " and tt.school_uid=$" + strconv.Itoa(len(args))
	}
	if f.SchoolIds != nil {
		args = append(args, *f.SchoolIds)
		wheres += " and tt.school_uid = ANY($" + strconv.Itoa(len(args)) + "::uuid[])"
	}
	if f.ClassroomIds != nil && len(*f.ClassroomIds) > 0 {
		args = append(args, *f.ClassroomIds)
		wheres += " and tt.classroom_uid = ANY($" + strconv.Itoa(len(args)) + "::uuid[])"
	}
	if f.ShiftIds != nil && len(*f.ShiftIds) > 0 {
		args = append(args, *f.ShiftIds)
		wheres += " and (tt.shift_uid = ANY($" + strconv.Itoa(len(args)) + ") )"
	}
	if f.Search != nil && *f.Search != "" {
		*f.Search = strings.ToLower(*f.Search)
		args = append(args, *f.Search)
		wheres += " and (lower(cl.name) like '%' || $" + strconv.Itoa(len(args)) + "|| '%')"
	}
	wheres += " group by tt.uid"
	if f.Sort != nil && *f.Sort != "" {
		dir := "desc"
		if strings.HasSuffix(*f.Sort, "~") {
			dir = "asc"
		}
		*f.Sort = strings.ReplaceAll(*f.Sort, "~", "")
		wheres += " order by " + *f.Sort + " " + dir
	} else {
		wheres += " order by LPAD(lower(MAX(cl.name)), 3, '0') asc"
	}
	qs := sqlTimetableSelectMany
	qs = strings.ReplaceAll(qs, "tt.uid=tt.uid", "tt.uid=tt.uid "+wheres+" ")
	return qs, args
}

func (d *PgxStore) TimetablesLoadRelations(ctx context.Context, l *[]*models.Timetable) error {
	ids := []string{}
	for _, m := range *l {
		ids = append(ids, m.ID)
	}
	if len(ids) < 1 {
		return nil
	}
	// load school
	if rs, err := d.TimetablesLoadSchool(ctx, ids); err == nil {
		var schoolParents []*models.School
		for _, r := range rs {
			for _, m := range *l {
				if r.ID == m.ID {
					schoolParents = append(schoolParents, r.Relation)
					m.School = r.Relation
				}
			}
		}
		err = d.SchoolsLoadParents(ctx, &schoolParents)
		if err != nil {
			return err
		}
	} else {
		return err
	}
	// load classroom
	if rs, err := d.TimetablesLoadClassroom(ctx, ids); err == nil {
		for _, r := range rs {
			for _, m := range *l {
				if r.ID == m.ID {
					m.Classroom = r.Relation
				}
			}
		}
	} else {
		return err
	}
	// load shift
	if rs, err := d.TimetablesLoadShift(ctx, ids); err == nil {
		for _, r := range rs {
			for _, m := range *l {
				if r.ID == m.ID {
					m.Shift = r.Relation
				}
			}
		}
	} else {
		return err
	}
	return nil
}

type TimetablesLoadSchoolItem struct {
	ID       string
	Relation *models.School
}

func (d *PgxStore) TimetablesLoadSchool(ctx context.Context, ids []string) ([]TimetablesLoadSchoolItem, error) {
	res := []TimetablesLoadSchoolItem{}
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		rows, err := tx.Query(ctx, sqlTimetableSchool, (ids))
		for rows.Next() {
			sub := models.School{}
			pid := ""
			err = scanSchool(rows, &sub, &pid)
			if err != nil {
				return err
			}
			res = append(res, TimetablesLoadSchoolItem{ID: pid, Relation: &sub})
		}
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}

	return res, nil
}

type TimetablesLoadClassroomItem struct {
	ID       string
	Relation *models.Classroom
}

func (d *PgxStore) TimetablesLoadClassroom(ctx context.Context, ids []string) ([]TimetablesLoadClassroomItem, error) {
	res := []TimetablesLoadClassroomItem{}
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		rows, err := tx.Query(ctx, sqlTimetableClassroom, (ids))
		for rows.Next() {
			sub := models.Classroom{}
			pid := ""
			err = scanClassroom(rows, &sub, &pid)
			if err != nil {
				return err
			}
			res = append(res, TimetablesLoadClassroomItem{ID: pid, Relation: &sub})
		}
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}

	return res, nil
}

type TimetablesLoadShiftItem struct {
	ID       string
	Relation *models.Shift
}

func (d *PgxStore) TimetablesLoadShift(ctx context.Context, ids []string) ([]TimetablesLoadShiftItem, error) {
	res := []TimetablesLoadShiftItem{}
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		rows, err := tx.Query(ctx, sqlTimetableShift, (ids))
		for rows.Next() {
			sub := models.Shift{}
			pid := ""
			err = scanShift(rows, &sub, &pid)
			if err != nil {
				return err
			}
			res = append(res, TimetablesLoadShiftItem{ID: pid, Relation: &sub})
		}
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}

	return res, nil
}
