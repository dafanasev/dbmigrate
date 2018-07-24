package migrate

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"
	
	"github.com/pkg/errors"
)

// Migration holds info about Migration
type Migration struct {
	// Timestamp when the migration was created
	Timestamp time.Time
	Name      string
	direction Direction
	driver    string
}

type byTimestamp []*Migration

func (bts byTimestamp) Len() int           { return len(bts) }
func (bts byTimestamp) Swap(i, j int)      { bts[i], bts[j] = bts[j], bts[i] }
func (bts byTimestamp) Less(i, j int) bool { return bts[i].Timestamp.Unix() < bts[j].Timestamp.Unix() }

// run executes single migration
func (m *Migration) run() error {
	return nil
}

func (m *Migration) buildPath(dir string) string {
	parts := []string{m.Timestamp.Format(timestampFromFileFormat), m.Name}
	if m.driver != "" {
		parts = append(parts, m.driver)
	}
	parts = append(parts, m.direction.String(), "sql")
	
	filename := strings.Join(parts, ".")
	return filepath.FromSlash(fmt.Sprintf("%s/%s", dir, filename))
}

func (m *Migration) FullName() string {
	return fmt.Sprintf("%s %s", m.Timestamp, m.Name)
}

func migrationFromFilename(fname string) (*Migration, error) {
	errMsg := fmt.Sprintf("can't parse migration from filename %s", fname)
	
	if strings.ToLower(filepath.Ext(fname)) != "sql" {
		return nil, errors.Errorf("%s, file is not sql", errMsg)
	}
	
	parts := strings.Split(fname, ".")
	
	ts, err := time.Parse(timestampFromFileFormat, parts[0])
	if err != nil {
		return nil, errors.Wrap(err, errMsg)
	}
	
	name := parts[1]
	
	direction, err := DirectionFromString(parts[2])
	if err != nil {
	    return nil, errors.Wrap(err, errMsg)
	}
	
	// Migration that should be run on specific dbWrapper only
	var driver string
	if len(parts) > 3 {
		if _, ok := providers[parts[3]]; !ok {
			return nil, errors.Errorf("%s, drives is not known", errMsg)
		}
		driver = parts[3]
	}
	
	return &Migration{Timestamp: ts, Name: name, direction: direction, driver: driver}, nil
}