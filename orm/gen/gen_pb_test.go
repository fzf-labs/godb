package gen

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/fzf-labs/godb/internal/testenv"
	"github.com/fzf-labs/godb/orm/gormx"
)

// TestNewGenerationPb 验证 proto 文件生成。
func TestNewGenerationPb(t *testing.T) {
	db, err := gormx.NewSimpleGormClient(gormx.Postgres, testenv.PostgresDSN("gorm_gen"))
	if err != nil {
		testenv.SkipIfUnavailable(t, "postgres unavailable: %v", err)
	}
	err = NewGenerationPB(
		db,
		"../example/pb",
		"api.gorm_gen.v1",
		"api/gorm_gen/v1;v1",
		WithPBOpts(ModelOptionRemoveDefault(), ModelOptionUnderline("ul_")),
	).Do()
	if err != nil {
		t.Fatal(err)
	}
}

func TestNewGenerationPbRejectsNilDB(t *testing.T) {
	err := NewGenerationPB(nil, t.TempDir(), "api.gorm_gen.v1", "api/gorm_gen/v1;v1").Do()
	if err == nil || !strings.Contains(err.Error(), "db cannot be nil") {
		t.Fatalf("expected nil db error, got %v", err)
	}
}

func TestNewGenerationPbRejectsBlankPackage(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "proto-gen.db")
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}

	err = NewGenerationPB(db, t.TempDir(), "", "api/gorm_gen/v1;v1").Do()
	if err == nil || !strings.Contains(err.Error(), "package cannot be empty") {
		t.Fatalf("expected blank package error, got %v", err)
	}
}

func TestNewGenerationPb_ReturnsGenerationErrors(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "proto-gen.db")
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	type protoExample struct {
		ID     uint `gorm:"primaryKey"`
		Name   string
		Status int
	}
	if err := db.AutoMigrate(&protoExample{}); err != nil {
		t.Fatal(err)
	}
	tables, err := db.Migrator().GetTables()
	if err != nil {
		t.Fatal(err)
	}
	if len(tables) == 0 {
		t.Fatal("expected at least one table")
	}

	outDir := t.TempDir()
	outFile := filepath.Join(outDir, tables[0]+".proto")
	if err := os.WriteFile(outFile, []byte("existing"), 0600); err != nil {
		t.Fatal(err)
	}

	err = NewGenerationPB(
		db,
		outDir,
		"api.gorm_gen.v1",
		"api/gorm_gen/v1;v1",
		WithPBTables([]string{tables[0]}),
	).Do()
	if err == nil {
		t.Fatal("expected generation error, got nil")
	}
	if !strings.Contains(err.Error(), tables[0]) {
		t.Fatalf("expected error to mention table %q, got %v", tables[0], err)
	}
}

func TestNewGenerationPbWithSQLiteSuccess(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "proto-gen.db")
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	type protoSuccessExample struct {
		ID   uint `gorm:"primaryKey"`
		Name string
	}
	if err := db.AutoMigrate(&protoSuccessExample{}); err != nil {
		t.Fatal(err)
	}

	err = NewGenerationPB(
		db,
		t.TempDir(),
		"api.demo.v1",
		"api/demo/v1;v1",
		WithPBOpts(ModelOptionRemoveDefault()),
		WithPBTables([]string{"proto_success_examples"}),
	).Do()
	if err != nil {
		t.Fatal(err)
	}
}

func TestNewGenerationPbPartitionError(t *testing.T) {
	db, err := gorm.Open(generationNamedDialector{Dialector: sqlite.Open(":memory:"), name: gormx.Postgres}, &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	type protoPartitionExample struct {
		ID uint `gorm:"primaryKey"`
	}
	if err := db.AutoMigrate(&protoPartitionExample{}); err != nil {
		t.Fatal(err)
	}

	err = NewGenerationPB(
		db,
		t.TempDir(),
		"api.demo.v1",
		"api/demo/v1;v1",
		WithPBTables([]string{"proto_partition_examples"}),
	).Do()
	if err == nil {
		t.Fatal("expected partition query error")
	}
	if !strings.Contains(err.Error(), "get partition table children") {
		t.Fatalf("unexpected error: %v", err)
	}
}
