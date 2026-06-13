package main

import (
	"context"
	"crypto/x509"
	"errors"
	"net/netip"
	"testing"
)

func TestCheckRebind(t *testing.T) {
	lan := netip.MustParseAddr("192.168.1.50")

	tests := []struct {
		name  string
		addrs []netip.Addr
		err   error
		want  bool
	}{
		{
			name:  "löst auf eigene LAN-IP → ok",
			addrs: []netip.Addr{netip.MustParseAddr("192.168.1.50")},
			want:  true,
		},
		{
			name:  "löst auf fremde IP (Rebind-Schutz) → blockiert",
			addrs: []netip.Addr{netip.MustParseAddr("203.0.113.7")},
			want:  false,
		},
		{
			name:  "mehrere Antworten, eine passt → ok",
			addrs: []netip.Addr{netip.MustParseAddr("10.0.0.1"), netip.MustParseAddr("192.168.1.50")},
			want:  true,
		},
		{
			name: "Auflösung schlägt fehl → blockiert",
			err:  errors.New("nxdomain"),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := checkRebind("host.lokal.jotti.rocks", lan, func(context.Context, string) ([]netip.Addr, error) {
				return tt.addrs, tt.err
			})
			if got != tt.want {
				t.Errorf("checkRebind = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClassifyHandshakeError(t *testing.T) {
	if got := classifyHandshakeError(x509.CertificateInvalidError{Reason: x509.Expired}); got != certExpired {
		t.Errorf("abgelaufenes Zertifikat: got %d, want certExpired", got)
	}
	if got := classifyHandshakeError(x509.UnknownAuthorityError{}); got != certNone {
		t.Errorf("interne/unbekannte CA: got %d, want certNone", got)
	}
	if got := classifyHandshakeError(errors.New("connection refused")); got != certNone {
		t.Errorf("Verbindungsfehler: got %d, want certNone", got)
	}
}
