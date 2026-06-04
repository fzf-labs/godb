package goredisdbcache

import (
	"context"
	"errors"
	"math/rand"
	"time"

	"github.com/fzf-labs/godb/orm/dbcache"
	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/singleflight"
)

// Cache 是基于 go-redis 的数据库查询缓存实现。
type Cache struct {
	name   string
	client *redis.Client
	ttl    time.Duration
	sf     singleflight.Group
}

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
		r.name = name
	}
}

// WithTTL 设置默认缓存过期时间。
func WithTTL(ttl time.Duration) CacheOption {
	return func(r *Cache) {
		r.ttl = ttl
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
	return r.ttl - time.Duration(rand.Float64()*0.1*float64(r.ttl)) //nolint:gosec
}

// Fetch 查询单个缓存值，未命中时调用回源函数并写入缓存。
func (r *Cache) Fetch(ctx context.Context, key string, fn func() (string, error), expire time.Duration) (string, error) {
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

// FetchHash 查询哈希字段缓存，未命中时回源并设置 hash key 过期时间。
func (r *Cache) FetchHash(ctx context.Context, key string, field string, fn func() (string, error), expire time.Duration) (string, error) {
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
	return r.client.Del(ctx, key).Err()
}

// DelBatch 批量删除缓存 key。
func (r *Cache) DelBatch(ctx context.Context, keys []string) error {
	return r.client.Del(ctx, keys...).Err()
}

// DelHash 删除 hash key 下的指定字段。
func (r *Cache) DelHash(ctx context.Context, key string, field string) error {
	return r.client.HDel(ctx, key, field).Err()
}
