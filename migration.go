package dbmigrate

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/errors"
)

// Migration holds metadata of migration
type Migration struct {
	// Version holds the created at timestamp
	Version   time.Time
	Name      string
	AppliedAt time.Time
	Direction Direction
	Engine    string
}

type byVersion []*Migration

func (bv byVersion) Len() int           { return len(bv) }
func (bv byVersion) Swap(i, j int)      { bv[i], bv[j] = bv[j], bv[i] }
func (bv byVersion) Less(i, j int) bool { return bv[i].Version.Unix() < bv[j].Version.Unix() }

// FileName builds migration file name from metadata
func (m *Migration) FileName() string {
	parts := []string{m.Version.Format(TimestampFormat), m.Name, m.Direction.String()}
	if m.Engine != "" {
		parts = append(parts, m.Engine)
	}
	parts = append(parts, "sql")

	return strings.Join(parts, ".")
}

// HumanName returns migration's name without underscores
func (m *Migration) HumanName() string {
	return strings.Replace(m.Name, "_", " ", -1)
}

// migrationFromFileName tries to parse migration metadata from the filename
func migrationFromFileName(fname string) (*Migration, error) {
	errMsg := fmt.Sprintf("can't parse migration from filename %s", fname)

	if strings.ToLower(filepath.Ext(fname)) != ".sql" {
		return nil, errors.Errorf("%s, file name is not sql", errMsg)
	}

	parts := strings.Split(fname, ".")

	ts, err := time.Parse(TimestampFormat, parts[0])
	if err != nil {
		return nil, errors.Wrap(err, errMsg)
	}

	name := parts[1]

	direction, err := DirectionFromString(parts[2])
	if err != nil {
		return nil, errors.Wrap(err, errMsg)
	}

	// engine specific migration
	var engine string
	if len(parts) > 4 {
		if _, ok := providers[strings.ToLower(parts[3])]; !ok {
			return nil, errors.Errorf("%s, engine is not known", errMsg)
		}
		engine = strings.ToLower(parts[3])
	}

	return &Migration{Version: ts, Name: name, Direction: direction, Engine: engine}, nil
}
