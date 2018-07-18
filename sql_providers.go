package migrate

var providers = make(map[string]provider)

type dsnProvider interface {
	dsn(cr *Credentials) (string, error)
}

type placeholdersProvider interface {
	setPlaceholders(string) string
}

type sqlProvider interface {

}

type provider interface {
	dsnProvider
	sqlProvider
}
