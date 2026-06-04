package sqltopb

import "testing"

func TestRunReturnsDriverErrorAfterParsingTables(t *testing.T) {
	oldDB, oldDSN, oldTables := db, dsn, targetTables
	oldPBPackage, oldPBGoPackage, oldOut := pbPackage, pbGoPackage, outPutPath
	defer func() {
		db, dsn, targetTables = oldDB, oldDSN, oldTables
		pbPackage, pbGoPackage, outPutPath = oldPBPackage, oldPBGoPackage, oldOut
	}()

	db = "sqlite"
	dsn = ":memory:"
	targetTables = "users,roles"
	pbPackage = "pb"
	pbGoPackage = "example.com/project/pb;pb"
	outPutPath = t.TempDir()

	if err := Run(nil, nil); err == nil {
		t.Fatal("expected unknown driver error")
	}
}
