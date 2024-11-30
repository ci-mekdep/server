package cmd

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/mekdep/server/internal/app"
	"github.com/mekdep/server/internal/models"
	"github.com/mekdep/server/internal/store"
)

func SendReminderTeacherAllSchools(isPrincipal bool) error {
	schools, _, err := store.Store().SchoolsFindBy(context.Background(), models.SchoolFilterRequest{})
	if err != nil {
		return err
	}
	for _, school := range schools {
		SendReminderTeacherBySchool(*school, isPrincipal)
	}
	return nil
}

func SendReminderTeacherBySchool(school models.School, isPrincipal bool) error {
	date := time.Now()
	usersArg := models.UserFilterRequest{
		SchoolId: &school.ID,
		Role:     new(string),
	}
	subjectsArg := models.SubjectFilterRequest{
		SchoolId: &school.ID,
	}
	*usersArg.Role = string(models.RoleTeacher)
	usersArg.Limit = new(int)
	*usersArg.Limit = 200
	subjectsArg.Limit = new(int)
	*subjectsArg.Limit = 200

	teachers, _, err := store.Store().UsersFindBy(context.Background(), usersArg)
	if err != nil {
		return err
	}

	subjects, _, err := store.Store().SubjectsListFilters(context.Background(), &subjectsArg)
	if err != nil {
		return err
	}
	subjectIds := []string{}
	for _, v := range subjects {
		subjectIds = append(subjectIds, v.ID)
	}

	subjectPercents, err := store.Store().SubjectsPercents(context.Background(), subjectIds, date, date)
	if err != nil {
		return err
	}

	if !isPrincipal {
		for _, teacher := range teachers {
			sp := []models.DashboardSubjectsPercent{}
			for _, v := range subjects {
				if v.BelongsToTeacher(teacher) {
					for _, vv := range subjectPercents {
						if v.ID == vv.SubjectId {
							sp = append(sp, vv)
						}
					}
				}
			}
			SendReminderTeacher(date, teacher, sp)
		}

	} else {
		if school.AdminUid == nil {
			return errors.New("School " + *school.Code + " has no admin")
		}
		principal, err := store.Store().UsersFindById(context.Background(), (*school.AdminUid))
		if err != nil {
			return err
		}
		SendReminderPrincipal(principal, teachers, subjects, subjectPercents)
	}
	return nil
}

func SendReminderTeacher(date time.Time, teacher *models.User, subjectPercents []models.DashboardSubjectsPercent) error {
	msg := date.Format(time.DateOnly) + " senesinde " + strconv.Itoa(len(subjectPercents)) + ` sapak hasaba alyndy, olar:
	
	`
	noFullCount := 0
	for _, sp := range subjectPercents {
		if !sp.IsGradeFull {
			noFullCount++
			msg += sp.ClassroomName + ` ` + sp.SubjectName + ` - žurnal doly däl
`
		} else {
			msg += sp.ClassroomName + ` ` + sp.SubjectName + ` - žurnal doly 
`
		}
	}
	if noFullCount > 0 {
		msg += `
` +
			strconv.Itoa(noFullCount) + ` sany doly däl žurnallary girizmek: ` + app.AppOpenLink
		phone, err := teacher.FormattedPhone()
		if err != nil {
			return nil
		}
		app.SendSMS([]string{phone}, msg, models.SmsTypeReminder)
	}
	return nil
}

func SendReminderPrincipal(principal *models.User, teachers []*models.User, subjects []*models.Subject, subjectPercents []models.DashboardSubjectsPercent) error {
	msg := "Hormatly Mekdep müdiri! Jemi " + strconv.Itoa(len(teachers)) + ` mugallym hasaba alynyp, olaryň žurnal dolduryş hasabaty:

`
	noFullTeacherCount := 0
	for _, teacher := range teachers {
		sp := []models.DashboardSubjectsPercent{}
		noFullCount := 0
		for _, v := range subjects {
			if v.BelongsToTeacher(teacher) {
				for _, vv := range subjectPercents {
					if v.ID == vv.SubjectId {
						sp = append(sp, vv)
						if !vv.IsGradeFull {
							noFullCount++
						}
					}
				}
			}
		}
		if noFullCount > 0 {
			noFullTeacherCount++
			msg += teacher.FullName() + ` jemi: ` + strconv.Itoa(len(sp)) + ` - doly däl: ` + strconv.Itoa(noFullCount) + `
`
		}
	}
	if noFullTeacherCount > 0 {
		msg += `
Giňişleýin ` + strconv.Itoa(noFullTeacherCount) + ` doldurmadyklary görmek: ` + app.AppOpenLink
		phone, err := principal.FormattedPhone()
		if err != nil {
			return nil
		}
		app.SendSMS([]string{phone}, msg, models.SmsTypeReminder)
	}
	return nil
}
