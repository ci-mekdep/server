package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
)

func Init() {
	rootCmd := cobra.Command{}
	rootCmd.AddCommand(&cobra.Command{
		Use: "version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("eMekdep version v1.0")
			os.Exit(0)
		},
	})
	rootCmd.AddCommand(MigrateUpCmd())
	rootCmd.AddCommand(MigrateDownCmd())
	rootCmd.AddCommand(MigrateDataCmd())
	rootCmd.AddCommand(GiftPlusCmd())
	rootCmd.AddCommand(UpdateLessonPeriodsCmd())
	rootCmd.AddCommand(UpdateLessonDatesCmd())
	rootCmd.AddCommand(UpdatePeriodGradesCmd())
	rootCmd.AddCommand(UpdatePaymentStatusCmd())
	rootCmd.AddCommand(MakeAdminCmd())
	err := rootCmd.Execute()
	if err != nil {
		log.Fatal(err)
	}
}
