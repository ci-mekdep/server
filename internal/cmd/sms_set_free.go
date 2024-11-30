package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"os"
	"time"

	apiutils "github.com/mekdep/server/internal/api/utils"
	"github.com/mekdep/server/internal/app"
	"github.com/mekdep/server/internal/models"
	"github.com/mekdep/server/internal/store"
	"github.com/mekdep/server/internal/utils"
	"github.com/spf13/cobra"
)

func GiftPlusCmd() *cobra.Command {
	return &cobra.Command{
		Use: "gift-plus",
		Run: SetFreePlusRun,
	}
}
func SetFreePlusRun(cmd *cobra.Command, args []string) {
	args = append(args, "", "", "")
	schoolId := args[0]
	classroomId := args[1]
	studentId := args[2]
	err := SetFreePlus(&apiutils.Session{}, models.PaymentPlus, schoolId, classroomId, studentId)
	if err != nil {
		log.Fatalln(err)
	}
	os.Exit(0)
}
func SetFreePlus(ses *apiutils.Session, tariffType models.PaymentTariffType, schoolId string, classroomId string, studentId string) error {
	args := models.UserFilterRequest{}
	args.Limit = new(int)
	args.Offset = new(int)
	*args.Limit = 500
	if studentId != "" {
		args.ID = &studentId
	} else if classroomId != "" {
		cl, err := store.Store().ClassroomsFindById(ses.Context(), classroomId)
		if err != nil {
			return err
		}
		sch, err := store.Store().SchoolsFindById(ses.Context(), cl.SchoolId)
		if err != nil {
			return err
		}
		log.Println("For whole classroom", *cl.Name, " of school", *sch.Name)
		args.ClassroomId = &cl.ID

	} else if schoolId != "" {
		sch, err := store.Store().SchoolsFindById(ses.Context(), classroomId)
		if err != nil {
			return err
		}
		log.Println("For whole classroom", *sch.Name)
		args.SchoolId = &sch.ID
	} else {
		return app.ErrRequired.SetKey("school_id")
	}
	studentModels, total, err := store.Store().UsersFindBy(context.Background(), args)
	if err != nil {
		return err
	}
	err = store.Store().UsersLoadRelationsParents(context.Background(), &studentModels)
	if err != nil {
		return err
	}
	for total > 0 {
		log.Println("Found", total, "students to give away plus! yet ", len(studentModels), " in 5 sec...")
		time.Sleep(time.Second * 5)
		errs := []error{}
		for _, student := range studentModels {
			log.Println("Student", student.ID, *student.FirstName, "with parents", len(student.Parents))
			err = app.UserTariffUpgrade(ses, &models.PaymentTransaction{TariffType: tariffType}, *student, student.Parents)
			if err != nil {
				errs = append(errs, errors.New(err.Error()+" "+student.ID))
			}
		}
		if len(errs) > 0 {
			tmp, _ := json.Marshal(errs)
			utils.LoggerDesc("in SetFreePlus after upgrade, all errors: " + string(tmp))
			return errs[0]
		}

		total -= *args.Limit
		*args.Offset += *args.Limit
		studentModels, _, err := store.Store().UsersFindBy(context.Background(), args)
		if err != nil {
			return err
		}
		err = store.Store().UsersLoadRelationsParents(context.Background(), &studentModels)
		if err != nil {
			return err
		}
	}
	if err != nil {
		return err
	}
	return nil
}
