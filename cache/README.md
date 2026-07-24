## 缓存

### key管理

`cache/keymanage` 用于集中声明服务内缓存 key 前缀、TTL 和说明文档，生成的 key 会转义 `:` 和 `\`，避免不同字段组合产生碰撞。

### Redis 客户端

- `cache/gorediscache` 封装 go-redis 客户端、基础 Redis 信息读取和分布式锁。
- `cache/rueidiscache` 封装 rueidis 客户端、cache-aside 客户端和 rueidislock 分布式锁。

默认测试使用 `miniredis` 或 `redismock`，不需要启动 Redis。依赖真实 Redis 的测试使用 `integration` build tag，通过 `make test-integration` 显式运行；Rueidis cache-aside 场景要求 Redis 7 或更高版本。测试默认连接 `127.0.0.1:6379`，密码为 `123456`，本地可以通过下面的环境变量覆盖：

```bash
GODB_TEST_REDIS_ADDR=127.0.0.1:6379
GODB_TEST_REDIS_PASSWORD=123456
```

CI 的 integration job 会启动 Redis 7 服务并设置同样的密码；显式运行集成测试时，服务不可用会直接失败而不是跳过。

### 进程内缓存

`github.com/zeromicro/go-zero/core/collection/cache`

### 一致性缓存

rockscache: https://github.com/dtm-labs/rockscache
