package file

import (
	"testing"
)

func TestFileExists(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "test1",
			args: args{
				path: "./file.go",
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Exists(tt.args.path); got != tt.want {
				t.Errorf("Exists() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMkdirPath(t *testing.T) {
	type args struct {
		relativePath string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "test1",
			args: args{
				relativePath: "./test",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := MkdirPath(tt.args.relativePath); (err != nil) != tt.wantErr {
				t.Errorf("MkdirPath() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
