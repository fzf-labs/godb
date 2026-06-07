package repo_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/fzf-labs/godb/internal/testenv"
	ormgen "github.com/fzf-labs/godb/orm/gen"
	"github.com/fzf-labs/godb/orm/gormx"
)

func TestGenerationDBGoldenMatchesExampleUserDemoRepo(t *testing.T) {
	db, err := gormx.NewSimpleGormClient(gormx.Postgres, testenv.PostgresDSN("gorm_gen"))
	if err != nil {
		testenv.SkipIfUnavailable(t, "postgres unavailable: %v", err)
	}

	workspace := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(workspace, "go.mod"), []byte("module github.com/fzf-labs/godb\ngo 1.24\n"), 0600))

	outDir := filepath.Join(workspace, "orm", "example", "gorm", "postgres")
	err = ormgen.NewGenerationDB(
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
