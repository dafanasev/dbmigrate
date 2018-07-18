package migrate

import (
	"database/sql"
	"path/filepath"
	
	"github.com/pkg/errors"
)

func init() {
	drivers["sqlite"] = &sqliteDriver{}
}

type sqliteDriver struct {
	commonDriver
}

func (d *sqliteDriver) name() string {
	return "sqlite"
}

func (d *sqliteDriver) open() error {
	dsn, err := d.dsn()
	if err != nil {
		return err
	}
	
	d.db, err = sql.Open("sqlite", dsn)
	if err != nil {
		return errors.Wrap(err, "can't open database")
	}
	
	return nil
}

func (d *sqliteDriver) dsn() (string, error) {
	if d.cr.DBName == "" {
		return "", errDBNameNotProvided
	}
	
	if filepath.IsAbs(d.cr.DBName) {
		return d.cr.DBName, nil
	}
	
	return "./" + d.cr.DBName, nil
}

