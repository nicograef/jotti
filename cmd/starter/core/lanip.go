package core

import (
	"fmt"
	"net"
	"strings"
)

// NetInterface beschreibt ein Netzwerk-Interface mit seinen IPv4-Adressen.
type NetInterface struct {
	Name string
	IPs  []string
}

// SelectLANIP waehlt die LAN-IPv4 fuer die LAN_IP-Env des Caddy-Containers.
// Bevorzugt wird die Outbound-Route-IP, sofern sie privat (RFC 1918) ist — auf
// Windows-Rechnern mit Docker Desktop tragen vEthernet-/WSL-Adapter eigene
// private 172.x-Adressen, die Smartphones nicht erreichen, weshalb "erste
// private IPv4" falsch waere. Schlaegt das fehl, greift eine Heuristik ueber die
// Interfaces mit Praeferenz 192.168.x > 10.x > 172.16-31.x. Loopback (127.x) und
// Link-Local (169.254.x) werden ignoriert.
func SelectLANIP(outboundIP string, interfaces []NetInterface) (string, error) {
	if privateRank(outboundIP) >= 0 {
		return strings.TrimSpace(outboundIP), nil
	}

	best := ""
	bestRank := -1
	for _, iface := range interfaces {
		for _, ip := range iface.IPs {
			rank := privateRank(ip)
			if rank < 0 {
				continue
			}
			// Kleinerer Rank = hoehere Praeferenz; bei Gleichstand gewinnt der
			// zuerst aufgefuehrte Adapter.
			if best == "" || rank < bestRank {
				best = strings.TrimSpace(ip)
				bestRank = rank
			}
		}
	}
	if best == "" {
		return "", fmt.Errorf("keine private LAN-IPv4-Adresse gefunden")
	}
	return best, nil
}

// privateRank bewertet eine IPv4 nach LAN-Eignung (kleiner = besser): 0 fuer
// 192.168.x, 1 fuer 10.x, 2 fuer 172.16-31.x. -1 fuer alles andere, inklusive
// Loopback, Link-Local, oeffentliche und IPv6-Adressen.
func privateRank(raw string) int {
	ip := net.ParseIP(strings.TrimSpace(raw))
	if ip == nil {
		return -1
	}
	ip4 := ip.To4()
	if ip4 == nil {
		return -1
	}
	switch {
	case ip4[0] == 192 && ip4[1] == 168:
		return 0
	case ip4[0] == 10:
		return 1
	case ip4[0] == 172 && ip4[1] >= 16 && ip4[1] <= 31:
		return 2
	default:
		return -1
	}
}
