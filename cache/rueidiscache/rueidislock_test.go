package rueidiscache

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/redis/rueidis"
	"github.com/redis/rueidis/rueidislock"
	"github.com/stretchr/testify/assert"
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

// rueidislockOption 构造带 TTL 的 rueidis 分布式锁配置。
func rueidislockOption(ttl time.Duration) rueidislock.LockerOption {
	return rueidislock.LockerOption{
		KeyValidity: ttl,
	}
}

// TestLocker_AutoLock 验证 AutoLock 加锁和释放流程。
func TestLocker_AutoLock(t *testing.T) {
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
	locker := NewLocker(NewDefaultLockerOption(client))
	ctx := context.Background()
	err = locker.LockOnce(ctx, "test_lock", 10*time.Second, func() error {
		fmt.Println("test_lock do ")
		return nil
	})
	assert.NoError(t, err)
}
