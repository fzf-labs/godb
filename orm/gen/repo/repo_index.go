package repo

import (
	"sort"
	"strings"

	"github.com/fzf-labs/godb/orm/gormx"
)

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
			ColumnKey:  joinIndexColumnKey(columns),
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
				columnKey := joinIndexColumnKey(v.Columns[0:i])
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

func joinIndexColumnKey(columns []string) string {
	return strings.Join(columns, ":")
}
