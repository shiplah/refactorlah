package php

import (
	"path/filepath"
	"testing"

	"github.com/shiplah/refactorlah/internal/testfixtures"
)

func TestReadComposerPsr4MapNormalisesComposerRelativePaths(t *testing.T) {
	root := testfixtures.CopyDir(t, "tests/fixtures/php-composer/relative-paths")
	composerRoot := filepath.Join(root, "platform")

	psr4, err := ReadComposerPsr4Map(root, composerRoot)
	if err != nil {
		t.Fatalf("read psr4 map: %v", err)
	}

	resolved, ok := Psr4NamespaceResolver{}.DeriveSymbol(psr4, "platform/src/Domain/Item.php")
	if !ok {
		t.Fatal("expected app symbol to resolve")
	}
	if resolved.Symbol != "App\\Domain\\Item" {
		t.Fatalf("expected App\\Domain\\Item, got %q", resolved.Symbol)
	}

	resolved, ok = Psr4NamespaceResolver{}.DeriveSymbol(psr4, "platform/tests/Feature/ItemTest.php")
	if !ok {
		t.Fatal("expected test symbol to resolve")
	}
	if resolved.Symbol != "App\\Tests\\Feature\\ItemTest" {
		t.Fatalf("expected App\\Tests\\Feature\\ItemTest, got %q", resolved.Symbol)
	}
}
