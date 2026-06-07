package testenv

import (
	"fmt"
	"testing"
)

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

func TestRedisAddrUsesOverrideWhenSet(t *testing.T) {
	t.Setenv("GODB_TEST_REDIS_ADDR", "redis.example:6380")

	got := RedisAddr()
	if got != "redis.example:6380" {
		t.Fatalf("unexpected redis addr override: %q", got)
	}
}

func TestRedisPasswordUsesDefaultWhenUnset(t *testing.T) {
	t.Setenv("GODB_TEST_REDIS_PASSWORD", "")

	got := RedisPassword()
	if got != "123456" {
		t.Fatalf("unexpected redis password: %q", got)
	}
}

func TestRedisPasswordUsesOverrideWhenSet(t *testing.T) {
	t.Setenv("GODB_TEST_REDIS_PASSWORD", "secret")

	got := RedisPassword()
	if got != "secret" {
		t.Fatalf("unexpected redis password override: %q", got)
	}
}

func TestSkipIfUnavailableSkipsOutsideCI(t *testing.T) {
	tb := &fakeUnavailableReporter{}

	skipIfUnavailable(tb, false, "service unavailable: %v", "dial")

	if tb.helperCalls != 1 {
		t.Fatalf("expected Helper to be called once, got %d", tb.helperCalls)
	}
	if tb.skipMessage != "service unavailable: dial" {
		t.Fatalf("unexpected skip message: %q", tb.skipMessage)
	}
	if tb.fatalMessage != "" {
		t.Fatalf("unexpected fatal message: %q", tb.fatalMessage)
	}
}

func TestSkipIfUnavailableFailsInCI(t *testing.T) {
	tb := &fakeUnavailableReporter{}

	skipIfUnavailable(tb, true, "service unavailable: %v", "dial")

	if tb.helperCalls != 1 {
		t.Fatalf("expected Helper to be called once, got %d", tb.helperCalls)
	}
	if tb.fatalMessage != "service unavailable: dial" {
		t.Fatalf("unexpected fatal message: %q", tb.fatalMessage)
	}
	if tb.skipMessage != "" {
		t.Fatalf("unexpected skip message: %q", tb.skipMessage)
	}
}

func TestSkipIfUnavailableUsesEnvironmentDecision(t *testing.T) {
	t.Setenv("CI", "")
	t.Setenv("GITHUB_ACTIONS", "")
	called := false

	t.Run("local skip", func(t *testing.T) {
		called = true
		SkipIfUnavailable(t, "local service unavailable")
		t.Fatal("SkipIfUnavailable returned without skipping")
	})

	if !called {
		t.Fatal("expected local skip subtest to run")
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

type fakeUnavailableReporter struct {
	helperCalls  int
	fatalMessage string
	skipMessage  string
}

// Fatalf records the fatal message that would be reported.
func (f *fakeUnavailableReporter) Fatalf(format string, args ...any) {
	f.fatalMessage = fmt.Sprintf(format, args...)
}

// Helper records helper registration calls.
func (f *fakeUnavailableReporter) Helper() {
	f.helperCalls++
}

// Skipf records the skip message that would be reported.
func (f *fakeUnavailableReporter) Skipf(format string, args ...any) {
	f.skipMessage = fmt.Sprintf(format, args...)
}
