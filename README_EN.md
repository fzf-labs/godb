# godb

[![Go Reference](https://pkg.go.dev/badge/github.com/fzf-labs/godb.svg)](https://pkg.go.dev/github.com/fzf-labs/godb)
[![Go Report Card](https://goreportcard.com/badge/github.com/fzf-labs/godb)](https://goreportcard.com/report/github.com/fzf-labs/godb)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8.svg)](go.mod)

Language: English | [Simplified Chinese](README.md)

godb is a database engineering toolkit for Go backend services. It does not try to replace GORM. Instead, it builds around GORM and fills in the pieces that appear again and again in production projects: database clients, code generation, repository layers, query caching, dynamic conditions, batch updates, schema dumping, logical backups, and proto generation.

If your project repeatedly needs DAO code, repository methods, cache keys, index-based queries, transaction helpers, and paginated condition queries from database schemas, godb turns that boilerplate into generated, repeatable code so you can focus on business logic.

## Highlights

- **GORM-first**: Built on GORM and gorm/gen while keeping the native GORM model, transaction, query, and plugin ecosystem.
- **Schema-driven code generation**: Generate `dao`, `model`, and `repo` layers from MySQL/PostgreSQL schemas.
- **Index-aware repositories**: Generate query/update/delete methods from primary keys, unique indexes, normal indexes, and left-prefix compound indexes.
- **Cache-ready repositories**: Generate repository methods with cache reads, writes, and invalidation logic behind a replaceable cache interface.
- **Multiple cache backends**: Provides wrappers for go-redis, rueidis, RocksCache, Ristretto, and related cache infrastructure.
- **Dynamic conditions**: Convert structured API query parameters into GORM clauses with comparisons, IN, LIKE, NULL checks, RAW expressions, ordering, and pagination.
- **Operational CLI**: One `godb` command for ORM code generation, SQL schema dumping, logical backups, and proto generation.
- **Production-oriented helpers**: Connection pools, health checks, OpenTelemetry tracing, CASE WHEN batch updates, and sharding plugin helpers.

## Packages

| Package | Purpose |
| --- | --- |
| `orm/gormx` | MySQL/PostgreSQL GORM clients, pool settings, health checks, DB stats, and tracing |
| `orm/gen` | Generate GORM model/dao/repo code and proto files from database schemas |
| `orm/gen/repo` | Repository template generator for CRUD, cache, transaction, and condition-query methods |
| `orm/dbcache` | Database query cache interface plus go-redis, rueidis, and RocksCache implementations |
| `orm/condition` | Structured dynamic query, ordering, and pagination parameters |
| `orm/batch` | MySQL/PostgreSQL batch update SQL generation |
| `orm/encoding` | JSON, Sonic, MsgPack, and Zlib codec adapters |
| `orm/plugin` | GORM sharding plugin helpers |
| `cache/*` | Redis, Rueidis, Ristretto, RocksCache, and lock helpers |
| `cmd/godb` | CLI commands: `ormgen`, `sqldump`, `sqlbackup`, and `sqltopb` |

## Installation

Install the library:

```bash
go get github.com/fzf-labs/godb
```

Install the CLI:

```bash
go install github.com/fzf-labs/godb/cmd/godb@latest
```

Requirements:

- Go 1.24+
- MySQL or PostgreSQL
- Redis is optional and only required when using cache implementations or generated cache repositories
- `pg_dump` is required to dump PostgreSQL schemas or create PostgreSQL `sqlbackup` artifacts
- `mysqldump` is required to create MySQL `sqlbackup` artifacts

## Quick Start

### Create a GORM Client

```go
package main

import (
	"time"

	"github.com/fzf-labs/godb/orm/gormx"
)

func main() {
	db, err := gormx.NewGormClient(&gormx.ClientConfig{
		Driver:          gormx.Postgres,
		DataSourceName:  "host=127.0.0.1 port=5432 user=postgres password=123456 dbname=app sslmode=disable TimeZone=Asia/Shanghai",
		MaxIdleConn:     10,
		MaxOpenConn:     100,
		ConnMaxIdleTime: 10 * time.Minute,
		ConnMaxLifeTime: time.Hour,
		ShowLog:         true,
		Tracing:         false,
	})
	if err != nil {
		panic(err)
	}

	_ = db
}
```

For MySQL, switch the driver and DSN:

```go
db, err := gormx.NewGormClient(&gormx.ClientConfig{
	Driver:         gormx.MySQL,
	DataSourceName: "user:password@tcp(127.0.0.1:3306)/app?charset=utf8mb4&parseTime=True&loc=Local",
	MaxIdleConn:    10,
	MaxOpenConn:    100,
})
```

### Generate ORM Code

```bash
godb ormgen \
  --db postgres \
  --dsn "host=127.0.0.1 port=5432 user=postgres password=123456 dbname=app sslmode=disable TimeZone=Asia/Shanghai" \
  --outPutPath ./internal/data/gorm \
  --tables users,orders
```

Generated code is grouped by database name:

```text
internal/data/gorm/
  app_dao/
  app_model/
  app_repo/
```

Generated repositories include common method families:

- `CreateOne`, `CreateBatch`, `UpsertOne`
- `UpdateOne`, `UpdateOneWithZero`, `UpdateBatchBy...`
- `DeleteOneBy...`, `DeleteMultiBy...`
- `FindOneBy...`, `FindMultiBy...`
- `FindOneCacheBy...`, `FindMultiCacheBy...`
- `FindMultiByCondition`
- `...ByTx` transaction variants

### Use a Generated Repository with Cache

The example below uses generated packages from this repository. In your own project, replace the package paths with your generated `*_repo`, `*_model`, and `*_dao` packages.

```go
package main

import (
	"context"

	"github.com/fzf-labs/godb/orm/dbcache/goredisdbcache"
	"github.com/fzf-labs/godb/orm/encoding"
	"github.com/fzf-labs/godb/orm/example/gorm/postgres/gorm_gen_repo"
	"github.com/fzf-labs/godb/orm/gen/config"
	"github.com/fzf-labs/godb/orm/gormx"
	"github.com/redis/go-redis/v9"
)

func main() {
	db, err := gormx.NewSimpleGormClient(
		gormx.Postgres,
		"host=127.0.0.1 port=5432 user=postgres password=123456 dbname=gorm_gen sslmode=disable TimeZone=Asia/Shanghai",
	)
	if err != nil {
		panic(err)
	}

	redisClient := redis.NewClient(&redis.Options{Addr: "127.0.0.1:6379"})
	dbCache := goredisdbcache.NewGoRedisDBCache(redisClient)

	repoCfg := config.NewRepoConfig(db, dbCache, encoding.NewMsgPack())
	userRepo := gorm_gen_repo.NewUserDemoRepo(repoCfg)

	user, err := userRepo.FindOneCacheByID(context.Background(), "user-id")
	if err != nil {
		panic(err)
	}

	_ = user
}
```

## CLI

`godb` provides four subcommands:

```bash
godb ormgen   # Generate GORM model/dao/repo code
godb sqldump  # Export database table schema SQL
godb sqlbackup # Export database schema and data
godb sqltopb  # Generate proto files from database tables
```

### `ormgen`

```bash
godb ormgen \
  -d postgres \
  -s "host=127.0.0.1 port=5432 user=postgres password=123456 dbname=app sslmode=disable TimeZone=Asia/Shanghai" \
  -o ./internal/data/gorm \
  -t users,orders
```

| Flag | Default | Description |
| --- | --- | --- |
| `-d, --db` | | Database type: `mysql` or `postgres` |
| `-s, --dsn` | | Database connection string |
| `-o, --outPutPath` | `./internal/data/gorm` | Output directory |
| `-t, --tables` | | Target tables, comma-separated. Empty means all tables |
| `-u, --optionUnderline` | `UL` | Replacement prefix for generated fields that start with `_` |
| `-p, --optionPgDefaultString` | `true` | Normalize PostgreSQL `character varying` default tags |
| `-r, --optionRemoveDefault` | `true` | Remove `default` from GORM tags |
| `-g, --optionRemoveGormTypeTag` | `false` | Remove `type` from GORM tags |

### `sqldump`

```bash
godb sqldump \
  -d mysql \
  -s "user:password@tcp(127.0.0.1:3306)/app?charset=utf8mb4&parseTime=True&loc=Local" \
  -o ./doc/sql \
  -t users,orders \
  -f
```

| Flag | Default | Description |
| --- | --- | --- |
| `-d, --db` | | Database type: `mysql` or `postgres` |
| `-s, --dsn` | | Database connection string |
| `-o, --outPutPath` | `./doc/sql` | Output directory |
| `-t, --tables` | | Target tables, comma-separated. Empty means all tables |
| `-f, --fileOverwrite` | `false` | Overwrite existing files |

MySQL uses `SHOW CREATE TABLE`. PostgreSQL uses local `pg_dump -s -t` and removes some environment-specific statements.

### `sqlbackup`

`sqlbackup` exports the schema and data of the single database selected by the DSN. The output must be a target file. Existing artifacts are protected unless `--force` is supplied, and the command writes a temporary file in the same directory before publishing the final backup.

```bash
godb sqlbackup \
  -d postgres \
  -s "host=127.0.0.1 port=5432 user=postgres password=123456 dbname=app sslmode=disable" \
  -o ./backup/app.dump

godb sqlbackup \
  -d mysql \
  -s "backup:password@tcp(127.0.0.1:3306)/app?parseTime=true" \
  -o ./backup/app.sql
```

| Flag | Default | Description |
| --- | --- | --- |
| `-d, --db` | | Database type: `mysql` or `postgres` |
| `-s, --dsn` | | Database connection string |
| `-o, --output` | | Backup target file; `.dump` is recommended for PostgreSQL and `.sql` for MySQL |
| `-f, --force` | `false` | Replace an existing target after the backup completes successfully |

PostgreSQL uses a `pg_dump` custom archive. Create a target database first, then restore it with local `pg_restore`:

```bash
createdb -h 127.0.0.1 -p 5432 -U postgres app_restore
PGPASSWORD=123456 pg_restore \
  --no-owner --no-privileges --no-tablespaces \
  -h 127.0.0.1 -p 5432 -U postgres -d app_restore \
  ./backup/app.dump
```

MySQL uses `mysqldump` to produce SQL. Restore into a newly created database with the local `mysql` client:

```bash
mysql -h 127.0.0.1 -P 3306 -u backup -p -e "CREATE DATABASE app_restore;"
mysql -h 127.0.0.1 -P 3306 -u backup -p app_restore < ./backup/app.sql
```

Backups do not include PostgreSQL roles/tablespaces, MySQL user accounts, or other instance-level objects. MySQL uses a single-transaction snapshot by default, which is consistent only for transactional tables such as InnoDB. The first release supports standard single-endpoint TCP/Unix socket connections; multi-host, service, and named Go-driver TLS configurations return an error.

### `sqltopb`

```bash
godb sqltopb \
  -d postgres \
  -s "host=127.0.0.1 port=5432 user=postgres password=123456 dbname=app sslmode=disable TimeZone=Asia/Shanghai" \
  -o ./api/pb \
  -p app \
  -g "github.com/example/project/api/pb;pb" \
  -t users,orders
```

| Flag | Default | Description |
| --- | --- | --- |
| `-d, --db` | | Database type: `mysql` or `postgres` |
| `-s, --dsn` | | Database connection string |
| `-o, --outPutPath` | `./pb` | Output directory |
| `-p, --pbPackage` | `pb` | Proto package |
| `-g, --pbGoPackage` | `github.com/fzf-labs/godb/orm/example/pb;pb` | `go_package` option |
| `-t, --tables` | | Target tables, comma-separated. Empty means all tables |

## Library APIs

### Dynamic Query Conditions

`orm/condition` describes query filters as structured parameters and converts them into GORM clauses. It is useful when API-layer filters, sorting, and pagination need to be safely passed to repository methods.

```go
req := &condition.Req{
	Page:     1,
	PageSize: 20,
	Query: []*condition.QueryParam{
		{Field: "status", Value: 1, Exp: condition.EQ},
		{Field: "username", Value: "%admin%", Exp: condition.LIKE},
		{Field: "tenant_id", Value: []int64{1, 2}, Exp: condition.IN},
	},
	Order: []*condition.OrderParam{
		{Field: "created_at", Order: condition.DESC},
	},
}
```

Supported expressions:

- `=`, `!=`, `>`, `>=`, `<`, `<=`
- `IN`, `NOT IN`
- `LIKE`, `NOT LIKE`
- `IS NULL`, `IS NOT NULL`
- `RAW`
- `AND`, `OR`
- `ASC`, `DESC`

### Batch Update SQL

`orm/batch` converts a slice of struct pointers into CASE WHEN batch update SQL. The current batch size is 200 rows per statement.

```go
sqls, err := batch.PostgresBatchUpdateToSQLArray("public.users", users)
if err != nil {
	return err
}

for _, sql := range sqls {
	if err := db.Exec(sql).Error; err != nil {
		return err
	}
}
```

Struct fields must have `gorm:"column:xxx"` tags, and the struct must include an `id` column.

### Cache Abstraction

Generated cache repositories depend on the `dbcache.IDBCache` interface:

```go
type IDBCache interface {
	Key(fields ...any) string
	TTL() time.Duration
	Fetch(ctx context.Context, key string, fn func() (string, error), expire time.Duration) (string, error)
	FetchBatch(ctx context.Context, keys []string, fn func(miss []string) (map[string]string, error), expire time.Duration) (map[string]string, error)
	FetchHash(ctx context.Context, key string, field string, fn func() (string, error), expire time.Duration) (string, error)
	Del(ctx context.Context, key string) error
	DelBatch(ctx context.Context, keys []string) error
	DelHash(ctx context.Context, key string, field string) error
}
```

Built-in implementations:

- `orm/dbcache/goredisdbcache`
- `orm/dbcache/rueidisdbcache`
- `orm/dbcache/rocksdbcache`

Cache value serialization is controlled by `orm/encoding`, which includes JSON, Sonic, MsgPack, and Zlib implementations.

## Code Generation Rules

The repository generator reads database indexes and generates methods in the following priority:

1. Primary keys
2. Unique indexes
3. Normal indexes
4. Left-prefix derived indexes from compound indexes

Different index shapes produce different query methods:

| Index Type | Generated Query Methods |
| --- | --- |
| Single-column unique index | `FindOneBy<Field>`, `FindOneCacheBy<Field>`, `FindMultiBy<FieldPlural>` |
| Multi-column unique index | `FindOneBy<Fields>`, `FindOneCacheBy<Fields>` |
| Single-column normal index | `FindMultiBy<Field>`, `FindMultiCacheBy<Field>`, batch-input variants |
| Multi-column normal index | `FindMultiBy<Fields>`, `FindMultiCacheBy<Fields>` |

Write methods are generated in normal, cache-aware, and transaction-aware variants, for example:

- `CreateOne` / `CreateOneCache` / `CreateOneByTx` / `CreateOneCacheByTx`
- `UpdateOne` / `UpdateOneCache` / `UpdateOneByTx` / `UpdateOneCacheByTx`
- `UpsertOne` / `UpsertOneCache` / `UpsertOneByTx` / `UpsertOneCacheByTx`
- `DeleteOneBy<Field>` / `DeleteOneCacheBy<Field>` / `DeleteOneBy<Field>Tx`

See [orm/gen/repo/README.md](orm/gen/repo/README.md) for more template rules.

## Project Layout

```text
.
├── cache/                 # Redis, Rueidis, Ristretto, RocksCache, and cache helpers
├── cmd/godb/              # godb CLI
├── orm/batch/             # Batch update SQL generation
├── orm/condition/         # Dynamic query conditions
├── orm/dbcache/           # Repository query cache interface and implementations
├── orm/encoding/          # Cache serialization strategies
├── orm/example/           # Generated code examples
├── orm/gen/               # ORM/proto code generators
├── orm/gormx/             # GORM clients and database utilities
├── orm/plugin/            # GORM plugin helpers
└── orm/utils/             # File, string, and template utilities
```

## Development

Common commands:

```bash
make fmt
make lint
make comments
make vet
make test
make test-integration
make bootstrap-postgres
make cover
make ci-unit
make ci-integration
make ci
```

Install and inspect the CLI locally:

```bash
go install ./cmd/godb
godb --help
```

`make test` and `make cover` run only the default tests that do not require external services. They use `sqlmock`, SQLite, `miniredis`, or `redismock`, so PostgreSQL and Redis do not need to be running locally. Real-service tests use the `integration` build tag and run explicitly through `make test-integration`; unavailable services fail those tests instead of skipping them.

Local PostgreSQL configuration uses the standard `PGHOST`, `PGPORT`, `PGUSER`, and `PGPASSWORD` environment variables. Redis uses `GODB_TEST_REDIS_ADDR` and `GODB_TEST_REDIS_PASSWORD`. If PostgreSQL uses the default user and password, run `make bootstrap-postgres` to recreate the `gorm_gen`, `user`, and `fkratos_sys` test databases and import the example schema. This command drops and recreates those test databases, so do not point it at production. CI runs default checks and real-service tests separately and provisions PostgreSQL, Redis, and seed data automatically.

Release flow:

```bash
make release-snapshot
make release-tag
```

`make release-tag` creates and pushes the next patch tag; every `v*` tag triggers the GitHub Actions release workflow, which uses GoReleaser to publish cross-platform binaries, checksums, and the GitHub Release.

## Design Notes

godb is designed to turn database schemas into the code you actually write in backend projects:

- Tables define models and DAOs.
- Indexes define repository query methods.
- Cache interfaces make cache implementations replaceable.
- Codec interfaces make cache payload formats replaceable.
- CLI commands make generation, schema dumping, and proto generation easy to put into Makefiles, CI jobs, or project bootstrap scripts.

The result is a database access layer that is generated and repeatable, while still keeping your business code independent from a specific cache client or serialization format.

## Related Documentation

- [ORM code generation](orm/gen/README.md)
- [Repository generation rules](orm/gen/repo/README.md)
- [Dynamic query conditions](orm/condition/README.md)
- [Batch updates](orm/batch/README.md)
- [CLI documentation](cmd/godb/README.md)
- [Cache notes](cache/README.md)

## License

godb is released under the [MIT License](LICENSE).
