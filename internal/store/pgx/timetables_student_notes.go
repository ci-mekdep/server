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
const sqlStudentNoteFields = `sn.uid, sn.school_uid, sn.subject_uid, sn.student_uid, sn.teacher_uid, sn.note, sn.updated_at, sn.created_at`
const sqlStudentNoteSelect = `select ` + sqlStudentNoteFields + ` from student_notes sn where uid = ANY($1::uuid[])`
const sqlStudentNoteSelectMany = `select ` + sqlStudentNoteFields + `, count(*) over() as total from student_notes sn where uid=uid limit $1 offset $2`
const sqlStudentNoteInsert = `insert into student_notes`
const sqlStudentNoteUpdate = `update student_notes sn set uid=uid`
const sqlStudentNoteDelete = `delete from student_notes sn where uid = ANY($1::uuid[])`

// relations
const sqlStudentNoteSchool = `select ` + sqlSchoolFields + `, sn.uid from student_notes sn 
	right join schools s on (s.uid=sn.school_uid) where sn.uid = ANY($1::uuid[])`
const sqlStudentNoteSubject = `select ` + sqlSubjectFields + `, sn.uid from student_notes sn 
	right join subjects sb on (sb.uid=sn.subject_uid)  where sn.uid = ANY($1::uuid[])`
const sqlStudentNoteStudent = `select ` + sqlUserFields + `, sn.uid from student_notes sn 
	right join users u on (u.uid=sn.student_uid) where sn.uid = ANY($1::uuid[])`
const sqlStudentNoteTeacher = `select ` + sqlUserFields + `, sn.uid from student_notes sn 
	right join users u on (u.uid=sn.teacher_uid) where sn.uid = ANY($1::uuid[])`

func scanStudentNote(rows pgx.Row, m *models.StudentNote, addColumns ...interface{}) (err error) {
	err = rows.Scan(parseColumnsForScan(m, addColumns...)...)
	return
}

func (d *PgxStore) StudentNotesFindOrCreate(ctx context.Context, data *models.StudentNote) (models.StudentNote, error) {
	if data.StudentId == "" {
		return models.StudentNote{}, errors.New("data student uid is nil")
	}
	// try to fetch grade
	qs := sqlStudentNoteSelectMany
	qs = strings.ReplaceAll(qs, "uid=uid", "sn.subject_uid=$3 and sn.student_uid=$4")
	m := models.StudentNote{}
	total := 0
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		err = scanStudentNote(
			tx.QueryRow(ctx, qs, 1, 0, data.SubjectId, data.StudentId),
			&m,
			&total)
		return
	})
	if err != nil {
		if !strings.Contains(err.Error(), "no rows") {
			utils.LoggerDesc("Query error").Error(err)
			return models.StudentNote{}, err
		}
		// grade not exists, create
		qs, args := StudentNoteCreateQuery(data)
		qs += " RETURNING " + strings.ReplaceAll(sqlStudentNoteFields, "sn.", "")
		err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
			err = scanStudentNote(tx.QueryRow(ctx, qs, args...), &m)
			return
		})
		if err != nil {
			utils.LoggerDesc("Scan error").Error(err)
			return models.StudentNote{}, err
		}
	}
	return m, nil
}
func (d *PgxStore) StudentNotesUpdateOrCreate(ctx context.Context, data *models.StudentNote) (models.StudentNote, error) {
	m, err := d.StudentNotesFindOrCreate(ctx, data)
	if err != nil {
		return models.StudentNote{}, err
	}
	data.ID = m.ID
	return d.StudentNoteUpdate(ctx, data)
}

func (d *PgxStore) StudentNotesFindByIds(ctx context.Context, ids []string) ([]*models.StudentNote, error) {
	l := []*models.StudentNote{}
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		rows, err := tx.Query(ctx, sqlStudentNoteSelect, (ids))
		for rows.Next() {
			m := models.StudentNote{}
			err := scanStudentNote(rows, &m)
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

func (d *PgxStore) StudentNoteFindById(ctx context.Context, id string) (*models.StudentNote, error) {
	l, err := d.StudentNotesFindByIds(ctx, []string{id})
	if err != nil {
		return nil, err
	}
	if len(l) < 1 {
		return nil, errors.New("timetable not found by uid: " + id)
	}
	return l[0], nil
}

func (d *PgxStore) StudentNotesFindBy(ctx context.Context, f models.StudentNoteFilterRequest) ([]*models.StudentNote, int, error) {
	args := []interface{}{1000, 0}
	qs, args := StudentNotesListBuildQuery(f, args)

	l := []*models.StudentNote{}
	var total int
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		rows, err := tx.Query(ctx, qs, args...)
		for rows.Next() {
			sub := models.StudentNote{}
			err = scanStudentNote(rows, &sub, &total)
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

func (d *PgxStore) StudentNoteUpdate(ctx context.Context, data *models.StudentNote) (models.StudentNote, error) {
	qs, args := StudentNoteUpdateQuery(data)
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		_, err = tx.Query(ctx, qs, args...)
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return models.StudentNote{}, err
	}

	editModel, err := d.StudentNoteFindById(ctx, data.ID)
	if err != nil {
		return models.StudentNote{}, err
	}
	d.StudentNoteUpdateRelations(ctx, data, editModel)
	return *editModel, nil
}

func (d *PgxStore) StudentNoteCreate(ctx context.Context, m *models.StudentNote) (*models.StudentNote, error) {
	qs, args := StudentNoteCreateQuery(m)
	qs += " RETURNING uid"
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		err = tx.QueryRow(ctx, qs, args...).Scan(&m.ID)
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}

	editModel, err := d.StudentNoteFindById(ctx, m.ID)
	if err != nil {
		return nil, err
	}
	d.StudentNoteUpdateRelations(ctx, m, editModel)
	return editModel, nil
}

func (d *PgxStore) StudentNotesDelete(ctx context.Context, l []*models.StudentNote) ([]*models.StudentNote, error) {
	ids := []string{}
	for _, i := range l {
		ids = append(ids, i.ID)
	}
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		_, err = tx.Query(ctx, sqlStudentNoteDelete, (ids))
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}
	return l, nil
}

func (d *PgxStore) StudentNoteUpdateRelations(ctx context.Context, data *models.StudentNote, model *models.StudentNote) {

}

func StudentNoteCreateQuery(m *models.StudentNote) (string, []interface{}) {
	args := []interface{}{}
	cols := ""
	vals := ""
	q := StudentNoteAtomicQuery(m, true)
	for k, v := range q {
		args = append(args, v)
		cols += ", " + k
		vals += ", $" + strconv.Itoa(len(args))
	}
	qs := sqlStudentNoteInsert + " (" + strings.Trim(cols, ", ") + ") VALUES (" + strings.Trim(vals, ", ") + ")"
	return qs, args
}

func StudentNoteUpdateQuery(m *models.StudentNote) (string, []interface{}) {
	args := []interface{}{}
	sets := ""
	q := StudentNoteAtomicQuery(m, false)
	for k, v := range q {
		args = append(args, v)
		sets += ", " + k + "=$" + strconv.Itoa(len(args))
	}
	args = append(args, m.ID)
	qs := strings.ReplaceAll(sqlStudentNoteUpdate, "set uid=uid", "set uid=uid "+sets+" ") + "where uid=$" + strconv.Itoa(len(args))
	return qs, args
}

func StudentNoteAtomicQuery(m *models.StudentNote, isCreate bool) map[string]interface{} {
	q := map[string]interface{}{}
	if m.SchoolId != "" {
		q["school_uid"] = m.SchoolId
	}
	if m.SubjectId != nil {
		q["subject_uid"] = *m.SubjectId
	}
	if m.StudentId != "" {
		q["student_uid"] = m.StudentId
	}
	q["teacher_uid"] = m.TeacherId
	q["note"] = m.Note
	if isCreate {
		q["created_at"] = time.Now()
	}
	q["updated_at"] = time.Now()
	return q
}

func StudentNotesListBuildQuery(f models.StudentNoteFilterRequest, args []interface{}) (string, []interface{}) {
	var wheres string = ""

	if f.StudentIds != nil && len(*f.StudentIds) > 0 {
		args = append(args, *f.StudentIds)
		wheres += " and (student_uid = ANY($" + strconv.Itoa(len(args)) + ") )"
	}
	if f.SchoolId != nil && *f.SchoolId != "" {
		args = append(args, *f.SchoolId)
		wheres += " and school_uid=$" + strconv.Itoa(len(args))
	}
	if f.SubjectId != nil && *f.SubjectId != "" {
		args = append(args, *f.SubjectId)
		wheres += " and subject_uid=$" + strconv.Itoa(len(args))
	}
	if f.TeacherId != nil && *f.TeacherId != "" {
		args = append(args, *f.TeacherId)
		wheres += " and teacher_uid=$" + strconv.Itoa(len(args))
	}
	if f.StudentId != nil && *f.StudentId != "" {
		args = append(args, *f.StudentId)
		wheres += " and student_uid=$" + strconv.Itoa(len(args))
	}
	wheres += " order by uid desc"
	qs := sqlStudentNoteSelectMany
	qs = strings.ReplaceAll(qs, "uid=uid", "uid=uid "+wheres+" ")
	return qs, args
}

func (d *PgxStore) StudentNotesLoadRelations(ctx context.Context, l *[]*models.StudentNote) error {
	ids := []string{}
	for _, m := range *l {
		ids = append(ids, m.ID)
	}
	if len(ids) < 1 {
		return nil
	}
	// load school
	if rs, err := d.StudentNotesLoadSchool(ctx, ids); err == nil {
		for _, r := range rs {
			for _, m := range *l {
				if r.ID == m.ID {
					m.School = r.Relation
				}
			}
		}
	} else {
		return nil
	}
	// load subject
	if rs, err := d.StudentNotesLoadSubject(ctx, ids); err == nil {
		for _, r := range rs {
			for _, m := range *l {
				if r.ID == m.ID {
					m.Subject = r.Relation
				}
			}
		}
	} else {
		return nil
	}
	// load teacher
	if rs, err := d.StudentNotesLoadTeacher(ctx, ids); err == nil {
		for _, r := range rs {
			for _, m := range *l {
				if r.ID == m.ID {
					m.Teacher = r.Relation
				}
			}
		}
	} else {
		return nil
	}
	// load student
	if rs, err := d.StudentNotesLoadStudent(ctx, ids); err == nil {
		for _, r := range rs {
			for _, m := range *l {
				if r.ID == m.ID {
					m.Student = r.Relation
				}
			}
		}
	} else {
		return nil
	}
	return nil
}

type StudentNotesLoadSchoolItem struct {
	ID       string
	Relation *models.School
}

func (d *PgxStore) StudentNotesLoadSchool(ctx context.Context, ids []string) ([]StudentNotesLoadSchoolItem, error) {
	res := []StudentNotesLoadSchoolItem{}
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		rows, err := tx.Query(ctx, sqlStudentNoteSchool, (ids))
		for rows.Next() {
			sub := models.School{}
			pid := ""
			err = scanSchool(rows, &sub, &pid)
			if err != nil {
				return err
			}
			res = append(res, StudentNotesLoadSchoolItem{ID: pid, Relation: &sub})
		}
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}

	return res, nil
}

type StudentNotesLoadSubjectItem struct {
	ID       string
	Relation *models.Subject
}

func (d *PgxStore) StudentNotesLoadSubject(ctx context.Context, ids []string) ([]StudentNotesLoadSubjectItem, error) {
	res := []StudentNotesLoadSubjectItem{}
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		rows, err := tx.Query(ctx, sqlStudentNoteSubject, (ids))
		for rows.Next() {
			sub := models.Subject{}
			pid := ""
			err = scanSubject(rows, &sub, &pid)
			if err != nil {
				return err
			}
			res = append(res, StudentNotesLoadSubjectItem{ID: pid, Relation: &sub})
		}
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}

	return res, nil
}

type StudentNotesLoadTeacherItem struct {
	ID       string
	Relation *models.User
}

func (d *PgxStore) StudentNotesLoadTeacher(ctx context.Context, ids []string) ([]StudentNotesLoadTeacherItem, error) {
	res := []StudentNotesLoadTeacherItem{}
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		rows, err := tx.Query(ctx, sqlStudentNoteTeacher, (ids))
		for rows.Next() {
			sub := models.User{}
			pid := ""
			err = scanUser(rows, &sub, &pid)
			if err != nil {
				return err
			}
			res = append(res, StudentNotesLoadTeacherItem{ID: pid, Relation: &sub})
		}

		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}
	return res, nil
}

type StudentNotesLoadStudentItem struct {
	ID       string
	Relation *models.User
}

func (d *PgxStore) StudentNotesLoadStudent(ctx context.Context, ids []string) ([]StudentNotesLoadStudentItem, error) {
	res := []StudentNotesLoadStudentItem{}
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		rows, err := tx.Query(ctx, sqlStudentNoteStudent, (ids))
		for rows.Next() {
			sub := models.User{}
			pid := ""
			err = scanUser(rows, &sub, &pid)
			if err != nil {
				return err
			}
			res = append(res, StudentNotesLoadStudentItem{ID: pid, Relation: &sub})
		}
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}

	return res, nil
}
