package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/golang-migrate/migrate"
	"github.com/mekdep/server/config"
	"github.com/spf13/cobra"
)

func MigrateUpCmd() *cobra.Command {
	return &cobra.Command{
		Use: "migrate-up",
		Run: MigrateUp,
	}
}
func MigrateDownCmd() *cobra.Command {
	return &cobra.Command{
		Use: "migrate-down",
		Run: MigrateDown,
	}
}
func MigrateDataCmd() *cobra.Command {
	return &cobra.Command{
		Use: "migrate-data",
		Run: MigrateData,
	}
}

func migrateInit() (*migrate.Migrate, error) {
	dbUrl := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		config.Conf.DbUsername, config.Conf.DbPassword, config.Conf.DbHost, config.Conf.DbPort, config.Conf.DbDatabase)

	return migrate.New("database/migrations", dbUrl)
}

func MigrateUp(cmd *cobra.Command, args []string) {
	m, err := migrateInit()
	if err != nil {
		log.Fatal(err)
	}

	err = m.Up()
	if err != nil {
		log.Fatal(err)
	}
	os.Exit(0)
}

func MigrateDown(cmd *cobra.Command, args []string) {
	m, err := migrateInit()
	if err != nil {
		log.Fatal(err)
	}
	err = m.Down()
	if err != nil {
		log.Fatal(err)
	}
	os.Exit(0)
}

func MigrateData(cmd *cobra.Command, args []string) {
	m, err := migrateInit()
	if err != nil {
		log.Fatal(err)
	}
	err = m.Up()
	if err != nil {
		log.Fatal(err)
	}
	os.Exit(0)
}
