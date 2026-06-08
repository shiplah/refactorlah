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
