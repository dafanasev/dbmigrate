package migrate

import (
	"os"
	"path"
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
	// db operations
	dbWrapper
}

// NewMigrator returns migrator instance
func NewMigrator(credentials Credentials) (*Migrator, error) {
	m := &Migrator{}
	
	wd, err := os.Getwd()
	if err != nil {
		return nil, errors.Wrap(err,"Can't get working directory")
	}
	
	m.projectDir, err = m.findProjectDir(wd)
	if err != nil {
		return nil, err
	}
	
	driverName := strings.ToLower(credentials.DriverName)
	if !isValidString(driverName, supportedDrivers) {
		return nil, errors.Errorf("unknown database driverName %s", driverName)
	}
	
	switch driverName {
	case "postgresql", "postgres", "pg":
		m.dbWrapper = &postgresWrapper{}
	case "mysql":
		m.dbWrapper = &mySQLWrapper{}
	case "sqlite":
		m.dbWrapper = &sqliteWrapper{}
	}
	
	return m, nil
}

func (m *Migrator) CreateDB() error {
	return m.createDB()
}

func (m *Migrator) DropDB() error {
	return m.dropDB()
}

func (m *Migrator) GetCurrentVersion() (time.Time, error) {
	return m.getCurrentVersion()
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
	if dirExists(path.Join(dirPath, m.migrationsDir)) {
		return dirPath, nil
	}
	
	if dirPath == "/" {
		return "", errors.New("Project dir not found")
	}
	
	return m.findProjectDir(path.Dir(dirPath))
}

// findNeededMigrations finds all valid migrations in the migrations dir
func (m *Migrator) findNeededMigrations(direction Direction, steps uint) []*migration {
	migrations := make([]*migration, 0)
	migrationsDirPath := path.Join(m.projectDir, m.migrationsDir)
	
	filepath.Walk(migrationsDirPath, func(mpath string, info os.FileInfo, err error) error {
		if mpath != migrationsDirPath && info.IsDir() {
			return filepath.SkipDir
		}
		if info.IsDir() {
			return nil
		}
		if strings.ToLower(path.Ext(mpath)) != "sql" {
			return nil
		}
		
		parts := strings.Split(info.Name(), ".")
		
		if parts[2] != direction.String() {
			return nil
		}
		
		// migration that should be run on specific dbWrapper only
		if len(parts) > 3 {
			if parts[3] != m.dbWrapper.driverName() {
			    return nil
			}
		}
		
		ts, err := time.Parse("20060102150405", parts[0])
		if err != nil {
			return nil
		}
		name := parts[1]
		
		migrations = append(migrations, &migration{name: name, timestamp: ts})
		return nil
	})
	
	return migrations
}
