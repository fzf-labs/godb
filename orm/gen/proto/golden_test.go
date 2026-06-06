package proto_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/fzf-labs/godb/internal/testenv"
	ormgen "github.com/fzf-labs/godb/orm/gen"
	"github.com/fzf-labs/godb/orm/gormx"
	"github.com/stretchr/testify/require"
)

func TestGenerationPBGoldenMatchesExampleUserDemo(t *testing.T) {
	db, err := gormx.NewSimpleGormClient(gormx.Postgres, testenv.PostgresDSN("gorm_gen"))
	if err != nil {
		testenv.SkipIfUnavailable(t, "postgres unavailable: %v", err)
	}

	outDir := t.TempDir()
	err = ormgen.NewGenerationPB(
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
