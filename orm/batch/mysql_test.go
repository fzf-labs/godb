package batch

import (
	"database/sql"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// TestBatchUpdateToSQLArray 验证 MySQL 批量更新 SQL 生成。
func TestBatchUpdateToSQLArray(t *testing.T) {
	type TestStruct struct {
		ID        int64  `json:"id" gorm:"column:id;primary_key"`
		Name      string `json:"name" gorm:"column:name"`
		Age       int    `json:"age" gorm:"column:age"`
		IsDeleted bool   `json:"is_deleted" gorm:"column:is_deleted"`
	}

	testData := []*TestStruct{
		{
			ID:        1,
			Name:      "test1",
			Age:       18,
			IsDeleted: false,
		},
		{
			ID:        2,
			Name:      "test2",
			Age:       20,
			IsDeleted: true,
		},
	}

	sqlArray, err := MysqlBatchUpdateToSQLArray("test_table", testData)
	t.Log(sqlArray)
	assert.NoError(t, err)
	assert.NotEmpty(t, sqlArray)
	assert.Len(t, sqlArray, 1)
	assert.Contains(t, sqlArray[0], "`age` = CASE `id`")
	assert.Contains(t, sqlArray[0], "`is_deleted` = CASE `id`")
	assert.Contains(t, sqlArray[0], "`name` = CASE `id`")
	assert.True(t, strings.Index(sqlArray[0], "`age` = CASE `id`") < strings.Index(sqlArray[0], "`is_deleted` = CASE `id`"))
	assert.True(t, strings.Index(sqlArray[0], "`is_deleted` = CASE `id`") < strings.Index(sqlArray[0], "`name` = CASE `id`"))
}

// TestMysqlBatchUpdateToSQLArray_InvalidIdentifier 验证非法表名会返回错误。
func TestMysqlBatchUpdateToSQLArray_InvalidIdentifier(t *testing.T) {
	type BadStruct struct {
		ID   int64  `gorm:"column:id"`
		Name string `gorm:"column:user-name"`
	}

	_, err := MysqlBatchUpdateToSQLArray("bad-table", []*BadStruct{{ID: 1, Name: "test"}})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid SQL identifier")
}

// TestMysqlBatchUpdateToSQLArray_QualifiedTableName 验证限定表名会按段转义。
func TestMysqlBatchUpdateToSQLArray_QualifiedTableName(t *testing.T) {
	type TestStruct struct {
		ID   int64  `gorm:"column:id"`
		Name string `gorm:"column:name"`
	}

	sqlArray, err := MysqlBatchUpdateToSQLArray("test_db.test_table", []*TestStruct{{ID: 1, Name: "test"}})
	assert.NoError(t, err)
	assert.Len(t, sqlArray, 1)
	assert.Contains(t, sqlArray[0], "UPDATE `test_db`.`test_table` SET ")
}

// TestMysqlBatchUpdateToSQLArray_InvalidColumnIdentifier 验证非法列名会返回错误。
func TestMysqlBatchUpdateToSQLArray_InvalidColumnIdentifier(t *testing.T) {
	type BadStruct struct {
		ID   int64  `gorm:"column:id"`
		Name string `gorm:"column:user-name"`
	}

	_, err := MysqlBatchUpdateToSQLArray("test_table", []*BadStruct{{ID: 1, Name: "test"}})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid SQL identifier")
}

func TestMysqlBatchUpdateToSQLArray_NilElementReturnsError(t *testing.T) {
	type TestStruct struct {
		ID   int64  `gorm:"column:id"`
		Name string `gorm:"column:name"`
	}

	_, err := MysqlBatchUpdateToSQLArray("test_table", []*TestStruct{nil})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "dataList[0] cannot be nil")
}

func TestMysqlBatchUpdateToSQLArray_SupportsCommonComplexTypes(t *testing.T) {
	fixedTime := time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC)
	type TestStruct struct {
		ID        int64          `gorm:"column:id"`
		CreatedAt time.Time      `gorm:"column:created_at"`
		DeletedAt gorm.DeletedAt `gorm:"column:deleted_at"`
		Payload   datatypes.JSON `gorm:"column:payload"`
		Raw       []byte         `gorm:"column:raw"`
		OptTime   sql.NullTime   `gorm:"column:opt_time"`
	}

	sqlArray, err := MysqlBatchUpdateToSQLArray("test_table", []*TestStruct{
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

func TestMysqlBatchUpdateToSQLArray_UsesConfiguredIDColumn(t *testing.T) {
	type TestStruct struct {
		ID   int64  `gorm:"column:user_id"`
		Name string `gorm:"column:name"`
	}

	sqlArray, err := MysqlBatchUpdateToSQLArray("test_table", []*TestStruct{{ID: 7, Name: "alice"}})
	require.NoError(t, err)
	require.Len(t, sqlArray, 1)
	assert.Contains(t, sqlArray[0], "`name` = CASE `user_id` WHEN 7 THEN 'alice' END")
	assert.Contains(t, sqlArray[0], "WHERE `user_id` IN (7)")
	assert.NotContains(t, sqlArray[0], "CASE id")
	assert.NotContains(t, sqlArray[0], "WHERE id IN")
}

func TestMysqlBatchUpdateToSQLArray_EscapesSingleQuotesByDoubling(t *testing.T) {
	type TestStruct struct {
		ID   int64  `gorm:"column:id"`
		Name string `gorm:"column:name"`
	}

	sqlArray, err := MysqlBatchUpdateToSQLArray("test_table", []*TestStruct{{ID: 1, Name: "O'Reilly"}})
	require.NoError(t, err)
	require.Len(t, sqlArray, 1)
	assert.Contains(t, sqlArray[0], "'O''Reilly'")
	assert.NotContains(t, sqlArray[0], `O\'Reilly`)
}

func TestMysqlBatchUpdateToSQLArray_RejectsIDOnlyStruct(t *testing.T) {
	type TestStruct struct {
		ID int64 `gorm:"column:id"`
	}

	_, err := MysqlBatchUpdateToSQLArray("test_table", []*TestStruct{{ID: 1}})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no update columns found")
}
