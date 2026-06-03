package gen

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fzf-labs/godb/orm/gormx"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// TestNewGenerationPb 验证 proto 文件生成。
func TestNewGenerationPb(t *testing.T) {
	db, err := gormx.NewSimpleGormClient(gormx.Postgres, "host=0.0.0.0 port=5432 user=postgres password=123456 dbname=gorm_gen sslmode=disable TimeZone=Asia/Shanghai")
	if err != nil {
		t.Skipf("postgres unavailable: %v", err)
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
