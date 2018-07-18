package migrate

import "fmt"

func init() {
	providers["mysql"] = &mysqlProvider{}
}

type mysqlProvider struct {}

func (d *mysqlProvider) dsn(cr *Credentials) (string, error) {
	if cr.DBName == "" {
		return "", errDBNameNotProvided
	}
	
	if cr.User == "" {
		return "", errUserNotProvided
	}
	
	host := cr.Host
	if host == "" {
		host = "127.0.0.1"
	}
	port := cr.Port
	if port == "" {
		port = "3306"
	}
	
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", cr.User, cr.Passwd, port, cr.DBName), nil
}