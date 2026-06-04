package sqldump

import "testing"

func TestRunCommandReturnsSQLDumpError(t *testing.T) {
	oldDB, oldDSN, oldOut, oldTables, oldOverwrite := db, dsn, outPutPath, targetTables, fileOverwrite
	defer func() {
		db, dsn, outPutPath, targetTables, fileOverwrite = oldDB, oldDSN, oldOut, oldTables, oldOverwrite
	}()

	db = "sqlite"
	dsn = ":memory:"
	outPutPath = t.TempDir()
	targetTables = "users"
	fileOverwrite = true

	if err := Run(nil, nil); err == nil {
		t.Fatal("expected unknown database type error")
	}
}
