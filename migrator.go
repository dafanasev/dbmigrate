package migrate

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

var ErrEmptyQuery = errors.New("empty query")

const allSteps = 0

type Migrator struct {
	*Settings
	// dbWrapper wraps db ops
	dbWrapper  *dbWrapper
	projectDir string
}

// NewMigrator returns migrator instance
func NewMigrator(settings *Settings) (*Migrator, error) {
	if settings.Driver == "" {
		return nil, errors.New("database driver not specified")
	}
	if settings.DB == "" {
		return nil, errors.New("database name not specified")
	}

	if settings.MigrationsDir == "" {
		settings.MigrationsDir = "migrations"
	}
	if settings.MigrationsTable == "" {
		settings.MigrationsTable = "migrations"
	}

	m := &Migrator{Settings: settings}

	provider, ok := providers[settings.Driver]
	if !ok {
		return nil, errors.Errorf("unknown database driver %s", settings.Driver)
	}

	m.dbWrapper = newDBWrapper(settings, provider)
	err := m.dbWrapper.open()
	if err != nil {
		return nil, errors.Wrap(err, "can't create database connection")
	}

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

	wd, err := os.Getwd()
	if err != nil {
		return nil, errors.Wrap(err, "can't get working directory")
	}

	m.projectDir, err = m.findProjectDir(wd)
	if err != nil {
		return nil, err
	}

	return m, nil
}

func (m *Migrator) GenerateMigration(descr string, isSpecific bool) ([]string, error) {
	var fnames []string

	ts := time.Now().UTC()
	re := regexp.MustCompile(`\s+`)

	for _, direction := range []string{"up", "down"} {
		parts := []string{ts.Format(timestampFormat), re.ReplaceAllString(strings.TrimSpace(strings.ToLower(descr)), "_")}
		if isSpecific {
			parts = append(parts, m.Driver)
		}
		parts = append(parts, direction, "sql")

		fname := strings.Join(parts, ".")
		fpath := filepath.Join(m.MigrationsDir, fname)

		if fileExists(fpath) {
			return nil, errors.Errorf("migration file %s already exists", fname)
		}

		_, err := os.Create(fpath)
		if err != nil {
			return nil, errors.Wrapf(err, "can't create migration file %s", fname)
		}

		fnames = append(fnames, fname)
	}
	return fnames, nil
}

// Close frees resources acquired by migrator
func (m *Migrator) Close() error {
	err := m.dbWrapper.close()
	if err != nil {
		return errors.Wrap(err, "error closing migrator")
	}
	return nil
}

func (m *Migrator) Up() (int, error) {
	return m.UpSteps(allSteps)
}

func (m *Migrator) UpSteps(steps int) (int, error) {
	migrations, err := m.unappliedMigrations()
	if err != nil {
		return 0, errors.Wrap(err, "can't find migrations")
	}

	if steps == allSteps || steps > len(migrations) {
		steps = len(migrations)
	}

	appliedAt := time.Now().UTC()
	for i, migration := range migrations[:steps] {
		migration.appliedAt = appliedAt
		err = m.run(migration)
		if err != nil {
			return i, errors.Wrapf(err, "can't execute migration %s", migration.HumanName())
		}
	}
	return len(migrations[:steps]), nil
}

func (m *Migrator) Down() (int, error) {
	steps, err := m.dbWrapper.countMigrationsInLastBatch()
	if err != nil {
		return 0, err
	}
	return m.DownSteps(steps)
}

func (m *Migrator) DownSteps(steps int) (int, error) {
	appliedMigrationsTimestamps, err := m.dbWrapper.appliedMigrationsTimestamps("DESC")
	if err != nil {
		return 0, errors.Wrap(err, "can't rollback")
	}

	if steps > len(appliedMigrationsTimestamps) {
		steps = len(appliedMigrationsTimestamps)
	}

	var migrations []*Migration
	for _, ts := range appliedMigrationsTimestamps[:steps] {
		migration, err := m.getMigration(ts, directionDown)
		if err == nil {
			migrations = append(migrations, migration)
		} else {
			err = errors.Wrapf(err, "can't get migration for version %s", ts.Format(printTimestampFormat))
			if !m.AllowMissingDowns {
				return 0, err
			}
			if m.ErrorsCh != nil {
				m.ErrorsCh <- err
			}
		}
	}

	for i, migration := range migrations {
		err = m.run(migration)
		if err != nil {
			return i, errors.Wrapf(err, "can't execute migration %s", migration.HumanName())
		}
	}
	return len(migrations), nil
}

func (m *Migrator) run(migration *Migration) error {
	fpath := filepath.Join(m.MigrationsDir, migration.fileName())

	query, err := ioutil.ReadFile(fpath)
	if err != nil {
		return errors.Wrapf(err, "can't read file for migration %s", migration.HumanName())
	}

	if strings.TrimSpace(string(query)) == "" {
		if migration.direction == directionUp || (migration.direction == directionDown && !m.AllowMissingDowns) {
			return ErrEmptyQuery
		}
		if m.ErrorsCh != nil {
			m.ErrorsCh <- ErrEmptyQuery
		}
		return nil
	}

	afterFunc := func(tx *sql.Tx) error {
		err = m.dbWrapper.insertMigrationVersion(migration.Version, migration.appliedAt, tx)
		if err != nil {
			return errors.Wrapf(err, "can't insert version for migration %s", migration.HumanName())
		}
		return nil
	}
	if migration.direction == directionDown {
		afterFunc = func(tx *sql.Tx) error {
			err := m.dbWrapper.deleteMigrationVersion(migration.Version, tx)
			if err != nil {
				return errors.Wrapf(err, "can't delete timestamp %s from db", migration.Version.Format(printTimestampFormat))
			}
			return nil
		}
	}

	err = m.dbWrapper.execMigrationQueries(string(query), afterFunc)
	if err != nil {
		return errors.Wrapf(err, "can't exec query for migration %s", migration.HumanName())
	}

	if m.MigrationsCh != nil {
		m.MigrationsCh <- migration
	}

	return nil
}

func (m *Migrator) LatestVersionMigration() (*Migration, error) {
	ts, err := m.dbWrapper.latestMigrationVersion()
	if err != nil {
		return nil, errors.Wrap(err, "can't get latest migration")
	}

	if ts == (time.Time{}) {
		return nil, nil
	}

	migration, err := m.getMigration(ts, directionUp)
	if err != nil {
		return nil, errors.Wrapf(err, "can't get latest migration with version %s", ts.Format(timestampFormat))
	}

	return migration, nil
}

func (m *Migrator) LastAppliedAtMigration() (*Migration, error) {
	ts, err := m.dbWrapper.lastAppliedMigrationVersion()
	if err != nil {
		return nil, errors.Wrap(err, "can't get last applied migration")
	}

	if ts == (time.Time{}) {
		return nil, nil
	}

	migration, err := m.getMigration(ts, directionUp)
	if err != nil {
		return nil, errors.Wrapf(err, "can't get last applied migration with %s", ts.Format(timestampFormat))
	}

	return migration, nil
}

// findProjectDir recursively find project dir (the one that has migrations subdir)
func (m *Migrator) findProjectDir(dir string) (string, error) {
	if dirExists(filepath.Join(dir, m.MigrationsDir)) {
		return dir, nil
	}

	if isRootDir(dir) {
		return "", errors.New("project dir not found")
	}

	return m.findProjectDir(filepath.Dir(dir))
}

// findMigrations finds all valid migrations in the migrations dir
func (m *Migrator) findMigrations(direction Direction) ([]*Migration, error) {
	var migrations []*Migration
	migrationsDirPath := filepath.Join(m.projectDir, m.MigrationsDir)

	filepath.Walk(migrationsDirPath, func(mpath string, info os.FileInfo, err error) error {
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

		if migration.direction != direction {
			return nil
		}

		// Migration that should be run on isSpecific dbWrapper only
		if migration.driverName != "" && migration.driverName != m.Driver {
			return nil
		}

		migrations = append(migrations, migration)
		return nil
	})

	sort.Sort(byTimestamp(migrations))
	for i := 0; i < len(migrations)-1; i++ {
		if migrations[i].Version == migrations[i+1].Version {
			return nil, errors.Errorf("migrations with %s are duplicated", migrations[i].Version.Format(printTimestampFormat))
		}
	}

	return migrations, nil
}

func (m *Migrator) unappliedMigrations() ([]*Migration, error) {
	migrations, err := m.findMigrations(directionUp)
	if err != nil {
		return nil, errors.Wrap(err, "can't get migrations")
	}

	appliedMigrationsTimestamps, err := m.dbWrapper.appliedMigrationsTimestamps("ASC")
	if err != nil {
		return nil, err
	}

	var unappliedMigrations []*Migration
	for _, m := range migrations {
		found := false
		for _, ts := range appliedMigrationsTimestamps {
			if m.Version == ts {
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

func (m *Migrator) getMigration(ts time.Time, direction Direction) (*Migration, error) {
	timestampStr := ts.Format(timestampFormat)

	pattern := filepath.FromSlash(fmt.Sprintf("%s/%s.*.%v.sql", m.MigrationsDir, timestampStr, direction))
	files, _ := filepath.Glob(pattern)

	if len(files) == 0 {
		pattern = filepath.FromSlash(fmt.Sprintf("%s/%s.*.%v.%s.sql", m.MigrationsDir, timestampStr, direction, m.Driver))
		files, _ = filepath.Glob(pattern)
	}

	if len(files) == 0 {
		return nil, errors.Errorf("migration %v with version %s does not exist", direction, timestampStr)
	}
	if len(files) > 1 {
		return nil, errors.Errorf("got %d %v migration with version %s, should be only one", len(files), direction, timestampStr)
	}

	migration, err := migrationFromFileName(filepath.Base(files[0]))
	if err != nil {
		return nil, err
	}

	return migration, nil
}
