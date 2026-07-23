package repo

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"

	"github.com/fzf-labs/godb/internal/testenv"
	"github.com/fzf-labs/godb/orm/gormx"
	tpl "github.com/fzf-labs/godb/orm/utils/template"
)

// newDB 创建 repo 生成测试用数据库连接。
func newDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gormx.NewSimpleGormClient(gormx.Postgres, testenv.PostgresDSN("gorm_gen"))
	if err != nil {
		testenv.SkipIfUnavailable(t, "postgres unavailable: %v", err)
	}
	testenv.CleanupGormDB(t, db)
	return db
}

func TestExtractFormatErrorLine(t *testing.T) {
	line, ok := extractFormatErrorLine(nil)
	if ok || line != 0 {
		t.Fatalf("expected nil error to return no line, got ok=%v line=%d", ok, line)
	}

	line, ok = extractFormatErrorLine(errors.New(filepath.Join("/tmp", "bad.go") + ":12:7: expected ';'"))
	if !ok || line != 12 {
		t.Fatalf("unexpected result: ok=%v line=%d", ok, line)
	}

	line, ok = extractFormatErrorLine(errors.New("imports failed"))
	if ok || line != 0 {
		t.Fatalf("expected no line, got ok=%v line=%d", ok, line)
	}
}

func TestExtractFormatErrorLineInvalidNumbers(t *testing.T) {
	if line, ok := extractFormatErrorLine(errors.New("bad.go:0:1: nope")); ok || line != 0 {
		t.Fatalf("expected invalid line to be ignored, got ok=%v line=%d", ok, line)
	}
}

func TestFormatErrorContext_ClampsAroundRequestedLine(t *testing.T) {
	content := []byte(strings.Join([]string{
		"package demo",
		"func one() {}",
		"func two() {}",
		"func three() {}",
		"func four() {}",
	}, "\n"))

	got := formatErrorContext(content, 3)
	for _, want := range []string{"1 package demo", "3 func two() {}", "5 func four() {}"} {
		if !strings.Contains(got, want) {
			t.Fatalf("expected %q in context:\n%s", want, got)
		}
	}

	got = formatErrorContext(content, 1)
	if !strings.Contains(got, "1 package demo") {
		t.Fatalf("expected context to clamp around the first line, got:\n%s", got)
	}
}

func TestRepoOutput_InvalidSourceReturnsError(t *testing.T) {
	r := &Repo{}
	err := r.output(filepath.Join(t.TempDir(), "broken.repo.go"), []byte("package repo\nfunc (\n"))
	if err == nil {
		t.Fatal("expected format error, got nil")
	}
}

func TestRepoOutput_WritesFormattedSource(t *testing.T) {
	r := &Repo{}
	out := filepath.Join(t.TempDir(), "ok.repo.go")
	if err := r.output(out, []byte("package repo\nfunc Answer()int{return 42}\n")); err != nil {
		t.Fatalf("unexpected output error: %v", err)
	}
	content, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	if !strings.Contains(string(content), "func Answer() int") {
		t.Fatalf("expected formatted function, got:\n%s", string(content))
	}
}

func TestRepoOutput_WriteError(t *testing.T) {
	r := &Repo{}
	err := r.output(filepath.Join(t.TempDir(), "missing", "ok.repo.go"), []byte("package repo\nfunc Answer() int { return 42 }\n"))
	if err == nil {
		t.Fatal("expected write error for missing directory")
	}
}

func TestGenerationTableRejectsEmptyTable(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}

	err = GenerationTable(db, "gorm_gen", t.TempDir(), t.TempDir(), t.TempDir(), "", nil, map[string]string{}, map[string]string{}, map[string]string{})
	if err == nil {
		t.Fatal("expected empty table error")
	}
}

func TestGenerationTableRejectsNilDB(t *testing.T) {
	err := GenerationTable(nil, "gorm_gen", t.TempDir(), t.TempDir(), t.TempDir(), "users", nil, map[string]string{}, map[string]string{}, map[string]string{})
	if err == nil || !strings.Contains(err.Error(), "repo generation db cannot be nil") {
		t.Fatalf("expected repo-specific nil db error, got %v", err)
	}
}

func TestGenerationTableRejectsMissingPackagePaths(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}

	err = GenerationTable(db, "gorm_gen", t.TempDir(), t.TempDir(), t.TempDir(), "users", nil, map[string]string{}, map[string]string{}, map[string]string{})
	if err == nil || !strings.Contains(err.Error(), "cannot resolve dao/model package path") {
		t.Fatalf("expected package path error, got %v", err)
	}
}

type emptySchemaNamer struct {
	schema.NamingStrategy
}

func (emptySchemaNamer) SchemaName(string) string {
	return ""
}

type emptyNameExample struct {
	ID int64 `gorm:"primaryKey"`
}

func (emptyNameExample) TableName() string {
	return "empty_name_examples"
}

func TestGenerationTableRejectsEmptyGoIdentifier(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		NamingStrategy: emptySchemaNamer{},
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&emptyNameExample{}); err != nil {
		t.Fatal(err)
	}

	err = GenerationTable(db, "gorm_gen", ".", ".", t.TempDir(), "empty_name_examples", nil, map[string]string{"id": "int64"}, map[string]string{"id": "ID"}, map[string]string{"id": "Int64"})
	if err == nil || !strings.Contains(err.Error(), "cannot convert table") {
		t.Fatalf("expected empty identifier error, got %v", err)
	}
}

func TestUpsertOneCacheByFieldsTemplatesGuardNilData(t *testing.T) {
	params := map[string]any{
		"firstTableChar": "u",
		"dbName":         "gorm_gen",
		"upperTableName": "UserDemo",
	}
	rendered, err := tpl.NewTemplate().Parse(UpsertOneCacheByFields).Execute(params)
	if err != nil {
		t.Fatalf("unexpected render error: %v", err)
	}
	if !strings.Contains(rendered.String(), "if data == nil {") {
		t.Fatalf("template missing nil guard: %s", rendered.String())
	}

	rendered, err = tpl.NewTemplate().Parse(UpsertOneCacheByFieldsTx).Execute(params)
	if err != nil {
		t.Fatalf("unexpected render error: %v", err)
	}
	if !strings.Contains(rendered.String(), "if data == nil {") {
		t.Fatalf("template missing nil guard: %s", rendered.String())
	}
}

func newTemplateRepo() *Repo {
	return &Repo{
		gorm:           &gorm.DB{Config: &gorm.Config{NamingStrategy: schema.NamingStrategy{}}},
		daoPath:        "../dao",
		modelPath:      "../model",
		repoPath:       "../repo",
		table:          "user_demo",
		dbName:         "gorm_gen",
		firstTableChar: "u",
		lowerTableName: "userDemo",
		upperTableName: "UserDemo",
		daoPkgPath:     "github.com/fzf-labs/godb/example/dao",
		modelPkgPath:   "github.com/fzf-labs/godb/example/model",
		index: []DBIndex{
			{Name: "pk_user", ColumnKey: "id", PrimaryKey: true, Unique: true, Columns: []string{"id"}},
			{Name: "uidx_username", ColumnKey: "username", Unique: true, Columns: []string{"username"}},
			{Name: "uidx_tenant_dept", ColumnKey: "tenant_id:dept_id", Unique: true, Columns: []string{"tenant_id", "dept_id"}},
			{Name: "idx_status", ColumnKey: "status", Columns: []string{"status"}},
			{Name: "idx_nickname", ColumnKey: "nickname", Columns: []string{"nickname"}},
			{Name: "idx_org_role", ColumnKey: "org_id:role_id", Columns: []string{"org_id", "role_id"}},
			{Name: "idx_metadata", ColumnKey: "metadata", Columns: []string{"metadata"}},
		},
		haveDeletedAt: true,
		columnNameToName: map[string]string{
			"id":         "ID",
			"username":   "Username",
			"tenant_id":  "TenantID",
			"dept_id":    "DeptID",
			"status":     "Status",
			"nickname":   "Nickname",
			"org_id":     "OrgID",
			"role_id":    "RoleID",
			"metadata":   "Metadata",
			"dao":        "Dao",
			"cacheKey":   "CacheKey",
			"deleted_at": "DeletedAt",
		},
		columnNameToDataType: map[string]string{
			"id":         "string",
			"username":   "string",
			"tenant_id":  "int64",
			"dept_id":    "int64",
			"status":     "bool",
			"nickname":   "string",
			"org_id":     "string",
			"role_id":    "string",
			"metadata":   "datatypes.JSON",
			"dao":        "string",
			"cacheKey":   "string",
			"deleted_at": "gorm.DeletedAt",
		},
		columnNameToFieldType: map[string]string{
			"id":         "String",
			"username":   "String",
			"tenant_id":  "Int64",
			"dept_id":    "Int64",
			"status":     "Bool",
			"nickname":   "String",
			"org_id":     "String",
			"role_id":    "String",
			"metadata":   "Field",
			"dao":        "String",
			"cacheKey":   "String",
			"deleted_at": "Field",
		},
	}
}

func TestRepoFieldHelpers(t *testing.T) {
	r := newTemplateRepo()

	if got := r.upperFields([]string{"id", "tenant_id"}); got != "IDTenantID" {
		t.Fatalf("unexpected upper fields: %s", got)
	}
	if got := r.fieldAndDataTypes([]string{"username", "tenant_id"}); got != "username string,tenantID int64" {
		t.Fatalf("unexpected typed params: %s", got)
	}
	if got := r.cacheFields([]string{"id", "username"}); got != "IDUsername" {
		t.Fatalf("unexpected cache fields: %s", got)
	}
	if got := r.cacheFieldsJoin([]string{"username", "tenant_id"}); got != "username,tenantID" {
		t.Fatalf("unexpected cache join: %s", got)
	}
	if got := r.lowerFieldName("dao"); got != "_dao" {
		t.Fatalf("keyword field should be prefixed, got %s", got)
	}
	if got := r.lowerFieldName("id"); got != "ID" {
		t.Fatalf("initialism field should be preserved, got %s", got)
	}
	if got := r.lowerFieldName("cacheKey"); got != "_cacheKey" {
		t.Fatalf("reserved helper name should be prefixed, got %s", got)
	}
	if got := r.plural("status"); got != "statuses" {
		t.Fatalf("unexpected plural: %s", got)
	}
	if got := r.plural("data"); got != "dataplural" {
		t.Fatalf("unchanged plural should get suffix, got %s", got)
	}
	if got := r.plural(""); got != "" {
		t.Fatalf("empty plural should stay empty, got %s", got)
	}
	if got := r.lowerName(""); got != "" {
		t.Fatalf("empty lower name should stay empty, got %s", got)
	}
	if !r.checkDaoFieldType([]string{"metadata"}) {
		t.Fatal("expected metadata Field type to be skipped")
	}
	if r.checkDaoFieldType([]string{"username"}) {
		t.Fatal("string field should not be skipped")
	}
	if got := r.whereFields([]string{"status", "username"}); got != "dao.Status.Is(status),dao.Username.Eq(username)" {
		t.Fatalf("unexpected where fields: %s", got)
	}
	if got := r.primaryKeyWhereFields([]string{"status", "id"}); got != "dao.Status.Is(data.Status),dao.ID.Eq(data.ID)" {
		t.Fatalf("unexpected primary where fields: %s", got)
	}
}

func TestHasDeletedAt(t *testing.T) {
	if !hasDeletedAt(map[string]string{"deleted_at": "gorm.DeletedAt"}) {
		t.Fatal("expected gorm deleted-at to be detected")
	}
	if !hasDeletedAt(map[string]string{"deleted_at": "soft_delete.DeletedAt"}) {
		t.Fatal("expected soft-delete deleted-at to be detected")
	}
	if hasDeletedAt(map[string]string{"deleted_at": "time.Time"}) {
		t.Fatal("plain time should not be treated as deleted-at marker")
	}
}

func TestRepoTemplateSectionsCoverIndexBranches(t *testing.T) {
	r := newTemplateRepo()
	tests := []struct {
		name     string
		fn       func() (string, error)
		contains []string
	}{
		{
			name: "package",
			fn:   r.generatePkg,
			contains: []string{
				"package gorm_gen_repo",
			},
		},
		{
			name: "imports",
			fn:   r.generateImport,
			contains: []string{
				`"github.com/fzf-labs/godb/example/dao"`,
				`"github.com/fzf-labs/godb/example/model"`,
			},
		},
		{
			name: "vars",
			fn:   r.generateVar,
			contains: []string{
				"CacheUserDemoByIDPrefix",
				"CacheUserDemoByTenantIDDeptIDPrefix",
				"CacheUserDemoUnscopedByConditionPrefix",
			},
		},
		{
			name: "interface types",
			fn:   r.generateTypes,
			contains: []string{
				"type (",
				"FindOneByUsername",
				"FindOneByTenantIDDeptID",
				"FindMultiByStatus",
				"DeleteIndexCache",
			},
		},
		{
			name: "constructor",
			fn:   r.generateNew,
			contains: []string{
				"func NewUserDemoRepo",
			},
		},
		{
			name: "common funcs",
			fn:   r.generateCommonFunc,
			contains: []string{
				"func (u *UserDemoRepo) NewData",
				"func (u *UserDemoRepo) DeepCopy",
			},
		},
		{
			name: "create funcs",
			fn:   r.generateCreateFunc,
			contains: []string{
				"CreateOneCache",
				"UpsertOneCacheByFieldsTx",
				"dao.ID.Eq(data.ID)",
			},
		},
		{
			name: "update funcs",
			fn:   r.generateUpdateFunc,
			contains: []string{
				"UpdateOneWithZero",
				"UpdateBatchByTenantIDDeptID",
				"UpdateBatchByNicknames",
			},
		},
		{
			name: "read funcs",
			fn:   r.generateReadFunc,
			contains: []string{
				"FindOneCacheByUsername",
				"FindMultiCacheByStatus",
				"FindMultiCacheByOrgIDRoleID",
				"FindMultiCacheByCondition",
			},
		},
		{
			name: "delete funcs",
			fn:   r.generateDelFunc,
			contains: []string{
				"DeleteOneCacheByUsername",
				"DeleteMultiCacheByNicknames",
				"DeleteMultiCacheByOrgIDRoleID",
				"DeleteIndexCache",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.fn()
			if err != nil {
				t.Fatalf("unexpected generate error: %v", err)
			}
			for _, want := range tt.contains {
				if !strings.Contains(got, want) {
					t.Fatalf("expected %q in output:\n%s", want, got)
				}
			}
		})
	}
}

func TestRepoTemplateErrorBranches(t *testing.T) {
	r := newTemplateRepo()
	badTemplate := "{{"
	assertTemplateError := func(t *testing.T, target *string, fn func() (string, error)) {
		t.Helper()
		old := *target
		*target = badTemplate
		defer func() { *target = old }()

		if _, err := fn(); err == nil {
			t.Fatal("expected template error")
		}
	}

	tests := []struct {
		name   string
		target *string
		fn     func() (string, error)
	}{
		{name: "pkg", target: &Pkg, fn: r.generatePkg},
		{name: "import", target: &Import, fn: r.generateImport},
		{name: "var cache global", target: &VarCacheGlobal, fn: r.generateVar},
		{name: "var cache", target: &VarCache, fn: r.generateVar},
		{name: "var", target: &Var, fn: r.generateVar},
		{name: "var cache keys", target: &VarCacheKeys, fn: r.generateVar},
		{name: "interface new data", target: &InterfaceNewData, fn: r.generateCommonMethods},
		{name: "interface deep copy", target: &InterfaceDeepCopy, fn: r.generateCommonMethods},
		{name: "interface create one", target: &InterfaceCreateOne, fn: r.generateCreateMethods},
		{name: "interface create one cache", target: &InterfaceCreateOneCache, fn: r.generateCreateMethods},
		{name: "interface create one by tx", target: &InterfaceCreateOneByTx, fn: r.generateCreateMethods},
		{name: "interface create one cache by tx", target: &InterfaceCreateOneCacheByTx, fn: r.generateCreateMethods},
		{name: "interface create batch", target: &InterfaceCreateBatch, fn: r.generateCreateMethods},
		{name: "interface create batch cache", target: &InterfaceCreateBatchCache, fn: r.generateCreateMethods},
		{name: "interface create batch by tx", target: &InterfaceCreateBatchByTx, fn: r.generateCreateMethods},
		{name: "interface create batch cache by tx", target: &InterfaceCreateBatchCacheByTx, fn: r.generateCreateMethods},
		{name: "interface upsert one", target: &InterfaceUpsertOne, fn: r.generateCreateMethods},
		{name: "interface upsert one cache", target: &InterfaceUpsertOneCache, fn: r.generateCreateMethods},
		{name: "interface upsert one by tx", target: &InterfaceUpsertOneByTx, fn: r.generateCreateMethods},
		{name: "interface upsert one cache by tx", target: &InterfaceUpsertOneCacheByTx, fn: r.generateCreateMethods},
		{name: "interface upsert one by fields", target: &InterfaceUpsertOneByFields, fn: r.generateCreateMethods},
		{name: "interface upsert one cache by fields", target: &InterfaceUpsertOneCacheByFields, fn: r.generateCreateMethods},
		{name: "interface upsert one by fields tx", target: &InterfaceUpsertOneByFieldsTx, fn: r.generateCreateMethods},
		{name: "interface upsert one cache by fields tx", target: &InterfaceUpsertOneCacheByFieldsTx, fn: r.generateCreateMethods},
		{name: "interface update one", target: &InterfaceUpdateOne, fn: r.generateUpdateMethods},
		{name: "interface update one cache", target: &InterfaceUpdateOneCache, fn: r.generateUpdateMethods},
		{name: "interface update one by tx", target: &InterfaceUpdateOneByTx, fn: r.generateUpdateMethods},
		{name: "interface update one cache by tx", target: &InterfaceUpdateOneCacheByTx, fn: r.generateUpdateMethods},
		{name: "interface update one with zero", target: &InterfaceUpdateOneWithZero, fn: r.generateUpdateMethods},
		{name: "interface update one cache with zero", target: &InterfaceUpdateOneCacheWithZero, fn: r.generateUpdateMethods},
		{name: "interface update one with zero by tx", target: &InterfaceUpdateOneWithZeroByTx, fn: r.generateUpdateMethods},
		{name: "interface update one cache with zero by tx", target: &InterfaceUpdateOneCacheWithZeroByTx, fn: r.generateUpdateMethods},
		{name: "interface update batch by fields", target: &InterfaceUpdateBatchByFields, fn: r.generateUpdateMethods},
		{name: "interface update batch by fields tx", target: &InterfaceUpdateBatchByFieldsTx, fn: r.generateUpdateMethods},
		{name: "interface update batch by field plural", target: &InterfaceUpdateBatchByFieldPlural, fn: r.generateUpdateMethods},
		{name: "interface update batch by field plural tx", target: &InterfaceUpdateBatchByFieldPluralTx, fn: r.generateUpdateMethods},
		{name: "interface find one by field", target: &InterfaceFindOneByField, fn: r.generateReadMethods},
		{name: "interface find one cache by field", target: &InterfaceFindOneCacheByField, fn: r.generateReadMethods},
		{name: "interface find multi by field plural", target: &InterfaceFindMultiByFieldPlural, fn: r.generateReadMethods},
		{name: "interface find multi cache by field plural unique true", target: &InterfaceFindMultiCacheByFieldPluralUniqueTrue, fn: r.generateReadMethods},
		{name: "interface find one by fields", target: &InterfaceFindOneByFields, fn: r.generateReadMethods},
		{name: "interface find one cache by fields", target: &InterfaceFindOneCacheByFields, fn: r.generateReadMethods},
		{name: "interface find multi by field", target: &InterfaceFindMultiByField, fn: r.generateReadMethods},
		{name: "interface find multi cache by field", target: &InterfaceFindMultiCacheByField, fn: r.generateReadMethods},
		{name: "interface find multi cache by field plural unique false", target: &InterfaceFindMultiCacheByFieldPluralUniqueFalse, fn: r.generateReadMethods},
		{name: "interface find multi by fields", target: &InterfaceFindMultiByFields, fn: r.generateReadMethods},
		{name: "interface find multi cache by fields", target: &InterfaceFindMultiCacheByFields, fn: r.generateReadMethods},
		{name: "interface find multi by condition", target: &InterfaceFindMultiByCondition, fn: r.generateReadMethods},
		{name: "interface find multi by cache condition", target: &InterfaceFindMultiByCacheCondition, fn: r.generateReadMethods},
		{name: "interface delete one by field", target: &InterfaceDeleteOneByField, fn: r.generateDelMethods},
		{name: "interface delete one cache by field", target: &InterfaceDeleteOneCacheByField, fn: r.generateDelMethods},
		{name: "interface delete one by field tx", target: &InterfaceDeleteOneByFieldTx, fn: r.generateDelMethods},
		{name: "interface delete one cache by field tx", target: &InterfaceDeleteOneCacheByFieldTx, fn: r.generateDelMethods},
		{name: "interface delete multi by field plural", target: &InterfaceDeleteMultiByFieldPlural, fn: r.generateDelMethods},
		{name: "interface delete multi cache by field plural", target: &InterfaceDeleteMultiCacheByFieldPlural, fn: r.generateDelMethods},
		{name: "interface delete multi by field plural tx", target: &InterfaceDeleteMultiByFieldPluralTx, fn: r.generateDelMethods},
		{name: "interface delete multi cache by field plural tx", target: &InterfaceDeleteMultiCacheByFieldPluralTx, fn: r.generateDelMethods},
		{name: "interface delete one by fields", target: &InterfaceDeleteOneByFields, fn: r.generateDelMethods},
		{name: "interface delete one cache by fields", target: &InterfaceDeleteOneCacheByFields, fn: r.generateDelMethods},
		{name: "interface delete one by fields tx", target: &InterfaceDeleteOneByFieldsTx, fn: r.generateDelMethods},
		{name: "interface delete one cache by fields tx", target: &InterfaceDeleteOneCacheByFieldsTx, fn: r.generateDelMethods},
		{name: "interface delete multi by field", target: &InterfaceDeleteMultiByField, fn: r.generateDelMethods},
		{name: "interface delete multi cache by field", target: &InterfaceDeleteMultiCacheByField, fn: r.generateDelMethods},
		{name: "interface delete multi by field tx", target: &InterfaceDeleteMultiByFieldTx, fn: r.generateDelMethods},
		{name: "interface delete multi cache by field tx", target: &InterfaceDeleteMultiCacheByFieldTx, fn: r.generateDelMethods},
		{name: "interface delete multi by fields", target: &InterfaceDeleteMultiByFields, fn: r.generateDelMethods},
		{name: "interface delete multi cache by fields", target: &InterfaceDeleteMultiCacheByFields, fn: r.generateDelMethods},
		{name: "interface delete multi by fields tx", target: &InterfaceDeleteMultiByFieldsTx, fn: r.generateDelMethods},
		{name: "interface delete multi cache by fields tx", target: &InterfaceDeleteMultiCacheByFieldsTx, fn: r.generateDelMethods},
		{name: "interface delete index cache", target: &InterfaceDeleteIndexCache, fn: r.generateDelMethods},
		{name: "types", target: &Types, fn: r.generateTypes},
		{name: "new", target: &New, fn: r.generateNew},
		{name: "new data", target: &NewData, fn: r.generateCommonFunc},
		{name: "deep copy", target: &DeepCopy, fn: r.generateCommonFunc},
		{name: "create one", target: &CreateOne, fn: r.generateCreateFunc},
		{name: "create one cache", target: &CreateOneCache, fn: r.generateCreateFunc},
		{name: "create one by tx", target: &CreateOneByTx, fn: r.generateCreateFunc},
		{name: "create one cache by tx", target: &CreateOneCacheByTx, fn: r.generateCreateFunc},
		{name: "create batch", target: &CreateBatch, fn: r.generateCreateFunc},
		{name: "create batch cache", target: &CreateBatchCache, fn: r.generateCreateFunc},
		{name: "create batch by tx", target: &CreateBatchByTx, fn: r.generateCreateFunc},
		{name: "create batch cache by tx", target: &CreateBatchCacheByTx, fn: r.generateCreateFunc},
		{name: "upsert one", target: &UpsertOne, fn: r.generateCreateFunc},
		{name: "upsert one cache", target: &UpsertOneCache, fn: r.generateCreateFunc},
		{name: "upsert one by tx", target: &UpsertOneByTx, fn: r.generateCreateFunc},
		{name: "upsert one cache by tx", target: &UpsertOneCacheByTx, fn: r.generateCreateFunc},
		{name: "upsert one by fields", target: &UpsertOneByFields, fn: r.generateCreateFunc},
		{name: "upsert one cache by fields", target: &UpsertOneCacheByFields, fn: r.generateCreateFunc},
		{name: "upsert one by fields tx", target: &UpsertOneByFieldsTx, fn: r.generateCreateFunc},
		{name: "upsert one cache by fields tx", target: &UpsertOneCacheByFieldsTx, fn: r.generateCreateFunc},
		{name: "update one", target: &UpdateOne, fn: r.generateUpdateFunc},
		{name: "update one cache", target: &UpdateOneCache, fn: r.generateUpdateFunc},
		{name: "update one by tx", target: &UpdateOneByTx, fn: r.generateUpdateFunc},
		{name: "update one cache by tx", target: &UpdateOneCacheByTx, fn: r.generateUpdateFunc},
		{name: "update one with zero", target: &UpdateOneWithZero, fn: r.generateUpdateFunc},
		{name: "update one cache with zero", target: &UpdateOneCacheWithZero, fn: r.generateUpdateFunc},
		{name: "update one with zero by tx", target: &UpdateOneWithZeroByTx, fn: r.generateUpdateFunc},
		{name: "update one cache with zero by tx", target: &UpdateOneCacheWithZeroByTx, fn: r.generateUpdateFunc},
		{name: "update batch by fields", target: &UpdateBatchByFields, fn: r.generateUpdateFunc},
		{name: "update batch by fields tx", target: &UpdateBatchByFieldsTx, fn: r.generateUpdateFunc},
		{name: "update batch by field plural", target: &UpdateBatchByFieldPlural, fn: r.generateUpdateFunc},
		{name: "update batch by field plural tx", target: &UpdateBatchByFieldPluralTx, fn: r.generateUpdateFunc},
		{name: "find one by field", target: &FindOneByField, fn: r.generateReadFunc},
		{name: "find one cache by field", target: &FindOneCacheByField, fn: r.generateReadFunc},
		{name: "find multi by field plural", target: &FindMultiByFieldPlural, fn: r.generateReadFunc},
		{name: "find multi cache by field plural unique true", target: &FindMultiCacheByFieldPluralUniqueTrue, fn: r.generateReadFunc},
		{name: "find one by fields", target: &FindOneByFields, fn: r.generateReadFunc},
		{name: "find one cache by fields", target: &FindOneCacheByFields, fn: r.generateReadFunc},
		{name: "find multi by field", target: &FindMultiByField, fn: r.generateReadFunc},
		{name: "find multi cache by field", target: &FindMultiCacheByField, fn: r.generateReadFunc},
		{name: "find multi cache by field plural unique false", target: &FindMultiCacheByFieldPluralUniqueFalse, fn: r.generateReadFunc},
		{name: "find multi by fields", target: &FindMultiByFields, fn: r.generateReadFunc},
		{name: "find multi cache by fields", target: &FindMultiCacheByFields, fn: r.generateReadFunc},
		{name: "find multi by condition", target: &FindMultiByCondition, fn: r.generateReadFunc},
		{name: "find multi by cache condition", target: &FindMultiByCacheCondition, fn: r.generateReadFunc},
		{name: "var cache del key", target: &VarCacheDelKey, fn: r.generateDelFunc},
		{name: "delete one by field", target: &DeleteOneByField, fn: r.generateDelFunc},
		{name: "delete one cache by field", target: &DeleteOneCacheByField, fn: r.generateDelFunc},
		{name: "delete one by field tx", target: &DeleteOneByFieldTx, fn: r.generateDelFunc},
		{name: "delete one cache by field tx", target: &DeleteOneCacheByFieldTx, fn: r.generateDelFunc},
		{name: "delete multi by field plural", target: &DeleteMultiByFieldPlural, fn: r.generateDelFunc},
		{name: "delete multi cache by field plural", target: &DeleteMultiCacheByFieldPlural, fn: r.generateDelFunc},
		{name: "delete multi by field plural tx", target: &DeleteMultiByFieldPluralTx, fn: r.generateDelFunc},
		{name: "delete multi cache by field plural tx", target: &DeleteMultiCacheByFieldPluralTx, fn: r.generateDelFunc},
		{name: "delete one by fields", target: &DeleteOneByFields, fn: r.generateDelFunc},
		{name: "delete one cache by fields", target: &DeleteOneCacheByFields, fn: r.generateDelFunc},
		{name: "delete one by fields tx", target: &DeleteOneByFieldsTx, fn: r.generateDelFunc},
		{name: "delete one cache by fields tx", target: &DeleteOneCacheByFieldsTx, fn: r.generateDelFunc},
		{name: "delete multi by field", target: &DeleteMultiByField, fn: r.generateDelFunc},
		{name: "delete multi cache by field", target: &DeleteMultiCacheByField, fn: r.generateDelFunc},
		{name: "delete multi by field tx", target: &DeleteMultiByFieldTx, fn: r.generateDelFunc},
		{name: "delete multi cache by field tx", target: &DeleteMultiCacheByFieldTx, fn: r.generateDelFunc},
		{name: "delete multi by fields", target: &DeleteMultiByFields, fn: r.generateDelFunc},
		{name: "delete multi cache by fields", target: &DeleteMultiCacheByFields, fn: r.generateDelFunc},
		{name: "delete multi by fields tx", target: &DeleteMultiByFieldsTx, fn: r.generateDelFunc},
		{name: "delete multi cache by fields tx", target: &DeleteMultiCacheByFieldsTx, fn: r.generateDelFunc},
		{name: "delete index cache", target: &DeleteIndexCache, fn: r.generateDelFunc},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertTemplateError(t, tt.target, tt.fn)
		})
	}
}

func TestGenerateTypesPropagatesSectionErrors(t *testing.T) {
	r := newTemplateRepo()
	badTemplate := "{{"
	assertTemplateError := func(t *testing.T, target *string) {
		t.Helper()
		old := *target
		*target = badTemplate
		defer func() { *target = old }()

		if _, err := r.generateTypes(); err == nil {
			t.Fatal("expected generateTypes error")
		}
	}

	tests := []struct {
		name   string
		target *string
	}{
		{name: "common", target: &InterfaceNewData},
		{name: "create", target: &InterfaceCreateOne},
		{name: "update", target: &InterfaceUpdateOne},
		{name: "read", target: &InterfaceFindOneByField},
		{name: "delete", target: &InterfaceDeleteOneByField},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertTemplateError(t, tt.target)
		})
	}
}

type repoIndexModel struct {
	ID       string `gorm:"primaryKey"`
	Email    string `gorm:"uniqueIndex"`
	TenantID int64  `gorm:"index:idx_repo_tenant_dept,priority:1"`
	DeptID   int64  `gorm:"index:idx_repo_tenant_dept,priority:2"`
	Status   bool   `gorm:"index"`
}

// TableName 返回 repo 索引测试模型表名。
func (repoIndexModel) TableName() string {
	return "repo_index_models"
}

func TestRepoProcessIndex_SortsAndDeduplicatesIndexes(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&repoIndexModel{}); err != nil {
		t.Fatal(err)
	}
	r := &Repo{gorm: db, table: "repo_index_models"}

	indexes, err := r.processIndex()
	if err != nil {
		t.Fatalf("process indexes: %v", err)
	}
	if len(indexes) == 0 {
		t.Fatal("expected indexes")
	}
	var hasPrimary, hasUnique, hasLeftmost bool
	for _, idx := range indexes {
		switch {
		case idx.PrimaryKey && len(idx.Columns) == 1 && idx.Columns[0] == "id":
			hasPrimary = true
		case idx.Unique && len(idx.Columns) == 1 && idx.Columns[0] == "email":
			hasUnique = true
		case idx.ColumnKey == "tenant_id":
			hasLeftmost = true
		}
	}
	if !hasPrimary || !hasUnique || !hasLeftmost {
		t.Fatalf("missing expected index categories: primary=%v unique=%v leftmost=%v indexes=%+v", hasPrimary, hasUnique, hasLeftmost, indexes)
	}
}

func TestRepoProcessIndexUsesPartitionChildTable(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&repoIndexModel{}); err != nil {
		t.Fatal(err)
	}
	r := &Repo{gorm: db, table: "parent_table", partitionTables: []string{"repo_index_models"}}
	indexes, err := r.processIndex()
	if err != nil {
		t.Fatalf("process indexes: %v", err)
	}
	if len(indexes) == 0 {
		t.Fatal("expected child table indexes")
	}
}

func TestGenerationTableReturnsIndexError(t *testing.T) {
	db, err := gorm.Open(namedRepoDialector{Dialector: sqlite.Open(":memory:"), name: gormx.Postgres}, &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	err = GenerationTable(db, "demo", "dao", "model", t.TempDir(), "missing", nil, nil, nil, nil)
	if err == nil {
		t.Fatal("expected index query error")
	}
}

type namedRepoDialector struct {
	gorm.Dialector
	name string
}

// Name 返回测试包装后的 repo 方言名称。
func (d namedRepoDialector) Name() string {
	return d.name
}

type repoGenerationModel struct {
	ID       string `gorm:"primaryKey"`
	Email    string `gorm:"uniqueIndex"`
	TenantID int64  `gorm:"index:idx_repo_generation_tenant_status,priority:1"`
	Status   bool   `gorm:"index:idx_repo_generation_tenant_status,priority:2"`
	Deleted  gorm.DeletedAt
}

// TableName 返回 repo 生成测试模型表名。
func (repoGenerationModel) TableName() string {
	return "repo_generation_models"
}

func TestGenerationTableWithSQLite(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&repoGenerationModel{}); err != nil {
		t.Fatal(err)
	}
	repoPath := t.TempDir()

	err = GenerationTable(
		db,
		"gorm_gen",
		"../../../orm/example/gorm/postgres/gorm_gen_dao",
		"../../../orm/example/gorm/postgres/gorm_gen_model",
		repoPath,
		"repo_generation_models",
		nil,
		map[string]string{
			"id":        "string",
			"email":     "string",
			"tenant_id": "int64",
			"status":    "bool",
			"deleted":   "gorm.DeletedAt",
		},
		map[string]string{
			"id":        "ID",
			"email":     "Email",
			"tenant_id": "TenantID",
			"status":    "Status",
			"deleted":   "Deleted",
		},
		map[string]string{
			"id":        "String",
			"email":     "String",
			"tenant_id": "Int64",
			"status":    "Bool",
			"deleted":   "Field",
		},
	)
	if err != nil {
		t.Fatalf("generation table: %v", err)
	}
	content, err := os.ReadFile(filepath.Join(repoPath, "repo_generation_models.repo.go"))
	if err != nil {
		t.Fatalf("read generated repo: %v", err)
	}
	for _, want := range []string{"func NewRepoGenerationModelRepo", "FindOneByEmail", "FindMultiByTenantIDStatus"} {
		if !strings.Contains(string(content), want) {
			t.Fatalf("expected %q in generated repo:\n%s", want, string(content))
		}
	}
}

func TestJoinIndexColumnKeyUsesColons(t *testing.T) {
	got := joinIndexColumnKey([]string{"a", "b", "c"})
	if got != "a:b:c" {
		t.Fatalf("unexpected key: %s", got)
	}
}

// TestGenerationTable 验证表级 repo 代码生成。
func TestGenerationTable(t *testing.T) {
	db := newDB(t)
	type args struct {
		db                    *gorm.DB
		dbname                string
		daoPath               string
		modelPath             string
		repoPath              string
		table                 string
		partitionTable        []string
		columnNameToDataType  map[string]string
		columnNameToName      map[string]string
		columnNameToFieldType map[string]string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "user_demo",
			args: args{
				db:        db,
				dbname:    "gorm_gen",
				daoPath:   "../../example/gorm/postgres/gorm_gen_dao",
				modelPath: "../../example/gorm/postgres/gorm_gen_model",
				table:     "user_demo",
				columnNameToDataType: map[string]string{
					"password":   "string",
					"avatar":     "string",
					"login_ip":   "string",
					"login_date": "time.Time",
					"created_at": "time.Time",
					"id":         "string",
					"uid":        "string",
					"post_ids":   "string",
					"email":      "string",
					"mobile":     "string",
					"status":     "int16",
					"tenant_id":  "int64",
					"username":   "string",
					"nickname":   "string",
					"dept_id":    "int64",
					"deleted_at": "gorm.DeletedAt",
					"updated_at": "time.Time",
					"remark":     "string",
					"sex":        "int16",
				},
				columnNameToName: map[string]string{
					"login_ip":   "LoginIP",
					"id":         "ID",
					"remark":     "Remark",
					"password":   "Password",
					"nickname":   "Nickname",
					"mobile":     "Mobile",
					"sex":        "Sex",
					"status":     "Status",
					"created_at": "CreatedAt",
					"uid":        "UID",
					"username":   "Username",
					"tenant_id":  "TenantID",
					"updated_at": "UpdatedAt",
					"deleted_at": "DeletedAt",
					"dept_id":    "DeptID",
					"post_ids":   "PostIds",
					"login_date": "LoginDate",
					"email":      "Email",
					"avatar":     "Avatar",
				},
				columnNameToFieldType: map[string]string{
					"post_ids":   "String",
					"email":      "String",
					"status":     "Int16",
					"login_ip":   "String",
					"tenant_id":  "Int64",
					"created_at": "Time",
					"username":   "String",
					"nickname":   "String",
					"dept_id":    "Int64",
					"mobile":     "String",
					"updated_at": "Time",
					"id":         "String",
					"password":   "String",
					"login_date": "Time",
					"deleted_at": "Field",
					"remark":     "String",
					"avatar":     "String",
					"uid":        "String",
					"sex":        "Int16",
				},
			},
			wantErr: false,
		},
		{
			name: "partition_table",
			args: args{
				db:             db,
				dbname:         "gorm_gen",
				daoPath:        "../../example/gorm/postgres/gorm_gen_dao",
				modelPath:      "../../example/gorm/postgres/gorm_gen_model",
				table:          "partition_table",
				partitionTable: []string{},
				columnNameToDataType: map[string]string{
					"id":         "string",
					"user_id":    "string",
					"created_at": "time.Time",
				},
				columnNameToName: map[string]string{
					"id":         "ID",
					"user_id":    "UserID",
					"created_at": "CreatedAt",
				},
				columnNameToFieldType: map[string]string{
					"id":         "String",
					"user_id":    "String",
					"created_at": "Time",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.args.table == "partition_table" {
				if err := tt.args.db.Exec(`
CREATE TABLE IF NOT EXISTS partition_table (
    id uuid PRIMARY KEY,
    user_id uuid NOT NULL,
    created_at timestamptz NOT NULL
)`).Error; err != nil {
					t.Fatalf("create partition_table: %v", err)
				}
			}
			repoPath := t.TempDir()
			if err := GenerationTable(tt.args.db, tt.args.dbname, tt.args.daoPath, tt.args.modelPath, repoPath, tt.args.table, tt.args.partitionTable, tt.args.columnNameToDataType, tt.args.columnNameToName, tt.args.columnNameToFieldType); (err != nil) != tt.wantErr {
				t.Errorf("GenerationTable() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
