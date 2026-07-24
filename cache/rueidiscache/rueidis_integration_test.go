//go:build integration

package rueidiscache

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/redis/rueidis"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func integrationRueidisOption() *rueidis.ClientOption {
	addr := os.Getenv("GODB_TEST_REDIS_ADDR")
	if addr == "" {
		addr = "127.0.0.1:6379"
	}
	password := os.Getenv("GODB_TEST_REDIS_PASSWORD")
	if password == "" {
		password = "123456"
	}
	return &rueidis.ClientOption{
		Password:          password,
		InitAddress:       []string{addr},
		SelectDB:          0,
		PipelineMultiplex: -1,
	}
}

// TestNewRueiis 验证 rueidis 客户端基础缓存命令。
func TestNewRueiis(t *testing.T) {
	client, err := NewRueidisClient(integrationRueidisOption())
	require.NoError(t, err)
	t.Cleanup(client.Close)

	ctx := context.Background()
	require.NoError(t, client.Do(ctx, client.B().Ping().Build()).Error())
	client.DoMulti(
		ctx,
		client.B().Hmset().Key("myhash").FieldValue().FieldValue("1", "a").FieldValue("2", "b").Build(),
		client.B().Expire().Key("myhash").Seconds(1000).Build(),
	)

	array, err := client.DoCache(ctx, client.B().Hmget().Key("myhash").Field("1", "2").Cache(), time.Minute).ToArray()
	require.NoError(t, err)
	require.Len(t, array, 2)
	got := make([]string, 0, len(array))
	for _, msg := range array {
		value, err := msg.ToString()
		require.NoError(t, err)
		got = append(got, value)
	}
	assert.Equal(t, []string{"a", "b"}, got)
}

// TestNewRueidisAside 验证 cache-aside 客户端的加载和缓存命中。
func TestNewRueidisAside(t *testing.T) {
	ctx := context.Background()
	client, err := NewRueidisAsideClient(integrationRueidisOption())
	require.NoError(t, err)
	t.Cleanup(client.Close)

	redisClient := client.Client()
	require.NoError(t, redisClient.Do(ctx, redisClient.B().Ping().Build()).Error())

	key := fmt.Sprintf("godb:rueidisaside:%d", time.Now().UnixNano())
	probeKey := key + ":probe"
	t.Cleanup(func() {
		_ = client.Del(context.Background(), key)
		_ = client.Del(context.Background(), probeKey)
	})
	require.NoError(t, client.Del(ctx, key))

	_, err = client.Get(ctx, time.Minute, probeKey, func(_ context.Context, _ string) (val string, err error) {
		return "probe", nil
	})
	require.NoError(t, err)

	loaderCalls := 0
	val, err := client.Get(ctx, time.Minute, key, func(_ context.Context, _ string) (val string, err error) {
		loaderCalls++
		return "abcd", nil
	})
	assert.NoError(t, err)
	assert.Equal(t, "abcd", val)
	assert.Equal(t, 1, loaderCalls)

	val, err = client.Get(ctx, time.Minute, key, func(_ context.Context, _ string) (val string, err error) {
		loaderCalls++
		return "updated", nil
	})
	assert.NoError(t, err)
	assert.Equal(t, "abcd", val)
	assert.Equal(t, 1, loaderCalls)
}

// TestLocker_AutoLock 验证 AutoLock 加锁和释放流程。
func TestLocker_AutoLock(t *testing.T) {
	client, err := NewRueidisClient(integrationRueidisOption())
	require.NoError(t, err)
	t.Cleanup(client.Close)

	ctx := context.Background()
	require.NoError(t, client.Do(ctx, client.B().Ping().Build()).Error())

	option := NewDefaultLockerOption(client)
	option.FallbackSETPX = true
	option.KeyMajority = 1
	option.TryNextAfter = time.Second
	locker := NewLocker(option)
	key := fmt.Sprintf("test_lock:%d", time.Now().UnixNano())
	t.Cleanup(func() {
		_ = client.Do(context.Background(), client.B().Del().Key("rueidislock:0:"+key).Build()).Error()
	})

	err = locker.LockOnce(ctx, key, 10*time.Second, func() error {
		return nil
	})
	assert.NoError(t, err)
}
