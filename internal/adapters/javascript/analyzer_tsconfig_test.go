package javascript

import (
	"testing"

	"github.com/NickSdot/refactorlah/internal/adapters/javascript/rules"
	"github.com/NickSdot/refactorlah/internal/planning"
)

func TestAnalyzerRewritesTypeScriptPathAliasImport(t *testing.T) {
	root := t.TempDir()
	writeJavaScriptFixture(t, root, "tsconfig.json", `{
  // Real tsconfig files commonly use JSONC.
  "compilerOptions": {
    "baseUrl": ".",
    "paths": {
      "@/*": ["src/*"],
    },
  },
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
	replacement, found := findJavaScriptReplacement(response.Replacements, "src/consumer.ts", rules.TypeScriptPathAliasReason)
	if !found {
		t.Fatalf("expected typescript path alias replacement, got %#v", response.Replacements)
	}
	if replacement.Adapter != "javascript" || replacement.Rule != rules.TypeScriptPathAliasRuleName {
		t.Fatalf("unexpected replacement metadata %#v", replacement)
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
