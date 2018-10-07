package dbmigrate

import (
	"path/filepath"

	"github.com/pkg/errors"
)

// MigrationsDir is the directory to store migrations
const MigrationsDir = "dbmigrations"

const (
	// TimestampFormat defines format for migration versioning
	// and is used in migrations file names and db table
	TimestampFormat = "20060102150405"
	// PrintTimestampFormat defines format for printing timestamps
	PrintTimestampFormat = "2006.01.02 15:04:05"
)

// Settings used by Migrator
type Settings struct {
	Engine            string
	Database          string
	User              string
	Password          string
	Host              string
	Port              int
	MigrationsTable   string
	AllowMissingDowns bool
	// MigrationsCh is the channel for applied migrations
	MigrationsCh chan *Migration
	// ErrorsChan is the channel for errors that happened during the work but are not fatal
	ErrorsCh chan error
}

// Direction specifies if migration is used to migrate or rollback schema
type Direction int

const (
	directionError = Direction(iota)
	// DirectionUp is the direction to migrate schema
	DirectionUp
	// DirectionDown is the direction to rollback schema
	DirectionDown
)

// String returns string representation of direction
func (d Direction) String() string {
	if d == DirectionUp {
		return "up"
	}
	return "down"
}

// FindProjectDir recursively finds project dir (the one that has dbmigrations subdir)
func FindProjectDir(fromDir string) (string, error) {
	if DirExists(filepath.Join(fromDir, MigrationsDir)) {
		return fromDir, nil
	}

	if isRootDir(fromDir) {
		return "", errors.New("project dir not found")
	}

	return FindProjectDir(filepath.Dir(fromDir))
}
