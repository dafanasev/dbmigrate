package migrate

var timestampFormat  = "20060102150405"

type Direction int

const (
	directionError = Direction(iota)
	DirectionUp
	DirectionDown
)

func (d Direction) String() string {
	var s string
	switch d {
	case DirectionUp:
		s = "up"
	case DirectionDown:
		s = "down"
	}
	return s
}

type Credentials struct {
	DriverName      string
	DBName          string
	Host            string
	Port            string
	User            string
	Passwd          string
	MigrationsDir   string
	MigrationsTable string
}