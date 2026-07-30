package rules_test

import (
	"testing"

	"github.com/shiplah/refactorlah/internal/adapters/javascript/rules"
	"github.com/shiplah/refactorlah/internal/planning"
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
	expected := "{\n  \"imports\": {\n    \"#helper\": \"./src/new-helper.js\"\n  }\n}\n"
	if updated := applyRuleReplacements(content, replacements); updated != expected {
		t.Fatalf("unexpected rewritten package.json:\n%s", updated)
	}
}

func TestPackageImportTargetRuleIgnoresNonJavaScriptTargets(t *testing.T) {
	content := `{
  "imports": {
    "#styles": "./src/old.css"
  }
}
`
	replacements := rules.PackageImportTargetRule{}.Collect(rules.PackageImportTargetInput{
		File:    "package.json",
		Content: []byte(content),
		Targets: []rules.PackageImportTarget{{
			Target: "./src/old.css",
		}},
		Moves: []planning.FileMove{{
			OldPath: "src/old.css",
			NewPath: "src/new.css",
		}},
	})

	if len(replacements) != 0 {
		t.Fatalf("expected non-javascript package target to be skipped, got %#v", replacements)
	}
}

func TestPackageImportWarningRuleWarnsForNestedConditionalArrayTargets(t *testing.T) {
	warnings := rules.PackageImportWarningRule{}.Collect(rules.PackageImportWarningInput{
		File: "package.json",
		ConditionalImports: []rules.PackageConditionalImport{{
			Key:     "#feature",
			Targets: []string{"./src/old-helper.ts", "./dist/old-helper.js"},
		}},
		Moves: []planning.FileMove{{
			OldPath: "src/old-helper.ts",
			NewPath: "src/new-helper.ts",
		}},
	})

	if len(warnings) != 1 {
		t.Fatalf("expected conditional warning, got %#v", warnings)
	}
	if warnings[0].Message != `Package imports entry "#feature" uses conditional targets; skipped conservatively.` {
		t.Fatalf("unexpected conditional warning %#v", warnings[0])
	}
}

func TestPackageSelfReferenceMappingsRejectsInvalidNames(t *testing.T) {
	for _, packageName := range []string{"", "./local", "/absolute", "bad name", "@scope", "@scope/"} {
		if mappings := rules.PackageSelfReferenceMappings(packageName); len(mappings) != 0 {
			t.Fatalf("expected invalid package name %q to be rejected, got %#v", packageName, mappings)
		}
	}
}
