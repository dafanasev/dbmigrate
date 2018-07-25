package migrate

import (
	"database/sql"
	"log"
	"strings"
	"time"
	
	"github.com/pkg/errors"
	
	// drivers, imported only to exec their init functions
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

var (
	errDBNameNotProvided = errors.New("database name name is not provided")
	errUserNotProvided   = errors.New("user is not provided")
)

type dbWrapper struct {
	settings *Settings
	db       *sql.DB
	provider
	placeholdersProvider
}

func newDBWrapper(settings *Settings, provider provider) *dbWrapper {
	w := &dbWrapper{
		settings: settings,
		provider: provider,
	}
	if pp, ok := w.provider.(placeholdersProvider); ok {
		w.placeholdersProvider = pp
	}
	return w
}

func (w *dbWrapper) open() error {
	dsn, err := w.provider.dsn(w.settings)
	if err != nil {
		return err
	}
	
	w.db, err = sql.Open(w.provider.driverName(), dsn)
	if err != nil {
		return errors.Wrap(err, "can't open database")
	}
	
	return nil
}

func (w *dbWrapper) close() error {
	err := w.db.Close()
	if err != nil {
		return errors.Wrap(err, "can't close db")
	}
	return nil
}

func (w *dbWrapper) setPlaceholders(s string) string {
	if w.placeholdersProvider == nil {
		return s
	}
	return w.placeholdersProvider.setPlaceholders(s)
}

func (w *dbWrapper) hasMigrationsTable() (bool, error) {
	var table string
	err := w.db.QueryRow(w.setPlaceholders(w.provider.hasTableQuery()), w.settings.MigrationsTable).Scan(&table)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	
	return true, nil
}

func (w *dbWrapper) createMigrationsTable() error {
	_, err := w.db.Exec(w.setPlaceholders("CREATE TABLE ? (version TIMESTAMP NOT NULL, PRIMARY KEY(version));"), w.settings.MigrationsTable)
	if err != nil {
		return errors.Wrap(err, "can't create migrations table")
	}
	return nil
}

func (w *dbWrapper) lastMigrationData() (time.Time, error) {
	var version time.Time
	err := w.db.QueryRow(w.setPlaceholders("SELECT version, FROM ? ORDER BY version DESC LIMIT 1"), w.settings.MigrationsTable).Scan(&version)
	if err != nil {
	    return time.Time{}, errors.Wrap(err,"can't select last Migration version from database")
	}
	return version, nil
}

func (w *dbWrapper) appliedMigrationsTimestamps(order string) ([]time.Time, error) {
	rows, err := w.db.Query(w.setPlaceholders("SELECT version FROM ? ORDER BY version ?"), w.settings.MigrationsTable, order)
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

func (w *dbWrapper) insertMigration(version time.Time) error {
	_, err := w.db.Exec(w.setPlaceholders("INSERT INTO ? (version) VALUES (?)"), w.settings.MigrationsTable, version)
	if err != nil {
		return errors.Wrap(err, "can't insert migration")
	}
	return nil
}

func (w *dbWrapper) deleteMigration(version time.Time) error {
	_, err := w.db.Exec(w.setPlaceholders("DELETE FROM ? WHERE version = ?"), w.settings.MigrationsTable, version)
	if err != nil {
		return errors.Wrap(err, "can't delete migration")
	}
	return nil
}

func (w *dbWrapper) execQuery(query string) error {
	// using transactions, although only postgres supports DDL ones
	tx, err := w.db.Begin()
	if err != nil {
		log.Fatal(errors.Wrap(err, "can't execute migration query, can't begin transaction"))
	}
	
	// split queries, because mysql driver can't exec multiple queries at once
	queries := strings.Split(query, ";")
	if len(queries) > 0 {
		for _, q := range queries {
			q := strings.TrimSpace(q)
			if q != "" {
				_, err = tx.Exec(q + ";")
				if err != nil {
					tx.Rollback()
					return errors.Wrapf(err, "can't execute query %s", q)
				}
			}
		}
	}
	
	err = tx.Commit()
	if err != nil {
		log.Fatal(errors.Wrap(err, "can't commit transaction"))
	}
	
	return nil
}