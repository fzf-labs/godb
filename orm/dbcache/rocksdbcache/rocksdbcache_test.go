package rocksdbcache

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/dtm-labs/rockscache"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

var client = redis.NewClient(&redis.Options{
	Addr: "0.0.0.0:6379",
})

func NewWeakRocksCacheClient(rdb *redis.Client) *rockscache.Client {
	rc := rockscache.NewClient(rdb, rockscache.NewDefaultOptions())
	// 常用参数设置
	// 1、强一致性(默认关闭强一致性，如果开启的话会影响性能)
	rc.Options.StrongConsistency = false
	// 2、redis出现问题需要缓存降级时设置为true
	rc.Options.DisableCacheRead = false   // 关闭缓存读，默认false；如果打开，那么Fetch就不从缓存读取数据，而是直接调用fn获取数据
	rc.Options.DisableCacheDelete = false // 关闭缓存删除，默认false；如果打开，那么TagAsDeleted就什么操作都不做，直接返回
	// 3、其他设置
	// 标记删除的延迟时间，默认10秒，设置为3秒表示：被删除的key在3秒后才从redis中彻底清除
	rc.Options.Delay = time.Second * time.Duration(3)
	// 防穿透：若fn返回空字符串，空结果在缓存中的缓存时间，默认60秒
	rc.Options.EmptyExpire = time.Second * time.Duration(120)
	// 防雪崩,默认0.1,当前设置为0.1的话，如果设定为600的过期时间，那么过期时间会被设定为540s - 600s中间的一个随机数，避免数据出现同时到期
	rc.Options.RandomExpireAdjustment = 0.1 // 设置为默认就行
	return rc
}

func TestRocksCache_Fetch(t *testing.T) {
	rocksCacheClient := NewWeakRocksCacheClient(client)
	cache := NewRocksDBCache(client, rocksCacheClient, WithName("test"), WithTTL(time.Minute), WithBatchSize(100))
	ctx := context.Background()
	fetch, err := cache.Fetch(ctx, "RocksCache_Fetch", func() (string, error) {
		fmt.Println(1)
		return "RocksCache_Fetch:result", nil
	}, cache.TTL())
	fmt.Println(fetch)
	fmt.Println(err)
	assert.Equal(t, nil, err)
}

func TestRocksCache_FetchBatch(t *testing.T) {
	rocksCacheClient := NewWeakRocksCacheClient(client)
	cache := NewRocksDBCache(client, rocksCacheClient, WithName("test"), WithTTL(time.Minute), WithBatchSize(100))
	ctx := context.Background()
	keys := []string{
		"RocksCache_FetchBatch_a",
		"RocksCache_FetchBatch_b",
		"RocksCache_FetchBatch_c",
	}
	take, err := cache.FetchBatch(ctx, keys, func(miss []string) (map[string]string, error) {
		resp := make(map[string]string)
		for _, v := range miss {
			resp[v] = v + ":result"
		}
		return resp, nil
	}, cache.TTL())
	fmt.Println(take)
	fmt.Println(err)
	assert.Equal(t, nil, err)
}

func TestCache_Del(t *testing.T) {
	rocksCacheClient := NewWeakRocksCacheClient(client)
	cache := NewRocksDBCache(client, rocksCacheClient, WithName("test"), WithTTL(time.Minute), WithBatchSize(100))
	ctx := context.Background()
	err := cache.Del(ctx, "RocksCache_Fetch")
	fmt.Println(err)
	assert.Equal(t, nil, err)
}

func TestCache_DelBatch(t *testing.T) {
	rocksCacheClient := NewWeakRocksCacheClient(client)
	cache := NewRocksDBCache(client, rocksCacheClient, WithName("test"), WithTTL(time.Minute), WithBatchSize(100))
	ctx := context.Background()
	keys := []string{
		"RocksCache_FetchBatch_a",
		"RocksCache_FetchBatch_b",
		"RocksCache_FetchBatch_c",
	}
	err := cache.DelBatch(ctx, keys)
	if err != nil {
		return
	}
	assert.Equal(t, nil, err)
}

func TestCache_Key(t *testing.T) {
	rocksCacheClient := NewWeakRocksCacheClient(client)
	cache := NewRocksDBCache(client, rocksCacheClient, WithName("test"), WithTTL(time.Minute), WithBatchSize(100))
	key := cache.Key("a", "b", "c")
	assert.Equal(t, key, "test:a:b:c")
}
