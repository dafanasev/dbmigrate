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
	// Version when the migration was created
	Version    time.Time
	AppliedAt  time.Time
	Name       string
	direction  Direction
	driverName string
}

type byTimestamp []*Migration

func (bts byTimestamp) Len() int           { return len(bts) }
func (bts byTimestamp) Swap(i, j int)      { bts[i], bts[j] = bts[j], bts[i] }
func (bts byTimestamp) Less(i, j int) bool { return bts[i].Version.Unix() < bts[j].Version.Unix() }

func (m *Migration) fileName() string {
	parts := []string{m.Version.Format(timestampFormat), m.Name, m.direction.String()}
	if m.driverName != "" {
		parts = append(parts, m.driverName)
	}
	parts = append(parts, "sql")

	return strings.Join(parts, ".")
}

func (m *Migration) HumanName() string {
	return fmt.Sprintf("%s %s", m.Version.Format(printTimestampFormat), m.Name)
}

func migrationFromFileName(fname string) (*Migration, error) {
	errMsg := fmt.Sprintf("can't parse migration from filename %s", fname)

	if strings.ToLower(filepath.Ext(fname)) != ".sql" {
		return nil, errors.Errorf("%s, file name is not sql", errMsg)
	}

	parts := strings.Split(fname, ".")

	ts, err := time.Parse(timestampFormat, parts[0])
	if err != nil {
		return nil, errors.Wrap(err, errMsg)
	}

	name := parts[1]

	direction, err := DirectionFromString(parts[2])
	if err != nil {
		return nil, errors.Wrap(err, errMsg)
	}

	// Migration that should be run on isSpecific dbWrapper only
	var driver string
	// 4 for .sql extension
	if len(parts) > 4 {
		if _, ok := providers[strings.ToLower(parts[3])]; !ok {
			return nil, errors.Errorf("%s, driverName is not known", errMsg)
		}
		driver = strings.ToLower(parts[3])
	}

	return &Migration{Version: ts, Name: name, direction: direction, driverName: driver}, nil
}
