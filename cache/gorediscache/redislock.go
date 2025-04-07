package gorediscache

import (
	"context"
	"time"

	"github.com/bsm/redislock"
	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
)

// NotObtained 未获取到锁
// 业务层使用errors.Is(err, NotObtained)判断是否未获取到锁，并抛出业务异常。
var NotObtained = errors.New("lock not obtained")

func NewRedisLock(rd *redis.Client) *RedisLock {
	return &RedisLock{
		locker: redislock.New(rd),
	}
}

// RedisLock Redis分布式锁
type RedisLock struct {
	locker *redislock.Client
}

// LockRetry 自动锁-重试
// 自动加锁与释放，间隔100ms 重试3次
// 使用场景：常规业务中使用，使用最频繁
func (r *RedisLock) LockRetry(ctx context.Context, key string, ttl time.Duration, fn func() error) error {
	lock, err := r.locker.Obtain(ctx, key, ttl, &redislock.Options{
		RetryStrategy: redislock.LimitRetry(redislock.LinearBackoff(100*time.Millisecond), 3),
	})
	if err != nil {
		return errors.Wrapf(NotObtained, "origin error is: %v", err)
	}
	defer func(lock *redislock.Lock, ctx context.Context) {
		_ = lock.Release(ctx)
	}(lock, ctx)
	return fn()
}

// LockRetryWithCustom 自动锁-重试
// 自定义时间间隔和重试次数
// 使用场景：常规业务中使用，使用频繁
func (r *RedisLock) LockRetryWithCustom(ctx context.Context, key string, ttl, retryDuration time.Duration, retryNum int, fn func() error) error {
	lock, err := r.locker.Obtain(ctx, key, ttl, &redislock.Options{
		RetryStrategy: redislock.LimitRetry(redislock.LinearBackoff(retryDuration), retryNum),
	})
	if err != nil {
		return errors.Wrapf(NotObtained, "origin error is: %v", err)
	}
	defer func(lock *redislock.Lock, ctx context.Context) {
		_ = lock.Release(ctx)
	}(lock, ctx)
	return fn()
}

// LockOnce 自动锁-一次
// 自动加锁与释放，不重试
// 使用场景：特殊业务中使用，使用较少
func (r *RedisLock) LockOnce(ctx context.Context, key string, ttl time.Duration, fn func() error) error {
	lock, err := r.locker.Obtain(ctx, key, ttl, nil)
	if err != nil {
		return errors.Wrapf(NotObtained, "origin error is: %v", err)
	}
	defer func(lock *redislock.Lock, ctx context.Context) {
		_ = lock.Release(ctx)
	}(lock, ctx)
	return fn()
}
