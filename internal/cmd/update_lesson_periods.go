package cmd

import (
	"context"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/mekdep/server/internal/api/utils"
	"github.com/mekdep/server/internal/models"
	"github.com/mekdep/server/internal/store"
	"github.com/spf13/cobra"
)

func UpdateLessonPeriodsCmd() *cobra.Command {
	return &cobra.Command{
		Use: "update-lesson-periods",
		Run: UpdateLessonPeriodsRun,
	}
}
func UpdateLessonPeriodsRun(cmd *cobra.Command, args []string) {
	args = append(args, " ", " ")
	timeLvl, _ := strconv.Atoi(args[0])
	schoolCode := (args[1])
	err := UpdateLessonPeriods(&utils.Session{}, timeLvl, schoolCode)
	if err != nil {
		log.Fatalln(err)
	}
	os.Exit(0)
}
func UpdateLessonPeriods(ses *utils.Session, timeLvl int, schoolCode string) error {
	startDate := time.Now()
	startDate = startDate.AddDate(0, 0, 6)
	startDate = startDate.AddDate(0, 0, 1-int(startDate.Weekday()))

	if schoolCode == "*" {
		schoolCode = ""
	}
	sarg := models.SchoolFilterRequest{
		Code: &schoolCode,
	}
	sarg.Limit = new(int)
	*sarg.Limit = 500
	schoolList, _, err := store.Store().SchoolsFindBy(ses.Context(), sarg)
	if err != nil {
		return err
	}
	schoolIds := []string{}
	for _, v := range schoolList {
		schoolIds = append(schoolIds, v.ID)
	}
	pargs := models.PeriodFilterRequest{
		SchoolIds: &schoolIds,
	}
	pargs.Limit = new(int)
	*pargs.Limit = 500
	periodList, _, err := store.Store().PeriodsListFilters(context.Background(), pargs)
	if err != nil {
		return err
	}
	log.Println("# Found schools: ", len(schoolList))
	ull := []models.Lesson{}
	for _, s := range schoolList {
		log.Println("#", *s.Name)
		var p *models.Period
		for _, pp := range periodList {
			if s.ID == *pp.SchoolId {
				p = pp
				break
			}
		}
		if p == nil {
			log.Println("No period created. Continue...")
			continue
		}
		pStartDate, pEndDate, err := (*p).Dates()
		if err != nil {
			return err
		}
		endDate := pEndDate
		if timeLvl == 1 {
			startDate = pStartDate
		}
		larg := models.LessonFilterRequest{
			SchoolId:  &s.ID,
			DateRange: &[]string{startDate.Format(time.DateOnly), endDate.Format(time.DateOnly)},
		}
		larg.Limit = new(int)
		*larg.Limit = 150000
		ll, _, err := store.Store().LessonsFindBy(context.Background(), larg)
		if err != nil {
			return err
		}
		log.Println(startDate.Format(time.DateOnly), endDate.Format(time.DateOnly), "Lessons found ", len(ll))
		// to update ll
		for k, v := range ll {
			pk, _ := p.GetKey(v.Date, true)
			if pk != 0 {
				if v.PeriodKey == nil || v.PeriodId == nil || *v.PeriodId != p.ID || *v.PeriodKey != pk {
					ll[k].PeriodKey = new(int)
					ll[k].PeriodId = new(string)
					*ll[k].PeriodKey = pk
					*ll[k].PeriodId = p.ID
					ull = append(ull, *v)
				}
			}
		}
		log.Println("Lessons update ", len(ull))
		err = store.Store().LessonsUpdateBatch(context.Background(), ull)
		ull = []models.Lesson{}
		if err != nil {
			return err
		}
	}
	log.Println("Success.")

	return nil
}
