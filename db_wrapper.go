package migrate

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/pkg/errors"
)

var (
	errDBNameNotProvided = errors.New("database name name is not provided")
	errUserNotProvided   = errors.New("user is not provided")
)

type dbWrapper struct {
	*Settings
	db *sql.DB
	provider
	placeholdersProvider
}

type executor interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
}

func newDBWrapper(settings *Settings, provider provider) *dbWrapper {
	w := &dbWrapper{
		Settings: settings,
		provider: provider,
	}
	if pp, ok := w.provider.(placeholdersProvider); ok {
		w.placeholdersProvider = pp
	}

	return w
}

func (w *dbWrapper) open() error {
	dsn, err := w.provider.dsn(w.Settings)
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
	err := w.db.QueryRow(w.setPlaceholders(w.provider.hasTableQuery()), w.MigrationsTable).Scan(&table)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (w *dbWrapper) createMigrationsTable() error {
	_, err := w.db.Exec(fmt.Sprintf(
		"CREATE TABLE %s (version VARCHAR(14) NOT NULL, applied_at VARCHAR(14) NOT NULL, PRIMARY KEY(version));", w.MigrationsTable))
	if err != nil {
		return errors.Wrap(err, "can't create migrations table")
	}
	_, err = w.db.Exec("CREATE INDEX migrations_applied_at ON migrations (applied_at)")
	if err != nil {
		return errors.Wrap(err, "can't create applied_at index on migrations table")
	}
	return nil
}

func (w *dbWrapper) latestMigrationVersion() (time.Time, error) {
	ts, err := w.getAttrOrderedBy("version", "version DESC")
	if err != nil {
		return time.Time{}, errors.Wrap(err, "can't select latest migration version from database")
	}
	return ts, nil
}

func (w *dbWrapper) lastAppliedMigrationVersion() (time.Time, error) {
	ts, err := w.getAttrOrderedBy("version", "applied_at DESC, version DESC")
	if err != nil {
		return time.Time{}, errors.Wrap(err, "can't select last applied migration version from database")
	}
	return ts, nil
}

func (w *dbWrapper) getAttrOrderedBy(attr string, order string) (time.Time, error) {
	var result string
	err := w.db.QueryRow(fmt.Sprintf("SELECT %s FROM %s ORDER BY %s LIMIT 1", attr, w.MigrationsTable, order)).Scan(&result)
	if err == sql.ErrNoRows {
		return time.Time{}, nil
	}
	if err != nil {
		return time.Time{}, err
	}
	ts, _ := time.Parse(timestampFormat, result)
	return ts, nil
}

func (w *dbWrapper) appliedMigrationsTimestamps(order string) ([]time.Time, error) {
	rows, err := w.db.Query(fmt.Sprintf("SELECT version FROM %s ORDER BY %s", w.MigrationsTable, order))
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

func (w *dbWrapper) insertMigrationVersion(ts time.Time, appliedAtTs time.Time, executor executor) error {
	if executor == nil {
		executor = w.db
	}

	_, err := executor.Exec(w.setPlaceholders(fmt.Sprintf("INSERT INTO %s (version, applied_at) VALUES (?, ?)", w.MigrationsTable)),
		ts.UTC().Format(timestampFormat), appliedAtTs.UTC().Format(timestampFormat))
	if err != nil {
		return errors.Wrap(err, "can't insert migration")
	}

	return nil
}

func (w *dbWrapper) countMigrationsInLastBatch() (int, error) {
	var count int
	err := w.db.QueryRow(w.setPlaceholders("SELECT COUNT(*) FROM migrations GROUP BY applied_at ORDER BY applied_at DESC LIMIT 1")).Scan(&count)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	if err != nil {
		return 0, errors.Wrapf(err, "can't count migrations for last batch")
	}

	return count, nil
}

func (w *dbWrapper) deleteMigrationVersion(ts time.Time, executor executor) error {
	if executor == nil {
		executor = w.db
	}

	_, err := executor.Exec(w.setPlaceholders(fmt.Sprintf(
		"DELETE FROM %s WHERE version = ?", w.MigrationsTable)),
		ts.UTC().Format(timestampFormat))
	if err != nil {
		return errors.Wrap(err, "can't delete migration")
	}

	return nil
}

func (w *dbWrapper) execMigrationQueries(query string, afterFunc func(tx *sql.Tx) error) error {
	// using transactions, although only postgres supports supports DDL ones
	tx, err := w.db.Begin()
	if err != nil {
		return errors.Wrap(err, "can't begin transaction")
	}

	// split queries, because mysql driver can't exec multiple queries at once
	queries := strings.Split(query, ";")
	for _, q := range queries {
		q := strings.TrimSpace(q)
		if q != "" {
			_, err := tx.Exec(q + ";")
			if err != nil {
				tx.Rollback()
				return errors.Wrapf(err, "can't execute query %s", q)
			}
		}
	}

	err = afterFunc(tx)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = tx.Commit()
	if err != nil {
		return errors.Wrap(err, "can't commit transaction")
	}

	return nil
}
