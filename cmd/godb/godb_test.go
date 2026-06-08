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
	configureRootCommandForTest(t, "--version")

	main()
}

func TestRootCommandStateIsRestoredBetweenRuns(t *testing.T) {
	t.Run("unknown command", func(t *testing.T) {
		configureRootCommandForTest(t, "definitely-not-a-godb-command")
		if err := rootCmd.Execute(); err == nil {
			t.Fatal("expected unknown command error")
		}
	})

	configureRootCommandForTest(t, "--version")
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("expected second command to ignore stale args, got %v", err)
	}
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

func configureRootCommandForTest(t *testing.T, args ...string) {
	t.Helper()
	rootCmd.SetArgs(args)
	rootCmd.SetOut(io.Discard)
	rootCmd.SetErr(io.Discard)
	t.Cleanup(func() {
		rootCmd.SetArgs(nil)
		rootCmd.SetOut(nil)
		rootCmd.SetErr(nil)
	})
}
