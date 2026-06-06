package php

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadComposerPsr4MapNormalisesComposerRelativePaths(t *testing.T) {
	root := t.TempDir()
	composerRoot := filepath.Join(root, "platform")
	if err := os.MkdirAll(composerRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(composerRoot, "composer.json"), []byte(`{
		"autoload": {"psr-4": {"App\\": "src/"}},
		"autoload-dev": {"psr-4": {"App\\Tests\\": ["tests/"]}}
	}`), 0o644); err != nil {
		t.Fatal(err)
	}

	psr4, err := ReadComposerPsr4Map(root, composerRoot)
	if err != nil {
		t.Fatalf("read psr4 map: %v", err)
	}

	resolved, ok := Psr4NamespaceResolver{}.DeriveSymbol(psr4, "platform/src/Domain/Invoice.php")
	if !ok {
		t.Fatal("expected app symbol to resolve")
	}
	if resolved.Symbol != "App\\Domain\\Invoice" {
		t.Fatalf("expected App\\Domain\\Invoice, got %q", resolved.Symbol)
	}

	resolved, ok = Psr4NamespaceResolver{}.DeriveSymbol(psr4, "platform/tests/Feature/InvoiceTest.php")
	if !ok {
		t.Fatal("expected test symbol to resolve")
	}
	if resolved.Symbol != "App\\Tests\\Feature\\InvoiceTest" {
		t.Fatalf("expected App\\Tests\\Feature\\InvoiceTest, got %q", resolved.Symbol)
	}
}
