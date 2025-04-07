package sqldump

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/fzf-labs/godb/orm/gormx"
	"github.com/fzf-labs/godb/orm/utils"
	"github.com/fzf-labs/godb/orm/utils/file"
)

// DumpMySQL 导出创建语句
func (s *SQLDump) DumpMySQL() {
	dbClient := gormx.NewSimpleGormClient(s.db, s.dsn)
	var tables []string
	var err error
	if s.targetTables != "" {
		tables = strings.Split(s.targetTables, ",")
	} else {
		tables, err = dbClient.Migrator().GetTables()
		if err != nil {
			return
		}
	}
	outPath := filepath.Join(strings.Trim(s.outPutPath, "/"), dbClient.Migrator().CurrentDatabase())
	err = os.MkdirAll(outPath, os.ModePerm)
	if err != nil {
		log.Println("DumpMySQL create path err:", err)
		return
	}
	for _, v := range tables {
		result := make(map[string]any)
		err := dbClient.Raw(fmt.Sprintf("SHOW CREATE TABLE `%s`.`%s`", dbClient.Migrator().CurrentDatabase(), v)).Scan(result).Error
		if err != nil {
			log.Println("DumpMySQL sql err:", err)
			return
		}
		outFile := filepath.Join(outPath, fmt.Sprintf("%s.sql", v))
		if !s.fileOverwrite {
			if file.Exists(outFile) {
				continue
			}
		}
		tableContent := utils.ConvToString(result["Create Table"])
		if tableContent != "" {
			err := file.WriteContentCover(outFile, tableContent)
			if err != nil {
				log.Println("DumpMySQL file write err:", err)
				return
			}
		}
	}
}
