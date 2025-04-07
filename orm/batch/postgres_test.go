package batch

import (
	"strconv"
	"strings"
	"testing"
)

type testUser struct {
	ID       int64  `gorm:"column:id"`
	Name     string `gorm:"column:name"`
	Age      int    `gorm:"column:age"`
	IsActive bool   `gorm:"column:is_active"`
}

func TestPostgresBatchUpdateToSQLArray(t *testing.T) {
	tests := []struct {
		name      string
		tableName string
		dataList  []*testUser
		want      string
		wantErr   bool
	}{
		{
			name:      "empty table name",
			tableName: "",
			dataList:  []*testUser{},
			want:      "",
			wantErr:   true,
		},
		{
			name:      "empty data list",
			tableName: "users",
			dataList:  []*testUser{},
			want:      "",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := PostgresBatchUpdateToSQLArray(tt.tableName, tt.dataList)
			if (err != nil) != tt.wantErr {
				t.Errorf("PostgresBatchUpdateToSQLArray() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}

			if len(got) != 1 {
				t.Errorf("PostgresBatchUpdateToSQLArray() got %d SQL statements, want 1", len(got))
				return
			}

			// 规范化 SQL 语句(移除多余的空格)以便比较
			gotNormalized := normalizeSQL(got[0])
			wantNormalized := normalizeSQL(tt.want)

			if gotNormalized != wantNormalized {
				t.Errorf("PostgresBatchUpdateToSQLArray() got = %v, want %v", gotNormalized, wantNormalized)
			}
		})
	}
}

func TestPostgresBatchUpdateToSQLArray_LargeBatch(t *testing.T) {
	// 创建超过 batchSize 的数据
	dataList := make([]*testUser, 250)
	for i := 0; i < 250; i++ {
		dataList[i] = &testUser{
			ID:       int64(i + 1),
			Name:     "User" + strconv.Itoa(i+1),
			Age:      25 + i,
			IsActive: i%2 == 0,
		}
	}
	got, err := PostgresBatchUpdateToSQLArray("users", dataList)
	if err != nil {
		t.Errorf("PostgresBatchUpdateToSQLArray() error = %v", err)
		return
	}
	// 打印生成的SQL语句
	for _, sql := range got {
		t.Log(sql)
	}
	if len(got) != 2 {
		t.Errorf("PostgresBatchUpdateToSQLArray() got %d SQL statements, want 2", len(got))
		return
	}
	// 验证第一个批次包含200条记录
	ids := strings.Count(got[0], "WHEN")
	if ids != 200 {
		t.Errorf("First batch contains %d records, want 200", ids)
	}
	// 验证第二个批次包含50条记录
	ids = strings.Count(got[1], "WHEN")
	if ids != 50 {
		t.Errorf("Second batch contains %d records, want 50", ids)
	}
}

// normalizeSQL 规范化 SQL 语句以便比较
func normalizeSQL(sql string) string {
	// 移除多余的空格
	sql = strings.Join(strings.Fields(sql), " ")
	return sql
}
