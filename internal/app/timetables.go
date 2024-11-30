package app

import (
	"context"
	"log"
	"math"
	"sort"
	"time"

	"github.com/mekdep/server/internal/api/utils"
	"github.com/mekdep/server/internal/models"
	"github.com/mekdep/server/internal/store"
	"go.elastic.co/apm/v2"
)

func (a App) TimetablesList(ses *utils.Session, data models.TimetableFilterRequest) ([]*models.TimetableResponse, int, error) {
	sp, ctx := apm.StartSpan(ses.Context(), "TimetablesList", "app")
	ses.SetContext(ctx)
	defer sp.End()
	if data.ID == nil {
		data.ID = new(string)
	}
	l, total, err := store.Store().TimetablesFindBy(ses.Context(), data)
	if err != nil {
		return nil, 0, err
	}
	err = store.Store().TimetablesLoadRelations(ses.Context(), &l)
	if err != nil {
		return nil, 0, err
	}
	res := []*models.TimetableResponse{}
	for _, m := range l {
		item := models.TimetableResponse{}
		item.FromModel(m)
		res = append(res, &item)
	}
	return res, total, nil
}

func TimetableDetail(ses *utils.Session, id string) (*models.TimetableResponse, error) {
	sp, ctx := apm.StartSpan(ses.Context(), "TimetableDetail", "app")
	ses.SetContext(ctx)
	defer sp.End()
	m, err := store.Store().TimetableFindById(ses.Context(), id)
	if err != nil {
		return nil, err
	}
	err = store.Store().TimetablesLoadRelations(ses.Context(), &[]*models.Timetable{m})
	if err != nil {
		return nil, err
	}
	res := &models.TimetableResponse{}
	res.FromModel(m)
	return res, nil
}

func (a App) TimetableCreate(ses *utils.Session, data models.TimetableRequest) (*models.TimetableResponse, error) {
	sp, ctx := apm.StartSpan(ses.Context(), "TimetableCreate", "app")
	ses.SetContext(ctx)
	defer sp.End()
	if data.ShiftId != nil && *data.ShiftId == "" {
		return nil, ErrInvalid.SetKey("shift_id")
	}
	m := &models.Timetable{}
	data.ToModel(m)
	var err error
	m, err = store.Store().TimetableCreate(ses.Context(), m)
	if err != nil {
		return nil, err
	}
	res := &models.TimetableResponse{}
	err = store.Store().TimetablesLoadRelations(ses.Context(), &[]*models.Timetable{m})
	if err != nil {
		return nil, err
	}
	res.FromModel(m)
	return res, nil
}

func (a App) TimetableUpdate(ses *utils.Session, u *models.User, data models.TimetableRequest) (*models.TimetableResponse, error) {
	sp, ctx := apm.StartSpan(ses.Context(), "TimetableUpdate", "app")
	ses.SetContext(ctx)
	defer sp.End()

	m := &models.Timetable{}
	var err error
	err = data.ToModel(m)
	if u != nil {
		m.UpdatedBy = &u.ID
	}
	if err != nil {
		return nil, err
	}
	m, err = store.Store().TimetableUpdate(ses.Context(), m)
	if err != nil {
		return nil, err
	}
	err = store.Store().TimetablesLoadRelations(ses.Context(), &[]*models.Timetable{m})
	if err != nil {
		return nil, err
	}
	res := &models.TimetableResponse{}
	res.FromModel(m)

	err = store.Store().TimetablesLoadRelations(ses.Context(), &[]*models.Timetable{m})
	if err != nil {
		return nil, err
	}
	desc := ""
	if m.Classroom != nil {
		desc += *m.Classroom.Name + " "
	} else {
		desc += "? "
	}
	if m.Shift != nil {
		desc += *m.Shift.Name + " "
	} else {
		desc += "? "
	}
	userLog(ses, models.UserLog{
		SchoolId:           ses.GetSchoolId(),
		SessionId:          ses.GetSessionId(),
		UserId:             u.ID,
		SubjectId:          &m.ID,
		Subject:            models.LogSubjectTimetable,
		SubjectDescription: &desc,
		SubjectAction:      models.LogActionUpdate,
		SubjectProperties:  data,
	})

	go TimetableUpdateValue(ses, *m, res.Value, data.IsThisWeek, false, false)
	return res, nil
}

func TimetableUpdateValue(ses *utils.Session, timetable models.Timetable, tValue models.TimetableValue, isThisWeek bool, isThisYear bool, disableLog bool) error {
	sp, ctx := apm.StartSpan(ses.Context(), "TimetableUpdateValue", "app")
	ses.SetContext(ctx)
	defer sp.End()

	ses.SetContext(context.Background())
	if !disableLog {
		log.Println("Worker timetable starter", isThisWeek)
	}
	err := lessonsSyncByTimetable(ses, timetable, tValue, isThisWeek, isThisYear, disableLog)
	if err == nil {
		if !disableLog {
			log.Println("Worker timetable finished for: " + timetable.ID)
		}
	} else {
		if !disableLog {
			log.Println("Worker timetable did not finish for: "+timetable.ID, "err ", err)
		}
	}
	return err
}

func lessonsSyncByTimetable(ses *utils.Session, timetable models.Timetable, tValue models.TimetableValue, isThisWeek bool, _ bool, disableLog bool) error {
	sp, ctx := apm.StartSpan(ses.Context(), "lessonsSyncByTimetable", "app")
	ses.SetContext(ctx)
	defer sp.End()
	// set startDate and endDate
	period, _, err := periodsGetByDate(ses, time.Now(), timetable.SchoolId)
	if err != nil {
		return err
	}
	if period == nil {
		return ErrRequired.SetKey("period")
	}
	periodStartDate, periodEndDate, err := period.Dates()
	if err != nil {
		return err
	}

	// get lessons
	date := time.Now().Truncate(time.Hour*24*2).AddDate(0, 0, 7)
	if isThisWeek {
		date = date.AddDate(0, 0, -7)
	}
	if date.Before(periodStartDate) {
		date = periodStartDate
	}
	date = date.AddDate(0, 0, int(date.Weekday()*-1))
	args := models.LessonFilterRequest{
		ClassroomId: &timetable.ClassroomId,
		DateRange:   &[]string{periodStartDate.Format(time.DateOnly), periodEndDate.Format(time.DateOnly)},
	}
	args.Limit = new(int)
	*args.Limit = 2000
	allLessons, _, err := store.Store().LessonsFindBy(ses.Context(), args)
	if err != nil {
		return err
	}

	if !disableLog {
		log.Println("Found len: ", len(allLessons), periodStartDate.Format(time.DateOnly), periodEndDate.Format(time.DateOnly))
	}

	lessonCountByWeek := map[int]int{}
	existLessons := map[int]map[string][]*models.Lesson{}
	for _, item := range allLessons {
		_, wn := item.Date.ISOWeek()
		lessonCountByWeek[wn]++

		if existLessons[wn] == nil {
			existLessons[wn] = map[string][]*models.Lesson{}
		}
		if existLessons[wn][item.SubjectId] == nil {
			existLessons[wn][item.SubjectId] = []*models.Lesson{}
		}
		existLessons[wn][item.SubjectId] = append(existLessons[wn][item.SubjectId], item)
	}

	// sync day by day
	periodEndDate = periodEndDate.AddDate(0, 0, 1)
	updateLessons := []models.Lesson{}
	createLessons := []models.Lesson{}

	// create pre lessons from start of period
	preDate := periodStartDate.AddDate(0, 0, -1)
	for preDate.Before(date) {
		preDate = preDate.AddDate(0, 0, 1)
		if isDateVacation(preDate, *period) || isDateHoliday(preDate) {
			continue
		}
		weekDay := int(preDate.Weekday()) - 1
		if weekDay == -1 {
			weekDay = 6
		}
		// check if week contains lessons, existLessons[weekNumber]
		_, weekNumber := preDate.ISOWeek()
		weekLessonsContains := false
		for _, lessons := range existLessons[weekNumber] {
			if len(lessons) > 0 {
				weekLessonsContains = true
			}
		}
		// not create for weeks with lessons
		if weekLessonsContains {
			continue
		}
		if len(tValue) > weekDay && len(tValue[weekDay]) > 0 {
			tDayValue := tValue[weekDay]
			err = lessonSyncDayItem(tDayValue, preDate, *period, true, &existLessons, nil, &createLessons)
			if err != nil {
				return err
			}
		}
	}

	loggedUserId := ""
	if ses.GetUser() != nil {
		loggedUserId = ses.GetUser().ID
	}
	logDescription := "timetable:" + timetable.ID
	startDate := date
	for date.Before(periodEndDate) {
		date = date.AddDate(0, 0, 1)
		// check is vacation or holiday
		if isDateVacation(date, *period) || isDateHoliday(date) {
			continue
		}
		// set weekday monday=0 and sunday=6
		weekDay := int(date.Weekday()) - 1
		if weekDay == -1 {
			weekDay = 6
		}
		// update lessons by timetable
		if len(tValue) > weekDay && len(tValue[weekDay]) > 0 {
			tDayValue := tValue[weekDay]
			// TODO: arg subjects
			err = lessonSyncDayItem(tDayValue, date, *period, false, &existLessons, &updateLessons, &createLessons)
			if err != nil {
				return err
			}
		}

		// log only end of week (full first week as example)
		if startDate.AddDate(0, 0, 6) == date {
			for _, v := range updateLessons {
				userLog(ses, models.UserLog{
					SchoolId:           ses.GetSchoolId(),
					SessionId:          ses.GetSessionId(),
					UserId:             loggedUserId,
					Subject:            models.LogSubjectLesson,
					SubjectId:          &v.ID,
					SubjectAction:      models.LogActionUpdate,
					SubjectProperties:  v,
					SubjectDescription: &logDescription,
				})
			}
			for _, v := range createLessons {
				userLog(ses, models.UserLog{
					SchoolId:           ses.GetSchoolId(),
					SessionId:          ses.GetSessionId(),
					UserId:             loggedUserId,
					Subject:            models.LogSubjectLesson,
					SubjectId:          &v.ID,
					SubjectAction:      models.LogActionUpdate,
					SubjectProperties:  v,
					SubjectDescription: &logDescription,
				})
			}
		}
	}
	// delete left leftLessons
	deleteIds := []string{}
	for wn, bySubjects := range existLessons {
		for _, items := range bySubjects {
			for _, item := range items {
				if item != nil {
					if item.Date.Compare(startDate) >= 0 {
						deleteIds = append(deleteIds, item.ID)
					}
				}
			}
		}
		if _, swn := startDate.ISOWeek(); wn == swn || wn == swn+1 {
			for _, items := range bySubjects {
				for _, item := range items {
					if item != nil {
						if item.Date.Compare(startDate) >= 0 {
							userLog(ses, models.UserLog{
								SchoolId:           ses.GetSchoolId(),
								SessionId:          ses.GetSessionId(),
								UserId:             loggedUserId,
								Subject:            models.LogSubjectLesson,
								SubjectId:          &item.ID,
								SubjectAction:      models.LogActionDelete,
								SubjectProperties:  item,
								SubjectDescription: &logDescription,
							})
						}
					}
				}
			}
		}
	}
	// perform all sql
	if !disableLog {
		log.Println("Deleted len: ", len(deleteIds), err)
	}
	err = store.Store().LessonsDeleteBatch(ses.Context(), deleteIds)
	if err != nil {
		return err
	}
	// var dt, dtt []byte
	// if len(deleteIds) > 0 {
	// 	log.Println("Deleted len: ", deleteIds)
	// }

	if !disableLog {
		log.Println("Created len: ", len(createLessons), err)
	}
	err = store.Store().LessonsCreateBatch(ses.Context(), createLessons)
	if err != nil {
		return err
	}
	// if len(createLessons) > 0 {
	// 	dt, _ = json.Marshal(createLessons[0])
	// 	dtt, _ = json.Marshal(createLessons[len(createLessons)-1])
	// 	log.Println("Created len: ", string(dt), string(dtt))
	// }

	if !disableLog {
		log.Println("Updated len: ", len(updateLessons), err)
	}
	err = store.Store().LessonsUpdateBatch(ses.Context(), updateLessons)
	if err != nil {
		return err
	}
	// if len(updateLessons) > 0 {
	// 	dt, _ = json.Marshal(updateLessons[0])
	// 	dtt, _ = json.Marshal(updateLessons[len(updateLessons)-1])
	// 	log.Println("Updated len: ", string(dt), string(dtt))
	// }

	return nil
}

func isDateHoliday(date time.Time) bool {
	return GetHolidayByDate(date) != ""
}

func GetHolidayByDate(date time.Time) string {
	for _, v := range models.DefaultHolidays {
		if (date.Month() == v.StartDate.Month() && date.Day() >= v.StartDate.Day()) &&
			(date.Month() == v.EndDate.Month() && date.Day() <= v.EndDate.Day()) {
			return v.Name
		}
	}
	return ""
}

func isDateVacation(date time.Time, p models.Period) bool {
	k, _ := p.GetKey(date, true)
	return k == 0
}

func isDateVacationBySchool(ses *utils.Session, date time.Time, schoolId string) (bool, error) {
	p, _, err := periodsGetByDate(ses, date, schoolId)
	if err != nil {
		return false, err
	}
	return isDateVacation(date, *p), nil
}

func lessonSyncDayItem(tDayValue []string, date time.Time, period models.Period, onlyCreate bool,
	existLessonsPtr *map[int]map[string][]*models.Lesson, updateLessons *[]models.Lesson, createLessons *[]models.Lesson) error {
	_, weekNumber := date.ISOWeek()
	periodKey, _ := periodGetKey(period, date)

	var err error
	existLessons := *existLessonsPtr
	for hourNumber, subjectId := range tDayValue {
		if subjectId == "" {
			continue
		}
		if existLessons[weekNumber] == nil {
			existLessons[weekNumber] = map[string][]*models.Lesson{}
		}
		var updateOrCreate *models.Lesson
		existLessons[weekNumber][subjectId], updateOrCreate, err =
			lessonSyncHour(existLessons[weekNumber][subjectId], date, hourNumber+1)
		if err != nil {
			return err
		}
		if updateOrCreate != nil {
			updateOrCreate.SchoolId = *period.SchoolId
			updateOrCreate.SubjectId = subjectId
			updateOrCreate.PeriodId = &period.ID
			updateOrCreate.PeriodKey = &periodKey
			if updateOrCreate.ID != "" {
				if !onlyCreate && updateLessons != nil {
					*updateLessons = append(*updateLessons, *updateOrCreate)
				}
			} else if createLessons != nil {
				*createLessons = append(*createLessons, *updateOrCreate)
			}

			// TODO: find subject by id
			// TODO: foreach children:
			// TODO: foreach lesson.children
			// TODO: update or create by lesson.sub_group_id
			// TODO: set fields same as updateOrCreate, set parent_id, sub_group_id
			// TODO: append all of children to createLessons or updateLessons (see example old code)
		}
	}
	*existLessonsPtr = existLessons
	return nil
}

func lessonSyncHour(lessons []*models.Lesson, date time.Time, hourNumber int) ([]*models.Lesson, *models.Lesson, error) {
	var updateOrCreate *models.Lesson
	// get lesson.Date nearest to Date
	if lessons == nil || len(lessons) < 1 || lessons[0] == nil {
		// if not found, create one
		updateOrCreate = (&models.Lesson{
			HourNumber: &hourNumber,
			Date:       date,
		})
	} else {
		var lesson models.Lesson
		// mark it moved: unset -> lessons[weekNumber][subjectId]
		lessonK := lessonFindCloseDate(date, hourNumber, lessons)
		lesson = *lessons[lessonK]
		lessons = append(lessons[:lessonK], lessons[lessonK+1:]...)
		// lesson, lessons = *lessons[0], lessons[1:]
		// set lesson.Date -> Date
		if lesson.Date.Format(time.DateOnly) != date.Format(time.DateOnly) ||
			lesson.HourNumber != nil && *lesson.HourNumber != hourNumber ||
			lesson.HourNumber == nil && hourNumber != 0 {
			lesson.Date = date
			lesson.HourNumber = &hourNumber
			updateOrCreate = &lesson
		}
	}
	// synced
	return lessons, updateOrCreate, nil
}

func lessonFindCloseDate(date time.Time, hourNumber int, l []*models.Lesson) int {
	unixDiff := float64(-1)
	activeK := -1
	sort.Slice(l, func(i, j int) bool {
		return l[i].ID > l[j].ID
	})
	for k, v := range l {
		if v.HourNumber == nil {
			v.HourNumber = new(int)
			*v.HourNumber = -100
		}
		df := float64(v.Date.Unix() - date.Unix())
		hdf := math.Abs(float64(*v.HourNumber-hourNumber))*100 + 1
		df = df * hdf
		if activeK != -1 && math.Abs(df) <= math.Abs(unixDiff) || activeK == -1 {
			unixDiff = df
			activeK = k
		}
	}
	return activeK
}

func (a App) TimetablesDelete(ses *utils.Session, ids []string) ([]*models.Timetable, error) {
	sp, ctx := apm.StartSpan(ses.Context(), "TimetablesDelete", "app")
	ses.SetContext(ctx)
	defer sp.End()
	l, err := store.Store().TimetablesFindByIds(ses.Context(), ids)
	if err != nil {
		return nil, err
	}
	if len(l) < 1 {
		return nil, ErrNotExists.SetKey("ids")
	}
	return store.Store().TimetablesDelete(ses.Context(), l)
}
