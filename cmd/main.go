package main

import (
	"github.com/dafanasev/dbmigrate"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	_ "github.com/spf13/viper/remote"
)

type appFlags struct {
	appName string
	env     string

	configFile string

	kvsParamsStr      string
	secretKeyRingPath string
}

var (
	migrator *dbmigrate.Migrator
	flags    *appFlags
	steps    int
)

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

	migrateCmd.PersistentFlags().StringVarP(&flags.appName, "appname", "a", "", "app name (used as prefix for env vars)")
	migrateCmd.PersistentFlags().StringVarP(&flags.env, "env", "e", "", "optional environment (to support more than one database)")

	migrateCmd.PersistentFlags().StringVarP(&flags.configFile, "config", "c", "dbmigrate", "config file (default is dbmigrate.yml)")
	migrateCmd.PersistentFlags().StringVarP(&flags.kvsParamsStr, "kvsparams", "k", "", "key value connection string, (provider://host:port/path.format")
	migrateCmd.PersistentFlags().StringVarP(&flags.secretKeyRingPath, "secretkeyring", "r", "", "secret key ring path")

	migrateCmd.PersistentFlags().StringVarP(&migrateFlags.engine, "engine", "n", "", "database engine")
	migrateCmd.PersistentFlags().StringVarP(&migrateFlags.database, "database", "d", "", "database name")
	migrateCmd.PersistentFlags().StringVarP(&migrateFlags.user, "user", "u", "", "database user")
	migrateCmd.PersistentFlags().StringVarP(&migrateFlags.password, "password", "p", "", "database password")
	migrateCmd.PersistentFlags().StringVarP(&migrateFlags.host, "host", "b", "", "database host")
	migrateCmd.PersistentFlags().IntVarP(&migrateFlags.port, "port", "o", 0, "database port")
	migrateCmd.PersistentFlags().StringVarP(&migrateFlags.migrationsTable, "table", "t", "", "migrations table")
	migrateCmd.PersistentFlags().BoolVarP(&migrateFlags.allowMissingDowns, "missingdowns", "m", false, "allow missing down migrations")

	migrateCmd.AddCommand(generateCmd, statusCmd, rollbackCmd, reapplyCmd)

	cobra.OnInitialize(func() {
		v, err := (&viperConfigurator{viper: viper.GetViper(), flags: flags}).configure()
		if err != nil {
			exitWithError(err)
		}

		migrator, err = dbmigrate.NewMigrator(&dbmigrate.Settings{
			Engine:            v.GetString("engine"),
			Database:          v.GetString("database"),
			User:              v.GetString("user"),
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

	migrator.Close()
}
