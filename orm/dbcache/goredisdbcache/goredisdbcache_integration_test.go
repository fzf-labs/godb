//go:build integration

package goredisdbcache

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

func TestGoRedisCache_Fetch(t *testing.T) {
	client := newIntegrationRedisClient(t)
	cache := NewGoRedisDBCache(client, WithName("test"), WithTTL(time.Minute))
	ctx := context.Background()
	fetch, err := cache.Fetch(ctx, "GoRedisCache_Fetch", func() (string, error) {
		return "GoRedisCache_Fetch: result", nil
	}, cache.TTL())
	assert.NoError(t, err)
	assert.Equal(t, "GoRedisCache_Fetch: result", fetch)
}

func TestGoRedisCache_FetchBatch(t *testing.T) {
	client := newIntegrationRedisClient(t)
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
	client := newIntegrationRedisClient(t)
	cache := NewGoRedisDBCache(client)
	ctx := context.Background()
	err := cache.Del(ctx, "GoRedisCache_Fetch")
	assert.NoError(t, err)
}

func TestCache_DelBatch(t *testing.T) {
	client := newIntegrationRedisClient(t)
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
