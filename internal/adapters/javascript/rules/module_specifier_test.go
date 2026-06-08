package rules_test

import (
	"reflect"
	"testing"

	"refactorlah/internal/adapters/javascript/rules"
	"refactorlah/internal/planning"
)

func TestModuleSpecifierRuleCollectsExplicitAndImplicitReferences(t *testing.T) {
	rewrites := rules.ModuleSpecifierRule{}.Collect("src/pages/home.tsx", []planning.FileMove{{
		OldPath: "src/widgets/card/index.tsx",
		NewPath: "src/ui/card/index.tsx",
	}})

	expected := map[string]string{
		"../widgets/card/index.tsx": "../ui/card/index.tsx",
		"../widgets/card":           "../ui/card",
	}
	if len(rewrites) != len(expected) {
		t.Fatalf("expected %d rewrites, got %#v", len(expected), rewrites)
	}
	for _, rewrite := range rewrites {
		if expected[rewrite.OldSpecifier] != rewrite.NewSpecifier {
			t.Fatalf("unexpected rewrite %#v", rewrite)
		}
		if rewrite.Reason != rules.ModuleSpecifierReason || rewrite.Rule != rules.ModuleSpecifierRuleName || rewrite.Adapter != "javascript" {
			t.Fatalf("unexpected rewrite metadata %#v", rewrite)
		}
	}
}

func TestModuleCandidateNeedlesIncludesIndexDirectoryNeedles(t *testing.T) {
	needles := rules.ModuleCandidateNeedles([]planning.FileMove{{
		OldPath: "src/components/card/index.tsx",
		NewPath: "src/ui/card/index.tsx",
	}})

	expected := []string{
		"src/components/card/index.tsx",
		"index.tsx",
		"src/components/card",
		"card",
	}
	if !reflect.DeepEqual(needles, expected) {
		t.Fatalf("expected needles %#v, got %#v", expected, needles)
	}
}
