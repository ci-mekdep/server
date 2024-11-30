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

const (
	sqlAssignmentFields     = "a.uid, a.lesson_uid, a.title, a.files, a.updated_at, a.created_at, a.updated_by_uid, a.created_by_uid"
	sqlAssignmentSelect     = `select ` + sqlAssignmentFields + `  from assignments a where a.uid = ANY($1::uuid[])`
	sqlAssignmentSelectMany = `select ` + sqlAssignmentFields + `, count(*) over() as total  from assignments a where uid=uid limit $1 offset $2`
	sqlAssignmentUpdate     = "update assignments a set uid=uid"
	sqlAssignmentInsert     = "insert into assignments"
	sqlAssignmentDelete     = "delete from assignments where uid = ANY($1::uuid[])"
)

func scanAssignment(rows pgx.Row, m *models.Assignment, addColumns ...interface{}) (err error) {
	err = rows.Scan(parseColumnsForScan(m, addColumns...)...)
	return
}

func (d *PgxStore) AssignmentUpdateOrCreate(ctx context.Context, data models.Assignment) (models.Assignment, error) {
	m, err := d.AssignmentFindOrCreate(ctx, data)
	if err != nil {
		return models.Assignment{}, err
	}
	data.ID = m.ID
	return d.AssignmentUpdate(ctx, data)
}

func (d *PgxStore) AssignmentsFindBy(ctx context.Context, f models.AssignmentFilterRequest) ([]models.Assignment, int, error) {
	if f.Limit == nil {
		f.Limit = new(int)
		*f.Limit = 1000
	}
	if f.Offset == nil {
		f.Offset = new(int)
		*f.Offset = 0
	}
	args := []interface{}{f.Limit, f.Offset}
	var qs string
	qs, args = AssignmentsListBuildQuery(f, args)

	l := []models.Assignment{}
	var total int
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		rows, err := tx.Query(ctx, qs, args...)
		for rows.Next() {
			sub := models.Assignment{}
			err = scanAssignment(rows, &sub, &total)
			if err != nil {
				utils.LoggerDesc("Scan error").Error(err)
				return err
			}
			l = append(l, sub)
		}
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, 0, err
	}

	return l, total, nil
}

func (d *PgxStore) AssignmentsFindByIds(ctx context.Context, ids []string) ([]models.Assignment, error) {
	l := []models.Assignment{}

	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		rows, err := tx.Query(ctx, sqlAssignmentSelect, (ids))
		for rows.Next() {
			m := models.Assignment{}
			err := scanAssignment(rows, &m)
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

func (d *PgxStore) AssignmentFindById(ctx context.Context, id string) (models.Assignment, error) {
	l, err := d.AssignmentsFindByIds(ctx, []string{id})
	if err != nil {
		return models.Assignment{}, err
	}
	if len(l) < 1 {
		err = pgx.ErrNoRows
		utils.LoggerDesc("Scan error").Error(err)
		return models.Assignment{}, err
	}
	return l[0], nil
}

func (d *PgxStore) AssignmentUpdate(ctx context.Context, data models.Assignment) (models.Assignment, error) {
	qs, args := AssignmentUpdateQuery(data)

	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		_, err = tx.Query(ctx, qs, args...)
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return models.Assignment{}, err
	}

	editModel, err := d.AssignmentFindById(ctx, data.ID)
	if err != nil {
		return models.Assignment{}, err
	}
	return editModel, nil
}

func (d *PgxStore) AssignmentFindOrCreate(ctx context.Context, data models.Assignment) (models.Assignment, error) {
	// fetch lesson
	lesson, err := d.LessonsFindById(ctx, data.LessonId)
	if err != nil || lesson.ID == "" {
		return models.Assignment{}, err
	}

	// try to fetch assignment
	qs := sqlAssignmentSelectMany
	qs = strings.ReplaceAll(qs, "uid=uid", "a.lesson_uid=$3")
	var row pgx.Row
	err = d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		row = tx.QueryRow(ctx, qs, 1, 0, data.LessonId)
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return models.Assignment{}, err
	}
	m := models.Assignment{}
	total := 0
	err = row.Scan(parseColumnsForScan(&m, &total)...)
	// if not exists then create
	if err != nil {
		if !strings.Contains(err.Error(), "no rows") {
			utils.LoggerDesc("Scan error").Error(err)
			return models.Assignment{}, err
		}
		qs, args := AssignmentCreateQuery(data)
		qs += " RETURNING " + strings.ReplaceAll(sqlAssignmentFields, "a.", "")

		err = d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
			err = scanAssignment(tx.QueryRow(ctx, qs, args...),
				&m)
			return
		})
		if err != nil {
			utils.LoggerDesc("Query error").Error(err)
			return models.Assignment{}, err
		}
	}
	return m, nil
}

func AssignmentCreateQuery(m models.Assignment) (string, []interface{}) {
	args := []interface{}{}
	cols := ""
	vals := ""
	if m.Title == "" {
		m.Title = " "
	}
	q := AssignmentAtomicQuery(m, true)
	for k, v := range q {
		args = append(args, v)
		cols += ", " + k
		vals += ", $" + strconv.Itoa(len(args))
	}
	qs := sqlAssignmentInsert + " (" + strings.Trim(cols, ", ") + ") VALUES (" + strings.Trim(vals, ", ") + ")"
	return qs, args
}

func AssignmentUpdateQuery(m models.Assignment) (string, []interface{}) {
	args := []interface{}{}
	sets := ""
	q := AssignmentAtomicQuery(m, false)
	for k, v := range q {
		args = append(args, v)
		sets += ", " + k + "=$" + strconv.Itoa(len(args))
	}
	args = append(args, m.ID)
	qs := strings.ReplaceAll(sqlAssignmentUpdate, "set uid=uid", "set uid=uid "+sets+" ") + "where uid=$" + strconv.Itoa(len(args))
	return qs, args
}

func AssignmentAtomicQuery(m models.Assignment, isCreate bool) map[string]interface{} {
	q := map[string]interface{}{}
	if m.LessonId != "" {
		q["lesson_uid"] = m.LessonId
	}
	q["title"] = m.Title
	if m.Files != nil {
		q["files"] = *m.Files
	}
	if isCreate {
		q["created_at"] = time.Now()
		// todo: current user
		q["created_by_uid"] = m.CreatedBy
	}
	q["updated_by_uid"] = m.UpdatedBy
	q["updated_at"] = time.Now()
	return q
}

func AssignmentsListBuildQuery(f models.AssignmentFilterRequest, args []interface{}) (string, []interface{}) {
	var wheres string = ""

	if f.ID != nil && *f.ID != "" {
		args = append(args, *f.ID)
		wheres += " and a.uid=$" + strconv.Itoa(len(args))
	}
	if f.LessonID != nil && *f.LessonID != "" {
		args = append(args, *f.LessonID)
		wheres += " and a.lesson_uid=$" + strconv.Itoa(len(args))
	}
	if f.LessonIDs != nil {
		args = append(args, *f.LessonIDs)
		wheres += " and a.lesson_uid = ANY($" + strconv.Itoa(len(args)) + "::uuid[])"
	}
	if f.Search != nil && *f.Search != "" {
		*f.Search = strings.ToLower(*f.Search)
		args = append(args, *f.Search)
		wheres += " and (lower(a.title) like '%' || $" + strconv.Itoa(len(args)) +
			" || '%' or lower(a.content) like '%' || $" + strconv.Itoa(len(args)) +
			" || '%' or lower(a.files) like '%' || $" + strconv.Itoa(len(args)) + " || '%')"
	}
	wheres += " group by a.uid "
	if f.Sort != nil && *f.Sort != "" {
		args = append(args, *f.Sort)
		dir := "desc"
		if strings.HasSuffix(*f.Sort, "~") {
			dir = "asc"
		}
		*f.Sort = strings.ReplaceAll(*f.Sort, "~", "")
		wheres += " order by $" + strconv.Itoa(len(args)) + " " + dir
	} else {
		wheres += " order by uid desc"
	}
	qs := sqlAssignmentSelectMany
	qs = strings.ReplaceAll(qs, "uid=uid", "uid=uid "+wheres+" ")
	return qs, args
}
