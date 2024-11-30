package app

import (
	"context"
	"sort"
	"time"

	"github.com/mekdep/server/internal/api/utils"
	"github.com/mekdep/server/internal/models"
	"github.com/mekdep/server/internal/store"
	"go.elastic.co/apm/v2"
)

func (a App) ParentAnalyticsWeekly(ses *utils.Session, student *models.User, classroomId string, date time.Time, p Permission) (*models.ParentAnalyticsWeekly, error) {
	sp, ctx := apm.StartSpan(ses.Context(), "ParentAnalyticsWeekly", "app")
	ses.SetContext(ctx)
	defer sp.End()
	var err error
	if err := a.CheckPayment(student, p); err != nil {
		return nil, err
	}
	if len(student.Classrooms) < 1 {
		return nil, ErrNotSet.SetKey("classroom_id")
	}
	var schoolId string
	_, schoolId, err = studentClassroom(student)
	if err != nil {
		return nil, err
	}
	date = date.AddDate(0, 0, -1)
	startDate := date.AddDate(0, 0, -int(date.Weekday()))
	endDate := startDate.AddDate(0, 0, 6)
	period, periodNumber, err := periodsGetByDate(ses, date, schoolId)
	if err != nil {
		return nil, err
	}
	periodStart, periodEnd, err := periodsGetByKey(*period, periodNumber)
	if err != nil {
		return nil, err
	}

	res := models.ParentAnalyticsWeekly{}
	res.SubjectPercentByAreas, err = analyticsSubjectPercent(student, classroomId, startDate, endDate)
	if err != nil {
		return nil, err
	}
	res.SubjectRating, err = analyticsRating(ses, student, classroomId, schoolId, periodNumber, periodStart, periodEnd)
	if err != nil {
		return nil, err
	}
	res.GradeStreak, err = analyticsGradeStreak(student, classroomId)
	if err != nil {
		return nil, err
	}

	return &res, nil
}

func (a App) ParentAnalyticsSummary(ses *utils.Session, student *models.User, classroomId string, date time.Time, p Permission) (*models.ParentAnalyticsSummary, error) {
	sp, ctx := apm.StartSpan(ses.Context(), "ParentAnalyticsSummary", "app")
	ses.SetContext(ctx)
	defer sp.End()
	var err error
	if err := a.CheckPayment(student, p); err != nil {
		return nil, err
	}
	if len(student.Classrooms) < 1 {
		return nil, ErrNotSet.SetKey("classroom_id")
	}
	var schoolId string
	_, schoolId, err = studentClassroom(student)
	if err != nil {
		return nil, err
	}
	date = date.AddDate(0, 0, -1)
	startDate := date.AddDate(0, 0, -int(date.Weekday())+1)
	endDate := startDate.AddDate(0, 0, 6)
	period, periodNumber, err := periodsGetByDate(ses, date, schoolId)
	if err != nil {
		return nil, err
	}
	periodStart, periodEnd, err := periodsGetByKey(*period, periodNumber)
	if err != nil {
		return nil, err
	}

	res := models.ParentAnalyticsSummary{}
	res.SubjectRating, err = analyticsRating(ses, student, classroomId, schoolId, periodNumber, periodStart, periodEnd)
	if err != nil {
		return nil, err
	}
	res.SubjectGrades, err = store.Store().SubjectGrades(ses.Context(), student.ID, startDate, endDate)
	if err != nil {
		return nil, err
	}
	res.GradeStreak, err = analyticsGradeStreak(student, classroomId)
	if err != nil {
		return nil, err
	}

	return &res, nil
}

func (a App) ParentAnalyticsGrades(ses *utils.Session, child *models.User, subjectId string, classroomId string, periodNumber, limit *int, endDate *time.Time, showOnlyAbsent, showOnlyGrade bool, p Permission) ([]models.GradeDetail, error) {
	sp, ctx := apm.StartSpan(ses.Context(), "ParentAnalyticsGrades", "app")
	ses.SetContext(ctx)
	defer sp.End()
	var err error
	if len(child.Classrooms) < 1 {
		return nil, ErrNotSet.SetKey("classroom_id")
	}

	argL := models.LessonFilterRequest{
		ClassroomId:  &classroomId,
		SubjectId:    &subjectId,
		PeriodNumber: periodNumber,
	}
	argL.Limit = new(int)
	*argL.Limit = 500
	if limit != nil {
		argL.Limit = limit
	}
	lessonList, _, err := store.Store().LessonsFindBy(ses.Context(), argL)
	if err != nil {
		return nil, err
	}

	if endDate != nil {
		nextDay := endDate.Add(24 * time.Hour)
		filteredLessons := []*models.Lesson{}
		for _, lesson := range lessonList {
			if lesson.Date.Before(nextDay) {
				filteredLessons = append(filteredLessons, lesson)
			}
		}
		lessonList = filteredLessons
	}

	lessonIds := []string{}
	for _, v := range lessonList {
		lessonIds = []string{v.ID}
	}

	var gradesList []*models.Grade
	if !showOnlyAbsent {
		argG := models.GradeFilterRequest{
			LessonIds: &lessonIds,
			StudentId: &child.ID,
		}
		argG.Limit = new(int)
		*argG.Limit = 500
		if limit != nil {
			argG.Limit = limit
		}
		gradesList, _, err = store.Store().GradesFindBy(ses.Context(), argG)
		if err != nil {
			return nil, err
		}
	}

	var absentsList []*models.Absent
	if !showOnlyGrade {
		argA := models.AbsentFilterRequest{
			LessonIds: &lessonIds,
			StudentId: &child.ID,
		}
		argA.Limit = new(int)
		*argA.Limit = 500
		if limit != nil {
			argA.Limit = limit
		}
		absentsList, _, err = store.Store().AbsentsFindBy(ses.Context(), argA)
		if err != nil {
			return nil, err
		}
	}
	var result []models.GradeDetail
	for _, lesson := range lessonList {
		var grade *models.Grade
		if !showOnlyAbsent {
			for _, g := range gradesList {
				if g.LessonId == lesson.ID {
					grade = g
					break
				}
			}
		}

		var absent *models.Absent
		if !showOnlyGrade {
			for _, a := range absentsList {
				if a.LessonId == lesson.ID {
					absent = a
					break
				}
			}
		}

		if showOnlyAbsent && absent == nil {
			continue
		}
		if showOnlyGrade && grade == nil {
			continue
		}

		result = append(result, models.GradeDetail{
			Lesson: lesson,
			Grade:  grade,
			Absent: absent,
		})
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Absent != nil && result[j].Absent == nil {
			return true
		}
		if result[i].Absent == nil && result[j].Absent != nil {
			return false
		}
		if result[i].Grade != nil && result[j].Grade == nil {
			return true
		}
		if result[i].Grade == nil && result[j].Grade != nil {
			return false
		}
		if result[i].Lesson != nil && result[j].Lesson == nil {
			return true
		}
		return result[i].Lesson.ID > result[j].Lesson.ID
	})
	return result, nil
}

func (a App) ParentAnalyticsExams(ses *utils.Session, child *models.User, subjectId string, classroomId string, p Permission) (*models.RatingCenterSchool, error) {
	sp, ctx := apm.StartSpan(ses.Context(), "ParentAnalyticsExams", "app")
	ses.SetContext(ctx)
	defer sp.End()
	var err error
	if err := a.CheckPayment(child, p); err != nil {
		return nil, err
	}
	if len(child.Classrooms) < 1 {
		return nil, ErrNotSet.SetKey("classroom_id")
	}
	classroom, err := store.Store().ClassroomsFindById(ses.Context(), classroomId)
	if err != nil {
		return nil, err
	}
	err = store.Store().ClassroomsLoadRelations(ses.Context(), &[]*models.Classroom{classroom}, true)
	if err != nil {
		return nil, err
	}
	// Fetch all subjects with the same BaseSubjectId
	subjects := []*models.Subject{}
	for _, subject := range classroom.Subjects {
		argsS := models.SubjectFilterRequest{
			BaseSubjectId: subject.BaseSubjectId,
		}
		subjectsList, _, err := store.Store().SubjectsListFilters(ses.Context(), &argsS)
		if err != nil {
			return nil, err
		}
		subjects = append(subjects, subjectsList...)
	}

	// Fetch all exams for the subjects sharing the same BaseSubjectId
	exams := []*models.SubjectExam{}
	for _, subject := range subjects {
		// TODO: subjectIds filter (N+1 problem)
		argE := models.SubjectExamFilterRequest{
			SubjectIds: []string{subject.ID},
		}
		subjectExamsList, _, err := store.Store().SubjectExamsFindBy(ses.Context(), &argE)
		if err != nil {
			return nil, err
		}
		exams = append(exams, subjectExamsList...)
	}
	// Calculate RatingByLevel points for each student based on all exams for BaseSubjectId
	pointsMapLevel := make(map[string]int)
	pointsCountLevelMap := make(map[string]int)
	studentsMap := make(map[string]models.User)
	var ratingByLevel []models.RatingByLevel
	for _, exam := range exams {
		argG := models.PeriodGradeFilterRequest{
			ExamId: &exam.ID,
		}
		grades, _, err := store.Store().PeriodGradesFindBy(ses.Context(), argG)
		if err != nil {
			return nil, err
		}
		err = store.Store().PeriodGradesLoadRelations(ses.Context(), &grades)
		if err != nil {
			return nil, err
		}
		for _, grade := range grades {
			pointsMapLevel[*grade.StudentId] += grade.GradeSum
			pointsCountLevelMap[*grade.StudentId] += 1
			studentsMap[*grade.StudentId] = *grade.Student
		}
	}

	// Calculate RatingByGroup points for each student based on exam
	pointsMap := make(map[string]int)
	pointsCountGroupMap := make(map[string]int)
	studentsGroupMap := make(map[string]models.User)
	// TODO: struct menzesler merge + response ayratyn
	var ratingByGroup []models.RatingByGroup
	argE := models.SubjectExamFilterRequest{
		ClassroomId: &classroomId,
		SubjectId:   &subjectId,
	}
	subjectExamsList, _, err := store.Store().SubjectExamsFindBy(ses.Context(), &argE)
	if err != nil {
		return nil, err
	}
	for _, ss := range subjectExamsList {
		argG := models.PeriodGradeFilterRequest{
			ExamId: &ss.ID,
		}
		grades, _, err := store.Store().PeriodGradesFindBy(ses.Context(), argG)
		if err != nil {
			return nil, err
		}
		err = store.Store().PeriodGradesLoadRelations(ses.Context(), &grades)
		if err != nil {
			return nil, err
		}
		for _, grade := range grades {
			pointsMap[*grade.StudentId] += grade.GradeSum
			pointsCountGroupMap[*grade.StudentId] += 1
			studentsGroupMap[*grade.StudentId] = *grade.Student
		}
	}

	for id, student := range studentsGroupMap {
		if pointsCountGroupMap[id] > 0 {
			pointsMap[id] = pointsMap[id] / pointsCountGroupMap[id]
		}
		ratingByGroup = append(ratingByGroup, models.RatingByGroup{
			User:   student,
			Points: pointsMap[id],
		})
	}

	for id, student := range studentsMap {
		if pointsCountLevelMap[id] > 0 {
			pointsMapLevel[id] = pointsMapLevel[id] / pointsCountLevelMap[id]
		}
		ratingByLevel = append(ratingByLevel, models.RatingByLevel{
			// TODO: optimize fetch student
			User:   student,
			Points: pointsMapLevel[id],
		})
	}

	//+++ Sort the ratingByGroup and ratingByLevel slice by points in descending order
	sort.Slice(ratingByGroup, func(i, j int) bool {
		return ratingByGroup[i].Points > ratingByGroup[j].Points
	})

	sort.Slice(ratingByLevel, func(i, j int) bool {
		return ratingByLevel[i].Points > ratingByLevel[j].Points
	})

	//+++ Assign indexes
	for i := range ratingByGroup {
		ratingByGroup[i].Index = i + 1
	}
	for i := range ratingByLevel {
		ratingByLevel[i].Index = i + 1
	}

	//+++ Return the results
	return &models.RatingCenterSchool{
		RatingByGroup: ratingByGroup,
		RatingByLevel: ratingByLevel,
	}, nil
}

func analyticsSubjectPercent(student *models.User, classroomId string, startDate time.Time, endDate time.Time) ([]models.SubjectPercentByArea, error) {
	args := models.SubjectFilterRequest{
		ClassroomIds: []string{},
	}
	args.Limit = new(int)
	*args.Limit = 100

	subjectPercents, err := store.Store().SubjectsPercentByStudentWithPrev(context.Background(), student.ID, classroomId, startDate, endDate)
	if err != nil {
		return nil, err
	}

	resPercent := []models.SubjectPercentByArea{}

	for _, v := range models.DefaultSubjectAreas {
		resItem := models.SubjectPercentByArea{
			Title:     string(v),
			BySubject: []models.SubjectPercent{},
		}

		// find subjectPercents which belongs to v(current subject area)
		for _, sp := range subjectPercents {
			for _, s := range models.DefaultSubjects {
				if len(s) > 2 && s[2] == string(v) || len(s) == 2 && v == models.SubjectAreaHumanity {
					if sp.SubjectFullName == s[1] {
						resItem.BySubject = append(resItem.BySubject, sp)
					}
				}
			}
		}
		resItem.CalcPoint()
		resPercent = append(resPercent, resItem)
	}
	models.CalcPercentByArea(&resPercent)
	return resPercent, nil
}

func analyticsRating(ses *utils.Session, student *models.User, classroomId string, schoolId string, periodNum int, startDate time.Time, endDate time.Time) (*models.SubjectRatingByPeriodGrade, error) {
	sp, ctx := apm.StartSpan(ses.Context(), "analyticsRating", "app")
	ses.SetContext(ctx)
	defer sp.End()
	subjectRatingAll, err := store.Store().SubjectsRatingByStudentWithPrev(ses.Context(), classroomId, startDate, endDate)
	if err != nil {
		return nil, err
	}
	// show only selected student
	subjectRating := []models.SubjectRating{}
	for _, v := range subjectRatingAll {
		if v.StudentId == student.ID {
			subjectRating = append(subjectRating, v)
		}
	}
	// add period grades
	periodGrades, _, err := store.Store().PeriodGradesFindBy(ses.Context(), models.PeriodGradeFilterRequest{
		StudentId: &student.ID,
		PeriodKey: &periodNum,
	})
	if err != nil {
		return nil, err
	}
	for _, v := range periodGrades {
		for kk, vv := range subjectRating {
			if *v.SubjectId == vv.SubjectId {
				resItem := models.PeriodGradeResponse{}
				resItem.FromModel(v)
				subjectRating[kk].PeriodGrade = resItem
			}
		}
	}

	resRating := models.SubjectRatingByPeriodGrade{}
	resRating.BySubject = subjectRating
	resRating.CalcRating()
	return &resRating, nil
}

func analyticsGradeStreak(student *models.User, classroomId string) (int, error) {
	strike := 0
	strikeGrades, err := store.Store().SubjectsGradeStrike(context.Background(), classroomId, student.ID)
	if err != nil {
		return 0, err
	}
	yesterday := time.Now().AddDate(0, 0, -1)
	// more strikes and less than or equal yesterday
	if len(strikeGrades) > 0 && yesterday.Compare(strikeGrades[0].LessonDate) >= 0 {
		for _, v := range strikeGrades {
			if v.GradesGoodCount > 0 {
				strike++
			} else {
				break
			}
		}
	}
	return strike, nil
}
