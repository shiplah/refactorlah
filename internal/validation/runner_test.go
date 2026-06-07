package validation

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunnerRunsConfiguredChecks(t *testing.T) {
	root := t.TempDir()
	results, err := NewRunner().Run(t.Context(), root, []Check{helperCheck("pass")}, nil, RunOptions{})
	if err != nil {
		t.Fatalf("expected check to pass, got %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 validation result, got %d", len(results))
	}
	if results[0].Status != "ok" || results[0].Message != "passed" {
		t.Fatalf("unexpected result: %#v", results[0])
	}
}

func TestRunnerReportsValidationFailure(t *testing.T) {
	root := t.TempDir()
	results, err := NewRunner().Run(t.Context(), root, []Check{helperCheck("fail")}, nil, RunOptions{})
	if !errors.Is(err, ErrValidationFailed) {
		t.Fatalf("expected validation failure, got %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 validation result, got %d", len(results))
	}
	if results[0].Status != "failed" {
		t.Fatalf("expected failed status, got %#v", results[0])
	}
	if !strings.Contains(results[0].Stderr, "helper failed") {
		t.Fatalf("expected captured stderr, got %#v", results[0])
	}
}

func TestRunnerRunsTestsOnlyWhenRequested(t *testing.T) {
	root := t.TempDir()
	tests := []Check{helperCheck("pass")}

	results, err := NewRunner().Run(t.Context(), root, nil, tests, RunOptions{})
	if err != nil {
		t.Fatalf("expected skipped tests not to fail, got %v", err)
	}
	if len(results) != 1 || results[0].Name != "tests" || results[0].Status != "skipped" {
		t.Fatalf("expected skipped tests result, got %#v", results)
	}

	results, err = NewRunner().Run(t.Context(), root, nil, tests, RunOptions{RunTests: true})
	if err != nil {
		t.Fatalf("expected tests to pass, got %v", err)
	}
	if len(results) != 1 || results[0].Name == "tests" || results[0].Status != "ok" {
		t.Fatalf("expected executed test result, got %#v", results)
	}
}

func TestRunnerPlansValidationCommands(t *testing.T) {
	results := NewRunner().Plan([]Check{{Command: []string{"composer", "stan"}}}, []Check{{Command: []string{"composer", "test"}}}, RunOptions{})
	if len(results) != 2 {
		t.Fatalf("expected check plan and skipped tests, got %#v", results)
	}
	if results[0].Name != "composer stan" || results[0].Message != "would run" {
		t.Fatalf("unexpected check plan: %#v", results[0])
	}
	if results[1].Name != "tests" || results[1].Status != "skipped" {
		t.Fatalf("unexpected tests plan: %#v", results[1])
	}
}

func TestRunnerUsesProjectRelativeWorkingDirectory(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "platform"), 0o755); err != nil {
		t.Fatal(err)
	}

	results, err := NewRunner().Run(t.Context(), root, []Check{{
		Directory: "platform",
		Command:   helperCommand("cwd"),
	}}, nil, RunOptions{})
	if err != nil {
		t.Fatalf("expected check to pass, got %v", err)
	}
	if !strings.Contains(filepath.ToSlash(results[0].Stdout), "/platform") {
		t.Fatalf("expected command to run in platform directory, got %#v", results[0])
	}
}

func TestRunnerSkipsValidationWhenDisabled(t *testing.T) {
	results, err := NewRunner().Run(t.Context(), t.TempDir(), []Check{helperCheck("fail")}, []Check{helperCheck("fail")}, RunOptions{
		SkipValidation: true,
		RunTests:       true,
	})
	if err != nil {
		t.Fatalf("expected skipped validation not to fail, got %v", err)
	}
	if len(results) != 1 || results[0].Name != "validation" || results[0].Status != "skipped" {
		t.Fatalf("unexpected skipped validation result: %#v", results)
	}
}

func TestValidationHelperCommand(t *testing.T) {
	mode, ok := helperMode()
	if !ok {
		return
	}

	switch mode {
	case "pass":
		_, _ = os.Stdout.WriteString("helper passed\n")
	case "fail":
		_, _ = os.Stderr.WriteString("helper failed\n")
		os.Exit(1)
	case "cwd":
		cwd, err := os.Getwd()
		if err != nil {
			os.Exit(1)
		}
		_, _ = os.Stdout.WriteString(cwd)
	default:
		os.Exit(2)
	}
	os.Exit(0)
}

func helperCheck(mode string) Check {
	return Check{Command: helperCommand(mode)}
}

func helperCommand(mode string) []string {
	return []string{os.Args[0], "-test.run=TestValidationHelperCommand", "--", mode}
}

func helperMode() (string, bool) {
	for index, argument := range os.Args {
		if argument == "--" && index+1 < len(os.Args) {
			return os.Args[index+1], true
		}
	}
	return "", false
}
