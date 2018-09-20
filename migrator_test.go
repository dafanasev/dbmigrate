package migrate

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_NewMigrator(t *testing.T) {
	os.Remove("test.db")
	s := &Settings{}
	m, err := NewMigrator(s)
	assert.EqualError(t, err, "database driver not specified")

	s.Driver = "nosql"
	m, err = NewMigrator(s)
	assert.EqualError(t, err, "database name not specified")

	s.DB = "test.db"

	_, err = NewMigrator(s)
	assert.Contains(t, err.Error(), "unknown database driver")

	s.Driver = "sqlite"
	m, err = NewMigrator(s)
	require.NoError(t, err)
	assert.Equal(t, "migrations", m.migrationsDir)
	assert.Equal(t, "migrations", m.migrationsTable)
	projectDir, _ := os.Getwd()
	assert.Equal(t, projectDir, m.projectDir)
	assert.Equal(t, "sqlite3", m.dbWrapper.driverName())
	m.Close()
}

func Test_Migrator_Close(t *testing.T) {
	m, err := NewMigrator(&Settings{Driver: "sqlite", DB: "test.db"})
	require.NoError(t, err)
	err = m.Close()
	assert.NoError(t, err)
}

func Test_Migrator_getMigration(t *testing.T) {
	os.Remove("test.db")
	m, _ := NewMigrator(&Settings{Driver: "sqlite", DB: "test.db", MigrationsDir: "test_migrations"})
	defer m.Close()

	// does not exist at all
	_, err := m.getMigration(time.Date(2018, 9, 10, 11, 12, 13, 0, time.UTC), directionUp)
	assert.Contains(t, err.Error(), "does not exist")

	// does not exist for needed direction
	os.Rename("test_migrations/20180918200453.correct.down.sql", "./20180918200453.correct.down.sql")
	defer os.Rename("./20180918200453.correct.down.sql", "test_migrations/20180918200453.correct.down.sql")
	_, err = m.getMigration(time.Date(2018, 9, 18, 20, 4, 53, 0, time.UTC), directionDown)
	assert.Contains(t, err.Error(), "does not exist")

	// does not exist for needed driver
	_, err = m.getMigration(time.Date(2018, 9, 18, 20, 7, 42, 0, time.UTC), directionUp)
	assert.Contains(t, err.Error(), "does not exist")

	// multiple migrations for the timestamp
	_, err = m.getMigration(time.Date(2018, 9, 18, 20, 6, 32, 0, time.UTC), directionUp)
	assert.Contains(t, err.Error(), "should be only one")

	// correct for any driver
	ts := time.Date(2018, 9, 18, 20, 4, 53, 0, time.UTC)
	migration, err := m.getMigration(ts, directionUp)
	require.NoError(t, err)
	assert.NotNil(t, migration)
	expected := &Migration{Timestamp: ts, Name: "correct", direction: directionUp}
	assert.Equal(t, expected, migration)

	// correct for the isSpecific driver
	ts = time.Date(2018, 9, 18, 20, 10, 19, 0, time.UTC)
	migration, err = m.getMigration(ts, directionUp)
	require.NoError(t, err)
	assert.NotNil(t, migration)
	expected = &Migration{Timestamp: ts, Name: "specific_driver_correct", direction: directionUp, driverName: "sqlite"}
	assert.Equal(t, expected, migration)
}

func Test_Migrator_findMigrations(t *testing.T) {
	os.Remove("test.db")
	m, _ := NewMigrator(&Settings{Driver: "sqlite", DB: "test.db", MigrationsDir: "test_migrations"})
	defer m.Close()

	_, err := m.findMigrations(directionUp)
	assert.Contains(t, err.Error(), "are duplicated")

	os.Rename("test_migrations/20180918200632.duplicate2.up.sql", "./20180918200632.duplicate2.up.sql")
	defer os.Rename("./20180918200632.duplicate2.up.sql", "test_migrations/20180918200632.duplicate2.up.sql")
	migrations, err := m.findMigrations(directionUp)
	require.NoError(t, err)
	assert.Len(t, migrations, 3)
}

func Test_Migrator_unappliedMigrations(t *testing.T) {
	os.Remove("test.db")
	os.Rename("test_migrations/20180918200632.duplicate2.up.sql", "./20180918200632.duplicate2.up.sql")
	defer os.Rename("./20180918200632.duplicate2.up.sql", "test_migrations/20180918200632.duplicate2.up.sql")
	defer os.Remove("test.db")

	m, _ := NewMigrator(&Settings{Driver: "sqlite", DB: "test.db", MigrationsDir: "test_migrations"})
	defer m.Close()
	migrations, _ := m.findMigrations(directionUp)

	for i := 0; i < 4; i++ {
		unappliedMigrations, err := m.unappliedMigrations()
		require.NoError(t, err)
		assert.Len(t, unappliedMigrations, 3-i)

		// we've got migrations we were actually needed
		for j, um := range unappliedMigrations {
			assert.Equal(t, um.Timestamp, migrations[i+j].Timestamp)
		}

		if i < 3 {
			m.dbWrapper.insertMigrationTimestamp(migrations[i].Timestamp, time.Now())
		}
	}
}

func Test_Migrator_findProjectDir(t *testing.T) {
	os.Remove("test.db")
	m, _ := NewMigrator(&Settings{Driver: "sqlite", DB: "test.db", MigrationsDir: "test_migrations"})
	defer m.Close()

	wd, _ := os.Getwd()
	projectDir, err := m.findProjectDir(wd)
	require.NoError(t, err)
	assert.Equal(t, wd, projectDir)

	projectDir, err = m.findProjectDir(filepath.Join(wd, "cmd"))
	require.NoError(t, err)
	assert.Equal(t, wd, projectDir)

	m.migrationsDir = "migrationsss"
	_, err = m.findProjectDir(wd)
	assert.EqualError(t, err, "project dir not found")
}

func TestMigrator_Migrator_LastMigration(t *testing.T) {
	os.Remove("test.db")
	m, _ := NewMigrator(&Settings{Driver: "sqlite", DB: "test.db", MigrationsDir: "test_migrations"})
	defer m.Close()

	migration, err := m.LastMigration()
	require.NoError(t, err)
	assert.Nil(t, migration)

	ts := time.Date(2018, 9, 18, 20, 4, 53, 0, time.UTC)
	_ = m.dbWrapper.insertMigrationTimestamp(ts, time.Now())
	migration, err = m.LastMigration()
	require.NoError(t, err)
	assert.Equal(t, ts, migration.Timestamp)

	ts = time.Date(2018, 9, 18, 22, 2, 34, 0, time.UTC)
	_ = m.dbWrapper.insertMigrationTimestamp(ts, time.Now())
	_, err = m.LastMigration()
	assert.Contains(t, err.Error(), "can't get last migration with timestamp")
}

func Test_Migrator_run(t *testing.T) {
	os.Remove("test.db")
	m, _ := NewMigrator(&Settings{Driver: "sqlite", DB: "test.db", MigrationsDir: "test_migrations"})
	defer m.Close()

	migration, _ := migrationFromFileName("20180918100423.incorrect.up.sql")
	err := m.run(migration)
	assert.Contains(t, err.Error(), "can't read file for migration")

	migration, _ = migrationFromFileName("20180918200742.wrong_driver.up.postgres.sql")
	err = m.run(migration)
	assert.Contains(t, err.Error(), "can't exec query for migration")

	migration, _ = migrationFromFileName("20180918200453.correct.up.sql")
	err = m.run(migration)
	require.NoError(t, err)
}

func Test_Migrator_UpSteps_DownSteps(t *testing.T) {
	os.Remove("test.db")

	os.Rename("test_migrations/20180918200632.duplicate2.up.sql", "./20180918200632.duplicate2.up.sql")
	defer os.Rename("./20180918200632.duplicate2.up.sql", "test_migrations/20180918200632.duplicate2.up.sql")

	m, _ := NewMigrator(&Settings{Driver: "sqlite", DB: "test.db", MigrationsDir: "test_migrations"})
	defer m.Close()

	n, err := m.Down()
	require.NoError(t, err)
	assert.Equal(t, 0, n)
	lm, _ := m.LastMigration()
	assert.Nil(t, lm)

	n, err = m.DownSteps(1)
	require.NoError(t, err)
	assert.Equal(t, 0, n)
	lm, _ = m.LastMigration()
	assert.Nil(t, lm)

	n, err = m.UpSteps(1)
	require.NoError(t, err)
	assert.Equal(t, 1, n)
	lm, _ = m.LastMigration()
	assert.Equal(t, time.Date(2018, 9, 18, 20, 4, 53, 0, time.UTC), lm.Timestamp)

	n, err = m.Down()
	require.NoError(t, err)
	assert.Equal(t, 1, n)
	lm, _ = m.LastMigration()
	assert.Nil(t, lm)

	n, err = m.UpSteps(2)
	require.NoError(t, err)
	assert.Equal(t, 2, n)
	lm, _ = m.LastMigration()
	assert.Equal(t, time.Date(2018, 9, 18, 20, 6, 32, 0, time.UTC), lm.Timestamp)

	n, err = m.DownSteps(2)
	require.NoError(t, err)
	assert.Equal(t, 2, n)
	lm, _ = m.LastMigration()
	assert.Nil(t, lm)

	n, err = m.Up()
	require.NoError(t, err)
	assert.Equal(t, 3, n)
	lm, _ = m.LastMigration()
	assert.Equal(t, time.Date(2018, 9, 18, 20, 10, 19, 0, time.UTC), lm.Timestamp)

	n, err = m.DownSteps(1)
	require.NoError(t, err)
	assert.Equal(t, 1, n)
	lm, _ = m.LastMigration()
	assert.Equal(t, time.Date(2018, 9, 18, 20, 6, 32, 0, time.UTC), lm.Timestamp)

	n, err = m.Down()
	require.NoError(t, err)
	assert.Equal(t, 2, n)
	lm, _ = m.LastMigration()
	assert.Nil(t, lm)
}

func Test_Migrator_GenerateMigration(t *testing.T) {
	m, _ := NewMigrator(&Settings{Driver: "sqlite", DB: "test.db", MigrationsDir: "test_migrations"})
	defer m.Close()

	testData := []struct {
		descr      string
		isSpecific bool
	}{
		{" test  migration \n ", false},
		{" test\tspecific migration \n ", true},
	}
	for _, data := range testData {
		fnames, err := m.GenerateMigration(data.descr, data.isSpecific)
		assert.NoError(t, err)
		for _, fname := range fnames {
			descrPart := "test_migration"
			if data.isSpecific {
				descrPart = "test_specific_migration"
			}
			assert.Contains(t, fname, descrPart)
			assert.True(t, fileExists(filepath.Join(m.migrationsDir, fname)))
		}

		_, err = m.GenerateMigration(data.descr, data.isSpecific)
		assert.Contains(t, err.Error(), "already exists")

		for _, fname := range fnames {
			os.Remove(filepath.Join(m.migrationsDir, fname))
		}
	}
}
