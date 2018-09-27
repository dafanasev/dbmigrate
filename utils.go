package dbmigrate

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

// DirExists checks if directory at path is exist
func DirExists(dirpath string) bool {
	stats, err := os.Stat(dirpath)
	if os.IsNotExist(err) || !stats.IsDir() {
		return false
	}
	return true
}

func FileExists(fpath string) bool {
	stats, err := os.Stat(fpath)
	if os.IsNotExist(err) || stats.IsDir() {
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
	switch strings.ToLower(s) {
	case "up":
		return directionUp, nil
	case "down":
		return directionDown, nil
	default:
		return directionError, errors.Errorf("can't parse direction from string %s", s)
	}
}

func isRootDir(dir string) bool {
	// second check is for windows
	if dir == "/" || dir == strings.Split(dir, string(filepath.Separator))[0] {
		return true
	}
	return false
}
