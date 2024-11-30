package pgx

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/mekdep/server/internal/models"
	"github.com/mekdep/server/internal/utils"
)

const sqlTopicFields = `t.uid, t.book_uid, t.book_page, t.subject_name, t.classyear, t.period, t.level, t.language, t.tags, t.title, t.content, t.files`

const sqlTopicsInsert = `insert into topics`

const sqlTopicUpdate = `update topics set uid=uid`
const sqlTopicsDelete = `DELETE FROM topics t WHERE uid = ANY($1::uuid[])`

const sqlTopicSelect = `select ` + sqlTopicFields + ` from topics t where t.uid = ANY($1::uuid[])`

const sqlTopicSelectMany = `select ` + sqlTopicFields + `, count(*) over() as total from topics t where t.uid=t.uid limit $1 offset $2 `

// relations
const sqlTopicBook = `select ` + sqlBookFields + `, t.uid from topics t
	right join books b on (b.uid=t.book_uid) where t.uid=ANY($1::uuid[])`

func scanTopics(rows pgx.Rows, m *models.Topics, addColumns ...interface{}) (err error) {
	err = rows.Scan(parseColumnsForScan(m, addColumns...)...)
	return
}

func (d *PgxStore) TopicsFindBy(ctx context.Context, f models.TopicsFilterRequest) (topics []*models.Topics, total int, err error) {
	args := []interface{}{f.Limit, f.Offset}
	qs, args := TopicsListBuildQuery(f, args)
	err = d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		rows, err := tx.Query(ctx, qs, args...)
		for rows.Next() {
			t := models.Topics{}
			err = scanTopics(rows, &t, &total)
			if err != nil {
				utils.LoggerDesc("Scan error").Error(err)
				return err
			}
			topics = append(topics, &t)
		}
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, 0, err
	}

	return topics, total, nil
}

func (d *PgxStore) TopicsFindById(ctx context.Context, ID string) (*models.Topics, error) {
	row, err := d.TopicsFindByIds(ctx, []string{ID})
	if err != nil {
		return nil, err
	}
	if len(row) < 1 {
		return nil, errors.New("topic not found by uid: " + ID)
	}
	return row[0], nil
}

func (d *PgxStore) TopicsFindByIds(ctx context.Context, Ids []string) ([]*models.Topics, error) {
	topics := []*models.Topics{}

	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		rows, err := tx.Query(ctx, sqlTopicSelect, (Ids))
		for rows.Next() {
			t := models.Topics{}
			err := scanTopics(rows, &t)
			if err != nil {
				utils.LoggerDesc("Scan error").Error(err)
				return err
			}
			topics = append(topics, &t)
		}
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}

	return topics, nil
}

func (d *PgxStore) TopicsCreate(ctx context.Context, model *models.Topics) (*models.Topics, error) {
	qs, args := TopicsCreateQuery(model)
	qs += " RETURNING uid"
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		err = tx.QueryRow(ctx, qs, args...).Scan(&model.ID)
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}
	editModel, err := d.TopicsFindById(ctx, model.ID)
	if err != nil {
		return nil, err
	}
	return editModel, nil
}

func (d *PgxStore) TopicsUpdate(ctx context.Context, model *models.Topics) (*models.Topics, error) {
	qs, args := TopicsUpdateQuery(model)
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		_, err = tx.Query(ctx, qs, args...)
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}
	editModel, err := d.TopicsFindById(ctx, model.ID)
	if err != nil {
		return nil, err
	}
	return editModel, nil
}

func (d *PgxStore) TopicsDelete(ctx context.Context, items []*models.Topics) ([]*models.Topics, error) {
	ids := []string{}
	for _, i := range items {
		ids = append(ids, i.ID)
	}
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		_, err = tx.Query(ctx, sqlTopicsDelete, (ids))
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}
	return items, nil
}

func TopicsCreateQuery(m *models.Topics) (string, []interface{}) {
	args := []interface{}{}
	cols := ""
	vals := ""
	q := TopicsAtomicQuery(m, true)
	for k, v := range q {
		args = append(args, v)
		cols += ", " + k
		vals += ", $" + strconv.Itoa(len(args))
	}
	qs := sqlTopicsInsert + " (" + strings.Trim(cols, ", ") + ") VALUES (" + strings.Trim(vals, ", ") + ")"
	return qs, args
}

func TopicsUpdateQuery(m *models.Topics) (string, []interface{}) {
	args := []interface{}{}
	sets := ""
	q := TopicsAtomicQuery(m, false)
	for k, v := range q {
		args = append(args, v)
		sets += ", " + k + "=$" + strconv.Itoa(len(args))
	}
	args = append(args, m.ID)
	qs := strings.ReplaceAll(sqlTopicUpdate, "set uid=uid", "set uid=uid"+sets+" ") + "where uid=$" + strconv.Itoa(len(args))
	return qs, args
}

func TopicsAtomicQuery(m *models.Topics, isCreate bool) map[string]interface{} {
	q := map[string]interface{}{}
	if m.SubjectName != nil {
		q["subject_name"] = m.SubjectName
	}
	if m.Classyear != nil {
		q["classyear"] = m.Classyear
	}
	if m.Period != nil {
		q["period"] = m.Period
	}
	if m.Level != nil {
		q["level"] = m.Level
	}
	if m.Language != nil {
		q["language"] = m.Language
	}
	if m.Tags != nil {
		q["tags"] = *m.Tags
	}
	if m.Title != nil {
		q["title"] = *m.Title
	}
	if m.Content != nil {
		q["content"] = *m.Content
	}
	if m.BookPage != nil {
		q["book_page"] = m.BookPage
	}
	if m.BookId != nil {
		q["book_uid"] = m.BookId
	}
	if m.Files != nil {
		q["files"] = *m.Files
	}
	return q
}

func TopicsListBuildQuery(f models.TopicsFilterRequest, args []interface{}) (string, []interface{}) {
	var wheres string = ""

	if f.ID != nil && *f.ID != "" {
		args = append(args, *f.ID)
		wheres += " and t.uid=$" + strconv.Itoa(len(args)) + "::uuid"
	}
	if f.IDs != nil && len(*f.IDs) > 0 {
		args = append(args, *f.IDs)
		wheres += " and t.uid = ANY($" + strconv.Itoa(len(args)) + "::uuid[])"
	}
	if f.SubjectName != nil {
		args = append(args, *f.SubjectName)
		wheres += " and t.subject_name=$" + strconv.Itoa(len(args))
	}
	if f.Classyear != nil && *f.Classyear != "" {
		args = append(args, *f.Classyear)
		wheres += " and t.classyear=$" + strconv.Itoa(len(args))
	}
	if f.PeriodNumber != nil && *f.PeriodNumber != "" {
		args = append(args, *f.PeriodNumber)
		wheres += " and t.period=$" + strconv.Itoa(len(args))
	}
	if f.Level != nil && *f.Level != "" {
		args = append(args, *f.Level)
		wheres += " and t.level=$" + strconv.Itoa(len(args))
	}
	if f.Language != nil && *f.Language != "" {
		args = append(args, *f.Language)
		wheres += " and t.language=$" + strconv.Itoa(len(args))
	}
	if f.Search != nil && *f.Search != "" {
		*f.Search = strings.ToLower(*f.Search)
		args = append(args, *f.Search)
		wheres += " and (lower(t.title) like '%' || $" + strconv.Itoa(len(args)) + "|| '%')"
	}
	wheres += " group by t.uid "
	if f.Sort != nil && *f.Sort != "" {
		dir := "desc"
		if strings.HasSuffix(*f.Sort, "~") {
			dir = "asc"
		}
		*f.Sort = strings.ReplaceAll(*f.Sort, "~", "")
		wheres += " order by " + *f.Sort + " " + dir
	} else {
		wheres += " order by t.subject_name, t.title desc"
	}
	qs := sqlTopicSelectMany
	qs = strings.ReplaceAll(qs, "t.uid=t.uid", "t.uid=t.uid "+wheres+" ")
	return qs, args
}

func (d *PgxStore) TopicsLoadRelations(ctx context.Context, l *[]*models.Topics) error {
	ids := []string{}
	for _, m := range *l {
		ids = append(ids, m.ID)
	}
	if len(ids) < 1 {
		return nil
	}

	if rs, err := d.TopicsLoadBook(ctx, ids); err != nil {
		return err
	} else {

		for _, r := range rs {
			for _, m := range *l {
				if r.Id == m.ID {
					m.Book = r.Relation
				}
			}
		}
	}
	return nil
}

type TopicsLoadBookItem struct {
	Id       string
	Relation *models.Book
}

func (d *PgxStore) TopicsLoadBook(ctx context.Context, ids []string) ([]TopicsLoadBookItem, error) {
	res := []TopicsLoadBookItem{}
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		rows, err := tx.Query(ctx, sqlTopicBook, (ids))
		for rows.Next() {
			sub := models.Book{}
			pid := ""
			err = scanBook(rows, &sub, &pid)
			if err != nil {
				return err
			}
			res = append(res, TopicsLoadBookItem{Id: pid, Relation: &sub})
		}
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}

	return res, nil
}
