package main

import (
	"fmt"
	"net/netip"
	"strings"
)

// deriveHostname bildet `<lan-ip-mit-bindestrichen>.<install-id>.<zone>` (z. B.
// `192-168-1-50.8e5700b1-….lokal.jotti.rocks`). Der jotti-Resolver löst den Namen
// rein rechnerisch auf die eingebettete IP auf; das Wildcard-Zertifikat
// `*.<install-id>.<zone>` deckt jede LAN-IP ab, ein DHCP-Wechsel ändert also nur
// den Namen, nie das Zertifikat.
func deriveHostname(lanIP netip.Addr, installID, zone string) string {
	return dashedIPv4(lanIP) + "." + installID + "." + zone
}

// dashedIPv4 schreibt eine IPv4-Adresse mit Bindestrichen statt Punkten, damit
// sie ein einzelnes DNS-Label bildet (192.168.1.50 → 192-168-1-50).
func dashedIPv4(ip netip.Addr) string {
	o := ip.As4()
	return fmt.Sprintf("%d-%d-%d-%d", o[0], o[1], o[2], o[3])
}

// resolveLANIP liest die LAN-IP des Hosts aus LAN_IP. Jeder Stack betreibt den Proxy in
// einem Bridge-Netz; dort zeigt die Default-Route nur auf die Docker-Bridge (172.x), die
// kein Smartphone erreicht. Ohne gültiges LAN_IP gibt es darum keine Zugangsadresse.
func resolveLANIP(envValue string) (netip.Addr, bool) {
	ip, err := netip.ParseAddr(strings.TrimSpace(envValue))
	if err != nil || !ip.Is4() {
		return netip.Addr{}, false
	}
	return ip, true
}
