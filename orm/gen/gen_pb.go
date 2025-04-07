package gen

import (
	"log"
	"sync"

	"github.com/fzf-labs/godb/orm/gen/proto"
	"github.com/fzf-labs/godb/orm/gormx"
	"github.com/fzf-labs/godb/orm/utils"
	"gorm.io/gen"
	"gorm.io/gorm"
)

// NewGenerationPB SQL 生成 proto
func NewGenerationPB(db *gorm.DB, outPutPath, packageStr, goPackageStr string, opts ...OptionPB) *GenerationPb {
	g := &GenerationPb{
		gorm:         db,
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

type GenerationPb struct {
	gorm         *gorm.DB       // 数据库
	outPutPath   string         // 文件生成地址
	opts         []gen.ModelOpt // 特殊处理逻辑函
	packageStr   string
	goPackageStr string
}

type OptionPB func(gen *GenerationPb)

// WithPBOpts 选项函数-自定义特殊设置
func WithPBOpts(opts ...gen.ModelOpt) OptionPB {
	return func(r *GenerationPb) {
		r.opts = opts
	}
}

func (g *GenerationPb) Do() {
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
		return
	}
	// 查询分区表父级到子表的映射
	partitionTableToChildTables, err := gormx.GetPartitionTableToChildTables(g.gorm)
	if err != nil {
		return
	}
	partitionChildTables := make([]string, 0)
	for _, v := range partitionTableToChildTables {
		partitionChildTables = append(partitionChildTables, v...)
	}
	// 去掉tables中的partitionChildTables
	tables = utils.SliRemove(tables, partitionChildTables)
	var wg sync.WaitGroup
	wg.Add(len(tables))
	for _, v := range tables {
		table := v
		// 表字段对应的名称
		columnNameToName := make(map[string]string)
		queryStructMeta := generator.GenerateModel(table)
		for _, vv := range queryStructMeta.Fields {
			columnNameToName[vv.ColumnName] = vv.Name
		}
		go func(db *gorm.DB, outPutPath, packageStr, goPackageStr, table string, columnNameToName map[string]string) {
			defer wg.Done()
			// 数据表repo代码生成
			err2 := proto.GenerationPB(db, outPutPath, packageStr, goPackageStr, table, columnNameToName)
			if err2 != nil {
				log.Println("repo GenerationTable err:", err2)
				return
			}
		}(g.gorm, g.outPutPath, g.packageStr, g.goPackageStr, table, columnNameToName)
	}
	wg.Wait()
}
