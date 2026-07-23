package sqldump

import (
	"fmt"
	"os"
	"strings"

	"github.com/fzf-labs/godb/cmd/godb/internal/tablelist"
	"github.com/fzf-labs/godb/orm/gormx"
	"github.com/fzf-labs/godb/orm/utils/fileutil"
	"github.com/fzf-labs/godb/orm/utils/strutil"
)

var newSimpleGormClient = gormx.NewSimpleGormClient

// DumpMySQL 导出创建语句
func (s *SQLDump) DumpMySQL() error {
	var err error
	tables, err := tablelist.ParseCSV(s.targetTables)
	if err != nil {
		return err
	}
	if len(tables) > 0 {
		if err := validateMySQLTablePatterns(tables); err != nil {
			return err
		}
	}
	dbClient, err := newSimpleGormClient(s.db, s.dsn)
	if err != nil {
		return err
	}
	if dbClient == nil {
		return fmt.Errorf("sqldump database client cannot be nil")
	}
	defer closeGormDB(dbClient)
	if len(tables) == 0 {
		tables, err = dbClient.Migrator().GetTables()
		if err != nil {
			return err
		}
	}
	if len(tables) == 0 {
		return fmt.Errorf("no tables to dump")
	}
	if err := validateMySQLTablePatterns(tables); err != nil {
		return err
	}
	currentDatabase := dbClient.Migrator().CurrentDatabase()
	outPath := outputDir(s.outPutPath, currentDatabase)
	err = os.MkdirAll(outPath, os.ModePerm)
	if err != nil {
		return fmt.Errorf("create output path: %w", err)
	}
	for _, v := range tables {
		outFile, err := fileutil.JoinOutputFilePath(outPath, v, ".sql")
		if err != nil {
			return err
		}
		if !s.fileOverwrite {
			if fileutil.Exists(outFile) {
				continue
			}
		}
		result := make(map[string]any)
		err = dbClient.Raw(buildMySQLShowCreateTableSQL(currentDatabase, v)).Scan(result).Error
		if err != nil {
			return fmt.Errorf("show create table %s: %w", v, err)
		}
		tableContent := strutil.ConvToString(result["Create Table"])
		err = fileutil.WriteContentCover(outFile, tableContent)
		if err != nil {
			return fmt.Errorf("write %s: %w", outFile, err)
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

func validateMySQLTablePatterns(tables []string) error {
	for _, table := range tables {
		if err := validateMySQLTablePattern(table); err != nil {
			return err
		}
	}
	return nil
}

func validateMySQLTablePattern(table string) error {
	parts := strings.Split(table, ".")
	for _, part := range parts {
		if !pgDumpTableIdentifierPattern.MatchString(part) {
			return fmt.Errorf("invalid mysql table pattern: %q", table)
		}
	}
	return nil
}
