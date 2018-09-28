package main

import (
	"fmt"
	"strings"

	"github.com/dafanasev/dbmigrate"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var migrationsGeneratorEngines []string

func init() {
	generateCmd.Flags().StringSliceVarP(&migrationsGeneratorEngines, "genengines", "g", nil, "specific engine")
	// if flag is set without an option use this placeholder to later set specific engine to the one from migrator settings
	generateCmd.Flags().Lookup("genengines").NoOptDefVal = "current"
}

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "generate migration",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(migrationsGeneratorEngines) > 0 && migrationsGeneratorEngines[0] != "all" && migrationsGeneratorEngines[0] != "current" {
			for _, engine := range migrationsGeneratorEngines {
				if !dbmigrate.EngineExists(engine) {
					return errors.Errorf("can't generate migration, engines %s is not exists/supported", engine)
				}
			}
		}

		if len(migrationsGeneratorEngines) == 0 {
			// generate migration for all engines
			migrationsGeneratorEngines = []string{""}
		}

		if len(migrationsGeneratorEngines) == 1 && migrationsGeneratorEngines[0] == "current" {
			migrationsGeneratorEngines[0] = migrator.Engine
		}

		if len(migrationsGeneratorEngines) == 1 && migrationsGeneratorEngines[0] == "all" {
			migrationsGeneratorEngines = dbmigrate.Engines()
		}

		for _, engine := range migrationsGeneratorEngines {
			fpaths, err := migrator.GenerateMigration(strings.Join(args, " "), engine)
			if err != nil {
				return errors.Wrap(err, "can't generate migration")
			}

			for _, fpath := range fpaths {
				fmt.Printf("created %s\n", fpath)
			}
		}

		return nil
	},
}
