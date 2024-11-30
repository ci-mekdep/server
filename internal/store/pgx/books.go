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

const sqlBookFields = `b.uid, b.title, b.categories, b.description, b.year, b.pages, b.authors, b.file, b.file_size, b.file_preview, b.is_downloadable, b.created_at, b.updated_at`

const sqlBookInsert = `insert into books`

const sqlBookUpdate = `update books set uid=uid`
const sqlBookDelete = `DELETE FROM books b WHERE uid = ANY($1::uuid[])`

const sqlBookSelect = `select ` + sqlBookFields + ` from books b where b.uid = ANY($1::uuid[])`

const sqlBookSelectMany = `select ` + sqlBookFields + `, count(*) over() as total from books b where b.uid=b.uid limit $1 offset $2 `

const sqlBookGetAuthors = `select distinct unnest(authors) from books`

func scanBook(rows pgx.Rows, m *models.Book, addColumns ...interface{}) (err error) {
	err = rows.Scan(parseColumnsForScan(m, addColumns...)...)
	return
}

func (d *PgxStore) BookFindBy(ctx context.Context, f models.BookFilterRequest) (books []*models.Book, total int, err error) {
	args := []interface{}{f.Limit, f.Offset}
	qs, args := BookListBuildQuery(f, args)
	err = d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		rows, err := tx.Query(ctx, qs, args...)
		for rows.Next() {
			t := models.Book{}
			err = scanBook(rows, &t, &total)
			if err != nil {
				utils.LoggerDesc("Scan error").Error(err)
				return err
			}
			books = append(books, &t)
		}
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, 0, err
	}

	return books, total, nil
}

func (d *PgxStore) BookFindById(ctx context.Context, ID string) (*models.Book, error) {
	row, err := d.BookFindByIds(ctx, []string{ID})
	if err != nil {
		return nil, err
	}
	if len(row) < 1 {
		return nil, errors.New("book not found by id: " + ID)
	}
	return row[0], nil
}

func (d *PgxStore) BookFindByIds(ctx context.Context, Ids []string) ([]*models.Book, error) {
	books := []*models.Book{}

	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		rows, err := tx.Query(ctx, sqlBookSelect, (Ids))
		for rows.Next() {
			t := models.Book{}
			err := scanBook(rows, &t)
			if err != nil {
				utils.LoggerDesc("Scan error").Error(err)
				return err
			}
			books = append(books, &t)
		}
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}

	return books, nil
}

func (d *PgxStore) BookGetAuthors(ctx context.Context) ([]string, error) {
	var authors []string
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		rows, err := tx.Query(ctx, sqlBookGetAuthors)
		for rows.Next() {
			var author string
			if err := rows.Scan(&author); err != nil {
				return err
			}
			authors = append(authors, author)
		}
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}
	return authors, nil
}

func (d *PgxStore) BookCreate(ctx context.Context, model *models.Book) (*models.Book, error) {
	qs, args := BookCreateQuery(model)
	qs += " RETURNING uid"
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		err = tx.QueryRow(ctx, qs, args...).Scan(&model.ID)
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}
	editModel, err := d.BookFindById(ctx, model.ID)
	if err != nil {
		return nil, err
	}
	return editModel, nil
}

func (d *PgxStore) BookUpdate(ctx context.Context, model *models.Book) (*models.Book, error) {
	qs, args := BookUpdateQuery(model)
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		_, err = tx.Query(ctx, qs, args...)
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}
	editModel, err := d.BookFindById(ctx, model.ID)
	if err != nil {
		return nil, err
	}
	return editModel, nil
}

func (d *PgxStore) BookDelete(ctx context.Context, items []*models.Book) ([]*models.Book, error) {
	ids := []string{}
	for _, i := range items {
		ids = append(ids, i.ID)
	}
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		_, err = tx.Query(ctx, sqlBookDelete, (ids))
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}
	return items, nil
}

func BookCreateQuery(m *models.Book) (string, []interface{}) {
	args := []interface{}{}
	cols := ""
	vals := ""
	q := BookAtomicQuery(m, true)
	for k, v := range q {
		args = append(args, v)
		cols += ", " + k
		vals += ", $" + strconv.Itoa(len(args))
	}
	qs := sqlBookInsert + " (" + strings.Trim(cols, ", ") + ") VALUES (" + strings.Trim(vals, ", ") + ")"
	return qs, args
}

func BookUpdateQuery(m *models.Book) (string, []interface{}) {
	args := []interface{}{}
	sets := ""
	q := BookAtomicQuery(m, false)
	for k, v := range q {
		args = append(args, v)
		sets += ", " + k + "=$" + strconv.Itoa(len(args))
	}
	args = append(args, m.ID)
	qs := strings.ReplaceAll(sqlBookUpdate, "set uid=uid", "set uid=uid"+sets+" ") + "where uid=$" + strconv.Itoa(len(args))
	return qs, args
}

func BookAtomicQuery(m *models.Book, isCreate bool) map[string]interface{} {
	q := map[string]interface{}{}
	if m.Title != nil {
		q["title"] = m.Title
	}
	if m.Categories != nil {
		q["categories"] = m.Categories
	}
	if m.Description != nil {
		q["Description"] = m.Description
	}
	if m.Year != nil {
		q["year"] = m.Year
	}
	if m.Pages != nil {
		q["pages"] = m.Pages
	}
	if m.Authors != nil {
		q["authors"] = *m.Authors
	}
	if m.File != nil {
		q["file"] = *m.File
	}
	if m.FileSize != nil {
		q["file_size"] = *m.FileSize
	}
	if m.FilePreview != nil {
		q["file_preview"] = *m.FilePreview
	}
	q["is_downloadable"] = m.IsDownloadable
	if isCreate {
		q["created_at"] = time.Now()
	}
	q["updated_at"] = time.Now()
	return q
}

func BookListBuildQuery(f models.BookFilterRequest, args []interface{}) (string, []interface{}) {
	var wheres string = ""

	if f.ID != "" {
		args = append(args, f.ID)
		wheres += " and b.uid=$" + strconv.Itoa(len(args))
	}
	if f.IDs != nil {
		args = append(args, *f.IDs)
		wheres += " and b.uid = ANY($" + strconv.Itoa(len(args)) + "::uuid[])"
	}
	if f.Categories != nil {
		categories := "{" + strings.Join(*f.Categories, ",") + "}"
		args = append(args, categories)
		wheres += " and b.categories::varchar[] @> $" + strconv.Itoa(len(args)) + "::varchar[]"
	}
	if f.Year != nil && *f.Year != 0 {
		args = append(args, *f.Year)
		wheres += " and b.year=$" + strconv.Itoa(len(args))
	}
	if f.Authors != nil {
		authors := "{" + strings.Join(*f.Authors, ",") + "}"
		args = append(args, authors)
		wheres += " and b.authors::varchar[] @> $" + strconv.Itoa(len(args)) + "::varchar[]"
	}
	if f.Search != nil && *f.Search != "" {
		*f.Search = strings.ToLower(*f.Search)
		args = append(args, *f.Search)
		wheres += " and (lower(b.title) like '%' || $" + strconv.Itoa(len(args)) + "|| '%')"
	}
	wheres += " group by b.uid "
	if f.Sort != nil && *f.Sort != "" {
		dir := "desc"
		if strings.HasSuffix(*f.Sort, "~") {
			dir = "asc"
		}
		*f.Sort = strings.ReplaceAll(*f.Sort, "~", "")
		wheres += " order by " + *f.Sort + " " + dir
	} else {
		wheres += " order by b.title desc"
	}
	qs := sqlBookSelectMany
	qs = strings.ReplaceAll(qs, "b.uid=b.uid", "b.uid=b.uid "+wheres+" ")
	return qs, args
}
