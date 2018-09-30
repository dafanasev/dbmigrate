package main

import (
	"github.com/dafanasev/dbmigrate"
	"github.com/pkg/errors"
	_ "github.com/spf13/viper/remote"
)

var migrator *dbmigrate.Migrator

var (
	appName string
	env     string

	configFile string

	kvsParamsStr      string
	secretKeyRingPath string

	engine            string
	database          string
	user              string
	password          string
	host              string
	port              int
	migrationsTable   string
	allowMissingDowns bool

	steps int
)

func init() {
	migrateCmd.PersistentFlags().StringVarP(&appName, "appname", "a", "", "app name (used as prefix for env vars)")
	migrateCmd.PersistentFlags().StringVarP(&env, "env", "e", "", "optional environment (to support more than one database)")

	migrateCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "dbmigrations", "config file (default is dbmigrations.yml)")
	migrateCmd.PersistentFlags().StringVarP(&kvsParamsStr, "kvsparams", "k", "", "key value connection string, (provider://host:port/path.format")
	migrateCmd.PersistentFlags().StringVarP(&secretKeyRingPath, "secretkeyring", "r", "", "secret key ring path")

	migrateCmd.PersistentFlags().StringVarP(&engine, "engine", "n", "", "database engine")
	migrateCmd.PersistentFlags().StringVarP(&database, "database", "d", "", "database name")
	migrateCmd.PersistentFlags().StringVarP(&user, "user", "u", "", "database user")
	migrateCmd.PersistentFlags().StringVarP(&password, "password", "p", "", "database password")
	migrateCmd.PersistentFlags().StringVarP(&host, "host", "b", "", "database host")
	migrateCmd.PersistentFlags().IntVarP(&port, "port", "o", 0, "database port")
	migrateCmd.PersistentFlags().StringVarP(&migrationsTable, "table", "t", "", "migrations table")
	migrateCmd.PersistentFlags().BoolVarP(&allowMissingDowns, "missingdowns", "m", false, "allow missing down migrations")
}

func main() {
	startErrStr := "can't start dbmigrate"
	v, err := setupViper()
	if err != nil {
		exitWithError(errors.Wrap(err, startErrStr))
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
		exitWithError(errors.Wrapf(err, startErrStr))
	}

	migrateCmd.AddCommand(generateCmd, statusCmd, rollbackCmd, redoCmd)
	err = migrateCmd.Execute()
	if err != nil {
		exitWithError(err)
	}

	migrator.Close()
}
