package app

import (
	"context"
	"slices"
	"sort"
	"time"

	"github.com/mekdep/server/internal/api/utils"
	"github.com/mekdep/server/internal/models"
	"github.com/mekdep/server/internal/store"
	"go.elastic.co/apm/v2"
)

type UserDashboard map[string]interface{}

func (a App) Dashboards(ses *utils.Session, startDate time.Time, endDate time.Time) (map[models.Role]UserDashboard, error) {
	sp, ctx := apm.StartSpan(ses.Context(), "Dashboards", "app")
	ses.SetContext(ctx)
	defer sp.End()
	d := map[models.Role]UserDashboard{}
	err := a.UserActionCheckRead(ses, PermUser, func(u *models.User) error {
		sesId := ""
		if ses.GetSessionId() != nil {
			sesId = *ses.GetSessionId()
		}
		r := ses.GetRole()
		if r == nil || u == nil {
			return ErrUnauthorized
		}
		var err error
		if *r == models.RoleTeacher {
			d[*r], err = dashboardTeacher(ses, *ses.GetSchoolId(), u.ID, startDate, endDate)
			if err != nil {
				return err
			}
		}
		if *r == models.RolePrincipal {
			cKey := sesId + "_dashboard_users_count" + startDate.Format(time.DateOnly) + endDate.Format(time.DateOnly)
			if val, found := a.cache.Get(cKey); found {
				d[*r] = val.(UserDashboard)
			} else {
				d[*r], err = dashboardPrincipal(ses, *ses.GetSchoolId(), startDate, endDate)
				if err != nil {
					return err
				}
				a.cache.Set(cKey, d[*r], 0)
			}
			dd := d[*r]

			err = dashboardPrincipalCurrentLesson(ses, *ses.GetSchoolId(), startDate, &dd)

			if err != nil {
				return err
			}
			d[*r] = dd
		}
		if *r == models.RoleAdmin || *r == models.RoleOrganization {
			cKey := sesId + "_dashboard_users_count" + startDate.Format(time.DateOnly) + endDate.Format(time.DateOnly)
			if val, found := a.cache.Get(cKey); found {
				d[*r] = val.(UserDashboard)
			} else {
				d[*r], err = dashboardAdmin(ses, ses.SchoolAllIds(), startDate, endDate)
				if err != nil {
					return err
				}
				a.cache.Set(cKey, d[*r], 0)
			}
		}
		return nil
	})
	return d, err
}
func (a App) DashboardDetails(ses *utils.Session, startDate time.Time, endDate time.Time, detailId *string) (map[models.Role]UserDashboard, error) {
	sp, ctx := apm.StartSpan(ses.Context(), "Dashboards", "app")
	ses.SetContext(ctx)
	defer sp.End()
	d := map[models.Role]UserDashboard{}
	err := a.UserActionCheckRead(ses, PermUser, func(u *models.User) error {
		sesId := ""
		if ses.GetSessionId() != nil {
			sesId = *ses.GetSessionId()
		}
		r := ses.GetRole()
		if r == nil || u == nil {
			return ErrUnauthorized
		}
		var err error
		if *r == models.RoleTeacher {
			d[*r], err = dashboardTeacher(ses, *ses.GetSchoolId(), u.ID, startDate, endDate)
			if err != nil {
				return err
			}
		}
		if *r == models.RolePrincipal {
			cKey := sesId + "_dashboard_details_users_count" + startDate.Format(time.DateOnly) + endDate.Format(time.DateOnly)
			if detailId != nil {
				cKey += *detailId
			}
			// if val, found := a.cache.Get(cKey); found {
			// 	d[*r] = val.(UserDashboard)
			// } else {
			d[*r], err = dashboardPrincipalDetails(ses, *ses.GetSchoolId(), startDate, endDate)
			if err != nil {
				return err
			}
			// 	a.cache.Set(cKey, d[*r], 0)
			// }
			dd := d[*r]

			err = dashboardPrincipalCurrentLesson(ses, *ses.GetSchoolId(), startDate, &dd)

			if err != nil {
				return err
			}
			d[*r] = dd
		}
		if *r == models.RoleAdmin || *r == models.RoleOrganization {
			cKey := sesId + "_dashboard_details_users_count" + startDate.Format(time.DateOnly) + endDate.Format(time.DateOnly)
			if detailId != nil {
				cKey += *detailId
			}
			if val, found := a.cache.Get(cKey); found {
				d[*r] = val.(UserDashboard)
			} else {
				d[*r], err = dashboardAdminDetails(ses, ses.SchoolAllIds(), startDate, endDate, detailId)
				if err != nil {
					return err
				}
				a.cache.Set(cKey, d[*r], 0)
			}
		}
		return nil
	})
	return d, err
}

func dashboardTeacher(ses *utils.Session, schoolId string, userId string, startDate time.Time, endDate time.Time) (UserDashboard, error) {
	sp, ctx := apm.StartSpan(ses.Context(), "dashboardTeacher", "app")
	ses.SetContext(ctx)
	defer sp.End()
	d := UserDashboard{}
	args := models.SubjectFilterRequest{
		TeacherIds: []string{userId},
	}
	args.Limit = new(int)
	*args.Limit = 100
	subjects, _, err := store.Store().SubjectsListFilters(context.Background(), &args)
	if err != nil {
		return nil, err
	}
	subjectHours := uint(0)
	classroomIds := []string{}
	for _, s := range subjects {
		if s.WeekHours != nil {
			subjectHours += *s.WeekHours
		}
		if !slices.Contains(classroomIds, s.ClassroomId) {
			classroomIds = append(classroomIds, s.ClassroomId)
		}
	}
	d["subject_hours_sum"] = subjectHours
	d["classroom_count"] = len(classroomIds)
	// set current shift
	return d, nil
}

func dashboardAdmin(ses *utils.Session, schoolIds []string, startDate, endDate time.Time) (UserDashboard, error) {
	sp, ctx := apm.StartSpan(ses.Context(), "dashboardAdmin", "app")
	ses.SetContext(ctx)
	defer sp.End()
	d := UserDashboard{}
	var err error
	// set user counts
	usersCount, err := store.Store().UsersLoadCount(ses.Context(), schoolIds)
	if err != nil {
		return nil, err
	}
	d["students_count"] = usersCount.StudentsCount
	d["parents_count"] = usersCount.ParentsCount
	d["teachers_count"] = usersCount.TeachersCount
	d["principals_count"] = usersCount.PrincipalsCount
	d["schools_count"] = usersCount.SchoolsCount
	d["organizations_count"] = usersCount.OrganizationsCount
	return d, nil
}
func dashboardAdminDetails(ses *utils.Session, schoolIds []string, startDate, endDate time.Time, detailSchoolId *string) (UserDashboard, error) {
	sp, ctx := apm.StartSpan(ses.Context(), "dashboardAdmin", "app")
	ses.SetContext(ctx)
	defer sp.End()
	d := UserDashboard{}
	var err error
	// set by_schools
	if detailSchoolId != nil {
		d, err = dashboardPrincipalDetails(ses, *detailSchoolId, startDate, endDate)
		if err != nil {
			return nil, err
		}
	} else {
		d["by_school"], err = dashboardAdminBySchool(ses, schoolIds, startDate, endDate)
		if err != nil {
			return nil, err
		}
	}
	return d, nil
}

func dashboardPrincipal(ses *utils.Session, schoolId string, startDate, endDate time.Time) (UserDashboard, error) {
	sp, ctx := apm.StartSpan(ses.Context(), "dashboardPrincipal", "app")
	ses.SetContext(ctx)
	defer sp.End()
	d := UserDashboard{}
	var err error
	// set counts
	usersCount, err := store.Store().UsersLoadCount(ses.Context(), []string{schoolId})
	if err != nil {
		return nil, err
	}
	d["students_count"] = usersCount.StudentsCount
	d["parents_count"] = usersCount.ParentsCount
	d["teachers_count"] = usersCount.TeachersCount
	d["principals_count"] = usersCount.PrincipalsCount
	d["classrooms_count"] = usersCount.ClassroomsCount
	return d, nil
}

func dashboardPrincipalDetails(ses *utils.Session, schoolId string, startDate, endDate time.Time) (UserDashboard, error) {
	sp, ctx := apm.StartSpan(ses.Context(), "dashboardPrincipal", "app")
	ses.SetContext(ctx)
	defer sp.End()
	d := UserDashboard{}
	var err error
	d["by_teacher"], err = dashboardPrincipalByTeacher(ses, schoolId, startDate, endDate)
	if err != nil {
		return nil, err
	}
	d["birthday_users"] = dashboardPrincipalBirthdaysBySchool(ses, schoolId, time.Now())
	return d, nil
}

func dashboardAdminBySchool(ses *utils.Session, schoolIds []string, startDate, endDate time.Time) (interface{}, error) {
	sp, ctx := apm.StartSpan(ses.Context(), "dashboardAdminBySchool", "app")
	ses.SetContext(ctx)
	defer sp.End()
	args := models.SchoolFilterRequest{
		Uids: &schoolIds,
	}
	args.Limit = new(int)
	*args.Limit = 500
	schoolList, _, err := store.Store().SchoolsFindBy(ses.Context(), args)
	if err != nil {
		return nil, err
	}
	err = store.Store().SchoolsLoadRelations(ses.Context(), &schoolList)
	if err != nil {
		return nil, err
	}
	subjectsPercents, err := store.Store().SubjectsPercentsBySchool(ses.Context(), schoolIds, startDate, endDate)
	if err != nil {
		return nil, err
	}
	type bySchoolType struct {
		School         models.SchoolResponse                    `json:"school"`
		SubjectPercent *models.DashboardSubjectsPercentBySchool `json:"subject_percent"`
	}
	bySchool := []bySchoolType{}
	for _, item := range schoolList {
		bySchoolItem := bySchoolType{
			School: models.SchoolResponse{},
		}
		bySchoolItem.School.FromModel(item)
		for k, sp := range subjectsPercents {
			if sp.SchoolID == item.ID {
				bySchoolItem.SubjectPercent = &subjectsPercents[k]
			}
		}
		bySchool = append(bySchool, bySchoolItem)
	}
	return bySchool, nil
}

func dashboardPrincipalBirthdaysBySchool(ses *utils.Session, schoolId string, date time.Time) []models.UserResponse {
	var err error
	sp, ctx := apm.StartSpan(ses.Context(), "dashboardPrincipalBirthdaysBySchool", "app")
	ses.SetContext(ctx)
	defer sp.End()
	res := []models.UserResponse{}
	args := models.UserFilterRequest{
		SchoolId: &schoolId,
		Roles:    &[]string{string(models.RoleTeacher), string(models.RolePrincipal), string(models.RoleStudent)},
	}
	args.Limit = new(int)
	*args.Limit = 100
	args.SchoolId = &schoolId
	args.BirthdayToday = &time.Time{}
	*args.BirthdayToday = time.Now()
	uList, _, _ := store.Store().UsersFindBy(ses.Context(), args)
	err = store.Store().UsersLoadRelationsClassrooms(ses.Context(), &uList)
	if err != nil {
		return nil
	}
	for _, v := range uList {
		resItem := models.UserResponse{}
		resItem.FromModel(v)
		res = append(res, resItem)
	}
	return res
}

func dashboardPrincipalByTeacher(ses *utils.Session, schoolId string, startDate, endDate time.Time) (interface{}, error) {
	sp, ctx := apm.StartSpan(ses.Context(), "dashboardPrincipalByTeacher", "app")
	ses.SetContext(ctx)
	defer sp.End()
	// set by_teacher journals
	args := models.SubjectFilterRequest{
		SchoolId: &schoolId,
	}
	args.Limit = new(int)
	*args.Limit = 2000
	subjectList, _, err := store.Store().SubjectsListFilters(ses.Context(), &args)
	if err != nil {
		return nil, err
	}
	subjectIds := []string{}
	for _, v := range subjectList {
		subjectIds = append(subjectIds, v.ID)
	}
	subjectsPercents, err := store.Store().SubjectsPercents(ses.Context(), subjectIds, startDate, endDate)
	if err != nil {
		return nil, err
	}
	args2 := models.UserFilterRequest{
		SchoolId: &schoolId,
	}
	args2.Limit = new(int)
	*args2.Limit = 500
	args2.Role = new(string)
	*args2.Role = string(models.RoleTeacher)
	teacherList, _, err := store.Store().UsersFindBy(ses.Context(), args2)
	if err != nil {
		return nil, err
	}
	type byTeacherType struct {
		Teacher         models.UserResponse               `json:"teacher"`
		SubjectPercents []models.DashboardSubjectsPercent `json:"subject_percents"`
	}
	byTeacher := []byTeacherType{}
	for _, teacherItem := range teacherList {
		byTeacherItem := byTeacherType{
			Teacher:         models.UserResponse{},
			SubjectPercents: []models.DashboardSubjectsPercent{},
		}
		byTeacherItem.Teacher.FromModel(teacherItem)
		for _, subjectItem := range subjectList {
			if subjectItem.SecondTeacherId != nil && *subjectItem.SecondTeacherId == teacherItem.ID ||
				subjectItem.SecondTeacherId == nil && subjectItem.TeacherId != nil && *subjectItem.TeacherId == teacherItem.ID {
				for _, sp := range subjectsPercents {
					if sp.SubjectId == subjectItem.ID {
						byTeacherItem.SubjectPercents = append(byTeacherItem.SubjectPercents, sp)
					}
				}
			}
		}
		byTeacher = append(byTeacher, byTeacherItem)
	}
	sort.Slice(byTeacher, func(i, j int) bool {
		return len(byTeacher[i].SubjectPercents) > len(byTeacher[j].SubjectPercents)
	})

	return byTeacher, nil
}

func dashboardPrincipalCurrentLesson(ses *utils.Session, schoolId string, date time.Time, d *UserDashboard) error {
	sp, ctx := apm.StartSpan(ses.Context(), "dashboardPrincipalCurrentLesson", "app")
	ses.SetContext(ctx)
	defer sp.End()
	shiftList, _, err := store.Store().ShiftsFindBy(ses.Context(), models.ShiftFilterRequest{
		SchoolId: &schoolId,
	})
	resShifts := []models.ShiftResponse{}
	for _, v := range shiftList {
		item := models.ShiftResponse{}
		item.FromModel(v)
		if len(item.Value) < 1 || len(item.Value[0]) < 1 {
			continue
		}
		resShifts = append(resShifts, item)
	}
	sort.Slice(resShifts, func(i, j int) bool {
		t1 := resShifts[i].Value[0][0][0]
		t2 := resShifts[j].Value[0][0][0]
		if len(t1) < len("09:09") {
			t1 = "0" + t1
		}
		if len(t2) < len("09:09") {
			t2 = "0" + t2
		}
		return t1 < t2
	})
	dateWeekday := int(date.Weekday()) - 1
	if dateWeekday == -1 {
		dateWeekday = 6
	}
	currentLessonNumber := 0
	currentLessonWeekday := 0
	currentLessonTimes := []string{}
	for _, shiftRes := range resShifts {
		currentDate := date

		// today is exact
		if len(shiftRes.Value) > dateWeekday {
			shiftHours := shiftRes.Value[dateWeekday]
			prevTimes := []string{}
			if len(shiftHours) > 0 {
				shiftHours = append([][]string{shiftHours[0]}, shiftHours...)
			}
			for shiftHour, times := range shiftHours {
				if len(times[0]) < len("09:09") {
					times[0] = "0" + times[0]
				}
				if len(times[1]) < len("09:09") {
					times[1] = "0" + times[1]
				}
				if len(prevTimes) > 0 && (currentDate.Format(time.TimeOnly) >= prevTimes[1] && currentDate.Format(time.TimeOnly) <= times[1] || currentDate.Format(time.TimeOnly) <= prevTimes[0]) {
					currentLessonNumber = shiftHour
					currentLessonTimes = times
					currentLessonWeekday = dateWeekday

					break
				}
				prevTimes = times
			}
		}
		// tomorrow exists
		if currentLessonNumber == 0 && dateWeekday != 6 {
			currentDate = currentDate.AddDate(0, 0, 1)
			dateWeekday = int(currentDate.Weekday()) - 1
			if dateWeekday == -1 {
				dateWeekday = 6
			}
			if len(shiftRes.Value) > dateWeekday {
				shiftHours := shiftRes.Value[dateWeekday]
				if len(shiftHours) > 0 {
					currentLessonNumber = 1
					currentLessonTimes = shiftHours[0]
					currentLessonWeekday = dateWeekday
				}
			}
		}
		// tomorrow is next week
		if currentLessonNumber == 0 {
			currentDate = currentDate.AddDate(0, 0, 7)
			currentDate = currentDate.AddDate(0, 0, -7+int(currentDate.Weekday())+1)
			dateWeekday = int(currentDate.Weekday()) - 1
			if dateWeekday == -1 {
				dateWeekday = 6
			}
			if len(shiftRes.Value) > dateWeekday {
				shiftHours := shiftRes.Value[dateWeekday]
				if len(shiftHours) > 0 {
					currentLessonNumber = 1
					currentLessonTimes = shiftHours[0]
					currentLessonWeekday = dateWeekday
				}
			}
		}
		if currentLessonNumber != 0 {
			date = currentDate
			break
		}
	}
	(*d)["current_lesson_date"] = date.Format(time.DateOnly)
	(*d)["current_lesson_number"] = currentLessonNumber
	(*d)["current_lesson_times"] = currentLessonTimes
	// set current_lesson_subjects
	timetablesList, _, err := store.Store().TimetablesFindBy(ses.Context(), models.TimetableFilterRequest{
		SchoolId: &schoolId,
	})
	if err != nil {
		return err
	}
	todaySubjectIds := []string{}
	for _, timetableItem := range timetablesList {
		timetableRes := models.TimetableResponse{}
		timetableRes.FromModel(timetableItem)
		if len(timetableRes.Value) > currentLessonWeekday {
			tValueHours := timetableRes.Value[currentLessonWeekday]
			if len(tValueHours) >= currentLessonNumber && currentLessonNumber > 0 {
				if tValueHours[currentLessonNumber-1] != "" {
					todaySubjectIds = append(todaySubjectIds, tValueHours[currentLessonNumber-1])
				}

			}
		}
	}
	(*d)["current_lesson_subjects"] = len(todaySubjectIds)
	return nil
}
