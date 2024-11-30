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

// FIELDS
const sqlContactItemFieldsMany = `ct.uid, ct.user_uid, ct.school_uid, ct.message, ct.type, ct.status, ct.files, ct.note, ct.classroom_name, ct.parent_phone, ct.birth_cert_number, count(distinct ctt.uid), ct.related_uid, ct.created_at, ct.updated_at, ct.updated_by`
const sqlContactItemFields = `ct.uid, ct.user_uid, ct.school_uid, ct.message, ct.type, ct.status, ct.files, ct.note, ct.classroom_name, ct.parent_phone, ct.birth_cert_number, 0, ct.related_uid, ct.created_at, ct.updated_at, ct.updated_by`

// CRUD
const sqlContactItemSelect = `SELECT ` + sqlContactItemFieldsMany + ` FROM contact_items ct 
	left join contact_items ctt on ctt.related_uid=ct.uid
	WHERE ct.uid = ANY($1::uuid[]) group by ct.uid`
const sqlContactItemSelectMany = `SELECT ` + sqlContactItemFieldsMany + `, count(*) over() as total FROM contact_items ct 
	left join contact_items ctt on ctt.related_uid=ct.uid
	where ct.uid=ct.uid limit $1 offset $2 `
const sqlContactItemUpdate = `UPDATE contact_items ct set uid=uid`
const sqlContactItemInsert = `INSERT INTO contact_items`
const sqlContactItemDelete = `delete from contact_items ct where uid = ANY($1::uuid[])`

// relations
const sqlContactItemUser = `select ` + sqlUserFields + `, ct.uid from contact_items ct
right join users u on (u.uid=ct.user_uid) where ct.uid = ANY($1::uuid[])`

const sqlContactItemUpdatedBy = `select ` + sqlUserFields + `, ct.uid from contact_items ct
	right join users u on (u.uid=ct.updated_by) where ct.uid = ANY($1::uuid[])`

const sqlContactItemSchool = `select ` + sqlSchoolFields + `, ct.uid from contact_items ct
	right join schools s on (s.uid=ct.school_uid) where ct.uid = ANY($1::uuid[])`

const sqlContactItemRelated = `select ` + sqlContactItemFields + `, ctt.uid from contact_items ctt 
	right join contact_items ct on ctt.related_uid=ct.uid  where ctt.uid = ANY($1::uuid[])`

const sqlContactItemRelatedChildren = `select ` + sqlContactItemFields + `, ct.related_uid from contact_items ct 
	where ct.related_uid = ANY($1::uuid[])`

// COUNTS
const sqlContactItemsCountByType = `SELECT
	s.code,
	COUNT(*) as total_count,
	SUM(CASE WHEN ct.type = 'review' THEN 1 ELSE 0 END) AS review_count,
	SUM(CASE WHEN ct.type = 'suggestion' THEN 1 ELSE 0 END) AS suggestion_count,
	SUM(CASE WHEN ct.type = 'complaint' THEN 1 ELSE 0 END) AS complaint_count,
	SUM(CASE WHEN ct.type = 'data_complaint' THEN 1 ELSE 0 END) AS data_complaint_count
FROM contact_items ct JOIN schools s ON ct.school_uid = s.uid GROUP BY s.code`

func scanContactItem(rows pgx.Row, m *models.ContactItems, addColumns ...interface{}) (err error) {
	err = rows.Scan(parseColumnsForScan(m, addColumns...)...)
	return
}

func (d *PgxStore) ContactItemsCountByType(ctx context.Context, f models.ContactItemsFilterRequest) ([]models.ContactItemsCount, error) {
	items := []models.ContactItemsCount{}
	args := []interface{}{}
	qs, args := ContactItemsListBuildQuery(f, args, sqlContactItemsCountByType)
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		rows, err := tx.Query(ctx, qs, args...)
		for rows.Next() {
			item := models.ContactItemsCount{}
			err = rows.Scan(&item.SchoolCode, &item.TotalCount, &item.ReviewCount, &item.ComplaintCount, &item.SuggestionCount, &item.DataComplaintCount)
			if err != nil {
				return err
			}
			items = append(items, item)
		}
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}
	return items, nil
}

func (d *PgxStore) ContactItemsFindBy(ctx context.Context, f models.ContactItemsFilterRequest) (contactItems []*models.ContactItems, total int, err error) {
	args := []interface{}{f.Limit, f.Offset}
	qs, args := ContactItemsListBuildQuery(f, args, "")
	err = d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		rows, err := tx.Query(ctx, qs, args...)
		for rows.Next() {
			contactItem := models.ContactItems{}
			err = scanContactItem(rows, &contactItem, &total)
			if err != nil {
				return err
			}
			contactItems = append(contactItems, &contactItem)
		}
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, 0, err
	}

	return contactItems, total, nil
}

func (d *PgxStore) ContactItemsFindById(ctx context.Context, Id string) (*models.ContactItems, error) {
	row, err := d.ContactItemsFindByIds(ctx, []string{Id})
	if err != nil {
		return nil, err
	}
	if len(row) < 1 {
		return nil, errors.New("contact_item not found by id: " + Id)
	}
	return row[0], nil
}

func (d *PgxStore) ContactItemsFindByIds(ctx context.Context, Ids []string) ([]*models.ContactItems, error) {
	contactItems := []*models.ContactItems{}
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		rows, err := tx.Query(ctx, sqlContactItemSelect, (Ids))
		for rows.Next() {
			contactItem := models.ContactItems{}
			err := scanContactItem(rows, &contactItem)
			if err != nil {
				return err
			}
			contactItems = append(contactItems, &contactItem)
		}
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}

	return contactItems, nil
}

func (d *PgxStore) ContactItemUpdate(ctx context.Context, model *models.ContactItems) (*models.ContactItems, error) {
	qs, args := ContactItemUpdateQuery(model)
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		_, err = tx.Query(ctx, qs, args...)
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}
	editModel, err := d.ContactItemsFindById(ctx, model.ID)
	if err != nil {
		return nil, err
	}
	d.ContactItemUpdateRelations(ctx, model, editModel)
	return editModel, nil
}

func (d *PgxStore) ContactItemCreate(ctx context.Context, model *models.ContactItems) (*models.ContactItems, error) {
	qs, args := ContactItemCreateQuery(model)
	qs += " RETURNING uid"
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		err = tx.QueryRow(ctx, qs, args...).Scan(&model.ID)
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}
	editModel, err := d.ContactItemsFindById(ctx, model.ID)
	if err != nil {
		return nil, err
	}
	d.ContactItemUpdateRelations(ctx, model, editModel)
	return editModel, nil
}

func (d *PgxStore) ContactItemsDelete(ctx context.Context, items []*models.ContactItems) ([]*models.ContactItems, error) {
	ids := []string{}
	for _, i := range items {
		ids = append(ids, i.ID)
	}
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		_, err = tx.Query(ctx, sqlContactItemDelete, (ids))
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}
	return items, nil
}

func (d *PgxStore) ContactItemUpdateRelations(ctx context.Context, data *models.ContactItems, model *models.ContactItems) {

}

func ContactItemCreateQuery(m *models.ContactItems) (string, []interface{}) {
	args := []interface{}{}
	cols := ""
	vals := ""
	q := ContactItemAtomicQuery(m, true)
	for k, v := range q {
		args = append(args, v)
		cols += ", " + k
		vals += ", $" + strconv.Itoa(len(args))
	}
	qs := sqlContactItemInsert + " (" + strings.Trim(cols, ", ") + ") VALUES (" + strings.Trim(vals, ", ") + ")"
	return qs, args
}

func ContactItemUpdateQuery(m *models.ContactItems) (string, []interface{}) {
	args := []interface{}{}
	sets := ""
	q := ContactItemAtomicQuery(m, false)
	for k, v := range q {
		args = append(args, v)
		sets += ", " + k + "=$" + strconv.Itoa(len(args))
	}
	args = append(args, m.ID)
	qs := strings.ReplaceAll(sqlContactItemUpdate, "set uid=uid", "set uid=uid "+sets+" ") + "where uid=$" + strconv.Itoa(len(args))
	return qs, args
}

func ContactItemAtomicQuery(m *models.ContactItems, isCreate bool) map[string]interface{} {
	q := map[string]interface{}{}
	if m.Message != nil {
		q["message"] = m.Message
	}
	if m.UserId != nil {
		q["user_uid"] = m.UserId
	}
	if m.SchoolId != nil {
		q["school_uid"] = m.SchoolId
	}
	if m.Type != "" {
		q["type"] = m.Type
	}
	if m.Status != "" {
		q["status"] = m.Status
	}
	if m.Note != nil {
		q["note"] = m.Note
	}
	if m.ClassroomName != nil {
		q["classroom_name"] = m.ClassroomName
	}
	if m.ParentPhone != nil {
		q["parent_phone"] = m.ParentPhone
	}
	if m.BirthCertNumber != nil {
		q["birth_cert_number"] = m.BirthCertNumber
	}
	if m.RelatedId != nil {
		q["related_uid"] = m.RelatedId
	}
	if m.Status != "" {
		q["status"] = m.Status
	}
	if m.Files != nil {
		q["files"] = m.Files
	}
	if m.UpdatedBy != nil {
		q["updated_by"] = m.UpdatedBy
	}
	if isCreate {
		q["created_at"] = time.Now()
	}
	q["updated_at"] = time.Now()
	return q
}

func ContactItemsListBuildQuery(f models.ContactItemsFilterRequest, args []interface{}, qs string) (string, []interface{}) {
	var wheres string = ""
	if f.ID != nil && *f.ID != "" {
		args = append(args, *f.ID)
		wheres += " and ct.uid=$" + strconv.Itoa(len(args))
	}
	if f.IDs != nil && len(*f.IDs) > 0 {
		args = append(args, *f.IDs)
		wheres += " and (ct.uid = ANY($" + strconv.Itoa(len(args)) + ") )"
	}
	if f.RelatedIds != nil {
		args = append(args, *f.RelatedIds)
		wheres += " and ANY(ct.related_uid)=$" + strconv.Itoa(len(args))
	}
	if f.UserId != nil && *f.UserId != "" {
		args = append(args, *f.UserId)
		wheres += " and ct.user_uid=$" + strconv.Itoa(len(args))
	}
	if f.UpdatedBy != nil && *f.UpdatedBy != "" {
		args = append(args, *f.UpdatedBy)
		wheres += " and ct.updated_by=$" + strconv.Itoa(len(args))
	}
	if f.SchoolId != nil && *f.SchoolId != "" {
		args = append(args, *f.SchoolId)
		wheres += " and ct.school_uid=$" + strconv.Itoa(len(args))
	}
	if f.SchoolIds != nil {
		args = append(args, *f.SchoolIds)
		wheres += " and ct.school_uid = ANY($" + strconv.Itoa(len(args)) + "::uuid[])"
	}
	if f.OnlyNotRelated != nil && *f.OnlyNotRelated {
		wheres += " and ct.related_uid IS NULL"
	}
	if f.NotStatus != nil && *f.NotStatus != "" {
		args = append(args, *f.NotStatus)
		wheres += " and ct.status!=$" + strconv.Itoa(len(args))
	}
	if f.Status != nil && *f.Status != "" {
		args = append(args, *f.Status)
		wheres += " and ct.status=$" + strconv.Itoa(len(args))
	}
	if f.Type != nil && *f.Type != "" {
		args = append(args, *f.Type)
		wheres += " and ct.type=$" + strconv.Itoa(len(args))
	}
	if f.Search != nil && *f.Search != "" {
		*f.Search = strings.ToLower(*f.Search)
		args = append(args, *f.Search)
		tmp := "$" + strconv.Itoa(len(args))
		wheres += " and (lower(ct.message) like '%' ||" + tmp + "|| '%' or lower(ct.note) like '%' ||" + tmp +
			"||'%' or lower(ct.parent_phone) like '%'||" + tmp + "||'%' or lower(ct.birth_cert_number) like '%'||" + tmp +
			"||'%')"
	}
	if f.StartDate != nil && f.EndDate != nil {
		startDate, _ := time.Parse(time.DateOnly, *f.StartDate)
		startDate = startDate.Add(time.Hour * 0)
		endDate, _ := time.Parse(time.DateOnly, *f.EndDate)
		endDate = endDate.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
		args = append(args, startDate, endDate)
		wheres += " and ct.created_at BETWEEN $" + strconv.Itoa(len(args)-1) + " AND $" + strconv.Itoa(len(args))
	} else if f.StartDate != nil {
		startDate, _ := time.Parse(time.DateOnly, *f.StartDate)
		startDate = startDate.Add(time.Hour * 0)
		// Only start_date is provided
		args = append(args, startDate)
		wheres += " and ct.created_at >= $" + strconv.Itoa(len(args))
	} else if f.EndDate != nil {
		endDate, _ := time.Parse(time.DateOnly, *f.EndDate)
		endDate = endDate.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
		// Only end_date is provided
		args = append(args, endDate)
		wheres += " and ct.created_at <= $" + strconv.Itoa(len(args))
	}
	if qs == "" {
		wheres += " group by ct.uid "
		wheres += " order by ct.created_at desc"
		qs = sqlContactItemSelectMany
		qs = strings.ReplaceAll(qs, "ct.uid=ct.uid", "ct.uid=ct.uid "+wheres+" ")
	}
	qs = strings.ReplaceAll(qs, "ct.school_uid = s.uid", "ct.school_uid = s.uid "+wheres+" ")
	return qs, args
}

func (d *PgxStore) ContactItemLoadRelations(ctx context.Context, l *[]*models.ContactItems, isDetail bool) error {
	ids := []string{}
	for _, m := range *l {
		ids = append(ids, m.ID)
	}
	if len(ids) < 1 {
		return nil
	}
	var err error

	err = d.ContactItemLoadSchool(ctx, l)
	if err != nil {
		return err
	}
	err = d.ContactItemLoadUser(ctx, l)
	if err != nil {
		return err
	}
	err = d.ContactItemLoadUpdatedBy(ctx, l)
	if err != nil {
		return err
	}
	if isDetail {
		if err := d.ContactItemLoadRelated(ctx, l); err != nil {
			return err
		}
		if err := d.ContactItemLoadRelatedChildren(ctx, l); err != nil {
			return err
		}
	}

	return nil
}

func (d *PgxStore) ContactItemLoadRelated(ctx context.Context, l *[]*models.ContactItems) error {
	ids := []string{}
	for _, m := range *l {
		ids = append(ids, m.ID)
	}
	if len(ids) < 1 {
		return nil
	}
	// load
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		rows, err := tx.Query(ctx, sqlContactItemRelated, (ids))
		for rows.Next() {
			sub := models.ContactItems{}
			subId := ""
			err = scanContactItem(rows, &sub, &subId)
			if err != nil {
				return err
			}
			for _, m := range *l {
				if (m.ID) == subId {
					m.Related = &sub
				}
			}
		}
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return err
	}
	// Load relations for the related items (User and School)
	relatedItems := []*models.ContactItems{}
	for _, m := range *l {
		if m.Related != nil {
			relatedItems = append(relatedItems, m.Related)
		}
	}

	if len(relatedItems) > 0 {
		err = d.ContactItemLoadRelations(ctx, &relatedItems, false)
		if err != nil {
			return err
		}
	}
	return nil
}

func (d *PgxStore) ContactItemLoadRelatedChildren(ctx context.Context, l *[]*models.ContactItems) error {
	ids := []string{}
	for _, m := range *l {
		ids = append(ids, m.ID)
	}
	if len(ids) < 1 {
		return nil
	}
	// load
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		rows, err := tx.Query(ctx, sqlContactItemRelatedChildren, (ids))
		for rows.Next() {
			sub := models.ContactItems{}
			subId := ""
			err = scanContactItem(rows, &sub, &subId)
			if err != nil {
				return err
			}
			for _, m := range *l {
				if m.ID == subId {
					if m.RelatedChildren == nil {
						m.RelatedChildren = &[]*models.ContactItems{}
					}
					*m.RelatedChildren = append(*m.RelatedChildren, &sub)
				}
			}
		}

		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return err
	}
	childrenItems := []*models.ContactItems{}
	for _, parent := range *l {
		if parent.RelatedChildren != nil {
			for _, child := range *parent.RelatedChildren {
				childrenItems = append(childrenItems, child)
			}
		}
	}
	// Load relations for the related children (User and School models)
	if len(childrenItems) > 0 {
		err = d.ContactItemLoadRelations(ctx, &childrenItems, false)
		if err != nil {
			return err
		}
	}
	return nil
}

func (d *PgxStore) ContactItemLoadSchool(ctx context.Context, l *[]*models.ContactItems) error {
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
		rows, err := tx.Query(ctx, sqlContactItemSchool, ids)
		for rows.Next() {
			sub := models.School{}
			pid := ""
			err = scanSchool(rows, &sub, &pid)
			if err != nil {
				utils.LoggerDesc("Scan error").Error(err)
				return err
			}
			for _, m := range *l {
				if m.ID == pid {
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

func (d *PgxStore) ContactItemLoadUser(ctx context.Context, l *[]*models.ContactItems) error {
	ids := []string{}

	for _, m := range *l {
		ids = append(ids, m.ID)
	}
	if len(ids) < 1 {
		return nil
	}
	// load
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		rows, err := tx.Query(ctx, sqlContactItemUser, ids)
		for rows.Next() {
			sub := models.User{}
			pid := ""
			err = scanUser(rows, &sub, &pid)
			if err != nil {
				utils.LoggerDesc("Scan error").Error(err)
				return err
			}
			for _, m := range *l {
				if m.ID == pid {
					m.User = &sub
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

func (d *PgxStore) ContactItemLoadUpdatedBy(ctx context.Context, l *[]*models.ContactItems) error {
	ids := []string{}

	for _, m := range *l {
		ids = append(ids, m.ID)
	}
	if len(ids) < 1 {
		return nil
	}
	// load
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		rows, err := tx.Query(ctx, sqlContactItemUpdatedBy, ids)
		for rows.Next() {
			sub := models.User{}
			pid := ""
			err = scanUser(rows, &sub, &pid)
			if err != nil {
				utils.LoggerDesc("Scan error").Error(err)
				return err
			}
			for _, m := range *l {
				if m.ID == pid {
					m.UpdatedByUser = &sub
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
