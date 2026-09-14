package core

import (
	"fmt"
	"net"
	"strings"
)

type NetInterface struct {
	Name string
	IPs  []string
}

// SelectLANIP waehlt die LAN-IPv4 fuer die LAN_IP-Env des Caddy-Containers:
// bevorzugt die Outbound-Route-IP, sofern privat (RFC 1918). Auf Rechnern mit
// Docker Desktop tragen vEthernet-/WSL-Adapter eigene private 172.x-Adressen, die
// Smartphones nicht erreichen — "erste private IPv4" waere also falsch. Sonst greift
// die Heuristik 192.168.x > 10.x > 172.16-31.x; Loopback und Link-Local zaehlen nie.
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
