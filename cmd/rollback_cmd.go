package main

import (
	"fmt"
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func init() {
	rollbackCmd.Flags().IntVarP(&steps, "steps", "s", 0, "steps")
}

var rollbackCmd = &cobra.Command{
	Use:   "rollback",
	Short: "rollback database schema",
	RunE: func(cmd *cobra.Command, args []string) error {
		done := make(chan struct{})
		gdone := make(chan struct{})

		go func() {
			for {
				select {
				case err := <-migrator.ErrorsCh:
					fmt.Fprintln(os.Stderr, errors.Wrap(err, "rollback error"))
				case migration := <-migrator.MigrationsCh:
					fmt.Printf("migration %s has been successfully rolled back\n", migration.FileName())
				case <-done:
					close(gdone)
					return
				}
			}
		}()

		var n int
		var err error
		if steps == 0 {
			n, err = migrator.Rollback()
		} else {
			n, err = migrator.RollbackSteps(steps)
		}
		close(done)
		if err != nil {
			return errors.Wrap(err, "can't rollback")
		}

		<-gdone
		fmt.Printf("%d migrations successfully rolled back\n", n)

		return nil
	},
}
