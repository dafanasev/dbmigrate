package migrate

import (
	"fmt"
	"strings"
	
)

func init() {
	providers["postgres"] = &postgresProvider{}
}

type postgresProvider struct {
	defaultProvider
}

func (p *postgresProvider) dsn(cr *Credentials) (string, error) {
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

func (p *postgresProvider) setPlaceholders(s string) string {
	counter := 0
	for strings.Index(s, "?") != -1 {
		counter++
		s = strings.Replace(s, "?", fmt.Sprintf("$%d", counter), 1)
	}
	return s
}