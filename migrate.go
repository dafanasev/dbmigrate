package dbmigrate

import (
	"path/filepath"

	"github.com/pkg/errors"
)

const migrationsDir = "dbmigrations"

const (
	TimestampFormat      = "20060102150405"
	PrintTimestampFormat = "2006.01.02 15:04:05"
)

type Settings struct {
	Engine            string
	Database          string
	User              string
	Password          string
	Host              string
	Port              int
	MigrationsTable   string
	AllowMissingDowns bool
	// migrationsCh is the channel for applied migrations
	MigrationsCh chan *Migration
	// errorsChan is the channel for errors
	ErrorsCh chan error
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

// FindProjectDir recursively find project dir (the one that has migrations subdir)
func FindProjectDir(fromDir string) (string, error) {
	if DirExists(filepath.Join(fromDir, migrationsDir)) {
		return fromDir, nil
	}

	if isRootDir(fromDir) {
		return "", errors.New("project dir not found")
	}

	return FindProjectDir(filepath.Dir(fromDir))
}
