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

func MakeSchoolDataCmd() *cobra.Command {
	return &cobra.Command{
		Use: "make-school",
		Run: makeSchoolDataRun,
	}
}
func makeSchoolDataRun(cmd *cobra.Command, args []string) {
	args = append(args, " ", " ", "")
	periodNum, _ := strconv.Atoi(args[0])
	schoolCode := (args[1])
	classroomName := (args[2])
	err := makeSchoolData(&utils.Session{}, periodNum, schoolCode, classroomName)
	if err != nil {
		log.Fatalln(err)
	}
	os.Exit(0)
}
func makeSchoolData(ses *utils.Session, periodNum int, schoolCode string, classroomName string) error {
	if schoolCode == "0" {
		schoolCode = ""
	}
	scarg := models.SchoolFilterRequest{
		Code: &schoolCode,
	}
	scarg.Limit = new(int)
	*scarg.Limit = 500
	schoolList, _, err := store.Store().SchoolsFindBy(ses.Context(), scarg)
	if err != nil {
		return err
	}
	log.Println("# Found schools: ", len(schoolList))
	for _, s := range schoolList {
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
			return app.ErrNotExists.SetKey("period_id")
		}
		p := pp[0]
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
		log.Println("Found subjects", len(subjectList))
		log.Println("Found students", len(studentList), "updating...")
		for k, v := range subjectList {
			log.Println("Update", k, "/", len(subjectList))
			for _, st := range studentList {
				for _, cl := range st.Classrooms {
					if cl.ClassroomId == v.ClassroomId &&
						cl.Type == v.ClassroomType && cl.TypeKey == v.ClassroomTypeKey {
						if classroomName == "" || (cl.Classroom != nil && *cl.Classroom.Name == classroomName || cl.ClassroomId == (classroomName)) {
							store.Store().PeriodGradesUpdateValues(ses.Context(), models.PeriodGrade{
								SubjectId: &v.ID,
								PeriodId:  &p.ID,
								PeriodKey: periodNum,
								StudentId: &st.ID,
							})
						}
					}
				}
			}
		}
	}
	log.Println("Success.")

	return nil
}
