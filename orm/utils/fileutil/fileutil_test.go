package fileutil

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestFileExists 验证文件存在性判断。
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
				path: "./fileutil.go",
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

func TestExistsMissingPath(t *testing.T) {
	if Exists(filepath.Join(t.TempDir(), "missing")) {
		t.Fatal("missing path should not exist")
	}
}

func TestExistsRejectsNonDirectoryLookupErrors(t *testing.T) {
	parentFile := filepath.Join(t.TempDir(), "parent")
	if err := os.WriteFile(parentFile, []byte("file"), 0600); err != nil {
		t.Fatal(err)
	}
	if Exists(filepath.Join(parentFile, "child")) {
		t.Fatal("child path under a regular file should not count as existing")
	}
}

// TestMkdirPath 验证目录创建工具函数。
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

func TestWriteContentCover(t *testing.T) {
	file := filepath.Join(t.TempDir(), "nested", "file.txt")
	if err := WriteContentCover(file, "first"); err != nil {
		t.Fatalf("write content: %v", err)
	}
	content, err := os.ReadFile(file)
	if err != nil {
		t.Fatalf("read content: %v", err)
	}
	if string(content) != "first" {
		t.Fatalf("unexpected content: %q", string(content))
	}
	if err := WriteContentCover(file, "second"); err != nil {
		t.Fatalf("overwrite content: %v", err)
	}
	content, err = os.ReadFile(file)
	if err != nil {
		t.Fatalf("read overwritten content: %v", err)
	}
	if string(content) != "second" {
		t.Fatalf("unexpected overwritten content: %q", string(content))
	}
}

func TestWriteContentCoverReturnsPathErrors(t *testing.T) {
	parentFile := filepath.Join(t.TempDir(), "parent")
	if err := os.WriteFile(parentFile, []byte("file"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := WriteContentCover(filepath.Join(parentFile, "child.txt"), "content"); err == nil {
		t.Fatal("expected mkdir error when parent is a file")
	}

	dirPath := filepath.Join(t.TempDir(), "dir-as-file")
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		t.Fatal(err)
	}
	if err := WriteContentCover(dirPath, "content"); err == nil {
		t.Fatal("expected open error when target is a directory")
	}
}

func TestFillModelPkgPath(t *testing.T) {
	if got := FillModelPkgPath("."); !strings.HasSuffix(got, "/orm/utils/fileutil") {
		t.Fatalf("unexpected package path: %s", got)
	}
	if got := FillModelPkgPath(filepath.Join(t.TempDir(), "missing")); got != "" {
		t.Fatalf("missing package should return empty path, got %s", got)
	}
	if got := FillModelPkgPath(t.TempDir()); got != "" {
		t.Fatalf("empty package dir should return empty path, got %s", got)
	}
}

func TestFillModelPkgPathRejectsEmptyResolvedPackage(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/test\n\ngo 1.24\n"), 0600); err != nil {
		t.Fatal(err)
	}
	emptyPkgDir := filepath.Join(dir, "dao")
	if err := os.MkdirAll(emptyPkgDir, 0755); err != nil {
		t.Fatal(err)
	}
	if got := FillModelPkgPath(emptyPkgDir); got != "" {
		t.Fatalf("empty package dir should not resolve to %q", got)
	}
}
