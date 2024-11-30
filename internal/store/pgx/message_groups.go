package pgx

import (
	"context"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/mekdep/server/internal/models"
	"github.com/mekdep/server/internal/utils"
)

// base
const sqlMessageGroupFields = `mg.uid, mg.admin_uid, mg.title, mg.school_uid, mg.type, mg.description, mg.classroom_uid`
const sqlMessageGroupSelect = `SELECT ` + sqlMessageGroupFields + ` FROM message_groups mg WHERE uid = ANY($1::uuid[])`

// const sqlMessageGroupSelectByClassroom = `SELECT ` + sqlMessageGroupFields + ` FROM message_groups mg WHERE classroom_uid = ANY($1::uuid[])`

const sqlSelectMessageGroupWithCountUnread = `
	SELECT ` + sqlMessageGroupFields + `, 
	COUNT(m.uid) FILTER(where mr.uid is null) AS unread_messages,
	COUNT(mg.uid) OVER() AS total
	FROM message_groups mg 
	LEFT JOIN messages m ON mg.uid = m.group_uid
	LEFT JOIN messages_reads mr ON m.uid = mr.message_uid AND mr.user_uid = $1
	WHERE mg.uid=mg.uid LIMIT $2 OFFSET $3`

const sqlMessageGroupInsert = `INSERT INTO message_groups`

func scanMessageGroup(rows pgx.Row, m *models.MessageGroup, addColumns ...interface{}) (err error) {
	err = rows.Scan(parseColumnsForScan(m, addColumns...)...)
	return
}

func scanStaticMessageGroup(rows pgx.Row, messageGroup *models.MessageGroup) (err error) {
	err = rows.Scan(
		&messageGroup.ID,
		&messageGroup.AdminId,
		&messageGroup.Title,
		&messageGroup.SchoolId,
		&messageGroup.Type,
		&messageGroup.Description,
		&messageGroup.ClassroomId,
	)
	return
}

func (d *PgxStore) MessageGroupsFindByIds(ctx context.Context, ids []string) ([]models.MessageGroup, error) {
	l := []models.MessageGroup{}

	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		rows, err := tx.Query(ctx, sqlMessageGroupSelect, (ids))
		for rows.Next() {
			m := models.MessageGroup{}
			err := scanStaticMessageGroup(rows, &m)
			if err != nil {
				utils.LoggerDesc("Scan error").Error(err)
				return err
			}
			l = append(l, m)
		}
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}

	return l, nil
}

func (d *PgxStore) MessageGroupsFindById(ctx context.Context, id string) (models.MessageGroup, error) {
	l, err := d.MessageGroupsFindByIds(ctx, []string{id})
	if err != nil {
		return models.MessageGroup{}, err
	}
	if len(l) < 1 {
		err = pgx.ErrNoRows
		utils.LoggerDesc("Scan error").Error(err)
		return models.MessageGroup{}, err
	}
	return l[0], nil
}

func (d *PgxStore) MessageGroupsFindBy(ctx context.Context, dto models.GetMessageGroupsRequest) ([]*models.MessageGroup, int, error) {
	if dto.Limit == nil {
		dto.Limit = new(int)
		*dto.Limit = 12
	}
	if dto.Offset == nil {
		dto.Offset = new(int)
		*dto.Offset = 0
	}
	args := []interface{}{dto.CurrentUserId, dto.Limit, dto.Offset}
	qs, args := MessageGroupListBuildQuery(dto, args)

	l := []*models.MessageGroup{}
	var total int
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		rows, err := tx.Query(ctx, qs, args...)
		for rows.Next() {
			sub := models.MessageGroup{}
			err = scanMessageGroup(rows, &sub, &total)
			if err != nil {
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

func (store *PgxStore) CreateMessageGroupCommand(ctx context.Context, messageGroup models.MessageGroup) (models.MessageGroup, error) {
	stmt := `
		INSERT INTO message_groups 
		(admin_uid, title, description, school_uid, classroom_uid, type)
		VALUES ($1, $2, $3, $4, $5, $6) RETURNING uid
	`

	err := store.runQuery(ctx, func(conn *pgxpool.Conn) (err error) {
		row := conn.QueryRow(
			ctx,
			stmt,
			messageGroup.AdminId,
			messageGroup.Title,
			messageGroup.Description,
			messageGroup.SchoolId,
			messageGroup.ClassroomId,
			messageGroup.Type,
		)
		err = row.Scan(&messageGroup.ID)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return models.MessageGroup{}, err
	}
	return messageGroup, nil
}

func MessageGroupListBuildQuery(dto models.GetMessageGroupsRequest, args []interface{}) (string, []interface{}) {
	var wheres string = ""

	if dto.ID != nil && *dto.ID != "" {
		args = append(args, *dto.ID)
		wheres += " AND mg.uid=$" + strconv.Itoa(len(args))
	}
	if dto.SchoolId != nil {
		args = append(args, dto.SchoolId)
		wheres += " AND mg.school_uid=$" + strconv.Itoa(len(args))
	}
	if dto.SchoolIds != nil {
		args = append(args, *dto.SchoolIds)
		wheres += " AND mg.school_uid = ANY($" + strconv.Itoa(len(args)) + "::uuid[])"
	}
	if dto.AdminId != nil {
		args = append(args, dto.AdminId)
		wheres += " AND mg.admin_uid=$" + strconv.Itoa(len(args))
	}
	if dto.AdminIds != nil {
		args = append(args, *dto.AdminIds)
		wheres += " AND mg.admin_uid = ANY($" + strconv.Itoa(len(args)) + "::uuid[])"
	}
	if dto.ClassroomId != nil {
		args = append(args, dto.ClassroomId)
		wheres += " AND mg.classroom_uid=$" + strconv.Itoa(len(args))
	}
	if dto.ClassroomIds != nil {
		args = append(args, *dto.ClassroomIds)
		wheres += " AND mg.classroom_uid = ANY($" + strconv.Itoa(len(args)) + "::uuid[])"
	}
	if dto.Types != nil {
		args = append(args, *dto.Types)
		wheres += " AND mg.type = ANY($" + strconv.Itoa(len(args)) + "::string[])"
	} else if dto.Type != nil {
		args = append(args, *dto.Type)
		wheres += " AND mg.type=$" + strconv.Itoa(len(args))
	}

	if dto.Search != nil {
		*dto.Search = strings.ToLower(*dto.Search)
		args = append(args, dto.Search)
		wheres += " AND (lower(title) LIKE '%' || $" + strconv.Itoa(len(args)) +
			" || '%' OR lower(description) LIKE '%' || $" + strconv.Itoa(len(args)) + " || '%')"
	}
	if dto.Ids != nil {
		args = append(args, dto.Ids)
		wheres += " AND mg.uid = ANY($" + strconv.Itoa(len(args)) + "::uuid[])"
	}
	wheres += " GROUP BY mg.uid"
	if dto.Sort != nil {
		dir := "DESC"
		if strings.HasSuffix(*dto.Sort, "~") {
			dir = "ASC"
		}
		*dto.Sort = strings.ReplaceAll(*dto.Sort, "~", "")
		args = append(args, dto.Sort)
		wheres += " ORDER BY $" + strconv.Itoa(len(args)) + " " + dir
	} else {
		wheres += " ORDER BY mg.school_uid ASC, mg.title ASC, mg.classroom_uid DESC"
	}
	qs := sqlSelectMessageGroupWithCountUnread
	qs = strings.ReplaceAll(qs, "mg.uid=mg.uid", "mg.uid=mg.uid "+wheres+" ")

	return qs, args
}
