package cmd

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/mekdep/server/internal/api/utils"
	"github.com/mekdep/server/internal/app"
	"github.com/mekdep/server/internal/models"
	"github.com/mekdep/server/internal/store"
	"github.com/spf13/cobra"
)

func UpdatePeriodGradesCmd() *cobra.Command {
	return &cobra.Command{
		Use: "update-period-grades",
		Run: UpdatePeriodGradesRun,
	}
}
func UpdatePeriodGradesRun(cmd *cobra.Command, args []string) {
	args = append(args, "", "", "", "")
	periodNum, _ := strconv.Atoi(args[0])
	schoolCode := (args[1])
	nextSchoolCode := (args[2])
	selectLeast := (args[3])
	err := UpdatePeriodGrades(&utils.Session{}, periodNum, schoolCode, nextSchoolCode, selectLeast == "1")
	if err != nil {
		log.Fatalln(err)
	}
	os.Exit(0)
}
func UpdatePeriodGrades(ses *utils.Session, periodNum int, schoolCode string, nextSchoolCode string, selectLeast bool) error {
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
	for _, s := range schoolList {
		if isContinue {
			if *s.Code == nextSchoolCode {
				isContinue = false
			} else {
				continue
			}
		}
		log.Println("#", *s.Name)
		// get students
		uarg := models.UserFilterRequest{
			SchoolId: &s.ID,
		}
		uarg.Limit = new(int)
		*uarg.Limit = 5000
		uarg.Role = new(string)
		*uarg.Role = string(models.RoleStudent)
		studentList, _, err := store.Store().UsersFindBy(ses.Context(), uarg)
		if err != nil {
			return err
		}
		err = store.Store().UsersLoadRelationsClassrooms(ses.Context(), &studentList)
		if err != nil {
			return err
		}
		// get period
		pp, _, err := store.Store().PeriodsListFilters(ses.Context(), models.PeriodFilterRequest{
			SchoolId: &s.ID,
		})
		if err != nil {
			return err
		}
		if len(pp) < 1 {
			log.Println("ERR", app.ErrNotExists.SetKey("period_id"))
		}
		p := pp[0]
		schoolPeriodNum := 1
		if periodNum == 0 {
			schoolPeriodNum, _ = p.GetKey(time.Now(), false)
		} else {
			schoolPeriodNum = periodNum
		}
		// get subject
		sarg := models.SubjectFilterRequest{
			SchoolId: &s.ID,
		}
		sarg.Limit = new(int)
		*sarg.Limit = 5000
		subjectList, _, err := store.Store().SubjectsListFilters(ses.Context(), &sarg)
		if err != nil {
			return err
		}
		// get period grades if necessary
		periodGrades := []*models.PeriodGrade{}
		studentIds := []string{}
		subjectIds := []string{}
		for _, v := range studentList {
			studentIds = append(studentIds, v.ID)
		}
		for _, v := range subjectList {
			subjectIds = append(subjectIds, v.ID)
		}
		if selectLeast {
			pgargs := models.PeriodGradeFilterRequest{}
			pgargs.StudentIds = &studentIds
			pgargs.SubjectIds = &subjectIds
			periodGrades, _, _ = store.Store().PeriodGradesFindBy(ses.Context(), pgargs)
		}
		log.Println("Found students", len(studentList), "updating...")
		updates := []models.PeriodGrade{}
		totalUpdates := 0
		for _, subject := range subjectList {
			// log.Println("Updating ", *subject.Name, " of ", len(subjectList))
			for _, student := range studentList {
				for _, v := range periodGrades {
					if *v.StudentId == student.ID && *v.SubjectId == subject.ID {
						if v.GradeCount > 2 {
							continue
						}
					}
				}

				for _, classroom := range student.Classrooms {
					if classroom.ClassroomId == subject.ClassroomId {
						updates = append(updates, models.PeriodGrade{
							SubjectId: &subject.ID,
							PeriodId:  &p.ID,
							PeriodKey: schoolPeriodNum,
							StudentId: &student.ID,
						})
						if subject.ParentId != nil {
							updates = append(updates, models.PeriodGrade{
								SubjectId: subject.ParentId,
								PeriodId:  &p.ID,
								PeriodKey: schoolPeriodNum,
								StudentId: &student.ID,
							})
						}
						if len(updates) > 1000 {
							err = store.Store().PeriodGradesUpdateBatch(ses.Context(), updates)
							totalUpdates = totalUpdates + len(updates)
							log.Println("Updated ", len(updates), "#", *s.Name, "totaled", totalUpdates)
							if err != nil {
								return err
							}
							updates = []models.PeriodGrade{}
						}
					}
				}
			}
		}
		if len(updates) > 0 {
			err = store.Store().PeriodGradesUpdateBatch(ses.Context(), updates)
			totalUpdates = totalUpdates + len(updates)
			log.Println("Updated ", len(updates), "#", *s.Name, "totaled", totalUpdates)
			if err != nil {
				return err
			}
			updates = []models.PeriodGrade{}
		}
	}
	log.Println("Success.")

	return nil
}
