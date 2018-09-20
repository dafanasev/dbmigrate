package migrate

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_mysqlProviderExist(t *testing.T) {
	_, ok := providers["mysql"]
	assert.True(t, ok)
}

func Test_mysqlProvider_driverName(t *testing.T) {
	assert.Equal(t, "mysql", (&mysqlProvider{}).driverName())
}

func Test_mysqlProvider_dsn(t *testing.T) {
	p := &mysqlProvider{}
	s := &Settings{}

	_, err := p.dsn(s)
	assert.EqualError(t, err, errDBNameNotProvided.Error())

	s.DB = "test"
	_, err = p.dsn(s)
	assert.EqualError(t, err, errUserNotProvided.Error())

	s.User = "root"
	dsn, err := p.dsn(s)
	require.NoError(t, err)
	assert.Equal(t, "root:@tcp(127.0.0.1:3306)/test?parseTime=true", dsn)

	s.Passwd = "12345"
	s.Host = "myhost"
	s.Port = 3307
	dsn, err = p.dsn(s)
	require.NoError(t, err)
	assert.Equal(t, "root:12345@tcp(myhost:3307)/test?parseTime=true", dsn)
}
