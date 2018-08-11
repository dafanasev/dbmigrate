package migrate

var timestampFormat = "20060102150405"
var printTimestampFormat = "2006.01.02 15:04:05"

type Settings struct {
	DriverName      string
	DBName          string
	Host            string
	Port            int
	User            string
	Passwd          string
	MigrationsDir   string
	MigrationsTable string
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
