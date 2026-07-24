//go:build integration

package rueidisdbcache

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/redis/rueidis"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/fzf-labs/godb/cache/rueidiscache"
)

func newIntegrationRueidisClient(t *testing.T) rueidis.Client {
	t.Helper()
	addr := os.Getenv("GODB_TEST_REDIS_ADDR")
	if addr == "" {
		addr = "127.0.0.1:6379"
	}
	password := os.Getenv("GODB_TEST_REDIS_PASSWORD")
	if password == "" {
		password = "123456"
	}
	client, err := rueidiscache.NewRueidisClient(&rueidis.ClientOption{
		Password:    password,
		InitAddress: []string{addr},
		SelectDB:    0,
	})
	require.NoError(t, err)
	t.Cleanup(client.Close)
	require.NoError(t, client.Do(context.Background(), client.B().Ping().Build()).Error())
	return client
}

func TestRueidisCache_Take(t *testing.T) {
	client := newIntegrationRueidisClient(t)
	ctx := context.Background()
	rueidisCache := NewRueidisDBCache(client)
	take, err := rueidisCache.Fetch(ctx, "take_test", func() (string, error) {
		return "take", nil
	}, rueidisCache.TTL())
	assert.NoError(t, err)
	assert.Equal(t, "take", take)
}

func TestRueidisCache_TakeBatch(t *testing.T) {
	client := newIntegrationRueidisClient(t)
	ctx := context.Background()
	rueidisCache := NewRueidisDBCache(client)
	prefix := "batch:" + time.Now().Format("20060102150405.000000000")
	keys := []string{
		prefix + ":a",
		prefix + ":b",
		prefix + ":c",
		prefix + ":d",
	}
	t.Cleanup(func() {
		_ = rueidisCache.DelBatch(ctx, keys)
	})
	take, err := rueidisCache.FetchBatch(ctx, keys, func(miss []string) (map[string]string, error) {
		assert.Equal(t, keys, miss)
		return map[string]string{
			keys[0]: "test1",
			keys[1]: "test2",
			keys[2]: "test3",
			keys[3]: "test4",
		}, nil
	}, rueidisCache.TTL())
	assert.NoError(t, err)
	assert.Equal(t, map[string]string{
		keys[0]: "test1",
		keys[1]: "test2",
		keys[2]: "test3",
		keys[3]: "test4",
	}, take)
}

func TestRueidisCache_Del(t *testing.T) {
	client := newIntegrationRueidisClient(t)
	ctx := context.Background()
	rueidisCache := NewRueidisDBCache(client)
	err := rueidisCache.Del(ctx, "a")
	assert.NoError(t, err)
}

func TestRueidisCache_DelBatch(t *testing.T) {
	client := newIntegrationRueidisClient(t)
	ctx := context.Background()
	rueidisCache := NewRueidisDBCache(client)
	err := rueidisCache.DelBatch(ctx, []string{"a", "b", "f"})
	assert.NoError(t, err)
}

func TestRueidisCacheFetchBatchDeduplicatesKeys(t *testing.T) {
	client := newIntegrationRueidisClient(t)
	cache := NewRueidisDBCache(client)
	ctx := context.Background()
	keys := []string{"batch:dup", "batch:other"}
	require.NoError(t, cache.DelBatch(ctx, keys))
	t.Cleanup(func() {
		_ = cache.DelBatch(context.Background(), keys)
	})

	got, err := cache.FetchBatch(ctx, []string{"batch:dup", "batch:dup", "batch:other"}, func(miss []string) (map[string]string, error) {
		require.Equal(t, keys, miss)
		return map[string]string{
			"batch:dup":   "loaded-dup",
			"batch:other": "loaded-other",
		}, nil
	}, time.Minute)

	assert.NoError(t, err)
	assert.Equal(t, map[string]string{"batch:dup": "loaded-dup", "batch:other": "loaded-other"}, got)
}
