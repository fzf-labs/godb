package sqltopb

import (
	"errors"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/fzf-labs/godb/orm/gen"
)

func TestRunReturnsDriverErrorAfterParsingTables(t *testing.T) {
	oldDB, oldDSN, oldTables := db, dsn, targetTables
	oldPBPackage, oldPBGoPackage, oldOut := pbPackage, pbGoPackage, outPutPath
	defer func() {
		db, dsn, targetTables = oldDB, oldDSN, oldTables
		pbPackage, pbGoPackage, outPutPath = oldPBPackage, oldPBGoPackage, oldOut
	}()

	db = "sqlite"
	dsn = ":memory:"
	targetTables = "users,roles"
	pbPackage = "pb"
	pbGoPackage = "example.com/project/pb;pb"
	outPutPath = t.TempDir()

	if err := Run(nil, nil); err == nil {
		t.Fatal("expected unknown driver error")
	}
}

func TestRunRejectsBlankTables(t *testing.T) {
	oldDB, oldDSN, oldTables := db, dsn, targetTables
	oldPBPackage, oldPBGoPackage, oldOut := pbPackage, pbGoPackage, outPutPath
	defer func() {
		db, dsn, targetTables = oldDB, oldDSN, oldTables
		pbPackage, pbGoPackage, outPutPath = oldPBPackage, oldPBGoPackage, oldOut
	}()

	db = "sqlite"
	dsn = ":memory:"
	targetTables = " , "
	pbPackage = "pb"
	pbGoPackage = "example.com/project/pb;pb"
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
		{name: "db", opts: runOptions{dsn: "dsn", outPutPath: t.TempDir(), pbPackage: "pb", pbGoPackage: "example.com/pb;pb"}, want: "db cannot be empty"},
		{name: "dsn", opts: runOptions{db: "postgres", outPutPath: t.TempDir(), pbPackage: "pb", pbGoPackage: "example.com/pb;pb"}, want: "dsn cannot be empty"},
		{name: "output", opts: runOptions{db: "postgres", dsn: "dsn", outPutPath: " \t\n", pbPackage: "pb", pbGoPackage: "example.com/pb;pb"}, want: "output path cannot be empty"},
		{name: "package", opts: runOptions{db: "postgres", dsn: "dsn", outPutPath: t.TempDir(), pbGoPackage: "example.com/pb;pb"}, want: "pb package cannot be empty"},
		{name: "go package", opts: runOptions{db: "postgres", dsn: "dsn", outPutPath: t.TempDir(), pbPackage: "pb"}, want: "pb go package cannot be empty"},
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

func TestRunWithOptionsUsesProvidedSnapshotAndClosesDB(t *testing.T) {
	oldDB, oldDSN, oldTables := db, dsn, targetTables
	oldPBPackage, oldPBGoPackage, oldOut := pbPackage, pbGoPackage, outPutPath
	oldNewSimple := newSimpleGormClient
	defer func() {
		db, dsn, targetTables = oldDB, oldDSN, oldTables
		pbPackage, pbGoPackage, outPutPath = oldPBPackage, oldPBGoPackage, oldOut
		newSimpleGormClient = oldNewSimple
	}()

	db = ""
	dsn = ""
	targetTables = " , "
	pbPackage = ""
	pbGoPackage = ""
	outPutPath = ""

	sentinelErr := errors.New("pb generation failed")
	dbClient, assertClosed := closeTrackingDB(t)
	newSimpleGormClient = func(driver, dataSource string) (*gorm.DB, error) {
		if driver != "postgres" || dataSource != "dsn" {
			t.Fatalf("unexpected connection args: %s %s", driver, dataSource)
		}
		return dbClient, nil
	}
	generatePBDo = func(*gen.GenerationPb) error {
		return sentinelErr
	}
	defer func() { generatePBDo = (*gen.GenerationPb).Do }()

	err := runWithOptions(runOptions{
		db:           "postgres",
		dsn:          "dsn",
		targetTables: "users",
		pbPackage:    "pb",
		pbGoPackage:  "example.com/pb;pb",
		outPutPath:   t.TempDir(),
	})
	if !errors.Is(err, sentinelErr) {
		t.Fatalf("expected generation error, got %v", err)
	}
	assertClosed()
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
