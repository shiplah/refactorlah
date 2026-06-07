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

	"refactorlah/internal/planning"
	"refactorlah/internal/reporting"
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

func TestApplyRunsConfiguredChecks(t *testing.T) {
	root := plainProject(t, "app/Services/Billing/InvoiceService.php")
	check := commandValidationHelperCommand("pass")
	writeRefactorlahConfig(t, root, [][]string{check}, nil)

	report, exitCode := NewCommand().runWithOptions(t.Context(), root, Options{
		OldPath: "app/Services/Billing/InvoiceService.php",
		NewPath: "app/Domain/Billing/InvoiceService.php",
		Apply:   true,
		Format:  FormatText,
	}, io.Discard)
	if exitCode != ExitSuccess {
		t.Fatalf("unexpected exit code: %d %#v", exitCode, report.Errors)
	}
	if !hasValidation(report.Validation, strings.Join(check, " "), "ok") {
		t.Fatalf("expected configured check to run, got %#v", report.Validation)
	}
}

func TestApplyRunsConfiguredTestsOnlyWhenRequested(t *testing.T) {
	root := plainProject(t, "app/Services/Billing/InvoiceService.php")
	testCommand := commandValidationHelperCommand("pass")
	writeRefactorlahConfig(t, root, nil, [][]string{testCommand})

	report, exitCode := NewCommand().runWithOptions(t.Context(), root, Options{
		OldPath: "app/Services/Billing/InvoiceService.php",
		NewPath: "app/Domain/Billing/InvoiceService.php",
		Apply:   true,
		Format:  FormatText,
	}, io.Discard)
	if exitCode != ExitSuccess {
		t.Fatalf("unexpected exit code: %d %#v", exitCode, report.Errors)
	}
	if !hasValidation(report.Validation, "tests", "skipped") {
		t.Fatalf("expected configured tests to be skipped, got %#v", report.Validation)
	}

	root = plainProject(t, "app/Services/Billing/InvoiceService.php")
	writeRefactorlahConfig(t, root, nil, [][]string{testCommand})
	report, exitCode = NewCommand().runWithOptions(t.Context(), root, Options{
		OldPath:  "app/Services/Billing/InvoiceService.php",
		NewPath:  "app/Domain/Billing/InvoiceService.php",
		Apply:    true,
		RunTests: true,
		Format:   FormatText,
	}, io.Discard)
	if exitCode != ExitSuccess {
		t.Fatalf("unexpected exit code with tests: %d %#v", exitCode, report.Errors)
	}
	if !hasValidation(report.Validation, strings.Join(testCommand, " "), "ok") {
		t.Fatalf("expected configured tests to run, got %#v", report.Validation)
	}
}

func TestApplyFailsOnConfiguredCheckFailure(t *testing.T) {
	root := plainProject(t, "app/Services/Billing/InvoiceService.php")
	writeRefactorlahConfig(t, root, [][]string{commandValidationHelperCommand("fail")}, nil)

	report, exitCode := NewCommand().runWithOptions(t.Context(), root, Options{
		OldPath: "app/Services/Billing/InvoiceService.php",
		NewPath: "app/Domain/Billing/InvoiceService.php",
		Apply:   true,
		Format:  FormatText,
	}, io.Discard)
	if exitCode != ExitValidationFailed {
		t.Fatalf("expected validation failure exit, got %d %#v", exitCode, report.Errors)
	}
	if len(report.Errors) != 1 || !strings.Contains(report.Errors[0].Message, "validation failed") {
		t.Fatalf("expected validation failure error, got %#v", report.Errors)
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
	mustWriteFile(t, filepath.Join(root, "internal", "adapters", "php", "parser.go"), `package php

import "refactorlah/internal/parsing/treesitter"

func Parse() {
	treesitter.Parse()
}
`)
	mustWriteFile(t, filepath.Join(root, "internal", "parsing", "treesitter", "document.go"), `package treesitter

func Parse() {}
`)

	command := NewCommand()
	report, exitCode := command.runWithOptions(t.Context(), root, Options{
		OldPath:      "internal/parsing/treesitter",
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

	parser := mustReadFile(t, filepath.Join(root, "internal", "adapters", "php", "parser.go"))
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

func TestApplyGoFileRenameUpdatesMatchingSymbolReferences(t *testing.T) {
	root := plainProject(t)
	mustWriteFile(t, filepath.Join(root, "go.mod"), "module refactorlah\n")
	mustWriteFile(t, filepath.Join(root, "internal", "models", "old_thing.go"), `package models

type OldThing struct{}

func (thing OldThing) Clone() OldThing {
	return OldThing{}
}
`)
	mustWriteFile(t, filepath.Join(root, "internal", "models", "use.go"), `package models

func Build(value OldThing) OldThing {
	return OldThing{}
}
`)
	mustWriteFile(t, filepath.Join(root, "internal", "consumer", "use.go"), `package consumer

import "refactorlah/internal/models"

func Build() models.OldThing {
	return models.OldThing{}
}
`)

	command := NewCommand()
	report, exitCode := command.runWithOptions(t.Context(), root, Options{
		OldPath:      "internal/models/old_thing.go",
		NewPath:      "internal/models/new_thing.go",
		Apply:        true,
		NoValidation: true,
		Format:       FormatText,
	}, io.Discard)
	if exitCode != ExitSuccess {
		t.Fatalf("unexpected exit code: %d %#v", exitCode, report.Errors)
	}

	movedFile := mustReadFile(t, filepath.Join(root, "internal", "models", "new_thing.go"))
	for _, expected := range []string{"type NewThing struct{}", "func (thing NewThing) Clone() NewThing", "return NewThing{}"} {
		if !strings.Contains(movedFile, expected) {
			t.Fatalf("expected %q in moved Go file, got:\n%s", expected, movedFile)
		}
	}

	samePackageConsumer := mustReadFile(t, filepath.Join(root, "internal", "models", "use.go"))
	if !strings.Contains(samePackageConsumer, "func Build(value NewThing) NewThing") {
		t.Fatalf("expected same-package symbol reference rewrite, got:\n%s", samePackageConsumer)
	}

	externalConsumer := mustReadFile(t, filepath.Join(root, "internal", "consumer", "use.go"))
	if !strings.Contains(externalConsumer, "models.NewThing") {
		t.Fatalf("expected imported symbol reference rewrite, got:\n%s", externalConsumer)
	}
}

func TestApplyGoBroadCorpusUpdatesPackageAndSymbolReferences(t *testing.T) {
	root := copyNamedFixture(t, filepath.Join("tests", "fixtures", "native-mixed"))
	report, exitCode := NewCommand().runWithOptions(t.Context(), root, Options{
		MoveRequests: []planning.RequestedMove{
			{OldPath: "internal/oldpkg/old_service.go", NewPath: "internal/newpkg/new_service.go"},
			{OldPath: "internal/oldpkg/old_worker.go", NewPath: "internal/newpkg/new_worker.go"},
			{OldPath: "internal/oldpkg/helper.go", NewPath: "internal/newpkg/helper.go"},
			{OldPath: "internal/oldpkg/service_test.go", NewPath: "internal/newpkg/service_test.go"},
		},
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
	for _, warning := range report.Warnings {
		if warning.File == "internal/unrelated/broken.go" {
			t.Fatalf("expected unrelated broken Go file to avoid parsing warnings, got %#v", warning)
		}
	}

	service := mustReadFile(t, filepath.Join(root, "internal", "newpkg", "new_service.go"))
	for _, expected := range []string{"package newpkg", "type NewService struct{}", "func (service NewService) Build(worker NewWorker) NewWorker", "return NewWorker{}"} {
		if !strings.Contains(service, expected) {
			t.Fatalf("expected %q in moved service, got:\n%s", expected, service)
		}
	}
	helper := mustReadFile(t, filepath.Join(root, "internal", "newpkg", "helper.go"))
	for _, expected := range []string{"package newpkg", "func BuildDefault() NewService", "return NewService{}"} {
		if !strings.Contains(helper, expected) {
			t.Fatalf("expected %q in moved helper, got:\n%s", expected, helper)
		}
	}
	testFile := mustReadFile(t, filepath.Join(root, "internal", "newpkg", "service_test.go"))
	for _, expected := range []string{`"example.com/project/internal/newpkg"`, "newpkg.NewService{}", "newpkg.NewWorker{}"} {
		if !strings.Contains(testFile, expected) {
			t.Fatalf("expected %q in moved Go test, got:\n%s", expected, testFile)
		}
	}
	api := mustReadFile(t, filepath.Join(root, "internal", "consumer", "api.go"))
	for _, expected := range []string{`"example.com/project/internal/newpkg"`, "func Build() newpkg.NewService", "return newpkg.NewService{}"} {
		if !strings.Contains(api, expected) {
			t.Fatalf("expected %q in external consumer, got:\n%s", expected, api)
		}
	}
	more := mustReadFile(t, filepath.Join(root, "internal", "consumer", "more.go"))
	for _, expected := range []string{"func Worker() newpkg.NewWorker", "return newpkg.BuildWorker()"} {
		if !strings.Contains(more, expected) {
			t.Fatalf("expected %q in second external consumer, got:\n%s", expected, more)
		}
	}
	noise := mustReadFile(t, filepath.Join(root, "internal", "unrelated", "noise.go"))
	if strings.Contains(noise, "newpkg.NewService") {
		t.Fatalf("expected unrelated string-like text to remain unchanged, got:\n%s", noise)
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
		t.Skip("git status output assertion is unix-oriented")
	}

	root := copyFixture(t)
	runGitForCliTest(t, root, "init")
	runGitForCliTest(t, root, "config", "user.email", "test@example.com")
	runGitForCliTest(t, root, "config", "user.name", "Test User")
	runGitForCliTest(t, root, "add", ".")
	runGitForCliTest(t, root, "commit", "-m", "initial")

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

	status := runGitForCliTestOutput(t, root, "status", "--short")
	if !strings.Contains(status, "app/Services/Billing/InvoiceService.php -> app/Domain/Billing/InvoiceService.php") {
		t.Fatalf("expected git mv rename to be reported, got status:\n%s", status)
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
	return copyNamedFixture(t, filepath.Join("tests", "fixtures", "php-basic"))
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

func writeRefactorlahConfig(t *testing.T, root string, checks [][]string, tests [][]string) {
	t.Helper()
	payload := struct {
		Checks [][]string `json:"checks,omitempty"`
		Tests  [][]string `json:"tests,omitempty"`
	}{
		Checks: checks,
		Tests:  tests,
	}
	content, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}
	mustWriteFile(t, filepath.Join(root, ".refactorlah.json"), string(content))
}

func hasValidation(results []reporting.ValidationResult, name string, status string) bool {
	for _, result := range results {
		if result.Name == name && result.Status == status {
			return true
		}
	}
	return false
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

func TestCommandValidationHelper(t *testing.T) {
	mode, ok := commandValidationHelperMode()
	if !ok {
		return
	}
	switch mode {
	case "pass":
		_, _ = os.Stdout.WriteString("check passed\n")
	case "fail":
		_, _ = os.Stderr.WriteString("check failed\n")
		os.Exit(1)
	default:
		os.Exit(2)
	}
	os.Exit(0)
}

func commandValidationHelperCommand(mode string) []string {
	return []string{os.Args[0], "-test.run=TestCommandValidationHelper", "--", mode}
}

func commandValidationHelperMode() (string, bool) {
	for index, argument := range os.Args {
		if argument == "--" && index+1 < len(os.Args) {
			return os.Args[index+1], true
		}
	}
	return "", false
}
