package produkt

import (
	"errors"
	"testing"
)

func TestNewVariante_RejectsZeroPrice(t *testing.T) {
	_, err := NewVariante("Bier", 0)
	if err == nil {
		t.Fatal("erwartete Ablehnung eines 0-Cent-Preises, bekam nil")
	}
	if !errors.Is(err, ErrPreisUngueltig) {
		t.Errorf("erwartete ErrPreisUngueltig, bekam: %v", err)
	}
}

func TestNewVariante_RejectsNegativePrice(t *testing.T) {
	_, err := NewVariante("Bier", -1)
	if err == nil {
		t.Fatal("erwartete Ablehnung eines negativen Preises, bekam nil")
	}
	if !errors.Is(err, ErrPreisUngueltig) {
		t.Errorf("erwartete ErrPreisUngueltig, bekam: %v", err)
	}
}

func TestNewVariante_AcceptsOneCent(t *testing.T) {
	v, err := NewVariante("Bier", 1)
	if err != nil {
		t.Fatalf("erwartete Erfolg bei 1 Cent, bekam: %v", err)
	}
	if v.PreisCents != 1 {
		t.Errorf("erwartete PreisCents=1, bekam: %d", v.PreisCents)
	}
}

func TestNewVariante_AcceptsNormalPrice(t *testing.T) {
	v, err := NewVariante("Bier", 350)
	if err != nil {
		t.Fatalf("erwartete Erfolg bei normalem Preis, bekam: %v", err)
	}
	if v.PreisCents != 350 {
		t.Errorf("erwartete PreisCents=350, bekam: %d", v.PreisCents)
	}
	if v.Status != InactiveStatus {
		t.Errorf("erwartete Status=inactive, bekam: %s", v.Status)
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

// GTE(1) allein würde den Zero-Value 0 überspringen; .Required() fängt ihn ab.
func TestPreisCentsSchema_RejectsZeroValue(t *testing.T) {
	zero := 0
	issues := PreisCentsSchema.Validate(&zero)
	if issues == nil {
		t.Error("erwartete Validierungsissue für 0 Cent, bekam nil")
	}

	valid := 1
	if issues := PreisCentsSchema.Validate(&valid); issues != nil {
		t.Errorf("erwartete kein Issue für 1 Cent, bekam: %v", issues)
	}
}
