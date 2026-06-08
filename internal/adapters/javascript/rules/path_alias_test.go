package rules_test

import (
	"testing"

	"refactorlah/internal/adapters/javascript/rules"
	"refactorlah/internal/planning"
)

func TestPathAliasSpecifierRuleRewritesModulesWithinConfiguredTarget(t *testing.T) {
	rewrites := rules.PathAliasSpecifierRule{
		Reason: "javascript-test-alias",
		Rule:   "javascript.TestAliasRule",
	}.Collect([]rules.PathAliasMapping{{
		AliasPrefix:  "@/",
		TargetPrefix: "src/",
	}}, []planning.FileMove{{
		OldPath: "src/billing/old-helper.ts",
		NewPath: "src/billing/new-helper.ts",
	}})

	if len(rewrites) != 1 {
		t.Fatalf("expected 1 rewrite, got %#v", rewrites)
	}
	rewrite := rewrites[0]
	if rewrite.OldSpecifier != "@/billing/old-helper" || rewrite.NewSpecifier != "@/billing/new-helper" {
		t.Fatalf("unexpected rewrite %#v", rewrite)
	}
	if rewrite.Reason != "javascript-test-alias" || rewrite.Rule != "javascript.TestAliasRule" || rewrite.Adapter != "javascript" {
		t.Fatalf("unexpected rewrite metadata %#v", rewrite)
	}
}

func TestPathAliasSpecifierRuleSkipsConflictingAliases(t *testing.T) {
	rewrites := rules.PathAliasSpecifierRule{
		Reason: "javascript-test-alias",
		Rule:   "javascript.TestAliasRule",
	}.Collect([]rules.PathAliasMapping{
		{AliasPrefix: "@/", TargetPrefix: "src/"},
		{AliasPrefix: "@/", TargetPrefix: "client/"},
	}, []planning.FileMove{
		{OldPath: "src/shared/helper.ts", NewPath: "src/shared/new-helper.ts"},
		{OldPath: "client/shared/helper.ts", NewPath: "client/shared/other-helper.ts"},
	})

	if len(rewrites) != 0 {
		t.Fatalf("expected conflicting alias to be skipped, got %#v", rewrites)
	}
}

func TestPathAliasSpecifierRuleCollapsesIndexTargets(t *testing.T) {
	rewrites := rules.PathAliasSpecifierRule{
		Reason: "javascript-test-alias",
		Rule:   "javascript.TestAliasRule",
	}.Collect([]rules.PathAliasMapping{{
		AliasPrefix:  "@features/",
		TargetPrefix: "src/features/",
	}}, []planning.FileMove{{
		OldPath: "src/features/checkout/index.ts",
		NewPath: "src/features/billing/index.ts",
	}})

	if len(rewrites) != 1 {
		t.Fatalf("expected 1 rewrite, got %#v", rewrites)
	}
	if rewrites[0].OldSpecifier != "@features/checkout" || rewrites[0].NewSpecifier != "@features/billing" {
		t.Fatalf("unexpected index alias rewrite %#v", rewrites[0])
	}
}

func TestPathAliasSpecifierRuleSkipsMovesLeavingConfiguredTarget(t *testing.T) {
	rewrites := rules.PathAliasSpecifierRule{
		Reason: "javascript-test-alias",
		Rule:   "javascript.TestAliasRule",
	}.Collect([]rules.PathAliasMapping{{
		AliasPrefix:  "@/internal/",
		TargetPrefix: "src/internal/",
	}}, []planning.FileMove{{
		OldPath: "src/internal/helper.ts",
		NewPath: "src/public/helper.ts",
	}})

	if len(rewrites) != 0 {
		t.Fatalf("expected alias move leaving configured target to be skipped, got %#v", rewrites)
	}
}
