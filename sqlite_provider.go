package migrate

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

func (p *sqliteProvider) driverName() string {
	return "sqlite3"
}

func (p *sqliteProvider) dsn(settings *Settings) (string, error) {
	if settings.DB == "" {
		return "", errDBNameNotProvided
	}

	if filepath.IsAbs(settings.DB) {
		return settings.DB, nil
	}

	dir, err := os.Getwd()
	if err != nil {
		return "", errors.Wrap(err, "can't get working directory")
	}
	dbPath := settings.DB
	for !dirExists(filepath.Join(dir, migrationsDir)) {
		if isRootDir(dir) {
			return "", errors.New("project root is not found")
		}
		dir = filepath.Dir(dir)
		dbPath = filepath.FromSlash("../") + dbPath
	}

	return dbPath, nil
}

func (p *sqliteProvider) hasTableQuery() string {
	return "SELECT name FROM sqlite_master WHERE type = 'table' AND name = ?"
}
