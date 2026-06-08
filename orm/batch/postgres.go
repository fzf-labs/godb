package batch

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

// PostgresBatchUpdateToSQLArray 批量更新
// tableName: 表名
// dataList: 需要更新的数据列表,必须是指向结构体的切片
func PostgresBatchUpdateToSQLArray(tableName string, dataList any) ([]string, error) {
	if tableName == "" {
		return nil, errors.New("tableName cannot be empty")
	}
	if err := validateQualifiedIdentifier(tableName); err != nil {
		return nil, err
	}

	// 检查 dataList 是否为切片
	rv := reflect.ValueOf(dataList)
	if !rv.IsValid() {
		return nil, errors.New("dataList must be a slice")
	}
	if rv.Kind() != reflect.Slice {
		return nil, errors.New("dataList must be a slice")
	}

	if rv.Len() == 0 {
		return nil, errors.New("dataList cannot be empty")
	}

	// 获取元素类型
	elemType := rv.Type().Elem()
	if elemType.Kind() != reflect.Ptr || elemType.Elem().Kind() != reflect.Struct {
		return nil, errors.New("dataList must be a slice of struct pointers")
	}

	// 获取结构体字段信息
	fields, err := getStructFields(elemType.Elem())
	if err != nil {
		return nil, fmt.Errorf("get struct fields error: %w", err)
	}

	// 检查是否存在 "id" 字段
	idColumn, idInfo, ok := findIDField(fields)
	if !ok {
		return nil, errors.New("struct must have a field with json tag 'id'")
	}
	if len(fields) == 1 {
		return nil, errors.New("no update columns found")
	}

	// 准备数据
	ids := make([]string, 0, rv.Len())
	updateMap := make(map[string][]string)
	for i := 0; i < rv.Len(); i++ {
		// 获取每个结构体实例
		item := rv.Index(i)
		if item.IsNil() {
			return nil, fmt.Errorf("dataList[%d] cannot be nil", i)
		}
		structVal := item.Elem()
		idField := structVal.FieldByName(idInfo.name)
		if !idField.IsValid() {
			return nil, fmt.Errorf("id field not found in struct at index %d", i)
		}

		idStr, err := formatBatchIDValue(idField, quotePostgresString)
		if err != nil {
			return nil, fmt.Errorf("%w at index %d", err, i)
		}

		ids = append(ids, idStr)

		// 处理其他字段
		for fieldName, fieldInfo := range fields {
			if fieldName == idColumn {
				continue
			}
			fieldValue := structVal.FieldByName(fieldInfo.name)
			if !fieldValue.IsValid() {
				return nil, fmt.Errorf("field %s not found in struct at index %d", fieldName, i)
			}

			valStr, err := formatPostgresFieldValue(fieldValue)
			if err != nil {
				return nil, fmt.Errorf("format field %s error at index %d: %w", fieldName, i, err)
			}
			updateMap[fieldName] = append(updateMap[fieldName], valStr)
		}
	}

	// 计算 SQL 语句数量
	length := len(ids)
	const batchSize = 200
	sqlQuantity := getSQLQuantity(length, batchSize)

	// 生成 SQL 语句
	sqlArray := make([]string, 0, sqlQuantity)
	for i := 0; i < sqlQuantity; i++ {
		batchStart := i * batchSize
		batchEnd := min((i+1)*batchSize, length)

		sql, err := buildPostgresBatchUpdateSQLWithIDColumn(tableName, idColumn, updateMap, batchStart, batchEnd, ids[batchStart:batchEnd])
		if err != nil {
			return nil, fmt.Errorf("build batch update SQL error: %w", err)
		}
		sqlArray = append(sqlArray, sql)
	}

	return sqlArray, nil
}

// formatPostgresFieldValue 格式化 PostgreSQL 字段值
func formatPostgresFieldValue(field reflect.Value) (string, error) {
	return formatSQLValueWithBool(field, quotePostgresString, postgresBoolLiteral)
}

// buildPostgresBatchUpdateSQL 生成 PostgreSQL 批量更新 SQL
func buildPostgresBatchUpdateSQL(tableName string, updateMap map[string][]string, batchStart, batchEnd int, batchIDs []string) (string, error) {
	return buildPostgresBatchUpdateSQLWithIDColumn(tableName, "id", updateMap, batchStart, batchEnd, batchIDs)
}

func buildPostgresBatchUpdateSQLWithIDColumn(tableName, idColumn string, updateMap map[string][]string, batchStart, batchEnd int, batchIDs []string) (string, error) {
	if len(batchIDs) == 0 {
		return "", errors.New("batchIDs cannot be empty")
	}
	if len(updateMap) == 0 {
		return "", errors.New("no update columns found")
	}

	var sqlBuilder strings.Builder
	sqlBuilder.Grow(4096)

	sqlBuilder.WriteString("UPDATE " + escapeQualifiedIdentifier(tableName, escapePostgresIdentifier) + " SET ")

	fieldNames := sortedFieldNames(updateMap)
	setClauses := make([]string, 0, len(fieldNames))
	for _, fieldName := range fieldNames {
		fieldValueList, err := sliceBatchValues(updateMap[fieldName], batchStart, batchEnd)
		if err != nil {
			return "", err
		}
		clause := escapePostgresIdentifier(fieldName) + " = CASE " + escapePostgresIdentifier(idColumn)
		for i, id := range batchIDs {
			clause += " WHEN " + id + " THEN " + fieldValueList[i]
		}
		clause += " END"
		setClauses = append(setClauses, clause)
	}

	sqlBuilder.WriteString(strings.Join(setClauses, ", "))
	sqlBuilder.WriteString(" WHERE " + escapePostgresIdentifier(idColumn) + " IN (" + strings.Join(batchIDs, ",") + ")")

	return sqlBuilder.String(), nil
}

func quotePostgresString(s string) string {
	return fmt.Sprintf("'%s'", strings.ReplaceAll(s, "'", "''"))
}

func postgresBoolLiteral(value bool) string {
	if value {
		return "TRUE"
	}
	return "FALSE"
}
