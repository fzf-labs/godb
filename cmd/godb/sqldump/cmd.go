package sqldump

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// CmdSQLDump 是导出数据库表结构 SQL 的 cobra 子命令。
var CmdSQLDump = &cobra.Command{
	Use:   "sqldump",
	Short: "Export database table structure",
	Long:  "Export database table structure",
	RunE:  Run,
}

var (
	db            string // 数据库类型 mysql postgres
	dsn           string // 数据库连接
	outPutPath    string // 输出路径
	targetTables  string // 指定表
	fileOverwrite bool   // 是否覆盖
)

type runOptions struct {
	db            string
	dsn           string
	outPutPath    string
	targetTables  string
	fileOverwrite bool
}

// init 注册 sqldump 命令行参数。
//
//nolint:gochecknoinits
func init() {
	CmdSQLDump.Flags().StringVarP(&db, "db", "d", "", "db")
	CmdSQLDump.Flags().StringVarP(&dsn, "dsn", "s", "", "dsn")
	CmdSQLDump.Flags().StringVarP(&outPutPath, "outPutPath", "o", "./doc/sql", "outPutPath")
	CmdSQLDump.Flags().StringVarP(&targetTables, "tables", "t", "", "tables")
	CmdSQLDump.Flags().BoolVarP(&fileOverwrite, "fileOverwrite", "f", false, "file overwrite")
}

// Run 执行数据库结构导出命令。
func Run(_ *cobra.Command, _ []string) error {
	return runWithOptions(snapshotRunOptions())
}

func snapshotRunOptions() runOptions {
	return runOptions{
		db:            db,
		dsn:           dsn,
		outPutPath:    outPutPath,
		targetTables:  targetTables,
		fileOverwrite: fileOverwrite,
	}
}

func runWithOptions(opts runOptions) error {
	if err := opts.validate(); err != nil {
		return err
	}
	return NewSQLDump(opts.db, opts.dsn, opts.outPutPath, opts.targetTables, opts.fileOverwrite).Run()
}

func (o runOptions) validate() error {
	if strings.TrimSpace(o.db) == "" {
		return fmt.Errorf("db cannot be empty")
	}
	if strings.TrimSpace(o.dsn) == "" {
		return fmt.Errorf("dsn cannot be empty")
	}
	if strings.TrimSpace(o.outPutPath) == "" {
		return fmt.Errorf("output path cannot be empty")
	}
	return nil
}
