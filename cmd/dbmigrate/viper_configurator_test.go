package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/dafanasev/dbmigrate"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	crypt "github.com/xordataexchange/crypt/config"
)

var dbmigrateConfigJSON = []byte(`{"engine": "sqlite", "database": "test.db"}`)
var envScopedDbmigrateConfigJSON = []byte(`{"test": {"engine": "sqlite", "database": "test.db"}}`)

func createConfigFiles() {
	os.Mkdir(dbmigrate.MigrationsDir, os.ModePerm)
	ioutil.WriteFile("dbmigrate.json", dbmigrateConfigJSON, os.ModePerm)
	ioutil.WriteFile("dbmigrate.test.json", envScopedDbmigrateConfigJSON, os.ModePerm)
}

func removeConfigFiles() {
	os.RemoveAll(dbmigrate.MigrationsDir)
	os.Remove("dbmigrate.json")
	os.Remove("dbmigrate.test.json")
}

func Test_viperConfigurator_readConfigFile(t *testing.T) {
	createConfigFiles()
	defer removeConfigFiles()

	projectDir, _ := os.Getwd()

	vc := &viperConfigurator{viper: viper.New(), flags: &appFlags{configFile: "dbmigrate"}, projectDir: projectDir}
	err := vc.readConfigFile()
	require.NoError(t, err)
	for key, val := range map[string]string{"engine": "sqlite", "database": "test.db"} {
		assert.Equal(t, val, vc.viper.GetString(key))
	}

	vc = &viperConfigurator{viper: viper.New(), flags: &appFlags{configFile: "not_exist"}, projectDir: projectDir}
	err = vc.readConfigFile()
	require.NoError(t, err)
	for _, val := range []string{"engine", "database"} {
		assert.Empty(t, vc.viper.Get(val))
	}

	vc = &viperConfigurator{viper: viper.New(), flags: &appFlags{configFile: "dbmigrate.test"}, projectDir: projectDir}
	err = vc.readConfigFile()
	require.NoError(t, err)
	for key, val := range map[string]string{"engine": "sqlite", "database": "test.db"} {
		assert.Equal(t, val, vc.viper.GetString("test."+key))
	}
}

func Test_viperConfigurator_readKVS(t *testing.T) {
	kvsManagerCreationFnsMap := map[string]func([]string) (crypt.ConfigManager, error){
		"etcd":   crypt.NewStandardEtcdConfigManager,
		"consul": crypt.NewStandardConsulConfigManager,
	}

	for kvs, managerFn := range kvsManagerCreationFnsMap {
		var scheme string
		port := 2379
		if kvs == "etcd" {
			scheme = "http://"
		}
		if kvs == "consul" {
			port = 8500
		}

		cm, err := managerFn([]string{fmt.Sprintf("%s127.0.0.1:%d", scheme, port)})
		require.NoError(t, err)

		err = cm.Set("/configs/theservice", dbmigrateConfigJSON)
		require.NoError(t, err)

		vc := &viperConfigurator{viper: viper.New(), flags: &appFlags{kvsParamsStr: kvs + "://localhost/configs/theservice.json"}}
		err = vc.readKVS()
		require.NoError(t, err)
		assert.Equal(t, "sqlite", vc.viper.GetString("engine"))
	}
}

func Test_viperConfigurator_scopeToEnv(t *testing.T) {
	createConfigFiles()
	defer removeConfigFiles()

	projectDir, _ := os.Getwd()

	vc := &viperConfigurator{viper: viper.New(), flags: &appFlags{env: "test"}, projectDir: projectDir}
	vc.readConfigFile()
	vc.scopeToEnv()
	assert.Empty(t, vc.viper.GetString("engine"))

	vc = &viperConfigurator{viper: viper.New(), flags: &appFlags{env: "test", configFile: "dbmigrate.test"}, projectDir: projectDir}
	vc.readConfigFile()
	vc.scopeToEnv()
	assert.Equal(t, "sqlite", vc.viper.GetString("engine"))
}

func Test_viperConfigurator_readEnv(t *testing.T) {
	projectDir, _ := os.Getwd()

	for _, appName := range []string{"", "theservice"} {
		for _, env := range []string{"", "test"} {
			vc := &viperConfigurator{viper: viper.New(), flags: &appFlags{prefix: appName, env: env}, projectDir: projectDir}
			vc.readEnv()

			envVarsPrefix := "THESERVICE"
			if appName == "" {
				envVarsPrefix = "DBMIGRATE"
			}
			if env != "" {
				envVarsPrefix += "_TEST"
			}

			os.Setenv(envVarsPrefix+"_ENGINE", "sqlite")
			assert.Equalf(t, "sqlite", vc.viper.GetString("engine"), envVarsPrefix)
			os.Unsetenv(envVarsPrefix + "_ENGINE")
		}
	}
}

func Test_viperConfigurator_readFlags(t *testing.T) {
	vc := &viperConfigurator{viper: viper.New(), flags: &appFlags{}}
	require.NoError(t, vc.readFlags())
}

func Test_viperConfigurator_configure(t *testing.T) {
	createConfigFiles()
	defer removeConfigFiles()

	vc := &viperConfigurator{viper: viper.New(), flags: &appFlags{}}
	_, err := vc.configure()
	require.NoError(t, err)
}
