//go:build cgo

package javascript

import (
	"path/filepath"
	"strconv"
	"testing"

	"refactorlah/internal/planning"
)

func TestAnalyzerRewritesViteAliasObjectConfig(t *testing.T) {
	root := t.TempDir()
	aliasTarget := filepath.ToSlash(filepath.Join(root, "src"))
	writeJavaScriptFixture(t, root, "vite.config.js", `import { defineConfig } from 'vite';

export default defineConfig({
  resolve: {
    alias: {
      '@': `+strconv.Quote(aliasTarget)+`,
    },
  },
});
`)
	consumer := `import helper from '@/old-helper';
`
	writeJavaScriptFixture(t, root, "src/consumer.ts", consumer)
	writeJavaScriptFixture(t, root, "src/old-helper.ts", "export default function helper() {}\n")

	response, relevant, err := analyzeJavaScript(t, root, planning.MovePlan{
		Moves: []planning.FileMove{{
			OldPath: "src/old-helper.ts",
			NewPath: "src/new-helper.ts",
		}},
	})
	if err != nil {
		t.Fatalf("analyze vite alias object config: %v", err)
	}
	if !relevant {
		t.Fatal("expected javascript analyzer to be relevant")
	}

	updated := applyJavaScriptReplacements(consumer, response.Replacements, "src/consumer.ts")
	if updated != `import helper from '@/new-helper';
` {
		t.Fatalf("unexpected rewritten vite alias import:\n%s", updated)
	}
	replacement, found := findJavaScriptReplacement(response.Replacements, "src/consumer.ts", viteAliasReason)
	if !found {
		t.Fatalf("expected vite alias replacement, got %#v", response.Replacements)
	}
	if replacement.Adapter != "javascript" || replacement.Rule != viteAliasRule {
		t.Fatalf("unexpected replacement metadata %#v", replacement)
	}
}

func TestAnalyzerRewritesViteAliasArrayConfig(t *testing.T) {
	root := t.TempDir()
	aliasTarget := filepath.ToSlash(filepath.Join(root, "lib"))
	writeJavaScriptFixture(t, root, "vite.config.ts", `export default {
  resolve: {
    alias: [
      { find: '@lib', replacement: `+strconv.Quote(aliasTarget)+` },
    ],
  },
};
`)
	consumer := `import helper from '@lib/old-helper';
`
	writeJavaScriptFixture(t, root, "src/consumer.ts", consumer)
	writeJavaScriptFixture(t, root, "lib/old-helper.ts", "export default function helper() {}\n")

	response, relevant, err := analyzeJavaScript(t, root, planning.MovePlan{
		Moves: []planning.FileMove{{
			OldPath: "lib/old-helper.ts",
			NewPath: "lib/new-helper.ts",
		}},
	})
	if err != nil {
		t.Fatalf("analyze vite alias array config: %v", err)
	}
	if !relevant {
		t.Fatal("expected javascript analyzer to be relevant")
	}

	updated := applyJavaScriptReplacements(consumer, response.Replacements, "src/consumer.ts")
	if updated != `import helper from '@lib/new-helper';
` {
		t.Fatalf("unexpected rewritten vite array alias import:\n%s", updated)
	}
}

func TestAnalyzerRewritesWebpackAliasObjectConfig(t *testing.T) {
	root := t.TempDir()
	aliasTarget := filepath.ToSlash(filepath.Join(root, "client"))
	writeJavaScriptFixture(t, root, "webpack.config.cjs", `module.exports = {
  resolve: {
    alias: {
      '@client': `+strconv.Quote(aliasTarget)+`,
    },
  },
};
`)
	consumer := `const helper = require('@client/old-helper');
`
	writeJavaScriptFixture(t, root, "src/consumer.cjs", consumer)
	writeJavaScriptFixture(t, root, "client/old-helper.cjs", "module.exports = function helper() {};\n")

	response, relevant, err := analyzeJavaScript(t, root, planning.MovePlan{
		Moves: []planning.FileMove{{
			OldPath: "client/old-helper.cjs",
			NewPath: "client/new-helper.cjs",
		}},
	})
	if err != nil {
		t.Fatalf("analyze webpack alias object config: %v", err)
	}
	if !relevant {
		t.Fatal("expected javascript analyzer to be relevant")
	}

	updated := applyJavaScriptReplacements(consumer, response.Replacements, "src/consumer.cjs")
	if updated != `const helper = require('@client/new-helper');
` {
		t.Fatalf("unexpected rewritten webpack alias import:\n%s", updated)
	}
	replacement, found := findJavaScriptReplacement(response.Replacements, "src/consumer.cjs", webpackAliasReason)
	if !found {
		t.Fatalf("expected webpack alias replacement, got %#v", response.Replacements)
	}
	if replacement.Adapter != "javascript" || replacement.Rule != webpackAliasRule {
		t.Fatalf("unexpected replacement metadata %#v", replacement)
	}
}
