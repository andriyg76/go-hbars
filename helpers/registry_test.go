package helpers

import (
	"testing"
)

func TestRegistry(t *testing.T) {
	reg := Registry()
	if reg == nil {
		t.Fatal("Registry() returned nil")
	}
	if len(reg) == 0 {
		t.Fatal("Registry() returned empty map")
	}

	// Spot-check known helpers
	for name, ref := range map[string]HelperRef{
		"upper":      {ImportPath: "github.com/andriyg76/go-hbars/helpers/handlebars", Ident: "Upper"},
		"formatDate": {ImportPath: "github.com/andriyg76/go-hbars/helpers/handlebars", Ident: "FormatDate"},
		"eq":         {ImportPath: "github.com/andriyg76/go-hbars/helpers/handlebars", Ident: "Eq"},
		"default":   {ImportPath: "github.com/andriyg76/go-hbars/helpers/handlebars", Ident: "Default"},
	} {
		r, ok := reg[name]
		if !ok {
			t.Errorf("Registry() missing helper %q", name)
			continue
		}
		if r.ImportPath != ref.ImportPath {
			t.Errorf("Registry()[%q].ImportPath = %q, want %q", name, r.ImportPath, ref.ImportPath)
		}
		if r.Ident != ref.Ident {
			t.Errorf("Registry()[%q].Ident = %q, want %q", name, r.Ident, ref.Ident)
		}
	}
}
