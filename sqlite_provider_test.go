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

func Test_sqliteProvider_driverName(t *testing.T) {
	assert.Equal(t, "sqlite3", (&sqliteProvider{}).driverName())
}

func Test_sqliteProvider_dsn(t *testing.T) {
	p := &sqliteProvider{}
	s := &Settings{}

	_, err := p.dsn(s)
	assert.EqualError(t, err, errDBNameNotProvided.Error())

	s.MigrationsDir = "not_existed"
	s.DBName = "test.db"
	_, err = p.dsn(s)
	assert.Error(t, err)

	s.MigrationsDir = "migrations"

	// from project root dir
	for _, dir := range []string{".", "..", "test", "/some/absolute/path"} {
		s.DBName = filepath.Join(dir, "test.db")
		dsn, err := p.dsn(s)
		assert.NoError(t, err)
		assert.Equal(t, s.DBName, dsn)
	}

	// from project subdir
	wd, _ := os.Getwd()
	os.Chdir(filepath.Join(wd, "migrations"))

	for _, dir := range []string{".", "..", "test"} {
		s.DBName = filepath.Join(dir, "test.db")
		dsn, err := p.dsn(s)
		assert.NoError(t, err)
		assert.Equal(t, filepath.Join("..", s.DBName), dsn)
	}

	s.DBName = "/some/absolute/path/test.db"
	dsn, err := p.dsn(s)
	assert.NoError(t, err)
	assert.Equal(t, "/some/absolute/path/test.db", dsn)

	os.Chdir(wd)
}

func Test_sqliteProvider_hasTableQuery(t *testing.T) {
	p := &sqliteProvider{}
	assert.Contains(t, p.hasTableQuery(), "sqlite")
}
