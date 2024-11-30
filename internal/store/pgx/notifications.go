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

const sqlNotificationFields = `nn.uid, nn.school_uids, nn.roles, nn.user_uids, nn.author_uid, nn.title, nn.content, nn.files, nn.updated_at, nn.created_at`

const sqlNotificationsInsert = `insert into notifications`

const sqlNotificationUpdate = `update notifications nn set uid=uid`

const sqlNotificationDelete = `delete from notifications where uid = $1`

const sqlUserNotificationDelete = `delete from user_notifications where notification_uid = $1`

const sqlNotificationSelect = `select ` + sqlNotificationFields + ` from notifications nn where uid = ANY($1::uuid[])`

const sqlNotificationSelectMany = `select ` + sqlNotificationFields + `, count(*) over() as total from notifications nn where nn.uid=nn.uid limit $1 offset $2`

const sqlNotificationSchool = `select ` + sqlSchoolFields + `, nn.uid from notifications nn
	right join schools s on (s.uid = ANY(nn.school_uids::uuid[])) where nn.uid = ANY($1::uuid[])`

const sqlNotificationAuthor = `select ` + sqlUserFields + `, nn.uid from notifications nn
	right join users u on (u.uid = nn.author_uid) where nn.uid = ANY($1::uuid[])`

func scanNotifications(rows pgx.Rows, m *models.Notifications, addColumns ...interface{}) (err error) {
	err = rows.Scan(parseColumnsForScan(m, addColumns...)...)
	return
}

func (d *PgxStore) NotificationsFindBy(ctx context.Context, f models.NotificationsFilterRequest) (notifications []*models.Notifications, total int, err error) {
	args := []interface{}{f.Limit, f.Offset}
	qs, args := NotificationsListBuildQuery(f, args)
	err = d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		rows, err := tx.Query(ctx, qs, args...)
		for rows.Next() {
			sub := models.Notifications{}
			err = scanNotifications(rows, &sub, &total)
			if err != nil {
				utils.LoggerDesc("Scan error").Error(err)
				return err
			}
			notifications = append(notifications, &sub)
		}
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, 0, err
	}

	return notifications, total, nil
}

func (d *PgxStore) NotificationFindById(ctx context.Context, ID string) (*models.Notifications, error) {
	row, err := d.NotificationFindByIds(ctx, []string{ID})
	if err != nil {
		return nil, err
	}
	if len(row) < 1 {
		return nil, errors.New("notification not found by uid: " + ID)
	}
	return row[0], nil
}

func (d *PgxStore) NotificationFindByIds(ctx context.Context, Ids []string) ([]*models.Notifications, error) {
	notifications := []*models.Notifications{}
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		rows, err := tx.Query(ctx, sqlNotificationSelect, (Ids))
		for rows.Next() {
			notification := models.Notifications{}
			err := scanNotifications(rows, &notification)
			if err != nil {
				utils.LoggerDesc("Scan error").Error(err)
				return err
			}
			notifications = append(notifications, &notification)
		}
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}

	return notifications, nil
}

func (d *PgxStore) NotificationCreate(ctx context.Context, model *models.Notifications) (*models.Notifications, error) {
	qs, args := NotificationCreateQuery(model)
	qs += " RETURNING uid"
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		err = tx.QueryRow(ctx, qs, args...).Scan(&model.ID)
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}
	editModel, err := d.NotificationFindById(ctx, model.ID)
	if err != nil {
		return nil, err
	}
	return editModel, nil
}

func (d *PgxStore) NotificationUpdate(ctx context.Context, model *models.Notifications) (*models.Notifications, error) {
	qs, args := NotificationUpdateQuery(model)
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		_, err = tx.Query(ctx, qs, args...)
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}
	editModel, err := d.NotificationFindById(ctx, model.ID)
	if err != nil {
		return nil, err
	}
	return editModel, nil
}

func (d *PgxStore) NotificationDelete(ctx context.Context, ID string) error {
	// Begin a transaction
	tx, err := d.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) // Rollback if an error occurs

	// Delete related entries in user_notifications table
	_, err = tx.Exec(ctx, sqlUserNotificationDelete, ID)
	if err != nil {
		return err
	}

	// Delete the notification itself
	_, err = tx.Exec(ctx, sqlNotificationDelete, ID)
	if err != nil {
		return err
	}

	// Commit the transaction
	err = tx.Commit(ctx)
	if err != nil {
		return err
	}

	return nil
}
func NotificationCreateQuery(m *models.Notifications) (string, []interface{}) {
	args := []interface{}{}
	cols := ""
	vals := ""
	q := NotificationAtomicQuery(m, true)
	for k, v := range q {
		args = append(args, v)
		cols += ", " + k
		vals += ", $" + strconv.Itoa(len(args))
	}
	qs := sqlNotificationsInsert + " (" + strings.Trim(cols, ", ") + ") VALUES (" + strings.Trim(vals, ", ") + ")"
	return qs, args
}

func NotificationUpdateQuery(m *models.Notifications) (string, []interface{}) {
	args := []interface{}{}
	sets := ""
	q := NotificationAtomicQuery(m, false)
	for k, v := range q {
		args = append(args, v)
		sets += ", " + k + "=$" + strconv.Itoa(len(args))
	}
	args = append(args, m.ID)
	qs := strings.ReplaceAll(sqlNotificationUpdate, "set uid=uid", "set uid=uid"+sets+" ") + "where uid=$" + strconv.Itoa(len(args))
	return qs, args
}

func NotificationAtomicQuery(m *models.Notifications, isCreate bool) map[string]interface{} {
	q := map[string]interface{}{}
	if m.SchoolIds != nil {
		q["school_uids"] = m.SchoolIds
	}
	if m.UserIds != nil {
		q["user_uids"] = m.UserIds
	}
	if m.Roles != nil {
		q["roles"] = m.Roles
	}
	if m.AuthorID != nil {
		q["author_uid"] = m.AuthorID
	}
	if m.Title != nil {
		q["title"] = *m.Title
	}
	if m.Content != nil {
		q["content"] = *m.Content
	}
	if m.Files != nil {
		q["files"] = *m.Files
	}
	if isCreate {
		q["created_at"] = time.Now()
	}
	q["updated_at"] = time.Now()
	return q
}

func NotificationsListBuildQuery(f models.NotificationsFilterRequest, args []interface{}) (string, []interface{}) {
	var wheres string = ""

	if f.ID != nil && *f.ID != "" {
		args = append(args, *f.ID)
		wheres += " and uid=$" + strconv.Itoa(len(args))
	}
	if f.UserIds != nil && len(*f.UserIds) != 0 {
		args = append(args, f.UserIds)
		wheres += " and user_uid = ANY($" + strconv.Itoa(len(args)) + "::uuid[])"
	}
	if f.AuthorId != nil && *f.AuthorId != "" {
		args = append(args, *f.AuthorId)
		wheres += " and author_uid=$" + strconv.Itoa(len(args))
	}
	if f.Search != nil && *f.Search != "" {
		*f.Search = strings.ToLower(*f.Search)
		args = append(args, *f.Search)
		tmp := "$" + strconv.Itoa(len(args))
		wheres += " and (lower(name) like '%" + tmp + "%' or lower(full_name) like '%" + tmp + "%')"
	}
	if f.Sort != nil && *f.Sort != "" {
		dir := "desc"
		if strings.HasSuffix(*f.Sort, "~") {
			dir = "asc"
		}
		*f.Sort = strings.ReplaceAll(*f.Sort, "~", "")
		wheres += " order by " + *f.Sort + " " + dir
	} else {
		wheres += " order by nn.created_at desc"
	}

	qs := sqlNotificationSelectMany
	qs = strings.ReplaceAll(qs, "nn.uid=nn.uid", "nn.uid=nn.uid "+wheres+" ")

	return qs, args
}

func (d *PgxStore) NotificationsLoadRelations(ctx context.Context, l *[]*models.Notifications) error {
	ids := []string{}
	for _, m := range *l {
		ids = append(ids, m.ID)
	}
	if len(ids) < 1 {
		return nil
	}

	if rs, err := d.NotificationsLoadSchools(ctx, ids); err != nil {
		return err
	} else {
		for _, r := range rs {
			for _, m := range *l {
				if m.Schools == nil {
					m.Schools = &[]models.School{}
				}
				if r.ID == m.ID && r.Relation != nil {
					*m.Schools = append(*m.Schools, *r.Relation)
				}
			}
		}
	}
	if rs, err := d.NotificationsLoadAuthor(ctx, ids); err != nil {
		return err
	} else {
		for _, r := range rs {
			for _, m := range *l {
				if r.ID == m.ID && r.Relation != nil {
					m.Author = r.Relation
				}
			}
		}
	}
	return nil
}

type NotificationsLoadSchoolItem struct {
	ID       string
	Relation *models.School
}

func (d *PgxStore) NotificationsLoadSchools(ctx context.Context, ids []string) ([]NotificationsLoadSchoolItem, error) {
	res := []NotificationsLoadSchoolItem{}

	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		rows, err := tx.Query(ctx, sqlNotificationSchool, ids)
		for rows.Next() {
			sub := models.School{}
			pid := ""
			err = scanSchool(rows, &sub, &pid)
			if err != nil {
				utils.LoggerDesc("Scan error").Error(err)
				return err
			}
			res = append(res, NotificationsLoadSchoolItem{ID: pid, Relation: &sub})
		}
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}

	return res, nil
}

type NotificationsLoadAuthorItem struct {
	ID       string
	Relation *models.User
}

func (d *PgxStore) NotificationsLoadAuthor(ctx context.Context, ids []string) ([]NotificationsLoadAuthorItem, error) {
	res := []NotificationsLoadAuthorItem{}

	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		rows, err := tx.Query(ctx, sqlNotificationAuthor, ids)
		for rows.Next() {
			sub := models.User{}
			pid := ""
			err = scanUser(rows, &sub, &pid)
			if err != nil {
				utils.LoggerDesc("Scan error").Error(err)
				return err
			}
			res = append(res, NotificationsLoadAuthorItem{ID: pid, Relation: &sub})
		}
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}

	return res, nil
}
