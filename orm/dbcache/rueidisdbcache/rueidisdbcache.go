//nolint:gosec
package rueidisdbcache

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/redis/rueidis"
	"golang.org/x/sync/singleflight"

	"github.com/fzf-labs/godb/orm/dbcache"
)

// Cache 是基于 rueidis 的数据库查询缓存实现。
type Cache struct {
	name          string
	client        rueidis.Client
	ttl           time.Duration
	delayedDelete time.Duration
	sf            singleflight.Group
}

var delayedDeleteAfterFunc = time.AfterFunc
var errNilFetchCallback = errors.New("fetch callback cannot be nil")
var errNilFetchBatchCallback = errors.New("fetch batch callback cannot be nil")
var errNilFetchHashCallback = errors.New("fetch hash callback cannot be nil")

// CacheOption 配置 rueidis 数据库查询缓存。
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

// WithDelayedDelete enables best-effort delayed double delete after the given delay.
func WithDelayedDelete(delay time.Duration) CacheOption {
	return func(r *Cache) {
		r.delayedDelete = delay
	}
}

// NewRueidisDBCache 创建 rueidis 数据库查询缓存。
func NewRueidisDBCache(client rueidis.Client, opts ...CacheOption) *Cache {
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

// Key 生成带缓存名称前缀的缓存 key。
func (r *Cache) Key(keys ...any) string {
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
	return r.ttl - time.Duration(rand.Float64()*0.1*float64(r.ttl))
}

func (r *Cache) ensureClient() error {
	if r == nil || r.client == nil {
		return fmt.Errorf("rueidisdbcache client cannot be nil")
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
	do, err, _ := r.sf.Do(key, func() (any, error) {
		cacheValue := r.client.DoCache(ctx, r.client.B().Get().Key(key).Cache(), expire)
		if cacheValue.Error() != nil && !rueidis.IsRedisNil(cacheValue.Error()) {
			return "", cacheValue.Error()
		}
		if !rueidis.IsRedisNil(cacheValue.Error()) {
			resp, err := cacheValue.ToString()
			if err != nil {
				return "", err
			}
			return resp, nil
		}
		resp, err := fn()
		if err != nil {
			return "", err
		}
		err = r.client.Do(ctx, r.client.B().Set().Key(key).Value(resp).Ex(expire).Build()).Error()
		if err != nil {
			return "", err
		}
		return resp, nil
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
	commands := make([]rueidis.CacheableTTL, 0)
	for _, v := range keys {
		commands = append(commands, rueidis.CT(r.client.B().Get().Key(v).Cache(), expire))
	}
	cacheValue := r.client.DoMultiCache(ctx, commands...)
	miss := make([]string, 0)
	for k, v := range cacheValue {
		// Redis Nil 表示缓存未命中，不能继续 ToString，否则会把解析错误和未命中混在一起。
		if rueidis.IsRedisNil(v.Error()) {
			miss = append(miss, keys[k])
			resp[keys[k]] = ""
			continue
		}
		if v.Error() != nil {
			return nil, v.Error()
		}
		toString, err := v.ToString()
		if err != nil {
			return nil, err
		}
		resp[keys[k]] = toString
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
		completes := make([]rueidis.Completed, 0)
		for k, v := range dbValue {
			completes = append(completes, r.client.B().Set().Key(k).Value(v).Ex(expire).Build())
			resp[k] = v
		}
		multi := r.client.DoMulti(ctx, completes...)
		for _, result := range multi {
			err = result.Error()
			if err != nil {
				return nil, err
			}
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
	do, err, _ := r.sf.Do(hashFlightKey(key, field), func() (any, error) {
		cacheValue := r.client.DoCache(ctx, r.client.B().Hget().Key(key).Field(field).Cache(), expire)
		if cacheValue.Error() != nil && !rueidis.IsRedisNil(cacheValue.Error()) {
			return "", cacheValue.Error()
		}
		if !rueidis.IsRedisNil(cacheValue.Error()) {
			resp, err := cacheValue.ToString()
			if err != nil {
				return "", err
			}
			return resp, nil
		}
		resp, err := fn()
		if err != nil {
			return "", err
		}
		// HSET 不会继承 TTL，写入 hash 后需要显式设置 key 的过期时间。
		results := r.client.DoMulti(
			ctx,
			r.client.B().Hset().Key(key).FieldValue().FieldValue(field, resp).Build(),
			r.client.B().Pexpire().Key(key).Milliseconds(expire.Milliseconds()).Build(),
		)
		for _, result := range results {
			if err := result.Error(); err != nil {
				return "", err
			}
		}
		return resp, nil
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
	err := r.client.Do(ctx, r.client.B().Del().Key(key).Build()).Error()
	if err != nil {
		return err
	}
	if r.delayedDelete > 0 {
		delayedDeleteAfterFunc(r.delayedDelete, func() {
			_ = r.client.Do(context.Background(), r.client.B().Del().Key(key).Build()).Error()
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
	completes := make([]rueidis.Completed, 0)
	for _, v := range keys {
		completes = append(completes, r.client.B().Del().Key(v).Build())
	}
	multi := r.client.DoMulti(ctx, completes...)
	for _, result := range multi {
		err := result.Error()
		if err != nil {
			return err
		}
	}
	if r.delayedDelete > 0 {
		delayedKeys := append([]string(nil), keys...)
		delayedDeleteAfterFunc(r.delayedDelete, func() {
			completes := make([]rueidis.Completed, 0, len(delayedKeys))
			for _, v := range delayedKeys {
				completes = append(completes, r.client.B().Del().Key(v).Build())
			}
			multi := r.client.DoMulti(context.Background(), completes...)
			for _, result := range multi {
				_ = result.Error()
			}
		})
	}
	return nil
}

// DelHash 删除 hash key 下的指定字段。
func (r *Cache) DelHash(ctx context.Context, key string, field string) error {
	if err := r.ensureClient(); err != nil {
		return err
	}
	err := r.client.Do(ctx, r.client.B().Hdel().Key(key).Field(field).Build()).Error()
	if err != nil {
		return err
	}
	if r.delayedDelete > 0 {
		delayedDeleteAfterFunc(r.delayedDelete, func() {
			_ = r.client.Do(context.Background(), r.client.B().Hdel().Key(key).Field(field).Build()).Error()
		})
	}
	return nil
}
