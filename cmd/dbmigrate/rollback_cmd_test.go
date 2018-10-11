package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/dafanasev/dbmigrate"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_rollback(t *testing.T) {
	createTestMigrations()
	defer os.RemoveAll(dbmigrate.MigrationsDir)
	defer os.Remove("test.db")

	migrator, _ := dbmigrate.NewMigrator(&dbmigrate.Settings{
		Engine: "sqlite", Database: "test.db",
		MigrationsCh: make(chan *dbmigrate.Migration), ErrorsCh: make(chan error),
	})
	defer migrator.Close()

	migrate(migrator, dbmigrate.AllSteps)
	n, err := rollback(migrator, 0)
	require.NoError(t, err)
	assert.Equal(t, 3, n)

	migrate(migrator, dbmigrate.AllSteps)
	n, err = rollback(migrator, 1)
	require.NoError(t, err)
	assert.Equal(t, 1, n)
	n, err = rollback(migrator, 0)
	require.NoError(t, err)
	assert.Equal(t, 2, n)
	n, err = rollback(migrator, 0)
	require.NoError(t, err)
	assert.Equal(t, 0, n)

	os.Rename(filepath.Join(dbmigrate.MigrationsDir, "20180918200453.first.down.sql"), "./20180918200453.first.down.sql")
	migrate(migrator, dbmigrate.AllSteps)
	n, err = rollback(migrator, 0)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "can't rollback")
	assert.Equal(t, 0, n)

	migrate(migrator, dbmigrate.AllSteps)
	migrator.AllowMissingDowns = true
	n, err = rollback(migrator, 0)
	require.NoError(t, err)
	assert.Equal(t, 2, n)
	os.Rename("./20180918200453.first.down.sql", filepath.Join(dbmigrate.MigrationsDir, "20180918200453.first.down.sql"))
}
