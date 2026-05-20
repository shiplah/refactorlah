package adapters

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDiscoveryFindsProjectLocalPHPAdapter(t *testing.T) {
	root := t.TempDir()
	adapterPath := filepath.Join(root, "adapters", "php", "bin", "refactorlah-php")
	if err := os.MkdirAll(filepath.Dir(adapterPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(adapterPath, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}

	discovery := NewDiscovery()
	foundPath, ok := discovery.FindPHPAdapter(root)
	if !ok {
		t.Fatal("expected adapter to be found")
	}
	if foundPath != adapterPath {
		t.Fatalf("expected %s, got %s", adapterPath, foundPath)
	}
}
