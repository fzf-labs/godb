package rueidiscache

import (
	"context"
	"errors"
	"time"

	"github.com/redis/rueidis"
	"github.com/redis/rueidis/rueidislock"
)

var errNilLockCallback = errors.New("lock callback cannot be nil")

// NewLocker 创建基于 rueidislock 的分布式锁封装。
func NewLocker(option rueidislock.LockerOption) *Locker {
	return &Locker{option: option}
}

// NewDefaultLockerOption 使用已有 rueidis 客户端构造默认锁配置。
func NewDefaultLockerOption(client rueidis.Client) rueidislock.LockerOption {
	return rueidislock.LockerOption{
		ClientBuilder: func(_ rueidis.ClientOption) (rueidis.Client, error) {
			return client, nil
		},
	}
}

// Locker 封装 rueidis 分布式锁的常用加锁流程。
type Locker struct {
	option rueidislock.LockerOption
}

// optionWithTTL 返回带有指定 ttl 的配置副本。
// ttl 小于等于 0 时保留原 KeyValidity，用于沿用默认锁有效期。
func (l *Locker) optionWithTTL(ttl time.Duration) rueidislock.LockerOption {
	option := l.option
	if ttl > 0 {
		option.KeyValidity = ttl
	}
	return option
}

// LockOnce 自动锁-一次
// 自动加锁与释放
func (l *Locker) LockOnce(ctx context.Context, key string, ttl time.Duration, fn func() error) error {
	if fn == nil {
		return errNilLockCallback
	}
	locker, err := rueidislock.NewLocker(l.optionWithTTL(ttl))
	if err != nil {
		return err
	}
	defer locker.Close()
	_, cancel, err := locker.TryWithContext(ctx, key)
	if err != nil {
		return err
	}
	defer cancel()
	return fn()
}

// LockRetry 自动锁-重试
// 自动加锁与释放，间隔100ms 重试3次
func (l *Locker) LockRetry(ctx context.Context, key string, ttl time.Duration, fn func() error) error {
	if fn == nil {
		return errNilLockCallback
	}
	locker, err := rueidislock.NewLocker(l.optionWithTTL(ttl))
	if err != nil {
		return err
	}
	defer locker.Close()
	_, cancel, err := locker.WithContext(ctx, key)
	if err != nil {
		return err
	}
	defer cancel()
	return fn()
}

// LockOnceNotRelease 自动锁-一次-不释放
func (l *Locker) LockOnceNotRelease(ctx context.Context, key string, ttl time.Duration, fn func() error) error {
	if fn == nil {
		return errNilLockCallback
	}
	locker, err := rueidislock.NewLocker(l.optionWithTTL(ttl))
	if err != nil {
		return err
	}
	defer locker.Close()
	_, _, err = locker.WithContext(ctx, key)
	if err != nil {
		return err
	}
	return fn()
}
