package pgx

import (
	"context"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/mekdep/server/config"
	"github.com/mekdep/server/internal/models"
	"github.com/mekdep/server/internal/utils"
)

// TODO change this
const CODE_EXPIRE_MIN = "10"

var CODE_EXPIRE_MIN_ALT = 10 * time.Minute

const SQL_CONFIRM_CODE_FIELDS = `cc.uid, cc.user_uid, cc.phone, cc.email, cc.code, cc.expire_at`
const SQL_CONFIRM_CODE_INSERT = `insert into confirm_codes (user_uid, phone, code, expire_at) values ($1, $2, $3, current_timestamp + (` + CODE_EXPIRE_MIN + ` * interval '1 minute'))`
const SQL_CONFIRM_CODE_INSERT_ALT = `insert into confirm_codes (user_uid, phone, code, expire_at) values ($1, $2, $3, $4))`
const SQL_CONFIRM_CODE_LIST = `select cc.uid, cc.user_uid 
	from confirm_codes cc where uid=uid and cc.phone = $1 and cc.code = $2`

// TODO remove current_timestamp since it is a part of business logic
const SQL_CONFIRM_CODE_CLEAR = `delete from confirm_codes where phone=$1 and expire_at < current_timestamp`
const SQL_CONFIRM_CODE_DELETE = `delete from confirm_codes where uid=$1`

const sqlSessionFields = `ss.uid, ss.token, ss.user_uid, ss.device_token, ss.agent, ss.ip, ss.iat, ss.exp, ss.lat`
const sqlSessionSelectMany = `select ` + sqlSessionFields + ` from sessions ss where uid=uid`
const sqlSessionDeleteMany = `delete from sessions ss where uid=uid`

var sqlSessionInsert = `insert into sessions  
	(token, user_uid, device_token, agent, ip, iat, exp, lat) values ($1, $2, $3, $4, $5, $6, $7, $8) 
	RETURNING ` + strings.ReplaceAll(sqlSessionFields, "ss.", "")

func scanSession(rows pgx.Row, m *models.Session, addColumns ...interface{}) (err error) {
	err = rows.Scan(parseColumnsForScan(m, addColumns...)...)
	return
}

func (d *PgxStore) ConfirmCodeGenerate(ctx context.Context, m *models.User) (string, error) {
	code := rand.Intn(89999) + 10000
	phone, err := m.FormattedPhone()
	if err != nil {
		return "", nil
	}

	err = d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		_, err = tx.Query(ctx, SQL_CONFIRM_CODE_INSERT, m.ID, phone, strconv.Itoa(code))
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return "", err
	}

	return strconv.Itoa(code), nil
}

func (d *PgxStore) ConfirmCodeClear(ctx context.Context, m *models.User) error {
	phone, err := m.FormattedPhone()
	if err != nil {
		return nil
	}

	err = d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		_, err = tx.Query(ctx, SQL_CONFIRM_CODE_CLEAR, phone)
		return
	})
	if err != nil {
		if err != pgx.ErrNoRows {
			utils.LoggerDesc("Query error").Error(err)
		}
		return err
	}

	return nil
}

func (d *PgxStore) ConfirmCodeDelete(ctx context.Context, id string) error {
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		_, err = tx.Query(ctx, SQL_CONFIRM_CODE_DELETE, id)
		return
	})
	if err != nil {
		if err != pgx.ErrNoRows {
			utils.LoggerDesc("Query error").Error(err)
		}
		return err
	}
	return nil
}

func (d *PgxStore) CheckConfirmCode(ctx context.Context, m *models.User, code string) (string, error) {
	err := d.ConfirmCodeClear(ctx, m)
	if err != nil {
		return "", err
	}

	phone, err := m.FormattedPhone()
	if err != nil {
		return "", nil
	}

	if !*config.Conf.OtpValidationEnabled {
		return "", nil
	}

	var userId, codeId string

	err = d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		err = tx.QueryRow(ctx, SQL_CONFIRM_CODE_LIST, phone, code).Scan(&codeId, &userId)
		return
	})
	if err != nil {
		if err != pgx.ErrNoRows {
			utils.LoggerDesc("Query error").Error(err)
		}
		return "", err
	}

	err = d.ConfirmCodeDelete(ctx, codeId)
	if err != nil {
		return "", err
	}
	return userId, nil
}

func (d *PgxStore) SessionsSelect(ctx context.Context, f models.SessionFilter) ([]models.Session, error) {
	args := []interface{}{}
	qs, args := sessionsBuildWhere(f, args, sqlSessionSelectMany)
	l := []models.Session{}
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		rows, err := tx.Query(ctx, qs, args...)
		for rows.Next() {
			sub := models.Session{}
			err = scanSession(rows, &sub)
			if err != nil {
				return err
			}
			l = append(l, sub)
		}
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}
	return l, nil
}

func (d *PgxStore) SessionsDelete(ctx context.Context, f models.SessionFilter) error {
	args := []interface{}{}
	qs, args := sessionsBuildWhere(f, args, sqlSessionDeleteMany)

	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		_, err = tx.Query(ctx, qs, args...)
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return err
	}
	return nil
}

func (d *PgxStore) SessionsCreate(ctx context.Context, m models.Session) (models.Session, error) {
	mm := models.Session{}
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		row := tx.QueryRow(ctx, sqlSessionInsert, m.Token, m.UserId, m.DeviceToken, m.Agent, m.Ip, m.Iat, m.Exp, m.Lat)
		err = scanSession(row, &mm)
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return mm, err
	}
	return mm, nil
}

func (d *PgxStore) SessionsClear(ctx context.Context, now time.Time) error {
	return d.SessionsDelete(ctx, models.SessionFilter{
		Exp: &now,
	})
}
func sessionsBuildWhere(f models.SessionFilter, args []interface{}, qs string) (string, []interface{}) {
	var wheres string = ""

	if f.ID != nil {
		args = append(args, *f.ID)
		wheres += " and uid=$" + strconv.Itoa(len(args))
	}
	if f.DeviceToken != nil && *f.DeviceToken != "" {
		args = append(args, *f.DeviceToken)
		wheres += " and device_token=$" + strconv.Itoa(len(args))
	}
	if f.Token != nil && *f.Token != "" {
		args = append(args, *f.Token)
		wheres += " and token=$" + strconv.Itoa(len(args))
	}
	if f.UserId != nil {
		args = append(args, *f.UserId)
		wheres += " and user_uid=$" + strconv.Itoa(len(args))
	}
	if f.Ip != nil {
		args = append(args, *f.Ip)
		wheres += " and ip=$" + strconv.Itoa(len(args))
	}
	if f.Exp != nil {
		args = append(args, *f.Exp)
		wheres += " and exp <= $" + strconv.Itoa(len(args))
	}
	if f.Lat != nil {
		args = append(args, *f.Lat)
		wheres += " and lat >= $" + strconv.Itoa(len(args))
	}

	qs = strings.ReplaceAll(qs, "uid=uid", "uid=uid "+wheres+" ")
	return qs, args
}
