package batch

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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
}
