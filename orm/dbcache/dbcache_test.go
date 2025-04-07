package dbcache

import (
	"testing"
	"time"
)

func TestKeyFormat(t *testing.T) {
	nt := time.Now()
	ntStr := nt.Format("2006-01-02 15:04:05")
	var nilTime *time.Time
	type args struct {
		any any
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "case nil",
			args: args{
				any: nil,
			},
			want: "",
		},
		{
			name: "case true",
			args: args{
				any: true,
			},
			want: "true",
		},
		{
			name: "case false",
			args: args{
				any: false,
			},
			want: "false",
		},
		{
			name: "case int",
			args: args{
				any: 8,
			},
			want: "8",
		},
		{
			name: "case int8",
			args: args{
				any: int8(8),
			},
			want: "8",
		},
		{
			name: "case int16",
			args: args{
				any: int16(8),
			},
			want: "8",
		},
		{
			name: "case int32",
			args: args{
				any: int32(8),
			},
			want: "8",
		},
		{
			name: "case int64",
			args: args{
				any: int64(8),
			},
			want: "8",
		},
		{
			name: "case uint",
			args: args{
				any: uint(8),
			},
			want: "8",
		},
		{
			name: "case uint8",
			args: args{
				any: uint8(8),
			},
			want: "8",
		},
		{
			name: "case uint16",
			args: args{
				any: uint16(8),
			},
			want: "8",
		},
		{
			name: "case uint32",
			args: args{
				any: uint32(8),
			},
			want: "8",
		},
		{
			name: "case uint64",
			args: args{
				any: uint64(8),
			},
			want: "8",
		},
		{
			name: "case float32",
			args: args{
				any: float32(8),
			},
			want: "8",
		},
		{
			name: "case float64",
			args: args{
				any: float64(8),
			},
			want: "8",
		},
		{
			name: "case string",
			args: args{
				any: "8",
			},
			want: "8",
		},
		{
			name: "case []byte",
			args: args{
				any: "8",
			},
			want: "8",
		},
		{
			name: "case struct",
			args: args{
				any: struct{}{},
			},
			want: "{}",
		},
		{
			name: "case time",
			args: args{
				any: nt,
			},
			want: ntStr,
		},
		{
			name: "case *time",
			args: args{
				any: &nt,
			},
			want: ntStr,
		},
		{
			name: "case null time",
			args: args{
				any: time.Time{},
			},
			want: "",
		},
		{
			name: "case null time",
			args: args{
				any: nilTime,
			},
			want: "",
		},
		{
			name: "case []byte",
			args: args{
				any: []byte("8"),
			},
			want: "8",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := KeyFormat(tt.args.any); got != tt.want {
				t.Errorf("KeyFormat() = %v, want %v", got, tt.want)
			}
		})
	}
}
