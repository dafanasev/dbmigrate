package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/dafanasev/dbmigrate"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_migrate(t *testing.T) {
	createTestMigrations()
	defer os.RemoveAll(dbmigrate.MigrationsDir)
	defer os.Remove("test.db")

	migrator, _ := dbmigrate.NewMigrator(&dbmigrate.Settings{
		Engine: "sqlite", Database: "test.db",
		MigrationsCh: make(chan *dbmigrate.Migration), ErrorsCh: make(chan error),
	})
	defer migrator.Close()

	n, err := migrate(migrator, dbmigrate.AllSteps)
	require.NoError(t, err)
	assert.Equal(t, 3, n)

	n, err = migrate(migrator, dbmigrate.AllSteps)
	require.NoError(t, err)
	assert.Zero(t, n)

	ioutil.WriteFile(filepath.Join(dbmigrate.MigrationsDir, "20180918202020.error.up.sqlite.sql"), []byte("error"), os.ModePerm)
	_, err = migrate(migrator, dbmigrate.AllSteps)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "can't migrate")
	os.Remove(filepath.Join(dbmigrate.MigrationsDir, "20180918202020.error.up.sqlite.sql"))

}
