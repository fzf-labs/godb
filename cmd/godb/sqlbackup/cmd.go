// Package sqlbackup 提供逻辑数据库备份命令，支持 PostgreSQL 与 MySQL。
package sqlbackup

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

// CmdSQLBackup 是创建逻辑数据库备份的 cobra 子命令。
var CmdSQLBackup = &cobra.Command{
	Use:   "sqlbackup",
	Short: "Create a logical database backup",
	Long:  "Create a logical PostgreSQL or MySQL backup containing schema and data.",
	Args:  cobra.NoArgs,
	RunE:  Run,
}

var (
	db     string // 数据库类型：postgres 或 mysql
	dsn    string // 数据库连接串
	output string // 备份输出文件路径
	force  bool   // 是否覆盖已存在的输出文件
)

// runOptions 保存一次命令执行所需的运行参数。
type runOptions struct {
	db     string
	dsn    string
	output string
	force  bool
}

// init 注册 sqlbackup 命令行参数。
//
//nolint:gochecknoinits
func init() {
	CmdSQLBackup.Flags().StringVarP(&db, "db", "d", "", "database type: postgres or mysql")
	CmdSQLBackup.Flags().StringVarP(&dsn, "dsn", "s", "", "database connection string")
	CmdSQLBackup.Flags().StringVarP(&output, "output", "o", "", "backup output file")
	CmdSQLBackup.Flags().BoolVarP(&force, "force", "f", false, "replace an existing output file")
}

// Run 执行 sqlbackup 命令，成功后将输出文件路径打印到标准输出。
func Run(cmd *cobra.Command, _ []string) error {
	outputPath, err := runWithOptions(commandContext(cmd), snapshotRunOptions())
	if err != nil {
		return err
	}
	if cmd == nil {
		return nil
	}
	_, err = fmt.Fprintln(cmd.OutOrStdout(), outputPath)
	return err
}

// snapshotRunOptions 从全局 flag 变量复制一份运行参数，避免并发或后续修改影响当前执行。
func snapshotRunOptions() runOptions {
	return runOptions{
		db:     db,
		dsn:    dsn,
		output: output,
		force:  force,
	}
}

// runWithOptions 校验参数并执行备份，返回最终输出文件路径。
func runWithOptions(ctx context.Context, opts runOptions) (string, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	opts = opts.normalize()
	if err := opts.validate(); err != nil {
		return "", err
	}

	var err error
	switch opts.db {
	case databasePostgres:
		err = opts.runPostgres(ctx)
	case databaseMySQL:
		err = opts.runMySQL(ctx)
	default:
		err = fmt.Errorf("unknown database type: %s", opts.db)
	}
	if err != nil {
		return "", err
	}
	return opts.output, nil
}

// normalize 规范化运行参数，去除空白并统一数据库类型大小写。
func (o runOptions) normalize() runOptions {
	o.db = strings.ToLower(strings.TrimSpace(o.db))
	o.dsn = strings.TrimSpace(o.dsn)
	o.output = strings.TrimSpace(o.output)
	if o.output != "" {
		o.output = filepath.Clean(o.output)
	}
	return o
}

// validate 检查运行参数是否完整且合法。
func (o runOptions) validate() error {
	if o.db == "" {
		return fmt.Errorf("db cannot be empty")
	}
	if o.db != databasePostgres && o.db != databaseMySQL {
		return fmt.Errorf("unknown database type: %s", o.db)
	}
	if o.dsn == "" {
		return fmt.Errorf("dsn cannot be empty")
	}
	if o.output == "" {
		return fmt.Errorf("output path cannot be empty")
	}
	return nil
}

// commandContext 返回命令关联的 context；命令为空时回退到 Background。
func commandContext(cmd *cobra.Command) context.Context {
	if cmd == nil || cmd.Context() == nil {
		return context.Background()
	}
	return cmd.Context()
}
