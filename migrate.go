package migrate

type database int
const (
	databaseNone = database(iota)
	databasePostgres
	databaseMySQL
	databaseSQLite
)

type direction int
const (
	directionError   = direction(iota)
	directionUp
	directionDown
)

type Conf struct {
	Direction string
	Driver string
}