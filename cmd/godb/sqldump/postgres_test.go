package sqldump

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestPostgresDsnParse_KeywordValue(t *testing.T) {
	s := &SQLDump{
		dsn: "host=127.0.0.1 port=5432 user=postgres password='pa ss=word' dbname=test_db sslmode=disable",
	}
	got, err := s.postgresDsnParse()
	if err != nil {
		t.Fatal(err)
	}
	if got.Host != "127.0.0.1" || got.Port != 5432 || got.User != "postgres" || got.Password != "pa ss=word" || got.Dbname != "test_db" {
		t.Fatalf("unexpected parse result: %+v", got)
	}
}

func TestPostgresDsnParse_URL(t *testing.T) {
	s := &SQLDump{
		dsn: "postgres://pguser:p%40ss@localhost:5433/app_db?sslmode=disable",
	}
	got, err := s.postgresDsnParse()
	if err != nil {
		t.Fatal(err)
	}
	if got.Host != "localhost" || got.Port != 5433 || got.User != "pguser" || got.Password != "p@ss" || got.Dbname != "app_db" {
		t.Fatalf("unexpected parse result: %+v", got)
	}
}

func TestPostgresDsnParse_Invalid(t *testing.T) {
	s := &SQLDump{dsn: ":::bad dsn:::"}
	if _, err := s.postgresDsnParse(); err == nil {
		t.Fatal("expected parse error, got nil")
	}
}

func TestOutputDir_PreservesAbsoluteBasePath(t *testing.T) {
	got := outputDir("/tmp/sql", "app_db")
	if got != "/tmp/sql/app_db" {
		t.Fatalf("unexpected output dir: %s", got)
	}
}

func TestBuildPgDumpArgs_UsesDatabaseFlag(t *testing.T) {
	args := buildPgDumpArgs(&PostgresDsn{
		Host:   "127.0.0.1",
		Port:   5432,
		User:   "postgres",
		Dbname: "app_db",
	}, "users")
	want := []string{"-h", "127.0.0.1", "-p", "5432", "-U", "postgres", "-d", "app_db", "-s", "-t", `"users"`}
	if len(args) != len(want) {
		t.Fatalf("unexpected arg len: got=%d want=%d", len(args), len(want))
	}
	for i := range want {
		if args[i] != want[i] {
			t.Fatalf("unexpected args: got=%v want=%v", args, want)
		}
	}
}

func TestBuildPgDumpArgs_QuotesQualifiedAndMixedCaseTablePattern(t *testing.T) {
	args := buildPgDumpArgs(&PostgresDsn{
		Host:   "127.0.0.1",
		Port:   5432,
		User:   "postgres",
		Dbname: "app_db",
	}, `app_schema.User.Events`)
	got := args[len(args)-1]
	want := `"app_schema"."User"."Events"`
	if got != want {
		t.Fatalf("unexpected table pattern: got=%q want=%q", got, want)
	}
}

func TestShouldSkipLine(t *testing.T) {
	dump := &SQLDump{}
	tests := []struct {
		line string
		want bool
	}{
		{line: "", want: true},
		{line: "-- comment", want: true},
		{line: "SELECT pg_catalog.set_config('search_path', '', false);", want: true},
		{line: "SET statement_timeout = 0;", want: true},
		{line: `\restrict abc`, want: true},
		{line: `\unrestrict abc`, want: true},
		{line: "ALTER TABLE public.users OWNER TO postgres;", want: true},
		{line: "ALTER TABLE public.users OWNER TO app_owner;", want: true},
		{line: "CREATE TABLE public.users (id bigint);", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.line, func(t *testing.T) {
			if got := dump.shouldSkipLine(tt.line); got != tt.want {
				t.Fatalf("got %v want %v", got, tt.want)
			}
		})
	}
}

func TestPostgresRemoveFiltersAndFlattensStatements(t *testing.T) {
	dump := &SQLDump{}
	input := `-- dumped by pg_dump
SET statement_timeout = 0;

CREATE TABLE public.users (
    id bigint NOT NULL
);
ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);
ALTER TABLE public.users OWNER TO postgres;
SELECT pg_catalog.set_config('search_path', '', false);
`

	got := dump.postgresRemove(input)
	for _, forbidden := range []string{"dumped by", "SET statement_timeout", "OWNER TO postgres", "SELECT pg_catalog"} {
		if strings.Contains(got, forbidden) {
			t.Fatalf("expected %q to be removed from:\n%s", forbidden, got)
		}
	}
	wantAlter := "ALTER TABLE ONLY public.users ADD CONSTRAINT users_pkey PRIMARY KEY (id);"
	if !strings.Contains(got, wantAlter) {
		t.Fatalf("expected flattened alter statement %q in:\n%s", wantAlter, got)
	}
	if !strings.Contains(got, "CREATE TABLE public.users") {
		t.Fatalf("expected create table to remain:\n%s", got)
	}
}

func TestPostgresRemoveFiltersMultilineOwnerStatement(t *testing.T) {
	dump := &SQLDump{}
	input := `CREATE TABLE public.users (
    id bigint NOT NULL
);
ALTER TABLE public.users
    OWNER TO postgres;
ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);
`

	got := dump.postgresRemove(input)
	if strings.Contains(got, "OWNER TO postgres") {
		t.Fatalf("expected multiline owner statement to be removed:\n%s", got)
	}
	wantAlter := "ALTER TABLE ONLY public.users ADD CONSTRAINT users_pkey PRIMARY KEY (id);"
	if !strings.Contains(got, wantAlter) {
		t.Fatalf("expected non-owner alter statement to remain:\n%s", got)
	}
}

func TestPostgresRemoveEmpty(t *testing.T) {
	if got := (&SQLDump{}).postgresRemove(""); got != "" {
		t.Fatalf("got %q want empty", got)
	}
}

func TestPostgresRemovePreservesLongLines(t *testing.T) {
	longDefault := strings.Repeat("x", 70*1024)
	input := "CREATE TABLE public.big_defaults (\n    payload text DEFAULT '" + longDefault + "'\n);\n"

	got := (&SQLDump{}).postgresRemove(input)
	if !strings.Contains(got, longDefault) {
		t.Fatalf("expected long line to be preserved, got length %d", len(got))
	}
	if !strings.Contains(got, "CREATE TABLE public.big_defaults") {
		t.Fatalf("expected create table statement to remain:\n%s", got)
	}
}

func TestDumpPostgresWritesAndSkipsExistingFiles(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	oldNewSimple := newSimpleGormClient
	newSimpleGormClient = func(driver, dsn string) (*gorm.DB, error) {
		if driver != "postgres" || !strings.Contains(dsn, "dbname=app") {
			t.Fatalf("unexpected connection args: %s %s", driver, dsn)
		}
		return db, nil
	}
	defer func() { newSimpleGormClient = oldNewSimple }()

	binDir := t.TempDir()
	pgDump := filepath.Join(binDir, "pg_dump")
	script := `#!/bin/sh
case " $* " in
  *" -t users "*)
    echo "unexpected pg_dump invocation for existing users file" >&2
    exit 99
    ;;
esac
cat <<'SQL'
-- comment
SET statement_timeout = 0;
CREATE TABLE public.users (id bigint);
ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);
ALTER TABLE public.users OWNER TO postgres;
SQL
`
	if err := os.WriteFile(pgDump, []byte(script), 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	outDir := t.TempDir()
	existingDir := filepath.Join(outDir, "app")
	if err := os.MkdirAll(existingDir, 0755); err != nil {
		t.Fatal(err)
	}
	existingFile := filepath.Join(existingDir, "users.sql")
	if err := os.WriteFile(existingFile, []byte("keep"), 0600); err != nil {
		t.Fatal(err)
	}

	dsn := "host=127.0.0.1 port=5432 user=pg password=secret dbname=app sslmode=disable"
	if err := NewSQLDump("postgres", dsn, outDir, "users,roles", false).DumpPostgres(); err != nil {
		t.Fatal(err)
	}

	existingContent, err := os.ReadFile(existingFile)
	if err != nil {
		t.Fatal(err)
	}
	if string(existingContent) != "keep" {
		t.Fatalf("existing file should not be overwritten: %s", string(existingContent))
	}

	content, err := os.ReadFile(filepath.Join(existingDir, "roles.sql"))
	if err != nil {
		t.Fatal(err)
	}
	got := string(content)
	if strings.Contains(got, "OWNER TO postgres") || strings.Contains(got, "SET statement_timeout") {
		t.Fatalf("dump content was not cleaned:\n%s", got)
	}
	if !strings.Contains(got, "ALTER TABLE ONLY public.users ADD CONSTRAINT users_pkey PRIMARY KEY (id);") {
		t.Fatalf("expected flattened alter statement:\n%s", got)
	}
}

func TestDumpPostgresOverwritesExistingFileWithEmptyCleanedOutput(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	oldNewSimple := newSimpleGormClient
	newSimpleGormClient = func(driver, dsn string) (*gorm.DB, error) {
		if driver != "postgres" || !strings.Contains(dsn, "dbname=app") {
			t.Fatalf("unexpected connection args: %s %s", driver, dsn)
		}
		return db, nil
	}
	defer func() { newSimpleGormClient = oldNewSimple }()

	binDir := t.TempDir()
	pgDump := filepath.Join(binDir, "pg_dump")
	script := `#!/bin/sh
cat <<'SQL'
-- comment
SET statement_timeout = 0;
ALTER TABLE public.users OWNER TO postgres;
SQL
`
	if err := os.WriteFile(pgDump, []byte(script), 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	outDir := t.TempDir()
	existingDir := filepath.Join(outDir, "app")
	if err := os.MkdirAll(existingDir, 0755); err != nil {
		t.Fatal(err)
	}
	existingFile := filepath.Join(existingDir, "users.sql")
	if err := os.WriteFile(existingFile, []byte("stale"), 0600); err != nil {
		t.Fatal(err)
	}

	dsn := "host=127.0.0.1 port=5432 user=pg password=secret dbname=app sslmode=disable"
	if err := NewSQLDump("postgres", dsn, outDir, "users", true).DumpPostgres(); err != nil {
		t.Fatal(err)
	}

	content, err := os.ReadFile(existingFile)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "" {
		t.Fatalf("expected empty overwritten file, got %q", string(content))
	}
}

func TestDumpPostgresRejectsUnsafeOutputFileName(t *testing.T) {
	installPgDump(t, "#!/bin/sh\nexit 0\n")
	restore := replacePostgresDumpClient(t)
	defer restore()

	dsn := "host=127.0.0.1 port=5432 user=pg password=secret dbname=app sslmode=disable"
	err := NewSQLDump("postgres", dsn, t.TempDir(), "../users", true).DumpPostgres()
	if err == nil || !strings.Contains(err.Error(), "unsafe output file name") {
		t.Fatalf("expected unsafe output file name error, got %v", err)
	}
}

func TestDumpPostgresReturnsMissingCommand(t *testing.T) {
	t.Setenv("PATH", t.TempDir())

	err := NewSQLDump("postgres", "host=127.0.0.1 dbname=app", t.TempDir(), "users", true).DumpPostgres()
	if err == nil || !strings.Contains(err.Error(), "pg_dump not found") {
		t.Fatalf("expected pg_dump lookup error, got %v", err)
	}
}

func TestDumpPostgresReturnsClientError(t *testing.T) {
	installPgDump(t, "#!/bin/sh\nexit 0\n")
	clientErr := errors.New("client failed")
	oldNewSimple := newSimpleGormClient
	newSimpleGormClient = func(string, string) (*gorm.DB, error) {
		return nil, clientErr
	}
	defer func() { newSimpleGormClient = oldNewSimple }()

	err := NewSQLDump("postgres", "host=127.0.0.1 dbname=app", t.TempDir(), "users", true).DumpPostgres()
	if !errors.Is(err, clientErr) {
		t.Fatalf("expected client error, got %v", err)
	}
}

func TestDumpPostgresReturnsGetTablesError(t *testing.T) {
	installPgDump(t, "#!/bin/sh\nexit 0\n")
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	if err := sqlDB.Close(); err != nil {
		t.Fatal(err)
	}
	oldNewSimple := newSimpleGormClient
	newSimpleGormClient = func(string, string) (*gorm.DB, error) {
		return db, nil
	}
	defer func() { newSimpleGormClient = oldNewSimple }()

	dsn := "host=127.0.0.1 port=5432 user=pg password=secret dbname=app sslmode=disable"
	err = NewSQLDump("postgres", dsn, t.TempDir(), "", true).DumpPostgres()
	if err == nil {
		t.Fatal("expected get tables error")
	}
}

func TestDumpPostgresRejectsEmptyTableSet(t *testing.T) {
	installPgDump(t, "#!/bin/sh\nexit 0\n")
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	oldNewSimple := newSimpleGormClient
	newSimpleGormClient = func(string, string) (*gorm.DB, error) {
		return db, nil
	}
	defer func() { newSimpleGormClient = oldNewSimple }()

	err = NewSQLDump("postgres", "host=127.0.0.1 port=5432 user=pg password=secret dbname=app sslmode=disable", t.TempDir(), "", true).DumpPostgres()
	if err == nil || !strings.Contains(err.Error(), "no tables to dump") {
		t.Fatalf("expected empty table set error, got %v", err)
	}
}

func TestDumpPostgresReturnsDSNParseError(t *testing.T) {
	installPgDump(t, "#!/bin/sh\nexit 0\n")
	restore := replacePostgresDumpClient(t)
	defer restore()

	err := NewSQLDump("postgres", ":::bad dsn:::", t.TempDir(), "users", true).DumpPostgres()
	if err == nil || !strings.Contains(err.Error(), "parse postgres dsn") {
		t.Fatalf("expected dsn parse error, got %v", err)
	}
}

func TestDumpPostgresClosesClientAfterDSNParseError(t *testing.T) {
	installPgDump(t, "#!/bin/sh\nexit 0\n")
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer sqlDB.Close()
	mock.ExpectClose()
	db, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	oldNewSimple := newSimpleGormClient
	newSimpleGormClient = func(string, string) (*gorm.DB, error) {
		return db, nil
	}
	defer func() { newSimpleGormClient = oldNewSimple }()

	err = NewSQLDump("postgres", ":::bad dsn:::", t.TempDir(), "users", true).DumpPostgres()
	if err == nil || !strings.Contains(err.Error(), "parse postgres dsn") {
		t.Fatalf("expected dsn parse error, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestDumpPostgresReturnsMkdirError(t *testing.T) {
	installPgDump(t, "#!/bin/sh\nexit 0\n")
	restore := replacePostgresDumpClient(t)
	defer restore()

	outDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(outDir, "app"), []byte("file"), 0600); err != nil {
		t.Fatal(err)
	}
	dsn := "host=127.0.0.1 port=5432 user=pg password=secret dbname=app sslmode=disable"
	err := NewSQLDump("postgres", dsn, outDir, "users", true).DumpPostgres()
	if err == nil || !strings.Contains(err.Error(), "create output path") {
		t.Fatalf("expected mkdir error, got %v", err)
	}
}

func TestDumpPostgresReturnsCommandError(t *testing.T) {
	installPgDump(t, "#!/bin/sh\necho broken dump >&2\nexit 3\n")
	restore := replacePostgresDumpClient(t)
	defer restore()

	dsn := "host=127.0.0.1 port=5432 user=pg password=secret dbname=app sslmode=disable"
	err := NewSQLDump("postgres", dsn, t.TempDir(), "users", true).DumpPostgres()
	if err == nil || !strings.Contains(err.Error(), "pg_dump table users") || !strings.Contains(err.Error(), "broken dump") {
		t.Fatalf("expected command error, got %v", err)
	}
}

func TestDumpPostgresReturnsCommandTimeout(t *testing.T) {
	restore := replacePostgresDumpClient(t)
	defer restore()
	oldPgDumpTimeout := pgDumpTimeout
	pgDumpTimeout = time.Nanosecond
	defer func() { pgDumpTimeout = oldPgDumpTimeout }()

	oldExec := execPgDumpCommand
	execPgDumpCommand = func(ctx context.Context, _ *PostgresDsn, _ string) ([]byte, []byte, error) {
		<-ctx.Done()
		return nil, []byte("still running"), ctx.Err()
	}
	defer func() { execPgDumpCommand = oldExec }()

	dsn := "host=127.0.0.1 port=5432 user=pg password=secret dbname=app sslmode=disable"
	err := NewSQLDump("postgres", dsn, t.TempDir(), "users", true).DumpPostgres()
	if err == nil || !strings.Contains(err.Error(), "timed out") || !strings.Contains(err.Error(), "still running") {
		t.Fatalf("expected timeout error with stderr, got %v", err)
	}
}

func TestExecPgDumpCommandBuildsCommandWithPassword(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, _, err := execPgDumpCommand(ctx, &PostgresDsn{Host: "127.0.0.1", Port: 5432, User: "pg", Password: "secret", Dbname: "app"}, "users")
	if !errors.Is(err, context.Canceled) && !errors.Is(err, exec.ErrNotFound) {
		t.Fatalf("expected canceled or missing command error, got %v", err)
	}
}

func TestDumpPostgresReturnsWriteError(t *testing.T) {
	installPgDump(t, `#!/bin/sh
cat <<'SQL'
CREATE TABLE public.users (id bigint);
SQL
`)
	restore := replacePostgresDumpClient(t)
	defer restore()

	outDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(outDir, "app", "users.sql"), 0755); err != nil {
		t.Fatal(err)
	}
	dsn := "host=127.0.0.1 port=5432 user=pg password=secret dbname=app sslmode=disable"
	err := NewSQLDump("postgres", dsn, outDir, "users", true).DumpPostgres()
	if err == nil || !strings.Contains(err.Error(), "write") {
		t.Fatalf("expected write error, got %v", err)
	}
}

func installPgDump(t *testing.T, script string) {
	t.Helper()
	binDir := t.TempDir()
	pgDump := filepath.Join(binDir, "pg_dump")
	if err := os.WriteFile(pgDump, []byte(script), 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
}

func replacePostgresDumpClient(t *testing.T) func() {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	oldNewSimple := newSimpleGormClient
	newSimpleGormClient = func(driver, dsn string) (*gorm.DB, error) {
		if driver != "postgres" {
			t.Fatalf("unexpected driver: %s", driver)
		}
		return db, nil
	}
	return func() { newSimpleGormClient = oldNewSimple }
}
