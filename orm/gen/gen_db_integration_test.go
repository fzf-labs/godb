//go:build integration

package gen

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/gen"
	"gorm.io/gen/field"
	"gorm.io/gorm"

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

// TestGenerationPostgresWithOutRepo 验证不生成 repo 的 PostgreSQL 代码生成。
func TestGenerationPostgresWithOutRepo(t *testing.T) {
	db := newIntegrationPostgresDB(t, "gorm_gen")
	outDir := newModuleDir(t)
	err := NewGenerationDB(
		db,
		outDir,
		WithOutRepo(),
		WithDBNameOpts(DBNameOpts()),
		WithTables([]string{"admin_demo", "admin_log_demo", "admin_role_demo"}),
		WithDataMap(DataTypeMap()), // 设置数据类型映射
		WithDBOpts(ModelOptionRemoveDefault(), ModelOptionPgDefaultString(), ModelOptionRemoveGormTypeTag(), ModelOptionUnderline("UL")), // 设置自定义选项
		WithFieldNullable(),
	).Do()
	if err != nil {
		t.Fatal(err)
	}
}

// TestGenerationPostgres 验证 PostgreSQL 模型和 repo 代码生成。
func TestGenerationPostgres(t *testing.T) {
	db := newIntegrationPostgresDB(t, "gorm_gen")
	outDir := newModuleDir(t)
	err := NewGenerationDB(
		db,
		outDir,
		WithGenerateModel(func(g *gen.Generator) map[string]any { // 设置表关联关系(1对多,多对多...)
			adminLogDemo := g.GenerateModel("admin_log_demo")
			adminRoleDemo := g.GenerateModel("admin_role_demo",
				gen.FieldRelate(field.Many2Many, "Admins", g.GenerateModel("admin_demo"),
					&field.RelateConfig{
						RelateSlicePointer: true,
						JSONTag:            JSONTagNameStrategy("Admins"),
						GORMTag:            field.GormTag{"joinForeignKey": []string{"role_id"}, "joinReferences": []string{"admin_id"}, "many2many": []string{"admin_to_role_demo"}},
					},
				),
			)
			adminDemo := g.GenerateModel("admin_demo",
				gen.FieldRelate(field.HasMany, "AdminLogDemos", adminLogDemo,
					&field.RelateConfig{
						RelateSlicePointer: true,
						JSONTag:            JSONTagNameStrategy("AdminLogDemos"),
						GORMTag:            field.GormTag{"foreignKey": []string{"admin_id"}},
					},
				),
				gen.FieldRelate(field.Many2Many, "AdminRoles", adminRoleDemo,
					&field.RelateConfig{
						RelateSlicePointer: true,
						JSONTag:            JSONTagNameStrategy("AdminRoles"),
						GORMTag:            field.GormTag{"joinForeignKey": []string{"admin_id"}, "joinReferences": []string{"role_id"}, "many2many": []string{"admin_to_role_demo"}},
					},
				),
			)
			return map[string]any{
				"admin_demo":      adminDemo,
				"admin_log_demo":  adminLogDemo,
				"admin_role_demo": adminRoleDemo,
			}
		}),
		WithDataMap(DataTypeMap()), // 设置数据类型映射
		WithDBOpts(ModelOptionRemoveDefault(), ModelOptionUnderline("UL")), // 设置自定义选项
	).Do()
	if err != nil {
		t.Fatal(err)
	}
}

// TestGenerationPostgresFieldNullable 验证可空字段的 PostgreSQL 代码生成。
func TestGenerationPostgresFieldNullable(t *testing.T) {
	db := newIntegrationPostgresDB(t, "gorm_gen")
	outDir := newModuleDir(t)
	err := NewGenerationDB(
		db,
		outDir,
		WithGenerateModel(func(g *gen.Generator) map[string]any { // 设置表关联关系(1对多,多对多...)
			adminLogDemo := g.GenerateModel("admin_log_demo")
			AdminRoleDemo := g.GenerateModel("admin_role_demo",
				gen.FieldRelate(field.Many2Many, "Admins", g.GenerateModel("admin_demo"),
					&field.RelateConfig{
						RelateSlicePointer: true,
						JSONTag:            JSONTagNameStrategy("Admins"),
						GORMTag:            field.GormTag{"joinForeignKey": []string{"role_id"}, "joinReferences": []string{"admin_id"}, "many2many": []string{"admin_to_role_demo"}},
					},
				),
			)
			adminDemo := g.GenerateModel("admin_demo",
				gen.FieldRelate(field.HasMany, "AdminLogDemos", adminLogDemo,
					&field.RelateConfig{
						RelateSlicePointer: true,
						JSONTag:            JSONTagNameStrategy("AdminLogDemos"),
						GORMTag:            field.GormTag{"foreignKey": []string{"admin_id"}},
					},
				),
				gen.FieldRelate(field.Many2Many, "AdminRoles", AdminRoleDemo,
					&field.RelateConfig{
						RelateSlicePointer: true,
						JSONTag:            JSONTagNameStrategy("AdminRoles"),
						GORMTag:            field.GormTag{"joinForeignKey": []string{"admin_id"}, "joinReferences": []string{"role_id"}, "many2many": []string{"admin_to_role_demo"}},
					},
				),
			)
			return map[string]any{
				"admin_demo":      adminDemo,
				"admin_log_demo":  adminLogDemo,
				"admin_role_demo": AdminRoleDemo,
			}
		}),
		WithDataMap(DataTypeMap()), // 设置数据类型映射
		WithDBOpts(ModelOptionRemoveDefault(), ModelOptionUnderline("UL")), // 设置自定义选项
		WithFieldNullable(),
	).Do()
	if err != nil {
		t.Fatal(err)
	}
}
