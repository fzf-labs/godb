package sqldump

import (
	"errors"
	"strings"
	"testing"

	"gorm.io/gorm"
)

func TestRunCommandReturnsSQLDumpError(t *testing.T) {
	oldDB, oldDSN, oldOut, oldTables, oldOverwrite := db, dsn, outPutPath, targetTables, fileOverwrite
	oldNewSimple := newSimpleGormClient
	defer func() {
		db, dsn, outPutPath, targetTables, fileOverwrite = oldDB, oldDSN, oldOut, oldTables, oldOverwrite
		newSimpleGormClient = oldNewSimple
	}()

	sentinelErr := errors.New("connect failed")
	newSimpleGormClient = func(driver, dataSource string) (*gorm.DB, error) {
		if driver != "mysql" || dataSource != "dsn" {
			t.Fatalf("unexpected connection args: %s %s", driver, dataSource)
		}
		return nil, sentinelErr
	}

	db = " MYSQL "
	dsn = " dsn "
	outPutPath = " " + t.TempDir() + " "
	targetTables = " users "
	fileOverwrite = true

	if err := Run(nil, nil); !errors.Is(err, sentinelErr) {
		t.Fatalf("expected connection error, got %v", err)
	}
}

func TestRunCommandRejectsBlankTablesBeforeConnecting(t *testing.T) {
	oldDB, oldDSN, oldOut, oldTables, oldOverwrite := db, dsn, outPutPath, targetTables, fileOverwrite
	oldNewSimple := newSimpleGormClient
	defer func() {
		db, dsn, outPutPath, targetTables, fileOverwrite = oldDB, oldDSN, oldOut, oldTables, oldOverwrite
		newSimpleGormClient = oldNewSimple
	}()

	db = " MYSQL "
	dsn = " dsn "
	outPutPath = " " + t.TempDir() + " "
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
