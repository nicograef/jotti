package main

import (
	"fmt"
	"net"

	"github.com/nicograef/jotti/windows/starter/core"
)

// detectLANIP ermittelt die LAN-IP fuer die LAN_IP-Env des Caddy-Containers; ohne
// sie laeuft der Start weiter, Caddy rendert dann nur die Fallback-Site.
func detectLANIP() string {
	ip, err := core.SelectLANIP(outboundIP(), localInterfaces())
	if err != nil {
		fmt.Printf("Hinweis: LAN-IP konnte nicht ermittelt werden (%v) - die Zugangsadresse fuers WLAN "+
			"erscheint erst, sobald eine LAN-IP erkannt wird.\n", err)
		return ""
	}
	fmt.Printf("LAN-IP: %s\n", ip)
	return ip
}

// outboundIP liefert die IP des Default-Route-Interfaces ueber einen UDP-"Connect"
// (es wird kein Paket gesendet).
func outboundIP() string {
	conn, err := net.Dial("udp", "1.1.1.1:80")
	if err != nil {
		return ""
	}
	defer func() { _ = conn.Close() }()
	if addr, ok := conn.LocalAddr().(*net.UDPAddr); ok {
		return addr.IP.String()
	}
	return ""
}

func localInterfaces() []core.NetInterface {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil
	}
	result := make([]core.NetInterface, 0, len(ifaces))
	for _, iface := range ifaces {
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		var ips []string
		for _, addr := range addrs {
			if ipNet, ok := addr.(*net.IPNet); ok {
				if ip4 := ipNet.IP.To4(); ip4 != nil {
					ips = append(ips, ip4.String())
				}
			}
		}
		if len(ips) > 0 {
			result = append(result, core.NetInterface{Name: iface.Name, IPs: ips})
		}
	}
	return result
}
