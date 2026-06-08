package sqldump

import (
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestBuildMySQLShowCreateTableSQL_QuotesIdentifiers(t *testing.T) {
	got := buildMySQLShowCreateTableSQL("app`db", "user`log")
	want := "SHOW CREATE TABLE `app``db`.`user``log`"
	if got != want {
		t.Fatalf("unexpected sql: got=%q want=%q", got, want)
	}
}

func TestDumpMySQLReturnsDriverError(t *testing.T) {
	err := NewSQLDump("sqlite", "ignored", t.TempDir(), "users", false).DumpMySQL()
	if err == nil {
		t.Fatal("expected unknown driver error")
	}
}

func TestDumpMySQLWritesCreateTable(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer sqlDB.Close()

	db, err := gorm.Open(mysql.New(mysql.Config{Conn: sqlDB, SkipInitializeWithVersion: true}), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}

	oldNewSimple := newSimpleGormClient
	newSimpleGormClient = func(driver, dsn string) (*gorm.DB, error) {
		if driver != "mysql" || dsn != "dsn" {
			t.Fatalf("unexpected connection args: %s %s", driver, dsn)
		}
		return db, nil
	}
	defer func() { newSimpleGormClient = oldNewSimple }()

	currentDBQuery := regexp.QuoteMeta("SELECT SCHEMA_NAME from Information_schema.SCHEMATA where SCHEMA_NAME LIKE ? ORDER BY SCHEMA_NAME=? DESC,SCHEMA_NAME limit 1")
	mock.ExpectQuery(currentDBQuery).
		WithArgs("%", "").
		WillReturnRows(sqlmock.NewRows([]string{"SCHEMA_NAME"}).AddRow("app"))
	mock.ExpectQuery(regexp.QuoteMeta("SHOW CREATE TABLE `app`.`users`")).
		WillReturnRows(sqlmock.NewRows([]string{"Table", "Create Table"}).
			AddRow("users", "CREATE TABLE users (id bigint)"))
	mock.ExpectClose()

	outDir := t.TempDir()
	if err := NewSQLDump("mysql", "dsn", outDir, "users", true).DumpMySQL(); err != nil {
		t.Fatal(err)
	}

	content, err := os.ReadFile(filepath.Join(outDir, "app", "users.sql"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(content), "CREATE TABLE users") {
		t.Fatalf("unexpected dump content: %s", string(content))
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestDumpMySQLReturnsGetTablesError(t *testing.T) {
	db, _ := openMySQLDumpMock(t)
	restore := replaceDumpClient(t, db)
	defer restore()

	err := NewSQLDump("mysql", "dsn", t.TempDir(), "", true).DumpMySQL()
	if err == nil {
		t.Fatal("expected get tables error")
	}
}

func TestDumpMySQLRejectsEmptyTableSet(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	oldNewSimple := newSimpleGormClient
	newSimpleGormClient = func(driver, dsn string) (*gorm.DB, error) {
		if driver != "mysql" || dsn != "dsn" {
			t.Fatalf("unexpected connection args: %s %s", driver, dsn)
		}
		return db, nil
	}
	defer func() { newSimpleGormClient = oldNewSimple }()

	err = NewSQLDump("mysql", "dsn", t.TempDir(), "", true).DumpMySQL()
	if err == nil || !strings.Contains(err.Error(), "no tables to dump") {
		t.Fatalf("expected empty table set error, got %v", err)
	}
}

func TestDumpMySQLReturnsMkdirError(t *testing.T) {
	db, mock := openMySQLDumpMock(t)
	restore := replaceDumpClient(t, db)
	defer restore()
	expectDumpCurrentDatabase(mock, "app")

	outDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(outDir, "app"), []byte("file"), 0600); err != nil {
		t.Fatal(err)
	}

	err := NewSQLDump("mysql", "dsn", outDir, "users", true).DumpMySQL()
	if err == nil || !strings.Contains(err.Error(), "create output path") {
		t.Fatalf("expected mkdir error, got %v", err)
	}
}

func TestDumpMySQLReturnsRawError(t *testing.T) {
	db, mock := openMySQLDumpMock(t)
	restore := replaceDumpClient(t, db)
	defer restore()
	expectDumpCurrentDatabase(mock, "app")
	rawErr := errors.New("raw failed")
	mock.ExpectQuery(regexp.QuoteMeta("SHOW CREATE TABLE `app`.`users`")).WillReturnError(rawErr)

	err := NewSQLDump("mysql", "dsn", t.TempDir(), "users", true).DumpMySQL()
	if !errors.Is(err, rawErr) {
		t.Fatalf("expected raw error, got %v", err)
	}
}

func TestDumpMySQLSkipsExistingFile(t *testing.T) {
	db, mock := openMySQLDumpMock(t)
	restore := replaceDumpClient(t, db)
	defer restore()
	expectDumpCurrentDatabase(mock, "app")

	outDir := t.TempDir()
	appDir := filepath.Join(outDir, "app")
	if err := os.MkdirAll(appDir, 0755); err != nil {
		t.Fatal(err)
	}
	outFile := filepath.Join(appDir, "users.sql")
	if err := os.WriteFile(outFile, []byte("keep"), 0600); err != nil {
		t.Fatal(err)
	}

	if err := NewSQLDump("mysql", "dsn", outDir, "users", false).DumpMySQL(); err != nil {
		t.Fatal(err)
	}
	content, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "keep" {
		t.Fatalf("expected existing file to be preserved, got %s", string(content))
	}
}

func TestDumpMySQLRejectsUnsafeOutputFileName(t *testing.T) {
	db, mock := openMySQLDumpMock(t)
	restore := replaceDumpClient(t, db)
	defer restore()
	expectDumpCurrentDatabase(mock, "app")

	err := NewSQLDump("mysql", "dsn", t.TempDir(), "../users", true).DumpMySQL()
	if err == nil || !strings.Contains(err.Error(), "unsafe output file name") {
		t.Fatalf("expected unsafe output file name error, got %v", err)
	}
}

func TestDumpMySQLOverwritesExistingFileWithEmptyCleanedOutput(t *testing.T) {
	db, mock := openMySQLDumpMock(t)
	restore := replaceDumpClient(t, db)
	defer restore()
	expectDumpCurrentDatabase(mock, "app")
	mock.ExpectQuery(regexp.QuoteMeta("SHOW CREATE TABLE `app`.`users`")).
		WillReturnRows(sqlmock.NewRows([]string{"Table", "Create Table"}).
			AddRow("users", ""))

	outDir := t.TempDir()
	appDir := filepath.Join(outDir, "app")
	if err := os.MkdirAll(appDir, 0755); err != nil {
		t.Fatal(err)
	}
	outFile := filepath.Join(appDir, "users.sql")
	if err := os.WriteFile(outFile, []byte("stale"), 0600); err != nil {
		t.Fatal(err)
	}

	if err := NewSQLDump("mysql", "dsn", outDir, "users", true).DumpMySQL(); err != nil {
		t.Fatal(err)
	}

	content, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "" {
		t.Fatalf("expected empty overwritten file, got %q", string(content))
	}
}

func TestDumpMySQLReturnsWriteError(t *testing.T) {
	db, mock := openMySQLDumpMock(t)
	restore := replaceDumpClient(t, db)
	defer restore()
	expectDumpCurrentDatabase(mock, "app")
	mock.ExpectQuery(regexp.QuoteMeta("SHOW CREATE TABLE `app`.`users`")).
		WillReturnRows(sqlmock.NewRows([]string{"Table", "Create Table"}).
			AddRow("users", "CREATE TABLE users (id bigint)"))

	outDir := t.TempDir()
	outFile := filepath.Join(outDir, "app", "users.sql")
	if err := os.MkdirAll(outFile, 0755); err != nil {
		t.Fatal(err)
	}

	err := NewSQLDump("mysql", "dsn", outDir, "users", true).DumpMySQL()
	if err == nil || !strings.Contains(err.Error(), "write") {
		t.Fatalf("expected write error, got %v", err)
	}
}

func openMySQLDumpMock(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	t.Helper()
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	db, err := gorm.Open(mysql.New(mysql.Config{Conn: sqlDB, SkipInitializeWithVersion: true}), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	return db, mock
}

func replaceDumpClient(t *testing.T, db *gorm.DB) func() {
	t.Helper()
	oldNewSimple := newSimpleGormClient
	newSimpleGormClient = func(driver, dsn string) (*gorm.DB, error) {
		if driver != "mysql" || dsn != "dsn" {
			t.Fatalf("unexpected connection args: %s %s", driver, dsn)
		}
		return db, nil
	}
	return func() { newSimpleGormClient = oldNewSimple }
}

func expectDumpCurrentDatabase(mock sqlmock.Sqlmock, dbName string) {
	currentDBQuery := regexp.QuoteMeta("SELECT SCHEMA_NAME from Information_schema.SCHEMATA where SCHEMA_NAME LIKE ? ORDER BY SCHEMA_NAME=? DESC,SCHEMA_NAME limit 1")
	mock.ExpectQuery(currentDBQuery).
		WithArgs("%", "").
		WillReturnRows(sqlmock.NewRows([]string{"SCHEMA_NAME"}).AddRow(dbName))
}
