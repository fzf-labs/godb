package gen

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/fzf-labs/godb/internal/testenv"
	"github.com/fzf-labs/godb/orm/gormx"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gen"
	"gorm.io/gen/field"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

type fakeColumnType struct {
	name     string
	nullable bool
}

// Name 返回测试列名。
func (f fakeColumnType) Name() string { return f.name }

// DatabaseTypeName 返回测试数据库类型名。
func (f fakeColumnType) DatabaseTypeName() string { return "" }

// ColumnType 返回测试列类型。
func (f fakeColumnType) ColumnType() (string, bool) { return "", false }

// PrimaryKey 返回测试主键信息。
func (f fakeColumnType) PrimaryKey() (bool, bool) { return false, false }

// AutoIncrement 返回测试自增信息。
func (f fakeColumnType) AutoIncrement() (bool, bool) { return false, false }

// Length 返回测试列长度。
func (f fakeColumnType) Length() (int64, bool) { return 0, false }

// DecimalSize 返回测试小数精度。
func (f fakeColumnType) DecimalSize() (int64, int64, bool) { return 0, 0, false }

// Nullable 返回测试列可空信息。
func (f fakeColumnType) Nullable() (bool, bool) { return f.nullable, true }

// Unique 返回测试唯一约束信息。
func (f fakeColumnType) Unique() (bool, bool) { return false, false }

// ScanType 返回测试扫描类型。
func (f fakeColumnType) ScanType() reflect.Type { return nil }

// Comment 返回测试列注释。
func (f fakeColumnType) Comment() (string, bool) { return "", false }

// DefaultValue 返回测试默认值。
func (f fakeColumnType) DefaultValue() (string, bool) { return "", false }

type generationDBExample struct {
	ID   int64  `gorm:"primaryKey"`
	Name string `gorm:"size:64"`
}

// TableName 返回测试模型表名。
func (generationDBExample) TableName() string {
	return "generation_db_examples"
}

func TestGenerationOutputPaths_PreserveAbsoluteBasePath(t *testing.T) {
	daoPath, modelPath, repoPath := generationOutputPaths("/tmp/godb", "demo")
	if daoPath != "/tmp/godb/demo_dao" {
		t.Fatalf("unexpected dao path: %s", daoPath)
	}
	if modelPath != "/tmp/godb/demo_model" {
		t.Fatalf("unexpected model path: %s", modelPath)
	}
	if repoPath != "/tmp/godb/demo_repo" {
		t.Fatalf("unexpected repo path: %s", repoPath)
	}
}

func TestNewGenerationDBOptions(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	dataMap := map[string]func(columnType gorm.ColumnType) string{
		"custom": func(gorm.ColumnType) string { return "string" },
	}
	dbNameOpt := func(*gorm.DB) string { return "demo_db" }
	generateModelOpt := func(*gen.Generator) map[string]any { return map[string]any{"users": struct{}{}} }

	generation := NewGenerationDB(
		db,
		"/tmp/out",
		WithOutRepo(),
		WithTables([]string{"users"}),
		WithDataMap(dataMap),
		WithDBOpts(ModelOptionRemoveDefault()),
		WithDBNameOpts(dbNameOpt),
		WithGenerateModel(generateModelOpt),
		WithFieldNullable(),
	)

	if generation.db != db {
		t.Fatal("expected db to be assigned")
	}
	if generation.outPutPath != "/tmp/out" {
		t.Fatalf("unexpected output path: %s", generation.outPutPath)
	}
	if generation.genRepo {
		t.Fatal("WithOutRepo should disable repo generation")
	}
	if len(generation.tables) != 1 || generation.tables[0] != "users" {
		t.Fatalf("unexpected tables: %#v", generation.tables)
	}
	if generation.dataMap["custom"](fakeColumnType{}) != "string" {
		t.Fatal("custom data map was not assigned")
	}
	if len(generation.opts) != 1 {
		t.Fatalf("expected one model option, got %d", len(generation.opts))
	}
	if generation.dbNameOpt == nil || generation.dbNameOpt(db) != "demo_db" {
		t.Fatal("db name option was not assigned")
	}
	if generation.generateModelOpt == nil || len(generation.generateModelOpt(nil)) != 1 {
		t.Fatal("generate model option was not assigned")
	}
	if !generation.fieldNullable {
		t.Fatal("field nullable option was not assigned")
	}
}

func TestGetDBNameAppliesOverrideAndTablePrefix(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{TablePrefix: "tenant_"},
	})
	require.NoError(t, err)

	if got := GetDBName(db, func(*gorm.DB) string { return "app" }); got != "tenant_app" {
		t.Fatalf("unexpected db name: %s", got)
	}
	if got := GetDBName(db, func(*gorm.DB) string { return "tenant_app" }); got != "tenant_app" {
		t.Fatalf("db name should not be double-prefixed: %s", got)
	}
}

func TestJSONTagNameStrategy(t *testing.T) {
	if got := JSONTagNameStrategy("AdminID"); got != "adminId" {
		t.Fatalf("unexpected json tag: %s", got)
	}
	if got := JSONTagNameStrategy("data_type_json"); got != "dataTypeJson" {
		t.Fatalf("unexpected snake json tag: %s", got)
	}
}

func TestModelOptionsModifyFields(t *testing.T) {
	underlined := applyModelOption(t, ModelOptionUnderline("UL"), map[string]any{
		"Name":       "_id",
		"ColumnName": "_id",
		"Tag":        field.Tag{},
	})
	if got := stringField(t, underlined, "Name"); got != "ULid" {
		t.Fatalf("unexpected underlined name: %s", got)
	}
	if got := tagField(t, underlined, "Tag"); got[field.TagKeyJson] != "_id" {
		t.Fatalf("unexpected json tag: %#v", got)
	}

	unchanged := applyModelOption(t, ModelOptionUnderline("UL"), map[string]any{
		"Name":       "ID",
		"ColumnName": "id",
		"Tag":        field.Tag{},
	})
	if got := stringField(t, unchanged, "Name"); got != "ID" {
		t.Fatalf("non-underscore field should be unchanged: %s", got)
	}

	withDefault := applyModelOption(t, ModelOptionPgDefaultString(), map[string]any{
		"GORMTag": field.GormTag{"default": []string{"'active'::character varying"}},
	})
	if got := gormTagField(t, withDefault).Build(); !strings.Contains(got, "default:active") {
		t.Fatalf("expected postgres default string to be normalized, got %q", got)
	}

	withoutType := applyModelOption(t, ModelOptionRemoveGormTypeTag(), map[string]any{
		"GORMTag": field.GormTag{"type": []string{"uuid"}, "column": []string{"id"}},
	})
	if got := gormTagField(t, withoutType).Build(); strings.Contains(got, "type:") {
		t.Fatalf("expected type tag to be removed, got %q", got)
	}

	withoutDefault := applyModelOption(t, ModelOptionRemoveDefault(), map[string]any{
		"GORMTag": field.GormTag{"default": []string{"gen_random_uuid()"}},
	})
	if got := gormTagField(t, withoutDefault).Build(); strings.Contains(got, "default:") {
		t.Fatalf("expected non-primary default to be removed, got %q", got)
	}
	primaryDefault := applyModelOption(t, ModelOptionRemoveDefault(), map[string]any{
		"GORMTag": field.GormTag{"primaryKey": nil, "default": []string{"gen_random_uuid()"}},
	})
	if got := gormTagField(t, primaryDefault).Build(); !strings.Contains(got, "default:gen_random_uuid()") {
		t.Fatalf("expected primary default to be preserved, got %q", got)
	}
}

func applyModelOption(t *testing.T, opt gen.ModelOpt, fields map[string]any) reflect.Value {
	t.Helper()
	operatorValues := reflect.ValueOf(opt).MethodByName("Operator").Call(nil)
	if len(operatorValues) != 1 {
		t.Fatalf("unexpected operator return count: %d", len(operatorValues))
	}
	operator := operatorValues[0]
	fieldValue := reflect.New(operator.Type().In(0).Elem())
	for name, value := range fields {
		target := fieldValue.Elem().FieldByName(name)
		if !target.IsValid() {
			t.Fatalf("unknown field %s", name)
		}
		target.Set(reflect.ValueOf(value))
	}
	return operator.Call([]reflect.Value{fieldValue})[0]
}

func stringField(t *testing.T, value reflect.Value, name string) string {
	t.Helper()
	return value.Elem().FieldByName(name).String()
}

func tagField(t *testing.T, value reflect.Value, name string) field.Tag {
	t.Helper()
	return value.Elem().FieldByName(name).Interface().(field.Tag)
}

func gormTagField(t *testing.T, value reflect.Value) field.GormTag {
	t.Helper()
	return value.Elem().FieldByName("GORMTag").Interface().(field.GormTag)
}

func TestDataTypeMap(t *testing.T) {
	mapping := DataTypeMap()

	tests := []struct {
		dbType   string
		column   fakeColumnType
		expected string
	}{
		{dbType: "json", expected: "datatypes.JSON"},
		{dbType: "jsonb", expected: "datatypes.JSON"},
		{dbType: "timestamptz", column: fakeColumnType{name: "deleted_at"}, expected: "gorm.DeletedAt"},
		{dbType: "timestamptz", column: fakeColumnType{name: "created_at", nullable: true}, expected: SQLNullTime},
		{dbType: "timestamptz", column: fakeColumnType{name: "created_at"}, expected: TimeTime},
		{dbType: "interval[]", expected: "pq.StringArray"},
		{dbType: "bytea[]", expected: "pq.ByteaArray"},
		{dbType: "\"char\"[]", expected: "pq.StringArray"},
		{dbType: "character varying[]", expected: "pq.StringArray"},
		{dbType: "text[]", expected: "pq.StringArray"},
		{dbType: "uuid[]", expected: "pq.StringArray"},
		{dbType: "json[]", expected: "pq.StringArray"},
		{dbType: "jsonb[]", expected: "pq.StringArray"},
		{dbType: "xml[]", expected: "pq.StringArray"},
		{dbType: "numeric[]", expected: "pq.Float64Array"},
		{dbType: "smallint[]", expected: "pq.Int32Array"},
		{dbType: "integer[]", expected: "pq.Int32Array"},
		{dbType: "bigint[]", expected: "pq.Int64Array"},
		{dbType: "real[]", expected: "pq.Float32Array"},
		{dbType: "double precision[]", expected: "pq.Float64Array"},
		{dbType: "boolean[]", expected: "pq.BoolArray"},
		{dbType: "date[]", expected: "pq.StringArray"},
		{dbType: "time without time zone[]", expected: "pq.StringArray"},
		{dbType: "time with time zone[]", expected: "pq.StringArray"},
		{dbType: "timestamp without time zone[]", expected: "pq.StringArray"},
		{dbType: "timestamp with time zone[]", expected: "pq.StringArray"},
		{dbType: "timestamp[]", expected: "pq.StringArray"},
		{dbType: "timestamptz[]", expected: "pq.StringArray"},
	}

	for _, tt := range tests {
		t.Run(tt.dbType, func(t *testing.T) {
			if got := mapping[tt.dbType](tt.column); got != tt.expected {
				t.Fatalf("unexpected type: got %s want %s", got, tt.expected)
			}
		})
	}
}

func TestDBNameOptsNormalizesDatabaseName(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "gorm-gen test.db")
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	require.NoError(t, err)

	got := DBNameOpts()(db)
	if strings.ContainsAny(got, "- ") {
		t.Fatalf("expected normalized db name, got %s", got)
	}
}

func TestGenerationPBOptions(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	pb := NewGenerationPB(
		db,
		t.TempDir(),
		"pb",
		"example.com/project/pb;pb",
		WithPBOpts(ModelOptionRemoveDefault()),
		WithPBTables([]string{"users"}),
	)

	if len(pb.opts) != 1 {
		t.Fatalf("expected one pb option, got %d", len(pb.opts))
	}
	if len(pb.tables) != 1 || pb.tables[0] != "users" {
		t.Fatalf("unexpected pb tables: %#v", pb.tables)
	}
}

func TestGenerationDBDoWithSQLiteWithoutRepo(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&generationDBExample{}))
	outDir := t.TempDir()

	err = NewGenerationDB(
		db,
		outDir,
		WithOutRepo(),
		WithTables([]string{"generation_db_examples"}),
		WithDataMap(map[string]func(gorm.ColumnType) string{
			"integer": func(gorm.ColumnType) string { return "int64" },
		}),
		WithDBOpts(ModelOptionRemoveDefault()),
		WithDBNameOpts(func(*gorm.DB) string { return "demo" }),
		WithFieldNullable(),
	).Do()
	require.NoError(t, err)

	for _, path := range []string{
		filepath.Join(outDir, "demo_dao", "gen.go"),
		filepath.Join(outDir, "demo_model", "generation_db_examples.gen.go"),
	} {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("expected generated file %s: %v", path, err)
		}
	}
}

func TestGenerationDBDoErrorBranches(t *testing.T) {
	t.Run("panic is returned as error", func(t *testing.T) {
		err := NewGenerationDB(nil, t.TempDir()).Do()
		require.Error(t, err)
		require.Contains(t, err.Error(), "generate db code panic")
	})

	t.Run("partition query error", func(t *testing.T) {
		db, err := gorm.Open(generationNamedDialector{Dialector: sqlite.Open(":memory:"), name: gormx.Postgres}, &gorm.Config{})
		require.NoError(t, err)
		require.NoError(t, db.AutoMigrate(&generationDBExample{}))

		err = NewGenerationDB(
			db,
			t.TempDir(),
			WithOutRepo(),
			WithTables([]string{"generation_db_examples"}),
			WithDBNameOpts(func(*gorm.DB) string { return "demo" }),
		).Do()
		require.Error(t, err)
		require.Contains(t, err.Error(), "get partition table children")
	})

	t.Run("repo path is a file", func(t *testing.T) {
		db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
		require.NoError(t, err)
		require.NoError(t, db.AutoMigrate(&generationDBExample{}))
		outDir := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(outDir, "demo_repo"), []byte("not a directory"), 0600))

		err = NewGenerationDB(
			db,
			outDir,
			WithTables([]string{"generation_db_examples"}),
			WithDBNameOpts(func(*gorm.DB) string { return "demo" }),
		).Do()
		require.Error(t, err)
		require.Contains(t, err.Error(), "create repo path")
	})
}

type generationNamedDialector struct {
	gorm.Dialector
	name string
}

// Name 返回测试包装后的方言名称。
func (d generationNamedDialector) Name() string {
	return d.name
}

func TestGenerationDBDoWithSQLiteRepo(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&generationDBExample{}))
	outDir := t.TempDir()

	err = NewGenerationDB(
		db,
		outDir,
		WithTables([]string{"generation_db_examples"}),
		WithDBNameOpts(func(*gorm.DB) string { return "demo" }),
	).Do()
	require.NoError(t, err)

	for _, path := range []string{
		filepath.Join(outDir, "demo_dao", "gen.go"),
		filepath.Join(outDir, "demo_model", "generation_db_examples.gen.go"),
		filepath.Join(outDir, "demo_repo", "generation_db_examples.repo.go"),
	} {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("expected generated file %s: %v", path, err)
		}
	}
}

// TestGenerationPostgresWithOutRepo 验证不生成 repo 的 PostgreSQL 代码生成。
func TestGenerationPostgresWithOutRepo(t *testing.T) {
	// 初始化数据库
	db, err := gormx.NewSimpleGormClient(gormx.Postgres, testenv.PostgresDSN("gorm_gen"))
	if err != nil {
		testenv.SkipIfUnavailable(t, "postgres unavailable: %v", err)
	}
	// 生成代码
	err = NewGenerationDB(
		db,
		"../example/gorm/postgres/",
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
	// 初始化数据库
	db, err := gormx.NewSimpleGormClient(gormx.Postgres, testenv.PostgresDSN("gorm_gen"))
	if err != nil {
		testenv.SkipIfUnavailable(t, "postgres unavailable: %v", err)
	}
	// 生成代码
	err = NewGenerationDB(
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
	if err != nil {
		t.Fatal(err)
	}
}

// TestGenerationPostgresFieldNullable 验证可空字段的 PostgreSQL 代码生成。
func TestGenerationPostgresFieldNullable(t *testing.T) {
	// 初始化数据库
	db, err := gormx.NewSimpleGormClient(gormx.Postgres, testenv.PostgresDSN("gorm_gen"))
	if err != nil {
		testenv.SkipIfUnavailable(t, "postgres unavailable: %v", err)
	}
	// 生成代码
	err = NewGenerationDB(
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
	if err != nil {
		t.Fatal(err)
	}
}
