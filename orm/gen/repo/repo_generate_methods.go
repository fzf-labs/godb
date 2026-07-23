//nolint:all
package repo

import (
	"fmt"

	"github.com/fzf-labs/godb/orm/utils/template"
)

func (r *Repo) generateCommonMethods() (string, error) {
	var commonMethods string
	tplParams := map[string]any{
		"dbName":         r.dbName,
		"upperTableName": r.upperTableName,
	}
	interfaceNewData, err := template.NewTemplate().Parse(InterfaceNewData).Execute(tplParams)
	if err != nil {
		return "", err
	}
	commonMethods += fmt.Sprintln(interfaceNewData.String())
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
		"haveDeletedAt":  r.haveDeletedAt,
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
	for _, v := range r.index {
		if v.PrimaryKey {
			havePrimaryKey = true
			break
		}
	}
	var updateMethods string
	tplParams := map[string]any{
		"dbName":         r.dbName,
		"upperTableName": r.upperTableName,
		"haveDeletedAt":  r.haveDeletedAt,
	}
	// 有主键索引
	if havePrimaryKey {
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
	}
	for _, v := range r.index {
		if r.checkDaoFieldType(v.Columns) {
			continue
		}
		tplParams["dataType"] = r.columnNameToDataType[v.Columns[0]]
		tplParams["upperFields"] = r.upperFields(v.Columns)
		tplParams["fieldAndDataTypes"] = r.fieldAndDataTypes(v.Columns)
		tplParams["whereFields"] = r.whereFields(v.Columns)
		interfaceUpdateBatchByFields, err := template.NewTemplate().Parse(InterfaceUpdateBatchByFields).Execute(tplParams)
		if err != nil {
			return "", err
		}
		updateMethods += fmt.Sprintln(interfaceUpdateBatchByFields.String())

		interfaceUpdateBatchByFieldsTx, err := template.NewTemplate().Parse(InterfaceUpdateBatchByFieldsTx).Execute(tplParams)
		if err != nil {
			return "", err
		}
		updateMethods += fmt.Sprintln(interfaceUpdateBatchByFieldsTx.String())
		if len(v.Columns) == 1 {
			tplParams["upperField"] = r.upperFieldName(v.Columns[0])
			tplParams["lowerField"] = r.lowerFieldName(v.Columns[0])
			tplParams["upperFieldPlural"] = r.plural(r.upperFieldName(v.Columns[0]))
			tplParams["lowerFieldPlural"] = r.plural(r.lowerFieldName(v.Columns[0]))
			tplParams["dataType"] = r.columnNameToDataType[v.Columns[0]]
			switch tplParams["dataType"] {
			case "bool":
			default:
				interfaceUpdateBatchByFieldPlural, err := template.NewTemplate().Parse(InterfaceUpdateBatchByFieldPlural).Execute(tplParams)
				if err != nil {
					return "", err
				}
				updateMethods += fmt.Sprintln(interfaceUpdateBatchByFieldPlural.String())
				interfaceUpdateBatchByFieldPluralTx, err := template.NewTemplate().Parse(InterfaceUpdateBatchByFieldPluralTx).Execute(tplParams)
				if err != nil {
					return "", err
				}
				updateMethods += fmt.Sprintln(interfaceUpdateBatchByFieldPluralTx.String())
			}
		}
	}
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
			"haveDeletedAt":     r.haveDeletedAt,
		}
		// 唯一 && 字段数于1
		if v.Unique && len(v.Columns) == 1 {
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

			switch tplParams["dataType"] {
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
			switch tplParams["dataType"] {
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
	tplParams := map[string]any{
		"dbName":         r.dbName,
		"upperTableName": r.upperTableName,
		"lowerTableName": r.lowerTableName,
		"haveDeletedAt":  r.haveDeletedAt,
	}
	interfaceFindMultiByCondition, err := template.NewTemplate().Parse(InterfaceFindMultiByCondition).Execute(tplParams)
	if err != nil {
		return "", err
	}
	readMethods += fmt.Sprintln(interfaceFindMultiByCondition.String())

	interfaceFindMultiByCacheCondition, err := template.NewTemplate().Parse(InterfaceFindMultiByCacheCondition).Execute(tplParams)
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
			"haveDeletedAt":     r.haveDeletedAt,
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
