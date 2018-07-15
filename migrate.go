package migrate

type Direction int

type Credentials struct {
	DriverName string
	Host       string
	Port       string
	DBName     string
	User       string
	Passwd     string
}

var supportedDrivers = []string{"postgres", "mysql", "sqlite"}

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