package localcache

import (
	"github.com/dgraph-io/ristretto/v2"
)

// NewRistrettoStringCache 创建一个字符串键值对的 Ristretto 缓存实例
// 默认配置:
// - NumCounters: 10M (用于跟踪键的频率计数器数量)
// - MaxCost: 1GB (缓存的最大内存占用)
// - BufferItems: 64 (Get 操作的缓冲区大小)
func NewRistrettoStringCache() (*ristretto.Cache[string, string], error) {
	return newRistrettoCache[string]()
}

// NewRistrettoAnyCache 创建一个支持任意值类型的 Ristretto 缓存实例
// 默认配置与 NewRistrettoStringCache 相同
func NewRistrettoAnyCache() (*ristretto.Cache[string, any], error) {
	return newRistrettoCache[any]()
}

// newRistrettoCache 创建一个通用的 Ristretto 缓存实例
// T 为缓存值的类型参数
func newRistrettoCache[T any]() (*ristretto.Cache[string, T], error) {
	return ristretto.NewCache(&ristretto.Config[string, T]{
		NumCounters: 1e7,     // 跟踪键频率的计数器数量 (10M)
		MaxCost:     1 << 30, // 缓存的最大内存占用 (1GB)
		BufferItems: 64,      // Get 操作的缓冲区大小
	})
}
