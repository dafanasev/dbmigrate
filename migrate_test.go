package migrate

import (
	"os"
	"testing"

	"github.com/magiconair/properties/assert"
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
	os.MkdirAll("migrations", os.ModeDir|os.ModePerm)
}

func removeTempStuff() {
	os.RemoveAll("test")
	os.RemoveAll("migrations")
	os.Remove("test.db")
	os.Remove("migrate_test")
}

func Test_Direction_String(t *testing.T) {
	d := directionUp
	assert.Equal(t, "up", d.String())

	d = directionDown
	assert.Equal(t, "down", d.String())
}
