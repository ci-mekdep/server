package app

import (
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"github.com/mekdep/server/internal/api/utils"
	"github.com/mekdep/server/internal/models"
	"github.com/mekdep/server/internal/store"
	"go.elastic.co/apm/v2"
)

func (a App) StatisticsJournalCached(ses *utils.Session, dto ReportsRequestDto) (*StatisticsResponse, error) {
	reqStr, _ := json.Marshal(dto)
	k := "StatisticsJournalCached_" + string(reqStr)
	res := StatisticsResponse{}
	if v, ok := a.cache.Get(k); ok {
		res = v.(StatisticsResponse)
		return &res, nil
	} else {
		v, err := a.StatisticsJournal(ses, dto)
		if err != nil {
			return nil, err
		}
		a.cache.Set(k, *v, time.Hour)
		return v, nil
	}
}

func (a App) StatisticsJournal(ses *utils.Session, dto ReportsRequestDto) (*StatisticsResponse, error) {
	sp, ctx := apm.StartSpan(ses.Context(), "StatisticsJournal", "app")
	ses.SetContext(ctx)
	defer sp.End()
	if dto.PeriodNumber != nil && ses.GetSchoolId() == nil {
		periods := appSettingsPeriods()
		if *dto.PeriodNumber > 0 && *dto.PeriodNumber <= len(periods.Value) {
			period := periods.Value[*dto.PeriodNumber-1]
			startDate, err := time.Parse("2006-01-02", period[0])
			if err != nil {
				return nil, err
			}
			endDate, err := time.Parse("2006-01-02", period[1])
			if err != nil {
				return nil, err
			}
			*dto.StartDate = startDate
			*dto.EndDate = endDate
		}
	}
	args := models.SchoolFilterRequest{
		Uids: dto.SchoolIds,
	}
	args.IsParent = new(bool)
	*args.IsParent = false
	args.Limit = new(int)
	*args.Limit = 500
	schools, _, err := store.Store().SchoolsFindBy(ses.Context(), args)
	if err != nil {
		return nil, err
	}
	err = store.Store().SchoolsLoadRelations(ses.Context(), &schools)
	if err != nil {
		return nil, err
	}
	// get all subjectsPercents
	subjectsPercents, err := store.Store().SubjectsPercentsBySchool(ses.Context(), *dto.SchoolIds, *dto.StartDate, *dto.EndDate)
	if err != nil {
		return nil, err
	}

	resRows := []*StatisticsRow{}
	for _, schoolItem := range schools {
		// get school setting by key
		var subjectPercent *models.DashboardSubjectsPercentBySchool
		for _, v := range subjectsPercents {
			if v.SchoolID == schoolItem.ID {
				subjectPercent = &v
				break
			}
		}

		// form StatisticsRow
		rowItem := StatisticsRow{}
		resRows = append(resRows, &rowItem)
		rowItem.FromSchool(schoolItem)

		// place all by keys, calc total users
		// calc Data percent and place
		// Fetch journal percent and place
		rowItem.Values = []StatisticsCell{}
		rowItem.Values = append(rowItem.Values, StatisticsCell(*schoolItem.Name))
		rowItem.Percent = new(uint)

		if subjectPercent == nil {
			continue
		}

		*rowItem.Percent = uint(subjectPercent.GradeFullPercent)
		rowItem.Values = append(rowItem.Values, StatisticsCell(strconv.Itoa(int(subjectPercent.StudentsCount))))
		rowItem.Values = append(rowItem.Values, StatisticsCell(strconv.Itoa(subjectPercent.LessonsCount)+" / "+strconv.Itoa(subjectPercent.TopicsCount)))
		rowItem.Values = append(rowItem.Values, StatisticsCell(strconv.Itoa(int(*rowItem.Percent))))
	}
	resRowsNoPtr := []StatisticsRow{}
	for _, v := range resRows {
		resRowsNoPtr = append(resRowsNoPtr, *v)
	}

	// set totals, school count, region count,
	// data percent less 80 count, data percent below 10 count
	// journal full (10%) percent, journal empty (1%) count
	// timetables set percent, total classrooms
	totalCount := len(resRowsNoPtr)
	percentSum := 0
	percentCount := 0
	averagePercent := 0
	totalTrue := 0
	totalFalse := 0
	for _, v := range resRowsNoPtr {
		if v.Region == v.School {
			continue
		}
		item := *v.Percent
		percentSum += int(item)
		percentCount++

		if item >= 50 {
			totalTrue++
		} else {
			totalFalse++
		}
	}
	if percentCount > 0 {
		averagePercent = percentSum / percentCount
	}

	totals := []StatisticsTotal{
		StatisticsTotal{
			Title: "Jemi mekdep sany",
			Value: strconv.Itoa(totalCount),
			Type:  "number",
		},
		StatisticsTotal{
			Title: "Doly sany (50% +)",
			Value: strconv.Itoa(totalTrue),
			Type:  "number",
		},
		StatisticsTotal{
			Title: "Doly däl sany",
			Value: strconv.Itoa(totalFalse),
			Type:  "number",
		},
		StatisticsTotal{
			Title: "Ortaça dolulyk (%)",
			Value: strconv.Itoa(averagePercent),
			Type:  "number",
		},
	}

	res := StatisticsResponse{
		Headers:        []StatisticsHeader{"Mekdep", "Okuwçy sany", "Geçilen sapak sany", "Žurnal dolulyk %"},
		Rows:           resRowsNoPtr,
		HasDetail:      true,
		HasBetweenDate: true,
		Totals:         totals,
	}

	return &res, nil
}

func (a App) StatisticsJournalByTeacher(ses *utils.Session, dto ReportsRequestDto) (*StatisticsResponse, error) {
	sp, ctx := apm.StartSpan(ses.Context(), "StatisticsJournal", "app")
	ses.SetContext(ctx)
	defer sp.End()

	if dto.SchoolId == nil || *dto.SchoolId == "" {
		return nil, errors.New("no school given")
	}
	if dto.PeriodNumber != nil && ses.GetSchoolId() != nil {
		argsP := models.PeriodFilterRequest{
			SchoolId: ses.GetSchoolId(),
		}
		period, _, err := store.Store().PeriodsListFilters(ses.Context(), argsP)
		if err != nil {
			return nil, err
		}
		periodStart, periodEnd, err := period[0].DatesByKey(*dto.PeriodNumber)
		if err != nil {
			return nil, err
		}
		*dto.StartDate = periodStart
		*dto.EndDate = periodEnd
	}
	schoolIds := []string{*dto.SchoolId}
	args := models.SchoolFilterRequest{
		Uids: &schoolIds,
	}
	args.IsParent = new(bool)
	*args.IsParent = false
	args.Limit = new(int)
	*args.Limit = 500
	schools, _, err := store.Store().SchoolsFindBy(ses.Context(), args)
	if err != nil {
		return nil, err
	}
	err = store.Store().SchoolsLoadRelations(ses.Context(), &schools)
	if err != nil {
		return nil, err
	}
	schoolItem := schools[0]

	// get teachers
	argsU := models.UserFilterRequest{
		SchoolIds: &schoolIds,
	}
	argsU.Limit = new(int)
	argsU.Offset = new(int)
	argsU.Role = new(string)
	*argsU.Limit = 1000
	*argsU.Offset = 0
	*argsU.Role = string(models.RoleTeacher)
	teachers, _, err := store.Store().UsersFindBy(ses.Context(), argsU)
	if err != nil {
		return nil, err
	}

	// get subjects
	argsS := models.SubjectFilterRequest{
		SchoolIds: schoolIds,
	}
	argsS.Limit = new(int)
	argsS.Offset = new(int)
	*argsS.Limit = 2000
	*argsS.Offset = 0
	subjects, _, err := store.Store().SubjectsListFilters(ses.Context(), &argsS)
	if err != nil {
		return nil, err
	}
	subjectIds := []string{}
	for _, v := range subjects {
		subjectIds = append(subjectIds, v.ID)
	}

	// get all subjectsPercents
	subjectsPercents, err := store.Store().SubjectsPercents(ses.Context(), subjectIds, *dto.StartDate, *dto.EndDate)
	if err != nil {
		return nil, err
	}

	resRows := []*StatisticsRow{}
	for _, teacherItem := range teachers {
		// get school setting by key
		totalTrue := 0
		totalFalse := 0
		for _, v := range subjects {
			if v.IsTeacherEq(teacherItem.ID) {
				for _, vv := range subjectsPercents {
					if vv.SubjectId == v.ID {
						if vv.IsGradeFull {
							totalTrue++
						} else {
							totalFalse++
						}
					}
				}
			}
		}

		// form StatisticsRow
		rowItem := StatisticsRow{}
		resRows = append(resRows, &rowItem)
		rowItem.FromSchool(schoolItem)
		rowItem.FromUser(teacherItem)

		// place all by keys, calc total users
		// calc Data percent and place
		// Fetch journal percent and place
		rowItem.Values = []StatisticsCell{}
		rowItem.Values = append(rowItem.Values, StatisticsCell(teacherItem.FullName()))

		percent := uint(0)
		total := totalTrue + totalFalse
		if totalTrue > 0 {
			percent = uint(totalTrue * 100 / total)
		}
		rowItem.Percent = &percent
		rowItem.Values = append(rowItem.Values, StatisticsCell(strconv.Itoa(total)))
		rowItem.Values = append(rowItem.Values, StatisticsCell(strconv.Itoa(totalTrue)))
		rowItem.Values = append(rowItem.Values, StatisticsCell(strconv.Itoa(totalFalse)))
		rowItem.Values = append(rowItem.Values, StatisticsCell(strconv.Itoa(int(*rowItem.Percent))))
	}
	resRowsNoPtr := []StatisticsRow{}
	for _, v := range resRows {
		resRowsNoPtr = append(resRowsNoPtr, *v)
	}

	// set totals, school count, region count,
	// data percent less 80 count, data percent below 10 count
	// journal full (10%) percent, journal empty (1%) count
	// timetables set percent, total classrooms
	totalCount := len(resRowsNoPtr)
	percentSum := 0
	percentCount := 0
	averagePercent := 0
	totalTrue := 0
	totalFalse := 0
	for _, v := range resRowsNoPtr {
		if v.Region == v.School {
			continue
		}
		item := *v.Percent
		percentSum += int(item)
		percentCount++

		if item >= 90 {
			totalTrue++
		} else {
			totalFalse++
		}
	}
	if percentCount > 0 {
		averagePercent = percentSum / percentCount
	}

	totals := []StatisticsTotal{
		StatisticsTotal{
			Title: "Jemi mugallym sany",
			Value: strconv.Itoa(totalCount),
			Type:  "number",
		},
		StatisticsTotal{
			Title: "Doly sany",
			Value: strconv.Itoa(totalTrue),
			Type:  "number",
		},
		StatisticsTotal{
			Title: "Doly däl sany",
			Value: strconv.Itoa(totalFalse),
			Type:  "number",
		},
		StatisticsTotal{
			Title: "Ortaça dolulyk (%)",
			Value: strconv.Itoa(averagePercent),
			Type:  "number",
		},
	}

	res := StatisticsResponse{
		Headers:        []StatisticsHeader{"Mugallym", "Sagat sany", "Doly", "Doly däl", "Dolulyk %"},
		HasBetweenDate: true,
		HasDetail:      true,
		Rows:           resRowsNoPtr,
		Totals:         totals,
	}

	return &res, nil
}

func (a App) StatisticsJournalByLesson(ses *utils.Session, dto ReportsRequestDto) (*StatisticsResponse, error) {
	sp, ctx := apm.StartSpan(ses.Context(), "StatisticsJournalByLesson", "app")
	ses.SetContext(ctx)
	defer sp.End()

	if dto.SchoolId == nil || *dto.SchoolId == "" {
		return nil, errors.New("no school given")
	}
	var schoolItem *models.School
	for _, v := range ses.GetSchools() {
		if v.School != nil && v.School.ID == *dto.SchoolId {
			schoolItem = v.School
			break
		}
	}
	if schoolItem == nil {
		return nil, errors.New("no school given")
	}
	if dto.UserId == nil || *dto.UserId == "" {
		return nil, errors.New("no user given")
	}

	// get teachers
	teacherItem, err := store.Store().UsersFindById(ses.Context(), *dto.UserId)

	// get subjects
	argsS := models.SubjectFilterRequest{
		SchoolIds: []string{*dto.SchoolId},
		TeacherId: &teacherItem.ID,
	}
	argsS.Limit = new(int)
	argsS.Offset = new(int)
	*argsS.Limit = 200
	*argsS.Offset = 0
	subjects, _, err := store.Store().SubjectsListFilters(ses.Context(), &argsS)
	if err != nil {
		return nil, err
	}
	subjectIds := []string{}
	for _, v := range subjects {
		subjectIds = append(subjectIds, v.ID)
	}

	// get all subjectsPercents
	subjectsPercents, err := store.Store().SubjectsPercents(ses.Context(), subjectIds, *dto.StartDate, *dto.EndDate)
	if err != nil {
		return nil, err
	}

	resRows := []*StatisticsRow{}
	// get school setting by key
	for _, subjectItem := range subjectsPercents {
		// form StatisticsRow
		rowItem := StatisticsRow{}
		resRows = append(resRows, &rowItem)
		rowItem.FromSchool(schoolItem)
		rowItem.FromUser(teacherItem)
		percent := uint(subjectItem.GradeFullPercent)
		rowItem.Percent = &percent
		if subjectItem.LessonTitle == nil {
			subjectItem.LessonTitle = new(string)
		}

		// place all by keys, calc total users
		// calc Data percent and place
		// Fetch journal percent and place
		rowItem.Values = []StatisticsCell{}
		rowItem.Values = append(rowItem.Values, StatisticsCell(
			subjectItem.ClassroomName+" "+subjectItem.SubjectName+" "+subjectItem.LessonDate.Format(time.DateOnly)))
		rowItem.Values = append(rowItem.Values, StatisticsCell(*subjectItem.LessonTitle))
		rowItem.Values = append(rowItem.Values, StatisticsCell(strconv.Itoa(subjectItem.GradeFullPercent)))
		rowItem.Values = append(rowItem.Values, StatisticsCell(strconv.Itoa(subjectItem.AbsentPercent)))
		rowItem.Values = append(rowItem.Values, StatisticsCell(strconv.Itoa(subjectItem.StudentsCount)))

	}
	resRowsNoPtr := []StatisticsRow{}
	for _, v := range resRows {
		resRowsNoPtr = append(resRowsNoPtr, *v)
	}

	// set totals, school count, region count,
	// data percent less 80 count, data percent below 10 count
	// journal full (10%) percent, journal empty (1%) count
	// timetables set percent, total classrooms
	totalCount := len(resRowsNoPtr)
	percentSum := 0
	percentCount := 0
	averagePercent := 0
	for _, v := range resRowsNoPtr {
		if v.Region == v.School {
			continue
		}
		item := *v.Percent
		percentSum += int(item)
		percentCount++

	}
	if percentCount > 0 {
		averagePercent = percentSum / percentCount
	}

	totals := []StatisticsTotal{
		StatisticsTotal{
			Title: "Jemi ders sany",
			Value: strconv.Itoa(totalCount),
			Type:  "number",
		},
		StatisticsTotal{
			Title: "Ortaça dolulyk (%)",
			Value: strconv.Itoa(averagePercent),
			Type:  "number",
		},
	}

	res := StatisticsResponse{
		Headers:        []StatisticsHeader{"Ders ýükiň ady", "Temanyň ady", "Dolulyk %", "Gatnaşyk %", "Okuwçy sany"},
		HasBetweenDate: true,
		HasDetail:      true,
		Rows:           resRowsNoPtr,
		Totals:         totals,
	}

	return &res, nil
}
