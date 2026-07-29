//go:build integration

package sqlbackup

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestSQLBackupPostgresIntegration(t *testing.T) {
	const (
		sourceDatabase = "godb_sqlbackup_source"
		targetDatabase = "godb_sqlbackup_target"
	)

	pg := postgresIntegrationConfig()
	dropPostgresIntegrationDatabase(t, pg, sourceDatabase)
	dropPostgresIntegrationDatabase(t, pg, targetDatabase)
	t.Cleanup(func() {
		dropPostgresIntegrationDatabase(t, pg, sourceDatabase)
		dropPostgresIntegrationDatabase(t, pg, targetDatabase)
	})
	createPostgresIntegrationDatabase(t, pg, sourceDatabase)
	runPostgresIntegrationCommand(t, pg, sourceDatabase, []string{
		"-c", "CREATE TABLE backup_records (id bigint PRIMARY KEY, name text NOT NULL);",
		"-c", "INSERT INTO backup_records (id, name) VALUES (1, 'row-one');",
	})

	outputPath := filepath.Join(t.TempDir(), "postgres.dump")
	_, err := runWithOptions(context.Background(), runOptions{
		db:     databasePostgres,
		dsn:    postgresIntegrationDSN(pg, sourceDatabase),
		output: outputPath,
	})
	if err != nil {
		t.Fatal(err)
	}

	archiveContents := runPostgresIntegrationTool(t, pg, []string{"pg_restore", "--list", outputPath})
	if !strings.Contains(archiveContents, "backup_records") {
		t.Fatalf("pg_restore list did not include backup_records:\n%s", archiveContents)
	}

	createPostgresIntegrationDatabase(t, pg, targetDatabase)
	runPostgresIntegrationTool(t, pg, []string{
		"pg_restore",
		"--exit-on-error",
		"--no-owner",
		"--no-privileges",
		"--no-tablespaces",
		"--host", pg.host,
		"--port", pg.port,
		"--username", pg.user,
		"--dbname", targetDatabase,
		outputPath,
	})
	result := runPostgresIntegrationCommand(t, pg, targetDatabase, []string{
		"-At",
		"-c", "SELECT name FROM backup_records WHERE id = 1;",
	})
	if strings.TrimSpace(result) != "row-one" {
		t.Fatalf("restored row = %q, want row-one", result)
	}
}

func TestSQLBackupMySQLIntegration(t *testing.T) {
	const (
		sourceDatabase = "godb_sqlbackup_source"
		targetDatabase = "godb_sqlbackup_target"
	)

	installMySQLDockerClients(t)
	mysql := mysqlIntegrationConfig()
	dropMySQLIntegrationDatabase(t, mysql, sourceDatabase)
	dropMySQLIntegrationDatabase(t, mysql, targetDatabase)
	t.Cleanup(func() {
		dropMySQLIntegrationDatabase(t, mysql, sourceDatabase)
		dropMySQLIntegrationDatabase(t, mysql, targetDatabase)
	})
	createMySQLIntegrationDatabase(t, mysql, sourceDatabase)
	runMySQLIntegrationCommand(t, mysql, sourceDatabase, nil,
		"-e", "CREATE TABLE backup_records (id bigint PRIMARY KEY, name varchar(64) NOT NULL); INSERT INTO backup_records (id, name) VALUES (1, 'row-one');",
	)

	outputPath := filepath.Join(t.TempDir(), "mysql.sql")
	_, err := runWithOptions(context.Background(), runOptions{
		db:     databaseMySQL,
		dsn:    fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", mysql.user, mysql.password, mysql.host, mysql.port, sourceDatabase),
		output: outputPath,
	})
	if err != nil {
		t.Fatal(err)
	}
	if content := readTestFile(t, outputPath); !strings.Contains(content, "backup_records") {
		t.Fatalf("mysqldump output did not include backup_records:\n%s", content)
	}

	createMySQLIntegrationDatabase(t, mysql, targetDatabase)
	backupFile, err := os.Open(outputPath)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = backupFile.Close() })
	runMySQLIntegrationCommand(t, mysql, targetDatabase, backupFile)
	result := runMySQLIntegrationCommand(t, mysql, targetDatabase, nil,
		"--batch",
		"--skip-column-names",
		"-e", "SELECT name FROM backup_records WHERE id = 1;",
	)
	if strings.TrimSpace(result) != "row-one" {
		t.Fatalf("restored row = %q, want row-one", result)
	}
}

type postgresIntegrationSettings struct {
	host     string
	port     string
	user     string
	password string
}

func postgresIntegrationConfig() postgresIntegrationSettings {
	return postgresIntegrationSettings{
		host:     envOrDefault("PGHOST", "127.0.0.1"),
		port:     envOrDefault("PGPORT", "5432"),
		user:     envOrDefault("PGUSER", "postgres"),
		password: envOrDefault("PGPASSWORD", "123456"),
	}
}

func postgresIntegrationDSN(settings postgresIntegrationSettings, database string) string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		quotePostgresDSNValue(settings.host),
		quotePostgresDSNValue(settings.port),
		quotePostgresDSNValue(settings.user),
		quotePostgresDSNValue(settings.password),
		quotePostgresDSNValue(database),
	)
}

func quotePostgresDSNValue(value string) string {
	value = strings.ReplaceAll(value, "\\", "\\\\")
	value = strings.ReplaceAll(value, "'", "\\'")
	return "'" + value + "'"
}

func createPostgresIntegrationDatabase(t *testing.T, settings postgresIntegrationSettings, database string) {
	t.Helper()
	runPostgresIntegrationTool(t, settings, []string{
		"createdb",
		"--host", settings.host,
		"--port", settings.port,
		"--username", settings.user,
		database,
	})
}

func dropPostgresIntegrationDatabase(t *testing.T, settings postgresIntegrationSettings, database string) {
	t.Helper()
	cmd := exec.Command("dropdb", "--if-exists", "--host", settings.host, "--port", settings.port, "--username", settings.user, database)
	cmd.Env = commandEnvironment(os.Environ(), "PGPASSWORD", settings.password)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Logf("drop postgres integration database %s: %v: %s", database, err, output)
	}
}

func runPostgresIntegrationCommand(t *testing.T, settings postgresIntegrationSettings, database string, args []string) string {
	t.Helper()
	command := append([]string{
		"psql",
		"--host", settings.host,
		"--port", settings.port,
		"--username", settings.user,
		"--dbname", database,
		"--set", "ON_ERROR_STOP=1",
	}, args...)
	return runPostgresIntegrationTool(t, settings, command)
}

func runPostgresIntegrationTool(t *testing.T, settings postgresIntegrationSettings, command []string) string {
	t.Helper()
	cmd := exec.Command(command[0], command[1:]...)
	cmd.Env = commandEnvironment(os.Environ(), "PGPASSWORD", settings.password)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("%s failed: %v: %s", command[0], err, output)
	}
	return string(output)
}

type mysqlIntegrationSettings struct {
	host     string
	port     string
	user     string
	password string
}

func mysqlIntegrationConfig() mysqlIntegrationSettings {
	return mysqlIntegrationSettings{
		host:     envOrDefault("MYSQL_HOST", "127.0.0.1"),
		port:     envOrDefault("MYSQL_PORT", "3306"),
		user:     envOrDefault("MYSQL_USER", "root"),
		password: envOrDefault("MYSQL_PASSWORD", "123456"),
	}
}

func createMySQLIntegrationDatabase(t *testing.T, settings mysqlIntegrationSettings, database string) {
	t.Helper()
	runMySQLIntegrationCommand(t, settings, "", nil, "-e", "CREATE DATABASE `"+database+"`;")
}

func dropMySQLIntegrationDatabase(t *testing.T, settings mysqlIntegrationSettings, database string) {
	t.Helper()
	cmd := mysqlIntegrationCommand(settings, "", nil, "-e", "DROP DATABASE IF EXISTS `"+database+"`;")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Logf("drop mysql integration database %s: %v: %s", database, err, output)
	}
}

func runMySQLIntegrationCommand(t *testing.T, settings mysqlIntegrationSettings, database string, stdin *os.File, args ...string) string {
	t.Helper()
	cmd := mysqlIntegrationCommand(settings, database, stdin, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("mysql failed: %v: %s", err, output)
	}
	return string(output)
}

func mysqlIntegrationCommand(settings mysqlIntegrationSettings, database string, stdin *os.File, args ...string) *exec.Cmd {
	command := []string{
		"--user=" + settings.user,
		"--protocol=TCP",
		"--host=" + settings.host,
		"--port=" + settings.port,
	}
	command = append(command, args...)
	if database != "" {
		command = append(command, database)
	}
	cmd := exec.Command("mysql", command...)
	cmd.Env = commandEnvironment(os.Environ(), "MYSQL_PWD", settings.password)
	cmd.Stdin = stdin
	return cmd
}

func installMySQLDockerClients(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("docker"); err != nil {
		t.Fatalf("docker is required for MySQL integration tests: %v", err)
	}
	binDir := t.TempDir()
	for name, script := range map[string]string{
		"mysqldump": "#!/bin/sh\nexec docker run --rm --network host -e MYSQL_PWD mysql:8.4 mysqldump \"$@\"\n",
		"mysql":     "#!/bin/sh\nexec docker run --rm --network host -i -e MYSQL_PWD mysql:8.4 mysql \"$@\"\n",
	} {
		path := filepath.Join(binDir, name)
		if err := os.WriteFile(path, []byte(script), 0755); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
}

func envOrDefault(name, defaultValue string) string {
	if value := strings.TrimSpace(os.Getenv(name)); value != "" {
		return value
	}
	return defaultValue
}
