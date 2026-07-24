package rueidiscache

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/rueidis"
	"github.com/redis/rueidis/rueidislock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestLockerOptionWithTTL 验证自定义锁 TTL 配置。
func TestLockerOptionWithTTL(t *testing.T) {
	defaultTTL := 5 * time.Second
	customTTL := 10 * time.Second
	locker := NewLocker(rueidislockOption(defaultTTL))

	option := locker.optionWithTTL(customTTL)
	assert.Equal(t, customTTL, option.KeyValidity)

	option = locker.optionWithTTL(0)
	assert.Equal(t, defaultTTL, option.KeyValidity)
}

func TestNewLockerAndDefaultOption(t *testing.T) {
	option := NewDefaultLockerOption(nil)
	client, err := option.ClientBuilder(rueidis.ClientOption{})
	assert.ErrorContains(t, err, "rueidis client cannot be nil")
	assert.Nil(t, client)

	locker := NewLocker(option)
	assert.NotNil(t, locker)
	assert.NotNil(t, locker.option.ClientBuilder)
}

func TestLockerMethodsReturnBuilderError(t *testing.T) {
	builderErr := assert.AnError
	locker := NewLocker(rueidislock.LockerOption{
		ClientBuilder: func(rueidis.ClientOption) (rueidis.Client, error) {
			return nil, builderErr
		},
	})

	tests := []struct {
		name string
		fn   func(func() error) error
	}{
		{
			name: "once",
			fn: func(callback func() error) error {
				return locker.LockOnce(context.Background(), "lock-key", time.Second, callback)
			},
		},
		{
			name: "retry",
			fn: func(callback func() error) error {
				return locker.LockRetry(context.Background(), "lock-key", time.Second, callback)
			},
		},
		{
			name: "not release",
			fn: func(callback func() error) error {
				return locker.LockOnceNotRelease(context.Background(), "lock-key", time.Second, callback)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			called := false
			err := tt.fn(func() error {
				called = true
				return nil
			})
			assert.ErrorIs(t, err, builderErr)
			assert.False(t, called)
		})
	}
}

func TestLockerMethodsRejectNilCallback(t *testing.T) {
	locker := NewLocker(rueidislock.LockerOption{
		ClientBuilder: func(rueidis.ClientOption) (rueidis.Client, error) {
			t.Fatal("expected nil callback validation before locker creation")
			return nil, nil
		},
	})

	tests := []struct {
		name string
		run  func() error
	}{
		{name: "once", run: func() error { return locker.LockOnce(context.Background(), "lock-key", time.Second, nil) }},
		{name: "retry", run: func() error { return locker.LockRetry(context.Background(), "lock-key", time.Second, nil) }},
		{name: "not release", run: func() error { return locker.LockOnceNotRelease(context.Background(), "lock-key", time.Second, nil) }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.ErrorContains(t, tt.run(), "lock callback cannot be nil")
		})
	}
}

func TestLockerMethodsRejectNilReceiver(t *testing.T) {
	tests := []struct {
		name string
		run  func(*Locker) error
	}{
		{
			name: "once",
			run: func(locker *Locker) error {
				return locker.LockOnce(context.Background(), "lock-key", time.Second, func() error { return nil })
			},
		},
		{
			name: "retry",
			run: func(locker *Locker) error {
				return locker.LockRetry(context.Background(), "lock-key", time.Second, func() error { return nil })
			},
		},
		{
			name: "not release",
			run: func(locker *Locker) error {
				return locker.LockOnceNotRelease(context.Background(), "lock-key", time.Second, func() error { return nil })
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var locker *Locker
			assert.ErrorContains(t, tt.run(locker), "rueidis locker cannot be nil")
		})
	}
}

func TestLockerMethodsUseRueidisLockClient(t *testing.T) {
	server, err := miniredis.Run()
	require.NoError(t, err)
	defer server.Close()

	client, err := rueidis.NewClient(rueidis.ClientOption{InitAddress: []string{server.Addr()}, DisableCache: true})
	require.NoError(t, err)
	defer client.Close()

	option := NewDefaultLockerOption(client)
	option.ClientOption.DisableCache = true
	option.FallbackSETPX = true
	option.KeyMajority = 1
	// TryNextAfter 同时是单次 acquire 超时；1ms 在 CI 负载下会偶发 DeadlineExceeded。
	option.TryNextAfter = 100 * time.Millisecond
	locker := NewLocker(option)
	ctx := context.Background()

	tests := []struct {
		name       string
		key        string
		wantExists bool
		fn         func(func() error) error
	}{
		{
			name:       "retry",
			key:        "lock:retry",
			wantExists: false,
			fn: func(callback func() error) error {
				return locker.LockRetry(ctx, "lock:retry", 2*time.Second, callback)
			},
		},
		{
			name:       "not release",
			key:        "lock:not-release",
			wantExists: true,
			fn: func(callback func() error) error {
				return locker.LockOnceNotRelease(ctx, "lock:not-release", 2*time.Second, callback)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			called := false
			err := tt.fn(func() error {
				called = true
				return nil
			})

			assert.NoError(t, err)
			assert.True(t, called)
			assert.Equal(t, tt.wantExists, server.Exists("rueidislock:0:"+tt.key))
		})
	}
}

// rueidislockOption 构造带 TTL 的 rueidis 分布式锁配置。
func rueidislockOption(ttl time.Duration) rueidislock.LockerOption {
	return rueidislock.LockerOption{
		KeyValidity: ttl,
	}
}
