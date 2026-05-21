package planning

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPlannerSingleFile(t *testing.T) {
	root := t.TempDir()
	mustWriteFile(t, filepath.Join(root, "app", "Foo.php"))

	planner := NewPlanner()
	plan, err := planner.Build(t.Context(), root, "app/Foo.php", "app/Bar.php", func(path string) (bool, error) {
		return path == "app/Foo.php", nil
	})
	if err != nil {
		t.Fatalf("build failed: %v", err)
	}

	if len(plan.Moves) != 1 {
		t.Fatalf("expected 1 move, got %d", len(plan.Moves))
	}
	if !plan.Moves[0].Tracked || plan.Moves[0].Mover != "git mv" {
		t.Fatalf("unexpected move strategy: %#v", plan.Moves[0])
	}
}

func TestPlannerDirectorySkipsIgnoredDirectories(t *testing.T) {
	root := t.TempDir()
	mustWriteFile(t, filepath.Join(root, "app", "Services", "Billing", "InvoiceService.php"))
	mustWriteFile(t, filepath.Join(root, "app", "Services", "Billing", "storage", "framework", "cache.php"))
	mustWriteFile(t, filepath.Join(root, "app", "Services", "Billing", "vendor", "package.php"))

	planner := NewPlanner()
	plan, err := planner.Build(t.Context(), root, "app/Services/Billing", "app/Domain/Billing", func(path string) (bool, error) {
		return false, nil
	})
	if err != nil {
		t.Fatalf("build failed: %v", err)
	}

	if len(plan.Moves) != 1 {
		t.Fatalf("expected 1 move, got %d", len(plan.Moves))
	}
	if plan.Moves[0].OldPath != "app/Services/Billing/InvoiceService.php" {
		t.Fatalf("unexpected move path: %s", plan.Moves[0].OldPath)
	}
}

func TestPlannerExistingTargetFails(t *testing.T) {
	root := t.TempDir()
	mustWriteFile(t, filepath.Join(root, "app", "Foo.php"))
	mustWriteFile(t, filepath.Join(root, "app", "Bar.php"))

	planner := NewPlanner()
	if _, err := planner.Build(t.Context(), root, "app/Foo.php", "app/Bar.php", func(path string) (bool, error) {
		return false, nil
	}); err == nil {
		t.Fatal("expected target exists error")
	}
}

func TestPlannerBuildManyRejectsDuplicateTargets(t *testing.T) {
	root := t.TempDir()
	mustWriteFile(t, filepath.Join(root, "app", "Foo.php"))
	mustWriteFile(t, filepath.Join(root, "app", "Bar.php"))

	planner := NewPlanner()
	_, err := planner.BuildMany(t.Context(), root, []RequestedMove{
		{OldPath: "app/Foo.php", NewPath: "app/Baz.php"},
		{OldPath: "app/Bar.php", NewPath: "app/Baz.php"},
	}, func(path string) (bool, error) {
		return false, nil
	})
	if err == nil {
		t.Fatal("expected duplicate target error")
	}
}

func TestPlannerBuildManyAllowsLaterMoveInsideEarlierTarget(t *testing.T) {
	root := t.TempDir()
	mustWriteFile(t, filepath.Join(root, "src", "Workers", "FooWorker.php"))
	mustWriteFile(t, filepath.Join(root, "src", "Workers", "BarWorker.php"))

	planner := NewPlanner()
	plan, err := planner.BuildMany(t.Context(), root, []RequestedMove{
		{OldPath: "src/Workers", NewPath: "src/Rules"},
		{OldPath: "src/Rules/FooWorker.php", NewPath: "src/Rules/FooRule.php"},
	}, func(path string) (bool, error) {
		return true, nil
	})
	if err != nil {
		t.Fatalf("build failed: %v", err)
	}

	if len(plan.Moves) != 2 {
		t.Fatalf("expected 2 moves, got %d", len(plan.Moves))
	}
	if plan.Moves[0].OldPath != "src/Workers/BarWorker.php" || plan.Moves[0].NewPath != "src/Rules/BarWorker.php" {
		t.Fatalf("unexpected first move: %#v", plan.Moves[0])
	}
	if plan.Moves[1].OldPath != "src/Workers/FooWorker.php" || plan.Moves[1].NewPath != "src/Rules/FooRule.php" {
		t.Fatalf("unexpected second move: %#v", plan.Moves[1])
	}
}

func mustWriteFile(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte("<?php\n"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
}
