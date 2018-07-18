package migrate

import (
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"time"
	
	"github.com/pkg/errors"
	
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

const allSteps = 0

type Migrator struct {
	// dir holding migrations
	migrationsDir string
	// project dir (the one that has migrationsDir as straight subdir)
	projectDir string
	driver string
	dbname string
	db *sql.DB
	placeholdersProvider placeholdersProvider
}

// NewMigrator returns migrator instance
func NewMigrator(credentials *Credentials) (*Migrator, error) {
	m := &Migrator{}
	
	driver, ok := drivers[credentials.DriverName]
	if !ok {
		return nil, errors.Errorf("unknown database driver name %s", credentials.DriverName)
	}
	
	dsn, err := driver.dsn(credentials)
	if err != nil {
		return nil, err
	}
	
	m.driver = credentials.DriverName
	m.dbname = credentials.DBName
	
	wd, err := os.Getwd()
	if err != nil {
		return nil, errors.Wrap(err,"Can't get working directory")
	}
	
	m.projectDir, err = m.findProjectDir(wd)
	if err != nil {
		return nil, err
	}
	
	m.db, err = sql.Open(credentials.DriverName, dsn)
	
	if pp, ok := driver.(placeholdersProvider); ok {
		m.placeholdersProvider = pp
	}
	
	return m, nil
}

// Done frees resources acquired by migrator
func (m *Migrator) Done() error {
	err := m.db.Close()
	if err != nil {
	    return errors.Wrap(err, "Error shutting down migrator")
	}
	return nil
}

func (m *Migrator) setPlaceholders(sql string) string {
	if m.placeholdersProvider != nil {
		return m.placeholdersProvider.setPlaceholders(sql)
	}
	return sql
}

func (m *Migrator) createMigrationsTable() error {
	_, err := m.db.Query(m.setPlaceholders("CREATE TABLE migrations (version timestamp NOT NULL, PRIMARY KEY(version));"))
	if err != nil {
		return errors.Wrapf(err, "Can't create migrations table")
	}
	return nil
}

func (m *Migrator) GetCurrentVersion() (time.Time, error) {
	var v time.Time
	err := m.db.QueryRow("SELECT version FROM migrations ORDER BY version DESC LIMIT 1").Scan(&v)
	if err != nil {
	    return time.Time{}, errors.Wrap(err, "Can't get current version")
	}
	return v, nil
}

func (m *Migrator) findAppliedMigrationsVersions() ([]time.Time, error) {
	rows, err := m.db.Query("SELECT version FROM migrations")
	if err != nil {
	    return nil, errors.Wrap(err, "can't get executed migrations versions")
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

func (m *Migrator) Run(direction Direction) {
	m.RunSteps(direction, allSteps)
}

func (m *Migrator) RunSteps(direction Direction, steps uint) {
	migrations := m.findUnappliedMigrations(direction, steps)
	for _, migration := range migrations {
		migration.run()
	}
}

// findProjectDir recursively find project dir (the one that has migrations subdir)
func (m *Migrator) findProjectDir(dirPath string) (string, error) {
	if dirExists(filepath.Join(dirPath, m.migrationsDir)) {
		return dirPath, nil
	}
	
	if dirPath == "/" {
		return "", errors.New("Project dir not found")
	}
	
	return m.findProjectDir(filepath.Dir(dirPath))
}

// findUnappliedMigrations finds all valid migrations in the migrations dir
func (m *Migrator) findUnappliedMigrations(direction Direction, steps uint) []*migration {
	migrations := []*migration{}
	migrationsDirPath := filepath.Join(m.projectDir, m.migrationsDir)
	
	filepath.Walk(migrationsDirPath, func(mpath string, info os.FileInfo, err error) error {
		if mpath != migrationsDirPath && info.IsDir() {
			return filepath.SkipDir
		}
		if info.IsDir() {
			return nil
		}
		if strings.ToLower(filepath.Ext(mpath)) != "sql" {
			return nil
		}
		
		parts := strings.Split(info.Name(), ".")
		
		ts, err := time.Parse(timestampFormat, parts[0])
		if err != nil {
			return nil
		}
		
		name := parts[1]
		
		if parts[2] != direction.String() {
			return nil
		}
		
		// migration that should be run on specific db only
		if len(parts) > 3 && parts[3] != m.driver {
			return nil
		}
		
		migrations = append(migrations, &migration{name: name, timestamp: ts})
		return nil
	})
	
	return migrations
}
