package migrate

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_sqliteProviderExist(t *testing.T) {
	_, ok := providers["sqlite"]
	assert.True(t, ok)
}

func Test_sqliteProvider_driver(t *testing.T) {
	assert.Equal(t, "sqlite3", (&sqliteProvider{}).driver())
}

func Test_sqliteProvider_dsn(t *testing.T) {
	p := &sqliteProvider{}
	s := &Settings{}

	_, err := p.dsn(s)
	assert.EqualError(t, err, errDBNameNotProvided.Error())

	os.Rename(migrationsDir, "!"+migrationsDir)
	s.Database = "test.db"
	_, err = p.dsn(s)
	assert.Error(t, err)

	os.Rename("!"+migrationsDir, migrationsDir)

	// from project root dir
	for i, dir := range []string{"/some/absolute/path", ".", "..", "test"} {
		s.Database = filepath.Join(dir, "test.db")
		dsn, err := p.dsn(s)
		assert.NoError(t, err)
		wd, _ := os.Getwd()
		if i == 0 {
			assert.Equal(t, s.Database, dsn)
		} else {
			assert.Equal(t, filepath.Join(wd, dir, "test.db"), dsn)
		}
	}

	// from project subdir
	wd, _ := os.Getwd()
	os.Chdir(filepath.Join(wd, "dbmigrations"))

	for _, dir := range []string{".", "..", "test"} {
		s.Database = filepath.Join(dir, "test.db")
		dsn, err := p.dsn(s)
		assert.NoError(t, err)
		wd, _ := os.Getwd()
		assert.Equal(t, filepath.Join(wd, "..", dir, "test.db"), dsn)
	}

	s.Database = "/some/absolute/path/test.db"
	dsn, err := p.dsn(s)
	assert.NoError(t, err)
	assert.Equal(t, "/some/absolute/path/test.db", dsn)

	os.Chdir(wd)
}

func Test_sqliteProvider_hasTableQuery(t *testing.T) {
	p := &sqliteProvider{}
	assert.Contains(t, p.hasTableQuery(), "sqlite")
}
