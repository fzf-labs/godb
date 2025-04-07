package sqldump

import "log"

type SQLDump struct {
	db            string // 数据库类型 mysql postgres
	dsn           string // 数据库连接
	outPutPath    string // 输出路径
	targetTables  string // 指定表
	fileOverwrite bool   // 是否覆盖
}

func NewSQLDump(db, dsn, outPutPath, targetTables string, fileCover bool) *SQLDump {
	return &SQLDump{db: db, dsn: dsn, outPutPath: outPutPath, targetTables: targetTables, fileOverwrite: fileCover}
}

func (s *SQLDump) Run() {
	switch s.db {
	case "mysql":
		s.DumpMySQL()
	case "postgres":
		s.DumpPostgres()
	default:
		log.Println("unknown database type: ", s.db)
	}
}
