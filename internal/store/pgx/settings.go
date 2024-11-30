package pgx

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/mekdep/server/internal/models"
	"github.com/mekdep/server/internal/utils"
)

const (
	sqlSettingsTable = `settings s`
	sqlSettingsOrder = `order by s.uid`
	sqlSettingsKeys  = `s.uid, s.key, s.value, s.updated_at`
)

func scanSettings(rows pgx.Row, model *models.Settings, addColumns ...interface{}) error {
	scanColumns := append([]interface{}{
		&model.ID, &model.Key, &model.Value, &model.UpdatedAt,
	}, addColumns...)
	return rows.Scan(scanColumns...)
}

func (d *PgxStore) SettingsFindById(ctx context.Context, id string) (model *models.Settings, err error) {
	model = &models.Settings{}
	err = d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		sql := fmt.Sprintf(`select %v from %v where uid=$1`, sqlSettingsKeys, sqlSettingsTable)
		err = scanSettings(tx.QueryRow(ctx, sql, id),
			model)
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}
	return model, err
}

func (d *PgxStore) SettingsUpsert(ctx context.Context, data *models.Settings) (model *models.Settings, err error) {
	err = d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		sql, sqlArgs := settingsInsertOrUpdate(data)
		err = tx.QueryRow(ctx, sql, sqlArgs...).Scan(&data.ID)
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}
	return model, nil
}

func settingsInsertOrUpdate(data *models.Settings) (sql string, sqlArgs []interface{}) {
	sqlArgs = []interface{}{}

	dataMap := map[string]interface{}{}
	dataStr, _ := json.Marshal(data)
	_ = json.Unmarshal(dataStr, &dataMap)

	sqlKeys := ""
	sqlValues := ""
	for key, value := range dataMap {
		if value == nil || slices.Contains([]string{"id"}, key) {
			continue
		}
		sqlArgs = append(sqlArgs, value)
		sqlValues += fmt.Sprintf(", $%v", len(sqlArgs))
		sqlKeys += fmt.Sprintf(", %v", key)
	}
	sqlKeys = strings.Trim(sqlKeys, ", ")
	sqlValues = strings.Trim(sqlValues, ", ")

	sqlTableParts := strings.Split(sqlSettingsTable, " ")
	sqlTable := sqlTableParts[0]
	sql = fmt.Sprintf(`INSERT INTO %v (%v) VALUES (%v)
	ON CONFLICT (key) DO UPDATE 
	SET value = excluded.value, updated_at = excluded.updated_at RETURNING uid;`, sqlTable, sqlKeys, sqlValues)
	return
}

func (d *PgxStore) SettingsFindBy(ctx context.Context, opts map[string]interface{}) (list []*models.Settings, err error) {
	err = d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		sql, sqlArgs := settingsFindBySql(opts)
		rows, err := tx.Query(ctx, sql, sqlArgs...)
		if err != nil {
			return err
		}
		for rows.Next() {
			model := &models.Settings{}
			err = scanSettings(rows, model)
			if err != nil {
				return err
			}
			list = append(list, model)
		}
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}
	return
}

func settingsFindBySql(opts map[string]interface{}) (sql string, sqlArgs []interface{}) {
	sqlArgs = []interface{}{}
	sqlWheres := "1=1"
	sql = fmt.Sprintf(`select %v from %v where %v`, sqlSettingsKeys, sqlSettingsTable, sqlWheres)
	return
}
