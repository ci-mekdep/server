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
	sqlSchoolSettingFields = `st.school_uid, st.key, st.value, st.updated_at`
	sqlSchoolSettingSelect = `select ` + sqlSchoolSettingFields + ` from school_settings st where st.school_uid=ANY($1::uuid[])`
	sqlSchoolSettingUpsert = `insert into school_settings (key, value, school_uid, updated_at) VALUES ON CONFLICT(key,school_uid) DO update set value=EXCLUDED.value, updated_at=EXCLUDED.updated_at;`
)

func scanSchoolSetting(rows pgx.Row, m *models.SchoolSetting, addColumns ...interface{}) (err error) {
	err = rows.Scan(parseColumnsForScan(m, addColumns...)...)
	return
}

func (d *PgxStore) SchoolSettingsGet(ctx context.Context, schoolIds []string) ([]models.SchoolSetting, error) {
	l := []models.SchoolSetting{}
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		rows, err := tx.Query(ctx, sqlSchoolSettingSelect, schoolIds)
		for rows.Next() {
			sub := models.SchoolSetting{}
			err = scanSchoolSetting(rows, &sub)
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
		return nil, err
	}

	return l, nil
}

func (d *PgxStore) SchoolSettingsUpdate(ctx context.Context, schoolId string, values []models.SchoolSettingRequest) error {
	qs, args := d.SchoolSettingsUpdateQuery(ctx, values)
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		_, err = tx.Exec(ctx, qs, args...)
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return err
	}
	return nil
}

func (d *PgxStore) SchoolSettingsUpdateQuery(ctx context.Context, m []models.SchoolSettingRequest) (string, []interface{}) {
	args := []interface{}{}
	values := ""
	for _, v := range m {
		values += ", ($" + strconv.Itoa(len(args)+1) + ", $" + strconv.Itoa(len(args)+2) + ", $" + strconv.Itoa(len(args)+3) + ", $" + strconv.Itoa(len(args)+4) + ")"
		args = append(args, v.Key)
		args = append(args, v.Value)
		args = append(args, v.SchoolId)
		args = append(args, time.Now())
	}
	values = strings.Trim(values, ", ")
	qs := strings.ReplaceAll(sqlSchoolSettingUpsert, "VALUES ON", "VALUES "+values+" ON")
	return qs, args
}
