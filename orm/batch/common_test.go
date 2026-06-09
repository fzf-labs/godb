package batch

import (
	"database/sql/driver"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"
)

type sqlValueValuer struct {
	value driver.Value
	err   error
}

// Value 返回测试用 driver.Value 或预设错误。
func (v sqlValueValuer) Value() (driver.Value, error) {
	return v.value, v.err
}

type sqlValueStringer string

// String 返回测试用 SQL 字符串值。
func (s sqlValueStringer) String() string {
	return string(s)
}

type sqlValueRawBytes []byte

func testQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "''") + "'"
}

func TestFormatSQLValueCoversSupportedKinds(t *testing.T) {
	ts := time.Date(2026, 6, 4, 1, 2, 3, 0, time.UTC)
	str := "ptr"

	tests := []struct {
		name  string
		value reflect.Value
		want  string
	}{
		{name: "nil pointer", value: reflect.ValueOf((*string)(nil)), want: "NULL"},
		{name: "time", value: reflect.ValueOf(ts), want: "'2026-06-04 01:02:03'"},
		{name: "time pointer", value: reflect.ValueOf(&ts), want: "'2026-06-04 01:02:03'"},
		{name: "driver valuer", value: reflect.ValueOf(sqlValueValuer{value: int64(12)}), want: "12"},
		{name: "stringer", value: reflect.ValueOf(sqlValueStringer("hello")), want: "'hello'"},
		{name: "int", value: reflect.ValueOf(int(-1)), want: "-1"},
		{name: "int8", value: reflect.ValueOf(int8(-2)), want: "-2"},
		{name: "int16", value: reflect.ValueOf(int16(-3)), want: "-3"},
		{name: "int32", value: reflect.ValueOf(int32(-4)), want: "-4"},
		{name: "int64", value: reflect.ValueOf(int64(-5)), want: "-5"},
		{name: "uint", value: reflect.ValueOf(uint(1)), want: "1"},
		{name: "uint8", value: reflect.ValueOf(uint8(2)), want: "2"},
		{name: "uint16", value: reflect.ValueOf(uint16(3)), want: "3"},
		{name: "uint32", value: reflect.ValueOf(uint32(4)), want: "4"},
		{name: "uint64", value: reflect.ValueOf(uint64(5)), want: "5"},
		{name: "string", value: reflect.ValueOf("a'b"), want: "'a''b'"},
		{name: "float32", value: reflect.ValueOf(float32(1.25)), want: "1.25"},
		{name: "float64", value: reflect.ValueOf(float64(2.5)), want: "2.5"},
		{name: "bool", value: reflect.ValueOf(true), want: "true"},
		{name: "bytes", value: reflect.ValueOf([]byte("blob")), want: "'blob'"},
		{name: "pointer", value: reflect.ValueOf(&str), want: "'ptr'"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := formatSQLValue(tt.value, testQuote)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("got %q want %q", got, tt.want)
			}
		})
	}
}

func TestFormatSQLValueErrors(t *testing.T) {
	if _, err := formatSQLValue(reflect.Value{}, testQuote); err == nil {
		t.Fatal("expected invalid reflect value error")
	}

	wantErr := errors.New("valuer failed")
	if _, err := formatSQLValue(reflect.ValueOf(sqlValueValuer{err: wantErr}), testQuote); !errors.Is(err, wantErr) {
		t.Fatalf("got %v want %v", err, wantErr)
	}

	if _, err := formatSQLValue(reflect.ValueOf([]int{1}), testQuote); err == nil {
		t.Fatal("expected unsupported slice error")
	}
}

func TestFormatSQLValueFromAnyCoversSupportedKinds(t *testing.T) {
	ts := time.Date(2026, 6, 4, 1, 2, 3, 0, time.UTC)
	ptr := 9

	tests := []struct {
		name  string
		value any
		want  string
	}{
		{name: "nil", value: nil, want: "NULL"},
		{name: "nil time pointer", value: (*time.Time)(nil), want: "NULL"},
		{name: "time", value: ts, want: "'2026-06-04 01:02:03'"},
		{name: "time pointer", value: &ts, want: "'2026-06-04 01:02:03'"},
		{name: "driver valuer", value: sqlValueValuer{value: "db"}, want: "'db'"},
		{name: "stringer", value: sqlValueStringer("shown"), want: "'shown'"},
		{name: "string", value: "plain", want: "'plain'"},
		{name: "bytes", value: []byte("raw"), want: "'raw'"},
		{name: "bool", value: false, want: "false"},
		{name: "int", value: int(-1), want: "-1"},
		{name: "int8", value: int8(-2), want: "-2"},
		{name: "int16", value: int16(-3), want: "-3"},
		{name: "int32", value: int32(-4), want: "-4"},
		{name: "int64", value: int64(-5), want: "-5"},
		{name: "uint", value: uint(1), want: "1"},
		{name: "uint8", value: uint8(2), want: "2"},
		{name: "uint16", value: uint16(3), want: "3"},
		{name: "uint32", value: uint32(4), want: "4"},
		{name: "uint64", value: uint64(5), want: "5"},
		{name: "float32", value: float32(1.25), want: "1.25"},
		{name: "float64", value: float64(2.5), want: "2.5"},
		{name: "pointer", value: &ptr, want: "9"},
		{name: "named bytes", value: sqlValueRawBytes("named"), want: "'named'"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := formatSQLValueFromAny(tt.value, testQuote)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("got %q want %q", got, tt.want)
			}
		})
	}
}

func TestFormatSQLValueFromAnyErrors(t *testing.T) {
	wantErr := errors.New("valuer failed")
	if _, err := formatSQLValueFromAny(sqlValueValuer{err: wantErr}, testQuote); !errors.Is(err, wantErr) {
		t.Fatalf("got %v want %v", err, wantErr)
	}

	if _, err := formatSQLValueFromAny(struct{}{}, testQuote); err == nil {
		t.Fatal("expected unsupported struct error")
	}
}

func TestFormatBatchIDValueCoversSupportedKinds(t *testing.T) {
	id := int64(42)
	tests := []struct {
		name  string
		value any
		want  string
	}{
		{name: "int", value: int(1), want: "1"},
		{name: "int pointer", value: &id, want: "42"},
		{name: "uint", value: uint(2), want: "2"},
		{name: "string", value: "abc'123", want: "'abc''123'"},
		{name: "bool fallback", value: true, want: "true"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := formatBatchIDValue(reflect.ValueOf(tt.value), testQuote)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("got %q want %q", got, tt.want)
			}
		})
	}
}

func TestFormatBatchIDValueRejectsInvalidIDs(t *testing.T) {
	var nilID *int64
	tests := []struct {
		name  string
		value reflect.Value
		want  string
	}{
		{name: "invalid", value: reflect.Value{}, want: "id field is invalid"},
		{name: "nil pointer", value: reflect.ValueOf(nilID), want: "empty id value"},
		{name: "zero int", value: reflect.ValueOf(0), want: "id value must be greater than 0"},
		{name: "negative int", value: reflect.ValueOf(-1), want: "id value must be greater than 0"},
		{name: "zero uint", value: reflect.ValueOf(uint(0)), want: "id value must be greater than 0"},
		{name: "blank string", value: reflect.ValueOf(" \t\n"), want: "empty id value"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := formatBatchIDValue(tt.value, testQuote)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("expected %q error, got %v", tt.want, err)
			}
		})
	}
}
