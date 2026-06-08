package javascript

import (
	"testing"

	"github.com/NickSdot/refactorlah/internal/adapters/javascript/rules"
	"github.com/NickSdot/refactorlah/internal/planning"
)

func TestAnalyzerRewritesComplexApplicationCorpus(t *testing.T) {
	root := t.TempDir()
	tsconfig := `{
  "compilerOptions": {
    "baseUrl": ".",
    "paths": {
      "@app/*": ["src/*"],
      "@feature": ["src/features/checkout/old-card.tsx"],
      "@ambiguous": ["src/features/checkout/old-card.tsx", "src/fallback/card.tsx"]
    }
  }
}
`
	packageJSON := `{
  "name": "@example/shop",
  "imports": {
    "#app/*": "./src/*",
    "#feature": "./src/features/checkout/old-card.tsx",
    "#conditional/*": {
      "default": "./src/*"
    }
  }
}
`
	consumer := `import Card from '../features/checkout/old-card';
export { Card as CheckoutCard } from '../features/checkout/old-card';
const lazy = import('../features/checkout/old-card');
const common = require('../features/checkout/old-card');
import AliasCard from '@app/features/checkout/old-card';
import PackageCard from '#app/features/checkout/old-card';
import SelfCard from '@example/shop/src/features/checkout/old-card';
import StableTypeScriptAlias from '@feature';
import AmbiguousAlias from '@ambiguous';
import ConditionalPackageAlias from '#conditional/features/checkout/old-card';
const dynamic = import(` + "`../features/checkout/${name}`" + `);
const concatenated = require('../features/checkout/' + name);
`

	writeJavaScriptFixture(t, root, "tsconfig.json", tsconfig)
	writeJavaScriptFixture(t, root, "package.json", packageJSON)
	writeJavaScriptFixture(t, root, "src/pages/home.tsx", consumer)
	writeJavaScriptFixture(t, root, "src/features/checkout/old-card.tsx", "export const Card = () => null;\n")

	response, relevant, err := analyzeJavaScript(t, root, planning.MovePlan{
		Moves: []planning.FileMove{{
			OldPath: "src/features/checkout/old-card.tsx",
			NewPath: "src/features/payment/card.tsx",
		}},
	})
	if err != nil {
		t.Fatalf("analyze complex javascript corpus: %v", err)
	}
	if !relevant {
		t.Fatal("expected javascript analyzer to be relevant")
	}

	updatedConsumer := applyJavaScriptReplacements(consumer, response.Replacements, "src/pages/home.tsx")
	if updatedConsumer != `import Card from '../features/payment/card';
export { Card as CheckoutCard } from '../features/payment/card';
const lazy = import('../features/payment/card');
const common = require('../features/payment/card');
import AliasCard from '@app/features/payment/card';
import PackageCard from '#app/features/payment/card';
import SelfCard from '@example/shop/src/features/payment/card';
import StableTypeScriptAlias from '@feature';
import AmbiguousAlias from '@ambiguous';
import ConditionalPackageAlias from '#conditional/features/checkout/old-card';
const dynamic = import(`+"`../features/checkout/${name}`"+`);
const concatenated = require('../features/checkout/' + name);
` {
		t.Fatalf("unexpected rewritten corpus consumer:\n%s", updatedConsumer)
	}

	updatedTSConfig := applyJavaScriptReplacements(tsconfig, response.Replacements, "tsconfig.json")
	if updatedTSConfig != `{
  "compilerOptions": {
    "baseUrl": ".",
    "paths": {
      "@app/*": ["src/*"],
      "@feature": ["src/features/payment/card.tsx"],
      "@ambiguous": ["src/features/checkout/old-card.tsx", "src/fallback/card.tsx"]
    }
  }
}
` {
		t.Fatalf("unexpected rewritten tsconfig:\n%s", updatedTSConfig)
	}

	updatedPackageJSON := applyJavaScriptReplacements(packageJSON, response.Replacements, "package.json")
	if updatedPackageJSON != `{
  "name": "@example/shop",
  "imports": {
    "#app/*": "./src/*",
    "#feature": "./src/features/payment/card.tsx",
    "#conditional/*": {
      "default": "./src/*"
    }
  }
}
` {
		t.Fatalf("unexpected rewritten package.json:\n%s", updatedPackageJSON)
	}

	assertJavaScriptReplacement(t, response.Replacements, "src/pages/home.tsx", rules.ModuleSpecifierReason)
	assertJavaScriptReplacement(t, response.Replacements, "src/pages/home.tsx", rules.TypeScriptPathAliasReason)
	assertJavaScriptReplacement(t, response.Replacements, "src/pages/home.tsx", rules.PackageImportsReason)
	assertJavaScriptReplacement(t, response.Replacements, "src/pages/home.tsx", rules.PackageSelfReferenceReason)
	assertJavaScriptReplacement(t, response.Replacements, "tsconfig.json", rules.TypeScriptPathTargetReason)
	assertJavaScriptReplacement(t, response.Replacements, "package.json", rules.PackageImportTargetReason)
	if !hasJavaScriptWarning(response.Warnings, "tsconfig.json", `TypeScript path alias "@ambiguous" has multiple targets; skipped conservatively.`) {
		t.Fatalf("expected ambiguous tsconfig warning, got %#v", response.Warnings)
	}
	if !hasJavaScriptWarning(response.Warnings, "package.json", `Package imports entry "#conditional/*" uses conditional targets; skipped conservatively.`) {
		t.Fatalf("expected conditional package warning, got %#v", response.Warnings)
	}
}

func TestAnalyzerRewritesMultiMoveDirectoryCorpus(t *testing.T) {
	root := t.TempDir()
	writeJavaScriptFixture(t, root, "tsconfig.json", `{
  "compilerOptions": {
    "paths": {
      "@features/*": ["src/features/*"]
    }
  }
}
`)
	consumer := `import feature from '../features/checkout';
import { price } from '../features/checkout/pricing';
import featureAlias from '@features/checkout';
import { price as aliasPrice } from '@features/checkout/pricing';
`
	writeJavaScriptFixture(t, root, "src/pages/summary.ts", consumer)
	writeJavaScriptFixture(t, root, "src/features/checkout/index.ts", "export default 'checkout';\n")
	writeJavaScriptFixture(t, root, "src/features/checkout/pricing.ts", "export const price = 10;\n")

	response, relevant, err := analyzeJavaScript(t, root, planning.MovePlan{
		Moves: []planning.FileMove{
			{OldPath: "src/features/checkout/index.ts", NewPath: "src/features/billing/index.ts"},
			{OldPath: "src/features/checkout/pricing.ts", NewPath: "src/features/billing/pricing.ts"},
		},
	})
	if err != nil {
		t.Fatalf("analyze directory corpus: %v", err)
	}
	if !relevant {
		t.Fatal("expected javascript analyzer to be relevant")
	}

	updated := applyJavaScriptReplacements(consumer, response.Replacements, "src/pages/summary.ts")
	if updated != `import feature from '../features/billing';
import { price } from '../features/billing/pricing';
import featureAlias from '@features/billing';
import { price as aliasPrice } from '@features/billing/pricing';
` {
		t.Fatalf("unexpected rewritten directory corpus:\n%s", updated)
	}
}

func TestAnalyzerKeepsSimilarAndDynamicSpecifiersInCorpus(t *testing.T) {
	root := t.TempDir()
	consumer := `import exact from './old-helper';
import similar from './old-helper-extra';
const exactRequire = require("./old-helper");
const dynamicImport = import(` + "`./${name}`" + `);
const concatenatedRequire = require('./old-' + name);
`
	writeJavaScriptFixture(t, root, "src/consumer.js", consumer)
	writeJavaScriptFixture(t, root, "src/old-helper.js", "export default 'old';\n")
	writeJavaScriptFixture(t, root, "src/old-helper-extra.js", "export default 'extra';\n")

	response, relevant, err := analyzeJavaScript(t, root, planning.MovePlan{
		Moves: []planning.FileMove{{
			OldPath: "src/old-helper.js",
			NewPath: "src/new-helper.js",
		}},
	})
	if err != nil {
		t.Fatalf("analyze conservative corpus: %v", err)
	}
	if !relevant {
		t.Fatal("expected javascript analyzer to be relevant")
	}

	updated := applyJavaScriptReplacements(consumer, response.Replacements, "src/consumer.js")
	if updated != `import exact from './new-helper';
import similar from './old-helper-extra';
const exactRequire = require("./new-helper");
const dynamicImport = import(`+"`./${name}`"+`);
const concatenatedRequire = require('./old-' + name);
` {
		t.Fatalf("unexpected conservative corpus rewrite:\n%s", updated)
	}
}
