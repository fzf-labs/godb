package testenv

import (
	"fmt"
	"os"
	"testing"
)

const (
	defaultRedisAddr     = "127.0.0.1:6379"
	defaultRedisPassword = "123456"
)

// PostgresDSN returns the PostgreSQL DSN used by tests, honoring GODB_TEST_POSTGRES_DSN when set.
func PostgresDSN(dbname string) string {
	if dsn := os.Getenv("GODB_TEST_POSTGRES_DSN"); dsn != "" {
		return dsn
	}
	return fmt.Sprintf("host=127.0.0.1 port=5432 user=postgres password=123456 dbname=%s sslmode=disable TimeZone=Asia/Shanghai", dbname)
}

// RedisAddr returns the Redis address used by tests, honoring GODB_TEST_REDIS_ADDR when set.
func RedisAddr() string {
	if addr := os.Getenv("GODB_TEST_REDIS_ADDR"); addr != "" {
		return addr
	}
	return defaultRedisAddr
}

// RedisPassword returns the Redis password used by tests, honoring GODB_TEST_REDIS_PASSWORD when set.
func RedisPassword() string {
	if password := os.Getenv("GODB_TEST_REDIS_PASSWORD"); password != "" {
		return password
	}
	return defaultRedisPassword
}

// SkipIfUnavailable skips locally for optional services, but fails in CI where services must be provisioned.
func SkipIfUnavailable(t testing.TB, format string, args ...any) {
	t.Helper()
	skipIfUnavailable(t, ciEnabled(), format, args...)
}

type unavailableReporter interface {
	Helper()
	Fatalf(format string, args ...any)
	Skipf(format string, args ...any)
}

func skipIfUnavailable(t unavailableReporter, fail bool, format string, args ...any) {
	t.Helper()
	if fail {
		t.Fatalf(format, args...)
		return
	}
	t.Skipf(format, args...)
}

func ciEnabled() bool {
	return os.Getenv("CI") != "" || os.Getenv("GITHUB_ACTIONS") != ""
}
