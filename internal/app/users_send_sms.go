package app

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	apiutils "github.com/mekdep/server/internal/api/utils"
	"github.com/mekdep/server/internal/models"
	"github.com/mekdep/server/internal/store"
	"github.com/mekdep/server/internal/utils"
	"go.elastic.co/apm/v2"
)

const AppOpenLink = `mekdep.edu.tm/redirect/app`

// Caganyz Meret eMekdep programmada "Goreldeli" nyrhnama birikdi. Gutarmagyna 30 gun galdy.
// Elektron gundelik, Jemleyji bahalar, SMS habarlar, Gyzyklanma analitika we basgalar.
// Peydalanmak:
func SendWelcomeSms(child *models.User, parents []*models.User, expiresAt time.Time) error {
	if child.Classrooms == nil {
		return ErrRequired.SetKey("classroom_id")
	}
	leftDays := (expiresAt.Unix() - time.Now().Unix()) / (60 * 60 * 24)
	tariffName := ""
	for _, v := range models.DefaultTariff {
		if v.Code == models.PaymentTariffType(*child.Classrooms[0].TariffType) {
			tariffName = v.Name
		}
	}
	msg := `Caganyz ` + *child.FirstName + ` eMekdep programmada "` + tariffName + `" nyrhnama birikdi. Gutarmagyna ` + strconv.Itoa(int(leftDays)) + ` gun galdy.
Elektron gundelik, Jemleyji bahalar, SMS habarlar, Gyzyklanma analitika we basgalar.
Peydalanmak: ` + AppOpenLink

	var err error
	phones := []string{}
	for _, v := range parents {
		if ph, err := v.FormattedPhone(); err == nil && ph != "" {
			phones = append(phones, ph)
		}
	}
	for _, v := range phones {
		err = SendSMS([]string{v}, LettersRemoveTurkmen(msg), models.SmsTypeDaily)
	}
	return err
}

// Caganyz Meret eMekdep programmada nyrhnamasynyň gutarmagyna 7 gun galdy. Dowam etmek üçin töleg etmeli.
func SendTariffEndsSmsAll(ses *apiutils.Session, isLate bool, today time.Time, student *models.User, parents []*models.User, daysLeft string) error {
	sp, ctx := apm.StartSpan(ses.Context(), "SendDailySms", "app")
	ses.SetContext(ctx)
	defer sp.End()
	if len(student.Classrooms) < 1 {
		return ErrNotSet.SetKey("classroom_id").SetComment("ID: " + student.ID)
	}
	childName := *student.FirstName

	smsText := "Çagaňyz " + childName + " eMekdep programmada nyrhnamasynyň gutarmagyna " + daysLeft + " gün galdy! Programmada töleg edip uzaltmagy unutmaň!"
	phones := []string{}

	for _, parent := range parents {
		if ph, err := parent.FormattedPhone(); err == nil && ph != "" {
			phones = append(phones, ph)
		}
	}

	utils.LoggerDesc("SendExpirationReminderSms").Info(phones, smsText)
	for _, phone := range phones {
		err := SendSMS([]string{phone}, LettersRemoveTurkmen(smsText), models.SmsTypeReminder)
		if err != nil {
			utils.LoggerDesc("SendExpirationReminderSms").Error(err)
		}
	}

	return nil
}

func SendMessageUnreadReminder(ses *apiutils.Session, count int, user *models.User) (err error) {
	title := "Size " + strconv.Itoa(count) + " sany täze hat geldi"
	body := ""
	// add payload type of push, and id
	sendNotificationPush(models.Notifications{
		Title:   &title,
		Content: &body,
		UserIds: []string{user.ID},
	}, []string{user.ID}, PushTypeChat, "")
	return nil
}

// Caganyz Meret
func SendDailySms(ses *apiutils.Session, isLate bool, today time.Time, student *models.User, parents []*models.User) error {
	sp, ctx := apm.StartSpan(ses.Context(), "SendDailySms", "app")
	ses.SetContext(ctx)
	defer sp.End()
	if len(student.Classrooms) < 1 {
		return ErrNotSet.SetKey("classroom_id").SetComment("ID: " + student.ID)
	}
	classroomId := student.Classrooms[0].ClassroomId
	childName := *student.FirstName
	lessonList, err := SendDailyGetLessons(ses, today, classroomId, student.ID)
	if err != nil {
		return err
	}
	if len(lessonList) < 1 {
		return nil
	}
	endShiftTime, err := SendDailyGetShiftTime(ses, classroomId)
	if err != nil {
		return err
	}
	// disable when isLate and time is before 13:50
	if isLate && endShiftTime <= "13:50" {
		return nil
	}
	if !isLate && endShiftTime > "13:50" {
		return nil
	}
	subjectNames, err := SendDailyGetSubjects(classroomId)
	if err != nil {
		return err
	}
	gradeList, err := SendDailyGetGrades(ses, lessonList, student.ID)
	if err != nil {
		return err
	}
	absentList, err := SendDailyGetAbsents(ses, lessonList, student.ID)
	if err != nil {
		return err
	}
	if len(gradeList) == 0 && len(absentList) == 0 {
		return nil
	}

	smsTextSubGrade := "Elektron gündeliginde %s baha aldy. "
	smsTextSubAbsent := "%s sapagyna gatnaşmady. "
	smsTextSub := ""
	// message grades
	tmp := ""
	for _, v := range gradeList {
		tmp += subjectNames[v.Lesson.SubjectId] + " " + v.ValueString() + ", "
	}
	if tmp != "" {
		smsTextSub += fmt.Sprintf(smsTextSubGrade, strings.Trim(tmp, ", "))
	}
	// message absents
	tmp = ""
	for _, v := range absentList {
		tmp += subjectNames[v.Lesson.SubjectId] + ", "
	}
	if tmp != "" {
		smsTextSub += fmt.Sprintf(smsTextSubAbsent, strings.Trim(tmp, ", "))
	}
	// message all
	smsText := "Çagaňyz " + childName + " " + smsTextSub + "Giňişleýin " + AppOpenLink
	phones := []string{}
	for _, p := range parents {
		if ph, err := p.FormattedPhone(); err == nil && ph != "" {
			phones = append(phones, ph)
		}
	}
	utils.LoggerDesc("SendSmsItem").Info(phones, smsText)
	for _, v := range phones {
		err = SendSMS([]string{v}, LettersRemoveTurkmen(smsText), models.SmsTypeDaily)
	}
	return err
}

func SendDailyGetLessons(ses *apiutils.Session, today time.Time, classroomId string, studentId string) ([]*models.Lesson, error) {
	sp, ctx := apm.StartSpan(ses.Context(), "SendDailyGetLessons", "app")
	ses.SetContext(ctx)
	defer sp.End()
	ll, _, err := store.Store().LessonsFindBy(ses.Context(), models.LessonFilterRequest{
		ClassroomId: &classroomId,
		Date:        &today,
	})
	if err != nil {
		return nil, err
	}
	return ll, nil
}

func SendDailyGetAbsents(ses *apiutils.Session, lessonList []*models.Lesson, studentId string) ([]*models.Absent, error) {
	sp, ctx := apm.StartSpan(ses.Context(), "SendDailyGetAbsents", "app")
	ses.SetContext(ctx)
	defer sp.End()
	lessonIds := []string{}
	for _, v := range lessonList {
		lessonIds = append(lessonIds, v.ID)
	}
	aa, _, err := store.Store().AbsentsFindBy(ses.Context(), models.AbsentFilterRequest{
		StudentId: &studentId,
		LessonIds: &lessonIds,
	})
	if err != nil {
		return nil, err
	}
	for k, v := range aa {
		for _, vv := range lessonList {
			if v.LessonId == vv.ID {
				aa[k].Lesson = vv
			}
		}
	}
	return aa, nil
}

func SendDailyGetGrades(ses *apiutils.Session, lessonList []*models.Lesson, studentId string) ([]*models.Grade, error) {
	sp, ctx := apm.StartSpan(ses.Context(), "SendDailyGetGrades", "app")
	ses.SetContext(ctx)
	defer sp.End()
	lessonIds := []string{}
	for _, v := range lessonList {
		lessonIds = append(lessonIds, v.ID)
	}
	gg, _, err := store.Store().GradesFindBy(ses.Context(), models.GradeFilterRequest{
		StudentId: &studentId,
		LessonIds: &lessonIds,
	})
	if err != nil {
		return nil, err
	}
	for k, v := range gg {
		for _, vv := range lessonList {
			if v.LessonId == vv.ID {
				gg[k].Lesson = new(models.Lesson)
				gg[k].Lesson = vv
			}
		}
	}
	return gg, nil
}

func SendDailyGetSubjects(classroomId string) (map[string]string, error) {
	ss := map[string]string{}
	if v, ok := CacheSubjects[classroomId]; ok {
		ss = v
	} else {
		subjectsList, _, err := store.Store().SubjectsListFilters(context.Background(), &models.SubjectFilterRequest{
			ClassroomIds: []string{classroomId},
		})
		if err != nil {
			return ss, err
		}
		for _, v := range subjectsList {
			ss[v.ID] = *v.Name
		}
	}
	return ss, nil
}

func SendDailyGetShiftTime(ses *apiutils.Session, classroomId string) (string, error) {
	sp, ctx := apm.StartSpan(ses.Context(), "SendDailyGetShiftTime", "app")
	ses.SetContext(ctx)
	defer sp.End()
	today := time.Now()
	endShiftTime := ""
	if v, ok := CacheShiftTimes[classroomId]; ok {
		endShiftTime = v
	} else {
		endShiftNumber := 0

		tt, _, err := store.Store().TimetablesFindBy(ses.Context(), models.TimetableFilterRequest{
			ClassroomIds: &[]string{classroomId},
		})
		if err != nil {
			return endShiftTime, err
		}
		if len(tt) < 1 {
			return endShiftTime, ErrNotSet.SetKey("timetable_id")
		}
		ttR := models.TimetableResponse{}
		ttR.FromModel(tt[0])
		if v := ttR.Value[int(today.Weekday())-1]; v != nil {
			endShiftNumber = len(v)
		} else {
			return endShiftTime, ErrNotSet.SetKey("timetable.value " + today.Weekday().String())
		}

		sh, _, err := store.Store().ShiftsFindBy(ses.Context(), models.ShiftFilterRequest{
			TimetableId: &ttR.ID,
		})
		if err != nil {
			return endShiftTime, err
		}
		if len(sh) < 1 {
			return endShiftTime, ErrNotSet.SetKey("shift_id")
		}
		shR := models.ShiftResponse{}
		shR.FromModel(sh[0])
		if v := shR.Value[int(today.Weekday())-1]; v != nil {
			if len(v) > endShiftNumber-1 && len(v[endShiftNumber-1]) > 1 {
				endShiftTime = v[endShiftNumber-1][1]
			} else {
				return endShiftTime, ErrNotSet.SetKey("shift_id.value " + today.Weekday().String())
			}
		} else {
			return endShiftTime, ErrNotSet.SetKey("shift_id.value " + today.Weekday().String())
		}
	}
	return endShiftTime, nil
}

func LettersRemoveTurkmen(msg string) string {
	m := map[string]string{
		"ä": "a",
		"ň": "n",
		"ž": "z",
		"ü": "u",
		"ç": "c",
		"ý": "y",
		"ş": "s",
		"ö": "o",
		"Ä": "A",
		"Ň": "N",
		"Ž": "Z",
		"Ü": "U",
		"Ç": "C",
		"Ý": "Y",
		"Ş": "S",
		"Ö": "O",
	}
	for old, new := range m {
		msg = strings.ReplaceAll(msg, old, new)
	}
	return msg
}

var CacheShiftTimes map[string]string = map[string]string{}
var CacheSubjects map[string]map[string]string = map[string]map[string]string{}
