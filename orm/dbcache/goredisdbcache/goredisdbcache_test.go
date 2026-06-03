package goredisdbcache

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/go-redis/redismock/v9"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

var client = redis.NewClient(&redis.Options{
	Addr:     "0.0.0.0:6379",
	Password: "123456",
})

func requireRedis(t *testing.T) {
	t.Helper()
	if err := client.Ping(context.Background()).Err(); err != nil {
		t.Skipf("redis unavailable: %v", err)
	}
}

func TestGoRedisCache_Fetch(t *testing.T) {
	requireRedis(t)
	cache := NewGoRedisDBCache(client, WithName("test"), WithTTL(time.Minute))
	ctx := context.Background()
	fetch, err := cache.Fetch(ctx, "GoRedisCache_Fetch", func() (string, error) {
		fmt.Println("do Fetch")
		return "GoRedisCache_Fetch: result", nil
	}, cache.TTL())
	fmt.Println(fetch)
	fmt.Println(err)
	assert.Equal(t, nil, err)
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
	fmt.Println(fetch)
	fmt.Println(err)
	assert.Equal(t, nil, err)
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
