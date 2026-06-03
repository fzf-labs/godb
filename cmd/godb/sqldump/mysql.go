package sqldump

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fzf-labs/godb/orm/gormx"
	"github.com/fzf-labs/godb/orm/utils/fileutil"
	"github.com/fzf-labs/godb/orm/utils/strutil"
)

// DumpMySQL 导出创建语句
func (s *SQLDump) DumpMySQL() error {
	dbClient, err := gormx.NewSimpleGormClient(s.db, s.dsn)
	if err != nil {
		return err
	}
	var tables []string
	if s.targetTables != "" {
		tables = strings.Split(s.targetTables, ",")
	} else {
		tables, err = dbClient.Migrator().GetTables()
		if err != nil {
			return err
		}
	}
	outPath := outputDir(s.outPutPath, dbClient.Migrator().CurrentDatabase())
	err = os.MkdirAll(outPath, os.ModePerm)
	if err != nil {
		return fmt.Errorf("create output path: %w", err)
	}
	for _, v := range tables {
		result := make(map[string]any)
		err := dbClient.Raw(buildMySQLShowCreateTableSQL(dbClient.Migrator().CurrentDatabase(), v)).Scan(result).Error
		if err != nil {
			return fmt.Errorf("show create table %s: %w", v, err)
		}
		outFile := filepath.Join(outPath, fmt.Sprintf("%s.sql", v))
		if !s.fileOverwrite {
			if fileutil.Exists(outFile) {
				continue
			}
		}
		tableContent := strutil.ConvToString(result["Create Table"])
		if tableContent != "" {
			err := fileutil.WriteContentCover(outFile, tableContent)
			if err != nil {
				return fmt.Errorf("write %s: %w", outFile, err)
			}
		}
	}
	return nil
}

func buildMySQLShowCreateTableSQL(dbName, table string) string {
	return fmt.Sprintf("SHOW CREATE TABLE %s.%s", quoteMySQLIdentifier(dbName), quoteMySQLIdentifier(table))
}

func quoteMySQLIdentifier(identifier string) string {
	return "`" + strings.ReplaceAll(identifier, "`", "``") + "`"
}
