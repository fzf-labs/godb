# Release Hygiene Foundation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make the repository's baseline developer loop and CI reliable by adding service-backed test defaults, first-class make targets, and coverage reporting that excludes generated example packages.

**Architecture:** Keep the shared test inputs in a tiny internal helper package so local runs and CI use the same defaults. Put the CI logic in GitHub Actions and keep database bootstrap in a small script so the workflow stays readable. Coverage should remain a `make` concern, not a test concern, so package filtering happens in one place.

**Tech Stack:** Go 1.24, GitHub Actions, `make`, PostgreSQL, Redis.

---

### Task 1: Shared test environment helpers

**Files:**
- Create: `internal/testenv/testenv.go`
- Modify: `cache/gorediscache/gorediscache_test.go`
- Modify: `cache/rueidiscache/rueidis_test.go`
- Modify: `cache/rueidiscache/rueidislock_test.go`
- Modify: `cache/rockscache/rockscache_test.go`
- Modify: `orm/dbcache/goredisdbcache/goredisdbcache_test.go`
- Modify: `orm/dbcache/rueidisdbcache/rueidisdbcache_test.go`
- Modify: `orm/dbcache/rocksdbcache/rocksdbcache_test.go`
- Modify: `orm/gormx/gormx_test.go`
- Modify: `orm/plugin/shard_test.go`
- Modify: `orm/gen/repo/repo_test.go`
- Modify: `orm/gen/gen_db_test.go`
- Modify: `orm/gen/gen_pb_test.go`
- Modify: `orm/gen/proto/proto_test.go`
- Modify: `orm/example/gorm/example_test.go`

- [x] **Step 1: Add the shared defaults**

```go
package testenv

import (
	"fmt"
	"os"
)

func PostgresDSN(dbname string) string {
	if dsn := os.Getenv("GODB_TEST_POSTGRES_DSN"); dsn != "" {
		return dsn
	}
	return fmt.Sprintf(
		"host=127.0.0.1 port=5432 user=postgres password=123456 dbname=%s sslmode=disable TimeZone=Asia/Shanghai",
		dbname,
	)
}

func RedisAddr() string {
	if addr := os.Getenv("GODB_TEST_REDIS_ADDR"); addr != "" {
		return addr
	}
	return "127.0.0.1:6379"
}

func RedisPassword() string {
	if password := os.Getenv("GODB_TEST_REDIS_PASSWORD"); password != "" {
		return password
	}
	return "123456"
}
```

- [x] **Step 2: Replace hard-coded local service addresses**

Use `testenv.PostgresDSN("gorm_gen")`, `testenv.PostgresDSN("user")`, `testenv.PostgresDSN("fkratos_sys")`, `testenv.RedisAddr()`, and `testenv.RedisPassword()` in the files listed above.

- [x] **Step 3: Run the affected packages**

Run: `go test ./cache/... ./orm/gormx ./orm/plugin ./orm/dbcache/... ./orm/gen/... ./orm/example/gorm`
Expected: live-service tests connect through the shared defaults instead of embedding `0.0.0.0`.

### Task 2: CI workflow and PostgreSQL bootstrap

**Files:**
- Create: `.github/workflows/ci.yml`
- Create: `scripts/ci/bootstrap-postgres.sql`
- Create: `scripts/ci/bootstrap-postgres.sh`

- [x] **Step 1: Add the workflow**

Run these checks in CI:

```yaml
name: ci
on:
  push:
  pull_request:

jobs:
  test:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:15
        env:
          POSTGRES_PASSWORD: 123456
        ports:
          - 5432:5432
        options: >-
          --health-cmd="pg_isready -U postgres"
          --health-interval=5s
          --health-timeout=5s
          --health-retries=20
      redis:
        image: redis:7
        ports:
          - 6379:6379
        options: >-
          --health-cmd="redis-cli ping"
          --health-interval=5s
          --health-timeout=5s
          --health-retries=20
```

- [x] **Step 2: Bootstrap the databases**

Create `gorm_gen`, `user`, and `fkratos_sys`, enable `pgcrypto`, apply the schema files under `orm/example/sql/fdatabase`, seed the two `user_demo` rows used by the repository tests, and create a minimal `sys_admin_202301` table for the sharding plugin tests.

- [x] **Step 3: Run the checks**

Run:

```bash
go test ./...
go vet ./...
test -z "$(gofmt -l .)"
```

Expected: CI fails on formatting or static issues, and service-backed tests run instead of skipping.

### Task 3: Makefile coverage and developer entrypoints

**Files:**
- Modify: `Makefile`

- [x] **Step 1: Add package filtering for coverage**

```make
TEST_PKGS := ./...
COVER_PKGS := $(shell go list ./... | grep -v '/orm/example/gorm/postgres/gorm_gen_')

.PHONY: test
test:
	@go test $(TEST_PKGS)

.PHONY: cover
cover:
	@go test $(COVER_PKGS) -coverprofile=/tmp/godb.cover
	@go tool cover -func=/tmp/godb.cover | tail -n 1

.PHONY: ci
ci: fmt vet test cover
```

- [x] **Step 2: Keep the existing helpers**

Leave `fmt`, `vet`, and `help` in place so the new targets extend the current workflow instead of replacing it.

- [x] **Step 3: Run the entrypoints**

Run: `make test && make cover`
Expected: `make cover` reports coverage without counting the generated example repo/dao/model packages.

### Task 4: README sync for the new workflow

**Files:**
- Modify: `README.md`
- Modify: `README_EN.md`

- [x] **Step 1: Update the development section**

Document `make test`, `make cover`, and `make ci` as the preferred local checks.

- [x] **Step 2: Clarify external-service behavior**

State that PostgreSQL and Redis-backed tests now have stable CI defaults and use local overrides when those environment variables are set.

- [x] **Step 3: Verify the prose**

Run: `rg -n "make test|make cover|make ci|0.0.0.0" README.md README_EN.md`
Expected: the new commands are mentioned and the old hard-coded service address guidance is gone.

---

### Out of scope for this plan

- `orm/gen` and `orm/gen/repo` golden tests
- release/tag automation
- splitting `orm/gen/repo/repo.go`
- additional benchmark coverage
- broader docs and API cleanup
