package dbmigrate

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/pkg/errors"
)

var (
	errDBNameNotProvided = errors.New("database name is not provided")
	errUserNotProvided   = errors.New("user is not provided")
)

// dbWrapper encapsulates all database access operations
type dbWrapper struct {
	*Settings
	db *sql.DB
	provider
	placeholdersProvider
}

// executor is an interface to exec sql so we could pass db instance as well as tx one
type executor interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
}

// migrationData holds info about migration from migrations table
type migrationData struct {
	version   time.Time
	appliedAt time.Time
}

// newDBWrapper creates new dbWrapper instance
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

// open creates new database connection
func (w *dbWrapper) open() error {
	dsn, err := w.provider.dsn(w.Settings)
	if err != nil {
		return err
	}

	w.db, err = sql.Open(w.provider.driver(), dsn)
	if err != nil {
		return errors.Wrap(err, "can't open database")
	}

	return nil
}

// close shuts down database connection
func (w *dbWrapper) close() error {
	err := w.db.Close()
	if err != nil {
		return errors.Wrap(err, "can't close db")
	}
	return nil
}

// setPlaceholders calls placeholdersProvider's placeholdersProvider if it is provided
func (w *dbWrapper) setPlaceholders(s string) string {
	if w.placeholdersProvider == nil {
		return s
	}
	return w.placeholdersProvider.setPlaceholders(s)
}

// hasMigrationsTable checks if the table with applied migrations data already exists
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

// createMigrationsTable creates new table for applied migrations data
func (w *dbWrapper) createMigrationsTable() error {
	_, err := w.db.Exec(fmt.Sprintf(
		"CREATE TABLE %s (version VARCHAR(14) NOT NULL, applied_at VARCHAR(14) NOT NULL, PRIMARY KEY(version));", w.MigrationsTable))
	if err != nil {
		return errors.Wrap(err, "can't create migrations table")
	}
	return nil
}

// latestMigrationVersion returns a timestamp for latest migration version
func (w *dbWrapper) latestMigrationVersion() (time.Time, error) {
	version, err := w.getAttrOrderedBy("version", "version DESC")
	if err != nil {
		return time.Time{}, errors.Wrap(err, "can't select latest migration version from database")
	}
	return version, nil
}

// lastAppliedMigrationVersion returns a latest applied migration version timestamp
func (w *dbWrapper) lastAppliedMigrationVersion() (time.Time, error) {
	version, err := w.getAttrOrderedBy("version", "applied_at DESC, version DESC")
	if err != nil {
		return time.Time{}, errors.Wrap(err, "can't select last applied migration version from database")
	}
	return version, nil
}

// getAttrOrderedBy returns first attr ordered by order
func (w *dbWrapper) getAttrOrderedBy(attr string, order string) (time.Time, error) {
	var result string
	err := w.db.QueryRow(fmt.Sprintf("SELECT %s FROM %s ORDER BY %s LIMIT 1", attr, w.MigrationsTable, order)).Scan(&result)
	if err == sql.ErrNoRows {
		return time.Time{}, nil
	}
	if err != nil {
		return time.Time{}, err
	}
	ts, _ := time.Parse(TimestampFormat, result)
	return ts, nil
}

// appliedMigrationsData returns all data from migrations table ordered by provided order variable
func (w *dbWrapper) appliedMigrationsData(order string) ([]*migrationData, error) {
	rows, err := w.db.Query(fmt.Sprintf("SELECT version, applied_at FROM %s ORDER BY %s", w.MigrationsTable, order))
	if err != nil {
		return nil, errors.Wrap(err, "can't get applied migrations versions")
	}
	defer rows.Close()

	var mds []*migrationData
	var version, appliedAt string
	for rows.Next() {
		err = rows.Scan(&version, &appliedAt)
		if err != nil {
			return nil, errors.Wrap(err, "can't scan migrations table's row")
		}

		md := &migrationData{}
		md.version, _ = time.Parse(TimestampFormat, version)
		md.appliedAt, _ = time.Parse(TimestampFormat, appliedAt)
		mds = append(mds, md)
	}
	return mds, nil
}

// insertMigrationData inserts data for applied migration
func (w *dbWrapper) insertMigrationData(version time.Time, appliedAtTs time.Time, executor executor) error {
	if executor == nil {
		executor = w.db
	}

	_, err := executor.Exec(w.setPlaceholders(fmt.Sprintf("INSERT INTO %s (version, applied_at) VALUES (?, ?)", w.MigrationsTable)),
		version.UTC().Format(TimestampFormat), appliedAtTs.UTC().Format(TimestampFormat))
	if err != nil {
		return errors.Wrap(err, "can't insert migration")
	}

	return nil
}

// countMigrationsInLastBatch returns number of migrations which were applied during the last database operation
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

// deleteMigrationVersion removes database row with given migration version
func (w *dbWrapper) deleteMigrationVersion(version time.Time, executor executor) error {
	if executor == nil {
		executor = w.db
	}

	_, err := executor.Exec(w.setPlaceholders(fmt.Sprintf(
		"DELETE FROM %s WHERE version = ?", w.MigrationsTable)),
		version.UTC().Format(TimestampFormat))
	if err != nil {
		return errors.Wrap(err, "can't delete migration")
	}

	return nil
}

// execMigrationQueries executes queries from the migration file, calling func after
func (w *dbWrapper) execMigrationQueries(query string, afterFunc func(tx *sql.Tx) error) error {
	// using transactions, although only postgres supports supports DDL ones
	tx, err := w.db.Begin()
	if err != nil {
		return errors.Wrap(err, "can't begin transaction")
	}

	// split queries and exec them one by one, because mysql driver can't exec multiple queries using one Exec call
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
