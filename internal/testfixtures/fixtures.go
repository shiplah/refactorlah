package testfixtures

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func Path(t testing.TB, relativePath string) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("locate test fixture helper")
	}

	return filepath.Join(filepath.Dir(file), "..", "..", filepath.FromSlash(relativePath))
}

func Read(t testing.TB, relativePath string) []byte {
	t.Helper()

	source, err := os.ReadFile(Path(t, relativePath))
	if err != nil {
		t.Fatalf("read fixture file %s: %v", relativePath, err)
	}

	return source
}

func AssertFileMatches(t testing.TB, actualPath string, expectedFixture string) {
	t.Helper()

	actual, err := os.ReadFile(actualPath)
	if err != nil {
		t.Fatalf("read actual file %s: %v", actualPath, err)
	}

	expected := Read(t, expectedFixture)
	actualText := NormalizeNewlines(string(actual))
	expectedText := NormalizeNewlines(string(expected))
	if actualText != expectedText {
		t.Fatalf("expected %s to match %s\ngot:\n%s\nwant:\n%s", actualPath, expectedFixture, actualText, expectedText)
	}
}

func NormalizeNewlines(text string) string {
	return strings.ReplaceAll(text, "\r\n", "\n")
}

func CopyDir(t testing.TB, relativePath string) string {
	t.Helper()

	root := t.TempDir()
	sourceRoot := Path(t, relativePath)
	err := filepath.Walk(sourceRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(sourceRoot, path)
		if err != nil {
			return err
		}

		target := filepath.Join(root, rel)
		if info.IsDir() {
			return os.MkdirAll(target, 0o755)
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		return os.WriteFile(target, data, info.Mode())
	})
	if err != nil {
		t.Fatalf("copy fixture %s: %v", relativePath, err)
	}

	return root
}
