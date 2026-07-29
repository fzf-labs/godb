package sqlbackup

import (
	"context"
	"fmt"
	"net"
	"os"
	"strings"

	mysqlDriver "github.com/go-sql-driver/mysql"
)

type mysqlConnection struct {
	user     string
	password string
	network  string
	host     string
	port     string
	socket   string
	sslMode  string
	database string
}

func (b *sqlBackup) runMySQL(ctx context.Context) error {
	connection, err := parseMySQLConnection(b.dsn)
	if err != nil {
		return err
	}
	executable, err := lookupExecutable("mysqldump")
	if err != nil {
		return fmt.Errorf("command mysqldump not found, please install it: %w", err)
	}

	output, err := prepareAtomicOutput(b.output, b.force)
	if err != nil {
		return err
	}
	defer output.abort()
	writer, err := output.writer()
	if err != nil {
		return err
	}

	stderr, err := runExternalCommand(
		ctx,
		executable,
		buildMySQLDumpArgs(connection),
		commandEnvironment(os.Environ(), "MYSQL_PWD", connection.password),
		writer,
	)
	if err != nil {
		return formatCommandError(ctx, "mysqldump", stderr, err, b.dsn, connection.password)
	}
	if err := output.commit(); err != nil {
		return err
	}
	return nil
}

func parseMySQLConnection(dsn string) (*mysqlConnection, error) {
	config, err := mysqlDriver.ParseDSN(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse mysql dsn: %w", err)
	}
	if strings.TrimSpace(config.User) == "" {
		return nil, fmt.Errorf("mysql dsn user cannot be empty")
	}
	if strings.TrimSpace(config.DBName) == "" {
		return nil, fmt.Errorf("mysql dsn database cannot be empty")
	}
	sslMode, err := mysqlSSLMode(config.TLSConfig)
	if err != nil {
		return nil, err
	}

	connection := &mysqlConnection{
		user:     config.User,
		password: config.Passwd,
		network:  config.Net,
		sslMode:  sslMode,
		database: config.DBName,
	}
	switch config.Net {
	case "tcp", "tcp4", "tcp6":
		host, port, err := net.SplitHostPort(config.Addr)
		if err != nil {
			return nil, fmt.Errorf("parse mysql tcp address: %w", err)
		}
		connection.host = host
		connection.port = port
	case "unix":
		if strings.TrimSpace(config.Addr) == "" {
			return nil, fmt.Errorf("mysql unix socket cannot be empty")
		}
		connection.socket = config.Addr
	default:
		return nil, fmt.Errorf("mysql network %q is not supported for backup", config.Net)
	}
	return connection, nil
}

func mysqlSSLMode(tlsConfig string) (string, error) {
	switch tlsConfig {
	case "":
		return "", nil
	case "false":
		return "DISABLED", nil
	case "true", "skip-verify":
		return "REQUIRED", nil
	case "preferred":
		return "PREFERRED", nil
	default:
		return "", fmt.Errorf("mysql custom TLS configuration %q is not supported for backup", tlsConfig)
	}
}

func buildMySQLDumpArgs(connection *mysqlConnection) []string {
	args := []string{
		"--user=" + connection.user,
	}
	if connection.password == "" {
		args = append(args, "--skip-password")
	}
	switch connection.network {
	case "unix":
		args = append(args, "--protocol=SOCKET", "--socket="+connection.socket)
	default:
		args = append(args, "--protocol=TCP", "--host="+connection.host, "--port="+connection.port)
	}
	if connection.sslMode != "" {
		args = append(args, "--ssl-mode="+connection.sslMode)
	}
	args = append(args,
		"--single-transaction",
		"--quick",
		"--routines",
		"--events",
		"--triggers",
		"--no-tablespaces",
		"--set-gtid-purged=OFF",
		"--",
		connection.database,
	)
	return args
}
