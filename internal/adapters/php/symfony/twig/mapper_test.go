package twig

import (
	"testing"

	"refactorlah/internal/planning"
)

func TestTemplateMapperDerivesNamespacedTwigMappings(t *testing.T) {
	mappings := TemplateMapper{}.DeriveMappings([]planning.FileMove{{
		OldPath: "templates/billing/archive.html.twig",
		NewPath: "src/Billing/Archive/Listing/Ui/Web/Twig/archive.html.twig",
	}}, PathConfiguration{Roots: []PathRoot{
		{Path: "templates"},
		{Path: "src/Billing/Archive/Listing/Ui/Web/Twig", Namespace: "Billing"},
	}})

	if len(mappings) != 2 {
		t.Fatalf("expected file and directory mapping, got %#v", mappings)
	}
	if mappings[0].OldReference != "billing/archive.html.twig" || mappings[0].NewReference != "@Billing/archive.html.twig" {
		t.Fatalf("unexpected template mapping %#v", mappings[0])
	}
	if mappings[1].OldReference != "billing" || mappings[1].NewReference != "@Billing" {
		t.Fatalf("unexpected directory mapping %#v", mappings[1])
	}
}

func TestTemplateMapperPrefersLongestTwigRoot(t *testing.T) {
	mappings := TemplateMapper{}.DeriveMappings([]planning.FileMove{{
		OldPath: "templates/admin/card.html.twig",
		NewPath: "templates/backoffice/card.html.twig",
	}}, PathConfiguration{Roots: []PathRoot{
		{Path: "templates"},
		{Path: "templates/admin", Namespace: "Admin"},
	}})

	if len(mappings) == 0 {
		t.Fatal("expected mapping")
	}
	if mappings[0].OldReference != "@Admin/card.html.twig" {
		t.Fatalf("expected longest old root, got %#v", mappings[0])
	}
}

func TestTemplateMapperSkipsNonTwigMoves(t *testing.T) {
	mappings := TemplateMapper{}.DeriveMappings([]planning.FileMove{{
		OldPath: "assets/app.css",
		NewPath: "assets/site.css",
	}}, PathConfiguration{Roots: []PathRoot{{Path: "templates"}}})

	if len(mappings) != 0 {
		t.Fatalf("expected no mappings, got %#v", mappings)
	}
}
