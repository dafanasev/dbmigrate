package migrate

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/pkg/errors"
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
	_, err := w.db.Exec(fmt.Sprintf("CREATE TABLE %s (version VARCHAR(14) NOT NULL, PRIMARY KEY(version));", w.settings.MigrationsTable))
	if err != nil {
		return errors.Wrap(err, "can't create migrations table")
	}
	return nil
}

func (w *dbWrapper) lastMigrationTimestamp() (time.Time, error) {
	var version string
	err := w.db.QueryRow(fmt.Sprintf("SELECT version FROM %s ORDER BY version DESC LIMIT 1", w.settings.MigrationsTable)).Scan(&version)
	if err == sql.ErrNoRows {
		return time.Time{}, nil
	}
	if err != nil {
		return time.Time{}, errors.Wrap(err, "can't select last migration version from database")
	}
	ts, _ := time.Parse(timestampFormat, version)
	return ts, nil
}

func (w *dbWrapper) appliedMigrationsTimestamps(order string) ([]time.Time, error) {
	rows, err := w.db.Query(fmt.Sprintf("SELECT version FROM %s ORDER BY version %s", w.settings.MigrationsTable, order))
	if err != nil {
		return nil, errors.Wrap(err, "can't get applied migrations versions")
	}
	defer rows.Close()

	var tss []time.Time
	var v string
	for rows.Next() {
		err = rows.Scan(&v)
		if err != nil {
			return nil, errors.Wrap(err, "can't scan migration version row")
		}
		ts, _ := time.Parse(timestampFormat, v)
		tss = append(tss, ts)
	}
	return tss, nil
}

func (w *dbWrapper) insertMigrationTimestamp(ts time.Time) error {
	_, err := w.db.Exec(w.setPlaceholders(fmt.Sprintf("INSERT INTO %s (version) VALUES (?)", w.settings.MigrationsTable)), ts.Format(timestampFormat))
	if err != nil {
		return errors.Wrap(err, "can't insert migration")
	}
	return nil
}

func (w *dbWrapper) deleteMigrationTimestamp(ts time.Time) error {
	_, err := w.db.Exec(w.setPlaceholders(fmt.Sprintf("DELETE FROM %s WHERE version = ?", w.settings.MigrationsTable)), ts.Format(timestampFormat))
	if err != nil {
		return errors.Wrap(err, "can't delete migration")
	}
	return nil
}

func (w *dbWrapper) execQuery(query string) error {
	if strings.TrimSpace(query) == "" {
		// TODO: option to skip empty migrations instead of returning an error
		return errors.New("empty query")
	}

	// using transactions, although only mysql doesn't supports DDL ones
	tx, err := w.db.Begin()
	if err != nil {
		log.Fatal(errors.Wrap(err, "can't begin transaction"))
	}

	// split queries, because mysql driver can't exec multiple queries at once
	queries := strings.Split(query, ";")
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

	err = tx.Commit()
	if err != nil {
		return errors.Wrap(err, "can't commit transaction")
	}

	return nil
}
