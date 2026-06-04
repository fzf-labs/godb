package gormx

import (
	"strings"
	"testing"
)

func TestBuildPostgresTableCommentsQuery_IncludesPartitionedTables(t *testing.T) {
	query := buildPostgresTableCommentsQuery()
	if !strings.Contains(query, "c.relkind IN ('r', 'p')") {
		t.Fatalf("unexpected query: %s", query)
	}
}
