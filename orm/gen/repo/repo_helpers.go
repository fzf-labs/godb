package repo

import (
	"fmt"
	"go/token"
	"strings"
	"unicode"

	"github.com/fzf-labs/godb/orm/utils/strutil"
	"github.com/jinzhu/inflection"
)

// KeyWords 定义生成 repo 代码时需要避让的局部变量名。
var KeyWords = []string{
	"dao",
	"parameters",
	"cacheKey",
	"cacheKeys",
	"cacheValue",
	"keyToParam",
	"resp",
	"result",
	"marshal",
	"item",
}

func (r *Repo) upperFields(columns []string) string {
	var upperFields string
	for _, v := range columns {
		upperFields += r.upperFieldName(v)
	}
	return upperFields
}

func (r *Repo) fieldAndDataTypes(columns []string) string {
	var fieldAndDataTypes string
	for _, v := range columns {
		fieldAndDataTypes += fmt.Sprintf("%s %s,", r.lowerFieldName(v), r.columnNameToDataType[v])
	}
	return strings.Trim(fieldAndDataTypes, ",")
}

func (r *Repo) cacheFields(columns []string) string {
	var cacheFields string
	for _, v := range columns {
		cacheFields += r.upperFieldName(v)
	}
	return cacheFields
}

func (r *Repo) cacheFieldsJoin(columns []string) string {
	var cacheFieldsJoin string
	for _, v := range columns {
		cacheFieldsJoin += fmt.Sprintf("%s,", r.lowerFieldName(v))
	}
	return strings.Trim(cacheFieldsJoin, ",")
}

// upperFieldName 字段名称大写
func (r *Repo) upperFieldName(s string) string {
	return r.columnNameToName[s]
}

// lowerFieldName 字段名称小写
func (r *Repo) lowerFieldName(s string) string {
	str := r.upperFieldName(s)
	if str == "" {
		return str
	}
	words := []string{"API", "ASCII", "CPU", "CSS", "DNS", "EOF", "GUID", "HTML", "HTTP", "HTTPS", "ID", "IP", "JSON", "LHS", "QPS", "RAM", "RHS", "RPC", "SLA", "SMTP", "SSH", "TLS", "ttl", "UID", "UI", "UUID", "URI", "URL", "UTF8", "VM", "XML", "XSRF", "XSS"}
	// 如果第一个单词命中  则不处理
	for _, v := range words {
		if strings.HasPrefix(str, v) {
			return str
		}
	}
	rs := []rune(str)
	f := rs[0]
	if 'A' <= f && f <= 'Z' {
		str = string(unicode.ToLower(f)) + string(rs[1:])
	}
	if token.Lookup(str).IsKeyword() || strutil.StrSliFind(KeyWords, str) {
		str = "_" + str
	}
	return str
}

// upperName 大写
func (r *Repo) upperName(s string) string {
	return r.gorm.NamingStrategy.SchemaName(s)
}

// lowerName 小写
func (r *Repo) lowerName(s string) string {
	str := r.upperName(s)
	if str == "" {
		return str
	}
	words := []string{"API", "ASCII", "CPU", "CSS", "DNS", "EOF", "GUID", "HTML", "HTTP", "HTTPS", "ID", "IP", "JSON", "LHS", "QPS", "RAM", "RHS", "RPC", "SLA", "SMTP", "SSH", "TLS", "ttl", "UID", "UI", "UUID", "URI", "URL", "UTF8", "VM", "XML", "XSRF", "XSS"}
	// 如果第一个单词命中  则不处理
	for _, v := range words {
		if strings.HasPrefix(str, v) {
			return str
		}
	}
	rs := []rune(str)
	f := rs[0]
	if 'A' <= f && f <= 'Z' {
		str = string(unicode.ToLower(f)) + string(rs[1:])
	}
	return str
}

// plural 复数形式
func (r *Repo) plural(s string) string {
	if s == "" {
		return s
	}
	str := inflection.Plural(s)
	if str == s {
		str += "plural"
	}
	return str
}

// checkDaoFieldType  检查字段是否是 dao 中的 Field类型
func (r *Repo) checkDaoFieldType(s []string) bool {
	for _, v := range s {
		if r.columnNameToFieldType[v] == "Field" {
			return true
		}
	}
	return false
}

func (r *Repo) whereFields(columns []string) string {
	var whereFields string
	for _, v := range columns {
		switch r.columnNameToDataType[v] {
		case "bool":
			whereFields += fmt.Sprintf("dao.%s.Is(%s),", r.upperFieldName(v), r.lowerFieldName(v))
		default:
			whereFields += fmt.Sprintf("dao.%s.Eq(%s),", r.upperFieldName(v), r.lowerFieldName(v))
		}
	}
	return strings.TrimRight(whereFields, ",")
}

func (r *Repo) primaryKeyWhereFields(columns []string) string {
	var whereFields string
	for _, v := range columns {
		switch r.columnNameToDataType[v] {
		case "bool":
			whereFields += fmt.Sprintf("dao.%s.Is(%s),", r.upperFieldName(v), "data."+r.upperFieldName(v))
		default:
			whereFields += fmt.Sprintf("dao.%s.Eq(%s),", r.upperFieldName(v), "data."+r.upperFieldName(v))
		}
	}
	return strings.TrimRight(whereFields, ",")
}

// hasDeletedAt 是否有删除标记
func hasDeletedAt(columnNameToDataType map[string]string) bool {
	for _, v := range columnNameToDataType {
		if v == "gorm.DeletedAt" || v == "soft_delete.DeletedAt" {
			return true
		}
	}
	return false
}
