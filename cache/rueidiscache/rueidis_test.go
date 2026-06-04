package rueidiscache

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/redis/rueidis"
	"github.com/stretchr/testify/assert"
)

// TestNewRueiis 验证 rueidis 客户端基础缓存命令。
func TestNewRueiis(t *testing.T) {
	client, err := NewRueidisClient(&rueidis.ClientOption{
		Username:    "",
		Password:    "123456",
		InitAddress: []string{"127.0.0.1:6379"},
		SelectDB:    0,
	})
	if err != nil {
		t.Skipf("redis unavailable: %v", err)
	}
	defer client.Close()
	if err := client.Do(context.Background(), client.B().Ping().Build()).Error(); err != nil {
		t.Skipf("redis unavailable: %v", err)
	}
	client.DoMulti(
		context.Background(),
		client.B().Hmset().Key("myhash").FieldValue().FieldValue("1", "a").FieldValue("2", "b").Build(),
		client.B().Expire().Key("myhash").Seconds(1000).Build(),
	)

	array, err2 := client.DoCache(context.Background(), client.B().Hmget().Key("myhash").Field("1", "2").Cache(), time.Minute).ToArray()
	if err2 != nil {
		t.Fatal(err2)
	}
	fmt.Printf("%+v \n", array)
	assert.Len(t, array, 2)
	assert.Equal(t, nil, err)
}

func TestNewRueidisClientReturnsConfigError(t *testing.T) {
	client, err := NewRueidisClient(&rueidis.ClientOption{})
	assert.Nil(t, client)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no alive address")
}

// TestNewRueidisAside 验证 cache-aside 客户端的加载和缓存命中。
func TestNewRueidisAside(t *testing.T) {
	ctx := context.Background()
	client, err := NewRueidisAsideClient(&rueidis.ClientOption{
		Username:    "",
		Password:    "123456",
		InitAddress: []string{"127.0.0.1:6379"},
		SelectDB:    0,
	})
	if err != nil {
		t.Skipf("redis unavailable: %v", err)
	}
	defer client.Close()
	redisClient := client.Client()
	if err := redisClient.Do(ctx, redisClient.B().Ping().Build()).Error(); err != nil {
		t.Skipf("redis unavailable: %v", err)
	}

	key := fmt.Sprintf("godb:rueidisaside:%d", time.Now().UnixNano())
	defer client.Del(context.Background(), key)
	if err := client.Del(ctx, key); err != nil {
		t.Skipf("redis unavailable: %v", err)
	}

	probeKey := key + ":probe"
	defer client.Del(context.Background(), probeKey)
	if _, err := client.Get(ctx, time.Minute, probeKey, func(_ context.Context, _ string) (val string, err error) {
		return "probe", nil
	}); err != nil {
		t.Skipf("redis cache-aside unavailable: %v", err)
	}

	loaderCalls := 0
	val, err := client.Get(ctx, time.Minute, key, func(_ context.Context, _ string) (val string, err error) {
		loaderCalls++
		return "abcd", nil
	})
	fmt.Println(err)
	fmt.Println(val)
	assert.Equal(t, nil, err)
	assert.Equal(t, "abcd", val)
	assert.Equal(t, 1, loaderCalls)

	val, err = client.Get(ctx, time.Minute, key, func(_ context.Context, _ string) (val string, err error) {
		loaderCalls++
		return "updated", nil
	})
	assert.Equal(t, nil, err)
	assert.Equal(t, "abcd", val)
	assert.Equal(t, 1, loaderCalls)
}
