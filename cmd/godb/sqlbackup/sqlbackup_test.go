package sqlbackup

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunWithOptionsPostgresCreatesCustomArchive(t *testing.T) {
	argsFile, envFile := installBackupCommand(t, "pg_dump", `#!/bin/sh
printf '%s\n' "$@" > "$GODB_SQLBACKUP_ARGS"
printf 'PGPASSWORD=%s\n' "$PGPASSWORD" > "$GODB_SQLBACKUP_ENV"
output=""
while [ "$#" -gt 0 ]; do
  if [ "$1" = "--file" ]; then
    output="$2"
    shift 2
    continue
  fi
  shift
done
printf 'postgres archive' > "$output"
`)

	outputPath := filepath.Join(t.TempDir(), "app.dump")
	dsn := "host=127.0.0.1 port=5432 user=backup password='secret value' dbname=app sslmode=disable"
	got, err := runWithOptions(context.Background(), runOptions{
		db:     " POSTGRES ",
		dsn:    dsn,
		output: outputPath,
	})
	if err != nil {
		t.Fatal(err)
	}
	if got != outputPath {
		t.Fatalf("output path = %q, want %q", got, outputPath)
	}

	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "postgres archive" {
		t.Fatalf("unexpected archive content: %q", content)
	}
	info, err := os.Stat(outputPath)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0600 {
		t.Fatalf("output permissions = %o, want 0600", info.Mode().Perm())
	}

	args := readTestFile(t, argsFile)
	for _, argument := range []string{"--format=custom", "--no-tablespaces", "--no-password", "--dbname", "--file"} {
		if !strings.Contains(args, argument) {
			t.Fatalf("expected %q in pg_dump args:\n%s", argument, args)
		}
	}
	if strings.Contains(args, "secret value") {
		t.Fatalf("pg_dump arguments leaked password:\n%s", args)
	}
	if !strings.Contains(readTestFile(t, envFile), "PGPASSWORD=secret value") {
		t.Fatalf("pg_dump did not receive password through its environment")
	}
	if matches, err := filepath.Glob(filepath.Join(filepath.Dir(outputPath), ".app.dump.tmp-*")); err != nil || len(matches) != 0 {
		t.Fatalf("temporary files = %v, err = %v", matches, err)
	}
}

func TestRunWithOptionsMySQLCreatesSQLBackup(t *testing.T) {
	argsFile, envFile := installBackupCommand(t, "mysqldump", `#!/bin/sh
printf '%s\n' "$@" > "$GODB_SQLBACKUP_ARGS"
printf 'MYSQL_PWD=%s\n' "$MYSQL_PWD" > "$GODB_SQLBACKUP_ENV"
printf '%s\n' 'CREATE TABLE users (id bigint);'
printf '%s\n' 'INSERT INTO users VALUES (1);'
`)

	outputPath := filepath.Join(t.TempDir(), "app.sql")
	got, err := runWithOptions(context.Background(), runOptions{
		db:     "mysql",
		dsn:    "backup:secret@tcp(127.0.0.1:3307)/app?parseTime=true",
		output: outputPath,
	})
	if err != nil {
		t.Fatal(err)
	}
	if got != outputPath {
		t.Fatalf("output path = %q, want %q", got, outputPath)
	}
	content := readTestFile(t, outputPath)
	if !strings.Contains(content, "CREATE TABLE users") || !strings.Contains(content, "INSERT INTO users") {
		t.Fatalf("unexpected mysql backup content:\n%s", content)
	}

	args := readTestFile(t, argsFile)
	for _, argument := range []string{
		"--user=backup",
		"--protocol=TCP",
		"--host=127.0.0.1",
		"--port=3307",
		"--single-transaction",
		"--quick",
		"--routines",
		"--events",
		"--triggers",
		"--no-tablespaces",
		"--set-gtid-purged=OFF",
		"app",
	} {
		if !strings.Contains(args, argument) {
			t.Fatalf("expected %q in mysqldump args:\n%s", argument, args)
		}
	}
	if strings.Contains(args, "secret") {
		t.Fatalf("mysqldump arguments leaked password:\n%s", args)
	}
	if !strings.Contains(readTestFile(t, envFile), "MYSQL_PWD=secret") {
		t.Fatalf("mysqldump did not receive password through its environment")
	}
}

func TestRunWithOptionsKeepsExistingOutputWhenBackupFails(t *testing.T) {
	installBackupCommand(t, "pg_dump", `#!/bin/sh
printf '%s\n' 'failed for secret' >&2
exit 7
`)

	outputPath := filepath.Join(t.TempDir(), "app.dump")
	if err := os.WriteFile(outputPath, []byte("previous backup"), 0600); err != nil {
		t.Fatal(err)
	}

	_, err := runWithOptions(context.Background(), runOptions{
		db:     "postgres",
		dsn:    "host=127.0.0.1 port=5432 user=backup password=secret dbname=app sslmode=disable",
		output: outputPath,
		force:  true,
	})
	if err == nil {
		t.Fatal("expected pg_dump error")
	}
	if strings.Contains(err.Error(), "secret") {
		t.Fatalf("backup error leaked password: %v", err)
	}
	if got := readTestFile(t, outputPath); got != "previous backup" {
		t.Fatalf("existing output = %q, want previous backup", got)
	}
	if matches, globErr := filepath.Glob(filepath.Join(filepath.Dir(outputPath), ".app.dump.tmp-*")); globErr != nil || len(matches) != 0 {
		t.Fatalf("temporary files = %v, err = %v", matches, globErr)
	}
}

func TestRunWithOptionsForceReplacesExistingOutput(t *testing.T) {
	installBackupCommand(t, "pg_dump", `#!/bin/sh
output=""
while [ "$#" -gt 0 ]; do
  if [ "$1" = "--file" ]; then
    output="$2"
    shift 2
    continue
  fi
  shift
done
printf 'new backup' > "$output"
`)

	outputPath := filepath.Join(t.TempDir(), "app.dump")
	if err := os.WriteFile(outputPath, []byte("previous backup"), 0600); err != nil {
		t.Fatal(err)
	}

	_, err := runWithOptions(context.Background(), runOptions{
		db:     "postgres",
		dsn:    "host=127.0.0.1 port=5432 user=backup password=secret dbname=app sslmode=disable",
		output: outputPath,
		force:  true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if got := readTestFile(t, outputPath); got != "new backup" {
		t.Fatalf("output = %q, want new backup", got)
	}
}

func TestRunWithOptionsReturnsValidationErrorsBeforeRunningCommands(t *testing.T) {
	tests := []struct {
		name string
		opts runOptions
		want string
	}{
		{name: "missing db", opts: runOptions{dsn: "dsn", output: "out"}, want: "db cannot be empty"},
		{name: "unknown db", opts: runOptions{db: "sqlite", dsn: "dsn", output: "out"}, want: "unknown database type"},
		{name: "missing dsn", opts: runOptions{db: "postgres", output: "out"}, want: "dsn cannot be empty"},
		{name: "missing output", opts: runOptions{db: "postgres", dsn: "dsn"}, want: "output path cannot be empty"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := runWithOptions(context.Background(), tt.opts)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("error = %v, want %q", err, tt.want)
			}
		})
	}
}

func TestRunWithOptionsRejectsExistingOutputWithoutForce(t *testing.T) {
	installBackupCommand(t, "mysqldump", "#!/bin/sh\nprintf 'unexpected backup'\n")

	outputPath := filepath.Join(t.TempDir(), "app.sql")
	if err := os.WriteFile(outputPath, []byte("previous backup"), 0600); err != nil {
		t.Fatal(err)
	}

	_, err := runWithOptions(context.Background(), runOptions{
		db:     "mysql",
		dsn:    "backup:secret@tcp(127.0.0.1:3306)/app",
		output: outputPath,
	})
	if err == nil || !strings.Contains(err.Error(), "output file already exists") {
		t.Fatalf("expected output exists error, got %v", err)
	}
	if got := readTestFile(t, outputPath); got != "previous backup" {
		t.Fatalf("existing output = %q, want previous backup", got)
	}
}

func TestRunWithOptionsHonorsCanceledContext(t *testing.T) {
	installBackupCommand(t, "mysqldump", `#!/bin/sh
sleep 5
printf '%s\n' 'unreachable'
`)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	outputPath := filepath.Join(t.TempDir(), "app.sql")
	_, err := runWithOptions(ctx, runOptions{
		db:     "mysql",
		dsn:    "backup:secret@tcp(127.0.0.1:3306)/app",
		output: outputPath,
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("error = %v, want context canceled", err)
	}
	if _, statErr := os.Stat(outputPath); !errors.Is(statErr, os.ErrNotExist) {
		t.Fatalf("output file should not exist, stat error = %v", statErr)
	}
}

func installBackupCommand(t *testing.T, name, script string) (string, string) {
	t.Helper()
	binDir := t.TempDir()
	commandPath := filepath.Join(binDir, name)
	if err := os.WriteFile(commandPath, []byte(script), 0755); err != nil {
		t.Fatal(err)
	}
	argsFile := filepath.Join(t.TempDir(), "args")
	envFile := filepath.Join(t.TempDir(), "env")
	t.Setenv("GODB_SQLBACKUP_ARGS", argsFile)
	t.Setenv("GODB_SQLBACKUP_ENV", envFile)
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
	return argsFile, envFile
}

func readTestFile(t *testing.T, path string) string {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(content)
}
