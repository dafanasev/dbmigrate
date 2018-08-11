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

func (p *postgresProvider) driverName() string {
	return "postgres"
}

func (p *postgresProvider) dsn(settings *Settings) (string, error) {
	var kvs []string

	if settings.DBName == "" {
		return "", errDBNameNotProvided
	}
	kvs = append(kvs, "dbname="+settings.DBName)

	if settings.User == "" {
		return "", errUserNotProvided
	}
	kvs = append(kvs, "user="+settings.User)

	if settings.Passwd != "" {
		kvs = append(kvs, "password="+settings.Passwd)
	}

	if settings.Host != "" {
		kvs = append(kvs, "host="+settings.Host)
	}

	if settings.Port != 0 {
		kvs = append(kvs, fmt.Sprintf("port=%d", settings.Port))
	}

	kvs = append(kvs, "sslmode=disable")

	return strings.Join(kvs, " "), nil
}

func (p *postgresProvider) setPlaceholders(s string) string {
	counter := 0
	for strings.Contains(s, "?") {
		counter++
		s = strings.Replace(s, "?", fmt.Sprintf("$%d", counter), 1)
	}
	return s
}
