package pgx

import (
	"context"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/mekdep/server/internal/models"
	"github.com/mekdep/server/internal/utils"
)

// base
const sqlSubjectFields = `sb.uid as uid, sb.school_uid, sb.classroom_uid, sb.classroom_type, sb.classroom_type_key, sb.base_subject_uid, sb.parent_uid, sb.teacher_uid, sb.second_teacher_uid, sb.name, sb.full_name, sb.week_hours, sb.updated_at, sb.created_at`
const sqlSubjectSelect = `select ` + sqlSubjectFields + ` from subjects sb where uid = ANY($1::uuid[])`
const sqlSubjectSelectMany = `select ` + sqlSubjectFields + `, count(*) over() as total from subjects sb 
	right join classrooms c on c.uid=sb.classroom_uid
	left join subject_exams se on se.subject_uid=sb.uid
	where sb.uid=sb.uid limit $1 offset $2`
const sqlSubjectSelectParentId = `select sb.uid from subjects sb where sb.school_uid=$1 and sb.classroom_uid=$2 and sb.name=$3 and sb.full_name=$4 limit 1`
const sqlSubjectInsert = `insert into subjects`
const sqlSubjectUpdate = `update subjects sb set uid=uid`
const sqlSubjectDelete = `delete from subjects sb where uid = ANY($1::uuid[])`
const sqlFindSubjectByClassroomId = `select ` + sqlSubjectFields + ` from subjects sb where sb.classroom_uid = $1`

// relations
const sqlSubjectSchool = `select ` + sqlSchoolFields + `, sb.uid from subjects sb 
	right join schools s on (s.uid=sb.school_uid)  where sb.uid = ANY($1::uuid[])`

const sqlSubjectClassroom = `select ` + sqlClassroomFields + `, sb.uid from subjects sb 
	right join classrooms c on (c.uid=sb.classroom_uid) where sb.uid = ANY($1::uuid[])`

const sqlSubjectTeacher = `select ` + sqlUserFields + `, sb.uid, sb.classroom_uid, sb.name from subjects sb 
	right join users u on (u.uid=sb.teacher_uid)  where sb.uid = ANY($1::uuid[])`

const sqlSubjectSecondTeacher = `select ` + sqlUserFields + `, sb.uid from subjects sb 
	right join users u on (u.uid=sb.second_teacher_uid) where sb.uid = ANY($1::uuid[])`

const sqlSubjectSelectParent = `select ` + sqlSubjectFields + `, s.uid from subjects s 
	right join subjects sb on (sb.uid=s.parent_uid) where s.uid = ANY($1::uuid[])`

const sqlSubjectsChildren = `select ` + sqlSubjectFields + ` from subjects sb where sb.parent_uid = ANY($1::uuid[])`

const sqlSubjectSelectExams = `select ` + sqlSubjectExamFields + `, s.uid from subjects s 
	right join subject_exams se on (se.subject_uid=s.uid) where s.uid = ANY($1::uuid[])`

const sqlSubjectsBaseSubjects = `select ` + sqlBaseSubjectsFields + `, sb.uid from subjects sb
	right join base_subjects bs on (bs.uid=sb.base_subject_uid) where sb.uid = ANY($1::uuid[])`

// many to many
const sqlSubjectStudents = `select ` + sqlUserFields + `, usb.type, usb.subject_uid from user_subjects uc 
	left join users u on usb.user_uid=u.uid  where usb.subject_uid = ANY($1::uuid[])`
const sqlSubjectStudentsDelete = `delete from user_subjects where subject_uid=$1`
const sqlSubjectStudentsInsert = `insert into user_subjects (subject_uid, user_uid, type) values ($1, $2, $3)`

func scanSubject(rows pgx.Row, m *models.Subject, addColumns ...interface{}) (err error) {
	err = rows.Scan(parseColumnsForScan(m, addColumns...)...)
	return
}

func (d *PgxStore) SubjectsFindByIds(ctx context.Context, ids []string) ([]*models.Subject, error) {
	l := []*models.Subject{}
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		rows, err := tx.Query(ctx, sqlSubjectSelect, (ids))
		for rows.Next() {
			m := models.Subject{}
			err := scanSubject(rows, &m)
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

func (d *PgxStore) SubjectsFindById(ctx context.Context, id string) (*models.Subject, error) {
	l, err := d.SubjectsFindByIds(ctx, []string{id})
	if err != nil {
		return nil, err
	}
	if len(l) < 1 {
		err = pgx.ErrNoRows
		utils.LoggerDesc("Scan error uid " + id).Error(err)
		return nil, err
	}
	return l[0], nil
}

func (d *PgxStore) SubjectsListFilters(ctx context.Context, f *models.SubjectFilterRequest) ([]*models.Subject, int, error) {
	args := []interface{}{f.Limit, f.Offset}
	qs, args := SubjectsListBuildQuery(f, args)
	l := []*models.Subject{}
	var total int

	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		rows, err := tx.Query(ctx, qs, args...)
		for rows.Next() {
			sub := models.Subject{}
			err = scanSubject(rows, &sub, &total)
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

func (d *PgxStore) SubjectsUpdate(ctx context.Context, m *models.Subject) (*models.Subject, error) {
	// update
	qs, args := SubjectsUpdateQuery(m)
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		_, err = tx.Exec(ctx, qs, args...)
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}
	// load new
	editModel, err := d.SubjectsFindById(ctx, m.ID)
	if err != nil {
		return nil, err
	}
	d.SubjectsUpdateRelations(ctx, m, editModel)
	return editModel, nil
}

func (d *PgxStore) SubjectsCreate(ctx context.Context, m *models.Subject) (*models.Subject, error) {
	// set parent id
	var err error
	// TODO: ayratyn function etmeli we create, update-da cagyrmaly
	parentId := ""
	d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		err = tx.QueryRow(ctx, sqlSubjectSelectParentId, m.SchoolId, m.ClassroomId, m.Name, m.FullName).Scan(&parentId)
		return
	})
	if parentId != "" {
		m.ParentId = &parentId
	}

	// create
	qs, args := SubjectsCreateQuery(m)
	qs += " RETURNING uid"
	err = d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		err = tx.QueryRow(ctx, qs, args...).Scan(&m.ID)
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}
	// load new
	editModel, err := d.SubjectsFindById(ctx, m.ID)
	if err != nil {
		return nil, err
	}
	d.SubjectsUpdateRelations(ctx, m, editModel)
	return editModel, nil
}

func (d *PgxStore) SubjectsDelete(ctx context.Context, l []*models.Subject) ([]*models.Subject, error) {
	ids := []string{}
	for _, i := range l {
		ids = append(ids, i.ID)
	}
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		_, err = tx.Query(ctx, sqlSubjectDelete, (ids))
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}
	return l, nil
}

func (d *PgxStore) SubjectsUpdateRelations(ctx context.Context, data *models.Subject, model *models.Subject) {

}

func SubjectsCreateQuery(m *models.Subject) (string, []interface{}) {
	args := []interface{}{}
	cols := ""
	vals := ""
	q := SubjectAtomicQuery(m, true)
	for k, v := range q {
		args = append(args, v)
		cols += ", " + k
		vals += ", $" + strconv.Itoa(len(args))
	}
	qs := sqlSubjectInsert + " (" + strings.Trim(cols, ", ") + ") VALUES (" + strings.Trim(vals, ", ") + ")"
	return qs, args
}

func SubjectsUpdateQuery(m *models.Subject) (string, []interface{}) {
	args := []interface{}{}
	sets := ""
	q := SubjectAtomicQuery(m, false)
	for k, v := range q {
		args = append(args, v)
		sets += ", " + k + "=$" + strconv.Itoa(len(args))
	}
	args = append(args, m.ID)
	qs := strings.ReplaceAll(sqlSubjectUpdate, "set uid=uid", "set uid=uid "+sets+" ") + "where uid=$" + strconv.Itoa(len(args))
	return qs, args
}

func SubjectAtomicQuery(m *models.Subject, isCreate bool) map[string]interface{} {
	q := map[string]interface{}{}
	if m.ClassroomId != "" {
		q["classroom_uid"] = m.ClassroomId
	}
	q["classroom_type"] = m.ClassroomType
	q["classroom_type_key"] = m.ClassroomTypeKey
	if m.SecondTeacherId != nil && *m.SecondTeacherId == "" {
		m.SecondTeacherId = nil
	}
	q["second_teacher_uid"] = m.SecondTeacherId
	if m.TeacherId != nil {
		q["teacher_uid"] = m.TeacherId
	}
	if m.ParentId != nil {
		q["parent_uid"] = m.ParentId
	}
	if m.BaseSubjectId != nil {
		q["base_subject_uid"] = m.BaseSubjectId
	}
	if m.Name != nil {
		q["name"] = m.Name
	}
	if m.FullName != nil {
		q["full_name"] = m.FullName
	}
	if m.WeekHours != nil {
		q["week_hours"] = m.WeekHours
	}
	if m.SchoolId != "" {
		q["school_uid"] = m.SchoolId
	}
	if isCreate {
		q["created_at"] = time.Now()
	}
	q["updated_at"] = time.Now()
	return q
}

func SubjectsListBuildQuery(f *models.SubjectFilterRequest, args []interface{}) (string, []interface{}) {
	var wheres string = ""

	if f.ID != nil && *f.ID != "" {
		args = append(args, *f.ID)
		wheres += " and sb.uid=$" + strconv.Itoa(len(args))
	}
	if f.NotIds != nil && len(*f.NotIds) > 0 {
		args = append(args, *f.NotIds)
		wheres += " and sb.uid <> ANY($" + strconv.Itoa(len(args)) + "::uuid[])"
	}
	if f.Ids != nil && len(*f.Ids) > 0 {
		args = append(args, *f.Ids)
		wheres += " and sb.uid = ANY($" + strconv.Itoa(len(args)) + "::uuid[])"
	}
	if f.ClassroomId != nil {
		args = append(args, *f.ClassroomId)
		wheres += " and sb.classroom_uid=$" + strconv.Itoa(len(args))
	}
	if len(f.ClassroomIds) > 0 {
		args = append(args, (f.ClassroomIds))
		wheres += " and sb.classroom_uid = ANY($" + strconv.Itoa(len(args)) + "::uuid[])"
	}
	if f.BaseSubjectId != nil {
		args = append(args, *f.BaseSubjectId)
		wheres += " and sb.base_subject_uid=$" + strconv.Itoa(len(args))
	}
	if f.ClassroomTypeKey != nil {
		args = append(args, f.ClassroomTypeKey)
		wheres += " and sb.classroom_type_key=$" + strconv.Itoa(len(args))
	}
	if f.TeacherId != nil && *f.TeacherId != "" {
		args = append(args, *f.TeacherId)
		wheres += " and sb.teacher_uid=$" + strconv.Itoa(len(args))
	}
	if len(f.TeacherIds) > 0 {
		args = append(args, (f.TeacherIds))
		wheres += " and (sb.teacher_uid = ANY($" + strconv.Itoa(len(args)) + ") or sb.second_teacher_uid = ANY($" + strconv.Itoa(len(args)) + ") or se.teacher_uid = ANY($" + strconv.Itoa(len(args)) + ") )"
	}
	if f.IsSecondTeacher != nil && *f.IsSecondTeacher {
		wheres += " and sb.second_teacher_uid is not null"
	}
	if f.IsSubjectExam != nil && *f.IsSubjectExam {
		if *f.IsSubjectExam {
			wheres += " and se.uid is not null"
		} else {
			wheres += " and se.uid is null"
		}
	}
	if f.SchoolIds != nil {
		args = append(args, (f.SchoolIds))
		wheres += " and sb.school_uid = ANY($" + strconv.Itoa(len(args)) + ")"
	}
	if f.SchoolId != nil && *f.SchoolId != "" {
		args = append(args, *f.SchoolId)
		wheres += " and sb.school_uid=$" + strconv.Itoa(len(args))
	}

	if f.WeekHours != nil {
		args = append(args, *f.WeekHours)
		wheres += " and sb.week_hours = $" + strconv.Itoa(len(args))
	}
	if len(f.WeekHoursRange) > 0 {
		args = append(args, f.WeekHoursRange[0])
		wheres += " and sb.week_hours >= $" + strconv.Itoa(len(args))
		if len(f.WeekHoursRange) > 1 {
			args = append(args, f.WeekHoursRange[1])
			wheres += " and sb.week_hours <= $" + strconv.Itoa(len(args))
		}
	}
	if len(f.SubjectNames) > 0 {
		args = append(args, (f.SubjectNames))
		wheres += " and sb.name = ANY($" + strconv.Itoa(len(args)) + ")"
	}
	if f.Search != nil && *f.Search != "" {
		*f.Search = strings.ToLower(*f.Search)
		args = append(args, *f.Search)
		tmp := "$" + strconv.Itoa(len(args))
		wheres += " and (lower(sb.name) like '%' || $" + strconv.Itoa(len(args)) + "|| '%' or lower(sb.full_name) like '%" + tmp + "%')"
	}
	wheres += " group by sb.uid"
	if f.Sort != nil && *f.Sort != "" {
		dir := "desc"
		if strings.HasSuffix(*f.Sort, "~") {
			dir = "asc"
		}
		*f.Sort = strings.ReplaceAll(*f.Sort, "~", "")
		wheres += " order by sb." + *f.Sort + " " + dir
	} else {
		wheres += " order by LPAD(lower(max(c.name)), 3, '0') asc, sb.name asc"
	}
	qs := sqlSubjectSelectMany
	qs = strings.ReplaceAll(qs, "sb.uid=sb.uid", "sb.uid=sb.uid "+wheres+" ")
	return qs, args
}

func (d *PgxStore) SubjectsLoadRelations(ctx context.Context, l *[]*models.Subject, isDetail bool) error {
	ids := []string{}
	for _, m := range *l {
		ids = append(ids, m.ID)
	}
	if len(ids) < 1 {
		return nil
	}
	// load admin
	if rs, err := d.SubjectsLoadSchool(ctx, ids); err == nil {
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
	// load classrooms
	if rs, err := d.SubjectsLoadClassroom(ctx, ids); err == nil {
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
	if rs, err := d.SubjectsLoadBaseSubject(ctx, ids); err == nil {
		for _, r := range rs {
			for _, m := range *l {
				if r.ID == m.ID {
					m.BaseSubject = r.Relation
				}
			}
		}
	} else {
		return err
	}
	// load teachers
	if rs, err := d.SubjectsLoadTeacher(ctx, ids); err == nil {
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
	if rs, err := d.SubjectsLoadExam(ctx, ids); err == nil {
		for _, r := range rs {
			err = d.SubjectExamLoadRelations(ctx, &r.Relation)
			if err != nil {
				return err
			}
			for _, m := range *l {
				if r.ID == m.ID {
					m.Exams = r.Relation
				}
			}
		}
	} else {
		return err
	}
	if rs, err := d.SubjectsLoadParent(ctx, ids); err == nil {
		for _, r := range rs {
			for _, m := range *l {
				if r.ID == m.ID {
					m.Parent = r.Relation
				}
			}
		}
	}
	err := d.SubjectsLoadChildren(ctx, l)
	if err != nil {
		return err
	}
	if isDetail {
		// load second teachers
		if rs, err := d.SubjectsLoadSecondTeacher(ctx, ids); err == nil {
			for _, r := range rs {
				for _, m := range *l {
					if r.ID == m.ID {
						m.SecondTeacher = r.Relation
					}
				}
			}
		} else {
			return err
		}
	}
	return nil
}

type SubjectsLoadSchoolItem struct {
	ID       string
	Relation *models.School
}

func (d *PgxStore) SubjectsLoadSchool(ctx context.Context, ids []string) ([]SubjectsLoadSchoolItem, error) {
	res := []SubjectsLoadSchoolItem{}

	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		rows, err := tx.Query(ctx, sqlSubjectSchool, (ids))
		for rows.Next() {
			sub := models.School{}
			pid := ""
			err = scanSchool(rows, &sub, &pid)
			if err != nil {
				return err
			}
			res = append(res, SubjectsLoadSchoolItem{ID: pid, Relation: &sub})
		}
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}

	return res, nil
}

type SubjectsLoadClassroomItem struct {
	ID       string
	Relation *models.Classroom
}

func (d *PgxStore) SubjectsLoadClassroom(ctx context.Context, ids []string) ([]SubjectsLoadClassroomItem, error) {
	res := []SubjectsLoadClassroomItem{}

	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		rows, err := tx.Query(ctx, sqlSubjectClassroom, (ids))
		for rows.Next() {
			sub := models.Classroom{}
			pid := ""
			err = scanClassroom(rows, &sub, &pid)
			if err != nil {
				return err
			}
			res = append(res, SubjectsLoadClassroomItem{ID: pid, Relation: &sub})
		}
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}

	return res, nil
}

type SubjectsLoadTeacherItem struct {
	ID          string
	ClassroomId string
	SubjectName string
	Relation    *models.User
}

func (d *PgxStore) SubjectsLoadTeacher(ctx context.Context, ids []string) ([]SubjectsLoadTeacherItem, error) {
	res := []SubjectsLoadTeacherItem{}

	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		rows, err := tx.Query(ctx, sqlSubjectTeacher, (ids))
		for rows.Next() {
			item := SubjectsLoadTeacherItem{}
			sub := models.User{}
			err = scanUser(rows, &sub, &item.ID, &item.ClassroomId, &item.SubjectName)
			item.Relation = &sub
			if err != nil {
				return err
			}
			res = append(res, item)
		}
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}

	return res, nil
}

type SubjectsLoadSecondTeacherItem struct {
	ID       string
	Relation *models.User
}

func (d *PgxStore) SubjectsLoadSecondTeacher(ctx context.Context, ids []string) ([]SubjectsLoadSecondTeacherItem, error) {
	res := []SubjectsLoadSecondTeacherItem{}

	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		rows, err := tx.Query(ctx, sqlSubjectSecondTeacher, (ids))
		for rows.Next() {
			sub := models.User{}
			pid := ""
			err = scanUser(rows, &sub, &pid)
			if err != nil {
				return err
			}
			res = append(res, SubjectsLoadSecondTeacherItem{ID: pid, Relation: &sub})
		}
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}

	return res, nil
}

type SubjectsLoadParentItem struct {
	ID       string
	Relation *models.Subject
}

func (d *PgxStore) SubjectsLoadParent(ctx context.Context, ids []string) ([]SubjectsLoadParentItem, error) {
	res := []SubjectsLoadParentItem{}

	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		rows, err := tx.Query(ctx, sqlSubjectSelectParent, (ids))
		for rows.Next() {
			sub := models.Subject{}
			pid := ""
			err = scanSubject(rows, &sub, &pid)
			if err != nil {
				return err
			}
			res = append(res, SubjectsLoadParentItem{ID: pid, Relation: &sub})
		}
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}

	return res, nil
}

func (d *PgxStore) SubjectsLoadChildren(ctx context.Context, l *[]*models.Subject) error {
	ids := []string{}

	for _, m := range *l {
		ids = append(ids, m.ID)
	}
	if len(ids) < 1 {
		return nil
	}
	// load
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		rows, err := tx.Query(ctx, sqlSubjectsChildren, (ids))
		for rows.Next() {
			sub := models.Subject{}
			err = scanSubject(rows, &sub)
			if err != nil {
				utils.LoggerDesc("Scan error").Error(err)
				return err
			}
			for _, m := range *l {
				if slices.Contains(ids, m.ID) {
					m.Children = append(m.Children, &sub)
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

type SubjectsLoadExamItem struct {
	ID       string
	Relation []*models.SubjectExam
}

func (d *PgxStore) SubjectsLoadExam(ctx context.Context, ids []string) ([]SubjectsLoadExamItem, error) {
	res := []SubjectsLoadExamItem{}
	for _, v := range ids {
		res = append(res, SubjectsLoadExamItem{ID: v, Relation: []*models.SubjectExam{}})
	}
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		rows, err := tx.Query(ctx, sqlSubjectSelectExams, (ids))
		for rows.Next() {
			sub := models.SubjectExam{}
			pid := ""
			err = scanSubjectExam(rows, &sub, &pid)
			if err != nil {
				return err
			}
			for k, v := range res {
				if v.ID == pid {
					res[k].Relation = append(res[k].Relation, &sub)
				}
			}
		}
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}

	return res, nil
}

type SubjectsLoadBaseSubjectItem struct {
	ID       string
	Relation *models.BaseSubjects
}

func (d *PgxStore) SubjectsLoadBaseSubject(ctx context.Context, ids []string) ([]SubjectsLoadBaseSubjectItem, error) {
	res := []SubjectsLoadBaseSubjectItem{}

	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		rows, err := tx.Query(ctx, sqlSubjectsBaseSubjects, (ids))
		for rows.Next() {
			sub := models.BaseSubjects{}
			pid := ""
			err = scanBaseSubjects(rows, &sub, &pid)
			if err != nil {
				return err
			}
			res = append(res, SubjectsLoadBaseSubjectItem{ID: pid, Relation: &sub})
		}
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}

	return res, nil
}

func (d *PgxStore) SubjectsFindByClassroomId(ctx context.Context, classroomId string) ([]*models.Subject, error) {
	subjects := []*models.Subject{}

	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		rows, err := tx.Query(ctx, sqlFindSubjectByClassroomId, classroomId)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			subject := models.Subject{}
			if err := scanSubject(rows, &subject); err != nil {
				return err
			}
			subjects = append(subjects, &subject)
		}
		return
	})

	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}

	return subjects, nil
}
