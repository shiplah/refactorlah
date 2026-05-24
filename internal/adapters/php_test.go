package adapters

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"refactorlah/internal/planning"
)

func TestInvokerAcceptsValidAdapterResponse(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell script helper is unix-only")
	}

	root := t.TempDir()
	script := filepath.Join(root, "fake-adapter")
	if err := os.WriteFile(script, []byte("#!/bin/sh\ncat >/dev/null\nprintf '%s' '{\"protocolVersion\":1,\"adapter\":\"php\",\"symbolMappings\":[],\"pathMappings\":[],\"replacements\":[],\"warnings\":[],\"errors\":[]}'\n"), 0o755); err != nil {
		t.Fatal(err)
	}

	invoker := NewInvoker()
	selection := Selection{Adapters: []Config{{Name: "php", Path: script}}}
	_, err := invoker.Invoke(t.Context(), root, planning.MovePlan{}, true, selection)
	if err != nil {
		t.Fatalf("invoke failed: %v", err)
	}
}

func TestInvokerPassesScanOptionsToAdapter(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell script helper is unix-only")
	}

	root := t.TempDir()
	requestPath := filepath.Join(root, "request.json")
	script := filepath.Join(root, "fake-adapter")
	scriptContent := "#!/bin/sh\ncat > request.json\nprintf '%s' '{\"protocolVersion\":1,\"adapter\":\"php\",\"symbolMappings\":[],\"pathMappings\":[],\"replacements\":[],\"warnings\":[],\"errors\":[]}'\n"
	if err := os.WriteFile(script, []byte(scriptContent), 0o755); err != nil {
		t.Fatal(err)
	}

	invoker := NewInvoker()
	selection := Selection{Adapters: []Config{{
		Name: "php",
		Path: script,
		Options: RequestOptions{
			IncludePHP:  true,
			IncludeTwig: true,
			ScanInclude: []string{"platform/local/fixtures/Allowed.php"},
			ScanExclude: []string{"platform/local/fixtures/**"},
		},
	}}}
	_, err := invoker.Invoke(t.Context(), root, planning.MovePlan{}, true, selection)
	if err != nil {
		t.Fatalf("invoke failed: %v", err)
	}

	content, err := os.ReadFile(requestPath)
	if err != nil {
		t.Fatal(err)
	}

	var request Request
	if err := json.Unmarshal(content, &request); err != nil {
		t.Fatal(err)
	}
	if got := request.Options.ScanInclude; len(got) != 1 || got[0] != "platform/local/fixtures/Allowed.php" {
		t.Fatalf("unexpected scan include options: %#v", got)
	}
	if got := request.Options.ScanExclude; len(got) != 1 || got[0] != "platform/local/fixtures/**" {
		t.Fatalf("unexpected scan exclude options: %#v", got)
	}
}

func TestInvokerRejectsInvalidJSON(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell script helper is unix-only")
	}

	root := t.TempDir()
	script := filepath.Join(root, "fake-adapter")
	if err := os.WriteFile(script, []byte("#!/bin/sh\ncat >/dev/null\nprintf '%s' 'not-json'\n"), 0o755); err != nil {
		t.Fatal(err)
	}

	invoker := NewInvoker()
	selection := Selection{Adapters: []Config{{Name: "php", Path: script}}}
	if _, err := invoker.Invoke(t.Context(), root, planning.MovePlan{}, true, selection); err == nil {
		t.Fatal("expected invalid JSON error")
	}
}
