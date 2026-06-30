package main

import (
	"context"
	"strings"
	"testing"
)

func TestRunReturnsConfigError(t *testing.T) {
	err := run(context.Background(), []string{"server", "-config", "missing.yaml"})
	if err == nil {
		t.Fatalf("run() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "load config") {
		t.Fatalf("run() error = %q, want load config error", err.Error())
	}
}
