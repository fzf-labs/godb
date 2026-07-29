package sqlbackup

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestRunWritesFinalOutputPath(t *testing.T) {
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
printf 'archive' > "$output"
`)

	oldDB, oldDSN, oldOutput, oldForce := db, dsn, output, force
	defer func() {
		db, dsn, output, force = oldDB, oldDSN, oldOutput, oldForce
		CmdSQLBackup.SetOut(nil)
		CmdSQLBackup.SetContext(nil)
	}()

	outputPath := filepath.Join(t.TempDir(), "app.dump")
	db = "postgres"
	dsn = "host=127.0.0.1 port=5432 user=backup password=secret dbname=app sslmode=disable"
	output = outputPath
	force = false

	var stdout bytes.Buffer
	CmdSQLBackup.SetOut(&stdout)
	CmdSQLBackup.SetContext(context.Background())
	if err := Run(CmdSQLBackup, nil); err != nil {
		t.Fatal(err)
	}
	if stdout.String() != outputPath+"\n" {
		t.Fatalf("stdout = %q, want %q", stdout.String(), outputPath+"\n")
	}
	if _, err := os.Stat(outputPath); err != nil {
		t.Fatal(err)
	}
}
