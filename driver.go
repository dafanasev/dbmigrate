package migrate

import (
	"database/sql"
	"time"
	
	"github.com/pkg/errors"
	
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

var (
	errDBNameNotProvided = errors.New("db name is not provided")
	errUserNotProvided   = errors.New("user is not provided")
)

type driver struct {
	credentials *Credentials
	db          *sql.DB
	provider
	placeholdersProvider
}

func newDriver(cr *Credentials, provider provider) *driver {
	d := &driver{
		credentials: cr,
		provider:    provider,
	}
	if pp, ok := d.provider.(placeholdersProvider); ok {
		d.placeholdersProvider = pp
	}
	return d
}

func (d *driver) open() error {
	dsn, err := d.provider.dsn(d.credentials)
	if err != nil {
		return err
	}
	
	d.db, err = sql.Open(d.credentials.DriverName, dsn)
	if err != nil {
		return errors.Wrap(err, "can't open database")
	}
	
	return nil
}

func (d *driver) close() error {
	err := d.db.Close()
	if err != nil {
		return errors.Wrap(err, "Error shutting down migrator")
	}
	return nil
}

func (d *driver) setPlaceholders(s string) string {
	if d.placeholdersProvider == nil {
		return s
	}
	return d.placeholdersProvider.setPlaceholders(s)
}

func (d *driver) hasTable() (bool, error) {
	var table string
	err := d.db.QueryRow(d.setPlaceholders(d.provider.hasTableQuery()), d.credentials.MigrationsTable).Scan(&table)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	
	return true, nil
}

func (d *driver) createMigrationsTable() error {
	_, err := d.db.Exec(d.setPlaceholders("CREATE TABLE ? (version TIMESTAMP NOT NULL, PRIMARY KEY(version));"), d.credentials.MigrationsTable)
	if err != nil {
		return errors.Wrapf(err, "Can't create migrations table")
	}
	return nil
}

func (d *driver) lastMigrationTimestamp() (time.Time, error) {
	var v time.Time
	err := d.db.QueryRow(d.setPlaceholders("SELECT version FROM ? ORDER BY version DESC LIMIT 1"), d.credentials.MigrationsTable).Scan(&v)
	if err != nil {
	    return time.Time{}, errors.Wrap(err,"can't get last migration version")
	}
	return v, nil
}

func (d *driver) appliedMigrationsTimestamps() ([]time.Time, error) {
	rows, err := d.db.Query(d.setPlaceholders("SELECT version FROM ? ORDER BY version ASC"), d.credentials.MigrationsTable)
	if err != nil {
		return nil, errors.Wrap(err, "can't get applied migrations versions")
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

func (d *driver) insertMigrationTimestamp(version time.Time) error {
	_, err := d.db.Exec(d.setPlaceholders("INSERT INTO ? (migration) VALUES (?)"), d.credentials.MigrationsTable, version)
	if err != nil {
		return errors.Wrap(err, "can't insert migration version")
	}
	return nil
}

func (d *driver) deleteMigrationTimestamp(version time.Time) error {
	_, err := d.db.Exec(d.setPlaceholders("DELETE FROM ? WHERE migration = ?"), d.credentials.MigrationsTable, version)
	if err != nil {
		return errors.Wrap(err, "can't delete migration version")
	}
	return nil
}