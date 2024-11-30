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

const sqlSchoolTransfersTable = `school_transfers st`
const sqlSchoolTransferFields = `st.uid, st.student_uid, st.target_school_uid, st.source_school_uid, st.target_classroom_uid, st.source_classroom_uid, st.sender_note, st.sender_files, st.receiver_note, st.sent_by, st.received_by, st.status, st.created_at, st.updated_at`

func scanSchoolTransfer(rows pgx.Row, model *models.SchoolTransfer, addColumns ...interface{}) error {
	scanColumns := append([]interface{}{
		&model.ID, &model.StudentId, &model.TargetSchoolId, &model.SourceSchoolId, &model.TargetClassroomId, &model.SourceClassroomId, &model.SenderNote,
		&model.SenderFiles, &model.ReceiverNote, &model.SentBy, &model.ReceivedBy, &model.Status, &model.CreatedAt, &model.UpdatedAt,
	}, addColumns...)
	return rows.Scan(scanColumns...)
}

func (d *PgxStore) SchoolTransfersFindById(ctx context.Context, id string) (model *models.SchoolTransfer, err error) {
	model = &models.SchoolTransfer{}
	err = d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		sql := fmt.Sprintf(`select %v from %v where uid=$1`, sqlSchoolTransferFields, sqlSchoolTransfersTable)
		err = scanSchoolTransfer(tx.QueryRow(ctx, sql, id),
			model)
		return
	})
	if err != nil {
		return nil, err
	}
	return model, err
}

func (d *PgxStore) SchoolTransfersFindBy(ctx context.Context, opts map[string]interface{}) (list *models.SchoolTransfers, err error) {
	list = &models.SchoolTransfers{
		SchoolTransfers: []*models.SchoolTransfer{},
	}

	err = d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		sql, sqlArgs := schoolTransfersFindBySql(opts)
		rows, err := tx.Query(ctx, sql, sqlArgs...)
		if err != nil {
			return err
		}
		for rows.Next() {
			model := &models.SchoolTransfer{}
			err = scanSchoolTransfer(rows, model, &list.Total)
			if err != nil {
				return err
			}
			list.SchoolTransfers = append(list.SchoolTransfers, model)
		}
		return
	})

	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}
	return list, nil
}

func schoolTransfersFindBySql(opts map[string]interface{}) (sql string, sqlArgs []interface{}) {
	sqlArgs = []interface{}{}
	sqlWheres := "1=1"
	sqlLimit := 12
	sqlOffset := 0

	if v, ok := opts["id"]; ok {
		sqlArgs = append(sqlArgs, v.(string))
		sqlWheres += fmt.Sprintf(` AND st.uid = $%v`, len(sqlArgs))
	}
	if v, ok := opts["ids"]; ok {
		sqlArgs = append(sqlArgs, v)
		sqlWheres += fmt.Sprintf(` AND te.uid = ANY($%v::uuid[])`, len(sqlArgs))
	}
	if v, ok := opts["student_id"]; ok {
		sqlArgs = append(sqlArgs, v.(string))
		sqlWheres += fmt.Sprintf(` AND st.student_uid = $%v`, len(sqlArgs))
	}
	if v, ok := opts["target_school_id"]; ok {
		sqlArgs = append(sqlArgs, v.(string))
		sqlWheres += fmt.Sprintf(` AND st.target_school_uid = $%v`, len(sqlArgs))
	}
	if v, ok := opts["source_school_id"]; ok {
		sqlArgs = append(sqlArgs, v.(string))
		sqlWheres += fmt.Sprintf(` AND st.source_school_uid = $%v`, len(sqlArgs))
	}
	if v, ok := opts["status"]; ok {
		sqlArgs = append(sqlArgs, v.(string))
		sqlWheres += fmt.Sprintf(` AND st.status = $%v`, len(sqlArgs))
	}
	sqlWheres += " GROUP BY st.uid"
	if v, ok := opts["limit"].(int); ok {
		sqlLimit = v
	}
	if v, ok := opts["offset"].(int); ok {
		sqlOffset = v
	}

	sqlAppend := fmt.Sprintf(` LIMIT %v OFFSET %v`, sqlLimit, sqlOffset)

	// final query
	sql = fmt.Sprintf(
		`SELECT %v, count(*) OVER() AS total_count 
		FROM %v
		WHERE %v %v`,
		sqlSchoolTransferFields,
		sqlSchoolTransfersTable,
		sqlWheres,
		sqlAppend,
	)

	return sql, sqlArgs
}

func (d *PgxStore) SchoolTransfersUpdate(ctx context.Context, data *models.SchoolTransfer) (model *models.SchoolTransfer, err error) {
	err = d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		sql, sqlArgs := schoolTransferUpdateSql(data)
		_, err = tx.Exec(ctx, sql, sqlArgs...)
		return
	})
	if err != nil {
		utils.LoggerDesc("Update Query error").Error(err)
		return nil, err
	}

	editModel, err := d.SchoolTransfersFindById(ctx, data.ID)
	if err != nil {
		return nil, err
	}
	return editModel, nil
}

func schoolTransferUpdateSql(data *models.SchoolTransfer) (sql string, sqlArgs []interface{}) {
	sqlArgs = []interface{}{data.ID}

	dataMap := map[string]interface{}{}
	dataStr, _ := json.Marshal(data)
	_ = json.Unmarshal(dataStr, &dataMap)

	sqlSets := ""
	for key, value := range dataMap {
		if value == nil || slices.Contains([]string{"id"}, key) {
			continue
		}
		if key == "target_classroom_id" {
			key = "target_classroom_uid"
		}
		sqlArgs = append(sqlArgs, value)
		sqlSets += fmt.Sprintf("%v=$%v, ", key, len(sqlArgs))
	}
	sqlSets = strings.TrimRight(sqlSets, ", ")

	sqlTableParts := strings.Split(sqlSchoolTransfersTable, " ")
	sqlTable := sqlTableParts[0]
	sql = fmt.Sprintf(`UPDATE %v SET %v WHERE uid=$1`, sqlTable, sqlSets)

	return sql, sqlArgs
}

func (d *PgxStore) SchoolTransfersInsert(ctx context.Context, data *models.SchoolTransfer) (model *models.SchoolTransfer, err error) {
	err = d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		sql, sqlArgs := schoolTransferInsertSql(data)
		err = tx.QueryRow(ctx, sql, sqlArgs...).Scan(&data.ID)
		return
	})
	if err != nil {
		utils.LoggerDesc("Insert Query error").Error(err)
		return nil, err
	}
	insertedSchoolTransfer, err := d.SchoolTransfersFindById(ctx, data.ID)
	if err != nil {
		return nil, err
	}

	return insertedSchoolTransfer, nil

}

func schoolTransferInsertSql(data *models.SchoolTransfer) (sql string, sqlArgs []interface{}) {
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
		if key == "student_id" {
			key = "student_uid"
		}
		if key == "target_school_id" {
			key = "target_school_uid"
		}
		if key == "source_school_id" {
			key = "source_school_uid"
		}
		if key == "source_classroom_id" {
			key = "source_classroom_uid"
		}
		sqlArgs = append(sqlArgs, value)
		sqlValues += fmt.Sprintf(", $%v", len(sqlArgs))
		sqlKeys += fmt.Sprintf(", %v", key)
	}

	sqlKeys = strings.Trim(sqlKeys, ", ")
	sqlValues = strings.Trim(sqlValues, ", ")

	sqlTableParts := strings.Split(sqlSchoolTransfersTable, " ")
	sqlTable := sqlTableParts[0]
	sql = fmt.Sprintf(`INSERT INTO %v (%v) VALUES (%v) RETURNING uid`, sqlTable, sqlKeys, sqlValues)

	return sql, sqlArgs
}

func (d *PgxStore) SchoolTransfersDelete(ctx context.Context, ids []string) (list *models.SchoolTransfers, err error) {
	list, err = d.SchoolTransfersFindBy(ctx, map[string]interface{}{
		"ids": ids,
	})

	if err != nil {
		return nil, err
	}
	if len(list.SchoolTransfers) < 1 {
		return nil, err
	}

	err = d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		sql := `DELETE FROM school_transfers WHERE uid = ANY($1)`
		_, err = tx.Exec(ctx, sql, ids)
		return
	})
	if err != nil {
		utils.LoggerDesc("Delete Query error").Error(err)
		return nil, err
	}

	return list, nil
}

func (d *PgxStore) SchoolTransfersLoadRelations(ctx context.Context, list *models.SchoolTransfers) error {
	err := d.SchoolTransfersLoadStudent(ctx, *list)
	if err != nil {
		return err
	}
	err = d.SchoolTransfersLoadTargetSchool(ctx, *list)
	if err != nil {
		return err
	}
	err = d.SchoolTransfersLoadSourceSchool(ctx, *list)
	if err != nil {
		return err
	}
	err = d.SchoolTransfersLoadTargetClassroom(ctx, *list)
	if err != nil {
		return err
	}
	err = d.SchoolTransfersLoadSourceClassroom(ctx, *list)
	if err != nil {
		return err
	}
	err = d.SchoolTransfersLoadSentByUser(ctx, *list)
	if err != nil {
		return err
	}
	err = d.SchoolTransfersLoadReceivedByUser(ctx, *list)
	if err != nil {
		return err
	}
	return nil
}

func (d *PgxStore) SchoolTransfersLoadStudent(ctx context.Context, l models.SchoolTransfers) error {
	ids := []string{}

	for _, m := range l.SchoolTransfers {
		ids = append(ids, m.ID)
	}
	if len(ids) < 1 {
		return nil
	}
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		query := `
				SELECT ` + sqlUserFields + `, st.uid
				FROM school_transfers st
				RIGHT JOIN users u ON (u.uid=st.student_uid)
				WHERE st.uid = ANY($1::uuid[])`

		rows, err := tx.Query(ctx, query, ids)
		if err != nil {
			return err
		}
		defer rows.Close()

		studentMap := make(map[string]*models.User)

		for rows.Next() {
			student := models.User{}
			var schoolTransferID string
			err = scanUser(rows, &student, &schoolTransferID)
			if err != nil {
				utils.LoggerDesc("Scan error").Error(err)
				return err
			}
			studentMap[schoolTransferID] = &student

		}
		// Assign the single school to each subject
		for _, m := range l.SchoolTransfers {
			if student, exists := studentMap[m.ID]; exists {
				m.Student = student
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

func (d *PgxStore) SchoolTransfersLoadTargetClassroom(ctx context.Context, l models.SchoolTransfers) error {
	ids := []string{}

	for _, m := range l.SchoolTransfers {
		ids = append(ids, m.ID)
	}
	if len(ids) < 1 {
		return nil
	}
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		query := `
				SELECT ` + sqlClassroomFields + `, st.uid
				FROM school_transfers st
				RIGHT JOIN classrooms c ON (c.uid=st.target_classroom_uid)
				WHERE st.uid = ANY($1::uuid[])`

		rows, err := tx.Query(ctx, query, ids)
		if err != nil {
			return err
		}
		defer rows.Close()
		classroomMap := make(map[string]*models.Classroom)

		for rows.Next() {
			classroom := models.Classroom{}
			var schoolTransferID string
			err = scanClassroom(rows, &classroom, &schoolTransferID)
			if err != nil {
				utils.LoggerDesc("Scan error").Error(err)
				return err
			}
			classroomMap[schoolTransferID] = &classroom
		}
		// Assign the single classroom to each subject
		for _, m := range l.SchoolTransfers {
			if classroom, exists := classroomMap[m.ID]; exists {
				m.TargetClassroom = classroom
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

func (d *PgxStore) SchoolTransfersLoadSourceClassroom(ctx context.Context, l models.SchoolTransfers) error {
	ids := []string{}

	for _, m := range l.SchoolTransfers {
		ids = append(ids, m.ID)
	}
	if len(ids) < 1 {
		return nil
	}
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		query := `
				SELECT ` + sqlClassroomFields + `, st.uid
				FROM school_transfers st
				RIGHT JOIN classrooms c ON (c.uid=st.source_classroom_uid)
				WHERE st.uid = ANY($1::uuid[])`

		rows, err := tx.Query(ctx, query, ids)
		if err != nil {
			return err
		}
		defer rows.Close()
		classroomMap := make(map[string]*models.Classroom)

		for rows.Next() {
			classroom := models.Classroom{}
			var schoolTransferID string
			err = scanClassroom(rows, &classroom, &schoolTransferID)
			if err != nil {
				utils.LoggerDesc("Scan error").Error(err)
				return err
			}
			classroomMap[schoolTransferID] = &classroom
		}
		// Assign the single classroom to each subject
		for _, m := range l.SchoolTransfers {
			if classroom, exists := classroomMap[m.ID]; exists {
				m.SourceClassroom = classroom
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

func (d *PgxStore) SchoolTransfersLoadTargetSchool(ctx context.Context, l models.SchoolTransfers) error {
	ids := []string{}

	for _, m := range l.SchoolTransfers {
		ids = append(ids, m.ID)
	}
	if len(ids) < 1 {
		return nil
	}
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		query := `
				SELECT ` + sqlSchoolFields + `, st.uid
				FROM school_transfers st
				RIGHT JOIN schools s ON (s.uid=st.target_school_uid)
				WHERE st.uid = ANY($1::uuid[])`

		rows, err := tx.Query(ctx, query, ids)
		if err != nil {
			return err
		}
		defer rows.Close()

		schoolMap := make(map[string]*models.School)

		for rows.Next() {
			school := models.School{}
			var schoolTransferID string
			err = scanSchool(rows, &school, &schoolTransferID)
			if err != nil {
				utils.LoggerDesc("Scan error").Error(err)
				return err
			}
			schoolMap[schoolTransferID] = &school

		}
		// Assign the single school to each subject
		for _, m := range l.SchoolTransfers {
			if school, exists := schoolMap[m.ID]; exists {
				m.TargetSchool = school
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

func (d *PgxStore) SchoolTransfersLoadSourceSchool(ctx context.Context, l models.SchoolTransfers) error {
	ids := []string{}

	for _, m := range l.SchoolTransfers {
		ids = append(ids, m.ID)
	}
	if len(ids) < 1 {
		return nil
	}
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		query := `
				SELECT ` + sqlSchoolFields + `, st.uid
				FROM school_transfers st
				RIGHT JOIN schools s ON (s.uid=st.source_school_uid)
				WHERE st.uid = ANY($1::uuid[])`

		rows, err := tx.Query(ctx, query, ids)
		if err != nil {
			return err
		}
		defer rows.Close()

		schoolMap := make(map[string]*models.School)

		for rows.Next() {
			school := models.School{}
			var schoolTransferID string
			err = scanSchool(rows, &school, &schoolTransferID)
			if err != nil {
				utils.LoggerDesc("Scan error").Error(err)
				return err
			}
			schoolMap[schoolTransferID] = &school

		}
		// Assign the single school to each subject
		for _, m := range l.SchoolTransfers {
			if school, exists := schoolMap[m.ID]; exists {
				m.SourceSchool = school
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

func (d *PgxStore) SchoolTransfersLoadSentByUser(ctx context.Context, l models.SchoolTransfers) error {
	ids := []string{}

	for _, m := range l.SchoolTransfers {
		ids = append(ids, m.ID)
	}
	if len(ids) < 1 {
		return nil
	}
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		query := `
				SELECT ` + sqlUserFields + `, st.uid
				FROM school_transfers st
				RIGHT JOIN users u ON (u.uid=st.sent_by)
				WHERE st.uid = ANY($1::uuid[])`

		rows, err := tx.Query(ctx, query, ids)
		if err != nil {
			return err
		}
		defer rows.Close()

		sentByUserMap := make(map[string]*models.User)

		for rows.Next() {
			sentByUser := models.User{}
			var schoolTransferID string
			err = scanUser(rows, &sentByUser, &schoolTransferID)
			if err != nil {
				utils.LoggerDesc("Scan error").Error(err)
				return err
			}
			sentByUserMap[schoolTransferID] = &sentByUser

		}
		// Assign the single school to each subject
		for _, m := range l.SchoolTransfers {
			if sentByUser, exists := sentByUserMap[m.ID]; exists {
				m.SentByUser = sentByUser
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

func (d *PgxStore) SchoolTransfersLoadReceivedByUser(ctx context.Context, l models.SchoolTransfers) error {
	ids := []string{}

	for _, m := range l.SchoolTransfers {
		ids = append(ids, m.ID)
	}
	if len(ids) < 1 {
		return nil
	}
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		query := `
				SELECT ` + sqlUserFields + `, st.uid
				FROM school_transfers st
				RIGHT JOIN users u ON (u.uid=st.received_by)
				WHERE st.uid = ANY($1::uuid[])`

		rows, err := tx.Query(ctx, query, ids)
		if err != nil {
			return err
		}
		defer rows.Close()

		receivedByUserMap := make(map[string]*models.User)

		for rows.Next() {
			receivedByUser := models.User{}
			var schoolTransferID string
			err = scanUser(rows, &receivedByUser, &schoolTransferID)
			if err != nil {
				utils.LoggerDesc("Scan error").Error(err)
				return err
			}
			receivedByUserMap[schoolTransferID] = &receivedByUser

		}
		// Assign the single school to each subject
		for _, m := range l.SchoolTransfers {
			if receivedByUser, exists := receivedByUserMap[m.ID]; exists {
				m.ReceivedByUser = receivedByUser
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
