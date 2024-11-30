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

const sqlUserLogFields = sqlUserLogFieldsNoRelation + ``
const sqlUserLogFieldsNoRelation = `ul.uid, ul.school_uid, ul.user_uid, ul.session_uid, ul.subject_uid, ul.subject, ul.subject_action, ul.subject_description, ul.properties, ul.created_at`
const sqlUserLogSelect = `SELECT ` + sqlUserLogFields + ` FROM user_logs ul WHERE ul.uid = ANY($1::uuid[])`
const sqlUserLogSelectMany = `SELECT ` + sqlUserLogFieldsNoRelation + `, 9999999 as total FROM user_logs ul
	left join sessions ss on ss.uid=ul.session_uid
	where ul.uid=ul.uid limit $1 offset $2 `
const sqlUserLogUpdate = `UPDATE user_logs ul set uid=uid`
const sqlUserLogInsert = `INSERT INTO user_logs`
const sqlUserLogDelete = `delete from user_logs ul where uid = ANY($1::uuid[])`

const sqlUserLogSchool = `select ` + sqlSchoolFields + `, ul.uid from user_logs ul
	right join schools s on (s.uid=ul.school_uid) where ul.uid = ANY($1::uuid[])`
const sqlUserLogUser = `select ` + sqlUserFields + `, ul.uid from user_logs ul
right join users u on (u.uid=ul.user_uid) where ul.uid = ANY($1::uuid[])`
const sqlUserLogSession = `select ` + sqlSessionFields + `, ul.uid from user_logs ul
right join sessions ss on (ss.uid=ul.session_uid) where ul.uid = ANY($1::uuid[])`

func scanUserLog(rows pgx.Row, m *models.UserLog, addColumns ...interface{}) (err error) {
	err = rows.Scan(parseColumnsForScan(m, addColumns...)...)
	return
}

func (d *PgxStore) UserLogsFindBy(ctx context.Context, f models.UserLogFilterRequest) (user_logs []*models.UserLog, total int, err error) {
	args := []interface{}{f.Limit, f.Offset}
	qs, args := UserLogsListBuildQuery(f, args)
	err = d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		rows, err := tx.Query(ctx, qs, args...)
		for rows.Next() {
			user_log := models.UserLog{}
			err = scanUserLog(rows, &user_log, &total)
			if err != nil {
				utils.LoggerDesc("Scan error").Error(err)
				return err
			}
			user_logs = append(user_logs, &user_log)
		}
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, 0, err
	}
	return user_logs, total, nil
}

func (d *PgxStore) UserLogsFindById(ctx context.Context, Id string) (*models.UserLog, error) {
	row, err := d.UserLogsFindByIds(ctx, []string{Id})
	if err != nil {
		return nil, err
	}
	if len(row) < 1 {
		return nil, errors.New("user_log not found by uid: " + Id)
	}
	return row[0], nil
}

func (d *PgxStore) UserLogsFindByIds(ctx context.Context, Ids []string) ([]*models.UserLog, error) {
	items := []*models.UserLog{}
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		rows, err := tx.Query(ctx, sqlUserLogSelect, (Ids))
		for rows.Next() {
			item := models.UserLog{}
			err := scanUserLog(rows, &item)
			if err != nil {
				utils.LoggerDesc("Scan error").Error(err)
				return err
			}
			items = append(items, &item)
		}
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}

	return items, nil
}

func (d *PgxStore) UserLogsUpdate(ctx context.Context, model models.UserLog) (*models.UserLog, error) {
	qs, args := UserLogUpdateQuery(model)
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		_, err = tx.Query(ctx, qs, args...)
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}
	editModel, err := d.UserLogsFindById(ctx, model.ID)
	if err != nil {
		return nil, err
	}
	return editModel, nil
}

func (d *PgxStore) UserLogsCreate(ctx context.Context, model models.UserLog) (*models.UserLog, error) {
	qs, args := UserLogCreateQuery(model)
	qs += " RETURNING uid"
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		err = tx.QueryRow(ctx, qs, args...).Scan(&model.ID)
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}
	editModel, err := d.UserLogsFindById(ctx, model.ID)
	if err != nil {
		return nil, err
	}
	return editModel, nil
}

func (d *PgxStore) UserLogsDelete(ctx context.Context, items []*models.UserLog) ([]*models.UserLog, error) {
	ids := []string{}
	for _, i := range items {
		ids = append(ids, i.ID)
	}
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		_, err = tx.Query(ctx, sqlUserLogDelete, (ids))
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}
	return items, nil
}

func UserLogCreateQuery(m models.UserLog) (string, []interface{}) {
	args := []interface{}{}
	cols := ""
	vals := ""
	q := UserLogAtomicQuery(m, true)
	for k, v := range q {
		args = append(args, v)
		cols += ", " + k
		vals += ", $" + strconv.Itoa(len(args))
	}
	qs := sqlUserLogInsert + " (" + strings.Trim(cols, ", ") + ") VALUES (" + strings.Trim(vals, ", ") + ")"
	return qs, args
}

func UserLogUpdateQuery(m models.UserLog) (string, []interface{}) {
	args := []interface{}{}
	sets := ""
	q := UserLogAtomicQuery(m, false)
	for k, v := range q {
		args = append(args, v)
		sets += ", " + k + "=$" + strconv.Itoa(len(args))
	}
	args = append(args, m.ID)
	qs := strings.ReplaceAll(sqlUserLogUpdate, "set uid=uid", "set uid=uid "+sets+" ") + "where uid=$" + strconv.Itoa(len(args))
	return qs, args
}

func UserLogAtomicQuery(m models.UserLog, isCreate bool) map[string]interface{} {
	q := map[string]interface{}{}
	if m.SchoolId != nil && *m.SchoolId != "" {
		q["school_uid"] = m.SchoolId
	}
	if m.UserId != "" {
		q["user_uid"] = m.UserId
	}
	if m.SessionId != nil && *m.SessionId != "" {
		q["session_uid"] = m.SessionId
	}
	if m.SubjectId != nil && *m.SubjectId != "" {
		q["subject_uid"] = m.SubjectId
	}
	if m.Subject != "" {
		q["subject"] = m.Subject
	}
	if m.SubjectAction != "" {
		q["subject_action"] = m.SubjectAction
	}
	if m.SubjectDescription != nil {
		q["subject_description"] = *m.SubjectDescription
	}
	if m.SubjectProperties != nil {
		q["properties"] = m.SubjectProperties
	}
	if isCreate {
		q["created_at"] = time.Now()
	}
	return q
}

func UserLogsListBuildQuery(f models.UserLogFilterRequest, args []interface{}) (string, []interface{}) {
	var wheres string = ""

	if f.ID != nil && *f.ID != "" {
		args = append(args, *f.ID)
		wheres += " and ul.uid=$" + strconv.Itoa(len(args))
	}
	if f.SchoolIds != nil {
		args = append(args, *f.SchoolIds)
		wheres += " and ul.school_uid = ANY($" + strconv.Itoa(len(args)) + "::uuid[])"
	}
	if f.SchoolId != nil && *f.SchoolId != "" {
		args = append(args, *f.SchoolId)
		wheres += " and ul.school_uid=$" + strconv.Itoa(len(args))
	}
	if f.UserId != nil && *f.UserId != "" {
		args = append(args, *f.UserId)
		wheres += " and ul.user_uid=$" + strconv.Itoa(len(args))
	}
	if f.SessionId != nil && *f.SessionId != "" {
		args = append(args, *f.SessionId)
		wheres += " and ul.session_uid=$" + strconv.Itoa(len(args))
	}
	if f.SubjectName != nil && *f.SubjectName != "" {
		args = append(args, *f.SubjectName)
		wheres += " and ul.subject=$" + strconv.Itoa(len(args))
	}
	if f.SubjectAction != nil && *f.SubjectAction != "" {
		args = append(args, *f.SubjectAction)
		wheres += " and ul.subject_action=$" + strconv.Itoa(len(args))
	}
	if f.SubjectId != nil && *f.SubjectId != "" {
		args = append(args, *f.SubjectId)
		wheres += " and ul.subject_uid=$" + strconv.Itoa(len(args))
	}
	if f.Ip != nil && *f.Ip != "" {
		args = append(args, *f.Ip)
		wheres += " and ss.ip=$" + strconv.Itoa(len(args))
	}
	if f.StartDate != nil && f.EndDate != nil {
		args = append(args, *f.StartDate, *f.EndDate)
		wheres += " and ul.created_at BETWEEN $" + strconv.Itoa(len(args)) + " AND $" + strconv.Itoa(len(args))
	}
	if f.StartDate != nil {
		args = append(args, *f.StartDate)
		wheres += " and ul.created_at >= $" + strconv.Itoa(len(args))
	}
	if f.EndDate != nil {
		args = append(args, *f.EndDate)
		wheres += " and ul.created_at <= $" + strconv.Itoa(len(args))
	}
	if f.Search != nil && *f.Search != "" && f.SearchType != nil {
		*f.Search = strings.ToLower(*f.Search)
		args = append(args, *f.Search)

		switch *f.SearchType {
		case "ip":
			wheres += " and ss.ip = $" + strconv.Itoa(len(args))
		case "id":
			wheres += " and ul.subject_uid = $" + strconv.Itoa(len(args))
		case "agent":
			wheres += " and ss.agent LIKE '%' || $" + strconv.Itoa(len(args)) + " || '%'"
		case "properties":
			wheres += " and ul.subject_properties LIKE '%' || $" + strconv.Itoa(len(args)) + " || '%'"
		case "subject_description":
			wheres += " and ul.subject_description LIKE '%' || $" + strconv.Itoa(len(args)) + " || '%'"
		}
	}
	wheres += " group by ul.uid "
	wheres += " order by ul.created_at desc"

	qs := sqlUserLogSelectMany
	qs = strings.ReplaceAll(qs, "ul.uid=ul.uid", "ul.uid=ul.uid "+wheres+" ")
	return qs, args
}

func (d *PgxStore) UserLogsLoadRelations(ctx context.Context, l *[]*models.UserLog) error {
	ids := []string{}
	for _, m := range *l {
		ids = append(ids, m.ID)
	}
	if len(ids) < 1 {
		return nil
	}

	// load school
	if rs, err := d.UserLogLoadSchool(ctx, ids); err == nil {
		var schoolParents []*models.School
		for _, r := range rs {
			for _, m := range *l {
				if r.Id == m.ID {
					schoolParents = append(schoolParents, r.Relation)
					m.School = r.Relation
				}
			}
		}
		err = d.SchoolsLoadParents(ctx, &schoolParents)
	} else {
		return err
	}
	// load user
	if rs, err := d.UserLogLoadUser(ctx, ids); err == nil {
		for _, r := range rs {
			for _, m := range *l {
				if r.Id == m.ID {
					m.User = r.Relation
				}
			}
		}
	} else {
		return err
	}
	// load session
	if rs, err := d.UserLogLoadSession(ctx, ids); err == nil {
		for _, r := range rs {
			for _, m := range *l {
				// TODO: NEED TO COME HERE AND FIX
				if r.Id == "" {
					m.Session = r.Relation
				}
			}
		}
	} else {
		return err
	}
	// load relation
	if err := d.UserLogLoadSubject(ctx, *l); err != nil {
		return err
	}
	return nil
}

type UserLogLoadSchoolItem struct {
	Id       string
	Relation *models.School
}

func (d *PgxStore) UserLogLoadSchool(ctx context.Context, ids []string) ([]UserLogLoadSchoolItem, error) {
	res := []UserLogLoadSchoolItem{}
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		rows, err := tx.Query(ctx, sqlUserLogSchool, (ids))
		for rows.Next() {
			sub := models.School{}
			pid := ""
			err = scanSchool(rows, &sub, &pid)
			if err != nil {
				utils.LoggerDesc("Scan error").Error(err)
				return err
			}
			res = append(res, UserLogLoadSchoolItem{Id: pid, Relation: &sub})
		}
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}
	return res, nil
}

type UserLogLoadUserItem struct {
	Id       string
	Relation *models.User
}

func (d *PgxStore) UserLogLoadUser(ctx context.Context, ids []string) ([]UserLogLoadUserItem, error) {
	res := []UserLogLoadUserItem{}
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		rows, err := tx.Query(ctx, sqlUserLogUser, (ids))
		for rows.Next() {
			sub := models.User{}
			pid := ""
			err = scanUser(rows, &sub, &pid)
			if err != nil {
				utils.LoggerDesc("Scan error").Error(err)
				return err
			}
			res = append(res, UserLogLoadUserItem{Id: pid, Relation: &sub})
		}
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}
	return res, nil
}

type UserLogLoadSessionItem struct {
	Id       string
	Relation *models.Session
}

func (d *PgxStore) UserLogLoadSession(ctx context.Context, ids []string) ([]UserLogLoadSessionItem, error) {
	res := []UserLogLoadSessionItem{}
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		rows, err := tx.Query(ctx, sqlUserLogSession, (ids))
		for rows.Next() {
			sub := models.Session{}
			pid := string(0)
			err = scanSession(rows, &sub, &pid)
			if err != nil {
				utils.LoggerDesc("Scan error").Error(err)
				return err
			}
			res = append(res, UserLogLoadSessionItem{Id: pid, Relation: &sub})
		}
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}
	return res, nil
}

func (d *PgxStore) UserLogLoadSubject(ctx context.Context, l []*models.UserLog) error {
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		for k, v := range l {
			var qs string
			var sub interface{}
			// var scanFunc func
			if v.Subject == models.LogSubjectUsers && v.SubjectId != nil && *v.SubjectId != "" {
				qs = `select ` + sqlUserFields + ` from users u where uid=$1`
				item := models.User{}
				err = scanUser(tx.QueryRow(ctx, qs, v.SubjectId), &item)
				if err != nil {
					utils.LoggerDesc("Query error").Error(err)
					return err
				}
				sub = item
			}
			if v.Subject == models.LogSubjectSchools && v.SubjectId != nil && *v.SubjectId != "" {
				qs = `select ` + sqlSchoolFields + ` from schools s where uid=$1`
				item := models.School{}
				err = scanSchool(tx.QueryRow(ctx, qs, v.SubjectId), &item)
				if err != nil {
					utils.LoggerDesc("Query error").Error(err)
					return err
				}
				sub = item
			}
			if v.Subject == models.LogSubjectClassrooms && v.SubjectId != nil && *v.SubjectId != "" {
				qs = `select ` + sqlClassroomFields + ` from classrooms c where uid=$1`
				item := models.Classroom{}
				err = scanClassroom(tx.QueryRow(ctx, qs, v.SubjectId), &item)
				if err != nil {
					utils.LoggerDesc("Query error").Error(err)
					return err
				}
				sub = item
			}
			if v.Subject == models.LogSubjectGrade && v.SubjectId != nil && *v.SubjectId != "" {
				qs = `select ` + sqlGradeFields + ` from grades g where uid=$1`
				item := models.Grade{}
				err = scanGrade(tx.QueryRow(ctx, qs, v.SubjectId), &item)
				if err != nil {
					utils.LoggerDesc("Query error").Error(err)
					return err
				}
				sub = item
			}
			if v.Subject == models.LogSubjectAbsent && v.SubjectId != nil && *v.SubjectId != "" {
				qs = `select ` + sqlAbsentFields + ` from absents ab where uid=$1`
				item := models.Absent{}
				err = scanAbsent(tx.QueryRow(ctx, qs, v.SubjectId), &item)
				if err != nil {
					utils.LoggerDesc("Query error").Error(err)
					return err
				}
				sub = item
			}
			if v.Subject == models.LogSubjectTimetable && v.SubjectId != nil && *v.SubjectId != "" {
				qs = `select ` + sqlTimetableFields + ` from timetables tt where uid=$1`
				item := models.Timetable{}
				err = scanTimetable(tx.QueryRow(ctx, qs, v.SubjectId), &item)
				if err != nil {
					utils.LoggerDesc("Query error").Error(err)
					return err
				}
				sub = item
			}
			if v.Subject == models.LogSubjectSubject && v.SubjectId != nil && *v.SubjectId != "" {
				qs = `select ` + sqlSubjectFields + ` from subjects sb where uid=$1`
				item := models.Subject{}
				err = scanSubject(tx.QueryRow(ctx, qs, v.SubjectId), &item)
				if err != nil {
					utils.LoggerDesc("Query error").Error(err)
					return err
				}
				sub = item
			}
			if v.Subject == models.LogSubjectShift && v.SubjectId != nil && *v.SubjectId != "" {
				qs = `select ` + sqlShiftFields + ` from shifts sh where uid=$1`
				item := models.Shift{}
				err = scanShift(tx.QueryRow(ctx, qs, v.SubjectId), &item)
				if err != nil {
					utils.LoggerDesc("Query error").Error(err)
					return err
				}
				sub = item
			}
			if v.Subject == models.LogSubjectPeriod && v.SubjectId != nil && *v.SubjectId != "" {
				qs = `select ` + sqlPeriodFields + ` from periods p where uid=$1`
				item := models.Period{}
				err = scanPeriod(tx.QueryRow(ctx, qs, v.SubjectId), &item)
				if err != nil {
					utils.LoggerDesc("Query error").Error(err)
					return err
				}
				sub = item
			}
			l[k].SubjectModel = sub
		}
		return
	})
	if err != nil && err != pgx.ErrNoRows {
		return err
	}
	return nil
}
