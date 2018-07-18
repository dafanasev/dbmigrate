package migrate

import "path/filepath"

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
	
	return "./" + cr.DBName, nil
}

func (p *sqliteProvider) hasTableQuery() string {
	return "SELECT name FROM sqlite_master WHERE type = 'table' AND name = ?"
}