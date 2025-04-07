package config

import (
	"github.com/fzf-labs/godb/orm/dbcache"
	"github.com/fzf-labs/godb/orm/encoding"
	"gorm.io/gorm"
)

func NewRepoConfig(db *gorm.DB, cache dbcache.IDBCache, encode encoding.API) *Repo {
	return &Repo{
		DB:       db,
		Cache:    cache,
		Encoding: encode,
	}
}

type Repo struct {
	DB       *gorm.DB
	Cache    dbcache.IDBCache
	Encoding encoding.API
}
