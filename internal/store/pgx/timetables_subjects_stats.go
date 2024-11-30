package pgx

import (
	"context"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/mekdep/server/internal/models"
	"github.com/mekdep/server/internal/utils"
)

const sqlSubjectLessonsCount = `select 
sb.uid, 
sb.classroom_uid,
sb."name", 
max(c."name"), 
l.date as lesson_date, 
max(l.title) as lesson_title, 
max(l.assignment_title) as lesson_assignment, 
count(distinct uc.user_uid) as students_count, 
(select count(*) from grades g where  g.lesson_uid = ANY(array_agg(l.uid))) as grades_count,
(select count(*) from absents a where  a.lesson_uid = ANY(array_agg(l.uid))) as absents_count
from subjects sb 
	RIGHT JOIN lessons l ON (sb.uid=l.subject_uid)
	LEFT JOIN classrooms c ON (c.uid=sb.classroom_uid)
	LEFT JOIN user_classrooms uc ON (uc.classroom_uid = sb.classroom_uid
			and((COALESCE(uc.type_key, 0) = COALESCE(sb.classroom_type_key, 0))))
	where sb.uid = ANY($1::uuid[]) and l.date>=$2 and l.date<=$3 and l.is_teacher_excused = false
	group by sb.uid, l.date`

const sqlSubjectLessonsCountBySchool = `select 
ll.school_uid,
count(ll.uid) as lessons_count,
sum(ll.topics_count) as topics_count,
(select count(uc.user_uid) from classrooms c left join user_classrooms uc on (uc.classroom_uid=c.uid and uc."type" is null and uc.type_key is null) where c.school_uid=ll.school_uid) as students_count,
sum(ll.grades_count) as grades_count,
sum(ll.absents_count) as absents_count,
COALESCE(avg(ll.grade_full_percent),0) as grade_full_percent
from 
(select 
s.uid as school_uid,
l.uid as uid,
count(distinct lt.uid) as topics_count, 
(select count(*) from grades g where  g.lesson_uid = ANY(array_agg(l.uid))) as grades_count,
(select count(*) from absents a where  a.lesson_uid = ANY(array_agg(l.uid))) as absents_count,
((select count(*) from grades g where  g.lesson_uid = ANY(array_agg(l.uid))) * 100 / nullif(count(distinct uc.user_uid),0)) as grade_full_percent
from schools s 
	left join subjects sb on (sb.school_uid =s.uid)
	left join lessons l on (sb.uid=l.subject_uid and l.date>=$2 and l.date<=$3)
	left join lessons lt on (lt.uid=l.uid and lt.title != '')
	LEFT JOIN user_classrooms uc ON (uc.classroom_uid = sb.classroom_uid
			and((COALESCE(uc.type_key, 0) = COALESCE(sb.classroom_type_key, 0))))
	where s.uid=ANY($1::uuid[])
	group by s.uid,l.uid) ll 
group by ll.school_uid`

const sqlSubjectPercentByStudent = `select s.uid as subject_uid, s.name as subject, count(distinct l.uid) as lessons_count, count(distinct g.uid) as grades_count, 
array_agg(g.value::int) FILTER (WHERE g.value IS NOT null)::integer[] as values1, 
array_remove(array_agg((g.values[1]+g.values[2])/2) FILTER (WHERE g.values IS NOT NULL), null) as values2
from subjects s 
left join lessons l on (l.subject_uid =s.uid and l."date" >= $3 and l."date" <= $4)
left join grades g on (l.uid=g.lesson_uid and g.student_uid=$1)
where s.classroom_uid=$2
group by s.uid`

const sqlSubjectRatingByStudent = `select uc.user_uid, s.uid as subject_uid, s.name as subject_name, count(distinct l.uid) as lessons_count, count(distinct g.uid) as grades_count, 
array_agg(g.value::int) FILTER (WHERE g.value IS NOT null) as values1, 
array_remove(array_agg((g.values[1]+g.values[2])/2) FILTER (WHERE g.values IS NOT NULL), null) as values2
from subjects s 
left join lessons l on (l.subject_uid =s.uid and l."date" >= $2 and l."date" <= $3)
left join user_classrooms uc on (uc.classroom_uid=s.classroom_uid)
left join grades g on (l.uid=g.lesson_uid and g.student_uid=uc.user_uid)
where s.classroom_uid=$1
group by s.uid, uc.user_uid`

const sqlSubjectGradeStreak = `select l.date, count(distinct l.uid),  
array_agg(g.value::int) FILTER (WHERE g.value IS NOT null) as values1, 
array_remove(array_agg((g.values[1]+g.values[2])/2) FILTER (WHERE g.values IS NOT NULL), null) as values2
from lessons l 
left join grades g on (g.lesson_uid = l.uid and g.student_uid = $2)
right join subjects s on (s.uid=l.subject_uid and s.classroom_uid = $1)
where l.date <= $3
group by l.date
order by l.date desc`

const sqlSubjectGrades = `select max(s.uid::text), max(s.name), (l.date),
CASE WHEN g.value is not null THEN g.value::integer ELSE (coalesce(g.values[1],0)+coalesce(g.values[2],g.values[1],0))/2 END::text
from grades g
right join lessons l on l.uid=g.lesson_uid 
left join subjects s on s.uid=l.subject_uid
where g.student_uid=$1 and l.date >= $2 and l.date <= $3
group by g.uid, l.uid
order by l.date`

const sqlStudentRating = `select ` + sqlUserFields + `, count(distinct g.uid)
from users u
left join grades g on g.student_uid = u.uid 
right join lessons l on l.uid = g.lesson_uid 
left join user_schools us on us.user_uid = u.uid 
where g.value::integer=5
and us.school_uid=$1 and l.date >= $2 and l.date <= $3 
group by u.uid
having count(g.uid) > 0
order by count(g.uid) desc, max(g.created_at) asc
limit 50`

const sqlSubjectPeriodGradeFinished = `select ` + sqlSubjectFields + ` , max(c.	name) , count(uc) as students_count, count(pg) as finished_count 
from subjects sb
left join classrooms c on c.uid = sb.classroom_uid 
left join user_classrooms uc on (uc.classroom_uid = sb.classroom_uid and uc."type" is null)
left join period_grades pg 
on (pg.subject_uid = sb.uid and pg.period_key = $2 and (pg.grade_count >= 3 or pg.lesson_count -3 <= pg.absent_count))
where sb.school_uid=ANY($1::uuid[])
group by sb.uid`

func (d *PgxStore) SubjectsRatingByStudentWithPrev(ctx context.Context, classroomId string, startDate time.Time, endDate time.Time) ([]models.SubjectRating, error) {
	sp, err := d.SubjectsRatingByStudent(ctx, classroomId, startDate, endDate)
	if err != nil {
		return nil, err
	}
	startDate = startDate.AddDate(0, 0, -7)
	endDate = endDate.AddDate(0, 0, -7)
	spPrev, err := d.SubjectsRatingByStudent(ctx, classroomId, startDate, endDate)
	if err != nil {
		return nil, err
	}
	for _, v := range spPrev {
		for kk, vv := range sp {
			if v.SubjectName == vv.SubjectName && v.StudentId == vv.StudentId {
				sp[kk].PointPrev = v.Point
				sp[kk].RatingPrev = v.Rating
			}
		}
	}
	return sp, nil
}
func (d *PgxStore) SubjectsRatingByStudent(ctx context.Context, classroomId string, startDate time.Time, endDate time.Time) ([]models.SubjectRating, error) {
	res := []models.SubjectRating{}
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		rows, err := tx.Query(ctx, sqlSubjectRatingByStudent, classroomId, startDate, endDate)
		for rows.Next() {
			item := models.SubjectRating{}
			err = rows.Scan(&item.StudentId, &item.SubjectId, &item.SubjectName, &item.LessonsCount, &item.GradesCount, &item.GradesValues1, &item.GradesValues2)
			if err != nil {
				return err
			}
			item.CalcPoint()
			res = append(res, item)
		}
		return
	})
	// calc rating
	models.CalcRatingStudents(&res)

	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}
	return res, nil
}

func (d *PgxStore) SubjectsPercentByStudentWithPrev(ctx context.Context, studentId string, classroomId string, startDate time.Time, endDate time.Time) ([]models.SubjectPercent, error) {
	sp, err := d.SubjectsPercentByStudent(ctx, studentId, classroomId, startDate, endDate)
	if err != nil {
		return nil, err
	}
	startDate = startDate.AddDate(0, 0, -7)
	endDate = endDate.AddDate(0, 0, -7)
	spPrev, err := d.SubjectsPercentByStudent(ctx, studentId, classroomId, startDate, endDate)
	if err != nil {
		return nil, err
	}
	for _, v := range spPrev {
		for kk, vv := range sp {
			if v.SubjectId == vv.SubjectId {
				sp[kk].PointPrev = v.Point
				sp[kk].PercentPrev = v.Percent
			}
		}
	}
	return sp, nil
}

func (d *PgxStore) SubjectsPercentByStudent(ctx context.Context, studentId string, classroomId string, startDate time.Time, endDate time.Time) ([]models.SubjectPercent, error) {
	res := []models.SubjectPercent{}
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		rows, err := tx.Query(ctx, sqlSubjectPercentByStudent, studentId, classroomId, startDate, endDate)
		for rows.Next() {
			item := models.SubjectPercent{}
			err = rows.Scan(&item.SubjectId, &item.SubjectFullName, &item.LessonsCount, &item.GradesCount, &item.GradesValues1, &item.GradesValues2)
			if err != nil {
				return err
			}
			item.CalcPoint()
			res = append(res, item)
		}
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}
	return res, nil
}

func (d *PgxStore) SubjectsPercents(ctx context.Context, subjectIds []string, startDate time.Time, endDate time.Time) ([]models.DashboardSubjectsPercent, error) {
	res := []models.DashboardSubjectsPercent{}
	if len(subjectIds) < 1 {
		return res, nil
	}

	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		rows, err := tx.Query(ctx, sqlSubjectLessonsCount, subjectIds, startDate, endDate)
		for rows.Next() {
			item := models.DashboardSubjectsPercent{}
			err = rows.Scan(&item.SubjectId, &item.ClassroomId, &item.SubjectName, &item.ClassroomName, &item.LessonDate, &item.LessonTitle, &item.AssignmentTitle, &item.StudentsCount, &item.GradesCount, &item.AbsentsCount)
			if err != nil {
				return err
			}
			item.SetOtherKeys()
			res = append(res, item)
		}
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}
	// set item
	return res, nil
}

func (d *PgxStore) SubjectsPercentsBySchool(ctx context.Context, schoolIds []string, startDate time.Time, endDate time.Time) ([]models.DashboardSubjectsPercentBySchool, error) {
	res := []models.DashboardSubjectsPercentBySchool{}
	if len(schoolIds) < 1 {
		return res, nil
	}

	diffDays := uint(endDate.Unix()/60/60/24 - startDate.Unix()/60/60/24)
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		rows, err := tx.Query(ctx, sqlSubjectLessonsCountBySchool, schoolIds, startDate, endDate)
		if err != nil {
			return err
		}
		for rows.Next() {
			item := models.DashboardSubjectsPercentBySchool{}
			err = rows.Scan(&item.SchoolID, &item.LessonsCount, &item.TopicsCount, &item.StudentsCount, &item.GradesCount, &item.AbsentsCount, &item.GradeFullPercent)
			if err != nil {
				return err
			}
			item.DaysCount = int(diffDays) + 1
			item.SetOtherKeys()
			res = append(res, item)
		}
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}
	// set item
	return res, nil
}
func (d *PgxStore) SubjectsPeriodGradeFinished(ctx context.Context, schoolIds []string, periodNumber int) ([]models.SubjectPeriodGradeFinished, error) {
	res := []models.SubjectPeriodGradeFinished{}
	if len(schoolIds) < 1 {
		return res, nil
	}

	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		rows, err := tx.Query(ctx, sqlSubjectPeriodGradeFinished, schoolIds, periodNumber)
		for rows.Next() {
			item := models.SubjectPeriodGradeFinished{
				Subject: models.Subject{
					Classroom: &models.Classroom{},
				},
			}
			err = scanSubject(rows, &item.Subject, &item.ClassroomName, &item.StudentsCount, &item.FinishedCount)
			item.Subject.Classroom.Name = item.ClassroomName

			if err != nil {
				return err
			}
			if models.SubjectIsNoGrade(item.Subject) {
				item.FinishedCount = item.StudentsCount
			}
			res = append(res, item)
		}
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}
	// set item
	return res, nil
}

func (d *PgxStore) SubjectsGradeStrike(ctx context.Context, classroomId, studentId string) ([]models.SubjectLessonGrades, error) {
	res := []models.SubjectLessonGrades{}

	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		rows, err := tx.Query(ctx, sqlSubjectGradeStreak, classroomId, studentId, time.Now().AddDate(0, 0, -1))
		for rows.Next() {
			item := models.SubjectLessonGrades{}
			err = rows.Scan(&item.LessonDate, &item.LessonHours, &item.GradesValues1, &item.GradesValues2)
			if err != nil {
				return err
			}
			item.SetOtherKeys()
			res = append(res, item)
		}
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}
	// set item
	return res, nil
}

func (d *PgxStore) SubjectGrades(ctx context.Context, studentId string, startDate time.Time, endDate time.Time) ([]models.SubjectGrade, error) {
	res := []models.SubjectGrade{}
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		rows, err := tx.Query(ctx, sqlSubjectGrades, studentId, startDate, endDate)
		for rows.Next() {
			item := models.SubjectGrade{}
			err = rows.Scan(&item.SubjectId, &item.SubjectName, &item.Date, &item.GradeValue)
			if err != nil {
				return err
			}
			res = append(res, item)
		}
		return
	})

	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}
	return res, nil
}

func (d *PgxStore) StudentRatingBySchool(ctx context.Context, schoolId string, startDate time.Time, endDate time.Time) ([]*models.User, []int, error) {
	res := []*models.User{}
	vals := []int{}
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		rows, err := tx.Query(ctx, sqlStudentRating, schoolId, startDate, endDate)
		for rows.Next() {
			sub := models.User{}
			count := 0
			err = scanUser(rows, &sub, &count)
			if err != nil {
				return err
			}
			res = append(res, &sub)
			vals = append(vals, count)
		}
		return
	})

	if err != nil && err != pgx.ErrNoRows {
		utils.LoggerDesc("Query error").Error(err)
		return nil, nil, err
	}
	return res, vals, nil
}
