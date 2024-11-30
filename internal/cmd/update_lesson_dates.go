package cmd

import (
	"log"
	"os"
	"strconv"

	"github.com/mekdep/server/internal/api/utils"
	"github.com/mekdep/server/internal/app"
	"github.com/mekdep/server/internal/models"
	"github.com/mekdep/server/internal/store"
	"github.com/spf13/cobra"
)

func UpdateLessonDatesCmd() *cobra.Command {
	return &cobra.Command{
		Use: "update-lesson-dates",
		Run: UpdateLessonDatesRun,
	}
}
func UpdateLessonDatesRun(cmd *cobra.Command, args []string) {
	args = append(args, "", "", "")
	timeLvl, _ := strconv.Atoi(args[0])
	schoolCode := (args[1])
	nextSchoolCode := (args[2])
	err := UpdateLessonDates(&utils.Session{}, timeLvl, schoolCode, nextSchoolCode)
	if err != nil {
		log.Fatalln(err)
	}
	os.Exit(0)
}
func UpdateLessonDates(ses *utils.Session, timeLvl int, schoolCode string, nextSchoolCode string) error {
	f := false
	t := true
	scarg := models.SchoolFilterRequest{
		IsSecondarySchool: &t,
		IsParent:          &f,
	}
	if schoolCode != "" && schoolCode != "0" {
		scarg.Code = &schoolCode
	}
	scarg.Limit = new(int)
	*scarg.Limit = 500
	schoolList, _, err := store.Store().SchoolsFindBy(ses.Context(), scarg)
	if err != nil {
		return err
	}
	log.Println("# Found schools: ", len(schoolList))
	isContinue := false
	if nextSchoolCode != "" && nextSchoolCode != "0" {
		isContinue = true
	}
	isCurrentWeek := timeLvl == 1
	log.Println("Is current week", isCurrentWeek)
	for _, s := range schoolList {
		if isContinue {
			if *s.Code == nextSchoolCode {
				isContinue = false
			} else {
				continue
			}
		}
		log.Println("#", *s.Name)
		targ := models.TimetableFilterRequest{
			SchoolId: &s.ID,
		}
		targ.Limit = new(int)
		*targ.Limit = 5000
		timetableList, _, err := store.Store().TimetablesFindBy(ses.Context(), targ)
		if err != nil {
			return err
		}
		log.Println("Found", len(timetableList), "updating...")
		for _, v := range timetableList {
			vr := models.TimetableResponse{}
			vr.FromModel(v)
			err := app.TimetableUpdateValue(ses, *v, vr.Value, isCurrentWeek, true, true)
			if err != nil {
				log.Println("err", err)
			}
		}
	}
	log.Println("Success.")

	return nil
}
