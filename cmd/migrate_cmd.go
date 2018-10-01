package main

import (
	"fmt"
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func init() {
	migrateCmd.Flags().IntVarP(&steps, "steps", "s", 0, "steps")
}

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "migrate database schema",
	RunE: func(cmd *cobra.Command, args []string) error {
		done := make(chan struct{})
		gdone := make(chan struct{})

		go func() {
			for {
				select {
				case err := <-migrator.ErrorsCh:
					fmt.Fprintln(os.Stderr, errors.Wrap(err, "migration error"))
				case migration := <-migrator.MigrationsCh:
					fmt.Printf("migration %s has been successfully applied\n", migration.FileName())
				case <-done:
					close(gdone)
					return
				}
			}
		}()

		n, err := migrator.MigrateSteps(steps)
		close(done)

		<-gdone
		if err != nil {
			return errors.Wrap(err, "can't migrate")
		}

		if n == 0 {
			fmt.Println("there are no migrations to apply")
			return nil
		}
		fmt.Printf("%d %s successfully applied\n", n, pluralize("migration", n))

		return nil
	},
}
