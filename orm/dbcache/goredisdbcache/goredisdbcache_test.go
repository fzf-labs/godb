package goredisdbcache

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redismock/v9"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/fzf-labs/godb/internal/testenv"
)

var client = redis.NewClient(&redis.Options{
	Addr:     testenv.RedisAddr(),
	Password: testenv.RedisPassword(),
})

func requireRedis(t *testing.T) {
	t.Helper()
	if err := client.Ping(context.Background()).Err(); err != nil {
		testenv.SkipIfUnavailable(t, "redis unavailable: %v", err)
	}
}

func TestGoRedisCache_Fetch(t *testing.T) {
	requireRedis(t)
	cache := NewGoRedisDBCache(client, WithName("test"), WithTTL(time.Minute))
	ctx := context.Background()
	fetch, err := cache.Fetch(ctx, "GoRedisCache_Fetch", func() (string, error) {
		return "GoRedisCache_Fetch: result", nil
	}, cache.TTL())
	assert.NoError(t, err)
	assert.Equal(t, "GoRedisCache_Fetch: result", fetch)
}

func TestGoRedisCache_FetchBatch(t *testing.T) {
	requireRedis(t)
	cache := NewGoRedisDBCache(client)
	ctx := context.Background()
	keys := []string{
		"GoRedisCache_Fetch_a",
		"GoRedisCache_Fetch_b",
		"GoRedisCache_Fetch_c",
		"GoRedisCache_Fetch_d",
	}
	fetch, err := cache.FetchBatch(ctx, keys, func(miss []string) (map[string]string, error) {
		resp := make(map[string]string)
		for _, v := range miss {
			resp[v] = v + ": result"
		}
		return resp, nil
	}, cache.TTL())
	assert.NoError(t, err)
	assert.Equal(t, map[string]string{
		"GoRedisCache_Fetch_a": "GoRedisCache_Fetch_a: result",
		"GoRedisCache_Fetch_b": "GoRedisCache_Fetch_b: result",
		"GoRedisCache_Fetch_c": "GoRedisCache_Fetch_c: result",
		"GoRedisCache_Fetch_d": "GoRedisCache_Fetch_d: result",
	}, fetch)
}

func TestCache_Del(t *testing.T) {
	requireRedis(t)
	cache := NewGoRedisDBCache(client)
	ctx := context.Background()
	err := cache.Del(ctx, "GoRedisCache_Fetch")
	assert.NoError(t, err)
}

func TestCache_DelBatch(t *testing.T) {
	requireRedis(t)
	cache := NewGoRedisDBCache(client)
	ctx := context.Background()
	keys := []string{
		"GoRedisCache_Fetch_a",
		"GoRedisCache_Fetch_b",
		"GoRedisCache_Fetch_c",
		"GoRedisCache_Fetch_d",
	}
	err := cache.DelBatch(ctx, keys)
	assert.NoError(t, err)
}

func TestCache_Key(t *testing.T) {
	cache := NewGoRedisDBCache(client, WithName("test"), WithTTL(time.Minute))
	key := cache.Key("a", "b", "c")
	assert.Equal(t, key, "test:a:b:c")
}

func TestGoRedisCacheOptionsAndTTL(t *testing.T) {
	rdb, _ := redismock.NewClientMock()
	cache := NewGoRedisDBCache(rdb, WithName("custom"), WithTTL(time.Minute))

	assert.Equal(t, "custom:a", cache.Key("a"))
	ttl := cache.TTL()
	assert.LessOrEqual(t, ttl, time.Minute)
	assert.GreaterOrEqual(t, ttl, 54*time.Second)
}

func TestGoRedisCacheFetchHit(t *testing.T) {
	rdb, mock := redismock.NewClientMock()
	cache := NewGoRedisDBCache(rdb)
	mock.ExpectGet("key").SetVal("cached")

	got, err := cache.Fetch(context.Background(), "key", func() (string, error) {
		t.Fatal("loader should not run on cache hit")
		return "", nil
	}, time.Minute)

	assert.NoError(t, err)
	assert.Equal(t, "cached", got)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGoRedisCacheFetchMissStoresValue(t *testing.T) {
	rdb, mock := redismock.NewClientMock()
	cache := NewGoRedisDBCache(rdb)
	mock.ExpectGet("key").RedisNil()
	mock.ExpectSet("key", "loaded", time.Minute).SetVal("OK")

	got, err := cache.Fetch(context.Background(), "key", func() (string, error) {
		return "loaded", nil
	}, time.Minute)

	assert.NoError(t, err)
	assert.Equal(t, "loaded", got)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGoRedisCacheFetchErrors(t *testing.T) {
	tests := []struct {
		name  string
		setup func(redismock.ClientMock)
		fn    func() (string, error)
	}{
		{
			name: "get error",
			setup: func(mock redismock.ClientMock) {
				mock.ExpectGet("key").SetErr(context.Canceled)
			},
			fn: func() (string, error) { return "unused", nil },
		},
		{
			name: "loader error",
			setup: func(mock redismock.ClientMock) {
				mock.ExpectGet("key").RedisNil()
			},
			fn: func() (string, error) { return "", context.Canceled },
		},
		{
			name: "set error",
			setup: func(mock redismock.ClientMock) {
				mock.ExpectGet("key").RedisNil()
				mock.ExpectSet("key", "loaded", time.Minute).SetErr(context.Canceled)
			},
			fn: func() (string, error) { return "loaded", nil },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rdb, mock := redismock.NewClientMock()
			cache := NewGoRedisDBCache(rdb)
			tt.setup(mock)

			_, err := cache.Fetch(context.Background(), "key", tt.fn, time.Minute)

			assert.Error(t, err)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestGoRedisCacheFetchBatchHitsWithMock(t *testing.T) {
	rdb, mock := redismock.NewClientMock()
	cache := NewGoRedisDBCache(rdb)
	mock.ExpectGet("a").SetVal("cached-a")
	mock.ExpectGet("b").SetVal("cached-b")

	got, err := cache.FetchBatch(context.Background(), []string{"a", "b"}, func(miss []string) (map[string]string, error) {
		t.Fatalf("loader should not run on cache hit: %#v", miss)
		return nil, nil
	}, time.Minute)

	assert.NoError(t, err)
	assert.Equal(t, map[string]string{"a": "cached-a", "b": "cached-b"}, got)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGoRedisCacheFetchBatchMissStoresValue(t *testing.T) {
	rdb, mock := redismock.NewClientMock()
	cache := NewGoRedisDBCache(rdb)
	mock.ExpectGet("a").RedisNil()
	mock.ExpectSet("a", "loaded-a", time.Minute).SetVal("OK")

	got, err := cache.FetchBatch(context.Background(), []string{"a"}, func(miss []string) (map[string]string, error) {
		assert.Equal(t, []string{"a"}, miss)
		return map[string]string{"a": "loaded-a"}, nil
	}, time.Minute)

	assert.NoError(t, err)
	assert.Equal(t, map[string]string{"a": "loaded-a"}, got)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGoRedisCacheFetchBatchDeduplicatesKeys(t *testing.T) {
	server, err := miniredis.Run()
	require.NoError(t, err)
	t.Cleanup(server.Close)

	rdb := redis.NewClient(&redis.Options{Addr: server.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })
	cache := NewGoRedisDBCache(rdb)

	got, err := cache.FetchBatch(context.Background(), []string{"a", "a", "b"}, func(miss []string) (map[string]string, error) {
		assert.Equal(t, []string{"a", "b"}, miss)
		return map[string]string{"a": "loaded-a", "b": "loaded-b"}, nil
	}, time.Minute)

	assert.NoError(t, err)
	assert.Equal(t, map[string]string{"a": "loaded-a", "b": "loaded-b"}, got)
}

func TestGoRedisCacheFetchBatchErrors(t *testing.T) {
	tests := []struct {
		name  string
		setup func(redismock.ClientMock)
		fn    func([]string) (map[string]string, error)
	}{
		{
			name: "get error",
			setup: func(mock redismock.ClientMock) {
				mock.ExpectGet("a").SetErr(context.Canceled)
			},
			fn: func([]string) (map[string]string, error) { return nil, nil },
		},
		{
			name: "loader error",
			setup: func(mock redismock.ClientMock) {
				mock.ExpectGet("a").RedisNil()
			},
			fn: func([]string) (map[string]string, error) { return nil, context.Canceled },
		},
		{
			name: "set error",
			setup: func(mock redismock.ClientMock) {
				mock.ExpectGet("a").RedisNil()
				mock.ExpectSet("a", "loaded-a", time.Minute).SetErr(context.Canceled)
			},
			fn: func([]string) (map[string]string, error) { return map[string]string{"a": "loaded-a"}, nil },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rdb, mock := redismock.NewClientMock()
			cache := NewGoRedisDBCache(rdb)
			tt.setup(mock)

			_, err := cache.FetchBatch(context.Background(), []string{"a"}, tt.fn, time.Minute)

			assert.Error(t, err)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestGoRedisCacheFetchBatchRejectsMissingLoaderValues(t *testing.T) {
	server, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(server.Close)

	rdb := redis.NewClient(&redis.Options{Addr: server.Addr()})
	cache := NewGoRedisDBCache(rdb)

	_, err = cache.FetchBatch(context.Background(), []string{"a"}, func([]string) (map[string]string, error) {
		return map[string]string{}, nil
	}, time.Minute)

	assert.Error(t, err)
}

func TestGoRedisCacheFetchHashHitAndErrors(t *testing.T) {
	t.Run("hit", func(t *testing.T) {
		rdb, mock := redismock.NewClientMock()
		cache := NewGoRedisDBCache(rdb)
		mock.ExpectHGet("hash", "field").SetVal("cached")

		got, err := cache.FetchHash(context.Background(), "hash", "field", func() (string, error) {
			t.Fatal("loader should not run on cache hit")
			return "", nil
		}, time.Minute)

		assert.NoError(t, err)
		assert.Equal(t, "cached", got)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

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
			cache := NewGoRedisDBCache(rdb)
			tt.setup(mock)

			_, err := cache.FetchHash(context.Background(), "hash", "field", tt.fn, time.Minute)

			assert.Error(t, err)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestGoRedisCacheDeletesWithMock(t *testing.T) {
	rdb, mock := redismock.NewClientMock()
	cache := NewGoRedisDBCache(rdb)

	mock.ExpectDel("key").SetVal(1)
	assert.NoError(t, cache.Del(context.Background(), "key"))

	mock.ExpectDel("a", "b").SetVal(2)
	assert.NoError(t, cache.DelBatch(context.Background(), []string{"a", "b"}))

	mock.ExpectHDel("hash", "field").SetVal(1)
	assert.NoError(t, cache.DelHash(context.Background(), "hash", "field"))

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGoRedisCacheDelayedDelete(t *testing.T) {
	rdb, mock := redismock.NewClientMock()
	cache := NewGoRedisDBCache(rdb, WithDelayedDelete(time.Millisecond))
	oldAfterFunc := delayedDeleteAfterFunc
	t.Cleanup(func() {
		delayedDeleteAfterFunc = oldAfterFunc
	})
	delayedDeleteAfterFunc = func(_ time.Duration, fn func()) *time.Timer {
		fn()
		return nil
	}

	mock.ExpectDel("key").SetVal(1)
	mock.ExpectDel("key").SetVal(1)
	assert.NoError(t, cache.Del(context.Background(), "key"))

	mock.ExpectDel("a", "b").SetVal(2)
	mock.ExpectDel("a", "b").SetVal(2)
	assert.NoError(t, cache.DelBatch(context.Background(), []string{"a", "b"}))

	mock.ExpectHDel("hash", "field").SetVal(1)
	mock.ExpectHDel("hash", "field").SetVal(1)
	assert.NoError(t, cache.DelHash(context.Background(), "hash", "field"))

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGoRedisCache_FetchHash_ScopesSingleflightByField(t *testing.T) {
	client, mock := redismock.NewClientMock()
	mock.MatchExpectationsInOrder(false)

	cache := NewGoRedisDBCache(client)
	ctx := context.Background()
	key := "shared-hash"
	expire := time.Minute

	mock.ExpectHGet(key, "field1").RedisNil()
	mock.ExpectHGet(key, "field2").RedisNil()
	mock.ExpectHSet(key, "field1", "value1").SetVal(1)
	mock.ExpectExpire(key, expire).SetVal(true)
	mock.ExpectHSet(key, "field2", "value2").SetVal(1)
	mock.ExpectExpire(key, expire).SetVal(true)

	field1Started := make(chan struct{})
	field2Started := make(chan struct{})
	releaseField1 := make(chan struct{})

	type result struct {
		value string
		err   error
	}

	result1Ch := make(chan result, 1)
	result2Ch := make(chan result, 1)

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		value, err := cache.FetchHash(ctx, key, "field1", func() (string, error) {
			close(field1Started)
			<-releaseField1
			return "value1", nil
		}, expire)
		result1Ch <- result{value: value, err: err}
	}()

	<-field1Started

	go func() {
		defer wg.Done()
		value, err := cache.FetchHash(ctx, key, "field2", func() (string, error) {
			close(field2Started)
			return "value2", nil
		}, expire)
		result2Ch <- result{value: value, err: err}
	}()

	select {
	case <-field2Started:
	case <-time.After(time.Second):
		t.Fatal("second field fetch did not start while first field was inflight")
	}

	close(releaseField1)
	wg.Wait()

	result1 := <-result1Ch
	result2 := <-result2Ch
	assert.NoError(t, result1.err)
	assert.NoError(t, result2.err)
	assert.Equal(t, "value1", result1.value)
	assert.Equal(t, "value2", result2.value)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHashFlightKey_SeparatesKeyAndField(t *testing.T) {
	assert.Equal(t, "user:profile", hashFlightKey("user", "profile"))
	assert.NotEqual(t, hashFlightKey("ab", "c"), hashFlightKey("a", "bc"))
	assert.NotEqual(t, hashFlightKey("a:b", "c"), hashFlightKey("a", "b:c"))
}
