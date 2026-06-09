package batch

import (
	"database/sql"
	"math"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"
	"gorm.io/gorm"
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
	assert.Contains(t, got[0], `"is_active" = CASE "id" WHEN 1 THEN TRUE WHEN 2 THEN FALSE`)
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

func TestPostgresBatchUpdateToSQLArray_TrimsTableName(t *testing.T) {
	sqlArray, err := PostgresBatchUpdateToSQLArray(" users ", []*testUser{{ID: 1, Name: "test", Age: 18, IsActive: true}})
	assert.NoError(t, err)
	assert.Len(t, sqlArray, 1)
	assert.Contains(t, sqlArray[0], `UPDATE "users" SET `)
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

func TestPostgresBatchUpdateToSQLArray_NilElementReturnsError(t *testing.T) {
	_, err := PostgresBatchUpdateToSQLArray("users", []*testUser{nil})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "dataList[0] cannot be nil")
}

func TestPostgresBatchUpdateToSQLArray_SupportsCommonComplexTypes(t *testing.T) {
	fixedTime := time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC)
	type TestStruct struct {
		ID        int64          `gorm:"column:id"`
		CreatedAt time.Time      `gorm:"column:created_at"`
		DeletedAt gorm.DeletedAt `gorm:"column:deleted_at"`
		Payload   datatypes.JSON `gorm:"column:payload"`
		Raw       []byte         `gorm:"column:raw"`
		OptTime   sql.NullTime   `gorm:"column:opt_time"`
	}

	sqlArray, err := PostgresBatchUpdateToSQLArray("users", []*TestStruct{
		{
			ID:        1,
			CreatedAt: fixedTime,
			DeletedAt: gorm.DeletedAt{Time: fixedTime, Valid: true},
			Payload:   datatypes.JSON([]byte(`{"mode":"test"}`)),
			Raw:       []byte("abc"),
			OptTime:   sql.NullTime{Time: fixedTime, Valid: true},
		},
	})
	assert.NoError(t, err)
	assert.Len(t, sqlArray, 1)
	assert.Contains(t, sqlArray[0], "'2024-01-02 03:04:05'")
	assert.Contains(t, sqlArray[0], "'abc'")
	assert.Contains(t, sqlArray[0], `{"mode":"test"}`)
}

func TestPostgresBatchUpdateToSQLArray_SupportsPointerToSlice(t *testing.T) {
	data := []*testUser{{ID: 1, Name: "alice", Age: 18, IsActive: true}}

	sqlArray, err := PostgresBatchUpdateToSQLArray("users", &data)
	assert.NoError(t, err)
	assert.Len(t, sqlArray, 1)
	assert.Contains(t, sqlArray[0], `"name" = CASE "id" WHEN 1 THEN 'alice' END`)
}

func TestPostgresBatchUpdateToSQLArray_UsesConfiguredIDColumn(t *testing.T) {
	type TestStruct struct {
		ID   int64  `gorm:"column:user_id"`
		Name string `gorm:"column:name"`
	}

	sqlArray, err := PostgresBatchUpdateToSQLArray("users", []*TestStruct{{ID: 7, Name: "alice"}})
	assert.NoError(t, err)
	assert.Len(t, sqlArray, 1)
	assert.Contains(t, sqlArray[0], `"name" = CASE "user_id" WHEN 7 THEN 'alice' END`)
	assert.Contains(t, sqlArray[0], `WHERE "user_id" IN (7)`)
	assert.NotContains(t, sqlArray[0], `CASE "id"`)
	assert.NotContains(t, sqlArray[0], `WHERE "id" IN`)
}

func TestPostgresBatchUpdateToSQLArray_RejectsEmptyStringID(t *testing.T) {
	type TestStruct struct {
		ID   string `gorm:"column:id"`
		Name string `gorm:"column:name"`
	}

	_, err := PostgresBatchUpdateToSQLArray("users", []*TestStruct{{Name: "alice"}})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "empty id value at index 0")
}

func TestPostgresBatchUpdateToSQLArray_RejectsIDOnlyStruct(t *testing.T) {
	type TestStruct struct {
		ID int64 `gorm:"column:id"`
	}

	_, err := PostgresBatchUpdateToSQLArray("users", []*TestStruct{{ID: 1}})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no update columns found")
}

func TestPostgresBatchUpdateToSQLArray_RejectsNonPositiveNumericID(t *testing.T) {
	type TestStruct struct {
		ID   int64  `gorm:"column:id"`
		Name string `gorm:"column:name"`
	}

	for _, id := range []int64{0, -1} {
		_, err := PostgresBatchUpdateToSQLArray("users", []*TestStruct{{ID: id, Name: "alice"}})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "id value must be greater than 0 at index 0")
	}
}

func TestPostgresBatchUpdateToSQLArray_RejectsInvalidFloatValues(t *testing.T) {
	type TestStruct struct {
		ID    int64   `gorm:"column:id"`
		Score float64 `gorm:"column:score"`
	}

	_, err := PostgresBatchUpdateToSQLArray("users", []*TestStruct{{ID: 1, Score: math.Inf(1)}})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported float value")
}
