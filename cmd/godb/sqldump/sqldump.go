package sqldump

import "fmt"

type SQLDump struct {
	db            string // 数据库类型 mysql postgres
	dsn           string // 数据库连接
	outPutPath    string // 输出路径
	targetTables  string // 指定表
	fileOverwrite bool   // 是否覆盖
}

// NewSQLDump 创建数据库结构导出器。
func NewSQLDump(db, dsn, outPutPath, targetTables string, fileCover bool) *SQLDump {
	return &SQLDump{db: db, dsn: dsn, outPutPath: outPutPath, targetTables: targetTables, fileOverwrite: fileCover}
}

// Run 根据数据库类型执行结构导出。
func (s *SQLDump) Run() error {
	switch s.db {
	case "mysql":
		return s.DumpMySQL()
	case "postgres":
		return s.DumpPostgres()
	default:
		return fmt.Errorf("unknown database type: %s", s.db)
	}
}
