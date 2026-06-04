package main

import "testing"

func TestCommandVersionUsesInjectedVersion(t *testing.T) {
	oldVersion := version
	defer func() { version = oldVersion }()

	version = "v1.2.3"
	if got := commandVersion(); got != "v1.2.3" {
		t.Fatalf("got %q want injected version", got)
	}
}
