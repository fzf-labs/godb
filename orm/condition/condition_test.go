package condition

import (
	"reflect"
	"testing"

	"gorm.io/gorm/clause"
)

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
