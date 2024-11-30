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

type StatisticsHeader string
type StatisticsCell string
type StatisticsRow struct {
	Region      string           `json:"region"`
	RegionCode  string           `json:"region_code"`
	School      string           `json:"school"`
	SchoolId    string           `json:"school_id"`
	SchoolCode  *string          `json:"school_code"`
	Percent     *uint            `json:"percent"`
	Classroom   *string          `json:"classroom"`
	ClassroomId *string          `json:"classroom_id"`
	User        *string          `json:"user"`
	UserId      *string          `json:"user_id"`
	Values      []StatisticsCell `json:"values"`
}

func (rowItem *StatisticsRow) FromSchool(s *models.School) {
	rowItem.School = *s.Name
	rowItem.SchoolId = s.ID
	rowItem.SchoolCode = s.Code
	if s.Parent != nil {
		rowItem.Region = *s.Parent.Name
		rowItem.RegionCode = *s.Parent.Code
	}
}

func (rowItem *StatisticsRow) FromClassroom(c *models.Classroom) {
	rowItem.Classroom = c.Name
	rowItem.ClassroomId = &c.ID
}

func (rowItem *StatisticsRow) FromUser(u *models.User) {
	rowItem.User = new(string)
	*rowItem.User = u.FullName()
	rowItem.UserId = &u.ID
}

type StatisticsTotal struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Type  string `json:"type"`
}

type StatisticsResponse struct {
	Headers         []StatisticsHeader `json:"headers"`
	Rows            []StatisticsRow    `json:"rows"`
	Totals          []StatisticsTotal  `json:"totals"`
	HasDetail       bool               `json:"has_detail"`
	HasPeriodNumber bool               `json:"has_period_number"`
	HasBetweenDate  bool               `json:"has_between_date"`
}

type ReportsRequestDto struct {
	StartDate    *time.Time `json:"start_date" form:"start_date" time_format:"2006-01-02"`
	EndDate      *time.Time `json:"end_date" form:"end_date" time_format:"2006-01-02"`
	SchoolId     *string    `json:"school_id" form:"school_id"`
	UserId       *string    `json:"user_id" form:"user_id"`
	SchoolIds    *[]string  ``
	PeriodNumber *int       `json:"period_number" form:"period_number"`
}

func (a App) StatisticsPeriodFinished(ses *utils.Session, schoolIds []string, periodNumber int) (StatisticsResponse, error) {
	sp, ctx := apm.StartSpan(ses.Context(), "StatisticsPeriodFinished", "app")
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

	log.Println("period completed grades")
	pg, err := store.Store().SubjectsPeriodGradeFinished(ctx, schoolIds, periodNumber)
	if err != nil {
		return StatisticsResponse{}, err
	}

	resRows := []*StatisticsRow{}
	for _, schoolItem := range sl {
		// form StatisticsRow
		rowItem := StatisticsRow{}
		resRows = append(resRows, &rowItem)
		rowItem.FromSchool(schoolItem)

		// place all by keys, calc total users
		// calc Data percent and place
		// Fetch journal percent and place
		rowItem.Values = []StatisticsCell{}
		rowItem.Values = append(rowItem.Values, StatisticsCell(*schoolItem.Name))
		weekHours := uint(0)
		total := 0
		totalFinished := 0
		for _, p := range pg {
			subjectID := p.Subject.ID
			if p.Subject.ParentId != nil {
				subjectID = *p.Subject.ParentId
			}
			if p.Subject.SchoolId == schoolItem.ID && subjectID == p.Subject.ID && p.Subject.WeekHours != nil {
				weekHours += *p.Subject.WeekHours
				total++
				if p.StudentsCount <= p.FinishedCount {
					totalFinished++
				}
			}
		}
		rowItem.Values = append(rowItem.Values, StatisticsCell(strconv.Itoa(total)))
		rowItem.Values = append(rowItem.Values, StatisticsCell(strconv.Itoa(total-totalFinished)))
		rowItem.Values = append(rowItem.Values, StatisticsCell(strconv.Itoa(totalFinished)))
		percent := uint(0)
		if total > 0 {
			percent = uint(totalFinished * 100 / total)
		}
		rowItem.Percent = &percent
	}
	resRowsNoPtr := []StatisticsRow{}
	for _, v := range resRows {
		resRowsNoPtr = append(resRowsNoPtr, *v)
	}

	// insights
	totalSchoolFinished := 0
	totalSchool := len(resRowsNoPtr)
	totalRegion := []string{}
	for _, v := range resRowsNoPtr {
		if v.Region == v.School {
			continue
		}
		if n, _ := strconv.Atoi(string(v.Values[2])); n < 1 {
			totalSchoolFinished++
		}
		if !slices.Contains(totalRegion, v.Region) {
			totalRegion = append(totalRegion, v.Region)
		}
	}

	totals := []StatisticsTotal{
		StatisticsTotal{
			Title: "Jemi mekdep sany",
			Value: strconv.Itoa(totalSchool),
			Type:  "number",
		},
		StatisticsTotal{
			Title: "Doly jemlän mekdepler",
			Value: strconv.Itoa(totalSchoolFinished),
			Type:  "number",
		},
		StatisticsTotal{
			Title: "Jemi etrap sany",
			Value: strconv.Itoa(len(totalRegion)),
			Type:  "number",
		},
	}

	res := StatisticsResponse{
		Headers: []StatisticsHeader{"Mekdep",
			"Jemi žurnallar", "Jemlenenmedik", "Jemlenen"},
		Rows:   resRowsNoPtr,
		Totals: totals,
	}

	return res, nil
}

func (a App) StatisticsPeriodFinishedByTeacher(ses *utils.Session, schoolId string, periodNumber int) (StatisticsResponse, error) {
	sp, ctx := apm.StartSpan(ses.Context(), "StatisticsPeriodFinished", "app")
	ses.SetContext(ctx)
	defer sp.End()

	school, err := store.Store().SchoolsFindById(ses.Context(), schoolId)
	if err != nil {
		return StatisticsResponse{}, err
	}

	args := models.UserFilterRequest{
		SchoolId: &schoolId,
		Roles:    &[]string{string(models.RoleTeacher)},
	}
	args.Limit = new(int)
	*args.Limit = 500
	teachers, _, err := store.Store().UsersFindBy(ses.Context(), args)
	if err != nil {
		return StatisticsResponse{}, err
	}

	log.Println("period completed grades")
	periodGrades, err := store.Store().SubjectsPeriodGradeFinished(ctx, []string{schoolId}, periodNumber)
	if err != nil {
		return StatisticsResponse{}, err
	}

	resRows := []*StatisticsRow{}
	for _, teacherItem := range teachers {
		// form StatisticsRow
		rowItem := StatisticsRow{}
		resRows = append(resRows, &rowItem)
		rowItem.FromSchool(school)

		// place all by keys, calc total users
		// calc Data percent and place
		// Fetch journal percent and place
		rowItem.Values = []StatisticsCell{}
		rowItem.Values = append(rowItem.Values, StatisticsCell(teacherItem.FullName()))
		fullName := teacherItem.FullName()
		rowItem.User = &fullName
		rowItem.UserId = &teacherItem.ID
		weekHours := uint(0)
		total := 0
		totalFinished := 0
		for _, p := range periodGrades {
			subjectID := p.Subject.ID
			if p.Subject.ParentId != nil {
				subjectID = *p.Subject.ParentId
			}
			if subjectID == p.Subject.ID && (p.Subject.TeacherId != nil && *p.Subject.TeacherId == teacherItem.ID ||
				p.Subject.SecondTeacherId != nil && *p.Subject.SecondTeacherId == teacherItem.ID) &&
				p.Subject.WeekHours != nil {
				weekHours += *p.Subject.WeekHours
				total++
				if p.StudentsCount <= p.FinishedCount {
					totalFinished++
				}
			}
		}
		rowItem.Values = append(rowItem.Values, StatisticsCell(strconv.Itoa(int(weekHours))))
		rowItem.Values = append(rowItem.Values, StatisticsCell(strconv.Itoa(total)))
		rowItem.Values = append(rowItem.Values, StatisticsCell(strconv.Itoa(total-totalFinished)))
		rowItem.Values = append(rowItem.Values, StatisticsCell(strconv.Itoa(totalFinished)))
		percent := uint(0)
		if total > 0 {
			percent = uint(totalFinished * 100 / total)
		}
		rowItem.Percent = &percent

	}
	resRowsNoPtr := []StatisticsRow{}
	for _, v := range resRows {
		resRowsNoPtr = append(resRowsNoPtr, *v)
	}

	// insights
	totalTeachers := len(resRowsNoPtr)
	totalTeachersFinished := 0
	for _, v := range resRowsNoPtr {
		if v.Region == v.School {
			continue
		}
		if n, _ := strconv.Atoi(string(v.Values[3])); n < 1 {
			totalTeachersFinished++
		}
	}

	totals := []StatisticsTotal{
		StatisticsTotal{
			Title: "Jemi mugallym sany",
			Value: strconv.Itoa(totalTeachers),
			Type:  "number",
		},
		StatisticsTotal{
			Title: "Doly jemlänler",
			Value: strconv.Itoa(totalTeachersFinished),
			Type:  "number",
		},
		StatisticsTotal{
			Title: "Doly jemlemedikler",
			Value: strconv.Itoa(totalTeachers - totalTeachersFinished),
			Type:  "number",
		},
	}

	res := StatisticsResponse{
		Headers: []StatisticsHeader{"Mugallym",
			"Hepdelik sagat", "Ders ýük sany (žurnal)", "Jemlenenmedikler", "Jemlenenler"},
		Rows:   resRowsNoPtr,
		Totals: totals,
	}

	return res, nil
}

func (a App) StatisticsParentsCached(ses *utils.Session, schoolIds []string) (StatisticsResponse, error) {
	schoolIdsStr, _ := json.Marshal(schoolIds)
	k := "StatisticsParentsCached_" + string(schoolIdsStr)
	if v, ok := a.cache.Get(k); ok {
		return v.(StatisticsResponse), nil
	} else {
		v, err := a.StatisticsParents(ses, schoolIds)
		if err != nil {
			return StatisticsResponse{}, err
		}
		a.cache.Set(k, v, time.Minute*60)
		return v, nil
	}
}

func (a App) StatisticsParents(ses *utils.Session, schoolIds []string) (StatisticsResponse, error) {
	sp, ctx := apm.StartSpan(ses.Context(), "StatisticsAttendance", "app")
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

	// get all current period and finish

	log.Println("current period")

	log.Println("period completed grades")
	usersCount, err := store.Store().UsersLoadCountByClassroom(ses.Context(), schoolIds)
	if err != nil {
		return StatisticsResponse{}, err
	}

	resRows := []*StatisticsRow{}
	for _, schoolItem := range sl {
		// form StatisticsRow
		rowItem := StatisticsRow{}
		rowItem.FromSchool(schoolItem)
		percent := uint(100)
		classroomCount := 0
		studentsCount := 0
		parentsCount := 0
		parentsOnline := 0
		parentsOnlinePercent := 0
		for _, uc := range usersCount {
			if uc.SchoolId == schoolItem.ID {
				classroomCount++
				studentsCount += uc.StudentsCount
				parentsCount += uc.ParentsCount
				parentsOnline += uc.ParentsOnlineCount
			}
		}
		if studentsCount > 0 {
			percent = uint(parentsCount * 100 / (studentsCount * 2))
		}
		if parentsCount > 0 {
			parentsOnlinePercent = (parentsOnline * 100 / parentsCount)
		}
		rowItem.Percent = &percent

		// place all by keys, calc total users
		// calc Data percent and place
		// Fetch journal percent and place
		rowItem.Values = []StatisticsCell{}
		rowItem.Values = append(rowItem.Values, StatisticsCell(*schoolItem.Name))
		rowItem.Values = append(rowItem.Values, StatisticsCell(strconv.Itoa(classroomCount)))
		rowItem.Values = append(rowItem.Values, StatisticsCell(strconv.Itoa(studentsCount)))
		rowItem.Values = append(rowItem.Values, StatisticsCell(strconv.Itoa(parentsCount)))
		rowItem.Values = append(rowItem.Values, StatisticsCell(strconv.Itoa(int(percent))))
		rowItem.Values = append(rowItem.Values, StatisticsCell(strconv.Itoa(parentsOnlinePercent)))
		resRows = append(resRows, &rowItem)
	}
	resRowsNoPtr := []StatisticsRow{}
	for _, v := range resRows {
		resRowsNoPtr = append(resRowsNoPtr, *v)
	}

	// insights
	totalSchool := len(resRowsNoPtr)
	totalRegion := []string{}
	for _, v := range resRowsNoPtr {
		if !slices.Contains(totalRegion, v.Region) {
			totalRegion = append(totalRegion, v.Region)
		}
	}
	totalEq100 := []string{}
	for _, v := range resRowsNoPtr {
		if v.Percent != nil && *v.Percent >= 99 {
			totalEq100 = append(totalEq100, *v.SchoolCode)
		}
	}
	totalMore90 := []string{}
	for _, v := range resRowsNoPtr {
		if v.Percent != nil && *v.Percent >= 90 {
			totalMore90 = append(totalMore90, *v.SchoolCode)
		}
	}
	totalLess80 := []string{}
	for _, v := range resRowsNoPtr {
		if v.Percent != nil && *v.Percent < 60 {
			totalLess80 = append(totalLess80, *v.SchoolCode)
		}
	}
	totalLess60 := []string{}
	for _, v := range resRowsNoPtr {
		if v.Percent != nil && *v.Percent < 60 {
			totalLess60 = append(totalLess60, *v.SchoolCode)
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
			Title: "Ata-eneler doly 100 mekdepler",
			Value: strconv.Itoa(len(totalEq100)),
			Type:  "number",
		},
		{
			Title: "Ata-eneler doly 90 kop mekdepler",
			Value: strconv.Itoa(len(totalMore90)),
			Type:  "number",
		},
		{
			Title: "Ata-eneler doly 80 az mekdepler",
			Value: strconv.Itoa(len(totalLess80)),
			Type:  "number",
		},
		{
			Title: "Ata-eneler doly 60 az mekdepler",
			Value: strconv.Itoa(len(totalLess60)),
			Type:  "number",
		},
	}

	res := StatisticsResponse{
		Headers: []StatisticsHeader{"Mekdep", "Synp sany",
			"Jemi okuwçy", "Jemi ata-ene", "Ata-ene doly %", "Ata-ene işjeň %"},
		Rows:   resRowsNoPtr,
		Totals: totals,
	}

	return res, nil
}

func (a App) StatisticsParentsBySchoolCached(ses *utils.Session, schoolId string) (StatisticsResponse, error) {
	schoolIdsStr, _ := json.Marshal(schoolId)
	k := "StatisticsParentsBySchoolCached_" + string(schoolIdsStr)
	if v, ok := a.cache.Get(k); ok {
		return v.(StatisticsResponse), nil
	} else {
		v, err := a.StatisticsParentsBySchool(ses, schoolId)
		if err != nil {
			return StatisticsResponse{}, err
		}
		a.cache.Set(k, v, time.Minute*10)
		return v, nil
	}
}

func (a App) StatisticsParentsBySchool(ses *utils.Session, schoolId string) (StatisticsResponse, error) {
	sp, ctx := apm.StartSpan(ses.Context(), "StatisticsAttendance", "app")
	ses.SetContext(ctx)
	defer sp.End()

	school, err := store.Store().SchoolsFindById(ses.Context(), schoolId)
	if err != nil {
		return StatisticsResponse{}, err
	}

	err = store.Store().SchoolsLoadRelations(ses.Context(), &[]*models.School{school})
	if err != nil {
		return StatisticsResponse{}, err
	}

	log.Println("SchoolsLoadRelations")

	// get all current period and finish

	log.Println("current period")

	log.Println("period completed grades")
	usersCount, err := store.Store().UsersLoadCountByClassroom(ses.Context(), []string{school.ID})
	if err != nil {
		return StatisticsResponse{}, err
	}

	resRows := []*StatisticsRow{}
	for _, uc := range usersCount {
		// form StatisticsRow
		rowItem := StatisticsRow{}
		rowItem.FromSchool(school)
		percent := uint(100)
		if uc.StudentsCount > 0 {
			percent = uint(uc.ParentsCount * 100 / (uc.StudentsCount * 2))
		}
		rowItem.Percent = &percent
		rowItem.Classroom = uc.Classroom.Name
		rowItem.ClassroomId = &uc.Classroom.ID

		// place all by keys, calc total users
		// calc Data percent and place
		// Fetch journal percent and place
		rowItem.Values = []StatisticsCell{}
		rowItem.Values = append(rowItem.Values, StatisticsCell(*uc.Classroom.Name))
		rowItem.Values = append(rowItem.Values, StatisticsCell(strconv.Itoa(uc.StudentsCount)))
		rowItem.Values = append(rowItem.Values, StatisticsCell(strconv.Itoa(uc.ParentsCount)))
		rowItem.Values = append(rowItem.Values, StatisticsCell(strconv.Itoa(int(percent))))
		rowItem.Values = append(rowItem.Values, StatisticsCell(strconv.Itoa(uc.ParentsOnlineCount)))
		resRows = append(resRows, &rowItem)
	}
	resRowsNoPtr := []StatisticsRow{}
	for _, v := range resRows {
		resRowsNoPtr = append(resRowsNoPtr, *v)
	}

	// insights
	totalSchool := len(resRowsNoPtr)
	totalEq100 := []string{}
	for _, v := range resRowsNoPtr {
		if v.Percent != nil && *v.Percent >= 99 {
			totalEq100 = append(totalEq100, *v.SchoolCode)
		}
	}
	totalMore90 := []string{}
	for _, v := range resRowsNoPtr {
		if v.Percent != nil && *v.Percent >= 90 {
			totalMore90 = append(totalMore90, *v.SchoolCode)
		}
	}
	totalLess80 := []string{}
	for _, v := range resRowsNoPtr {
		if v.Percent != nil && *v.Percent < 60 {
			totalLess80 = append(totalLess80, *v.SchoolCode)
		}
	}
	totalLess60 := []string{}
	for _, v := range resRowsNoPtr {
		if v.Percent != nil && *v.Percent < 60 {
			totalLess60 = append(totalLess60, *v.SchoolCode)
		}
	}

	totals := []StatisticsTotal{
		{
			Title: "Jemi synp sany",
			Value: strconv.Itoa(totalSchool),
			Type:  "number",
		},
		{
			Title: "Ata-eneler doly 100 mekdepler",
			Value: strconv.Itoa(len(totalEq100)),
			Type:  "number",
		},
		{
			Title: "Ata-eneler doly 90 kop mekdepler",
			Value: strconv.Itoa(len(totalMore90)),
			Type:  "number",
		},
		{
			Title: "Ata-eneler doly 80 az mekdepler",
			Value: strconv.Itoa(len(totalLess80)),
			Type:  "number",
		},
		{
			Title: "Ata-eneler doly 60 az mekdepler",
			Value: strconv.Itoa(len(totalLess60)),
			Type:  "number",
		},
	}

	res := StatisticsResponse{
		Headers: []StatisticsHeader{"Synp",
			"Jemi okuwçy", "Jemi ata-ene", "Ata-ene doly %", "Ata-ene işjeň"},
		Rows:   resRowsNoPtr,
		Totals: totals,
	}

	return res, nil
}

func (a App) StatisticsOnline(ses *utils.Session, schoolIds []string, startDate time.Time, endDate time.Time) (StatisticsResponse, error) {
	sp, ctx := apm.StartSpan(ses.Context(), "StatisticsOnline", "app")
	ses.SetContext(ctx)
	defer sp.End()
	args := models.SchoolFilterRequest{
		Uids: &schoolIds,
	}
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

	resRows := []*StatisticsRow{}
	for _, schoolItem := range sl {
		// form StatisticsRow
		rowItem := StatisticsRow{}
		resRows = append(resRows, &rowItem)
		rowItem.FromSchool(schoolItem)

		// place all by keys, calc total users
		// calc Data percent and place
		// Fetch journal percent and place
		rowItem.Values = []StatisticsCell{}
		rowItem.Values = append(rowItem.Values, StatisticsCell(*schoolItem.Name))

	}
	resRowsNoPtr := []StatisticsRow{}
	for _, v := range resRows {
		resRowsNoPtr = append(resRowsNoPtr, *v)
	}

	// insights
	totalSchool := len(resRowsNoPtr)
	totalRegion := []string{}
	for _, v := range resRowsNoPtr {
		if !slices.Contains(totalRegion, v.Region) {
			totalRegion = append(totalRegion, v.Region)
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
	}

	res := StatisticsResponse{
		Headers: []StatisticsHeader{"Mekdep", "Jemi ulanyjy",
			"Onlaýn ulanyjy sany, 24 sagat", "7 gün", "3 aý"},
		Rows:   resRowsNoPtr,
		Totals: totals,
	}

	return res, nil
}

func (a App) StatisticsStudentsCached(ses *utils.Session, schoolIds []string) (StatisticsResponse, error) {
	schoolIdsStr, _ := json.Marshal(schoolIds)
	k := "StatisticsStudentsCached_" + string(schoolIdsStr)
	if v, ok := a.cache.Get(k); ok {
		return v.(StatisticsResponse), nil
	} else {
		v, err := a.StatisticsStudents(ses, schoolIds)
		if err != nil {
			return StatisticsResponse{}, err
		}
		a.cache.Set(k, v, time.Minute*60)
		return v, nil
	}
}

func (a App) StatisticsStudents(ses *utils.Session, schoolIds []string) (StatisticsResponse, error) {
	sp, ctx := apm.StartSpan(ses.Context(), "StatisticsStudents", "app")
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
	classroomStudentsCounts, err := store.Store().ClassroomStudentsCountBySchool(ctx)
	if err != nil {
		return StatisticsResponse{}, err
	}

	resRows := []*StatisticsRow{}
	for _, schoolItem := range sl {
		// form StatisticsRow
		rowItemCopy := StatisticsRow{
			School:     *schoolItem.Name,
			SchoolId:   schoolItem.ID,
			Region:     *schoolItem.Parent.Name,
			RegionCode: *schoolItem.Parent.Code,
		}
		rowItemAll := rowItemCopy
		rowItemAll.Values = []StatisticsCell{StatisticsCell(*schoolItem.Name + ", " + *schoolItem.Code), "Jemi"}
		rowItemGirls := rowItemCopy
		rowItemGirls.Values = []StatisticsCell{StatisticsCell(*schoolItem.Name + ", " + *schoolItem.Code), "Gyzlar"}
		rowItemBoys := rowItemCopy
		rowItemBoys.Values = []StatisticsCell{StatisticsCell(*schoolItem.Name + ", " + *schoolItem.Code), "Oglanlar"}
		rowItemGroups := rowItemCopy
		rowItemGroups.Values = []StatisticsCell{StatisticsCell(*schoolItem.Name + ", " + *schoolItem.Code), "Synp sany"}
		for _, count := range classroomStudentsCounts {
			if count.SchoolCode == *schoolItem.Code {
				boysSum := 0
				girlsSum := 0
				groupsSum := 0
				for _, i := range count.Counts {
					groupsSum += i[0]
					boysSum += i[1]
					girlsSum += i[2]
					rowItemAll.Values = append(rowItemAll.Values, StatisticsCell(strconv.Itoa(i[1]+i[2])))
					rowItemBoys.Values = append(rowItemBoys.Values, StatisticsCell(strconv.Itoa(i[1])))
					rowItemGirls.Values = append(rowItemGirls.Values, StatisticsCell(strconv.Itoa(i[2])))
					rowItemGroups.Values = append(rowItemGroups.Values, StatisticsCell(strconv.Itoa(i[0])))
				}
				rowItemAll.Values = append(rowItemAll.Values, StatisticsCell(strconv.Itoa(boysSum+girlsSum)))
				rowItemBoys.Values = append(rowItemBoys.Values, StatisticsCell(strconv.Itoa(boysSum)))
				rowItemGirls.Values = append(rowItemGirls.Values, StatisticsCell(strconv.Itoa(girlsSum)))
				rowItemGroups.Values = append(rowItemGroups.Values, StatisticsCell(strconv.Itoa(groupsSum)))
				break
			}
		}
		rowItem := []*StatisticsRow{
			&rowItemAll,
			&rowItemBoys,
			&rowItemGirls,
			&rowItemGroups,
		}

		// TODO:
		resRows = append(resRows, rowItem...)
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
	}

	res := StatisticsResponse{
		Headers: []StatisticsHeader{"Mekdep", "",
			"1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12",
			"Jemi"},
		Rows:   resRowsNoPtr,
		Totals: totals,
	}

	return res, nil
}
