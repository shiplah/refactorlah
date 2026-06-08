package javascript

import (
	"testing"

	"refactorlah/internal/adapters/javascript/rules"
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
	replacement, found := findJavaScriptReplacement(response.Replacements, "src/consumer.ts", rules.ModuleSpecifierReason)
	if !found {
		t.Fatalf("expected javascript module replacement, got %#v", response.Replacements)
	}
	if replacement.Adapter != "javascript" || replacement.Rule != rules.ModuleSpecifierRuleName {
		t.Fatalf("unexpected replacement metadata %#v", replacement)
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

func TestModuleCandidateNeedlesIncludeExtensionlessModuleNames(t *testing.T) {
	needles := rules.ModuleCandidateNeedles([]planning.FileMove{{
		OldPath: "src/old/index.ts",
		NewPath: "src/new/index.ts",
	}})

	for _, expected := range []string{"src/old/index.ts", "index.ts", "src/old", "old"} {
		if !containsJavaScriptString(needles, expected) {
			t.Fatalf("expected module needle %q in %#v", expected, needles)
		}
	}
}
