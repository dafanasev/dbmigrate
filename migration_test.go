package migrate

import (
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_byTimestamp(t *testing.T) {
	ts1 := time.Date(2010, 6, 7, 8, 9, 10, 11, time.UTC)
	ts2 := ts1.Add(time.Second)
	migrations := []*Migration{{Version: ts2}, {Version: ts1}}
	sort.Sort(byTimestamp(migrations))
	assert.Equal(t, []*Migration{{Version: ts1}, {Version: ts2}}, migrations)
}

func Test_Migration_fileName(t *testing.T) {
	ts := time.Date(2010, 6, 7, 8, 9, 10, 11, time.UTC)
	m := &Migration{Version: ts, Name: "test_migration", direction: directionUp}
	assert.Equal(t, "20100607080910.test_migration.up.sql", m.fileName())

	m.driverName = "postgres"
	assert.Equal(t, "20100607080910.test_migration.up.postgres.sql", m.fileName())
}

func Test_Migration_HumanName(t *testing.T) {
	ts := time.Date(2010, 6, 7, 8, 9, 10, 11, time.UTC)
	m := &Migration{Version: ts, Name: "test_migration"}
	assert.Equal(t, "2010.06.07 08:09:10 test_migration", m.HumanName())
}

func Test_migrationFromFileName(t *testing.T) {
	incorrectNames := []string{
		"20100607080910.test_migration.up.wrong",
		"201000607080910.test_migration.up.sql",
		"20100607080910.test_migration.upp.sql",
		"20100607080910.test_migration.up.msql.sql",
	}
	for _, fname := range incorrectNames {
		_, err := migrationFromFileName(fname)
		assert.Error(t, err)
	}

	correctNames := []string{
		"20100607080910.test_migration.up.sql",
		"20100607080910.test_migration.up.postgres.sql",
		"20100607080910.test_migration.up.postGres.sql",
		"20100607080910.test_migration.up.Sql",
		"20100607080910.test_migration.up.postGres.sql",
		"20100607080910.test migration.up.postGres.sql",
		"20100607080910.test 123 $^* migration.up.postgres.sql",
	}
	for _, fname := range correctNames {
		m, err := migrationFromFileName(fname)
		assert.NoError(t, err)
		ts := time.Date(2010, 6, 7, 8, 9, 10, 0, time.UTC)
		assert.Equal(t, ts, m.Version)
		parts := strings.Split(fname, ".")
		assert.Equal(t, parts[1], m.Name)
		assert.Equal(t, directionUp, m.direction)
		if strings.ToLower(parts[3]) == "sql" {
			assert.Empty(t, m.driverName)
		} else {
			assert.Equal(t, strings.ToLower(parts[3]), m.driverName)
		}
	}
}
