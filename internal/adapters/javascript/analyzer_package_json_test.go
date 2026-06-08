package javascript

import (
	"testing"

	"refactorlah/internal/adapters/javascript/rules"
	"refactorlah/internal/planning"
)

func TestAnalyzerRewritesPackageImportsAlias(t *testing.T) {
	root := t.TempDir()
	writeJavaScriptFixture(t, root, "package.json", `{
  "imports": {
    "#app/*": "./src/*"
  }
}
`)
	consumer := `import helper from '#app/old-helper';
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
		t.Fatalf("analyze package imports aliases: %v", err)
	}
	if !relevant {
		t.Fatal("expected javascript analyzer to be relevant")
	}

	updated := applyJavaScriptReplacements(consumer, response.Replacements, "src/consumer.ts")
	if updated != `import helper from '#app/new-helper';
` {
		t.Fatalf("unexpected rewritten package import:\n%s", updated)
	}
	replacement, found := findJavaScriptReplacement(response.Replacements, "src/consumer.ts", rules.PackageImportsReason)
	if !found {
		t.Fatalf("expected package imports replacement, got %#v", response.Replacements)
	}
	if replacement.Adapter != "javascript" || replacement.Rule != rules.PackageImportsRuleName {
		t.Fatalf("unexpected replacement metadata %#v", replacement)
	}
}

func TestAnalyzerRewritesPackageImportExactTarget(t *testing.T) {
	root := t.TempDir()
	packageJSON := `{
  "scripts": {
    "demo": "./src/old-helper.js"
  },
  "imports": {
    "#helper": "./src/old-helper.js"
  }
}
`
	writeJavaScriptFixture(t, root, "package.json", packageJSON)
	consumer := `import helper from '#helper';
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
		t.Fatalf("analyze exact package imports: %v", err)
	}
	if !relevant {
		t.Fatal("expected javascript analyzer to be relevant")
	}

	updatedPackageJSON := applyJavaScriptReplacements(packageJSON, response.Replacements, "package.json")
	if updatedPackageJSON != `{
  "scripts": {
    "demo": "./src/old-helper.js"
  },
  "imports": {
    "#helper": "./src/new-helper.js"
  }
}
` {
		t.Fatalf("unexpected rewritten package json:\n%s", updatedPackageJSON)
	}
	updatedConsumer := applyJavaScriptReplacements(consumer, response.Replacements, "src/consumer.js")
	if updatedConsumer != consumer {
		t.Fatalf("expected package import specifier to stay stable, got:\n%s", updatedConsumer)
	}
	replacement, found := findJavaScriptReplacement(response.Replacements, "package.json", rules.PackageImportTargetReason)
	if !found {
		t.Fatalf("expected package import target replacement, got %#v", response.Replacements)
	}
	if replacement.Adapter != "javascript" || replacement.Rule != rules.PackageImportTargetRuleName {
		t.Fatalf("unexpected replacement metadata %#v", replacement)
	}
}

func TestAnalyzerSkipsPackageImportConditions(t *testing.T) {
	root := t.TempDir()
	packageJSON := `{
  "imports": {
    "#app/*": {
      "default": "./src/*"
    }
  }
}
`
	writeJavaScriptFixture(t, root, "package.json", packageJSON)
	consumer := `import helper from '#app/old-helper';
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
		t.Fatalf("analyze package import conditions: %v", err)
	}
	if !relevant {
		t.Fatal("expected javascript analyzer to be relevant")
	}
	if _, found := findJavaScriptReplacement(response.Replacements, "src/consumer.ts", rules.PackageImportsReason); found {
		t.Fatalf("expected conditional package imports to be skipped, got %#v", response.Replacements)
	}
	if updatedPackageJSON := applyJavaScriptReplacements(packageJSON, response.Replacements, "package.json"); updatedPackageJSON != packageJSON {
		t.Fatalf("expected conditional package imports config to stay unchanged, got:\n%s", updatedPackageJSON)
	}
	if !hasJavaScriptWarning(response.Warnings, "package.json", `Package imports entry "#app/*" uses conditional targets; skipped conservatively.`) {
		t.Fatalf("expected conditional package imports warning, got %#v", response.Warnings)
	}
}

func TestAnalyzerRewritesPackageSelfReferenceImport(t *testing.T) {
	root := t.TempDir()
	writeJavaScriptFixture(t, root, "package.json", `{
  "name": "@example/app"
}
`)
	consumer := `import helper from '@example/app/src/old-helper';
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
		t.Fatalf("analyze package self-reference imports: %v", err)
	}
	if !relevant {
		t.Fatal("expected javascript analyzer to be relevant")
	}

	updated := applyJavaScriptReplacements(consumer, response.Replacements, "src/consumer.ts")
	if updated != `import helper from '@example/app/src/new-helper';
` {
		t.Fatalf("unexpected rewritten package self-reference:\n%s", updated)
	}
	replacement, found := findJavaScriptReplacement(response.Replacements, "src/consumer.ts", rules.PackageSelfReferenceReason)
	if !found {
		t.Fatalf("expected package self-reference replacement, got %#v", response.Replacements)
	}
	if replacement.Adapter != "javascript" || replacement.Rule != rules.PackageSelfReferenceRuleName {
		t.Fatalf("unexpected replacement metadata %#v", replacement)
	}
}
