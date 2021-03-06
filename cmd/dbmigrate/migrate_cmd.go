package main

import (
	"fmt"

	"github.com/dafanasev/dbmigrate"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func init() {
	migrateCmd.Flags().IntVarP(&steps, "steps", "s", 0, "steps")
}

// migrateCmd is the root Cobra command, used to migrate database schema
var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Apply migrations",
	Long: `Migrate database schema.
By default, all unapplied migrations will be applied.
If --steps (-s) flag is provided, only -s migrations will.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		_, err := migrate(migrator, steps)
		return err
	},
}

// migrate is the actual migration function
func migrate(migrator *dbmigrate.Migrator, steps int) (int, error) {
	done := make(chan struct{})
	gdone := make(chan struct{})

	go func() {
		for {
			select {
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
		return n, errors.Wrap(err, "can't migrate")
	}

	if n == 0 {
		fmt.Println("there are no migrations to apply")
		return n, nil
	}
	fmt.Printf("%d %s successfully applied\n", n, pluralize("migration", n))

	return n, nil
}
