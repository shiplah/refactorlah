//go:build cgo

package php

import (
	"reflect"
	"testing"

	adapterproto "refactorlah/internal/adapters/contract"
	"refactorlah/internal/planning"
)

func TestPHPSanityChecksUseEditedPHPFilesAndComposerRoot(t *testing.T) {
	plan := planning.MovePlan{Moves: []planning.FileMove{{
		OldPath: "platform/app/Old.php",
		NewPath: "platform/app/New.php",
	}}}
	replacements := []adapterproto.Replacement{
		{File: "platform/app/New.php"},
		{File: "platform/app/Controller.php"},
		{File: "platform/templates/card.html.twig"},
	}

	checks := phpSanityChecks("/repo", "/repo/platform", plan, replacements, func(name string) bool {
		return name == "php" || name == "composer"
	})

	want := []adapterproto.Check{
		{Command: []string{"php", "-l", "platform/app/Controller.php"}},
		{Command: []string{"php", "-l", "platform/app/New.php"}},
		{Directory: "platform", Command: []string{"composer", "dump-autoload"}},
	}
	if !reflect.DeepEqual(checks, want) {
		t.Fatalf("unexpected checks:\nwant %#v\ngot  %#v", want, checks)
	}
}

func TestPHPSanityChecksSkipUnavailableTools(t *testing.T) {
	checks := phpSanityChecks("/repo", "/repo", planning.MovePlan{Moves: []planning.FileMove{{
		NewPath: "app/New.php",
	}}}, nil, func(string) bool {
		return false
	})

	if len(checks) != 0 {
		t.Fatalf("expected no checks for unavailable tools, got %#v", checks)
	}
}
