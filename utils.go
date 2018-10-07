package dbmigrate

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

// DirExists checks if directory at path exists
func DirExists(dirpath string) bool {
	stats, err := os.Stat(dirpath)
	if os.IsNotExist(err) || !stats.IsDir() {
		return false
	}
	return true
}

// FileExists checks if file at path exists
func FileExists(fpath string) bool {
	stats, err := os.Stat(fpath)
	if os.IsNotExist(err) || stats.IsDir() {
		return false
	}
	return true
}

// DirectionFromString tries to build Direction from string,
// checking for valid ones and returning an error if check was unsuccessful
func DirectionFromString(s string) (Direction, error) {
	switch strings.ToLower(s) {
	case "up":
		return DirectionUp, nil
	case "down":
		return DirectionDown, nil
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

// EngineExists checks is specified database provider exists
func EngineExists(engine string) bool {
	_, ok := providers[engine]
	return ok
}

// Engines returns list of supported database engines
func Engines() []string {
	var engines []string
	for engine := range providers {
		engines = append(engines, engine)
	}
	return engines
}
