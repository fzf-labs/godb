//nolint:all
package repo

import (
	"fmt"
	"strings"

	"github.com/fzf-labs/godb/orm/utils/template"
)

// generateCreateFunc
func (r *Repo) generateCreateFunc() (string, error) {
	// 是否有索引
	haveIndex := len(r.index) > 0
	// 是否有主键索引
	havePrimaryKey := false
	primaryKeyFields := make([]string, 0)
	for _, v := range r.index {
		if v.PrimaryKey {
			havePrimaryKey = true
			primaryKeyFields = v.Columns
			break
		}
	}
	var createFunc string
	tplParams := map[string]any{
		"firstTableChar":        r.firstTableChar,
		"dbName":                r.dbName,
		"upperTableName":        r.upperTableName,
		"lowerTableName":        r.lowerTableName,
		"primaryKeyWhereFields": r.primaryKeyWhereFields(primaryKeyFields),
		"haveDeletedAt":         r.haveDeletedAt,
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
	for _, v := range r.index {
		if v.PrimaryKey {
			havePrimaryKey = true
			break
		}
	}
	var updateFunc string
	//参数
	tplParams := map[string]any{
		"firstTableChar": r.firstTableChar,
		"dbName":         r.dbName,
		"upperTableName": r.upperTableName,
		"lowerTableName": r.lowerTableName,
		"haveDeletedAt":  r.haveDeletedAt,
	}
	// 有主键索引
	if havePrimaryKey {
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
	}
	// 有索引
	for _, v := range r.index {
		if r.checkDaoFieldType(v.Columns) {
			continue
		}
		tplParams["upperFields"] = r.upperFields(v.Columns)
		tplParams["fieldAndDataTypes"] = r.fieldAndDataTypes(v.Columns)
		tplParams["whereFields"] = r.whereFields(v.Columns)
		updateBatchByFields, err := template.NewTemplate().Parse(UpdateBatchByFields).Execute(tplParams)
		if err != nil {
			return "", err
		}
		updateFunc += fmt.Sprintln(updateBatchByFields.String())
		updateBatchByFieldsTx, err := template.NewTemplate().Parse(UpdateBatchByFieldsTx).Execute(tplParams)
		if err != nil {
			return "", err
		}
		updateFunc += fmt.Sprintln(updateBatchByFieldsTx.String())
		if len(v.Columns) == 1 {
			tplParams["upperField"] = r.upperFieldName(v.Columns[0])
			tplParams["lowerField"] = r.lowerFieldName(v.Columns[0])
			tplParams["upperFieldPlural"] = r.plural(r.upperFieldName(v.Columns[0]))
			tplParams["lowerFieldPlural"] = r.plural(r.lowerFieldName(v.Columns[0]))
			tplParams["dataType"] = r.columnNameToDataType[v.Columns[0]]
			switch tplParams["dataType"] {
			case "bool":
			default:
				updateBatchByFieldPlural, err := template.NewTemplate().Parse(UpdateBatchByFieldPlural).Execute(tplParams)
				if err != nil {
					return "", err
				}
				updateFunc += fmt.Sprintln(updateBatchByFieldPlural.String())
				updateBatchByFieldPluralTx, err := template.NewTemplate().Parse(UpdateBatchByFieldPluralTx).Execute(tplParams)
				if err != nil {
					return "", err
				}
				updateFunc += fmt.Sprintln(updateBatchByFieldPluralTx.String())
			}
		}
	}
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
			"haveDeletedAt":     r.haveDeletedAt,
			"whereFields":       r.whereFields(v.Columns),
		}
		// 唯一 && 字段数于1
		if v.Unique && len(v.Columns) == 1 {
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
			switch tplParams["dataType"] {
			case "bool":
			default:
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

			switch tplParams["dataType"] {
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
	tplParams := map[string]any{
		"firstTableChar": r.firstTableChar,
		"dbName":         r.dbName,
		"upperTableName": r.upperTableName,
		"lowerTableName": r.lowerTableName,
		"haveDeletedAt":  r.haveDeletedAt,
	}
	findMultiByCondition, err := template.NewTemplate().Parse(FindMultiByCondition).Execute(tplParams)
	if err != nil {
		return "", err
	}
	readFunc += fmt.Sprintln(findMultiByCondition.String())
	findMultiByCacheCondition, err := template.NewTemplate().Parse(FindMultiByCacheCondition).Execute(tplParams)
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
			cacheFieldsJoinSli = append(cacheFieldsJoinSli, fmt.Sprintf("item.%s", r.upperFieldName(column)))
		}
		// 缓存删除key
		varCacheDelKeyTpl, err := template.NewTemplate().Parse(VarCacheDelKey).Execute(map[string]any{
			"firstTableChar":      r.firstTableChar,
			"upperTableName":      r.upperTableName,
			"cacheFields":         r.cacheFields(v.Columns),
			"delCacheFieldsParam": strings.Join(cacheFieldsJoinSli, ","),
			"haveDeletedAt":       r.haveDeletedAt,
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
			"haveDeletedAt":     r.haveDeletedAt,
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
			tplParams["whereFields"] = r.whereFields(v.Columns)
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
			tplParams["whereFields"] = r.whereFields(v.Columns)
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
			"haveDeletedAt":  r.haveDeletedAt,
		})
		if err != nil {
			return "", err
		}
		delMethods += fmt.Sprintln(deleteUniqueIndexCacheTpl.String())
	}
	return delMethods, nil
}
