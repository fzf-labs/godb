package ormgen

import (
	"errors"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/fzf-labs/godb/orm/gen"
)

func TestRunReturnsDriverErrorAfterBuildingOptions(t *testing.T) {
	oldDB, oldDSN, oldTables, oldOut := db, dsn, targetTables, outPutPath
	oldUnderline := optionUnderline
	oldPGDefault := optionPgDefaultString
	oldRemoveDefault := optionRemoveDefault
	oldRemoveType := optionRemoveGormTypeTag
	oldNewSimple := newSimpleGormClient
	defer func() {
		db, dsn, targetTables, outPutPath = oldDB, oldDSN, oldTables, oldOut
		optionUnderline = oldUnderline
		optionPgDefaultString = oldPGDefault
		optionRemoveDefault = oldRemoveDefault
		optionRemoveGormTypeTag = oldRemoveType
		newSimpleGormClient = oldNewSimple
	}()

	sentinelErr := errors.New("connect failed")
	newSimpleGormClient = func(driver, dataSource string) (*gorm.DB, error) {
		if driver != "postgres" || dataSource != "dsn" {
			t.Fatalf("unexpected connection args: %s %s", driver, dataSource)
		}
		return nil, sentinelErr
	}

	db = " POSTGRES "
	dsn = " dsn "
	targetTables = " users , roles "
	outPutPath = " " + t.TempDir() + " "
	optionUnderline = "UL"
	optionPgDefaultString = true
	optionRemoveDefault = true
	optionRemoveGormTypeTag = true

	if err := Run(nil, nil); !errors.Is(err, sentinelErr) {
		t.Fatalf("expected connection error, got %v", err)
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

func TestRunWithOptionsRejectsUnsafeTableNamesBeforeConnecting(t *testing.T) {
	oldNewSimple := newSimpleGormClient
	defer func() { newSimpleGormClient = oldNewSimple }()
	newSimpleGormClient = func(string, string) (*gorm.DB, error) {
		t.Fatal("expected table validation to fail before connecting")
		return nil, nil
	}

	err := runWithOptions(runOptions{
		db:           "postgres",
		dsn:          "dsn",
		outPutPath:   t.TempDir(),
		targetTables: "users,bad name",
	})
	if err == nil || !strings.Contains(err.Error(), "whitespace or control characters") {
		t.Fatalf("expected unsafe table name error, got %v", err)
	}
}

func TestRunWithOptionsRejectsNilDBClient(t *testing.T) {
	oldNewSimple := newSimpleGormClient
	defer func() { newSimpleGormClient = oldNewSimple }()
	newSimpleGormClient = func(string, string) (*gorm.DB, error) {
		return nil, nil
	}

	err := runWithOptions(runOptions{
		db:           "postgres",
		dsn:          "dsn",
		outPutPath:   t.TempDir(),
		targetTables: "users",
	})
	if err == nil || !strings.Contains(err.Error(), "ormgen database client cannot be nil") {
		t.Fatalf("expected ormgen nil db client error, got %v", err)
	}
}

func TestRunWithOptionsUsesProvidedSnapshotAndClosesDB(t *testing.T) {
	oldDB, oldDSN, oldTables, oldOut := db, dsn, targetTables, outPutPath
	oldNewSimple := newSimpleGormClient
	defer func() {
		db, dsn, targetTables, outPutPath = oldDB, oldDSN, oldTables, oldOut
		newSimpleGormClient = oldNewSimple
	}()

	db = ""
	dsn = ""
	targetTables = " , "
	outPutPath = ""

	sentinelErr := errors.New("generation failed")
	dbClient, assertClosed := closeTrackingDB(t)
	newSimpleGormClient = func(driver, dataSource string) (*gorm.DB, error) {
		if driver != "postgres" || dataSource != "dsn" {
			t.Fatalf("unexpected connection args: %s %s", driver, dataSource)
		}
		return dbClient, nil
	}
	generateDBDo = func(*gen.GenerationDB) error {
		return sentinelErr
	}
	defer func() { generateDBDo = (*gen.GenerationDB).Do }()

	err := runWithOptions(runOptions{
		db:                    " postgres ",
		dsn:                   " dsn ",
		outPutPath:            t.TempDir(),
		targetTables:          "users",
		optionUnderline:       "UL",
		optionPgDefaultString: true,
		optionRemoveDefault:   true,
	})
	if !errors.Is(err, sentinelErr) {
		t.Fatalf("expected generation error, got %v", err)
	}
	assertClosed()
}

func TestCloseGormDBHandlesNilAndClosesSQLite(t *testing.T) {
	closeGormDB(nil)

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}

	closeGormDB(db)
	if err := sqlDB.Ping(); err == nil {
		t.Fatal("expected sqlite db to be closed")
	}
}

func closeTrackingDB(t *testing.T) (*gorm.DB, func()) {
	t.Helper()
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	mock.ExpectClose().WillReturnError(nil)
	db, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	return db, func() {
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
	}
}
