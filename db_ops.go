package migrate

import (
	"database/sql"
	"fmt"
	"time"
)

type dbWrapper interface {
	driverName() string
	dsnFormat() string
	
	beforeQuery()
	placeholder() string
	
	createDB() error
	dropDB() error
	
	getCurrentVersion() (time.Time, error)
	insertVersion(ts time.Time) error
	deleteVersion(ts time.Time) error
}


type baseSQLWrapper struct {
	Credentials
	db *sql.DB
}

func (w *baseSQLWrapper) driverName() string {
	return w.DriverName
}

func (w *baseSQLWrapper) beforeQuery() {
}

func (w *baseSQLWrapper) placeholder() string {
	return "?"
}

func (w *baseSQLWrapper) createDB() error {
	return nil
}

func (w *baseSQLWrapper) dropDB() error {
	return nil
}

func (w *baseSQLWrapper) getCurrentVersion() (time.Time, error) {
	return time.Now(), nil
}

func (w *baseSQLWrapper) insertVersion(ts time.Time) error {
	return nil
}

func (w *baseSQLWrapper) deleteVersion(ts time.Time) error {
	return nil
}


type postgresWrapper struct {
	baseSQLWrapper
	varCounter uint
}

func (w *postgresWrapper) dsnFormat() string {
	return ""
}

func (w *postgresWrapper) placeholder() string {
	w.varCounter++
	return fmt.Sprintf("$%d", w.varCounter)
}

func (w *postgresWrapper) beforeQuery() {
	w.varCounter = 0
}


type mySQLWrapper struct {
	baseSQLWrapper
}

func (w *mySQLWrapper) dsnFormat() string {
	return ""
}


type sqliteWrapper struct {
	baseSQLWrapper
}

func (w *sqliteWrapper) dsnFormat() string {
	return ""
}
