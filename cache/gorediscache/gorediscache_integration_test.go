//go:build integration

package gorediscache

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func integrationRedisConfig() GoRedisConfig {
	addr := os.Getenv("GODB_TEST_REDIS_ADDR")
	if addr == "" {
		addr = "127.0.0.1:6379"
	}
	password := os.Getenv("GODB_TEST_REDIS_PASSWORD")
	if password == "" {
		password = "123456"
	}
	return GoRedisConfig{Addr: addr, Password: password, DB: 0}
}

func TestNewGoRedis(t *testing.T) {
	client, err := NewGoRedis(integrationRedisConfig())
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, client.Close())
	})

	key := "godb:gorediscache:test"
	require.NoError(t, client.Set(context.Background(), key, "ok", time.Minute).Err())
	value, err := client.Get(context.Background(), key).Result()
	require.NoError(t, err)
	assert.Equal(t, "ok", value)
}
