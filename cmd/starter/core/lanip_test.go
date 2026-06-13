package core

import "testing"

func TestSelectLANIPPrefersPrivateOutbound(t *testing.T) {
	ip, err := SelectLANIP("192.168.1.50", []NetInterface{
		{Name: "vEthernet (WSL)", IPs: []string{"172.20.0.1"}},
	})
	if err != nil {
		t.Fatalf("unerwarteter Fehler: %v", err)
	}
	if ip != "192.168.1.50" {
		t.Fatalf("got %q, want 192.168.1.50", ip)
	}
}

func TestSelectLANIPIgnoresPublicOutboundAndPicksLAN(t *testing.T) {
	// Oeffentliche Outbound-IP -> Fallback ueber die Interfaces.
	ip, err := SelectLANIP("203.0.113.7", []NetInterface{
		{Name: "vEthernet (WSL)", IPs: []string{"172.20.0.1"}},
		{Name: "WLAN", IPs: []string{"192.168.178.42"}},
	})
	if err != nil {
		t.Fatalf("unerwarteter Fehler: %v", err)
	}
	if ip != "192.168.178.42" {
		t.Fatalf("got %q, want 192.168.178.42 (192.168.x schlaegt 172.x)", ip)
	}
}

func TestSelectLANIP192BeatsDockerBridge(t *testing.T) {
	// Kein/Loopback-Outbound -> Interface-Heuristik; der vEthernet-172.x-Adapter
	// darf nie gegen ein echtes 192.168-WLAN gewinnen.
	ip, err := SelectLANIP("127.0.0.1", []NetInterface{
		{Name: "vEthernet (Default Switch)", IPs: []string{"172.18.32.1"}},
		{Name: "Ethernet", IPs: []string{"10.0.0.5"}},
		{Name: "WLAN", IPs: []string{"192.168.0.10"}},
	})
	if err != nil {
		t.Fatalf("unerwarteter Fehler: %v", err)
	}
	if ip != "192.168.0.10" {
		t.Fatalf("got %q, want 192.168.0.10", ip)
	}
}

func TestSelectLANIP10BeatsDockerBridgeWhenNo192(t *testing.T) {
	ip, err := SelectLANIP("", []NetInterface{
		{Name: "vEthernet", IPs: []string{"172.30.0.1"}},
		{Name: "Ethernet", IPs: []string{"10.1.2.3"}},
	})
	if err != nil {
		t.Fatalf("unerwarteter Fehler: %v", err)
	}
	if ip != "10.1.2.3" {
		t.Fatalf("got %q, want 10.1.2.3 (10.x schlaegt 172.16-31.x)", ip)
	}
}

func TestSelectLANIPIgnoresLoopbackAndLinkLocal(t *testing.T) {
	_, err := SelectLANIP("", []NetInterface{
		{Name: "Loopback", IPs: []string{"127.0.0.1"}},
		{Name: "WLAN (kein DHCP)", IPs: []string{"169.254.10.20"}},
	})
	if err == nil {
		t.Fatal("erwartete Fehlermeldung, weil keine private LAN-IP vorhanden ist")
	}
}
