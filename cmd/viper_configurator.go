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
	viper      *viper.Viper
	flags      *appFlags
	projectDir string
}

func (vc *viperConfigurator) configure() (*viper.Viper, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, errors.Wrap(err, "can't get working directory")
	}

	vc.projectDir, err = dbmigrate.FindProjectDir(wd)
	if err != nil {
		return nil, err
	}

	err = vc.readConfigFile()
	if err != nil {
		return nil, err
	}
	if vc.flags.kvsParamsStr != "" {
		err = vc.readKVS()
		if err != nil {
			return nil, err
		}
	}
	if vc.flags.env != "" {
		vc.scopeViper()
	}
	vc.readEnv()
	vc.readFlags()

	return vc.viper, nil
}

func (vc *viperConfigurator) readConfigFile() error {
	vc.viper.AddConfigPath(vc.projectDir)
	vc.viper.SetConfigName(vc.flags.configFile)
	err := vc.viper.ReadInConfig()
	// if there is no config - it is not an error, we allow it
	if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
		return err
	}
	return nil
}

func (vc *viperConfigurator) readKVS() error {
	kvsParams, err := parseKVSConnectionString(vc.flags.kvsParamsStr)
	if err != nil {
		return errors.Wrap(err, "wrong key value store connection string")
	}

	kvsErrorString := fmt.Sprintf("can't connect to key value store using connection string %s", vc.flags.kvsParamsStr)
	if vc.flags.secretKeyRingPath != "" {
		err = vc.viper.AddSecureRemoteProvider(kvsParams.provider, kvsParams.formatEndpoint(), kvsParams.path, vc.flags.secretKeyRingPath)
		if err != nil {
			return errors.Wrapf(err, "%s and key ring path %s", kvsErrorString, vc.flags.secretKeyRingPath)
		}
	} else {
		err = vc.viper.AddRemoteProvider(kvsParams.provider, kvsParams.formatEndpoint(), kvsParams.path)
		if err != nil {
			return errors.Wrap(err, vc.flags.kvsParamsStr)
		}
	}

	vc.viper.SetConfigType(kvsParams.format)
	err = vc.viper.ReadRemoteConfig()
	if err != nil {
		return errors.Wrap(err, kvsErrorString)
	}
	return nil
}

func (vc *viperConfigurator) scopeViper() {
	if vc.viper.IsSet(vc.flags.env) {
		vc.viper = vc.viper.Sub(vc.flags.env)
	} else {
		vc.viper = viper.New()
	}
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

	vc.viper.SetEnvPrefix(envVarsPrefix)
	vc.viper.AutomaticEnv()
}

func (vc *viperConfigurator) readFlags() error {
	for _, flag := range []string{"engine", "database", "user", "password", "host", "port", "table", "missingdowns"} {
		err := vc.viper.BindPFlag(flag, migrateCmd.PersistentFlags().Lookup(flag))
		if err != nil {
			return errors.Wrapf(err, "can't bind flag %s", flag)
		}
	}
	return nil
}
