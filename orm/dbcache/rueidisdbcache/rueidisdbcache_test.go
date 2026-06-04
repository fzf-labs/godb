package rueidisdbcache

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/fzf-labs/godb/cache/rueidiscache"
	"github.com/redis/rueidis"
	"github.com/stretchr/testify/assert"
)

func requireRueidis(t *testing.T) rueidis.Client {
	t.Helper()
	client, err := rueidiscache.NewRueidisClient(&rueidis.ClientOption{
		Username:    "",
		Password:    "123456",
		InitAddress: []string{"127.0.0.1:6379"},
		SelectDB:    0,
	})
	if err != nil {
		t.Skipf("redis unavailable: %v", err)
	}
	if err := client.Do(context.Background(), client.B().Ping().Build()).Error(); err != nil {
		client.Close()
		t.Skipf("redis unavailable: %v", err)
	}
	t.Cleanup(client.Close)
	return client
}

func TestRueidisCache_Take(t *testing.T) {
	client := requireRueidis(t)
	ctx := context.Background()
	rueidisCache := NewRueidisDBCache(client)
	take, err := rueidisCache.Fetch(ctx, "take_test", func() (string, error) {
		return "take", nil
	}, rueidisCache.TTL())
	fmt.Println(take)
	fmt.Println(err)
	assert.Equal(t, nil, err)
}

func TestRueidisCache_TakeBatch(t *testing.T) {
	client := requireRueidis(t)
	ctx := context.Background()
	rueidisCache := NewRueidisDBCache(client)
	keys := []string{
		"a",
		"b",
		"c",
		"d",
	}
	take, err := rueidisCache.FetchBatch(ctx, keys, func(miss []string) (map[string]string, error) {
		fmt.Println(miss)
		return map[string]string{
			"a": "test1",
			"b": "test2",
			"c": "test3",
			"d": "test4",
		}, nil
	}, rueidisCache.TTL())
	fmt.Println(take)
	fmt.Println(err)
	assert.Equal(t, nil, err)
}

func TestRueidisCache_Del(t *testing.T) {
	client := requireRueidis(t)
	ctx := context.Background()
	rueidisCache := NewRueidisDBCache(client)
	err := rueidisCache.Del(ctx, "a")
	assert.NoError(t, err)
}

func TestRueidisCache_DelBatch(t *testing.T) {
	client := requireRueidis(t)
	ctx := context.Background()
	rueidisCache := NewRueidisDBCache(client)
	err := rueidisCache.DelBatch(ctx, []string{"a", "b", "f"})
	assert.NoError(t, err)
}

func TestHashFlightKey_SeparatesKeyAndField(t *testing.T) {
	assert.Equal(t, "user:profile", hashFlightKey("user", "profile"))
	assert.NotEqual(t, hashFlightKey("ab", "c"), hashFlightKey("a", "bc"))
	assert.NotEqual(t, hashFlightKey("a:b", "c"), hashFlightKey("a", "b:c"))
}

func TestRueidisCacheOptionsKeyAndTTL(t *testing.T) {
	cache := NewRueidisDBCache(nil, WithName("custom"), WithTTL(time.Minute))

	assert.Equal(t, "custom:a:b", cache.Key("a", "b"))
	ttl := cache.TTL()
	assert.LessOrEqual(t, ttl, time.Minute)
	assert.GreaterOrEqual(t, ttl, 54*time.Second)
}

func newMiniRueidisClient(t *testing.T) (*miniredis.Miniredis, rueidis.Client) {
	t.Helper()
	server, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}
	client, err := rueidis.NewClient(rueidis.ClientOption{InitAddress: []string{server.Addr()}, DisableCache: true})
	if err != nil {
		server.Close()
		t.Fatal(err)
	}
	t.Cleanup(func() {
		client.Close()
		server.Close()
	})
	return server, client
}

func TestRueidisCacheFetchWithMiniredis(t *testing.T) {
	_, client := newMiniRueidisClient(t)
	cache := NewRueidisDBCache(client)
	ctx := context.Background()
	loads := 0

	got, err := cache.Fetch(ctx, "fetch:key", func() (string, error) {
		loads++
		return "loaded", nil
	}, time.Minute)

	assert.NoError(t, err)
	assert.Equal(t, "loaded", got)
	assert.Equal(t, 1, loads)

	got, err = cache.Fetch(ctx, "fetch:key", func() (string, error) {
		t.Fatal("loader should not run on cache hit")
		return "", nil
	}, time.Minute)
	assert.NoError(t, err)
	assert.Equal(t, "loaded", got)
}

func TestRueidisCacheFetchBatchAndHashWithMiniredis(t *testing.T) {
	server, client := newMiniRueidisClient(t)
	cache := NewRueidisDBCache(client)
	ctx := context.Background()
	assert.NoError(t, server.Set("batch:hit", "cached"))

	got, err := cache.FetchBatch(ctx, []string{"batch:miss", "batch:hit"}, func(miss []string) (map[string]string, error) {
		assert.Equal(t, []string{"batch:miss"}, miss)
		return map[string]string{"batch:miss": "loaded"}, nil
	}, time.Minute)

	assert.NoError(t, err)
	assert.Equal(t, map[string]string{"batch:miss": "loaded", "batch:hit": "cached"}, got)

	hashValue, err := cache.FetchHash(ctx, "hash:key", "field", func() (string, error) {
		return "hash-loaded", nil
	}, time.Minute)
	assert.NoError(t, err)
	assert.Equal(t, "hash-loaded", hashValue)

	assert.NoError(t, cache.DelHash(ctx, "hash:key", "field"))
}

func TestRueidisCacheHitAndLoaderErrorsWithMiniredis(t *testing.T) {
	server, client := newMiniRueidisClient(t)
	cache := NewRueidisDBCache(client)
	ctx := context.Background()

	server.HSet("hash:hit", "field", "cached")
	got, err := cache.FetchHash(ctx, "hash:hit", "field", func() (string, error) {
		t.Fatal("loader should not run on hash hit")
		return "", nil
	}, time.Minute)
	assert.NoError(t, err)
	assert.Equal(t, "cached", got)

	_, err = cache.Fetch(ctx, "fetch:error", func() (string, error) {
		return "", context.Canceled
	}, time.Minute)
	assert.ErrorIs(t, err, context.Canceled)

	_, err = cache.FetchBatch(ctx, []string{"batch:error"}, func([]string) (map[string]string, error) {
		return nil, context.Canceled
	}, time.Minute)
	assert.ErrorIs(t, err, context.Canceled)

	_, err = cache.FetchHash(ctx, "hash:error", "field", func() (string, error) {
		return "", context.Canceled
	}, time.Minute)
	assert.ErrorIs(t, err, context.Canceled)
}
