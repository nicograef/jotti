//go:build unit

package druckstation

import "testing"

func TestNewDruckstation(t *testing.T) {
	cases := []struct {
		name      string
		kategorie Kategorie
		druckerIP string
		bonmodus  Bonmodus
		wantError bool
	}{
		{"produktkategorie pro_position", KategorieEssen, "192.168.1.51", BonmodusProPosition, false},
		{"produktkategorie pro_bestellung", KategorieGetraenk, "", BonmodusProBestellung, false},
		{"produktkategorie ohne bonmodus", KategorieSonstiges, "", "", true},
		{"produktkategorie ungültiger bonmodus", KategorieEssen, "", Bonmodus("invalid"), true},
		{"kassenbeleg ohne bonmodus", KategorieKassenbeleg, "192.168.1.60", "", false},
		{"abholbon pro_bestellung", KategorieAbholbon, "192.168.1.70", BonmodusProBestellung, false},
		{"abholbon pro_position", KategorieAbholbon, "", BonmodusProPosition, false},
		{"abholbon ohne bonmodus abgelehnt", KategorieAbholbon, "", "", true},
		{"kassenbeleg mit bonmodus abgelehnt", KategorieKassenbeleg, "", BonmodusProPosition, true},
		{"ungültige kategorie", Kategorie("foo"), "", "", true},
		{"ungültige IPv4", KategorieEssen, "256.1.1.1", BonmodusProPosition, true},
		{"IPv6 abgelehnt", KategorieEssen, "2001:db8::1", BonmodusProPosition, true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			station, err := NewDruckstation(tc.kategorie, tc.druckerIP, tc.bonmodus)
			if tc.wantError {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if station.Kategorie != tc.kategorie {
				t.Errorf("Kategorie = %q, want %q", station.Kategorie, tc.kategorie)
			}
			if station.DruckerIP != tc.druckerIP {
				t.Errorf("DruckerIP = %q, want %q", station.DruckerIP, tc.druckerIP)
			}
			if station.Bonmodus != tc.bonmodus {
				t.Errorf("Bonmodus = %q, want %q", station.Bonmodus, tc.bonmodus)
			}
		})
	}
}

func TestKategorieHatBonmodus(t *testing.T) {
	mitBonmodus := []Kategorie{KategorieEssen, KategorieGetraenk, KategorieSonstiges, KategorieAbholbon}
	for _, k := range mitBonmodus {
		if !k.HatBonmodus() {
			t.Errorf("%q should carry a Bonmodus", k)
		}
	}

	if KategorieKassenbeleg.HatBonmodus() {
		t.Errorf("%q should not carry a Bonmodus", KategorieKassenbeleg)
	}
}
