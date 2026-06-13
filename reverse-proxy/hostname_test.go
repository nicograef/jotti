package main

import (
	"net/netip"
	"testing"
)

func TestDeriveHostname(t *testing.T) {
	const installID = "8e5700b1-3fa2-4b91-bc4a-1234567890ab"

	tests := []struct {
		name      string
		lanIP     string
		installID string
		zone      string
		want      string
	}{
		{
			name:      "private LAN-IP",
			lanIP:     "192.168.1.50",
			installID: installID,
			zone:      "lokal.jotti.rocks",
			want:      "192-168-1-50." + installID + ".lokal.jotti.rocks",
		},
		{
			name:      "einstellige Oktette",
			lanIP:     "10.0.0.1",
			installID: "abc",
			zone:      "lokal.jotti.rocks",
			want:      "10-0-0-1.abc.lokal.jotti.rocks",
		},
		{
			name:      "dreistellige Oktette und Grenzen",
			lanIP:     "192.168.123.255",
			installID: "id",
			zone:      "lokal.jotti.rocks",
			want:      "192-168-123-255.id.lokal.jotti.rocks",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := deriveHostname(netip.MustParseAddr(tt.lanIP), tt.installID, tt.zone)
			if got != tt.want {
				t.Fatalf("deriveHostname = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestResolveLANIPPrefersEnv(t *testing.T) {
	ip, ok := resolveLANIP("192.168.2.200")
	if !ok || ip.String() != "192.168.2.200" {
		t.Fatalf("resolveLANIP(env) = %v, %v; want 192.168.2.200, true", ip, ok)
	}
}
