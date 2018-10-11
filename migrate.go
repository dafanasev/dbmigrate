// Package dbmigrate is a sql database migration tool.
//
// dbmigrate can be used both as a CLI application and as a Go package, does not use any DSL for migrations,
// just plain old SQL we all know and love so it is compatible with any framework and programming language.
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
	// Engines is the used RDBMS. Currently PostgreSQL, MySQL and SQLite are supported
	Engine   string
	Database string
	User     string
	Password string
	Host     string
	Port     int
	// MigrationsTable is the database table to store applied migrations data
	MigrationsTable string
	// AllowMissingDowns flag specifies if Migrator should allow empty or missing down migrations files
	// which means that there will be no rollback for the corresponding up migrations and that this is ok
	AllowMissingDowns bool
	// MigrationsCh is the channel for applied migrations
	MigrationsCh chan *Migration
	// ErrorsChan is the channel for errors that happened during the work but are not fatal
	ErrorsCh chan error
}

// Direction specifies if migration is used to migrate or rollback schema
type Direction int

const (
	// directionError is just to start the iota cycle and to specify that the default Direction value is incorrect
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
		return "", errors.New("dbmigrations dir not found")
	}

	return FindProjectDir(filepath.Dir(fromDir))
}
