package gormx

import (
	"strings"
	"testing"
)

func TestBuildPgIndexesQuery_UsesPlaceholder(t *testing.T) {
	query, args := buildPgIndexesQuery(`user'name`)
	if !strings.Contains(query, "t.relname=?") {
		t.Fatalf("expected placeholder query, got %q", query)
	}
	if len(args) != 1 || args[0] != `user'name` {
		t.Fatalf("unexpected args: %#v", args)
	}
}

func TestBuildPGSortIndexColumnsQuery_UsesPlaceholder(t *testing.T) {
	query, args := buildPGSortIndexColumnsQuery(`user'name`)
	if !strings.Contains(query, "t.relname=?") {
		t.Fatalf("expected placeholder query, got %q", query)
	}
	if len(args) != 1 || args[0] != `user'name` {
		t.Fatalf("unexpected args: %#v", args)
	}
}
