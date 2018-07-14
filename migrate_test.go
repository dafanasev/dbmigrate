package migrate

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"strings"
)

func Test_filenameSplit(t *testing.T)  {
	s := "20180711084831.create_users.postgres"
	assert.Equal(t, []string{"20180711084831", "create_users", "postgres"}, strings.Split(s, "."))
	
	s = "20180711084831.create_users"
	assert.Equal(t, []string{"20180711084831", "create_users"}, strings.Split(s, "."))
}