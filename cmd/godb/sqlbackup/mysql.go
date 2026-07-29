package sqlbackup

import (
	"context"
	"fmt"
	"net"
	"os"
	"strings"

	mysqlDriver "github.com/go-sql-driver/mysql"
)

// mysqlConnection 保存解析后的 MySQL 连接信息，用于组装 mysqldump 参数。
type mysqlConnection struct {
	user     string
	password string
	network  string // tcp / unix
	host     string
	port     string
	socket   string
	sslMode  string // mysqldump --ssl-mode 取值
	database string
}

// runMySQL 调用 mysqldump 创建包含 schema 与数据的逻辑备份。
func (o runOptions) runMySQL(ctx context.Context) error {
	connection, err := parseMySQLConnection(o.dsn)
	if err != nil {
		return err
	}
	executable, err := lookupExecutable("mysqldump")
	if err != nil {
		return fmt.Errorf("command mysqldump not found, please install it: %w", err)
	}

	output, err := prepareAtomicOutput(o.output, o.force)
	if err != nil {
		return err
	}
	defer output.abort()

	stderr, err := runExternalCommand(
		ctx,
		executable,
		buildMySQLDumpArgs(connection),
		commandEnvironment(os.Environ(), "MYSQL_PWD", connection.password),
		output.file,
	)
	if err != nil {
		return formatCommandError(ctx, "mysqldump", stderr, err, o.dsn, connection.password)
	}
	if err := output.commit(); err != nil {
		return err
	}
	return nil
}

// parseMySQLConnection 解析 go-sql-driver 风格的 MySQL DSN。
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

// mysqlSSLMode 将 go-sql-driver 的 TLSConfig 映射为 mysqldump --ssl-mode。
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

// buildMySQLDumpArgs 组装 mysqldump 命令行参数。
// 密码通过 MYSQL_PWD 环境变量传递，避免出现在进程参数中。
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
