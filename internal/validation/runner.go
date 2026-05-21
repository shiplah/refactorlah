package validation

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"

	"refactorlah/internal/reporting"
)

var ErrValidationFailed = errors.New("validation failed")

const postApplyValidationDisclaimer = "failed; post-apply validation only, pre-existing project issues are not baselined yet"

type RunOptions struct {
	SkipValidation bool
	RunTests       bool
}

type Runner struct{}

func NewRunner() *Runner {
	return &Runner{}
}

func (r *Runner) Run(ctx context.Context, projectRoot string, options RunOptions) ([]reporting.ValidationResult, error) {
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

	if _, err := os.Stat(filepath.Join(projectRoot, "vendor", "bin", "phpstan")); err == nil {
		result, runErr := runCommand(ctx, projectRoot, "phpstan", filepath.Join(projectRoot, "vendor", "bin", "phpstan"))
		results = append(results, result)
		if runErr != nil {
			return results, ErrValidationFailed
		}
	}

	if _, err := os.Stat(filepath.Join(projectRoot, "vendor", "bin", "psalm")); err == nil {
		result, runErr := runCommand(ctx, projectRoot, "psalm", filepath.Join(projectRoot, "vendor", "bin", "psalm"))
		results = append(results, result)
		if runErr != nil {
			return results, ErrValidationFailed
		}
	}

	if options.RunTests && composerHasTestScript(projectRoot) {
		result, err := runCommand(ctx, projectRoot, "composer test", "composer", "test")
		results = append(results, result)
		if err != nil {
			return results, ErrValidationFailed
		}
	}

	return results, nil
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
		result.Message = postApplyValidationDisclaimer
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
