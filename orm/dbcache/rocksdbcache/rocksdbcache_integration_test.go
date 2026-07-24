//go:build integration

package rocksdbcache

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newIntegrationRedisClient(t *testing.T) *redis.Client {
	t.Helper()
	addr := os.Getenv("GODB_TEST_REDIS_ADDR")
	if addr == "" {
		addr = "127.0.0.1:6379"
	}
	password := os.Getenv("GODB_TEST_REDIS_PASSWORD")
	if password == "" {
		password = "123456"
	}
	client := redis.NewClient(&redis.Options{Addr: addr, Password: password})
	t.Cleanup(func() {
		_ = client.Close()
	})
	require.NoError(t, client.Ping(context.Background()).Err())
	return client
}

// TestRocksCache_Fetch 验证单 key 缓存查询。
func TestRocksCache_Fetch(t *testing.T) {
	client := newIntegrationRedisClient(t)
	rocksCacheClient := newWeakRocksCacheClient(client)
	cache := NewRocksDBCache(client, rocksCacheClient, WithName("test"), WithTTL(time.Minute), WithBatchSize(100))
	ctx := context.Background()
	fetch, err := cache.Fetch(ctx, "RocksCache_Fetch", func() (string, error) {
		return "RocksCache_Fetch:result", nil
	}, cache.TTL())
	assert.NoError(t, err)
	assert.Equal(t, "RocksCache_Fetch:result", fetch)
}

// TestRocksCache_FetchBatch 验证批量缓存查询。
func TestRocksCache_FetchBatch(t *testing.T) {
	client := newIntegrationRedisClient(t)
	rocksCacheClient := newWeakRocksCacheClient(client)
	cache := NewRocksDBCache(client, rocksCacheClient, WithName("test"), WithTTL(time.Minute), WithBatchSize(100))
	ctx := context.Background()
	keys := []string{
		"RocksCache_FetchBatch_a",
		"RocksCache_FetchBatch_b",
		"RocksCache_FetchBatch_c",
	}
	take, err := cache.FetchBatch(ctx, keys, func(miss []string) (map[string]string, error) {
		resp := make(map[string]string)
		for _, v := range miss {
			resp[v] = v + ":result"
		}
		return resp, nil
	}, cache.TTL())
	assert.NoError(t, err)
	assert.Equal(t, map[string]string{
		"RocksCache_FetchBatch_a": "RocksCache_FetchBatch_a:result",
		"RocksCache_FetchBatch_b": "RocksCache_FetchBatch_b:result",
		"RocksCache_FetchBatch_c": "RocksCache_FetchBatch_c:result",
	}, take)
}

// TestCache_Del 验证单 key 删除标记。
func TestCache_Del(t *testing.T) {
	client := newIntegrationRedisClient(t)
	rocksCacheClient := newWeakRocksCacheClient(client)
	cache := NewRocksDBCache(client, rocksCacheClient, WithName("test"), WithTTL(time.Minute), WithBatchSize(100))
	ctx := context.Background()
	err := cache.Del(ctx, "RocksCache_Fetch")
	assert.NoError(t, err)
}

// TestCache_DelBatch 验证批量 key 删除标记。
func TestCache_DelBatch(t *testing.T) {
	client := newIntegrationRedisClient(t)
	rocksCacheClient := newWeakRocksCacheClient(client)
	cache := NewRocksDBCache(client, rocksCacheClient, WithName("test"), WithTTL(time.Minute), WithBatchSize(100))
	ctx := context.Background()
	keys := []string{
		"RocksCache_FetchBatch_a",
		"RocksCache_FetchBatch_b",
		"RocksCache_FetchBatch_c",
	}
	err := cache.DelBatch(ctx, keys)
	assert.NoError(t, err)
}
