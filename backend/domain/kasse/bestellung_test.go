//go:build unit

package kasse

import "testing"

func TestPositionBezeichnung(t *testing.T) {
	tests := []struct {
		name         string
		produktName  string
		varianteName string
		want         string
	}{
		{name: "Normalfall", produktName: "Pommes", varianteName: "mit Ketchup", want: "Pommes mit Ketchup"},
		{name: "gleichlautend ohne Dedup", produktName: "Cola", varianteName: "Cola", want: "Cola Cola"},
		{name: "leerer Variantenname ohne Trailing-Space", produktName: "Maß Bier", varianteName: "", want: "Maß Bier"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			pos := Position{ProduktName: tc.produktName, VarianteName: tc.varianteName}
			if got := pos.Bezeichnung(); got != tc.want {
				t.Errorf("Bezeichnung() = %q, want %q", got, tc.want)
			}
		})
	}
}
