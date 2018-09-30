package dbmigrate

var providers = make(map[string]provider)

type provider interface {
	driver() string
	dsn(settings *Settings) (string, error)
	hasTableQuery() string
}

type placeholdersProvider interface {
	setPlaceholders(string) string
}

type defaultProvider struct{}

func (p *defaultProvider) hasTableQuery() string {
	return "SELECT table_name FROM information_schema.tables WHERE table_name = ?"
}
