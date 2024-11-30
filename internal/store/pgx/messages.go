package pgx

import (
	"context"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/mekdep/server/internal/models"
	"github.com/mekdep/server/internal/utils"
)

func scanMessage(rows pgx.Row, m *models.Message, addColumns ...interface{}) (err error) {
	err = rows.Scan(parseColumnsForScan(m, addColumns...)...)
	return
}

func (store *PgxStore) GetMessagesQuery(
	ctx context.Context,
	dto models.GetMessagesQueryDto,
) ([]*models.Message, error) {

	dto.SetDefaults()
	stmt := `
		SELECT uid, user_uid, session_uid, group_uid, message, files, created_at
		FROM messages WHERE 1=1
	`
	args := []interface{}{}

	if dto.GroupId != nil {
		args = append(args, *dto.GroupId)
		stmt += " AND group_uid=$" + strconv.Itoa(len(args))
	}

	stmt += " ORDER BY created_at DESC"

	args = append(args, *dto.Limit)
	stmt += " LIMIT $" + strconv.Itoa(len(args))

	args = append(args, dto.Offset)
	stmt += " OFFSET $" + strconv.Itoa(len(args))

	messages := []*models.Message{}
	err := store.runQuery(ctx, func(conn *pgxpool.Conn) (err error) {
		rows, err := conn.Query(ctx, stmt, args...)
		for rows.Next() {
			message := models.Message{}
			err = rows.Scan(
				&message.ID,
				&message.UserId,
				&message.SessionId,
				&message.GroupId,
				&message.Message,
				&message.Files,
				&message.CreatedAt,
			)
			if err != nil {
				return err
			}
			messages = append(messages, &message)
		}
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}

	return messages, nil
}

func (store *PgxStore) GetMessageReadsQuery(
	ctx context.Context,
	dto models.GetMessageReadsQueryDto,
) (map[string]int, error) {

	wherePoint := "WHERE 1=1"
	wheres := ""
	stmt := `
		SELECT user_uid, count(*)
		FROM messages_reads ` + wherePoint + `
		GROUP BY user_uid`
	args := []interface{}{}

	if dto.UserIds != nil {
		args = append(args, *dto.UserIds)
		wheres += " AND user_uid=ANY($" + strconv.Itoa(len(args)) + "::uuid[])"
	}
	if dto.IsRead != nil {
		if *dto.IsRead {
			wheres += " AND read_at is not null"
		} else {
			wheres += " AND read_at is null"
		}
	}
	if dto.IsNotified != nil {
		if *dto.IsRead {
			wheres += " AND notified_at is not null"
		} else {
			wheres += " AND notified_at is null"
		}
	}
	if dto.NotifiedAtMax != nil {
		args = append(args, *dto.NotifiedAtMax)
		wheres += " AND (notified_at < $" + strconv.Itoa(len(args)) + " or notified_at is null)"
	}

	stmt = strings.ReplaceAll(stmt, wherePoint, wherePoint+wheres)

	userCounts := map[string]int{}
	err := store.runQuery(ctx, func(conn *pgxpool.Conn) (err error) {
		rows, err := conn.Query(ctx, stmt, args...)
		for rows.Next() {
			userId := ""
			count := int(0)
			err = rows.Scan(
				&userId,
				&count,
			)
			if err != nil {
				return err
			}
			userCounts[userId] = count
		}
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}

	if dto.SetNotified != nil && *dto.SetNotified {
		// set all notified now depending which count we get
		args = append(args, time.Now())
		updateStmt := `update messages_reads set notified_at=$` + strconv.Itoa(len(args)) + ` ` + wherePoint
		updateStmt = strings.ReplaceAll(updateStmt, wherePoint, wherePoint+wheres)
		log.Println("setted", time.Now(), updateStmt)
		err = store.runQuery(ctx, func(conn *pgxpool.Conn) (err error) {
			_, err = conn.Exec(ctx, updateStmt, args...)
			return
		})
		if err != nil {
			utils.LoggerDesc("Query error").Error(err)
			return nil, err
		}
	}

	return userCounts, nil
}

func (store *PgxStore) CreateMessageCommand(ctx context.Context, message models.Message) (models.Message, error) {
	stmt := `
		INSERT INTO messages 
		(user_uid, session_uid, group_uid, parent_uid, message, files, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW()) RETURNING uid
	`
	err := store.runQuery(ctx, func(conn *pgxpool.Conn) (err error) {
		row := conn.QueryRow(
			ctx,
			stmt,
			message.UserId,
			message.SessionId,
			message.GroupId,
			message.ParentId,
			message.Message,
			message.Files,
		)
		err = row.Scan(&message.ID)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return models.Message{}, err
	}

	return message, nil
}

func (d *PgxStore) CreateMessageReadsCommand(ctx context.Context, messageReads []models.MessageRead) error {
	if len(messageReads) <= 0 {
		return nil
	}

	err := d.runQuery(ctx, func(conn *pgxpool.Conn) (err error) {
		batch := pgx.Batch{}
		for _, messageRead := range messageReads {
			stmt := `
				INSERT INTO messages_reads 
				(user_uid, session_uid, message_uid, read_at) 
				VALUES ($1, $2, $3, $4)
				ON CONFLICT (user_uid, message_uid) DO NOTHING
			`
			batch.Queue(
				stmt,
				messageRead.UserId,
				messageRead.SessionId,
				messageRead.MessageId,
				messageRead.ReadAt,
			)
		}

		batchResults := conn.SendBatch(ctx, &batch)
		for range messageReads {
			_, err := batchResults.Exec()
			if err != nil {
				return err
			}
		}
		return err
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return err
	}
	return err
}

type MessageLoadParentMessageItem struct {
	Id       string
	Relation *models.Message
}

func (d *PgxStore) LoadMessagesWithParents(ctx context.Context, l *[]*models.Message) error {
	stmt := `
		SELECT p.uid, p.user_uid, p.session_uid, p.group_uid, p.message, p.files, p.created_at,
		m.uid FROM messages m
		right JOIN messages p ON (p.uid = m.parent_uid)
		WHERE m.uid = ANY($1::uuid[])
	`

	ids := []string{}
	for _, m := range *l {
		ids = append(ids, m.ID)
	}
	if len(ids) < 1 {
		return nil
	}

	results := []MessageLoadParentMessageItem{}
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		rows, err := tx.Query(ctx, stmt, (ids))
		for rows.Next() {
			parent := models.Message{}
			pid := ""
			err = scanMessage(rows, &parent, &pid)
			if err != nil {
				return err
			}
			results = append(results, MessageLoadParentMessageItem{Id: pid, Relation: &parent})
		}
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return err
	}
	for _, r := range results {
		for _, m := range *l {
			if r.Id == m.ID {
				m.Parent = r.Relation
			}
		}
	}
	return nil
}
