package sqldump

import (
	"strings"
	"testing"

	"gorm.io/gorm"
)

func TestRunCommandReturnsSQLDumpError(t *testing.T) {
	oldDB, oldDSN, oldOut, oldTables, oldOverwrite := db, dsn, outPutPath, targetTables, fileOverwrite
	defer func() {
		db, dsn, outPutPath, targetTables, fileOverwrite = oldDB, oldDSN, oldOut, oldTables, oldOverwrite
	}()

	db = "sqlite"
	dsn = ":memory:"
	outPutPath = t.TempDir()
	targetTables = "users"
	fileOverwrite = true

	if err := Run(nil, nil); err == nil {
		t.Fatal("expected unknown database type error")
	}
}

func TestRunCommandRejectsBlankTablesBeforeConnecting(t *testing.T) {
	oldDB, oldDSN, oldOut, oldTables, oldOverwrite := db, dsn, outPutPath, targetTables, fileOverwrite
	oldNewSimple := newSimpleGormClient
	defer func() {
		db, dsn, outPutPath, targetTables, fileOverwrite = oldDB, oldDSN, oldOut, oldTables, oldOverwrite
		newSimpleGormClient = oldNewSimple
	}()

	db = "mysql"
	dsn = "dsn"
	outPutPath = t.TempDir()
	targetTables = " , "
	fileOverwrite = true
	newSimpleGormClient = func(string, string) (*gorm.DB, error) {
		t.Fatal("expected table parsing to fail before connecting")
		return nil, nil
	}

	err := Run(nil, nil)
	if err == nil || !strings.Contains(err.Error(), "no table names found") {
		t.Fatalf("expected table parsing error, got %v", err)
	}
}

func TestRunWithOptionsRejectsMissingRequiredFieldsBeforeConnecting(t *testing.T) {
	oldNewSimple := newSimpleGormClient
	defer func() { newSimpleGormClient = oldNewSimple }()
	newSimpleGormClient = func(string, string) (*gorm.DB, error) {
		t.Fatal("expected validation to fail before connecting")
		return nil, nil
	}

	tests := []struct {
		name string
		opts runOptions
		want string
	}{
		{name: "db", opts: runOptions{dsn: "dsn", outPutPath: t.TempDir()}, want: "db cannot be empty"},
		{name: "dsn", opts: runOptions{db: "postgres", outPutPath: t.TempDir()}, want: "dsn cannot be empty"},
		{name: "output", opts: runOptions{db: "postgres", dsn: "dsn", outPutPath: " \t\n"}, want: "output path cannot be empty"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := runWithOptions(tt.opts)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("expected %q error, got %v", tt.want, err)
			}
		})
	}
}
