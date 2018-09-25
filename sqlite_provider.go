package dbmigrate

import (
	"os"
	"path/filepath"

	"github.com/pkg/errors"

	// sqlite driver, imported only to exec init function
	_ "github.com/mattn/go-sqlite3"
)

func init() {
	providers["sqlite"] = &sqliteProvider{}
}

type sqliteProvider struct{}

func (p *sqliteProvider) driver() string {
	return "sqlite3"
}

func (p *sqliteProvider) dsn(settings *Settings) (string, error) {
	if settings.Database == "" {
		return "", errDBNameNotProvided
	}

	if filepath.IsAbs(settings.Database) {
		return settings.Database, nil
	}

	wd, err := os.Getwd()
	if err != nil {
		return "", errors.Wrap(err, "can't get working directory")
	}
	projectDir, err := FindProjectDir(wd)
	if err != nil {
		return "", err
	}

	return filepath.Join(projectDir, settings.Database), nil
}

func (p *sqliteProvider) hasTableQuery() string {
	return "SELECT name FROM sqlite_master WHERE type = 'table' AND name = ?"
}
