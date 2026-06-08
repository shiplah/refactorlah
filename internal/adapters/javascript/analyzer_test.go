package javascript

import (
	"os"
	"path/filepath"
	"sort"
	"testing"

	adapterproto "refactorlah/internal/adapters/contract"
	"refactorlah/internal/adapters/scan"
	"refactorlah/internal/config"
	"refactorlah/internal/planning"
)

func TestAnalyzerRewritesRelativeImportForMovedTypeScriptModule(t *testing.T) {
	root := t.TempDir()
	consumer := `import helper from './old-helper';

export function run() {
    return helper();
}
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
		t.Fatalf("analyze javascript imports: %v", err)
	}
	if !relevant {
		t.Fatal("expected javascript analyzer to be relevant")
	}

	updated := applyJavaScriptReplacements(consumer, response.Replacements, "src/consumer.ts")
	if updated != `import helper from './new-helper';

export function run() {
    return helper();
}
` {
		t.Fatalf("unexpected rewritten consumer:\n%s", updated)
	}
}

func TestAnalyzerRewritesRequireForMovedCommonJSModule(t *testing.T) {
	root := t.TempDir()
	consumer := `const helper = require('./old-helper');

module.exports = helper;
`
	writeJavaScriptFixture(t, root, "src/consumer.cjs", consumer)
	writeJavaScriptFixture(t, root, "src/old-helper.cjs", "module.exports = function helper() {};\n")

	response, relevant, err := analyzeJavaScript(t, root, planning.MovePlan{
		Moves: []planning.FileMove{{
			OldPath: "src/old-helper.cjs",
			NewPath: "src/new-helper.cjs",
		}},
	})
	if err != nil {
		t.Fatalf("analyze javascript requires: %v", err)
	}
	if !relevant {
		t.Fatal("expected javascript analyzer to be relevant")
	}

	updated := applyJavaScriptReplacements(consumer, response.Replacements, "src/consumer.cjs")
	if updated != `const helper = require('./new-helper');

module.exports = helper;
` {
		t.Fatalf("unexpected rewritten consumer:\n%s", updated)
	}
}

func TestAnalyzerRewritesDirectoryImportForMovedIndexModule(t *testing.T) {
	root := t.TempDir()
	consumer := `import { helper } from './old';
`
	writeJavaScriptFixture(t, root, "src/consumer.ts", consumer)
	writeJavaScriptFixture(t, root, "src/old/index.ts", "export const helper = () => 'ok';\n")

	response, relevant, err := analyzeJavaScript(t, root, planning.MovePlan{
		Moves: []planning.FileMove{{
			OldPath: "src/old/index.ts",
			NewPath: "src/new/index.ts",
		}},
	})
	if err != nil {
		t.Fatalf("analyze javascript index imports: %v", err)
	}
	if !relevant {
		t.Fatal("expected javascript analyzer to be relevant")
	}

	updated := applyJavaScriptReplacements(consumer, response.Replacements, "src/consumer.ts")
	if updated != `import { helper } from './new';
` {
		t.Fatalf("unexpected rewritten consumer:\n%s", updated)
	}
}

func TestAnalyzerRewritesTypeScriptPathAliasImport(t *testing.T) {
	root := t.TempDir()
	writeJavaScriptFixture(t, root, "tsconfig.json", `{
  "compilerOptions": {
    "baseUrl": ".",
    "paths": {
      "@/*": ["src/*"]
    }
  }
}
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
		t.Fatalf("analyze typescript path aliases: %v", err)
	}
	if !relevant {
		t.Fatal("expected javascript analyzer to be relevant")
	}

	updated := applyJavaScriptReplacements(consumer, response.Replacements, "src/consumer.ts")
	if updated != `import helper from '@/new-helper';
` {
		t.Fatalf("unexpected rewritten alias import:\n%s", updated)
	}
}

func TestAnalyzerRewritesJsConfigPathAliasDirectoryImport(t *testing.T) {
	root := t.TempDir()
	writeJavaScriptFixture(t, root, "jsconfig.json", `{
  "compilerOptions": {
    "paths": {
      "~/*": ["client/*"]
    }
  }
}
`)
	consumer := `import { helper } from '~/old';
`
	writeJavaScriptFixture(t, root, "client/consumer.js", consumer)
	writeJavaScriptFixture(t, root, "client/old/index.js", "export const helper = () => 'ok';\n")

	response, relevant, err := analyzeJavaScript(t, root, planning.MovePlan{
		Moves: []planning.FileMove{{
			OldPath: "client/old/index.js",
			NewPath: "client/new/index.js",
		}},
	})
	if err != nil {
		t.Fatalf("analyze jsconfig path aliases: %v", err)
	}
	if !relevant {
		t.Fatal("expected javascript analyzer to be relevant")
	}

	updated := applyJavaScriptReplacements(consumer, response.Replacements, "client/consumer.js")
	if updated != `import { helper } from '~/new';
` {
		t.Fatalf("unexpected rewritten alias import:\n%s", updated)
	}
}

func TestAnalyzerSkipsNonJavaScriptMoves(t *testing.T) {
	root := t.TempDir()

	response, relevant, err := analyzeJavaScript(t, root, planning.MovePlan{
		Moves: []planning.FileMove{{
			OldPath: "src/old.css",
			NewPath: "src/new.css",
		}},
	})
	if err != nil {
		t.Fatalf("analyze non-javascript move: %v", err)
	}
	if relevant {
		t.Fatalf("expected analyzer to be irrelevant, got %#v", response)
	}
}

func analyzeJavaScript(t *testing.T, root string, plan planning.MovePlan) (adapterproto.AggregatedResponse, bool, error) {
	t.Helper()
	scanConfig := config.Config{}
	return NewAnalyzer().Analyze(root, plan, scanConfig, scan.NewIndex(root, scanConfig))
}

func writeJavaScriptFixture(t *testing.T, root string, relativePath string, content string) {
	t.Helper()

	absolutePath := filepath.Join(root, filepath.FromSlash(relativePath))
	if err := os.MkdirAll(filepath.Dir(absolutePath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(absolutePath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func applyJavaScriptReplacements(content string, replacements []adapterproto.Replacement, file string) string {
	fileReplacements := make([]adapterproto.Replacement, 0, len(replacements))
	for _, replacement := range replacements {
		if replacement.File == file {
			fileReplacements = append(fileReplacements, replacement)
		}
	}
	sort.Slice(fileReplacements, func(left int, right int) bool {
		return fileReplacements[left].Start > fileReplacements[right].Start
	})

	result := []byte(content)
	for _, replacement := range fileReplacements {
		next := make([]byte, 0, len(result)-replacement.End+replacement.Start+len(replacement.Replacement))
		next = append(next, result[:replacement.Start]...)
		next = append(next, []byte(replacement.Replacement)...)
		next = append(next, result[replacement.End:]...)
		result = next
	}
	return string(result)
}
