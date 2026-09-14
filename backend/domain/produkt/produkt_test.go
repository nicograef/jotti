//go:build unit

package produkt

import "testing"

func TestRichtungSchema_AkzeptiertNurHochUndRunter(t *testing.T) {
	for _, gueltig := range []Richtung{RichtungHoch, RichtungRunter} {
		if issues := RichtungSchema.Validate(&gueltig); issues != nil {
			t.Errorf("erwartete kein Issue für %q, bekam: %v", gueltig, issues)
		}
	}

	ungueltig := Richtung("links")
	if issues := RichtungSchema.Validate(&ungueltig); issues == nil {
		t.Error("erwartete Validierungsissue für \"links\", bekam nil")
	}
}
