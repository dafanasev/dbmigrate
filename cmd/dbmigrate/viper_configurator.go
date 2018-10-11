package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/dafanasev/dbmigrate"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

// viperConfigurator is the struct which sole purpose is to create proper viper instance
type viperConfigurator struct {
	// initial viper, that can be substituted for
	viper      *viper.Viper
	flags      *appFlags
	projectDir string
}

// configure returns properly initialized viper instance
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
		vc.scopeToEnv()
	}
	vc.readEnv()
	vc.readFlags()

	return vc.viper, nil
}

// readConfigFile tries to read configuration from a file
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

// readKVS reads configuration from the key value store
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

// scopeToEnv returns new viper instance using environment flag
// new instance is either subviper of initial viper, if the needed key exists or entirely new clean viper
func (vc *viperConfigurator) scopeToEnv() {
	if vc.viper.IsSet(vc.flags.env) {
		vc.viper = vc.viper.Sub(vc.flags.env)
	} else {
		vc.viper = viper.New()
	}
}

// readEnv builds full prefix for env vars env and reads them
func (vc *viperConfigurator) readEnv() {
	var envVarsPrefix string
	if vc.flags.prefix != "" {
		envVarsPrefix = vc.flags.prefix
	} else {
		envVarsPrefix = filepath.Base(vc.projectDir)
	}
	if vc.flags.env != "" {
		envVarsPrefix += "_" + vc.flags.env
	}

	vc.viper.SetEnvPrefix(envVarsPrefix)
	vc.viper.AutomaticEnv()
}

// readFlags binds cobra flags to viper
func (vc *viperConfigurator) readFlags() error {
	for _, flag := range []string{"engine", "database", "user", "password", "host", "port", "table", "missingdowns"} {
		err := vc.viper.BindPFlag(flag, migrateCmd.PersistentFlags().Lookup(flag))
		if err != nil {
			return errors.Wrapf(err, "can't bind flag %s", flag)
		}
	}
	return nil
}
