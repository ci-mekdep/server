package app

import (
	"context"
	"encoding/json"
	"slices"
	"sort"
	"time"

	"github.com/mekdep/server/config"
	"github.com/mekdep/server/internal/api/utils"
	"github.com/mekdep/server/internal/models"
	"github.com/mekdep/server/internal/store"
	apputils "github.com/mekdep/server/internal/utils"
	"go.elastic.co/apm/v2"
)

type DashboardNumbersRequest struct {
	UserId   *string `json:"user_id"`
	SchoolId *string `json:"school_id"`
}

type DashboardDetailsRequest struct {
	StartDate *time.Time `json:"start_date" form:"start_date" time_format:"2006-01-02"`
	EndDate   *time.Time `json:"end_date" form:"end_date" time_format:"2006-01-02"`
	UserId    *string    `json:"user_id" form:"user_id"`
	SchoolId  *string    `json:"school_id" form:"school_id"`
	RegionId  *string    `json:"region_id" form:"region_id"`
}

type DashboardNumbersResponse struct {
	OrganizationsCount   *int    `json:"organizations_count"`
	SchoolsCount         *int    `json:"schools_count"`
	PrincipalsCount      *int    `json:"principals_count"`
	TeachersCount        *int    `json:"teachers_count"`
	ClassroomsCount      *int    `json:"classrooms_count"` // todo: classroom_count also
	StudentsCount        *int    `json:"students_count"`
	ParentsCount         *int    `json:"parents_count"`
	SubjectHoursSum      *uint   `json:"subject_hours_sum"`
	TeacherClassroomName *string `json:"teacher_classroom_name"`
}

type DashboardDetailsResponse struct {
	ReportByTeacher []ReportByTeacherType `json:"report_by_teacher"`
	ReportBySchool  []ReportBySchoolType  `json:"report_by_school"`

	CurrentLessonDate     *string   `json:"current_lesson_date"`
	CurrentLessonNumber   *int      `json:"current_lesson_number"`
	CurrentLessonTimes    *[]string `json:"current_lesson_times"`
	CurrentLessonSubjects *int      `json:"current_lesson_subjects"`

	BirthdayUsers *[]models.UserResponse `json:"users_birthday"`
}

func (a App) DashboardNumbersV2Cached(ses *utils.Session, req DashboardNumbersRequest) (*DashboardNumbersResponse, error) {
	reqStr, _ := json.Marshal(req)
	cacheKey := *ses.GetSessionId() + "DashboardNumbersV2Cached" + string(reqStr)
	res := DashboardNumbersResponse{}
	var err error
	if val, found := a.cache.Get(cacheKey); found {
		res = val.(DashboardNumbersResponse)
	} else {
		resPtr, err := a.DashboardNumbersV2(ses, req)
		if err != nil {
			return nil, err
		}
		a.cache.Set(cacheKey, *resPtr, time.Minute*60)
		res = *resPtr
	}
	return &res, err
}

func (a App) DashboardDetailsV2Cached(ses *utils.Session, req DashboardDetailsRequest) (*DashboardDetailsResponse, error) {
	reqStr, _ := json.Marshal(req)
	cacheKey := *ses.GetSessionId() + "DashboardDetailsV2Cached" + string(reqStr)
	res := DashboardDetailsResponse{}
	var err error
	if val, found := a.cache.Get(cacheKey); found {
		resStr := val.(string)
		err = json.Unmarshal([]byte(resStr), &res)
	} else {
		resPtr, err := a.DashboardDetailsV2(ses, req)
		if err != nil {
			return nil, err
		}
		resStr := []byte{}
		resStr, err = json.Marshal(*resPtr)
		res = *resPtr
		a.cache.Set(cacheKey, string(resStr), time.Minute*10)
	}
	return &res, err
}

func (a App) DashboardNumbersV2(ses *utils.Session, req DashboardNumbersRequest) (*DashboardNumbersResponse, error) {
	sp, ctx := apm.StartSpan(ses.Context(), "DashboardNumbers", "app")
	ses.SetContext(ctx)
	defer sp.End()
	roleCode := ses.GetRole()
	req.UserId = &ses.GetUser().ID
	req.SchoolId = ses.GetSchoolId()

	var err error
	res := DashboardNumbersResponse{}

	if slices.Contains([]models.Role{
		models.RoleAdmin,
		models.RoleOrganization,
		models.RolePrincipal}, *roleCode) {
		err = dashboardNumbersAdmin(ses, req, &res)
	} else if models.RoleTeacher == *roleCode {
		err = dashboardNumbersTeacher(ses, req, &res)
	}
	if err != nil {
		return nil, err
	}
	return &res, nil
}

func (a App) DashboardDetailsV2(ses *utils.Session, req DashboardDetailsRequest) (*DashboardDetailsResponse, error) {
	sp, ctx := apm.StartSpan(ses.Context(), "DashboardNumbers", "app")
	ses.SetContext(ctx)
	defer sp.End()
	isDetail := false
	if req.RegionId != nil {
		req.SchoolId = req.RegionId
	}
	if req.SchoolId != nil && slices.Contains(ses.SchoolAllIds(), *req.SchoolId) {
		isDetail = true
	} else {
		req.SchoolId = ses.GetSchoolId()
	}
	roleCode := ses.GetRole()
	req.UserId = &ses.GetUser().ID

	var err error
	res := DashboardDetailsResponse{}

	if slices.Contains([]models.Role{
		models.RoleAdmin, models.RoleOrganization}, *roleCode) {
		if isDetail {
			err = dashboardDetailsPrincipal(ses, req, &res)
		} else {
			req.SchoolId = nil
			req.RegionId = nil
			err = dashboardDetailsAdmin(ses, req, &res)
		}
	} else if models.RolePrincipal == *roleCode {
		err = dashboardDetailsPrincipal(ses, req, &res)
	} else if models.RoleTeacher == *roleCode {
		err = dashboardDetailsTeacher(ses, req, &res)
	}
	if err != nil {
		return nil, err
	}
	return &res, nil
}

func dashboardNumbersAdmin(ses *utils.Session, req DashboardNumbersRequest, res *DashboardNumbersResponse) error {
	sp, ctx := apm.StartSpan(ses.Context(), "dashboardNumbersTeacher", "app")
	ses.SetContext(ctx)
	defer sp.End()
	schoolIds := ses.GetSchoolsByAdminRoles()

	usersCount, err := store.Store().UsersLoadCount(ses.Context(), schoolIds)
	if err != nil {
		return err
	}
	res.StudentsCount = &usersCount.StudentsCount
	res.ParentsCount = &usersCount.ParentsCount
	res.TeachersCount = &usersCount.TeachersCount
	res.PrincipalsCount = &usersCount.PrincipalsCount
	res.ClassroomsCount = &usersCount.ClassroomsCount
	res.SchoolsCount = &usersCount.SchoolsCount
	res.OrganizationsCount = &usersCount.OrganizationsCount
	return nil
}

func dashboardNumbersTeacher(ses *utils.Session, req DashboardNumbersRequest, res *DashboardNumbersResponse) error {
	sp, ctx := apm.StartSpan(ses.Context(), "dashboardNumbersTeacher", "app")
	ses.SetContext(ctx)
	defer sp.End()
	args := models.SubjectFilterRequest{
		TeacherIds: []string{*req.UserId},
	}
	args.Limit = new(int)
	*args.Limit = 100
	subjects, _, err := store.Store().SubjectsListFilters(context.Background(), &args)
	if err != nil {
		return err
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
	if ses.GetUser().TeacherClassroom != nil {
		classroom, err := store.Store().ClassroomsFindById(ses.Context(), ses.GetUser().TeacherClassroom.ID)
		if err != nil {
			return err
		}
		res.TeacherClassroomName = classroom.Name
		res.StudentsCount = classroom.StudentsCount
	}
	res.SubjectHoursSum = &subjectHours
	res.ClassroomsCount = new(int)
	*res.ClassroomsCount = len(classroomIds)
	return nil
}

type ReportByTeacherType struct {
	Teacher         models.UserResponse               `json:"teacher"`
	SubjectPercents []models.DashboardSubjectsPercent `json:"subject_percents"`
}

type ReportBySchoolType struct {
	School         models.SchoolResponse                    `json:"school"`
	SubjectPercent *models.DashboardSubjectsPercentBySchool `json:"subject_percent"`
}

func dashboardDetailsTeacher(ses *utils.Session, req DashboardDetailsRequest, res *DashboardDetailsResponse) error {
	sp, ctx := apm.StartSpan(ses.Context(), "dashboardDetailsTeacher", "app")
	ses.SetContext(ctx)
	defer sp.End()
	var err error

	res.BirthdayUsers = &[]models.UserResponse{}
	*res.BirthdayUsers = dashboardDetailsBirthdays(ses, *req.SchoolId, time.Now())
	err = dashboardDetailsCurrentLesson(ses, req, res)
	if err != nil {
		return err
	}
	return nil
}

func dashboardDetailsPrincipal(ses *utils.Session, req DashboardDetailsRequest, res *DashboardDetailsResponse) error {
	sp, ctx := apm.StartSpan(ses.Context(), "dashboardDetailsPrincipal", "app")
	ses.SetContext(ctx)
	defer sp.End()
	var err error

	// set by_teacher journals
	args := models.SubjectFilterRequest{
		SchoolId: req.SchoolId,
	}
	args.Limit = new(int)
	*args.Limit = 5000
	subjectList, _, err := store.Store().SubjectsListFilters(ses.Context(), &args)
	if err != nil {
		return err
	}
	subjectIds := []string{}
	for _, v := range subjectList {
		subjectIds = append(subjectIds, v.ID)
	}
	subjectsPercents, err := store.Store().SubjectsPercents(ses.Context(), subjectIds, *req.StartDate, *req.EndDate)
	if err != nil {
		return err
	}
	args2 := models.UserFilterRequest{
		SchoolId: req.SchoolId,
	}
	args2.Limit = new(int)
	*args2.Limit = 2000
	args2.Role = new(string)
	*args2.Role = string(models.RoleTeacher)
	teacherList, _, err := store.Store().UsersFindBy(ses.Context(), args2)
	if err != nil {
		return err
	}
	byTeacher := []ReportByTeacherType{}
	for _, teacherItem := range teacherList {
		byTeacherItem := ReportByTeacherType{
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
	res.ReportByTeacher = byTeacher

	if err != nil {
		return err
	}
	res.BirthdayUsers = &[]models.UserResponse{}
	*res.BirthdayUsers = dashboardDetailsBirthdays(ses, *req.SchoolId, time.Now())
	err = dashboardDetailsCurrentLesson(ses, req, res)
	if err != nil {
		return err
	}
	return nil
}

func dashboardDetailsAdmin(ses *utils.Session, req DashboardDetailsRequest, res *DashboardDetailsResponse) error {
	sp, ctx := apm.StartSpan(ses.Context(), "dashboardDetailsPrincipal", "app")
	ses.SetContext(ctx)
	defer sp.End()
	var err error

	// set by_teacher journals
	args := models.SchoolFilterRequest{}
	if req.SchoolId != nil {
		args.Uids = &[]string{*req.SchoolId}
	} else if req.RegionId != nil {
		args.ParentUid = req.RegionId
	} else {
		args.Uids = &[]string{}
		*args.Uids = ses.SchoolAllIds()
	}
	args.Limit = new(int)
	*args.Limit = 500
	args.IsParent = new(bool)
	*args.IsParent = false
	schoolList, _, err := store.Store().SchoolsFindBy(ses.Context(), args)
	if err != nil {
		return err
	}
	schoolIds := &[]string{}
	for _, v := range schoolList {
		*schoolIds = append(*schoolIds, v.ID)
	}
	err = store.Store().SchoolsLoadRelations(ses.Context(), &schoolList)
	if err != nil {
		return err
	}
	subjectsPercents, err := store.Store().SubjectsPercentsBySchool(ses.Context(), *schoolIds, *req.StartDate, *req.EndDate)
	if err != nil {
		return err
	}
	bySchool := []ReportBySchoolType{}
	for _, item := range schoolList {
		bySchoolItem := ReportBySchoolType{
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
	res.ReportBySchool = bySchool
	return nil
}

func dashboardDetailsBirthdays(ses *utils.Session, schoolId string, date time.Time) []models.UserResponse {
	var err error
	sp, ctx := apm.StartSpan(ses.Context(), "dashboardPrincipalBirthdaysBySchool", "app")
	ses.SetContext(ctx)
	defer sp.End()
	res := []models.UserResponse{}
	args := models.UserFilterRequest{
		Roles: &[]string{string(models.RoleTeacher), string(models.RolePrincipal), string(models.RoleStudent)},
	}
	args.Limit = new(int)
	*args.Limit = 100
	args.SchoolId = &schoolId
	args.BirthdayToday = &date
	if *ses.GetRole() == models.RoleTeacher && ses.GetUser().TeacherClassroom != nil {
		args.ClassroomIdForBirthday = &ses.GetUser().TeacherClassroom.ID
	}
	uList, _, _ := store.Store().UsersFindBy(ses.Context(), args)
	if err != nil {
		apputils.LoggerDesc("dashboardDetailsBirthdays").Error(err)
		return nil
	}
	err = store.Store().UsersLoadRelationsClassrooms(ses.Context(), &uList)
	if err != nil {
		apputils.LoggerDesc("dashboardDetailsBirthdays").Error(err)
		return nil
	}
	for _, v := range uList {
		resItem := models.UserResponse{}
		resItem.FromModel(v)
		res = append(res, resItem)
	}
	return res
}

func dashboardDetailsCurrentLesson(ses *utils.Session, req DashboardDetailsRequest, res *DashboardDetailsResponse) error {
	sp, ctx := apm.StartSpan(ses.Context(), "dashboardDetailsCurrentLesson", "app")
	ses.SetContext(ctx)
	defer sp.End()
	shiftList, _, err := store.Store().ShiftsFindBy(ses.Context(), models.ShiftFilterRequest{
		SchoolId: req.SchoolId,
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
		if len(resShifts[i].Value) < 1 || len(resShifts[i].Value[0]) < 1 || len(resShifts[i].Value[0][0]) < 1 {
			return false
		}
		if len(resShifts[j].Value) < 1 || len(resShifts[j].Value[0]) < 1 || len(resShifts[j].Value[0][0]) < 1 {
			return false
		}
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
	date := *req.StartDate
	if req.StartDate.Format(time.DateOnly) == time.Now().Format(time.DateOnly) {
		date = time.Now().In(config.RequestLocation)
	}
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
				if len(times) < 1 {
					continue
				}
				if len(times[0]) < len("09:09") {
					times[0] = "0" + times[0]
				}
				if len(times[1]) < len("09:09") {
					times[1] = "0" + times[1]
				}
				if len(prevTimes) > 0 &&
					(currentDate.Format(time.TimeOnly) >= prevTimes[1] &&
						currentDate.Format(time.TimeOnly) <= times[1] ||
						currentDate.Format(time.TimeOnly) <= prevTimes[0]) {
					currentLessonNumber = shiftHour
					currentLessonTimes = times
					currentLessonWeekday = dateWeekday
					break
				}
				prevTimes = times
			}
		}
	}
	for _, shiftRes := range resShifts {
		currentDate := date

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
		}
	}
	res.CurrentLessonDate = new(string)
	*res.CurrentLessonDate = date.Format(time.DateOnly)
	res.CurrentLessonNumber = &currentLessonNumber
	res.CurrentLessonTimes = &currentLessonTimes
	// set current_lesson_subjects
	timetablesList, _, err := store.Store().TimetablesFindBy(ses.Context(), models.TimetableFilterRequest{
		SchoolId: req.SchoolId,
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
	res.CurrentLessonSubjects = new(int)
	*res.CurrentLessonSubjects = len(todaySubjectIds)
	return nil
}
