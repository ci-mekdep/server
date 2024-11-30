package app

import (
	"encoding/json"
	"errors"
	"sort"
	"time"

	"github.com/mekdep/server/internal/api/utils"
	"github.com/mekdep/server/internal/models"
	"github.com/mekdep/server/internal/store"
	tools "github.com/mekdep/server/internal/utils"
	"go.elastic.co/apm/v2"
)

func (a *App) TeacherTimetable(ses *utils.Session, date time.Time) ([]models.TeacherTimetableResponse, error) {
	sp, ctx := apm.StartSpan(ses.Context(), "TeacherTimetable", "app")
	ses.SetContext(ctx)
	defer sp.End()
	res := []models.TeacherTimetableResponse{}
	user := ses.GetUser()
	period, _, err := periodsGetByDate(ses, date, *ses.GetSchoolId())
	defaultResponse := models.TeacherTimetableResponse{
		Shift:     models.ShiftResponse{},
		Timetable: models.TeacherTimetableValue{},
	}
	if period != nil {
		periodStart, periodEnd, err := period.Dates()
		if err != nil {
			return nil, err
		}
		isWithinPeriod := false
		if date.After(periodStart) && date.Before(periodEnd) {
			isWithinPeriod = true
		}
		if !isWithinPeriod {
			return []models.TeacherTimetableResponse{defaultResponse}, nil
		}
	}
	// fetch subjects
	subjects, _, err := store.Store().SubjectsListFilters(ses.Context(), &models.SubjectFilterRequest{
		TeacherIds: []string{user.ID},
	})
	if err != nil {
		return nil, err
	}
	err = store.Store().SubjectsLoadRelations(ses.Context(), &subjects, false)
	if err != nil {
		return nil, err
	}
	subjectIds := []string{}
	cIds := []string{}
	for _, s := range subjects {
		cIds = append(cIds, s.ClassroomId)
		subjectIds = append(subjectIds, s.ID)
	}
	// fetch timetables
	tf := models.TimetableFilterRequest{
		ClassroomIds: &cIds,
	}
	tf.Limit = new(int)
	*tf.Limit = 1000
	timetables, _, err := store.Store().TimetablesFindBy(ses.Context(), tf)
	if err != nil {
		return nil, err
	}
	shIds := []string{}
	for _, t := range timetables {
		if t.ShiftId != nil {
			shIds = append(shIds, *t.ShiftId)
		}
	}
	// fetch shifts
	shifts, _, err := store.Store().ShiftsFindBy(ses.Context(), models.ShiftFilterRequest{
		IDs: &shIds,
	})
	if err != nil {
		return nil, err
	}

	// make teacher table
	values := map[string]models.TeacherTimetableValue{}
	for _, t := range timetables {
		if t.Value == nil {
			continue
		}
		tv := [][]string{}
		json.Unmarshal([]byte(*t.Value), &tv)
		// find shift
		shv := models.ShiftValue{}
		shid := ""
		if t.ShiftId != nil {
			for _, s := range shifts {
				if *t.ShiftId == s.Id {
					if s.Value != nil {
						json.Unmarshal([]byte(*s.Value), &shv)
						shid = s.Id
					}
				}
			}
		}
		if shid == "" {
			tools.Logger.Error(errors.New("Shift not found for teacher"))
			continue
		}

		value := models.TeacherTimetableValue{}
		if v, ok := values[shid]; ok {
			value = v
		}

		// set schema by shift value
		day := date
		dayOffset := time.Monday - day.Weekday()
		day = day.AddDate(0, 0, int(dayOffset)-1)

		for dayNum, hours := range shv {
			if len(value) <= dayNum {
				value = append(value, []*models.TeacherTimetableItemResponse{})
			}
			day = day.AddDate(0, 0, 1)
			for hourNum, v := range hours {
				if len(value[dayNum]) <= hourNum {
					value[dayNum] = append(value[dayNum], &models.TeacherTimetableItemResponse{})
					value[dayNum][hourNum].ShiftTimes = v
					value[dayNum][hourNum].Date = day.Format(time.DateOnly)
				}
			}
		}
		// find teacher subjects
		for tDayNum, tHours := range tv {
			for tHourNum, tSubjectId := range tHours {
				if len(value) <= tDayNum || len(value[tDayNum]) <= tHourNum {
					break
				}
				for _, s := range subjects {
					if tSubjectId == s.ID || s.ParentId != nil && tSubjectId == *s.ParentId {
						sr := models.SubjectResponse{}
						sr.FromModel(s)
						if value[tDayNum][tHourNum].Subject == nil {
							value[tDayNum][tHourNum].Subject = &sr
						} else {
							// TODO: do something when conflict subject for teacher
						}
					}
				}
			}
		}
		values[shid] = value
	}

	// set subject percent
	startDate := date
	startDate = startDate.AddDate(0, 0, int(time.Monday)-int(date.Weekday()))
	endDate := startDate.AddDate(0, 0, 6)
	subjectPercentList, err := store.Store().SubjectsPercents(ses.Context(), subjectIds, startDate, endDate)
	for vKey, vv := range values {
		for vDayKey, vDay := range vv {
			for vHourKey, vHour := range vDay {
				for _, sp := range subjectPercentList {
					if vHour.Subject != nil && sp.SubjectId == vHour.Subject.ID && sp.LessonDate.Format(time.DateOnly) == vHour.Date {
						values[vKey][vDayKey][vHourKey].SubjectPercent = &models.DashboardSubjectsPercent{}
						*values[vKey][vDayKey][vHourKey].SubjectPercent = sp
					}
				}
			}
		}
	}

	// set shifts
	for shid, value := range values {
		filteredValue := models.TeacherTimetableValue{}
		for _, day := range value {
			nonEmptyHours := []*models.TeacherTimetableItemResponse{}
			for _, hour := range day {
				if len(hour.ShiftTimes) > 0 {
					nonEmptyHours = append(nonEmptyHours, hour)
				}
			}
			if len(nonEmptyHours) > 0 {
				filteredValue = append(filteredValue, nonEmptyHours)
			}
		}
		if len(filteredValue) == 0 {
			continue // Skip empty shifts
		}
		mshift := models.Shift{}
		for _, s := range shifts {
			if shid == s.Id {
				mshift = *s
			}
		}
		rshift := models.ShiftResponse{}
		rshift.FromModel(&mshift)
		res = append(res, models.TeacherTimetableResponse{
			Shift:     rshift,
			Timetable: filteredValue,
		})
	}
	// sort by shift
	sort.Slice(res, func(i, j int) bool {
		if res[i].Shift.Id == "" || res[j].Shift.Id == "" {
			return len(res[i].Timetable[0]) > len(res[j].Timetable[0])
		}
		return len(res[i].Shift.Name) > len(res[j].Shift.Name) || res[i].Shift.Id > res[j].Shift.Id
	})
	return res, err
}
