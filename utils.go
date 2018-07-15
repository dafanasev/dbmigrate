package migrate

import (
	"os"
	"strings"
)

// dirExists checks if directory at path is exist,
// the path should be the absolute one
func dirExists(fpath string) bool {
	stats, err := os.Stat(fpath)
	if os.IsNotExist(err) || !stats.IsDir() {
		return false
	}
	return true
}

// isValidString checks whether the given string is the one of passed options,
// case insensible
func isValidString(s string, validOptions []string) bool {
	for _, v := range validOptions {
		if v == strings.ToLower(s) {
			return true
		}
	}
	return false
}

// DirectionFromString tries to build Direction from string,
// checking for valid ones
func DirectionFromString(s string) (Direction, error) {
	s = strings.ToLower(s)
	if !isValidString(s, []string{"up", "down"}) {
		return directionError, nil
	}
	
	var d Direction
	switch s {
	case "up":
		d = DirectionUp
	case "down":
		d = DirectionDown
	}
	return d, nil
}