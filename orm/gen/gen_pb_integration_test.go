//go:build integration

package gen

import "testing"

// TestNewGenerationPb 验证 proto 文件生成。
func TestNewGenerationPb(t *testing.T) {
	db := newIntegrationPostgresDB(t, "gorm_gen")
	err := NewGenerationPB(
		db,
		t.TempDir(),
		"api.gorm_gen.v1",
		"api/gorm_gen/v1;v1",
		WithPBOpts(ModelOptionRemoveDefault(), ModelOptionUnderline("ul_")),
		WithPBTables([]string{"user_demo"}),
	).Do()
	if err != nil {
		t.Fatal(err)
	}
}
