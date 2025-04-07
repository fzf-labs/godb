package utils

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"time"

	"golang.org/x/tools/go/packages"
)

// FillModelPkgPath 返回模型文件的包路径
func FillModelPkgPath(dir string) string {
	pkg, err := packages.Load(&packages.Config{
		Mode: packages.NeedName,
		Dir:  dir,
	})
	if err != nil {
		return ""
	}
	if len(pkg) > 0 {
		return pkg[0].PkgPath
	}
	return ""
}

// StrSliFind 判断字符串切片中是否存在某个元素
func StrSliFind(collection []string, element string) bool {
	for _, s := range collection {
		if s == element {
			return true
		}
	}
	return false
}

// SliRemove 删除字符串切片中的某个元素
func SliRemove(collection, element []string) []string {
	for _, s := range element {
		for i, v := range collection {
			if s == v {
				collection = append(collection[:i], collection[i+1:]...)
			}
		}
	}
	return collection
}

// ConvToString 任意类型转字符串
func ConvToString(any any) string {
	if any == nil {
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
		return value.String()
	case *time.Time:
		if value == nil {
			return ""
		}
		return value.String()
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
			return ConvToString(rv.Elem().Interface())
		}
		// Finally, we use json.Marshal to convert.
		if jsonContent, err := json.Marshal(value); err != nil {
			return fmt.Sprint(value)
		} else {
			return string(jsonContent)
		}
	}
}
