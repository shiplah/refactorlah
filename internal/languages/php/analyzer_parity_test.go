//go:build cgo

package php

import (
	"sort"
	"strings"
	"testing"

	adapterproto "refactorlah/internal/adapters"
	"refactorlah/internal/planning"
)

func TestAnalyzerKeepsOverlappingPHPRenamesTokenScoped(t *testing.T) {
	root := t.TempDir()
	writeAnalyzerFixtureFile(t, root, "composer.json", `{"autoload":{"psr-4":{"App\\":"src/"}}}`)
	memoryIndex := `<?php

declare(strict_types=1);

namespace App\Billing\Invoice\Infrastructure\Cache;

use App\Billing\Invoice\Application\InvoiceLookup;

final readonly class CacheInvoiceIndex implements InvoiceLookup {}
`
	services := `<?php

use App\Billing\Invoice\Application\InvoiceLookup;
use App\Billing\Invoice\Infrastructure\Cache\CacheInvoiceIndex;

return static function ($services): void {
    $services->set(CacheInvoiceIndex::class);
    $services->alias(InvoiceLookup::class, CacheInvoiceIndex::class);
};
`
	writeAnalyzerFixtureFile(t, root, "src/Billing/Invoice/Application/InvoiceLookup.php", "<?php\nnamespace App\\Billing\\Invoice\\Application;\ninterface InvoiceLookup {}\n")
	writeAnalyzerFixtureFile(t, root, "src/Billing/Invoice/Infrastructure/Cache/CacheInvoiceIndex.php", memoryIndex)
	writeAnalyzerFixtureFile(t, root, "services.php", services)

	response, _, err := NewAnalyzer().Analyze(root, planning.MovePlan{
		Moves: []planning.FileMove{{
			OldPath: "src/Billing/Invoice/Infrastructure/Cache/CacheInvoiceIndex.php",
			NewPath: "src/Billing/Invoice/Infrastructure/Cache/CacheInvoiceLookup.php",
		}},
	})
	if err != nil {
		t.Fatalf("analyze php: %v", err)
	}

	updatedCacheIndex := applyPHPAdapterReplacements(memoryIndex, response.Replacements, "src/Billing/Invoice/Infrastructure/Cache/CacheInvoiceIndex.php")
	assertPHPTextEqual(t, `<?php

declare(strict_types=1);

namespace App\Billing\Invoice\Infrastructure\Cache;

use App\Billing\Invoice\Application\InvoiceLookup;

final readonly class CacheInvoiceLookup implements InvoiceLookup {}
`, updatedCacheIndex)

	updatedServices := applyPHPAdapterReplacements(services, response.Replacements, "services.php")
	assertPHPTextEqual(t, `<?php

use App\Billing\Invoice\Application\InvoiceLookup;
use App\Billing\Invoice\Infrastructure\Cache\CacheInvoiceLookup;

return static function ($services): void {
    $services->set(CacheInvoiceLookup::class);
    $services->alias(InvoiceLookup::class, CacheInvoiceLookup::class);
};
`, updatedServices)
	assertPHPTextNotContains(t, updatedServices, "AssetRegistCache")
	assertPHPTextNotContains(t, updatedServices, "CacheInvoiceLookupectIndex")
}

func TestAnalyzerUpdatesImportedShortReferencesWhenNamespaceAndBasenameChange(t *testing.T) {
	root := t.TempDir()
	writeAnalyzerFixtureFile(t, root, "composer.json", `{"autoload":{"psr-4":{"App\\":"src/"}},"autoload-dev":{"psr-4":{"App\\Tests\\":"tests/"}}}`)
	movedRenderer := `<?php

declare(strict_types=1);

namespace App\Shared\RichText\Ui\Web\Block;

final class AccordionBlockWebRenderer
{
    public static function make(): self
    {
        return new self();
    }
}
`
	consumer := `<?php

declare(strict_types=1);

namespace App\Shared\RichText\Ui\Web;

use App\Shared\RichText\Ui\Web\Block\AccordionBlockWebRenderer;

final class HtmlRichTextRenderer
{
    private ?AccordionBlockWebRenderer $renderer = null;

    public function render(AccordionBlockWebRenderer $renderer): AccordionBlockWebRenderer
    {
        $this->renderer = $renderer;

        if (!$renderer instanceof AccordionBlockWebRenderer) {
            return new AccordionBlockWebRenderer();
        }

        return AccordionBlockWebRenderer::make();
    }
}
`
	services := `<?php

use App\Shared\RichText\Ui\Web\Block\AccordionBlockWebRenderer;

return static function ($services): void {
    $services->instanceof(AccordionBlockWebRenderer::class);
    $services->set(AccordionBlockWebRenderer::class);
};
`
	testFile := `<?php

declare(strict_types=1);

namespace App\Tests\Shared\RichText;

use App\Shared\RichText\Ui\Web\Block\AccordionBlockWebRenderer;

$renderer = new AccordionBlockWebRenderer();
$matches = $renderer instanceof AccordionBlockWebRenderer;
`
	writeAnalyzerFixtureFile(t, root, "src/Shared/RichText/Ui/Web/Block/AccordionBlockWebRenderer.php", movedRenderer)
	writeAnalyzerFixtureFile(t, root, "src/Shared/RichText/Ui/Web/HtmlRichTextRenderer.php", consumer)
	writeAnalyzerFixtureFile(t, root, "services.php", services)
	writeAnalyzerFixtureFile(t, root, "tests/Shared/RichText/AccordionRendererTest.php", testFile)

	response, _, err := NewAnalyzer().Analyze(root, planning.MovePlan{
		Moves: []planning.FileMove{{
			OldPath: "src/Shared/RichText/Ui/Web/Block/AccordionBlockWebRenderer.php",
			NewPath: "src/Shared/RichText/Ui/Web/Renderer/AccordionRenderableWebRenderer.php",
		}},
	})
	if err != nil {
		t.Fatalf("analyze php: %v", err)
	}

	updatedRenderer := applyPHPAdapterReplacements(movedRenderer, response.Replacements, "src/Shared/RichText/Ui/Web/Block/AccordionBlockWebRenderer.php")
	assertPHPTextContains(t, updatedRenderer, "namespace App\\Shared\\RichText\\Ui\\Web\\Renderer;")
	assertPHPTextContains(t, updatedRenderer, "final class AccordionRenderableWebRenderer")

	updatedConsumer := applyPHPAdapterReplacements(consumer, response.Replacements, "src/Shared/RichText/Ui/Web/HtmlRichTextRenderer.php")
	assertPHPTextContains(t, updatedConsumer, "use App\\Shared\\RichText\\Ui\\Web\\Renderer\\AccordionRenderableWebRenderer;")
	assertPHPTextContains(t, updatedConsumer, "private ?AccordionRenderableWebRenderer $renderer = null;")
	assertPHPTextContains(t, updatedConsumer, "public function render(AccordionRenderableWebRenderer $renderer): AccordionRenderableWebRenderer")
	assertPHPTextContains(t, updatedConsumer, "$renderer instanceof AccordionRenderableWebRenderer")
	assertPHPTextContains(t, updatedConsumer, "return new AccordionRenderableWebRenderer();")
	assertPHPTextContains(t, updatedConsumer, "return AccordionRenderableWebRenderer::make();")
	assertPHPTextNotContains(t, updatedConsumer, "AccordionBlockWebRenderer")

	updatedServices := applyPHPAdapterReplacements(services, response.Replacements, "services.php")
	assertPHPTextContains(t, updatedServices, "use App\\Shared\\RichText\\Ui\\Web\\Renderer\\AccordionRenderableWebRenderer;")
	assertPHPTextContains(t, updatedServices, "$services->instanceof(AccordionRenderableWebRenderer::class);")
	assertPHPTextContains(t, updatedServices, "$services->set(AccordionRenderableWebRenderer::class);")
	assertPHPTextNotContains(t, updatedServices, "AccordionBlockWebRenderer")

	updatedTest := applyPHPAdapterReplacements(testFile, response.Replacements, "tests/Shared/RichText/AccordionRendererTest.php")
	assertPHPTextContains(t, updatedTest, "use App\\Shared\\RichText\\Ui\\Web\\Renderer\\AccordionRenderableWebRenderer;")
	assertPHPTextContains(t, updatedTest, "$renderer = new AccordionRenderableWebRenderer();")
	assertPHPTextContains(t, updatedTest, "$matches = $renderer instanceof AccordionRenderableWebRenderer;")
	assertPHPTextNotContains(t, updatedTest, "AccordionBlockWebRenderer")
}

func TestAnalyzerUpdatesImportedEnumCaseReferencesWhenBasenameChanges(t *testing.T) {
	root := t.TempDir()
	writeAnalyzerFixtureFile(t, root, "platform/composer.json", `{"autoload":{"psr-4":{"App\\":"src/"}}}`)
	enumSource := `<?php

declare(strict_types=1);

namespace App\Shared\RichText\Application;

enum RichTextComponentKind
{
    case Accordion;
}
`
	renderer := `<?php

declare(strict_types=1);

namespace App\Shared\RichText\Ui\Web\Renderer;

use App\Shared\RichText\Application\RichTextComponentKind;

final class AccordionRenderableWebRenderer
{
    public function kind(): RichTextComponentKind
    {
        return RichTextComponentKind::Accordion;
    }
}
`
	writeAnalyzerFixtureFile(t, root, "platform/src/Shared/RichText/Application/RichTextComponentKind.php", enumSource)
	writeAnalyzerFixtureFile(t, root, "platform/src/Shared/RichText/Ui/Web/Renderer/AccordionRenderableWebRenderer.php", renderer)

	response, _, err := NewAnalyzer().Analyze(root, planning.MovePlan{
		Moves: []planning.FileMove{{
			OldPath: "platform/src/Shared/RichText/Application/RichTextComponentKind.php",
			NewPath: "platform/src/Shared/RichText/Application/RichTextDirectiveKind.php",
		}},
	})
	if err != nil {
		t.Fatalf("analyze php: %v", err)
	}

	updatedEnum := applyPHPAdapterReplacements(enumSource, response.Replacements, "platform/src/Shared/RichText/Application/RichTextComponentKind.php")
	assertPHPTextContains(t, updatedEnum, "enum RichTextDirectiveKind")

	updatedRenderer := applyPHPAdapterReplacements(renderer, response.Replacements, "platform/src/Shared/RichText/Ui/Web/Renderer/AccordionRenderableWebRenderer.php")
	assertPHPTextContains(t, updatedRenderer, "use App\\Shared\\RichText\\Application\\RichTextDirectiveKind;")
	assertPHPTextContains(t, updatedRenderer, "public function kind(): RichTextDirectiveKind")
	assertPHPTextContains(t, updatedRenderer, "return RichTextDirectiveKind::Accordion;")
	assertPHPTextNotContains(t, updatedRenderer, "RichTextComponentKind")
}

func TestAnalyzerUpdatesSameNamespaceShortReferencesWhenBasenameChanges(t *testing.T) {
	root := t.TempDir()
	writeAnalyzerFixtureFile(t, root, "composer.json", `{"autoload":{"psr-4":{"App\\":"src/"}}}`)
	component := `<?php

declare(strict_types=1);

namespace App\Shared\RichText;

interface ComponentRenderer {}
`
	consumer := `<?php

declare(strict_types=1);

namespace App\Shared\RichText;

final class DirectiveNodeRenderer
{
    /** @param iterable<ComponentRenderer> $renderers */
    public function __construct(private iterable $renderers) {}

    public function renderer(?ComponentRenderer $renderer): ComponentRenderer
    {
        if (!$renderer instanceof ComponentRenderer) {
            return new ComponentRenderer();
        }

        return ComponentRenderer::make();
    }
}
`
	writeAnalyzerFixtureFile(t, root, "src/Shared/RichText/ComponentRenderer.php", component)
	writeAnalyzerFixtureFile(t, root, "src/Shared/RichText/DirectiveNodeRenderer.php", consumer)

	response, _, err := NewAnalyzer().Analyze(root, planning.MovePlan{
		Moves: []planning.FileMove{{
			OldPath: "src/Shared/RichText/ComponentRenderer.php",
			NewPath: "src/Shared/RichText/DirectiveRenderer.php",
		}},
	})
	if err != nil {
		t.Fatalf("analyze php: %v", err)
	}

	updatedComponent := applyPHPAdapterReplacements(component, response.Replacements, "src/Shared/RichText/ComponentRenderer.php")
	assertPHPTextContains(t, updatedComponent, "interface DirectiveRenderer")

	updatedConsumer := applyPHPAdapterReplacements(consumer, response.Replacements, "src/Shared/RichText/DirectiveNodeRenderer.php")
	assertPHPTextContains(t, updatedConsumer, "@param iterable<DirectiveRenderer> $renderers")
	assertPHPTextContains(t, updatedConsumer, "public function renderer(?DirectiveRenderer $renderer): DirectiveRenderer")
	assertPHPTextContains(t, updatedConsumer, "$renderer instanceof DirectiveRenderer")
	assertPHPTextContains(t, updatedConsumer, "return new DirectiveRenderer();")
	assertPHPTextContains(t, updatedConsumer, "return DirectiveRenderer::make();")
	assertPHPTextNotContains(t, updatedConsumer, "ComponentRenderer")
	assertPHPTextNotContains(t, updatedConsumer, "use App\\Shared\\RichText\\DirectiveRenderer;")
}

func TestAnalyzerKeepsSameFileHelperClassesNamespaceLocal(t *testing.T) {
	root := t.TempDir()
	writeAnalyzerFixtureFile(t, root, "composer.json", `{"autoload-dev":{"psr-4":{"App\\Tests\\":"tests/"}}}`)
	testSource := `<?php

declare(strict_types=1);

namespace App\Tests\Application\Billing\Invoice;

final class RewriteInvoiceRichTextLinksTest
{
    public function helper(): Helper
    {
        return new Helper();
    }
}

final class Helper {}
`
	writeAnalyzerFixtureFile(t, root, "tests/Application/Billing/Invoice/RewriteInvoiceRichTextLinksTest.php", testSource)

	response, _, err := NewAnalyzer().Analyze(root, planning.MovePlan{
		Moves: []planning.FileMove{{
			OldPath: "tests/Application/Billing/Invoice/RewriteInvoiceRichTextLinksTest.php",
			NewPath: "tests/Billing/Archive/Detailed/Application/RewriteInvoiceRichTextLinksTest.php",
		}},
	})
	if err != nil {
		t.Fatalf("analyze php: %v", err)
	}

	updated := applyPHPAdapterReplacements(testSource, response.Replacements, "tests/Application/Billing/Invoice/RewriteInvoiceRichTextLinksTest.php")
	assertPHPTextContains(t, updated, "namespace App\\Tests\\Billing\\Archive\\Detailed\\Application;")
	assertPHPTextContains(t, updated, "public function helper(): Helper")
	assertPHPTextContains(t, updated, "return new Helper();")
	assertPHPTextNotContains(t, updated, "use App\\Tests\\Application\\Billing\\Invoice\\Helper;")
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

func assertPHPTextContains(t *testing.T, actual string, expected string) {
	t.Helper()

	if !strings.Contains(actual, expected) {
		t.Fatalf("expected PHP to contain %q, got:\n%s", expected, actual)
	}
}

func assertPHPTextNotContains(t *testing.T, actual string, unexpected string) {
	t.Helper()

	if strings.Contains(actual, unexpected) {
		t.Fatalf("expected PHP not to contain %q, got:\n%s", unexpected, actual)
	}
}
