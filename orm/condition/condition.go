package condition

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"sync"

	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"
)

type Exp string // 操作

// 查询表达式操作符。
const (
	RAW       Exp = "RAW" // 原始表达式
	EQ        Exp = "="
	NEQ       Exp = "!="
	GT        Exp = ">"
	GTE       Exp = ">="
	LT        Exp = "<"
	LTE       Exp = "<="
	IN        Exp = "IN"
	NOTIN     Exp = "NOT IN"
	LIKE      Exp = "LIKE"
	NOTLIKE   Exp = "NOT LIKE"
	ISNULL    Exp = "IS NULL"
	ISNOTNULL Exp = "IS NOT NULL"
)

type Logic string // 逻辑关系

// 查询条件逻辑连接符。
const (
	AND Logic = "AND" // 逻辑关系 and
	OR  Logic = "OR"  // 逻辑关系 or
)

type Order string // 排序

// 排序方向。
const (
	ASC  Order = "ASC"  // 升序
	DESC Order = "DESC" // 降序
)

// QueryParam 查询条件
type QueryParam struct {
	Field string      `json:"field"` // 字段
	Value interface{} `json:"value"` // 值（当Exp为IN, NOTIN 时为[]interface{}）
	Exp   Exp         `json:"exp"`   // 操作 "=", "!=", ">", ">=", "<", "<=", "IN", "NOT IN", "LIKE", "NOT LIKE"
	Logic Logic       `json:"logic"` // 逻辑关系 AND OR
}

// OrderParam 排序条件
type OrderParam struct {
	Field string `json:"field"` // 字段
	Order Order  `json:"order"` // 排序 ASC DESC
}

// Req 请求-自定义查询
type Req struct {
	Page     int32         `json:"page"`     // 页码
	PageSize int32         `json:"pageSize"` // 页数
	Query    []*QueryParam `json:"query"`    // 查询条件
	Order    []*OrderParam `json:"order"`    // 排序条件
}

// Reply 返回-自定义查询
type Reply struct {
	Page      int32 `json:"page"`      // 第几页
	PageSize  int32 `json:"pageSize"`  // 页大小
	Total     int32 `json:"total"`     // 总数
	PrevPage  int32 `json:"prevPage"`  // 上一页
	NextPage  int32 `json:"nextPage"`  // 下一页
	TotalPage int32 `json:"totalPage"` // 总页数
}

// ExpValidate 验证Exp是否合法
func ExpValidate(s Exp) bool {
	s = normalizeExp(s)
	switch s {
	case EQ, NEQ, GT, GTE, LT, LTE, IN, NOTIN, LIKE, NOTLIKE, ISNULL, ISNOTNULL, RAW:
		return true
	default:
		return false
	}
}

// LogicValidate 验证Logic是否合法
func LogicValidate(s Logic) bool {
	s = normalizeLogic(s)
	switch s {
	case AND, OR:
		return true
	default:
		return false
	}
}

// OrderValidate 验证Order是否合法
func OrderValidate(s Order) bool {
	s = normalizeOrder(s)
	switch s {
	case ASC, DESC:
		return true
	default:
		return false
	}
}

// ToInterfaceSlice 将任意类型的切片转换为 []interface{}
func (p *Req) ToInterfaceSlice(val interface{}) ([]interface{}, error) {
	if val == nil {
		return nil, fmt.Errorf("value is not a slice")
	}
	if values, ok := val.([]interface{}); ok {
		return values, nil
	}
	rv := reflect.ValueOf(val)
	for rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			return nil, fmt.Errorf("value is not a slice")
		}
		rv = rv.Elem()
	}
	// 如果不是切片类型，返回错误
	if rv.Kind() != reflect.Slice && rv.Kind() != reflect.Array {
		return nil, fmt.Errorf("value is not a slice")
	}
	// 遍历切片中的每个元素，将其转换为 interface{} 类型
	values := make([]interface{}, rv.Len())
	for i := 0; i < rv.Len(); i++ {
		values[i] = rv.Index(i).Interface()
	}
	return values, nil
}

// ConvertToCacheField 将 req 转换为缓存 hash 中的 field
func (p *Req) ConvertToCacheField() string {
	marshal, err := json.Marshal(p.canonicalCachePayload())
	if err != nil {
		return ""
	}
	// 使用 sha256 加密
	hash := sha256.New()
	hash.Write(marshal)
	return fmt.Sprintf("%x", hash.Sum(nil))
}

func (p *Req) canonicalCachePayload() any {
	if p == nil {
		return nil
	}
	query := make([]*QueryParam, len(p.Query))
	for i, item := range p.Query {
		if item == nil {
			continue
		}
		exp := normalizeExp(item.Exp)
		if exp == "" {
			exp = EQ
		}
		logic := normalizeLogic(item.Logic)
		if logic == "" {
			logic = AND
		}
		query[i] = &QueryParam{
			Field: strings.TrimSpace(item.Field),
			Value: item.Value,
			Exp:   exp,
			Logic: logic,
		}
	}
	order := make([]*OrderParam, len(p.Order))
	for i, item := range p.Order {
		if item == nil {
			continue
		}
		direction := normalizeOrder(item.Order)
		if direction == "" {
			direction = ASC
		}
		order[i] = &OrderParam{
			Field: strings.TrimSpace(item.Field),
			Order: direction,
		}
	}
	return &Req{
		Page:     p.Page,
		PageSize: p.PageSize,
		Query:    query,
		Order:    order,
	}
}

// ConvertToGormExpression 根据SearchColumn参数转换为符合gorm where clause.Expression
func (p *Req) ConvertToGormExpression(model interface{}) (whereExpressions, orderExpressions []clause.Expression, err error) {
	whereExpressions = make([]clause.Expression, 0)
	orderExpressions = make([]clause.Expression, 0)
	if p == nil {
		return whereExpressions, orderExpressions, nil
	}
	column := fieldToColumn(model)
	if len(p.Query) > 0 {
		for _, v := range p.Query {
			if v == nil {
				return whereExpressions, orderExpressions, fmt.Errorf("query cannot be nil")
			}
			queryField := strings.TrimSpace(v.Field)
			if queryField == "" {
				return whereExpressions, orderExpressions, fmt.Errorf("field cannot be empty")
			}
			field, ok := column[strings.ToLower(queryField)]
			if !ok {
				return whereExpressions, orderExpressions, fmt.Errorf("field '%s' is not db column name", v.Field)
			}
			exp := normalizeExp(v.Exp)
			if exp == "" {
				exp = EQ
			}
			if !ExpValidate(exp) {
				return whereExpressions, orderExpressions, fmt.Errorf("unknown exp type '%s'", exp)
			}
			logic := normalizeLogic(v.Logic)
			if logic == "" {
				logic = AND
			}
			if !LogicValidate(logic) {
				return whereExpressions, orderExpressions, fmt.Errorf("unknown logic type '%s'", logic)
			}
			var expression clause.Expression
			switch exp {
			case EQ:
				expression = clause.Eq{Column: field, Value: v.Value}
			case NEQ:
				expression = clause.Neq{Column: field, Value: v.Value}
			case GT:
				expression = clause.Gt{Column: field, Value: v.Value}
			case GTE:
				expression = clause.Gte{Column: field, Value: v.Value}
			case LT:
				expression = clause.Lt{Column: field, Value: v.Value}
			case LTE:
				expression = clause.Lte{Column: field, Value: v.Value}
			case IN:
				values, err := p.ToInterfaceSlice(v.Value)
				if err != nil {
					return nil, nil, err
				}
				expression = clause.IN{Column: field, Values: values}
			case NOTIN:
				values, err := p.ToInterfaceSlice(v.Value)
				if err != nil {
					return nil, nil, err
				}
				expression = clause.Not(clause.IN{Column: field, Values: values})
			case LIKE:
				expression = clause.Like{Column: field, Value: v.Value}
			case NOTLIKE:
				expression = clause.Not(clause.Like{Column: field, Value: v.Value})
			case ISNULL:
				expression = clause.Eq{Column: field, Value: nil}
			case ISNOTNULL:
				expression = clause.Neq{Column: field, Value: nil}
			case RAW:
				expression, ok = v.Value.(clause.Expr)
				if !ok {
					return nil, nil, fmt.Errorf("RAW value is not a clause.Expr")
				}
			}
			if logic == AND {
				whereExpressions = append(whereExpressions, clause.And(expression))
			} else {
				whereExpressions = append(whereExpressions, clause.Or(expression))
			}
		}
	}
	if len(p.Order) > 0 {
		for _, v := range p.Order {
			if v == nil {
				return whereExpressions, orderExpressions, fmt.Errorf("order cannot be nil")
			}
			orderField := strings.TrimSpace(v.Field)
			if orderField == "" {
				return whereExpressions, orderExpressions, fmt.Errorf("field cannot be empty")
			}
			field, ok := column[strings.ToLower(orderField)]
			if !ok {
				return whereExpressions, orderExpressions, fmt.Errorf("field '%s' is not db column name", v.Field)
			}
			order := normalizeOrder(v.Order)
			if order == "" {
				order = ASC
			}
			if !OrderValidate(order) {
				return whereExpressions, orderExpressions, fmt.Errorf("order is err")
			}
			orderExpressions = append(orderExpressions, clause.OrderBy{
				Columns: []clause.OrderByColumn{
					{
						Column:  clause.Column{Name: field},
						Desc:    order == DESC,
						Reorder: false,
					},
				},
			})
		}
	}
	return whereExpressions, orderExpressions, nil
}

// ConvertToPage 转换为page
func (p *Req) ConvertToPage(total int32) (*Reply, error) {
	if total < 0 {
		return &Reply{}, fmt.Errorf("total cannot be less than 0")
	}
	resp := &Reply{
		Page:      0,
		PageSize:  0,
		Total:     total,
		PrevPage:  0,
		NextPage:  0,
		TotalPage: 0,
	}
	if p == nil {
		return resp, nil
	}
	if p.Page < 0 {
		return resp, fmt.Errorf("page cannot be less than 0")
	}
	if p.PageSize < 0 {
		return resp, fmt.Errorf("pageSize cannot be less than 0")
	}
	if (p.Page != 0 && p.PageSize == 0) || (p.Page == 0 && p.PageSize != 0) {
		return resp, fmt.Errorf("page and pageSize must be a pair")
	}
	if p.Page == 0 && p.PageSize == 0 {
		return resp, nil
	}
	resp.Page = p.Page
	resp.PageSize = p.PageSize
	resp.TotalPage = int32((int64(total) + int64(p.PageSize) - 1) / int64(p.PageSize))
	if resp.TotalPage == 0 {
		return resp, nil
	}
	if p.Page > resp.TotalPage {
		resp.NextPage = resp.TotalPage
		resp.PrevPage = resp.TotalPage
		return resp, nil
	}
	currentPage := p.Page
	if currentPage >= resp.TotalPage {
		resp.NextPage = resp.TotalPage
	} else {
		resp.NextPage = currentPage + 1
	}
	resp.PrevPage = currentPage - 1
	if resp.PrevPage <= 0 {
		resp.PrevPage = 1
	}
	return resp, nil
}

// fieldToColumn 将model的tag中gorm的tag的Column转换为map[string]string
func fieldToColumn(model interface{}) map[string]string {
	m := make(map[string]string)
	if model == nil {
		return m
	}
	if parsed, err := schema.Parse(model, &sync.Map{}, schema.NamingStrategy{}); err == nil {
		for _, field := range parsed.Fields {
			column := strings.TrimSpace(field.DBName)
			if column != "" {
				m[strings.ToLower(column)] = column
			}
		}
	}
	t := reflect.TypeOf(model)
	if t == nil {
		return m
	}
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return m
	}
	collectFieldColumns(t, m)
	return m
}

func collectFieldColumns(t reflect.Type, columns map[string]string) {
	namer := schema.NamingStrategy{}
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fieldType := field.Type
		for fieldType.Kind() == reflect.Ptr {
			fieldType = fieldType.Elem()
		}
		if field.Anonymous && fieldType.Kind() == reflect.Struct {
			collectFieldColumns(fieldType, columns)
			continue
		}
		if field.PkgPath != "" {
			continue
		}
		gormTag := field.Tag.Get("gorm")
		if gormTag == "-" {
			continue
		}
		columnName := ""
		if gormTag != "" {
			gormTags := strings.Split(gormTag, ";")
			for _, v := range gormTags {
				tagPart := strings.TrimSpace(v)
				if strings.HasPrefix(strings.ToLower(tagPart), "column:") {
					column := strings.SplitN(tagPart, ":", 2)
					if len(column) == 2 {
						columnName = strings.TrimSpace(column[1])
					}
					break
				}
			}
		}
		if columnName == "" {
			if !isDefaultColumnCandidate(fieldType) {
				continue
			}
			columnName = namer.ColumnName("", field.Name)
		}
		columns[strings.ToLower(columnName)] = columnName
	}
}

func isDefaultColumnCandidate(t reflect.Type) bool {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	switch t.Kind() {
	case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64, reflect.String:
		return true
	case reflect.Slice, reflect.Array:
		return t.Elem().Kind() == reflect.Uint8
	default:
		return false
	}
}

func normalizeExp(exp Exp) Exp {
	return Exp(strings.ToUpper(strings.Join(strings.Fields(string(exp)), " ")))
}

func normalizeLogic(logic Logic) Logic {
	return Logic(strings.ToUpper(strings.TrimSpace(string(logic))))
}

func normalizeOrder(order Order) Order {
	return Order(strings.ToUpper(strings.TrimSpace(string(order))))
}
