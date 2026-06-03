package sqldump

import "testing"

func TestBuildMySQLShowCreateTableSQL_QuotesIdentifiers(t *testing.T) {
	got := buildMySQLShowCreateTableSQL("app`db", "user`log")
	want := "SHOW CREATE TABLE `app``db`.`user``log`"
	if got != want {
		t.Fatalf("unexpected sql: got=%q want=%q", got, want)
	}
}
