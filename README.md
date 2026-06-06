# godb

[![Go Reference](https://pkg.go.dev/badge/github.com/fzf-labs/godb.svg)](https://pkg.go.dev/github.com/fzf-labs/godb)
[![Go Report Card](https://goreportcard.com/badge/github.com/fzf-labs/godb)](https://goreportcard.com/report/github.com/fzf-labs/godb)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8.svg)](go.mod)

语言: 简体中文 | [English](README_EN.md)

godb 是一个面向 Go 服务端项目的数据库工程工具箱。它并不试图替代 GORM，而是围绕 GORM 补齐真实业务项目里反复出现的工程化能力：数据库连接、代码生成、Repository 层、查询缓存、动态条件查询、批量更新、表结构导出和 proto 文件生成。

如果你的项目里经常需要根据表结构写重复的 DAO、Repo、缓存 Key、按索引查询、事务方法和分页条件查询，godb 可以把这些样板代码稳定地生成出来，让你把注意力放回业务规则。

## Highlights

- **GORM-first**: 基于 GORM 和 gorm/gen，保留原生 GORM 的模型、事务、查询和插件生态。
- **Schema-driven codegen**: 从 MySQL/PostgreSQL 表结构生成 `dao`、`model`、`repo` 三层代码。
- **Index-aware repository**: 根据主键、唯一索引、普通索引和联合索引最左匹配规则生成查询/更新/删除方法。
- **Cache-ready repository**: 生成带缓存读写和失效逻辑的 Repo 方法，缓存实现通过接口解耦。
- **Multiple cache backends**: 提供 go-redis、rueidis、RocksCache、Ristretto 等缓存相关封装。
- **Dynamic conditions**: 把结构化查询参数转换为 GORM clause，支持比较、IN、LIKE、NULL、RAW、排序和分页。
- **Operational CLI**: 一个 `godb` 命令覆盖 ORM 代码生成、SQL 结构导出、proto 文件生成。
- **Production-oriented helpers**: 连接池、健康检查、OpenTelemetry tracing、批量 CASE WHEN 更新、分表插件辅助。

## Packages

| Package | Purpose |
| --- | --- |
| `orm/gormx` | MySQL/PostgreSQL GORM 客户端、连接池配置、健康检查、状态读取、Tracing 插件 |
| `orm/gen` | 从数据库表结构生成 GORM model/dao/repo 和 proto 文件 |
| `orm/gen/repo` | Repo 模板生成器，按索引生成 CRUD、缓存、事务和条件查询方法 |
| `orm/dbcache` | 数据库查询缓存接口，以及 go-redis、rueidis、RocksCache 实现 |
| `orm/condition` | 结构化动态查询、排序和分页参数 |
| `orm/batch` | MySQL/PostgreSQL 批量更新 SQL 生成 |
| `orm/encoding` | JSON、Sonic、MsgPack、Zlib 编解码适配 |
| `orm/plugin` | GORM sharding 插件辅助构造 |
| `cache/*` | Redis、Rueidis、Ristretto、RocksCache、锁等缓存基础设施封装 |
| `cmd/godb` | CLI：`ormgen`、`sqldump`、`sqltopb` |

## Installation

安装库：

```bash
go get github.com/fzf-labs/godb
```

安装 CLI：

```bash
go install github.com/fzf-labs/godb/cmd/godb@latest
```

要求：

- Go 1.24+
- MySQL 或 PostgreSQL
- Redis 可选，仅在使用缓存实现或缓存 Repo 时需要
- `pg_dump` 可选，仅在导出 PostgreSQL 表结构时需要

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
		DataSourceName:  "host=localhost port=5432 user=postgres password=123456 dbname=app sslmode=disable TimeZone=Asia/Shanghai",
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

MySQL 只需要切换 driver 和 DSN：

```go
db, err := gormx.NewGormClient(&gormx.ClientConfig{
	Driver:         gormx.MySQL,
	DataSourceName: "user:password@tcp(localhost:3306)/app?charset=utf8mb4&parseTime=True&loc=Local",
	MaxIdleConn:    10,
	MaxOpenConn:    100,
})
```

### Generate ORM Code

```bash
godb ormgen \
  --db postgres \
  --dsn "host=localhost port=5432 user=postgres password=123456 dbname=app sslmode=disable TimeZone=Asia/Shanghai" \
  --outPutPath ./internal/data/gorm \
  --tables users,orders
```

生成目录会按数据库名拆分：

```text
internal/data/gorm/
  app_dao/
  app_model/
  app_repo/
```

生成的 Repo 会包含常见方法族：

- `CreateOne`、`CreateBatch`、`UpsertOne`
- `UpdateOne`、`UpdateOneWithZero`、`UpdateBatchBy...`
- `DeleteOneBy...`、`DeleteMultiBy...`
- `FindOneBy...`、`FindMultiBy...`
- `FindOneCacheBy...`、`FindMultiCacheBy...`
- `FindMultiByCondition`
- `...ByTx` 事务版本

### Use Generated Repository with Cache

下面示例使用仓库内生成的 example 包展示调用方式。你在业务项目中应替换为自己生成出的 `*_repo`、`*_model`、`*_dao` 包路径。

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
		"host=localhost port=5432 user=postgres password=123456 dbname=gorm_gen sslmode=disable TimeZone=Asia/Shanghai",
	)
	if err != nil {
		panic(err)
	}

	redisClient := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
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

`godb` 提供三个子命令：

```bash
godb ormgen   # 生成 GORM model/dao/repo
godb sqldump  # 导出数据库表结构 SQL
godb sqltopb  # 根据表结构生成 proto 文件
```

### `ormgen`

```bash
godb ormgen \
  -d postgres \
  -s "host=localhost port=5432 user=postgres password=123456 dbname=app sslmode=disable TimeZone=Asia/Shanghai" \
  -o ./internal/data/gorm \
  -t users,orders
```

| Flag | Default | Description |
| --- | --- | --- |
| `-d, --db` | | 数据库类型：`mysql`、`postgres` |
| `-s, --dsn` | | 数据库连接字符串 |
| `-o, --outPutPath` | `./internal/data/gorm` | 输出目录 |
| `-t, --tables` | | 指定表，逗号分隔；为空时读取全部表 |
| `-u, --optionUnderline` | `UL` | 模型字段以下划线开头时的替换前缀 |
| `-p, --optionPgDefaultString` | `true` | 处理 PostgreSQL `character varying` 默认值 tag |
| `-r, --optionRemoveDefault` | `true` | 移除 gorm tag 中的 default |
| `-g, --optionRemoveGormTypeTag` | `false` | 移除 gorm tag 中的 type |

### `sqldump`

```bash
godb sqldump \
  -d mysql \
  -s "user:password@tcp(localhost:3306)/app?charset=utf8mb4&parseTime=True&loc=Local" \
  -o ./doc/sql \
  -t users,orders \
  -f
```

| Flag | Default | Description |
| --- | --- | --- |
| `-d, --db` | | 数据库类型：`mysql`、`postgres` |
| `-s, --dsn` | | 数据库连接字符串 |
| `-o, --outPutPath` | `./doc/sql` | 输出目录 |
| `-t, --tables` | | 指定表，逗号分隔；为空时读取全部表 |
| `-f, --fileOverwrite` | `false` | 是否覆盖已存在文件 |

MySQL 使用 `SHOW CREATE TABLE`；PostgreSQL 使用本机 `pg_dump -s -t` 并清理部分环境相关语句。

### `sqltopb`

```bash
godb sqltopb \
  -d postgres \
  -s "host=localhost port=5432 user=postgres password=123456 dbname=app sslmode=disable TimeZone=Asia/Shanghai" \
  -o ./api/pb \
  -p app \
  -g "github.com/example/project/api/pb;pb" \
  -t users,orders
```

| Flag | Default | Description |
| --- | --- | --- |
| `-d, --db` | | 数据库类型：`mysql`、`postgres` |
| `-s, --dsn` | | 数据库连接字符串 |
| `-o, --outPutPath` | `./pb` | 输出目录 |
| `-p, --pbPackage` | `pb` | proto package |
| `-g, --pbGoPackage` | `github.com/fzf-labs/godb/orm/example/pb;pb` | `go_package` option |
| `-t, --tables` | | 指定表，逗号分隔；为空时读取全部表 |

## Library APIs

### Dynamic Query Conditions

`orm/condition` 用结构化参数描述查询条件，并转换为 GORM clause。它适合 API 层把筛选、排序、分页请求安全地传给 Repo 层。

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

支持的表达式：

- `=`、`!=`、`>`、`>=`、`<`、`<=`
- `IN`、`NOT IN`
- `LIKE`、`NOT LIKE`
- `IS NULL`、`IS NOT NULL`
- `RAW`
- `AND`、`OR`
- `ASC`、`DESC`

### Batch Update SQL

`orm/batch` 可以把结构体切片转换为 CASE WHEN 批量更新 SQL，当前每 200 条切分一次。

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

结构体字段需要带 `gorm:"column:xxx"` tag，并且必须包含 `id` 列。

### Cache Abstraction

生成的缓存 Repo 依赖 `dbcache.IDBCache` 接口：

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

内置实现：

- `orm/dbcache/goredisdbcache`
- `orm/dbcache/rueidisdbcache`
- `orm/dbcache/rocksdbcache`

缓存值的序列化由 `orm/encoding` 决定，内置 JSON、Sonic、MsgPack、Zlib。

## Code Generation Rules

Repo 生成器会读取数据库索引并按以下优先级生成方法：

1. 主键索引
2. 唯一索引
3. 普通索引
4. 联合索引的最左匹配派生索引

不同索引会生成不同语义的方法：

| Index Type | Generated Query Methods |
| --- | --- |
| 单字段唯一索引 | `FindOneBy<Field>`、`FindOneCacheBy<Field>`、`FindMultiBy<FieldPlural>` |
| 多字段唯一索引 | `FindOneBy<Fields>`、`FindOneCacheBy<Fields>` |
| 单字段普通索引 | `FindMultiBy<Field>`、`FindMultiCacheBy<Field>`、批量入参版本 |
| 多字段普通索引 | `FindMultiBy<Fields>`、`FindMultiCacheBy<Fields>` |

写操作同时生成普通版本、缓存版本和事务版本，例如：

- `CreateOne` / `CreateOneCache` / `CreateOneByTx` / `CreateOneCacheByTx`
- `UpdateOne` / `UpdateOneCache` / `UpdateOneByTx` / `UpdateOneCacheByTx`
- `UpsertOne` / `UpsertOneCache` / `UpsertOneByTx` / `UpsertOneCacheByTx`
- `DeleteOneBy<Field>` / `DeleteOneCacheBy<Field>` / `DeleteOneBy<Field>Tx`

更多模板规则可以参考 [orm/gen/repo/README.md](orm/gen/repo/README.md)。

## Project Layout

```text
.
├── cache/                 # Redis、Rueidis、Ristretto、RocksCache 等基础缓存封装
├── cmd/godb/              # godb CLI
├── orm/batch/             # 批量更新 SQL 生成
├── orm/condition/         # 动态查询条件
├── orm/dbcache/           # Repo 查询缓存接口与实现
├── orm/encoding/          # 缓存序列化策略
├── orm/example/           # 生成代码示例
├── orm/gen/               # ORM/proto 代码生成器
├── orm/gormx/             # GORM 客户端和数据库工具
├── orm/plugin/            # GORM 插件辅助
└── orm/utils/             # 文件、字符串、模板工具
```

## Development

常用命令：

```bash
make fmt
make vet
make test
make cover
make ci
```

手动安装并检查 CLI：

```bash
go install ./cmd/godb
godb --help
```

本仓库的 PostgreSQL/Redis 示例测试在 CI 中会自动准备服务和种子数据；本地可通过 `GODB_TEST_POSTGRES_DSN`、`GODB_TEST_REDIS_ADDR` 和 `GODB_TEST_REDIS_PASSWORD` 覆盖默认地址与凭据。如果本机没有对应服务，相关测试会跳过或使用 mock；在 CI 环境中，服务不可用会让测试失败，避免误把集成测试跳过当成通过。

发布流程：

```bash
make release-snapshot
make release-tag
```

`make release-tag` 会创建并推送下一个 patch tag；所有 `v*` tag 会触发 GitHub Actions 的 release workflow，并由 GoReleaser 产出跨平台二进制、校验和和 GitHub Release。

## Design Notes

godb 的设计目标是把“数据库表结构”转换成“项目里真正会写的代码”：

- 表结构决定 model 和 DAO。
- 索引决定 Repository 查询方法。
- 缓存接口决定缓存实现可替换。
- 编解码接口决定缓存 payload 可替换。
- CLI 让生成、导出、proto 构建可以放进 Makefile、CI 或初始化脚本。

这使得数据库访问层既保持可生成、可重复，又不会把业务代码绑死在某个缓存客户端或序列化格式上。

## Related Documentation

- [ORM 代码生成](orm/gen/README.md)
- [Repo 生成规则](orm/gen/repo/README.md)
- [动态条件查询](orm/condition/README.md)
- [批量更新](orm/batch/README.md)
- [CLI 文档](cmd/godb/README.md)
- [缓存说明](cache/README.md)

## License

godb is released under the [MIT License](LICENSE).
