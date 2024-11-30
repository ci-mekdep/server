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

const sqlSmsSenderFields = `ss.uid, ss.phones, ss.message, ss.type, ss.error_msg, ss.is_completed, ss.left_try, ss.tried_at, ss.created_at`
const sqlSmsSenderSelect = `SELECT ` + sqlSmsSenderFields + ` FROM sms_sender ss WHERE ss.uid = ANY($1::uuid[])`
const sqlSmsSenderSelectMany = `SELECT ` + sqlSmsSenderFields + `, count(*) over() as total FROM sms_sender ss where ss.uid=ss.uid limit $1 offset $2 `
const sqlSmsSenderInsert = `INSERT INTO sms_sender`

func scanSmsSender(rows pgx.Row, m *models.SmsSender, addColumns ...interface{}) (err error) {
	err = rows.Scan(parseColumnsForScan(m, addColumns...)...)
	return
}

func (d *PgxStore) SmsSendersFindBy(ctx context.Context, f models.SmsSenderFilterRequest) (smsSenders []*models.SmsSender, total int, err error) {
	args := []interface{}{f.Limit, f.Offset}
	qs, args := SmsSendersListBuildQuery(f, args)
	err = d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		rows, err := tx.Query(ctx, qs, args...)
		for rows.Next() {
			smsSender := models.SmsSender{}
			err = scanSmsSender(rows, &smsSender, &total)
			if err != nil {
				return err
			}
			smsSenders = append(smsSenders, &smsSender)
		}
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, 0, err
	}

	return smsSenders, total, nil
}

func (d *PgxStore) SmsSendersFindById(ctx context.Context, ID string) (*models.SmsSender, error) {
	row, err := d.SmsSendersFindByIds(ctx, []string{ID})
	if err != nil {
		return nil, err
	}
	if len(row) < 1 {
		return nil, errors.New("sms_sender not found by uid: " + ID)
	}
	return row[0], nil
}

func (d *PgxStore) SmsSendersFindByIds(ctx context.Context, IDs []string) ([]*models.SmsSender, error) {
	smsSenders := []*models.SmsSender{}
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		rows, err := tx.Query(ctx, sqlSmsSenderSelect, (IDs))
		for rows.Next() {
			smsSender := models.SmsSender{}
			err := scanSmsSender(rows, &smsSender)
			if err != nil {
				return err
			}
			smsSenders = append(smsSenders, &smsSender)
		}
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}

	return smsSenders, nil
}

func (d *PgxStore) SmsSenderCreate(ctx context.Context, model *models.SmsSender) (*models.SmsSender, error) {
	qs, args := SmsSenderCreateQuery(model)
	qs += " RETURNING uid"
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		err = tx.QueryRow(ctx, qs, args...).Scan(&model.ID)
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}
	editModel, err := d.SmsSendersFindById(ctx, model.ID)
	if err != nil {
		return nil, err
	}
	return editModel, nil
}

func SmsSenderCreateQuery(m *models.SmsSender) (string, []interface{}) {
	args := []interface{}{}
	cols := ""
	vals := ""
	q := SmsSenderAtomicQuery(m, true)
	for k, v := range q {
		args = append(args, v)
		cols += ", " + k
		vals += ", $" + strconv.Itoa(len(args))
	}
	qs := sqlSmsSenderInsert + " (" + strings.Trim(cols, ", ") + ") VALUES (" + strings.Trim(vals, ", ") + ")"
	return qs, args
}

func SmsSenderAtomicQuery(m *models.SmsSender, isCreate bool) map[string]interface{} {
	q := map[string]interface{}{}
	if m.Phones != nil {
		q["phones"] = m.Phones
	}
	if m.Message != "" {
		q["message"] = m.Message
	}
	if m.Type != "" {
		q["type"] = m.Type
	}
	if m.ErrorMsg != nil {
		q["error_msg"] = m.Type
	}
	q["is_completed"] = m.IsCompleted
	q["left_try"] = m.LeftTry
	if isCreate {
		q["created_at"] = time.Now()
	}
	q["tried_at"] = time.Now()
	return q
}

func SmsSendersListBuildQuery(f models.SmsSenderFilterRequest, args []interface{}) (string, []interface{}) {
	var wheres string = ""

	if f.ID != nil && *f.ID != "" {
		args = append(args, *f.ID)
		wheres += " and ss.uid=$" + strconv.Itoa(len(args))
	}
	wheres += " group by ss.uid "
	wheres += " order by ss.created_at desc"

	qs := sqlSmsSenderSelectMany
	qs = strings.ReplaceAll(qs, "ss.uid=ss.uid", "ss.uid=ss.uid "+wheres+" ")

	return qs, args
}
