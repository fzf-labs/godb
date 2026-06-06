//nolint:all
package repo

import (
	"fmt"

	"github.com/fzf-labs/godb/orm/utils/fileutil"
	"gorm.io/gorm"
)

// GenerationTable 为单张表生成 repo 层代码文件。
func GenerationTable(db *gorm.DB, dbname, daoPath, modelPath, repoPath, table string, partitionTables []string, columnNameToDataType, columnNameToName, columnNameToFieldType map[string]string) error {
	var file string
	g := Repo{
		gorm:                  db,
		daoPath:               daoPath,
		modelPath:             modelPath,
		repoPath:              repoPath,
		table:                 table,
		partitionTables:       partitionTables,
		columnNameToDataType:  columnNameToDataType,
		columnNameToName:      columnNameToName,
		columnNameToFieldType: columnNameToFieldType,
		dbName:                dbname,
		firstTableChar:        "",
		lowerTableName:        "",
		upperTableName:        "",
		daoPkgPath:            fileutil.FillModelPkgPath(daoPath),
		modelPkgPath:          fileutil.FillModelPkgPath(modelPath),
		index:                 make([]DBIndex, 0),
		haveDeletedAt:         hasDeletedAt(columnNameToDataType),
	}
	// 查询当前db的索引
	index, err := g.processIndex()
	if err != nil {
		return err
	}
	g.index = index
	g.lowerTableName = g.lowerName(table)
	g.upperTableName = g.upperName(table)
	g.firstTableChar = g.lowerTableName[0:1]
	generatePkg, err := g.generatePkg()
	if err != nil {
		return err
	}
	generateImport, err := g.generateImport()
	if err != nil {
		return err
	}
	generateVar, err := g.generateVar()
	if err != nil {
		return err
	}
	generateTypes, err := g.generateTypes()
	if err != nil {
		return err
	}
	generateNew, err := g.generateNew()
	if err != nil {
		return err
	}
	generateCommonFunc, err := g.generateCommonFunc()
	if err != nil {
		return err
	}
	generateCreateFunc, err := g.generateCreateFunc()
	if err != nil {
		return err
	}
	generateUpdateFunc, err := g.generateUpdateFunc()
	if err != nil {
		return err
	}
	generateReadFunc, err := g.generateReadFunc()
	if err != nil {
		return err
	}
	generateDelFunc, err := g.generateDelFunc()
	if err != nil {
		return err
	}
	file += fmt.Sprintln(generatePkg)
	file += fmt.Sprintln(generateImport)
	file += fmt.Sprintln(generateVar)
	file += fmt.Sprintln(generateTypes)
	file += fmt.Sprintln(generateNew)
	file += fmt.Sprintln(generateCommonFunc)
	file += fmt.Sprintln(generateCreateFunc)
	file += fmt.Sprintln(generateUpdateFunc)
	file += fmt.Sprintln(generateReadFunc)
	file += fmt.Sprintln(generateDelFunc)
	outputFile := g.repoPath + "/" + table + ".repo.go"
	err = g.output(outputFile, []byte(file))
	if err != nil {
		return err
	}
	return nil
}

// Repo 保存单表 repo 代码生成过程中的模板上下文。
type Repo struct {
	gorm                  *gorm.DB          // 数据库
	daoPath               string            // dao所在的路径
	modelPath             string            // model所在的路径
	repoPath              string            // repo所在的路径
	table                 string            // 表名称
	partitionTables       []string          // 子分区表名称
	columnNameToDataType  map[string]string // 字段名称对应的类型
	columnNameToName      map[string]string // 字段名称对应的Go名称
	columnNameToFieldType map[string]string // 字段名称对应的dao类型
	dbName                string            // 数据库名称
	firstTableChar        string            // 表名称第一个字母
	lowerTableName        string            // 表名称小写
	upperTableName        string            // 表名称大写
	daoPkgPath            string            // go文件中daoPkgPath
	modelPkgPath          string            // go文件中modelPkgPath
	index                 []DBIndex         // 索引
	haveDeletedAt         bool              // 是否有删除标记
}

// DBIndex 描述生成 repo 方法时使用的数据库索引信息。
type DBIndex struct {
	Name       string   // 索引名称
	ColumnKey  string   // 索引字段KEY
	PrimaryKey bool     // 是否是主键
	Unique     bool     // 是否是唯一索引
	Columns    []string // 索引字段
}
