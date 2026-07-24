//go:build integration

package repo_test

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

func TestGenerationDBGoldenMatchesExampleUserDemoRepo(t *testing.T) {
	db := newIntegrationPostgresDB(t, "gorm_gen")

	workspace := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(workspace, "go.mod"), []byte("module github.com/fzf-labs/godb\ngo 1.24\n"), 0600))

	outDir := filepath.Join(workspace, "orm", "example", "gorm", "postgres")
	err := ormgen.NewGenerationDB(
		db,
		outDir,
		ormgen.WithTables([]string{"user_demo"}),
		ormgen.WithDataMap(ormgen.DataTypeMap()),
		ormgen.WithDBOpts(ormgen.ModelOptionRemoveDefault(), ormgen.ModelOptionUnderline("UL")),
	).Do()
	require.NoError(t, err)

	got, err := os.ReadFile(filepath.Join(outDir, "gorm_gen_repo", "user_demo.repo.go"))
	require.NoError(t, err)
	want, err := os.ReadFile(filepath.Join("..", "..", "example", "gorm", "postgres", "gorm_gen_repo", "user_demo.repo.go"))
	require.NoError(t, err)

	require.Equal(t, string(want), string(got))
}
