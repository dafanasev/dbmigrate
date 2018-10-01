package main

import (
	"os"
	"testing"

	"github.com/dafanasev/dbmigrate"
)

func TestMain(m *testing.M) {
	removeTempStuff()
	createTempStuff()
	code := m.Run()
	removeTempStuff()
	os.Exit(code)
}

func createTempStuff() {
	os.Mkdir(dbmigrate.MigrationsDir, os.ModePerm)
}

func removeTempStuff() {
	os.RemoveAll(dbmigrate.MigrationsDir)
	os.Remove("test.db")
}
