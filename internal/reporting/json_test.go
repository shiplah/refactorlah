package reporting

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestRenderJSONIncludesStructuredPerformanceDiagnostic(t *testing.T) {
	var output bytes.Buffer
	if err := RenderJSON(&output, Result{
		DryRun: true,
		Diagnostics: []Diagnostic{{
			Code:           "large-candidate-file",
			Message:        "Large candidate may slow parsing.",
			File:           "var/phpstan/resultCache.php",
			DurationMS:     12000,
			CandidateFiles: 18,
			FileBytes:      11 << 20,
		}},
	}); err != nil {
		t.Fatal(err)
	}
	var report struct {
		Warnings    []Message    `json:"warnings"`
		Diagnostics []Diagnostic `json:"diagnostics"`
	}
	if err := json.Unmarshal(output.Bytes(), &report); err != nil {
		t.Fatal(err)
	}
	if len(report.Diagnostics) != 1 || report.Diagnostics[0].Code != "large-candidate-file" || report.Diagnostics[0].FileBytes != 11<<20 || report.Diagnostics[0].CandidateFiles != 18 {
		t.Fatalf("unexpected diagnostics: %#v", report.Diagnostics)
	}
	if len(report.Warnings) != 0 {
		t.Fatalf("unexpected semantic warnings: %#v", report.Warnings)
	}
}
