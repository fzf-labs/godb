package rueidiscache

import (
	"context"
	"time"

	"github.com/redis/rueidis"
	"github.com/redis/rueidis/rueidislock"
)

func NewLocker(option rueidislock.LockerOption) *Locker {
	return &Locker{option: option}
}

func NewDefaultLockerOption(client rueidis.Client) rueidislock.LockerOption {
	return rueidislock.LockerOption{
		ClientBuilder: func(_ rueidis.ClientOption) (rueidis.Client, error) {
			return client, nil
		},
	}
}

type Locker struct {
	option rueidislock.LockerOption
}

// LockOnce 自动锁-一次
// 自动加锁与释放
func (l *Locker) LockOnce(ctx context.Context, key string, ttl time.Duration, fn func() error) error {
	locker, err := rueidislock.NewLocker(l.option)
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
	locker, err := rueidislock.NewLocker(l.option)
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
	locker, err := rueidislock.NewLocker(l.option)
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
