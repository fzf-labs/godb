package gormx

import (
	"regexp"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestBuildPostgresTableCommentsQuery_IncludesPartitionedTables(t *testing.T) {
	query := buildPostgresTableCommentsQuery()
	if !strings.Contains(query, "c.relkind IN ('r', 'p')") {
		t.Fatalf("unexpected query: %s", query)
	}
}

func TestPartitionAndCommentsDefaultDialect(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	partitions, err := GetPartitionTableToChildTables(db)
	if err != nil {
		t.Fatalf("unexpected partition error: %v", err)
	}
	if partitions != nil {
		t.Fatalf("sqlite should return nil partition map, got %#v", partitions)
	}
	comments, err := GetTableComments(db)
	if err != nil {
		t.Fatalf("unexpected comments error: %v", err)
	}
	if comments != nil {
		t.Fatalf("sqlite should return nil comments, got %#v", comments)
	}
}

func TestPartitionAndCommentsNamedDialectErrors(t *testing.T) {
	for _, dialect := range []string{MySQL, Postgres} {
		t.Run(dialect, func(t *testing.T) {
			db := openNamedSQLite(t, dialect)

			if _, err := GetPartitionTableToChildTables(db); err == nil {
				t.Fatal("expected partition query error")
			}
			if _, err := GetTableComments(db); err == nil {
				t.Fatal("expected table comments query error")
			}
		})
	}
}

func TestMySQLPartitionAndComments(t *testing.T) {
	db, mock := openMockMySQL(t)

	expectMySQLCurrentDatabase(mock, "app")
	mock.ExpectQuery(regexp.QuoteMeta("SELECT DISTINCT TABLE_NAME FROM INFORMATION_SCHEMA.PARTITIONS WHERE PARTITION_NAME IS NOT NULL AND TABLE_SCHEMA=? ORDER BY TABLE_NAME")).
		WithArgs("app").
		WillReturnRows(sqlmock.NewRows([]string{"TABLE_NAME"}).AddRow("orders"))

	partitions, err := GetPartitionTableToChildTables(db)
	if err != nil {
		t.Fatalf("unexpected partition error: %v", err)
	}
	if _, ok := partitions["orders"]; !ok {
		t.Fatalf("expected orders partition entry: %#v", partitions)
	}

	expectMySQLCurrentDatabase(mock, "app")
	mock.ExpectQuery(regexp.QuoteMeta("SELECT TABLE_NAME,TABLE_COMMENT FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_SCHEMA=?")).
		WithArgs("app").
		WillReturnRows(sqlmock.NewRows([]string{"TABLE_NAME", "TABLE_COMMENT"}).AddRow("orders", "Orders"))

	comments, err := GetTableComments(db)
	if err != nil {
		t.Fatalf("unexpected comments error: %v", err)
	}
	if comments["orders"] != "Orders" {
		t.Fatalf("unexpected comments: %#v", comments)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestPostgresPartitionAndComments(t *testing.T) {
	db, mock := openMockPostgres(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT p.relname AS partitioned_table,array_to_string(array_agg(c.relname),',')AS child_tables FROM pg_catalog.pg_class c JOIN pg_catalog.pg_inherits i ON c.oid=i.inhrelid JOIN pg_catalog.pg_class p ON p.oid=i.inhparent GROUP BY p.relname;")).
		WillReturnRows(sqlmock.NewRows([]string{"partitioned_table", "child_tables"}).AddRow("orders", "orders_202401,orders_202402"))

	partitions, err := GetPartitionTableToChildTables(db)
	if err != nil {
		t.Fatalf("unexpected partition error: %v", err)
	}
	if got := strings.Join(partitions["orders"], ","); got != "orders_202401,orders_202402" {
		t.Fatalf("unexpected partitions: %#v", partitions)
	}

	mock.ExpectQuery(regexp.QuoteMeta(buildPostgresTableCommentsQuery())).
		WillReturnRows(sqlmock.NewRows([]string{"table_name", "table_comment"}).AddRow("orders", "Orders"))

	comments, err := GetTableComments(db)
	if err != nil {
		t.Fatalf("unexpected comments error: %v", err)
	}
	if comments["orders"] != "Orders" {
		t.Fatalf("unexpected comments: %#v", comments)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
