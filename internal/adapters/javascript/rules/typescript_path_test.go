package rules_test

import (
	"path/filepath"
	"testing"

	"refactorlah/internal/adapters/javascript/rules"
	"refactorlah/internal/planning"
)

func TestTypeScriptPathTargetRuleUpdatesExactTargets(t *testing.T) {
	root := t.TempDir()
	content := `{
  "compilerOptions": {
    "baseUrl": "src",
    "paths": {
      "@helper": ["old-helper.ts"]
    }
  }
}
`

	replacements := rules.TypeScriptPathTargetRule{}.Collect(rules.TypeScriptPathTargetInput{
		ProjectRoot: root,
		File:        "tsconfig.json",
		Content:     []byte(content),
		PathBase:    filepath.Join(root, "src"),
		Targets: []rules.TypeScriptPathTarget{{
			Target: "old-helper.ts",
		}},
		Moves: []planning.FileMove{{
			OldPath: "src/old-helper.ts",
			NewPath: "src/new-helper.ts",
		}},
	})

	if len(replacements) != 1 {
		t.Fatalf("expected 1 replacement, got %#v", replacements)
	}
	if replacements[0].Reason != rules.TypeScriptPathTargetReason || replacements[0].Rule != rules.TypeScriptPathTargetRuleName {
		t.Fatalf("unexpected replacement metadata %#v", replacements[0])
	}
	if updated := applyRuleReplacements(content, replacements); updated != `{
  "compilerOptions": {
    "baseUrl": "src",
    "paths": {
      "@helper": ["new-helper.ts"]
    }
  }
}
` {
		t.Fatalf("unexpected rewritten tsconfig:\n%s", updated)
	}
}

func TestTypeScriptPathWarningRuleWarnsOnAmbiguousMovedTarget(t *testing.T) {
	root := t.TempDir()
	warnings := rules.TypeScriptPathWarningRule{}.Collect(rules.TypeScriptPathWarningInput{
		ProjectRoot: root,
		File:        "tsconfig.json",
		PathBase:    root,
		Ambiguities: []rules.TypeScriptPathAmbiguity{{
			Alias:   "@/*",
			Targets: []string{"src/*", "generated/*"},
		}},
		Moves: []planning.FileMove{{
			OldPath: "src/old-helper.ts",
			NewPath: "src/new-helper.ts",
		}},
	})

	if len(warnings) != 1 {
		t.Fatalf("expected ambiguity warning, got %#v", warnings)
	}
	if warnings[0].File != "tsconfig.json" {
		t.Fatalf("unexpected warning file %#v", warnings[0])
	}
}
