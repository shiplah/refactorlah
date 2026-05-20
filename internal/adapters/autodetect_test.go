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
