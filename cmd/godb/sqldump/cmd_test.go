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
