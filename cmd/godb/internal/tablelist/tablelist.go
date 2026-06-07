package tablelist

import (
	"fmt"
	"strings"
)

// ParseCSV trims and splits a comma-separated table list.
// Blank input returns nil so callers can treat it as "no filter".
// Inputs that contain only separators or whitespace return an error.
func ParseCSV(raw string) ([]string, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, nil
	}
	parts := strings.Split(raw, ",")
	tables := make([]string, 0, len(parts))
	for _, part := range parts {
		table := strings.TrimSpace(part)
		if table == "" {
			continue
		}
		tables = append(tables, table)
	}
	if len(tables) == 0 {
		return nil, fmt.Errorf("no table names found")
	}
	return tables, nil
}
