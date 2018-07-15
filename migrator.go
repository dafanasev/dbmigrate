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
	// migration direction
	direction
	// sqlBuilder driver
	driver sqlBuilder
}

// NewMigrator returns migrator instance
func NewMigrator(conf Config) (*Migrator, error) {
	m := &Migrator{}
	
	wd, err := os.Getwd()
	if err != nil {
		return nil, errors.Wrap(err,"Can't get working directory")
	}
	
	m.projectDir, err = m.findProjectDir(wd)
	if err != nil {
		return nil, err
	}
	
	m.direction, err = m.directionFromString(conf.Direction)
	if err != nil {
		return nil, err
	}
	
	m.driver, err = m.driverFromString(conf.Driver)
	if err != nil {
		return nil, err
	}
	
	return m, nil
}

func (m *Migrator) Run() {
	migrations := m.findNeededMigrations()
	for _, migration := range migrations {
		migration.run()
	}
}

// directionFromString tries to build direction from string,
// checking for valid ones
func (m *Migrator) directionFromString(s string) (direction, error) {
	s = strings.ToLower(s)
	if !isValidString(s, []string{"up", "down"}) {
		return directionError, nil
	}
	
	var d direction
	switch s {
	case "up":
		d = directionUp
	case "down":
		d = directionDown
	}
	return d, nil
}

// driverFromString tries to build dialect from string,
// checking for valid ones
func (m *Migrator) driverFromString(s string) (sqlBuilder, error) {
	// TODO: more effective array lookup
	s = strings.ToLower(s)
	if !isValidString(s, []string{"postgresql", "postgres", "pg", "mysql", "sqlite"}) {
		return nil, errors.Errorf("unknown sqlBuilder %s", s)
	}
	
	var d sqlBuilder
	switch s {
	case "postgresql", "postgres", "pg":
		d = &postgresBuilder{}
	case "mysql":
		d = &mySQLBuilder{}
	case "sqlite":
		d = &sqliteBuilder{}
	}
	return d, nil
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
func (m *Migrator) findNeededMigrations() []*migration {
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
		
		if parts[2] != m.direction.String() {
			return nil
		}
		
		// migration that should be run on specific db only
		if len(parts) > 3 {
			mDriver, err := m.driverFromString(parts[3])
			if err != nil || mDriver != m.driver {
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
