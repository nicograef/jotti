package main

import (
	"fmt"
	"net"
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

// resolveLANIP bestimmt die LAN-IP des Hosts. Primär gilt LAN_IP: in einem
// Bridge-Container ist das der einzige verlässliche Weg, weil die Default-Route nur
// auf die Docker-Bridge (172.x) zeigt, die die Smartphones nicht erreichen. Ohne
// LAN_IP die Quell-IP der Default-Route — nur sinnvoll direkt im Host-Netz.
func resolveLANIP(envValue string) (netip.Addr, bool) {
	if ip, err := netip.ParseAddr(strings.TrimSpace(envValue)); err == nil && ip.Is4() {
		return ip, true
	}
	return outboundIPv4()
}

// outboundIPv4 liefert die Quell-IP der Default-Route. Der UDP-„Connect" sendet
// kein Paket, er löst nur die Routenwahl aus; ohne Default-Route schlägt er fehl.
func outboundIPv4() (netip.Addr, bool) {
	// 192.0.2.1 ist TEST-NET-1 (RFC 5737) — nie erreichbar, aber für die
	// Routenwahl ausreichend.
	conn, err := net.Dial("udp", "192.0.2.1:53")
	if err != nil {
		return netip.Addr{}, false
	}
	defer func() { _ = conn.Close() }()

	addr, ok := conn.LocalAddr().(*net.UDPAddr)
	if !ok {
		return netip.Addr{}, false
	}
	ip, ok := netip.AddrFromSlice(addr.IP)
	if !ok {
		return netip.Addr{}, false
	}
	if ip = ip.Unmap(); !ip.Is4() {
		return netip.Addr{}, false
	}
	return ip, true
}
