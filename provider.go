package dbmigrate

// providers is the map where keys are database engines names and values are implementations of provider inerface
var providers = make(map[string]provider)

// provider is the interface for database engines specific stuff
type provider interface {
	// driver returns driver name string by the database/sql lib to connect to database
	driver() string
	// dsn returns database connection string
	dsn(settings *Settings) (string, error)
	// hasTableQuery returns SQL query to check if the table used to store migrations exists
	hasTableQuery() string
}

// placeholdersProvider is the interface to set database specific variables placeholders in a SQL string
type placeholdersProvider interface {
	// setPlaceholders sets database specific variables placeholders in a SQL string
	setPlaceholders(string) string
}

// defaultProvider is the default implementation of provider interface
type defaultProvider struct{}

func (p *defaultProvider) hasTableQuery() string {
	return "SELECT table_name FROM information_schema.tables WHERE table_name = ?"
}
