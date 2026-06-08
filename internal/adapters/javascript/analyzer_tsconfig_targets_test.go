package javascript

import (
	"testing"

	"refactorlah/internal/adapters/javascript/rules"
	"refactorlah/internal/planning"
)

func TestAnalyzerRewritesTypeScriptPathAliasExactTarget(t *testing.T) {
	root := t.TempDir()
	tsconfig := `{
  "compilerOptions": {
    "paths": {
      "@helper": ["src/old-helper.ts"]
    }
  }
}
`
	writeJavaScriptFixture(t, root, "tsconfig.json", tsconfig)
	consumer := `import helper from '@helper';
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
		t.Fatalf("analyze exact typescript path alias: %v", err)
	}
	if !relevant {
		t.Fatal("expected javascript analyzer to be relevant")
	}

	updatedTSConfig := applyJavaScriptReplacements(tsconfig, response.Replacements, "tsconfig.json")
	if updatedTSConfig != `{
  "compilerOptions": {
    "paths": {
      "@helper": ["src/new-helper.ts"]
    }
  }
}
` {
		t.Fatalf("unexpected rewritten tsconfig:\n%s", updatedTSConfig)
	}
	updatedConsumer := applyJavaScriptReplacements(consumer, response.Replacements, "src/consumer.ts")
	if updatedConsumer != consumer {
		t.Fatalf("expected exact alias specifier to stay stable, got:\n%s", updatedConsumer)
	}
	replacement, found := findJavaScriptReplacement(response.Replacements, "tsconfig.json", rules.TypeScriptPathTargetReason)
	if !found {
		t.Fatalf("expected typescript path target replacement, got %#v", response.Replacements)
	}
	if replacement.Adapter != "javascript" || replacement.Rule != rules.TypeScriptPathTargetRuleName {
		t.Fatalf("unexpected replacement metadata %#v", replacement)
	}
}

func TestAnalyzerRewritesJsConfigPathAliasExactTargetWithBaseUrl(t *testing.T) {
	root := t.TempDir()
	jsconfig := `{
  // JSONC comments should not break offset-preserving config updates.
  "compilerOptions": {
    "baseUrl": "src",
    "paths": {
      "@helper": ["./old-helper.js",],
    },
  },
}
`
	writeJavaScriptFixture(t, root, "jsconfig.json", jsconfig)
	consumer := `import helper from '@helper';
`
	writeJavaScriptFixture(t, root, "src/consumer.js", consumer)
	writeJavaScriptFixture(t, root, "src/old-helper.js", "export default function helper() {}\n")

	response, relevant, err := analyzeJavaScript(t, root, planning.MovePlan{
		Moves: []planning.FileMove{{
			OldPath: "src/old-helper.js",
			NewPath: "src/new-helper.js",
		}},
	})
	if err != nil {
		t.Fatalf("analyze exact jsconfig path alias: %v", err)
	}
	if !relevant {
		t.Fatal("expected javascript analyzer to be relevant")
	}

	updatedJSConfig := applyJavaScriptReplacements(jsconfig, response.Replacements, "jsconfig.json")
	if updatedJSConfig != `{
  // JSONC comments should not break offset-preserving config updates.
  "compilerOptions": {
    "baseUrl": "src",
    "paths": {
      "@helper": ["./new-helper.js",],
    },
  },
}
` {
		t.Fatalf("unexpected rewritten jsconfig:\n%s", updatedJSConfig)
	}
}

func TestAnalyzerSkipsAmbiguousTypeScriptPathAliasTargets(t *testing.T) {
	root := t.TempDir()
	tsconfig := `{
  "compilerOptions": {
    "paths": {
      "@helper": ["src/old-helper.ts", "src/fallback-helper.ts"]
    }
  }
}
`
	writeJavaScriptFixture(t, root, "tsconfig.json", tsconfig)
	consumer := `import helper from '@helper';
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
		t.Fatalf("analyze ambiguous typescript path alias: %v", err)
	}
	if !relevant {
		t.Fatal("expected javascript analyzer to be relevant")
	}
	if updatedTSConfig := applyJavaScriptReplacements(tsconfig, response.Replacements, "tsconfig.json"); updatedTSConfig != tsconfig {
		t.Fatalf("expected ambiguous path target to stay unchanged, got:\n%s", updatedTSConfig)
	}
	if _, found := findJavaScriptReplacement(response.Replacements, "tsconfig.json", rules.TypeScriptPathTargetReason); found {
		t.Fatalf("expected ambiguous path target to be skipped, got %#v", response.Replacements)
	}
	if !hasJavaScriptWarning(response.Warnings, "tsconfig.json", `TypeScript path alias "@helper" has multiple targets; skipped conservatively.`) {
		t.Fatalf("expected ambiguous path warning, got %#v", response.Warnings)
	}
}
