package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"net"
	"net/netip"
	"time"
)

const (
	certProbeTimeout    = 4 * time.Second
	rebindLookupTimeout = 3 * time.Second
)

// probeCert öffnet einen TLS-Handshake zum eigenen Caddy und setzt serverName als
// SNI, sodass Caddy das Wildcard-Zertifikat ausliefert. Verifiziert wird gegen die
// System-Roots: nur eine öffentlich vertrauenswürdige Kette gilt als „grün"; ein
// abgelaufenes Zertifikat wird als solches erkannt, alles andere ⇒ certNone.
func probeCert(addr, serverName string) certState {
	dialer := &net.Dialer{Timeout: certProbeTimeout}
	conn, err := tls.DialWithDialer(dialer, "tcp", addr, &tls.Config{ServerName: serverName})
	if err != nil {
		return classifyHandshakeError(err)
	}
	_ = conn.Close()
	return certValid
}

func classifyHandshakeError(err error) certState {
	var invalid x509.CertificateInvalidError
	if errors.As(err, &invalid) && invalid.Reason == x509.Expired {
		return certExpired
	}
	return certNone
}

// checkRebind meldet, ob der eigene Hostname über den System-Resolver (= Router)
// auf die eigene LAN-IP auflöst. Tut er das nicht, greift vermutlich der
// DNS-Rebind-Schutz des Routers — dann ist die grüne Adresse im WLAN unerreichbar.
func checkRebind(hostname string, lanIP netip.Addr, lookup func(context.Context, string) ([]netip.Addr, error)) bool {
	ctx, cancel := context.WithTimeout(context.Background(), rebindLookupTimeout)
	defer cancel()

	addrs, err := lookup(ctx, hostname)
	if err != nil {
		return false
	}
	for _, a := range addrs {
		if a.Unmap() == lanIP {
			return true
		}
	}
	return false
}

// systemLookupIP löst host über den System-Resolver des Containers auf (→ Host →
// Router) — repräsentativ für die Auflösung der Handys im selben WLAN.
func systemLookupIP(ctx context.Context, host string) ([]netip.Addr, error) {
	return net.DefaultResolver.LookupNetIP(ctx, "ip4", host)
}
