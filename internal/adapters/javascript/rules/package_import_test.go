package rules_test

import (
	"testing"

	"refactorlah/internal/adapters/javascript/rules"
	"refactorlah/internal/planning"
)

func TestPackageImportAliasRuleCollectsImportsAndSelfReferences(t *testing.T) {
	rewrites := rules.PackageImportAliasRule{}.Collect(
		[]rules.PathAliasMapping{{
			AliasPrefix:  "#internal/",
			TargetPrefix: "src/",
		}},
		rules.PackageSelfReferenceMappings("@example/app"),
		[]planning.FileMove{{
			OldPath: "src/billing/old-helper.ts",
			NewPath: "src/billing/new-helper.ts",
		}},
	)

	expected := map[string]string{
		"#internal/billing/old-helper":        "#internal/billing/new-helper",
		"@example/app/src/billing/old-helper": "@example/app/src/billing/new-helper",
	}
	if len(rewrites) != len(expected) {
		t.Fatalf("expected %d rewrites, got %#v", len(expected), rewrites)
	}
	for _, rewrite := range rewrites {
		if expected[rewrite.OldSpecifier] != rewrite.NewSpecifier {
			t.Fatalf("unexpected rewrite %#v", rewrite)
		}
	}
}

func TestPackageImportTargetRuleUpdatesExactTargets(t *testing.T) {
	content := `{
  "imports": {
    "#helper": "./src/old-helper.js"
  }
}
`
	replacements := rules.PackageImportTargetRule{}.Collect(rules.PackageImportTargetInput{
		File:    "package.json",
		Content: []byte(content),
		Targets: []rules.PackageImportTarget{{
			Target: "./src/old-helper.js",
		}},
		Moves: []planning.FileMove{{
			OldPath: "src/old-helper.js",
			NewPath: "src/new-helper.js",
		}},
	})

	if len(replacements) != 1 {
		t.Fatalf("expected 1 replacement, got %#v", replacements)
	}
	if replacements[0].Reason != rules.PackageImportTargetReason || replacements[0].Rule != rules.PackageImportTargetRuleName {
		t.Fatalf("unexpected replacement metadata %#v", replacements[0])
	}
	if updated := applyRuleReplacements(content, replacements); updated != `{
  "imports": {
    "#helper": "./src/new-helper.js"
  }
}
` {
		t.Fatalf("unexpected rewritten package.json:\n%s", updated)
	}
}
