package testenv

import "testing"

func TestPostgresDSNUsesDefaultWhenUnset(t *testing.T) {
	t.Setenv("GODB_TEST_POSTGRES_DSN", "")

	got := PostgresDSN("gorm_gen")
	want := "host=127.0.0.1 port=5432 user=postgres password=123456 dbname=gorm_gen sslmode=disable TimeZone=Asia/Shanghai"
	if got != want {
		t.Fatalf("unexpected postgres dsn: got %q want %q", got, want)
	}
}

func TestPostgresDSNUsesOverrideWhenSet(t *testing.T) {
	t.Setenv("GODB_TEST_POSTGRES_DSN", "postgres://example")

	got := PostgresDSN("gorm_gen")
	if got != "postgres://example" {
		t.Fatalf("unexpected postgres dsn override: %q", got)
	}
}

func TestRedisAddrUsesDefaultWhenUnset(t *testing.T) {
	t.Setenv("GODB_TEST_REDIS_ADDR", "")

	got := RedisAddr()
	if got != "127.0.0.1:6379" {
		t.Fatalf("unexpected redis addr: %q", got)
	}
}

func TestRedisPasswordUsesDefaultWhenUnset(t *testing.T) {
	t.Setenv("GODB_TEST_REDIS_PASSWORD", "")

	got := RedisPassword()
	if got != "123456" {
		t.Fatalf("unexpected redis password: %q", got)
	}
}

func TestCIEnabled(t *testing.T) {
	t.Setenv("CI", "")
	t.Setenv("GITHUB_ACTIONS", "")
	if ciEnabled() {
		t.Fatal("expected CI to be disabled")
	}

	t.Setenv("CI", "true")
	if !ciEnabled() {
		t.Fatal("expected CI env to enable CI mode")
	}

	t.Setenv("CI", "")
	t.Setenv("GITHUB_ACTIONS", "true")
	if !ciEnabled() {
		t.Fatal("expected GITHUB_ACTIONS env to enable CI mode")
	}
}
