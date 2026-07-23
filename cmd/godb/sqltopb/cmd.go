package sqltopb

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"gorm.io/gorm"

	"github.com/fzf-labs/godb/cmd/godb/internal/tablelist"
	"github.com/fzf-labs/godb/orm/gen"
	genproto "github.com/fzf-labs/godb/orm/gen/proto"
	"github.com/fzf-labs/godb/orm/gormx"
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

var (
	newSimpleGormClient = gormx.NewSimpleGormClient
	generatePBDo        = (*gen.GenerationPb).Do
)

type runOptions struct {
	db           string
	dsn          string
	targetTables string
	pbPackage    string
	pbGoPackage  string
	outPutPath   string
}

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
	return runWithOptions(snapshotRunOptions())
}

func snapshotRunOptions() runOptions {
	return runOptions{
		db:           db,
		dsn:          dsn,
		targetTables: targetTables,
		pbPackage:    pbPackage,
		pbGoPackage:  pbGoPackage,
		outPutPath:   outPutPath,
	}
}

func runWithOptions(opts runOptions) error {
	opts = opts.normalize()
	if err := opts.validate(); err != nil {
		return err
	}
	tables, err := tablelist.ParseCSV(opts.targetTables)
	if err != nil {
		return err
	}
	dbClient, err := newSimpleGormClient(opts.db, opts.dsn)
	if err != nil {
		return err
	}
	if dbClient == nil {
		return fmt.Errorf("sqltopb database client cannot be nil")
	}
	defer closeGormDB(dbClient)
	return generatePBDo(gen.NewGenerationPB(
		dbClient,
		opts.outPutPath,
		opts.pbPackage,
		opts.pbGoPackage,
		gen.WithPBOpts(
			gen.ModelOptionRemoveDefault(),
			gen.ModelOptionUnderline("UL"),
		),
		gen.WithPBTables(tables),
	))
}

func (o runOptions) normalize() runOptions {
	o.db = strings.ToLower(strings.TrimSpace(o.db))
	o.dsn = strings.TrimSpace(o.dsn)
	o.targetTables = strings.TrimSpace(o.targetTables)
	o.pbPackage = strings.TrimSpace(o.pbPackage)
	o.pbGoPackage = strings.TrimSpace(o.pbGoPackage)
	o.outPutPath = strings.TrimSpace(o.outPutPath)
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
	if strings.TrimSpace(o.pbPackage) == "" {
		return fmt.Errorf("pb package cannot be empty")
	}
	if strings.TrimSpace(o.pbGoPackage) == "" {
		return fmt.Errorf("pb go package cannot be empty")
	}
	if err := genproto.ValidateProtoPackageStr(o.pbPackage); err != nil {
		return err
	}
	if err := genproto.ValidateGoPackageStr(o.pbGoPackage); err != nil {
		return err
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
