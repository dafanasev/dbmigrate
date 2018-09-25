package migrate

const migrationsDir = "dbmigrations"

const (
	timestampFormat      = "20060102150405"
	printTimestampFormat = "2006.01.02 15:04:05"
)

type Settings struct {
	Driver            string
	DB                string
	Host              string
	Port              int
	User              string
	Passwd            string
	MigrationsTable   string
	AllowMissingDowns bool
	// migrationsCh is the channel for applied migrations
	MigrationsCh chan<- *Migration
	// errorsChan is the channel for errors
	ErrorsCh chan<- error
}

type Direction int

const (
	directionError = Direction(iota)
	directionUp
	directionDown
)

func (d Direction) String() string {
	if d == directionUp {
		return "up"
	}
	return "down"
}
