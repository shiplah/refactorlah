package validation

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"refactorlah/internal/reporting"
)

var ErrValidationFailed = errors.New("validation failed")

const newValidationFailureMessage = "failed after refactor; new validator output detected"
const unchangedValidationFailureMessage = "failed before and after refactor; no new validator output detected"

type RunOptions struct {
	SkipValidation bool
	RunTests       bool
}

type Runner struct{}

func NewRunner() *Runner {
	return &Runner{}
}

func (r *Runner) Run(ctx context.Context, projectRoot string, options RunOptions) ([]reporting.ValidationResult, error) {
	return r.RunCompared(ctx, projectRoot, options, nil)
}

func (r *Runner) Baseline(ctx context.Context, projectRoot string, options RunOptions) []reporting.ValidationResult {
	if options.SkipValidation {
		return nil
	}

	results := []reporting.ValidationResult{}
	for _, check := range r.readOnlyChecks(projectRoot, options) {
		result, _ := runCommand(ctx, projectRoot, check.name, check.args...)
		if result.Status == "failed" {
			result.Message = "failed before refactor"
		}
		results = append(results, result)
	}

	return results
}

func (r *Runner) RunCompared(ctx context.Context, projectRoot string, options RunOptions, baseline []reporting.ValidationResult) ([]reporting.ValidationResult, error) {
	results := []reporting.ValidationResult{}

	if _, err := os.Stat(filepath.Join(projectRoot, "composer.json")); err == nil && composerAvailable() {
		result, err := runCommand(ctx, projectRoot, "composer dump-autoload", "composer", "dump-autoload")
		results = append(results, result)
		if err != nil {
			return results, ErrValidationFailed
		}
	}

	if options.SkipValidation {
		results = append(results, reporting.ValidationResult{
			Name:    "validation",
			Status:  "skipped",
			Message: "validation disabled by --no-validation",
		})
		return results, nil
	}

	baselineIndex := indexResults(baseline)
	for _, check := range r.readOnlyChecks(projectRoot, options) {
		result, runErr := runCommand(ctx, projectRoot, check.name, check.args...)
		if runErr != nil {
			if sameFailureOutput(result, baselineIndex[result.Name]) {
				result.Status = "unchanged-failure"
				result.Message = unchangedValidationFailureMessage
				results = append(results, result)
				continue
			}

			result.Message = newValidationFailureMessage
			results = append(results, result)
			return results, ErrValidationFailed
		}

		results = append(results, result)
	}

	return results, nil
}

type validationCheck struct {
	name string
	args []string
}

func (r *Runner) readOnlyChecks(projectRoot string, options RunOptions) []validationCheck {
	checks := []validationCheck{}

	if _, err := os.Stat(filepath.Join(projectRoot, "vendor", "bin", "phpstan")); err == nil {
		checks = append(checks, validationCheck{name: "phpstan", args: []string{filepath.Join(projectRoot, "vendor", "bin", "phpstan")}})
	}

	if _, err := os.Stat(filepath.Join(projectRoot, "vendor", "bin", "psalm")); err == nil {
		checks = append(checks, validationCheck{name: "psalm", args: []string{filepath.Join(projectRoot, "vendor", "bin", "psalm")}})
	}

	if options.RunTests && composerHasTestScript(projectRoot) {
		checks = append(checks, validationCheck{name: "composer test", args: []string{"composer", "test"}})
	}

	return checks
}

func runCommand(ctx context.Context, dir string, name string, args ...string) (reporting.ValidationResult, error) {
	command := exec.CommandContext(ctx, args[0], args[1:]...)
	command.Dir = dir
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr

	result := reporting.ValidationResult{
		Name: name,
	}
	if err := command.Run(); err != nil {
		result.Status = "failed"
		result.Message = "failed"
		result.Stdout = stdout.String()
		result.Stderr = stderr.String()
		return result, err
	}

	result.Status = "ok"
	result.Message = "passed"
	result.Stdout = stdout.String()
	result.Stderr = stderr.String()
	return result, nil
}

func indexResults(results []reporting.ValidationResult) map[string]reporting.ValidationResult {
	index := map[string]reporting.ValidationResult{}
	for _, result := range results {
		index[result.Name] = result
	}

	return index
}

func sameFailureOutput(current reporting.ValidationResult, baseline reporting.ValidationResult) bool {
	return baseline.Status == "failed" &&
		current.Status == "failed" &&
		strings.TrimSpace(current.Stdout) == strings.TrimSpace(baseline.Stdout) &&
		strings.TrimSpace(current.Stderr) == strings.TrimSpace(baseline.Stderr)
}
