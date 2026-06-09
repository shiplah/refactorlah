//go:build cgo

package python

import (
	"reflect"
	"testing"

	adapterproto "github.com/shiplah/refactorlah/internal/adapters/contract"
	"github.com/shiplah/refactorlah/internal/planning"
)

func TestPythonSanityChecksCompileEditedPythonFiles(t *testing.T) {
	plan := planning.MovePlan{Moves: []planning.FileMove{{
		OldPath: "src/example/current/registry.py",
		NewPath: "src/example/archive/registry.py",
	}}}
	replacements := []adapterproto.Replacement{
		{File: "src/example/current/registry.py"},
		{File: "src/example/consumer.py"},
		{File: "README.md"},
	}

	checks := pythonSanityChecks(plan, replacements, func(name string) (string, bool) {
		if name == "python3" {
			return "/bin/python3", true
		}
		return "", false
	})

	want := []adapterproto.Check{
		{Command: []string{"/bin/python3", "-m", "py_compile", "src/example/archive/registry.py"}},
		{Command: []string{"/bin/python3", "-m", "py_compile", "src/example/consumer.py"}},
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
