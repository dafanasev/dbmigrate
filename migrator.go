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
	
	wd, err := os.Getwd()
	if err != nil {
		return nil, errors.Wrap(err,"Can't get working directory")
	}

	// TODO: validate credentials
	m.driver = credentials.DriverName
	m.dbname = credentials.DBName
	
	m.projectDir, err = m.findProjectDir(wd)
	if err != nil {
		return nil, err
	}
	
	if !isValidString(credentials.DriverName, supportedDrivers) {
		return nil, errors.Errorf("unknown database driver name %s", credentials.DriverName)
	}
	
	var provider dsnProvider
	switch credentials.DriverName {
	case "postgres":
		provider = &postgresProvider{}
	case "mysql":
		provider = &mysqlProvider{}
	case "sqlite":
		provider = &sqliteProvider{}
	}
	
	dsn, err := provider.dsn(credentials)
	if err != nil {
	    return nil, err
	}
	
	m.db, err = sql.Open(credentials.DriverName, dsn)
	
	if pp, ok := provider.(placeholdersProvider); ok {
		m.placeholdersProvider = pp
	}
	
	return m, nil
}

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

func (m *Migrator) CreateDB() error {
	_, err := m.db.Query(m.setPlaceholders("CREATE DATABASE ?"), m.dbname)
	if err != nil {
	    return errors.Wrapf(err, "Can't create database %s, probably it is already exists", m.dbname)
	}
	return nil
}

func (m *Migrator) DropDB() error {
	// TODO: cannot drop the currently open database, use other method
	return nil
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

func (m *Migrator) Run(direction Direction) {
	m.RunSteps(direction, 0)
}

func (m *Migrator) RunSteps(direction Direction, steps uint) {
	migrations := m.findNeededMigrations(direction, steps)
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

// findNeededMigrations finds all valid migrations in the migrations dir
func (m *Migrator) findNeededMigrations(direction Direction, steps uint) []*migration {
	migrations := make([]*migration, 0)
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
		
		if parts[2] != direction.String() {
			return nil
		}
		
		// migration that should be run on specific dbWrapper only
		if len(parts) > 3 {
			if parts[3] != m.driver {
			    return nil
			}
		}
		
		ts, err := time.Parse(timestampFormat, parts[0])
		if err != nil {
			return nil
		}
		name := parts[1]
		
		migrations = append(migrations, &migration{name: name, timestamp: ts})
		return nil
	})
	
	return migrations
}
