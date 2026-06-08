package keymanage

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/fzf-labs/godb/orm/dbcache"
)

// KeyManage 管理服务内缓存 key 前缀及其过期时间说明。
type KeyManage struct {
	ServerName string
	List       map[string]KeyPrefix
}

// New 实例化key管理器
func New(serverName string) *KeyManage {
	return &KeyManage{
		ServerName: strings.TrimSpace(serverName),
		List:       make(map[string]KeyPrefix),
	}
}

// AddKey 添加一个key prefix
func (p *KeyManage) AddKey(prefix string, expirationTime time.Duration, remark string) (*KeyPrefix, error) {
	if p == nil {
		return nil, fmt.Errorf("key manager cannot be nil")
	}
	serverName := strings.TrimSpace(p.ServerName)
	if serverName == "" {
		return nil, fmt.Errorf("server name cannot be empty")
	}
	prefix = strings.TrimSpace(prefix)
	if prefix == "" {
		return nil, fmt.Errorf("prefix cannot be empty")
	}
	if _, ok := p.List[prefix]; ok {
		return nil, fmt.Errorf("key %s is exist, please change one", prefix)
	}
	key := KeyPrefix{
		ServerName:     serverName,
		PrefixName:     prefix,
		Remark:         remark,
		ExpirationTime: expirationTime,
	}
	p.List[prefix] = key
	return &key, nil
}

// Document 导出MD文档
func (p *KeyManage) Document() string {
	if p == nil {
		return ""
	}
	str := `|ServerName|PrefixName|ttl(s)|Remark` + "\n" + `|--|--|--|--|` + "\n"

	if len(p.List) > 0 {
		prefixes := make([]string, 0, len(p.List))
		for prefix := range p.List {
			prefixes = append(prefixes, prefix)
		}
		sort.Strings(prefixes)
		for _, prefix := range prefixes {
			m := p.List[prefix]
			str += `|` + escapeMarkdownTableCell(m.ServerName) + `|` + escapeMarkdownTableCell(m.PrefixName) + `|` + strconv.FormatFloat(m.ExpirationTime.Seconds(), 'f', -1, 64) + `|` + escapeMarkdownTableCell(m.Remark) + `|` + "\n"
		}
	}
	return str
}

// KeyPrefix key前缀
type KeyPrefix struct {
	ServerName     string
	PrefixName     string
	Remark         string
	ExpirationTime time.Duration
}

// Key 获取key
func (p *KeyPrefix) Key(keys ...string) string {
	if p == nil {
		return ""
	}
	parts := make([]any, 0, len(keys)+2)
	parts = append(parts, p.ServerName, p.PrefixName)
	for _, key := range keys {
		parts = append(parts, key)
	}
	return dbcache.BuildKey(parts...)
}

// Keys 获取keys
func (p *KeyPrefix) Keys(keys []string) []string {
	if p == nil {
		return []string{}
	}
	result := make([]string, 0)
	if len(keys) > 0 {
		for _, key := range keys {
			result = append(result, p.Key(key))
		}
	}
	return result
}

// TTL 获取key的过期时间time.Duration
func (p *KeyPrefix) TTL() time.Duration {
	if p == nil {
		return 0
	}
	return p.ExpirationTime
}

// TTLSecond 获取key的过期时间 Second
func (p *KeyPrefix) TTLSecond() int {
	if p == nil {
		return 0
	}
	return int(p.ExpirationTime / time.Second)
}

func escapeMarkdownTableCell(cell string) string {
	cell = strings.ReplaceAll(cell, `\`, `\\`)
	return strings.ReplaceAll(cell, `|`, `\|`)
}
