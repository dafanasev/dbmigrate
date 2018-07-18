package migrate

import (
	"fmt"
	"strings"
	
	"github.com/pkg/errors"
)

var drivers = make(map[string]driver)
var (
	errDBNameNotProvided = errors.New("db name is not provided")
	errUserNotProvided   = errors.New("user is not provided")
)

func init() {
	drivers["postgres"] = &postgresDriver{}
	drivers["mysql"]    = &mysqlDriver{}
	drivers["sqlite"]   = &sqliteDriver{}
}


type driver interface {
	dsn(cr *Credentials) (string, error)
}


type placeholdersProvider interface {
	setPlaceholders(string) string
}


type postgresDriver struct {}

func (w *postgresDriver) dsn(cr *Credentials) (string, error) {
	kvs := []string{}
	
	if cr.DBName == "" {
		return "", errDBNameNotProvided
	}
	kvs = append(kvs, "dbname=" + cr.DBName)
	
	if cr.User == "" {
		return "", errUserNotProvided
	}
	kvs = append(kvs, "user=" +  cr.User)
	
	if cr.Passwd != "" {
		kvs = append(kvs, "password=" + cr.Passwd)
	}
	
	if cr.Host != "" {
		kvs = append(kvs, "host=" + cr.Host)
	}
	
	if cr.Port != "" {
		kvs = append(kvs, "port=" + cr.Port)
	}
	
	return strings.Join(kvs, " "), nil
}

func (w *postgresDriver) setPlaceholders(s string) string {
	counter := 1
	for strings.Index(s, "?") != -1 {
		s = strings.Replace("s", "?", fmt.Sprintf("$%d", counter), 1)
		counter++
	}
	return s
}


type mysqlDriver struct {}

func (w *mysqlDriver) dsn(cr *Credentials) (string, error) {
	if cr.DBName == "" {
		return "", errDBNameNotProvided
	}
	
	if cr.User == "" {
		return "", errUserNotProvided
	}
	
	up := cr.User
	
	if cr.Passwd != "" {
		up += ":" + cr.Passwd
	}
	
	host := cr.Host
	if host == "" {
		host = "127.0.0.1"
	}
	port := cr.Port
	if port == "" {
		port = "3306"
	}
	
	return fmt.Sprintf("%s@tcp(%s:%s)/%s?parseTime=true", up, host, port, cr.DBName), nil
}


type sqliteDriver struct {}

func (w *sqliteDriver) dsn(cr *Credentials) (string, error) {
	if cr.DBName == "" {
		return "", errDBNameNotProvided
	}
	
	return "./" + cr.DBName, nil
}
