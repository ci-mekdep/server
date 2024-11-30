package app

import (
	"context"
	"time"

	"github.com/mekdep/server/internal/api/utils"
	"github.com/mekdep/server/internal/models"
	"github.com/mekdep/server/internal/store"
	"go.elastic.co/apm/v2"
)

func UserCacheClear(ses *utils.Session, parentId string) {
	cKey := "user_children_" + parentId
	Ap().cache.Set(cKey, nil, 0)

}

func UserChildrenGet(ses *utils.Session, user *models.User) ([]*models.User, error) {
	cKey := "user_children_" + user.ID
	var err error
	if val, found := Ap().cache.Get(cKey); found && val != nil {
		user.Children = val.([]*models.User)
	} else {
		user.Children = nil
		err = store.Store().UsersLoadRelationsChildren(ses.Context(), &[]*models.User{user})
		if err != nil {
			return nil, err
		}
		Ap().cache.Set(cKey, user.Children, 0)
	}
	return user.Children, nil
}

func (a App) ParentChildren(ses *utils.Session, schoolId *string, user *models.User) ([]models.UserResponse, error) {
	sp, ctx := apm.StartSpan(ses.Context(), "ParentChildren", "app")
	ses.SetContext(ctx)
	defer sp.End()
	var err error
	// get children

	user.Children, err = UserChildrenGet(ses, user)
	if err != nil {
		return nil, err
	}
	users := user.Children
	// if student
	if *ses.GetRole() == models.RoleStudent {
		user.Classrooms = nil
		users = append(users, user)
	}
	userIds := []string{}
	for _, v := range users {
		// users[k].Classrooms = nil
		userIds = append(userIds, v.ID)
	}

	// get classrooms and schools
	err = store.Store().UsersLoadRelationsClassrooms(ses.Context(), &users)
	if err != nil {
		return nil, err
	}
	classrooms := []*models.Classroom{}
	for _, v := range users {
		for _, cc := range v.Classrooms {
			classrooms = append(classrooms, cc.Classroom)
		}
	}
	err = store.Store().ClassroomsLoadSchool(ses.Context(), &classrooms)
	if err != nil {
		return nil, err
	}

	res := []models.UserResponse{}
	for _, v := range users {
		resItem := models.UserResponse{}
		resItem.FromModel(v)
		resClassrooms := []*models.ClassroomResponse{}
		for _, cc := range v.Classrooms {
			resClass := models.ClassroomResponse{}
			resClass.FromModel(cc.Classroom)
			resClass.TariffEndAt = cc.TariffEndAt
			resClass.TariffType = cc.TariffType
			if cc.TariffEndAt != nil && cc.TariffEndAt.Before(time.Now()) || cc.TariffEndAt == nil {
				resClass.TariffType = nil
				resClass.TariffEndAt = nil
			}
			school, err := store.Store().SchoolsFindById(ses.Context(), cc.Classroom.SchoolId)
			if err != nil {
				return nil, err
			}
			resClass.School = &models.SchoolResponse{}
			resClass.School.FromModel(school)

			resClassrooms = append(resClassrooms, &resClass)
		}
		resItem.Classrooms = resClassrooms
		res = append(res, resItem)
	}

	return res, nil
}

func studentClassroom(student *models.User) (string, string, error) {
	if len(student.Classrooms) < 1 {
		return "", "", ErrRequired.SetKey("classroom_id")
	}
	classroomId := student.Classrooms[0].ClassroomId
	schoolId := ""
	for _, v := range student.Schools {
		if v.SchoolUid != nil {
			schoolId = *v.SchoolUid
			break
		}
	}
	return classroomId, schoolId, nil
}

func (a App) GetAllGradesForSubject(ses *utils.Session, student *models.User, subjectID string, periodNumber int) ([]models.DiaryLessonResponse, error) {
	sp, ctx := apm.StartSpan(ses.Context(), "GetAllGradesForSubject", "app")
	ses.SetContext(ctx)
	defer sp.End()
	classroomId, _, err := studentClassroom(student)
	if err != nil {
		return nil, err
	}
	argL := models.LessonFilterRequest{
		ClassroomId:  &classroomId,
		SubjectId:    &subjectID,
		PeriodNumber: &periodNumber,
		SchoolId:     ses.GetSchoolId(),
	}
	argL.Limit = new(int)
	*argL.Limit = 500
	lessonList, _, err := store.Store().LessonsFindBy(ses.Context(), argL)
	if err != nil {
		return nil, err
	}
	lessonIds := make([]string, len(lessonList))
	for i, lesson := range lessonList {
		lessonIds[i] = lesson.ID
	}

	argG := models.GradeFilterRequest{
		LessonIds: &lessonIds,
		StudentId: &student.ID,
	}
	argG.Limit = new(int)
	*argG.Limit = 1000

	gradesList, _, err := store.Store().GradesFindBy(ses.Context(), argG)
	if err != nil {
		return nil, err
	}

	argA := models.AbsentFilterRequest{
		LessonIds: &lessonIds,
		StudentId: &student.ID,
	}
	argA.Limit = new(int)
	*argA.Limit = 1000
	absentList, _, err := store.Store().AbsentsFindBy(ses.Context(), argA)
	if err != nil {
		return nil, err
	}

	argAs := models.AssignmentFilterRequest{
		LessonIDs: &lessonIds,
	}
	argAs.Limit = new(int)
	*argAs.Limit = 1000
	assignmentList, _, err := store.Store().AssignmentsFindBy(ses.Context(), argAs)
	if err != nil {
		return nil, err
	}

	subject, err := store.Store().SubjectsFindById(ses.Context(), lessonList[0].SubjectId)
	if err != nil {
		return nil, err
	}

	var res []models.DiaryLessonResponse
	for _, lesson := range lessonList {
		response := models.DiaryLessonResponse{}
		var grade *models.Grade
		for _, g := range gradesList {
			if g.LessonId == lesson.ID {
				grade = g
				break
			}
		}

		var absent *models.Absent
		for _, a := range absentList {
			if a.LessonId == lesson.ID {
				absent = a
				break
			}
		}
		if grade == nil && absent == nil {
			continue
		}
		var assignment *models.Assignment
		for _, as := range assignmentList {
			if as.LessonId == lesson.ID {
				assignment = &as
				break
			}
		}

		response.Lesson = &models.LessonResponse{}
		response.Lesson.FromModel(lesson)
		if grade != nil {
			response.Grade = &models.GradeResponse{}
			response.Grade.FromModel(grade)
		}
		if absent != nil {
			response.Absent = &models.AbsentResponse{}
			response.Absent.FromModel(absent)
		}
		if assignment != nil {
			response.Assignment = &models.AssignmentResponse{}
			response.Assignment.FromModel(assignment)
		}

		response.Subject = &models.SubjectResponse{}
		response.Subject.FromModel(subject)

		res = append(res, response)
	}
	return res, nil
}

func (a App) StudentDiary(ses *utils.Session, student models.User, classroom models.Classroom, date time.Time, isReview bool) (*models.DiaryResponse, error) {
	sp, ctx := apm.StartSpan(ses.Context(), "StudentDiary", "app")
	ses.SetContext(ctx)
	defer sp.End()
	var err error

	// set date
	date = date.AddDate(0, 0, -int(date.Weekday()))
	// get subjects
	args := models.SubjectFilterRequest{
		ClassroomIds: []string{classroom.ID},
	}
	args.Limit = new(int)
	*args.Limit = 100
	subjectList, _, err := store.Store().SubjectsListFilters(ses.Context(), &args)
	if err != nil {
		return nil, err
	}
	teacherIds := []string{}
	for _, v := range subjectList {
		if v.TeacherId != nil {
			teacherIds = append(teacherIds, *v.TeacherId)
		}
	}
	teacherList, err := store.Store().UsersFindByIds(ses.Context(), teacherIds)
	if err != nil {
		return nil, err
	}
	for k, v := range subjectList {
		for _, vv := range teacherList {
			if v.TeacherId != nil && vv.ID == *v.TeacherId {
				subjectList[k].Teacher = vv
			}
		}
	}

	// get timetables
	argsT := models.TimetableFilterRequest{
		ClassroomIds: &[]string{classroom.ID},
	}
	argsT.Limit = new(int)
	*argsT.Limit = 10
	timetables, _, err := store.Store().TimetablesFindBy(ses.Context(), argsT)
	if len(timetables) < 1 {
		return nil, ErrNotExists.SetKey("timetable_id").SetComment("Classroom does not have any timetable")
	}
	timetableModel := timetables[0]
	tRes := models.TimetableResponse{}
	tRes.FromModel(timetableModel)

	// get shifts
	if timetableModel == nil || timetableModel.ShiftId == nil {
		return nil, ErrNotExists.SetKey("timetable_id").SetComment("Classroom does not have any timetable")
	}
	shiftModel, err := store.Store().ShiftsFindById(ses.Context(), *timetableModel.ShiftId)
	if err != nil {
		return nil, err
	}
	if shiftModel == nil {
		return nil, ErrNotExists.SetKey("shift_id").SetComment("Timetable does not have any shift")
	}
	resShift := models.ShiftResponse{}
	resShift.FromModel(shiftModel)
	shiftValue := resShift.Value

	// get lessons with grades, absents
	argL := models.LessonFilterRequest{
		ClassroomId: &classroom.ID,
		DateRange:   &[]string{date.Format(time.DateOnly), date.AddDate(0, 0, 7).Format(time.DateOnly)},
	}
	argL.Limit = new(int)
	*argL.Limit = 1000
	lessonList, _, err := store.Store().LessonsFindBy(ses.Context(), argL)
	if err != nil {
		return nil, err
	}
	lessonList = lessonsSort(lessonList)
	lessonIds := []string{}
	for _, v := range lessonList {
		lessonIds = append(lessonIds, v.ID)
	}
	err = store.Store().LessonsLoadRelations(ses.Context(), &lessonList)
	if err != nil {
		return nil, err
	}
	argG := models.GradeFilterRequest{
		LessonIds: &lessonIds,
		StudentId: &student.ID,
	}
	argG.Limit = new(int)
	*argG.Limit = 1000
	gradesList, _, err := store.Store().GradesFindBy(ses.Context(), argG)
	if err != nil {
		return nil, err
	}

	argA := models.AbsentFilterRequest{
		LessonIds: &lessonIds,
		StudentId: &student.ID,
	}
	argA.Limit = new(int)
	*argA.Limit = 1000
	absentsList, _, err := store.Store().AbsentsFindBy(ses.Context(), argA)
	if err != nil {
		return nil, err
	}

	// form diary
	res := models.DiaryResponse{}
	res.Days = []models.DiaryDayResponse{}
	for weekday := 1; weekday <= 6; weekday++ {
		date = date.AddDate(0, 0, 1)
		resDay := models.DiaryDayResponse{}
		resDay.Hours = []models.DiaryLessonResponse{}

		// set holiday
		if h := GetHolidayByDate(date); h != "" {
			resDay.Holiday = &h
		} else if is, _ := isDateVacationBySchool(ses, date, classroom.SchoolId); is {
			resDay.Holiday = new(string)
			*resDay.Holiday = "Dynç alyş"
		} else {
			resDay.Holiday = nil
		}
		resDay.Date = date.Format(time.DateOnly)

		if resDay.Holiday == nil {
			hour := -1
			for _, lesson := range lessonList {
				if lesson.Date.Format(time.DateOnly) == date.Format(time.DateOnly) {
					hour++
					var grade *models.Grade
					var absent *models.Absent
					for _, g := range gradesList {
						if g.LessonId == lesson.ID {
							grade = g
						}
					}
					for _, a := range absentsList {
						if a.LessonId == lesson.ID {
							absent = a
						}
					}
					for _, s := range subjectList {
						if s.ID == lesson.SubjectId {
							lesson.Subject = s
						}
					}

					resHour := models.DiaryLessonResponse{}
					// set shift
					if len(shiftValue) >= weekday {
						if len(shiftValue[weekday-1]) > hour {
							resHour.ShiftTimes = shiftValue[weekday-1][hour]
						}
					}

					// form diary hour
					if lesson.Subject != nil {
						resHour.Subject = &models.SubjectResponse{}
						resHour.Subject.FromModel(lesson.Subject)
						if lesson != nil {
							resHour.Lesson = &models.LessonResponse{}
							resHour.Lesson.FromModel(lesson)
						}
						if grade != nil {
							resHour.Grade = &models.GradeResponse{}
							resHour.Grade.FromModel(grade)
						}
						if absent != nil {
							resHour.Absent = &models.AbsentResponse{}
							resHour.Absent.FromModel(absent)
						}
					}
					resDay.Hours = append(resDay.Hours, resHour)
				}
			}
		}
		res.Days = append(res.Days, resDay)
	}
	// set lastReviewed
	res.LastReviewedAt, err = handleParentReview(ctx, isReview, gradesList, absentsList)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

func handleParentReview(ctx context.Context, isReview bool, gradesList []*models.Grade, absentsList []*models.Absent) (*time.Time, error) {
	var err error
	now := time.Now()
	if isReview {
		updateGrades := []*models.Grade{}
		updateAbsents := []*models.Absent{}
		for _, v := range gradesList {
			if v.ParentReviewedAt == nil {
				v.ParentReviewedAt = &now
				updateGrades = append(updateGrades, v)
			}
		}
		for _, v := range absentsList {
			if v.ParentReviewedAt == nil {
				v.ParentReviewedAt = &now
				updateAbsents = append(updateAbsents, v)
			}
		}
		// TODO: make batch
		if len(updateGrades) > 0 {
			for _, v := range updateGrades {
				_, err = store.Store().GradesUpdate(ctx, *v)
			}
		}
		if len(updateAbsents) > 0 {
			for _, v := range updateAbsents {
				_, err = store.Store().AbsentsUpdate(ctx, *v)
			}
		}
		if err != nil {
			return nil, err
		}
	}

	if len(gradesList) < 1 && len(absentsList) < 1 {
		return &now, nil
	}
	var lastReviewedAt *time.Time
	for _, v := range gradesList {
		if v.ParentReviewedAt == nil {
			return nil, nil
		}
		lastReviewedAt = v.ParentReviewedAt
	}
	for _, v := range absentsList {
		if v.ParentReviewedAt == nil {
			return nil, nil
		}
		lastReviewedAt = v.ParentReviewedAt
	}
	return lastReviewedAt, nil
}

func (a App) StudentSubjects(ses *utils.Session, student *models.User, classroomId string, date time.Time) (*models.StudentSubjectResponse, error) {
	sp, ctx := apm.StartSpan(ses.Context(), "StudentSubjects", "app")
	ses.SetContext(ctx)
	defer sp.End()
	var err error
	var schoolId string
	_, schoolId, err = studentClassroom(student)
	if err != nil {
		return nil, err
	}

	// get subjects
	args := models.SubjectFilterRequest{
		ClassroomIds: []string{classroomId},
	}
	args.Limit = new(int)
	*args.Limit = 100
	subjectList, _, err := store.Store().SubjectsListFilters(ses.Context(), &args)
	if err != nil {
		return nil, err
	}
	// assign teachers
	teacherIds := []string{}
	for _, v := range subjectList {
		if v.TeacherId != nil {
			teacherIds = append(teacherIds, *v.TeacherId)
		}
	}
	teacherList, err := store.Store().UsersFindByIds(ses.Context(), teacherIds)
	if err != nil {
		return nil, err
	}
	for k, v := range subjectList {
		for _, vv := range teacherList {
			if v.TeacherId != nil && vv.ID == *v.TeacherId {
				subjectList[k].Teacher = vv
			}
		}
	}

	// get period grades
	periodModel, _, err := periodsGetByDate(ses, time.Now(), schoolId)
	if err != nil {
		return nil, err
	}
	periodGrades, _, err := store.Store().PeriodGradesFindBy(ses.Context(), models.PeriodGradeFilterRequest{
		StudentIds: &[]string{student.ID},
	})
	if err != nil {
		return nil, err
	}
	// TODO: where period number + sql join
	absents, _, err := store.Store().AbsentsFindBy(ses.Context(), models.AbsentFilterRequest{StudentId: &student.ID})
	if err != nil {
		return nil, err
	}

	// response
	res := models.StudentSubjectResponse{
		PeriodCounts: []models.StudentPeriodCount{},
		Items:        []models.StudentSubjectResponseItem{},
	}

	// period grades
	for _, subjectItem := range subjectList {
		// set subject
		resItem := models.StudentSubjectResponseItem{
			Subject: &models.SubjectResponse{},
		}
		resItem.Subject.FromModel(subjectItem)
		// set period grades
		resItem.PeriodGrades = map[int]models.PeriodGradeResponse{}
		finalGrade := models.PeriodGrade{}
		for _, v := range periodGrades {
			if subjectItem.ID == *v.SubjectId {
				if v.PeriodKey == models.PeriodGradeExamKey {
					resItem.ExamGrade = &models.PeriodGradeResponse{}
					resItem.ExamGrade.FromModel(v)
					finalGrade.AppendPowerGrade(v)
				} else {
					pg := models.PeriodGradeResponse{}
					pg.FromModel(v)
					pg.SetValueByRules()
					resItem.PeriodGrades[v.PeriodKey] = pg
					finalGrade.AppendGrade(v)
				}
			}
		}
		resItem.FinalGrade = &models.PeriodGradeResponse{}
		resItem.FinalGrade.FromModel(&finalGrade)
		res.Items = append(res.Items, resItem)
	}
	// period count
	for _, v := range periodModel.GetPeriodKeys() {
		resItem := models.StudentPeriodCount{
			PeriodKey: v,
		}
		for _, vv := range res.Items {
			for _, vvv := range vv.PeriodGrades {
				if vvv.PeriodKey == v {
					resItem.LessonsCount += vvv.LessonCount
					resItem.AbsentsYesCount += vvv.AbsentCount
					// TODO: add other yes, ill types (need to create new store query)
				}
			}
		}
		for _, vv := range absents {
			// TODO: grade reason constant
			if vv.StudentId == student.ID && vv.Reason != nil && *vv.Reason == "no" {
				resItem.AbsentsNoCount++
			}
		}
		res.PeriodCounts = append(res.PeriodCounts, resItem)
	}

	return &res, nil
}
