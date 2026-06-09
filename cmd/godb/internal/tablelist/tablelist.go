package tablelist

import (
	"fmt"
	"strings"
)

// ParseCSV trims and splits a comma-separated table list.
// Blank input returns nil so callers can treat it as "no filter".
// Empty entries inside a non-empty list return an error.
func ParseCSV(raw string) ([]string, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, nil
	}
	parts := strings.Split(raw, ",")
	tables := make([]string, 0, len(parts))
	seen := make(map[string]struct{}, len(parts))
	for i, part := range parts {
		table := strings.TrimSpace(part)
		if table == "" {
			return nil, fmt.Errorf("empty table name at position %d", i+1)
		}
		if _, ok := seen[table]; ok {
			continue
		}
		seen[table] = struct{}{}
		tables = append(tables, table)
	}
	if len(tables) == 0 {
		return nil, fmt.Errorf("no table names found")
	}
	return tables, nil
}
