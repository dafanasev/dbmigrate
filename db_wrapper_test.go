package migrate

import (
		"testing"
	"time"
	
	"github.com/stretchr/testify/assert"
	)

func Test_newDBWrapper(t *testing.T) {
	s := &Settings{}
	for _, p := range []provider{&postgresProvider{}, &mysqlProvider{}, &sqliteProvider{}} 	{
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
	for _, p := range []provider{&postgresProvider{}, &mysqlProvider{}, &sqliteProvider{}} 	{
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
	for driverName, provider := range providers {
		s := &Settings{DriverName: driverName, DBName: "migrate_test", User: "migrate", Passwd: "mysecretpassword", MigrationsTable: "migrations"}
		if driverName == "postgres" {
			s.Port = 5433
		}
		if driverName == "mysql" {
			s.Port = 3307
		}
		
		w := newDBWrapper(s, provider)
		err := w.open()
		assert.NoError(t, err)
		
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
		
		ts, err := w.lastMigrationData()
		// no error and null time value means there are no migrations in the table
		assert.NoError(t, err)
		assert.Equal(t, time.Time{}, ts)
		
		tss, err := w.appliedMigrationsTimestamps("DESC")
		assert.NoError(t, err)
		assert.Equal(t, []time.Time(nil), tss)
		
		baseTs := time.Date(2010,6,7,8,9, 10, 0, time.UTC)
		for n := 0; n < 2; n++ {
			err = w.insertMigrationTimestamp(baseTs.Add(time.Duration(n) * time.Second))
			assert.NoError(t, err)
		}
		
		ts, err = w.lastMigrationData()
		assert.NoError(t, err)
		assert.Equal(t, baseTs.Add(time.Second), ts)
		
		tss, err = w.appliedMigrationsTimestamps("DESC")
		assert.NoError(t, err)
		assert.Equal(t, []time.Time{baseTs.Add(time.Second), baseTs}, tss)
		
		tss, err = w.appliedMigrationsTimestamps("ASC")
		assert.NoError(t, err)
		assert.Equal(t, []time.Time{baseTs, baseTs.Add(time.Second)}, tss)
		
		err = w.deleteMigrationTimestamp(baseTs.Add(time.Second))
		assert.NoError(t, err)
		
		ts, err = w.lastMigrationData()
		assert.NoError(t, err)
		assert.Equal(t, baseTs, ts)
		
		w.db.Exec("DROP TABLE migrations;")
		
		err = w.close()
		assert.NoError(t, err)
	}
}

func Test_dbWrapper_execQuery(t *testing.T) {
	for driverName, provider := range providers {
		s := &Settings{DriverName: driverName, DBName: "migrate_test", User: "migrate", Passwd: "mysecretpassword", MigrationsTable: "migrations"}
		if driverName == "postgres" {
			s.Port = 5433
		}
		if driverName == "mysql" {
			s.Port = 3307
		}
		
		w := newDBWrapper(s, provider)
		w.open()
		
		// wrong one command querie
		wrongQueries := []string{
			"",
			"CREATE TABLE posts ERROR title VARCHAR(255) NOT NULL, PRIMARY KEY(title));",
		}
		for _, query := range wrongQueries {
			err := w.execQuery(query)
			assert.Error(t, err)
			tableExists, _ := w.hasMigrationsTable()
			assert.False(t, tableExists)
		}
		
		// right one command query
		query := "CREATE TABLE posts (title VARCHAR(255) NOT NULL, PRIMARY KEY(title));"
		err := w.execQuery(query)
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
		err = w.execQuery(query)
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
		err = w.execQuery(query)
		assert.Error(t, err)
		title = ""
		err = w.db.QueryRow("SELECT title FROM posts LIMIT 1").Scan(&title)
		if s.DriverName == "mysql" {
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