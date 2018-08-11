package migrate

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/pkg/errors"
)

const allSteps = 0

type Migrator struct {
	// dir holding migrations
	migrationsDir string
	// migrations table
	migrationsTable string
	// project dir (the one that has migrationsDir as straight subdir)
	projectDir string
	dbWrapper  *dbWrapper
}

// NewMigrator returns migrator instance
func NewMigrator(settings *Settings) (*Migrator, error) {
	if settings.MigrationsDir == "" {
		settings.MigrationsDir = "migrations"
	}
	if settings.MigrationsTable == "" {
		settings.MigrationsTable = "migrations"
	}

	m := &Migrator{migrationsDir: settings.MigrationsDir, migrationsTable: settings.MigrationsTable}

	provider, ok := providers[settings.DriverName]
	if !ok {
		return nil, errors.Errorf("unknown database provider name %s", settings.DriverName)
	}
	m.dbWrapper = newDBWrapper(settings, provider)

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

// Close frees resources acquired by migrator
func (m *Migrator) Close() error {
	err := m.dbWrapper.close()
	if err != nil {
		return errors.Wrap(err, "error shutting down migrator")
	}
	return nil
}

func (m *Migrator) Up() (int, error) {
	return m.UpSteps(allSteps)
}

func (m *Migrator) UpSteps(steps int) (int, error) {
	migrations, err := m.unappliedMigrations()
	if err != nil {
		return 0, errors.Wrap(err, "can't find unapplied migrations")
	}

	if steps == allSteps || steps > len(migrations) {
		steps = len(migrations)
	}

	// TODO: think about prints
	for i, migration := range migrations[:steps] {
		err = m.run(migration)
		if err != nil {
			return i, errors.Wrapf(err, "can't execute migration %s", migration.HumanName())
		}
	}
	return len(migrations), nil
}

func (m *Migrator) Down() (int, error) {
	return m.UpSteps(1)
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
	fpath := filepath.FromSlash(fmt.Sprintf("%s/%s", m.migrationsDir, migration.fileName()))

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
	ts, err := m.dbWrapper.lastMigrationData()
	if err != nil {
		return nil, errors.Wrap(err, "can't get last migration")
	}
	migration, err := m.getMigration(ts, directionUp)
	if err != nil {
		return nil, errors.Wrapf(err, "can't get last migration", ts.Format(timestampFormat))
	}
	return migration, nil
}

// findProjectDir recursively find project dir (the one that has migrations subdir)
func (m *Migrator) findProjectDir(dir string) (string, error) {
	if isDirExists(filepath.Join(dir, m.migrationsDir)) {
		return dir, nil
	}

	if isRootDir(dir) {
		return "", errors.New("project dir not found")
	}

	return m.findProjectDir(filepath.Dir(dir))
}

// readMigrationsFromFiles finds all valid migrations in the migrations dir
func (m *Migrator) readMigrationsFromFiles(direction Direction) []*Migration {
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

		// Migration that should be run on specific dbWrapper only
		if migration.driverName != "" && migration.driverName != m.dbWrapper.settings.DriverName {
			return nil
		}

		migrations = append(migrations, migration)
		return nil
	})

	sort.Sort(byTimestamp(migrations))

	return migrations
}

func (m *Migrator) unappliedMigrations() ([]*Migration, error) {
	migrations := m.readMigrationsFromFiles(directionUp)
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
		pattern = filepath.FromSlash(fmt.Sprintf("%s/%s.*.%v.%s.sql", m.migrationsDir, timestampStr, direction, m.dbWrapper.settings.DriverName))
		files, _ = filepath.Glob(pattern)
	}

	if len(files) == 0 {
		return nil, errors.Errorf("can't get %v migration with timestamp %s", direction, ts.Format(timestampFormat))
	}
	if len(files) > 1 {
		return nil, errors.Errorf("got %d %v migration with timestamp %s, should be only one", len(files), direction, ts.Format(timestampFormat))
	}

	migration, err := migrationFromFileName(filepath.Base(files[0]))
	if err != nil {
		return nil, err
	}

	return migration, nil
}
