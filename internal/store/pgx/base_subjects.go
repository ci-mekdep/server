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

const sqlBaseSubjectsFields = `bs.uid, bs.school_uid, bs.name, bs.category, bs.price, bs.exam_min_grade, bs.age_category, bs.is_available, bs.updated_at, bs.created_at`

const sqlBaseSubjectsInsert = `insert into base_subjects`

const sqlBaseSubjectsUpdate = `update base_subjects set uid=uid`
const sqlBaseSubjectsDelete = `DELETE FROM base_subjects b WHERE uid = ANY($1::uuid[])`

const sqlBaseSubjectsSelect = `select ` + sqlBaseSubjectsFields + ` from base_subjects bs where bs.uid = ANY($1::uuid[])`

const sqlBaseSubjectsSelectMany = `select ` + sqlBaseSubjectsFields + `, count(*) over() as total from base_subjects bs where bs.uid=bs.uid limit $1 offset $2 `

const sqlBaseSubjectsSchool = `select ` + sqlSchoolFields + `, bs.uid from base_subjects bs
	right join schools s on (s.uid=bs.school_uid) where bs.uid = ANY($1::uuid[])`

func scanBaseSubjects(rows pgx.Rows, m *models.BaseSubjects, addColumns ...interface{}) (err error) {
	err = rows.Scan(parseColumnsForScan(m, addColumns...)...)
	return
}

func (d *PgxStore) BaseSubjectsFindBy(ctx context.Context, f models.BaseSubjectsFilterRequest) (baseSubjects []*models.BaseSubjects, total int, err error) {
	args := []interface{}{f.Limit, f.Offset}
	qs, args := BaseSubjectsListBuildQuery(f, args)
	err = d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		rows, err := tx.Query(ctx, qs, args...)
		for rows.Next() {
			t := models.BaseSubjects{}
			err = scanBaseSubjects(rows, &t, &total)
			if err != nil {
				utils.LoggerDesc("Scan error").Error(err)
				return err
			}
			baseSubjects = append(baseSubjects, &t)
		}
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, 0, err
	}

	return baseSubjects, total, nil
}

func (d *PgxStore) BaseSubjectsFindById(ctx context.Context, ID string) (*models.BaseSubjects, error) {
	row, err := d.BaseSubjectsFindByIds(ctx, []string{ID})
	if err != nil {
		return nil, err
	}
	if len(row) < 1 {
		return nil, errors.New("base_subjects not found by id: " + ID)
	}
	return row[0], nil
}

func (d *PgxStore) BaseSubjectsFindByIds(ctx context.Context, Ids []string) ([]*models.BaseSubjects, error) {
	baseSubjects := []*models.BaseSubjects{}

	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		rows, err := tx.Query(ctx, sqlBaseSubjectsSelect, (Ids))
		for rows.Next() {
			t := models.BaseSubjects{}
			err := scanBaseSubjects(rows, &t)
			if err != nil {
				utils.LoggerDesc("Scan error").Error(err)
				return err
			}
			baseSubjects = append(baseSubjects, &t)
		}
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}

	return baseSubjects, nil
}

func (d *PgxStore) BaseSubjectsCreate(ctx context.Context, model *models.BaseSubjects) (*models.BaseSubjects, error) {
	qs, args := BaseSubjectsCreateQuery(model)
	qs += " RETURNING uid"
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		err = tx.QueryRow(ctx, qs, args...).Scan(&model.ID)
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}
	editModel, err := d.BaseSubjectsFindById(ctx, model.ID)
	if err != nil {
		return nil, err
	}
	return editModel, nil
}

func (d *PgxStore) BaseSubjectsUpdate(ctx context.Context, model *models.BaseSubjects) (*models.BaseSubjects, error) {
	qs, args := BaseSubjectsUpdateQuery(model)
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		_, err = tx.Query(ctx, qs, args...)
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}
	editModel, err := d.BaseSubjectsFindById(ctx, model.ID)
	if err != nil {
		return nil, err
	}
	return editModel, nil
}

func (d *PgxStore) BaseSubjectsDelete(ctx context.Context, items []*models.BaseSubjects) ([]*models.BaseSubjects, error) {
	ids := []string{}
	for _, i := range items {
		ids = append(ids, i.ID)
	}
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		_, err = tx.Query(ctx, sqlBaseSubjectsDelete, (ids))
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}
	return items, nil
}

func BaseSubjectsCreateQuery(m *models.BaseSubjects) (string, []interface{}) {
	args := []interface{}{}
	cols := ""
	vals := ""
	q := BaseSubjectsAtomicQuery(m, true)
	for k, v := range q {
		args = append(args, v)
		cols += ", " + k
		vals += ", $" + strconv.Itoa(len(args))
	}
	qs := sqlBaseSubjectsInsert + " (" + strings.Trim(cols, ", ") + ") VALUES (" + strings.Trim(vals, ", ") + ")"
	return qs, args
}

func BaseSubjectsUpdateQuery(m *models.BaseSubjects) (string, []interface{}) {
	args := []interface{}{}
	sets := ""
	q := BaseSubjectsAtomicQuery(m, false)
	for k, v := range q {
		args = append(args, v)
		sets += ", " + k + "=$" + strconv.Itoa(len(args))
	}
	args = append(args, m.ID)
	qs := strings.ReplaceAll(sqlBaseSubjectsUpdate, "set uid=uid", "set uid=uid"+sets+" ") + "where uid=$" + strconv.Itoa(len(args))
	return qs, args
}

func BaseSubjectsAtomicQuery(m *models.BaseSubjects, isCreate bool) map[string]interface{} {
	q := map[string]interface{}{}
	if m.Name != nil {
		q["name"] = m.Name
	}
	if m.SchoolId != nil {
		q["school_uid"] = m.SchoolId
	}
	if m.Category != nil {
		q["category"] = m.Category
	}
	if m.Price != nil {
		q["price"] = m.Price
	}
	if m.ExamMinGrade != nil {
		q["exam_min_grade"] = m.ExamMinGrade
	}
	if m.AgeCategory != nil {
		q["age_category"] = m.AgeCategory
	}
	if m.IsAvailable != nil {
		q["is_available"] = m.IsAvailable
	}
	if isCreate {
		q["created_at"] = time.Now()
	}
	q["updated_at"] = time.Now()
	return q
}

func BaseSubjectsListBuildQuery(f models.BaseSubjectsFilterRequest, args []interface{}) (string, []interface{}) {
	var wheres string = ""

	if f.ID != "" {
		args = append(args, f.ID)
		wheres += " and bs.uid=$" + strconv.Itoa(len(args))
	}
	if f.IDs != nil {
		args = append(args, *f.IDs)
		wheres += " and bs.uid = ANY($" + strconv.Itoa(len(args)) + "::uuid[])"
	}
	if f.SchoolId != nil {
		args = append(args, *f.SchoolId)
		wheres += " and bs.school_uid = ANY($" + strconv.Itoa(len(args)) + "::uuid[])"
	}
	if f.SchoolIds != nil {
		args = append(args, *f.SchoolIds)
		wheres += " and bs.school_uid = ANY($" + strconv.Itoa(len(args)) + "::uuid[])"
	}
	if f.Categories != nil {
		categories := "{" + strings.Join(*f.Categories, ",") + "}"
		args = append(args, categories)
		wheres += " and bs.categories::varchar[] @> $" + strconv.Itoa(len(args)) + "::varchar[]"
	}
	if f.AgeCategories != nil {
		args = append(args, *f.AgeCategories)
		wheres += " and bs.age_category = ANY($" + strconv.Itoa(len(args)) + "::uuid[])"
	}
	if f.Search != nil && *f.Search != "" {
		*f.Search = strings.ToLower(*f.Search)
		args = append(args, *f.Search)
		wheres += " and (lower(bs.name) like '%' || $" + strconv.Itoa(len(args)) + "|| '%')"
	}
	wheres += " group by bs.uid "
	if f.Sort != nil && *f.Sort != "" {
		dir := "desc"
		if strings.HasSuffix(*f.Sort, "~") {
			dir = "asc"
		}
		*f.Sort = strings.ReplaceAll(*f.Sort, "~", "")
		wheres += " order by " + *f.Sort + " " + dir
	} else {
		wheres += " order by bs.name desc"
	}
	qs := sqlBaseSubjectsSelectMany
	qs = strings.ReplaceAll(qs, "bs.uid=bs.uid", "bs.uid=bs.uid "+wheres+" ")
	return qs, args
}

func (d *PgxStore) BaseSubjectsLoadRelations(ctx context.Context, l *[]*models.BaseSubjects) error {
	ids := []string{}
	for _, m := range *l {
		ids = append(ids, m.ID)
	}
	if len(ids) < 1 {
		return nil
	}

	if rs, err := d.BaseSubjectsLoadSchool(ctx, ids); err != nil {
		return err
	} else {
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
	}
	return nil
}

type BaseSubjectsLoadSchoolItem struct {
	ID       string
	Relation *models.School
}

func (d *PgxStore) BaseSubjectsLoadSchool(ctx context.Context, ids []string) ([]BaseSubjectsLoadSchoolItem, error) {
	res := []BaseSubjectsLoadSchoolItem{}
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		rows, err := tx.Query(ctx, sqlBaseSubjectsSchool, (ids))
		for rows.Next() {
			sub := models.School{}
			pid := ""
			err = scanSchool(rows, &sub, &pid)
			if err != nil {
				return err
			}
			res = append(res, BaseSubjectsLoadSchoolItem{ID: pid, Relation: &sub})
		}
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}

	return res, nil
}
