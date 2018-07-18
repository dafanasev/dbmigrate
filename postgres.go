package migrate

import (
	"database/sql"
	"fmt"
	"strings"
	
	"github.com/pkg/errors"
)

func init() {
	drivers["postgres"] = &postgresDriver{}
}

type postgresDriver struct {
	commonDriver
}

func (d *postgresDriver) name() string {
	return "postgres"
}

func (d *postgresDriver) open() error {
	dsn, err := d.dsn()
	if err != nil {
		return err
	}
	
	d.db, err = sql.Open("postgres", dsn)
	if err != nil {
		return errors.Wrap(err, "can't open database")
	}
	
	return nil
}

func (d *postgresDriver) dsn() (string, error) {
	kvs := []string{}
	
	if d.cr.DBName == "" {
		return "", errDBNameNotProvided
	}
	kvs = append(kvs, "dbname=" + d.cr.DBName)
	
	if d.cr.User == "" {
		return "", errUserNotProvided
	}
	kvs = append(kvs, "user=" +  d.cr.User)
	
	if d.cr.Passwd != "" {
		kvs = append(kvs, "password=" + d.cr.Passwd)
	}
	
	if d.cr.Host != "" {
		kvs = append(kvs, "host=" + d.cr.Host)
	}
	
	if d.cr.Port != "" {
		kvs = append(kvs, "port=" + d.cr.Port)
	}
	
	return strings.Join(kvs, " "), nil
}

func (d *postgresDriver) setPlaceholders(s string) string {
	counter := 1
	for strings.Index(s, "?") != -1 {
		s = strings.Replace("s", "?", fmt.Sprintf("$%d", counter), 1)
		counter++
	}
	return s
}
