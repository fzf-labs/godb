package sqlbackup

import (
	"strings"
	"testing"
)

func TestParseMySQLConnectionSupportsTCPAndUnixSockets(t *testing.T) {
	tests := []struct {
		name       string
		dsn        string
		wantSubnet string
	}{
		{name: "tcp", dsn: "backup:secret@tcp(127.0.0.1:3307)/app", wantSubnet: "--host=127.0.0.1"},
		{name: "unix", dsn: "backup:secret@unix(/tmp/mysql.sock)/app", wantSubnet: "--socket=/tmp/mysql.sock"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			connection, err := parseMySQLConnection(tt.dsn)
			if err != nil {
				t.Fatal(err)
			}
			args := strings.Join(buildMySQLDumpArgs(connection), " ")
			if !strings.Contains(args, tt.wantSubnet) {
				t.Fatalf("expected %q in args: %s", tt.wantSubnet, args)
			}
		})
	}
}

func TestParseMySQLConnectionRejectsUnsupportedConfigurations(t *testing.T) {
	tests := []struct {
		name string
		dsn  string
		want string
	}{
		{name: "empty user", dsn: "@tcp(127.0.0.1:3306)/app", want: "user cannot be empty"},
		{name: "empty database", dsn: "backup:secret@tcp(127.0.0.1:3306)/", want: "database cannot be empty"},
		{name: "custom tls", dsn: "backup:secret@tcp(127.0.0.1:3306)/app?tls=custom", want: "unknown config name"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseMySQLConnection(tt.dsn)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("error = %v, want %q", err, tt.want)
			}
		})
	}
}

func TestParseMySQLConnectionMapsStandardTLSModes(t *testing.T) {
	tests := []struct {
		dsn  string
		want string
	}{
		{dsn: "backup:secret@tcp(127.0.0.1:3306)/app?tls=false", want: "--ssl-mode=DISABLED"},
		{dsn: "backup:secret@tcp(127.0.0.1:3306)/app?tls=true", want: "--ssl-mode=REQUIRED"},
		{dsn: "backup:secret@tcp(127.0.0.1:3306)/app?tls=skip-verify", want: "--ssl-mode=REQUIRED"},
		{dsn: "backup:secret@tcp(127.0.0.1:3306)/app?tls=preferred", want: "--ssl-mode=PREFERRED"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			connection, err := parseMySQLConnection(tt.dsn)
			if err != nil {
				t.Fatal(err)
			}
			if args := strings.Join(buildMySQLDumpArgs(connection), " "); !strings.Contains(args, tt.want) {
				t.Fatalf("expected %q in args: %s", tt.want, args)
			}
		})
	}
}

func TestBuildMySQLDumpArgsSkipsPasswordPromptWithoutPassword(t *testing.T) {
	connection := &mysqlConnection{
		user:     "backup",
		network:  "tcp",
		host:     "127.0.0.1",
		port:     "3306",
		database: "app",
	}
	args := strings.Join(buildMySQLDumpArgs(connection), " ")
	if !strings.Contains(args, "--skip-password") {
		t.Fatalf("expected --skip-password in args: %s", args)
	}
}
