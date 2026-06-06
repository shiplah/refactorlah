package php

import "testing"

func TestPsr4NamespaceResolverDerivesSymbol(t *testing.T) {
	resolved, ok := Psr4NamespaceResolver{}.DeriveSymbol(
		NewPsr4Map(map[string][]string{"App\\": {"app/"}}),
		"app/Services/Billing/InvoiceService.php",
	)
	if !ok {
		t.Fatal("expected resolved symbol")
	}
	if resolved.Symbol != "App\\Services\\Billing\\InvoiceService" {
		t.Fatalf("unexpected symbol %q", resolved.Symbol)
	}
	if resolved.Namespace != "App\\Services\\Billing" {
		t.Fatalf("unexpected namespace %q", resolved.Namespace)
	}
	if resolved.ShortName != "InvoiceService" {
		t.Fatalf("unexpected short name %q", resolved.ShortName)
	}
}

func TestPsr4NamespaceResolverPrefersLongestRoot(t *testing.T) {
	resolved, ok := Psr4NamespaceResolver{}.DeriveSymbol(
		NewPsr4Map(map[string][]string{
			"App\\":         {"app/"},
			"App\\Shared\\": {"app/Shared/"},
		}),
		"app/Shared/Thing.php",
	)
	if !ok {
		t.Fatal("expected resolved symbol")
	}
	if resolved.Symbol != "App\\Shared\\Thing" {
		t.Fatalf("unexpected symbol %q", resolved.Symbol)
	}
}
