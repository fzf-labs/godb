package sqldump

import (
	"errors"
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
