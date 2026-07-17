package testfixtures

import (
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"

	adapterproto "github.com/shiplah/refactorlah/internal/adapters/contract"
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

func AssertStringMatches(t testing.TB, actual string, expectedFixture string) {
	t.Helper()

	expected := Read(t, expectedFixture)
	actualText := NormalizeNewlines(actual)
	expectedText := NormalizeNewlines(string(expected))
	if actualText != expectedText {
		t.Fatalf("expected text to match %s\ngot:\n%s\nwant:\n%s", expectedFixture, actualText, expectedText)
	}
}

func ApplyAdapterReplacements(content string, replacements []adapterproto.Replacement, file string) string {
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
