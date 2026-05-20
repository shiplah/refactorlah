package adapters

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"refactorlah/internal/planning"
)

func TestInvokerAcceptsValidAdapterResponse(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell script helper is unix-only")
	}

	root := t.TempDir()
	script := filepath.Join(root, "fake-adapter")
	if err := os.WriteFile(script, []byte("#!/bin/sh\ncat >/dev/null\nprintf '%s' '{\"protocolVersion\":1,\"adapter\":\"php\",\"symbolMappings\":[],\"pathMappings\":[],\"replacements\":[],\"warnings\":[],\"errors\":[]}'\n"), 0o755); err != nil {
		t.Fatal(err)
	}

	invoker := NewInvoker()
	selection := Selection{Adapters: []Config{{Name: "php", Path: script}}}
	_, err := invoker.Invoke(t.Context(), root, planning.MovePlan{}, true, selection)
	if err != nil {
		t.Fatalf("invoke failed: %v", err)
	}
}

func TestInvokerRejectsInvalidJSON(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell script helper is unix-only")
	}

	root := t.TempDir()
	script := filepath.Join(root, "fake-adapter")
	if err := os.WriteFile(script, []byte("#!/bin/sh\ncat >/dev/null\nprintf '%s' 'not-json'\n"), 0o755); err != nil {
		t.Fatal(err)
	}

	invoker := NewInvoker()
	selection := Selection{Adapters: []Config{{Name: "php", Path: script}}}
	if _, err := invoker.Invoke(t.Context(), root, planning.MovePlan{}, true, selection); err == nil {
		t.Fatal("expected invalid JSON error")
	}
}
