package main

import (
	"bytes"
	"io"
	"os"
	"testing"
	"time"

	"github.com/dafanasev/dbmigrate"
	"github.com/stretchr/testify/assert"
)

func Test_status(t *testing.T) {
	os.Mkdir(dbmigrate.MigrationsDir, os.ModePerm)
	defer os.RemoveAll(dbmigrate.MigrationsDir)
	defer os.Remove("test.db")

	migrator, _ := dbmigrate.NewMigrator(&dbmigrate.Settings{
		Engine: "sqlite", Database: "test.db",
		MigrationsCh: make(chan *dbmigrate.Migration), ErrorsCh: make(chan error),
	})
	defer migrator.Close()

	statusWrapper := func() string {
		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w
		status(migrator)

		resultCh := make(chan string)
		go func() {
			var b bytes.Buffer
			io.Copy(&b, r)
			r.Close()
			resultCh <- b.String()
		}()
		os.Stdout = old
		w.Close()
		return <-resultCh
	}

	output := statusWrapper()
	assert.Contains(t, output, "No migrations exist yet")

	createTestMigrations()

	output = statusWrapper()
	assert.Contains(t, output, "NAME")
	assert.Contains(t, output, "No migrations were applied yet")
	assert.Contains(t, output, "Database schema is not up to date")
	assert.Contains(t, output, "Database schema is not up to date", " - ")
	assert.NotContains(t, output, time.Now().UTC().Format("2006.01.02"))
	for _, migrationName := range []string{"first", "second", "third"} {
		assert.Contains(t, output, migrationName)
	}

	migrate(migrator, 2)
	output = statusWrapper()
	assert.Contains(t, output, "NAME")
	assert.Contains(t, output, "Latest version migration")
	assert.Contains(t, output, "Last applied migration ")
	assert.Contains(t, output, "Database schema is not up to date")
	assert.Contains(t, output, "Database schema is not up to date", " - ")
	assert.Contains(t, output, time.Now().UTC().Format("2006.01.02"))

	migrate(migrator, dbmigrate.AllSteps)
	output = statusWrapper()
	assert.Contains(t, output, "NAME")
	assert.Contains(t, output, "Latest version migration")
	assert.Contains(t, output, "Last applied migration ")
	assert.Contains(t, output, "Database schema is up to date")
	assert.NotContains(t, output, "Database schema is not up to date", " - ")
	assert.Contains(t, output, time.Now().UTC().Format("2006.01.02"))
}
