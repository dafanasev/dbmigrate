package migrate

type direction int

const (
	directionError = direction(iota)
	directionUp
	directionDown
)

func (d direction) String() string {
	var s string
	switch d {
	case directionUp:
		s = "up"
	case directionDown:
		s = "down"
	}
	return s
}

type Config struct {
	Driver string
	Direction string
	Steps uint
	DBName string
	User string
	Passwd string
}