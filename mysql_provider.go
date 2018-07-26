package migrate

import "fmt"

func init() {
	providers["mysql"] = &mysqlProvider{}
}

type mysqlProvider struct {
	defaultProvider
}

func (p *mysqlProvider) driverName() string {
	return "mysql"
}

func (p *mysqlProvider) dsn(settings *Settings) (string, error) {
	if settings.DBName == "" {
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
	
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true", settings.User, settings.Passwd, host, port, settings.DBName), nil
}