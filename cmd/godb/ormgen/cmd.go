package ormgen

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	gormGen "gorm.io/gen"
	"gorm.io/gorm"

	"github.com/fzf-labs/godb/cmd/godb/internal/tablelist"
	"github.com/fzf-labs/godb/orm/gen"
	"github.com/fzf-labs/godb/orm/gormx"
)

// CmdOrmGen 是生成 ORM model、dao 和 repo 代码的 cobra 子命令。
var CmdOrmGen = &cobra.Command{
	Use:   "ormgen",
	Short: "Generate GORM model code",
	Long:  "Generate GORM model code from database tables",
	RunE:  Run,
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

var (
	newSimpleGormClient = gormx.NewSimpleGormClient
	generateDBDo        = (*gen.GenerationDB).Do
)

type runOptions struct {
	db                      string
	dsn                     string
	targetTables            string
	outPutPath              string
	optionUnderline         string
	optionPgDefaultString   bool
	optionRemoveDefault     bool
	optionRemoveGormTypeTag bool
}

// init 注册 ormgen 命令行参数。
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

// Run 执行 ORM 代码生成命令。
func Run(_ *cobra.Command, _ []string) error {
	return runWithOptions(snapshotRunOptions())
}

func snapshotRunOptions() runOptions {
	return runOptions{
		db:                      db,
		dsn:                     dsn,
		targetTables:            targetTables,
		outPutPath:              outPutPath,
		optionUnderline:         optionUnderline,
		optionPgDefaultString:   optionPgDefaultString,
		optionRemoveDefault:     optionRemoveDefault,
		optionRemoveGormTypeTag: optionRemoveGormTypeTag,
	}
}

func runWithOptions(opts runOptions) error {
	opts = opts.normalize()
	if err := opts.validate(); err != nil {
		return err
	}
	dbOpts := make([]gormGen.ModelOpt, 0)
	if opts.optionUnderline != "" {
		dbOpts = append(dbOpts, gen.ModelOptionUnderline(opts.optionUnderline))
	}
	if opts.optionPgDefaultString {
		dbOpts = append(dbOpts, gen.ModelOptionPgDefaultString())
	}
	if opts.optionRemoveDefault {
		dbOpts = append(dbOpts, gen.ModelOptionRemoveDefault())
	}
	if opts.optionRemoveGormTypeTag {
		dbOpts = append(dbOpts, gen.ModelOptionRemoveGormTypeTag())
	}
	tables, err := tablelist.ParseCSV(opts.targetTables)
	if err != nil {
		return err
	}
	if err := gen.ValidateTableNames(tables); err != nil {
		return err
	}
	dbClient, err := newSimpleGormClient(opts.db, opts.dsn)
	if err != nil {
		return err
	}
	if dbClient == nil {
		return fmt.Errorf("ormgen database client cannot be nil")
	}
	defer closeGormDB(dbClient)
	return generateDBDo(gen.NewGenerationDB(
		dbClient,
		opts.outPutPath,
		gen.WithDataMap(gen.DataTypeMap()),
		gen.WithTables(tables),
		gen.WithDBNameOpts(gen.DBNameOpts()),
		gen.WithDBOpts(dbOpts...),
	))
}

func (o runOptions) normalize() runOptions {
	o.db = strings.ToLower(strings.TrimSpace(o.db))
	o.dsn = strings.TrimSpace(o.dsn)
	o.targetTables = strings.TrimSpace(o.targetTables)
	o.outPutPath = strings.TrimSpace(o.outPutPath)
	o.optionUnderline = strings.TrimSpace(o.optionUnderline)
	return o
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

func closeGormDB(db *gorm.DB) {
	if db == nil {
		return
	}
	sqlDB, err := db.DB()
	if err != nil {
		return
	}
	_ = sqlDB.Close()
}
