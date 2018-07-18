package migrate

import (
	"database/sql"
	"fmt"
	
	"github.com/pkg/errors"
)

func init() {
	drivers["mysql"] = &mysqlDriver{}
}

type mysqlDriver struct {
	commonDriver
}

func (d *mysqlDriver) name() string {
	return "mysql"
}

func (d *mysqlDriver) open() error {
	dsn, err := d.dsn()
	if err != nil {
	    return err
	}
	
	d.db, err = sql.Open("mysql", dsn)
	if err != nil {
		return errors.Wrap(err, "can't open database")
	}
	
	return nil
}

func (d *mysqlDriver) dsn() (string, error) {
	if d.cr.DBName == "" {
		return "", errDBNameNotProvided
	}
	
	if d.cr.User == "" {
		return "", errUserNotProvided
	}
	
	up := d.cr.User
	
	if d.cr.Passwd != "" {
		up += ":" + d.cr.Passwd
	}
	
	host := d.cr.Host
	if host == "" {
		host = "127.0.0.1"
	}
	port := d.cr.Port
	if port == "" {
		port = "3306"
	}
	
	return fmt.Sprintf("%s@tcp(%s:%s)/%s?parseTime=true", up, host, port, d.cr.DBName), nil
}