package gormx

import (
	"regexp"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type indexExample struct {
	ID       string `gorm:"primaryKey"`
	Email    string `gorm:"uniqueIndex"`
	TenantID int64  `gorm:"index:idx_index_example_tenant_dept,priority:1"`
	DeptID   int64  `gorm:"index:idx_index_example_tenant_dept,priority:2"`
}

func TestIndexHelpersRejectNilDB(t *testing.T) {
	if _, err := GetIndexes(nil, "users"); err == nil {
		t.Fatal("expected GetIndexes to reject nil db")
	}
	if _, err := SortIndexColumns(nil, "users"); err == nil {
		t.Fatal("expected SortIndexColumns to reject nil db")
	}
}

// TableName 返回索引测试模型表名。
func (indexExample) TableName() string {
	return "index_examples"
}

func TestBuildPgIndexesQuery_UsesPlaceholder(t *testing.T) {
	query, args := buildPgIndexesQuery(`user'name`)
	if !strings.Contains(query, "t.relname=?") {
		t.Fatalf("expected placeholder query, got %q", query)
	}
	if len(args) != 1 || args[0] != `user'name` {
		t.Fatalf("unexpected args: %#v", args)
	}
}

func TestBuildPGSortIndexColumnsQuery_UsesPlaceholder(t *testing.T) {
	query, args := buildPGSortIndexColumnsQuery(`user'name`)
	if !strings.Contains(query, "t.relname=?") {
		t.Fatalf("expected placeholder query, got %q", query)
	}
	if len(args) != 1 || args[0] != `user'name` {
		t.Fatalf("unexpected args: %#v", args)
	}
}

func TestBuildPGSortIndexColumnsQuery_IncludesPartitionedTables(t *testing.T) {
	query, _ := buildPGSortIndexColumnsQuery("users")
	if !strings.Contains(query, "t.relkind IN ('r','p')") {
		t.Fatalf("expected partitioned tables in query, got %q", query)
	}
}

func TestGetIndexesAndSortIndexColumnsDefaultDialect(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&indexExample{}); err != nil {
		t.Fatal(err)
	}

	indexes, err := GetIndexes(db, "index_examples")
	if err != nil {
		t.Fatalf("get indexes: %v", err)
	}
	if len(indexes) == 0 {
		t.Fatal("expected sqlite indexes")
	}
	var sawEmail bool
	for _, index := range indexes {
		if index.ColumnName == "email" && index.IsUnique {
			sawEmail = true
		}
	}
	if !sawEmail {
		t.Fatalf("expected unique email index, got %#v", indexes)
	}

	columns, err := SortIndexColumns(db, "index_examples")
	if err != nil {
		t.Fatalf("sort index columns: %v", err)
	}
	if got := columns["idx_index_example_tenant_dept"]; strings.Join(got, ",") != "tenant_id,dept_id" {
		t.Fatalf("unexpected compound index order: %#v", got)
	}
}

func TestPostgresIndexQueriesReturnErrors(t *testing.T) {
	db := openNamedSQLite(t, Postgres)

	if _, err := GetIndexes(db, "users"); err == nil {
		t.Fatal("expected postgres index query error")
	}
	if _, err := SortIndexColumns(db, "users"); err == nil {
		t.Fatal("expected postgres sort index query error")
	}
}

func TestIndexQueriesRejectUnsupportedDialects(t *testing.T) {
	db := openNamedSQLite(t, "oracle")

	if _, err := GetIndexes(db, "users"); err == nil {
		t.Fatal("expected unsupported index dialect error")
	}
	if _, err := SortIndexColumns(db, "users"); err == nil {
		t.Fatal("expected unsupported sort dialect error")
	}
}

func TestPostgresIndexQueriesWithMock(t *testing.T) {
	db, mock := openMockPostgres(t)

	query, _ := buildPgIndexesQuery("users")
	mock.ExpectQuery(regexp.QuoteMeta(strings.Replace(query, "?", "$1", 1))).
		WithArgs("users").
		WillReturnRows(sqlmock.NewRows([]string{"table_name", "index_name", "column_name", "is_unique", "primary"}).
			AddRow("users", "users_pkey", "id", true, true).
			AddRow("users", "idx_users_email", "email", true, false))

	indexes, err := GetIndexes(db, "users")
	if err != nil {
		t.Fatalf("unexpected indexes error: %v", err)
	}
	if len(indexes) != 2 || !indexes[0].Primary || indexes[1].ColumnName != "email" {
		t.Fatalf("unexpected indexes: %#v", indexes)
	}

	sortQuery, _ := buildPGSortIndexColumnsQuery("users")
	mock.ExpectQuery(regexp.QuoteMeta(strings.Replace(sortQuery, "?", "$1", 1))).
		WithArgs("users").
		WillReturnRows(sqlmock.NewRows([]string{"table_name", "index_name", "column_name"}).
			AddRow("users", "idx_users_tenant_dept", "tenant_id").
			AddRow("users", "idx_users_tenant_dept", "dept_id"))

	columns, err := SortIndexColumns(db, "users")
	if err != nil {
		t.Fatalf("unexpected sort error: %v", err)
	}
	if got := strings.Join(columns["idx_users_tenant_dept"], ","); got != "tenant_id,dept_id" {
		t.Fatalf("unexpected sorted columns: %#v", columns)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
