package config

import (
	"github.com/fzf-labs/godb/orm/dbcache"
	"github.com/fzf-labs/godb/orm/encoding"
	"gorm.io/gorm"
)

// NewRepoConfig 创建生成仓储使用的运行时依赖配置。
func NewRepoConfig(db *gorm.DB, cache dbcache.IDBCache, encode encoding.API) *Repo {
	return &Repo{
		DB:       db,
		Cache:    cache,
		Encoding: encode,
	}
}

// Repo 保存生成仓储依赖的数据库、缓存和编解码器。
type Repo struct {
	DB       *gorm.DB
	Cache    dbcache.IDBCache
	Encoding encoding.API
}
