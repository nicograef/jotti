package produkt

import (
	"errors"
	"testing"

	"github.com/nicograef/jotti/backend/domain/steuer"
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
		v, err := NewVariante("Bier", tc.preisCents, nil)
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
	v, err := NewVariante("Bier", 350, nil)
	if err != nil {
		t.Fatalf("Setup fehlgeschlagen: %v", err)
	}
	if err := v.UpdateDetails("Bier", 0, nil); err == nil {
		t.Error("erwartete Ablehnung eines 0-Cent-Preises bei UpdateDetails, bekam nil")
	}
	if v.PreisCents != 350 {
		t.Errorf("Preis sollte nach abgelehntem Update unverändert bleiben, bekam: %d", v.PreisCents)
	}
}

func TestPruefePfandVerknuepfung(t *testing.T) {
	pfand := func(steuersatz steuer.Steuersatz, status Status) Produkt {
		return Produkt{ID: 2, Name: "Pfand", Steuersatz: steuersatz, Pfand: true, Status: ActiveStatus,
			Varianten: []Variante{{ID: 20, Name: "Glas", PreisCents: 200, Status: status}}}
	}
	bier := func(steuersatz steuer.Steuersatz) Produkt {
		return Produkt{ID: 1, Name: "Bier", Steuersatz: steuersatz, Status: ActiveStatus}
	}
	ohneKennzeichen := pfand(steuer.RegelSteuersatz, ActiveStatus)
	ohneKennzeichen.Pfand = false

	cases := []struct {
		name   string
		quelle Produkt
		ziel   Produkt
		want   error
	}{
		{"gültig", bier(steuer.RegelSteuersatz), pfand(steuer.RegelSteuersatz, ActiveStatus), nil},
		{"gültig auf inaktive Pfand-Variante", bier(steuer.RegelSteuersatz), pfand(steuer.RegelSteuersatz, InactiveStatus), nil},
		{"Ziel ohne Pfand-Kennzeichen", bier(steuer.RegelSteuersatz), ohneKennzeichen, ErrPfandZielUngueltig},
		{"Ziel-Variante gelöscht", bier(steuer.RegelSteuersatz), pfand(steuer.RegelSteuersatz, DeletedStatus), ErrPfandZielUngueltig},
		{"Ziel-Variante fehlt", bier(steuer.RegelSteuersatz), Produkt{}, ErrPfandZielUngueltig},
		{"Quelle ist Pfand-Produkt", pfand(steuer.RegelSteuersatz, ActiveStatus), pfand(steuer.RegelSteuersatz, ActiveStatus), ErrPfandVerkettet},
		{"regel gegen ermaessigt", bier(steuer.RegelSteuersatz), pfand(steuer.ErmaessigtSteuersatz, ActiveStatus), ErrPfandSteuersatzAbweichend},
		{"kombi gegen kombi", bier(steuer.KombiSteuersatz), pfand(steuer.KombiSteuersatz, ActiveStatus), nil},
		{"kombi gegen regel", bier(steuer.KombiSteuersatz), pfand(steuer.RegelSteuersatz, ActiveStatus), ErrPfandSteuersatzAbweichend},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if err := PruefePfandVerknuepfung(tc.quelle, tc.ziel, 20); !errors.Is(err, tc.want) {
				t.Errorf("erwartete %v, bekam %v", tc.want, err)
			}
		})
	}
}
