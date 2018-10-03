package main

import (
	"fmt"
	"os"

	"github.com/dafanasev/dbmigrate"
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
		_, err := rollback(migrator, steps)
		return err
	},
}

func rollback(migrator *dbmigrate.Migrator, steps int) (int, error) {
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

	<-gdone
	if err != nil {
		return n, errors.Wrap(err, "can't rollback")
	}

	if n == 0 {
		fmt.Println("there are no migrations to rollback")
		return n, nil
	}
	fmt.Printf("%d %s successfully rolled back\n", n, pluralize("migration", n))

	return n, nil
}
