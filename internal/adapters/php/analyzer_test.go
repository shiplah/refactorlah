//go:build cgo

package php

import (
	"strings"
	"testing"

	adapterproto "github.com/shiplah/refactorlah/internal/adapters/contract"
	"github.com/shiplah/refactorlah/internal/config"
	"github.com/shiplah/refactorlah/internal/planning"
	"github.com/shiplah/refactorlah/internal/testfixtures"
)

func TestAnalyzerUpdatesNamespaceDeclarationAndUseStatement(t *testing.T) {
	root := testfixtures.CopyDir(t, "tests/fixtures/php-analyzer-basic/namespace-use")

	response, relevant, err := analyzePHP(t, root, planning.MovePlan{
		Moves: []planning.FileMove{{
			OldPath: "src/Source/Item.php",
			NewPath: "src/Target/Item.php",
		}},
	})
	if err != nil {
		t.Fatalf("analyze php: %v", err)
	}
	if !relevant {
		t.Fatal("expected php analyzer to be relevant")
	}
	if len(response.SymbolMappings) != 1 {
		t.Fatalf("expected 1 symbol mapping, got %#v", response.SymbolMappings)
	}
	if response.SymbolMappings[0].OldSymbol != "App\\Source\\Item" {
		t.Fatalf("unexpected old symbol %q", response.SymbolMappings[0].OldSymbol)
	}
	if response.SymbolMappings[0].NewSymbol != "App\\Target\\Item" {
		t.Fatalf("unexpected new symbol %q", response.SymbolMappings[0].NewSymbol)
	}

	assertReplacement(t, response.Replacements, "src/Source/Item.php", "App\\Source", "App\\Target")
	assertReplacement(t, response.Replacements, "src/Consumer/Consumer.php", "App\\Source\\Item", "App\\Target\\Item")
	assertReplacement(t, response.Replacements, "src/Consumer/Consumer.php", "\\App\\Source\\Item", "\\App\\Target\\Item")
}

func TestAnalyzerRenamesMovedClassDeclaration(t *testing.T) {
	root := testfixtures.CopyDir(t, "tests/fixtures/php-analyzer-basic/rename-class")

	response, _, err := analyzePHP(t, root, planning.MovePlan{
		Moves: []planning.FileMove{{
			OldPath: "src/Source/Item.php",
			NewPath: "src/Source/RenamedItem.php",
		}},
	})
	if err != nil {
		t.Fatalf("analyze php: %v", err)
	}

	assertReplacement(t, response.Replacements, "src/Source/Item.php", "Item", "RenamedItem")
	assertReplacement(t, response.Replacements, "src/Consumer/Consumer.php", "App\\Source\\Item", "App\\Source\\RenamedItem")
	assertReplacement(t, response.Replacements, "src/Consumer/Consumer.php", "Item", "RenamedItem")
}

func TestAnalyzerUpdatesDocblockReferences(t *testing.T) {
	root := testfixtures.CopyDir(t, "tests/fixtures/php-analyzer-basic/docblock")

	response, _, err := analyzePHP(t, root, planning.MovePlan{
		Moves: []planning.FileMove{{
			OldPath: "src/Source/Item.php",
			NewPath: "src/Target/RenamedItem.php",
		}},
	})
	if err != nil {
		t.Fatalf("analyze php: %v", err)
	}

	assertReplacement(t, response.Replacements, "src/Consumer/Consumer.php", "Item", "RenamedItem")
}

func TestAnalyzerSkipsUnrelatedInvalidPHPFiles(t *testing.T) {
	root := testfixtures.CopyDir(t, "tests/fixtures/php-analyzer-basic/invalid-unrelated")

	response, _, err := analyzePHP(t, root, planning.MovePlan{
		Moves: []planning.FileMove{{
			OldPath: "src/Source/Item.php",
			NewPath: "src/Target/Item.php",
		}},
	})
	if err != nil {
		t.Fatalf("analyze php: %v", err)
	}

	assertReplacement(t, response.Replacements, "src/Consumer/Consumer.php", "App\\Source\\Item", "App\\Target\\Item")
	assertNoWarningInFile(t, response.Warnings, "src/Fixtures/Broken.php")
}

func TestAnalyzerRewritesImportsInFilesWithRecoveredPHP85Syntax(t *testing.T) {
	root := testfixtures.CopyDir(t, "tests/fixtures/php-analyzer/recovered-syntax")

	response, _, err := analyzePHP(t, root, planning.MovePlan{
		Moves: []planning.FileMove{{
			OldPath: "src/Module/Record/Domain/Record.php",
			NewPath: "src/Module/Record.php",
		}},
	})
	if err != nil {
		t.Fatalf("analyze php: %v", err)
	}

	assertReplacement(t, response.Replacements, "src/Consumer/RecordMapper.php", "App\\Module\\Record\\Domain\\Record", "App\\Module\\Record")
	assertNoWarningInFile(t, response.Warnings, "src/Consumer/RecordMapper.php")
}

func TestAnalyzerAddsImportsForMovedFileNamespaceLocalDependencies(t *testing.T) {
	root := testfixtures.CopyDir(t, "tests/fixtures/php-analyzer/local-dependency-import")

	response, _, err := analyzePHP(t, root, planning.MovePlan{
		Moves: []planning.FileMove{{
			OldPath: "src/Module/Source/Container.php",
			NewPath: "src/Module/Target/Container.php",
		}},
	})
	if err != nil {
		t.Fatalf("analyze php: %v", err)
	}

	assertReplacementContaining(t, response.Replacements, "src/Module/Source/Container.php", "use App\\Module\\Source\\InputValue;")
}

func TestAnalyzerRemovesImportsThatBecomeSameNamespace(t *testing.T) {
	root := testfixtures.CopyDir(t, "tests/fixtures/php-analyzer/same-namespace-import-removal")

	response, _, err := analyzePHP(t, root, planning.MovePlan{
		Moves: []planning.FileMove{
			{
				OldPath: "src/Module/Source/Container.php",
				NewPath: "src/Module/Target/Container.php",
			},
			{
				OldPath: "src/Module/Source/ItemList.php",
				NewPath: "src/Module/Target/ItemList.php",
			},
		},
	})
	if err != nil {
		t.Fatalf("analyze php: %v", err)
	}

	assertReplacement(t, response.Replacements, "src/Module/Source/Container.php", "use App\\Module\\Source\\ItemList;", "")
	assertNoReplacement(t, response.Replacements, "src/Module/Source/Container.php", "App\\Module\\Target\\ItemList")
}

func TestAnalyzerUpdatesTwigTemplateReferences(t *testing.T) {
	root := testfixtures.CopyDir(t, "tests/fixtures/php-analyzer/twig-template")

	response, relevant, err := analyzePHP(t, root, planning.MovePlan{
		Moves: []planning.FileMove{{
			OldPath: "templates/module/detail.html.twig",
			NewPath: "src/Module/View/Twig/detail.html.twig",
		}},
	})
	if err != nil {
		t.Fatalf("analyze php: %v", err)
	}
	if !relevant {
		t.Fatal("expected php analyzer to be relevant for twig move")
	}
	if len(response.PathMappings) == 0 {
		t.Fatalf("expected twig path mappings, got %#v", response)
	}

	assertReplacement(t, response.Replacements, "templates/module/detail.html.twig", "'module/detail.html.twig'", "'@Module/detail.html.twig'")
	assertReplacement(t, response.Replacements, "src/Controller.php", "'module/detail.html.twig'", "'@Module/detail.html.twig'")
}

func TestAnalyzerUpdatesStaticImportsForMovedAssets(t *testing.T) {
	root := testfixtures.CopyDir(t, "tests/fixtures/php-analyzer/static-import")

	response, relevant, err := analyzePHP(t, root, planning.MovePlan{
		Moves: []planning.FileMove{{
			OldPath: "src/Module/View/entry.css",
			NewPath: "src/Module/Display/entry.css",
		}},
	})
	if err != nil {
		t.Fatalf("analyze php: %v", err)
	}
	if !relevant {
		t.Fatal("expected php analyzer to be relevant for static asset move")
	}

	assertReplacement(t, response.Replacements, "assets/app.js", "../src/Module/View/entry.css", "../src/Module/Display/entry.css")
}

func TestAnalyzerUpdatesAssetMapperPathForDirectoryMove(t *testing.T) {
	root := testfixtures.CopyDir(t, "tests/fixtures/php-analyzer/asset-mapper")

	response, relevant, err := analyzePHP(t, root, planning.MovePlan{
		OldPath: "src/Module/Ui/Web",
		NewPath: "src/Module/Ui/Browser",
		IsDir:   true,
		Moves: []planning.FileMove{{
			OldPath: "src/Module/Ui/Web/icon.svg",
			NewPath: "src/Module/Ui/Browser/icon.svg",
		}},
	})
	if err != nil {
		t.Fatalf("analyze php: %v", err)
	}
	if !relevant {
		t.Fatal("expected php analyzer to be relevant for project directory move")
	}

	assertReplacement(t, response.Replacements, "config/packages/asset_mapper.yaml", "'src/Module/Ui/Web/'", "'src/Module/Ui/Browser/'")
}

func TestAnalyzerUpdatesTwigComponentNamespaceDefaults(t *testing.T) {
	root := testfixtures.CopyDir(t, "tests/fixtures/php-analyzer/twig-component")

	response, _, err := analyzePHP(t, root, planning.MovePlan{
		Moves: []planning.FileMove{{
			OldPath: "src/Module/Widget/Ui/Web/WidgetComponent.php",
			NewPath: "src/Module/Widget/Ui/Browser/WidgetComponent.php",
		}},
	})
	if err != nil {
		t.Fatalf("analyze php: %v", err)
	}

	assertReplacement(t, response.Replacements, "config/packages/twig_component.yaml", "'App\\Module\\Widget\\Ui\\Web\\'", "'App\\Module\\Widget\\Ui\\Browser\\'")
}

func TestAnalyzerReportsSemanticRenameHints(t *testing.T) {
	root := testfixtures.CopyDir(t, "tests/fixtures/php-analyzer/semantic-rename-hints")

	response, _, err := analyzePHP(t, root, planning.MovePlan{
		Moves: []planning.FileMove{{
			OldPath: "src/Module/Handler/OldItemHandler.php",
			NewPath: "src/Module/Handler/NewItemHandler.php",
		}},
	})
	if err != nil {
		t.Fatalf("analyze php: %v", err)
	}

	assertWarning(t, response.Warnings, "src/Module/Consumer/Consumer.php", `Semantic name "oldItemHandlers" resembles moved symbol; consider "newItemHandlers". Not changed.`)
	assertWarning(t, response.Warnings, "config/packages/services.yaml", `Semantic name "old_item_handler" resembles moved symbol; consider "new_item_handler". Not changed.`)
}

func TestAnalyzerWarnsForSkippedGroupUseStatements(t *testing.T) {
	root := testfixtures.CopyDir(t, "tests/fixtures/php-analyzer/group-use-warning")

	response, _, err := analyzePHP(t, root, planning.MovePlan{
		Moves: []planning.FileMove{{
			OldPath: "src/Source/Item.php",
			NewPath: "src/Target/Item.php",
		}},
	})
	if err != nil {
		t.Fatalf("analyze php: %v", err)
	}

	assertWarning(t, response.Warnings, "src/Consumer/Consumer.php", "Group use statement references a moved symbol; skipped conservatively.")
	assertNoReplacementInFile(t, response.Replacements, "src/Consumer/Consumer.php")
}

func TestAnalyzerWarnsAboutStringLiteralsContainingMovedSymbols(t *testing.T) {
	root := testfixtures.CopyDir(t, "tests/fixtures/php-analyzer/string-literal-warning")

	response, _, err := analyzePHP(t, root, planning.MovePlan{
		Moves: []planning.FileMove{{
			OldPath: "src/Rules/OldRule.php",
			NewPath: "src/Rules/NewRule.php",
		}},
	})
	if err != nil {
		t.Fatalf("analyze php: %v", err)
	}

	assertWarning(t, response.Warnings, "tests/Rules/RuleReferenceTest.php", "String literal references a moved PHP symbol; not changed.")
	assertNoReplacementInFile(t, response.Replacements, "tests/Rules/RuleReferenceTest.php")
}

func TestAnalyzerHonoursScanExcludes(t *testing.T) {
	root := testfixtures.CopyDir(t, "tests/fixtures/php-analyzer/scan-excludes")

	response, _, err := analyzePHPWithConfig(t, root, planning.MovePlan{
		Moves: []planning.FileMove{{
			OldPath: "src/Module/Source/Resolver.php",
			NewPath: "src/Module/Source/RenamedResolver.php",
		}},
	}, config.Config{
		Exclude: []string{"local/phpstan/tests/fixtures/**"},
	})
	if err != nil {
		t.Fatalf("analyze php: %v", err)
	}

	assertReplacement(t, response.Replacements, "src/Consumer/Consumer.php", "App\\Module\\Source\\Resolver", "App\\Module\\Source\\RenamedResolver")
	assertNoReplacementInFile(t, response.Replacements, "local/phpstan/tests/fixtures/ArchitectureDependency.php")
}

func TestAnalyzerUsesComposerRootForMonorepoPaths(t *testing.T) {
	root := testfixtures.CopyDir(t, "tests/fixtures/php-analyzer/composer-root")

	response, relevant, err := analyzePHP(t, root, planning.MovePlan{
		Moves: []planning.FileMove{{
			OldPath: "platform/src/Source/Item.php",
			NewPath: "platform/src/Target/Item.php",
		}},
	})
	if err != nil {
		t.Fatalf("analyze php: %v", err)
	}
	if !relevant {
		t.Fatal("expected php analyzer to be relevant")
	}

	assertReplacement(t, response.Replacements, "platform/src/Source/Item.php", "App\\Source", "App\\Target")
	assertReplacement(t, response.Replacements, "platform/src/Consumer/Consumer.php", "App\\Source\\Item", "App\\Target\\Item")
}

func TestAnalyzerHandlesMultipleComposerRootsInOnePlan(t *testing.T) {
	root := testfixtures.CopyDir(t, "tests/fixtures/php-analyzer/multiple-composer-roots")

	response, relevant, err := analyzePHP(t, root, planning.MovePlan{
		Moves: []planning.FileMove{
			{
				OldPath: "platform/src/Source/Item.php",
				NewPath: "platform/src/Target/Item.php",
			},
			{
				OldPath: "admin/src/Source/AdminItem.php",
				NewPath: "admin/src/Target/AdminItem.php",
			},
		},
	})
	if err != nil {
		t.Fatalf("analyze php: %v", err)
	}
	if !relevant {
		t.Fatal("expected php analyzer to be relevant")
	}

	assertReplacement(t, response.Replacements, "platform/src/Consumer/Consumer.php", "App\\Source\\Item", "App\\Target\\Item")
	assertReplacement(t, response.Replacements, "admin/src/Consumer/AdminConsumer.php", "Admin\\Source\\AdminItem", "Admin\\Target\\AdminItem")
}

func TestAnalyzerIgnoresUnrelatedComposerRootsInMixedLanguageMoves(t *testing.T) {
	root := testfixtures.CopyDir(t, "tests/fixtures/php-analyzer/mixed-language-roots")

	response, relevant, err := analyzePHP(t, root, planning.MovePlan{
		Moves: []planning.FileMove{
			{
				OldPath: "platform/src/Source/Item.php",
				NewPath: "platform/src/Target/Item.php",
			},
			{
				OldPath: "tools/src/app/source/item.py",
				NewPath: "tools/src/app/target/item.py",
			},
		},
	})
	if err != nil {
		t.Fatalf("analyze php: %v", err)
	}
	if !relevant {
		t.Fatal("expected php analyzer to be relevant")
	}

	assertReplacement(t, response.Replacements, "platform/src/Consumer/Consumer.php", "App\\Source\\Item", "App\\Target\\Item")
	assertNoReplacementInFile(t, response.Replacements, "tools/src/app/source/item.py")
}

func assertReplacement(t *testing.T, replacements []adapterproto.Replacement, file string, oldText string, newText string) {
	t.Helper()

	for _, replacement := range replacements {
		if replacement.File == file && replacement.Replacement == newText {
			return
		}
	}
	t.Fatalf("expected replacement in %s from %q to %q, got %#v", file, oldText, newText, replacements)
}

func assertReplacementContaining(t *testing.T, replacements []adapterproto.Replacement, file string, text string) {
	t.Helper()

	for _, replacement := range replacements {
		if replacement.File == file && strings.Contains(replacement.Replacement, text) {
			return
		}
	}
	t.Fatalf("expected replacement in %s containing %q, got %#v", file, text, replacements)
}

func assertNoReplacement(t *testing.T, replacements []adapterproto.Replacement, file string, replacementText string) {
	t.Helper()

	for _, replacement := range replacements {
		if replacement.File == file && replacement.Replacement == replacementText {
			t.Fatalf("unexpected replacement in %s to %q: %#v", file, replacementText, replacement)
		}
	}
}

func assertNoReplacementInFile(t *testing.T, replacements []adapterproto.Replacement, file string) {
	t.Helper()

	for _, replacement := range replacements {
		if replacement.File == file {
			t.Fatalf("unexpected replacement in excluded file %s: %#v", file, replacement)
		}
	}
}

func assertWarning(t *testing.T, warnings []adapterproto.Warning, file string, message string) {
	t.Helper()

	for _, warning := range warnings {
		if warning.File == file && warning.Message == message {
			return
		}
	}
	t.Fatalf("expected warning in %s: %s, got %#v", file, message, warnings)
}

func assertNoWarningInFile(t *testing.T, warnings []adapterproto.Warning, file string) {
	t.Helper()

	for _, warning := range warnings {
		if warning.File == file {
			t.Fatalf("unexpected warning in %s: %#v", file, warning)
		}
	}
}
