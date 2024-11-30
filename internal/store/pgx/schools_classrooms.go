package pgx

import (
	"context"
	"errors"
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
const sqlClassroomFields = sqlClassroomFieldsNoRelation + `, null`
const sqlClassroomFieldsNoRelation = `c.uid, c.school_uid, c.shift_uid, c.name, c.name_canonical, c.avatar, c.description, c.language, c.level, c.teacher_uid, c.student_uid, c.parent_uid, c.period_uid, c.updated_at, c.created_at, c.archived_at`
const sqlClassroomSelect = `select ` + sqlClassroomFieldsNoRelation + `, count(uc.*) as students_count from classrooms c 
	left join user_classrooms uc on (uc.classroom_uid=c.uid and uc.type is null and uc.type_key is null) where c.uid = ANY($1::uuid[]) group by c.uid`
const sqlClassroomSelectMany = `select ` + sqlClassroomFieldsNoRelation + `, count(uc.*) as students_count, count(*) over() as total from classrooms c 
	left join user_classrooms uc on (uc.classroom_uid=c.uid and uc.type is null and uc.type_key is null) where c.uid=c.uid limit $1 offset $2 `
const sqlClassroomInsert = `insert into classrooms`
const sqlClassroomUpdate = `update classrooms c set uid=uid`
const sqlClassroomDelete = `delete from classrooms c where uid = ANY($1::uuid[])`

const sqlClassroomStudentsCountBySchool = `select s.code, 
(select  array[count(DISTINCT c1.uid), count(DISTINCT ub1.uid), count(DISTINCT ug1.uid)] from classrooms c1
left join user_classrooms uc1 ON c1.uid=uc1.classroom_uid
left join users ub1 ON (uc1.user_uid=ub1.uid and (ub1.gender=0 or ub1.gender=1 or ub1.gender IS NULL) )
left join users ug1 ON (uc1.user_uid=ug1.uid and ug1.gender=2)
where s.uid=c1.school_uid and c1.name similar to '1[^0-9]'),

(select  array[count(DISTINCT c1.uid), count(DISTINCT ub1.uid), count(DISTINCT ug1.uid)] from classrooms c1
left join user_classrooms uc1 ON c1.uid=uc1.classroom_uid
left join users ub1 ON (uc1.user_uid=ub1.uid and (ub1.gender=0 or ub1.gender=1 or ub1.gender IS NULL) )
left join users ug1 ON (uc1.user_uid=ug1.uid and ug1.gender=2)
where s.uid=c1.school_uid and c1.name similar to '2[^0-9]'),

(select  array[count(DISTINCT c1.uid), count(DISTINCT ub1.uid), count(DISTINCT ug1.uid)] from classrooms c1
left join user_classrooms uc1 ON c1.uid=uc1.classroom_uid
left join users ub1 ON (uc1.user_uid=ub1.uid and (ub1.gender=0 or ub1.gender=1 or ub1.gender IS NULL) )
left join users ug1 ON (uc1.user_uid=ug1.uid and ug1.gender=2)
where s.uid=c1.school_uid and c1.name similar to '3[^0-9]'),

(select  array[count(DISTINCT c1.uid), count(DISTINCT ub1.uid), count(DISTINCT ug1.uid)] from classrooms c1
left join user_classrooms uc1 ON c1.uid=uc1.classroom_uid
left join users ub1 ON (uc1.user_uid=ub1.uid and (ub1.gender=0 or ub1.gender=1 or ub1.gender IS NULL) )
left join users ug1 ON (uc1.user_uid=ug1.uid and ug1.gender=2)
where s.uid=c1.school_uid and c1.name similar to '4[^0-9]'),

(select  array[count(DISTINCT c1.uid), count(DISTINCT ub1.uid), count(DISTINCT ug1.uid)] from classrooms c1
left join user_classrooms uc1 ON c1.uid=uc1.classroom_uid
left join users ub1 ON (uc1.user_uid=ub1.uid and (ub1.gender=0 or ub1.gender=1 or ub1.gender IS NULL) )
left join users ug1 ON (uc1.user_uid=ug1.uid and ug1.gender=2)
where s.uid=c1.school_uid and c1.name similar to '5[^0-9]'),

(select  array[count(DISTINCT c1.uid), count(DISTINCT ub1.uid), count(DISTINCT ug1.uid)] from classrooms c1
left join user_classrooms uc1 ON c1.uid=uc1.classroom_uid
left join users ub1 ON (uc1.user_uid=ub1.uid and (ub1.gender=0 or ub1.gender=1 or ub1.gender IS NULL) )
left join users ug1 ON (uc1.user_uid=ug1.uid and ug1.gender=2)
where s.uid=c1.school_uid and c1.name similar to '6[^0-9]'),

(select  array[count(DISTINCT c1.uid), count(DISTINCT ub1.uid), count(DISTINCT ug1.uid)] from classrooms c1
left join user_classrooms uc1 ON c1.uid=uc1.classroom_uid
left join users ub1 ON (uc1.user_uid=ub1.uid and (ub1.gender=0 or ub1.gender=1 or ub1.gender IS NULL) )
left join users ug1 ON (uc1.user_uid=ug1.uid and ug1.gender=2)
where s.uid=c1.school_uid and c1.name similar to '7[^0-9]'),

(select  array[count(DISTINCT c1.uid), count(DISTINCT ub1.uid), count(DISTINCT ug1.uid)] from classrooms c1
left join user_classrooms uc1 ON c1.uid=uc1.classroom_uid
left join users ub1 ON (uc1.user_uid=ub1.uid and (ub1.gender=0 or ub1.gender=1 or ub1.gender IS NULL) )
left join users ug1 ON (uc1.user_uid=ug1.uid and ug1.gender=2)
where s.uid=c1.school_uid and c1.name similar to '8[^0-9]'),

(select  array[count(DISTINCT c1.uid), count(DISTINCT ub1.uid), count(DISTINCT ug1.uid)] from classrooms c1
left join user_classrooms uc1 ON c1.uid=uc1.classroom_uid
left join users ub1 ON (uc1.user_uid=ub1.uid and (ub1.gender=0 or ub1.gender=1 or ub1.gender IS NULL) )
left join users ug1 ON (uc1.user_uid=ug1.uid and ug1.gender=2)
where s.uid=c1.school_uid and c1.name similar to '9[^0-9]'),

(select  array[count(DISTINCT c1.uid), count(DISTINCT ub1.uid), count(DISTINCT ug1.uid)] from classrooms c1
left join user_classrooms uc1 ON c1.uid=uc1.classroom_uid
left join users ub1 ON (uc1.user_uid=ub1.uid and (ub1.gender=0 or ub1.gender=1 or ub1.gender IS NULL) )
left join users ug1 ON (uc1.user_uid=ug1.uid and ug1.gender=2)
where s.uid=c1.school_uid and c1.name similar to '10[^0-9]'),

(select  array[count(DISTINCT c1.uid), count(DISTINCT ub1.uid), count(DISTINCT ug1.uid)] from classrooms c1
left join user_classrooms uc1 ON c1.uid=uc1.classroom_uid
left join users ub1 ON (uc1.user_uid=ub1.uid and (ub1.gender=0 or ub1.gender=1 or ub1.gender IS NULL) )
left join users ug1 ON (uc1.user_uid=ug1.uid and ug1.gender=2)
where s.uid=c1.school_uid and c1.name similar to '11[^0-9]'),

(select  array[count(DISTINCT c1.uid), count(DISTINCT ub1.uid), count(DISTINCT ug1.uid)] from classrooms c1
left join user_classrooms uc1 ON c1.uid=uc1.classroom_uid
left join users ub1 ON (uc1.user_uid=ub1.uid and (ub1.gender=0 or ub1.gender=1 or ub1.gender IS NULL) )
left join users ug1 ON (uc1.user_uid=ug1.uid and ug1.gender=2)
where s.uid=c1.school_uid and c1.name similar to '12[^0-9]')

from schools s
GROUP by s.uid;`

// relations
const sqlClassroomSchool = `select ` + sqlSchoolFields + `, c.uid from classrooms c 
	right join schools s on (s.uid=c.school_uid) where c.uid = ANY($1::uuid[])`

const sqlClassroomShift = `select ` + sqlShiftFields + `, c.uid from classrooms c 
	right join shifts sh on (sh.uid=c.shift_uid) where c.uid = ANY($1::uuid[])`

const sqlClassroomParent = `select ` + sqlClassroomFields + `, cp.uid from classrooms cp 
	right join classrooms c on cp.parent_uid=c.uid  where cp.uid = ANY($1::uuid[])`

const sqlClassroomTeacher = `select ` + sqlUserFields + `, c.uid from classrooms c 
	right join users u on (u.uid=c.teacher_uid) where c.uid = ANY($1::uuid[])`

const sqlClassroomPeriod = `select ` + sqlPeriodFields + `, c.uid from classrooms c 
	right join periods p on (p.uid=c.period_uid) where c.uid = ANY($1::uuid[])`

const sqlClassroomStudent = `select ` + sqlUserFields + `, c.uid from classrooms c 
	right join users u on (u.uid=c.student_uid) where c.uid = ANY($1::uuid[])`

// many to many
const sqlClassroomStudents = `select ` + sqlUserFields + `, uc.type, uc.type_key, uc.classroom_uid from user_classrooms uc 
	right join users u on uc.user_uid=u.uid where uc.classroom_uid = ANY($1::uuid[])`
const sqlClassroomStudentsDeleteStudent = `delete from user_classrooms where user_uid=ANY($1::uuid[])`
const sqlClassroomStudentsDeleteType = `delete from user_classrooms where classroom_uid=$1 and type=$2 and type_key=$3`
const sqlClassroomStudentsDeleteTypeAny = `delete from user_classrooms where classroom_uid=$1 and type=$2`
const sqlClassroomStudentsInsert = `insert into user_classrooms (classroom_uid, user_uid, type, type_key) values ($1, $2, $3, $4)`
const sqlClassroomStudentsDeleteMain = `delete from user_classrooms where classroom_uid=$1 and type is null and type_key is null`
const sqlClassroomStudentsInsertMain = `insert into user_classrooms (classroom_uid, user_uid, type, type_key) values ($1, $2, null, null)`
const sqlClassroomSubjects = `select ` + sqlSubjectFields + `, sb.classroom_uid from subjects sb where sb.classroom_uid = ANY($1::uuid[])`

func scanClassroom(rows pgx.Row, m *models.Classroom, addColumns ...interface{}) (err error) {
	err = rows.Scan(parseColumnsForScan(m, addColumns...)...)
	return
}

func (d *PgxStore) ClassroomsFindByIds(ctx context.Context, ids []string) ([]*models.Classroom, error) {
	l := []*models.Classroom{}
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		rows, err := tx.Query(ctx, sqlClassroomSelect, (ids))
		for rows.Next() {
			m := models.Classroom{}
			err := scanClassroom(rows, &m)
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

func (d *PgxStore) ClassroomsFindById(ctx context.Context, id string) (*models.Classroom, error) {
	l, err := d.ClassroomsFindByIds(ctx, []string{id})
	if err != nil {
		return nil, err
	}
	if len(l) < 1 {
		return nil, errors.New("no rows was found by uid: " + id)
	}
	return l[0], nil
}

func (d *PgxStore) ClassroomsFindBy(ctx context.Context, f models.ClassroomFilterRequest) ([]*models.Classroom, int, error) {
	if f.Limit == nil {
		f.Limit = new(int)
		*f.Limit = 12
	}
	if f.Offset == nil {
		f.Offset = new(int)
		*f.Offset = 0
	}
	args := []interface{}{f.Limit, f.Offset}
	qs, args := ClassroomsListBuildQuery(f, args)

	l := []*models.Classroom{}
	var total int
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		rows, err := tx.Query(ctx, qs, args...)
		if err != nil {
			return err
		}
		for rows.Next() {
			sub := models.Classroom{}
			err := scanClassroom(rows, &sub, &total)
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

func (d *PgxStore) ClassroomStudentsCountBySchool(ctx context.Context) ([]models.SchoolStudentsCount, error) {
	var counts []models.SchoolStudentsCount

	err := d.runQuery(ctx, func(tx *pgxpool.Conn) error {
		rows, err := tx.Query(ctx, sqlClassroomStudentsCountBySchool)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			count := models.SchoolStudentsCount{
				Counts: [][]int{
					[]int{},
					[]int{},
					[]int{},
					[]int{},
					[]int{},
					[]int{},
					[]int{},
					[]int{},
					[]int{},
					[]int{},
					[]int{},
					[]int{},
				},
			}
			if err := rows.Scan(&count.SchoolCode, &count.Counts[0], &count.Counts[1], &count.Counts[2], &count.Counts[3], &count.Counts[4], &count.Counts[5], &count.Counts[6], &count.Counts[7], &count.Counts[8], &count.Counts[9], &count.Counts[10], &count.Counts[11]); err != nil {
				return err
			}
			counts = append(counts, count)
		}
		return nil
	})

	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}

	return counts, nil
}

func (d *PgxStore) ClassroomsUpdate(ctx context.Context, data *models.Classroom) (*models.Classroom, error) {
	// origModel := d.UsersFindById(strconv.Itoa(int(model.ID)))
	qs, args := ClassroomsUpdateQuery(data)
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		_, err = tx.Query(ctx, qs, args...)
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}

	editModel, err := d.ClassroomsFindById(ctx, data.ID)
	if err != nil {
		return nil, err
	}
	return editModel, nil
}

func (d *PgxStore) getLastPeriodUid(ctx context.Context, schoolUid string) (string, error) {
	var periodUid string
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		query := `SELECT uid FROM periods WHERE school_uid = $1 ORDER BY created_at DESC LIMIT 1`
		err = tx.QueryRow(ctx, query, schoolUid).Scan(&periodUid)
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return "", err
	}
	return periodUid, nil
}

func (d *PgxStore) ClassroomsCreate(ctx context.Context, m *models.Classroom) (*models.Classroom, error) {
	if m.PeriodId == nil {
		periodUid, err := d.getLastPeriodUid(ctx, m.SchoolId)
		if err != nil {
			return nil, err
		}
		m.PeriodId = &periodUid
	}
	qs, args := ClassroomsCreateQuery(m)
	qs += " RETURNING uid"
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		err = tx.QueryRow(ctx, qs, args...).Scan(&m.ID)
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}

	editModel, err := d.ClassroomsFindById(ctx, m.ID)
	if err != nil {
		return nil, err
	}
	return editModel, nil
}

func (d *PgxStore) ClassroomsDelete(ctx context.Context, l []*models.Classroom) ([]*models.Classroom, error) {
	ids := []string{}
	for _, i := range l {
		ids = append(ids, i.ID)
	}
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		_, err = tx.Query(ctx, sqlClassroomDelete, (ids))
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}
	return l, nil
}

func (d *PgxStore) ClassroomsDeleteStudent(ctx context.Context, userIds []string) error {
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		_, err = tx.Query(ctx, sqlClassroomStudentsDeleteStudent, userIds)
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
	}
	return err
}

func sliceUnique(l []string) []string {
	r := []string{}
	for _, v := range l {
		if !slices.Contains(r, v) {
			r = append(r, v)
		}
	}
	return r
}

func (d *PgxStore) ClassroomsUpdateRelations(ctx context.Context, data *models.Classroom, model *models.Classroom) error {

	// update students
	// TODO: refactor mess update students
	if data.Students != nil {
		err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
			_, err = tx.Query(ctx, sqlClassroomStudentsDeleteMain, data.ID)
			return
		})
		if err != nil {
			utils.LoggerDesc("Query error").Error(err)
			return err
		}
		toAddIds := []string{}
		for _, v := range data.Students {
			toAddIds = append(toAddIds, v.ID)
		}
		toAddIds = sliceUnique(toAddIds)
		toDeleteIds := []string{}
		for _, vOld := range model.Students {
			if !slices.Contains(toAddIds, vOld.ID) {
				toDeleteIds = append(toDeleteIds, vOld.ID)
			}
		}
		err = d.ClassroomsDeleteStudent(ctx, toDeleteIds)
		if err != nil {
			return err
		}

		for _, v := range toAddIds {
			err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
				_, err = tx.Query(ctx, sqlClassroomStudentsInsertMain, data.ID, v)
				if err != nil {
					return err
				}
				return nil
			})
			if err != nil {
				utils.LoggerDesc("Query error").Error(err)
				return err
			}
		}
	}
	if data.SubGroups != nil {
		for _, subGroupItem := range data.SubGroups {
			subGroupsKeys := 0
			for _, v := range data.SubGroups {
				if v.Type != nil && *v.Type == *subGroupItem.Type {
					if len(v.StudentIds) > 0 {
						subGroupsKeys++
					}
				}
			}
			if subGroupsKeys < 2 {
				err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
					_, err = tx.Query(ctx, sqlClassroomStudentsDeleteTypeAny, data.ID, subGroupItem.Type)
					return
				})
				if err != nil {
					utils.LoggerDesc("Query error").Error(err)
					return err
				}
				break
			}
			err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
				_, err = tx.Query(ctx, sqlClassroomStudentsDeleteType, data.ID, subGroupItem.Type, subGroupItem.TypeKey)
				return
			})
			if err != nil {
				utils.LoggerDesc("Query error").Error(err)
				return err
			}
			for _, studentId := range subGroupItem.StudentIds {
				// TODO: delete dublicate students
				err = d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
					_, err = tx.Query(ctx, sqlClassroomStudentsInsert, data.ID, studentId, subGroupItem.Type, subGroupItem.TypeKey)
					return
				})
				if err != nil {
					utils.LoggerDesc("Query error").Error(err)
					return err
				}
			}
		}
	}
	return nil
}

func ClassroomsCreateQuery(m *models.Classroom) (string, []interface{}) {
	args := []interface{}{}
	cols := ""
	vals := ""
	q := ClassroomAtomicQuery(m, true)
	for k, v := range q {
		args = append(args, v)
		cols += ", " + k
		vals += ", $" + strconv.Itoa(len(args))
	}
	qs := sqlClassroomInsert + " (" + strings.Trim(cols, ", ") + ") VALUES (" + strings.Trim(vals, ", ") + ")"
	return qs, args
}

func ClassroomsUpdateQuery(m *models.Classroom) (string, []interface{}) {
	args := []interface{}{}
	sets := ""
	q := ClassroomAtomicQuery(m, false)
	for k, v := range q {
		args = append(args, v)
		sets += ", " + k + "=$" + strconv.Itoa(len(args))
	}
	args = append(args, m.ID)
	qs := strings.ReplaceAll(sqlClassroomUpdate, "set uid=uid", "set uid=uid "+sets+" ") + "where uid=$" + strconv.Itoa(len(args))
	return qs, args
}

func ClassroomAtomicQuery(m *models.Classroom, isCreate bool) map[string]interface{} {
	q := map[string]interface{}{}
	if m.Name != nil {
		q["name"] = m.Name
	}
	if m.NameCanonical != nil {
		q["name_canonical"] = m.NameCanonical
	}
	if m.Avatar != nil {
		q["avatar"] = m.Avatar
	}
	if m.Description != nil {
		q["description"] = m.Description
	}
	if m.Language != nil {
		q["language"] = m.Language
	}
	if m.Level != nil {
		q["level"] = m.Level
	}
	if m.ArchivedAt != nil {
		q["archived_at"] = m.ArchivedAt
	}
	if m.TeacherId != nil {
		q["teacher_uid"] = m.TeacherId
	}
	if m.StudentId != nil {
		q["student_uid"] = m.StudentId
	}
	if m.SchoolId != "" {
		q["school_uid"] = m.SchoolId
	}
	if m.ShiftId != nil {
		q["shift_uid"] = m.ShiftId
	}
	if m.ParentId != nil {
		q["parent_uid"] = m.ParentId
	}
	if m.PeriodId != nil {
		q["period_uid"] = m.PeriodId
	}
	if isCreate {
		q["created_at"] = time.Now()
	}
	q["updated_at"] = time.Now()
	return q
}

// TODO: refactor store vulnerable-style
func ClassroomsListBuildQuery(f models.ClassroomFilterRequest, args []interface{}) (string, []interface{}) {
	var wheres string = ""
	qs := sqlClassroomSelectMany

	if f.ID != nil && *f.ID != "" {
		args = append(args, *f.ID)
		wheres += " and c.uid=$" + strconv.Itoa(len(args))
	}
	if f.Ids != nil && len(*f.Ids) >= 0 {
		args = append(args, *f.Ids)
		wheres += " and c.uid=ANY($" + strconv.Itoa(len(args)) + "::uuid[])"
	}
	if f.ShiftId != nil {
		args = append(args, *f.ShiftId)
		wheres += " and c.shift_uid=$" + strconv.Itoa(len(args))
	}
	if f.TeacherId != nil {
		args = append(args, *f.TeacherId)
		wheres += " and c.teacher_uid=$" + strconv.Itoa(len(args))
	}
	if f.SchoolId != nil {
		args = append(args, f.SchoolId)
		wheres += " and c.school_uid=$" + strconv.Itoa(len(args))
	}
	if f.SchoolIds != nil {
		args = append(args, *f.SchoolIds)
		wheres += " and c.school_uid = ANY($" + strconv.Itoa(len(args)) + "::uuid[])"
	}
	if f.ParentId != nil {
		args = append(args, f.ParentId)
		wheres += " and c.parent_uid=$" + strconv.Itoa(len(args))
	}
	if f.Name != nil {
		args = append(args, f.Name)
		wheres += " and c.name=$" + strconv.Itoa(len(args))
	}
	if f.ExamsCountBetween != nil && *f.ExamsCountBetween != "" {
		args = append(args, *f.ExamsCountBetween)
		wheres += " and (select count(*) from subject_exams se where se.classroom_uid = c.uid) = $" + strconv.Itoa(len(args))
	}
	if f.Search != nil {
		*f.Search = strings.ToLower(*f.Search)
		args = append(args, f.Search)
		wheres += " and (lower(name) like '%' || $" + strconv.Itoa(len(args)) + " || '%' or lower(name_canonical) like '%' || $" + strconv.Itoa(len(args)) +
			" || '%' or lower(description) like '%' || $" + strconv.Itoa(len(args)) + " || '%')"
	}
	wheres += " group by c.uid"
	if f.Sort != nil {
		dir := "desc"
		if strings.HasSuffix(*f.Sort, "~") {
			dir = "asc"
		}
		*f.Sort = strings.ReplaceAll(*f.Sort, "~", "")
		args = append(args, f.Sort)
		wheres += " order by $" + strconv.Itoa(len(args)) + " " + dir
	} else {
		wheres += " order by LPAD(lower(c.name), 3, '0') asc"
	}
	qs = strings.ReplaceAll(qs, "c.uid=c.uid", "c.uid=c.uid "+wheres+" ")
	return qs, args
}

func (d *PgxStore) ClassroomsLoadRelations(ctx context.Context, l *[]*models.Classroom, isDetail bool) error {
	ids := []string{}
	for _, m := range *l {
		ids = append(ids, m.ID)
	}
	if len(ids) < 1 {
		return nil
	}
	// load schools
	err := d.ClassroomsLoadSchool(ctx, l)
	if err != nil {
		return err
	}
	// load shifts
	if rs, err := d.ClassroomsLoadShift(ctx, ids); err == nil {
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
	// load teacher
	if rs, err := d.ClassroomsLoadTeacher(ctx, ids); err == nil {
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
	// load student
	if rs, err := d.ClassroomsLoadStudent(ctx, ids); err == nil {
		for _, r := range rs {
			for _, m := range *l {
				if r.ID == m.ID {
					m.Student = r.Relation
				}
			}
		}
	} else {
		return err
	}
	// load parent
	if rs, err := d.ClassroomsLoadParent(ctx, ids); err == nil {
		for _, r := range rs {
			for _, m := range *l {
				if r.ID == m.ID {
					m.Parent = r.Relation
				}
			}
		}
	} else {
		return err
	}
	if rs, err := d.ClassroomsLoadPeriod(ctx, ids); err == nil {
		for _, r := range rs {
			for _, m := range *l {
				if r.ID == m.ID {
					m.Period = r.Relation
				}
			}
		}
	} else {
		return err
	}
	// load subgroups
	if rs, err := d.ClassroomsLoadStudents(ctx, ids); err == nil {
		for _, r := range rs {
			for _, m := range *l {
				if r.ID == m.ID && r.Type != nil && r.TypeKey != nil {
					if m.SubGroups == nil {
						m.SubGroups = []models.ClassroomStudentsByType{}
					}
					tk := -1
					for k, v := range m.SubGroups {
						if *v.Type == *r.Type && *v.TypeKey == *r.TypeKey {
							tk = k
						}
					}
					if tk == -1 {
						m.SubGroups = append(m.SubGroups, models.ClassroomStudentsByType{Type: r.Type, TypeKey: r.TypeKey, StudentIds: []string{}})
						tk = len(m.SubGroups) - 1
					}
					for _, v := range r.Relation {
						m.SubGroups[tk].StudentIds = append(m.SubGroups[tk].StudentIds, v.ID)
					}
				}
			}
		}
	}
	if isDetail {
		// load students
		if rs, err := d.ClassroomsLoadStudents(ctx, ids); err == nil {
			for _, r := range rs {
				for _, m := range *l {
					if r.ID == m.ID && r.Type == nil && r.TypeKey == nil {
						m.Students = r.Relation
					}
				}
			}
		} else {
			return err
		}
		// load subjects
		if rs, err := d.ClassroomsLoadSubjects(ctx, ids); err == nil {
			for _, r := range rs {
				// TODO: change load method
				err = d.SubjectsLoadRelations(ctx, &r.Relation, true)
				if err != nil {
					return err
				}
				for _, m := range *l {
					if r.ID == m.ID {
						m.Subjects = r.Relation
					}
				}
			}
		} else {
			return err
		}
	}
	return nil
}

type ClassroomsLoadSchoolItem struct {
	ID       string
	Relation *models.School
}

func (d *PgxStore) ClassroomsLoadSchool(ctx context.Context, l *[]*models.Classroom) error {
	ids := []string{}
	for _, m := range *l {
		ids = append(ids, m.ID)
	}
	if len(ids) < 1 {
		return nil
	}
	// load
	var schoolParents []*models.School
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		rows, err := tx.Query(ctx, sqlClassroomSchool, (ids))

		for rows.Next() {
			sub := models.School{}
			pid := ""
			err = scanSchool(rows, &sub, &pid)
			if err != nil {
				return err
			}
			// res = append(res, ClassroomsLoadSchoolItem{ID: pid, Relation: &sub})
			for _, m := range *l {
				if (m.ID) == pid {
					schoolParents = append(schoolParents, &sub)
					m.School = &sub
				}
			}
		}

		return
	})
	err = d.SchoolsLoadParents(ctx, &schoolParents)
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return err
	}
	return nil
}

type ClassroomsLoadShiftItem struct {
	ID       string
	Relation *models.Shift
}

func (d *PgxStore) ClassroomsLoadShift(ctx context.Context, ids []string) ([]ClassroomsLoadShiftItem, error) {
	res := []ClassroomsLoadShiftItem{}
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		rows, err := tx.Query(ctx, sqlClassroomShift, (ids))
		for rows.Next() {
			sub := models.Shift{}
			pid := ""
			err = scanShift(rows, &sub, &pid)
			if err != nil {
				return err
			}
			res = append(res, ClassroomsLoadShiftItem{ID: pid, Relation: &sub})
		}
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}

	return res, nil
}

type ClassroomsLoadTeacherItem struct {
	ID       string
	Relation *models.User
}

func (d *PgxStore) ClassroomsLoadTeacher(ctx context.Context, ids []string) ([]ClassroomsLoadTeacherItem, error) {
	res := []ClassroomsLoadTeacherItem{}
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		rows, err := tx.Query(ctx, sqlClassroomTeacher, (ids))
		for rows.Next() {
			sub := models.User{}
			pid := ""
			err = scanUser(rows, &sub, &pid)
			if err != nil {
				return err
			}
			res = append(res, ClassroomsLoadTeacherItem{ID: pid, Relation: &sub})
		}
		return
	})
	if err != nil {
		return nil, err
	}
	return res, nil
}

type ClassroomsLoadStudentItem struct {
	ID       string
	Relation *models.User
}

func (d *PgxStore) ClassroomsLoadStudent(ctx context.Context, ids []string) ([]ClassroomsLoadStudentItem, error) {
	res := []ClassroomsLoadStudentItem{}
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		rows, err := tx.Query(ctx, sqlClassroomStudent, (ids))
		for rows.Next() {
			sub := models.User{}
			pid := ""
			err := scanUser(rows, &sub, &pid)
			if err != nil {
				return err
			}
			res = append(res, ClassroomsLoadStudentItem{ID: pid, Relation: &sub})
		}
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}
	return res, nil
}

type ClassroomsLoadParentItem struct {
	ID       string
	Relation *models.Classroom
}

func (d *PgxStore) ClassroomsLoadParent(ctx context.Context, ids []string) ([]ClassroomsLoadParentItem, error) {
	res := []ClassroomsLoadParentItem{}
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		rows, err := tx.Query(ctx, sqlClassroomParent, (ids))
		for rows.Next() {
			sub := models.Classroom{}
			pid := ""
			err := scanClassroom(rows, &sub, &pid)
			if err != nil {
				return err
			}
			res = append(res, ClassroomsLoadParentItem{ID: pid, Relation: &sub})
		}
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}
	return res, nil
}

type ClassroomsLoadStudentsItem struct {
	ID       string
	Type     *string
	TypeKey  *int
	Relation []*models.User
}

func (d *PgxStore) ClassroomsLoadStudents(ctx context.Context, ids []string) ([]ClassroomsLoadStudentsItem, error) {
	res := []ClassroomsLoadStudentsItem{}
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		rows, err := tx.Query(ctx, sqlClassroomStudents, ids)
		for rows.Next() {
			sub := models.User{}
			clType := new(string)
			clTypeKey := new(int)
			pid := ""
			err := scanUser(rows, &sub, &clType, &clTypeKey, &pid)
			if err != nil {
				utils.LoggerDesc("Scan error").Error(err)
				return err
			}
			resK := -1
			for k, v := range res {
				if v.ID == pid && v.Type == nil && v.TypeKey == nil && clType == nil && clTypeKey == nil {
					resK = k
				} else if v.ID == pid && v.Type != nil && v.TypeKey != nil && clType != nil && clTypeKey != nil &&
					*v.Type == *clType && *v.TypeKey == *clTypeKey {
					resK = k
				}
			}
			if resK == -1 {
				res = append(res, ClassroomsLoadStudentsItem{ID: pid, Type: clType, TypeKey: clTypeKey, Relation: []*models.User{}})
				resK = len(res) - 1
			}
			res[resK].Type = clType
			res[resK].TypeKey = clTypeKey
			res[resK].Relation = append(res[resK].Relation, &sub)
		}
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}
	return res, nil
}

type ClassroomsLoadSubjectsItem struct {
	ID       string
	Relation []*models.Subject
}

func (d *PgxStore) ClassroomsLoadSubjects(ctx context.Context, ids []string) ([]ClassroomsLoadSubjectsItem, error) {
	res := []ClassroomsLoadSubjectsItem{}
	for _, v := range ids {
		res = append(res, ClassroomsLoadSubjectsItem{ID: v, Relation: []*models.Subject{}})
	}
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		rows, err := tx.Query(ctx, sqlClassroomSubjects, (ids))
		for rows.Next() {
			sub := models.Subject{}
			pid := ""
			err = scanSubject(rows, &sub, &pid)
			if err != nil {
				utils.LoggerDesc("Scan error").Error(err)
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

func (store *PgxStore) GetClassroomIdByName(
	ctx context.Context,
	dto models.GetClassroomIdByNameQueryDto,
) (*string, error) {
	stmt := `
		SELECT uid
		FROM classrooms 
		WHERE 
			school_uid = $1
			AND name = $2
	`
	var id string
	err := store.runQuery(ctx, func(conn *pgxpool.Conn) (err error) {
		row := conn.QueryRow(
			ctx,
			stmt,
			dto.SchoolId,
			dto.Name,
		)
		err = row.Scan(&id)
		if err != nil {
			return err
		}
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}

	return &id, nil
}

type ClassroomsLoadPeriodItem struct {
	ID       string
	Relation *models.Period
}

func (d *PgxStore) ClassroomsLoadPeriod(ctx context.Context, ids []string) ([]ClassroomsLoadPeriodItem, error) {
	res := []ClassroomsLoadPeriodItem{}

	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		rows, err := tx.Query(ctx, sqlClassroomPeriod, (ids))
		for rows.Next() {
			sub := models.Period{}
			pid := ""
			err = scanPeriod(rows, &sub, &pid)
			if err != nil {
				return err
			}
			res = append(res, ClassroomsLoadPeriodItem{ID: pid, Relation: &sub})
		}
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}
	return res, nil
}
