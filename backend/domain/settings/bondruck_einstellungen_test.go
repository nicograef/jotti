//go:build unit

package settings

import "testing"

func TestNewBondruckEinstellungen(t *testing.T) {
	cases := []struct {
		name               string
		kassenbelegIP      string
		direktverkaufModus DirektverkaufModus
		abholbonIP         string
		wantError          bool
	}{
		{"valid all fields", "192.168.1.100", DirektverkaufModusAnStationen, "192.168.1.101", false},
		{"empty IPs allowed", "", DirektverkaufModusKeinBon, "", false},
		{"invalid kassenbeleg IPv4", "256.1.1.1", DirektverkaufModusKeinBon, "", true},
		{"invalid abholbon IPv4", "", DirektverkaufModusAbholbon, "999.999.999.999", true},
		{"abholbon rejects IPv6", "", DirektverkaufModusAbholbon, "2001:db8::1", true},
		{"invalid direktverkauf modus", "", DirektverkaufModus("invalid"), "", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			b, err := NewBondruckEinstellungen(tc.kassenbelegIP, tc.direktverkaufModus, tc.abholbonIP)
			if tc.wantError {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if b.KassenbelegDruckerIP != tc.kassenbelegIP {
				t.Errorf("KassenbelegDruckerIP = %q, want %q", b.KassenbelegDruckerIP, tc.kassenbelegIP)
			}
			if b.DirektverkaufModus != tc.direktverkaufModus {
				t.Errorf("DirektverkaufModus = %q, want %q", b.DirektverkaufModus, tc.direktverkaufModus)
			}
			if b.AbholbonDruckerIP != tc.abholbonIP {
				t.Errorf("AbholbonDruckerIP = %q, want %q", b.AbholbonDruckerIP, tc.abholbonIP)
			}
			if b.UpdatedAt.IsZero() {
				t.Error("UpdatedAt should be set")
			}
		})
	}
}

func TestBondruckEinstellungenValidate_RequiresUpdatedAt(t *testing.T) {
	b := BondruckEinstellungen{
		KassenbelegDruckerIP: "192.168.1.100",
		DirektverkaufModus:   DirektverkaufModusKeinBon,
	}
	if err := b.Validate(); err == nil {
		t.Fatal("expected error for zero UpdatedAt, got nil")
	}
}
