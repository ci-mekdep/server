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

const sqlReportsFields = `rp.uid, rp.title, rp.description, rp.value_types, rp.school_uids, rp.region_uids, rp.is_pinned, rp.is_center_rating, rp.is_classrooms_included, rp.created_at, rp.updated_at`
const sqlReportsSelect = `SELECT ` + sqlReportsFields + `, null as items_count, null as items_filled_count FROM reports rp 
	left join report_items ri on (ri.report_uid=rp.uid) WHERE rp.uid = ANY($1::uuid[]) GROUP BY rp.uid`
const sqlReportsSelectMany = `SELECT ` + sqlReportsFields + `, null as items_count, null as items_filled_count, count(1) over() as total FROM reports rp 
	left join report_items ri on (ri.report_uid=rp.uid) where rp.uid=rp.uid limit $1 offset $2 `
const sqlReportsInsert = `INSERT INTO reports`
const sqlReportsUpdate = `UPDATE reports rp set uid=uid`
const sqlReportsDelete = `DELETE FROM reports rp where uid = ANY($1::uuid[])`

func scanReports(rows pgx.Row, m *models.Reports, addColumns ...interface{}) (err error) {
	err = rows.Scan(parseColumnsForScan(m, addColumns...)...)
	return
}

func (d *PgxStore) ReportsFindBy(ctx context.Context, f models.ReportsFilterRequest) (reports []*models.Reports, total int, err error) {
	args := []interface{}{f.Limit, f.Offset}
	qs, args := ReportsListBuildQuery(f, args)
	err = d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		rows, err := tx.Query(ctx, qs, args...)
		for rows.Next() {
			shift := models.Reports{}
			err = scanReports(rows, &shift, &total)
			if err != nil {
				return err
			}
			reports = append(reports, &shift)
		}
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, 0, err
	}

	return reports, total, nil
}

func (d *PgxStore) ReportsFindById(ctx context.Context, Id string) (*models.Reports, error) {
	row, err := d.ReportsFindByIds(ctx, []string{Id})
	if err != nil {
		return nil, err
	}
	if len(row) < 1 {
		return nil, errors.New("report not found by uid: " + Id)
	}
	return row[0], nil
}

func (d *PgxStore) ReportsFindByIds(ctx context.Context, Ids []string) ([]*models.Reports, error) {
	reports := []*models.Reports{}
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		rows, err := tx.Query(ctx, sqlReportsSelect, (Ids))
		for rows.Next() {
			report := models.Reports{}
			err := scanReports(rows, &report)
			if err != nil {
				return err
			}
			reports = append(reports, &report)
		}
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}

	return reports, nil
}

func (d *PgxStore) ReportsDelete(ctx context.Context, items []*models.Reports) ([]*models.Reports, error) {
	ids := []string{}
	for _, i := range items {
		ids = append(ids, i.ID)
	}
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		_, err = tx.Query(ctx, sqlReportsDelete, (ids))
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}
	return items, nil
}

func (d *PgxStore) ReportsCreate(ctx context.Context, model *models.Reports) (*models.Reports, error) {
	qs, args := ReportsCreateQuery(model)
	qs += " RETURNING uid"
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		err = tx.QueryRow(ctx, qs, args...).Scan(&model.ID)
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}
	editModel, err := d.ReportsFindById(ctx, model.ID)
	if err != nil {
		return nil, err
	}
	return editModel, nil
}

func ReportsCreateQuery(m *models.Reports) (string, []interface{}) {
	args := []interface{}{}
	cols := ""
	vals := ""
	q := ReportsAtomicQuery(m, true)
	for k, v := range q {
		args = append(args, v)
		cols += ", " + k
		vals += ", $" + strconv.Itoa(len(args))
	}
	qs := sqlReportsInsert + " (" + strings.Trim(cols, ", ") + ") VALUES (" + strings.Trim(vals, ", ") + ")"
	return qs, args
}

func (d *PgxStore) ReportsUpdate(ctx context.Context, model *models.Reports) (*models.Reports, error) {
	qs, args := ReportsUpdateQuery(model)
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		_, err = tx.Query(ctx, qs, args...)
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}
	editModel, err := d.ReportsFindById(ctx, model.ID)
	if err != nil {
		return nil, err
	}
	return editModel, nil
}

func ReportsUpdateQuery(m *models.Reports) (string, []interface{}) {
	args := []interface{}{}
	sets := ""
	q := ReportsAtomicQuery(m, false)
	for k, v := range q {
		args = append(args, v)
		sets += ", " + k + "=$" + strconv.Itoa(len(args))
	}
	args = append(args, m.ID)
	qs := strings.ReplaceAll(sqlReportsUpdate, "set uid=uid", "set uid=uid "+sets+" ") + "where uid=$" + strconv.Itoa(len(args))
	return qs, args
}

func ReportsAtomicQuery(m *models.Reports, isCreate bool) map[string]interface{} {
	q := map[string]interface{}{}
	q["title"] = m.Title
	if m.Description != nil {
		q["description"] = m.Description
	}
	q["value_types"] = m.ValueTypes
	q["school_uids"] = m.SchoolIds
	q["region_uids"] = m.RegionIds
	if m.IsPinned != nil {
		q["is_pinned"] = m.IsPinned
	}
	if m.IsCenterRating != nil {
		q["is_center_rating"] = m.IsCenterRating
	}
	if m.IsClassroomsIncluded != nil {
		q["is_classrooms_included"] = m.IsClassroomsIncluded
	}
	if isCreate {
		q["created_at"] = time.Now()
	}
	q["updated_at"] = time.Now()
	return q
}

func ReportsListBuildQuery(f models.ReportsFilterRequest, args []interface{}) (string, []interface{}) {
	var wheres string = ""

	if f.ID != nil && *f.ID != "" {
		args = append(args, *f.ID)
		wheres += " and rp.uid=$" + strconv.Itoa(len(args))
	}
	if f.IDs != nil {
		args = append(args, *f.IDs)
		wheres += " and rp.uid = ANY($" + strconv.Itoa(len(args)) + "::uuid[])"
	}
	if f.SchoolIds != nil {
		args = append(args, *f.SchoolIds)
		wheres += " and ri.school_uid=ANY($" + strconv.Itoa(len(args)) + "::uuid[])"
	}
	if f.ClassroomUids != nil {
		args = append(args, *f.ClassroomUids)
		wheres += " and ri.classroom_uid=ANY($" + strconv.Itoa(len(args)) + "::uuid[])"
	}
	if f.Search != nil && *f.Search != "" {
		*f.Search = strings.ToLower(*f.Search)
		args = append(args, *f.Search)
		wheres += " and (lower(rp.title) like '%' || $" + strconv.Itoa(len(args)) + "|| '%')"
	}
	wheres += " group by rp.uid "
	wheres += " order by rp.is_pinned desc, rp.created_at desc"

	qs := sqlReportsSelectMany
	qs = strings.ReplaceAll(qs, "rp.uid=rp.uid", "rp.uid=rp.uid "+wheres+" ")
	return qs, args
}
