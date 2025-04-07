package rueidiscache

import (
	"github.com/redis/rueidis"
	"github.com/redis/rueidis/rueidisaside"
	"github.com/redis/rueidis/rueidisotel"
)

// NewRueidisClient  redis客户端rueidis
// redis > 6.0
func NewRueidisClient(clientOption *rueidis.ClientOption) (rueidis.Client, error) {
	client, err := rueidisotel.NewClient(*clientOption)
	if err != nil {
		return nil, err
	}
	return client, nil
}

// NewRueidisAsideClient 增强的 Cache-Aside 模式客户端
// redis > 7.0
func NewRueidisAsideClient(clientOption *rueidis.ClientOption) (rueidisaside.CacheAsideClient, error) {
	return rueidisaside.NewClient(rueidisaside.ClientOption{
		ClientOption: *clientOption,
	})
}
