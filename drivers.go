package migrate

import (
	"database/sql"
	"time"
	
	"github.com/pkg/errors"
)

var drivers = make(map[string]driver)
var (
	errDBNameNotProvided = errors.New("db name is not provided")
	errUserNotProvided   = errors.New("user is not provided")
)

type driver interface {
	name() string
	dsn() (string, error)
	open() error
	close() error
	setPlaceholders(string) string
}

type commonDriver struct {
	cr *Credentials
	db *sql.DB
}

func (d *commonDriver) close() error {
	err := d.db.Close()
	if err != nil {
		return errors.Wrap(err, "Error shutting down migrator")
	}
	return nil
}

func (d *commonDriver) setPlaceholders(s string) string {
	return s
}

func (d *commonDriver) findAppliedMigrationsVersions() ([]time.Time, error) {
	rows, err := d.db.Query("SELECT version FROM migrations")
	if err != nil {
		return nil, errors.Wrap(err, "can't get executed migrations versions")
	}
	defer rows.Close()
	
	vs := []time.Time{}
	var v time.Time
	for rows.Next() {
		err = rows.Scan(&v)
		if err != nil {
			return nil, errors.Wrap(err, "can't scan migration version row")
		}
		vs = append(vs, v)
	}
	return vs, nil
}

func (d *commonDriver) createMigrationsTable() error {
	_, err := d.db.Query(d.setPlaceholders("CREATE TABLE migrations (version timestamp NOT NULL, PRIMARY KEY(version));"))
	if err != nil {
		return errors.Wrapf(err, "Can't create migrations table")
	}
	return nil
}