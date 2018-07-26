package migrate

import (
	"testing"
	
	"github.com/stretchr/testify/assert"
)

func Test_defaultProvider_hasTableQuery(t *testing.T) {
	p := &defaultProvider{}
	assert.Contains(t, p.hasTableQuery(), "information_schema.tables")
}
