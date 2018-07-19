package migrate

import (
	"os"
	"path/filepath"
	
	"github.com/pkg/errors"
)

func init() {
	providers["sqlite"] = &sqliteProvider{}
}

type sqliteProvider struct {}

func (p *sqliteProvider) dsn(cr *Credentials) (string, error) {
	if cr.DBName == "" {
		return "", errDBNameNotProvided
	}
	
	if filepath.IsAbs(cr.DBName) {
		return cr.DBName, nil
	}
	
	dir, err := os.Getwd()
	if err != nil {
	    return "", errors.Wrap(err, "can't get working directory")
	}
	
	dbPath := "./" + cr.DBName
	for !fileExists(filepath.Join(dir, cr.DBName)) {
		if dir == "/" {
			return "", errors.Wrap(err, "database file is not found")
		}
		dir = filepath.Dir(dir)
		dbPath = "../" + dbPath
	}
	
	return dbPath, nil
}

func (p *sqliteProvider) hasTableQuery() string {
	return "SELECT name FROM sqlite_master WHERE type = 'table' AND name = ?"
}