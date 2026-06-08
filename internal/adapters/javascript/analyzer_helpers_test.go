package javascript

import (
	"os"
	"path/filepath"
	"sort"
	"testing"

	adapterproto "refactorlah/internal/adapters/contract"
	"refactorlah/internal/adapters/scan"
	"refactorlah/internal/config"
	"refactorlah/internal/planning"
)

func analyzeJavaScript(t *testing.T, root string, plan planning.MovePlan) (adapterproto.AggregatedResponse, bool, error) {
	t.Helper()
	scanConfig := config.Config{}
	return NewAnalyzer().Analyze(root, plan, scanConfig, scan.NewIndex(root, scanConfig))
}

func writeJavaScriptFixture(t *testing.T, root string, relativePath string, content string) {
	t.Helper()

	absolutePath := filepath.Join(root, filepath.FromSlash(relativePath))
	if err := os.MkdirAll(filepath.Dir(absolutePath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(absolutePath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func applyJavaScriptReplacements(content string, replacements []adapterproto.Replacement, file string) string {
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

func findJavaScriptReplacement(replacements []adapterproto.Replacement, file string, reason string) (adapterproto.Replacement, bool) {
	for _, replacement := range replacements {
		if replacement.File == file && replacement.Reason == reason {
			return replacement, true
		}
	}
	return adapterproto.Replacement{}, false
}

func containsJavaScriptString(values []string, expected string) bool {
	for _, value := range values {
		if value == expected {
			return true
		}
	}
	return false
}

func hasJavaScriptWarning(warnings []adapterproto.Warning, file string, message string) bool {
	for _, warning := range warnings {
		if warning.File == file && warning.Message == message {
			return true
		}
	}
	return false
}
