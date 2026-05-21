package adapters

import (
	"os"
	"path/filepath"
	"testing"

	"refactorlah/internal/planning"
)

func TestAutoDetectorDetectsComposerPHPProject(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "composer.json"), []byte(`{"autoload":{"psr-4":{"App\\":"app/"}}}`), 0o644); err != nil {
		t.Fatal(err)
	}

	detector := NewAutoDetector()
	signals, err := detector.Detect(t.Context(), root, planning.MovePlan{
		Moves: []planning.FileMove{{OldPath: "app/Foo.php", NewPath: "app/Bar.php"}},
	})
	if err != nil {
		t.Fatalf("detect failed: %v", err)
	}
	if !signals.PHPRelevant || !signals.IncludePHP {
		t.Fatalf("expected PHP adapter signal, got %#v", signals)
	}
}

func TestAutoDetectorDetectsTwigMoves(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "composer.json"), []byte(`{"autoload":{"psr-4":{"App\\":"app/"}}}`), 0o644); err != nil {
		t.Fatal(err)
	}

	detector := NewAutoDetector()
	signals, err := detector.Detect(t.Context(), root, planning.MovePlan{
		Moves: []planning.FileMove{{OldPath: "templates/admin/card.html.twig", NewPath: "templates/backoffice/card.html.twig"}},
	})
	if err != nil {
		t.Fatalf("detect failed: %v", err)
	}
	if !signals.PHPRelevant || !signals.IncludeTwig {
		t.Fatalf("expected Twig adapter signal, got %#v", signals)
	}
}

func TestAutoDetectorDetectsNestedComposerProjectTwigMove(t *testing.T) {
	root := t.TempDir()
	platformDir := filepath.Join(root, "platform")
	if err := os.MkdirAll(filepath.Join(platformDir, "config", "packages"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(platformDir, "composer.json"), []byte(`{"autoload":{"psr-4":{"App\\":"src/"}}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(platformDir, "config", "packages", "twig.yaml"), []byte("twig:\n  default_path: '%kernel.project_dir%/templates'\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	detector := NewAutoDetector()
	signals, err := detector.Detect(t.Context(), root, planning.MovePlan{
		Moves: []planning.FileMove{{
			OldPath: "platform/templates/billing/archive.html.twig",
			NewPath: "platform/src/Billing/Archive/Listing/Ui/Web/Twig/archive.html.twig",
		}},
	})
	if err != nil {
		t.Fatalf("detect failed: %v", err)
	}
	if !signals.PHPRelevant || !signals.IncludeTwig {
		t.Fatalf("expected nested Twig adapter signal, got %#v", signals)
	}
}

func TestAutoDetectorReturnsNoSignalsWithoutComposerRoot(t *testing.T) {
	root := t.TempDir()
	detector := NewAutoDetector()

	signals, err := detector.Detect(t.Context(), root, planning.MovePlan{
		Moves: []planning.FileMove{{OldPath: "app/Foo.php", NewPath: "app/Bar.php"}},
	})
	if err != nil {
		t.Fatalf("detect failed: %v", err)
	}
	if signals != (Signals{}) {
		t.Fatalf("expected no signals, got %#v", signals)
	}
}
