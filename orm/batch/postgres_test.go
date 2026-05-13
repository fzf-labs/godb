package batch

import (
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testUser struct {
	ID       int64  `gorm:"column:id"`
	Name     string `gorm:"column:name"`
	Age      int    `gorm:"column:age"`
	IsActive bool   `gorm:"column:is_active"`
}

// TestPostgresBatchUpdateToSQLArray 验证 PostgreSQL 批量更新 SQL 的基础参数校验。
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

// TestPostgresBatchUpdateToSQLArray_LargeBatch 验证 PostgreSQL 批量更新的分页和字段顺序。
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
	ids := countBatchIDs(got[0])
	if ids != 200 {
		t.Errorf("First batch contains %d records, want 200", ids)
	}
	// 验证第二个批次包含50条记录
	ids = countBatchIDs(got[1])
	if ids != 50 {
		t.Errorf("Second batch contains %d records, want 50", ids)
	}
	assert.Contains(t, got[1], `WHEN 201 THEN 'User201'`)
	assert.Contains(t, got[1], `WHEN 250 THEN 'User250'`)
	assert.NotContains(t, got[1], `WHEN 201 THEN 'User1'`)
	assert.True(t, strings.Index(got[0], `"age" = CASE "id"`) < strings.Index(got[0], `"is_active" = CASE "id"`))
	assert.True(t, strings.Index(got[0], `"is_active" = CASE "id"`) < strings.Index(got[0], `"name" = CASE "id"`))
}

// countBatchIDs 统计 SQL WHERE IN 子句中的 ID 数量。
func countBatchIDs(sql string) int {
	const wherePrefix = ` WHERE "id" IN (`
	idx := strings.LastIndex(sql, wherePrefix)
	if idx < 0 {
		return 0
	}
	ids := strings.TrimSuffix(sql[idx+len(wherePrefix):], ")")
	if ids == "" {
		return 0
	}
	return len(strings.Split(ids, ","))
}

// normalizeSQL 规范化 SQL 语句以便比较
func normalizeSQL(sql string) string {
	// 移除多余的空格
	sql = strings.Join(strings.Fields(sql), " ")
	return sql
}

// TestPostgresBatchUpdateToSQLArray_InvalidIdentifier 验证非法表名会返回错误。
func TestPostgresBatchUpdateToSQLArray_InvalidIdentifier(t *testing.T) {
	type BadStruct struct {
		ID   int64  `gorm:"column:id"`
		Name string `gorm:"column:user-name"`
	}

	_, err := PostgresBatchUpdateToSQLArray(`bad-table`, []*BadStruct{{ID: 1, Name: "test"}})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid SQL identifier")
}

// TestPostgresBatchUpdateToSQLArray_QualifiedTableName 验证限定表名会按段转义。
func TestPostgresBatchUpdateToSQLArray_QualifiedTableName(t *testing.T) {
	sqlArray, err := PostgresBatchUpdateToSQLArray("public.users", []*testUser{{ID: 1, Name: "test", Age: 18, IsActive: true}})
	assert.NoError(t, err)
	assert.Len(t, sqlArray, 1)
	assert.Contains(t, sqlArray[0], `UPDATE "public"."users" SET `)
}

// TestPostgresBatchUpdateToSQLArray_InvalidColumnIdentifier 验证非法列名会返回错误。
func TestPostgresBatchUpdateToSQLArray_InvalidColumnIdentifier(t *testing.T) {
	type BadStruct struct {
		ID   int64  `gorm:"column:id"`
		Name string `gorm:"column:user-name"`
	}

	_, err := PostgresBatchUpdateToSQLArray(`users`, []*BadStruct{{ID: 1, Name: "test"}})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid SQL identifier")
}
