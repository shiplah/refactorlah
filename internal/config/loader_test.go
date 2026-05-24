package config

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestLoaderMergesRootAndNestedConfigs(t *testing.T) {
	root := t.TempDir()
	mustWriteConfig(t, filepath.Join(root, ".refactorlah.json"), `{
		"exclude": ["platform/local/phpstan/tests/fixtures/**"],
		"include": ["/platform/local/phpstan/tests/fixtures/Allowed.php"]
	}`)
	mustWriteConfig(t, filepath.Join(root, "platform", ".refactorlah.json"), `{
		"exclude": ["local/generated/**"],
		"include": ["local/generated/Keep.php"]
	}`)

	config, err := NewLoader().Load(root)
	if err != nil {
		t.Fatal(err)
	}

	wantInclude := []string{
		"platform/local/phpstan/tests/fixtures/Allowed.php",
		"platform/local/generated/Keep.php",
	}
	if !reflect.DeepEqual(config.Include, wantInclude) {
		t.Fatalf("unexpected include patterns: %#v", config.Include)
	}

	wantExclude := []string{
		"platform/local/phpstan/tests/fixtures/**",
		"platform/local/generated/**",
	}
	if !reflect.DeepEqual(config.Exclude, wantExclude) {
		t.Fatalf("unexpected exclude patterns: %#v", config.Exclude)
	}
}

func TestLoaderSkipsIgnoredDirectories(t *testing.T) {
	root := t.TempDir()
	mustWriteConfig(t, filepath.Join(root, ".refactorlah.json"), `{"exclude":["src/generated/**"]}`)
	mustWriteConfig(t, filepath.Join(root, "vendor", "package", ".refactorlah.json"), `{"exclude":["src/**"]}`)

	config, err := NewLoader().Load(root)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(config.Exclude, []string{"src/generated/**"}) {
		t.Fatalf("unexpected exclude patterns: %#v", config.Exclude)
	}
}

func mustWriteConfig(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
