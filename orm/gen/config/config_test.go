package config

import (
	"testing"

	"github.com/fzf-labs/godb/orm/encoding"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestNewRepoConfig(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	encoder := encoding.NewJSON()

	cfg := NewRepoConfig(db, nil, encoder)
	if cfg.DB != db {
		t.Fatal("db was not assigned")
	}
	if cfg.Cache != nil {
		t.Fatal("cache should be nil")
	}
	if cfg.Encoding != encoder {
		t.Fatal("encoding was not assigned")
	}
}
