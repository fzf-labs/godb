package sqldump

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/fzf-labs/godb/orm/utils/fileutil"
	"github.com/jackc/pgx/v5/pgconn"
)

// DumpPostgres 导出创建语句
func (s *SQLDump) DumpPostgres() error {
	// 查找命令的可执行文件
	_, err := exec.LookPath("pg_dump")
	if err != nil {
		return fmt.Errorf("command pg_dump not found, please install: %w", err)
	}
	dbClient, err := newSimpleGormClient(s.db, s.dsn)
	if err != nil {
		return err
	}
	var tables []string
	if s.targetTables != "" {
		tables = strings.Split(s.targetTables, ",")
	} else {
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
		cmdArgs := buildPgDumpArgs(dsnParse, v)
		// 创建一个 Cmd 对象来表示将要执行的命令
		cmd := exec.Command("pg_dump", cmdArgs...)
		// 添加一个环境变量到命令中
		cmd.Env = append(os.Environ(), fmt.Sprintf("PGPASSWORD=%s", dsnParse.Password))
		// 执行命令，并捕获输出和错误信息
		output, err := cmd.Output()
		if err != nil {
			return fmt.Errorf("pg_dump table %s: %w", v, err)
		}
		if !s.fileOverwrite {
			if fileutil.Exists(outFile) {
				continue
			}
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
		"-t", table,
	}
}

// 预编译正则表达式，避免重复编译
var alterOwnerRegex = regexp.MustCompile(`ALTER TABLE .*? OWNER TO postgres`)

// remove 移除多余行
func (s *SQLDump) postgresRemove(str string) string {
	if str == "" {
		return ""
	}
	var result strings.Builder
	// 预估结果大小，减少内存重分配
	result.Grow(len(str))
	reader := strings.NewReader(str)
	scanner := bufio.NewScanner(reader)
	var currentStatement strings.Builder
	var inAlterStatement bool
	for scanner.Scan() {
		line := scanner.Text()
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
			result.WriteString(currentStatement.String())
			result.WriteByte('\n')
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
