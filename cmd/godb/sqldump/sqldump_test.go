package sqldump

import (
	"errors"
	"strings"
	"testing"

	"gorm.io/gorm"
)

func TestNewSQLDumpAndRunUnknownDatabase(t *testing.T) {
	dump := NewSQLDump("sqlite", "dsn", "/tmp/out", "users,roles", true)
	if dump.db != "sqlite" || dump.dsn != "dsn" || dump.outPutPath != "/tmp/out" || dump.targetTables != "users,roles" || !dump.fileOverwrite {
		t.Fatalf("unexpected dump config: %#v", dump)
	}

	if err := dump.Run(); err == nil {
		t.Fatal("expected unknown database type error")
	}
}

func TestSQLDumpRunDispatchesDrivers(t *testing.T) {
	clientErr := errors.New("client failed")
	oldNewSimple := newSimpleGormClient
	newSimpleGormClient = func(string, string) (*gorm.DB, error) {
		return nil, clientErr
	}
	defer func() { newSimpleGormClient = oldNewSimple }()

	if err := NewSQLDump("mysql", "dsn", t.TempDir(), "users", true).Run(); !errors.Is(err, clientErr) {
		t.Fatalf("expected mysql dispatch error, got %v", err)
	}

	installPgDump(t, "#!/bin/sh\nexit 0\n")
	dsn := "host=127.0.0.1 port=5432 user=pg password=secret dbname=app sslmode=disable"
	if err := NewSQLDump("postgres", dsn, t.TempDir(), "users", true).Run(); !errors.Is(err, clientErr) {
		t.Fatalf("expected postgres dispatch error, got %v", err)
	}
}

func TestDumpMySQLRejectsNilDBClient(t *testing.T) {
	oldNewSimple := newSimpleGormClient
	defer func() { newSimpleGormClient = oldNewSimple }()
	newSimpleGormClient = func(string, string) (*gorm.DB, error) {
		return nil, nil
	}

	err := NewSQLDump("mysql", "dsn", t.TempDir(), "users", true).DumpMySQL()
	if err == nil || !strings.Contains(err.Error(), "sqldump database client cannot be nil") {
		t.Fatalf("expected mysql nil db client error, got %v", err)
	}
}

func TestDumpPostgresRejectsNilDBClient(t *testing.T) {
	installPgDump(t, "#!/bin/sh\nexit 0\n")
	oldNewSimple := newSimpleGormClient
	defer func() { newSimpleGormClient = oldNewSimple }()
	newSimpleGormClient = func(string, string) (*gorm.DB, error) {
		return nil, nil
	}

	dsn := "host=127.0.0.1 port=5432 user=pg password=secret dbname=app sslmode=disable"
	err := NewSQLDump("postgres", dsn, t.TempDir(), "users", true).DumpPostgres()
	if err == nil || !strings.Contains(err.Error(), "sqldump database client cannot be nil") {
		t.Fatalf("expected postgres nil db client error, got %v", err)
	}
}
