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

	got, err = ParseCSV("users, roles, ,admin,,")
	require.NoError(t, err)
	assert.Equal(t, []string{"users", "roles", "admin"}, got)
}

func TestParseCSVRejectsAllBlankEntries(t *testing.T) {
	got, err := ParseCSV(" , , ")
	assert.Error(t, err)
	assert.Empty(t, got)
}

func TestParseCSVDeduplicatesWhilePreservingOrder(t *testing.T) {
	got, err := ParseCSV(" users , roles,users, admin ,roles ")
	require.NoError(t, err)
	assert.Equal(t, []string{"users", "roles", "admin"}, got)
}
