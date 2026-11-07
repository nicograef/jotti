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

func TestResolveLANIP(t *testing.T) {
	tests := []struct {
		name   string
		env    string
		want   string
		wantOK bool
	}{
		{name: "gültige IPv4", env: "192.168.2.200", want: "192.168.2.200", wantOK: true},
		{name: "Leerzeichen werden entfernt", env: " 10.0.0.5 ", want: "10.0.0.5", wantOK: true},
		{name: "leer", env: ""},
		{name: "keine IP", env: "not-an-ip"},
		{name: "IPv6", env: "fd00::1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ip, ok := resolveLANIP(tt.env)
			if ok != tt.wantOK {
				t.Fatalf("resolveLANIP(%q) ok = %v, want %v", tt.env, ok, tt.wantOK)
			}
			if ok && ip.String() != tt.want {
				t.Fatalf("resolveLANIP(%q) = %s, want %s", tt.env, ip, tt.want)
			}
		})
	}
}
