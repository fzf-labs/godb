package gorediscache

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewGoRedis(t *testing.T) {
	newGoRedis, err := NewGoRedis(GoRedisConfig{
		Addr:     "127.0.0.1:6379",
		Password: "123456",
		DB:       0,
	})
	if err != nil {
		t.Skipf("redis unavailable: %v", err)
	}
	key := "godb:gorediscache:test"
	require.NoError(t, newGoRedis.Set(context.Background(), key, "ok", time.Minute).Err())
	value, err := newGoRedis.Get(context.Background(), key).Result()
	require.NoError(t, err)
	assert.Equal(t, "ok", value)
}

func TestStringToKV_PreservesValueAfterFirstColon(t *testing.T) {
	key, value := stringToKV("module:name:1.0")
	assert.Equal(t, "module", key)
	assert.Equal(t, "name:1.0", value)
}
