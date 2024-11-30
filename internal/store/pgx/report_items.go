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

const sqlReportItemsFields = `ri.uid, ri.report_uid, ri.school_uid, ri.period_uid, ri.classroom_uid, ri.updated_by, ri.values, ri.is_edited_manually, ri.created_at, ri.updated_at`

const sqlReportItemsInsert = `insert into report_items`

const sqlReportItemsUpdate = `update report_items ri set uid=uid`

const sqlReportItemsSelect = `select ` + sqlReportItemsFields + ` from report_items ri where ri.uid = ANY($1::uuid[])`

const sqlReportItemsSelectMany = `select ` + sqlReportItemsFields + `, count(*) over() as total from 
	report_items ri 
	left join schools s on s.uid=ri.school_uid
	left join classrooms c on c.uid=ri.classroom_uid
	where ri.uid=ri.uid 
	limit $1 offset $2 `

const sqlReportItemsRelationReport = `select ` + sqlReportsFields + `, ri.uid from report_items ri
	right join reports rp on (rp.uid=ri.report_uid) where ri.uid = ANY($1::uuid[])`

const sqlReportItemsRelationSchool = `select ` + sqlSchoolFields + `, ri.uid from report_items ri
	right join schools s on (s.uid=ri.school_uid) where ri.uid = ANY($1::uuid[])`

const sqlReportItemsRelationPeriod = `select ` + sqlPeriodFields + `, ri.uid from report_items ri
	right join periods p on (p.uid=ri.period_uid) where ri.uid = ANY($1::uuid[])`

const sqlReportItemsRelationClassroom = `select ` + sqlClassroomFields + `, ri.uid from report_items ri
	right join classrooms c on (c.uid=ri.classroom_uid) where ri.uid = ANY($1::uuid[])`

const sqlReportItemsRelationUpdatedBy = `select ` + sqlUserFields + `, ri.uid from report_items ri
	right join users u on (u.uid=ri.updated_by) where ri.uid = ANY($1::uuid[])`

func scanReportItems(rows pgx.Rows, m *models.ReportItems, addColumns ...interface{}) (err error) {
	err = rows.Scan(parseColumnsForScan(m, addColumns...)...)
	return
}

func (d *PgxStore) ReportItemsFindBy(ctx context.Context, f models.ReportItemsFilterRequest) (reportItems []*models.ReportItems, total int, err error) {
	args := []interface{}{f.Limit, f.Offset}
	qs, args := ReportItemsListBuildQuery(f, args)
	err = d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		rows, err := tx.Query(ctx, qs, args...)
		for rows.Next() {
			reportItem := models.ReportItems{}
			err = scanReportItems(rows, &reportItem, &total)
			if err != nil {
				utils.LoggerDesc("Scan error").Error(err)
				return err
			}
			reportItems = append(reportItems, &reportItem)
		}
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, 0, err
	}

	return reportItems, total, nil
}

func (d *PgxStore) ReportItemsFindById(ctx context.Context, ID string) (*models.ReportItems, error) {
	row, err := d.ReportItemsFindByIds(ctx, []string{ID})
	if err != nil {
		return nil, err
	}
	if len(row) < 1 {
		return nil, errors.New("report_item not found by uid: " + ID)
	}
	return row[0], nil
}

func (d *PgxStore) ReportItemsFindByIds(ctx context.Context, Ids []string) ([]*models.ReportItems, error) {
	reportItems := []*models.ReportItems{}

	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		rows, err := tx.Query(ctx, sqlReportItemsSelect, (Ids))
		for rows.Next() {
			reportItem := models.ReportItems{}
			err := scanReportItems(rows, &reportItem)
			if err != nil {
				utils.LoggerDesc("Scan error").Error(err)
				return err
			}
			reportItems = append(reportItems, &reportItem)
		}
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}

	return reportItems, nil
}

func (d *PgxStore) ReportItemsCreate(ctx context.Context, model models.ReportItems) (*models.ReportItems, error) {
	qs, args := ReportItemsCreateQuery(model)
	qs += " RETURNING uid"
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		err = tx.QueryRow(ctx, qs, args...).Scan(&model.ID)
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}
	editModel, err := d.ReportItemsFindById(ctx, model.ID)
	if err != nil {
		return nil, err
	}
	return editModel, nil
}

func (d *PgxStore) ReportItemsUpdate(ctx context.Context, model models.ReportItems) (*models.ReportItems, error) {
	qs, args := ReportItemsUpdateQuery(model)
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		_, err = tx.Query(ctx, qs, args...)
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}
	editModel, err := d.ReportItemsFindById(ctx, model.ID)
	if err != nil {
		return nil, err
	}
	return editModel, nil
}

func ReportItemsUpdateQuery(m models.ReportItems) (string, []interface{}) {
	args := []interface{}{}
	sets := ""
	q := ReportItemsAtomicQuery(m, false)
	for k, v := range q {
		args = append(args, v)
		sets += ", " + k + "=$" + strconv.Itoa(len(args))
	}
	args = append(args, m.ID)
	qs := strings.ReplaceAll(sqlReportItemsUpdate, "set uid=uid", "set uid=uid "+sets+" ") + "where uid=$" + strconv.Itoa(len(args))
	return qs, args
}

func ReportItemsCreateQuery(m models.ReportItems) (string, []interface{}) {
	args := []interface{}{}
	cols := ""
	vals := ""
	q := ReportItemsAtomicQuery(m, true)
	for k, v := range q {
		args = append(args, v)
		cols += ", " + k
		vals += ", $" + strconv.Itoa(len(args))
	}
	qs := sqlReportItemsInsert + " (" + strings.Trim(cols, ", ") + ") VALUES (" + strings.Trim(vals, ", ") + ")"
	// qs += " ON CONFLICT DO NOTHING"
	return qs, args
}

func (d *PgxStore) ReportItemsCreateBatch(ctx context.Context, l []models.ReportItems) error {
	err := d.ReportItemsBatch(ctx, l, ReportItemsCreateQuery)
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return err
	}
	return err
}

func (d *PgxStore) ReportItemsBatch(ctx context.Context, l []models.ReportItems, f func(models.ReportItems) (string, []interface{})) error {
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

func ReportItemsAtomicQuery(m models.ReportItems, isCreate bool) map[string]interface{} {
	q := map[string]interface{}{}
	if m.ReportId != nil {
		q["report_uid"] = m.ReportId
	}
	if m.SchoolId != nil {
		q["school_uid"] = m.SchoolId
	}
	if m.PeriodId != nil {
		q["period_uid"] = m.PeriodId
	}
	if m.ClassroomId != nil {
		q["classroom_uid"] = m.ClassroomId
	}
	if m.UpdatedBy != nil {
		q["updated_by"] = m.UpdatedBy
	}
	if m.IsEditedManually != nil {
		q["is_edited_manually"] = m.IsEditedManually
	}
	if m.Values != nil {
		q["values"] = m.Values
	}
	if isCreate {
		q["created_at"] = time.Now()
		q["updated_at"] = time.Now()
	} else {
		q["updated_at"] = time.Now()
	}
	return q
}

func ReportItemsListBuildQuery(f models.ReportItemsFilterRequest, args []interface{}) (string, []interface{}) {
	var wheres string = ""
	if f.ID != nil {
		args = append(args, *f.ID)
		wheres += " and ri.uid=$" + strconv.Itoa(len(args))
	}
	if f.IDs != nil && len(*f.IDs) > 0 {
		args = append(args, *f.IDs)
		wheres += " and ri.uid= ANY($" + strconv.Itoa(len(args)) + "::uuid[])"
	}
	if f.ReportId != nil && *f.ReportId != "" {
		args = append(args, *f.ReportId)
		wheres += " and ri.report_uid=$" + strconv.Itoa(len(args))
	}
	if f.SchoolId != nil && *f.SchoolId != "" {
		args = append(args, *f.SchoolId)
		wheres += " and ri.school_uid=$" + strconv.Itoa(len(args))
	}
	if f.SchoolIds != nil && len(f.SchoolIds) != 0 {
		args = append(args, f.SchoolIds)
		wheres += " and ri.school_uid=ANY($" + strconv.Itoa(len(args)) + ")"
	}
	if f.ClassroomId != nil && *f.ClassroomId != "" {
		args = append(args, f.ClassroomId)
		wheres += " and ri.classroom_uid=$" + strconv.Itoa(len(args))
	}
	if f.PeriodId != nil && *f.PeriodId != "" {
		args = append(args, *f.PeriodId)
		wheres += " and ri.period_uid=$" + strconv.Itoa(len(args))
	}
	if f.OnlyClassroom != nil {
		if *f.OnlyClassroom {
			wheres += " and ri.classroom_uid IS NOT NULL"
		} else {
			wheres += " and ri.classroom_uid IS NULL"
		}
	}
	// wheres += " group by ri.uid"
	if f.Sort != nil && *f.Sort != "" {
		dir := "desc"
		if strings.HasSuffix(*f.Sort, "~") {
			dir = "asc"
		}
		*f.Sort = strings.ReplaceAll(*f.Sort, "~", "")
		wheres += " order by ri.created_at " + dir + ", ri.updated_at is null, ri.uid desc"
	} else {
		wheres += " order by LPAD(replace(lower(s.code), 'ag', ''), 3, '0'), s.parent_uid, CAST(REVERSE(SUBSTRING(REVERSE(s.code) FROM '[0-9]+')) AS INTEGER), s.code desc, LPAD(lower(c.name), 3, '0') asc"

	}
	qs := sqlReportItemsSelectMany
	qs = strings.ReplaceAll(qs, "ri.uid=ri.uid", "ri.uid=ri.uid "+wheres+" ")

	return qs, args
}

func (d *PgxStore) ReportItemsLoadRelations(ctx context.Context, l *[]*models.ReportItems) error {
	ids := []string{}
	for _, m := range *l {
		ids = append(ids, m.ID)
	}
	if len(ids) < 1 {
		return nil
	}

	// if rs, err := d.ReportItemsLoadReport(ctx, ids); err != nil {
	// 	return err
	// } else {
	// 	for _, r := range rs {
	// 		for _, m := range *l {
	// 			if r.ID == m.ID {
	// 				m.Report = r.Relation
	// 			}
	// 		}
	// 	}
	// }
	if rs, err := d.ReportItemsLoadSchool(ctx, ids); err != nil {
		return err
	} else {
		for _, r := range rs {
			for _, m := range *l {
				if r.ID == m.ID {
					m.School = r.Relation
				}
			}
		}
	}
	if rs, err := d.ReportItemsLoadPeriod(ctx, ids); err != nil {
		return err
	} else {
		for _, r := range rs {
			for _, m := range *l {
				if r.ID == m.ID {
					m.Period = r.Relation
				}
			}
		}
	}
	if rs, err := d.ReportItemsLoadClassroom(ctx, ids); err != nil {
		return err
	} else {
		for _, r := range rs {
			for _, m := range *l {
				if r.ID == m.ID {
					m.Classroom = r.Relation
				}
			}
		}
	}
	if rs, err := d.ReportItemsLoadUpdatedBy(ctx, ids); err != nil {
		return err
	} else {
		for _, r := range rs {
			for _, m := range *l {
				if r.ID == m.ID {
					m.UpdatedByUser = r.Relation
				}
			}
		}
	}

	return nil
}

type ReportItemsLoadReportItem struct {
	ID       string
	Relation *models.Reports
}

func (d *PgxStore) ReportItemsLoadReport(ctx context.Context, ids []string) ([]ReportItemsLoadReportItem, error) {
	res := []ReportItemsLoadReportItem{}
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		rows, err := tx.Query(ctx, sqlReportItemsRelationReport, ids)
		for rows.Next() {
			sub := models.Reports{}
			pid := ""
			err = scanReports(rows, &sub, &pid)
			if err != nil {
				return err
			}
			res = append(res, ReportItemsLoadReportItem{ID: pid, Relation: &sub})
		}
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}
	return res, nil
}

type ReportItemsLoadSchoolItem struct {
	ID       string
	Relation *models.School
}

func (d *PgxStore) ReportItemsLoadSchool(ctx context.Context, ids []string) ([]ReportItemsLoadSchoolItem, error) {
	res := []ReportItemsLoadSchoolItem{}
	err := d.runQuery(context.Background(), func(tx *pgxpool.Conn) (err error) {
		rows, err := tx.Query(context.Background(), sqlReportItemsRelationSchool, ids)
		for rows.Next() {
			sub := models.School{}
			pid := ""
			err = scanSchool(rows, &sub, &pid)
			if err != nil {
				return err
			}
			res = append(res, ReportItemsLoadSchoolItem{ID: pid, Relation: &sub})
		}
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}
	return res, nil
}

type ReportItemsLoadPeriodItem struct {
	ID       string
	Relation *models.Period
}

func (d *PgxStore) ReportItemsLoadPeriod(ctx context.Context, ids []string) ([]ReportItemsLoadPeriodItem, error) {
	res := []ReportItemsLoadPeriodItem{}
	err := d.runQuery(context.Background(), func(tx *pgxpool.Conn) (err error) {
		rows, err := tx.Query(context.Background(), sqlReportItemsRelationPeriod, ids)
		for rows.Next() {
			sub := models.Period{}
			pid := ""
			err = scanPeriod(rows, &sub, &pid)
			if err != nil {
				return err
			}
			res = append(res, ReportItemsLoadPeriodItem{ID: pid, Relation: &sub})
		}
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}
	return res, nil
}

type ReportItemsLoadClassroomItem struct {
	ID       string
	Relation *models.Classroom
}

func (d *PgxStore) ReportItemsLoadClassroom(ctx context.Context, ids []string) ([]ReportItemsLoadClassroomItem, error) {
	res := []ReportItemsLoadClassroomItem{}
	err := d.runQuery(context.Background(), func(tx *pgxpool.Conn) (err error) {
		rows, err := tx.Query(context.Background(), sqlReportItemsRelationClassroom, ids)
		for rows.Next() {
			sub := models.Classroom{}
			pid := ""
			err = scanClassroom(rows, &sub, &pid)
			if err != nil {
				return err
			}
			res = append(res, ReportItemsLoadClassroomItem{ID: pid, Relation: &sub})
		}
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}
	return res, nil
}

type ReportItemsLoadUpdatedByItem struct {
	ID       string
	Relation *models.User
}

func (d *PgxStore) ReportItemsLoadUpdatedBy(ctx context.Context, ids []string) ([]ReportItemsLoadUpdatedByItem, error) {
	res := []ReportItemsLoadUpdatedByItem{}
	err := d.runQuery(context.Background(), func(tx *pgxpool.Conn) (err error) {
		rows, err := tx.Query(context.Background(), sqlReportItemsRelationUpdatedBy, ids)
		for rows.Next() {
			sub := models.User{}
			pid := ""
			err = scanUser(rows, &sub, &pid)
			if err != nil {
				return err
			}
			res = append(res, ReportItemsLoadUpdatedByItem{ID: pid, Relation: &sub})
		}
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}
	return res, nil
}
