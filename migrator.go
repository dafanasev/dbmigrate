package dbmigrate

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/pkg/errors"
)

// AllSteps specifies that all migrations should be applied
const AllSteps = 0

// Migrator is the end user interface for all library operations
type Migrator struct {
	// Settings used by migrator
	*Settings
	// dbWrapper wraps database operations
	dbWrapper  *dbWrapper
	projectDir string
}

// NewMigrator creates new Migrator instance
func NewMigrator(settings *Settings) (*Migrator, error) {
	if settings.Engine == "" {
		return nil, errors.New("database engine not specified")
	}
	if settings.Database == "" {
		return nil, errors.New("database name not specified")
	}

	if settings.MigrationsTable == "" {
		settings.MigrationsTable = "migrations"
	}

	m := &Migrator{Settings: settings}

	wd, err := os.Getwd()
	if err != nil {
		return nil, errors.Wrap(err, "can't get working directory")
	}

	m.projectDir, err = FindProjectDir(wd)
	if err != nil {
		return nil, err
	}

	p, ok := providers[settings.Engine]
	if !ok {
		return nil, errors.Errorf("unknown database engine %s", settings.Engine)
	}

	m.dbWrapper = newDBWrapper(settings, p)
	err = m.dbWrapper.open()
	if err != nil {
		return nil, errors.Wrap(err, "can't create database connection")
	}

	// create migrations table if it is not exists yet
	migrationsTableExists, err := m.dbWrapper.hasMigrationsTable()
	if err != nil {
		return nil, errors.Wrap(err, "can't check if migrations table exists")
	}
	if !migrationsTableExists {
		err = m.dbWrapper.createMigrationsTable()
		if err != nil {
			return nil, errors.Wrap(err, "can't create migrations table")
		}
	}

	return m, nil
}

// GenerateMigration generates up and down migrations with given name for given engine
func (m *Migrator) GenerateMigration(descr string, engine string) ([]string, error) {
	if engine != "" {
		if _, ok := providers[engine]; !ok {
			return nil, errors.Errorf("database engine %s is not exists/supported", engine)
		}
	}

	var fpaths []string

	ts := time.Now().UTC()
	re := regexp.MustCompile(`\s+`)

	for _, direction := range []string{"up", "down"} {
		parts := []string{ts.Format(TimestampFormat), re.ReplaceAllString(strings.TrimSpace(strings.ToLower(descr)), "_"), direction}
		if engine != "" {
			parts = append(parts, engine)
		}
		parts = append(parts, "sql")

		fname := strings.Join(parts, ".")
		fpath := filepath.Join(MigrationsDir, fname)

		if FileExists(fpath) {
			return nil, errors.Errorf("migration file %s already exists", fname)
		}

		_, err := os.Create(fpath)
		if err != nil {
			return nil, errors.Wrapf(err, "can't create migration file %s", fname)
		}

		fpaths = append(fpaths, fpath)
	}
	return fpaths, nil
}

// Close frees resources acquired by migrator
func (m *Migrator) Close() error {
	err := m.dbWrapper.close()
	if err != nil {
		return errors.Wrap(err, "error closing migrator")
	}

	if m.MigrationsCh != nil {
		close(m.MigrationsCh)
		m.MigrationsCh = nil
	}

	if m.ErrorsCh != nil {
		close(m.ErrorsCh)
		m.ErrorsCh = nil
	}

	return nil
}

// Migrate applies all unapplied migrations
func (m *Migrator) Migrate() (int, error) {
	return m.MigrateSteps(AllSteps)
}

// MigrateSteps applies the number of migrations specified by the steps variable, returning number of applied migrations
func (m *Migrator) MigrateSteps(steps int) (int, error) {
	migrations, err := m.unappliedMigrations()
	if err != nil {
		return 0, errors.Wrap(err, "can't find migrations")
	}

	if steps == AllSteps || steps > len(migrations) {
		steps = len(migrations)
	}

	appliedAt := time.Now().UTC()
	for i, migration := range migrations[:steps] {
		migration.AppliedAt = appliedAt
		err = m.run(migration)
		if err != nil {
			return i, errors.Wrapf(err, "can't execute migration %s", migration.FileName())
		}
	}
	return len(migrations[:steps]), nil
}

// Rollback rolls back last migration operation
func (m *Migrator) Rollback() (int, error) {
	return m.RollbackSteps(0)
}

// RollbackSteps rolls back the number of migrations specified by the steps variable
func (m *Migrator) RollbackSteps(steps int) (int, error) {
	appliedMigrationsData, err := m.dbWrapper.appliedMigrationsData("applied_at DESC, version DESC")
	if err != nil {
		return 0, errors.Wrap(err, "can't rollback")
	}

	if steps == 0 {
		steps, err = m.dbWrapper.countMigrationsInLastBatch()
		if err != nil {
			return 0, errors.Wrap(err, "can't get migrations in last batch")
		}
	}

	if steps > len(appliedMigrationsData) {
		steps = len(appliedMigrationsData)
	}

	// get migrations list
	var migrations []*Migration
	for _, migrationData := range appliedMigrationsData[:steps] {
		migration, err := m.getMigration(migrationData.version, DirectionDown)
		if err == nil {
			migrations = append(migrations, migration)
		} else {
			err = errors.Wrapf(err, "can't get migration for version %s", migrationData.version.Format(PrintTimestampFormat))
			if !m.AllowMissingDowns {
				return 0, err
			}
			if m.ErrorsCh != nil {
				m.ErrorsCh <- err
			}
		}
	}

	// and run them
	for i, migration := range migrations {
		err = m.run(migration)
		if err != nil {
			return i, errors.Wrapf(err, "can't execute migration %s", migration.FileName())
		}
	}
	return len(migrations), nil
}

// run executes given migration
func (m *Migrator) run(migration *Migration) error {
	fpath := filepath.Join(MigrationsDir, migration.FileName())

	query, err := ioutil.ReadFile(fpath)
	if err != nil {
		return errors.Wrapf(err, "can't read migration %s", migration.FileName())
	}

	if strings.TrimSpace(string(query)) == "" {
		// optionally allow empty down migrations, notifying about it
		if migration.Direction == DirectionUp || (migration.Direction == DirectionDown && !m.AllowMissingDowns) {
			return errors.New("empty query")
		}
		if m.ErrorsCh != nil {
			m.ErrorsCh <- errors.New("empty query")
		}
		return nil
	}

	// insert/delete migration data from the database after executing migration
	afterFunc := func(tx *sql.Tx) error {
		err = m.dbWrapper.insertMigrationData(migration.Version, migration.AppliedAt, tx)
		if err != nil {
			return errors.Wrapf(err, "can't insert version for migration %s", migration.FileName())
		}
		return nil
	}
	if migration.Direction == DirectionDown {
		afterFunc = func(tx *sql.Tx) error {
			err := m.dbWrapper.deleteMigrationVersion(migration.Version, tx)
			if err != nil {
				return errors.Wrapf(err, "can't delete timestamp %s from db", migration.Version.Format(PrintTimestampFormat))
			}
			return nil
		}
	}

	err = m.dbWrapper.execMigrationQueries(string(query), afterFunc)
	if err != nil {
		return errors.Wrapf(err, "can't exec query for migration %s", migration.FileName())
	}

	if m.MigrationsCh != nil {
		m.MigrationsCh <- migration
	}

	return nil
}

// LatestVersionMigration returns the migration that has the most recent version (which is not necessarily the last applied one)
func (m *Migrator) LatestVersionMigration() (*Migration, error) {
	ts, err := m.dbWrapper.latestMigrationVersion()
	if err != nil {
		return nil, errors.Wrap(err, "can't get latest migration")
	}

	if ts == (time.Time{}) {
		return nil, nil
	}

	migration, err := m.getMigration(ts, DirectionUp)
	if err != nil {
		return nil, errors.Wrapf(err, "can't get latest migration with version %s", ts.Format(TimestampFormat))
	}

	return migration, nil
}

// LastAppliedMigration returns the migration which was applied last
func (m *Migrator) LastAppliedMigration() (*Migration, error) {
	ts, err := m.dbWrapper.lastAppliedMigrationVersion()
	if err != nil {
		return nil, errors.Wrap(err, "can't get last applied migration")
	}

	if ts == (time.Time{}) {
		return nil, nil
	}

	migration, err := m.getMigration(ts, DirectionUp)
	if err != nil {
		return nil, errors.Wrapf(err, "can't get last applied migration with version %s", ts.Format(TimestampFormat))
	}

	return migration, nil
}

// Status returns applied a timestamp for each migration or nil if it is not set
// along with the migration's name and version
func (m *Migrator) Status() ([]*Migration, error) {
	foundMigrations, err := m.findMigrations(DirectionUp)
	if err != nil {
		return nil, errors.Wrap(err, "can't get migrations")
	}

	appliedMigrationsData, err := m.dbWrapper.appliedMigrationsData("version ASC")
	if err != nil {
		return nil, err
	}

	for _, m := range foundMigrations {
		for _, migrationData := range appliedMigrationsData {
			if m.Version == migrationData.version {
				m.AppliedAt = migrationData.appliedAt
				break
			}
		}
	}

	return foundMigrations, nil
}

// findMigrations finds all valid migrations in the dbmigrations dir
func (m *Migrator) findMigrations(direction Direction) ([]*Migration, error) {
	var migrations []*Migration
	migrationsDirPath := filepath.Join(m.projectDir, MigrationsDir)

	err := filepath.Walk(migrationsDirPath, func(mpath string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if mpath != migrationsDirPath && info.IsDir() {
			return filepath.SkipDir
		}
		if info.IsDir() {
			return nil
		}

		migration, err := migrationFromFileName(info.Name())
		if err != nil {
			return nil
		}

		if migration.Direction != direction {
			return nil
		}

		// migration that should be run on isSpecific dbWrapper only
		if migration.Engine != "" && migration.Engine != m.Engine {
			return nil
		}

		migrations = append(migrations, migration)
		return nil
	})
	if err != nil {
		return nil, errors.Wrap(err, "can't scan migrations directory")
	}

	sort.Sort(byVersion(migrations))

	// return an error if there are multiple migrations with the same version
	for i := 0; i < len(migrations)-1; i++ {
		if migrations[i].Version == migrations[i+1].Version {
			return nil, errors.Errorf("migrations with %s are duplicated", migrations[i].Version.Format(PrintTimestampFormat))
		}
	}

	return migrations, nil
}

// unappliedMigrations returns list of all unapplied migrations
func (m *Migrator) unappliedMigrations() ([]*Migration, error) {
	foundMigrations, err := m.findMigrations(DirectionUp)
	if err != nil {
		return nil, errors.Wrap(err, "can't get migrations")
	}

	appliedMigrationsData, err := m.dbWrapper.appliedMigrationsData("version ASC")
	if err != nil {
		return nil, err
	}

	var unappliedMigrations []*Migration
	for _, m := range foundMigrations {
		found := false
		for _, migrationData := range appliedMigrationsData {
			if m.Version == migrationData.version {
				found = true
				break
			}
		}
		if !found {
			unappliedMigrations = append(unappliedMigrations, m)
		}
	}

	return unappliedMigrations, nil
}

// getMigration tries to find migration file and create Migration instance for given version and direction,
// returning an error if there are multiple ones
func (m *Migrator) getMigration(version time.Time, direction Direction) (*Migration, error) {
	versionStr := version.Format(TimestampFormat)

	pattern := filepath.FromSlash(fmt.Sprintf("%s/%s.*.%v.sql", MigrationsDir, versionStr, direction))
	files, _ := filepath.Glob(pattern)

	if len(files) == 0 {
		pattern = filepath.FromSlash(fmt.Sprintf("%s/%s.*.%v.%s.sql", MigrationsDir, versionStr, direction, m.Engine))
		files, _ = filepath.Glob(pattern)
	}

	if len(files) == 0 {
		return nil, errors.Errorf("migration %v with version %s does not exist", direction, versionStr)
	}
	if len(files) > 1 {
		return nil, errors.Errorf("got %d %v migration with version %s, should be only one", len(files), direction, versionStr)
	}

	migration, err := migrationFromFileName(filepath.Base(files[0]))
	if err != nil {
		return nil, err
	}

	return migration, nil
}
