package tablelist

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseCSV(t *testing.T) {
	got, err := ParseCSV("")
	require.NoError(t, err)
	assert.Empty(t, got)

	got, err = ParseCSV("users, roles, admin")
	require.NoError(t, err)
	assert.Equal(t, []string{"users", "roles", "admin"}, got)
}

func TestParseCSVRejectsEmptyEntries(t *testing.T) {
	got, err := ParseCSV(" , , ")
	assert.Error(t, err)
	assert.Empty(t, got)

	got, err = ParseCSV("users,,roles")
	assert.Error(t, err)
	assert.Empty(t, got)
	assert.Contains(t, err.Error(), "empty table name at position 2")

	got, err = ParseCSV("users,roles,")
	assert.Error(t, err)
	assert.Empty(t, got)
	assert.Contains(t, err.Error(), "empty table name at position 3")
}

func TestParseCSVDeduplicatesWhilePreservingOrder(t *testing.T) {
	got, err := ParseCSV(" users , roles,users, admin ,roles ")
	require.NoError(t, err)
	assert.Equal(t, []string{"users", "roles", "admin"}, got)
}
