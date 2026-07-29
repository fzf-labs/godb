package sqlbackup

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	databaseMySQL    = "mysql"
	databasePostgres = "postgres"
)

// 以下变量便于单测替换可执行文件查找与外部命令执行。
var (
	lookupExecutable   = exec.LookPath
	execCommandContext = exec.CommandContext
)

// atomicOutput 以临时文件写入备份内容，成功后再原子落到最终路径。
type atomicOutput struct {
	finalPath string   // 最终输出文件路径
	tempPath  string   // 同目录下的临时文件路径
	file      *os.File // 临时文件句柄
	force     bool     // 是否允许覆盖已存在文件
	committed bool     // 是否已成功提交到最终路径
}

// prepareAtomicOutput 创建同目录临时文件用于写入备份；目标已存在且未指定 force 时返回错误。
func prepareAtomicOutput(outputPath string, force bool) (*atomicOutput, error) {
	info, err := os.Lstat(outputPath)
	switch {
	case err == nil:
		if info.IsDir() {
			return nil, fmt.Errorf("output path is a directory: %s", outputPath)
		}
		if !force {
			return nil, fmt.Errorf("output file already exists: %s (use --force to replace it)", outputPath)
		}
	case errors.Is(err, fs.ErrNotExist):
		// 目标不存在，后续创建父目录即可。
	case err != nil:
		return nil, fmt.Errorf("inspect output path: %w", err)
	}

	outputDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(outputDir, 0700); err != nil {
		return nil, fmt.Errorf("create output directory: %w", err)
	}

	file, err := os.CreateTemp(outputDir, "."+filepath.Base(outputPath)+".tmp-*")
	if err != nil {
		return nil, fmt.Errorf("create temporary backup file: %w", err)
	}
	if err := file.Chmod(0600); err != nil {
		_ = file.Close()
		_ = os.Remove(file.Name())
		return nil, fmt.Errorf("set temporary backup file permissions: %w", err)
	}

	return &atomicOutput{
		finalPath: outputPath,
		tempPath:  file.Name(),
		file:      file,
		force:     force,
	}, nil
}

// close 关闭临时文件句柄。
func (o *atomicOutput) close() error {
	if o.file == nil {
		return nil
	}
	err := o.file.Close()
	o.file = nil
	return err
}

// commit 将临时文件原子落到最终路径。
// force 时使用 Rename（Windows 上必要时先删除目标再重试）；非 force 时用 Link 避免覆盖竞态。
func (o *atomicOutput) commit() error {
	if err := o.close(); err != nil {
		return fmt.Errorf("close temporary backup file: %w", err)
	}

	if o.force {
		if err := os.Rename(o.tempPath, o.finalPath); err != nil {
			if runtime.GOOS != "windows" {
				return fmt.Errorf("replace output file: %w", err)
			}
			if removeErr := os.Remove(o.finalPath); removeErr != nil && !errors.Is(removeErr, fs.ErrNotExist) {
				return fmt.Errorf("replace output file: %w", err)
			}
			if retryErr := os.Rename(o.tempPath, o.finalPath); retryErr != nil {
				return fmt.Errorf("replace output file: %w", retryErr)
			}
		}
		o.committed = true
		return nil
	}

	if err := os.Link(o.tempPath, o.finalPath); err != nil {
		if errors.Is(err, fs.ErrExist) {
			return fmt.Errorf("output file already exists: %s (use --force to replace it)", o.finalPath)
		}
		return fmt.Errorf("finalize output file: %w", err)
	}
	if err := os.Remove(o.tempPath); err != nil {
		return fmt.Errorf("remove temporary backup file: %w", err)
	}
	o.committed = true
	return nil
}

// abort 在未提交时清理临时文件；已提交则忽略。
func (o *atomicOutput) abort() {
	if o == nil || o.committed {
		return
	}
	_ = o.close()
	_ = os.Remove(o.tempPath)
}

// runExternalCommand 执行外部备份工具，将 stdout 写入指定 Writer，并返回 stderr 文本。
func runExternalCommand(ctx context.Context, executable string, args []string, env []string, stdout io.Writer) (string, error) {
	cmd := execCommandContext(ctx, executable, args...)
	cmd.Env = env
	cmd.Stdout = stdout

	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err := cmd.Run()
	return stderr.String(), err
}

// formatCommandError 将外部命令失败包装为可读错误，并脱敏 stderr 中的敏感信息。
func formatCommandError(ctx context.Context, tool string, stderr string, err error, secrets ...string) error {
	if ctx != nil && ctx.Err() != nil {
		err = ctx.Err()
	}
	detail := strings.TrimSpace(redactSensitive(stderr, secrets...))
	if detail != "" {
		return fmt.Errorf("%s failed: %w: %s", tool, err, detail)
	}
	return fmt.Errorf("%s failed: %w", tool, err)
}

// redactSensitive 将文本中的敏感片段替换为 [REDACTED]。
func redactSensitive(value string, secrets ...string) string {
	for _, secret := range secrets {
		if secret != "" {
			value = strings.ReplaceAll(value, secret, "[REDACTED]")
		}
	}
	return value
}

// commandEnvironment 基于环境变量副本设置或覆盖指定 key。
func commandEnvironment(base []string, key, value string) []string {
	prefix := key + "="
	env := make([]string, 0, len(base)+1)
	for _, item := range base {
		if !strings.HasPrefix(item, prefix) {
			env = append(env, item)
		}
	}
	if value != "" {
		env = append(env, prefix+value)
	}
	return env
}
