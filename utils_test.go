package dbmigrate

import (
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testDir() string {
	wd, _ := os.Getwd()
	return filepath.Join(wd, "test")
}

func Test_DirExists(t *testing.T) {
	assert.False(t, DirExists(filepath.Join(testDir(), "not_existed")))
	assert.False(t, DirExists(filepath.Join(testDir(), "file")))
	assert.False(t, DirExists("relPath"))

	assert.True(t, DirExists(filepath.Join("test", "dir")))
	assert.True(t, DirExists(filepath.Join(testDir(), "dir")))
}

func Test_FileExists(t *testing.T) {
	assert.False(t, FileExists(filepath.Join(testDir(), "not_existed")))
	assert.False(t, FileExists("relPath"))
	assert.False(t, FileExists(filepath.Join(testDir(), "dir")))

	assert.True(t, FileExists(filepath.Join("test", "file")))
	assert.True(t, FileExists(filepath.Join(testDir(), "file")))
}

func Test_DirectionFromString(t *testing.T) {
	for _, s := range []string{"UP", "Up", "up"} {
		d, err := DirectionFromString(s)
		require.NoError(t, err)
		assert.Equal(t, DirectionUp, d)
	}

	for _, s := range []string{"DOWN", "DoWn", "down"} {
		d, err := DirectionFromString(s)
		require.NoError(t, err)
		assert.Equal(t, DirectionDown, d)
	}

	for _, s := range []string{"down!", "1up", "rnd", " ", ""} {
		_, err := DirectionFromString(s)
		assert.Error(t, err)
	}
}

func Test_isRootDir(t *testing.T) {
	assert.False(t, isRootDir("/some/path"))
	assert.False(t, isRootDir(filepath.Join("C:", "windows")))
	assert.True(t, isRootDir("/"))
	assert.True(t, isRootDir(filepath.Join("C:")))
}

func Test_EngineExists(t *testing.T) {
	for _, p := range []string{"sqlite", "postgres", "mysql"} {
		assert.True(t, EngineExists(p))
	}

	for _, p := range []string{"", " ", "\n", "nodb"} {
		assert.False(t, EngineExists(p))
	}
}

func Test_Engines(t *testing.T) {
	exp := []string{"sqlite", "postgres", "mysql"}
	sort.Strings(exp)

	act := Engines()
	sort.Strings(act)

	assert.Equal(t, exp, act)
}
