//go:build cgo

package javascript

import (
	"path/filepath"
	"strconv"
	"testing"

	"github.com/shiplah/refactorlah/internal/adapters/javascript/rules"
	"github.com/shiplah/refactorlah/internal/planning"
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
	replacement, found := findJavaScriptReplacement(response.Replacements, "src/consumer.ts", rules.ViteAliasReason)
	if !found {
		t.Fatalf("expected vite alias replacement, got %#v", response.Replacements)
	}
	if replacement.Adapter != "javascript" || replacement.Rule != rules.ViteAliasRuleName {
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
	replacement, found := findJavaScriptReplacement(response.Replacements, "src/consumer.cjs", rules.WebpackAliasReason)
	if !found {
		t.Fatalf("expected webpack alias replacement, got %#v", response.Replacements)
	}
	if replacement.Adapter != "javascript" || replacement.Rule != rules.WebpackAliasRuleName {
		t.Fatalf("unexpected replacement metadata %#v", replacement)
	}
}

func TestAnalyzerRewritesMultipleBundlerAliasesInApplicationCorpus(t *testing.T) {
	root := t.TempDir()
	srcTarget := filepath.ToSlash(filepath.Join(root, "src"))
	legacyTarget := filepath.ToSlash(filepath.Join(root, "legacy"))
	writeJavaScriptFixture(t, root, "vite.config.js", `export default {
  resolve: {
    alias: {
      '@app': `+strconv.Quote(srcTarget)+`,
    },
  },
};
`)
	writeJavaScriptFixture(t, root, "webpack.config.cjs", `module.exports = {
  resolve: {
    alias: {
      '@legacy': `+strconv.Quote(legacyTarget)+`,
    },
  },
};
`)
	consumer := `import helper from '@app/features/old-helper';
const legacy = require('@legacy/old-helper');
`
	writeJavaScriptFixture(t, root, "src/pages/home.ts", consumer)
	writeJavaScriptFixture(t, root, "src/features/old-helper.ts", "export default 'helper';\n")
	writeJavaScriptFixture(t, root, "legacy/old-helper.cjs", "module.exports = 'legacy';\n")

	response, relevant, err := analyzeJavaScript(t, root, planning.MovePlan{
		Moves: []planning.FileMove{
			{OldPath: "src/features/old-helper.ts", NewPath: "src/features/new-helper.ts"},
			{OldPath: "legacy/old-helper.cjs", NewPath: "legacy/new-helper.cjs"},
		},
	})
	if err != nil {
		t.Fatalf("analyze mixed bundler alias corpus: %v", err)
	}
	if !relevant {
		t.Fatal("expected javascript analyzer to be relevant")
	}

	updated := applyJavaScriptReplacements(consumer, response.Replacements, "src/pages/home.ts")
	if updated != `import helper from '@app/features/new-helper';
const legacy = require('@legacy/new-helper');
` {
		t.Fatalf("unexpected rewritten bundler corpus:\n%s", updated)
	}
	assertJavaScriptReplacement(t, response.Replacements, "src/pages/home.ts", rules.ViteAliasReason)
	assertJavaScriptReplacement(t, response.Replacements, "src/pages/home.ts", rules.WebpackAliasReason)
}

func TestAnalyzerSkipsUnsupportedBundlerAliasShapes(t *testing.T) {
	root := t.TempDir()
	srcTarget := filepath.ToSlash(filepath.Join(root, "src"))
	writeJavaScriptFixture(t, root, "vite.config.js", `export default {
  resolve: {
    alias: {
      '@relative': './src',
      '@wild*': `+strconv.Quote(srcTarget)+`,
    },
  },
};
`)
	consumer := `import relative from '@relative/old-helper';
import wildcard from '@wild/old-helper';
`
	writeJavaScriptFixture(t, root, "src/pages/home.ts", consumer)
	writeJavaScriptFixture(t, root, "src/old-helper.ts", "export default 'old';\n")

	response, relevant, err := analyzeJavaScript(t, root, planning.MovePlan{
		Moves: []planning.FileMove{{
			OldPath: "src/old-helper.ts",
			NewPath: "src/new-helper.ts",
		}},
	})
	if err != nil {
		t.Fatalf("analyze unsupported bundler alias shapes: %v", err)
	}
	if !relevant {
		t.Fatal("expected javascript analyzer to be relevant")
	}
	if updated := applyJavaScriptReplacements(consumer, response.Replacements, "src/pages/home.ts"); updated != consumer {
		t.Fatalf("expected unsupported bundler aliases to stay unchanged, got:\n%s", updated)
	}
	if _, found := findJavaScriptReplacement(response.Replacements, "src/pages/home.ts", rules.ViteAliasReason); found {
		t.Fatalf("expected no vite alias replacement for unsupported aliases, got %#v", response.Replacements)
	}
}

func TestAnalyzerWarnsWhenBundlerConfigCannotBeParsed(t *testing.T) {
	root := t.TempDir()
	writeJavaScriptFixture(t, root, "vite.config.js", `export default { resolve: { alias: { '@': ; } } };`)
	writeJavaScriptFixture(t, root, "src/consumer.ts", "import helper from '@/old-helper';\n")
	writeJavaScriptFixture(t, root, "src/old-helper.ts", "export default 'old';\n")

	response, relevant, err := analyzeJavaScript(t, root, planning.MovePlan{
		Moves: []planning.FileMove{{
			OldPath: "src/old-helper.ts",
			NewPath: "src/new-helper.ts",
		}},
	})
	if err != nil {
		t.Fatalf("analyze invalid bundler config: %v", err)
	}
	if !relevant {
		t.Fatal("expected javascript analyzer to be relevant")
	}
	if !hasJavaScriptWarning(response.Warnings, "vite.config.js", "Vite config could not be parsed; alias rewrites skipped.") {
		t.Fatalf("expected vite parse warning, got %#v", response.Warnings)
	}
}
