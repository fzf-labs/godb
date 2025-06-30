package fileutil

import (
	"os"
	"path/filepath"

	"golang.org/x/tools/go/packages"
)

// FillModelPkgPath 返回模型文件的包路径
func FillModelPkgPath(dir string) string {
	pkg, err := packages.Load(&packages.Config{
		Mode: packages.NeedName,
		Dir:  dir,
	})
	if err != nil {
		return ""
	}
	if len(pkg) > 0 {
		return pkg[0].PkgPath
	}
	return ""
}

// Exists 判断文件是否存在
func Exists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// MkdirPath 生成文件夹
func MkdirPath(relativePath string) error {
	return os.MkdirAll(relativePath, os.ModePerm)
}

// WriteContentCover 数据写入，不存在则创建
func WriteContentCover(filePath, content string) error {
	fileDir := filepath.Dir(filePath)
	if err := os.MkdirAll(fileDir, 0775); err != nil {
		return err
	}
	dstFile, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0665)
	if err != nil {
		return err
	}
	defer dstFile.Close()
	_, err = dstFile.WriteString(content)
	if err != nil {
		return err
	}
	return err
}
