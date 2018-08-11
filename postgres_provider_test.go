package migrate

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_postgresProviderExist(t *testing.T) {
	_, ok := providers["postgres"]
	assert.True(t, ok)
}

func Test_postgresProvider_driverName(t *testing.T) {
	assert.Equal(t, "postgres", (&postgresProvider{}).driverName())
}

func Test_postgresProvider_dsn(t *testing.T) {
	p := &postgresProvider{}
	s := &Settings{}

	_, err := p.dsn(s)
	assert.EqualError(t, err, errDBNameNotProvided.Error())

	s.DBName = "test"
	_, err = p.dsn(s)
	assert.EqualError(t, err, errUserNotProvided.Error())

	s.User = "root"
	dsn, err := p.dsn(s)
	require.NoError(t, err)
	assert.Equal(t, "dbname=test user=root sslmode=disable", dsn)

	s.Passwd = "12345"
	s.Host = "myhost"
	s.Port = 5555
	dsn, err = p.dsn(s)
	require.NoError(t, err)
	assert.Equal(t, "dbname=test user=root password=12345 host=myhost port=5555 sslmode=disable", dsn)
}

func Test_postgresProvider_setPlaceholders(t *testing.T) {
	p := &postgresProvider{}
	s := "SELECT * FROM posts WHERE author_id = ? AND created_AT > ? LIMIT 10 ORDER BY created_at ?"
	expected := "SELECT * FROM posts WHERE author_id = $1 AND created_AT > $2 LIMIT 10 ORDER BY created_at $3"
	assert.Equal(t, expected, p.setPlaceholders(s))
}
