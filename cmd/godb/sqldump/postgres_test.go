package sqldump

import "testing"

func TestPostgresDsnParse_KeywordValue(t *testing.T) {
	s := &SQLDump{
		dsn: "host=127.0.0.1 port=5432 user=postgres password='pa ss=word' dbname=test_db sslmode=disable",
	}
	got, err := s.postgresDsnParse()
	if err != nil {
		t.Fatal(err)
	}
	if got.Host != "127.0.0.1" || got.Port != 5432 || got.User != "postgres" || got.Password != "pa ss=word" || got.Dbname != "test_db" {
		t.Fatalf("unexpected parse result: %+v", got)
	}
}

func TestPostgresDsnParse_URL(t *testing.T) {
	s := &SQLDump{
		dsn: "postgres://pguser:p%40ss@localhost:5433/app_db?sslmode=disable",
	}
	got, err := s.postgresDsnParse()
	if err != nil {
		t.Fatal(err)
	}
	if got.Host != "localhost" || got.Port != 5433 || got.User != "pguser" || got.Password != "p@ss" || got.Dbname != "app_db" {
		t.Fatalf("unexpected parse result: %+v", got)
	}
}

func TestPostgresDsnParse_Invalid(t *testing.T) {
	s := &SQLDump{dsn: ":::bad dsn:::"}
	if _, err := s.postgresDsnParse(); err == nil {
		t.Fatal("expected parse error, got nil")
	}
}

func TestOutputDir_PreservesAbsoluteBasePath(t *testing.T) {
	got := outputDir("/tmp/sql", "app_db")
	if got != "/tmp/sql/app_db" {
		t.Fatalf("unexpected output dir: %s", got)
	}
}

func TestBuildPgDumpArgs_UsesDatabaseFlag(t *testing.T) {
	args := buildPgDumpArgs(&PostgresDsn{
		Host:   "127.0.0.1",
		Port:   5432,
		User:   "postgres",
		Dbname: "app_db",
	}, "users")
	want := []string{"-h", "127.0.0.1", "-p", "5432", "-U", "postgres", "-d", "app_db", "-s", "-t", "users"}
	if len(args) != len(want) {
		t.Fatalf("unexpected arg len: got=%d want=%d", len(args), len(want))
	}
	for i := range want {
		if args[i] != want[i] {
			t.Fatalf("unexpected args: got=%v want=%v", args, want)
		}
	}
}
