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
const sqlGradeFields = `g.uid, g.lesson_uid, g.student_uid, g.value::int, g.values::int[], g.reason, g.comment, g.parent_reviewed_at, g.created_at, g.updated_at, g.created_by_uid, g.updated_by_uid, g.deleted_at`
const sqlGradeSelect = `select ` + sqlGradeFields + `  from grades g where uid = ANY($1::uuid[])`
const sqlGradeSelectMany = `select ` + sqlGradeFields + `, 1 from grades g where uid=uid limit $1 offset $2`
const sqlGradeInsert = `insert into grades`
const sqlGradeUpdate = `update grades g set uid=uid`
const sqlGradeCreateOrUpdate = `INSERT INTO grades (value, values, comment, student_uid, lesson_uid, created_by_uid, updated_by_uid) 
	values ($1, $2, $3, $4, $5, $6, $7) ON CONFLICT (lesson_uid, student_uid) DO UPDATE SET 
	value = EXCLUDED.value, values = EXCLUDED.values, comment = EXCLUDED.comment, updated_by_uid = EXCLUDED.updated_by_uid,  updated_at = NOW() RETURNING uid;`
const sqlGradeDelete = `delete from grades g where lesson_uid = ANY($1::uuid[]) and student_uid = ANY($2::uuid[])`

// relations
const sqlGradeLesson = `select ` + sqlLessonFields + `, g.uid from lessons l 
	right join grades g on g.lesson_uid=l.uid where g.uid = ANY($1::uuid[])`

// many to many
const sqlGradeStudents = `select ` + sqlUserFields + `, ug.type, ug.classroom_uid from user_classrooms uc 
	left join users u on ug.user_uid=u.uid where ug.classroom_uid = ANY($1::uuid[])`
const sqlGradeStudentsDelete = `delete from user_classrooms where classroom_uid=$1`
const sqlGradeStudentsInsert = `insert into user_classrooms (classroom_uid, user_uid, type) values ($1, $2, $3)`

func scanGrade(rows pgx.Row, m *models.Grade, addColumns ...interface{}) (err error) {
	err = rows.Scan(parseColumnsForScan(m, addColumns...)...)
	return
}

func (d *PgxStore) GradesFindByIds(ctx context.Context, ids []string) ([]models.Grade, error) {
	l := []models.Grade{}

	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		rows, err := tx.Query(ctx, sqlGradeSelect, (ids))
		for rows.Next() {
			m := models.Grade{}
			err := scanGrade(rows, &m)
			if err != nil {
				utils.LoggerDesc("Scan error").Error(err)
				return err
			}
			l = append(l, m)
		}
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}

	return l, nil
}

func (d *PgxStore) GradesFindById(ctx context.Context, id string) (models.Grade, error) {
	l, err := d.GradesFindByIds(ctx, []string{id})
	if err != nil {
		return models.Grade{}, err
	}
	if len(l) < 1 {
		err = pgx.ErrNoRows
		utils.LoggerDesc("Scan error").Error(err)
		return models.Grade{}, err
	}
	return l[0], nil
}

func (d *PgxStore) GradesFindBy(ctx context.Context, f models.GradeFilterRequest) ([]*models.Grade, int, error) {
	if f.Limit == nil {
		f.Limit = new(int)
		*f.Limit = 1000
	}
	if f.Offset == nil {
		f.Offset = new(int)
		*f.Offset = 0
	}
	args := []interface{}{f.Limit, f.Offset}
	qs, args := GradesListQuery(f, args)

	l := []*models.Grade{}
	var total int

	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		rows, err := tx.Query(ctx, qs, args...)
		for rows.Next() {
			sub := models.Grade{}
			err = scanGrade(rows, &sub, &total)
			if err != nil {
				utils.LoggerDesc("Scan error").Error(err)
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

func (d *PgxStore) GradesCreateOrUpdate(ctx context.Context, data models.Grade) (models.Grade, error) {
	// origModel := d.UsersFindById(strconv.Itoa(int(model.ID)))
	qs, args := gradesCreateOrUpdateQuery(data)
	uid := ""
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		r := tx.QueryRow(ctx, qs, args...)
		err = r.Scan(&uid)
		if err != nil {
			return err
		}
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return models.Grade{}, err
	}
	data.ID = uid
	return data, nil
}

func (d *PgxStore) GradesUpdate(ctx context.Context, data models.Grade) (models.Grade, error) {
	// origModel := d.UsersFindById(strconv.Itoa(int(model.ID)))
	qs, args := GradesUpdateQuery(data)
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		_, err = tx.Query(ctx, qs, args...)
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return models.Grade{}, err
	}

	editModel, err := d.GradesFindById(ctx, data.ID)
	if err != nil {
		return models.Grade{}, err
	}
	if editModel.ID == "" {
		return models.Grade{}, errors.New("model not found: " + string(data.ID))
	}
	d.GradesUpdateRelations(data, editModel)
	return editModel, nil
}

func (d *PgxStore) GradesCreate(ctx context.Context, m models.Grade) (models.Grade, error) {
	qs, args := GradesCreateQuery(m)
	qs += " RETURNING uid"
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		err = tx.QueryRow(ctx, qs, args...).Scan(&m.ID)
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return models.Grade{}, err
	}

	editModel, err := d.GradesFindById(ctx, m.ID)
	if err != nil {
		return models.Grade{}, err
	}
	if editModel.ID == "" {
		return models.Grade{}, errors.New("model not found: " + string(m.ID))
	}
	d.GradesUpdateRelations(m, editModel)
	return editModel, nil
}

func (d *PgxStore) GradesDelete(ctx context.Context, l []*models.Grade) ([]*models.Grade, error) {
	lids := []string{}
	for _, i := range l {
		lids = append(lids, i.LessonId)
	}
	sids := []string{}
	for _, i := range l {
		sids = append(sids, i.StudentId)
	}
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		_, err = tx.Query(ctx, sqlGradeDelete, lids, sids)
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}
	return l, nil
}

func (d *PgxStore) GradesUpdateRelations(data models.Grade, model models.Grade) error {

	return nil
}

func GradesCreateQuery(m models.Grade) (string, []interface{}) {
	args := []interface{}{}
	cols := ""
	vals := ""
	q := GradeAtomicQuery(m, true)
	for k, v := range q {
		args = append(args, v)
		cols += ", " + k
		vals += ", $" + strconv.Itoa(len(args))
	}
	qs := sqlGradeInsert + " (" + strings.Trim(cols, ", ") + ") VALUES (" + strings.Trim(vals, ", ") + ")"
	return qs, args
}

func GradesUpdateQuery(m models.Grade) (string, []interface{}) {
	args := []interface{}{}
	sets := ""
	q := GradeAtomicQuery(m, false)
	for k, v := range q {
		args = append(args, v)
		sets += ", " + k + "=$" + strconv.Itoa(len(args))
	}
	args = append(args, m.ID)
	qs := strings.ReplaceAll(sqlGradeUpdate, "set uid=uid", "set uid=uid "+sets+" ") + "where uid=$" + strconv.Itoa(len(args))
	return qs, args
}

func gradesCreateOrUpdateQuery(m models.Grade) (string, []interface{}) {
	args := []interface{}{}
	args = append(args, m.GetValue(), m.Values, m.Comment, m.StudentId, m.LessonId, m.CreatedBy, m.UpdatedBy)
	return sqlGradeCreateOrUpdate, args
}

func GradeAtomicQuery(m models.Grade, isCreate bool) map[string]interface{} {
	q := map[string]interface{}{}
	if m.Value != nil {
		q["value"] = strconv.Itoa(*m.Value)
	}
	if m.Values != nil {
		q["value"] = nil
		q["values"] = []string{}
		for _, v := range *m.Values {
			q["values"] = append(q["values"].([]string), strconv.Itoa(v))
		}
	}
	if m.Reason != nil {
		q["reason"] = *m.Reason
	}
	if m.Comment != nil {
		q["comment"] = *m.Comment
	}
	if m.ParentReviewedAt != nil {
		q["parent_reviewed_at"] = *m.ParentReviewedAt
	}
	if isCreate {
		if m.LessonId != "" {
			q["lesson_uid"] = m.LessonId
		}
		if m.StudentId != "" {
			q["student_uid"] = m.StudentId
		}
		q["created_at"] = time.Now()
		q["created_by_uid"] = m.UpdatedBy
	}
	q["updated_at"] = time.Now()
	q["updated_by_uid"] = m.UpdatedBy
	// todo: set by users to current session
	return q
}

func GradesListQuery(f models.GradeFilterRequest, args []interface{}) (string, []interface{}) {
	var wheres string = ""
	qs := sqlGradeSelectMany

	if f.ID != nil && *f.ID != "" {
		args = append(args, *f.ID)
		wheres += " and g.uid=$" + strconv.Itoa(len(args))
	}
	if f.SchoolId != nil && *f.SchoolId != "" {
		args = append(args, *f.SchoolId)
		wheres += " and g.school_uid=$" + strconv.Itoa(len(args))
	}
	if f.LessonId != nil && *f.LessonId != "" {
		args = append(args, *f.LessonId)
		wheres += " and g.lesson_uid=$" + strconv.Itoa(len(args))
	}
	if f.LessonIds != nil {
		args = append(args, *f.LessonIds)
		wheres += " and g.lesson_uid = ANY($" + strconv.Itoa(len(args)) + "::uuid[])"
	}
	if f.StudentId != nil && *f.StudentId != "" {
		args = append(args, *f.StudentId)
		wheres += " and g.student_uid=$" + strconv.Itoa(len(args))
	}
	if f.StudentIds != nil && len(*f.StudentIds) > 0 {
		args = append(args, *f.StudentIds)
		wheres += " and g.student_uid = ANY($" + strconv.Itoa(len(args)) + "::uuid[])"
	}
	if f.SchoolIds != nil {
		args = append(args, *f.SchoolIds)
		wheres += " and g.school_uid = ANY($" + strconv.Itoa(len(args)) + "::uuid[])"
	}
	//	wheres += " group by g.uid "
	//	wheres += " order by g.created_at desc"
	qs = strings.ReplaceAll(qs, "uid=uid", "uid=uid "+wheres+" ")
	return qs, args
}

// todo: check workings of grade relation lesson and apply others
func (d *PgxStore) GradesLoadRelations(ctx context.Context, l []*models.Grade) error {
	ids := []string{}
	for _, m := range l {
		ids = append(ids, m.ID)
	}
	if len(ids) < 1 {
		return nil
	}
	err := d.GradesLoadRelationLessons(ctx, l)
	if err != nil {
		return nil
	}
	return nil
}

func (d *PgxStore) GradesLoadRelationLessons(ctx context.Context, l []*models.Grade) error {
	ids := []string{}
	for _, m := range l {
		ids = append(ids, m.ID)
	}
	if len(ids) < 1 {
		return nil
	}

	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		rs, err := tx.Query(ctx, sqlGradeLesson, ids)
		for rs.Next() {
			m := models.Lesson{}
			pid := ""
			err = scanLesson(rs, &m, &pid)
			if err != nil {
				return err
			}
			for k, v := range l {
				if v.ID == pid {
					l[k].Lesson = new(models.Lesson)
					*l[k].Lesson = m
				}
			}
		}
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return err
	}
	return nil
}
