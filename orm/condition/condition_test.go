package condition

import (
	"reflect"
	"testing"

	"gorm.io/gorm/clause"
)

// UserTest 是动态条件测试使用的示例模型。
type UserTest struct {
	ID       string `gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid();comment:ID" json:"id"`    // ID
	UID      string `gorm:"column:uid;type:character varying(64);not null;comment:uid" json:"uid"`            // uid
	Username string `gorm:"column:username;type:character varying(30);not null;comment:用户账号" json:"username"` // 用户账号
	Password string `gorm:"column:password;type:character varying(100);not null;comment:密码" json:"password"`  // 密码
	Nickname string `gorm:"column:nickname;type:character varying(30);not null;comment:用户昵称" json:"nickname"` // 用户昵称
}

func TestPaginatorReq_ConvertToGormExpression(t *testing.T) {
	type fields struct {
		Page     int32
		PageSize int32
		Search   []*QueryParam
		Order    []*OrderParam
	}
	type args struct {
		model any
	}
	tests := []struct {
		name                 string
		fields               fields
		args                 args
		wantWhereExpressions []clause.Expression
		wantOrderExpressions []clause.Expression
		wantErr              bool
	}{
		{
			name: "test1",
			fields: fields{
				Page:     0,
				PageSize: 0,
				Search:   nil,
				Order:    nil,
			},
			args: args{
				model: UserTest{},
			},
			wantWhereExpressions: []clause.Expression{},
			wantOrderExpressions: []clause.Expression{},
			wantErr:              false,
		},
		{
			name: "test2",
			fields: fields{
				Page:     1,
				PageSize: 20,
				Search: []*QueryParam{
					{
						Field: "id",
						Value: "admin",
						Exp:   EQ,
						Logic: AND,
					},
					{
						Field: "id",
						Value: "admin",
						Exp:   NEQ,
						Logic: AND,
					},
					{
						Field: "id",
						Value: "admin",
						Exp:   GT,
						Logic: AND,
					},
					{
						Field: "id",
						Value: "admin",
						Exp:   GTE,
						Logic: AND,
					},
					{
						Field: "id",
						Value: "admin",
						Exp:   LT,
						Logic: AND,
					},
					{
						Field: "id",
						Value: "admin",
						Exp:   LTE,
						Logic: AND,
					},
					{
						Field: "id",
						Value: []any{"admin"},
						Exp:   IN,
						Logic: AND,
					},
					{
						Field: "id",
						Value: []any{"admin"},
						Exp:   NOTIN,
						Logic: AND,
					},
					{
						Field: "id",
						Value: "admin",
						Exp:   LIKE,
						Logic: AND,
					},
					{
						Field: "id",
						Value: "admin",
						Exp:   NOTLIKE,
						Logic: AND,
					},
					{
						Field: "id",
						Value: "admin",
						Exp:   EQ,
						Logic: OR,
					},
					{
						Field: "id",
						Value: "admin",
						Exp:   "",
						Logic: "",
					},
				},
				Order: nil,
			},
			args: args{
				model: UserTest{},
			},
			wantWhereExpressions: []clause.Expression{
				clause.And(clause.Eq{Column: "id", Value: "admin"}),
				clause.And(clause.Neq{Column: "id", Value: "admin"}),
				clause.And(clause.Gt{Column: "id", Value: "admin"}),
				clause.And(clause.Gte{Column: "id", Value: "admin"}),
				clause.And(clause.Lt{Column: "id", Value: "admin"}),
				clause.And(clause.Lte{Column: "id", Value: "admin"}),
				clause.And(clause.IN{Column: "id", Values: []any{"admin"}}),
				clause.And(clause.Not(clause.IN{Column: "id", Values: []any{"admin"}})),
				clause.And(clause.Like{Column: "id", Value: "admin"}),
				clause.And(clause.Not(clause.Like{Column: "id", Value: "admin"})),
				clause.Or(clause.Eq{Column: "id", Value: "admin"}),
				clause.And(clause.Eq{Column: "id", Value: "admin"}),
			},
			wantOrderExpressions: []clause.Expression{},
			wantErr:              false,
		},
		{
			name: "test-err-1",
			fields: fields{
				Page:     0,
				PageSize: 0,
				Search: []*QueryParam{
					{
						Field: "",
						Value: "admin",
						Exp:   EQ,
						Logic: AND,
					},
				},
				Order: nil,
			},
			args: args{
				model: UserTest{},
			},
			wantWhereExpressions: []clause.Expression{},
			wantOrderExpressions: []clause.Expression{},
			wantErr:              true,
		},
		{
			name: "test-err-2",
			fields: fields{
				Page:     0,
				PageSize: 0,
				Search: []*QueryParam{
					{
						Field: "id",
						Value: "admin",
						Exp:   "a",
						Logic: AND,
					},
				},
				Order: nil,
			},
			args: args{
				model: UserTest{},
			},
			wantWhereExpressions: []clause.Expression{},
			wantOrderExpressions: []clause.Expression{},
			wantErr:              true,
		},
		{
			name: "test-err-3",
			fields: fields{
				Page:     0,
				PageSize: 0,
				Search: []*QueryParam{
					{
						Field: "id",
						Value: "admin",
						Exp:   EQ,
						Logic: "b",
					},
				},
				Order: nil,
			},
			args: args{
				model: UserTest{},
			},
			wantWhereExpressions: []clause.Expression{},
			wantOrderExpressions: []clause.Expression{},
			wantErr:              true,
		},
		{
			name: "test4",
			fields: fields{
				Page:     0,
				PageSize: 0,
				Search: []*QueryParam{
					{
						Field: "pid",
						Value: "admin",
						Exp:   EQ,
						Logic: AND,
					},
				},
				Order: nil,
			},
			args: args{
				model: UserTest{},
			},
			wantWhereExpressions: []clause.Expression{},
			wantOrderExpressions: []clause.Expression{},
			wantErr:              true,
		},
		{
			name: "test-order",
			fields: fields{
				Page:     0,
				PageSize: 0,
				Search:   nil,
				Order: []*OrderParam{
					{
						Field: "id",
						Order: "",
					},
				},
			},
			args: args{
				model: UserTest{},
			},
			wantWhereExpressions: []clause.Expression{},
			wantOrderExpressions: []clause.Expression{clause.OrderBy{
				Columns: []clause.OrderByColumn{
					{
						Column:  clause.Column{Name: "id"},
						Desc:    false,
						Reorder: false,
					},
				},
			}},
			wantErr: false,
		},
		{
			name: "test-order-err1",
			fields: fields{
				Page:     0,
				PageSize: 0,
				Search:   nil,
				Order: []*OrderParam{
					{
						Field: "",
						Order: "",
					},
				},
			},
			args: args{
				model: UserTest{},
			},
			wantWhereExpressions: []clause.Expression{},
			wantOrderExpressions: []clause.Expression{},
			wantErr:              true,
		},
		{
			name: "test-order-err2",
			fields: fields{
				Page:     0,
				PageSize: 0,
				Search:   nil,
				Order: []*OrderParam{
					{
						Field: "id",
						Order: "a",
					},
				},
			},
			args: args{
				model: UserTest{},
			},
			wantWhereExpressions: []clause.Expression{},
			wantOrderExpressions: []clause.Expression{},
			wantErr:              true,
		},
		{
			name: "test-order-err3",
			fields: fields{
				Page:     0,
				PageSize: 0,
				Search:   nil,
				Order: []*OrderParam{
					{
						Field: "pid",
						Order: ASC,
					},
				},
			},
			args: args{
				model: UserTest{},
			},
			wantWhereExpressions: []clause.Expression{},
			wantOrderExpressions: []clause.Expression{},
			wantErr:              true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Req{
				Page:     tt.fields.Page,
				PageSize: tt.fields.PageSize,
				Query:    tt.fields.Search,
				Order:    tt.fields.Order,
			}
			gotWhereExpressions, gotOrderExpressions, err := p.ConvertToGormExpression(tt.args.model)
			if (err != nil) != tt.wantErr {
				t.Errorf("ConvertToGormExpression() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotWhereExpressions, tt.wantWhereExpressions) {
				t.Errorf("ConvertToGormExpression() gotWhereExpressions = %v, want %v", gotWhereExpressions, tt.wantWhereExpressions)
			}
			if !reflect.DeepEqual(gotOrderExpressions, tt.wantOrderExpressions) {
				t.Errorf("ConvertToGormExpression() gotOrderExpressions = %v, want %v", gotOrderExpressions, tt.wantOrderExpressions)
			}
		})
	}
}

func TestReqToInterfaceSliceBranches(t *testing.T) {
	req := &Req{}

	values, err := req.ToInterfaceSlice([]interface{}{"a", 1})
	if err != nil {
		t.Fatalf("unexpected []interface{} error: %v", err)
	}
	if !reflect.DeepEqual(values, []interface{}{"a", 1}) {
		t.Fatalf("unexpected []interface{} values: %#v", values)
	}

	source := []string{"a", "b"}
	values, err = req.ToInterfaceSlice(&source)
	if err != nil {
		t.Fatalf("unexpected pointer slice error: %v", err)
	}
	if !reflect.DeepEqual(values, []interface{}{"a", "b"}) {
		t.Fatalf("unexpected pointer slice values: %#v", values)
	}

	values, err = req.ToInterfaceSlice([2]int{1, 2})
	if err != nil {
		t.Fatalf("unexpected array error: %v", err)
	}
	if !reflect.DeepEqual(values, []interface{}{1, 2}) {
		t.Fatalf("unexpected array values: %#v", values)
	}

	var nilSlice *[]string
	if _, err := req.ToInterfaceSlice(nilSlice); err == nil {
		t.Fatal("expected nil pointer error")
	}
	if _, err := req.ToInterfaceSlice(1); err == nil {
		t.Fatal("expected non-slice error")
	}
}

func TestReqConvertToCacheField(t *testing.T) {
	req := &Req{
		Page:     1,
		PageSize: 20,
		Query: []*QueryParam{{
			Field: "id",
			Value: "42",
			Exp:   EQ,
			Logic: AND,
		}},
		Order: []*OrderParam{{
			Field: "id",
			Order: ASC,
		}},
	}

	got := req.ConvertToCacheField()
	if len(got) != 64 {
		t.Fatalf("expected sha256 hex string, got %q", got)
	}
	if got != req.ConvertToCacheField() {
		t.Fatal("cache field should be deterministic")
	}

	other := *req
	other.Page = 2
	if got == other.ConvertToCacheField() {
		t.Fatal("different request should produce a different cache field")
	}

	bad := (&Req{Query: []*QueryParam{{Field: "id", Value: make(chan int)}}}).ConvertToCacheField()
	if bad != "" {
		t.Fatalf("marshal failure should return empty cache field, got %q", bad)
	}
}

func TestReqConvertToGormExpressionAdditionalBranches(t *testing.T) {
	if where, order, err := (*Req)(nil).ConvertToGormExpression(UserTest{}); err != nil || len(where) != 0 || len(order) != 0 {
		t.Fatalf("nil request should return empty expressions, got where=%v order=%v err=%v", where, order, err)
	}

	req := &Req{
		Query: []*QueryParam{
			{Field: "id", Exp: ISNULL, Logic: AND},
			{Field: "uid", Exp: ISNOTNULL, Logic: OR},
			{Field: "username", Exp: RAW, Value: clause.Expr{SQL: "username <> ?", Vars: []any{"root"}}, Logic: AND},
		},
		Order: []*OrderParam{{Field: "id", Order: DESC}},
	}
	where, order, err := req.ConvertToGormExpression(UserTest{})
	if err != nil {
		t.Fatalf("unexpected expression error: %v", err)
	}
	if len(where) != 3 || len(order) != 1 {
		t.Fatalf("unexpected expression counts: where=%d order=%d", len(where), len(order))
	}

	badRaw := &Req{Query: []*QueryParam{{Field: "id", Exp: RAW, Value: "id = 1"}}}
	if _, _, err := badRaw.ConvertToGormExpression(UserTest{}); err == nil {
		t.Fatal("expected RAW value type error")
	}
	nilQuery := &Req{Query: []*QueryParam{nil}}
	if _, _, err := nilQuery.ConvertToGormExpression(UserTest{}); err == nil {
		t.Fatal("expected nil query error")
	}
	nilOrder := &Req{Order: []*OrderParam{nil}}
	if _, _, err := nilOrder.ConvertToGormExpression(UserTest{}); err == nil {
		t.Fatal("expected nil order error")
	}
}

func TestFieldToColumnBranches(t *testing.T) {
	if got := fieldToColumn(nil); len(got) != 0 {
		t.Fatalf("nil model should return empty map: %#v", got)
	}
	if got := fieldToColumn(1); len(got) != 0 {
		t.Fatalf("non-struct model should return empty map: %#v", got)
	}
	type mixedTags struct {
		ID   int    `gorm:"column:id;primaryKey"`
		Name string `json:"name"`
	}
	got := fieldToColumn(&mixedTags{})
	if !reflect.DeepEqual(got, map[string]string{"id": "id"}) {
		t.Fatalf("unexpected field map: %#v", got)
	}
}

func TestReq_ConvertToGormExpressionPointerModel(t *testing.T) {
	p := &Req{
		Query: []*QueryParam{
			{
				Field: "id",
				Value: []string{"admin"},
				Exp:   IN,
				Logic: AND,
			},
		},
	}
	whereExpressions, orderExpressions, err := p.ConvertToGormExpression(&UserTest{})
	if err != nil {
		t.Fatalf("ConvertToGormExpression() error = %v", err)
	}
	if len(whereExpressions) != 1 {
		t.Fatalf("ConvertToGormExpression() got %d where expressions, want 1", len(whereExpressions))
	}
	if len(orderExpressions) != 0 {
		t.Fatalf("ConvertToGormExpression() got %d order expressions, want 0", len(orderExpressions))
	}
}

func TestReq_ConvertToGormExpressionInvalidInValue(t *testing.T) {
	p := &Req{
		Query: []*QueryParam{
			{
				Field: "id",
				Value: "admin",
				Exp:   IN,
				Logic: AND,
			},
		},
	}
	if _, _, err := p.ConvertToGormExpression(UserTest{}); err == nil {
		t.Fatal("ConvertToGormExpression() error = nil, want non-nil")
	}
}

func TestPaginatorReq_ConvertToPage(t *testing.T) {
	type fields struct {
		Page     int32
		PageSize int32
		Search   []*QueryParam
		Order    []*OrderParam
	}
	type args struct {
		total int32
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *Reply
		wantErr bool
	}{
		{
			name: "test1",
			fields: fields{
				Page:     0,
				PageSize: 0,
				Search:   nil,
				Order:    nil,
			},
			args: args{
				total: 0,
			},
			want: &Reply{
				Page:      0,
				PageSize:  0,
				Total:     0,
				PrevPage:  0,
				NextPage:  0,
				TotalPage: 0,
			},
			wantErr: false,
		},
		{
			name: "test2",
			fields: fields{
				Page:     1,
				PageSize: 100,
				Search:   nil,
				Order:    nil,
			},
			args: args{
				total: 100,
			},
			want: &Reply{
				Page:      1,
				PageSize:  100,
				Total:     100,
				PrevPage:  1,
				NextPage:  1,
				TotalPage: 1,
			},
			wantErr: false,
		},
		{
			name: "test-err-1",
			fields: fields{
				Page:     -1,
				PageSize: 0,
				Search:   nil,
				Order:    nil,
			},
			args: args{
				total: 0,
			},
			want: &Reply{
				Page:      0,
				PageSize:  0,
				Total:     0,
				PrevPage:  0,
				NextPage:  0,
				TotalPage: 0,
			},
			wantErr: true,
		},
		{
			name: "test-err-2",
			fields: fields{
				Page:     0,
				PageSize: -1,
				Search:   nil,
				Order:    nil,
			},
			args: args{
				total: 0,
			},
			want: &Reply{
				Page:      0,
				PageSize:  0,
				Total:     0,
				PrevPage:  0,
				NextPage:  0,
				TotalPage: 0,
			},
			wantErr: true,
		},
		{
			name: "test-err-3",
			fields: fields{
				Page:     0,
				PageSize: 1,
				Search:   nil,
				Order:    nil,
			},
			args: args{
				total: 0,
			},
			want: &Reply{
				Page:      0,
				PageSize:  0,
				Total:     0,
				PrevPage:  0,
				NextPage:  0,
				TotalPage: 0,
			},
			wantErr: true,
		},
		{
			name: "test-err-3",
			fields: fields{
				Page:     1,
				PageSize: 0,
				Search:   nil,
				Order:    nil,
			},
			args: args{
				total: 0,
			},
			want: &Reply{
				Page:      0,
				PageSize:  0,
				Total:     0,
				PrevPage:  0,
				NextPage:  0,
				TotalPage: 0,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Req{
				Page:     tt.fields.Page,
				PageSize: tt.fields.PageSize,
				Query:    tt.fields.Search,
				Order:    tt.fields.Order,
			}
			got, err := p.ConvertToPage(tt.args.total)
			if (err != nil) != tt.wantErr {
				t.Errorf("ConvertToPage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ConvertToPage() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_jsonToColumn(t *testing.T) {
	type args struct {
		model any
	}
	tests := []struct {
		name string
		args args
		want map[string]string
	}{
		{
			name: "",
			args: args{
				model: UserTest{},
			},
			want: map[string]string{
				"id":       "id",
				"uid":      "uid",
				"username": "username",
				"password": "password",
				"nickname": "nickname",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := fieldToColumn(tt.args.model); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("fieldToColumn() = %v, want %v", got, tt.want)
			}
		})
	}
}
