package batch

import (
	"fmt"
	"math"
	"regexp"
	"sort"
	"strings"
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
