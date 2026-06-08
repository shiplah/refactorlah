package golang

import (
	"reflect"
	"testing"

	adapterproto "github.com/NickSdot/refactorlah/internal/adapters/contract"
)

func TestGoSanityChecksBuildFromGoRoot(t *testing.T) {
	checks := goSanityChecks("/repo", "/repo/platform", func(name string) bool {
		return name == "go"
	})

	want := []adapterproto.Check{{
		Directory: "platform",
		Command:   []string{"go", "build", "./..."},
	}}
	if !reflect.DeepEqual(checks, want) {
		t.Fatalf("unexpected checks:\nwant %#v\ngot  %#v", want, checks)
	}
}

func TestGoSanityChecksSkipUnavailableGo(t *testing.T) {
	checks := goSanityChecks("/repo", "/repo", func(string) bool {
		return false
	})

	if len(checks) != 0 {
		t.Fatalf("expected no checks for unavailable go, got %#v", checks)
	}
}
