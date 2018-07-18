package migrate

import (
	"fmt"
	"strings"
	
)

func init() {
	providers["postgres"] = &postgresDriver{}
}

type postgresDriver struct {}

func (d *postgresDriver) dsn(cr *Credentials) (string, error) {
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

func (d *postgresDriver) setPlaceholders(s string) string {
	counter := 1
	for strings.Index(s, "?") != -1 {
		s = strings.Replace("s", "?", fmt.Sprintf("$%d", counter), 1)
		counter++
	}
	return s
}
