package pgx

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/mekdep/server/internal/models"
	"github.com/mekdep/server/internal/utils"
)

const sqlSubjectExamFields = `se.uid, se.subject_uid, se.school_uid, se.classroom_uid, se.teacher_uid, se.head_teacher_uid, se.member_teacher_uids, se.room_number, se.start_time, se.time_length_min, se.exam_weight_percent, se.name, se.is_required, se.created_at, se.updated_at`
const sqlSubjectExamSelect = `select ` + sqlSubjectExamFields + ` from subject_exams se where uid = ANY($1::uuid[])`
const sqlSubjectExamSelectMany = `select ` + sqlSubjectExamFields + `, count(*) over() as total from subject_exams se
	left join subjects s on s.uid=se.subject_uid
	left join classrooms c on c.uid=s.classroom_uid
	where se.uid=se.uid limit $1 offset $2`
const sqlSubjectExamInsert = `insert into subject_exams`
const sqlSubjectExamUpdate = `update subject_exams se set uid=uid`
const sqlSubjectExamDelete = `delete from subject_exams se where se.uid = ANY($1::uuid[])`

// relations
const sqlSubjectExamSubject = `select ` + sqlSubjectFields + `, se.uid from subject_exams se
	right join subjects sb on (sb.uid=se.subject_uid) where se.uid = ANY($1::uuid[])`

const sqlSubjectExamSchool = `select ` + sqlSchoolFields + `, se.uid from subject_exams se
	right join schools s on (s.uid=se.school_uid) where se.uid = ANY($1::uuid[])`

const sqlSubjectExamTeacher = `select ` + sqlUserFields + `, se.uid from subject_exams se
	right join users u on (u.uid=se.teacher_uid) where se.uid = ANY($1::uuid[])`

const sqlSubjectExamHeadTeacher = `select ` + sqlUserFields + `, se.uid from subject_exams se
	right join users u on (u.uid=se.head_teacher_uid) where se.uid = ANY($1::uuid[])`

const sqlSubjectExamClassroom = `select ` + sqlClassroomFields + `, se.uid from subject_exams se
	right join classrooms c on (c.uid=se.classroom_uid) where se.uid = ANY($1::uuid[])`

const sqlSubjectExamMemberTeachers = `select ` + sqlUserFields + `, se.uid from subject_exams se
	right join users u on u.uid = ANY(se.member_teacher_uids::uuid[]) where se.uid = ANY($1::uuid[])`

func scanSubjectExam(rows pgx.Row, m *models.SubjectExam, addColumns ...interface{}) (err error) {
	err = rows.Scan(parseColumnsForScan(m, addColumns...)...)
	return
}

func (d *PgxStore) SubjectExamFindByIds(ctx context.Context, ids []string) ([]*models.SubjectExam, error) {
	l := []*models.SubjectExam{}
	err := d.runQuery(context.Background(), func(tx *pgxpool.Conn) (err error) {
		rows, err := tx.Query(context.Background(), sqlSubjectExamSelect, (ids))
		for rows.Next() {
			m := models.SubjectExam{}
			err := scanSubjectExam(rows, &m)
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

func (d *PgxStore) SubjectExamFindById(ctx context.Context, id string) (*models.SubjectExam, error) {
	l, err := d.SubjectExamFindByIds(ctx, []string{id})
	if err != nil {
		return nil, err
	}
	if len(l) < 1 {
		return nil, pgx.ErrNoRows
	}
	return l[0], nil
}

func (d *PgxStore) SubjectExamsFindBy(ctx context.Context, f *models.SubjectExamFilterRequest) (l []*models.SubjectExam, total int, err error) {
	args := []interface{}{f.Limit, f.Offset}
	qs, args := SubjectExamListBuildQuery(f, args)

	err = d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		rows, err := tx.Query(ctx, qs, args...)
		for rows.Next() {
			sub := models.SubjectExam{}
			err = scanSubjectExam(rows, &sub, &total)
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

func (d *PgxStore) SubjectExamUpdate(ctx context.Context, m *models.SubjectExam) (*models.SubjectExam, error) {
	qs, args := SubjectExamUpdateQuery(m)
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		_, err = tx.Exec(ctx, qs, args...)
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}

	editModel, err := d.SubjectExamFindById(ctx, m.ID)
	if err != nil {
		return nil, err
	}
	return editModel, nil
}

func (d *PgxStore) SubjectExamCreate(ctx context.Context, m *models.SubjectExam) (*models.SubjectExam, error) {
	qs, args := SubjectExamCreateQuery(m)
	qs += " RETURNING uid"
	err := d.runQuery(context.Background(), func(tx *pgxpool.Conn) (err error) {
		err = tx.QueryRow(context.Background(), qs, args...).Scan(&m.ID)
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}

	editModel, err := d.SubjectExamFindById(ctx, m.ID)
	if err != nil {
		return nil, err
	}
	return editModel, nil
}

func (d *PgxStore) SubjectExamDelete(ctx context.Context, model []*models.SubjectExam) ([]*models.SubjectExam, error) {
	ids := []string{}
	for _, i := range model {
		ids = append(ids, i.ID)
	}
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		_, err = tx.Query(ctx, sqlSubjectExamDelete, (ids))
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}
	return model, nil
}

func SubjectExamCreateQuery(m *models.SubjectExam) (string, []interface{}) {
	args := []interface{}{}
	cols := ""
	vals := ""
	q := SubjectExamAtomicQuery(m, true)
	for k, v := range q {
		args = append(args, v)
		cols += ", " + k
		vals += ", $" + strconv.Itoa(len(args))
	}
	qs := sqlSubjectExamInsert + " (" + strings.Trim(cols, ", ") + ") VALUES (" + strings.Trim(vals, ", ") + ")"
	return qs, args
}

func SubjectExamUpdateQuery(m *models.SubjectExam) (string, []interface{}) {
	args := []interface{}{}
	sets := ""
	q := SubjectExamAtomicQuery(m, false)
	for k, v := range q {
		args = append(args, v)
		sets += ", " + k + "=$" + strconv.Itoa(len(args))
	}
	args = append(args, m.ID)
	qs := strings.ReplaceAll(sqlSubjectExamUpdate, "set uid=uid", "set uid=uid "+sets+" ") + "where uid=$" + strconv.Itoa(len(args))
	return qs, args
}

func SubjectExamAtomicQuery(m *models.SubjectExam, isCreate bool) map[string]interface{} {
	q := map[string]interface{}{}
	if m.SubjectId != "" {
		q["subject_uid"] = m.SubjectId
	}
	if m.SchoolId != "" {
		q["school_uid"] = m.SchoolId
	}
	if m.TeacherId != "" {
		q["teacher_uid"] = m.TeacherId
	}
	if m.ClassroomId != "" {
		q["classroom_uid"] = m.ClassroomId
	}
	if m.StartTime != nil {
		q["start_time"] = m.StartTime
	}
	if m.HeadTeacherId != nil {
		q["head_teacher_uid"] = m.HeadTeacherId
	}
	if m.MemberTeacherIds != nil {
		q["member_teacher_uids"] = m.MemberTeacherIds
	}
	if m.RoomNumber != nil {
		q["room_number"] = m.RoomNumber
	}
	if m.TimeLengthMin != nil {
		q["time_length_min"] = m.TimeLengthMin
	}
	if m.ExamWeightPercent != nil {
		q["exam_weight_percent"] = m.ExamWeightPercent
	}
	if m.Name != nil {
		q["name"] = m.Name
	}
	if m.IsRequired != nil {
		q["is_required"] = m.IsRequired
	}
	if isCreate {
		q["created_at"] = time.Now()
	}
	q["updated_at"] = time.Now()
	return q
}

func SubjectExamListBuildQuery(f *models.SubjectExamFilterRequest, args []interface{}) (string, []interface{}) {
	var wheres string = ""

	if f.ID != nil && *f.ID != "" {
		args = append(args, *f.ID)
		wheres += " and se.uid=$" + strconv.Itoa(len(args)) + "::uuid"
	}
	if f.Ids != nil && len(*f.Ids) > 0 {
		args = append(args, *f.Ids)
		wheres += " and se.uid = ANY($" + strconv.Itoa(len(args)) + "::uuid[])"
	}
	if f.SubjectId != nil && *f.SubjectId != "" {
		args = append(args, *f.SubjectId)
		wheres += " and se.subject_uid=$" + strconv.Itoa(len(args)) + "::uuid"
	}
	if f.SubjectIds != nil && len(f.SubjectIds) > 0 {
		args = append(args, (f.SubjectIds))
		wheres += " and se.subject_uid = ANY($" + strconv.Itoa(len(args)) + "::uuid[])"
	}
	if len(f.SchoolIds) > 0 {
		args = append(args, (f.SchoolIds))
		wheres += " and se.school_uid = ANY($" + strconv.Itoa(len(args)) + "::uuid[])"
	}
	if f.SchoolId != nil && *f.SchoolId != "" {
		args = append(args, *f.SchoolId)
		wheres += " and se.school_uid=$" + strconv.Itoa(len(args)) + "::uuid"
	}
	if f.TeacherId != nil {
		args = append(args, f.TeacherId)
		wheres += " and se.teacher_uid=$" + strconv.Itoa(len(args)) + "::uuid"
	}
	if f.ClassroomNames != nil {
		args = append(args, f.ClassroomNames)
		wheres += " and regexp_replace(c.name, '[^0-9]+', '', 'g')=ANY($" + strconv.Itoa(len(args)) + ")"
	}
	if f.IsGraduate != nil && *f.IsGraduate {
		wheres += " and c.name LIKE '11%'"
	}
	if f.Search != nil && *f.Search != "" {
		*f.Search = strings.ToLower(*f.Search)
		args = append(args, *f.Search)
		tmp := "$" + strconv.Itoa(len(args))
		wheres += " and (start_time::text like '%' || " + tmp + " || '%'   or lower(room_number) like '%' || " + tmp + " || '%')"
	}
	if f.Sort != nil && *f.Sort != "" {
		dir := "desc"
		if strings.HasSuffix(*f.Sort, "~") {
			dir = "asc"
		}
		*f.Sort = strings.ReplaceAll(*f.Sort, "~", "")
		wheres += " order by " + *f.Sort + " " + dir
	} else {
		wheres += " order by uid desc"
	}
	qs := sqlSubjectExamSelectMany
	qs = strings.ReplaceAll(qs, "se.uid=se.uid", "se.uid=se.uid "+wheres+" ")
	return qs, args
}

func (d *PgxStore) SubjectExamLoadRelations(ctx context.Context, l *[]*models.SubjectExam) error {
	ids := []string{}
	for _, m := range *l {
		ids = append(ids, m.ID)
	}
	if len(ids) < 1 {
		return nil
	}
	// load subject
	if rs, err := d.SubjectExamLoadSubject(ctx, ids); err == nil {
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
	// load school
	if rs, err := d.SubjectExamLoadSchool(ctx, ids); err == nil {
		for _, r := range rs {
			for _, m := range *l {
				if r.ID == m.ID && r.Relation != nil {
					m.School = r.Relation
				}
			}
		}
	} else {
		return err
	}
	// load teacher
	if rs, err := d.SubjectExamLoadTeacher(ctx, ids); err == nil {
		for _, r := range rs {
			for _, m := range *l {
				if r.ID == m.ID {
					m.Teacher = r.Relation
				}
			}
		}
	} else {
		return err
	}
	// load head_teacher
	if rs, err := d.SubjectExamLoadHeadTeacher(ctx, ids); err == nil {
		for _, r := range rs {
			for _, m := range *l {
				if r.ID == m.ID {
					m.HeadTeacher = r.Relation
				}
			}
		}
	} else {
		return err
	}
	// load classroom
	if rs, err := d.SubjectExamLoadClassroom(ctx, ids); err == nil {
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
	// load member_teachers
	if rs, err := d.SubjectExamLoadMemberTeachers(ctx, ids); err == nil {
		for _, r := range rs {
			for _, m := range *l {
				if r.ID == m.ID {
					m.MemberTeachers = append(m.MemberTeachers, r.Relation)
				}
			}
		}
	} else {
		return err
	}
	return nil
}

type SubjectExamLoadSubjectItem struct {
	ID       string
	Relation *models.Subject
}

func (d *PgxStore) SubjectExamLoadSubject(ctx context.Context, ids []string) ([]SubjectExamLoadSubjectItem, error) {
	res := []SubjectExamLoadSubjectItem{}
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		rows, err := tx.Query(ctx, sqlSubjectExamSubject, (ids))
		for rows.Next() {
			sub := models.Subject{}
			pid := ""
			err = scanSubject(rows, &sub, &pid)
			if err != nil {
				return err
			}
			res = append(res, SubjectExamLoadSubjectItem{ID: pid, Relation: &sub})
		}
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}

	return res, nil
}

type SubjectExamLoadSchoolItem struct {
	ID       string
	Relation *models.School
}

func (d *PgxStore) SubjectExamLoadSchool(ctx context.Context, ids []string) ([]SubjectExamLoadSchoolItem, error) {
	res := []SubjectExamLoadSchoolItem{}
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		rows, err := tx.Query(ctx, sqlSubjectExamSchool, (ids))
		for rows.Next() {
			sub := models.School{}
			pid := ""
			err = scanSchool(rows, &sub, &pid)
			if err != nil {
				return err
			}
			res = append(res, SubjectExamLoadSchoolItem{ID: pid, Relation: &sub})
		}
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}

	return res, nil
}

type SubjectExamLoadTeacherItem struct {
	ID       string
	Relation *models.User
}

func (d *PgxStore) SubjectExamLoadTeacher(ctx context.Context, ids []string) ([]SubjectExamLoadTeacherItem, error) {
	res := []SubjectExamLoadTeacherItem{}
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		rows, err := tx.Query(ctx, sqlSubjectExamTeacher, (ids))
		for rows.Next() {
			sub := models.User{}
			pid := ""
			err = scanUser(rows, &sub, &pid)
			if err != nil {
				return err
			}
			res = append(res, SubjectExamLoadTeacherItem{ID: pid, Relation: &sub})
		}
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}

	return res, nil
}

type SubjectExamLoadHeadTeacherItem struct {
	ID       string
	Relation *models.User
}

func (d *PgxStore) SubjectExamLoadHeadTeacher(ctx context.Context, ids []string) ([]SubjectExamLoadHeadTeacherItem, error) {
	res := []SubjectExamLoadHeadTeacherItem{}
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		rows, err := tx.Query(ctx, sqlSubjectExamHeadTeacher, (ids))
		for rows.Next() {
			sub := models.User{}
			pid := ""
			err = scanUser(rows, &sub, &pid)
			if err != nil {
				return err
			}
			res = append(res, SubjectExamLoadHeadTeacherItem{ID: pid, Relation: &sub})
		}
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}

	return res, nil
}

type SubjectExamLoadMemberTeachersItems struct {
	ID       string
	Relation *models.User
}

func (d *PgxStore) SubjectExamLoadMemberTeachers(ctx context.Context, ids []string) ([]SubjectExamLoadMemberTeachersItems, error) {
	res := []SubjectExamLoadMemberTeachersItems{}
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		rows, err := tx.Query(ctx, sqlSubjectExamMemberTeachers, (ids))
		for rows.Next() {
			sub := models.User{}
			pid := ""
			err = scanUser(rows, &sub, &pid)
			if err != nil {
				return err
			}
			res = append(res, SubjectExamLoadMemberTeachersItems{ID: pid, Relation: &sub})
		}
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}

	return res, nil
}

type SubjectExamLoadClassroomItem struct {
	ID       string
	Relation *models.Classroom
}

func (d *PgxStore) SubjectExamLoadClassroom(ctx context.Context, ids []string) ([]SubjectExamLoadClassroomItem, error) {
	res := []SubjectExamLoadClassroomItem{}
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		rows, err := tx.Query(ctx, sqlSubjectExamClassroom, (ids))
		for rows.Next() {
			sub := models.Classroom{}
			pid := ""
			err = scanClassroom(rows, &sub, &pid)
			if err != nil {
				return err
			}
			res = append(res, SubjectExamLoadClassroomItem{ID: pid, Relation: &sub})
		}
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}

	return res, nil
}
