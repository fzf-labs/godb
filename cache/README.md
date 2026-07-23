## 缓存

### key管理

`cache/keymanage` 用于集中声明服务内缓存 key 前缀、TTL 和说明文档，生成的 key 会转义 `:` 和 `\`，避免不同字段组合产生碰撞。

### Redis 客户端

- `cache/gorediscache` 封装 go-redis 客户端、基础 Redis 信息读取和分布式锁。
- `cache/rueidiscache` 封装 rueidis 客户端、cache-aside 客户端和 rueidislock 分布式锁。

依赖真实 Redis 的测试默认连接 `127.0.0.1:6379`，密码为 `123456`。本地可以通过下面的环境变量覆盖：

```bash
GODB_TEST_REDIS_ADDR=127.0.0.1:6379
GODB_TEST_REDIS_PASSWORD=123456
```

CI 会启动 Redis 7 服务并设置同样的密码，因此 Redis 相关测试在 CI 中有稳定环境；如果 CI 中服务不可用，测试会失败而不是跳过。

### 进程内缓存

`github.com/zeromicro/go-zero/core/collection/cache`

### 一致性缓存

rockscache: https://github.com/dtm-labs/rockscache
