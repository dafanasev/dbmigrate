package dbmigrate

import (
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_byVersion(t *testing.T) {
	v1 := time.Date(2010, 6, 7, 8, 9, 10, 11, time.UTC)
	v2 := v1.Add(time.Second)
	migrations := []*Migration{{Version: v2}, {Version: v1}}
	sort.Sort(byVersion(migrations))
	assert.Equal(t, []*Migration{{Version: v1}, {Version: v2}}, migrations)
}

func Test_Migration_FileName(t *testing.T) {
	v := time.Date(2010, 6, 7, 8, 9, 10, 11, time.UTC)
	m := &Migration{Version: v, Name: "test_migration", Direction: DirectionUp}
	assert.Equal(t, "20100607080910.test_migration.up.sql", m.FileName())

	m.Engine = "postgres"
	assert.Equal(t, "20100607080910.test_migration.up.postgres.sql", m.FileName())
}

func Test_Migration_HumanName(t *testing.T) {
	v := time.Date(2010, 6, 7, 8, 9, 10, 11, time.UTC)
	m := &Migration{Version: v, Name: "test_migration"}
	assert.Equal(t, "test migration", m.HumanName())
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
		v := time.Date(2010, 6, 7, 8, 9, 10, 0, time.UTC)
		assert.Equal(t, v, m.Version)
		parts := strings.Split(fname, ".")
		assert.Equal(t, parts[1], m.Name)
		assert.Equal(t, DirectionUp, m.Direction)
		if strings.ToLower(parts[3]) == "sql" {
			assert.Empty(t, m.Engine)
		} else {
			assert.Equal(t, strings.ToLower(parts[3]), m.Engine)
		}
	}
}
