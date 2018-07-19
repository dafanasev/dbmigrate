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

type db struct {
	settings *Settings
	db       *sql.DB
	provider
	placeholdersProvider
}

func newDB(settings *Settings, provider provider) *db {
	d := &db{
		settings: settings,
		provider: provider,
	}
	if pp, ok := d.provider.(placeholdersProvider); ok {
		d.placeholdersProvider = pp
	}
	return d
}

func (d *db) open() error {
	dsn, err := d.provider.dsn(d.settings)
	if err != nil {
		return err
	}
	
	d.db, err = sql.Open(d.settings.DriverName, dsn)
	if err != nil {
		return errors.Wrap(err, "can't open database")
	}
	
	return nil
}

func (d *db) close() error {
	err := d.db.Close()
	if err != nil {
		return errors.Wrap(err, "Error shutting down migrator")
	}
	return nil
}

func (d *db) setPlaceholders(s string) string {
	if d.placeholdersProvider == nil {
		return s
	}
	return d.placeholdersProvider.setPlaceholders(s)
}

func (d *db) hasTable() (bool, error) {
	var table string
	err := d.db.QueryRow(d.setPlaceholders(d.provider.hasTableQuery()), d.settings.MigrationsTable).Scan(&table)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	
	return true, nil
}

func (d *db) createMigrationsTable() error {
	_, err := d.db.Exec(d.setPlaceholders("CREATE TABLE ? (version TIMESTAMP NOT NULL, PRIMARY KEY(version));"), d.settings.MigrationsTable)
	if err != nil {
		return errors.Wrapf(err, "can't create migrations table")
	}
	return nil
}

func (d *db) lastMigrationTimestamp() (time.Time, error) {
	var v time.Time
	err := d.db.QueryRow(d.setPlaceholders("SELECT version FROM ? ORDER BY version DESC LIMIT 1"), d.settings.MigrationsTable).Scan(&v)
	if err != nil {
	    return time.Time{}, errors.Wrap(err,"can't select last migration version from database")
	}
	return v, nil
}

func (d *db) appliedMigrationsTimestamps() ([]time.Time, error) {
	rows, err := d.db.Query(d.setPlaceholders("SELECT version FROM ? ORDER BY version ASC"), d.settings.MigrationsTable)
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

func (d *db) insertMigrationTimestamp(version time.Time) error {
	_, err := d.db.Exec(d.setPlaceholders("INSERT INTO ? (migration) VALUES (?)"), d.settings.MigrationsTable, version)
	if err != nil {
		return errors.Wrap(err, "can't insert migration version")
	}
	return nil
}

func (d *db) deleteMigrationTimestamp(version time.Time) error {
	_, err := d.db.Exec(d.setPlaceholders("DELETE FROM ? WHERE migration = ?"), d.settings.MigrationsTable, version)
	if err != nil {
		return errors.Wrap(err, "can't delete migration version")
	}
	return nil
}