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

const sqlUserNotificationFields = `un.uid, un.notification_uid, un.user_uid, un.read_at, un.role, un.comment, un.comment_files`

const sqlUserNotificationsInsert = `insert into user_notifications`

const sqlUserNotificationUpdate = `update user_notifications set uid=uid`
const sqlUserNotificationUpdateRead = `update user_notifications set read_at=$2 where read_at is null and uid=ANY($1::uuid[])`
const sqlUserNotificationSelectTotalUnread = `select count(uid) from user_notifications where user_uid=$1 and role=$2 and read_at is null`

const sqlUserNotificationSelect = `select ` + sqlUserNotificationFields + ` from user_notifications un where un.uid = ANY($1::uuid[])`

const sqlUserNotificationSelectMany = `select ` + sqlUserNotificationFields + `, count(*) over() as total from user_notifications un 
	join notifications nn on un.notification_uid = nn.uid
	where un.uid=un.uid limit $1 offset $2 `

const sqlUserNotificationRelation = `select ` + sqlNotificationFields + `, un.uid from user_notifications un
	right join notifications nn on (nn.uid=un.notification_uid) where un.uid = ANY($1::uuid[])`

const sqlUserNotificationRelationUser = `select ` + sqlUserFields + `, un.uid from user_notifications un
	right join users u on (u.uid=un.user_uid) where un.uid = ANY($1::uuid[])`

func scanUserNotifications(rows pgx.Rows, m *models.UserNotification, addColumns ...interface{}) (err error) {
	err = rows.Scan(parseColumnsForScan(m, addColumns...)...)
	return
}

func (d *PgxStore) UserNotificationsFindBy(ctx context.Context, f models.UserNotificationFilterRequest) (userNotifications []*models.UserNotification, total int, err error) {
	args := []interface{}{f.Limit, f.Offset}
	qs, args := UserNotificationsListBuildQuery(f, args)
	err = d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		rows, err := tx.Query(ctx, qs, args...)
		for rows.Next() {
			userNotification := models.UserNotification{}
			err = scanUserNotifications(rows, &userNotification, &total)
			if err != nil {
				utils.LoggerDesc("Scan error").Error(err)
				return err
			}
			userNotifications = append(userNotifications, &userNotification)
		}
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, 0, err
	}

	return userNotifications, total, nil
}

func (d *PgxStore) UserNotificationFindById(ctx context.Context, ID string) (*models.UserNotification, error) {
	row, err := d.UserNotificationFindByIds(ctx, []string{ID})
	if err != nil {
		return nil, err
	}
	if len(row) < 1 {
		return nil, errors.New("user_notification not found by uid: " + ID)
	}
	return row[0], nil
}

func (d *PgxStore) UserNotificationFindByIds(ctx context.Context, Ids []string) ([]*models.UserNotification, error) {
	userNotifications := []*models.UserNotification{}

	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		rows, err := tx.Query(ctx, sqlUserNotificationSelect, (Ids))
		for rows.Next() {
			userNotification := models.UserNotification{}
			err := scanUserNotifications(rows, &userNotification)
			if err != nil {
				utils.LoggerDesc("Scan error").Error(err)
				return err
			}
			userNotifications = append(userNotifications, &userNotification)
		}
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}

	return userNotifications, nil
}

func (d *PgxStore) UserNotificationsUpdate(ctx context.Context, model models.UserNotification) (*models.UserNotification, error) {
	qs, args := UserNotificationsUpdateQuery(model)
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		_, err = tx.Query(ctx, qs, args...)
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}
	editModel, err := d.UserNotificationFindById(ctx, model.ID)
	if err != nil {
		return nil, err
	}
	return editModel, nil
}

func (d *PgxStore) UserNotificationsUpdateRead(ctx context.Context, ids []string) error {
	now := time.Now()
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		_, err = tx.Query(ctx, sqlUserNotificationUpdateRead, ids, now)
		return
	})
	return err
}

func (d *PgxStore) UserNotificationsSelectTotalUnread(ctx context.Context, userId string, role string) (int, error) {
	total := 0
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		err = tx.QueryRow(ctx, sqlUserNotificationSelectTotalUnread, userId, role).Scan(&total)
		return
	})
	return total, err
}

func (d *PgxStore) UserNotificationsCreateBatch(ctx context.Context, l []models.UserNotification) error {
	err := d.UserNotificationsBatch(ctx, l, UserNotificationCreateQuery)
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return err
	}
	return err
}

func (d *PgxStore) UserNotificationsBatch(ctx context.Context, l []models.UserNotification, f func(models.UserNotification) (string, []interface{})) error {
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

func UserNotificationCreateQuery(m models.UserNotification) (string, []interface{}) {
	args := []interface{}{}
	cols := ""
	vals := ""
	q := UserNotificationAtomicQuery(m, true)
	for k, v := range q {
		args = append(args, v)
		cols += ", " + k
		vals += ", $" + strconv.Itoa(len(args))
	}
	qs := sqlUserNotificationsInsert + " (" + strings.Trim(cols, ", ") + ") VALUES (" + strings.Trim(vals, ", ") + ")"
	return qs, args
}

func UserNotificationsUpdateQuery(m models.UserNotification) (string, []interface{}) {
	args := []interface{}{}
	sets := ""
	q := UserNotificationAtomicQuery(m, false)
	for k, v := range q {
		args = append(args, v)
		sets += ", " + k + "=$" + strconv.Itoa(len(args))
	}
	args = append(args, m.ID)
	qs := strings.ReplaceAll(sqlUserNotificationUpdate, "set uid=uid", "set uid=uid"+sets+" ") + "where uid=$" + strconv.Itoa(len(args))
	return qs, args
}

func UserNotificationAtomicQuery(m models.UserNotification, isCreate bool) map[string]interface{} {
	q := map[string]interface{}{}
	if m.NotificationId != "" {
		q["notification_uid"] = m.NotificationId
	}
	if m.UserId != "" {
		q["user_uid"] = m.UserId
	}
	if m.Role != nil {
		q["role"] = m.Role
	}
	if m.Comment != nil {
		q["comment"] = m.Comment
	}
	if m.CommentFiles != nil {
		q["comment_files"] = *m.CommentFiles
	}
	if m.ReadAt != nil {
		q["read_at"] = *m.ReadAt
	}
	if isCreate {
		q["read_at"] = nil
	}
	return q
}

func UserNotificationsListBuildQuery(f models.UserNotificationFilterRequest, args []interface{}) (string, []interface{}) {
	var wheres string = ""

	if f.IDs != nil && len(*f.IDs) > 0 {
		args = append(args, *f.IDs)
		wheres += " and un.uid= ANY($" + strconv.Itoa(len(args)) + "::uuid[])"
	}
	if f.NotificationId != nil && *f.NotificationId != "" {
		args = append(args, *f.NotificationId)
		wheres += " and un.notification_uid=$" + strconv.Itoa(len(args))
	}
	if f.UserId != nil && *f.UserId != "" {
		args = append(args, *f.UserId)
		wheres += " and un.user_uid=$" + strconv.Itoa(len(args))
	}
	if f.Role != nil && *f.Role != "" {
		args = append(args, *f.Role)
		wheres += " and un.role=$" + strconv.Itoa(len(args))
	}
	wheres += " group by un.uid, nn.created_at "
	if f.Sort != nil && *f.Sort != "" {
		dir := "desc"
		if strings.HasSuffix(*f.Sort, "~") {
			dir = "asc"
		}
		*f.Sort = strings.ReplaceAll(*f.Sort, "~", "")
		wheres += " order by nn.created_at " + dir + ", un.read_at is null, un.uid desc"
	} else {
		wheres += " order by nn.created_at desc, un.read_at is null, un.uid desc"
	}
	qs := sqlUserNotificationSelectMany
	qs = strings.ReplaceAll(qs, "un.uid=un.uid", "un.uid=un.uid "+wheres+" ")

	return qs, args
}

func (d *PgxStore) UserNotificationsLoadRelations(ctx context.Context, l *[]*models.UserNotification) error {
	ids := []string{}
	for _, m := range *l {
		ids = append(ids, m.ID)
	}
	if len(ids) < 1 {
		return nil
	}

	if rs, err := d.UserNotificationsLoadNotification(ctx, ids); err != nil {
		return err
	} else {
		for _, r := range rs {
			for _, m := range *l {
				if r.ID == m.ID {
					m.Notifications = r.Relation
				}
			}
		}
	}
	return nil
}
func (d *PgxStore) UserNotificationsLoadRelationUser(l *[]*models.UserNotification) error {
	ids := []string{}
	for _, m := range *l {
		ids = append(ids, m.ID)
	}
	if len(ids) < 1 {
		return nil
	}

	if rs, err := d.UserNotificationLoadUser(ids); err != nil {
		return err
	} else {
		for _, r := range rs {
			for _, m := range *l {
				if r.ID == m.ID {
					m.User = r.Relation
				}
			}
		}
	}
	return nil
}

type UserNotLoadNotificationItem struct {
	ID       string
	Relation *models.Notifications
}

func (d *PgxStore) UserNotificationsLoadNotification(ctx context.Context, ids []string) ([]UserNotLoadNotificationItem, error) {
	res := []UserNotLoadNotificationItem{}
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		rows, err := tx.Query(ctx, sqlUserNotificationRelation, ids)
		for rows.Next() {
			sub := models.Notifications{}
			pid := ""
			err = scanNotifications(rows, &sub, &pid)
			if err != nil {
				return err
			}
			res = append(res, UserNotLoadNotificationItem{ID: pid, Relation: &sub})
		}
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}
	return res, nil
}

type UserNotificationLoadUserItem struct {
	ID       string
	Relation *models.User
}

func (d *PgxStore) UserNotificationLoadUser(ids []string) ([]UserNotificationLoadUserItem, error) {
	res := []UserNotificationLoadUserItem{}
	err := d.runQuery(context.Background(), func(tx *pgxpool.Conn) (err error) {
		rows, err := tx.Query(context.Background(), sqlUserNotificationRelationUser, ids)
		for rows.Next() {
			sub := models.User{}
			pid := ""
			err = scanUser(rows, &sub, &pid)
			if err != nil {
				return err
			}
			res = append(res, UserNotificationLoadUserItem{ID: pid, Relation: &sub})
		}
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}
	return res, nil
}
