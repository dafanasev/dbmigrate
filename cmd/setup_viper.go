package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/dafanasev/dbmigrate"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

type viperConfigurator struct {
	v          *viper.Viper
	flags      *appFlags
	projectDir string
}

func (vc *viperConfigurator) setup() (*viper.Viper, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, errors.Wrap(err, "can't get working directory")
	}

	vc.projectDir, err = dbmigrate.FindProjectDir(wd)
	if err != nil {
		return nil, err
	}

	vc.readConfigFile()
	if flags.kvsParamsStr != "" {
		err = vc.readKVS()
		if err != nil {
			return nil, err
		}
	}
	if flags.env != "" {
		vc.v = vc.newViper()
	}
	vc.readEnv()
	vc.readFlags()

	return vc.v, nil
}

func (vc *viperConfigurator) newViper() *viper.Viper {
	if vc.v.IsSet(vc.flags.env) {
		return vc.v.Sub(vc.flags.env)
	}
	return viper.New()
}

func (vc *viperConfigurator) readConfigFile() {
	vc.v.AddConfigPath(vc.projectDir)
	vc.v.SetConfigName(vc.flags.configFile)
	vc.v.ReadInConfig()
}

func (vc *viperConfigurator) readKVS() error {
	kvsParams, err := parseKVSConnectionString(flags.kvsParamsStr)
	if err != nil {
		return errors.Wrap(err, "wrong key value store connection")
	}

	kvsErrorString := fmt.Sprintf("can't connect to key value store using connection string %s", vc.flags.kvsParamsStr)
	if vc.flags.secretKeyRingPath != "" {
		err = vc.v.AddSecureRemoteProvider(kvsParams.provider, kvsParams.formatEndpoint(), kvsParams.path, vc.flags.secretKeyRingPath)
		if err != nil {
			return errors.Wrapf(err, "%s and key ring path %s", kvsErrorString, vc.flags.secretKeyRingPath)
		}
	} else {
		err = vc.v.AddRemoteProvider(kvsParams.provider, kvsParams.formatEndpoint(), kvsParams.path)
		if err != nil {
			return errors.Wrap(err, vc.flags.kvsParamsStr)
		}
	}

	vc.v.SetConfigType(kvsParams.format)
	err = vc.v.ReadRemoteConfig()
	if err != nil {
		return errors.Wrap(err, kvsErrorString)
	}
	return nil
}

func (vc *viperConfigurator) readEnv() {
	var envVarsPrefix string
	if vc.flags.appName != "" {
		envVarsPrefix = vc.flags.appName
	} else {
		envVarsPrefix = filepath.Base(vc.projectDir)
	}
	if vc.flags.env != "" {
		envVarsPrefix += "_" + vc.flags.env
	}

	vc.v.SetEnvPrefix(envVarsPrefix)
	vc.v.AutomaticEnv()
}

func (vc *viperConfigurator) readFlags() error {
	for _, flag := range []string{"engine", "database", "user", "password", "host", "port", "table", "missingdowns"} {
		err := vc.v.BindPFlag(flag, migrateCmd.PersistentFlags().Lookup(flag))
		if err != nil {
			return errors.Wrapf(err, "can't bind flag %s", flag)
		}
	}
	return nil
}
