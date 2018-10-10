package main

import (
	"fmt"
	"strings"

	"github.com/dafanasev/dbmigrate"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// enginesNoOptDefVal is the constant used to specify that migration should be created for the current (used in migrator) database engine only
const enginesNoOptDefVal = "currentengine"

// migrationsGeneratorEngines used by flag which specifies database engines to create migrations for
var migrationsGeneratorEngines []string

func init() {
	generateCmd.Flags().StringSliceVarP(&migrationsGeneratorEngines, "engines", "g", nil, "specific engines")
	// if flag is set without a value use this placeholder to later set specific engine to the one from migrator settings
	generateCmd.Flags().Lookup("engines").NoOptDefVal = enginesNoOptDefVal
}

// generateCmd is the Cobra command to generate migrations
var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate migration",
	Long: `Generate up and down migrations, use args to build migration name,
e.g. dbmigrate generate Create posts table will become create_posts_table in the generated migration name.`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return generateMigration(migrator, migrationsGeneratorEngines, args...)
	},
}

// generateMigration the the actual migration generation function
func generateMigration(migrator *dbmigrate.Migrator, engines []string, args ...string) error {
	if len(engines) == 1 && engines[0] == enginesNoOptDefVal {
		engines[0] = migrator.Engine
	}

	if len(engines) == 1 && engines[0] == "all" {
		engines = dbmigrate.Engines()
	}

	fpaths, err := migrator.GenerateMigration(strings.Join(args, " "), engines...)
	if err != nil {
		return errors.Wrap(err, "can't generate migration")
	}

	for _, fpath := range fpaths {
		fmt.Printf("created %s\n", fpath)
	}

	return nil
}
