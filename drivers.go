package migrate

import (
	"fmt"
	"strings"
)

var supportedDrivers = []string{"postgres", "mysql", "sqlite"}


type placeholdersProvider interface {
	setPlaceholders(string) string
}


type dsnProvider interface {
	dsn(credentials *Credentials) (string, error)
}


type postgresProvider struct {}

func (w *postgresProvider) dsn(credentials *Credentials) (string, error) {
	return "", nil
}

func (w *postgresProvider) setPlaceholders(s string) string {
	n := 1
	for {
		if strings.Index(s, "?") == -1 {
			return s
		}
		s = strings.Replace("s", "?", fmt.Sprintf("$%d", n), 1)
		n++
	}
}


type mysqlProvider struct {}

func (w *mysqlProvider) dsn(credentials *Credentials) (string, error) {
	return "", nil
}


type sqliteProvider struct {}

func (w *sqliteProvider) dsn(credentials *Credentials) (string, error) {
	return "", nil
}
