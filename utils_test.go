package migrate

import (
	"os"
	"path/filepath"
	"testing"
	
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testDir() string {
	wd, _ := os.Getwd()
	return filepath.Join(wd, "test")
}

func Test_isDirExists(t *testing.T) {
	assert.False(t, isDirExists(filepath.Join(testDir(), "not_existed")))
	assert.False(t, isDirExists(filepath.Join(testDir(), "file")))
	assert.False(t, isDirExists("relPath"))
	
	assert.True(t, isDirExists(filepath.Join("test", "dir")))
	assert.True(t, isDirExists(filepath.Join(testDir(), "dir")))
}

func Test_isFileExists(t *testing.T) {
	assert.False(t, isFileExists(filepath.Join(testDir(), "not_existed")))
	assert.False(t, isFileExists("relPath"))
	assert.False(t, isFileExists(filepath.Join(testDir(), "dir")))
	
	assert.True(t, isFileExists(filepath.Join("test", "file")))
	assert.True(t, isFileExists(filepath.Join(testDir(), "file")))
}

func Test_isValidString(t *testing.T) {
	validOptions := []string{"one", "two", "three"}
	assert.True(t, isValidString("one", validOptions))
	assert.False(t, isValidString("four",validOptions))
}

func Test_DirectionFromString(t *testing.T) {
	for _, s := range []string{"UP", "Up", "up"} {
		d, err := DirectionFromString(s)
		require.NoError(t, err)
		assert.Equal(t, directionUp, d)
	}
	
	for _, s := range []string{"DOWN", "DoWn", "down"} {
		d, err := DirectionFromString(s)
		require.NoError(t, err)
		assert.Equal(t, directionDown, d)
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