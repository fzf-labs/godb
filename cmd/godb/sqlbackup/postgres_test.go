package sqlbackup

import (
	"strings"
	"testing"
)

func TestParsePostgresConnectionSanitizesKeywordAndURLPasswords(t *testing.T) {
	tests := []struct {
		name string
		dsn  string
	}{
		{
			name: "keyword",
			dsn:  "host=127.0.0.1 port=5432 user=backup password='secret value' dbname=app sslmode=disable",
		},
		{
			name: "url",
			dsn:  "postgres://backup:secret%20value@127.0.0.1:5432/app?sslmode=disable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			connection, err := parsePostgresConnection(tt.dsn)
			if err != nil {
				t.Fatal(err)
			}
			if connection.password != "secret value" {
				t.Fatalf("password = %q, want secret value", connection.password)
			}
			if strings.Contains(connection.dsn, "secret") {
				t.Fatalf("sanitized dsn leaked password: %s", connection.dsn)
			}
			args := buildPgDumpArgs(connection, "/tmp/app.dump")
			if strings.Contains(strings.Join(args, " "), "secret") {
				t.Fatalf("pg_dump args leaked password: %v", args)
			}
		})
	}
}

func TestSanitizePostgresURLRemovesCaseInsensitivePasswordQuery(t *testing.T) {
	sanitized, _, err := sanitizePostgresURL("postgres://backup@127.0.0.1:5432/app?PASSWORD=secret&sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(sanitized, "secret") || strings.Contains(strings.ToLower(sanitized), "password") {
		t.Fatalf("sanitized dsn leaked password query parameter: %s", sanitized)
	}
}

func TestParsePostgresConnectionRejectsUnsupportedSettings(t *testing.T) {
	tests := []struct {
		name string
		dsn  string
		want string
	}{
		{name: "multiple hosts", dsn: "host=primary,standby user=backup dbname=app sslmode=disable", want: "one host"},
		{name: "service", dsn: "service=app user=backup dbname=app", want: "service DSNs"},
		{name: "ssl password", dsn: "host=127.0.0.1 user=backup dbname=app sslmode=disable sslpassword=secret", want: "sslpassword"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parsePostgresConnection(tt.dsn)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("error = %v, want %q", err, tt.want)
			}
		})
	}
}

func TestParsePostgresKeywordOptionsHandlesEscapes(t *testing.T) {
	options, err := parsePostgresKeywordOptions("user=backup password='p\\\\a\\'ss' dbname='app name'")
	if err != nil {
		t.Fatal(err)
	}
	if len(options) != 3 {
		t.Fatalf("option count = %d, want 3", len(options))
	}
	if options[1].value != "p\\a'ss" {
		t.Fatalf("password = %q", options[1].value)
	}
	formatted := formatPostgresKeywordOptions(options[:1])
	if formatted != "user='backup'" {
		t.Fatalf("formatted options = %q", formatted)
	}
}
