package ormgen

import (
	"strings"
	"testing"
)

func TestRunReturnsDriverErrorAfterBuildingOptions(t *testing.T) {
	oldDB, oldDSN, oldTables, oldOut := db, dsn, targetTables, outPutPath
	oldUnderline := optionUnderline
	oldPGDefault := optionPgDefaultString
	oldRemoveDefault := optionRemoveDefault
	oldRemoveType := optionRemoveGormTypeTag
	defer func() {
		db, dsn, targetTables, outPutPath = oldDB, oldDSN, oldTables, oldOut
		optionUnderline = oldUnderline
		optionPgDefaultString = oldPGDefault
		optionRemoveDefault = oldRemoveDefault
		optionRemoveGormTypeTag = oldRemoveType
	}()

	db = "sqlite"
	dsn = ":memory:"
	targetTables = "users,roles"
	outPutPath = t.TempDir()
	optionUnderline = "UL"
	optionPgDefaultString = true
	optionRemoveDefault = true
	optionRemoveGormTypeTag = true

	if err := Run(nil, nil); err == nil {
		t.Fatal("expected unknown driver error")
	}
}

func TestRunRejectsBlankTables(t *testing.T) {
	oldDB, oldDSN, oldTables, oldOut := db, dsn, targetTables, outPutPath
	defer func() {
		db, dsn, targetTables, outPutPath = oldDB, oldDSN, oldTables, oldOut
	}()

	db = "sqlite"
	dsn = ":memory:"
	targetTables = " , "
	outPutPath = t.TempDir()

	err := Run(nil, nil)
	if err == nil || !strings.Contains(err.Error(), "no table names found") {
		t.Fatalf("expected table parsing error, got %v", err)
	}
}
