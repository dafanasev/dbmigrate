package migrate

import (
	"fmt"
	"strings"
	
	"github.com/pkg/errors"
)

var supportedDrivers = []string{"postgres", "mysql", "sqlite"}
var (
	errDBNameNotProvided = errors.New("db name is not provided")
	errUserNotProvided   = errors.New("user is not provided")
)

type placeholdersProvider interface {
	setPlaceholders(string) string
}


type dsnProvider interface {
	dsn(cr *Credentials) (string, error)
}


type postgresProvider struct {}

func (w *postgresProvider) dsn(cr *Credentials) (string, error) {
	kvs := make([]string, 2)
	
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

func (w *postgresProvider) setPlaceholders(s string) string {
	counter := 1
	for strings.Index(s, "?") != -1 {
		s = strings.Replace("s", "?", fmt.Sprintf("$%d", counter), 1)
		counter++
	}
	return s
}


type mysqlProvider struct {}

func (w *mysqlProvider) dsn(cr *Credentials) (string, error) {
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


type sqliteProvider struct {}

func (w *sqliteProvider) dsn(cr *Credentials) (string, error) {
	if cr.DBName == "" {
		return "", errDBNameNotProvided
	}
	
	return "./" + cr.DBName, nil
}
