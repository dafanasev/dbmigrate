package main

import (
	"os"
	"path/filepath"

	"github.com/dafanasev/dbmigrate"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	_ "github.com/spf13/viper/remote"
)

var v *viper.Viper

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
)

func init() {
	migrateCmd.PersistentFlags().StringVarP(&appName, "app_name", "a", "", "app name (used as prefix for env vars)")
	migrateCmd.PersistentFlags().StringVarP(&env, "env", "e", "", "optional environment (to support more than one database)")

	migrateCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "dbmigrations.yml", "config file (default is dbmigrations.yml)")
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
		var err error
		v, err = getViper()
		if err != nil {
			exitWithError(errors.Wrap(err, "can't start dbmigrate"))
		}
	})
}

func main() {
	err := migrateCmd.Execute()
	if err != nil {
		exitWithError(errors.Wrap(err, "can't execute migrations"))
	}
}

func getViper() (*viper.Viper, error) {
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
		kvsParams, err := parseKVConnectionString(kvsParamsStr)
		if err != nil {
			return nil, errors.Wrap(err, "wrong key value store connection")
		}

		if secretKeyRingPath != "" {
			viper.AddSecureRemoteProvider(kvsParams.provider, kvsParams.formatEndpoint(), kvsParams.path, secretKeyRingPath)
		} else {
			viper.AddRemoteProvider(kvsParams.provider, kvsParams.formatEndpoint(), kvsParams.path)
		}

		viper.SetConfigType(kvsParams.format)
		viper.ReadRemoteConfig()
	}

	envVarsPrefix := filepath.Base(projectDir)
	var v *viper.Viper
	if env == "" {
		v = viper.GetViper()
	} else {
		if !viper.IsSet(env) {
			v = viper.Sub(env)
		} else {
			v = viper.New()
		}
		envVarsPrefix += "_" + env
	}
	v.SetEnvPrefix(envVarsPrefix)
	v.AutomaticEnv()

	for _, flag := range []string{"engine", "database", "user", "password", "host", "port", "table", "missingdowns"} {
		v.BindPFlag(flag, migrateCmd.PersistentFlags().Lookup(flag))
	}

	return v, nil
}
