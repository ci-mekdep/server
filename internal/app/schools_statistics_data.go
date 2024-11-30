package app

import (
	"encoding/json"
	"log"
	"slices"
	"strconv"
	"time"

	"github.com/mekdep/server/internal/api/utils"
	"github.com/mekdep/server/internal/models"
	"github.com/mekdep/server/internal/store"
	"go.elastic.co/apm/v2"
)

func (a App) StatisticsSchoolDataCached(ses *utils.Session, schoolIds []string, startDate time.Time, endDate time.Time) (StatisticsResponse, error) {
	schoolIdsStr, _ := json.Marshal(schoolIds)
	k := "StatisticsSchoolData_" + string(schoolIdsStr) + startDate.Format(time.DateOnly) + endDate.Format(time.DateOnly)
	if v, ok := a.cache.Get(k); ok {
		return v.(StatisticsResponse), nil
	} else {
		v, err := a.StatisticsSchoolData(ses, schoolIds, startDate, endDate)
		if err != nil {
			return StatisticsResponse{}, err
		}
		a.cache.Set(k, v, time.Minute*60)
		return v, nil
	}
}

func (a App) StatisticsSchoolData(ses *utils.Session, schoolIds []string, startDate time.Time, endDate time.Time) (StatisticsResponse, error) {
	sp, ctx := apm.StartSpan(ses.Context(), "StatisticsSchoolData", "app")
	ses.SetContext(ctx)
	defer sp.End()
	args := models.SchoolFilterRequest{
		Uids: &schoolIds,
	}
	args.IsParent = new(bool)
	*args.IsParent = false
	args.Limit = new(int)
	*args.Limit = 500
	sl, _, err := store.Store().SchoolsFindBy(ses.Context(), args)
	if err != nil {
		return StatisticsResponse{}, err
	}
	err = store.Store().SchoolsLoadRelations(ses.Context(), &sl)
	if err != nil {
		return StatisticsResponse{}, err
	}
	log.Println("SchoolsLoadRelations")
	// get all schoolSettings
	schoolSettings, err := store.Store().SchoolSettingsGet(ses.Context(), schoolIds)
	if err != nil {
		return StatisticsResponse{}, err
	}
	// get all userCounts
	userCounts, err := store.Store().UsersLoadCountBySchool(ses.Context(), schoolIds)
	if err != nil {
		return StatisticsResponse{}, err
	}

	log.Println("UsersLoadCountBySchool")
	// get journal percent
	// subjectPercents, err := store.Store().SubjectsPercentsBySchool(ses.Context(), schoolIds, startDate, endDate)
	// if err != nil {
	// 	return StatisticsResponse{}, err
	// }
	// log.Println("SubjectsLoadPercentsBySchool")

	resRows := []*StatisticsRow{}
	resRowsGrouped := map[string]map[int]int{}
	for _, schoolItem := range sl {
		// get school setting by key
		schoolSetting := map[models.SchoolSettingKey]models.SchoolSetting{}
		for _, v := range schoolSettings {
			if *v.SchoolId == schoolItem.ID {
				schoolSetting[v.Key] = v
			}
		}
		settingTeachersCount := 0
		settingValue, exists := schoolSetting[models.SchoolSettingTeachersCount]
		if exists && settingValue.Value != nil {
			settingTeachersCount, _ = strconv.Atoi(*settingValue.Value)
		}
		settingStudentsCount := 0
		settingValue, exists = schoolSetting[models.SchoolSettingStudentsCount]
		if exists && settingValue.Value != nil {
			settingStudentsCount, _ = strconv.Atoi(*settingValue.Value)
		}
		settingParentsCount := 0
		settingValue, exists = schoolSetting[models.SchoolSettingParentsCount]
		if exists && settingValue.Value != nil {
			settingParentsCount, _ = strconv.Atoi(*settingValue.Value)
		}
		settingClassroomsCount := 0
		settingValue, exists = schoolSetting[models.SchoolSettingClassroomsCount]
		if exists && settingValue.Value != nil {
			settingClassroomsCount, _ = strconv.Atoi(*settingValue.Value)
		}

		// get userCount
		userCount := models.DashboardUsersCount{}
		for _, v := range userCounts {
			if v.SchoolId != nil && *v.SchoolId == schoolItem.ID {
				userCount = v
				break
			}
		}
		if userCount.SchoolId == nil {
			continue
		}
		// get subject percent
		// subjectPercent := models.DashboardSubjectsPercentBySchool{}
		// for _, v := range subjectPercents {
		// 	if v.SchoolID == schoolItem.ID {
		// 		subjectPercent = v
		// 	}
		// }

		// form StatisticsRow
		rowItem := StatisticsRow{}
		resRows = append(resRows, &rowItem)
		rowItem.FromSchool(schoolItem)

		// place all by keys, calc total users
		// calc Data percent and place
		// Fetch journal percent and place
		rowItem.Values = []StatisticsCell{}
		rowItem.Values = append(rowItem.Values, StatisticsCell(*schoolItem.Name))
		totalDataPercentTeacher := 0
		if userCount.TeachersCount != 0 && settingTeachersCount != 0 {
			totalDataPercentTeacher = userCount.TeachersCount * 100 / settingTeachersCount
		}
		totalDataPercentStudent := 0
		if userCount.StudentsCount != 0 && settingStudentsCount != 0 {
			totalDataPercentStudent = userCount.StudentsCount * 100 / settingStudentsCount
		}
		totalDataPercentParent := 0
		if userCount.ParentsCount != 0 && settingParentsCount != 0 {
			totalDataPercentParent = userCount.ParentsCount * 100 / settingParentsCount
		}
		if totalDataPercentTeacher > 100 {
			totalDataPercentTeacher = 100
		}
		if totalDataPercentStudent > 100 {
			totalDataPercentStudent = 100
		}
		if totalDataPercentParent > 100 {
			totalDataPercentParent = 100
		}
		totalUsers := userCount.TeachersCount + userCount.StudentsCount + userCount.ParentsCount
		totalDataPercent := uint(totalDataPercentTeacher+totalDataPercentStudent+totalDataPercentParent) / 3
		rowItem.Percent = &totalDataPercent
		rowItem.Values = append(rowItem.Values, StatisticsCell(strconv.Itoa(int(totalDataPercent))))              // data %
		rowItem.Values = append(rowItem.Values, StatisticsCell(strconv.Itoa(userCount.TeachersCount)))            // teacher
		rowItem.Values = append(rowItem.Values, StatisticsCell(strconv.Itoa(settingTeachersCount)))               // setting teachers
		rowItem.Values = append(rowItem.Values, StatisticsCell(strconv.Itoa(userCount.StudentsCount)))            // student
		rowItem.Values = append(rowItem.Values, StatisticsCell(strconv.Itoa(settingStudentsCount)))               // setting students
		rowItem.Values = append(rowItem.Values, StatisticsCell(strconv.Itoa(userCount.ParentsCount)))             // parent
		rowItem.Values = append(rowItem.Values, StatisticsCell(strconv.Itoa(settingParentsCount)))                // setting parents
		rowItem.Values = append(rowItem.Values, StatisticsCell(strconv.Itoa(userCount.StudentsWithParentsCount))) // students with parents
		rowItem.Values = append(rowItem.Values, StatisticsCell(strconv.Itoa(userCount.TimetablesCount)))          // classrooms
		rowItem.Values = append(rowItem.Values, StatisticsCell(strconv.Itoa(settingClassroomsCount)))             // setting classrooms
		// rowItem.Values = append(rowItem.Values, StatisticsCell(strconv.Itoa(int(subjectPercent.GradeFullPercent))))                                   // journal %
		// rowItem.Values = append(rowItem.Values, StatisticsCell(strconv.Itoa(userCount.UsersOnlineCount)))                                             // users online
		rowItem.Values = append(rowItem.Values, StatisticsCell(strconv.Itoa(totalUsers))) // users total
		totalDataPercentUint := uint(totalDataPercent)
		rowItem.Percent = &totalDataPercentUint

		if resRowsGrouped[rowItem.Region] == nil {
			resRowsGrouped[rowItem.Region] = map[int]int{}
		}
		resRowsGrouped[rowItem.Region][1] = (resRowsGrouped[rowItem.Region][1] + int(totalDataPercent)) / 2
		resRowsGrouped[rowItem.Region][2] += userCount.TeachersCount
		resRowsGrouped[rowItem.Region][3] += settingTeachersCount
		resRowsGrouped[rowItem.Region][4] += userCount.StudentsCount
		resRowsGrouped[rowItem.Region][5] += settingStudentsCount
		resRowsGrouped[rowItem.Region][6] += userCount.ParentsCount
		resRowsGrouped[rowItem.Region][7] += settingParentsCount
		resRowsGrouped[rowItem.Region][8] += userCount.StudentsWithParentsCount
		resRowsGrouped[rowItem.Region][9] += userCount.TimetablesCount
		resRowsGrouped[rowItem.Region][10] += settingClassroomsCount
		resRowsGrouped[rowItem.Region][11] += totalUsers
	}
	for k, v := range resRowsGrouped {
		rowItem := StatisticsRow{}
		resRows = append(resRows, &rowItem)
		rowItem.School = k
		rowItem.Region = k
		rowItem.Values = make([]StatisticsCell, len(v)+1)
		rowItem.Values[0] = StatisticsCell(k)
		for kk, vv := range v {
			rowItem.Values[kk] = StatisticsCell(strconv.Itoa(vv))
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
	totalSchool := len(sl)
	totalRegion := []string{}
	for _, v := range resRowsNoPtr {
		if v.Region == v.School {
			continue
		}
		if !slices.Contains(totalRegion, v.Region) {
			totalRegion = append(totalRegion, v.Region)
		}
	}
	totalDataPercentSum := 0
	totalDataPercentCount := 0
	for _, v := range resRowsNoPtr {
		if v.Region == v.School {
			continue
		}
		item, _ := strconv.Atoi(string(v.Values[1]))
		if item < 1 {
			continue
		}
		totalDataPercentSum += item
		totalDataPercentCount++
	}
	totalDataPercent := 0
	if totalDataPercentCount > 0 {
		totalDataPercent = totalDataPercentSum / totalDataPercentCount
	}

	totalClassroomsCount := 0
	for _, uc := range userCounts {
		totalClassroomsCount += uc.ClassroomsCount
	}
	totalTimetablesCount := 0
	for _, uc := range userCounts {
		totalTimetablesCount += uc.TimetablesCount
	}

	totalTimetablesFullPercent := 0
	if totalClassroomsCount != 0 {
		totalTimetablesFullPercent = totalTimetablesCount * 100 / totalClassroomsCount
	}
	totalSettingsSet := 0
	for _, v := range schoolSettings {
		if v.Value == nil || *v.Value == "0" {
			continue
		}
		if v.Key == models.SchoolSettingStudentsCount {
			totalSettingsSet++
		}
	}
	totals := []StatisticsTotal{
		StatisticsTotal{
			Title: "Jemi mekdep sany",
			Value: strconv.Itoa(totalSchool),
			Type:  "number",
		},
		StatisticsTotal{
			Title: "Jemi etrap sany",
			Value: strconv.Itoa(len(totalRegion)),
			Type:  "number",
		},
		StatisticsTotal{
			Title: "Hasabat girizen mekdepler",
			Value: strconv.Itoa(totalSettingsSet),
			Type:  "number",
		},
		StatisticsTotal{
			Title: "Jemi ortaça maglumat (%)",
			Value: strconv.Itoa(totalDataPercent),
			Type:  "number",
		},
		StatisticsTotal{
			Title: "Ders reje girizilen (%)",
			Value: strconv.Itoa(totalTimetablesFullPercent),
			Type:  "number",
		},
		StatisticsTotal{
			Title: "Jemi synp sany",
			Value: strconv.Itoa(totalClassroomsCount),
			Type:  "number",
		},
	}

	res := StatisticsResponse{
		Headers: []StatisticsHeader{"Mekdep", "Maglumat dolulygy %",
			"Mugallymlar ulgamda", "Mugallymlar hakyky", "Okuwçylar ulgamda", "Okuwçylar hakyky", "Ata-eneler ulgamda", "Ata-eneler hakyky",
			"Okuwçylaryň ata-enesi girizilenleri", "Ders rejeler ulgamda", "Synplar hakyky",
			// "Žurnal dolulygy %", "Ulanyjy onlaýn (1aý)",
			"Jemi ulanyjy"},
		Rows:   resRowsNoPtr,
		Totals: totals,
	}

	return res, nil
}
