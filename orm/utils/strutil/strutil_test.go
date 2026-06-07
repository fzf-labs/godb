package strutil

import (
	"strings"
	"testing"
	"time"
)

func TestStrSliFind(t *testing.T) {
	if !StrSliFind([]string{"a", "b"}, "b") {
		t.Fatal("expected element to be found")
	}
	if StrSliFind([]string{"a", "b"}, "c") {
		t.Fatal("did not expect element to be found")
	}
}

func TestSliRemove(t *testing.T) {
	got := SliRemove([]string{"a", "b", "c", "d"}, []string{"b", "d"})
	want := []string{"a", "c"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("unexpected result: %#v", got)
	}
}

func TestSliRemoveRemovesAllMatchingValues(t *testing.T) {
	got := SliRemove([]string{"a", "b", "b", "c"}, []string{"b"})
	want := []string{"a", "c"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("unexpected result: %#v", got)
	}
}

func TestSliRemoveDoesNotMutateInput(t *testing.T) {
	collection := []string{"a", "b", "c"}
	original := append([]string(nil), collection...)

	got := SliRemove(collection, []string{"b"})
	if strings.Join(got, ",") != "a,c" {
		t.Fatalf("unexpected result: %#v", got)
	}
	if strings.Join(collection, ",") != strings.Join(original, ",") {
		t.Fatalf("input slice was mutated: before=%#v after=%#v", original, collection)
	}
}

func TestConvToString(t *testing.T) {
	now := time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC)
	zero := time.Time{}
	intPtr := 7
	var nilMap map[string]string
	tests := []struct {
		name string
		in   any
		want string
	}{
		{name: "nil", in: nil, want: ""},
		{name: "int", in: int(1), want: "1"},
		{name: "int8", in: int8(2), want: "2"},
		{name: "int16", in: int16(3), want: "3"},
		{name: "int32", in: int32(4), want: "4"},
		{name: "int64", in: int64(5), want: "5"},
		{name: "uint", in: uint(6), want: "6"},
		{name: "uint8", in: uint8(7), want: "7"},
		{name: "uint16", in: uint16(8), want: "8"},
		{name: "uint32", in: uint32(9), want: "9"},
		{name: "uint64", in: uint64(10), want: "10"},
		{name: "float32", in: float32(1.25), want: "1.25"},
		{name: "float64", in: float64(2.5), want: "2.5"},
		{name: "bool", in: true, want: "true"},
		{name: "string", in: "ok", want: "ok"},
		{name: "bytes", in: []byte("bytes"), want: "bytes"},
		{name: "zero time", in: time.Time{}, want: ""},
		{name: "time", in: now, want: now.String()},
		{name: "nil time pointer", in: (*time.Time)(nil), want: ""},
		{name: "zero time pointer", in: &zero, want: ""},
		{name: "time pointer", in: &now, want: now.String()},
		{name: "int pointer", in: &intPtr, want: "7"},
		{name: "nil map", in: nilMap, want: ""},
		{name: "json fallback", in: map[string]int{"a": 1}, want: `{"a":1}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ConvToString(tt.in); got != tt.want {
				t.Fatalf("ConvToString(%T)=%q want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestConvToStringUsesSprintWhenJSONFails(t *testing.T) {
	ch := make(chan int)
	got := ConvToString(struct {
		Ch chan int
	}{Ch: ch})
	if !strings.Contains(got, "0x") {
		t.Fatalf("expected fmt fallback for unsupported json value, got %q", got)
	}
}
