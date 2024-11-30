package app

import (
	"log"
	"slices"
	"strconv"

	"github.com/mekdep/server/internal/api/utils"
	"github.com/mekdep/server/internal/models"
	"github.com/mekdep/server/internal/store"
	"go.elastic.co/apm/v2"
)

func (a App) StatisticsAttendance(ses *utils.Session, req ReportsRequestDto) (StatisticsResponse, error) {
	sp, ctx := apm.StartSpan(ses.Context(), "StatisticsAttendance", "app")
	ses.SetContext(ctx)
	defer sp.End()
	args := models.SchoolFilterRequest{
		Uids: req.SchoolIds,
	}
	args.Limit = new(int)
	*args.Limit = 500
	args.IsParent = new(bool)
	*args.IsParent = false
	sl, _, err := store.Store().SchoolsFindBy(ses.Context(), args)
	if err != nil {
		return StatisticsResponse{}, err
	}

	err = store.Store().SchoolsLoadRelations(ses.Context(), &sl)
	if err != nil {
		return StatisticsResponse{}, err
	}

	log.Println("SchoolsLoadRelations")

	// get all current period and finish

	log.Println("current period")

	log.Println("period completed grades")
	subjectPercents, err := store.Store().SubjectsPercentsBySchool(ses.Context(), *req.SchoolIds, *req.StartDate, *req.EndDate)
	if err != nil {
		return StatisticsResponse{}, err
	}

	resRows := []*StatisticsRow{}
	for _, schoolItem := range sl {
		// form StatisticsRow
		rowItem := StatisticsRow{}
		rowItem.FromSchool(schoolItem)
		percent := uint(100)
		subjectPercent := models.DashboardSubjectsPercentBySchool{}
		for _, sp := range subjectPercents {
			if sp.SchoolID == schoolItem.ID {
				subjectPercent = sp
				percent = uint(sp.AbsentPercent)
			}
		}
		if subjectPercent.SchoolID == "" {
			continue
		}
		rowItem.Percent = &percent

		// place all by keys, calc total users
		// calc Data percent and place
		// Fetch journal percent and place
		rowItem.Values = []StatisticsCell{}
		rowItem.Values = append(rowItem.Values, StatisticsCell(*schoolItem.Name))
		rowItem.Values = append(rowItem.Values, StatisticsCell(strconv.Itoa(int(subjectPercent.LessonsCount))))
		rowItem.Values = append(rowItem.Values, StatisticsCell(strconv.Itoa(int(subjectPercent.StudentsCount))))
		rowItem.Values = append(rowItem.Values, StatisticsCell(strconv.Itoa(int(percent))))
		resRows = append(resRows, &rowItem)
	}
	resRowsNoPtr := []StatisticsRow{}
	for _, v := range resRows {
		resRowsNoPtr = append(resRowsNoPtr, *v)
	}

	// insights
	totalSchool := len(sl)
	totalRegion := []string{}
	for _, v := range resRowsNoPtr {
		if !slices.Contains(totalRegion, v.Region) {
			totalRegion = append(totalRegion, v.Region)
		}
	}
	totalEq100 := []string{}
	for _, v := range resRowsNoPtr {
		if v.Percent != nil && *v.Percent == 100 {
			totalEq100 = append(totalEq100, *v.SchoolCode)
		}
	}
	totalMore90 := []string{}
	for _, v := range resRowsNoPtr {
		if v.Percent != nil && *v.Percent >= 90 {
			totalMore90 = append(totalMore90, *v.SchoolCode)
		}
	}
	totalLess90 := []string{}
	for _, v := range resRowsNoPtr {
		if v.Percent != nil && *v.Percent < 90 {
			totalLess90 = append(totalLess90, *v.SchoolCode)
		}
	}

	totals := []StatisticsTotal{
		{
			Title: "Jemi mekdep sany",
			Value: strconv.Itoa(totalSchool),
			Type:  "number",
		},
		{
			Title: "Jemi etrap sany",
			Value: strconv.Itoa(len(totalRegion)),
			Type:  "number",
		},
		{
			Title: "Gatnasyk 100 mekdepler",
			Value: strconv.Itoa(len(totalEq100)),
			Type:  "number",
		},
		{
			Title: "Gatnasyk 90 köp mekdepler",
			Value: strconv.Itoa(len(totalMore90)),
			Type:  "number",
		},
		{
			Title: "Gatnasyk 90 az mekdepler",
			Value: strconv.Itoa(len(totalLess90)),
			Type:  "number",
		},
	}

	res := StatisticsResponse{
		Headers: []StatisticsHeader{"Mekdep", "Sapak okaldy",
			"Jemi okuwçy", "Gatnaşyk %"},
		Rows:   resRowsNoPtr,
		Totals: totals,
	}
	res.HasDetail = true

	return res, nil
}

func (a App) StatisticsAttendanceByClassroom(ses *utils.Session, req ReportsRequestDto) (StatisticsResponse, error) {
	sp, ctx := apm.StartSpan(ses.Context(), "StatisticsAttendance", "app")
	ses.SetContext(ctx)
	defer sp.End()

	args := models.ClassroomFilterRequest{
		SchoolId: req.SchoolId,
	}
	args.Limit = new(int)
	*args.Limit = 500
	classrooms, _, err := store.Store().ClassroomsFindBy(ses.Context(), args)
	if err != nil {
		return StatisticsResponse{}, err
	}

	argsSchool := models.SchoolFilterRequest{
		ID: req.SchoolId,
	}
	argsSchool.Limit = new(int)
	*argsSchool.Limit = 500
	sl, _, err := store.Store().SchoolsFindBy(ses.Context(), argsSchool)
	if err != nil {
		return StatisticsResponse{}, err
	}
	err = store.Store().SchoolsLoadRelations(ses.Context(), &sl)
	if err != nil {
		return StatisticsResponse{}, err
	}

	// get all current period and finish

	log.Println("current period")
	argsS := models.SubjectFilterRequest{
		SchoolIds: *req.SchoolIds,
	}

	subjectModels, _, err := store.Store().SubjectsListFilters(ses.Context(), &argsS)
	if err != nil {
		return StatisticsResponse{}, err
	}
	argsS.Limit = new(int)
	*argsS.Limit = 1000
	subjectIds := []string{}
	for _, subjectModel := range subjectModels {
		subjectIds = append(subjectIds, subjectModel.ID)
	}

	subjectPercents, err := store.Store().SubjectsPercents(ses.Context(), subjectIds, *req.StartDate, *req.EndDate)
	if err != nil {
		return StatisticsResponse{}, err
	}

	resRows := []*StatisticsRow{}
	totalStudents := 0
	for _, classroomItem := range classrooms {
		// form StatisticsRow
		rowItem := StatisticsRow{}
		rowItem.FromClassroom(classroomItem)
		rowItem.FromSchool(sl[0])
		totalAttendance := 0
		lessonCount := 0
		subjectPercent := models.DashboardSubjectsPercent{}
		for _, sp := range subjectPercents {
			if sp.ClassroomId == classroomItem.ID {
				totalAttendance += (sp.AbsentPercent)
				lessonCount++
				subjectPercent = sp
			}
		}

		percent := uint(0)
		if lessonCount > 0 {
			percent = uint(totalAttendance / lessonCount)
		}
		rowItem.Percent = &percent

		rowItem.UserId = classroomItem.TeacherId

		// place all by keys, calc total users
		// calc Data percent and place
		// Fetch journal percent and place
		rowItem.Values = []StatisticsCell{}
		rowItem.Values = append(rowItem.Values, StatisticsCell(*classroomItem.Name))
		rowItem.Values = append(rowItem.Values, StatisticsCell(strconv.Itoa(int(lessonCount))))
		rowItem.Values = append(rowItem.Values, StatisticsCell(strconv.Itoa(int(subjectPercent.StudentsCount))))
		rowItem.Values = append(rowItem.Values, StatisticsCell(strconv.Itoa(int(percent))))
		resRows = append(resRows, &rowItem)
		totalStudents += subjectPercent.StudentsCount
	}
	resRowsNoPtr := []StatisticsRow{}
	for _, v := range resRows {
		resRowsNoPtr = append(resRowsNoPtr, *v)
	}

	// insights
	totalSchool := len(resRowsNoPtr)
	totalEq100 := []string{}
	for _, v := range resRowsNoPtr {
		if v.Percent != nil && *v.Percent == 100 {
			totalEq100 = append(totalEq100, *v.Classroom)
		}
	}
	totalMore90 := []string{}
	for _, v := range resRowsNoPtr {
		if v.Percent != nil && *v.Percent >= 90 {
			totalMore90 = append(totalMore90, *v.Classroom)
		}
	}
	totalLess90 := []string{}
	for _, v := range resRowsNoPtr {
		if v.Percent != nil && *v.Percent < 90 {
			totalLess90 = append(totalLess90, *v.Classroom)
		}
	}

	totals := []StatisticsTotal{
		{
			Title: "Jemi synp sany",
			Value: strconv.Itoa(totalSchool),
			Type:  "number",
		},
		{
			Title: "Jemi okuwçy sany",
			Value: strconv.Itoa(totalStudents),
			Type:  "number",
		},
		{
			Title: "Gatnasyk 100 synplar",
			Value: strconv.Itoa(len(totalEq100)),
			Type:  "number",
		},
		{
			Title: "Gatnasyk 90 köp synplar",
			Value: strconv.Itoa(len(totalMore90)),
			Type:  "number",
		},
		{
			Title: "Gatnasyk 90 az synplar",
			Value: strconv.Itoa(len(totalLess90)),
			Type:  "number",
		},
	}

	res := StatisticsResponse{
		Headers: []StatisticsHeader{"Synp", "Sapak okaldy",
			"Jemi okuwçy", "Gatnaşyk %"},
		Rows:   resRowsNoPtr,
		Totals: totals,
	}

	return res, nil
}
