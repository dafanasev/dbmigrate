package main

import (
	"fmt"
	"os"

	"github.com/dafanasev/dbmigrate"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func init() {
	reapplyCmd.Flags().IntVarP(&steps, "steps", "s", 0, "steps")
}

var reapplyCmd = &cobra.Command{
	Use:   "reapply",
	Short: "Reapply migrations",
	Long: `Rollback migrations.
The latest migration operation will be reapplied, e.g. if 3 migrations have been applied, 3 migrations will be rolled back and reapplied.
If --steps (-s) flag is provided, -s migrations will be reapplied.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		_, err := reapply(migrator, steps)
		return err
	},
}

func reapply(migrator *dbmigrate.Migrator, steps int) (int, error) {
	done := make(chan struct{})
	gdone := make(chan struct{})

	go func() {
		for {
			select {
			case err := <-migrator.ErrorsCh:
				fmt.Fprintln(os.Stderr, errors.Wrap(err, "migration error"))
			case migration := <-migrator.MigrationsCh:
				switch migration.Direction {
				case dbmigrate.DirectionUp:
					fmt.Printf("migration %s has been successfully applied\n", migration.FileName())
				case dbmigrate.DirectionDown:
					fmt.Printf("migration %s has been successfully rolled back\n", migration.FileName())
				}
			case <-done:
				close(gdone)
				return
			}
		}
	}()

	n, err := migrator.RollbackSteps(steps)
	if err != nil {
		close(done)
		return n, errors.Wrap(err, "can't reapply: can't rollback")
	}
	if n == 0 {
		close(done)
		<-gdone
		fmt.Println("there are no migrations to reapply")
		return n, nil
	}

	n, err = migrator.MigrateSteps(n)
	close(done)

	<-gdone
	if err != nil {
		return n, errors.Wrap(err, "can't reapply: can't migrate")
	}
	fmt.Printf("%d %s successfully reapplied\n", n, pluralize("migration", n))

	return n, nil
}
