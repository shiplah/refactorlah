package validation

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/shiplah/refactorlah/internal/reporting"
)

var ErrValidationFailed = errors.New("validation failed")

type RunOptions struct {
	SkipValidation bool
	RunTests       bool
}

type Check struct {
	Directory string
	Command   []string
}

type Runner struct{}

func NewRunner() *Runner {
	return &Runner{}
}

func ChecksFromCommands(commands [][]string) []Check {
	checks := make([]Check, 0, len(commands))
	for _, command := range commands {
		if len(command) == 0 {
			continue
		}
		checks = append(checks, Check{Command: append([]string(nil), command...)})
	}
	return checks
}

func (r *Runner) Plan(checks []Check, tests []Check, options RunOptions) []reporting.ValidationResult {
	if options.SkipValidation {
		return []reporting.ValidationResult{validationSkippedResult()}
	}

	results := make([]reporting.ValidationResult, 0, len(checks)+len(tests)+1)
	for _, check := range checks {
		results = append(results, reporting.ValidationResult{
			Name:    check.DisplayName(),
			Status:  "planned",
			Message: "would run",
		})
	}
	if options.RunTests {
		for _, check := range tests {
			results = append(results, reporting.ValidationResult{
				Name:    check.DisplayName(),
				Status:  "planned",
				Message: "would run",
			})
		}
	} else if len(tests) > 0 {
		results = append(results, testsSkippedResult())
	}

	return results
}

func (r *Runner) Run(ctx context.Context, projectRoot string, checks []Check, tests []Check, options RunOptions) ([]reporting.ValidationResult, error) {
	if options.SkipValidation {
		return []reporting.ValidationResult{validationSkippedResult()}, nil
	}

	toRun := append([]Check(nil), checks...)
	if options.RunTests {
		toRun = append(toRun, tests...)
	}

	results := make([]reporting.ValidationResult, 0, len(toRun)+1)
	for _, check := range toRun {
		result, err := runCommand(ctx, projectRoot, check)
		results = append(results, result)
		if err != nil {
			return results, ErrValidationFailed
		}
	}

	if !options.RunTests && len(tests) > 0 {
		results = append(results, testsSkippedResult())
	}

	return results, nil
}

func validationSkippedResult() reporting.ValidationResult {
	return reporting.ValidationResult{
		Name:    "validation",
		Status:  "skipped",
		Message: "validation disabled by --no-validation",
	}
}

func testsSkippedResult() reporting.ValidationResult {
	return reporting.ValidationResult{
		Name:    "tests",
		Status:  "skipped",
		Message: "skipped; pass --run-tests to run configured tests",
	}
}

func runCommand(ctx context.Context, projectRoot string, check Check) (reporting.ValidationResult, error) {
	result := reporting.ValidationResult{
		Name: check.DisplayName(),
	}
	if len(check.Command) == 0 {
		result.Status = "failed"
		result.Message = "empty validation command"
		return result, errors.New("empty validation command")
	}

	dir, err := check.WorkingDirectory(projectRoot)
	if err != nil {
		result.Status = "failed"
		result.Message = err.Error()
		return result, err
	}

	command := exec.CommandContext(ctx, check.Command[0], check.Command[1:]...)
	command.Dir = dir
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr

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

func (c Check) DisplayName() string {
	name := strings.Join(c.Command, " ")
	if c.Directory == "" {
		return name
	}
	return fmt.Sprintf("%s: %s", c.Directory, name)
}

func (c Check) WorkingDirectory(projectRoot string) (string, error) {
	if c.Directory == "" {
		return projectRoot, nil
	}
	if filepath.IsAbs(c.Directory) {
		return "", fmt.Errorf("validation directory must be project-relative: %s", c.Directory)
	}

	cleaned := filepath.Clean(filepath.FromSlash(c.Directory))
	path := filepath.Join(projectRoot, cleaned)
	relative, err := filepath.Rel(projectRoot, path)
	if err != nil {
		return "", err
	}
	if relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) || filepath.IsAbs(relative) {
		return "", fmt.Errorf("validation directory is outside project root: %s", c.Directory)
	}
	return path, nil
}
