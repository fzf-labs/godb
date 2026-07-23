package gorediscache

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redismock/v9"
	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/fzf-labs/godb/internal/testenv"
)

func TestNewGoRedis(t *testing.T) {
	newGoRedis, err := NewGoRedis(GoRedisConfig{
		Addr:     testenv.RedisAddr(),
		Password: testenv.RedisPassword(),
		DB:       0,
	})
	if err != nil {
		testenv.SkipIfUnavailable(t, "redis unavailable: %v", err)
	}
	key := "godb:gorediscache:test"
	require.NoError(t, newGoRedis.Set(context.Background(), key, "ok", time.Minute).Err())
	value, err := newGoRedis.Get(context.Background(), key).Result()
	require.NoError(t, err)
	assert.Equal(t, "ok", value)
}

func TestNewGoRedisWithInstrumentation(t *testing.T) {
	server, err := miniredis.Run()
	require.NoError(t, err)
	defer server.Close()

	client, err := NewGoRedis(GoRedisConfig{
		Addr:    server.Addr(),
		Tracing: true,
		Metrics: true,
	})
	require.NoError(t, err)
	defer client.Close()
}

func TestNewGoRedisReturnsPingError(t *testing.T) {
	server, err := miniredis.Run()
	require.NoError(t, err)
	addr := server.Addr()
	server.Close()

	client, err := NewGoRedis(GoRedisConfig{
		Addr:         addr,
		DialTimeout:  time.Millisecond,
		ReadTimeout:  time.Millisecond,
		WriteTimeout: time.Millisecond,
	})
	assert.Nil(t, client)
	assert.Error(t, err)
}

func TestNewGoRedisRejectsEmptyAddr(t *testing.T) {
	client, err := NewGoRedis(GoRedisConfig{})
	assert.Nil(t, client)
	assert.ErrorContains(t, err, "redis addr cannot be empty")
}

func TestNewGoRedisClosesClientWhenTracingFails(t *testing.T) {
	oldTracing, oldMetrics, oldPing, oldClose := instrumentTracing, instrumentMetrics, pingRedisClient, closeRedisClient
	t.Cleanup(func() {
		instrumentTracing = oldTracing
		instrumentMetrics = oldMetrics
		pingRedisClient = oldPing
		closeRedisClient = oldClose
	})

	closed := 0
	instrumentTracing = func(redis.UniversalClient, ...redisotel.TracingOption) error {
		return errors.New("tracing failed")
	}
	instrumentMetrics = func(redis.UniversalClient, ...redisotel.MetricsOption) error {
		t.Fatal("metrics should not run after tracing failure")
		return nil
	}
	pingRedisClient = func(*redis.Client) error {
		t.Fatal("ping should not run after tracing failure")
		return nil
	}
	closeRedisClient = func(*redis.Client) error {
		closed++
		return nil
	}

	client, err := NewGoRedis(GoRedisConfig{Addr: "127.0.0.1:1", Tracing: true})
	assert.Nil(t, client)
	assert.Error(t, err)
	assert.Equal(t, 1, closed)
}

func TestNewGoRedisClosesClientWhenMetricsFail(t *testing.T) {
	oldTracing, oldMetrics, oldPing, oldClose := instrumentTracing, instrumentMetrics, pingRedisClient, closeRedisClient
	t.Cleanup(func() {
		instrumentTracing = oldTracing
		instrumentMetrics = oldMetrics
		pingRedisClient = oldPing
		closeRedisClient = oldClose
	})

	closed := 0
	instrumentTracing = func(redis.UniversalClient, ...redisotel.TracingOption) error {
		return nil
	}
	instrumentMetrics = func(redis.UniversalClient, ...redisotel.MetricsOption) error {
		return errors.New("metrics failed")
	}
	pingRedisClient = func(*redis.Client) error {
		t.Fatal("ping should not run after metrics failure")
		return nil
	}
	closeRedisClient = func(*redis.Client) error {
		closed++
		return nil
	}

	client, err := NewGoRedis(GoRedisConfig{Addr: "127.0.0.1:1", Tracing: true, Metrics: true})
	assert.Nil(t, client)
	assert.Error(t, err)
	assert.Equal(t, 1, closed)
}

func TestNewGoRedisClosesClientWhenPingFails(t *testing.T) {
	oldTracing, oldMetrics, oldPing, oldClose := instrumentTracing, instrumentMetrics, pingRedisClient, closeRedisClient
	t.Cleanup(func() {
		instrumentTracing = oldTracing
		instrumentMetrics = oldMetrics
		pingRedisClient = oldPing
		closeRedisClient = oldClose
	})

	closed := 0
	instrumentTracing = func(redis.UniversalClient, ...redisotel.TracingOption) error {
		return nil
	}
	instrumentMetrics = func(redis.UniversalClient, ...redisotel.MetricsOption) error {
		return nil
	}
	pingRedisClient = func(*redis.Client) error {
		return errors.New("ping failed")
	}
	closeRedisClient = func(*redis.Client) error {
		closed++
		return nil
	}

	client, err := NewGoRedis(GoRedisConfig{Addr: "127.0.0.1:1", Tracing: true, Metrics: true})
	assert.Nil(t, client)
	assert.Error(t, err)
	assert.Equal(t, 1, closed)
}

func TestStringToKV_PreservesValueAfterFirstColon(t *testing.T) {
	key, value := stringToKV("module:name:1.0")
	assert.Equal(t, "module", key)
	assert.Equal(t, "name:1.0", value)
}

func TestRedisInfoParsesInfo(t *testing.T) {
	client, mock := redismock.NewClientMock()
	mock.ExpectInfo("server").SetVal("# Server\r\nredis_version:7.2.0\r\nconnected_clients:3\r\n\r\n")

	info := RedisInfo(client, "server")
	assert.Equal(t, "7.2.0", info["redis_version"])
	assert.Equal(t, "3", info["connected_clients"])
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRedisInfoReturnsEmptyOnError(t *testing.T) {
	client, mock := redismock.NewClientMock()
	mock.ExpectInfo().SetErr(context.Canceled)

	info := RedisInfo(client)
	assert.Empty(t, info)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRedisInfoReturnsEmptyOnNilClient(t *testing.T) {
	assert.Empty(t, RedisInfo(nil))
}

func TestRedisInfoParsesLargeLines(t *testing.T) {
	client, mock := redismock.NewClientMock()
	mock.ExpectInfo().SetVal("redis_version:" + strings.Repeat("1", 70*1024) + "\n")

	info := RedisInfo(client)
	assert.Equal(t, strings.Repeat("1", 70*1024), info["redis_version"])
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestStringToLines(t *testing.T) {
	lines, err := stringToLines("a\nb\n")
	require.NoError(t, err)
	assert.Equal(t, []string{"a", "b"}, lines)
}

func TestDBSize(t *testing.T) {
	client, mock := redismock.NewClientMock()
	mock.ExpectDBSize().SetVal(12)
	assert.Equal(t, int64(12), DBSize(client))
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDBSizeReturnsZeroOnError(t *testing.T) {
	client, mock := redismock.NewClientMock()
	mock.ExpectDBSize().SetErr(context.Canceled)
	assert.Equal(t, int64(0), DBSize(client))
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDBSizeReturnsZeroOnNilClient(t *testing.T) {
	assert.Equal(t, int64(0), DBSize(nil))
}

func TestStringToKVWithoutSeparator(t *testing.T) {
	key, value := stringToKV("standalone")
	assert.Equal(t, "standalone", key)
	assert.Equal(t, "", value)
}
