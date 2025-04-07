package gormx

import (
	"fmt"

	"gorm.io/gorm"
)

type Index struct {
	TableName  string `json:"table_name" gorm:"column:table_name"`
	IndexName  string `json:"index_name" gorm:"column:index_name"`
	ColumnName string `json:"column_name" gorm:"column:column_name"`
	IsUnique   bool   `json:"is_unique" gorm:"column:is_unique"`
	Primary    bool   `json:"primary" gorm:"column:primary"`
}

// GetIndexes 获取索引
func GetIndexes(db *gorm.DB, table string) ([]*Index, error) {
	resp := make([]*Index, 0)
	var err error
	switch db.Dialector.Name() {
	case Postgres:
		resp, err = getPgIndexes(db, table)
		if err != nil {
			return nil, err
		}
		return resp, nil
	default:
		result, err := db.Migrator().GetIndexes(table)
		if err != nil {
			return nil, err
		}
		for _, v := range result {
			unique, _ := v.Unique()
			isPrimaryKey, _ := v.PrimaryKey()
			for _, vv := range v.Columns() {
				resp = append(resp, &Index{
					TableName:  table,
					IndexName:  v.Name(),
					ColumnName: vv,
					IsUnique:   unique,
					Primary:    isPrimaryKey,
				})
			}
		}
		return resp, nil
	}
}

// getPgIndexes 查询PG索引
func getPgIndexes(db *gorm.DB, table string) ([]*Index, error) {
	result := make([]*Index, 0)
	sql := fmt.Sprintf(`select t.relname as table_name,i.relname as index_name,a.attname as column_name,ix.indisunique as is_unique,ix.indisprimary as primary from pg_class t,pg_class i,pg_index ix,pg_attribute a where t.oid=ix.indrelid and i.oid=ix.indexrelid and a.attrelid=t.oid and a.attnum=ANY(ix.indkey)and t.relkind='r' and t.relname='%s'`, table)
	err := db.Raw(sql).Scan(&result).Error
	if err != nil {
		return nil, err
	}
	return result, nil
}

// SortIndexColumns 排序索引字段
func SortIndexColumns(db *gorm.DB, table string) (map[string][]string, error) {
	resp := make(map[string][]string)
	var err error
	switch db.Dialector.Name() {
	case Postgres:
		resp, err = pgSortIndexColumns(db, table)
		if err != nil {
			return nil, err
		}
	default:
		result, err := db.Migrator().GetIndexes(table)
		if err != nil {
			return nil, err
		}
		for _, v := range result {
			if _, ok := resp[v.Name()]; !ok {
				resp[v.Name()] = make([]string, 0)
			}
			resp[v.Name()] = v.Columns()
		}
	}
	return resp, nil
}

// pgSortIndexColumns  postgres索引字段排序
func pgSortIndexColumns(db *gorm.DB, table string) (map[string][]string, error) {
	resp := make(map[string][]string)
	type Tmp struct {
		TableName  string `gorm:"column:table_name" json:"table_name"`
		IndexName  string `gorm:"column:index_name" json:"index_name"`
		ColumnName string `gorm:"column:column_name" json:"column_name"`
	}
	result := make([]Tmp, 0)
	sql := fmt.Sprintf(`SELECT t.relname AS table_name,i.relname AS index_name,a.attname AS column_name,ix.indisunique AS is_unique,ix.indisprimary AS PRIMARY FROM pg_class t JOIN pg_index ix ON t.oid=ix.indrelid JOIN pg_class i ON i.oid=ix.indexrelid JOIN pg_attribute a ON a.attrelid=t.oid AND a.attnum=ANY(ix.indkey)WHERE t.relkind='r' AND t.relname='%s' ORDER BY ix.indrelid,(array_position(ix.indkey,a.attnum))`, table)
	err := db.Raw(sql).Scan(&result).Error
	if err != nil {
		return nil, err
	}
	for _, v := range result {
		if _, ok := resp[v.IndexName]; !ok {
			resp[v.IndexName] = make([]string, 0)
		}
		resp[v.IndexName] = append(resp[v.IndexName], v.ColumnName)
	}
	return resp, nil
}
