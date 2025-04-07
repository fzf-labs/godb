package utils

import (
	"reflect"
	"testing"
)

func TestFillModelPkgPath(t *testing.T) {
	type args struct {
		filePath string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "test",
			args: args{
				filePath: "./",
			},
			want: "github.com/fzf-labs/godb/orm/gen/utils/util",
		},
		{
			name: "test",
			args: args{
				filePath: "./util1",
			},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FillModelPkgPath(tt.args.filePath); got != tt.want {
				t.Errorf("FillModelPkgPath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStrSliFind(t *testing.T) {
	type args struct {
		collection []string
		element    string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "test1",
			args: args{
				collection: []string{"a", "b"},
				element:    "c",
			},
			want: false,
		},
		{
			name: "test2",
			args: args{
				collection: []string{"a", "b"},
				element:    "a",
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := StrSliFind(tt.args.collection, tt.args.element); got != tt.want {
				t.Errorf("StrSliFind() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSliRemove(t *testing.T) {
	type args struct {
		collection []string
		element    []string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "test1",
			args: args{
				collection: []string{"a", "b"},
				element:    []string{"c"},
			},
			want: []string{"a", "b"},
		},
		{
			name: "test1",
			args: args{
				collection: []string{"a", "b"},
				element:    []string{"a"},
			},
			want: []string{"b"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SliRemove(tt.args.collection, tt.args.element); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SliRemove() = %v, want %v", got, tt.want)
			}
		})
	}
}
