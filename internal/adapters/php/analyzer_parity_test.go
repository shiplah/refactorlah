//go:build cgo

package php

import (
	"sort"
	"strings"
	"testing"

	adapterproto "github.com/shiplah/refactorlah/internal/adapters/contract"
	"github.com/shiplah/refactorlah/internal/planning"
	"github.com/shiplah/refactorlah/internal/testfixtures"
)

func TestAnalyzerKeepsOverlappingPHPRenamesTokenScoped(t *testing.T) {
	const scenario = "overlapping-renames"
	root := analyzerParityRoot(t, scenario)

	response, _, err := analyzePHP(t, root, planning.MovePlan{
		Moves: []planning.FileMove{{
			OldPath: "src/Module/Cache/CacheServiceIndex.php",
			NewPath: "src/Module/Cache/CacheServiceLookup.php",
		}},
	})
	if err != nil {
		t.Fatalf("analyze php: %v", err)
	}

	updatedCacheIndex := applyPHPFixtureReplacements(t, scenario, response.Replacements, "src/Module/Cache/CacheServiceIndex.php")
	assertPHPFixtureEqual(t, scenario, "src/Module/Cache/CacheServiceIndex.php", updatedCacheIndex)

	updatedServices := applyPHPFixtureReplacements(t, scenario, response.Replacements, "services.php")
	assertPHPFixtureEqual(t, scenario, "services.php", updatedServices)
	assertPHPTextNotContains(t, updatedServices, "AssetRegistCache")
	assertPHPTextNotContains(t, updatedServices, "CacheServiceLookupectIndex")
}

func TestAnalyzerUpdatesImportedShortReferencesWhenNamespaceAndBasenameChange(t *testing.T) {
	const scenario = "imported-short-references"
	root := analyzerParityRoot(t, scenario)

	response, _, err := analyzePHP(t, root, planning.MovePlan{
		Moves: []planning.FileMove{{
			OldPath: "src/Module/Ui/Block/PanelBlockRenderer.php",
			NewPath: "src/Module/Ui/Renderer/PanelRenderableRenderer.php",
		}},
	})
	if err != nil {
		t.Fatalf("analyze php: %v", err)
	}

	assertPHPFixtureEqual(t, scenario, "src/Module/Ui/Block/PanelBlockRenderer.php",
		applyPHPFixtureReplacements(t, scenario, response.Replacements, "src/Module/Ui/Block/PanelBlockRenderer.php"))
	assertPHPFixtureEqual(t, scenario, "src/Module/Ui/HtmlRenderer.php",
		applyPHPFixtureReplacements(t, scenario, response.Replacements, "src/Module/Ui/HtmlRenderer.php"))
	assertPHPFixtureEqual(t, scenario, "services.php",
		applyPHPFixtureReplacements(t, scenario, response.Replacements, "services.php"))
	assertPHPFixtureEqual(t, scenario, "tests/Module/PanelRendererTest.php",
		applyPHPFixtureReplacements(t, scenario, response.Replacements, "tests/Module/PanelRendererTest.php"))
}

func TestAnalyzerUpdatesImportedEnumCaseReferencesWhenBasenameChanges(t *testing.T) {
	const scenario = "enum-case-references"
	root := analyzerParityRoot(t, scenario)

	response, _, err := analyzePHP(t, root, planning.MovePlan{
		Moves: []planning.FileMove{{
			OldPath: "platform/src/Module/Application/ElementKind.php",
			NewPath: "platform/src/Module/Application/DirectiveKind.php",
		}},
	})
	if err != nil {
		t.Fatalf("analyze php: %v", err)
	}

	assertPHPFixtureEqual(t, scenario, "platform/src/Module/Application/ElementKind.php",
		applyPHPFixtureReplacements(t, scenario, response.Replacements, "platform/src/Module/Application/ElementKind.php"))
	assertPHPFixtureEqual(t, scenario, "platform/src/Module/Ui/Renderer/PanelRenderer.php",
		applyPHPFixtureReplacements(t, scenario, response.Replacements, "platform/src/Module/Ui/Renderer/PanelRenderer.php"))
}

func TestAnalyzerUpdatesSameNamespaceShortReferencesWhenBasenameChanges(t *testing.T) {
	const scenario = "same-namespace-short-references"
	root := analyzerParityRoot(t, scenario)

	response, _, err := analyzePHP(t, root, planning.MovePlan{
		Moves: []planning.FileMove{{
			OldPath: "src/Module/Renderer/ComponentRenderer.php",
			NewPath: "src/Module/Renderer/DirectiveRenderer.php",
		}},
	})
	if err != nil {
		t.Fatalf("analyze php: %v", err)
	}

	assertPHPFixtureEqual(t, scenario, "src/Module/Renderer/ComponentRenderer.php",
		applyPHPFixtureReplacements(t, scenario, response.Replacements, "src/Module/Renderer/ComponentRenderer.php"))
	updatedConsumer := applyPHPFixtureReplacements(t, scenario, response.Replacements, "src/Module/Renderer/NodeRenderer.php")
	assertPHPFixtureEqual(t, scenario, "src/Module/Renderer/NodeRenderer.php", updatedConsumer)
	assertPHPTextNotContains(t, updatedConsumer, "use App\\Module\\Renderer\\DirectiveRenderer;")
}

func TestAnalyzerUpdatesTraitUseReferencesWhenBasenameChanges(t *testing.T) {
	const scenario = "trait-use-references"
	root := analyzerParityRoot(t, scenario)

	response, _, err := analyzePHP(t, root, planning.MovePlan{
		Moves: []planning.FileMove{{
			OldPath: "src/Services/Items/BuildsItem.php",
			NewPath: "src/Domain/Items/BuildsDomainItem.php",
		}},
	})
	if err != nil {
		t.Fatalf("analyze php: %v", err)
	}

	assertPHPFixtureEqual(t, scenario, "src/Services/Items/BuildsItem.php",
		applyPHPFixtureReplacements(t, scenario, response.Replacements, "src/Services/Items/BuildsItem.php"))
	assertPHPFixtureEqual(t, scenario, "src/Http/ItemController.php",
		applyPHPFixtureReplacements(t, scenario, response.Replacements, "src/Http/ItemController.php"))
}

func TestAnalyzerPreservesAliasedImportStyle(t *testing.T) {
	const scenario = "aliased-import-style"
	root := analyzerParityRoot(t, scenario)

	response, _, err := analyzePHP(t, root, planning.MovePlan{
		Moves: []planning.FileMove{{
			OldPath: "src/Services/Items/ItemService.php",
			NewPath: "src/Domain/Items/DomainItemService.php",
		}},
	})
	if err != nil {
		t.Fatalf("analyze php: %v", err)
	}

	assertPHPFixtureEqual(t, scenario, "src/Http/ItemController.php",
		applyPHPFixtureReplacements(t, scenario, response.Replacements, "src/Http/ItemController.php"))
}

func TestAnalyzerKeepsSameFileHelperClassesNamespaceLocal(t *testing.T) {
	const scenario = "same-file-helper-classes"
	root := analyzerParityRoot(t, scenario)

	response, _, err := analyzePHP(t, root, planning.MovePlan{
		Moves: []planning.FileMove{{
			OldPath: "tests/Application/Feature/RewriteLinksTest.php",
			NewPath: "tests/Feature/Archive/Detailed/Application/RewriteLinksTest.php",
		}},
	})
	if err != nil {
		t.Fatalf("analyze php: %v", err)
	}

	updated := applyPHPFixtureReplacements(t, scenario, response.Replacements, "tests/Application/Feature/RewriteLinksTest.php")
	assertPHPFixtureEqual(t, scenario, "tests/Application/Feature/RewriteLinksTest.php", updated)
	assertPHPTextNotContains(t, updated, "use App\\Tests\\Application\\Feature\\Helper;")
}

func analyzerParityRoot(t *testing.T, scenario string) string {
	t.Helper()

	return testfixtures.CopyDir(t, "tests/fixtures/php-analyzer-parity/"+scenario+"/before")
}

func applyPHPFixtureReplacements(t *testing.T, scenario string, replacements []adapterproto.Replacement, file string) string {
	t.Helper()

	content := string(testfixtures.Read(t, "tests/fixtures/php-analyzer-parity/"+scenario+"/before/"+file))
	return applyPHPAdapterReplacements(content, replacements, file)
}

func assertPHPFixtureEqual(t *testing.T, scenario string, file string, actual string) {
	t.Helper()

	expected := string(testfixtures.Read(t, "tests/fixtures/php-analyzer-parity/"+scenario+"/after/"+file))
	assertPHPTextEqual(t, expected, actual)
}

func applyPHPAdapterReplacements(content string, replacements []adapterproto.Replacement, file string) string {
	fileReplacements := make([]adapterproto.Replacement, 0, len(replacements))
	for _, replacement := range replacements {
		if replacement.File == file {
			fileReplacements = append(fileReplacements, replacement)
		}
	}
	sort.Slice(fileReplacements, func(left int, right int) bool {
		return fileReplacements[left].Start > fileReplacements[right].Start
	})

	result := []byte(content)
	for _, replacement := range fileReplacements {
		next := make([]byte, 0, len(result)-replacement.End+replacement.Start+len(replacement.Replacement))
		next = append(next, result[:replacement.Start]...)
		next = append(next, []byte(replacement.Replacement)...)
		next = append(next, result[replacement.End:]...)
		result = next
	}
	return string(result)
}

func assertPHPTextEqual(t *testing.T, expected string, actual string) {
	t.Helper()

	if actual != expected {
		t.Fatalf("unexpected updated PHP:\n%s\nexpected:\n%s", actual, expected)
	}
}

func assertPHPTextNotContains(t *testing.T, actual string, unexpected string) {
	t.Helper()

	if strings.Contains(actual, unexpected) {
		t.Fatalf("expected PHP not to contain %q, got:\n%s", unexpected, actual)
	}
}
