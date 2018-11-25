package main

import (
	"github.com/dafanasev/dbmigrate"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	_ "github.com/spf13/viper/remote"
)

// appFlags contains vars that can be specified only as flags
type appFlags struct {
	// prefix defines alternative prefix for environment variable names
	prefix string
	// Env defines optional alternative environment and thus alternative database configuration, e.g. for tests
	env string

	// config file name (without extension)
	configFile string

	// kvsParamsStr is key value store connection string (in store://host(:port)/path.type format)
	kvsParamsStr string
	// secretKeyRingPath is a path to key ring file
	secretKeyRingPath string
}

var (
	// migrator is the Migrator instance, suddenly
	migrator *dbmigrate.Migrator
	flags    *appFlags
	// steps variable, used for the corresponding flag in root (migrate)/rollback/reapply commands
	steps int
)

// migrateFlags holds variables used for flags that used by viper to provide settings for migrator
var migrateFlags struct {
	engine            string
	database          string
	user              string
	password          string
	host              string
	port              int
	migrationsTable   string
	allowMissingDowns bool
}

func init() {
	flags = &appFlags{}

	migrateCmd.PersistentFlags().StringVarP(&flags.prefix, "prefix", "x", "", "environment variables prefix, default is the project dir name")
	migrateCmd.PersistentFlags().StringVarP(&flags.env, "env", "e", "", "optional environment (to support more than one database, e.g. for tests)")

	migrateCmd.PersistentFlags().StringVarP(&flags.configFile, "config", "c", "dbmigrate", "config file, default is dbmigrate.yml")
	migrateCmd.PersistentFlags().StringVarP(&flags.kvsParamsStr, "kvsparams", "k", "", "key value connection string, format is provider://host:port/path.type")
	migrateCmd.PersistentFlags().StringVarP(&flags.secretKeyRingPath, "secretkeyring", "r", "", "secret key ring path")

	migrateCmd.PersistentFlags().StringVarP(&migrateFlags.engine, "engine", "n", "", "database engine (postgres, mysql or sqlite)")
	migrateCmd.PersistentFlags().StringVarP(&migrateFlags.database, "database", "d", "", "database name")
	migrateCmd.PersistentFlags().StringVarP(&migrateFlags.user, "user", "u", "", "database user")
	migrateCmd.PersistentFlags().StringVarP(&migrateFlags.password, "password", "p", "", "database password")
	migrateCmd.PersistentFlags().StringVarP(&migrateFlags.host, "host", "b", "", "database host, default is localhost")
	migrateCmd.PersistentFlags().IntVarP(&migrateFlags.port, "port", "o", 0, "database port, default is specific for each database engine")
	migrateCmd.PersistentFlags().StringVarP(&migrateFlags.migrationsTable, "table", "t", "", "migrations table, default is migrations")
	migrateCmd.PersistentFlags().BoolVarP(&migrateFlags.allowMissingDowns, "missingdowns", "m", false, "allow missing down migrations")

	migrateCmd.AddCommand(generateCmd, statusCmd, rollbackCmd, reapplyCmd)

	// only here flags are parsed and viper gives proper configuration,
	// so we initialize migrator here instead of main function
	cobra.OnInitialize(func() {
		v, err := (&viperConfigurator{viper: viper.GetViper(), flags: flags}).configure()
		if err != nil {
			exitWithError(err)
		}

		migrator, err = dbmigrate.NewMigrator(&dbmigrate.Settings{
			Engine:            v.GetString("engine"),
			Database:          v.GetString("database"),
			User:              v.GetString("user"),
			Password:          v.GetString("password"),
			Host:              v.GetString("host"),
			Port:              v.GetInt("port"),
			MigrationsTable:   v.GetString("table"),
			AllowMissingDowns: v.GetBool("missingdowns"),
			MigrationsCh:      make(chan *dbmigrate.Migration),
			ErrorsCh:          make(chan error),
		})
		if err != nil {
			exitWithError(err)
		}
	})
}

func main() {
	err := migrateCmd.Execute()
	if err != nil {
		exitWithError(err)
	}

	if migrator != nil {
		migrator.Close()
	}
}
