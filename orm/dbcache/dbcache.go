package dbcache

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// IDBCache 定义数据库查询缓存需要实现的 key、读取和失效接口。
type IDBCache interface {
	// Key 返回给定字段的字符串键
	Key(fields ...any) string
	// TTL 获取默认缓存过期时间
	TTL() time.Duration
	// Fetch 从缓存中获取值，如果没有找到，调用fn来获取数据并设置为缓存
	Fetch(ctx context.Context, key string, fn func() (string, error), expire time.Duration) (string, error)
	// FetchBatch 批量从缓存中获取值，如果没有找到，调用fn来获取数据并设置为缓存,并设置过期时间
	FetchBatch(ctx context.Context, keys []string, fn func(miss []string) (map[string]string, error), expire time.Duration) (map[string]string, error)
	// FetchHash 按照key和field从缓存中获取值，如果没有找到，调用fn来获取数据并设置为缓存,并设置过期时间
	FetchHash(ctx context.Context, key string, field string, fn func() (string, error), expire time.Duration) (string, error)
	// Del 删除缓存
	Del(ctx context.Context, key string) error
	// DelBatch 批量删除缓存
	DelBatch(ctx context.Context, keys []string) error
	// DelHash 删除缓存
	DelHash(ctx context.Context, key string, field string) error
}

var keyPartEscaper = strings.NewReplacer(`\`, `\\`, `:`, `\:`)

// EscapeKeyPart 对单个 key 分段做转义，避免分隔符碰撞。
func EscapeKeyPart(part string) string {
	return keyPartEscaper.Replace(part)
}

// BuildKey 将多个 key 分段转换并拼接成最终 key。
func BuildKey(parts ...any) string {
	formatted := make([]string, 0, len(parts))
	for _, part := range parts {
		formatted = append(formatted, KeyFormat(part))
	}
	return strings.Join(formatted, ":")
}

func isNilLikeValue(any any) bool {
	if any == nil {
		return true
	}
	rv := reflect.ValueOf(any)
	if !rv.IsValid() {
		return true
	}
	switch rv.Kind() {
	case reflect.Chan, reflect.Func, reflect.Map, reflect.Ptr, reflect.Slice, reflect.UnsafePointer:
		return rv.IsNil()
	default:
		return false
	}
}

// KeyFormat 将任意类型转换为字符串并转义为安全的 key 分段。
func KeyFormat(any any) string {
	return EscapeKeyPart(keyFormatRaw(any))
}

func keyFormatRaw(any any) string {
	if isNilLikeValue(any) {
		return ""
	}
	switch value := any.(type) {
	case int:
		return strconv.Itoa(value)
	case int8:
		return strconv.Itoa(int(value))
	case int16:
		return strconv.Itoa(int(value))
	case int32:
		return strconv.Itoa(int(value))
	case int64:
		return strconv.FormatInt(value, 10)
	case uint:
		return strconv.FormatUint(uint64(value), 10)
	case uint8:
		return strconv.FormatUint(uint64(value), 10)
	case uint16:
		return strconv.FormatUint(uint64(value), 10)
	case uint32:
		return strconv.FormatUint(uint64(value), 10)
	case uint64:
		return strconv.FormatUint(value, 10)
	case float32:
		return strconv.FormatFloat(float64(value), 'f', -1, 32)
	case float64:
		return strconv.FormatFloat(value, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(value)
	case string:
		return value
	case []byte:
		return string(value)
	case time.Time:
		if value.IsZero() {
			return ""
		}
		return value.Format("2006-01-02 15:04:05") // 转换为字符串
	case *time.Time:
		if value == nil || value.IsZero() {
			return ""
		}
		return value.Format("2006-01-02 15:04:05") // 转换为字符串
	default:
		// Empty checks.
		if value == nil {
			return ""
		}
		// Reflect checks.
		var (
			rv   = reflect.ValueOf(value)
			kind = rv.Kind()
		)
		switch kind {
		case reflect.Chan,
			reflect.Map,
			reflect.Slice,
			reflect.Func,
			reflect.Ptr,
			reflect.Interface,
			reflect.UnsafePointer:
			if rv.IsNil() {
				return ""
			}
		case reflect.String:
			return rv.String()
		}
		if kind == reflect.Ptr {
			return keyFormatRaw(rv.Elem().Interface())
		}
		// Finally, we use json.Marshal to convert.
		if jsonContent, err := json.Marshal(value); err != nil {
			return fmt.Sprint(value)
		} else {
			return string(jsonContent)
		}
	}
}
