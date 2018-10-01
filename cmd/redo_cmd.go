package main

import (
	"fmt"
	"os"

	"github.com/dafanasev/dbmigrate"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func init() {
	redoCmd.Flags().IntVarP(&steps, "steps", "s", 0, "steps")
}

var redoCmd = &cobra.Command{
	Use:   "redo",
	Short: "redo last batch",
	RunE: func(cmd *cobra.Command, args []string) error {
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

		var n int
		var err error
		n, err = migrator.Rollback()
		if err != nil {
			close(done)
			return errors.Wrap(err, "can't redo: can't rollback")
		}
		if n == 0 {
			close(done)
			<-gdone
			fmt.Println("there are no migrations to reapply")
			return nil
		}

		n, err = migrator.MigrateSteps(n)
		close(done)

		<-gdone
		if err != nil {
			return errors.Wrap(err, "can't redo: can't migrate")
		}
		fmt.Printf("%d %s successfully reapplied\n", n, pluralize("migration", n))

		return nil
	},
}
