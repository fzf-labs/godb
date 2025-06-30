package sqldump

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/fzf-labs/godb/orm/gormx"
	"github.com/fzf-labs/godb/orm/utils/fileutil"
)

// DumpPostgres 导出创建语句
func (s *SQLDump) DumpPostgres() {
	// 查找命令的可执行文件
	_, err := exec.LookPath("pg_dump")
	if err != nil {
		log.Println("command pg_dump not found,please install")
		return
	}
	dbClient := gormx.NewSimpleGormClient(s.db, s.dsn)
	var tables []string
	if s.targetTables != "" {
		tables = strings.Split(s.targetTables, ",")
	} else {
		tables, err = dbClient.Migrator().GetTables()
		if err != nil {
			return
		}
	}
	dsnParse := s.postgresDsnParse()
	outPath := filepath.Join(strings.Trim(s.outPutPath, "/"), dsnParse.Dbname)
	err = os.MkdirAll(outPath, os.ModePerm)
	if err != nil {
		log.Println("DumpPostgres create path err:", err)
		return
	}
	for _, v := range tables {
		outFile := filepath.Join(outPath, fmt.Sprintf("%s.sql", v))
		cmdArgs := []string{
			"-h", dsnParse.Host,
			"-p", strconv.Itoa(dsnParse.Port),
			"-U", dsnParse.User,
			"-s", dsnParse.Dbname,
			"-t", v,
		}
		// 创建一个 Cmd 对象来表示将要执行的命令
		cmd := exec.Command("pg_dump", cmdArgs...)
		// 添加一个环境变量到命令中
		cmd.Env = append(cmd.Env, fmt.Sprintf("PGPASSWORD=%s", dsnParse.Password))
		// 执行命令，并捕获输出和错误信息
		output, err := cmd.Output()
		if err != nil {
			log.Println("cmd exec err:", err)
			return
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
				log.Println("DumpPostgres err:", err)
				return
			}
		}
	}
}

type PostgresDsn struct {
	Host     string
	Port     int
	User     string
	Password string
	Dbname   string
}

// postgresDsnParse  数据库解析
func (s *SQLDump) postgresDsnParse() *PostgresDsn {
	result := new(PostgresDsn)
	// 分割连接字符串
	params := strings.Split(s.dsn, " ")

	// 解析参数
	for _, param := range params {
		keyValue := strings.Split(param, "=")
		if len(keyValue) != 2 {
			continue
		}
		key := keyValue[0]
		value := keyValue[1]
		switch key {
		case "host":
			result.Host = value
		case "port":
			if p, err := strconv.Atoi(value); err == nil {
				result.Port = p
			}
		case "user":
			result.User = value
		case "password":
			result.Password = value
		case "dbname":
			result.Dbname = value
		}
	}
	return result
}

// remove 移除多余行
func (s *SQLDump) postgresRemove(str string) string {
	var result string
	reader := strings.NewReader(str)
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" || strings.HasPrefix(line, "--") || strings.HasPrefix(line, "SELECT") || strings.HasPrefix(line, "SET") || regexp.MustCompile(`(ALTER TABLE .*? OWNER TO postgres)`).MatchString(line) {
			continue
		}
		result += fmt.Sprintln(line)
	}
	return result
}
