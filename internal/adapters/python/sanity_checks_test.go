//go:build cgo

package python

import (
	"reflect"
	"testing"

	adapterproto "refactorlah/internal/adapters/contract"
	"refactorlah/internal/planning"
)

func TestPythonSanityChecksCompileEditedPythonFiles(t *testing.T) {
	plan := planning.MovePlan{Moves: []planning.FileMove{{
		OldPath: "src/app/old.py",
		NewPath: "src/app/new.py",
	}}}
	replacements := []adapterproto.Replacement{
		{File: "src/app/new.py"},
		{File: "src/app/controller.py"},
		{File: "README.md"},
	}

	checks := pythonSanityChecks(plan, replacements, func(name string) (string, bool) {
		if name == "python3" {
			return "/bin/python3", true
		}
		return "", false
	})

	want := []adapterproto.Check{
		{Command: []string{"/bin/python3", "-m", "py_compile", "src/app/controller.py"}},
		{Command: []string{"/bin/python3", "-m", "py_compile", "src/app/new.py"}},
	}
	if !reflect.DeepEqual(checks, want) {
		t.Fatalf("unexpected checks:\nwant %#v\ngot  %#v", want, checks)
	}
}

func TestPythonSanityChecksSkipWhenNoPythonExists(t *testing.T) {
	checks := pythonSanityChecks(planning.MovePlan{Moves: []planning.FileMove{{
		NewPath: "src/app/new.py",
	}}}, nil, func(string) (string, bool) {
		return "", false
	})

	if len(checks) != 0 {
		t.Fatalf("expected no checks for unavailable python, got %#v", checks)
	}
}
