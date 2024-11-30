package cmd

import (
	"log"
	"os"

	"github.com/mekdep/server/internal/api/utils"
	"github.com/mekdep/server/internal/app"
	"github.com/mekdep/server/internal/models"
	"github.com/mekdep/server/internal/store"
	"github.com/spf13/cobra"
)

func MakeAdminCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "make-admin",
		Example: "{username} {password} {*roleCode} {*schoolCode}",
		Run:     makeAdminRun,
	}
}
func makeAdminRun(cmd *cobra.Command, args []string) {
	args = append(args, "", "", "", "")
	username := (args[0])
	password := (args[1])
	roleCode := (args[2])
	schoolCode := (args[3])

	err := makeAdmin(&utils.Session{}, username, password, roleCode, schoolCode)
	if err != nil {
		log.Fatalln(err)
	}
	os.Exit(0)
}
func makeAdmin(ses *utils.Session, username, password string, add ...string) error {
	var roleCode models.Role = models.RoleAdmin
	var schoolCode *string
	var schoolId *string

	if username == "" {
		return app.ErrRequired.SetKey("username")
	}
	if password == "" {
		return app.ErrRequired.SetKey("password")
	}

	if len(add) > 0 && add[0] != "" {
		roleCode = models.Role(add[0])
	}
	if len(add) > 1 && add[1] != "" {
		schoolCode = new(string)
		*schoolCode = add[1]
	}
	roleStr := string(roleCode)
	_, total, err := store.Store().UsersFindBy(ses.Context(), models.UserFilterRequest{
		Username: &username,
		Role:     &roleStr,
	})
	if err != nil {
		return err
	}
	if total > 0 {
		log.Println("Already exists: " + username)
		return nil
	}

	if schoolCode != nil {
		schoolItems, err := store.Store().SchoolsFindByCode(ses.Context(), []string{*schoolCode})
		if err != nil {
			return err
		}
		if len(schoolItems) < 1 {
			return app.ErrNotfound
		}
		school := schoolItems[0]
		schoolId = &school.ID
	}

	birthdayStr := "2000-02-02"
	_, _, err = app.UsersCreate(ses, &models.UserRequest{
		Username:  &username,
		FirstName: &username,
		LastName:  &username,
		Password:  &password,
		Birthday:  &birthdayStr,
		SchoolIds: &[]models.UserSchoolRequest{
			{
				RoleCode:  &roleCode,
				SchoolUid: schoolId,
			},
		},
	})
	if err != nil {
		return err
	}

	log.Println("Success.")

	return nil
}
