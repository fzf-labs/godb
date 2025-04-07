package ormgen

import (
	"strings"

	"github.com/fzf-labs/godb/orm/gen"
	"github.com/fzf-labs/godb/orm/gormx"
	"github.com/spf13/cobra"
	gormGen "gorm.io/gen"
)

var CmdOrmGen = &cobra.Command{
	Use:   "ormgen",
	Short: "Generate GORM model code",
	Long:  "Generate GORM model code from database tables",
	Run:   Run,
}

var (
	db                      string // 数据库类型 mysql postgres
	dsn                     string // 数据库自定义连接
	targetTables            string // 数据库指定表
	outPutPath              string // 输出路径
	optionUnderline         string // 选项：下划线转驼峰 (默认'_'替换为'UL')
	optionPgDefaultString   bool   // 选项：移除gorm tag default:'(.*?)'::character varying  (默认是 true)
	optionRemoveDefault     bool   // 选项：移除默认值 （默认是 true）
	optionRemoveGormTypeTag bool   // 选项：移除gorm tag :type (默认是 false)
)

func init() {
	CmdOrmGen.Flags().StringVarP(&db, "db", "d", "", "db: mysql postgres")
	CmdOrmGen.Flags().StringVarP(&dsn, "dsn", "s", "", "db dsn")
	CmdOrmGen.Flags().StringVarP(&targetTables, "tables", "t", "", "db tables")
	CmdOrmGen.Flags().StringVarP(&outPutPath, "outPutPath", "o", "./internal/data/gorm", "output path")
	CmdOrmGen.Flags().StringVarP(&optionUnderline, "optionUnderline", "u", "UL", "option: underline '_' replace 'UL'")
	CmdOrmGen.Flags().BoolVarP(&optionPgDefaultString, "optionPgDefaultString", "p", true, "option: pg default string ta removeying")
	CmdOrmGen.Flags().BoolVarP(&optionRemoveDefault, "optionRemoveDefault", "r", true, "option: remove tag default")
	CmdOrmGen.Flags().BoolVarP(&optionRemoveGormTypeTag, "optionRemoveGormTypeTag", "g", false, "option: remove gorm tag :type")
}

func Run(_ *cobra.Command, _ []string) {
	dbOpts := make([]gormGen.ModelOpt, 0)
	if optionUnderline != "" {
		dbOpts = append(dbOpts, gen.ModelOptionUnderline(optionUnderline))
	}
	if optionPgDefaultString {
		dbOpts = append(dbOpts, gen.ModelOptionPgDefaultString())
	}
	if optionRemoveDefault {
		dbOpts = append(dbOpts, gen.ModelOptionRemoveDefault())
	}
	if optionRemoveGormTypeTag {
		dbOpts = append(dbOpts, gen.ModelOptionRemoveGormTypeTag())
	}
	var tables []string
	if targetTables != "" {
		tables = strings.Split(targetTables, ",")
	}
	gen.NewGenerationDB(
		gormx.NewSimpleGormClient(db, dsn),
		outPutPath,
		gen.WithDataMap(gen.DataTypeMap()),
		gen.WithTables(tables),
		gen.WithDBNameOpts(gen.DBNameOpts()),
		gen.WithDBOpts(dbOpts...),
	).Do()
}
