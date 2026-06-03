package gorediscache

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/bsm/redislock"
	"github.com/redis/go-redis/v9"
)

// NotObtained 未获取到锁
// 业务层使用errors.Is(err, NotObtained)判断是否未获取到锁，并抛出业务异常。
var NotObtained = errors.New("lock not obtained")

func classifyObtainErr(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, redislock.ErrNotObtained) {
		return fmt.Errorf("%w: %v", NotObtained, err)
	}
	return err
}

func NewLocker(rd *redis.Client) *Locker {
	return &Locker{
		locker: redislock.New(rd),
	}
}

// Locker Redis分布式锁
type Locker struct {
	locker *redislock.Client
}

// LockOnce 自动锁-一次
// 自动加锁与释放
func (r *Locker) LockOnce(ctx context.Context, key string, ttl time.Duration, fn func() error) error {
	lock, err := r.locker.Obtain(ctx, key, ttl, nil)
	if err != nil {
		return classifyObtainErr(err)
	}
	defer func(lock *redislock.Lock, ctx context.Context) {
		_ = lock.Release(ctx)
	}(lock, ctx)
	return fn()
}

// LockRetry 自动锁-重试
// 自动加锁与释放，间隔100ms 重试3次
func (r *Locker) LockRetry(ctx context.Context, key string, ttl time.Duration, fn func() error) error {
	lock, err := r.locker.Obtain(ctx, key, ttl, &redislock.Options{
		RetryStrategy: redislock.LimitRetry(redislock.LinearBackoff(100*time.Millisecond), 3),
	})
	if err != nil {
		return classifyObtainErr(err)
	}
	defer func(lock *redislock.Lock, ctx context.Context) {
		_ = lock.Release(ctx)
	}(lock, ctx)
	return fn()
}

// LockWithCustom 自动锁-自定义
// 自定义时间间隔和重试次数
func (r *Locker) LockWithCustom(ctx context.Context, key string, ttl, retryDuration time.Duration, retryNum int, fn func() error) error {
	lock, err := r.locker.Obtain(ctx, key, ttl, &redislock.Options{
		RetryStrategy: redislock.LimitRetry(redislock.LinearBackoff(retryDuration), retryNum),
	})
	if err != nil {
		return classifyObtainErr(err)
	}
	defer func(lock *redislock.Lock, ctx context.Context) {
		_ = lock.Release(ctx)
	}(lock, ctx)
	return fn()
}

// LockOnceNotRelease 自动锁-一次-不释放
func (r *Locker) LockOnceNotRelease(ctx context.Context, key string, ttl time.Duration, fn func() error) error {
	_, err := r.locker.Obtain(ctx, key, ttl, nil)
	if err != nil {
		return classifyObtainErr(err)
	}
	return fn()
}
