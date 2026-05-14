package rocksdbcache

import (
	"context"
	"errors"
	"math/rand/v2"
	"strings"
	"time"

	"github.com/dtm-labs/rockscache"
	"github.com/fzf-labs/godb/orm/dbcache"
	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/errgroup"
)

type Cache struct {
	name             string // 缓存名称
	redisClient      *redis.Client
	rocksCacheClient *rockscache.Client // RocksCache缓存客户端
	ttl              time.Duration      // redis缓存过期时间
	batchSize        int                // redis lua 批量查询数量  默认100 有些云厂商对lua的keys有限制
}

// NewRocksDBCache 创建基于 RocksCache 的数据库缓存。
func NewRocksDBCache(redisClient *redis.Client, rocksCacheClient *rockscache.Client, opts ...CacheOption) *Cache {
	r := &Cache{
		name:             "GormCache",
		redisClient:      redisClient,
		rocksCacheClient: rocksCacheClient,
		ttl:              time.Hour * 24,
		batchSize:        100,
	}
	if len(opts) > 0 {
		for _, v := range opts {
			v(r)
		}
	}
	return r
}

type CacheOption func(cache *Cache)

// WithName 设置缓存名称
func WithName(name string) CacheOption {
	return func(r *Cache) {
		r.name = name
	}
}

// WithTTL 设置redis缓存过期时间
func WithTTL(ttl time.Duration) CacheOption {
	return func(r *Cache) {
		r.ttl = ttl
	}
}

// WithBatchSize 设置RocksCache批量查询数量
func WithBatchSize(batchSize int) CacheOption {
	return func(r *Cache) {
		r.batchSize = batchSize
	}
}

// Key 生成带缓存名称前缀的缓存 key。
func (r *Cache) Key(keys ...any) string {
	keyStr := make([]string, 0)
	keyStr = append(keyStr, r.name)
	for _, v := range keys {
		keyStr = append(keyStr, dbcache.KeyFormat(v))
	}
	return strings.Join(keyStr, ":")
}

// TTL 返回带随机抖动的缓存过期时间。
func (r *Cache) TTL() time.Duration {
	return r.ttl - time.Duration(rand.Float64()*0.1*float64(r.ttl))
}

// Fetch 查询单个缓存值，未命中时回源加载。
func (r *Cache) Fetch(ctx context.Context, key string, fn func() (string, error), expire time.Duration) (string, error) {
	// 查询redis缓存
	rocksCacheValue, err := r.rocksCacheClient.Fetch2(ctx, key, expire, fn)
	if err != nil {
		return "", err
	}
	return rocksCacheValue, nil
}

// FetchBatch 批量查询缓存值，未命中时按批次回源加载。
func (r *Cache) FetchBatch(ctx context.Context, keys []string, fn func(miss []string) (map[string]string, error), expire time.Duration) (map[string]string, error) {
	resp := make(map[string]string)
	// 去重
	keys = unique(keys)
	// 查询redis缓存
	batch, err := chunk(keys, r.batchSize)
	if err != nil {
		return nil, err
	}
	// 使用`errgroup`并发查询
	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(100)
	// 创建一个channel用于接收每个goroutine的结果
	resultCh := make(chan map[string]string, len(batch))
	for k := range batch {
		i := k
		g.Go(func() error {
			rocksCacheResult, err := r.fetchBatchItem(ctx, batch[i], fn, expire)
			if err != nil {
				return err
			}
			// 将结果发送到channel
			resultCh <- rocksCacheResult
			return nil
		})
	}
	// 等待所有goroutine执行完毕
	err = g.Wait()
	if err != nil {
		return nil, err
	}
	// 关闭channel
	close(resultCh)
	// 从channel中读取结果
	for result := range resultCh {
		for k, v := range result {
			resp[k] = v
		}
	}
	return resp, nil
}

// fetchBatchItem 批量查询
func (r *Cache) fetchBatchItem(ctx context.Context, keys []string, fn func(miss []string) (map[string]string, error), expire time.Duration) (map[string]string, error) {
	resp := make(map[string]string)
	// 查询redis缓存
	rocksCacheResult, err := r.rocksCacheClient.FetchBatch2(ctx, keys, expire, func(idx []int) (map[int]string, error) {
		result := make(map[int]string)
		miss := make([]string, 0)
		for _, v := range idx {
			result[v] = ""
			miss = append(miss, keys[v])
		}
		dbValue, err := fn(miss)
		if err != nil {
			return nil, err
		}
		keyToInt := make(map[string]int)
		for k, v := range keys {
			keyToInt[v] = k
		}
		for k, v := range dbValue {
			result[keyToInt[k]] = v
		}
		return result, nil
	})
	if err != nil {
		return nil, err
	}
	for k, v := range rocksCacheResult {
		resp[keys[k]] = v
	}
	return resp, nil
}

// FetchHash 查询哈希字段缓存，未命中时回源加载。
func (r *Cache) FetchHash(ctx context.Context, key string, field string, fn func() (string, error), expire time.Duration) (string, error) {
	result, err := r.redisClient.HGet(ctx, key, field).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		return "", err
	}
	if result == "" && errors.Is(err, redis.Nil) {
		result, err = fn()
		if err != nil {
			return "", err
		}
		err = r.redisClient.HSet(ctx, key, field, result).Err()
		if err != nil {
			return "", err
		}
		err = r.redisClient.Expire(ctx, key, expire).Err()
		if err != nil {
			return "", err
		}
	}
	return result, nil
}

// Del 标记单个缓存 key 已删除。
func (r *Cache) Del(ctx context.Context, key string) error {
	err := r.rocksCacheClient.TagAsDeleted2(ctx, key)
	if err != nil {
		return err
	}
	return nil
}

// DelBatch 批量标记缓存 key 已删除。
func (r *Cache) DelBatch(ctx context.Context, keys []string) error {
	keys = unique(keys)
	batch, err := chunk(keys, r.batchSize)
	if err != nil {
		return err
	}
	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(100)
	for k := range batch {
		i := k
		g.Go(func() error {
			err := r.rocksCacheClient.TagAsDeletedBatch2(ctx, batch[i])
			if err != nil {
				return err
			}
			return nil
		})
	}
	err = g.Wait()
	if err != nil {
		return err
	}
	return nil
}

// unique 去重
func unique(slice []string) []string {
	if len(slice) == 0 {
		return slice
	}
	// 用 map 做 O(n) 去重，同时按原始遍历顺序写入结果，保持调用方顺序稳定。
	result := make([]string, 0, len(slice))
	seen := make(map[string]struct{}, len(slice))
	for _, v := range slice {
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		result = append(result, v)
	}
	return result
}

// chunk 将一个数组分成多个数组，每个数组包含size个元素，最后一个数组可能包含少于size个元素。
func chunk(collection []string, size int) ([][]string, error) {
	if size <= 0 {
		return nil, errors.New("chunk size must be greater than 0")
	}
	chunksNum := len(collection) / size
	if len(collection)%size != 0 {
		chunksNum += 1
	}
	result := make([][]string, 0, chunksNum)
	for i := 0; i < chunksNum; i++ {
		last := (i + 1) * size
		if last > len(collection) {
			last = len(collection)
		}
		result = append(result, collection[i*size:last])
	}
	return result, nil
}

// DelHash 删除哈希字段缓存。
func (r *Cache) DelHash(ctx context.Context, key string, field string) error {
	return r.redisClient.HDel(ctx, key, field).Err()
}
