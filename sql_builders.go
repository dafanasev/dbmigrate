package migrate

import (
	"fmt"
	"time"
)

type sqlBuilder interface {
	beforeQuery()
	placeholder() string
	
	createDB() error
	dropDB() error
	
	getCurrentVersion() (time.Time, error)
	insertVersion(ts time.Time) error
	deleteVersion(ts time.Time) error
}

type baseSQLBuilder struct {
}

func (b *baseSQLBuilder) beforeQuery() {
}

func (b *baseSQLBuilder) placeholder() string {
	return "?"
}

func (b *baseSQLBuilder) createDB() error {
	return nil
}

func (b *baseSQLBuilder) dropDB() error {
	return nil
}

func (b *baseSQLBuilder) getCurrentVersion() (time.Time, error) {
	return time.Now(), nil
}

func (b *baseSQLBuilder) insertVersion(ts time.Time) error {
	return nil
}

func (b *baseSQLBuilder) deleteVersion(ts time.Time) error {
	return nil
}

type postgresBuilder struct {
	baseSQLBuilder
	varCounter uint
}

func (b *postgresBuilder) placeholder() string {
	b.varCounter++
	return fmt.Sprintf("$%d", b.varCounter)
}

func (b *postgresBuilder) beforeQuery() {
	b.varCounter = 0
}

type mySQLBuilder struct {
	baseSQLBuilder
}

type sqliteBuilder struct {
	baseSQLBuilder
}
