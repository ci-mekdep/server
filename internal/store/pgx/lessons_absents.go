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
const sqlAbsentFields = `ab.uid, ab.lesson_uid, ab.student_uid, ab.reason, ab.comment, ab.parent_reviewed_at, ab.updated_at, ab.created_at, ab.updated_by_uid, ab.created_by_uid, ab.deleted_at`
const sqlAbsentSelect = `select ` + sqlAbsentFields + `  from absents ab where uid = ANY($1::uuid[])`
const sqlAbsentSelectMany = `select ` + sqlAbsentFields + `, count(*) over() as total from absents ab where uid=uid limit $1 offset $2`
const sqlAbsentInsert = `insert into absents`
const sqlAbsentUpdate = `update absents ab set uid=uid`
const sqlAbsentDelete = `delete from absents ab where lesson_uid = ANY($1::uuid[]) and student_uid = ANY($2::uuid[])`
const sqlAbsentCreateOrUpdate = `INSERT INTO absents (reason, comment, student_uid, lesson_uid, created_by_uid, updated_by_uid) 
	values ($1, $2, $3, $4, $5, $6) ON CONFLICT (lesson_uid, student_uid) DO UPDATE SET 
	reason = EXCLUDED.reason, comment = EXCLUDED.comment, updated_by_uid = EXCLUDED.updated_by_uid,  updated_at = NOW() RETURNING uid;`

const sqlAbsentLesson = `select ` + sqlLessonFields + `, ab.uid from lessons l 
	right join absents ab on ab.lesson_uid=l.uid where ab.uid = ANY($1::uuid[])`

// many to many
const sqlAbsentStudents = `select ` + sqlUserFields + `, uc.type, uc.classroom_uid from user_classrooms uc 
	left join users u on uc.user_uid=u.uid where uc.classroom_uid = ANY($1::uuid[])`
const sqlAbsentStudentsDelete = `delete from user_classrooms where classroom_uid=$1`
const sqlAbsentStudentsInsert = `insert into user_classrooms (classroom_uid, user_uid, type) values ($1, $2, $3)`

func scanAbsent(rows pgx.Row, m *models.Absent, addColumns ...interface{}) (err error) {
	err = rows.Scan(parseColumnsForScan(m, addColumns...)...)
	return
}

func (d *PgxStore) AbsentsFindByIds(ctx context.Context, ids []string) ([]models.Absent, error) {
	l := []models.Absent{}

	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		rows, err := tx.Query(ctx, sqlAbsentSelect, (ids))
		for rows.Next() {
			m := models.Absent{}
			err := scanAbsent(rows, &m)
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

func (d *PgxStore) AbsentsFindById(ctx context.Context, id string) (models.Absent, error) {
	l, err := d.AbsentsFindByIds(ctx, []string{id})
	if err != nil {
		return models.Absent{}, err
	}
	if len(l) < 1 {
		err = pgx.ErrNoRows
		utils.LoggerDesc("Scan error").Error(err)
		return models.Absent{}, err
	}
	return l[0], nil
}

func (d *PgxStore) AbsentsFindBy(ctx context.Context, f models.AbsentFilterRequest) ([]*models.Absent, int, error) {
	// TODO refactor: Limit Offset-lary belli bir yerde goymak
	if f.Limit == nil {
		f.Limit = new(int)
		*f.Limit = 1000
	}
	if f.Offset == nil {
		f.Offset = new(int)
		*f.Offset = 0
	}
	args := []interface{}{f.Limit, f.Offset}
	qs, args := AbsentsListQuery(f, args)
	l := []*models.Absent{}
	var total int

	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		rows, err := tx.Query(ctx, qs, args...)
		if err != nil {
			return err
		}
		for rows.Next() {
			sub := models.Absent{}
			err = scanAbsent(rows, &sub, &total)
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

func (d *PgxStore) AbsentsCreateOrUpdate(ctx context.Context, data models.Absent) (models.Absent, error) {
	// origModel := d.UsersFindById(strconv.Itoa(int(model.ID)))
	qs, args := absentsCreateOrUpdateQuery(data)
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
		return models.Absent{}, err
	}
	data.ID = uid
	return data, nil
}

func absentsCreateOrUpdateQuery(m models.Absent) (string, []interface{}) {
	args := []interface{}{}
	args = append(args, m.Reason, m.Comment, m.StudentId, m.LessonId, m.CreatedBy, m.UpdatedBy)
	return sqlAbsentCreateOrUpdate, args
}

func (d *PgxStore) AbsentsUpdate(ctx context.Context, data models.Absent) (models.Absent, error) {
	// origModel := d.UsersFindById(strconv.Itoa(int(model.ID)))
	qs, args := AbsentsUpdateQuery(data)
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		_, err = tx.Query(ctx, qs, args...)
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return models.Absent{}, err
	}

	editModel, err := d.AbsentsFindById(ctx, data.ID)
	if err != nil {
		return models.Absent{}, err
	}
	if editModel.ID == "" {
		return models.Absent{}, errors.New("model not found: " + string(data.ID))
	}
	d.AbsentsUpdateRelations(data, editModel)
	return editModel, nil
}

func (d *PgxStore) AbsentsCreate(ctx context.Context, m models.Absent) (models.Absent, error) {
	qs, args := AbsentsCreateQuery(m)
	qs += " RETURNING uid"
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		err = tx.QueryRow(ctx, qs, args...).Scan(&m.ID)
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return models.Absent{}, err
	}

	editModel, err := d.AbsentsFindById(ctx, m.ID)
	if err != nil {
		return models.Absent{}, err
	}
	if editModel.ID == "" {
		return models.Absent{}, errors.New("model not found: " + string(m.ID))
	}
	d.AbsentsUpdateRelations(m, editModel)
	return editModel, nil
}

func (d *PgxStore) AbsentsDelete(ctx context.Context, l []*models.Absent) ([]*models.Absent, error) {
	lids := []string{}
	for _, i := range l {
		lids = append(lids, i.LessonId)
	}
	sids := []string{}
	for _, i := range l {
		sids = append(sids, i.StudentId)
	}
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		_, err = tx.Query(ctx, sqlAbsentDelete, lids, sids)
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}
	return l, nil
}

func (d *PgxStore) AbsentsUpdateRelations(data models.Absent, model models.Absent) error {

	return nil
}

func AbsentsCreateQuery(m models.Absent) (string, []interface{}) {
	args := []interface{}{}
	cols := ""
	vals := ""
	q := AbsentAtomicQuery(m, true)
	for k, v := range q {
		args = append(args, v)
		cols += ", " + k
		vals += ", $" + strconv.Itoa(len(args))
	}
	qs := sqlAbsentInsert + " (" + strings.Trim(cols, ", ") + ") VALUES (" + strings.Trim(vals, ", ") + ")"
	return qs, args
}

func AbsentsUpdateQuery(m models.Absent) (string, []interface{}) {
	args := []interface{}{}
	sets := ""
	q := AbsentAtomicQuery(m, false)
	for k, v := range q {
		args = append(args, v)
		sets += ", " + k + "=$" + strconv.Itoa(len(args))
	}
	args = append(args, m.ID)
	qs := strings.ReplaceAll(sqlAbsentUpdate, "set uid=uid", "set uid=uid "+sets+" ") + "where uid=$" + strconv.Itoa(len(args))
	return qs, args
}

func AbsentAtomicQuery(m models.Absent, isCreate bool) map[string]interface{} {
	q := map[string]interface{}{}
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

func AbsentsListQuery(f models.AbsentFilterRequest, args []interface{}) (string, []interface{}) {
	var wheres string = ""
	qs := sqlAbsentSelectMany

	if f.ID != nil && *f.ID != "" {
		args = append(args, *f.ID)
		wheres += " and ab.uid=$" + strconv.Itoa(len(args))
	}
	if f.SchoolId != nil && *f.SchoolId != "" {
		args = append(args, *f.SchoolId)
		wheres += " and ab.school_uid=$" + strconv.Itoa(len(args))
	}
	if f.LessonId != nil && *f.LessonId != "" {
		args = append(args, *f.LessonId)
		wheres += " and ab.lesson_uid=$" + strconv.Itoa(len(args))
	}
	if f.LessonIds != nil && len(*f.LessonIds) > 0 {
		args = append(args, *f.LessonIds)
		wheres += " and ab.lesson_uid = ANY($" + strconv.Itoa(len(args)) + "::uuid[])"
	}
	if f.StudentId != nil && *f.StudentId != "" {
		args = append(args, *f.StudentId)
		wheres += " and ab.student_uid=$" + strconv.Itoa(len(args))
	}
	if f.StudentIds != nil && len(*f.StudentIds) > 0 {
		args = append(args, *f.StudentIds)
		wheres += " and ab.student_uid = ANY($" + strconv.Itoa(len(args)) + "::uuid[])"
	}
	if f.SchoolIds != nil {
		args = append(args, *f.SchoolIds)
		wheres += " and ab.school_uid = ANY($" + strconv.Itoa(len(args)) + "::uuid[])"
	}
	wheres += " group by ab.uid "
	wheres += " order by uid desc"
	qs = strings.ReplaceAll(qs, "uid=uid", "uid=uid "+wheres+" ")
	return qs, args
}

// todo: check workings of absent relation lesson and apply others
func (d *PgxStore) AbsentsLoadRelations(ctx context.Context, l []*models.Absent) error {
	err := d.AbsentsLoadRelationsLessons(ctx, l)
	if err != nil {
		return err
	}
	return nil
}

func (d *PgxStore) AbsentsLoadRelationsLessons(ctx context.Context, l []*models.Absent) error {
	ids := []string{}
	for _, m := range l {
		ids = append(ids, m.ID)
	}
	if len(ids) < 1 {
		return nil
	}

	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		rs, err := tx.Query(ctx, sqlAbsentLesson, ids)
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
