package dbmigrate

import (
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_newDBWrapper(t *testing.T) {
	s := &Settings{}
	for _, p := range []provider{&postgresProvider{}, &mysqlProvider{}, &sqliteProvider{}} {
		w := newDBWrapper(s, p)
		if _, ok := p.(*postgresProvider); ok {
			assert.NotNil(t, w.placeholdersProvider)
		} else {
			assert.Nil(t, w.placeholdersProvider)
		}
	}
}

func Test_dbWrapper_setPlaceholders(t *testing.T) {
	s := &Settings{}
	str := "SELECT * FROM posts WHERE author_id = ? AND created_AT > ? LIMIT 10 ORDER BY created_at ?"
	for _, p := range []provider{&postgresProvider{}, &mysqlProvider{}, &sqliteProvider{}} {
		w := newDBWrapper(s, p)
		if _, ok := p.(*postgresProvider); ok {
			expected := "SELECT * FROM posts WHERE author_id = $1 AND created_AT > $2 LIMIT 10 ORDER BY created_at $3"
			assert.Equal(t, expected, w.setPlaceholders(str))
		} else {
			assert.Equal(t, str, w.setPlaceholders(str))
		}
	}
}

func Test_dbWrapper(t *testing.T) {
	for engine, provider := range providers {
		s := &Settings{Engine: engine, Database: "migrate_test", User: "dbmigrate", Password: "mysecretpassword", MigrationsTable: "migrations"}
		if engine == "postgres" {
			s.Port = 5433
		}
		if engine == "mysql" {
			s.Port = 3307
		}

		w := newDBWrapper(s, provider)
		err := w.open()
		require.NoError(t, err)

		tableExist, err := w.hasMigrationsTable()
		assert.NoError(t, err)
		assert.False(t, tableExist)

		err = w.createMigrationsTable()
		assert.NoError(t, err)

		err = w.createMigrationsTable()
		assert.Error(t, err)

		tableExist, err = w.hasMigrationsTable()
		assert.NoError(t, err)
		assert.True(t, tableExist)

		ts, err := w.latestMigrationVersion()
		// no error and null time value means there are no migrations in the table
		assert.NoError(t, err)
		assert.Equal(t, time.Time{}, ts)

		mds, err := w.appliedMigrationsData("version DESC")
		assert.NoError(t, err)
		assert.Equal(t, []*migrationData(nil), mds)

		n, err := w.countMigrationsInLastBatch()
		assert.NoError(t, err)
		assert.Zero(t, n)

		baseTs := time.Date(2010, 6, 7, 8, 9, 10, 0, time.UTC)
		now := time.Now().UTC().Truncate(time.Second)
		for n := 0; n < 2; n++ {
			err = w.insertMigrationVersion(baseTs.Add(time.Duration(n)*time.Second), now, nil)
			assert.NoError(t, err)
		}

		ts, err = w.latestMigrationVersion()
		assert.NoError(t, err)
		assert.Equal(t, baseTs.Add(time.Second), ts)

		_, err = w.getAttrOrderedBy("error_attr", "error_attr DESC")
		require.Error(t, err)

		ts, err = w.getAttrOrderedBy("applied_at", "applied_at DESC")
		assert.NoError(t, err)
		assert.Equal(t, now, ts)

		n, err = w.countMigrationsInLastBatch()
		assert.NoError(t, err)
		assert.Equal(t, 2, n)

		_, err = w.appliedMigrationsData("version RASC")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "can't get applied migrations versions")

		mds, err = w.appliedMigrationsData("version DESC")
		assert.NoError(t, err)
		assert.Equal(t, []time.Time{baseTs.Add(time.Second), baseTs}, []time.Time{mds[0].version, mds[1].version})
		for _, md := range mds {
			assert.NotEqual(t, time.Time{}, md.appliedAt)
		}

		mds, err = w.appliedMigrationsData("version ASC")
		assert.NoError(t, err)
		assert.Equal(t, []time.Time{baseTs, baseTs.Add(time.Second)}, []time.Time{mds[0].version, mds[1].version})
		for _, md := range mds {
			assert.NotEqual(t, time.Time{}, md.appliedAt)
		}

		err = w.deleteMigrationVersion(baseTs.Add(time.Second), nil)
		assert.NoError(t, err)

		ts, err = w.latestMigrationVersion()
		assert.NoError(t, err)
		assert.Equal(t, baseTs, ts)

		w.db.Exec("DROP TABLE migrations;")

		err = w.close()
		assert.NoError(t, err)
	}
}

func Test_dbWrapper_execMigrationQueries(t *testing.T) {

	for engine, provider := range providers {
		s := &Settings{Engine: engine, Database: "migrate_test", User: "dbmigrate", Password: "mysecretpassword", MigrationsTable: "migrations"}
		if engine == "postgres" {
			s.Port = 5433
		}
		if engine == "mysql" {
			s.Port = 3307
		}

		w := newDBWrapper(s, provider)
		w.open()

		afterFunc := func(tx *sql.Tx) error {
			return nil
		}

		// wrong one command query
		err := w.execMigrationQueries("CREATE TABLE posts ERROR title VARCHAR(255) NOT NULL, PRIMARY KEY(title));", afterFunc)
		assert.Error(t, err)
		tableExists, _ := w.hasMigrationsTable()
		assert.False(t, tableExists)

		// right one command query
		query := "CREATE TABLE posts (title VARCHAR(255) NOT NULL, PRIMARY KEY(title));"
		err = w.execMigrationQueries(query, afterFunc)
		assert.NoError(t, err)
		var table string
		err = w.db.QueryRow(w.setPlaceholders(w.provider.hasTableQuery()), "posts").Scan(&table)
		assert.NoError(t, err)
		assert.Equal(t, "posts", table)
		w.db.Exec("DROP TABLE posts;")

		// right multiple commands query
		query = `
			CREATE TABLE posts (title VARCHAR(255) NOT NULL, PRIMARY KEY(title));
			ALTER TABLE posts ADD content TEXT;
			INSERT INTO posts (title, content) VALUES ('First post', 'And its content');
		`
		err = w.execMigrationQueries(query, afterFunc)
		assert.NoError(t, err)
		var title string
		err = w.db.QueryRow("SELECT title FROM posts LIMIT 1").Scan(&title)
		assert.NoError(t, err)
		assert.Equal(t, "First post", title)
		w.db.Exec("DROP TABLE posts;")

		// wrong multiple commands query
		query = `
			CREATE TABLE posts (title VARCHAR(255) NOT NULL, PRIMARY KEY(title));
			INSERT INTO posts (title) VALUES ('First post');
			ERROR;
			ALTER TABLE posts ADD content TEXT;
		`
		err = w.execMigrationQueries(query, afterFunc)
		assert.Error(t, err)
		title = ""
		err = w.db.QueryRow("SELECT title FROM posts LIMIT 1").Scan(&title)
		if s.Engine == "mysql" {
			// mysql doesn't support DDL, therefore first two statements are not rolled back
			assert.NoError(t, err)
			assert.Equal(t, "First post", title)
		} else {
			// postgres and sqlite support DDL transactions
			// therefore there are no table and content because first two statements are rolled back
			assert.Error(t, err)
			assert.Equal(t, "", title)
		}
		w.db.Exec("DROP TABLE posts;")

		w.close()
	}
}
