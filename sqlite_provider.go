package migrate

import "path/filepath"

func init() {
	providers["sqlite"] = &sqliteProvider{}
}

type sqliteProvider struct {}

func (d *sqliteProvider) dsn(cr *Credentials) (string, error) {
	if cr.DBName == "" {
		return "", errDBNameNotProvided
	}
	
	if filepath.IsAbs(cr.DBName) {
		return cr.DBName, nil
	}
	
	return "./" + cr.DBName, nil
}

