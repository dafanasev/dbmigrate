package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_pluralize(t *testing.T) {
	assert.Equal(t, "migration", pluralize("migration", 1))
	assert.Equal(t, "migrations", pluralize("migration", 2))
	assert.Equal(t, "migrations", pluralize("migration", 0))
}
