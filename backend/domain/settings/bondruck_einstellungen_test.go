//go:build unit

package settings

import "testing"

func TestNewBondruckEinstellungen(t *testing.T) {
	cases := []struct {
		name      string
		ip        string
		wantError bool
	}{
		{"valid IPv4", "192.168.1.100", false},
		{"empty IP allowed", "", false},
		{"octet over 255", "256.1.1.1", true},
		{"out-of-range octets", "999.999.999.999", true},
		{"IPv6 rejected", "2001:db8::1", true},
		{"not an IP", "not-an-ip", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			b, err := NewBondruckEinstellungen(tc.ip)
			if tc.wantError {
				if err == nil {
					t.Fatalf("expected error for ip %q, got nil", tc.ip)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error for ip %q: %v", tc.ip, err)
			}
			if b.KassenbelegDruckerIP != tc.ip {
				t.Errorf("KassenbelegDruckerIP = %q, want %q", b.KassenbelegDruckerIP, tc.ip)
			}
			if b.UpdatedAt.IsZero() {
				t.Error("UpdatedAt should be set")
			}
		})
	}
}

func TestBondruckEinstellungenValidate_RequiresUpdatedAt(t *testing.T) {
	b := BondruckEinstellungen{KassenbelegDruckerIP: "192.168.1.100"}
	if err := b.Validate(); err == nil {
		t.Fatal("expected error for zero UpdatedAt, got nil")
	}
}
