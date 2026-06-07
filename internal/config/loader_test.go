package config

import (
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"
)

func TestLoaderMergesConfigsFromSearchRootToThreeLevelsDeep(t *testing.T) {
	root := t.TempDir()
	mustWriteConfig(t, filepath.Join(root, ".refactorlah.json"), `{
		"exclude": ["platform/local/phpstan/tests/fixtures/**"],
		"include": ["/platform/local/phpstan/tests/fixtures/Allowed.php"],
		"checks": [["composer", "stan"]],
		"tests": [["composer", "test"]]
	}`)
	mustWriteConfig(t, filepath.Join(root, "platform", ".refactorlah.json"), `{
		"exclude": ["local/generated/**"],
		"include": ["local/generated/Keep.php"],
		"checks": [["bin/check"]]
	}`)
	mustWriteConfig(t, filepath.Join(root, "one", "two", "three", ".refactorlah.json"), `{
		"exclude": ["fixtures/**"]
	}`)

	config, err := NewLoader().Load(root, root)
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
		"one/two/three/fixtures/**",
	}
	if !reflect.DeepEqual(config.Exclude, wantExclude) {
		t.Fatalf("unexpected exclude patterns: %#v", config.Exclude)
	}

	wantChecks := [][]string{{"composer", "stan"}, {"bin/check"}}
	if !reflect.DeepEqual(config.Checks, wantChecks) {
		t.Fatalf("unexpected checks: %#v", config.Checks)
	}

	wantTests := [][]string{{"composer", "test"}}
	if !reflect.DeepEqual(config.Tests, wantTests) {
		t.Fatalf("unexpected tests: %#v", config.Tests)
	}
}

func TestLoaderDoesNotSearchPastThreeLevels(t *testing.T) {
	root := t.TempDir()
	mustWriteConfig(t, filepath.Join(root, ".refactorlah.json"), `{"exclude":["src/generated/**"]}`)
	mustWriteConfig(t, filepath.Join(root, "one", "two", "three", "four", ".refactorlah.json"), `{"exclude":["fixtures/**"]}`)

	config, err := NewLoader().Load(root, root)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(config.Exclude, []string{"src/generated/**"}) {
		t.Fatalf("unexpected exclude patterns: %#v", config.Exclude)
	}
}

func TestLoaderSearchesFromCurrentDirectory(t *testing.T) {
	root := t.TempDir()
	mustWriteConfig(t, filepath.Join(root, ".refactorlah.json"), `{"exclude":["root-only/**"]}`)
	mustWriteConfig(t, filepath.Join(root, "platform", ".refactorlah.json"), `{"exclude":["local/phpstan/tests/fixtures/**"]}`)

	config, err := NewLoader().Load(root, filepath.Join(root, "platform"))
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(config.Exclude, []string{"platform/local/phpstan/tests/fixtures/**"}) {
		t.Fatalf("unexpected exclude patterns: %#v", config.Exclude)
	}
}

func TestLoaderDeduplicatesThroughAbsolutePatternIndex(t *testing.T) {
	root := t.TempDir()
	mustWriteConfig(t, filepath.Join(root, ".refactorlah.json"), `{"exclude":["platform/generated/**"]}`)
	mustWriteConfig(t, filepath.Join(root, "platform", ".refactorlah.json"), `{"exclude":["generated/**"]}`)

	config, err := NewLoader().Load(root, root)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(config.Exclude, []string{"platform/generated/**"}) {
		t.Fatalf("unexpected exclude patterns: %#v", config.Exclude)
	}
}

func TestLoaderRejectsPatternsOutsideProjectRoot(t *testing.T) {
	root := t.TempDir()
	mustWriteConfig(t, filepath.Join(root, "platform", ".refactorlah.json"), `{"exclude":["../../outside/**"]}`)

	if _, err := NewLoader().Load(root, root); err == nil {
		t.Fatal("expected outside-project config pattern to fail")
	}
}

func TestLoaderRejectsSearchRootOutsideProjectRoot(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()

	if _, err := NewLoader().Load(root, outside); err == nil {
		t.Fatal("expected outside-project search root to fail")
	}
}

func TestLoaderAcceptsSymlinkedSearchRootInsideProjectRoot(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink permissions vary on Windows")
	}

	parent := t.TempDir()
	root := filepath.Join(parent, "real")
	searchRoot := filepath.Join(parent, "link")
	mustWriteConfig(t, filepath.Join(root, ".refactorlah.json"), `{"exclude":["fixtures/**"]}`)
	if err := os.Symlink(root, searchRoot); err != nil {
		t.Fatal(err)
	}

	config, err := NewLoader().Load(root, searchRoot)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(config.Exclude, []string{"fixtures/**"}) {
		t.Fatalf("unexpected exclude patterns: %#v", config.Exclude)
	}
}

func TestLoaderSkipsIgnoredDirectories(t *testing.T) {
	root := t.TempDir()
	mustWriteConfig(t, filepath.Join(root, ".refactorlah.json"), `{"exclude":["src/generated/**"]}`)
	mustWriteConfig(t, filepath.Join(root, "vendor", "package", ".refactorlah.json"), `{"exclude":["src/**"]}`)

	config, err := NewLoader().Load(root, root)
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
