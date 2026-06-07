package rueidisdbcache

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/fzf-labs/godb/cache/rueidiscache"
	"github.com/fzf-labs/godb/internal/testenv"
	"github.com/redis/rueidis"
	"github.com/stretchr/testify/assert"
)

func requireRueidis(t *testing.T) rueidis.Client {
	t.Helper()
	client, err := rueidiscache.NewRueidisClient(&rueidis.ClientOption{
		Username:    "",
		Password:    testenv.RedisPassword(),
		InitAddress: []string{testenv.RedisAddr()},
		SelectDB:    0,
	})
	if err != nil {
		testenv.SkipIfUnavailable(t, "redis unavailable: %v", err)
	}
	if err := client.Do(context.Background(), client.B().Ping().Build()).Error(); err != nil {
		client.Close()
		testenv.SkipIfUnavailable(t, "redis unavailable: %v", err)
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
	assert.NoError(t, err)
	assert.Equal(t, "take", take)
}

func TestRueidisCache_TakeBatch(t *testing.T) {
	client := requireRueidis(t)
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

func TestRueidisCacheDelayedDelete(t *testing.T) {
	_, client := newMiniRueidisClient(t)
	cache := NewRueidisDBCache(client, WithDelayedDelete(time.Millisecond))
	oldAfterFunc := delayedDeleteAfterFunc
	t.Cleanup(func() {
		delayedDeleteAfterFunc = oldAfterFunc
	})
	delayedDeleteAfterFunc = func(_ time.Duration, fn func()) *time.Timer {
		fn()
		return nil
	}

	ctx := context.Background()
	assert.NoError(t, cache.Del(ctx, "del:key"))
	assert.NoError(t, cache.DelBatch(ctx, []string{"del:a", "del:b"}))
	assert.NoError(t, cache.DelHash(ctx, "hash:key", "field"))
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
	client, err := rueidis.NewClient(rueidis.ClientOption{
		InitAddress:      []string{server.Addr()},
		Dialer:           net.Dialer{Timeout: time.Second},
		ConnWriteTimeout: time.Second,
		DisableRetry:     true,
		DisableCache:     true,
	})
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

func TestRueidisCacheFetchBatchRejectsMissingLoaderValues(t *testing.T) {
	_, client := newMiniRueidisClient(t)
	cache := NewRueidisDBCache(client)

	_, err := cache.FetchBatch(context.Background(), []string{"batch:missing"}, func([]string) (map[string]string, error) {
		return map[string]string{}, nil
	}, time.Minute)

	assert.Error(t, err)
}

func TestRueidisCacheReturnsBackendErrorsWithMiniredis(t *testing.T) {
	server, client := newMiniRueidisClient(t)
	cache := NewRueidisDBCache(client)
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	server.Close()

	_, err := cache.Fetch(ctx, "fetch:backend-error", func() (string, error) {
		t.Fatal("loader should not run when redis returns a backend error")
		return "", nil
	}, time.Minute)
	assert.Error(t, err)

	_, err = cache.FetchBatch(ctx, []string{"batch:backend-error"}, func([]string) (map[string]string, error) {
		t.Fatal("loader should not run when redis returns a backend error")
		return nil, nil
	}, time.Minute)
	assert.Error(t, err)

	_, err = cache.FetchHash(ctx, "hash:backend-error", "field", func() (string, error) {
		t.Fatal("loader should not run when redis returns a backend error")
		return "", nil
	}, time.Minute)
	assert.Error(t, err)

	assert.Error(t, cache.DelBatch(ctx, []string{"delete:backend-error"}))
}
