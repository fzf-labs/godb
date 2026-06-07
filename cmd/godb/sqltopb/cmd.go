package sqltopb

import (
	"github.com/fzf-labs/godb/cmd/godb/internal/tablelist"
	"github.com/fzf-labs/godb/orm/gen"
	"github.com/fzf-labs/godb/orm/gormx"
	"github.com/spf13/cobra"
)

// CmdSQLToPb 是根据 SQL 表结构生成 proto 文件的 cobra 子命令。
var CmdSQLToPb = &cobra.Command{
	Use:   "sqltopb",
	Short: "sql generate proto file",
	Long:  "sql generate proto file",
	RunE:  Run,
}

var (
	db           string // 数据库类型 mysql postgres
	dsn          string // 数据库连接
	targetTables string // 指定表
	pbPackage    string // proto 包名
	pbGoPackage  string // proto go包名
	outPutPath   string // 输出路径
)

// init 注册 sqltopb 命令行参数。
//
//nolint:gochecknoinits
func init() {
	CmdSQLToPb.Flags().StringVarP(&db, "db", "d", "", "db: mysql postgres")
	CmdSQLToPb.Flags().StringVarP(&dsn, "dsn", "s", "", "dsn")
	CmdSQLToPb.Flags().StringVarP(&targetTables, "tables", "t", "", "tables")
	CmdSQLToPb.Flags().StringVarP(&pbPackage, "pbPackage", "p", "pb", "pbPackage")
	CmdSQLToPb.Flags().StringVarP(&pbGoPackage, "pbGoPackage", "g", "github.com/fzf-labs/godb/orm/example/pb;pb", "pbGoPackage")
	CmdSQLToPb.Flags().StringVarP(&outPutPath, "outPutPath", "o", "./pb", "outPutPath")
}

// Run 执行 SQL 转 proto 命令。
func Run(_ *cobra.Command, _ []string) error {
	tables, err := tablelist.ParseCSV(targetTables)
	if err != nil {
		return err
	}
	dbClient, err := gormx.NewSimpleGormClient(db, dsn)
	if err != nil {
		return err
	}
	return gen.NewGenerationPB(
		dbClient,
		outPutPath,
		pbPackage,
		pbGoPackage,
		gen.WithPBOpts(
			gen.ModelOptionRemoveDefault(),
			gen.ModelOptionUnderline("UL"),
		),
		gen.WithPBTables(tables),
	).Do()
}
