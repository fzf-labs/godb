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

var (
	lookupExecutable   = exec.LookPath
	execCommandContext = exec.CommandContext
)

type sqlBackup struct {
	db     string
	dsn    string
	output string
	force  bool
}

type atomicOutput struct {
	finalPath string
	tempPath  string
	file      *os.File
	force     bool
	committed bool
}

func newSQLBackup(opts runOptions) *sqlBackup {
	return &sqlBackup{
		db:     opts.db,
		dsn:    opts.dsn,
		output: filepath.Clean(opts.output),
		force:  opts.force,
	}
}

func (b *sqlBackup) run(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}
	switch b.db {
	case databasePostgres:
		return b.runPostgres(ctx)
	case databaseMySQL:
		return b.runMySQL(ctx)
	default:
		return fmt.Errorf("unknown database type: %s", b.db)
	}
}

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
		// The parent directory is created below.
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

func (o *atomicOutput) writer() (*os.File, error) {
	if o.file == nil {
		return nil, fmt.Errorf("temporary backup file is closed")
	}
	return o.file, nil
}

func (o *atomicOutput) close() error {
	if o.file == nil {
		return nil
	}
	err := o.file.Close()
	o.file = nil
	return err
}

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

func (o *atomicOutput) abort() {
	if o == nil || o.committed {
		return
	}
	_ = o.close()
	_ = os.Remove(o.tempPath)
}

func runExternalCommand(ctx context.Context, executable string, args []string, env []string, stdout io.Writer) (string, error) {
	cmd := execCommandContext(ctx, executable, args...)
	cmd.Env = env
	cmd.Stdout = stdout

	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err := cmd.Run()
	return stderr.String(), err
}

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

func redactSensitive(value string, secrets ...string) string {
	for _, secret := range secrets {
		if secret != "" {
			value = strings.ReplaceAll(value, secret, "[REDACTED]")
		}
	}
	return value
}

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
