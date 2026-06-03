package rueidisdbcache

import (
	"context"
	"fmt"
	"testing"

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
