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
	"go.elastic.co/apm/module/apmpgx/v2"
)

const sqlSchoolFields = sqlSchoolFieldsNoRelation + `, null, null`
const sqlSchoolFieldsNoRelation = `s.uid, s.code, s.name, s.full_name, s.description, s.address, s.avatar, s.background, s.phone, s.email, s.level, s.galleries, s.latitude, s.longitude, s.is_digitalized, s.is_secondary_school, s.parent_uid, s.admin_uid, s.specialist_uid, s.updated_at, s.created_at, s.archived_at`
const sqlSchoolSelect = `select ` + sqlSchoolFields + ` from schools s where s.uid = ANY($1::uuid[])`
const sqlSchoolSelectByCode = `select ` + sqlSchoolFields + ` from schools s where code = ANY($1::text[])`
const sqlSchoolSelectMany = `select ` + sqlSchoolFieldsNoRelation + `, count(distinct tt.uid) as timetables_count, count(c.uid) as classrooms_count, count(1) over() as total from schools s 
	left join classrooms c on (c.school_uid=s.uid) 
	left join timetables tt on (tt.classroom_uid=c.uid) where s.uid=s.uid limit $1 offset $2`
const sqlSchoolInsert = `insert into schools`
const sqlSchoolUpdate = `update schools set uid=uid`
const sqlSchoolDelete = `delete from schools where uid = ANY($1::uuid[])`

const sqlSchoolParent = `select ` + sqlSchoolFields + `, sp.uid from schools sp 
	right join schools s on sp.parent_uid=s.uid  where sp.uid = ANY($1::uuid[])`
const sqlSchoolAdmin = `select ` + sqlUserFields + `, s.uid from schools s 
	right join users u on (u.uid=s.admin_uid) where s.uid = ANY($1::uuid[])`
const sqlSchoolSpecialist = `select ` + sqlUserFields + `, s.uid from schools s 
	right join users u on (u.uid=s.specialist_uid) where s.uid = ANY($1::uuid[])`
const sqlSchoolAdminDelete = `delete from user_schools where school_uid=$1 and role_code='` + string(models.RolePrincipal) + `'`
const sqlSchoolAdminInsert = `insert into user_schools (school_uid, user_uid, role_code) values ($1, $2, '` + string(models.RolePrincipal) + `')`

func scanSchool(rows pgx.Row, m *models.School, addColumns ...interface{}) (err error) {
	err = rows.Scan(parseColumnsForScan(m, addColumns...)...)
	return
}

func (d *PgxStore) SchoolsFindByCode(ctx context.Context, codes []string) ([]*models.School, error) {
	l := []*models.School{}
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		rows, err := tx.Query(ctx, sqlSchoolSelectByCode, (codes))
		for rows.Next() {
			u := models.School{}
			err := scanSchool(rows, &u)
			if err != nil {
				utils.LoggerDesc("Scan error").Error(err)
				return err
			}
			l = append(l, &u)
		}
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}
	return l, nil
}

func (d *PgxStore) SchoolsFindByIds(ctx context.Context, ids []string) ([]*models.School, error) {
	l := []*models.School{}
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		rows, err := tx.Query(ctx, sqlSchoolSelect, (ids))
		for rows.Next() {
			u := models.School{}
			err := scanSchool(rows, &u)
			if err != nil {
				utils.LoggerDesc("Scan error").Error(err)
				return err
			}
			l = append(l, &u)
		}
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}
	return l, nil
}

func (d *PgxStore) SchoolsFindById(ctx context.Context, id string) (*models.School, error) {
	l, err := d.SchoolsFindByIds(ctx, []string{id})
	if err != nil {
		return nil, err
	}
	if len(l) < 1 {
		return nil, pgx.ErrNoRows
	}
	return l[0], err
}

func (d *PgxStore) SchoolsFindBy(ctx context.Context, f models.SchoolFilterRequest) ([]*models.School, int, error) {
	if f.Limit == nil {
		f.Limit = new(int)
		*f.Limit = 50
	}
	if f.Offset == nil {
		f.Offset = new(int)
		*f.Offset = 0
	}
	args := []interface{}{f.Limit, f.Offset}
	qs, args := SchoolsListBuildQuery(f, args)
	l := []*models.School{}
	var total int
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		apmpgx.Instrument(tx.Conn().Config())

		rows, err := tx.Query(ctx, qs, args...)

		if err != nil {
			return err
		}
		for rows.Next() {
			sub := models.School{}
			err = scanSchool(rows, &sub, &total)
			if err != nil {
				utils.LoggerDesc("Scan error").Error(err)
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

func (d *PgxStore) SchoolCreate(ctx context.Context, data *models.School) (*models.School, error) {
	qs, args := SchoolsCreateQuery(data)
	qs += " RETURNING uid"
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		err = tx.QueryRow(ctx, qs, args...).Scan(&data.ID)
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}

	editModel, err := d.SchoolsFindById(ctx, data.ID)
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}

	return editModel, nil
}

func (d *PgxStore) SchoolUpdate(ctx context.Context, data *models.School) (*models.School, error) {
	// origModel := d.UsersFindById(strconv.Itoa(int(model.ID)))
	qs, args := SchoolsUpdateQuery(data)
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		_, err = tx.Query(ctx, qs, args...)
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}
	editModel, err := d.SchoolsFindById(ctx, data.ID)
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}
	return editModel, nil
}

func (d *PgxStore) SchoolDelete(ctx context.Context, l []*models.School) ([]*models.School, error) {
	ids := []string{}
	for _, i := range l {
		ids = append(ids, i.ID)
	}
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		_, err = tx.Query(ctx, sqlSchoolDelete, (ids))
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}
	return l, nil
}

func SchoolsCreateQuery(m *models.School) (string, []interface{}) {
	args := []interface{}{}
	cols := ""
	vals := ""
	q := SchoolAtomicQuery(m, true)
	for k, v := range q {
		args = append(args, v)
		cols += ", " + k
		vals += ", $" + strconv.Itoa(len(args))
	}
	qs := sqlSchoolInsert + " (" + strings.Trim(cols, ", ") + ") VALUES (" + strings.Trim(vals, ", ") + ")"
	return qs, args
}

func SchoolsUpdateQuery(m *models.School) (string, []interface{}) {
	args := []interface{}{}
	sets := ""
	q := SchoolAtomicQuery(m, false)
	for k, v := range q {
		args = append(args, v)
		sets += ", " + k + "=$" + strconv.Itoa(len(args))
	}
	args = append(args, m.ID)
	qs := strings.ReplaceAll(sqlSchoolUpdate, "set uid=uid", "set uid=uid "+sets+" ") + "where uid=$" + strconv.Itoa(len(args))
	return qs, args
}

func (d *PgxStore) SchoolUpdateRelations(ctx context.Context, data *models.School, model *models.School) error {
	if data.Admin != nil {
		var r pgx.Rows
		err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
			r, err = tx.Query(ctx, sqlSchoolAdminDelete, data.ID)
			return
		})
		if err != nil {
			utils.LoggerDesc("Query error").Error(err)
			return err
		}
		r.Next()
		err = d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
			r, err = tx.Query(ctx, sqlSchoolAdminInsert, data.ID, data.Admin.ID)
			return
		})
		if err != nil {
			utils.LoggerDesc("Query error").Error(err)
			return err
		}
		r.Next()
	}
	return nil
}

func SchoolAtomicQuery(m *models.School, isCreate bool) map[string]interface{} {
	q := map[string]interface{}{}
	if m.Code != nil {
		q["code"] = m.Code
	}
	if m.Name != nil {
		q["name"] = m.Name
	}
	if m.FullName != nil {
		q["full_name"] = m.FullName
	}
	if m.Avatar != nil {
		q["avatar"] = m.Avatar
	}
	if m.Background != nil {
		q["archived_at"] = m.ArchivedAt
	}
	if m.Description != nil {
		q["description"] = m.Description
	}
	if m.Address != nil {
		q["address"] = m.Address
	}
	if m.Phone != nil {
		q["phone"] = m.Phone
	}
	if m.Email != nil {
		q["email"] = m.Email
	}
	if m.Level != nil {
		q["level"] = m.Level
	}
	if m.Galleries != nil {
		q["galleries"] = m.Galleries
	}
	if m.Latitude != nil {
		q["latitude"] = m.Latitude
	}
	if m.Longitude != nil {
		q["longitude"] = m.Longitude
	}
	if m.IsDigitalized != nil {
		q["is_digitalized"] = m.IsDigitalized
	}
	if m.IsSecondarySchool != nil {
		q["is_secondary_school"] = m.IsSecondarySchool
	}
	if m.ArchivedAt != nil {
		if m.ArchivedAt.Year() == 0 {
			q["archived_at"] = nil
		} else {
			q["archived_at"] = m.ArchivedAt
		}
	}
	if m.ParentUid != nil {
		q["parent_uid"] = m.ParentUid
	}
	if isCreate {
		q["created_at"] = time.Now()
	}
	if m.AdminUid != nil {
		q["admin_uid"] = m.AdminUid
	}
	if m.SpecialistUid != nil {
		q["specialist_uid"] = m.SpecialistUid
	}
	q["updated_at"] = time.Now()
	return q
}

func SchoolsListBuildQuery(f models.SchoolFilterRequest, args []interface{}) (string, []interface{}) {
	var wheres string = ""
	qs := sqlSchoolSelectMany
	groupBy := "group by s.uid"
	isJoinUsers := false
	joins := ""
	orderBy := ""
	orderByDir := "desc"
	orderByIsRating := false

	if f.Sort != nil && *f.Sort != "" {
		orderByDir = "desc"
		if strings.HasSuffix(*f.Sort, "~") {
			orderByDir = "asc"
			*f.Sort = strings.ReplaceAll(*f.Sort, "~", "")
		}
		orderBy = *f.Sort
		if *f.Sort == "rating" {
			orderByIsRating = true
			isJoinUsers = true
		}
	}

	if f.ID != nil && *f.ID != "" {
		args = append(args, *f.ID)
		wheres += " and s.uid=$" + strconv.Itoa(len(args))
	}
	if f.NotIds != nil && len(*f.NotIds) > 0 {
		args = append(args, *f.NotIds)
		wheres += " and s.uid <> ANY($" + strconv.Itoa(len(args)) + "::uuid[])"
	}
	if f.Code != nil && *f.Code != "" {
		args = append(args, *f.Code)
		wheres += " and s.code=$" + strconv.Itoa(len(args))
	}
	if f.Codes != nil && len(*f.Codes) > 0 {
		args = append(args, *f.Codes)
		wheres += " and s.code = ANY($" + strconv.Itoa(len(args)) + "::text[])"
	}
	if f.Uids != nil {
		args = append(args, *f.Uids)
		wheres += " and s.uid = ANY($" + strconv.Itoa(len(args)) + "::uuid[])"
	}
	if f.SpecialistId != nil {
		args = append(args, *f.SpecialistId)
		wheres += " and s.specialist_uid=$" + strconv.Itoa(len(args))
	}
	if f.ParentUid != nil && *f.ParentUid != "" {
		args = append(args, *f.ParentUid)
		wheres += " and s.parent_uid=$" + strconv.Itoa(len(args))
	}
	if f.ParentUids != nil {
		args = append(args, *f.ParentUids)
		wheres += " and s.parent_uid = ANY($" + strconv.Itoa(len(args)) + "::uuid[])"
	}
	if f.IsParent != nil {
		if *f.IsParent {
			wheres += " and s.parent_uid is null"
		} else {
			wheres += " and s.parent_uid is not null"
		}
	}
	if f.IsSecondarySchool != nil {
		if *f.IsSecondarySchool {
			wheres += " and s.is_secondary_school is true"
		} else {
			wheres += " and s.is_secondary_school is false"
		}
	}
	if f.Search != nil && *f.Search != "" {
		args = append(args, strings.ToLower(*f.Search))
		tmp := "$" + strconv.Itoa(len(args))
		wheres += " and (lower(s.name) like '%' ||" + tmp + "|| '%' or lower(s.code) like '%' ||" + tmp +
			"||'%' or lower(s.full_name) like '%'||" + tmp + "||'%' or lower(s.address) like '%'||" + tmp + "||'%')"
	}
	if isJoinUsers {
		joins += " left join user_schools us on (us.school_uid = s.uid)"
		joins += " left join users u on (u.uid = us.user_uid)"
	}

	if orderBy != "" {
		if orderByIsRating {
			orderBy = " order by count(us.*)"
			wheres += " and u.last_active > now() - INTERVAL '300 day'"
		} else {
			orderBy = " order by " + orderBy + " " + orderByDir
		}
	} else {
		orderBy = "order by LPAD(replace(lower(s.code), 'ag', ''), 3, '0'), s.parent_uid, CAST(REVERSE(SUBSTRING(REVERSE(code) FROM '[0-9]+')) AS INTEGER), s.code"
	}

	qs = strings.ReplaceAll(qs, "s.uid=s.uid", "s.uid=s.uid "+wheres+" "+groupBy+" "+orderBy+" ")
	qs = strings.ReplaceAll(qs, "where", joins+" where")
	return qs, args
}

func (d *PgxStore) SchoolsLoadParents(ctx context.Context, l *[]*models.School) error {
	ids := []string{}
	for _, m := range *l {
		ids = append(ids, m.ID)
	}
	if len(ids) < 1 {
		return nil
	}
	if rs, err := d.SchoolsLoadParent(ctx, ids); err == nil {
		for _, r := range rs {
			for _, m := range *l {
				if r.UID == m.ID {
					m.Parent = r.Relation
				}
			}
		}
	} else {
		return err
	}
	return nil
}

func (d *PgxStore) SchoolsLoadRelations(ctx context.Context, l *[]*models.School) error {
	ids := []string{}
	for _, m := range *l {
		ids = append(ids, m.ID)
	}
	if len(ids) < 1 {
		return nil
	}
	// load admin
	if rs, err := d.SchoolsLoadAdmin(ctx, ids); err == nil {
		for _, r := range rs {
			for _, m := range *l {
				if r.UID == m.ID {
					m.Admin = r.Relation
				}
			}
		}
	} else {
		return err
	}
	// load specialist
	if rs, err := d.SchoolsLoadSpecialist(ctx, ids); err == nil {
		for _, r := range rs {
			for _, m := range *l {
				if r.UID == m.ID {
					m.Specialist = r.Relation
				}
			}
		}
	} else {
		return err
	}
	// load parent
	if rs, err := d.SchoolsLoadParent(ctx, ids); err == nil {
		for _, r := range rs {
			for _, m := range *l {
				if r.UID == m.ID {
					m.Parent = r.Relation
				}
			}
		}
	} else {
		return err
	}

	return nil
}

type SchoolsLoadAdminItem struct {
	UID      string
	Relation *models.User
}

func (d *PgxStore) SchoolsLoadAdmin(ctx context.Context, ids []string) ([]SchoolsLoadAdminItem, error) {
	res := []SchoolsLoadAdminItem{}
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		rows, err := tx.Query(ctx, sqlSchoolAdmin, (ids))
		for rows.Next() {
			sub := models.User{}
			pid := string("")
			err = scanUser(rows, &sub, &pid)
			if err != nil {
				return err
			}
			res = append(res, SchoolsLoadAdminItem{UID: pid, Relation: &sub})
		}
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}

	return res, nil
}

type SchoolsLoadSpecialistItem struct {
	UID      string
	Relation *models.User
}

func (d *PgxStore) SchoolsLoadSpecialist(ctx context.Context, ids []string) ([]SchoolsLoadSpecialistItem, error) {
	res := []SchoolsLoadSpecialistItem{}
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		rows, err := tx.Query(ctx, sqlSchoolSpecialist, (ids))
		for rows.Next() {
			sub := models.User{}
			pid := string("")
			err = scanUser(rows, &sub, &pid)
			if err != nil {
				return err
			}
			res = append(res, SchoolsLoadSpecialistItem{UID: pid, Relation: &sub})
		}
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}

	return res, nil
}

type SchoolsLoadParentItem struct {
	UID      string
	Relation *models.School
}

func (d *PgxStore) SchoolsLoadParent(ctx context.Context, ids []string) ([]SchoolsLoadParentItem, error) {
	res := []SchoolsLoadParentItem{}
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		rows, err := tx.Query(ctx, sqlSchoolParent, (ids))
		for rows.Next() {
			sub := models.School{}
			pid := string("")
			err = scanSchool(rows, &sub, &pid)
			if err != nil {
				return err
			}
			res = append(res, SchoolsLoadParentItem{UID: pid, Relation: &sub})
		}
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}

	return res, nil
}
