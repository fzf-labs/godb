package sqlbackup

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
)

// postgresConnection 保存传给 pg_dump 的连接信息。
// dsn 中密码已剥离，实际密码通过 PGPASSWORD 环境变量传递。
type postgresConnection struct {
	dsn      string // 已脱敏的连接串（无 password）
	password string
}

// postgresOption 表示 keyword/value 形式 DSN 中的一对键值。
type postgresOption struct {
	key   string
	value string
}

// runPostgres 调用 pg_dump 创建 custom 格式的逻辑备份。
func (o runOptions) runPostgres(ctx context.Context) error {
	connection, err := parsePostgresConnection(o.dsn)
	if err != nil {
		return err
	}
	executable, err := lookupExecutable("pg_dump")
	if err != nil {
		return fmt.Errorf("command pg_dump not found, please install it: %w", err)
	}

	output, err := prepareAtomicOutput(o.output, o.force)
	if err != nil {
		return err
	}
	defer output.abort()
	// pg_dump 直接写入 --file，需先关闭临时文件句柄再交给外部进程。
	if err := output.close(); err != nil {
		return fmt.Errorf("close temporary backup file: %w", err)
	}

	stderr, err := runExternalCommand(
		ctx,
		executable,
		buildPgDumpArgs(connection, output.tempPath),
		commandEnvironment(os.Environ(), "PGPASSWORD", connection.password),
		nil,
	)
	if err != nil {
		return formatCommandError(ctx, "pg_dump", stderr, err, o.dsn, connection.password)
	}
	if err := output.commit(); err != nil {
		return err
	}
	return nil
}

// parsePostgresConnection 解析 PostgreSQL DSN，并剥离密码供环境变量传递。
func parsePostgresConnection(dsn string) (*postgresConnection, error) {
	sanitizedDSN, settings, err := sanitizePostgresDSN(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse postgres dsn: %w", err)
	}
	if err := validatePostgresBackupSettings(settings); err != nil {
		return nil, err
	}

	config, err := pgconn.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse postgres dsn: %w", err)
	}
	if strings.TrimSpace(config.Database) == "" {
		return nil, fmt.Errorf("postgres dsn database cannot be empty")
	}

	return &postgresConnection{
		dsn:      sanitizedDSN,
		password: config.Password,
	}, nil
}

// buildPgDumpArgs 组装 pg_dump 命令行参数，输出 custom 格式备份文件。
func buildPgDumpArgs(connection *postgresConnection, outputPath string) []string {
	return []string{
		"--dbname", connection.dsn,
		"--format=custom",
		"--no-tablespaces",
		"--no-password",
		"--file", outputPath,
	}
}

// sanitizePostgresDSN 从 URL 或 keyword/value DSN 中移除 password，并提取连接设置。
func sanitizePostgresDSN(dsn string) (string, map[string]string, error) {
	trimmedDSN := strings.TrimSpace(dsn)
	if strings.HasPrefix(strings.ToLower(trimmedDSN), "postgres://") || strings.HasPrefix(strings.ToLower(trimmedDSN), "postgresql://") {
		return sanitizePostgresURL(trimmedDSN)
	}
	return sanitizePostgresKeywordDSN(trimmedDSN)
}

// sanitizePostgresURL 处理 postgres:// / postgresql:// 形式的 DSN。
func sanitizePostgresURL(dsn string) (string, map[string]string, error) {
	parsed, err := url.Parse(dsn)
	if err != nil {
		return "", nil, err
	}

	settings := make(map[string]string)
	if parsed.User != nil {
		settings["user"] = parsed.User.Username()
		if password, ok := parsed.User.Password(); ok {
			settings["password"] = password
		}
		parsed.User = url.User(parsed.User.Username())
	}
	if parsed.Host != "" {
		settings["host"] = parsed.Host
	}
	if database := strings.TrimPrefix(parsed.EscapedPath(), "/"); database != "" {
		settings["database"] = database
	}

	query := parsed.Query()
	for key, values := range query {
		if len(values) > 0 {
			settings[strings.ToLower(key)] = values[len(values)-1]
		}
	}
	for key := range query {
		if strings.EqualFold(key, "password") {
			query.Del(key)
		}
	}
	parsed.RawQuery = query.Encode()

	return parsed.String(), settings, nil
}

// sanitizePostgresKeywordDSN 处理 host=... dbname=... 形式的 keyword/value DSN。
func sanitizePostgresKeywordDSN(dsn string) (string, map[string]string, error) {
	options, err := parsePostgresKeywordOptions(dsn)
	if err != nil {
		return "", nil, err
	}
	settings := make(map[string]string, len(options))
	filtered := make([]postgresOption, 0, len(options))
	for _, option := range options {
		key := strings.ToLower(option.key)
		settings[key] = option.value
		if key != "password" {
			filtered = append(filtered, option)
		}
	}
	return formatPostgresKeywordOptions(filtered), settings, nil
}

// validatePostgresBackupSettings 拒绝备份场景不支持的连接配置。
func validatePostgresBackupSettings(settings map[string]string) error {
	host := settings["host"]
	if strings.Contains(host, ",") || strings.Contains(settings["port"], ",") {
		return fmt.Errorf("postgres backup supports one host and port only")
	}
	if settings["service"] != "" || settings["servicefile"] != "" {
		return fmt.Errorf("postgres service DSNs are not supported for backup")
	}
	if settings["sslpassword"] != "" {
		return fmt.Errorf("postgres sslpassword is not supported for backup")
	}
	return nil
}

// parsePostgresKeywordOptions 解析 libpq keyword/value 连接串。
func parsePostgresKeywordOptions(dsn string) ([]postgresOption, error) {
	const whitespace = " \t\n\r\v\f"

	dsn = strings.TrimLeft(dsn, whitespace)
	options := make([]postgresOption, 0)
	for len(dsn) > 0 {
		equalsIndex := strings.IndexByte(dsn, '=')
		if equalsIndex < 0 {
			return nil, fmt.Errorf("invalid keyword/value dsn")
		}
		key := strings.Trim(dsn[:equalsIndex], whitespace)
		if key == "" {
			return nil, fmt.Errorf("invalid keyword/value dsn")
		}
		dsn = strings.TrimLeft(dsn[equalsIndex+1:], whitespace)

		value := ""
		if len(dsn) > 0 {
			if dsn[0] == '\'' {
				var err error
				value, dsn, err = readQuotedPostgresKeywordValue(dsn[1:])
				if err != nil {
					return nil, err
				}
			} else {
				var err error
				value, dsn, err = readUnquotedPostgresKeywordValue(dsn, whitespace)
				if err != nil {
					return nil, err
				}
			}
		}
		options = append(options, postgresOption{key: key, value: value})
		dsn = strings.TrimLeft(dsn, whitespace)
	}
	return options, nil
}

// readQuotedPostgresKeywordValue 读取单引号包裹的 keyword 值。
func readQuotedPostgresKeywordValue(value string) (string, string, error) {
	for index := 0; index < len(value); index++ {
		switch value[index] {
		case '\\':
			index++
			if index == len(value) {
				return "", "", fmt.Errorf("invalid backslash in postgres dsn")
			}
		case '\'':
			return unescapePostgresKeywordValue(value[:index]), value[index+1:], nil
		}
	}
	return "", "", fmt.Errorf("unterminated quoted value in postgres dsn")
}

// readUnquotedPostgresKeywordValue 读取未加引号的 keyword 值，直到空白分隔。
func readUnquotedPostgresKeywordValue(value, whitespace string) (string, string, error) {
	index := 0
	for index < len(value) {
		if strings.ContainsRune(whitespace, rune(value[index])) {
			break
		}
		if value[index] == '\\' {
			index++
			if index == len(value) {
				return "", "", fmt.Errorf("invalid backslash in postgres dsn")
			}
		}
		index++
	}
	return unescapePostgresKeywordValue(value[:index]), value[index:], nil
}

// unescapePostgresKeywordValue 还原 keyword 值中的转义字符。
func unescapePostgresKeywordValue(value string) string {
	value = strings.ReplaceAll(value, "\\\\", "\\")
	return strings.ReplaceAll(value, "\\'", "'")
}

// formatPostgresKeywordOptions 将选项重新格式化为 keyword/value DSN。
func formatPostgresKeywordOptions(options []postgresOption) string {
	parts := make([]string, 0, len(options))
	for _, option := range options {
		parts = append(parts, option.key+"='"+escapePostgresKeywordValue(option.value)+"'")
	}
	return strings.Join(parts, " ")
}

// escapePostgresKeywordValue 转义 keyword 值中的反斜杠与单引号。
func escapePostgresKeywordValue(value string) string {
	value = strings.ReplaceAll(value, "\\", "\\\\")
	return strings.ReplaceAll(value, "'", "\\'")
}
