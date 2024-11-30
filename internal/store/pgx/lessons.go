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
const sqlLessonFields = `l.uid, l.school_uid, l.subject_uid, l.book_uid, l.book_page, l.period_uid, l.period_key, l.date, l.hour_number, l.type_title, l.title, l.content, l.pro_title, l.pro_files, l.assignment_title, l.assignment_content, l.assignment_files, l.lesson_attributes, l.is_teacher_excused, l.updated_at, l.created_at`
const sqlLessonSelect = `select ` + sqlLessonFields + `  from lessons l where uid = ANY($1::uuid[])`
const sqlLessonSelectBySubject = `select ` + sqlLessonFields + `  from lessons l where subject_uid = ANY($1::uuid[])`
const sqlLessonSelectMany = `select ` + sqlLessonFields + `, 1 from lessons l 
	left join subjects s on (s.uid=l.subject_uid) where l.uid=l.uid limit $1 offset $2`
const sqlLessonInsert = `insert into lessons`
const sqlLessonUpdate = `update lessons l set uid=uid`
const sqlLessonDelete = `delete from lessons l where uid = ANY($1::uuid[])`

const sqlLessonLikeInsert = `insert into lesson_likes (lesson_uid, user_uid) VALUES ($1, $2)`

const sqlLessonLikedByUser = `select exists (select 1 from lesson_likes where lesson_uid = $1 and user_uid = $2)`

const sqlLessonUnlike = `DELETE FROM lesson_likes WHERE lesson_uid = $1 AND user_uid = $2`

// relations
const sqlLessonSchool = `select ` + sqlSchoolFields + `, l.uid from lessons l 
	right join schools s on (s.uid=l.school_uid)  where l.uid = ANY($1::uuid[])`
const sqlLessonParent = `select ` + sqlLessonFields + `, cp.uid from lessons lp 
	right join lessons l on cp.parent_uid=l.uid where cp.uid = ANY($1::uuid[])`
const sqlLessonTeacher = `select ` + sqlUserFields + `, l.uid from lessons l 
	right join users u on (u.uid=l.teacher_uid) where l.uid = ANY($1::uuid[])`
const sqlLessonStudent = `select ` + sqlUserFields + `, l.uid from lessons l 
	right join users u on (u.uid=l.student_uid) where l.uid = ANY($1::uuid[])`
const sqlLessonBook = `select ` + sqlBookFields + `, l.uid from lessons l
	right join books b on (b.uid=l.book_uid) where l.uid = ANY($1::uuid[])`
const sqlLessonSubject = `select ` + sqlSubjectFields + `, l.uid from lessons l
	right join subjects sb on (sb.uid=l.subject_uid) where l.uid = ANY($1::uuid[])`
const sqlLessonClassroom = `select ` + sqlClassroomFields + `, l.uid from lessons l
	right join subjects sb on (sb.uid=l.subject_uid) 
	right join classrooms c on (c.uid=sb.classroom_uid) where l.uid = ANY($1::uuid[])
`

// many to many
const sqlLessonStudents = `select ` + sqlUserFields + `, ul.type, ul.classroom_uid from user_classrooms uc 
	left join users u on ul.user_uid=u.uid where ul.classroom_uid = ANY($1::uuid[])`
const sqlLessonStudentsDelete = `delete from user_classrooms where classroom_uid=$1`
const sqlLessonStudentsInsert = `insert into user_classrooms (classroom_uid, user_uid, type) values ($1, $2, $3)`
const sqlLessonSubjects = `select ` + sqlSubjectFields + `, sb.classroom_uid from subjects sb where sb.classroom_uid = ANY($1::uuid[])`
const sqlLessonSubjectsDelete = `delete from user_classrooms where classroom_uid=$1`
const sqlLessonSubjectsInsert = `insert into user_classrooms (classroom_uid, user_uid, type) values ($1, $2, $3)`

func scanLesson(rows pgx.Row, m *models.Lesson, addColumns ...interface{}) (err error) {
	err = rows.Scan(parseColumnsForScan(m, addColumns...)...)
	return
}

func (d *PgxStore) LessonsFindByIds(ctx context.Context, ids []string) ([]models.Lesson, error) {
	l := []models.Lesson{}

	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		rows, err := tx.Query(ctx, sqlLessonSelect, (ids))
		for rows.Next() {
			m := models.Lesson{}
			err := scanLesson(rows, &m)
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

func (d *PgxStore) LessonsFindById(ctx context.Context, id string) (models.Lesson, error) {
	l, err := d.LessonsFindByIds(ctx, []string{id})
	if err != nil {
		return models.Lesson{}, err
	}
	if len(l) < 1 {
		err = pgx.ErrNoRows
		utils.LoggerDesc("Scan error").Error(err)
		return models.Lesson{}, err
	}
	return l[0], nil
}

func (d *PgxStore) LessonsFindBy(ctx context.Context, f models.LessonFilterRequest) ([]*models.Lesson, int, error) {
	if f.Limit == nil {
		f.Limit = new(int)
		*f.Limit = 12
	}
	if f.Offset == nil {
		f.Offset = new(int)
		*f.Offset = 0
	}
	args := []interface{}{f.Limit, f.Offset}
	qs, args := LessonsListBuildQuery(f, args)
	l := []*models.Lesson{}
	var total int
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		rows, err := tx.Query(ctx, qs, args...)
		for rows.Next() {
			sub := models.Lesson{}
			err = scanLesson(rows, &sub, &total)
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
func (d *PgxStore) LessonsUpdateBy(ctx context.Context, f models.LessonFilterRequest, sets map[string]interface{}) (int, error) {
	args := []interface{}{}
	qs, args := LessonsUpdateQueryByMap(f, sets)
	rowsAffected := int(0)
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		cmd, err := tx.Exec(ctx, qs, args...)
		rowsAffected = int(cmd.RowsAffected())
		return err
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return 0, err
	}
	return rowsAffected, nil
}

func (d *PgxStore) LessonsUpdate(ctx context.Context, data models.Lesson) (models.Lesson, error) {
	// origModel := d.UsersFindById(strconv.Itoa(int(model.ID)))
	qs, args := LessonsUpdateQuery(data)

	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		_, err = tx.Exec(ctx, qs, args...)
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return models.Lesson{}, err
	}

	return data, nil
}

func (d *PgxStore) LessonsCreate(ctx context.Context, m models.Lesson) (models.Lesson, error) {
	qs, args := LessonsCreateQuery(m)
	qs += " RETURNING uid"

	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		err = tx.QueryRow(ctx, qs, args...).Scan(&m.ID)
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return models.Lesson{}, err
	}

	editModel, err := d.LessonsFindById(ctx, m.ID)
	if err != nil {
		return models.Lesson{}, err
	}
	if editModel.ID == "" {
		return models.Lesson{}, errors.New("model not found: " + string(m.ID))
	}
	return editModel, nil
}

func (d *PgxStore) LessonsDelete(ctx context.Context, l []*models.Lesson) ([]*models.Lesson, error) {
	ids := []string{}
	for _, i := range l {
		ids = append(ids, i.ID)
	}
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		_, err = tx.Query(ctx, sqlLessonDelete, (ids))
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}
	return l, nil
}

func (d *PgxStore) LessonsLike(ctx context.Context, lessonID string, userID string) error {
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		_, err = tx.Exec(ctx, sqlLessonLikeInsert, lessonID, userID)
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return err
	}
	return nil
}

func (d *PgxStore) LessonsLikedByUser(ctx context.Context, lessonID string, userID string) (bool, error) {
	var liked bool
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) error {
		err := tx.QueryRow(ctx, sqlLessonLikedByUser, lessonID, userID).Scan(&liked)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				liked = false
				return nil
			}
			return err
		}
		return nil
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return false, err
	}
	return liked, nil
}

func (d *PgxStore) LessonsUnlike(ctx context.Context, lessonID string, userID string) error {
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) error {
		_, err := tx.Exec(ctx, sqlLessonUnlike, lessonID, userID)
		return err
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return err
	}
	return nil
}

func (d *PgxStore) LessonsUpdateBatch(ctx context.Context, l []models.Lesson) error {
	err := d.LessonsBatch(ctx, l, LessonsUpdateQuery)
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return err
	}
	return err
}

func (d *PgxStore) LessonsCreateBatch(ctx context.Context, l []models.Lesson) error {
	err := d.LessonsBatch(ctx, l, LessonsCreateQuery)
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return err
	}
	return err
}

func (d *PgxStore) LessonsBatch(ctx context.Context, l []models.Lesson, f func(models.Lesson) (string, []interface{})) error {
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

func (d *PgxStore) LessonsDeleteBatch(ctx context.Context, ids []string) error {
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		_, err = tx.Exec(ctx, sqlLessonDelete, ids)
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return err
	}
	return nil
}

func LessonsCreateQuery(m models.Lesson) (string, []interface{}) {
	args := []interface{}{}
	cols := ""
	vals := ""
	q := LessonAtomicQuery(m, true)
	for k, v := range q {
		args = append(args, v)
		cols += ", " + k
		vals += ", $" + strconv.Itoa(len(args))
	}
	qs := sqlLessonInsert + " (" + strings.Trim(cols, ", ") + ") VALUES (" + strings.Trim(vals, ", ") + ")"
	return qs, args
}

func LessonsUpdateQuery(m models.Lesson) (string, []interface{}) {
	args := []interface{}{}
	sets := ""
	q := LessonAtomicQuery(m, false)
	for k, v := range q {
		args = append(args, v)
		sets += ", " + k + "=$" + strconv.Itoa(len(args))
	}
	args = append(args, m.ID)
	qs := strings.ReplaceAll(sqlLessonUpdate, "set uid=uid", "set uid=uid "+sets+" ") + "where uid=$" + strconv.Itoa(len(args))
	return qs, args
}

func LessonsUpdateQueryByMap(f models.LessonFilterRequest, ss map[string]interface{}) (string, []interface{}) {
	args := []interface{}{}
	// sets
	sets := ""
	for k, v := range ss {
		args = append(args, v)
		sets += ", " + k + "=$" + strconv.Itoa(len(args))
	}
	//  wheres
	wheres, whereArgs := lessonsBuildWhereQuery(f, args)
	args = whereArgs
	qs := strings.ReplaceAll(sqlLessonUpdate, "set uid=uid", "set uid=uid "+sets) + " where uid=uid " + wheres
	return qs, args
}

func LessonAtomicQuery(m models.Lesson, isCreate bool) map[string]interface{} {
	q := map[string]interface{}{}
	if m.BookId != nil {
		q["book_uid"] = m.BookId
	}
	if m.BookPage != nil {
		q["book_page"] = m.BookPage
	}
	if m.HourNumber != nil {
		q["hour_number"] = m.HourNumber
	}
	if m.TypeTitle != nil {
		q["type_title"] = m.TypeTitle
	}
	if m.Title != nil {
		q["title"] = *m.Title
	}
	if m.Content != nil {
		q["content"] = *m.Content
	}
	if m.LessonAttributes != nil {
		q["lesson_attributes"] = *m.LessonAttributes
	}
	if m.ProTitle != nil {
		q["pro_title"] = m.ProTitle
	}
	if m.ProFiles != nil {
		q["pro_files"] = m.ProFiles
	}
	if m.AssignmentTitle != nil {
		q["assignment_title"] = m.AssignmentTitle
	}
	if m.AssignmentContent != nil {
		q["assignment_content"] = m.AssignmentContent
	}
	if m.AssignmentFiles != nil {
		q["assignment_files"] = m.AssignmentFiles
	}
	if m.IsTeacherExcused != nil {
		q["is_teacher_excused"] = *m.IsTeacherExcused
	}
	if m.PeriodKey != nil {
		q["period_key"] = *m.PeriodKey
	}
	if !m.Date.IsZero() {
		q["date"] = m.Date
	}
	q["updated_at"] = time.Now()
	if isCreate {
		if m.SchoolId != "" {
			q["school_uid"] = m.SchoolId
		}
		if m.SubjectId != "" {
			q["subject_uid"] = m.SubjectId
		}
		if m.PeriodId != nil {
			q["period_uid"] = *m.PeriodId
		}
		q["created_at"] = time.Now()
	}
	return q
}

func lessonsBuildWhereQuery(f models.LessonFilterRequest, args []interface{}) (string, []interface{}) {
	wheres := ""

	if f.ID != nil && *f.ID != "" {
		args = append(args, *f.ID)
		wheres += " and l.uid=$" + strconv.Itoa(len(args))
	}
	if f.SchoolId != nil {
		args = append(args, f.SchoolId)
		wheres += " and l.school_uid=$" + strconv.Itoa(len(args))
	}
	if f.SchoolIds != nil {
		args = append(args, *f.SchoolIds)
		wheres += " and l.school_uid = ANY($" + strconv.Itoa(len(args)) + "::uuid[])"
	}
	if f.SubjectId != nil {
		args = append(args, f.SubjectId)
		wheres += " and l.subject_uid=$" + strconv.Itoa(len(args))
	}
	if f.SubjectIds != nil && len(*f.SubjectIds) > 0 {
		args = append(args, *f.SubjectIds)
		wheres += " and l.subject_uid = ANY($" + strconv.Itoa(len(args)) + "::uuid[])"
	}
	if f.ClassroomId != nil {
		args = append(args, f.ClassroomId)
		wheres += " and s.classroom_uid=$" + strconv.Itoa(len(args))
	}
	if f.PeriodId != nil {
		args = append(args, f.PeriodId)
		wheres += " and l.period_uid=$" + strconv.Itoa(len(args))
	}
	if f.PeriodNumber != nil {
		args = append(args, f.PeriodNumber)
		wheres += " and l.period_key=$" + strconv.Itoa(len(args))
	}
	if f.Date != nil {
		args = append(args, f.Date)
		wheres += " and l.date = $" + strconv.Itoa(len(args))
	}
	if f.DateRange != nil && len(*f.DateRange) > 0 {
		if len(*f.DateRange) > 1 {
			args = append(args, (*f.DateRange)[0], (*f.DateRange)[1])
			wheres += " and date >= $" + strconv.Itoa(len(args)-1) + " and date <= $" + strconv.Itoa(len(args))
		} else {
			args = append(args, (*f.DateRange)[0])
			wheres += " and date >= $" + strconv.Itoa(len(args))
		}
	}
	if f.HourNumber != nil {
		args = append(args, f.HourNumber)
		wheres += " and hour_number = $" + strconv.Itoa(len(args))
	}
	if f.HourNumberRange != nil && len(*f.HourNumberRange) > 0 {
		if len(*f.HourNumberRange) > 1 {
			args = append(args, (*f.HourNumberRange)[0], (*f.HourNumberRange)[1])
			wheres += " and hour_number >= $" + strconv.Itoa(len(args)-1) + " and hour_number <= $" + strconv.Itoa(len(args))
		} else {
			args = append(args, (*f.HourNumberRange)[0])
			wheres += " and hour_number >= $" + strconv.Itoa(len(args))
		}
	}
	if f.IsTeacherExcused != nil {
		args = append(args, *f.IsTeacherExcused)
		wheres += " and is_teacher_excused = $" + strconv.Itoa(len(args))
	}
	if f.Search != nil {
		*f.Search = strings.ToLower(*f.Search)
		args = append(args, f.Search)
		wheres += " and (lower(title) like '%' || $" + strconv.Itoa(len(args)) + " || '%' or lower(type_title) like '%' || $" + strconv.Itoa(len(args)) +
			" || '%' or lower(content) like '%' || $" + strconv.Itoa(len(args)) + " || '%')"
	}
	return wheres, args
}

func LessonsListBuildQuery(f models.LessonFilterRequest, args []interface{}) (string, []interface{}) {
	wheres, args := lessonsBuildWhereQuery(f, args)
	wheres += " group by l.uid"

	if f.Sort != nil {
		dir := "desc"
		if strings.HasSuffix(*f.Sort, "~") {
			dir = "asc"
		}
		*f.Sort = strings.ReplaceAll(*f.Sort, "~", "")
		args = append(args, f.Sort)
		wheres += " order by $" + strconv.Itoa(len(args)) + " " + dir
	} else {
		// wheres += " order by l.date asc, l.hour_number desc"
	}
	qs := sqlLessonSelectMany
	qs = strings.ReplaceAll(qs, "l.uid=l.uid", "l.uid=l.uid "+wheres+" ")
	return qs, args
}

func (d *PgxStore) LessonsLoadRelations(ctx context.Context, l *[]*models.Lesson) error {
	var err error

	err = d.LessonsLoadSubject(ctx, l)
	if err != nil {
		return err
	}

	err = d.LessonsLoadClassroom(ctx, l)
	if err != nil {
		return err
	}

	err = d.LessonsLoadBook(ctx, l)
	if err != nil {
		return err
	}

	return nil
}

func (d *PgxStore) LessonsLoadSubject(ctx context.Context, l *[]*models.Lesson) error {
	ids := []string{}
	for _, m := range *l {
		ids = append(ids, m.ID)
	}
	if len(ids) < 1 {
		return nil
	}
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		rows, err := tx.Query(ctx, sqlLessonSubject, ids)
		for rows.Next() {
			sub := models.Subject{}
			pid := ""
			err = scanSubject(rows, &sub, &pid)
			if err != nil {
				utils.LoggerDesc("Scan error").Error(err)
				return err
			}
			for _, m := range *l {
				if (m.ID) == pid {
					if m.Subject != nil {
						sub.Classroom = m.Subject.Classroom
					}
					m.Subject = &sub
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

func (d *PgxStore) LessonsLoadClassroom(ctx context.Context, l *[]*models.Lesson) error {
	ids := []string{}
	for _, m := range *l {
		ids = append(ids, m.ID)
	}
	if len(ids) < 1 {
		return nil
	}
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		rows, err := tx.Query(ctx, sqlLessonClassroom, ids)
		for rows.Next() {
			sub := models.Classroom{}
			pid := ""
			err = scanClassroom(rows, &sub, &pid)
			if err != nil {
				utils.LoggerDesc("Scan error").Error(err)
				return err
			}
			for _, m := range *l {
				if (m.ID) == pid {
					if m.Subject == nil {
						m.Subject = &models.Subject{}
					}
					m.Subject.Classroom = &sub
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

func (d *PgxStore) LessonsLoadBook(ctx context.Context, l *[]*models.Lesson) error {
	ids := []string{}
	for _, m := range *l {
		ids = append(ids, m.ID)
	}
	if len(ids) < 1 {
		return nil
	}
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		rows, err := tx.Query(ctx, sqlLessonBook, ids)
		for rows.Next() {
			sub := models.Book{}
			pid := ""
			err = scanBook(rows, &sub, &pid)
			if err != nil {
				utils.LoggerDesc("Scan error").Error(err)
				return err
			}
			for _, m := range *l {
				if (m.ID) == pid {
					m.Book = &sub
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
