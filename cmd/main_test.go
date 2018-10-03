package main

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/dafanasev/dbmigrate"
)

func createTestMigrations() {
	filesData := map[string]string{
		"20180918200453.first.up.sql":           "CREATE TABLE posts (title VARCHAR NOT NULL, content TEXT NOT NULL, PRIMARY KEY(title));",
		"20180918200453.first.down.sql":         "DROP TABLE posts;",
		"20180918200632.second.up.sqlite.sql":   "CREATE TABLE users (name VARCHAR NOT NULL, email VARCHAR NOT NULL, PRIMARY KEY(email));",
		"20180918200632.second.down.sqlite.sql": "DROP TABLE users;",
		"20180918201019.third.up.sqlite.sql":    "CREATE TABLE comments (author VARCHAR NOT NULL, content TEXT NOT NULL, PRIMARY KEY(author))\n; CREATE TABLE tags (title VARCHAR NOT NULL, PRIMARY KEY(title));",
		"20180918201019.third.down.sqlite.sql":  "DROP TABLE comments;\n DROP TABLE tags;",
	}
	os.Mkdir(dbmigrate.MigrationsDir, os.ModePerm)
	for fname, content := range filesData {
		ioutil.WriteFile(filepath.Join(dbmigrate.MigrationsDir, fname), []byte(content), os.ModePerm)
	}
}
