//go:build unit

package produkt

import "testing"

// RichtungSchema lässt genau die beiden Verschieberichtungen zu; jeder andere
// Wert (etwa ein Frontend-Tippfehler oder eine fremde Himmelsrichtung) muss vor
// dem Repository scheitern.
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
