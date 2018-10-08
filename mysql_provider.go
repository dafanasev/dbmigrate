package dbmigrate

import (
	"fmt"

	// mysql driver, imported only to exec init function
	_ "github.com/go-sql-driver/mysql"
)

func init() {
	providers["mysql"] = &mysqlProvider{}
}

// mysqlProvider is the provider interface implementation for mysql
type mysqlProvider struct {
	defaultProvider
}

func (p *mysqlProvider) driver() string {
	return "mysql"
}

func (p *mysqlProvider) dsn(settings *Settings) (string, error) {
	if settings.Database == "" {
		return "", errDBNameNotProvided
	}

	if settings.User == "" {
		return "", errUserNotProvided
	}

	host := settings.Host
	if host == "" {
		host = "127.0.0.1"
	}
	port := settings.Port
	if port == 0 {
		port = 3306
	}

	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true", settings.User, settings.Password, host, port, settings.Database), nil
}
