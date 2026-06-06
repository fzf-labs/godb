package proto

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/fzf-labs/godb/internal/testenv"
	"github.com/fzf-labs/godb/orm/gormx"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// newDB 创建 proto 生成测试用数据库连接。
func newDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gormx.NewSimpleGormClient(gormx.Postgres, testenv.PostgresDSN("gorm_gen"))
	if err != nil {
		testenv.SkipIfUnavailable(t, "postgres unavailable: %v", err)
	}
	return db
}

type protoExample struct {
	ID        string `gorm:"primaryKey;size:36;comment:ID"`
	Name      string `gorm:"size:20;not null;comment:Name"`
	Status    int16  `gorm:"not null;comment:Status"`
	CreatedAt time.Time
	DeletedAt gorm.DeletedAt
}

// TableName 返回 proto 生成测试模型表名。
func (protoExample) TableName() string {
	return "proto_examples"
}

func newSQLiteProtoDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&protoExample{}))
	return db
}

func newProtoForTest(t *testing.T) *Proto {
	t.Helper()
	db := newSQLiteProtoDB(t)
	return &Proto{
		gorm:                db,
		outPutPath:          t.TempDir(),
		packageStr:          "api.demo.v1",
		goPackageStr:        "api/demo/v1;v1",
		tableName:           "proto_examples",
		tableNameComment:    "Example table",
		tableNameUnderScore: "proto_examples",
		lowerTableName:      "protoExample",
		upperTableName:      "ProtoExample",
		columnNameToName: map[string]string{
			"id":         "ID",
			"name":       "Name",
			"status":     "Status",
			"created_at": "CreatedAt",
			"deleted_at": "DeletedAt",
		},
		columnNameToDataType: map[string]string{
			"id":         "string",
			"name":       "string",
			"status":     "int16",
			"created_at": "time.Time",
			"deleted_at": "gorm.DeletedAt",
		},
	}
}

func TestProtoOutputWritesAndRejectsExistingFile(t *testing.T) {
	p := newProtoForTest(t)
	out := filepath.Join(t.TempDir(), "demo.proto")
	require.NoError(t, p.output(out, "syntax = \"proto3\";\n"))
	content, err := os.ReadFile(out)
	require.NoError(t, err)
	if string(content) != "syntax = \"proto3\";\n" {
		t.Fatalf("unexpected content: %q", string(content))
	}
	if err := p.output(out, "again"); err == nil {
		t.Fatal("expected existing file error")
	}

	blockingFile := filepath.Join(t.TempDir(), "not-a-dir")
	require.NoError(t, os.WriteFile(blockingFile, []byte("x"), 0600))
	if err := p.output(filepath.Join(blockingFile, "demo.proto"), "content"); err == nil {
		t.Fatal("expected mkdir error")
	}
}

func TestProtoTemplateSections(t *testing.T) {
	p := newProtoForTest(t)
	if got := p.genSyntax(); !strings.Contains(got, `syntax = "proto3"`) {
		t.Fatalf("unexpected syntax: %s", got)
	}
	if got := p.genPackage(); !strings.Contains(got, "package api.demo.v1") {
		t.Fatalf("unexpected package: %s", got)
	}
	if got := p.genImport(); !strings.Contains(got, "buf/validate/validate.proto") {
		t.Fatalf("unexpected imports: %s", got)
	}
	if got := p.genOption(); !strings.Contains(got, `option go_package = "api/demo/v1;v1"`) {
		t.Fatalf("unexpected option: %s", got)
	}
	service, err := p.genService()
	require.NoError(t, err)
	for _, want := range []string{"service ProtoExample", "/api/demo/v1/proto_examples", "UpdateProtoExampleStatus"} {
		if !strings.Contains(service, want) {
			t.Fatalf("expected %q in service:\n%s", want, service)
		}
	}
	message, err := p.genMessage()
	require.NoError(t, err)
	for _, want := range []string{"message ProtoExampleInfo", "message CreateProtoExampleReq", "message UpdateProtoExampleStatusReq", "status"} {
		if !strings.Contains(message, want) {
			t.Fatalf("expected %q in message:\n%s", want, message)
		}
	}
}

func TestGenerationPBWithSQLite(t *testing.T) {
	db := newSQLiteProtoDB(t)
	outDir := t.TempDir()
	err := GenerationPB(db, outDir, "api.demo.v1", "api/demo/v1;v1", "proto_examples",
		map[string]string{
			"id":         "ID",
			"name":       "Name",
			"status":     "Status",
			"created_at": "CreatedAt",
			"deleted_at": "DeletedAt",
		},
		map[string]string{
			"id":         "string",
			"name":       "string",
			"status":     "int16",
			"created_at": "time.Time",
			"deleted_at": "gorm.DeletedAt",
		},
	)
	require.NoError(t, err)
	content, err := os.ReadFile(filepath.Join(outDir, "proto_examples.proto"))
	require.NoError(t, err)
	if !strings.Contains(string(content), "message ProtoExample") {
		t.Fatalf("unexpected proto content:\n%s", string(content))
	}

	err = GenerationPB(db, outDir, "api.demo.v1", "api/demo/v1;v1", "proto_examples",
		map[string]string{
			"id":         "ID",
			"name":       "Name",
			"status":     "Status",
			"created_at": "CreatedAt",
			"deleted_at": "DeletedAt",
		},
		map[string]string{
			"id":         "string",
			"name":       "string",
			"status":     "int16",
			"created_at": "time.Time",
			"deleted_at": "gorm.DeletedAt",
		},
	)
	if err == nil {
		t.Fatal("expected existing output error")
	}
}

func TestProtoNameHelpers(t *testing.T) {
	p := newProtoForTest(t)
	if got := p.upperName("api_client"); got != "APIClient" {
		t.Fatalf("unexpected upper name: %s", got)
	}
	if got := p.lowerName("api_client"); got != "APIClient" {
		t.Fatalf("initialism prefix should be preserved: %s", got)
	}
	if got := p.lowerName("user_demo"); got != "userDemo" {
		t.Fatalf("unexpected lower name: %s", got)
	}
}

func TestLowerFieldName(t *testing.T) {
	tests := map[string]string{
		"ID":       "id",
		"UserID":   "userId",
		"URLValue": "URLValue",
		"Type":     "_type",
	}
	for input, expected := range tests {
		if got := lowerFieldName(input); got != expected {
			t.Fatalf("lowerFieldName(%q)=%q want %q", input, got, expected)
		}
	}
}

func TestDataTypeToPbType(t *testing.T) {
	tests := map[string]string{
		"int":       "int32",
		"int64":     "int32",
		"uint":      "uint32",
		"uint64":    "uint32",
		"float32":   "float",
		"float64":   "double",
		"bool":      "bool",
		"string":    "string",
		"time.Time": "string",
		"[]byte":    "bytes",
		"custom":    "string",
	}
	for input, expected := range tests {
		if got := dataTypeToPbType(input); got != expected {
			t.Fatalf("dataTypeToPbType(%q)=%q want %q", input, got, expected)
		}
	}
}

func TestPBTypeToValidate(t *testing.T) {
	tests := []struct {
		name     string
		pbType   string
		isNull   bool
		length   int64
		contains string
	}{
		{name: "required string without max", pbType: "string", contains: "min_len: 1"},
		{name: "required string with max", pbType: "string", length: 20, contains: "max_len: 20"},
		{name: "nullable string without max", pbType: "string", isNull: true, contains: "IGNORE_IF_UNPOPULATED"},
		{name: "nullable string with max", pbType: "string", isNull: true, length: 20, contains: "IGNORE_IF_UNPOPULATED"},
		{name: "nullable int", pbType: "int32", isNull: true, contains: "IGNORE_IF_UNPOPULATED"},
		{name: "required int", pbType: "int32", contains: ""},
		{name: "nullable int64", pbType: "int64", isNull: true, contains: "IGNORE_IF_UNPOPULATED"},
		{name: "nullable float", pbType: "float", isNull: true, contains: "IGNORE_IF_UNPOPULATED"},
		{name: "nullable double", pbType: "double", isNull: true, contains: "IGNORE_IF_UNPOPULATED"},
		{name: "required timestamp", pbType: "google.protobuf.Timestamp", contains: "required=true"},
		{name: "nullable timestamp", pbType: "google.protobuf.Timestamp", isNull: true, contains: "IGNORE_IF_UNPOPULATED"},
		{name: "unknown", pbType: "bytes", contains: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := pbTypeToValidate(tt.pbType, tt.isNull, tt.length)
			if tt.contains == "" {
				if got != "" {
					t.Fatalf("expected empty validation, got %q", got)
				}
				return
			}
			if !strings.Contains(got, tt.contains) {
				t.Fatalf("expected %q in %q", tt.contains, got)
			}
		})
	}
}

func TestJoinWithQuotes(t *testing.T) {
	if got := joinWithQuotes(nil); got != "" {
		t.Fatalf("expected empty join, got %q", got)
	}
	if got := joinWithQuotes([]string{"id", "name"}); got != `"id","name"` {
		t.Fatalf("unexpected join: %s", got)
	}
}

// TestGenerationPB 验证从数据库表生成 proto。
func TestGenerationPB(t *testing.T) {
	db := newDB(t)
	type args struct {
		db                   *gorm.DB
		outPutPath           string
		packageStr           string
		goPackageStr         string
		table                string
		columnNameToName     map[string]string
		columnNameToDataType map[string]string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "test",
			args: args{
				db:           db,
				outPutPath:   "../../example/pb",
				packageStr:   "api.gorm_gen.v1",
				goPackageStr: "api/gorm_gen/v1;v1",
				table:        "admin_log_demo",
				columnNameToName: map[string]string{
					"id":         "ID",
					"admin_id":   "adminID",
					"ip":         "IP",
					"uri":        "URI",
					"useragent":  "Useragent",
					"header":     "Header",
					"req":        "Req",
					"resp":       "Resp",
					"created_at": "CreatedAt",
					"status":     "Status",
				},
				columnNameToDataType: map[string]string{
					"id":        "int64",
					"admin_id":  "int64",
					"ip":        "string",
					"uri":       "string",
					"useragent": "string",
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := GenerationPB(tt.args.db, tt.args.outPutPath, tt.args.packageStr, tt.args.goPackageStr, tt.args.table, tt.args.columnNameToName, tt.args.columnNameToDataType); (err != nil) != tt.wantErr {
				t.Errorf("GenerationPB() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
