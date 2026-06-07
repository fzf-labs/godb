package batch

import (
	"reflect"
	"testing"
)

type batchNoGormTag struct {
	ID int `gorm:"column:id"`
	V  int
}

type batchNoColumnTag struct {
	ID int `gorm:"primaryKey"`
}

type batchDuplicateColumn struct {
	ID    int `gorm:"column:id"`
	Other int `gorm:"column:id"`
}

type batchInvalidColumn struct {
	ID int `gorm:"column:bad-name"`
}

type batchNoID struct {
	Name string `gorm:"column:name"`
}

type batchEmptyStringID struct {
	ID   string `gorm:"column:id"`
	Name string `gorm:"column:name"`
}

type batchUnsupportedField struct {
	ID     int   `gorm:"column:id"`
	Values []int `gorm:"column:values"`
}

func TestCommonHelpersErrorBranches(t *testing.T) {
	if got := sortedFieldNames(map[string][]string{"b": nil, "a": nil}, "a"); !reflect.DeepEqual(got, []string{"b"}) {
		t.Fatalf("unexpected sorted fields: %#v", got)
	}

	for _, args := range [][3]int{{-1, 0, 1}, {1, 0, 1}, {0, 2, 1}} {
		if _, err := sliceBatchValues([]string{"a"}, args[0], args[1]); err == nil {
			t.Fatalf("expected invalid range error for %#v", args)
		}
	}
}

func TestGetStructFieldsErrors(t *testing.T) {
	tests := []struct {
		name  string
		model any
	}{
		{name: "missing gorm tag", model: batchNoGormTag{}},
		{name: "missing column tag", model: batchNoColumnTag{}},
		{name: "duplicate column", model: batchDuplicateColumn{}},
		{name: "invalid column", model: batchInvalidColumn{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := getStructFields(reflect.TypeOf(tt.model)); err == nil {
				t.Fatal("expected getStructFields error")
			}
		})
	}
}

func TestBatchUpdateValidationErrors(t *testing.T) {
	tests := []struct {
		name        string
		data        any
		mysqlErr    bool
		postgresErr bool
	}{
		{name: "nil", data: nil, mysqlErr: true, postgresErr: true},
		{name: "not slice", data: batchNoID{}, mysqlErr: true, postgresErr: true},
		{name: "empty slice", data: []*batchNoID{}, mysqlErr: true, postgresErr: true},
		{name: "not pointer slice", data: []batchNoID{{Name: "a"}}, mysqlErr: true, postgresErr: true},
		{name: "missing id", data: []*batchNoID{{Name: "a"}}, mysqlErr: true, postgresErr: true},
		{name: "empty id", data: []*batchEmptyStringID{{Name: "a"}}, mysqlErr: true},
		{name: "unsupported field", data: []*batchUnsupportedField{{ID: 1, Values: []int{1}}}, mysqlErr: true, postgresErr: true},
	}

	for _, tt := range tests {
		t.Run("mysql "+tt.name, func(t *testing.T) {
			_, err := MysqlBatchUpdateToSQLArray("users", tt.data)
			if tt.mysqlErr && err == nil {
				t.Fatal("expected mysql batch error")
			}
			if !tt.mysqlErr && err != nil {
				t.Fatalf("unexpected mysql batch error: %v", err)
			}
		})
		t.Run("postgres "+tt.name, func(t *testing.T) {
			_, err := PostgresBatchUpdateToSQLArray("users", tt.data)
			if tt.postgresErr && err == nil {
				t.Fatal("expected postgres batch error")
			}
			if !tt.postgresErr && err != nil {
				t.Fatalf("unexpected postgres batch error: %v", err)
			}
		})
	}
}

func TestBuildBatchUpdateSQLErrors(t *testing.T) {
	if _, err := buildBatchUpdateSQL("users", map[string][]string{"name": {"a"}}, 0, 0, nil); err == nil {
		t.Fatal("expected empty mysql batch ids error")
	}
	if _, err := buildBatchUpdateSQL("users", map[string][]string{"name": {"a"}}, 0, 2, []string{"1", "2"}); err == nil {
		t.Fatal("expected mysql slice range error")
	}
	if _, err := buildPostgresBatchUpdateSQL("users", map[string][]string{"name": {"a"}}, 0, 0, nil); err == nil {
		t.Fatal("expected empty postgres batch ids error")
	}
	if _, err := buildPostgresBatchUpdateSQL("users", map[string][]string{"name": {"a"}}, 0, 2, []string{"1", "2"}); err == nil {
		t.Fatal("expected postgres slice range error")
	}
}
