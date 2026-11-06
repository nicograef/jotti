package produkt

import (
	"errors"
	"testing"
)

// GTE(1) allein würde den Zero-Value 0 überspringen; .Required() in PreisCentsSchema fängt ihn ab.
func TestNewVariante_Preisgrenze(t *testing.T) {
	cases := []struct {
		preisCents int
		wantErr    error
	}{
		{0, ErrPreisUngueltig},
		{-1, ErrPreisUngueltig},
		{1, nil},
		{350, nil},
	}
	for _, tc := range cases {
		v, err := NewVariante("Bier", tc.preisCents)
		if !errors.Is(err, tc.wantErr) {
			t.Errorf("NewVariante(%d): erwartete Fehler %v, bekam %v", tc.preisCents, tc.wantErr, err)
			continue
		}
		if tc.wantErr != nil {
			continue
		}
		if v.PreisCents != tc.preisCents {
			t.Errorf("NewVariante(%d): erwartete PreisCents=%d, bekam %d", tc.preisCents, tc.preisCents, v.PreisCents)
		}
		if v.Status != InactiveStatus {
			t.Errorf("NewVariante(%d): erwartete Status=inactive, bekam %s", tc.preisCents, v.Status)
		}
	}
}

func TestUpdateDetails_RejectsZeroPrice(t *testing.T) {
	v, err := NewVariante("Bier", 350)
	if err != nil {
		t.Fatalf("Setup fehlgeschlagen: %v", err)
	}
	if err := v.UpdateDetails("Bier", 0); err == nil {
		t.Error("erwartete Ablehnung eines 0-Cent-Preises bei UpdateDetails, bekam nil")
	}
	if v.PreisCents != 350 {
		t.Errorf("Preis sollte nach abgelehntem Update unverändert bleiben, bekam: %d", v.PreisCents)
	}
}
