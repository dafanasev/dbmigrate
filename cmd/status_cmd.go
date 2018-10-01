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

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "database schema status",
	RunE: func(cmd *cobra.Command, args []string) error {
		migrations, err := migrator.Status()
		if err != nil {
			return errors.Wrap(err, "can't get migrations status")
		}

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
	},
}
