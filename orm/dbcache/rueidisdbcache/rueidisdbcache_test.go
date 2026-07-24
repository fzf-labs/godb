package rueidisdbcache

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/rueidis"
	"github.com/stretchr/testify/assert"
)

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

func TestRueidisCacheWithNameTrimsAndKeepsDefaultForBlank(t *testing.T) {
	trimmed := NewRueidisDBCache(nil, WithName("  custom  "))
	assert.Equal(t, "custom:a", trimmed.Key("a"))

	defaulted := NewRueidisDBCache(nil, WithName("   "))
	assert.Equal(t, "GormCache:a", defaulted.Key("a"))
}

func TestRueidisCacheTTLReturnsZeroForNonPositiveTTL(t *testing.T) {
	assert.Equal(t, time.Duration(0), NewRueidisDBCache(nil, WithTTL(0)).TTL())
	assert.Equal(t, time.Duration(0), NewRueidisDBCache(nil, WithTTL(-time.Minute)).TTL())
}

func TestRueidisCacheRejectsNilClient(t *testing.T) {
	cache := NewRueidisDBCache(nil)
	ctx := context.Background()

	_, err := cache.Fetch(ctx, "key", func() (string, error) {
		t.Fatal("fetch callback should not run")
		return "", nil
	}, time.Minute)
	assert.ErrorContains(t, err, "rueidisdbcache client cannot be nil")

	_, err = cache.FetchBatch(ctx, []string{"a"}, func([]string) (map[string]string, error) {
		t.Fatal("fetch batch callback should not run")
		return nil, nil
	}, time.Minute)
	assert.ErrorContains(t, err, "rueidisdbcache client cannot be nil")

	_, err = cache.FetchHash(ctx, "key", "field", func() (string, error) {
		t.Fatal("fetch hash callback should not run")
		return "", nil
	}, time.Minute)
	assert.ErrorContains(t, err, "rueidisdbcache client cannot be nil")

	assert.ErrorContains(t, cache.Del(ctx, "key"), "rueidisdbcache client cannot be nil")
	assert.ErrorContains(t, cache.DelBatch(ctx, []string{"a"}), "rueidisdbcache client cannot be nil")
	assert.ErrorContains(t, cache.DelHash(ctx, "key", "field"), "rueidisdbcache client cannot be nil")
}

func TestRueidisCacheRejectsNilFetchCallbacks(t *testing.T) {
	cache := NewRueidisDBCache(nil)
	ctx := context.Background()

	_, err := cache.Fetch(ctx, "key", nil, time.Minute)
	assert.ErrorContains(t, err, "fetch callback cannot be nil")

	_, err = cache.FetchBatch(ctx, []string{"a"}, nil, time.Minute)
	assert.ErrorContains(t, err, "fetch batch callback cannot be nil")

	_, err = cache.FetchHash(ctx, "key", "field", nil, time.Minute)
	assert.ErrorContains(t, err, "fetch hash callback cannot be nil")
}

func TestRueidisCacheDelBatchEmptyIsNoop(t *testing.T) {
	cache := NewRueidisDBCache(nil)

	assert.NotPanics(t, func() {
		assert.NoError(t, cache.DelBatch(context.Background(), nil))
	})
	assert.NotPanics(t, func() {
		assert.NoError(t, cache.DelBatch(context.Background(), []string{}))
	})
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
