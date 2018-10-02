package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/dafanasev/dbmigrate"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_commands(t *testing.T) {
	os.Mkdir(dbmigrate.MigrationsDir, os.ModePerm)

	migrator, _ := dbmigrate.NewMigrator(&dbmigrate.Settings{Engine: "sqlite", Database: "test.db"})
	defer migrator.Close()

	testFn := func(migrator *dbmigrate.Migrator, pattern string, len int, engines []string, args ...string) {
		os.Mkdir(dbmigrate.MigrationsDir, os.ModePerm)
		err := generateMigration(migrator, engines, args...)
		require.NoError(t, err)
		matches, _ := filepath.Glob(filepath.Join(dbmigrate.MigrationsDir, pattern))
		assert.Len(t, matches, len)
		os.RemoveAll(dbmigrate.MigrationsDir)
	}

	testFn(migrator, "*first_migration.*.sql", 2, []string{}, "first", "migration")
	testFn(migrator, "*current_engine_migration.*.sqlite.sql", 2, []string{enginesNoOptDefVal}, "current", "engine", "migration")
	testFn(migrator, "*all_engines_migration.*.*.sql", 6, []string{"all"}, "all", "engines", "migration")
	testFn(migrator, "*two_engines_migration.*.*.sql", 4, []string{"sqlite", "postgres"}, "two", "engines", "migration")

	os.Mkdir(dbmigrate.MigrationsDir, os.ModePerm)
	err := generateMigration(migrator, []string{"nodb"}, "wrong", "engine", "migration")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "is not exists/supported")
	os.RemoveAll(dbmigrate.MigrationsDir)
}
