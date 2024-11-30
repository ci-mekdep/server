package pgx

import (
	"context"
	"errors"
	"log"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"

	"github.com/mekdep/server/internal/models"
	"github.com/mekdep/server/internal/utils"
)

// base
const sqlUserLat = `(select lat from sessions s where s.user_uid=u.uid order by lat desc limit 1)`
const sqlUserSecondarySchool = `u.passport_number, u.birth_cert_number, u.apply_number, u.work_title, u.work_place, u.district, u.reference, u.nickname, u.education_title, u.education_place, u.education_group`
const sqlUserFieldsNoRelation = `u.uid, u.first_name, u.last_name, u.middle_name, u.username, u.password, u.status, u.phone, u.phone_verified_at, u.email, u.email_verified_at, 
u.birthday, u.gender, u.address, u.avatar, ` + sqlUserLat + `,` + sqlUserSecondarySchool + `, u.documents, u.document_files, u.updated_at, u.created_at, u.archived_at`
const sqlUserFieldsSchoolForParent = sqlUserFieldsNoRelation + `, max('parent'), max(s.name), max(s.uid::text), null, null, null, null`
const sqlUserFieldsClassroom = sqlUserFieldsNoRelation + `, max(us.role_code), max(s.name), max(s.uid::text), max(sp.name), max(cc.uid::text), max(cc.name), null`
const sqlUserFieldsParent = sqlUserFieldsNoRelation + `, null, null, null, null, null, null, 
array_remove(array_agg(up.school_uid), NULL)  as school_uids`
const sqlUserFields = sqlUserFieldsNoRelation + `, null, null, null, null, null, null, null`
const sqlUserSelect = `select ` + sqlUserFieldsParent + ` from users u 
left join user_parents up on up.parent_uid=u.uid
where u.uid = ANY($1::uuid[])
group by u.uid`

const sqlUserSelectMany = `select ` + sqlUserFieldsClassroom + `, count(u.uid) over() from users u
	left join user_schools us on us.user_uid=u.uid 
	left join schools s on us.school_uid=s.uid 
	left join schools sp on s.uid=sp.parent_uid
	left join user_classrooms uc on uc.user_uid=u.uid
	left join classrooms cc on cc.uid=uc.classroom_uid
	left join classrooms c on c.teacher_uid=u.uid
	left join user_payments p on p.user_uid=u.uid and p.classroom_uid=uc.classroom_uid
	where u.uid=u.uid limit $1 offset $2`
const sqlUserSelectManyParent = `select ` + sqlUserFieldsParent + `, count(u.uid) over() from users u
	left join user_parents up on up.parent_uid=u.uid
	where u.uid=u.uid limit $1 offset $2`

const sqlUserSelectCount = `select 
null,
(select count(uid) from classrooms c where c.school_uid=ANY($1::uuid[]) and c.parent_uid is null) as classrooms_count,
(select count(*) from user_schools us where us.school_uid=ANY($1::uuid[]) and us.role_code ='student') as students_count, 
(select count(distinct us.parent_uid) from user_parents us where us.school_uid=ANY($1::uuid[])) as parents_count,
(select count(*) from user_schools us where us.school_uid=ANY($1::uuid[]) and us.role_code ='teacher') as teachers_count, 
(select count(*) from user_schools us where us.school_uid=ANY($1::uuid[]) and us.role_code ='principal') as principals_count, 
(select count(*) from user_schools us where us.school_uid=ANY($1::uuid[]) and us.role_code ='organization') as organizations_count, 
count(*) as schools_count,
(select count(t.uid) from classrooms c right join timetables t on t.classroom_uid =c.uid where c.school_uid=ANY($1::uuid[])) as timetables_count,
0,
0
from schools s where s.uid=ANY($1::uuid[]) and s.parent_uid is not null`

// (select count(distinct  ses.user_uid) from sessions ses right join user_schools us on us.user_uid=ses.user_uid  where us.school_uid=ANY($1::uuid[])) as users_online_count,
const sqlUserSelectCountBySchool = `select 
s.uid,
(select count(uid) from classrooms c where c.school_uid=s.uid and c.parent_uid is null) as classrooms_count,
(select count(*) from user_schools us where us.school_uid=s.uid and us.role_code ='student') as students_count, 
(select count(distinct us.parent_uid) from user_parents us where us.school_uid=s.uid) as parents_count,
(select count(*) from user_schools us where us.school_uid=s.uid and us.role_code ='teacher') as teachers_count, 
(select count(*) from user_schools us where us.school_uid=s.uid and us.role_code ='principal') as principals_count, 
(select count(*) from user_schools us where us.school_uid=s.uid and us.role_code ='organization') as organizations_count, 
count(uid) as schools_count,
(select count(t.uid) from classrooms c right join timetables t on t.classroom_uid =c.uid where c.school_uid=s.uid) as timetables_count,
0,
coalesce((select count(distinct us.*) from user_schools us right join user_parents up on up.child_uid =us.user_uid  where us.school_uid=s.uid and us.role_code ='student'  having count(up.*) > 1),0 ) as students_count
from schools s
where s.uid=ANY($1::uuid[])
group by s.uid`

const sqlUserSelectCountByClassroom = `select ` + sqlClassroomFields + `, c.school_uid, count(distinct uc.*), count(distinct up.*), count(distinct ss.user_uid), count(distinct ucc.uid)
from classrooms c
left join user_classrooms uc on uc.classroom_uid=c.uid and uc.type_key is null
left join user_parents up on up.child_uid=uc.user_uid
left join sessions ss on ss.user_uid=up.parent_uid and ss.iat >= now()-interval'14 days'
left join users ucc on ucc.uid=up.child_uid and uc.tariff_end_at >= now()
where c.school_uid=ANY($1::uuid[])
group by c.uid`

const sqlUserInsert = `insert into users`
const sqlUserUpdate = `update users set uid=uid`
const sqlUserDelete = `delete from users where uid = ANY($1::uuid[])`

// relations
// TODO: remove school_uid from children
const sqlUserChildren = `select ` + sqlUserFieldsClassroom + `, array_agg(distinct up.parent_uid) from user_parents up
	left join user_schools us on (us.user_uid=up.child_uid)
	left join schools s on us.school_uid=s.uid 
	left join schools sp on s.uid=sp.parent_uid
	right join users u on up.child_uid=u.uid 
	left join user_classrooms uc on uc.user_uid=u.uid
	left join classrooms cc on cc.uid=uc.classroom_uid
	where up.parent_uid = ANY($1::uuid[]) 
	group by u.uid`
const sqlUserParentsDelete = `delete from user_parents up using user_schools us where up.child_uid=us.user_uid and us.school_uid = $2 and up.parent_uid=$1`
const sqlUserParentsInsert = `insert into user_parents (parent_uid, child_uid, school_uid) values ($1, $2, $3)`
const sqlUserParents = `select ` + sqlUserFieldsSchoolForParent + `, array_agg(distinct up.child_uid) from user_parents up
	left join schools s on up.school_uid=s.uid 
	right join users u on up.parent_uid=u.uid  where up.child_uid = ANY($1::uuid[]) and u.uid=u.uid group by u.uid`

const sqlUserTeacherClassroom = `select ` + sqlClassroomFields + `, u.uid from users u
       right join classrooms c on c.teacher_uid=u.uid where u.uid = ANY($1::uuid[])`

const sqlUserChildrenDelete = `delete from user_parents where child_uid=$1`
const sqlUserChildrenInsert = `insert into user_parents (child_uid, parent_uid, school_uid) values ($1, $2, $3)`
const sqlUserSchools = `select us.role_code, us.user_uid, ` + sqlSchoolFields + `   from user_schools us 
	left join schools s on us.school_uid=s.uid  where us.user_uid = ANY($1::uuid[])`
const sqlUserSchoolsDelete = `delete from user_schools where user_uid=$1`
const sqlUserSchoolsDeleteBySchool = `delete from user_schools where user_uid=ANY($1::uuid[]) and COALESCE(school_uid,'00000000-0000-0000-0000-000000000000'::uuid)=ANY($2::uuid[])`
const sqlUserSchoolsDeleteByRole = `delete from user_schools where user_uid=ANY($1::uuid[]) and school_uid=ANY($2::uuid[]) and role_code=ANY($3::text[])`
const sqlUserParentsDeleteByRole = `delete from user_parents where parent_uid=ANY($1::uuid[]) and school_uid=ANY($2::uuid[])`
const sqlUserSchoolsInsert = `insert into user_schools (user_uid, school_uid, role_code) values ($1, $2, $3)`
const sqlUserClassrooms = `select ` + sqlClassroomFields + `, uc.user_uid, uc.type, p.expires_at, 'plus' from user_classrooms uc 
	left join user_payments p on p.user_uid=uc.user_uid and p.classroom_uid=uc.classroom_uid
	right join classrooms c on uc.classroom_uid=c.uid  
	where uc.user_uid = ANY($1::uuid[]) and uc.type is null and uc.type_key is null`
const sqlUserClassroomsAll = `select ` + sqlClassroomFields + `, uc.user_uid, uc.type, p.expires_at, 'plus' from user_classrooms uc 
	left join user_payments p on p.user_uid=uc.user_uid and p.classroom_uid=uc.classroom_uid
	right join classrooms c on uc.classroom_uid=c.uid  
	where uc.user_uid = ANY($1::uuid[])`

// const sqlUserClassroomsDeleteExcept = `delete from user_classrooms where user_uid=$1 and classroom_uid<>ANY($2::uuid[])`
const sqlUserClassroomsDelete = `delete from user_classrooms where user_uid=$1 and classroom_uid = ANY($2::uuid[])`
const sqlUserClassroomsInsert = `INSERT INTO user_classrooms (user_uid, classroom_uid, type)
SELECT $1, $2, null
WHERE NOT EXISTS (SELECT 1 FROM user_classrooms WHERE user_uid=$1 and classroom_uid=$2 and type is null)`

const sqlUserPaymentUpsert = `
    INSERT INTO user_payments (user_uid, expires_at, classroom_uid) 
    VALUES ($1, $2, $3)
    ON CONFLICT (user_uid, classroom_uid) 
    DO UPDATE SET expires_at = EXCLUDED.expires_at, classroom_uid = EXCLUDED.classroom_uid`

const sqlUpdateUserPaymentClassroom = `UPDATE user_payments SET classroom_uid = $2 WHERE user_uid = $1;`
const sqlUserPaymentGetDate = `select expires_at from user_payments where user_uid=$1 and classroom_uid=$2`

const sqlUserClassroomSelect = `SELECT uc.user_uid, uc.classroom_uid, up.expires_at AS tariff_end_at, 'plus'
								FROM user_classrooms uc 
								LEFT JOIN user_payments up ON uc.user_uid = up.user_uid AND uc.classroom_uid = up.classroom_uid 
								WHERE uc.user_uid = $1 AND uc.classroom_uid = $2`

const sqlUsersOnlineCount = `select count(*) from sessions where lat >= $1 and lat <= $2`

func scanUser(rows pgx.Row, m *models.User, addColumns ...interface{}) (err error) {
	err = rows.Scan(parseColumnsForScan(m, addColumns...)...)
	return
}

func (d *PgxStore) UserClassroomGet(ctx context.Context, userId string, classroomId string) (*models.UserClassroom, error) {
	var userClassroom models.UserClassroom
	// Execute the query
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) error {
		return tx.QueryRow(ctx, sqlUserClassroomSelect, userId, classroomId).Scan(
			&userClassroom.UserId,
			&userClassroom.ClassroomId,
			&userClassroom.TariffEndAt,
			&userClassroom.TariffType,
		)
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}

	return &userClassroom, nil
}

func (d *PgxStore) UpdateUserPayment(ctx context.Context, userUid string, expireAt time.Time, classroomId string) (*models.User, error) {
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		_, err = tx.Exec(ctx, sqlUserPaymentUpsert, userUid, expireAt, classroomId)
		return err
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}
	return nil, nil
}

func (d *PgxStore) UpdateUserPaymentClassroom(ctx context.Context, userUid string, classroomId string) (*models.User, error) {
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		_, err = tx.Exec(ctx, sqlUpdateUserPaymentClassroom, userUid, classroomId)
		return err
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}
	return nil, nil
}

func (d *PgxStore) GetDateUserPayment(ctx context.Context, userUid string, classroomId string) (time.Time, error) {
	var date time.Time
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		row := tx.QueryRow(ctx, sqlUserPaymentGetDate, userUid, classroomId)
		err = row.Scan(&date)
		if err == pgx.ErrNoRows {
			return nil
		}
		return err
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return time.Time{}, err
	}
	return date, nil
}

func (d *PgxStore) UsersFindByUsername(ctx context.Context, username string, schoolId *string, onlyAdmin bool) (models.User, error) {
	var user models.User
	var role string
	// first try to get default search by username
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		wheres := ""
		args := []interface{}{1, 0, username}
		if schoolId != nil {
			if onlyAdmin {
				wheres += " and (us.role_code='" + string(models.RoleAdmin) + "' or us.role_code='" + string(models.RoleOrganization) + "' or us.role_code='" + string(models.RoleOperator) + "' )"
			} else {
				args = append(args, schoolId)
				wheres += " and us.school_uid=$" + strconv.Itoa(len(args))
			}
		}
		qs := strings.ReplaceAll(`select `+sqlUserFields+`, max(us.role_code) from users u 
			RIGHT JOIN user_schools us on us.user_uid=u.uid 
			WHERE (username=REPLACE(REPLACE(REPLACE(REPLACE(REPLACE(REPLACE(LOWER($3), 'ý', 'y'), 'ö', 'o'), 'ä', 'a'), 'ň', 'n'), 'ü', 'u'), 'ç', 'c') or phone=$3) 
				AND 1=1 
			GROUP BY u.uid	
			LIMIT $1 OFFSET $2`, "1=1", "1=1"+wheres)
		err = scanUser(
			tx.QueryRow(ctx, qs, args...),
			&user, &role)
		if err != nil {
			return err
		}
		user.Role = &role
		if !slices.Contains(models.DefaultRoles, models.Role(*user.Role)) {
			user.Role = nil
		}
		return err
	})
	if err != nil && err != pgx.ErrNoRows {
		utils.LoggerDesc("Query error").Error(err)
		return models.User{}, err
	}
	var parent models.User
	// second, if role is student, then try parent search by username
	// if not student return
	if user.Role != nil && *user.Role != string(models.RoleStudent) {
		return user, nil
	}
	err = d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		qs := `
			SELECT ` + sqlUserFieldsParent + `
			FROM users u 
			RIGHT JOIN user_parents up ON up.parent_uid = u.uid
			WHERE phone=$1 and up.school_uid=$2
			GROUP BY u.uid`
		args := []interface{}{username, schoolId}

		err = scanUser(tx.QueryRow(ctx, qs, args...), &parent)
		if err != nil {
			return err
		}
		return err
	})
	if err != nil && err != pgx.ErrNoRows {
		utils.LoggerDesc("Query error").Error(err)
		return models.User{}, err
	}

	// return result
	if parent.ID != "" {
		user = parent
	}
	return user, nil
}

func (d *PgxStore) UsersFindByIds(ctx context.Context, ids []string) ([]*models.User, error) {
	l := []*models.User{}

	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		rows, err := tx.Query(ctx, sqlUserSelect, (ids))
		for rows.Next() {
			u := models.User{}
			err := scanUser(rows, &u)
			if err != nil {
				return err
			}
			l = append(l, &u)
		}
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}

	return l, nil
}

func (d *PgxStore) UsersFindById(ctx context.Context, id string) (*models.User, error) {
	l, err := d.UsersFindByIds(ctx, []string{id})
	if err != nil {
		return nil, err
	}
	if len(l) < 1 {
		return nil, pgx.ErrNoRows
	}
	return l[0], err
}

func (d *PgxStore) UsersFindBy(ctx context.Context, f models.UserFilterRequest) ([]*models.User, int, error) {
	if f.Limit == nil {
		f.Limit = new(int)
		*f.Limit = 12
	}
	if f.Offset == nil {
		f.Offset = new(int)
		*f.Offset = 0
	}
	args := []interface{}{f.Limit, f.Offset}
	var qs string
	qs, args = UsersListBuildQuery(f, args)
	l := []*models.User{}
	var total int
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		rows, err := tx.Query(ctx, qs, args...)
		for rows.Next() {
			u := models.User{}
			err = scanUser(rows, &u, &total)
			if err != nil {
				return err
			}
			l = append(l, &u)
		}
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, 0, err
	}

	return l, total, nil
}

// TODO: refactor below "mess"
// TODO: split into sub "UserUpdateRelations"
// example: UserUpdateRelationsChildren
// example: UserUpdateRelationsSchools
// example: UserUpdateRelationsParents
func (d *PgxStore) UserUpdateRelations(ctx context.Context, model *models.User) (*models.User, error) {
	var err error
	if len(model.Children) > 0 {
		schoolId := model.SchoolId
		// delete relations
		err = d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
			_, err = tx.Query(ctx, sqlUserParentsDelete, model.ID, schoolId)
			return
		})
		if err != nil {
			utils.LoggerDesc("Query error").Error(err)
			return nil, err
		}
		// create relations
		for _, child := range model.Children {
			err = d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
				_, err = tx.Query(ctx, sqlUserParentsInsert, model.ID, child.ID, schoolId)
				return
			})
			if err != nil {
				utils.LoggerDesc("Query error").Error(err)
				return nil, err
			}
		}
	}
	if model.Schools != nil {
		// delete relations
		err = d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
			_, err = tx.Query(ctx, sqlUserSchoolsDelete, model.ID)
			return
		})
		if err != nil {
			utils.LoggerDesc("Query error").Error(err)
			return nil, err
		}
		// create relations
		for _, c := range model.Schools {
			if c.RoleCode == models.RoleParent {
				continue
			}
			err = d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
				_, err = tx.Query(ctx, sqlUserSchoolsInsert, model.ID, c.SchoolUid, c.RoleCode)
				return
			})
			if err != nil {
				utils.LoggerDesc("Query error").Error(err)
				return nil, err
			}
		}
	}
	if len(model.Parents) > 0 {
		// delete relations
		err = d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
			_, err = tx.Query(ctx, sqlUserChildrenDelete, model.ID)
			return
		})
		if err != nil {
			utils.LoggerDesc("Query error").Error(err)
			return nil, err
		}
		// create relations
		for _, parent := range model.Parents {
			var schoolId string
			for _, v := range parent.Schools {
				schoolId = *v.SchoolUid
			}
			err = d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
				_, err = tx.Query(ctx, sqlUserChildrenInsert, model.ID, parent.ID, schoolId)
				return
			})
			if err != nil {
				utils.LoggerDesc("Query error").Error(err)
				return nil, err
			}
		}
	}

	// TODO: method UserUpdateRelationsChangeClassroom
	if model.Classrooms != nil {
		// Map of school_id -> classroom_id
		oldClassroomBySchool := make(map[string]string)
		newClassroomBySchool := make(map[string]string)

		// map of school_id -> [{}, {}, ...] -> list of classroom's subjects
		old_classroom_subjects := make(map[string][]*models.Subject)
		new_classroom_subjects := make(map[string][]*models.Subject)

		//new classroom id mappings with school id
		for _, v := range model.Classrooms {
			if v.ClassroomId != "" {
				class, err := d.ClassroomsFindById(ctx, v.ClassroomId)
				if err != nil {
					return nil, err
				}
				newClassroomBySchool[class.SchoolId] = v.ClassroomId
			}
		}

		//load relations
		err = d.UsersLoadRelations(ctx, &[]*models.User{model}, true)
		if err != nil {
			return nil, err
		}

		// Retrieve period grades for the student in the old classroom
		periodGrades, err := d.PeriodGradeByStudent(ctx, model.ID)
		if err != nil {
			return nil, err
		}

		classroomsToDelete := []string{}
		//old classroom id mappings with school id

		for _, class := range model.Classrooms {
			if class.Classroom == nil {
				continue
			}
			schoolID := class.Classroom.SchoolId
			existingClassroomID := class.ClassroomId
			if newClassroomID, exists := newClassroomBySchool[schoolID]; exists && existingClassroomID != newClassroomID {
				classroomsToDelete = append(classroomsToDelete, existingClassroomID)
			}
			oldClassroomBySchool[class.Classroom.SchoolId] = existingClassroomID
		}

		// Retrieve and map subjects for old classroom
		for schoolID, existingClassroomID := range oldClassroomBySchool {
			oldSubjects, err := d.SubjectsFindByClassroomId(ctx, existingClassroomID)
			if err != nil {
				return nil, err
			}
			old_classroom_subjects[schoolID] = oldSubjects
		}

		// Retrieve and map subjects for old classroom
		for schoolID, newClassroomID := range newClassroomBySchool {
			newSubjects, err := d.SubjectsFindByClassroomId(ctx, newClassroomID)
			if err != nil {
				return nil, err
			}
			new_classroom_subjects[schoolID] = newSubjects
		}

		// Get classrooms to delete where school id is identical with new classroom id

		// TODO: UserUpdateRelationsClassrooms
		//bulk delete from user_classrooms table
		if len(classroomsToDelete) > 0 {
			err = d.runQuery(ctx, func(tx *pgxpool.Conn) error {
				_, err := tx.Exec(ctx, sqlUserClassroomsDelete, model.ID, classroomsToDelete)
				return err
			})
			if err != nil {
				utils.LoggerDesc("Query error").Error(err)
				return nil, err
			}
		}

		// TODO: UserUpdateRelationsClassrooms
		// create relations in user_classrooms and update user_payments considering the school
		for schoolID, newClassroomID := range newClassroomBySchool {
			// Insert the new classrooms regardless of school (old classrooms in same school already deleted) refer classroomsToDelete
			err = d.runQuery(ctx, func(tx *pgxpool.Conn) error {
				_, err := tx.Exec(ctx, sqlUserClassroomsInsert, model.ID, newClassroomID)
				return err
			})
			if err != nil {
				utils.LoggerDesc("Insert Query error").Error(err)
				return nil, err
			}

			// Conditionally update payment if there was an existing classroom for the same school, we should not update classroom payment if schools are not matching. refer bilim merkezi payment
			if oldClassroomID, exists := oldClassroomBySchool[schoolID]; exists && oldClassroomID != newClassroomID {
				// Update the payment for the new classroom in the same school
				_, err = d.UpdateUserPaymentClassroom(ctx, model.ID, newClassroomID)
				if err != nil {
					utils.LoggerDesc("Update Payment error").Error(err)
					return nil, err
				}

				// Map old subjects to new subjects and insert into period grades in the same school
				if len(old_classroom_subjects[schoolID]) > 0 && len(new_classroom_subjects[schoolID]) > 0 {
					err = d.MapOldSubjectsToNewSubjectsInPeriodGrade(
						ctx,
						model.ID,
						periodGrades,
						old_classroom_subjects[schoolID],
						new_classroom_subjects[schoolID],
					)
					if err != nil {
						return nil, err
					}
				}

			}
		}

	}

	return model, nil
}

// TODO: rename UserUpdateRelationsMovePeriodGrades
func (d *PgxStore) MapOldSubjectsToNewSubjectsInPeriodGrade(
	ctx context.Context,
	student_id string,
	periodGrades []*models.PeriodGrade,
	oldSubjects []*models.Subject,
	newSubjects []*models.Subject,
) error {
	now := time.Now()
	newSubjectIDs := make([]string, len(newSubjects))
	for i, subject := range newSubjects {
		newSubjectIDs[i] = subject.ID
	}

	// Bulk delete newSubjects
	if len(newSubjectIDs) > 0 {
		err := d.DeletePeriodGradeByStudentAndSubjects(ctx, student_id, newSubjectIDs)
		if err != nil {
			return err
		}
	}

	for _, periodGrade := range periodGrades {
		// skip with no grade or absent
		if periodGrade.GetAbsentCount() == 0 && periodGrade.GetGradeCount() == 0 {
			continue
		}
		// Find the old subject details
		var oldSubject *models.Subject
		for _, subject := range oldSubjects {
			if subject.ID == *periodGrade.SubjectId {
				oldSubject = subject
				break
			}
		}

		if oldSubject == nil {
			continue // If no matching oldsubject is found in period_grades, skip to the next period grade
		}

		// finding the matching subject in the new classroom that has the same subject name and clasroom_type_key
		for _, newSubject := range newSubjects {
			if *oldSubject.Name == *newSubject.Name {
				if newSubject.ParentId != nil {
					continue
				}

				newPeriodGrade := &models.PeriodGrade{
					PeriodId:       periodGrade.PeriodId,
					PeriodKey:      periodGrade.PeriodKey,
					SubjectId:      &newSubject.ID,
					StudentId:      periodGrade.StudentId,
					ExamId:         periodGrade.ExamId,
					LessonCount:    periodGrade.LessonCount,
					AbsentCount:    0,
					GradeCount:     0,
					GradeSum:       0,
					PrevGradeCount: 0,
					PrevGradeSum:   0,
					OldGradeCount:  periodGrade.GetGradeCount(),
					OldGradeSum:    periodGrade.GetGradeSum(),
					OldAbsentCount: periodGrade.GetAbsentCount(),
					UpdatedAt:      &now,
					CreatedAt:      &now,
				}
				// TODO: append list periodGradesCreateList
				// TODO: and use PeriodGradesCreateBatch
				_, err := d.PeriodGradesCreate(ctx, newPeriodGrade)
				if err != nil {
					return err
				}

				break //found new subject to be inserted and inserted
			}
		}
	}
	return nil
}

func (d *PgxStore) UserChangeSchoolAndClassroom(ctx context.Context, studentId, schoolId, classroomId *string) error {
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		_, err = tx.Query(ctx, sqlUserSchoolsInsert, studentId, schoolId, models.RoleStudent)
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return err
	}

	err = d.runQuery(ctx, func(tx *pgxpool.Conn) error {
		_, err := tx.Exec(ctx, sqlUserClassroomsInsert, studentId, classroomId)
		return err
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return err
	}
	return nil
}

func (d PgxStore) UserCreate(ctx context.Context, model *models.User) (*models.User, error) {
	qs, args := UserCreateQuery(model)
	qs += " RETURNING uid"
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		err = tx.QueryRow(ctx, qs, args...).Scan(&model.ID)
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}

	editModel, err := d.UsersFindById(ctx, model.ID)
	if err != nil {
		return nil, err
	}
	return editModel, nil
}

func (d *PgxStore) UserUpdate(ctx context.Context, model *models.User) (*models.User, error) {
	// origModel := d.UsersFindById(strconv.Itoa(int(model.ID)))
	qs, args := UserUpdateQuery(model)
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		_, err = tx.Exec(ctx, qs, args...)
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}

	editModel, err := d.UsersFindById(ctx, model.ID)
	if err != nil {
		return nil, err
	}
	return editModel, nil
}

func (d *PgxStore) UsersOnlineCount(ctx context.Context, schoolId *string) (int, error) {
	// origModel := d.UsersFindById(strconv.Itoa(int(model.ID)))
	c := 0
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		now := time.Now()
		err = tx.QueryRow(ctx, sqlUsersOnlineCount, now.Add(time.Minute*-15), now).Scan(&c)
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return 0, err
	}
	return c, nil
}

func (d *PgxStore) UserDelete(ctx context.Context, l []*models.User) ([]*models.User, error) {
	ids := []string{}
	for _, i := range l {
		ids = append(ids, i.ID)
	}
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		_, err = tx.Query(ctx, sqlUserDelete, (ids))
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}
	return l, err
}

func (d *PgxStore) UserDeleteSchool(ctx context.Context, l []*models.User, schoolIds []string) ([]*models.User, error) {
	ids := []string{}
	for _, i := range l {
		ids = append(ids, i.ID)
	}
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		_, err = tx.Query(ctx, sqlUserSchoolsDeleteBySchool, ids, schoolIds)
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}
	return l, err
}

func (d *PgxStore) UserDeleteSchoolRole(ctx context.Context, userIds []string, schoolIds []string, roles []string) (int, error) {
	deleted := 0
	if slices.Contains(roles, string(models.RoleParent)) {
		err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
			cmd, err := tx.Exec(ctx, sqlUserParentsDeleteByRole, userIds, schoolIds)
			deleted = int(cmd.RowsAffected())
			return err
		})
		if err != nil {
			utils.LoggerDesc("Query error").Error(err)
			return 0, err
		}
		return deleted, nil
	}

	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		cmd, err := tx.Exec(ctx, sqlUserSchoolsDeleteByRole, userIds, schoolIds, roles)
		deleted = int(cmd.RowsAffected())
		return err
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return 0, err
	}
	return deleted, nil
}

func (d *PgxStore) UserDeleteFromClassroom(ctx context.Context, userId string, classroomIds []string) error {
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		_, err = tx.Exec(ctx, sqlUserClassroomsDelete, userId, classroomIds)
		return err
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return err
	}
	return nil
}

func (d *PgxStore) UserDeleteAllRelations(ctx context.Context, l *[]*models.User) error {
	// TODO: remove other relations manually
	return nil
}

func (d *PgxStore) UsersLoadRelations(ctx context.Context, users *[]*models.User, isDetail bool) error {
	ids := []string{}
	for _, user := range *users {
		if user.ID == "" {
			err := errors.New("user id is null")
			utils.LoggerDesc("Scan error").Error(err)
			return err
		}
		ids = append(ids, user.ID)
	}
	if len(ids) < 1 {
		return nil
	}
	// load schools
	var schoolParents []*models.School
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		rows, err := tx.Query(ctx, sqlUserSchools, (ids))
		for rows.Next() {
			school := &models.School{}
			userSchool := models.UserSchool{}
			mid := ""
			// scan 2 or more if there is also school cols
			cols := []interface{}{&userSchool.RoleCode, &mid}
			if vl, _ := rows.Values(); vl[2] != nil {
				cols = append(cols, parseColumnsForScan(school)...)
			} else {
				var t interface{}
				for range parseColumnsForScan(school) {
					cols = append(cols, &t)
				}
				school = nil
			}
			err = rows.Scan(cols...)
			if err != nil {
				utils.LoggerDesc("Scan error").Error(err)
				return err
			}
			// ignore incorrect roles from db
			if !slices.Contains(models.DefaultRoles, userSchool.RoleCode) {
				continue
			}
			for _, user := range *users {
				if (user.ID) == mid {
					var schoolUid *string
					if school != nil {
						schoolParents = append(schoolParents, school)
						schoolUid = &school.ID
					}
					user.Schools = append(user.Schools, &models.UserSchool{
						SchoolUid: schoolUid,
						UserId:    user.ID,
						School:    school,
						RoleCode:  userSchool.RoleCode,
					})
				}
			}
		}
		return
	})
	err = d.SchoolsLoadParents(ctx, &schoolParents)
	if err != nil {
		return err
	}

	// load teacher classroom
	err = d.UsersLoadRelationsTeacherClassroom(ctx, users)
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return err
	}
	// if (*users)[0].Schools[0].SchoolId {
	// }
	// load classrooms
	err = d.UsersLoadRelationsClassrooms(ctx, users)
	if err != nil {
		return err
	}

	if isDetail {
		// load parents
		err = d.UsersLoadRelationsParents(ctx, users)
		if err != nil {
			return err
		}

		// load children
		err = d.UsersLoadRelationsChildren(ctx, users)
		if err != nil {
			return err
		}
	}

	// TODO: if v.Schools == nil  then create v.Schools = []*models.UserSchool{}

	// fill children[].school_id by ChildrenSchoolIds
	for _, user := range *users {
		if user.Children == nil {
			user.Children = []*models.User{}
			for _, childSchoolId := range user.ChildrenSchoolIds {
				user.Children = append(user.Children, &models.User{
					SchoolId: &childSchoolId,
				})
			}
		}
	}

	// load schools for empty children[].school but with children[].school_id
	err = d.UsersLoadRelationsChildrenSchool(ctx, users)
	if err != nil {
		return err
	}

	// for parents load role and schools
	usedSchools := map[string]bool{}
	parentRole := string(models.RoleParent)
	for _, user := range *users {
		for _, child := range user.Children {
			if child.SchoolId != nil {
				if !usedSchools[user.ID+*child.SchoolId] {
					user.Role = &parentRole
					school := child.Schools[0].School
					user.SchoolId = &school.ID
					user.SchoolName = school.Name
					if school.Parent != nil {
						user.SchoolParent = school.Parent.Name
					}

					user.Schools = append(user.Schools, &models.UserSchool{
						SchoolUid: child.SchoolId,
						RoleCode:  models.RoleParent,
						School:    school,
					})
				}
				usedSchools[user.ID+*child.SchoolId] = true
			}
		}
	}

	// clear empty children (ref: parent load school)
	for _, user := range *users {
		if len(user.Children) > 0 && user.Children[0].ID == "" {
			user.Children = nil
		}
	}
	return nil
}

func (d *PgxStore) UsersLoadRelationsParentSchool(ctx context.Context, users *[]*models.User) error {
	parentSchoolIds := []string{}
	for _, user := range *users {
		for _, parent := range user.Parents {
			if parent.Schools == nil && parent.SchoolId != nil {
				parentSchoolIds = append(parentSchoolIds, *parent.SchoolId)
			}
		}
	}
	if len(parentSchoolIds) > 0 {
		schools, err := d.SchoolsFindByIds(ctx, parentSchoolIds)
		if err != nil {
			return err
		}
		err = d.SchoolsLoadParents(ctx, &schools)
		if err != nil {
			return err
		}
		for _, user := range *users {
			for _, parent := range user.Parents {
				for _, school := range schools {
					if parent.Schools == nil && parent.SchoolId != nil && *parent.SchoolId == school.ID {
						parent.SchoolId = &school.ID
						parent.SchoolName = school.Name
						if school.Parent != nil {
							parent.SchoolParent = school.Parent.Name
						}
						parent.Schools = []*models.UserSchool{
							{
								School:    school,
								SchoolUid: &school.ID,
								RoleCode:  models.RoleParent,
							},
						}
					}
				}
			}
		}
	}
	return nil
}

func (d *PgxStore) UsersLoadRelationsChildrenSchool(ctx context.Context, users *[]*models.User) error {
	childrenSchoolIds := []string{}
	for _, user := range *users {
		for _, child := range user.Children {
			if child.Schools == nil && child.SchoolId != nil {
				childrenSchoolIds = append(childrenSchoolIds, *child.SchoolId)
			}
		}
	}
	if len(childrenSchoolIds) > 0 {
		schools, err := d.SchoolsFindByIds(ctx, childrenSchoolIds)
		if err != nil {
			return err
		}
		err = d.SchoolsLoadParents(ctx, &schools)
		if err != nil {
			return err
		}
		for _, user := range *users {
			for _, child := range user.Children {
				for _, school := range schools {
					if child.Schools == nil && child.SchoolId != nil && *child.SchoolId == school.ID {
						child.SchoolId = &school.ID
						child.SchoolName = school.Name
						if school.Parent != nil {
							child.SchoolParent = school.Parent.Name
						}
						child.Schools = []*models.UserSchool{
							{
								School:    school,
								SchoolUid: &school.ID,
								RoleCode:  models.RoleStudent,
							},
						}
					}
				}
			}
		}
	}
	return nil
}

func (d *PgxStore) UsersLoadRelationsParents(ctx context.Context, l *[]*models.User) error {
	ids := []string{}
	for _, m := range *l {
		m.Parents = []*models.User{}
		ids = append(ids, m.ID)
	}
	if len(ids) < 1 {
		return nil
	}
	// load
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		rows, err := tx.Query(ctx, sqlUserParents, (ids))
		for rows.Next() {
			sub := models.User{}
			pid := []string{}
			err = scanUser(rows, &sub, &pid)
			if err != nil {
				utils.LoggerDesc("Scan error").Error(err)
				return err
			}
			for _, m := range *l {
				if slices.Contains(pid, m.ID) {
					m.Parents = append(m.Parents, &sub)
				}
			}
		}
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return err
	}
	err = d.UsersLoadRelationsParentSchool(ctx, l)
	if err != nil {
		return err
	}
	return nil
}

func (d *PgxStore) UsersLoadRelationsChildren(ctx context.Context, l *[]*models.User) error {
	ids := []string{}

	for _, m := range *l {
		m.Children = []*models.User{}
		ids = append(ids, m.ID)
	}
	if len(ids) < 1 {
		return nil
	}
	// load
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		rows, err := tx.Query(ctx, sqlUserChildren, ids)
		for rows.Next() {
			sub := models.User{}
			pid := []string{}
			err = scanUser(rows, &sub, &pid)
			if err != nil {
				utils.LoggerDesc("Scan error").Error(err)
				return err
			}

			for _, user := range *l {
				if slices.Contains(pid, user.ID) {
					user.Children = append(user.Children, &sub)
				}
			}
		}
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return err
	}
	err = d.UsersLoadRelationsChildrenSchool(ctx, l)
	if err != nil {
		return err
	}
	return nil
}

func (d *PgxStore) UsersLoadRelationsTeacherClassroom(ctx context.Context, l *[]*models.User) error {
	ids := []string{}
	for _, m := range *l {
		ids = append(ids, m.ID)
	}
	if len(ids) < 1 {
		return nil
	}
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		rows, err := tx.Query(ctx, sqlUserTeacherClassroom, (ids))
		for rows.Next() {
			sub := models.Classroom{}
			pid := ""
			err = scanClassroom(rows, &sub, &pid)
			if err != nil {
				utils.LoggerDesc("Scan error").Error(err)
				return err
			}
			for _, m := range *l {
				if (m.ID) == pid {
					m.TeacherClassroom = &sub
				}
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

func (d *PgxStore) UsersLoadRelationsClassroomsAll(ctx context.Context, l *[]*models.User) error {
	ids := []string{}
	for _, m := range *l {
		ids = append(ids, m.ID)
	}
	if len(ids) < 1 {
		return nil
	}
	// load
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		rows, err := tx.Query(ctx, sqlUserClassroomsAll, (ids))
		for rows.Next() {
			sub := models.Classroom{}
			piv := models.UserClassroom{}
			subId := ""
			err = scanClassroom(rows, &sub, &subId, &piv.Type)
			if err != nil {
				return err
			}
			for _, m := range *l {
				if (m.ID) == subId {
					m.Classrooms = append(m.Classrooms, &models.UserClassroom{
						ClassroomId: sub.ID,
						UserId:      m.ID,
						Classroom:   &sub,
						Type:        piv.Type,
					})
				}
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

func (d *PgxStore) UsersLoadRelationsClassrooms(ctx context.Context, l *[]*models.User) error {
	ids := []string{}
	for _, m := range *l {
		m.Classrooms = []*models.UserClassroom{}
		ids = append(ids, m.ID)
	}
	if len(ids) < 1 {
		return nil
	}
	// load
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		rows, err := tx.Query(ctx, sqlUserClassrooms, (ids))
		for rows.Next() {
			sub := models.Classroom{}
			piv := models.UserClassroom{}
			subId := ""
			err = scanClassroom(rows, &sub, &subId, &piv.Type, &piv.TariffEndAt, &piv.TariffType)
			if err != nil {
				return err
			}
			for _, m := range *l {
				if (m.ID) == subId {
					m.Classrooms = append(m.Classrooms, &models.UserClassroom{
						ClassroomId: sub.ID,
						UserId:      m.ID,
						Classroom:   &sub,
						Type:        piv.Type,
						TariffType:  piv.TariffType,
						TariffEndAt: piv.TariffEndAt,
					})
				}
			}
		}

		for k, m := range *l {
			classroomIds := []string{}
			classrooms := []*models.UserClassroom{}
			for _, c := range m.Classrooms {
				if !slices.Contains(classroomIds, c.ClassroomId) {
					classrooms = append(classrooms, c)
				}
				classroomIds = append(classroomIds, c.ClassroomId)
			}

			(*l)[k].Classrooms = classrooms
		}
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return err
	}
	return nil
}

func UserAtomicQuery(m *models.User, isCreate bool) map[string]interface{} {
	q := map[string]interface{}{}
	if m.FirstName != nil {
		q["first_name"] = m.FirstName
	}
	if m.LastName != nil {
		q["last_name"] = m.LastName
	}
	if m.MiddleName != nil {
		q["middle_name"] = m.MiddleName
	}
	if m.Username != nil {
		q["username"] = m.Username
	}
	if m.Status != nil {
		q["status"] = m.Status
	}
	q["phone"] = m.Phone
	if m.Email != nil {
		q["email"] = m.Email
	}
	q["birthday"] = m.Birthday
	if m.Gender != nil {
		q["gender"] = m.Gender
	}
	if m.Address != nil {
		q["address"] = m.Address
	}
	if m.Avatar != nil {
		if *m.Avatar == "" {
			q["avatar"] = nil
		} else {
			q["avatar"] = m.Avatar

		}
	}
	if m.Password != nil {
		q["password"] = m.Password
	}
	if m.PassportNumber != nil {
		q["passport_number"] = m.PassportNumber
	}
	if m.BirthCertNumber != nil {
		q["birth_cert_number"] = m.BirthCertNumber
	}
	if m.ApplyNumber != nil {
		q["apply_number"] = m.ApplyNumber
	}
	if m.WorkTitle != nil {
		q["work_title"] = m.WorkTitle
	}
	if m.WorkPlace != nil {
		q["work_place"] = m.WorkPlace
	}
	if m.District != nil {
		q["district"] = m.District
	}
	if m.Reference != nil {
		q["reference"] = m.Reference
	}
	if m.NickName != nil {
		q["nickname"] = m.NickName
	}
	if m.EducationTitle != nil {
		q["education_title"] = m.EducationTitle
	}
	if m.EducationPlace != nil {
		q["education_place"] = m.EducationPlace
	}
	if m.EducationGroup != nil {
		q["education_group"] = m.EducationGroup
	}
	if m.Documents != nil {
		q["documents"] = m.Documents
	}
	if m.DocumentFiles != nil {
		q["document_files"] = m.DocumentFiles
	}
	if isCreate {
		q["created_at"] = time.Now()
	}
	q["updated_at"] = time.Now()
	return q
}

func UsersListBuildQuery(f models.UserFilterRequest, args []interface{}) (string, []interface{}) {
	var wheres string = ""
	joins := ""
	isParentJoin := false
	isParentClassroomJoin := false
	isOnlyParentJoin := false
	if f.ID != nil && *f.ID != "" {
		args = append(args, *f.ID)
		wheres += " and u.uid=$" + strconv.Itoa(len(args))
	}
	if f.NotID != nil && *f.NotID != "" {
		args = append(args, *f.NotID)
		wheres += " and u.uid <> $" + strconv.Itoa(len(args))
	}
	if f.Ids != nil {
		args = append(args, *f.Ids)
		wheres += " and u.uid = ANY($" + strconv.Itoa(len(args)) + "::uuid[])"
	}
	if f.ParentId != nil {
		isParentJoin = true
		args = append(args, *f.ParentId)
		wheres += " and up_child.parent_uid =$" + strconv.Itoa(len(args))
	}
	if f.Username != nil {
		args = append(args, *f.Username)
		wheres += " and u.username =$" + strconv.Itoa(len(args))
	}
	if f.IsActive != nil && *f.IsActive {
		args = append(args, time.Now().Add(-time.Hour*15))
		wheres += " and " + sqlUserLat + " >= $" + strconv.Itoa(len(args))
	}
	if f.Role != nil && *f.Role != "" {
		if *f.Role == string(models.RoleParent) && f.NoParent != nil && *f.NoParent {
			isOnlyParentJoin = true
		}
		if *f.Role == string(models.RoleParent) && f.SchoolId != nil && *f.SchoolId != "" {
			isOnlyParentJoin = true
			args = append(args, *f.SchoolId)
			wheres += " and up.school_uid=$" + strconv.Itoa(len(args))
		} else {
			args = append(args, *f.Role)
			wheres += " and us.role_code=$" + strconv.Itoa(len(args))
		}
	}
	if f.TariffEndMin != nil && !f.TariffEndMin.IsZero() {
		args = append(args, *f.TariffEndMin)
		wheres += " and p.expires_at > $" + strconv.Itoa(len(args))
	}
	if f.Roles != nil && len(*f.Roles) > 0 {
		if slices.Contains(*f.Roles, string(models.RoleParent)) && f.SchoolId != nil && *f.SchoolId != "" {
			isOnlyParentJoin = true
			args = append(args, *f.SchoolId)
			wheres += " and up.school_uid=$" + strconv.Itoa(len(args))
		} else {
			args = append(args, *f.Roles)
			wheres += " and us.role_code = ANY($" + strconv.Itoa(len(args)) + "::text[])"
		}
	}
	if f.SchoolId != nil && *f.SchoolId != "" && !isOnlyParentJoin {
		args = append(args, *f.SchoolId)
		wheres += " and us.school_uid=$" + strconv.Itoa(len(args))
	}
	if f.SchoolIds != nil && !isOnlyParentJoin {
		args = append(args, *f.SchoolIds)
		wheres += " and us.school_uid = ANY($" + strconv.Itoa(len(args)) + "::uuid[])"
	}
	if f.Status != nil && *f.Status != "" {
		args = append(args, *f.Status)
		wheres += " and status=$" + strconv.Itoa(len(args))
	}
	if f.Gender != nil {
		args = append(args, *f.Gender)
		wheres += " and gender=$" + strconv.Itoa(len(args))
	}
	if f.Birthday != nil && *f.Birthday != "" {
		args = append(args, *f.Birthday)
		wheres += " and birthday=$" + strconv.Itoa(len(args))
	}
	if f.BirthdayToday != nil && !f.BirthdayToday.IsZero() {
		args = append(args, *f.BirthdayToday)
		wheres += " and to_char(birthday, 'MM-DD')=to_char($" + strconv.Itoa(len(args)) + "::timestamp, 'MM-DD')"
	}
	if f.ClassroomIdForBirthday != nil && *f.ClassroomIdForBirthday != "" {
		args = append(args, *f.ClassroomIdForBirthday)
		wheres += " and uc.classroom_uid=$" + strconv.Itoa(len(args))
	}
	if f.Phone != nil && *f.Phone != "" {
		args = append(args, *f.Phone)
		wheres += " and u.phone=$" + strconv.Itoa(len(args))
	}
	if f.FirstName != nil && *f.FirstName != "" {
		args = append(args, *f.FirstName)
		wheres += " and first_name=$" + strconv.Itoa(len(args))
	}
	if f.LastName != nil && *f.LastName != "" {
		args = append(args, *f.LastName)
		wheres += " and last_name=$" + strconv.Itoa(len(args))
	}
	if f.Address != nil && *f.Address != "" {
		args = append(args, strings.ToLower(*f.Address))
		wheres += " and lower(address) like '%$" + strconv.Itoa(len(args)) + "%'"
	}
	if f.ClassroomId != nil {
		if f.Roles != nil && slices.Contains(*f.Roles, string(models.RoleTeacher)) {
			if *f.ClassroomId != "" {
				args = append(args, *f.ClassroomId)
				wheres += " and c.uid=$" + strconv.Itoa(len(args))
			} else {
				wheres += " and c.uid is null"
			}
		} else if f.Roles != nil && slices.Contains(*f.Roles, string(models.RoleParent)) {
			isParentClassroomJoin = true
			if *f.ClassroomId != "" {
				args = append(args, *f.ClassroomId)
				wheres += " and uc_child.classroom_uid=$" + strconv.Itoa(len(args))
			} else {
				wheres += " and uc_child.classroom_uid is null"
			}
		} else {
			if *f.ClassroomId != "" {
				args = append(args, *f.ClassroomId)
				wheres += " and uc.classroom_uid=$" + strconv.Itoa(len(args))
			} else {
				wheres += " and uc.classroom_uid is null"
			}
		}
	}
	if f.ClassroomType != nil {
		args = append(args, *f.ClassroomType)
		wheres += " and uc.type=$" + strconv.Itoa(len(args))
	}
	if f.ClassroomName != nil {
		args = append(args, *f.ClassroomName)
		wheres += " and c.name=$" + strconv.Itoa(len(args))
	}
	if f.LowFirstName != nil {
		firstName := *f.LowFirstName
		args = append(args, replaceLetters(firstName))
		wheres += " and $" + strconv.Itoa(len(args)) + "=REPLACE(REPLACE(REPLACE(REPLACE(REPLACE(REPLACE(LOWER(u.first_name), 'ý', 'y'), 'ö', 'o'), 'ä', 'a'), 'ň', 'n'), 'ü', 'u'), 'ç', 'c')"
	}
	if f.LowLastName != nil {
		lastName := *f.LowLastName
		args = append(args, replaceLetters(lastName))
		wheres += " and $" + strconv.Itoa(len(args)) + "=REPLACE(REPLACE(REPLACE(REPLACE(REPLACE(REPLACE(LOWER(u.last_name), 'ý', 'y'), 'ö', 'o'), 'ä', 'a'), 'ň', 'n'), 'ü', 'u'), 'ç', 'c')"
	}
	if f.ClassroomTypeKey != nil {
		args = append(args, *f.ClassroomTypeKey)
		wheres += " and uc.type_key=$" + strconv.Itoa(len(args))
	}
	if f.Search != nil && *f.Search != "" {
		*f.Search = strings.ToLower(*f.Search)
		args = append(args, *f.Search)
		tmp := "$" + strconv.Itoa(len(args))
		wheres += " and (lower(u.first_name) like '%' ||" + tmp + "|| '%' or lower(u.last_name) like '%' ||" + tmp +
			"||'%' or lower(u.middle_name) like '%'||" + tmp + "||'%' or lower(u.username) like '%'||" + tmp +
			"||'%' or lower(u.phone) like '%'||" + tmp + "||'%' or lower(u.email) like '%'||" + tmp + "||'%')"
	}
	if f.LessonHours != nil && len(*f.LessonHours) == 2 {
		minHour := (*f.LessonHours)[0]
		maxHour := (*f.LessonHours)[1]

		joins += " LEFT JOIN subjects sb ON sb.teacher_uid = u.uid"

		args = append(args, minHour, maxHour)
		wheres += " AND sb.week_hours BETWEEN $" + strconv.Itoa(len(args)-1) + " AND $" + strconv.Itoa(len(args))
	}
	wheres += " group by u.uid"
	if f.ParentsCount != nil {
		isParentJoin = true
		args = append(args, *f.ParentsCount)
		wheres += " HAVING COUNT(distinct up_child.parent_uid)=$" + strconv.Itoa(len(args))
	}
	if f.Sort != nil && *f.Sort != "" {
		dir := "desc"
		if strings.HasSuffix(*f.Sort, "~") {
			dir = "asc"
		}
		*f.Sort = strings.ReplaceAll(*f.Sort, "~", "")
		wheres += " order by " + *f.Sort + " " + dir
	} else {
		wheres += ` order by u.last_name, u.first_name, u.middle_name asc`
	}

	// add sql, joins and execute
	qs := sqlUserSelectMany
	if isOnlyParentJoin {
		qs = sqlUserSelectManyParent
	}
	if isParentClassroomJoin {
		joins += " left join user_parents up_child on up_child.child_uid=u.uid"
		joins += " left join user_parents up_parent on up_parent.parent_uid=u.uid"
		joins += " left join user_classrooms uc_child on uc_child.user_uid=up_parent.child_uid"
	} else if isParentJoin {
		joins += " left join user_parents up_child on up_child.child_uid=u.uid"
		joins += " left join user_parents up_parent on up_parent.parent_uid=u.uid"
	}
	qs = strings.ReplaceAll(qs, "u.uid=u.uid", "u.uid=u.uid "+wheres+" ")
	qs = strings.ReplaceAll(qs, "where", " "+joins+" where")
	return qs, args
}

func replaceLetters(s string) string {
	replacements := map[string]string{
		"ý": "y",
		"ä": "a",
		"ü": "u",
		"ö": "o",
		"ň": "n",
		"ç": "c",
	}

	s = strings.ToLower(s)
	for old, new := range replacements {
		s = strings.ReplaceAll(s, old, new)
	}
	return s
}

func UserCreateQuery(m *models.User) (string, []interface{}) {
	args := []interface{}{}
	cols := ""
	vals := ""
	q := UserAtomicQuery(m, true)
	for k, v := range q {
		args = append(args, v)
		cols += ", " + k
		vals += ", $" + strconv.Itoa(len(args))
	}
	qs := sqlUserInsert + " (" + strings.Trim(cols, ", ") + ") VALUES (" + strings.Trim(vals, ", ") + ")"
	return qs, args
}

func UserUpdateQuery(m *models.User) (string, []interface{}) {
	args := []interface{}{}
	sets := ""
	q := UserAtomicQuery(m, false)
	for k, v := range q {
		args = append(args, v)
		sets += ", " + k + "=$" + strconv.Itoa(len(args))
	}
	args = append(args, m.ID)
	qs := strings.ReplaceAll(sqlUserUpdate, "set uid=uid", "set uid=uid "+sets+" ") + "where uid=$" + strconv.Itoa(len(args))
	return qs, args
}

func (d *PgxStore) UsersLoadCount(ctx context.Context, schoolIds []string) (models.DashboardUsersCount, error) {
	item := models.DashboardUsersCount{}
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		err = tx.QueryRow(ctx, sqlUserSelectCount, (schoolIds)).
			Scan(&item.SchoolId, &item.ClassroomsCount, &item.StudentsCount, &item.ParentsCount,
				&item.TeachersCount, &item.PrincipalsCount, &item.OrganizationsCount, &item.SchoolsCount, &item.TimetablesCount, &item.UsersOnlineCount, &item.StudentsWithParentsCount)
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return item, err
	}
	return item, nil
}

func (d *PgxStore) UsersLoadCountBySchool(ctx context.Context, schoolIds []string) ([]models.DashboardUsersCount, error) {
	items := []models.DashboardUsersCount{}
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		rows, err := tx.Query(ctx, sqlUserSelectCountBySchool, schoolIds) //time.Now().AddDate(0, -1, 0))
		if err != nil {
			return err
		}
		for rows.Next() {
			item := models.DashboardUsersCount{}
			err = rows.Scan(&item.SchoolId, &item.ClassroomsCount, &item.StudentsCount, &item.ParentsCount,
				&item.TeachersCount, &item.PrincipalsCount, &item.OrganizationsCount, &item.SchoolsCount, &item.TimetablesCount, &item.UsersOnlineCount, &item.StudentsWithParentsCount)
			if err != nil {
				return err
			}
			items = append(items, item)
		}
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return items, err
	}
	return items, nil
}

func (d *PgxStore) UsersLoadCountByClassroom(ctx context.Context, schoolIds []string) ([]models.DashboardUsersCountByClassroom, error) {
	items := []models.DashboardUsersCountByClassroom{}
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		rows, err := tx.Query(ctx, sqlUserSelectCountByClassroom, schoolIds) //time.Now().AddDate(0, -1, 0))
		if err != nil {
			return err
		}
		for rows.Next() {
			item := models.DashboardUsersCountByClassroom{
				Classroom: models.Classroom{},
			}
			err = scanClassroom(rows, &item.Classroom, &item.SchoolId, &item.StudentsCount, &item.ParentsCount, &item.ParentsOnlineCount, &item.ParentsPaidCount)
			if err != nil {
				return err
			}
			items = append(items, item)
		}
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return items, err
	}
	return items, nil
}

func (store *PgxStore) GetTeacherIdByName(
	ctx context.Context,
	dto models.GetTeacherIdByNameQueryDto,
) (*string, error) {

	stmt := `
		SELECT u.uid
		FROM users AS u
		LEFT JOIN user_schools AS user_school
			ON user_school.user_uid = u.uid
		WHERE
			user_school.role_code = 'teacher'
			AND user_school.school_uid = $1
			AND u.first_name = $2
	`
	args := []interface{}{
		dto.SchoolId,
		dto.FirstName,
	}

	if dto.LastName != nil {
		args = append(args, *dto.LastName)
		stmt += " AND last_name = $" + strconv.Itoa(len(args))
	}

	var id string
	err := store.runQuery(ctx, func(conn *pgxpool.Conn) (err error) {
		row := conn.QueryRow(
			ctx,
			stmt,
			args...,
		)
		err = row.Scan(&id)
		if err != nil {
			log.Println(err)
			return err
		}
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}

	return &id, nil
}
