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
	sqlTeacherExcusesTable = `teacher_excuses te`
	sqlTeacherExcusesOrder = `order by te.uid`
	sqlTeacherExcusesKeys  = `te.uid, te.teacher_uid, te.school_uid, te.start_date, te.end_date, te.reason, te.note, te.document_files, te.created_at, te.updated_at`
)

func scanTeacherExcuses(rows pgx.Row, model *models.TeacherExcuse, addColumns ...interface{}) error {
	scanColumns := append([]interface{}{
		&model.ID, &model.TeacherId, &model.SchoolId, &model.StartDate, &model.EndDate, &model.Reason, &model.Note, &model.DocumentFiles, &model.CreatedAt, &model.UpdatedAt,
	}, addColumns...)
	return rows.Scan(scanColumns...)
}

func (d *PgxStore) TeacherExcusesFindById(ctx context.Context, id string, loadRelations bool) (model *models.TeacherExcuse, err error) {
	model = &models.TeacherExcuse{}
	err = d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		sql := fmt.Sprintf(`select %v from %v where uid=$1`, sqlTeacherExcusesKeys, sqlTeacherExcusesTable)
		err = scanTeacherExcuses(tx.QueryRow(ctx, sql, id),
			model)
		return
	})
	if loadRelations {
		teacherExcuseList := []*models.TeacherExcuse{model}
		err = d.teacherExcusesLoadRelations(ctx, &teacherExcuseList)
		if err != nil {
			utils.LoggerDesc("Load relations error").Error(err)
			return
		}
	}
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}
	return model, err
}

func (d *PgxStore) TeacherExcuseInsert(ctx context.Context, data *models.TeacherExcuse) (model *models.TeacherExcuse, err error) {
	err = d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		sql, sqlArgs := teacherExcuseInsertSql(data)
		err = tx.QueryRow(ctx, sql, sqlArgs...).Scan(&data.ID)
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return
	}

	editTeacherExcuse, err := d.TeacherExcusesFindById(ctx, data.ID, false)
	if err != nil {
		return nil, err
	}
	return editTeacherExcuse, nil
}

func teacherExcuseInsertSql(data *models.TeacherExcuse) (sql string, sqlArgs []interface{}) {
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
		if key == "teacher_id" {
			key = "teacher_uid"
		}
		if key == "school_id" {
			key = "school_uid"
		}
		sqlArgs = append(sqlArgs, value)
		sqlValues += fmt.Sprintf(", $%v", len(sqlArgs))
		sqlKeys += fmt.Sprintf(", %v", key)
	}
	sqlKeys = strings.Trim(sqlKeys, ", ")
	sqlValues = strings.Trim(sqlValues, ", ")

	sqlTableParts := strings.Split(sqlTeacherExcusesTable, " ")
	sqlTable := sqlTableParts[0]
	sql = fmt.Sprintf(`insert into %v (%v) values (%v) RETURNING uid`, sqlTable, sqlKeys, sqlValues)
	return
}

func (d *PgxStore) TeacherExcuseUpdate(ctx context.Context, data *models.TeacherExcuse) (model *models.TeacherExcuse, err error) {
	err = d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		sql, sqlArgs := teacherExcuseUpdateSql(data)
		_, err = tx.Exec(ctx, sql, sqlArgs...)
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return
	}
	editTeacherExcuse, err := d.TeacherExcusesFindById(ctx, data.ID, false)
	if err != nil {
		return nil, err
	}
	return editTeacherExcuse, nil
}

func teacherExcuseUpdateSql(data *models.TeacherExcuse) (sql string, sqlArgs []interface{}) {
	sqlArgs = []interface{}{data.ID}

	dataMap := map[string]interface{}{}
	dataStr, _ := json.Marshal(data)
	_ = json.Unmarshal(dataStr, &dataMap)

	sqlSets := ""
	for key, value := range dataMap {
		if value == nil || slices.Contains([]string{"id", "start_date", "end_date", "reason"}, key) {
			continue
		}
		if key == "teacher_id" {
			key = "teacher_uid"
		}
		sqlArgs = append(sqlArgs, value)
		sqlSets += fmt.Sprintf("%v=$%v, ", key, len(sqlArgs))
	}

	sqlTableParts := strings.Split(sqlTeacherExcusesTable, " ")
	sqlTable := sqlTableParts[0]
	sql = fmt.Sprintf(`update %v set %v where uid=$1`, sqlTable, strings.Trim(sqlSets, ", "))
	return
}

func (d *PgxStore) TeacherExcusesFindBy(ctx context.Context, opts map[string]interface{}) (list *models.TeacherExcuses, err error) {
	list = &models.TeacherExcuses{
		TeacherExcuses: []*models.TeacherExcuse{},
	}
	err = d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		sql, sqlArgs := teacherExcusesFindBySql(opts)
		rows, err := tx.Query(ctx, sql, sqlArgs...)
		if err != nil {
			return err
		}
		for rows.Next() {
			model := &models.TeacherExcuse{}
			err = scanTeacherExcuses(rows, model, &list.Total)
			if err != nil {
				return err
			}
			list.TeacherExcuses = append(list.TeacherExcuses, model)
		}
		return
	})
	if loadRelations, ok := opts["load_relations"].(bool); ok && loadRelations {
		err = d.teacherExcusesLoadRelations(ctx, &list.TeacherExcuses)
		if err != nil {
			utils.LoggerDesc("Load relations error").Error(err)
			return nil, err
		}
	}
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}
	return
}

func teacherExcusesFindBySql(opts map[string]interface{}) (sql string, sqlArgs []interface{}) {
	sqlArgs = []interface{}{}
	sqlWheres := "1=1"
	sqlLimit := 12
	sqlOffset := 0
	if v, ok := opts["id"]; ok {
		sqlArgs = append(sqlArgs, v)
		sqlWheres += fmt.Sprintf(` AND te.uid = $%v`, len(sqlArgs))
	}
	if v, ok := opts["ids"]; ok {
		sqlArgs = append(sqlArgs, v)
		sqlWheres += fmt.Sprintf(` AND te.uid = ANY($%v::uuid[])`, len(sqlArgs))
	}
	if v, ok := opts["teacher_id"]; ok {
		sqlArgs = append(sqlArgs, v)
		sqlWheres += fmt.Sprintf(` AND te.teacher_uid = $%v`, len(sqlArgs))
	}
	if v, ok := opts["school_id"]; ok {
		sqlArgs = append(sqlArgs, v)
		sqlWheres += fmt.Sprintf(` AND te.school_uid = $%v`, len(sqlArgs))
	}
	if v, ok := opts["reason"]; ok {
		sqlArgs = append(sqlArgs, v)
		sqlWheres += fmt.Sprintf(` AND te.reason = $%v`, len(sqlArgs))
	}
	if v, ok := opts["date"]; ok {
		sqlArgs = append(sqlArgs, v)
		sqlWheres += fmt.Sprintf(` AND $%v BETWEEN te.start_date and te.end_date`, len(sqlArgs))
	}
	if v, ok := opts["limit"].(int); ok {
		sqlLimit = v
	}
	if v, ok := opts["offset"].(int); ok {
		sqlOffset = v
	}
	sqlAppend := fmt.Sprintf(`%v LIMIT %v OFFSET %v;`, sqlTeacherExcusesOrder, sqlLimit, sqlOffset)
	sql = fmt.Sprintf(`select %v, count(te.uid) over() from %v where %v %v`, sqlTeacherExcusesKeys, sqlTeacherExcusesTable, sqlWheres, sqlAppend)

	return
}

func (d *PgxStore) teacherExcusesLoadRelations(ctx context.Context, list *[]*models.TeacherExcuse) error {
	ids := []string{}
	for _, item := range *list {
		ids = append(ids, item.ID)
	}
	if len(ids) < 1 {
		return nil
	}
	if err := d.loadTeacher(ctx, ids, list); err != nil {
		return err
	}
	if err := d.loadSchool(ctx, ids, list); err != nil {
		return err
	}
	return nil
}

func (d *PgxStore) loadTeacher(ctx context.Context, ids []string, list *[]*models.TeacherExcuse) error {
	sql := `SELECT ` + sqlUserFields + `, te.uid from teacher_excuses te right join users u on (u.uid=te.teacher_uid) where te.uid = ANY($1::uuid[])`
	return d.loadRelation(ctx, sql, ids, func(rows pgx.Rows) error {
		var subModel models.User
		var id string
		if err := scanUser(rows, &subModel, &id); err != nil {
			return err
		}
		for _, item := range *list {
			if item.ID == id {
				item.Teacher = &subModel
				break
			}
		}
		return nil
	})
}

func (d *PgxStore) loadSchool(ctx context.Context, ids []string, list *[]*models.TeacherExcuse) error {
	sql := `SELECT ` + sqlSchoolFields + `, te.uid from teacher_excuses te right join schools s on (s.uid=te.school_uid) where te.uid = ANY($1::uuid[])`
	return d.loadRelation(ctx, sql, ids, func(rows pgx.Rows) (err error) {
		var subModel models.School
		var id string
		if err = scanSchool(rows, &subModel, &id); err != nil {
			return err
		}
		var schoolParents []*models.School
		for _, item := range *list {
			if item.ID == id {
				schoolParents = append(schoolParents, &subModel)
				item.School = &subModel
				break
			}
		}
		err = d.SchoolsLoadParents(ctx, &schoolParents)
		if err != nil {
			return err
		}
		return nil
	})
}

func (d *PgxStore) loadRelation(ctx context.Context, sql string, ids []string, scanFn func(rows pgx.Rows) error) error {
	return d.runQuery(ctx, func(tx *pgxpool.Conn) error {
		rows, err := tx.Query(ctx, sql, ids)
		if err != nil {
			return err
		}
		for rows.Next() {
			if err := scanFn(rows); err != nil {
				return err
			}
		}
		return nil
	})
}

func (d *PgxStore) TeacherExcusesDelete(ctx context.Context, ids []string) (list *models.TeacherExcuses, err error) {
	list, err = d.TeacherExcusesFindBy(ctx, map[string]interface{}{
		"ids": ids,
	})
	if err != nil {
		return nil, err
	}
	if len(list.TeacherExcuses) < 1 {
		return nil, err
	}

	err = d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		sqlTableParts := strings.Split(sqlTeacherExcusesTable, " ")
		sqlTable := sqlTableParts[0]
		sql := fmt.Sprintf(`delete from %v where uid=ANY($1)`, sqlTable)
		_, err = tx.Exec(ctx, sql, ids)
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return
	}
	return
}
