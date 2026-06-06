package config

import "testing"

func TestConfigAllowsUsesExcludes(t *testing.T) {
	policy := Config{Exclude: []string{"platform/local/phpstan/tests/fixtures/**"}}

	if policy.Allows("platform/local/phpstan/tests/fixtures/ArchitectureDependency.php") {
		t.Fatal("expected excluded fixture to be denied")
	}
	if !policy.Allows("platform/src/App.php") {
		t.Fatal("expected unrelated path to be allowed")
	}
}

func TestConfigAllowsIncludeOverridesExclude(t *testing.T) {
	policy := Config{
		Include: []string{"platform/local/phpstan/tests/fixtures/Allowed.php"},
		Exclude: []string{"platform/local/phpstan/tests/fixtures/**"},
	}

	if !policy.Allows("platform/local/phpstan/tests/fixtures/Allowed.php") {
		t.Fatal("expected explicit include to override exclude")
	}
}

func TestConfigAllowsSupportsSingleSegmentWildcards(t *testing.T) {
	policy := Config{Exclude: []string{"config/*.yaml"}}

	if policy.Allows("config/routes.yaml") {
		t.Fatal("expected yaml config to be excluded")
	}
	if !policy.Allows("config/packages/twig.yaml") {
		t.Fatal("single-star wildcard should not cross directories")
	}
}
