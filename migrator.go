package migrate

import (
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

const allSteps = 0

type Migrator struct {
	// dir holding migrations
	migrationsDir string
	// migrations table
	migrationsTable string
	// project dir (the one that has migrationsDir as first level subdir)
	projectDir string
	dbWrapper  *dbWrapper
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

	m := &Migrator{migrationsDir: settings.MigrationsDir, migrationsTable: settings.MigrationsTable}

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
	fnames := []string{}

	ts := time.Now().UTC()
	re := regexp.MustCompile(`\s+`)

	for _, direction := range []string{"up", "down"} {
		parts := []string{ts.Format(timestampFormat), re.ReplaceAllString(strings.TrimSpace(strings.ToLower(descr)), "_")}
		if isSpecific {
			parts = append(parts, m.dbWrapper.settings.Driver)
		}
		parts = append(parts, direction, "sql")

		fname := strings.Join(parts, ".")
		fpath := filepath.Join(m.migrationsDir, fname)

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

	appliedAt := time.Now()
	// TODO: think about prints
	for i, migration := range migrations[:steps] {
		err = m.run(migration)
		if err != nil {
			return i, errors.Wrapf(err, "can't execute migration %s", migration.HumanName())
		}

		err = m.dbWrapper.insertMigrationTimestamp(migration.Timestamp, appliedAt)
		if err != nil {
			return i, errors.Wrapf(err, "can't insert timestamp for migration %s", migration.HumanName())
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
		// TODO: option to skip not found migration instead of returning the error
		if err != nil {
			return 0, errors.Wrapf(err, "can't get migration for timestamp %s", ts.Format(printTimestampFormat))
		}
		migrations = append(migrations, migration)
	}

	for i, migration := range migrations {
		err = m.run(migration)
		if err != nil {
			return i, errors.Wrapf(err, "can't execute migration %s", migration.HumanName())
		}
		err := m.dbWrapper.deleteMigrationTimestamp(migration.Timestamp)
		if err != nil {
			return i, errors.Wrapf(err, "can't delete timestamp %s from db", migration.Timestamp.Format(printTimestampFormat))
		}
	}
	return len(migrations), nil
}

func (m *Migrator) run(migration *Migration) error {
	fpath := filepath.Join(m.migrationsDir, migration.fileName())

	query, err := ioutil.ReadFile(fpath)
	if err != nil {
		return errors.Wrapf(err, "can't read file for migration %s", migration.HumanName())
	}

	err = m.dbWrapper.execQuery(string(query))
	if err != nil {
		return errors.Wrapf(err, "can't exec query for migration %s", migration.HumanName())
	}

	return nil
}

func (m *Migrator) LastMigration() (*Migration, error) {
	ts, err := m.dbWrapper.lastMigrationTimestamp()
	if err != nil {
		return nil, errors.Wrap(err, "can't get last migration")
	}

	if ts == (time.Time{}) {
		return nil, nil
	}

	migration, err := m.getMigration(ts, directionUp)
	if err != nil {
		return nil, errors.Wrapf(err, "can't get last migration with timestamp %s", ts.Format(timestampFormat))
	}

	return migration, nil
}

// findProjectDir recursively find project dir (the one that has migrations subdir)
func (m *Migrator) findProjectDir(dir string) (string, error) {
	if dirExists(filepath.Join(dir, m.migrationsDir)) {
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
	migrationsDirPath := filepath.Join(m.projectDir, m.migrationsDir)

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
		if migration.driverName != "" && migration.driverName != m.dbWrapper.settings.Driver {
			return nil
		}

		migrations = append(migrations, migration)
		return nil
	})

	sort.Sort(byTimestamp(migrations))
	for i := 0; i < len(migrations)-1; i++ {
		if migrations[i].Timestamp == migrations[i+1].Timestamp {
			return nil, errors.Errorf("migrations with %s are duplicated", migrations[i].Timestamp.Format(printTimestampFormat))
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
			if m.Timestamp == ts {
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

	pattern := filepath.FromSlash(fmt.Sprintf("%s/%s.*.%v.sql", m.migrationsDir, timestampStr, direction))
	files, _ := filepath.Glob(pattern)

	if len(files) == 0 {
		pattern = filepath.FromSlash(fmt.Sprintf("%s/%s.*.%v.%s.sql", m.migrationsDir, timestampStr, direction, m.dbWrapper.settings.Driver))
		files, _ = filepath.Glob(pattern)
	}

	if len(files) == 0 {
		return nil, errors.Errorf("migration %v with timestamp %s does not exist", direction, timestampStr)
	}
	if len(files) > 1 {
		return nil, errors.Errorf("got %d %v migration with timestamp %s, should be only one", len(files), direction, timestampStr)
	}

	migration, err := migrationFromFileName(filepath.Base(files[0]))
	if err != nil {
		return nil, err
	}

	return migration, nil
}
