package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/dafanasev/dbmigrate"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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

	cobra.OnInitialize(func() {
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
	})
}

func main() {
	migrateCmd.AddCommand(generateCmd)
	err := migrateCmd.Execute()
	if err != nil {
		exitWithError(err)
	}
}

func setupViper() (*viper.Viper, error) {
	wd, err := os.Getwd()
	if err != nil {
		exitWithError(errors.Wrap(err, "can't get working directory"))
	}

	projectDir, err := dbmigrate.FindProjectDir(wd)
	if err != nil {
		exitWithError(err)
	}

	viper.AddConfigPath(projectDir)
	viper.SetConfigName(configFile)
	viper.ReadInConfig()

	if kvsParamsStr != "" {
		kvsParams, err := parseKVSConnectionString(kvsParamsStr)
		if err != nil {
			return nil, errors.Wrap(err, "wrong key value store connection")
		}

		kvsErrorString := fmt.Sprintf("can't connect to key value store using connection string %s", kvsParamsStr)
		if secretKeyRingPath != "" {
			err = viper.AddSecureRemoteProvider(kvsParams.provider, kvsParams.formatEndpoint(), kvsParams.path, secretKeyRingPath)
			if err != nil {
				return nil, errors.Wrapf(err, "%s and key ring path %s", kvsErrorString, secretKeyRingPath)
			}
		} else {
			err = viper.AddRemoteProvider(kvsParams.provider, kvsParams.formatEndpoint(), kvsParams.path)
			if err != nil {
				return nil, errors.Wrap(err, kvsParamsStr)
			}
		}

		viper.SetConfigType(kvsParams.format)
		err = viper.ReadRemoteConfig()
		if err != nil {
			return nil, errors.Wrap(err, kvsErrorString)
		}
	}

	var v *viper.Viper

	if env == "" {
		v = viper.GetViper()
	} else {
		if !viper.IsSet(env) {
			v = viper.Sub(env)
		} else {
			v = viper.New()
		}
	}

	var envVarsPrefix string
	if appName != "" {
		envVarsPrefix = appName
	} else {
		envVarsPrefix = filepath.Base(projectDir)
	}
	if env != "" {
		envVarsPrefix += "_" + env
	}

	v.SetEnvPrefix(envVarsPrefix)
	v.AutomaticEnv()

	for _, flag := range []string{"engine", "database", "user", "password", "host", "port", "table", "missingdowns"} {
		err = v.BindPFlag(flag, migrateCmd.PersistentFlags().Lookup(flag))
		if err != nil {
			return nil, errors.Wrapf(err, "can't bind flag %s", flag)
		}
	}

	return v, nil
}
