package syntax

import "testing"

func TestParentAndLeafSplitModuleNames(t *testing.T) {
	if Parent("collector.memory.snapshot_manifest") != "collector.memory" {
		t.Fatalf("unexpected parent")
	}
	if Leaf("collector.memory.snapshot_manifest") != "snapshot_manifest" {
		t.Fatalf("unexpected leaf")
	}
	if Parent("snapshot_manifest") != "" {
		t.Fatalf("single-segment module should not have parent")
	}
	if Leaf("snapshot_manifest") != "snapshot_manifest" {
		t.Fatalf("single-segment module leaf should be itself")
	}
}

func TestResolveRelativeModule(t *testing.T) {
	resolved, ok := ResolveRelativeModule("collector.assembly.cache_files", 1, "snapshot_manifest")
	if !ok {
		t.Fatal("expected relative module to resolve")
	}
	if resolved != "collector.assembly.cache_files.snapshot_manifest" {
		t.Fatalf("unexpected resolved module %q", resolved)
	}

	resolved, ok = ResolveRelativeModule("collector.assembly.cache_files", 2, "snapshot_manifest")
	if !ok {
		t.Fatal("expected parent relative module to resolve")
	}
	if resolved != "collector.assembly.snapshot_manifest" {
		t.Fatalf("unexpected resolved module %q", resolved)
	}
}
