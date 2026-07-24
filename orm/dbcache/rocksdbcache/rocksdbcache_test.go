package rocksdbcache

import (
	"context"
	"testing"
	"time"

	"github.com/dtm-labs/rockscache"
	"github.com/go-redis/redismock/v9"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

// newWeakRocksCacheClient 创建测试用 RocksCache 客户端。
func newWeakRocksCacheClient(rdb *redis.Client) *rockscache.Client {
	rc := rockscache.NewClient(rdb, rockscache.NewDefaultOptions())
	// 常用参数设置
	// 1、强一致性(默认关闭强一致性，如果开启的话会影响性能)
	rc.Options.StrongConsistency = false
	// 2、redis出现问题需要缓存降级时设置为true
	rc.Options.DisableCacheRead = false   // 关闭缓存读，默认false；如果打开，那么Fetch就不从缓存读取数据，而是直接调用fn获取数据
	rc.Options.DisableCacheDelete = false // 关闭缓存删除，默认false；如果打开，那么TagAsDeleted就什么操作都不做，直接返回
	// 3、其他设置
	// 标记删除的延迟时间，默认10秒，设置为3秒表示：被删除的key在3秒后才从redis中彻底清除
	rc.Options.Delay = time.Second * time.Duration(3)
	// 防穿透：若fn返回空字符串，空结果在缓存中的缓存时间，默认60秒
	rc.Options.EmptyExpire = time.Second * time.Duration(120)
	// 防雪崩,默认0.1,当前设置为0.1的话，如果设定为600的过期时间，那么过期时间会被设定为540s - 600s中间的一个随机数，避免数据出现同时到期
	rc.Options.RandomExpireAdjustment = 0.1 // 设置为默认就行
	return rc
}

// TestCache_Key 验证缓存 key 拼接。
func TestCache_Key(t *testing.T) {
	rdb, _ := redismock.NewClientMock()
	rocksCacheClient := newWeakRocksCacheClient(rdb)
	cache := NewRocksDBCache(rdb, rocksCacheClient, WithName("test"), WithTTL(time.Minute), WithBatchSize(100))
	key := cache.Key("a", "b", "c")
	assert.Equal(t, "test:a:b:c", key)
}

func TestNewRocksDBCacheOptionsAndTTL(t *testing.T) {
	rdb, _ := redismock.NewClientMock()
	rocksCacheClient := newWeakRocksCacheClient(rdb)
	cache := NewRocksDBCache(rdb, rocksCacheClient, WithName("custom"), WithTTL(time.Minute), WithBatchSize(2))

	assert.Equal(t, "custom", cache.name)
	assert.Equal(t, rdb, cache.redisClient)
	assert.Equal(t, rocksCacheClient, cache.rocksCacheClient)
	assert.Equal(t, time.Minute, cache.ttl)
	assert.Equal(t, 2, cache.batchSize)

	ttl := cache.TTL()
	assert.LessOrEqual(t, ttl, time.Minute)
	assert.GreaterOrEqual(t, ttl, 54*time.Second)
}

func TestRocksDBCacheWithNameTrimsAndKeepsDefaultForBlank(t *testing.T) {
	rdb, _ := redismock.NewClientMock()
	rocksCacheClient := newWeakRocksCacheClient(rdb)
	trimmed := NewRocksDBCache(rdb, rocksCacheClient, WithName("  custom  "))
	assert.Equal(t, "custom:a", trimmed.Key("a"))

	defaulted := NewRocksDBCache(rdb, rocksCacheClient, WithName("   "))
	assert.Equal(t, "GormCache:a", defaulted.Key("a"))
}

func TestRocksDBCacheTTLReturnsZeroForNonPositiveTTL(t *testing.T) {
	assert.Equal(t, time.Duration(0), NewRocksDBCache(nil, nil, WithTTL(0)).TTL())
	assert.Equal(t, time.Duration(0), NewRocksDBCache(nil, nil, WithTTL(-time.Minute)).TTL())
}

func TestRocksDBCacheRejectsNilClients(t *testing.T) {
	cache := NewRocksDBCache(nil, nil)
	ctx := context.Background()

	_, err := cache.Fetch(ctx, "key", func() (string, error) {
		t.Fatal("fetch callback should not run")
		return "", nil
	}, time.Minute)
	assert.ErrorContains(t, err, "rocksdbcache rocks cache client cannot be nil")

	_, err = cache.FetchBatch(ctx, []string{"a"}, func([]string) (map[string]string, error) {
		t.Fatal("fetch batch callback should not run")
		return nil, nil
	}, time.Minute)
	assert.ErrorContains(t, err, "rocksdbcache rocks cache client cannot be nil")

	_, err = cache.FetchHash(ctx, "key", "field", func() (string, error) {
		t.Fatal("fetch hash callback should not run")
		return "", nil
	}, time.Minute)
	assert.ErrorContains(t, err, "rocksdbcache redis client cannot be nil")

	assert.ErrorContains(t, cache.Del(ctx, "key"), "rocksdbcache rocks cache client cannot be nil")
	assert.ErrorContains(t, cache.DelBatch(ctx, []string{"a"}), "rocksdbcache rocks cache client cannot be nil")
	assert.ErrorContains(t, cache.DelHash(ctx, "key", "field"), "rocksdbcache redis client cannot be nil")
}

func TestRocksDBCacheRejectsNilFetchCallbacks(t *testing.T) {
	rdb, _ := redismock.NewClientMock()
	rocksCacheClient := newWeakRocksCacheClient(rdb)
	cache := NewRocksDBCache(rdb, rocksCacheClient)
	ctx := context.Background()

	_, err := cache.Fetch(ctx, "key", nil, time.Minute)
	assert.ErrorContains(t, err, "fetch callback cannot be nil")

	_, err = cache.FetchBatch(ctx, []string{"a"}, nil, time.Minute)
	assert.ErrorContains(t, err, "fetch batch callback cannot be nil")

	_, err = cache.FetchHash(ctx, "key", "field", nil, time.Minute)
	assert.ErrorContains(t, err, "fetch hash callback cannot be nil")
}

func TestUniquePreservesOrder(t *testing.T) {
	assert.Nil(t, unique(nil))
	assert.Equal(t, []string{"a", "b", "c"}, unique([]string{"a", "b", "a", "c", "b"}))
}

func TestChunkSplitsAndRejectsInvalidSize(t *testing.T) {
	got, err := chunk([]string{"a", "b", "c"}, 2)
	assert.NoError(t, err)
	assert.Equal(t, [][]string{{"a", "b"}, {"c"}}, got)

	got, err = chunk(nil, 2)
	assert.NoError(t, err)
	assert.Empty(t, got)

	_, err = chunk([]string{"a"}, 0)
	assert.Error(t, err)
}

func TestFetchBatchAndDelBatchRejectInvalidBatchSize(t *testing.T) {
	cache := NewRocksDBCache(nil, nil, WithBatchSize(0))

	_, err := cache.FetchBatch(context.Background(), []string{"a"}, func([]string) (map[string]string, error) {
		t.Fatal("fetch callback should not run when batch size is invalid")
		return nil, nil
	}, time.Minute)
	assert.Error(t, err)

	err = cache.DelBatch(context.Background(), []string{"a"})
	assert.Error(t, err)
}

func TestRocksDBCacheDelBatchEmptyIsNoop(t *testing.T) {
	cache := NewRocksDBCache(nil, nil, WithBatchSize(0))

	assert.NoError(t, cache.DelBatch(context.Background(), nil))
	assert.NoError(t, cache.DelBatch(context.Background(), []string{}))
}

func TestFetchUsesRocksCacheClient(t *testing.T) {
	rdb, _ := redismock.NewClientMock()
	rocksCacheClient := newWeakRocksCacheClient(rdb)
	rocksCacheClient.Options.DisableCacheRead = true
	cache := NewRocksDBCache(rdb, rocksCacheClient)

	got, err := cache.Fetch(context.Background(), "key", func() (string, error) {
		return "loaded", nil
	}, time.Minute)

	assert.NoError(t, err)
	assert.Equal(t, "loaded", got)
}

func TestFetchReturnsLoaderError(t *testing.T) {
	rdb, _ := redismock.NewClientMock()
	rocksCacheClient := newWeakRocksCacheClient(rdb)
	rocksCacheClient.Options.DisableCacheRead = true
	cache := NewRocksDBCache(rdb, rocksCacheClient)

	_, err := cache.Fetch(context.Background(), "key", func() (string, error) {
		return "", context.Canceled
	}, time.Minute)

	assert.ErrorIs(t, err, context.Canceled)
}

func TestFetchBatchUsesRocksCacheClient(t *testing.T) {
	rdb, _ := redismock.NewClientMock()
	rocksCacheClient := newWeakRocksCacheClient(rdb)
	rocksCacheClient.Options.DisableCacheRead = true
	cache := NewRocksDBCache(rdb, rocksCacheClient, WithBatchSize(2))

	got, err := cache.FetchBatch(context.Background(), []string{"a", "b", "a"}, func(miss []string) (map[string]string, error) {
		assert.Equal(t, []string{"a", "b"}, miss)
		return map[string]string{"a": "loaded-a", "b": "loaded-b"}, nil
	}, time.Minute)

	assert.NoError(t, err)
	assert.Equal(t, map[string]string{"a": "loaded-a", "b": "loaded-b"}, got)
}

func TestFetchBatchReturnsLoaderError(t *testing.T) {
	rdb, _ := redismock.NewClientMock()
	rocksCacheClient := newWeakRocksCacheClient(rdb)
	rocksCacheClient.Options.DisableCacheRead = true
	cache := NewRocksDBCache(rdb, rocksCacheClient, WithBatchSize(2))

	_, err := cache.FetchBatch(context.Background(), []string{"a", "b"}, func([]string) (map[string]string, error) {
		return nil, context.Canceled
	}, time.Minute)

	assert.ErrorIs(t, err, context.Canceled)
}

func TestFetchBatchRejectsMissingLoaderValues(t *testing.T) {
	rdb, _ := redismock.NewClientMock()
	rocksCacheClient := newWeakRocksCacheClient(rdb)
	rocksCacheClient.Options.DisableCacheRead = true
	cache := NewRocksDBCache(rdb, rocksCacheClient, WithBatchSize(2))

	_, err := cache.FetchBatch(context.Background(), []string{"a", "b"}, func([]string) (map[string]string, error) {
		return map[string]string{}, nil
	}, time.Minute)

	assert.Error(t, err)
}

func TestDelAndDelBatchUseRocksCacheClient(t *testing.T) {
	rdb, _ := redismock.NewClientMock()
	rocksCacheClient := newWeakRocksCacheClient(rdb)
	rocksCacheClient.Options.DisableCacheDelete = true
	cache := NewRocksDBCache(rdb, rocksCacheClient, WithBatchSize(2))

	assert.NoError(t, cache.Del(context.Background(), "key"))
	assert.NoError(t, cache.DelBatch(context.Background(), []string{"a", "b", "a"}))
}

func TestFetchHashCacheHit(t *testing.T) {
	rdb, mock := redismock.NewClientMock()
	cache := NewRocksDBCache(rdb, nil)

	mock.ExpectHGet("hash", "field").SetVal("cached")

	got, err := cache.FetchHash(context.Background(), "hash", "field", func() (string, error) {
		t.Fatal("loader should not run on cache hit")
		return "", nil
	}, time.Minute)

	assert.NoError(t, err)
	assert.Equal(t, "cached", got)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestFetchHashCacheMissStoresValue(t *testing.T) {
	rdb, mock := redismock.NewClientMock()
	cache := NewRocksDBCache(rdb, nil)

	mock.ExpectHGet("hash", "field").RedisNil()
	mock.ExpectHSet("hash", "field", "loaded").SetVal(1)
	mock.ExpectExpire("hash", time.Minute).SetVal(true)

	got, err := cache.FetchHash(context.Background(), "hash", "field", func() (string, error) {
		return "loaded", nil
	}, time.Minute)

	assert.NoError(t, err)
	assert.Equal(t, "loaded", got)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestFetchHashReturnsErrors(t *testing.T) {
	tests := []struct {
		name  string
		setup func(redismock.ClientMock)
		fn    func() (string, error)
	}{
		{
			name: "hget error",
			setup: func(mock redismock.ClientMock) {
				mock.ExpectHGet("hash", "field").SetErr(context.Canceled)
			},
			fn: func() (string, error) { return "unused", nil },
		},
		{
			name: "loader error",
			setup: func(mock redismock.ClientMock) {
				mock.ExpectHGet("hash", "field").RedisNil()
			},
			fn: func() (string, error) { return "", context.Canceled },
		},
		{
			name: "hset error",
			setup: func(mock redismock.ClientMock) {
				mock.ExpectHGet("hash", "field").RedisNil()
				mock.ExpectHSet("hash", "field", "loaded").SetErr(context.Canceled)
			},
			fn: func() (string, error) { return "loaded", nil },
		},
		{
			name: "expire error",
			setup: func(mock redismock.ClientMock) {
				mock.ExpectHGet("hash", "field").RedisNil()
				mock.ExpectHSet("hash", "field", "loaded").SetVal(1)
				mock.ExpectExpire("hash", time.Minute).SetErr(context.Canceled)
			},
			fn: func() (string, error) { return "loaded", nil },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rdb, mock := redismock.NewClientMock()
			cache := NewRocksDBCache(rdb, nil)
			tt.setup(mock)

			_, err := cache.FetchHash(context.Background(), "hash", "field", tt.fn, time.Minute)

			assert.Error(t, err)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestDelHash(t *testing.T) {
	rdb, mock := redismock.NewClientMock()
	cache := NewRocksDBCache(rdb, nil)
	mock.ExpectHDel("hash", "field").SetVal(1)

	assert.NoError(t, cache.DelHash(context.Background(), "hash", "field"))
	assert.NoError(t, mock.ExpectationsWereMet())
}
