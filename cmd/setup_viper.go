package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/dafanasev/dbmigrate"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

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
