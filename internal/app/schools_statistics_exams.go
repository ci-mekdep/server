package app

import (
	"encoding/json"
	"slices"
	"sort"
	"strconv"
	"time"

	"github.com/mekdep/server/config"
	"github.com/mekdep/server/internal/api/utils"
	"github.com/mekdep/server/internal/models"
	"github.com/mekdep/server/internal/store"
	"go.elastic.co/apm/v2"
)

func (a App) StatisticsExamCached(ses *utils.Session, dto ReportsRequestDto) (*StatisticsResponse, error) {
	var isGraduate *bool
	reqStr, _ := json.Marshal(dto)
	k := "StatisticsExamCached_" + string(reqStr)
	res := StatisticsResponse{}
	if v, ok := a.cache.Get(k); ok {
		res = v.(StatisticsResponse)
		return &res, nil
	} else {
		v, err := a.StatisticsExam(ses, dto, isGraduate)
		if err != nil {
			return nil, err
		}
		a.cache.Set(k, *v, time.Minute*10)
		return v, nil
	}
}

func (a App) StatisticsExam(ses *utils.Session, dto ReportsRequestDto, isGraduate *bool) (*StatisticsResponse, error) {
	sp, ctx := apm.StartSpan(ses.Context(), "StatisticsExam", "app")
	ses.SetContext(ctx)
	defer sp.End()

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
	// get all subjectExams
	argsE := models.SubjectExamFilterRequest{
		SchoolIds:  *dto.SchoolIds,
		IsGraduate: isGraduate,
	}
	argsE.Limit = new(int)
	*argsE.Limit = 100000
	subjectExams, _, err := store.Store().SubjectExamsFindBy(ses.Context(), &argsE)
	if err != nil {
		return nil, err
	}
	err = store.Store().SubjectExamLoadRelations(ses.Context(), &subjectExams)
	if err != nil {
		return nil, err
	}

	// map colors by subject
	subjectColors := map[string]string{}
	for _, v := range models.DefaultSubjects {
		if len(v) > 3 {
			subjectColors[v[0]] = v[3]
		} else {
			subjectColors[v[0]] = "#3F7CF2" // just default color
		}
	}

	examsCountBySubject := map[string]int{}

	// set header dates
	headerDates := []string{}
	for _, v := range subjectExams {
		if v.StartTime == nil {
			continue
		}
		dateStr := v.StartTime.Format(time.DateOnly)
		if !slices.Contains(headerDates, dateStr) {
			headerDates = append(headerDates, dateStr)
		}
	}
	sort.Slice(headerDates, func(i, j int) bool {
		return headerDates[i] < headerDates[j] // sort asc (is less)
	})

	resRows := []*StatisticsRow{}
	for _, schoolItem := range schools {
		hasExams := false
		subjectExamsByDate := map[string][]models.SubjectExam{}
		// get school setting by key
		for _, subjectExam := range subjectExams {
			if subjectExam.SchoolId == schoolItem.ID {
				hasExams = true
				if subjectExam.StartTime == nil {
					continue
				}
				*subjectExam.StartTime = subjectExam.StartTime.In(config.RequestLocation)
				subjectExamDate := subjectExam.StartTime.Format(time.DateOnly)
				if subjectExamsByDate[subjectExamDate] == nil {
					subjectExamsByDate[subjectExamDate] = []models.SubjectExam{}
				}
				subjectExamsByDate[subjectExamDate] = append(subjectExamsByDate[subjectExamDate], *subjectExam)
			}
		}
		if !hasExams {
			// empty row StatisticsRow
			rowItem := StatisticsRow{}
			resRows = append(resRows, &rowItem)
			rowItem.FromSchool(schoolItem)
			rowItem.Values = []StatisticsCell{}
			rowItem.Values = append(rowItem.Values, StatisticsCell(*schoolItem.Name))
			rowItem.Percent = new(uint)
		} else {
			maxNumber := 0
			for _, subjectExams := range subjectExamsByDate {
				if len(subjectExams) > maxNumber {
					maxNumber = len(subjectExams)
				}
			}

			for index := 0; index < maxNumber; index++ {
				// form StatisticsRow
				rowItem := StatisticsRow{}
				resRows = append(resRows, &rowItem)
				rowItem.FromSchool(schoolItem)
				rowItem.Values = []StatisticsCell{}
				rowItem.Values = append(rowItem.Values, StatisticsCell(*schoolItem.Name))
				rowItem.Percent = new(uint)

				for _, headerDate := range headerDates {
					if len(subjectExamsByDate[headerDate]) > index &&
						subjectExamsByDate[headerDate][index].Classroom != nil &&
						subjectExamsByDate[headerDate][index].Subject != nil {
						// place all by keys
						subjectExam := subjectExamsByDate[headerDate][index]
						startTime := subjectExam.StartTime.Format("15:04")
						subjectName := *subjectExam.Subject.Name
						rowItem.Values = append(rowItem.Values, StatisticsCell(*subjectExam.Classroom.Name))
						rowItem.Values = append(rowItem.Values, StatisticsCell(subjectName+subjectColors[subjectName]))
						rowItem.Values = append(rowItem.Values, StatisticsCell(startTime))
						examsCountBySubject[subjectName+subjectColors[subjectName]]++
					} else {
						// empty
						rowItem.Values = append(rowItem.Values, StatisticsCell(""))
						rowItem.Values = append(rowItem.Values, StatisticsCell(""))
						rowItem.Values = append(rowItem.Values, StatisticsCell(""))
					}
				}
			}
		}
	}
	resRowsNoPtr := []StatisticsRow{}
	for _, v := range resRows {
		resRowsNoPtr = append(resRowsNoPtr, *v)
	}

	// set totals, school count, region count,
	// data percent less 80 count, data percent below 10 count
	// journal full (10%) percent, journal empty (1%) count
	// timetables set percent, total classrooms
	totalCount := len(schools)
	percentSum := 0
	percentCount := 0
	totalTrue := 0
	totalFalse := 0
	totalExams := 0
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
	totalExams = len(subjectExams)

	totals := []StatisticsTotal{
		StatisticsTotal{
			Title: "Jemi mekdep sany",
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
			Title: "Jemi synag sany",
			Value: strconv.Itoa(totalExams),
			Type:  "number",
		},
	}
	for name, count := range examsCountBySubject {
		totals = append(totals, StatisticsTotal{
			Title: "Jemi " + name,
			Value: strconv.Itoa(count),
			Type:  "number",
		})
	}

	headers := []StatisticsHeader{"Mekdep"}
	for _, headerDate := range headerDates {
		headers = append(headers, StatisticsHeader(headerDate)) // add header 3 times: for classroom, subject, time
		headers = append(headers, StatisticsHeader(headerDate))
		headers = append(headers, StatisticsHeader(headerDate))
	}
	res := StatisticsResponse{
		Headers:        headers,
		Rows:           resRowsNoPtr,
		HasDetail:      true,
		HasBetweenDate: true,
		Totals:         totals,
	}

	return &res, nil
}
