package goredisdbcache

import (
	"context"
	"errors"
	"math/rand"
	"strings"
	"time"

	"github.com/fzf-labs/godb/orm/dbcache"
	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/singleflight"
)

type Cache struct {
	name   string
	client *redis.Client
	ttl    time.Duration
	sf     singleflight.Group
}

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

type CacheOption func(cache *Cache)

func WithName(name string) CacheOption {
	return func(r *Cache) {
		r.name = name
	}
}

func WithTTL(ttl time.Duration) CacheOption {
	return func(r *Cache) {
		r.ttl = ttl
	}
}

func (r *Cache) Key(keys ...interface{}) string {
	keyStr := make([]string, 0)
	keyStr = append(keyStr, r.name)
	for _, v := range keys {
		keyStr = append(keyStr, dbcache.KeyFormat(v))
	}
	return strings.Join(keyStr, ":")
}

func (r *Cache) TTL() time.Duration {
	return r.ttl - time.Duration(rand.Float64()*0.1*float64(r.ttl)) //nolint:gosec
}

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

func (r *Cache) FetchHash(ctx context.Context, key string, field string, fn func() (string, error), expire time.Duration) (string, error) {
	do, err, _ := r.sf.Do(key, func() (interface{}, error) {
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

func (r *Cache) Del(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}

func (r *Cache) DelBatch(ctx context.Context, keys []string) error {
	return r.client.Del(ctx, keys...).Err()
}

func (r *Cache) DelHash(ctx context.Context, key string, field string) error {
	return r.client.HDel(ctx, key, field).Err()
}
