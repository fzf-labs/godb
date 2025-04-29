package sqldump

import (
	"github.com/spf13/cobra"
)

var CmdSQLDump = &cobra.Command{
	Use:   "sqldump",
	Short: "Export database table structure",
	Long:  "Export database table structure",
	Run:   Run,
}

var (
	db            string // 数据库类型 mysql postgres
	dsn           string // 数据库连接
	outPutPath    string // 输出路径
	targetTables  string // 指定表
	fileOverwrite bool   // 是否覆盖
)

//nolint:gochecknoinits
func init() {
	CmdSQLDump.Flags().StringVarP(&db, "db", "d", "", "db")
	CmdSQLDump.Flags().StringVarP(&dsn, "dsn", "s", "", "dsn")
	CmdSQLDump.Flags().StringVarP(&outPutPath, "outPutPath", "o", "./doc/sql", "outPutPath")
	CmdSQLDump.Flags().StringVarP(&targetTables, "tables", "t", "", "tables")
	CmdSQLDump.Flags().BoolVarP(&fileOverwrite, "fileOverwrite", "f", false, "file overwrite")
}

func Run(_ *cobra.Command, _ []string) {
	NewSQLDump(db, dsn, outPutPath, targetTables, fileOverwrite).Run()
}
