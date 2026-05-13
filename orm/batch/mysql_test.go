package batch

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
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
	assert.Contains(t, sqlArray[0], "`age` = CASE id")
	assert.Contains(t, sqlArray[0], "`is_deleted` = CASE id")
	assert.Contains(t, sqlArray[0], "`name` = CASE id")
	assert.True(t, strings.Index(sqlArray[0], "`age` = CASE id") < strings.Index(sqlArray[0], "`is_deleted` = CASE id"))
	assert.True(t, strings.Index(sqlArray[0], "`is_deleted` = CASE id") < strings.Index(sqlArray[0], "`name` = CASE id"))
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
