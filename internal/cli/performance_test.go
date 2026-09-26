package cli

import (
	"bytes"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/shiplah/refactorlah/internal/adapters/scan"
)

func TestPerformanceDiagnosticsSuggestsActualPHPStanCachePath(t *testing.T) {
	stats := scan.CandidateStats{
		Files:        18,
		LargestFile:  "storage/quality/phpstan/resultCache.php",
		LargestBytes: 11 << 20,
	}
	if diagnostics := performanceDiagnostics(slowAnalysisThreshold-time.Nanosecond, stats); len(diagnostics) != 0 {
		t.Fatalf("unexpected fast-run diagnostic: %#v", diagnostics)
	}

	diagnostics := performanceDiagnostics(12*time.Second, stats)
	if len(diagnostics) != 1 || diagnostics[0].Code != "large-candidate-file" {
		t.Fatalf("unexpected diagnostics: %#v", diagnostics)
	}
	diagnostic := diagnostics[0]
	if diagnostic.File != stats.LargestFile || diagnostic.FileBytes != stats.LargestBytes || diagnostic.DurationMS != 12000 || diagnostic.CandidateFiles != stats.Files {
		t.Fatalf("unexpected diagnostic details: %#v", diagnostic)
	}
	if !strings.Contains(diagnostic.Message, `"storage/quality/phpstan/**"`) || !strings.Contains(diagnostic.Message, "if generated") {
		t.Fatalf("expected conditional path-specific advice, got %q", diagnostic.Message)
	}
}

func TestPerformanceDiagnosticsDoesNotAssumeEverySlowScanIsACache(t *testing.T) {
	diagnostics := performanceDiagnostics(6*time.Second, scan.CandidateStats{Files: 24})
	if len(diagnostics) != 1 || diagnostics[0].Code != "slow-reference-analysis" {
		t.Fatalf("unexpected diagnostics: %#v", diagnostics)
	}
	if strings.Contains(strings.ToLower(diagnostics[0].Message), "cache") {
		t.Fatalf("unexpected cache claim: %q", diagnostics[0].Message)
	}
}

func TestPHPStanCachePattern(t *testing.T) {
	tests := []struct {
		file    string
		pattern string
	}{
		{"storage/quality/phpstan/resultCache.php", "storage/quality/phpstan/**"},
		{"var/phpstan/cache/compiled.php", "var/phpstan/**"},
		{"src/phpstan/Analyzer.php", ""},
		{"src/resultCache.php", ""},
	}
	for _, test := range tests {
		pattern, ok := phpstanCachePattern(test.file)
		if pattern != test.pattern || ok != (test.pattern != "") {
			t.Errorf("phpstanCachePattern(%q) = %q, %t; want %q", test.file, pattern, ok, test.pattern)
		}
	}
}

func TestReportSlowAnalysisWritesStatusAndStops(t *testing.T) {
	writer := &notifyingWriter{wrote: make(chan struct{})}
	done := make(chan struct{})
	stopped := make(chan struct{})
	go reportSlowAnalysis(writer, done, stopped, time.Now(), 0, time.Hour)
	select {
	case <-writer.wrote:
	case <-time.After(time.Second):
		t.Fatal("expected delayed progress message")
	}
	close(done)
	<-stopped
	if !strings.Contains(writer.String(), "analysing references...") {
		t.Fatalf("unexpected progress output: %q", writer.String())
	}
}

type notifyingWriter struct {
	bytes.Buffer
	once  sync.Once
	wrote chan struct{}
}

func (w *notifyingWriter) Write(content []byte) (int, error) {
	w.once.Do(func() { close(w.wrote) })
	return w.Buffer.Write(content)
}
