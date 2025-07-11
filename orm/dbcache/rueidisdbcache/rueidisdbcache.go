//nolint:gosec
package rueidisdbcache

import (
	"context"
	"math/rand"
	"strings"
	"time"

	"github.com/fzf-labs/godb/orm/utils/strutil"
	"github.com/redis/rueidis"
	"golang.org/x/sync/singleflight"
)

type Cache struct {
	name   string
	client rueidis.Client
	ttl    time.Duration
	sf     singleflight.Group
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

func (r *Cache) Key(keys ...any) string {
	keyStr := make([]string, 0)
	keyStr = append(keyStr, r.name)
	for _, v := range keys {
		keyStr = append(keyStr, strutil.ConvToString(v))
	}
	return strings.Join(keyStr, ":")
}

func (r *Cache) TTL() time.Duration {
	return r.ttl - time.Duration(rand.Float64()*0.1*float64(r.ttl))
}

func (r *Cache) Fetch(ctx context.Context, key string, fn func() (string, error), expire time.Duration) (string, error) {
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
func (r *Cache) FetchBatch(ctx context.Context, keys []string, fn func(miss []string) (map[string]string, error), expire time.Duration) (map[string]string, error) {
	resp := make(map[string]string)
	commands := make([]rueidis.CacheableTTL, 0)
	for _, v := range keys {
		commands = append(commands, rueidis.CT(r.client.B().Get().Key(v).Cache(), expire))
	}
	cacheValue := r.client.DoMultiCache(ctx, commands...)
	miss := make([]string, 0)
	for k, v := range cacheValue {
		if rueidis.IsRedisNil(v.Error()) {
			miss = append(miss, keys[k])
		}
		toString, _ := v.ToString()
		resp[keys[k]] = toString
	}
	if len(miss) > 0 {
		dbValue, err := fn(miss)
		if err != nil {
			return nil, err
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

func (r *Cache) FetchHash(ctx context.Context, key string, field string, fn func() (string, error), expire time.Duration) (string, error) {
	do, err, _ := r.sf.Do(key+field, func() (any, error) {
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
		err = r.client.Do(ctx, r.client.B().Hset().Key(key).FieldValue().FieldValue(field, resp).Build()).Error()
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

func (r *Cache) Del(ctx context.Context, key string) error {
	return r.client.Do(ctx, r.client.B().Del().Key(key).Build()).Error()
}
func (r *Cache) DelBatch(ctx context.Context, keys []string) error {
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
	return nil
}
func (r *Cache) DelHash(ctx context.Context, key string, field string) error {
	return r.client.Do(ctx, r.client.B().Hdel().Key(key).Field(field).Build()).Error()
}
