package jsonconfig_test

import (
	"encoding/json"
	"sort"
	"strings"
	"testing"

	"refactorlah/internal/adapters/javascript/jsonconfig"
	"refactorlah/internal/replacements"
)

func TestNormaliseRemovesCommentsAndTrailingCommas(t *testing.T) {
	content := []byte(`{
  // comments are accepted by tsconfig and jsconfig.
  "compilerOptions": {
    "paths": {
      "@helper": ["src/helper.ts",],
    },
  },
}
`)

	var decoded map[string]any
	if err := json.Unmarshal(jsonconfig.Normalise(content), &decoded); err != nil {
		t.Fatalf("normalised JSON should decode: %v", err)
	}
}

func TestStringValueReplacementsUpdatesOnlyDirectObjectValues(t *testing.T) {
	content := `{
  // JSONC comments should not affect ranges.
  "imports": {
    "#helper": "./src/old-helper.js",
    "#nested": {
      "default": "./src/old-helper.js"
    }
  }
}
`
	objectRange, ok := jsonconfig.ObjectPropertyRange([]byte(content), "imports")
	if !ok {
		t.Fatal("expected imports object range")
	}

	items := jsonconfig.StringValueReplacements("package.json", []byte(content), objectRange, map[string]string{
		"./src/old-helper.js": "./src/new-helper.js",
	}, "javascript-test", "javascript.TestRule")

	if len(items) != 1 {
		t.Fatalf("expected 1 replacement, got %#v", items)
	}
	if updated := applyJSONConfigReplacements(content, items); updated != `{
  // JSONC comments should not affect ranges.
  "imports": {
    "#helper": "./src/new-helper.js",
    "#nested": {
      "default": "./src/old-helper.js"
    }
  }
}
` {
		t.Fatalf("unexpected rewritten JSON:\n%s", updated)
	}
}

func TestSingleStringArrayValueReplacementsAllowsTrailingComma(t *testing.T) {
	content := `{
  "compilerOptions": {
    "paths": {
      "@helper": ["old-helper.ts",],
      "@many": ["old-helper.ts", "other.ts"]
    }
  }
}
`
	compilerOptionsRange, ok := jsonconfig.ObjectPropertyRange([]byte(content), "compilerOptions")
	if !ok {
		t.Fatal("expected compilerOptions object range")
	}
	pathsRange, ok := jsonconfig.ObjectPropertyRangeIn([]byte(content), compilerOptionsRange, "paths")
	if !ok {
		t.Fatal("expected paths object range")
	}

	items := jsonconfig.SingleStringArrayValueReplacements("tsconfig.json", []byte(content), pathsRange, map[string]string{
		"old-helper.ts": "new-helper.ts",
	}, "javascript-test", "javascript.TestRule")

	if len(items) != 1 {
		t.Fatalf("expected 1 replacement, got %#v", items)
	}
	if updated := applyJSONConfigReplacements(content, items); updated != `{
  "compilerOptions": {
    "paths": {
      "@helper": ["new-helper.ts",],
      "@many": ["old-helper.ts", "other.ts"]
    }
  }
}
` {
		t.Fatalf("unexpected rewritten JSON:\n%s", updated)
	}
}

func applyJSONConfigReplacements(content string, items []replacements.Replacement) string {
	sort.Slice(items, func(left int, right int) bool {
		return items[left].Start > items[right].Start
	})

	builder := strings.Builder{}
	builder.WriteString(content)
	for _, item := range items {
		updated := builder.String()
		builder.Reset()
		builder.WriteString(updated[:item.Start])
		builder.WriteString(item.Replacement)
		builder.WriteString(updated[item.End:])
	}
	return builder.String()
}
