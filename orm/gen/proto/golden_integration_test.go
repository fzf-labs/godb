//go:build integration

package proto_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	ormgen "github.com/fzf-labs/godb/orm/gen"
	"github.com/fzf-labs/godb/orm/gormx"
)

func newIntegrationPostgresDB(t *testing.T, dbname string) *gorm.DB {
	t.Helper()
	envOrDefault := func(key, fallback string) string {
		if value := os.Getenv(key); value != "" {
			return value
		}
		return fallback
	}
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable connect_timeout=1 TimeZone=Asia/Shanghai",
		envOrDefault("PGHOST", "127.0.0.1"),
		envOrDefault("PGPORT", "5432"),
		envOrDefault("PGUSER", "postgres"),
		envOrDefault("PGPASSWORD", "123456"),
		dbname,
	)
	db, err := gormx.NewSimpleGormClient(gormx.Postgres, dsn)
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, sqlDB.Close())
	})
	return db
}

func TestGenerationPBGoldenMatchesExampleUserDemo(t *testing.T) {
	db := newIntegrationPostgresDB(t, "gorm_gen")

	outDir := t.TempDir()
	err := ormgen.NewGenerationPB(
		db,
		outDir,
		"api.gorm_gen.v1",
		"api/gorm_gen/v1;v1",
		ormgen.WithPBOpts(ormgen.ModelOptionRemoveDefault(), ormgen.ModelOptionUnderline("ul_")),
		ormgen.WithPBTables([]string{"user_demo"}),
	).Do()
	require.NoError(t, err)

	got, err := os.ReadFile(filepath.Join(outDir, "user_demo.proto"))
	require.NoError(t, err)
	want, err := os.ReadFile(filepath.Join("..", "..", "example", "pb", "user_demo.proto"))
	require.NoError(t, err)

	require.Equal(t, string(want), string(got))
}
