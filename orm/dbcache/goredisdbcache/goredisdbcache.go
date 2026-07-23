package goredisdbcache

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/singleflight"

	"github.com/fzf-labs/godb/orm/dbcache"
)

// Cache 是基于 go-redis 的数据库查询缓存实现。
type Cache struct {
	name          string
	client        *redis.Client
	ttl           time.Duration
	delayedDelete time.Duration
	sf            singleflight.Group
}

var delayedDeleteAfterFunc = time.AfterFunc
var errNilFetchCallback = errors.New("fetch callback cannot be nil")
var errNilFetchBatchCallback = errors.New("fetch batch callback cannot be nil")
var errNilFetchHashCallback = errors.New("fetch hash callback cannot be nil")

// NewGoRedisDBCache 创建 go-redis 数据库查询缓存。
func NewGoRedisDBCache(client *redis.Client, opts ...CacheOption) *Cache {
	r := &Cache{
		name:   "GormCache",
		client: client,
		ttl:    time.Hour * 24,
	}
	if len(opts) > 0 {
		for _, v := range opts {
			v(r)
		}
	}
	return r
}

// CacheOption 配置 go-redis 数据库查询缓存。
type CacheOption func(cache *Cache)

// WithName 设置缓存 key 的命名前缀。
func WithName(name string) CacheOption {
	return func(r *Cache) {
		name = strings.TrimSpace(name)
		if name == "" {
			return
		}
		r.name = name
	}
}

// WithTTL 设置默认缓存过期时间。
func WithTTL(ttl time.Duration) CacheOption {
	return func(r *Cache) {
		r.ttl = ttl
	}
}

// WithDelayedDelete enables best-effort delayed double delete after the given delay.
func WithDelayedDelete(delay time.Duration) CacheOption {
	return func(r *Cache) {
		r.delayedDelete = delay
	}
}

// Key 生成带缓存名称前缀的缓存 key。
func (r *Cache) Key(keys ...interface{}) string {
	parts := make([]any, 0, len(keys)+1)
	parts = append(parts, r.name)
	for _, v := range keys {
		parts = append(parts, v)
	}
	return dbcache.BuildKey(parts...)
}

// TTL 返回带随机抖动的默认缓存过期时间。
func (r *Cache) TTL() time.Duration {
	if r.ttl <= 0 {
		return 0
	}
	return r.ttl - time.Duration(rand.Float64()*0.1*float64(r.ttl)) //nolint:gosec
}

func (r *Cache) ensureClient() error {
	if r == nil || r.client == nil {
		return fmt.Errorf("goredisdbcache client cannot be nil")
	}
	return nil
}

// Fetch 查询单个缓存值，未命中时调用回源函数并写入缓存。
func (r *Cache) Fetch(ctx context.Context, key string, fn func() (string, error), expire time.Duration) (string, error) {
	if fn == nil {
		return "", errNilFetchCallback
	}
	if err := r.ensureClient(); err != nil {
		return "", err
	}
	do, err, _ := r.sf.Do(key, func() (interface{}, error) {
		result, err := r.client.Get(ctx, key).Result()
		if err != nil && !errors.Is(err, redis.Nil) {
			return "", err
		}
		if result == "" && errors.Is(err, redis.Nil) {
			result, err = fn()
			if err != nil {
				return "", err
			}
			err = r.client.Set(ctx, key, result, expire).Err()
			if err != nil {
				return "", err
			}
		}
		return result, nil
	})
	if err != nil {
		return "", err
	}
	return do.(string), nil
}

// FetchBatch 批量查询缓存值，未命中时按缺失 key 回源并写入缓存。
func (r *Cache) FetchBatch(ctx context.Context, keys []string, fn func(miss []string) (map[string]string, error), expire time.Duration) (map[string]string, error) {
	if fn == nil {
		return nil, errNilFetchBatchCallback
	}
	keys = uniqueStrings(keys)
	if len(keys) == 0 {
		return map[string]string{}, nil
	}
	if err := r.ensureClient(); err != nil {
		return nil, err
	}
	resp := make(map[string]string)
	miss := make([]string, 0)
	pipelined, err := r.client.Pipelined(ctx, func(p redis.Pipeliner) error {
		for _, v := range keys {
			_, err := p.Get(ctx, v).Result()
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil && !errors.Is(err, redis.Nil) {
		return nil, err
	}
	for k, cmder := range pipelined {
		if errors.Is(cmder.Err(), redis.Nil) {
			miss = append(miss, keys[k])
		}
		resp[keys[k]] = cmder.(*redis.StringCmd).Val()
	}
	if len(miss) > 0 {
		dbValue, err := fn(miss)
		if err != nil {
			return nil, err
		}
		for _, key := range miss {
			if _, ok := dbValue[key]; !ok {
				return nil, fmt.Errorf("missing fetched value for key %q", key)
			}
		}
		_, err = r.client.Pipelined(ctx, func(p redis.Pipeliner) error {
			for _, v := range miss {
				resp[v] = dbValue[v]
				err = p.Set(ctx, v, dbValue[v], expire).Err()
				if err != nil {
					return err
				}
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	return resp, nil
}

func uniqueStrings(values []string) []string {
	if len(values) == 0 {
		return values
	}
	result := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}

// FetchHash 查询哈希字段缓存，未命中时回源并设置 hash key 过期时间。
func (r *Cache) FetchHash(ctx context.Context, key string, field string, fn func() (string, error), expire time.Duration) (string, error) {
	if fn == nil {
		return "", errNilFetchHashCallback
	}
	if err := r.ensureClient(); err != nil {
		return "", err
	}
	// Hash field 的回源需要按 key:field 去重，避免同一个 hash key 下不同 field 的请求串值。
	do, err, _ := r.sf.Do(hashFlightKey(key, field), func() (interface{}, error) {
		result, err := r.client.HGet(ctx, key, field).Result()
		if err != nil && !errors.Is(err, redis.Nil) {
			return "", err
		}
		if result == "" && errors.Is(err, redis.Nil) {
			result, err = fn()
			if err != nil {
				return "", err
			}
			err = r.client.HSet(ctx, key, field, result).Err()
			if err != nil {
				return "", err
			}
			err = r.client.Expire(ctx, key, expire).Err()
			if err != nil {
				return "", err
			}
		}
		return result, nil
	})
	if err != nil {
		return "", err
	}
	return do.(string), nil
}

func hashFlightKey(key, field string) string {
	return dbcache.BuildKey(key, field)
}

// Del 删除单个缓存 key。
func (r *Cache) Del(ctx context.Context, key string) error {
	if err := r.ensureClient(); err != nil {
		return err
	}
	err := r.client.Del(ctx, key).Err()
	if err != nil {
		return err
	}
	if r.delayedDelete > 0 {
		delayedDeleteAfterFunc(r.delayedDelete, func() {
			_ = r.client.Del(context.Background(), key).Err()
		})
	}
	return nil
}

// DelBatch 批量删除缓存 key。
func (r *Cache) DelBatch(ctx context.Context, keys []string) error {
	if len(keys) == 0 {
		return nil
	}
	if err := r.ensureClient(); err != nil {
		return err
	}
	err := r.client.Del(ctx, keys...).Err()
	if err != nil {
		return err
	}
	if r.delayedDelete > 0 {
		delayedKeys := append([]string(nil), keys...)
		delayedDeleteAfterFunc(r.delayedDelete, func() {
			_ = r.client.Del(context.Background(), delayedKeys...).Err()
		})
	}
	return nil
}

// DelHash 删除 hash key 下的指定字段。
func (r *Cache) DelHash(ctx context.Context, key string, field string) error {
	if err := r.ensureClient(); err != nil {
		return err
	}
	err := r.client.HDel(ctx, key, field).Err()
	if err != nil {
		return err
	}
	if r.delayedDelete > 0 {
		delayedDeleteAfterFunc(r.delayedDelete, func() {
			_ = r.client.HDel(context.Background(), key, field).Err()
		})
	}
	return nil
}
