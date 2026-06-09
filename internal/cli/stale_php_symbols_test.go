package cli

import (
	"path/filepath"
	"strings"
	"testing"

	adapterproto "github.com/shiplah/refactorlah/internal/adapters/contract"
	"github.com/shiplah/refactorlah/internal/config"
)

func TestStalePHPSymbolValidationFindsCodeReferencesButSkipsStringsAndComments(t *testing.T) {
	root := t.TempDir()
	mustWriteFile(t, filepath.Join(root, "src", "Consumer.php"), `<?php
namespace App\Example;

use App\Old\Thing;

final class Consumer
{
    public function build(): \App\Old\Thing
    {
        return new \App\Old\Thing();
    }
}
`)
	mustWriteFile(t, filepath.Join(root, "src", "StringLiteral.php"), `<?php
return 'App\Old\Thing is mentioned in a diagnostic string';
`)
	mustWriteFile(t, filepath.Join(root, "src", "CommentOnly.php"), `<?php
// App\Old\Thing is mentioned in a comment.
`)
	mustWriteFile(t, filepath.Join(root, "fixtures", "Excluded.php"), `<?php
use App\Old\Thing;
`)

	result, err := stalePHPSymbolValidation(root, config.Config{
		Exclude: []string{"fixtures/**"},
	}, []adapterproto.SymbolMapping{{
		Kind:      "class",
		OldSymbol: "App\\Old\\Thing",
		NewSymbol: "App\\New\\Thing",
	}})

	if err == nil {
		t.Fatal("expected stale PHP symbol validation to fail")
	}
	if result.Name != "stale PHP symbol scan" || result.Status != "failed" {
		t.Fatalf("unexpected result: %#v", result)
	}
	if !strings.Contains(result.Stdout, "src/Consumer.php:4 App\\Old\\Thing") {
		t.Fatalf("expected stale use statement hit, got:\n%s", result.Stdout)
	}
	if !strings.Contains(result.Stdout, "src/Consumer.php:8 App\\Old\\Thing") {
		t.Fatalf("expected stale FQCN return type hit, got:\n%s", result.Stdout)
	}
	for _, unexpected := range []string{"StringLiteral.php", "CommentOnly.php", "Excluded.php"} {
		if strings.Contains(result.Stdout, unexpected) {
			t.Fatalf("did not expect %s in stale symbol output:\n%s", unexpected, result.Stdout)
		}
	}
}
