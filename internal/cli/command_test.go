package cli

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"refactorlah/internal/languages/native"
)

func TestDryRunWritesNothing(t *testing.T) {
	root := plainProject(t, "app/Services/Billing/InvoiceService.php")
	command := NewCommand()

	report, exitCode := command.runWithOptions(t.Context(), root, Options{
		OldPath: "app/Services/Billing/InvoiceService.php",
		NewPath: "app/Domain/Billing/InvoiceService.php",
		DryRun:  true,
		Format:  FormatText,
	}, io.Discard)
	if exitCode != ExitSuccess {
		t.Fatalf("unexpected exit code: %d %#v", exitCode, report.Errors)
	}

	content, err := os.ReadFile(filepath.Join(root, "app", "Services", "Billing", "InvoiceService.php"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(content, []byte("fixture: app/Services/Billing/InvoiceService.php")) {
		t.Fatal("fixture file changed during dry-run")
	}
}

func TestDefaultModeAppliesChanges(t *testing.T) {
	root := plainProject(t, "app/Services/Billing/InvoiceService.php")
	command := NewCommand()

	report, exitCode := command.runWithOptions(t.Context(), root, Options{
		OldPath:      "app/Services/Billing/InvoiceService.php",
		NewPath:      "app/Domain/Billing/InvoiceService.php",
		Apply:        true,
		NoValidation: true,
		Format:       FormatText,
	}, io.Discard)
	if exitCode != ExitSuccess {
		t.Fatalf("unexpected exit code: %d %#v", exitCode, report.Errors)
	}

	if _, err := os.Stat(filepath.Join(root, "app", "Domain", "Billing", "InvoiceService.php")); err != nil {
		t.Fatalf("moved file missing: %v", err)
	}
}

func TestJSONOutputIsValidAndUnpolluted(t *testing.T) {
	root := plainProject(t, "app/Services/Billing/InvoiceService.php")
	command := NewRootCommand()

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.Chdir(cwd)
	}()
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := command.Run(t.Context(), []string{
		"move",
		"--dry",
		"app/Services/Billing/InvoiceService.php",
		"app/Domain/Billing/InvoiceService.php",
		"--format=json",
	}, &stdout, &stderr)
	if exitCode != ExitSuccess {
		t.Fatalf("unexpected exit code: %d stderr=%s", exitCode, stderr.String())
	}
	if !bytes.HasPrefix(bytes.TrimSpace(stdout.Bytes()), []byte("{")) {
		t.Fatalf("stdout is not JSON: %s", stdout.String())
	}
}

func TestApplyMovesFixtureFile(t *testing.T) {
	root := plainProject(t, "app/Services/Billing/InvoiceService.php")
	command := NewCommand()

	report, exitCode := command.runWithOptions(t.Context(), root, Options{
		OldPath:      "app/Services/Billing/InvoiceService.php",
		NewPath:      "app/Domain/Billing/InvoiceService.php",
		Apply:        true,
		NoValidation: true,
		Format:       FormatText,
	}, io.Discard)
	if exitCode != ExitSuccess {
		t.Fatalf("unexpected exit code: %d %#v", exitCode, report.Errors)
	}

	if _, err := os.Stat(filepath.Join(root, "app", "Domain", "Billing", "InvoiceService.php")); err != nil {
		t.Fatalf("moved file missing: %v", err)
	}
}

func TestApplyGoPackageMoveUpdatesPackageReferences(t *testing.T) {
	root := plainProject(t)
	mustWriteFile(t, filepath.Join(root, "go.mod"), "module refactorlah\n")
	mustWriteFile(t, filepath.Join(root, "internal", "languages", "php", "parser.go"), `package php

import "refactorlah/internal/languages/treesitter"

func Parse() {
	treesitter.Parse()
}
`)
	mustWriteFile(t, filepath.Join(root, "internal", "languages", "treesitter", "document.go"), `package treesitter

func Parse() {}
`)

	command := NewCommand()
	report, exitCode := command.runWithOptions(t.Context(), root, Options{
		OldPath:      "internal/languages/treesitter",
		NewPath:      "internal/parsing/document",
		Apply:        true,
		NoValidation: true,
		Format:       FormatText,
	}, io.Discard)
	if exitCode != ExitSuccess {
		t.Fatalf("unexpected exit code: %d %#v", exitCode, report.Errors)
	}
	if !hasString(report.AutoDetectedAdapters, "go") {
		t.Fatalf("expected go semantic source, got %#v", report.AutoDetectedAdapters)
	}

	parser := mustReadFile(t, filepath.Join(root, "internal", "languages", "php", "parser.go"))
	if !strings.Contains(parser, `"refactorlah/internal/parsing/document"`) {
		t.Fatalf("expected Go import path rewrite, got:\n%s", parser)
	}
	if !strings.Contains(parser, "document.Parse()") {
		t.Fatalf("expected Go package qualifier rewrite, got:\n%s", parser)
	}
	movedFile := mustReadFile(t, filepath.Join(root, "internal", "parsing", "document", "document.go"))
	if !strings.Contains(movedFile, "package document") {
		t.Fatalf("expected moved Go package declaration rewrite, got:\n%s", movedFile)
	}
	if _, err := os.Stat(filepath.Join(root, "internal", "parsing", "document", "document.go")); err != nil {
		t.Fatalf("moved Go file missing: %v", err)
	}
}

func TestApplyFailsClearlyWhenRelevantAdapterIsUnavailable(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell script git shim is unix-only")
	}

	root := copyFixture(t)
	binDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(binDir, "git"), []byte("#!/bin/sh\nexit 1\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", binDir)
	t.Setenv("REFACTORLAH_PHP_ADAPTER", "")

	command := NewCommand()
	command.nativeAnalyzers = native.EmptyRegistry()
	report, exitCode := command.runWithOptions(t.Context(), root, Options{
		OldPath:      "app/Services/Billing/InvoiceService.php",
		NewPath:      "app/Domain/Billing/InvoiceService.php",
		Apply:        true,
		NoValidation: true,
		Format:       FormatText,
	}, io.Discard)
	if exitCode != ExitAdapterFailure {
		t.Fatalf("unexpected exit code: %d %#v", exitCode, report.Errors)
	}
	if len(report.Errors) != 1 {
		t.Fatalf("expected one error, got %#v", report.Errors)
	}
	if !strings.Contains(report.Errors[0].Message, "native PHP/Twig support is unavailable") {
		t.Fatalf("expected install guidance, got: %s", report.Errors[0].Message)
	}
	if _, err := os.Stat(filepath.Join(root, "app", "Services", "Billing", "InvoiceService.php")); err != nil {
		t.Fatalf("source file should not have moved: %v", err)
	}
}

func TestHelpShowsUsageWithoutError(t *testing.T) {
	command := NewRootCommand()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := command.Run(t.Context(), []string{"--help"}, &stdout, &stderr)
	if exitCode != ExitSuccess {
		t.Fatalf("unexpected exit code: %d", exitCode)
	}
	if !strings.Contains(stdout.String(), "Usage:") {
		t.Fatalf("expected usage output, got: %s", stdout.String())
	}
	if !strings.Contains(stdout.String(), "Commands:") {
		t.Fatalf("expected root command list, got: %s", stdout.String())
	}
	if strings.Contains(stdout.String(), "refactorlah <old-path> <new-path>") {
		t.Fatalf("did not expect shorthand usage in help: %s", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("expected empty stderr, got: %s", stderr.String())
	}
}

func TestNoArgsShowsUsageAndError(t *testing.T) {
	command := NewRootCommand()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := command.Run(t.Context(), nil, &stdout, &stderr)
	if exitCode != ExitInvalidArguments {
		t.Fatalf("unexpected exit code: %d", exitCode)
	}
	if !strings.Contains(stderr.String(), "Usage:") {
		t.Fatalf("expected usage output, got: %s", stderr.String())
	}
	if !strings.Contains(stderr.String(), "error: expected command") {
		t.Fatalf("expected missing-args error, got: %s", stderr.String())
	}
	if strings.Index(stderr.String(), "error: expected command") > strings.Index(stderr.String(), "Usage:") {
		t.Fatalf("expected error before usage, got: %s", stderr.String())
	}
	if !strings.Contains(stderr.String(), "Commands:") {
		t.Fatalf("expected command list, got: %s", stderr.String())
	}
	if stdout.Len() != 0 {
		t.Fatalf("expected empty stdout, got: %s", stdout.String())
	}
}

func TestInvalidFlagShowsErrorAboveUsage(t *testing.T) {
	command := NewRootCommand()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := command.Run(t.Context(), []string{"move", "--apply"}, &stdout, &stderr)
	if exitCode != ExitInvalidArguments {
		t.Fatalf("unexpected exit code: %d", exitCode)
	}
	if !strings.Contains(stderr.String(), "error: flag provided but not defined: -apply") {
		t.Fatalf("expected invalid flag error, got: %s", stderr.String())
	}
	if !strings.Contains(stderr.String(), "Usage:") {
		t.Fatalf("expected usage output, got: %s", stderr.String())
	}
	if strings.Index(stderr.String(), "error: flag provided but not defined: -apply") > strings.Index(stderr.String(), "Usage:") {
		t.Fatalf("expected error before usage, got: %s", stderr.String())
	}
	if stdout.Len() != 0 {
		t.Fatalf("expected empty stdout, got: %s", stdout.String())
	}
}

func TestMoveHelpShowsMoveOptions(t *testing.T) {
	command := NewRootCommand()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := command.Run(t.Context(), []string{"move", "--help"}, &stdout, &stderr)
	if exitCode != ExitSuccess {
		t.Fatalf("unexpected exit code: %d", exitCode)
	}
	if !strings.Contains(stdout.String(), "--require-clean-worktree") {
		t.Fatalf("expected move options in help: %s", stdout.String())
	}
	if strings.Contains(stdout.String(), "--require-clean ") {
		t.Fatalf("did not expect old require-clean flag in help: %s", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("expected empty stderr, got: %s", stderr.String())
	}
}

func TestMoveSubcommandDelegatesToMoveCommand(t *testing.T) {
	root := plainProject(t, "app/Services/Billing/InvoiceService.php")
	command := NewRootCommand()

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.Chdir(cwd)
	}()
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := command.Run(t.Context(), []string{
		"move",
		"--dry",
		"app/Services/Billing/InvoiceService.php",
		"app/Domain/Billing/InvoiceService.php",
	}, &stdout, &stderr)
	if exitCode != ExitSuccess {
		t.Fatalf("unexpected exit code: %d stderr=%s", exitCode, stderr.String())
	}
	if !strings.Contains(stdout.String(), "Mode: dry") {
		t.Fatalf("expected dry-run output, got: %s", stdout.String())
	}
}

func TestMoveSubcommandSupportsUseListPairs(t *testing.T) {
	root := plainProject(t,
		"app/Services/Billing/InvoiceService.php",
		"tests/Feature/BillingTest.php",
	)
	command := NewRootCommand()

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.Chdir(cwd)
	}()
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := command.Run(t.Context(), []string{
		"move",
		"--dry",
		"--use-list",
		"app/Services/Billing/InvoiceService.php,app/Domain/Billing/InvoiceService.php",
		"tests/Feature/BillingTest.php,tests/Feature/BillingTestMoved.php",
	}, &stdout, &stderr)
	if exitCode != ExitSuccess {
		t.Fatalf("unexpected exit code: %d stderr=%s", exitCode, stderr.String())
	}
	if !strings.Contains(stdout.String(), "app/Services/Billing/InvoiceService.php -> app/Domain/Billing/InvoiceService.php") {
		t.Fatalf("expected first move in output, got: %s", stdout.String())
	}
	if !strings.Contains(stdout.String(), "tests/Feature/BillingTest.php -> tests/Feature/BillingTestMoved.php") {
		t.Fatalf("expected second move in output, got: %s", stdout.String())
	}
}

func TestMoveSubcommandAllowsLaterPairInsideEarlierTarget(t *testing.T) {
	root := plainProject(t, "app/Services/Billing/InvoiceService.php")
	command := NewRootCommand()

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.Chdir(cwd)
	}()
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := command.Run(t.Context(), []string{
		"move",
		"--dry",
		"--use-list",
		"app/Services/Billing,app/Domain/Billing",
		"app/Domain/Billing/InvoiceService.php,app/Domain/Billing/BillingService.php",
	}, &stdout, &stderr)
	if exitCode != ExitSuccess {
		t.Fatalf("unexpected exit code: %d stderr=%s", exitCode, stderr.String())
	}
	if !strings.Contains(stdout.String(), "app/Services/Billing/InvoiceService.php -> app/Domain/Billing/BillingService.php") {
		t.Fatalf("expected refined move in output, got: %s", stdout.String())
	}
}

func TestMoveSubcommandSupportsUseFile(t *testing.T) {
	root := plainProject(t, "app/Services/Billing/InvoiceService.php")
	command := NewRootCommand()

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.Chdir(cwd)
	}()
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "moves.txt"), []byte("app/Services/Billing/InvoiceService.php,app/Domain/Billing/InvoiceService.php\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := command.Run(t.Context(), []string{
		"move",
		"--dry",
		"--use-file",
		"moves.txt",
	}, &stdout, &stderr)
	if exitCode != ExitSuccess {
		t.Fatalf("unexpected exit code: %d stderr=%s", exitCode, stderr.String())
	}
	if !strings.Contains(stdout.String(), "app/Services/Billing/InvoiceService.php -> app/Domain/Billing/InvoiceService.php") {
		t.Fatalf("expected move in output, got: %s", stdout.String())
	}
}

func TestMoveSubcommandExpandsWildcardPairs(t *testing.T) {
	root := plainProject(t, "app/Services/Billing/InvoiceService.php")
	command := NewRootCommand()

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.Chdir(cwd)
	}()
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := command.Run(t.Context(), []string{
		"move",
		"--dry",
		"app/Services/Billing/*Service.php",
		"app/Domain/Billing/*Service.php",
	}, &stdout, &stderr)
	if exitCode != ExitSuccess {
		t.Fatalf("unexpected exit code: %d stderr=%s", exitCode, stderr.String())
	}
	if !strings.Contains(stdout.String(), "app/Services/Billing/InvoiceService.php -> app/Domain/Billing/InvoiceService.php") {
		t.Fatalf("expected wildcard-expanded move in output, got: %s", stdout.String())
	}
}

func TestDirectMoveWithoutCommandIsRejected(t *testing.T) {
	root := copyFixture(t)
	command := NewRootCommand()

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.Chdir(cwd)
	}()
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := command.Run(t.Context(), []string{
		"app/Services/Billing/InvoiceService.php",
		"app/Domain/Billing/InvoiceService.php",
	}, &stdout, &stderr)
	if exitCode != ExitInvalidArguments {
		t.Fatalf("unexpected exit code: %d", exitCode)
	}
	if !strings.Contains(stderr.String(), "error: unknown command \"app/Services/Billing/InvoiceService.php\"") {
		t.Fatalf("expected unknown-command error, got: %s", stderr.String())
	}
	if stdout.Len() != 0 {
		t.Fatalf("expected empty stdout, got: %s", stdout.String())
	}
}

func TestApplyDoesNotStageSemanticEdits(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell script adapter helper is unix-only")
	}

	root := copyFixture(t)
	runGitForCliTest(t, root, "init")
	runGitForCliTest(t, root, "config", "user.email", "test@example.com")
	runGitForCliTest(t, root, "config", "user.name", "Test User")
	runGitForCliTest(t, root, "add", ".")
	runGitForCliTest(t, root, "commit", "-m", "initial")

	targetFile := "app/Http/Controllers/InvoiceController.php"
	controllerPath := filepath.Join(root, filepath.FromSlash(targetFile))
	controllerContent, err := os.ReadFile(controllerPath)
	if err != nil {
		t.Fatal(err)
	}
	oldImport := "use App\\Services\\Billing\\InvoiceService;"
	start := strings.Index(string(controllerContent), oldImport)
	if start < 0 {
		t.Fatalf("expected fixture to contain %q", oldImport)
	}

	adapterPath := filepath.Join(root, "fake-refactorlah-php")
	response, err := json.Marshal(map[string]any{
		"protocolVersion": 1,
		"adapter":         "php",
		"symbolMappings":  []any{},
		"pathMappings":    []any{},
		"replacements": []map[string]any{{
			"file":        targetFile,
			"start":       start,
			"end":         start + len(oldImport),
			"replacement": "use App\\Domain\\Billing\\InvoiceService;",
			"reason":      "php-use-statement",
		}},
		"warnings": []any{},
		"errors":   []any{},
	})
	if err != nil {
		t.Fatal(err)
	}
	script := "#!/bin/sh\ncat >/dev/null\ncat <<'JSON'\n" + string(response) + "\nJSON\n"
	if err := os.WriteFile(adapterPath, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	mustWriteFile(t, filepath.Join(root, "adapter.json"), `{
  "name": "php",
  "executable": "refactorlah-php",
  "runtime": {
    "command": "php",
    "minimumVersion": "8.2.0"
  }
}`)

	previousAdapterPath := os.Getenv("REFACTORLAH_PHP_ADAPTER")
	if err := os.Setenv("REFACTORLAH_PHP_ADAPTER", adapterPath); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if previousAdapterPath == "" {
			_ = os.Unsetenv("REFACTORLAH_PHP_ADAPTER")
			return
		}
		_ = os.Setenv("REFACTORLAH_PHP_ADAPTER", previousAdapterPath)
	}()

	command := NewCommand()
	command.nativeAnalyzers = native.EmptyRegistry()
	report, exitCode := command.runWithOptions(t.Context(), root, Options{
		OldPath:      "app/Services/Billing/InvoiceService.php",
		NewPath:      "app/Domain/Billing/InvoiceService.php",
		Apply:        true,
		NoValidation: true,
		Format:       FormatText,
	}, io.Discard)
	if exitCode != ExitSuccess {
		t.Fatalf("unexpected exit code: %d %#v", exitCode, report.Errors)
	}

	status := runGitForCliTestOutput(t, root, "status", "--short")
	if !strings.Contains(status, "R  app/Services/Billing/InvoiceService.php -> app/Domain/Billing/InvoiceService.php") {
		t.Fatalf("expected git mv rename to remain staged, got status:\n%s", status)
	}
	if !strings.Contains(status, " M app/Http/Controllers/InvoiceController.php") {
		t.Fatalf("expected semantic edit to remain unstaged, got status:\n%s", status)
	}
	if strings.Contains(status, "M  app/Http/Controllers/InvoiceController.php") {
		t.Fatalf("semantic edit was staged unexpectedly:\n%s", status)
	}
}

func copyFixture(t *testing.T) string {
	t.Helper()
	return copyNamedFixture(t, filepath.Join("adapters", "php", "tests", "fixtures", "php-basic"))
}

func plainProject(t *testing.T, files ...string) string {
	t.Helper()
	root := t.TempDir()
	for _, file := range files {
		mustWriteFile(t, filepath.Join(root, filepath.FromSlash(file)), "fixture: "+file+"\n")
	}
	return root
}

func copyNamedFixture(t *testing.T, source string) string {
	t.Helper()
	root := t.TempDir()
	sourceRoot := filepath.Join("..", "..", source)
	if runtime.GOOS == "windows" {
		sourceRoot = filepath.Clean(sourceRoot)
	}
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
		t.Fatalf("copy fixture: %v", err)
	}
	return root
}

func mustReadFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func mustWriteFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func hasString(values []string, expected string) bool {
	for _, value := range values {
		if value == expected {
			return true
		}
	}
	return false
}

func runGitForCliTest(t *testing.T, dir string, args ...string) {
	t.Helper()
	if output, err := runGitForCliTestCommand(dir, args...); err != nil {
		t.Fatalf("git %v failed: %v: %s", args, err, output)
	}
}

func runGitForCliTestOutput(t *testing.T, dir string, args ...string) string {
	t.Helper()
	output, err := runGitForCliTestCommand(dir, args...)
	if err != nil {
		t.Fatalf("git %v failed: %v: %s", args, err, output)
	}
	return output
}

func runGitForCliTestCommand(dir string, args ...string) (string, error) {
	command := exec.Command("git", args...)
	command.Dir = dir
	output, err := command.CombinedOutput()
	return string(output), err
}
