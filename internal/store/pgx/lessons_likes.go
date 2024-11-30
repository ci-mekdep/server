package pgx

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/mekdep/server/internal/models"
	"github.com/mekdep/server/internal/utils"
)

const sqlLessonLikesFields = `ll.uid, ll.lesson_uid, ll.user_uid`
const sqlLessonLikesInsert = `insert into lesson_likes (lesson_uid, user_uid) VALUES ($1, $2)`

const sqlLessonLikesByUser = `select exists (select 1 from lesson_likes ll where ll.lesson_uid = $1 and ll.user_uid = $2)`
const sqlLessonLikesThenUnlike = `DELETE FROM lesson_likes ll WHERE ll.lesson_uid = $1 AND ll.user_uid = $2`

// relations
const sqlLessonLikesLesson = `select ` + sqlLessonFields + `, ll.uid from lesson_likes ll
	right join lessons l on (l.uid=ll.lesson_uid) where ll.uid = ANY($1::uuid[])`
const sqlLessonLikesUser = `select ` + sqlUserFields + `, ll.uid from lesson_likes ll
	right join users u on (u.uid=ll.user_uid) where ll.uid = ANY($1::uuid[])`

const sqlLessonLikesCount = `select count(*) from lesson_likes ll
	inner join lessons l on (l.uid = ll.lesson_uid)
	inner join subjects s on (s.uid = l.subject_uid)
	where s.teacher_uid = $1
	and ll.created_at >= DATE_TRUNC('month', $2::DATE)
	and ll.created_at < DATE_TRUNC('month', $2::DATE) + INTERVAL '1 month'`

func (d *PgxStore) LessonsLikes(ctx context.Context, lessonID string, userID string) error {
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		_, err = tx.Exec(ctx, sqlLessonLikesInsert, lessonID, userID)
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return err
	}
	return nil
}

func (d *PgxStore) LessonsLikesByUser(ctx context.Context, lessonID string, userID string) (bool, error) {
	var liked bool
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) error {
		err := tx.QueryRow(ctx, sqlLessonLikesByUser, lessonID, userID).Scan(&liked)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				liked = false
				return nil
			}
			return err
		}
		return nil
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return false, err
	}
	return liked, nil
}

func (d *PgxStore) LessonsLikesThenUnlike(ctx context.Context, lessonID string, userID string) error {
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) error {
		_, err := tx.Exec(ctx, sqlLessonLikesThenUnlike, lessonID, userID)
		return err
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return err
	}
	return nil
}

func (d *PgxStore) LessonLikesCount(ctx context.Context, teacherId string, month time.Time) (int, error) {
	var count int
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) error {
		err := tx.QueryRow(ctx, sqlLessonLikesCount, teacherId, month).Scan(&count)
		return err
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return 0, err
	}
	count *= models.TeacherPointWeight
	return count, nil
}

func (d *PgxStore) LessonLikesLoadRelations(ctx context.Context, l *[]*models.LessonLikes) error {
	ids := []string{}
	for _, m := range *l {
		ids = append(ids, m.ID)
	}
	if len(ids) < 1 {
		return nil
	}

	if rs, err := d.LessonLikesLoadLesson(ctx, ids); err != nil {
		return err
	} else {
		for _, r := range rs {
			for _, m := range *l {
				if r.ID == m.ID {
					m.Lesson = r.Relation
				}
			}
		}
	}
	//load user
	if rs, err := d.LessonLikesLoadUser(ctx, ids); err != nil {
		return err
	} else {
		for _, r := range rs {
			for _, m := range *l {
				if r.ID == m.ID {
					m.User = r.Relation
				}
			}
		}
	}
	return nil
}

type LessonLikeLoadLessonItem struct {
	ID       string
	Relation *models.Lesson
}

func (d *PgxStore) LessonLikesLoadLesson(ctx context.Context, ids []string) ([]LessonLikeLoadLessonItem, error) {
	res := []LessonLikeLoadLessonItem{}
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		rows, err := tx.Query(ctx, sqlLessonLikesLesson, (ids))
		for rows.Next() {
			sub := models.Lesson{}
			pid := ""
			err = scanLesson(rows, &sub, &pid)
			if err != nil {
				return err
			}
			res = append(res, LessonLikeLoadLessonItem{ID: pid, Relation: &sub})
		}
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}

	return res, nil
}

type LessonLikeLoadUserItem struct {
	ID       string
	Relation *models.User
}

func (d *PgxStore) LessonLikesLoadUser(ctx context.Context, ids []string) ([]LessonLikeLoadUserItem, error) {
	res := []LessonLikeLoadUserItem{}
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		rows, err := tx.Query(ctx, sqlLessonLikesUser, (ids))
		for rows.Next() {
			sub := models.User{}
			pid := ""
			err = scanUser(rows, &sub, &pid)
			if err != nil {
				return err
			}
			res = append(res, LessonLikeLoadUserItem{ID: pid, Relation: &sub})
		}
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}

	return res, nil
}
