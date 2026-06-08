package gormx

import (
	"fmt"
	"strings"

	"gorm.io/gorm"
)

// GetPartitionTableToChildTables 获取分区表到子表的映射
func GetPartitionTableToChildTables(db *gorm.DB) (resp map[string][]string, err error) {
	if db == nil {
		return nil, fmt.Errorf("db cannot be nil")
	}
	switch db.Dialector.Name() {
	case Postgres:
		resp, err = getPGPartitionTableToChildTables(db)
		if err != nil {
			return nil, err
		}
		return resp, nil
	case MySQL:
		resp, err = getMySQLPartitionTableToChildTables(db)
		if err != nil {
			return nil, err
		}
		return resp, nil
	default:
		return nil, nil
	}
}

// getPartitionTableToChildTable 获取PG分区表到子表的映射
func getPGPartitionTableToChildTables(db *gorm.DB) (map[string][]string, error) {
	resp := make(map[string][]string)
	type tmp struct {
		PartitionedTable string `gorm:"column:partitioned_table" json:"partitioned_table"`
		ChildTables      string `gorm:"column:child_tables" json:"child_tables"`
	}
	result := make([]tmp, 0)
	sql := `SELECT p.relname AS partitioned_table,array_to_string(array_agg(c.relname),',')AS child_tables FROM pg_catalog.pg_class c JOIN pg_catalog.pg_inherits i ON c.oid=i.inhrelid JOIN pg_catalog.pg_class p ON p.oid=i.inhparent GROUP BY p.relname;`
	err := db.Raw(sql).Scan(&result).Error
	if err != nil {
		return nil, err
	}
	for _, v := range result {
		resp[v.PartitionedTable] = append(resp[v.PartitionedTable], strings.Split(v.ChildTables, ",")...)
	}
	return resp, nil
}

// getMySQLPartitionTableToChildTable 获取MySQL分区表到子表的映射
func getMySQLPartitionTableToChildTables(db *gorm.DB) (map[string][]string, error) {
	resp := make(map[string][]string)
	type tmp struct {
		TableName string `gorm:"column:TABLE_NAME"`
	}
	result := make([]tmp, 0)
	sql := `SELECT DISTINCT TABLE_NAME FROM INFORMATION_SCHEMA.PARTITIONS WHERE PARTITION_NAME IS NOT NULL AND TABLE_SCHEMA=? ORDER BY TABLE_NAME`
	err := db.Raw(sql, db.Migrator().CurrentDatabase()).Scan(&result).Error
	if err != nil {
		return nil, err
	}
	// MySQL partitions do not have child tables like PostgreSQL inheritance.
	for _, v := range result {
		resp[v.TableName] = nil
	}
	return resp, nil
}

// GetTableComments 获取数据库中所有表对应的 comment
func GetTableComments(db *gorm.DB) (map[string]string, error) {
	if db == nil {
		return nil, fmt.Errorf("db cannot be nil")
	}
	resp := make(map[string]string)
	switch db.Dialector.Name() {
	case MySQL:
		type tmp struct {
			TableName    string `gorm:"column:TABLE_NAME"`
			TableComment string `gorm:"column:TABLE_COMMENT"`
		}
		result := make([]tmp, 0)
		sql := `SELECT TABLE_NAME,TABLE_COMMENT FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_SCHEMA=?`
		err := db.Raw(sql, db.Migrator().CurrentDatabase()).Scan(&result).Error
		if err != nil {
			return nil, err
		}
		for _, v := range result {
			resp[v.TableName] = v.TableComment
		}
		return resp, nil
	case Postgres:
		type tmp struct {
			TableName    string `gorm:"column:table_name"`
			TableComment string `gorm:"column:table_comment"`
		}
		result := make([]tmp, 0)
		sql := buildPostgresTableCommentsQuery()
		err := db.Raw(sql).Scan(&result).Error
		if err != nil {
			return nil, err
		}
		for _, v := range result {
			resp[v.TableName] = v.TableComment
		}
		return resp, nil
	default:
		return nil, nil
	}
}

func buildPostgresTableCommentsQuery() string {
	return `SELECT c.relname AS table_name, obj_description(c.oid) AS table_comment FROM pg_class c WHERE c.relkind IN ('r', 'p') AND c.relname NOT LIKE 'pg_%' AND c.relname NOT LIKE 'sql_%'`
}
