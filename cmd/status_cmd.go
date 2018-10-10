package main

import (
	"fmt"
	"os"
	"time"

	"github.com/dafanasev/dbmigrate"
	"github.com/olekukonko/tablewriter"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// statusCmd is the Cobra command that shows status of all migrations, if they were applied and when
// also it shows latest version migration, last applied migration and if database schema is up to date or not
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Migrations status",
	Long: `Shows migrations list with names, versions and applied at times, if the migration was applied.
It also shows the latest version migration, the last applied migrations (they are not necessarily the same ones),
number of applied migrations and if the database schema is up to date or not.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return status(migrator)
	},
}

// status is the actual status function
func status(migrator *dbmigrate.Migrator) error {
	migrations, err := migrator.Status()
	if err != nil {
		return errors.Wrap(err, "can't get migrations status")
	}

	// function to format appliedAt timestamp
	appliedAtRowFn := func(appliedAt time.Time) string {
		if appliedAt == (time.Time{}) {
			return "-"
		}
		return appliedAt.Format(dbmigrate.PrintTimestampFormat)
	}

	if len(migrations) == 0 {
		fmt.Println("No migrations exist yet")
		return nil
	}

	isUpToDate := true

	// show igrations table
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Name", "Version", "Applied at"})
	table.SetAutoWrapText(false)
	for _, migration := range migrations {
		table.Append([]string{
			migration.HumanName(), migration.Version.Format(dbmigrate.TimestampFormat),
			appliedAtRowFn(migration.AppliedAt),
		})

		if migration.AppliedAt == (time.Time{}) {
			isUpToDate = false
		}
	}
	table.Render()

	lvm, err := migrator.LatestVersionMigration()
	if err != nil {
		return errors.Wrap(err, "can't get latest version migration")
	}
	if lvm != nil {
		fmt.Printf("Latest version migration is %s\n", lvm.FileName())
	}

	lam, err := migrator.LastAppliedMigration()
	if err != nil {
		return errors.Wrap(err, "can't get last applied migration")
	}
	if lam != nil {
		fmt.Printf("Last applied migration is %s\n", lam.FileName())
	} else {
		fmt.Println("No migrations were applied yet")
	}

	if isUpToDate {
		fmt.Println("Database schema is up to date")
	} else {
		fmt.Println("Database schema is not up to date")
	}

	return nil
}
