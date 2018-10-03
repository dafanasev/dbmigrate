package dbmigrate

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	removeTempStuff()
	createTempStuff()
	code := m.Run()
	removeTempStuff()
	os.Exit(code)
}

func createTempStuff() {
	removeTempStuff()
	os.MkdirAll("test/dir", os.ModeDir|os.ModePerm)
	os.Create("test/file")
	os.MkdirAll(MigrationsDir, os.ModeDir|os.ModePerm)

	os.MkdirAll(filepath.Join(MigrationsDir, "subfolder"), os.ModeDir|os.ModePerm)

	filesData := map[string]string{
		"8234234.incorrect_name.noql.sql":                        "",
		"20180918200453.correct.up.sql":                          "CREATE TABLE posts (title VARCHAR NOT NULL, content TEXT NOT NULL, PRIMARY KEY(title));",
		"20180918200453.correct.down.sql":                        "DROP TABLE posts;",
		"20180918200632.other_correct.up.sql":                    "CREATE TABLE authors (name VARCHAR NOT NULL, email VARCHAR NOT NULL, PRIMARY KEY(email));",
		"20180918200632.other_correct.down.sql":                  "DROP TABLE authors;",
		"20180918200742.wrong_engine.up.postgres.sql":            "",
		"20180918200742.wrong_engine.down.postgres.sql":          "",
		"20180918201019.specific_engine_correct.up.sqlite.sql":   "CREATE TABLE comments (author VARCHAR NOT NULL, content TEXT NOT NULL, PRIMARY KEY(author))\n; CREATE TABLE tags (title VARCHAR NOT NULL, PRIMARY KEY(title));",
		"20180918201019.specific_engine_correct.down.sqlite.sql": "DROP TABLE comments;\n DROP TABLE tags;",
	}
	for fname, content := range filesData {
		ioutil.WriteFile(filepath.Join(MigrationsDir, fname), []byte(content), 0644)
	}
}

func removeTempStuff() {
	os.RemoveAll("test")
	os.RemoveAll(MigrationsDir)
	os.Remove("test.db")
	os.Remove("migrate_test")
}

func Test_Direction_String(t *testing.T) {
	d := DirectionUp
	assert.Equal(t, "up", d.String())

	d = DirectionDown
	assert.Equal(t, "down", d.String())
}
