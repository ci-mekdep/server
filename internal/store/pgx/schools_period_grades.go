package pgx

import (
	"context"
	"errors"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/mekdep/server/internal/models"
	"github.com/mekdep/server/internal/utils"
)

// base
const sqlPeriodGradeFields = `pg.uid, pg.period_uid, pg.period_key, pg.subject_uid, pg.student_uid, pg.exam_uid, pg.lesson_count, pg.absent_count, pg.grade_count, pg.grade_sum, pg.old_absent_count, pg.old_grade_count, pg.old_grade_sum, pg.prev_grade_count, pg.prev_grade_sum, pg.updated_at, pg.created_at`
const sqlPeriodGradeSelect = `select ` + sqlPeriodGradeFields + `  from period_grades pg where uid = ANY($1::uuid[])`
const sqlPeriodGradeSelectMany = `select ` + sqlPeriodGradeFields + `, 1 as total from period_grades pg where uid=uid limit $1 offset $2`
const sqlPeriodGradeInsert = `insert into period_grades`
const sqlPeriodGradeUpdate = `update period_grades pg set uid=uid`
const sqlPeriodGradeDelete = `delete from period_grades pg where uid = ANY($1::uuid[])`

const sqlPeriodGradeBatchUpsert = `INSERT INTO period_grades 
(subject_uid, student_uid, period_key, lesson_count, absent_count, grade_count, grade_sum, updated_at, period_uid, exam_uid, created_at)
SELECT 
    $1::uuid,       -- subject_uid
    $2::uuid,       -- student_uid
    $3::int,        -- period_key
    COUNT(DISTINCT l.uid) AS lesson_count,
    COUNT(DISTINCT a.uid) AS absent_count,
    COUNT(DISTINCT g.uid) AS grade_count,
    SUM(COALESCE(g.value::int, 0)) AS grade_sum,
    NOW() AS updated_at,
    null::uuid AS period_uid,
    null::uuid AS exam_uid,
    NOW() AS created_at
FROM lessons l 
LEFT JOIN (
    SELECT
        uid,
        lesson_uid,
        student_uid,
        (COALESCE(value::int, 0) + (COALESCE(values[1]::int, 0) + COALESCE(values[2]::int, 0)) / 2) AS value
    FROM grades
) g ON (g.lesson_uid = l.uid AND g.student_uid = $2)
LEFT JOIN absents a ON (a.lesson_uid = l.uid AND a.student_uid = $2)
WHERE l.period_key = $3 AND l.subject_uid = $1
GROUP BY subject_uid, period_key

ON CONFLICT (subject_uid, student_uid, period_key)
DO UPDATE SET 
    lesson_count = EXCLUDED.lesson_count, 
    absent_count = EXCLUDED.absent_count, 
    grade_count = EXCLUDED.grade_count, 
    grade_sum = EXCLUDED.grade_sum,  
    updated_at = NOW();`

const sqlPeriodGradeStudent = `select ` + sqlUserFields + `, pg.uid from period_grades pg 
	right join users u on (u.uid=pg.student_uid) where pg.uid = ANY($1::uuid[])`
const sqlPeriodGradeByStudent = `select DISTINCT ON (subject_uid, period_uid, period_key) ` + sqlPeriodGradeFields + ` from period_grades pg where pg.student_uid = $1`
const sqlDeletePeriodGradeByStudentAndSubjects = `delete from period_grades where student_uid = $1 AND subject_uid = ANY($2::uuid[])`

func scanPeriodGrade(rows pgx.Row, m *models.PeriodGrade, addColumns ...interface{}) (err error) {
	err = rows.Scan(parseColumnsForScan(m, addColumns...)...)
	return
}

func (d *PgxStore) PeriodGradesFindOrCreate(ctx context.Context, data *models.PeriodGrade) (models.PeriodGrade, error) {
	if data.StudentId == nil {
		return models.PeriodGrade{}, errors.New("data student uid is nil")
	}
	// try to fetch grade
	qs := sqlPeriodGradeSelectMany
	qs = strings.ReplaceAll(qs, "uid=uid", " pg.period_key=$3 and pg.subject_uid=$4 and pg.student_uid=$5 and pg.exam_uid=$6")
	m := models.PeriodGrade{}
	total := 0
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		err = scanPeriodGrade(
			tx.QueryRow(ctx, qs, 1, 0, data.PeriodKey, data.SubjectId, data.StudentId, data.ExamId),
			&m,
			&total)
		return
	})
	if err != nil {
		if !strings.Contains(err.Error(), "no rows") {
			utils.LoggerDesc("Query error").Error(err)
			return models.PeriodGrade{}, err
		}
		// grade not exists, create
		qs, args := PeriodGradesCreateQuery(data)
		qs += " RETURNING " + strings.ReplaceAll(sqlPeriodGradeFields, "pg.", "")
		err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
			err = scanPeriodGrade(tx.QueryRow(ctx, qs, args...), &m)
			return
		})
		if err != nil {
			utils.LoggerDesc("Query error").Error(err)
			return models.PeriodGrade{}, err
		}
	}
	return m, nil
}
func (d *PgxStore) PeriodGradesUpdateOrCreate(ctx context.Context, data *models.PeriodGrade) (models.PeriodGrade, error) {
	m, err := d.PeriodGradesFindOrCreate(ctx, data)
	if err != nil {
		return models.PeriodGrade{}, err
	}
	data.ID = m.ID
	return d.PeriodGradesUpdate(ctx, data)
}

func (d *PgxStore) PeriodGradesUpdateValues(ctx context.Context, data models.PeriodGrade) (*models.PeriodGrade, error) {
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) error {
		qs, args := PeriodGradesUpdateQuery(&data)
		_, err := tx.Exec(ctx, qs, args...)
		return err
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}
	log.Println(data)
	return &data, nil
}

func (d *PgxStore) PeriodGradesUpdateBatch(ctx context.Context, l []models.PeriodGrade) error {
	return d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		sqls := pgx.Batch{}
		for _, m := range l {
			qs, args := PeriodGradesUpdateQuery(&m)
			sqls.Queue(qs, args...)
		}

		br := tx.SendBatch(ctx, &sqls)
		for range l {
			_, err := br.Exec()
			if err != nil {
				utils.LoggerDesc("Query error").Error(err)
				return err
			}
		}
		log.Println("Rows affected: " + strconv.Itoa(sqls.Len()))
		return err
	})
	return nil
}

func (d *PgxStore) PeriodGradesBatch(ctx context.Context, l []models.PeriodGrade, f func(models.PeriodGrade) (string, []interface{})) error {
	if len(l) <= 0 {
		return nil
	}

	return d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		sqls := pgx.Batch{}
		for _, m := range l {
			qs, args := f(m)
			sqls.Queue(qs, args...)
		}

		br := tx.SendBatch(ctx, &sqls)
		for range l {
			_, err := br.Exec()
			if err != nil {
				return err
			}
		}
		log.Println("Rows affected: " + strconv.Itoa(sqls.Len()))
		return err
	})
}

func (d *PgxStore) PeriodGradesFindByIds(ctx context.Context, ids []string) ([]models.PeriodGrade, error) {
	l := []models.PeriodGrade{}
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		rows, err := tx.Query(ctx, sqlPeriodGradeSelect, (ids))
		for rows.Next() {
			m := models.PeriodGrade{}
			err := scanPeriodGrade(rows, &m)
			if err != nil {
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

func (d *PgxStore) PeriodGradesFindById(ctx context.Context, id string) (models.PeriodGrade, error) {
	l, err := d.PeriodGradesFindByIds(ctx, []string{id})
	if err != nil {
		return models.PeriodGrade{}, err
	}
	if len(l) < 1 {
		return models.PeriodGrade{}, errors.New("grade not found by uid: " + id)
	}
	return l[0], nil
}

func (d *PgxStore) PeriodGradesFindBy(ctx context.Context, f models.PeriodGradeFilterRequest) ([]*models.PeriodGrade, int, error) {
	args := []interface{}{1000, 0}
	var qs string
	qs, args = PeriodGradesListBuildQuery(f, args)

	l := []*models.PeriodGrade{}
	var total int
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		rows, err := tx.Query(ctx, qs, args...)
		for rows.Next() {
			sub := models.PeriodGrade{}
			err = scanPeriodGrade(rows, &sub, &total)
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

func (d *PgxStore) PeriodGradesUpdate(ctx context.Context, data *models.PeriodGrade) (models.PeriodGrade, error) {
	// origModel := d.UsersFindById(strconv.Itoa(int(model.ID)))
	qs, args := PeriodGradesUpdateQuery(data)
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		_, err = tx.Query(ctx, qs, args...)
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return models.PeriodGrade{}, err
	}

	editModel, err := d.PeriodGradesFindById(ctx, data.ID)
	if err != nil {
		return models.PeriodGrade{}, err
	}
	if editModel.ID == "" {
		return models.PeriodGrade{}, errors.New("model not found: " + string(data.ID))
	}
	return editModel, nil
}

func (d *PgxStore) PeriodGradesCreate(ctx context.Context, m *models.PeriodGrade) (models.PeriodGrade, error) {
	qs, args := PeriodGradesCreateQuery(m)
	qs += " RETURNING uid"
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		err = tx.QueryRow(ctx, qs, args...).Scan(&m.ID)
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return models.PeriodGrade{}, err
	}

	editModel, err := d.PeriodGradesFindById(ctx, m.ID)
	if err != nil {
		return models.PeriodGrade{}, err
	}
	if editModel.ID == "" {
		return models.PeriodGrade{}, errors.New("model not found: " + string(m.ID))
	}
	return editModel, nil
}

func (d *PgxStore) PeriodGradesDelete(ctx context.Context, l []*models.PeriodGrade) ([]*models.PeriodGrade, error) {
	ids := []string{}
	for _, i := range l {
		ids = append(ids, i.ID)
	}
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		_, err = tx.Query(ctx, sqlPeriodGradeDelete, (ids))
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}
	return l, nil
}

func PeriodGradesCreateQuery(m *models.PeriodGrade) (string, []interface{}) {
	args := []interface{}{}
	cols := ""
	vals := ""
	q := PeriodGradeAtomicQuery(m, true)
	for k, v := range q {
		args = append(args, v)
		cols += ", " + k
		vals += ", $" + strconv.Itoa(len(args))
	}
	qs := sqlPeriodGradeInsert + " (" + strings.Trim(cols, ", ") + ") VALUES (" + strings.Trim(vals, ", ") + ")"
	return qs, args
}

func PeriodGradesUpdateQuery(m *models.PeriodGrade) (string, []interface{}) {
	args := []interface{}{m.SubjectId, m.StudentId, m.PeriodKey}
	return sqlPeriodGradeBatchUpsert, args
}

func PeriodGradeAtomicQuery(m *models.PeriodGrade, isCreate bool) map[string]interface{} {
	q := map[string]interface{}{}
	if m.PeriodId != nil {
		q["period_uid"] = m.PeriodId
	}
	if m.PeriodKey != 0 {
		q["period_key"] = m.PeriodKey
	}
	if m.SubjectId != nil {
		q["subject_uid"] = m.SubjectId
	}
	if m.StudentId != nil {
		q["student_uid"] = m.StudentId
	}
	if m.ExamId != nil && *m.ExamId != "" {
		q["exam_uid"] = m.ExamId
	}
	if m.LessonCount != 0 {
		q["lesson_count"] = m.LessonCount
	}
	if m.AbsentCount != 0 {
		q["absent_count"] = m.AbsentCount
	}
	if m.GradeCount != 0 {
		q["grade_count"] = m.GradeCount
	}
	if m.GradeSum != 0 {
		q["grade_sum"] = m.GradeSum
	}
	if m.OldAbsentCount != 0 {
		q["old_absent_count"] = m.OldAbsentCount
	}
	if m.OldGradeCount != 0 {
		q["old_grade_count"] = m.OldGradeCount
	}
	if m.OldGradeSum != 0 {
		q["old_grade_sum"] = m.OldGradeSum
	}
	if isCreate {
		q["created_at"] = time.Now()
	}
	q["updated_at"] = time.Now()
	return q
}

func PeriodGradesListBuildQuery(f models.PeriodGradeFilterRequest, args []interface{}) (string, []interface{}) {
	var wheres string = ""

	if f.SubjectId != nil {
		args = append(args, f.SubjectId)
		wheres += " and subject_uid=$" + strconv.Itoa(len(args)) + "::uuid"
	}
	if f.StudentId != nil {
		args = append(args, f.StudentId)
		wheres += " and student_uid=$" + strconv.Itoa(len(args)) + "::uuid"
	}
	if f.PeriodId != nil {
		args = append(args, f.PeriodId)
		wheres += " and period_uid=$" + strconv.Itoa(len(args)) + "::uuid"
	}
	if f.PeriodKey != nil {
		args = append(args, f.PeriodKey)
		wheres += " and period_key=$" + strconv.Itoa(len(args))
	}
	if f.PeriodKeys != nil && len(*f.PeriodKeys) > 0 {
		args = append(args, f.PeriodKeys)
		wheres += " and pg.period_key = ANY($" + strconv.Itoa(len(args)) + "::int[])"
	}
	if f.StudentIds != nil && len(*f.StudentIds) > 0 {
		args = append(args, f.StudentIds)
		wheres += " and pg.student_uid = ANY($" + strconv.Itoa(len(args)) + "::uuid[])"
	}
	if f.SubjectIds != nil && len(*f.SubjectIds) > 0 {
		args = append(args, f.SubjectIds)
		wheres += " and pg.subject_uid = ANY($" + strconv.Itoa(len(args)) + "::uuid[])"
	}
	if f.ExamId != nil {
		args = append(args, *f.ExamId)
		wheres += " and pg.exam_uid=$" + strconv.Itoa(len(args)) + "::uuid"
	}
	if f.ExamIds != nil && len(f.ExamIds) > 0 {
		args = append(args, f.ExamIds)
		wheres += " and pg.exam_uids = ANY($" + strconv.Itoa(len(args)) + "::uuid[])"
	}
	wheres += " group by pg.uid "
	wheres += " order by pg.created_at asc"
	qs := sqlPeriodGradeSelectMany
	qs = strings.ReplaceAll(qs, "uid=uid", "uid=uid "+wheres+" ")
	return qs, args
}

// todo: check workings of grade relation lesson and apply others
func (d *PgxStore) PeriodGradesLoadRelations(ctx context.Context, l *[]*models.PeriodGrade) error {
	ids := []string{}
	for _, m := range *l {
		ids = append(ids, m.ID)
	}
	if len(ids) < 1 {
		return nil
	}
	// load admin
	if rs, err := d.PeriodGradesLoadStudent(ctx, ids); err == nil {
		for _, r := range rs {
			for _, m := range *l {
				if r.UID == m.ID {
					m.Student = r.Relation
				}
			}
		}
	} else {
		return err
	}
	return nil
}

type PeriodGradesLoadStudentItem struct {
	UID      string
	Relation *models.User
}

func (d *PgxStore) PeriodGradesLoadStudent(ctx context.Context, ids []string) ([]PeriodGradesLoadStudentItem, error) {
	res := []PeriodGradesLoadStudentItem{}
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		rows, err := tx.Query(ctx, sqlPeriodGradeStudent, (ids))
		for rows.Next() {
			sub := models.User{}
			pid := string("")
			err = scanUser(rows, &sub, &pid)
			if err != nil {
				return err
			}
			res = append(res, PeriodGradesLoadStudentItem{UID: pid, Relation: &sub})
		}
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}

	return res, nil
}

func (d *PgxStore) PeriodGradeByStudent(ctx context.Context, student_id string) ([]*models.PeriodGrade, error) {
	res := []*models.PeriodGrade{}
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		rows, err := tx.Query(ctx, sqlPeriodGradeByStudent, (student_id))
		for rows.Next() {
			sub := models.PeriodGrade{}
			err = scanPeriodGrade(rows, &sub)
			if err != nil {
				return err
			}
			res = append(res, &sub)
		}
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}

	return res, nil
}

func (d *PgxStore) DeletePeriodGradeByStudentAndSubjects(ctx context.Context, student_id string, subjectIds []string) error {
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		_, err = tx.Exec(ctx, sqlDeletePeriodGradeByStudentAndSubjects, student_id, subjectIds)
		return
	})
	if err != nil {
		utils.LoggerDesc("Delete error").Error(err)
		return err
	}
	return nil
}
