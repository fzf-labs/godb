package sqldump

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgconn"

	"github.com/fzf-labs/godb/cmd/godb/internal/tablelist"
	"github.com/fzf-labs/godb/orm/utils/fileutil"
)

var (
	pgDumpTimeout     = 2 * time.Minute
	execPgDumpCommand = runPgDumpCommand
)

// DumpPostgres 导出创建语句
func (s *SQLDump) DumpPostgres() error {
	// 查找命令的可执行文件
	_, err := exec.LookPath("pg_dump")
	if err != nil {
		return fmt.Errorf("command pg_dump not found, please install: %w", err)
	}
	tables, err := tablelist.ParseCSV(s.targetTables)
	if err != nil {
		return err
	}
	dbClient, err := newSimpleGormClient(s.db, s.dsn)
	if err != nil {
		return err
	}
	defer closeGormDB(dbClient)
	if len(tables) == 0 {
		tables, err = dbClient.Migrator().GetTables()
		if err != nil {
			return err
		}
	}
	dsnParse, err := s.postgresDsnParse()
	if err != nil {
		return err
	}
	outPath := outputDir(s.outPutPath, dsnParse.Dbname)
	err = os.MkdirAll(outPath, os.ModePerm)
	if err != nil {
		return fmt.Errorf("create output path: %w", err)
	}
	for _, v := range tables {
		outFile := filepath.Join(outPath, fmt.Sprintf("%s.sql", v))
		if !s.fileOverwrite {
			if fileutil.Exists(outFile) {
				continue
			}
		}
		ctx, cancel := context.WithTimeout(context.Background(), pgDumpTimeout)
		output, stderr, err := execPgDumpCommand(ctx, dsnParse, v)
		cancel()
		if err != nil {
			return formatPgDumpError(v, stderr, err)
		}
		tableContent := s.postgresRemove(string(output))
		if tableContent != "" {
			err := fileutil.WriteContentCover(outFile, tableContent)
			if err != nil {
				return fmt.Errorf("write %s: %w", outFile, err)
			}
		}
	}
	return nil
}

func runPgDumpCommand(ctx context.Context, dsnParse *PostgresDsn, table string) ([]byte, []byte, error) {
	cmdArgs := buildPgDumpArgs(dsnParse, table)
	// 创建一个 Cmd 对象来表示将要执行的命令
	cmd := exec.CommandContext(ctx, "pg_dump", cmdArgs...)
	// 添加一个环境变量到命令中
	cmd.Env = append(os.Environ(), fmt.Sprintf("PGPASSWORD=%s", dsnParse.Password))
	output, err := cmd.Output()
	if err == nil {
		return output, nil, nil
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return output, exitErr.Stderr, err
	}
	return output, nil, err
}

func formatPgDumpError(table string, stderr []byte, err error) error {
	detail := strings.TrimSpace(string(stderr))
	if errors.Is(err, context.DeadlineExceeded) {
		if detail != "" {
			return fmt.Errorf("pg_dump table %s timed out: %w: %s", table, err, detail)
		}
		return fmt.Errorf("pg_dump table %s timed out: %w", table, err)
	}
	if detail != "" {
		return fmt.Errorf("pg_dump table %s: %w: %s", table, err, detail)
	}
	return fmt.Errorf("pg_dump table %s: %w", table, err)
}

// PostgresDsn 保存 pgconn 解析出的 PostgreSQL 连接参数。
type PostgresDsn struct {
	Host     string
	Port     int
	User     string
	Password string
	Dbname   string
}

// postgresDsnParse 数据库解析。
func (s *SQLDump) postgresDsnParse() (*PostgresDsn, error) {
	cfg, err := pgconn.ParseConfig(s.dsn)
	if err != nil {
		return nil, fmt.Errorf("parse postgres dsn: %w", err)
	}
	return &PostgresDsn{
		Host:     cfg.Host,
		Port:     int(cfg.Port),
		User:     cfg.User,
		Password: cfg.Password,
		Dbname:   cfg.Database,
	}, nil
}

func buildPgDumpArgs(dsnParse *PostgresDsn, table string) []string {
	return []string{
		"-h", dsnParse.Host,
		"-p", strconv.Itoa(dsnParse.Port),
		"-U", dsnParse.User,
		"-d", dsnParse.Dbname,
		"-s",
		"-t", quotePgDumpTablePattern(table),
	}
}

func quotePgDumpTablePattern(table string) string {
	parts := strings.Split(table, ".")
	for i, part := range parts {
		parts[i] = `"` + strings.ReplaceAll(part, `"`, `""`) + `"`
	}
	return strings.Join(parts, ".")
}

// 预编译正则表达式，避免重复编译
var alterOwnerRegex = regexp.MustCompile(`^ALTER TABLE .*?\s+OWNER TO\s+.+;?$`)

// remove 移除多余行
func (s *SQLDump) postgresRemove(str string) string {
	if str == "" {
		return ""
	}
	var result strings.Builder
	// 预估结果大小，减少内存重分配
	result.Grow(len(str))
	reader := bufio.NewReader(strings.NewReader(str))
	var currentStatement strings.Builder
	var inAlterStatement bool
	for {
		line, err := readDumpLine(reader)
		if err != nil {
			break
		}
		// 跳过需要过滤的行 - 优化条件判断顺序
		if s.shouldSkipLine(line) {
			continue
		}
		trimmedLine := strings.TrimSpace(line)
		// 处理 ALTER 语句
		if strings.HasPrefix(trimmedLine, "ALTER TABLE") {
			inAlterStatement = true
			currentStatement.Reset()
			currentStatement.WriteString(trimmedLine)
		} else if inAlterStatement {
			// 在 ALTER 语句中，继续拼接
			currentStatement.WriteByte(' ')
			currentStatement.WriteString(trimmedLine)
		}
		// 检测语句结束（以分号结尾）
		if inAlterStatement && strings.HasSuffix(trimmedLine, ";") {
			statement := currentStatement.String()
			if !alterOwnerRegex.MatchString(statement) {
				result.WriteString(statement)
				result.WriteByte('\n')
			}
			inAlterStatement = false
			currentStatement.Reset()
		} else if !inAlterStatement {
			// 不是 ALTER 语句，正常处理
			result.WriteString(line)
			result.WriteByte('\n')
		}
	}
	return result.String()
}

func readDumpLine(reader *bufio.Reader) (string, error) {
	line, err := reader.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return "", err
	}
	if err == io.EOF && line == "" {
		return "", err
	}
	return strings.TrimSuffix(line, "\n"), nil
}

// shouldSkipLine 判断是否应该跳过该行
func (s *SQLDump) shouldSkipLine(line string) bool {
	if line == "" {
		return true
	}
	// 优化：先检查最常见的情况
	if strings.HasPrefix(line, "--") {
		return true
	}
	// 其他前缀检查
	if strings.HasPrefix(line, "SELECT") ||
		strings.HasPrefix(line, "SET") {
		return true
	}
	// 包含特殊字符的检查
	if strings.Contains(line, "\\restrict") ||
		strings.Contains(line, "\\unrestrict") {
		return true
	}
	// 最后检查正则表达式（最昂贵的操作）
	return alterOwnerRegex.MatchString(line)
}
