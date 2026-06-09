package batch

import (
	"database/sql/driver"
	"fmt"
	"math"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

var identifierPattern = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

// getSQLQuantity 计算需要生成的 SQL 语句数量
func getSQLQuantity(length, batchSize int) int {
	return int(math.Ceil(float64(length) / float64(batchSize)))
}

// validateIdentifier 校验单段 SQL 标识符。
func validateIdentifier(name string) error {
	if !identifierPattern.MatchString(name) {
		return fmt.Errorf("invalid SQL identifier: %s", name)
	}
	return nil
}

// validateQualifiedIdentifier 校验支持 schema.table 的限定 SQL 标识符。
func validateQualifiedIdentifier(name string) error {
	parts := strings.Split(name, ".")
	for _, part := range parts {
		if part == "" || !identifierPattern.MatchString(part) {
			return fmt.Errorf("invalid SQL identifier: %s", name)
		}
	}
	return nil
}

// sortedFieldNames 返回稳定排序后的字段名列表。
func sortedFieldNames(fields map[string][]string, skip ...string) []string {
	skipSet := make(map[string]struct{}, len(skip))
	for _, item := range skip {
		skipSet[item] = struct{}{}
	}
	names := make([]string, 0, len(fields))
	for name := range fields {
		if _, ok := skipSet[name]; ok {
			continue
		}
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// sliceBatchValues 按批次范围切分字段值。
func sliceBatchValues(values []string, start, end int) ([]string, error) {
	if start < 0 || end < start || end > len(values) {
		return nil, fmt.Errorf("invalid batch range [%d:%d] for %d values", start, end, len(values))
	}
	return values[start:end], nil
}

func normalizeBatchSlice(dataList any) (reflect.Value, error) {
	rv := reflect.ValueOf(dataList)
	if !rv.IsValid() {
		return reflect.Value{}, fmt.Errorf("dataList must be a slice")
	}
	for rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			return reflect.Value{}, fmt.Errorf("dataList must be a slice")
		}
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Slice {
		return reflect.Value{}, fmt.Errorf("dataList must be a slice")
	}
	if rv.Len() == 0 {
		return reflect.Value{}, fmt.Errorf("dataList cannot be empty")
	}
	return rv, nil
}

// escapePostgresIdentifier 转义 PostgreSQL 标识符。
func escapePostgresIdentifier(name string) string {
	return `"` + strings.ReplaceAll(name, `"`, `""`) + `"`
}

// escapeQualifiedIdentifier 按段转义限定 SQL 标识符。
func escapeQualifiedIdentifier(name string, escape func(string) string) string {
	parts := strings.Split(name, ".")
	for i, part := range parts {
		parts[i] = escape(part)
	}
	return strings.Join(parts, ".")
}

const sqlTimeLayout = "2006-01-02 15:04:05"

// formatSQLValue 将字段值格式化为可直接拼入 SQL 的字面量。
func formatSQLValue(field reflect.Value, quote func(string) string) (string, error) {
	return formatSQLValueWithBool(field, quote, strconv.FormatBool)
}

func formatSQLValueWithBool(field reflect.Value, quote func(string) string, boolLiteral func(bool) string) (string, error) {
	if !field.IsValid() {
		return "", fmt.Errorf("unsupported field type: invalid")
	}
	if field.Kind() == reflect.Ptr && field.IsNil() {
		return "NULL", nil
	}

	if field.CanInterface() {
		switch value := field.Interface().(type) {
		case time.Time:
			return quote(value.Format(sqlTimeLayout)), nil
		case *time.Time:
			if value == nil {
				return "NULL", nil
			}
			return quote(value.Format(sqlTimeLayout)), nil
		case driver.Valuer:
			raw, err := value.Value()
			if err != nil {
				return "", err
			}
			return formatSQLValueFromAnyWithBool(raw, quote, boolLiteral)
		case fmt.Stringer:
			return quote(value.String()), nil
		}
	}
	if field.CanAddr() && field.Addr().CanInterface() {
		if value, ok := field.Addr().Interface().(driver.Valuer); ok {
			raw, err := value.Value()
			if err != nil {
				return "", err
			}
			return formatSQLValueFromAnyWithBool(raw, quote, boolLiteral)
		}
	}
	if field.CanInterface() && field.Kind() != reflect.Ptr {
		addressable := reflect.New(field.Type())
		addressable.Elem().Set(field)
		if value, ok := addressable.Interface().(driver.Valuer); ok {
			raw, err := value.Value()
			if err != nil {
				return "", err
			}
			return formatSQLValueFromAnyWithBool(raw, quote, boolLiteral)
		}
	}

	switch field.Kind() {
	case reflect.Int:
		return strconv.FormatInt(field.Int(), 10), nil
	case reflect.Int8:
		return strconv.FormatInt(field.Int(), 10), nil
	case reflect.Int16:
		return strconv.FormatInt(field.Int(), 10), nil
	case reflect.Int32:
		return strconv.FormatInt(field.Int(), 10), nil
	case reflect.Int64:
		return strconv.FormatInt(field.Int(), 10), nil
	case reflect.Uint:
		return strconv.FormatUint(field.Uint(), 10), nil
	case reflect.Uint8:
		return strconv.FormatUint(field.Uint(), 10), nil
	case reflect.Uint16:
		return strconv.FormatUint(field.Uint(), 10), nil
	case reflect.Uint32:
		return strconv.FormatUint(field.Uint(), 10), nil
	case reflect.Uint64:
		return strconv.FormatUint(field.Uint(), 10), nil
	case reflect.String:
		return quote(field.String()), nil
	case reflect.Float32:
		return formatSQLFloat(field.Float(), 32)
	case reflect.Float64:
		return formatSQLFloat(field.Float(), 64)
	case reflect.Bool:
		return boolLiteral(field.Bool()), nil
	case reflect.Slice:
		if field.Type().Elem().Kind() == reflect.Uint8 {
			return quote(string(field.Bytes())), nil
		}
	case reflect.Ptr:
		if field.IsNil() {
			return "NULL", nil
		}
		return formatSQLValueWithBool(field.Elem(), quote, boolLiteral)
	}

	return "", fmt.Errorf("unsupported field type: %v", field.Kind())
}

func formatSQLValueFromAny(value any, quote func(string) string) (string, error) {
	return formatSQLValueFromAnyWithBool(value, quote, strconv.FormatBool)
}

func formatSQLValueFromAnyWithBool(value any, quote func(string) string, boolLiteral func(bool) string) (string, error) {
	if value == nil {
		return "NULL", nil
	}

	rv := reflect.ValueOf(value)
	if !rv.IsValid() {
		return "NULL", nil
	}
	if rv.Kind() == reflect.Ptr && rv.IsNil() {
		return "NULL", nil
	}

	switch v := value.(type) {
	case time.Time:
		return quote(v.Format(sqlTimeLayout)), nil
	case *time.Time:
		if v == nil {
			return "NULL", nil
		}
		return quote(v.Format(sqlTimeLayout)), nil
	case driver.Valuer:
		raw, err := v.Value()
		if err != nil {
			return "", err
		}
		return formatSQLValueFromAnyWithBool(raw, quote, boolLiteral)
	case fmt.Stringer:
		return quote(v.String()), nil
	case string:
		return quote(v), nil
	case []byte:
		return quote(string(v)), nil
	case bool:
		return boolLiteral(v), nil
	case int:
		return strconv.FormatInt(int64(v), 10), nil
	case int8:
		return strconv.FormatInt(int64(v), 10), nil
	case int16:
		return strconv.FormatInt(int64(v), 10), nil
	case int32:
		return strconv.FormatInt(int64(v), 10), nil
	case int64:
		return strconv.FormatInt(v, 10), nil
	case uint:
		return strconv.FormatUint(uint64(v), 10), nil
	case uint8:
		return strconv.FormatUint(uint64(v), 10), nil
	case uint16:
		return strconv.FormatUint(uint64(v), 10), nil
	case uint32:
		return strconv.FormatUint(uint64(v), 10), nil
	case uint64:
		return strconv.FormatUint(v, 10), nil
	case float32:
		return formatSQLFloat(float64(v), 32)
	case float64:
		return formatSQLFloat(v, 64)
	}
	if rv.Kind() == reflect.Ptr {
		return formatSQLValueWithBool(rv.Elem(), quote, boolLiteral)
	}
	if rv.Kind() == reflect.Slice && rv.Type().Elem().Kind() == reflect.Uint8 {
		return quote(string(rv.Bytes())), nil
	}
	if stringer, ok := value.(fmt.Stringer); ok {
		return quote(stringer.String()), nil
	}

	return "", fmt.Errorf("unsupported field type: %T", value)
}

func formatSQLFloat(value float64, bitSize int) (string, error) {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return "", fmt.Errorf("unsupported float value: %v", value)
	}
	return strconv.FormatFloat(value, 'f', -1, bitSize), nil
}

func formatBatchIDValue(field reflect.Value, quote func(string) string) (string, error) {
	if !field.IsValid() {
		return "", fmt.Errorf("id field is invalid")
	}
	for field.Kind() == reflect.Ptr {
		if field.IsNil() {
			return "", fmt.Errorf("empty id value")
		}
		field = field.Elem()
	}

	switch field.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if field.Int() <= 0 {
			return "", fmt.Errorf("id value must be greater than 0")
		}
		return strconv.FormatInt(field.Int(), 10), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if field.Uint() == 0 {
			return "", fmt.Errorf("id value must be greater than 0")
		}
		return strconv.FormatUint(field.Uint(), 10), nil
	case reflect.String:
		if strings.TrimSpace(field.String()) == "" {
			return "", fmt.Errorf("empty id value")
		}
		return quote(field.String()), nil
	default:
		return formatSQLValue(field, quote)
	}
}
