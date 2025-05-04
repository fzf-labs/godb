package sqltopb

import (
	"strings"

	"github.com/fzf-labs/godb/orm/gen"
	"github.com/fzf-labs/godb/orm/gormx"
	"github.com/spf13/cobra"
)

var CmdSQLToPb = &cobra.Command{
	Use:   "sqltopb",
	Short: "sql generate proto file",
	Long:  "sql generate proto file",
	Run:   Run,
}

var (
	db           string // 数据库类型 mysql postgres
	dsn          string // 数据库连接
	targetTables string // 指定表
	pbPackage    string // proto 包名
	pbGoPackage  string // proto go包名
	outPutPath   string // 输出路径
)

//nolint:gochecknoinits
func init() {
	CmdSQLToPb.Flags().StringVarP(&db, "db", "d", "", "db: mysql postgres")
	CmdSQLToPb.Flags().StringVarP(&dsn, "dsn", "s", "", "dsn")
	CmdSQLToPb.Flags().StringVarP(&targetTables, "tables", "t", "", "tables")
	CmdSQLToPb.Flags().StringVarP(&pbPackage, "pbPackage", "p", "pb", "pbPackage")
	CmdSQLToPb.Flags().StringVarP(&pbGoPackage, "pbGoPackage", "g", "github.com/fzf-labs/godb/orm/example/pb;pb", "pbGoPackage")
	CmdSQLToPb.Flags().StringVarP(&outPutPath, "outPutPath", "o", "./pb", "outPutPath")
}

func Run(_ *cobra.Command, _ []string) {
	gen.NewGenerationPB(gormx.NewSimpleGormClient(db, dsn), outPutPath, pbPackage, pbGoPackage, gen.WithPBOpts(gen.ModelOptionRemoveDefault(), gen.ModelOptionUnderline("UL")), gen.WithPBTables(strings.Split(targetTables, ","))).Do()
}
