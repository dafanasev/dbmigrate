package migrate

var providers = make(map[string]provider)

type dsnProvider interface {
	dsn(settings *Settings) (string, error)
}

type placeholdersProvider interface {
	setPlaceholders(string) string
}

type sqlProvider interface {
	hasTableQuery() string
}

type provider interface {
	driverName() string
	dsnProvider
	sqlProvider
}

type defaultProvider struct{}

func (p *defaultProvider) hasTableQuery() string {
	return "SELECT table_name FROM information_schema.tables WHERE table_name = ?"
}
