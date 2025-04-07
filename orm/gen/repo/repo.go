//nolint:all
package repo

import (
	"fmt"
	"go/token"
	"os"
	"sort"
	"strconv"
	"strings"
	"unicode"

	"github.com/fzf-labs/godb/orm/gormx"
	"github.com/fzf-labs/godb/orm/utils"
	"github.com/fzf-labs/godb/orm/utils/template"
	"github.com/jinzhu/inflection"
	"golang.org/x/tools/imports"
	"gorm.io/gorm"
)

var KeyWords = []string{
	"dao",
	"parameters",
	"cacheKey",
	"cacheKeys",
	"cacheValue",
	"keyToParam",
	"resp",
	"result",
	"marshal",
}

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
		daoPkgPath:            utils.FillModelPkgPath(daoPath),
		modelPkgPath:          utils.FillModelPkgPath(modelPath),
		index:                 make([]DBIndex, 0),
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
}

type DBIndex struct {
	Name       string   // 索引名称
	ColumnKey  string   // 索引字段KEY
	PrimaryKey bool     // 是否是主键
	Unique     bool     // 是否是唯一索引
	Columns    []string // 索引字段
}

// processIndex 索引处理  索引去重和排序
func (r *Repo) processIndex() ([]DBIndex, error) {
	result := make([]DBIndex, 0)
	tmp := make([]DBIndex, 0)
	repeat := make(map[string]struct{})
	table := r.table
	// PG 需要特殊处理
	if len(r.partitionTables) > 0 {
		table = r.partitionTables[0]
	}
	// 获取索引
	indexes, err := gormx.GetIndexes(r.gorm, table)
	if err != nil {
		return nil, err
	}
	// 获取排序的索引字段
	sortIndexColumns, err := gormx.SortIndexColumns(r.gorm, table)
	if err != nil {
		return nil, err
	}
	for _, v := range indexes {
		columns := sortIndexColumns[v.IndexName]
		tmp = append(tmp, DBIndex{
			Name:       v.IndexName,
			ColumnKey:  strings.Join(columns, "_"),
			PrimaryKey: v.Primary,
			Unique:     v.IsUnique,
			Columns:    columns,
		})
	}
	sort.Slice(tmp, func(i, j int) bool {
		return tmp[i].ColumnKey < tmp[j].ColumnKey
	})
	// 主键索引
	for _, v := range tmp {
		if v.PrimaryKey {
			_, ok := repeat[v.ColumnKey]
			if !ok {
				repeat[v.ColumnKey] = struct{}{}
				result = append(result, v)
			}
		}
	}
	// 唯一索引
	for _, v := range tmp {
		if !v.PrimaryKey && v.Unique {
			_, ok := repeat[v.ColumnKey]
			if !ok {
				repeat[v.ColumnKey] = struct{}{}
				result = append(result, v)
			}

		}
	}
	// 普通索引
	for _, v := range tmp {
		if !v.PrimaryKey && !v.Unique {
			_, ok := repeat[v.ColumnKey]
			if !ok {
				repeat[v.ColumnKey] = struct{}{}
				result = append(result, v)
			}
		}
	}
	// 最左匹配原则索引
	for _, v := range tmp {
		if !v.PrimaryKey && len(v.Columns) > 1 {
			for i := len(v.Columns); i > 0; i-- {
				columnKey := strings.Join(v.Columns[0:i], "_")
				_, ok := repeat[columnKey]
				if !ok {
					repeat[columnKey] = struct{}{}
					result = append(result, DBIndex{
						Name:       v.Name,
						ColumnKey:  columnKey,
						PrimaryKey: false,
						Unique:     false,
						Columns:    v.Columns[0:i],
					})
				}
			}
		}
	}
	return result, nil
}

// output 导出文件
func (r *Repo) output(fileName string, content []byte) error {
	result, err := imports.Process(fileName, content, nil)
	if err != nil {
		lines := strings.Split(string(content), "\n")
		errLine, _ := strconv.Atoi(strings.Split(err.Error(), ":")[1])
		startLine, endLine := errLine-5, errLine+5
		fmt.Println("Format fail:", errLine, err)
		if startLine < 0 {
			startLine = 0
		}
		if endLine > len(lines)-1 {
			endLine = len(lines) - 1
		}
		for i := startLine; i <= endLine; i++ {
			fmt.Println(i, lines[i])
		}
		return fmt.Errorf("cannot format file: %w", err)
	}
	return os.WriteFile(fileName, result, 0600)
}

// generatePkg
func (r *Repo) generatePkg() (string, error) {
	tplParams := map[string]any{
		"dbName": r.dbName,
	}
	tpl, err := template.NewTemplate().Parse(Pkg).Execute(tplParams)
	if err != nil {
		return "", err
	}
	return tpl.String(), nil
}

// generateImport
func (r *Repo) generateImport() (string, error) {
	tplParams := map[string]any{
		"daoPkgPath":   r.daoPkgPath,
		"modelPkgPath": r.modelPkgPath,
	}
	tpl, err := template.NewTemplate().Parse(Import).Execute(tplParams)
	if err != nil {
		return "", err
	}
	return tpl.String(), nil
}

// generateVar
func (r *Repo) generateVar() (string, error) {
	var varStr string
	var cacheKeys string
	varCacheGlobalTpl, err := template.NewTemplate().Parse(VarCacheGlobal).Execute(map[string]any{
		"upperTableName": r.upperTableName,
	})
	if err != nil {
		return "", err
	}
	cacheKeys += varCacheGlobalTpl.String()
	// 生成缓存key
	for _, v := range r.index {
		if r.checkDaoFieldType(v.Columns) {
			continue
		}
		varCacheTpl, err := template.NewTemplate().Parse(VarCache).Execute(map[string]any{
			"dbName":         r.dbName,
			"upperTableName": r.upperTableName,
			"cacheFields":    r.cacheFields(v.Columns),
		})
		if err != nil {
			return "", err
		}
		cacheKeys += varCacheTpl.String()
	}
	// 生成变量
	varTpl, err := template.NewTemplate().Parse(Var).Execute(map[string]any{
		"upperTableName": r.upperTableName,
	})
	if err != nil {
		return "", err
	}
	varStr += fmt.Sprintln(varTpl.String())
	if len(cacheKeys) > 0 {
		varCacheKeysTpl, err := template.NewTemplate().Parse(VarCacheKeys).Execute(map[string]any{
			"cacheKeys": cacheKeys,
		})
		if err != nil {
			return "", err
		}
		varStr += fmt.Sprintln(varCacheKeysTpl.String())
	}
	return varStr, nil
}
func (r *Repo) generateCommonMethods() (string, error) {
	var commonMethods string
	tplParams := map[string]any{
		"dbName":         r.dbName,
		"upperTableName": r.upperTableName,
	}
	interfaceDeepCopy, err := template.NewTemplate().Parse(InterfaceDeepCopy).Execute(tplParams)
	if err != nil {
		return "", err
	}
	commonMethods += fmt.Sprintln(interfaceDeepCopy.String())
	return commonMethods, nil
}

// generateCreateMethods
func (r *Repo) generateCreateMethods() (string, error) {
	// 是否有索引
	haveIndex := len(r.index) > 0
	// 是否有主键索引
	havePrimaryKey := false
	for _, v := range r.index {
		if v.PrimaryKey {
			havePrimaryKey = true
			break
		}
	}

	var createMethods string
	tplParams := map[string]any{
		"dbName":         r.dbName,
		"upperTableName": r.upperTableName,
	}
	interfaceCreateOne, err := template.NewTemplate().Parse(InterfaceCreateOne).Execute(tplParams)
	if err != nil {
		return "", err
	}
	createMethods += fmt.Sprintln(interfaceCreateOne.String())

	if haveIndex {
		interfaceCreateOneCache, err := template.NewTemplate().Parse(InterfaceCreateOneCache).Execute(tplParams)
		if err != nil {
			return "", err
		}
		createMethods += fmt.Sprintln(interfaceCreateOneCache.String())
	}

	interfaceCreateOneByTx, err := template.NewTemplate().Parse(InterfaceCreateOneByTx).Execute(tplParams)
	if err != nil {
		return "", err
	}
	createMethods += fmt.Sprintln(interfaceCreateOneByTx.String())

	if haveIndex {
		interfaceCreateOneCacheByTx, err := template.NewTemplate().Parse(InterfaceCreateOneCacheByTx).Execute(tplParams)
		if err != nil {
			return "", err
		}
		createMethods += fmt.Sprintln(interfaceCreateOneCacheByTx.String())
	}

	interfaceCreateBatch, err := template.NewTemplate().Parse(InterfaceCreateBatch).Execute(tplParams)
	if err != nil {
		return "", err
	}
	createMethods += fmt.Sprintln(interfaceCreateBatch.String())
	if haveIndex {
		interfaceCreateBatchCache, err := template.NewTemplate().Parse(InterfaceCreateBatchCache).Execute(tplParams)
		if err != nil {
			return "", err
		}
		createMethods += fmt.Sprintln(interfaceCreateBatchCache.String())
	}

	interfaceCreateBatchByTx, err := template.NewTemplate().Parse(InterfaceCreateBatchByTx).Execute(tplParams)
	if err != nil {
		return "", err
	}
	createMethods += fmt.Sprintln(interfaceCreateBatchByTx.String())

	if haveIndex {
		interfaceCreateBatchCacheByTx, err := template.NewTemplate().Parse(InterfaceCreateBatchCacheByTx).Execute(tplParams)
		if err != nil {
			return "", err
		}
		createMethods += fmt.Sprintln(interfaceCreateBatchCacheByTx.String())
	}
	if havePrimaryKey {
		interfaceUpsertOne, err := template.NewTemplate().Parse(InterfaceUpsertOne).Execute(tplParams)
		if err != nil {
			return "", err
		}
		createMethods += fmt.Sprintln(interfaceUpsertOne.String())
		interfaceUpsertOneCache, err := template.NewTemplate().Parse(InterfaceUpsertOneCache).Execute(tplParams)
		if err != nil {
			return "", err
		}
		createMethods += fmt.Sprintln(interfaceUpsertOneCache.String())
		interfaceUpsertOneByTx, err := template.NewTemplate().Parse(InterfaceUpsertOneByTx).Execute(tplParams)
		if err != nil {
			return "", err
		}
		createMethods += fmt.Sprintln(interfaceUpsertOneByTx.String())
		interfaceUpsertOneCacheByTx, err := template.NewTemplate().Parse(InterfaceUpsertOneCacheByTx).Execute(tplParams)
		if err != nil {
			return "", err
		}
		createMethods += fmt.Sprintln(interfaceUpsertOneCacheByTx.String())
	}

	interfaceUpsertOneByFields, err := template.NewTemplate().Parse(InterfaceUpsertOneByFields).Execute(tplParams)
	if err != nil {
		return "", err
	}
	createMethods += fmt.Sprintln(interfaceUpsertOneByFields.String())

	if haveIndex {
		interfaceUpsertOneCacheByFields, err := template.NewTemplate().Parse(InterfaceUpsertOneCacheByFields).Execute(tplParams)
		if err != nil {
			return "", err
		}
		createMethods += fmt.Sprintln(interfaceUpsertOneCacheByFields.String())
	}

	interfaceUpsertOneByFieldsTx, err := template.NewTemplate().Parse(InterfaceUpsertOneByFieldsTx).Execute(tplParams)
	if err != nil {
		return "", err
	}
	createMethods += fmt.Sprintln(interfaceUpsertOneByFieldsTx.String())

	if haveIndex {
		interfaceUpsertOneCacheByFieldsTx, err := template.NewTemplate().Parse(InterfaceUpsertOneCacheByFieldsTx).Execute(tplParams)
		if err != nil {
			return "", err
		}
		createMethods += fmt.Sprintln(interfaceUpsertOneCacheByFieldsTx.String())
	}
	return createMethods, nil
}

// generateUpdateMethods
func (r *Repo) generateUpdateMethods() (string, error) {
	// 是否有主键索引
	havePrimaryKey := false
	singlePrimaryKeyField := ""
	for _, v := range r.index {
		if v.PrimaryKey {
			havePrimaryKey = true
			if len(v.Columns) == 1 {
				singlePrimaryKeyField = v.Columns[0]
			}
			break
		}
	}
	if !havePrimaryKey {
		return "", nil
	}
	var updateMethods string
	tplParams := map[string]any{
		"dbName":                      r.dbName,
		"upperTableName":              r.upperTableName,
		"dataTypeSinglePrimaryKey":    r.columnNameToDataType[singlePrimaryKeyField],
		"upperSinglePrimaryKey":       r.upperFieldName(singlePrimaryKeyField),
		"upperSinglePrimaryKeyPlural": r.plural(r.upperFieldName(singlePrimaryKeyField)),
		"lowerSinglePrimaryKeyPlural": r.plural(r.lowerFieldName(singlePrimaryKeyField)),
	}
	interfaceUpdateOne, err := template.NewTemplate().Parse(InterfaceUpdateOne).Execute(tplParams)
	if err != nil {
		return "", err
	}
	updateMethods += fmt.Sprintln(interfaceUpdateOne.String())

	interfaceUpdateOneCache, err := template.NewTemplate().Parse(InterfaceUpdateOneCache).Execute(tplParams)
	if err != nil {
		return "", err
	}
	updateMethods += fmt.Sprintln(interfaceUpdateOneCache.String())

	interfaceUpdateOneByTx, err := template.NewTemplate().Parse(InterfaceUpdateOneByTx).Execute(tplParams)
	if err != nil {
		return "", err
	}
	updateMethods += fmt.Sprintln(interfaceUpdateOneByTx.String())

	interfaceUpdateOneCacheByTx, err := template.NewTemplate().Parse(InterfaceUpdateOneCacheByTx).Execute(tplParams)
	if err != nil {
		return "", err
	}
	updateMethods += fmt.Sprintln(interfaceUpdateOneCacheByTx.String())

	interfaceUpdateOneWithZero, err := template.NewTemplate().Parse(InterfaceUpdateOneWithZero).Execute(tplParams)
	if err != nil {
		return "", err
	}
	updateMethods += fmt.Sprintln(interfaceUpdateOneWithZero.String())

	interfaceUpdateOneCacheWithZero, err := template.NewTemplate().Parse(InterfaceUpdateOneCacheWithZero).Execute(tplParams)
	if err != nil {
		return "", err
	}
	updateMethods += fmt.Sprintln(interfaceUpdateOneCacheWithZero.String())

	interfaceUpdateOneWithZeroByTx, err := template.NewTemplate().Parse(InterfaceUpdateOneWithZeroByTx).Execute(tplParams)
	if err != nil {
		return "", err
	}
	updateMethods += fmt.Sprintln(interfaceUpdateOneWithZeroByTx.String())

	interfaceUpdateOneCacheWithZeroByTx, err := template.NewTemplate().Parse(InterfaceUpdateOneCacheWithZeroByTx).Execute(tplParams)
	if err != nil {
		return "", err
	}
	updateMethods += fmt.Sprintln(interfaceUpdateOneCacheWithZeroByTx.String())

	interfaceUpdateBatchByPrimaryKeys, err := template.NewTemplate().Parse(InterfaceUpdateBatchByPrimaryKeys).Execute(tplParams)
	if err != nil {
		return "", err
	}
	updateMethods += fmt.Sprintln(interfaceUpdateBatchByPrimaryKeys.String())

	interfaceUpdateBatchByPrimaryKeysTx, err := template.NewTemplate().Parse(InterfaceUpdateBatchByPrimaryKeysTx).Execute(tplParams)
	if err != nil {
		return "", err
	}
	updateMethods += fmt.Sprintln(interfaceUpdateBatchByPrimaryKeysTx.String())

	return updateMethods, nil
}

// generateReadMethods
func (r *Repo) generateReadMethods() (string, error) {
	var readMethods string
	for _, v := range r.index {
		if r.checkDaoFieldType(v.Columns) {
			continue
		}
		tplParams := map[string]any{
			"firstTableChar":    r.firstTableChar,
			"dbName":            r.dbName,
			"upperTableName":    r.upperTableName,
			"lowerTableName":    r.lowerTableName,
			"upperField":        r.upperFieldName(v.Columns[0]),
			"lowerField":        r.lowerFieldName(v.Columns[0]),
			"upperFieldPlural":  r.plural(r.upperFieldName(v.Columns[0])),
			"lowerFieldPlural":  r.plural(r.lowerFieldName(v.Columns[0])),
			"dataType":          r.columnNameToDataType[v.Columns[0]],
			"upperFields":       r.upperFields(v.Columns),
			"fieldAndDataTypes": r.fieldAndDataTypes(v.Columns),
		}
		// 唯一 && 字段数于1
		if v.Unique && len(v.Columns) == 1 {
			columnNameToDataType := r.columnNameToDataType[v.Columns[0]]

			interfaceFindOneByField, err := template.NewTemplate().Parse(InterfaceFindOneByField).Execute(tplParams)
			if err != nil {
				return "", err
			}
			readMethods += fmt.Sprintln(interfaceFindOneByField.String())

			interfaceFindOneCacheByField, err := template.NewTemplate().Parse(InterfaceFindOneCacheByField).Execute(tplParams)
			if err != nil {
				return "", err
			}
			readMethods += fmt.Sprintln(interfaceFindOneCacheByField.String())

			switch columnNameToDataType {
			case "bool":
			default:
				interfaceFindMultiByFieldPlural, err := template.NewTemplate().Parse(InterfaceFindMultiByFieldPlural).Execute(tplParams)
				if err != nil {
					return "", err
				}
				readMethods += fmt.Sprintln(interfaceFindMultiByFieldPlural.String())

				interfaceFindMultiCacheByFieldPluralUniqueTrue, err := template.NewTemplate().Parse(InterfaceFindMultiCacheByFieldPluralUniqueTrue).Execute(tplParams)
				if err != nil {
					return "", err
				}
				readMethods += fmt.Sprintln(interfaceFindMultiCacheByFieldPluralUniqueTrue.String())
			}

		}
		// 唯一 && 字段数大于1
		if v.Unique && len(v.Columns) > 1 {
			interfaceFindOneByFields, err := template.NewTemplate().Parse(InterfaceFindOneByFields).Execute(tplParams)
			if err != nil {
				return "", err
			}
			readMethods += fmt.Sprintln(interfaceFindOneByFields.String())

			interfaceFindOneCacheByFields, err := template.NewTemplate().Parse(InterfaceFindOneCacheByFields).Execute(tplParams)
			if err != nil {
				return "", err
			}
			readMethods += fmt.Sprintln(interfaceFindOneCacheByFields.String())

		}
		// 不唯一 && 字段数等于1
		if !v.Unique && len(v.Columns) == 1 {
			interfaceFindMultiByField, err := template.NewTemplate().Parse(InterfaceFindMultiByField).Execute(tplParams)
			if err != nil {
				return "", err
			}
			readMethods += fmt.Sprintln(interfaceFindMultiByField.String())

			interfaceFindMultiCacheByField, err := template.NewTemplate().Parse(InterfaceFindMultiCacheByField).Execute(tplParams)
			if err != nil {
				return "", err
			}
			readMethods += fmt.Sprintln(interfaceFindMultiCacheByField.String())

			columnNameToDataType := r.columnNameToDataType[v.Columns[0]]
			switch columnNameToDataType {
			case "bool":
			default:
				interfaceFindMultiByFieldPlural, err := template.NewTemplate().Parse(InterfaceFindMultiByFieldPlural).Execute(tplParams)
				if err != nil {
					return "", err
				}
				readMethods += fmt.Sprintln(interfaceFindMultiByFieldPlural.String())

				interfaceFindMultiCacheByFieldPluralUniqueFalse, err := template.NewTemplate().Parse(InterfaceFindMultiCacheByFieldPluralUniqueFalse).Execute(tplParams)
				if err != nil {
					return "", err
				}
				readMethods += fmt.Sprintln(interfaceFindMultiCacheByFieldPluralUniqueFalse.String())
			}
		}
		// 不唯一 && 字段数大于1
		if !v.Unique && len(v.Columns) > 1 {
			interfaceFindMultiByFields, err := template.NewTemplate().Parse(InterfaceFindMultiByFields).Execute(tplParams)
			if err != nil {
				return "", err
			}
			readMethods += fmt.Sprintln(interfaceFindMultiByFields.String())

			interfaceFindMultiCacheByFields, err := template.NewTemplate().Parse(InterfaceFindMultiCacheByFields).Execute(tplParams)
			if err != nil {
				return "", err
			}
			readMethods += fmt.Sprintln(interfaceFindMultiCacheByFields.String())
		}
	}
	interfaceFindMultiByCondition, err := template.NewTemplate().Parse(InterfaceFindMultiByCondition).Execute(map[string]any{
		"dbName":         r.dbName,
		"upperTableName": r.upperTableName,
		"lowerTableName": r.lowerTableName,
	})
	if err != nil {
		return "", err
	}
	readMethods += fmt.Sprintln(interfaceFindMultiByCondition.String())

	interfaceFindMultiByCacheCondition, err := template.NewTemplate().Parse(InterfaceFindMultiByCacheCondition).Execute(map[string]any{
		"dbName":         r.dbName,
		"upperTableName": r.upperTableName,
		"lowerTableName": r.lowerTableName,
	})
	if err != nil {
		return "", err
	}
	readMethods += fmt.Sprintln(interfaceFindMultiByCacheCondition.String())

	return readMethods, nil
}

// generateDelMethods
func (r *Repo) generateDelMethods() (string, error) {
	var delMethods string
	// 是否有索引
	haveIndex := len(r.index) > 0
	for _, v := range r.index {
		if r.checkDaoFieldType(v.Columns) {
			continue
		}
		tplParams := map[string]any{
			"firstTableChar":    r.firstTableChar,
			"dbName":            r.dbName,
			"upperTableName":    r.upperTableName,
			"lowerTableName":    r.lowerTableName,
			"upperField":        r.upperFieldName(v.Columns[0]),
			"lowerField":        r.lowerFieldName(v.Columns[0]),
			"upperFieldPlural":  r.plural(r.upperFieldName(v.Columns[0])),
			"lowerFieldPlural":  r.plural(r.lowerFieldName(v.Columns[0])),
			"dataType":          r.columnNameToDataType[v.Columns[0]],
			"upperFields":       r.upperFields(v.Columns),
			"fieldAndDataTypes": r.fieldAndDataTypes(v.Columns),
		}

		// 唯一 && 字段数于1
		if v.Unique && len(v.Columns) == 1 {
			switch r.columnNameToDataType[v.Columns[0]] {
			case "bool":
			default:
				interfaceDeleteOneByField, err := template.NewTemplate().Parse(InterfaceDeleteOneByField).Execute(tplParams)
				if err != nil {
					return "", err
				}
				delMethods += fmt.Sprintln(interfaceDeleteOneByField.String())

				interfaceDeleteOneCacheByField, err := template.NewTemplate().Parse(InterfaceDeleteOneCacheByField).Execute(tplParams)
				if err != nil {
					return "", err
				}
				delMethods += fmt.Sprintln(interfaceDeleteOneCacheByField.String())

				interfaceDeleteOneByFieldTx, err := template.NewTemplate().Parse(InterfaceDeleteOneByFieldTx).Execute(tplParams)
				if err != nil {
					return "", err
				}
				delMethods += fmt.Sprintln(interfaceDeleteOneByFieldTx.String())

				interfaceDeleteOneCacheByFieldTx, err := template.NewTemplate().Parse(InterfaceDeleteOneCacheByFieldTx).Execute(tplParams)
				if err != nil {
					return "", err
				}
				delMethods += fmt.Sprintln(interfaceDeleteOneCacheByFieldTx.String())

				interfaceDeleteMultiByFieldPlural, err := template.NewTemplate().Parse(InterfaceDeleteMultiByFieldPlural).Execute(tplParams)
				if err != nil {
					return "", err
				}
				delMethods += fmt.Sprintln(interfaceDeleteMultiByFieldPlural.String())

				interfaceDeleteMultiCacheByFieldPlural, err := template.NewTemplate().Parse(InterfaceDeleteMultiCacheByFieldPlural).Execute(tplParams)
				if err != nil {
					return "", err
				}
				delMethods += fmt.Sprintln(interfaceDeleteMultiCacheByFieldPlural.String())

				interfaceDeleteMultiByFieldPluralTx, err := template.NewTemplate().Parse(InterfaceDeleteMultiByFieldPluralTx).Execute(tplParams)
				if err != nil {
					return "", err
				}
				delMethods += fmt.Sprintln(interfaceDeleteMultiByFieldPluralTx.String())
				interfaceDeleteMultiCacheByFieldPluralTx, err := template.NewTemplate().Parse(InterfaceDeleteMultiCacheByFieldPluralTx).Execute(tplParams)
				if err != nil {
					return "", err
				}
				delMethods += fmt.Sprintln(interfaceDeleteMultiCacheByFieldPluralTx.String())
			}
		}
		// 唯一 && 字段数大于1
		if v.Unique && len(v.Columns) > 1 {
			interfaceDeleteOneByFields, err := template.NewTemplate().Parse(InterfaceDeleteOneByFields).Execute(tplParams)
			if err != nil {
				return "", err
			}
			delMethods += fmt.Sprintln(interfaceDeleteOneByFields.String())

			interfaceDeleteOneCacheByFields, err := template.NewTemplate().Parse(InterfaceDeleteOneCacheByFields).Execute(tplParams)
			if err != nil {
				return "", err
			}
			delMethods += fmt.Sprintln(interfaceDeleteOneCacheByFields.String())

			interfaceDeleteOneByFieldsTx, err := template.NewTemplate().Parse(InterfaceDeleteOneByFieldsTx).Execute(tplParams)
			if err != nil {
				return "", err
			}
			delMethods += fmt.Sprintln(interfaceDeleteOneByFieldsTx.String())

			interfaceDeleteOneCacheByFieldsTx, err := template.NewTemplate().Parse(InterfaceDeleteOneCacheByFieldsTx).Execute(tplParams)
			if err != nil {
				return "", err
			}
			delMethods += fmt.Sprintln(interfaceDeleteOneCacheByFieldsTx.String())
		}
		// 不唯一 && 字段数等于1
		if !v.Unique && len(v.Columns) == 1 {
			switch r.columnNameToDataType[v.Columns[0]] {
			case "bool":
			default:
				interfaceDeleteMultiByField, err := template.NewTemplate().Parse(InterfaceDeleteMultiByField).Execute(tplParams)
				if err != nil {
					return "", err
				}
				delMethods += fmt.Sprintln(interfaceDeleteMultiByField.String())

				interfaceDeleteMultiCacheByField, err := template.NewTemplate().Parse(InterfaceDeleteMultiCacheByField).Execute(tplParams)
				if err != nil {
					return "", err
				}
				delMethods += fmt.Sprintln(interfaceDeleteMultiCacheByField.String())

				interfaceDeleteMultiByFieldTx, err := template.NewTemplate().Parse(InterfaceDeleteMultiByFieldTx).Execute(tplParams)
				if err != nil {
					return "", err
				}
				delMethods += fmt.Sprintln(interfaceDeleteMultiByFieldTx.String())

				interfaceDeleteMultiCacheByFieldTx, err := template.NewTemplate().Parse(InterfaceDeleteMultiCacheByFieldTx).Execute(tplParams)
				if err != nil {
					return "", err
				}
				delMethods += fmt.Sprintln(interfaceDeleteMultiCacheByFieldTx.String())

				interfaceDeleteMultiByFieldPlural, err := template.NewTemplate().Parse(InterfaceDeleteMultiByFieldPlural).Execute(tplParams)
				if err != nil {
					return "", err
				}
				delMethods += fmt.Sprintln(interfaceDeleteMultiByFieldPlural.String())

				interfaceDeleteMultiCacheByFieldPlural, err := template.NewTemplate().Parse(InterfaceDeleteMultiCacheByFieldPlural).Execute(tplParams)
				if err != nil {
					return "", err
				}
				delMethods += fmt.Sprintln(interfaceDeleteMultiCacheByFieldPlural.String())

				interfaceDeleteMultiByFieldPluralTx, err := template.NewTemplate().Parse(InterfaceDeleteMultiByFieldPluralTx).Execute(tplParams)
				if err != nil {
					return "", err
				}
				delMethods += fmt.Sprintln(interfaceDeleteMultiByFieldPluralTx.String())
				interfaceDeleteMultiCacheByFieldPluralTx, err := template.NewTemplate().Parse(InterfaceDeleteMultiCacheByFieldPluralTx).Execute(tplParams)
				if err != nil {
					return "", err
				}
				delMethods += fmt.Sprintln(interfaceDeleteMultiCacheByFieldPluralTx.String())
			}
		}
		// 不唯一 && 字段数大于1
		if !v.Unique && len(v.Columns) > 1 {
			interfaceDeleteMultiByFields, err := template.NewTemplate().Parse(InterfaceDeleteMultiByFields).Execute(tplParams)
			if err != nil {
				return "", err
			}
			delMethods += fmt.Sprintln(interfaceDeleteMultiByFields.String())

			interfaceDeleteMultiCacheByFields, err := template.NewTemplate().Parse(InterfaceDeleteMultiCacheByFields).Execute(tplParams)
			if err != nil {
				return "", err
			}
			delMethods += fmt.Sprintln(interfaceDeleteMultiCacheByFields.String())

			interfaceDeleteMultiByFieldsTx, err := template.NewTemplate().Parse(InterfaceDeleteMultiByFieldsTx).Execute(tplParams)
			if err != nil {
				return "", err
			}
			delMethods += fmt.Sprintln(interfaceDeleteMultiByFieldsTx.String())

			interfaceDeleteMultiCacheByFieldsTx, err := template.NewTemplate().Parse(InterfaceDeleteMultiCacheByFieldsTx).Execute(tplParams)
			if err != nil {
				return "", err
			}
			delMethods += fmt.Sprintln(interfaceDeleteMultiCacheByFieldsTx.String())
		}
	}
	// 有唯一索引
	if haveIndex {
		interfaceDeleteIndexCache, err := template.NewTemplate().Parse(InterfaceDeleteIndexCache).Execute(map[string]any{
			"dbName":         r.dbName,
			"upperTableName": r.upperTableName,
		})
		if err != nil {
			return "", err
		}
		delMethods += fmt.Sprintln(interfaceDeleteIndexCache.String())
	}
	return delMethods, nil
}

// generateTypes
func (r *Repo) generateTypes() (string, error) {
	var methods string
	commonMethods, err := r.generateCommonMethods()
	if err != nil {
		return "", err
	}
	createMethods, err := r.generateCreateMethods()
	if err != nil {
		return "", err
	}
	updateMethods, err := r.generateUpdateMethods()
	if err != nil {
		return "", err
	}
	readMethods, err := r.generateReadMethods()
	if err != nil {
		return "", err
	}
	delMethods, err := r.generateDelMethods()
	if err != nil {
		return "", err
	}
	methods += commonMethods
	methods += createMethods
	methods += updateMethods
	methods += readMethods
	methods += delMethods
	typesTpl, err := template.NewTemplate().Parse(Types).Execute(map[string]any{
		"dbName":         r.dbName,
		"upperTableName": r.upperTableName,
		"lowerTableName": r.lowerTableName,
		"methods":        methods,
	})
	if err != nil {
		return "", err
	}
	return typesTpl.String(), nil
}

// generateNew
func (r *Repo) generateNew() (string, error) {
	newTpl, err := template.NewTemplate().Parse(New).Execute(map[string]any{
		"dbName":         r.dbName,
		"upperTableName": r.upperTableName,
		"lowerTableName": r.lowerTableName,
	})
	if err != nil {
		return "", err
	}
	return newTpl.String(), nil
}

// generateCommonFunc
func (r *Repo) generateCommonFunc() (string, error) {
	var commonFunc string
	tplParams := map[string]any{
		"firstTableChar": r.firstTableChar,
		"dbName":         r.dbName,
		"upperTableName": r.upperTableName,
		"lowerTableName": r.lowerTableName,
	}
	deepCopy, err := template.NewTemplate().Parse(DeepCopy).Execute(tplParams)
	if err != nil {
		return "", err
	}
	commonFunc += fmt.Sprintln(deepCopy.String())
	return commonFunc, nil
}

// generateCreateFunc
func (r *Repo) generateCreateFunc() (string, error) {
	// 是否有索引
	haveIndex := len(r.index) > 0
	// 是否有主键索引
	havePrimaryKey := false
	singlePrimaryKeyField := ""
	for _, v := range r.index {
		if v.PrimaryKey {
			havePrimaryKey = true
			if len(v.Columns) == 1 {
				singlePrimaryKeyField = v.Columns[0]
			}
			break
		}
	}
	var createFunc string
	tplParams := map[string]any{
		"firstTableChar":        r.firstTableChar,
		"dbName":                r.dbName,
		"upperTableName":        r.upperTableName,
		"lowerTableName":        r.lowerTableName,
		"upperSinglePrimaryKey": r.upperFieldName(singlePrimaryKeyField),
	}
	createOne, err := template.NewTemplate().Parse(CreateOne).Execute(tplParams)
	if err != nil {
		return "", err
	}
	createFunc += fmt.Sprintln(createOne.String())
	if haveIndex {
		createOneCache, err := template.NewTemplate().Parse(CreateOneCache).Execute(tplParams)
		if err != nil {
			return "", err
		}
		createFunc += fmt.Sprintln(createOneCache.String())
	}
	createOneByTx, err := template.NewTemplate().Parse(CreateOneByTx).Execute(tplParams)
	if err != nil {
		return "", err
	}
	createFunc += fmt.Sprintln(createOneByTx.String())

	if haveIndex {
		createOneCacheByTx, err := template.NewTemplate().Parse(CreateOneCacheByTx).Execute(tplParams)
		if err != nil {
			return "", err
		}
		createFunc += fmt.Sprintln(createOneCacheByTx.String())
	}
	createBatch, err := template.NewTemplate().Parse(CreateBatch).Execute(tplParams)
	if err != nil {
		return "", err
	}
	createFunc += fmt.Sprintln(createBatch.String())
	if haveIndex {
		createBatchCache, err := template.NewTemplate().Parse(CreateBatchCache).Execute(tplParams)
		if err != nil {
			return "", err
		}
		createFunc += fmt.Sprintln(createBatchCache.String())
	}
	createBatchByTx, err := template.NewTemplate().Parse(CreateBatchByTx).Execute(tplParams)
	if err != nil {
		return "", err
	}
	createFunc += fmt.Sprintln(createBatchByTx.String())
	if haveIndex {
		createBatchCacheByTx, err := template.NewTemplate().Parse(CreateBatchCacheByTx).Execute(tplParams)
		if err != nil {
			return "", err
		}
		createFunc += fmt.Sprintln(createBatchCacheByTx.String())
	}

	if havePrimaryKey {
		upsertOne, err := template.NewTemplate().Parse(UpsertOne).Execute(tplParams)
		if err != nil {
			return "", err
		}
		createFunc += fmt.Sprintln(upsertOne.String())
		upsertOneCache, err := template.NewTemplate().Parse(UpsertOneCache).Execute(tplParams)
		if err != nil {
			return "", err
		}
		createFunc += fmt.Sprintln(upsertOneCache.String())

		upsertOneByTx, err := template.NewTemplate().Parse(UpsertOneByTx).Execute(tplParams)
		if err != nil {
			return "", err
		}
		createFunc += fmt.Sprintln(upsertOneByTx.String())
		upsertOneCacheByTx, err := template.NewTemplate().Parse(UpsertOneCacheByTx).Execute(tplParams)
		if err != nil {
			return "", err
		}
		createFunc += fmt.Sprintln(upsertOneCacheByTx.String())
	}
	upsertOneByFields, err := template.NewTemplate().Parse(UpsertOneByFields).Execute(tplParams)
	if err != nil {
		return "", err
	}
	createFunc += fmt.Sprintln(upsertOneByFields.String())

	if haveIndex {
		upsertOneCacheByFields, err := template.NewTemplate().Parse(UpsertOneCacheByFields).Execute(tplParams)
		if err != nil {
			return "", err
		}
		createFunc += fmt.Sprintln(upsertOneCacheByFields.String())
	}

	upsertOneByFieldsTx, err := template.NewTemplate().Parse(UpsertOneByFieldsTx).Execute(tplParams)
	if err != nil {
		return "", err
	}
	createFunc += fmt.Sprintln(upsertOneByFieldsTx.String())

	if haveIndex {
		upsertOneCacheByFieldsTx, err := template.NewTemplate().Parse(UpsertOneCacheByFieldsTx).Execute(tplParams)
		if err != nil {
			return "", err
		}
		createFunc += fmt.Sprintln(upsertOneCacheByFieldsTx.String())
	}
	return createFunc, nil
}

// generateUpdateFunc
func (r *Repo) generateUpdateFunc() (string, error) {
	// 是否有主键索引
	havePrimaryKey := false
	singlePrimaryKeyField := ""
	for _, v := range r.index {
		if v.PrimaryKey {
			havePrimaryKey = true
			if len(v.Columns) == 1 {
				singlePrimaryKeyField = v.Columns[0]
			}
			break
		}
	}
	// 没有主键索引,不生成更新方法
	if !havePrimaryKey {
		return "", nil
	}
	var updateFunc string
	//参数
	tplParams := map[string]any{
		"firstTableChar":              r.firstTableChar,
		"dbName":                      r.dbName,
		"upperTableName":              r.upperTableName,
		"lowerTableName":              r.lowerTableName,
		"dataTypeSinglePrimaryKey":    r.columnNameToDataType[singlePrimaryKeyField],
		"upperSinglePrimaryKey":       r.upperFieldName(singlePrimaryKeyField),
		"upperSinglePrimaryKeyPlural": r.plural(r.upperFieldName(singlePrimaryKeyField)),
		"lowerSinglePrimaryKeyPlural": r.plural(r.lowerFieldName(singlePrimaryKeyField)),
	}
	updateOneTpl, err := template.NewTemplate().Parse(UpdateOne).Execute(tplParams)
	if err != nil {
		return "", err
	}
	updateFunc += fmt.Sprintln(updateOneTpl.String())

	updateOneCacheTpl, err := template.NewTemplate().Parse(UpdateOneCache).Execute(tplParams)
	if err != nil {
		return "", err
	}
	updateFunc += fmt.Sprintln(updateOneCacheTpl.String())

	updateOneByTx, err := template.NewTemplate().Parse(UpdateOneByTx).Execute(tplParams)
	if err != nil {
		return "", err
	}
	updateFunc += fmt.Sprintln(updateOneByTx.String())
	updateOneCacheByTxTpl, err := template.NewTemplate().Parse(UpdateOneCacheByTx).Execute(tplParams)
	if err != nil {
		return "", err
	}
	updateFunc += fmt.Sprintln(updateOneCacheByTxTpl.String())

	updateOneWithZero, err := template.NewTemplate().Parse(UpdateOneWithZero).Execute(tplParams)
	if err != nil {
		return "", err
	}
	updateFunc += fmt.Sprintln(updateOneWithZero.String())

	updateOneCacheWithZero, err := template.NewTemplate().Parse(UpdateOneCacheWithZero).Execute(tplParams)
	if err != nil {
		return "", err
	}
	updateFunc += fmt.Sprintln(updateOneCacheWithZero.String())

	updateOneWithZeroByTx, err := template.NewTemplate().Parse(UpdateOneWithZeroByTx).Execute(tplParams)
	if err != nil {
		return "", err
	}
	updateFunc += fmt.Sprintln(updateOneWithZeroByTx.String())

	updateOneCacheWithZeroByTx, err := template.NewTemplate().Parse(UpdateOneCacheWithZeroByTx).Execute(tplParams)
	if err != nil {
		return "", err
	}
	updateFunc += fmt.Sprintln(updateOneCacheWithZeroByTx.String())

	updateBatchByPrimaryKeys, err := template.NewTemplate().Parse(UpdateBatchByPrimaryKeys).Execute(tplParams)
	if err != nil {
		return "", err
	}
	updateFunc += fmt.Sprintln(updateBatchByPrimaryKeys.String())

	updateBatchByPrimaryKeysTx, err := template.NewTemplate().Parse(UpdateBatchByPrimaryKeysTx).Execute(tplParams)
	if err != nil {
		return "", err
	}
	updateFunc += fmt.Sprintln(updateBatchByPrimaryKeysTx.String())

	return updateFunc, nil
}

// generateReadFunc
func (r *Repo) generateReadFunc() (string, error) {
	var readFunc string
	for _, v := range r.index {
		if r.checkDaoFieldType(v.Columns) {
			continue
		}
		//参数
		tplParams := map[string]any{
			"firstTableChar":    r.firstTableChar,
			"dbName":            r.dbName,
			"upperTableName":    r.upperTableName,
			"lowerTableName":    r.lowerTableName,
			"upperField":        r.upperFieldName(v.Columns[0]),
			"lowerField":        r.lowerFieldName(v.Columns[0]),
			"upperFieldPlural":  r.plural(r.upperFieldName(v.Columns[0])),
			"lowerFieldPlural":  r.plural(r.lowerFieldName(v.Columns[0])),
			"dataType":          r.columnNameToDataType[v.Columns[0]],
			"upperFields":       r.upperFields(v.Columns),
			"fieldAndDataTypes": r.fieldAndDataTypes(v.Columns),
			"cacheFields":       r.cacheFields(v.Columns),
			"cacheFieldsJoin":   r.cacheFieldsJoin(v.Columns),
		}
		// 唯一 && 字段数于1
		if v.Unique && len(v.Columns) == 1 {
			columnNameToDataType := r.columnNameToDataType[v.Columns[0]]
			switch columnNameToDataType {
			case "bool":
				tplParams["whereFields"] = fmt.Sprintf("dao.%s.Is(%s)", r.upperFieldName(v.Columns[0]), r.lowerFieldName(v.Columns[0]))

				findOneByField, err := template.NewTemplate().Parse(FindOneByField).Execute(tplParams)
				if err != nil {
					return "", err
				}
				readFunc += fmt.Sprintln(findOneByField.String())

				findOneCacheByField, err := template.NewTemplate().Parse(FindOneCacheByField).Execute(tplParams)
				if err != nil {
					return "", err
				}
				readFunc += fmt.Sprintln(findOneCacheByField.String())
			default:
				tplParams["whereFields"] = fmt.Sprintf("dao.%s.Eq(%s)", r.upperFieldName(v.Columns[0]), r.lowerFieldName(v.Columns[0]))

				findOneByField, err := template.NewTemplate().Parse(FindOneByField).Execute(tplParams)
				if err != nil {
					return "", err
				}
				readFunc += fmt.Sprintln(findOneByField.String())

				findOneCacheByField, err := template.NewTemplate().Parse(FindOneCacheByField).Execute(tplParams)
				if err != nil {
					return "", err
				}
				readFunc += fmt.Sprintln(findOneCacheByField.String())

				findMultiByFieldPlural, err := template.NewTemplate().Parse(FindMultiByFieldPlural).Execute(tplParams)
				if err != nil {
					return "", err
				}
				readFunc += fmt.Sprintln(findMultiByFieldPlural.String())

				findMultiCacheByFieldPluralUniqueTrue, err := template.NewTemplate().Parse(FindMultiCacheByFieldPluralUniqueTrue).Execute(tplParams)
				if err != nil {
					return "", err
				}
				readFunc += fmt.Sprintln(findMultiCacheByFieldPluralUniqueTrue.String())
			}

		}
		// 唯一 && 字段数大于1
		if v.Unique && len(v.Columns) > 1 {
			var whereFields string
			for _, v := range v.Columns {
				switch r.columnNameToDataType[v] {
				case "bool":
					whereFields += fmt.Sprintf("dao.%s.Is(%s),", r.upperFieldName(v), r.lowerFieldName(v))
				default:
					whereFields += fmt.Sprintf("dao.%s.Eq(%s),", r.upperFieldName(v), r.lowerFieldName(v))
				}
			}
			tplParams["whereFields"] = strings.TrimRight(whereFields, ",")
			findOneByFields, err := template.NewTemplate().Parse(FindOneByFields).Execute(tplParams)
			if err != nil {
				return "", err
			}
			readFunc += fmt.Sprintln(findOneByFields.String())

			findOneCacheByFields, err := template.NewTemplate().Parse(FindOneCacheByFields).Execute(tplParams)
			if err != nil {
				return "", err
			}
			readFunc += fmt.Sprintln(findOneCacheByFields.String())
		}
		// 不唯一 && 字段数等于1
		if !v.Unique && len(v.Columns) == 1 {
			var whereField string
			columnNameToDataType := r.columnNameToDataType[v.Columns[0]]
			switch columnNameToDataType {
			case "bool":
				whereField += fmt.Sprintf("dao.%s.Is(%s)", r.upperFieldName(v.Columns[0]), r.lowerFieldName(v.Columns[0]))
			default:
				whereField += fmt.Sprintf("dao.%s.Eq(%s)", r.upperFieldName(v.Columns[0]), r.lowerFieldName(v.Columns[0]))
			}
			tplParams["whereFields"] = whereField

			findMultiByField, err := template.NewTemplate().Parse(FindMultiByField).Execute(tplParams)
			if err != nil {
				return "", err
			}
			readFunc += fmt.Sprintln(findMultiByField.String())

			findMultiCacheByField, err := template.NewTemplate().Parse(FindMultiCacheByField).Execute(tplParams)
			if err != nil {
				return "", err
			}
			readFunc += fmt.Sprintln(findMultiCacheByField.String())

			switch columnNameToDataType {
			case "bool":
			default:
				findMultiByFieldPlural, err := template.NewTemplate().Parse(FindMultiByFieldPlural).Execute(tplParams)
				if err != nil {
					return "", err
				}
				readFunc += fmt.Sprintln(findMultiByFieldPlural.String())

				findMultiCacheByFieldPluralUniqueFalse, err := template.NewTemplate().Parse(FindMultiCacheByFieldPluralUniqueFalse).Execute(tplParams)
				if err != nil {
					return "", err
				}
				readFunc += fmt.Sprintln(findMultiCacheByFieldPluralUniqueFalse.String())
			}
		}
		// 不唯一 && 字段数大于1
		if !v.Unique && len(v.Columns) > 1 {
			var whereFields string
			for _, v := range v.Columns {
				switch r.columnNameToDataType[v] {
				case "bool":
					whereFields += fmt.Sprintf("dao.%s.Is(%s),", r.upperFieldName(v), r.lowerFieldName(v))
				default:
					whereFields += fmt.Sprintf("dao.%s.Eq(%s),", r.upperFieldName(v), r.lowerFieldName(v))
				}
			}
			tplParams["whereFields"] = strings.TrimRight(whereFields, ",")
			findMultiByFields, err := template.NewTemplate().Parse(FindMultiByFields).Execute(tplParams)
			if err != nil {
				return "", err
			}
			readFunc += fmt.Sprintln(findMultiByFields.String())
			findMultiCacheByFields, err := template.NewTemplate().Parse(FindMultiCacheByFields).Execute(tplParams)
			if err != nil {
				return "", err
			}
			readFunc += fmt.Sprintln(findMultiCacheByFields.String())
		}
	}
	findMultiByCondition, err := template.NewTemplate().Parse(FindMultiByCondition).Execute(map[string]any{
		"firstTableChar": r.firstTableChar,
		"dbName":         r.dbName,
		"upperTableName": r.upperTableName,
		"lowerTableName": r.lowerTableName,
	})
	if err != nil {
		return "", err
	}
	readFunc += fmt.Sprintln(findMultiByCondition.String())
	findMultiByCacheCondition, err := template.NewTemplate().Parse(FindMultiByCacheCondition).Execute(map[string]any{
		"firstTableChar": r.firstTableChar,
		"dbName":         r.dbName,
		"upperTableName": r.upperTableName,
		"lowerTableName": r.lowerTableName,
	})
	if err != nil {
		return "", err
	}
	readFunc += fmt.Sprintln(findMultiByCacheCondition.String())
	return readFunc, nil
}

// generateDelFunc
func (r *Repo) generateDelFunc() (string, error) {
	haveIndex := len(r.index) > 0

	var delMethods string
	var cacheDelKeys string
	for _, v := range r.index {
		if r.checkDaoFieldType(v.Columns) {
			continue
		}
		cacheFieldsJoinSli := make([]string, 0)
		for _, column := range v.Columns {
			cacheFieldsJoinSli = append(cacheFieldsJoinSli, fmt.Sprintf("v.%s", r.upperFieldName(column)))
		}
		// 缓存删除key
		varCacheDelKeyTpl, err := template.NewTemplate().Parse(VarCacheDelKey).Execute(map[string]any{
			"firstTableChar":      r.firstTableChar,
			"upperTableName":      r.upperTableName,
			"cacheFields":         r.cacheFields(v.Columns),
			"delCacheFieldsParam": strings.Join(cacheFieldsJoinSli, ","),
		})
		if err != nil {
			return "", err
		}
		cacheDelKeys += fmt.Sprintln(varCacheDelKeyTpl.String())

		tplParams := map[string]any{
			"firstTableChar":    r.firstTableChar,
			"dbName":            r.dbName,
			"upperTableName":    r.upperTableName,
			"lowerTableName":    r.lowerTableName,
			"upperField":        r.upperFieldName(v.Columns[0]),
			"lowerField":        r.lowerFieldName(v.Columns[0]),
			"upperFieldPlural":  r.plural(r.upperFieldName(v.Columns[0])),
			"lowerFieldPlural":  r.plural(r.lowerFieldName(v.Columns[0])),
			"dataType":          r.columnNameToDataType[v.Columns[0]],
			"upperFields":       r.upperFields(v.Columns),
			"fieldAndDataTypes": r.fieldAndDataTypes(v.Columns),
		}

		// 唯一 && 字段数于1
		if v.Unique && len(v.Columns) == 1 {
			columnNameToDataType := r.columnNameToDataType[v.Columns[0]]
			switch columnNameToDataType {
			case "bool":
			default:
				deleteOneByField, err := template.NewTemplate().Parse(DeleteOneByField).Execute(tplParams)
				if err != nil {
					return "", err
				}
				delMethods += fmt.Sprintln(deleteOneByField.String())
				deleteOneCacheByField, err := template.NewTemplate().Parse(DeleteOneCacheByField).Execute(tplParams)
				if err != nil {
					return "", err
				}
				delMethods += fmt.Sprintln(deleteOneCacheByField.String())
				deleteOneByFieldTx, err := template.NewTemplate().Parse(DeleteOneByFieldTx).Execute(tplParams)
				if err != nil {
					return "", err
				}
				delMethods += fmt.Sprintln(deleteOneByFieldTx.String())
				deleteOneCacheByFieldTx, err := template.NewTemplate().Parse(DeleteOneCacheByFieldTx).Execute(tplParams)
				if err != nil {
					return "", err
				}
				delMethods += fmt.Sprintln(deleteOneCacheByFieldTx.String())
				deleteMultiByFieldPlural, err := template.NewTemplate().Parse(DeleteMultiByFieldPlural).Execute(tplParams)
				if err != nil {
					return "", err
				}
				delMethods += fmt.Sprintln(deleteMultiByFieldPlural.String())
				deleteMultiCacheByFieldPlural, err := template.NewTemplate().Parse(DeleteMultiCacheByFieldPlural).Execute(tplParams)
				if err != nil {
					return "", err
				}
				delMethods += fmt.Sprintln(deleteMultiCacheByFieldPlural.String())
				deleteMultiByFieldPluralTx, err := template.NewTemplate().Parse(DeleteMultiByFieldPluralTx).Execute(tplParams)
				if err != nil {
					return "", err
				}
				delMethods += fmt.Sprintln(deleteMultiByFieldPluralTx.String())
				deleteMultiCacheByFieldPluralTx, err := template.NewTemplate().Parse(DeleteMultiCacheByFieldPluralTx).Execute(tplParams)
				if err != nil {
					return "", err
				}
				delMethods += fmt.Sprintln(deleteMultiCacheByFieldPluralTx.String())
			}
		}
		// 唯一 && 字段数大于1
		if v.Unique && len(v.Columns) > 1 {
			var whereFields string
			for _, v := range v.Columns {
				switch r.columnNameToDataType[v] {
				case "bool":
					whereFields += fmt.Sprintf("dao.%s.Is(%s),", r.upperFieldName(v), r.lowerFieldName(v))
				default:
					whereFields += fmt.Sprintf("dao.%s.Eq(%s),", r.upperFieldName(v), r.lowerFieldName(v))
				}
			}
			tplParams["whereFields"] = strings.TrimRight(whereFields, ",")
			deleteOneByFields, err := template.NewTemplate().Parse(DeleteOneByFields).Execute(tplParams)
			if err != nil {
				return "", err
			}
			delMethods += fmt.Sprintln(deleteOneByFields.String())
			deleteOneCacheByFields, err := template.NewTemplate().Parse(DeleteOneCacheByFields).Execute(tplParams)
			if err != nil {
				return "", err
			}
			delMethods += fmt.Sprintln(deleteOneCacheByFields.String())
			deleteOneByFieldsTx, err := template.NewTemplate().Parse(DeleteOneByFieldsTx).Execute(tplParams)
			if err != nil {
				return "", err
			}
			delMethods += fmt.Sprintln(deleteOneByFieldsTx.String())
			deleteOneCacheByFieldsTx, err := template.NewTemplate().Parse(DeleteOneCacheByFieldsTx).Execute(tplParams)
			if err != nil {
				return "", err
			}
			delMethods += fmt.Sprintln(deleteOneCacheByFieldsTx.String())
		}
		// 不唯一 && 字段数等于1
		if !v.Unique && len(v.Columns) == 1 {
			switch r.columnNameToDataType[v.Columns[0]] {
			case "bool":
			default:
				deleteMultiByField, err := template.NewTemplate().Parse(DeleteMultiByField).Execute(tplParams)
				if err != nil {
					return "", err
				}
				delMethods += fmt.Sprintln(deleteMultiByField.String())

				deleteMultiCacheByField, err := template.NewTemplate().Parse(DeleteMultiCacheByField).Execute(tplParams)
				if err != nil {
					return "", err
				}
				delMethods += fmt.Sprintln(deleteMultiCacheByField.String())

				deleteMultiByFieldTx, err := template.NewTemplate().Parse(DeleteMultiByFieldTx).Execute(tplParams)
				if err != nil {
					return "", err
				}
				delMethods += fmt.Sprintln(deleteMultiByFieldTx.String())

				deleteMultiCacheByFieldTx, err := template.NewTemplate().Parse(DeleteMultiCacheByFieldTx).Execute(tplParams)
				if err != nil {
					return "", err
				}
				delMethods += fmt.Sprintln(deleteMultiCacheByFieldTx.String())

				deleteMultiByFieldPlural, err := template.NewTemplate().Parse(DeleteMultiByFieldPlural).Execute(tplParams)
				if err != nil {
					return "", err
				}
				delMethods += fmt.Sprintln(deleteMultiByFieldPlural.String())

				deleteMultiCacheByFieldPlural, err := template.NewTemplate().Parse(DeleteMultiCacheByFieldPlural).Execute(tplParams)
				if err != nil {
					return "", err
				}
				delMethods += fmt.Sprintln(deleteMultiCacheByFieldPlural.String())

				deleteMultiByFieldPluralTx, err := template.NewTemplate().Parse(DeleteMultiByFieldPluralTx).Execute(tplParams)
				if err != nil {
					return "", err
				}
				delMethods += fmt.Sprintln(deleteMultiByFieldPluralTx.String())

				deleteMultiCacheByFieldPluralTx, err := template.NewTemplate().Parse(DeleteMultiCacheByFieldPluralTx).Execute(tplParams)
				if err != nil {
					return "", err
				}
				delMethods += fmt.Sprintln(deleteMultiCacheByFieldPluralTx.String())
			}
		}
		// 不唯一 && 字段数大于1
		if !v.Unique && len(v.Columns) > 1 {
			var whereFields string
			for _, v := range v.Columns {
				switch r.columnNameToDataType[v] {
				case "bool":
					whereFields += fmt.Sprintf("dao.%s.Is(%s),", r.upperFieldName(v), r.lowerFieldName(v))
				default:
					whereFields += fmt.Sprintf("dao.%s.Eq(%s),", r.upperFieldName(v), r.lowerFieldName(v))
				}
			}
			tplParams["whereFields"] = strings.TrimRight(whereFields, ",")
			deleteMultiByFields, err := template.NewTemplate().Parse(DeleteMultiByFields).Execute(tplParams)
			if err != nil {
				return "", err
			}
			delMethods += fmt.Sprintln(deleteMultiByFields.String())

			deleteMultiCacheByFields, err := template.NewTemplate().Parse(DeleteMultiCacheByFields).Execute(tplParams)
			if err != nil {
				return "", err
			}
			delMethods += fmt.Sprintln(deleteMultiCacheByFields.String())

			deleteMultiByFieldsTx, err := template.NewTemplate().Parse(DeleteMultiByFieldsTx).Execute(tplParams)
			if err != nil {
				return "", err
			}
			delMethods += fmt.Sprintln(deleteMultiByFieldsTx.String())

			deleteMultiCacheByFieldsTx, err := template.NewTemplate().Parse(DeleteMultiCacheByFieldsTx).Execute(tplParams)
			if err != nil {
				return "", err
			}
			delMethods += fmt.Sprintln(deleteMultiCacheByFieldsTx.String())
		}
	}
	// cacheDelKeys 去掉最后一个换行符
	cacheDelKeys = strings.TrimRight(cacheDelKeys, "\n")
	// 有唯一索引
	if haveIndex {
		deleteUniqueIndexCacheTpl, err := template.NewTemplate().Parse(DeleteIndexCache).Execute(map[string]any{
			"firstTableChar": r.firstTableChar,
			"dbName":         r.dbName,
			"upperTableName": r.upperTableName,
			"lowerTableName": r.lowerTableName,
			"cacheDelKeys":   cacheDelKeys,
		})
		if err != nil {
			return "", err
		}
		delMethods += fmt.Sprintln(deleteUniqueIndexCacheTpl.String())
	}
	return delMethods, nil
}

func (r *Repo) upperFields(columns []string) string {
	var upperFields string
	for _, v := range columns {
		upperFields += r.upperFieldName(v)
	}
	return upperFields
}

func (r *Repo) fieldAndDataTypes(columns []string) string {
	var fieldAndDataTypes string
	for _, v := range columns {
		fieldAndDataTypes += fmt.Sprintf("%s %s,", r.lowerFieldName(v), r.columnNameToDataType[v])
	}
	return strings.Trim(fieldAndDataTypes, ",")
}

func (r *Repo) cacheFields(columns []string) string {
	var cacheFields string
	for _, v := range columns {
		cacheFields += r.upperFieldName(v)
	}
	return cacheFields
}

func (r *Repo) cacheFieldsJoin(columns []string) string {
	var cacheFieldsJoin string
	for _, v := range columns {
		cacheFieldsJoin += fmt.Sprintf("%s,", r.lowerFieldName(v))
	}
	return strings.Trim(cacheFieldsJoin, ",")
}

// upperFieldName 字段名称大写
func (r *Repo) upperFieldName(s string) string {
	return r.columnNameToName[s]
}

// lowerFieldName 字段名称小写
func (r *Repo) lowerFieldName(s string) string {
	str := r.upperFieldName(s)
	if str == "" {
		return str
	}
	words := []string{"API", "ASCII", "CPU", "CSS", "DNS", "EOF", "GUID", "HTML", "HTTP", "HTTPS", "ID", "IP", "JSON", "LHS", "QPS", "RAM", "RHS", "RPC", "SLA", "SMTP", "SSH", "TLS", "ttl", "UID", "UI", "UUID", "URI", "URL", "UTF8", "VM", "XML", "XSRF", "XSS"}
	// 如果第一个单词命中  则不处理
	for _, v := range words {
		if strings.HasPrefix(str, v) {
			return str
		}
	}
	rs := []rune(str)
	f := rs[0]
	if 'A' <= f && f <= 'Z' {
		str = string(unicode.ToLower(f)) + string(rs[1:])
	}
	if token.Lookup(str).IsKeyword() || utils.StrSliFind(KeyWords, str) {
		str = "_" + str
	}
	return str
}

// upperName 大写
func (r *Repo) upperName(s string) string {
	return r.gorm.NamingStrategy.SchemaName(s)
}

// lowerName 小写
func (r *Repo) lowerName(s string) string {
	str := r.upperName(s)
	if str == "" {
		return str
	}
	words := []string{"API", "ASCII", "CPU", "CSS", "DNS", "EOF", "GUID", "HTML", "HTTP", "HTTPS", "ID", "IP", "JSON", "LHS", "QPS", "RAM", "RHS", "RPC", "SLA", "SMTP", "SSH", "TLS", "ttl", "UID", "UI", "UUID", "URI", "URL", "UTF8", "VM", "XML", "XSRF", "XSS"}
	// 如果第一个单词命中  则不处理
	for _, v := range words {
		if strings.HasPrefix(str, v) {
			return str
		}
	}
	rs := []rune(str)
	f := rs[0]
	if 'A' <= f && f <= 'Z' {
		str = string(unicode.ToLower(f)) + string(rs[1:])
	}
	return str
}

// plural 复数形式
func (r *Repo) plural(s string) string {
	str := inflection.Plural(s)
	if str == s {
		str += "plural"
	}
	return str
}

// checkDaoFieldType  检查字段是否是 dao 中的 Field类型
func (r *Repo) checkDaoFieldType(s []string) bool {
	for _, v := range s {
		if r.columnNameToFieldType[v] == "Field" {
			return true
		}
	}
	return false
}
