package gen

import (
	"testing"

	"github.com/fzf-labs/godb/orm/gormx"
	"gorm.io/gen"
	"gorm.io/gen/field"
)

func TestGenerationPostgresWithOutRepo(t *testing.T) {
	// 初始化数据库
	db := gormx.NewSimpleGormClient(gormx.Postgres, "host=0.0.0.0 port=5432 user=postgres password=123456 dbname=gorm_gen sslmode=disable TimeZone=Asia/Shanghai")
	if db == nil {
		return
	}
	// 生成代码
	NewGenerationDB(
		db,
		"../example/gorm/postgres/",
		WithOutRepo(),
		WithDBNameOpts(DBNameOpts()),
		WithTables([]string{"admin_demo", "admin_log_demo", "admin_role_demo"}),
		WithDataMap(DataTypeMap()), // 设置数据类型映射
		WithDBOpts(ModelOptionRemoveDefault(), ModelOptionPgDefaultString(), ModelOptionRemoveGormTypeTag(), ModelOptionUnderline("UL")), // 设置自定义选项
		WithFieldNullable(),
	).Do()
}

func TestGenerationPostgres(t *testing.T) {
	// 初始化数据库
	db := gormx.NewSimpleGormClient(gormx.Postgres, "host=0.0.0.0 port=5432 user=postgres password=123456 dbname=gorm_gen sslmode=disable TimeZone=Asia/Shanghai")
	if db == nil {
		return
	}
	// 生成代码
	NewGenerationDB(
		db,
		"../example/gorm/postgres/",
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
}

func TestGenerationPostgresFieldNullable(t *testing.T) {
	// 初始化数据库
	db := gormx.NewSimpleGormClient(gormx.Postgres, "host=0.0.0.0 port=5432 user=postgres password=123456 dbname=gorm_gen sslmode=disable TimeZone=Asia/Shanghai")
	if db == nil {
		return
	}
	// 生成代码
	NewGenerationDB(
		db,
		"../example/gorm/postgres/",
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
}
