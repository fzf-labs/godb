package main

import (
	"errors"
	"io"
	"testing"
)

func TestCommandVersionUsesInjectedVersion(t *testing.T) {
	oldVersion := version
	defer func() { version = oldVersion }()

	version = "v1.2.3"
	if got := commandVersion(); got != "v1.2.3" {
		t.Fatalf("got %q want injected version", got)
	}
}

func TestCommandVersionUsesDevFallback(t *testing.T) {
	oldVersion := version
	defer func() { version = oldVersion }()

	version = "dev"
	if got := commandVersion(); got != "dev" {
		t.Fatalf("got %q want dev fallback", got)
	}
}

func TestMainExecutesRootCommand(t *testing.T) {
	rootCmd.SetArgs([]string{"--version"})
	rootCmd.SetOut(io.Discard)
	rootCmd.SetErr(io.Discard)
	defer func() {
		rootCmd.SetArgs(nil)
		rootCmd.SetOut(nil)
		rootCmd.SetErr(nil)
	}()

	main()
}

func TestRunMainReportsExecuteError(t *testing.T) {
	oldFatal := logFatal
	defer func() { logFatal = oldFatal }()

	executeErr := errors.New("execute failed")
	var fatalArgs []any
	logFatal = func(v ...any) {
		fatalArgs = append(fatalArgs, v...)
	}

	runMain(func() error {
		return executeErr
	})

	if len(fatalArgs) != 1 || !errors.Is(fatalArgs[0].(error), executeErr) {
		t.Fatalf("fatal args = %#v, want execute error", fatalArgs)
	}
}
