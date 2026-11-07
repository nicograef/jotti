package druckstation

import (
	"slices"
	"testing"
)

func TestNewDruckstation(t *testing.T) {
	cases := []struct {
		name      string
		kategorie Kategorie
		druckerIP string
		bonmodus  Bonmodus
		wantError bool
	}{
		{"produktkategorie mit bonmodus", KategorieEssen, "192.168.1.51", BonmodusProPosition, false},
		{"kassenbeleg ohne bonmodus", KategorieKassenbeleg, "192.168.1.60", "", false},
		{"abholbon kein_bon ohne IP", KategorieAbholbon, "", BonmodusKeinBon, false},
		{"bonmodus, den die kategorie nicht erlaubt", KategorieEssen, "", BonmodusProStueck, true},
		{"ungültige kategorie", Kategorie("foo"), "", "", true},
		{"ungültige IPv4", KategorieEssen, "256.1.1.1", BonmodusProPosition, true},
		{"IPv6 abgelehnt", KategorieEssen, "2001:db8::1", BonmodusProPosition, true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			station, err := NewDruckstation(tc.kategorie, tc.druckerIP, tc.bonmodus)
			if tc.wantError {
				if err == nil {
					t.Errorf("expected error, got nil")
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

func TestKategorieErlaubtBonmodus(t *testing.T) {
	cases := []struct {
		kategorie Kategorie
		bonmodus  Bonmodus
		want      bool
	}{
		{KategorieEssen, BonmodusProPosition, true},
		{KategorieEssen, BonmodusProBestellung, true},
		{KategorieEssen, BonmodusProStueck, false},
		{KategorieEssen, BonmodusKeinBon, false},
		{KategorieEssen, "", false},
		{KategorieGetraenk, BonmodusProBestellung, true},
		{KategorieGetraenk, BonmodusProStueck, false},
		{KategorieGetraenk, BonmodusKeinBon, false},
		{KategorieSonstiges, BonmodusProStueck, false},
		{KategorieSonstiges, "", false},
		{KategorieAbholbon, BonmodusProPosition, true},
		{KategorieAbholbon, BonmodusProBestellung, true},
		{KategorieAbholbon, BonmodusProStueck, true},
		{KategorieAbholbon, BonmodusKeinBon, true},
		{KategorieAbholbon, "", false},
		{KategorieAbholbon, Bonmodus("invalid"), false},
		{KategorieKassenbeleg, "", true},
		{KategorieKassenbeleg, BonmodusProPosition, false},
	}

	for _, tc := range cases {
		t.Run(string(tc.kategorie)+"/"+string(tc.bonmodus), func(t *testing.T) {
			if got := tc.kategorie.ErlaubtBonmodus(tc.bonmodus); got != tc.want {
				t.Errorf("ErlaubtBonmodus(%q) = %v, want %v", tc.bonmodus, got, tc.want)
			}
		})
	}
}

func TestKategorieBonmodi(t *testing.T) {
	cases := []struct {
		kategorie Kategorie
		want      []Bonmodus
	}{
		{KategorieEssen, []Bonmodus{BonmodusProPosition, BonmodusProBestellung}},
		{KategorieAbholbon, []Bonmodus{BonmodusProPosition, BonmodusProBestellung, BonmodusProStueck, BonmodusKeinBon}},
		{KategorieKassenbeleg, []Bonmodus{}},
	}

	for _, tc := range cases {
		if got := tc.kategorie.Bonmodi(); !slices.Equal(got, tc.want) {
			t.Errorf("%s.Bonmodi() = %v, want %v", tc.kategorie, got, tc.want)
		}
	}
}
