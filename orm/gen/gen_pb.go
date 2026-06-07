package gen

import (
	"errors"
	"fmt"
	"strings"
	"sync"

	"golang.org/x/sync/errgroup"
	"gorm.io/gen"
	"gorm.io/gorm"

	"github.com/fzf-labs/godb/orm/gen/proto"
	"github.com/fzf-labs/godb/orm/gormx"
	"github.com/fzf-labs/godb/orm/utils/strutil"
)

// NewGenerationPB SQL 生成 proto
func NewGenerationPB(db *gorm.DB, outPutPath, packageStr, goPackageStr string, opts ...OptionPB) *GenerationPb {
	g := &GenerationPb{
		gorm:         db,
		tables:       make([]string, 0),
		outPutPath:   outPutPath,
		packageStr:   packageStr,
		goPackageStr: goPackageStr,
	}
	if len(opts) > 0 {
		for _, v := range opts {
			v(g)
		}
	}
	return g
}

// GenerationPb 保存 SQL 生成 proto 文件所需配置。
type GenerationPb struct {
	gorm         *gorm.DB       // 数据库
	tables       []string       // 指定表
	outPutPath   string         // 文件生成地址
	opts         []gen.ModelOpt // 特殊处理逻辑函
	packageStr   string         // 包名
	goPackageStr string         // 包路径
}

// OptionPB 配置 proto 文件生成器。
type OptionPB func(gen *GenerationPb)

// WithPBOpts 选项函数-自定义特殊设置
func WithPBOpts(opts ...gen.ModelOpt) OptionPB {
	return func(r *GenerationPb) {
		r.opts = opts
	}
}

// WithPBTables 选项函数-指定表
func WithPBTables(tables []string) OptionPB {
	return func(r *GenerationPb) {
		r.tables = tables
	}
}

// Do 执行 proto 文件生成。
func (g *GenerationPb) Do() (err error) {
	defer func() {
		if r := recover(); r != nil {
			// gorm/gen 的生成过程可能通过 panic 暴露失败，这里统一转为 error。
			err = fmt.Errorf("generate pb code panic: %v", r)
		}
	}()
	if g.gorm == nil {
		return fmt.Errorf("db cannot be nil")
	}
	if strings.TrimSpace(g.outPutPath) == "" {
		return fmt.Errorf("output path cannot be empty")
	}
	if strings.TrimSpace(g.packageStr) == "" {
		return fmt.Errorf("package cannot be empty")
	}
	if strings.TrimSpace(g.goPackageStr) == "" {
		return fmt.Errorf("go package cannot be empty")
	}
	// 初始化
	generator := gen.NewGenerator(gen.Config{})
	// 使用数据库
	generator.UseDB(g.gorm)
	// json 小驼峰模型命名
	generator.WithJSONTagNameStrategy(JSONTagNameStrategy)
	// 特殊处理逻辑
	if len(g.opts) > 0 {
		generator.WithOpts(g.opts...)
	}
	// 获取所有表
	tables, err := g.gorm.Migrator().GetTables()
	if err != nil {
		return fmt.Errorf("get database tables: %w", err)
	}
	if len(g.tables) > 0 {
		tables = g.tables
	}
	tables, err = normalizeTableNames(tables)
	if err != nil {
		return err
	}
	// 查询分区表父级到子表的映射
	partitionTableToChildTables, err := gormx.GetPartitionTableToChildTables(g.gorm)
	if err != nil {
		return fmt.Errorf("get partition table children: %w", err)
	}
	partitionChildTables := make([]string, 0)
	for _, v := range partitionTableToChildTables {
		partitionChildTables = append(partitionChildTables, v...)
	}
	// 去掉tables中的partitionChildTables
	tables = strutil.SliRemove(tables, partitionChildTables)
	var group errgroup.Group
	var mu sync.Mutex
	genErrs := make([]error, 0)
	for _, v := range tables {
		table := v
		// 表字段对应的名称
		columnNameToName := make(map[string]string)
		// 表字段对应的类型
		columnNameToDataType := make(map[string]string)
		queryStructMeta := generator.GenerateModel(table)
		for _, vv := range queryStructMeta.Fields {
			columnNameToName[vv.ColumnName] = vv.Name
			columnNameToDataType[vv.ColumnName] = strings.TrimLeft(vv.Type, "*")
		}
		group.Go(func() error {
			if err := proto.GenerationPB(g.gorm, g.outPutPath, g.packageStr, g.goPackageStr, table, columnNameToName, columnNameToDataType); err != nil {
				err = fmt.Errorf("generate proto for table %q: %w", table, err)
				mu.Lock()
				genErrs = append(genErrs, err)
				mu.Unlock()
				return err
			}
			return nil
		})
	}
	_ = group.Wait()
	if len(genErrs) > 0 {
		return errors.Join(genErrs...)
	}
	return nil
}
