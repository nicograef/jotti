package main

import (
	"net"
	"strings"
	"testing"
	"time"

	"github.com/miekg/dns"
)

func TestLoadConfigFromEnv(t *testing.T) {
	tests := []struct {
		name        string
		env         map[string]string
		wantErr     bool
		errContains string
		check       func(t *testing.T, config resolverConfig)
	}{
		{
			name: "uses defaults when optional values are missing",
			env: map[string]string{
				"RESOLVER_NS_IP":        "203.0.113.10",
				"RESOLVER_AUTH_FORWARD": "acme-dns:53",
			},
			check: func(t *testing.T, config resolverConfig) {
				if config.zones.zone != "lokal.jotti.rocks." {
					t.Fatalf("zone: got %q", config.zones.zone)
				}
				if config.zones.authZone != "auth.jotti.rocks." {
					t.Fatalf("auth zone: got %q", config.zones.authZone)
				}
				if config.zones.nsName != "dns.jotti.rocks." {
					t.Fatalf("ns name: got %q", config.zones.nsName)
				}
				if config.listenAddr != ":53" {
					t.Fatalf("listen addr: got %q", config.listenAddr)
				}
				if config.forwardAddr != "acme-dns:53" {
					t.Fatalf("forward addr: got %q", config.forwardAddr)
				}
			},
		},
		{
			name: "canonicalizes zone names and appends default forward port",
			env: map[string]string{
				"RESOLVER_ZONE":         "Beispiel.Test",
				"RESOLVER_AUTH_ZONE":    "Auth.Beispiel.Test",
				"RESOLVER_NS_NAME":      "NS.Beispiel.Test",
				"RESOLVER_NS_IP":        "203.0.113.10",
				"RESOLVER_AUTH_FORWARD": "acme-dns",
			},
			check: func(t *testing.T, config resolverConfig) {
				if config.zones.zone != "beispiel.test." {
					t.Fatalf("zone: got %q", config.zones.zone)
				}
				if config.zones.authZone != "auth.beispiel.test." {
					t.Fatalf("auth zone: got %q", config.zones.authZone)
				}
				if config.zones.nsName != "ns.beispiel.test." {
					t.Fatalf("ns name: got %q", config.zones.nsName)
				}
				if config.forwardAddr != "acme-dns:53" {
					t.Fatalf("forward addr: got %q", config.forwardAddr)
				}
			},
		},
		{
			name: "fails when ns ip is missing",
			env: map[string]string{
				"RESOLVER_AUTH_FORWARD": "acme-dns:53",
			},
			wantErr:     true,
			errContains: "RESOLVER_NS_IP",
		},
		{
			name: "fails when ns ip is not IPv4",
			env: map[string]string{
				"RESOLVER_NS_IP":        "2001:db8::1",
				"RESOLVER_AUTH_FORWARD": "acme-dns:53",
			},
			wantErr:     true,
			errContains: "RESOLVER_NS_IP",
		},
		{
			name: "fails when forward address is missing",
			env: map[string]string{
				"RESOLVER_NS_IP": "203.0.113.10",
			},
			wantErr:     true,
			errContains: "RESOLVER_AUTH_FORWARD",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := loadConfigFromEnv(func(key string) string {
				return tt.env[key]
			})

			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Fatalf("expected error to contain %q, got %q", tt.errContains, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
			tt.check(t, config)
		})
	}
}

// startTestServer startet den Resolver mit UDP- und TCP-Listener auf
// zufälligen Ports und liefert beide Adressen.
func startTestServer(t *testing.T, h dns.Handler) (udpAddr, tcpAddr string) {
	t.Helper()

	packetConn, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("udp listen: %v", err)
	}
	udpServer := &dns.Server{PacketConn: packetConn, Handler: h}
	go func() { _ = udpServer.ActivateAndServe() }()
	t.Cleanup(func() { _ = udpServer.Shutdown() })

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("tcp listen: %v", err)
	}
	tcpServer := &dns.Server{Listener: listener, Handler: h}
	go func() { _ = tcpServer.ActivateAndServe() }()
	t.Cleanup(func() { _ = tcpServer.Shutdown() })

	return packetConn.LocalAddr().String(), listener.Addr().String()
}

func exchange(t *testing.T, network, addr string, req *dns.Msg) *dns.Msg {
	t.Helper()
	client := &dns.Client{Net: network, Timeout: 5 * time.Second}
	resp, _, err := client.Exchange(req, addr)
	if err != nil {
		t.Fatalf("exchange (%s): %v", network, err)
	}
	return resp
}

func TestIntegrationAAnfrage(t *testing.T) {
	udpAddr, tcpAddr := startTestServer(t, &handler{cfg: testZoneConfig()})
	qname := "192-168-1-50." + testInstallID + ".lokal.jotti.rocks."

	for network, addr := range map[string]string{"udp": udpAddr, "tcp": tcpAddr} {
		resp := exchange(t, network, addr, newQuery(qname, dns.TypeA))

		if resp.Rcode != dns.RcodeSuccess {
			t.Fatalf("%s rcode: got %d, want %d", network, resp.Rcode, dns.RcodeSuccess)
		}
		if !resp.Authoritative {
			t.Fatalf("%s: Antwort ist nicht autoritativ", network)
		}
		if len(resp.Answer) != 1 {
			t.Fatalf("%s answer count: got %d, want 1", network, len(resp.Answer))
		}
		a, ok := resp.Answer[0].(*dns.A)
		if !ok {
			t.Fatalf("%s answer type: got %T, want *dns.A", network, resp.Answer[0])
		}
		if got := a.A.String(); got != "192.168.1.50" {
			t.Fatalf("%s ip: got %s, want 192.168.1.50", network, got)
		}
		if a.Hdr.Ttl != 86400 {
			t.Fatalf("%s ttl: got %d, want 86400", network, a.Hdr.Ttl)
		}
	}
}

func TestIntegrationChallengeCNAME(t *testing.T) {
	udpAddr, _ := startTestServer(t, &handler{cfg: testZoneConfig()})

	resp := exchange(t, "udp", udpAddr, newQuery("_acme-challenge."+testInstallID+".lokal.jotti.rocks.", dns.TypeTXT))

	if len(resp.Answer) != 1 {
		t.Fatalf("answer count: got %d, want 1", len(resp.Answer))
	}
	cname, ok := resp.Answer[0].(*dns.CNAME)
	if !ok {
		t.Fatalf("answer type: got %T, want *dns.CNAME", resp.Answer[0])
	}
	if want := testInstallID + ".auth.jotti.rocks."; cname.Target != want {
		t.Fatalf("target: got %s, want %s", cname.Target, want)
	}
}

func TestIntegrationNXDomainUndApex(t *testing.T) {
	udpAddr, _ := startTestServer(t, &handler{cfg: testZoneConfig()})

	nx := exchange(t, "udp", udpAddr, newQuery("unsinn.lokal.jotti.rocks.", dns.TypeA))
	if nx.Rcode != dns.RcodeNameError {
		t.Fatalf("rcode: got %d, want %d (NXDOMAIN)", nx.Rcode, dns.RcodeNameError)
	}
	if len(nx.Ns) != 1 {
		t.Fatalf("authority count: got %d, want 1 (SOA)", len(nx.Ns))
	}

	soa := exchange(t, "udp", udpAddr, newQuery("lokal.jotti.rocks.", dns.TypeSOA))
	if len(soa.Answer) != 1 {
		t.Fatalf("SOA answer count: got %d, want 1", len(soa.Answer))
	}
	ns := exchange(t, "udp", udpAddr, newQuery("lokal.jotti.rocks.", dns.TypeNS))
	if len(ns.Answer) != 1 {
		t.Fatalf("NS answer count: got %d, want 1", len(ns.Answer))
	}
}

func TestIntegrationForwardingZurAuthZone(t *testing.T) {
	// Fake-acme-dns: beantwortet TXT-Anfragen mit einem festen Token.
	fakeAuth := dns.HandlerFunc(func(w dns.ResponseWriter, req *dns.Msg) {
		msg := new(dns.Msg)
		msg.SetReply(req)
		msg.Authoritative = true
		msg.Answer = []dns.RR{&dns.TXT{
			Hdr: dns.RR_Header{Name: req.Question[0].Name, Rrtype: dns.TypeTXT, Class: dns.ClassINET, Ttl: 1},
			Txt: []string{"challenge-token"},
		}}
		_ = w.WriteMsg(msg)
	})
	authAddr, _ := startTestServer(t, fakeAuth)

	udpAddr, _ := startTestServer(t, &handler{cfg: testZoneConfig(), forwardAddr: authAddr})

	resp := exchange(t, "udp", udpAddr, newQuery(testInstallID+".auth.jotti.rocks.", dns.TypeTXT))

	if resp.Rcode != dns.RcodeSuccess {
		t.Fatalf("rcode: got %d, want %d", resp.Rcode, dns.RcodeSuccess)
	}
	if len(resp.Answer) != 1 {
		t.Fatalf("answer count: got %d, want 1", len(resp.Answer))
	}
	txt, ok := resp.Answer[0].(*dns.TXT)
	if !ok {
		t.Fatalf("answer type: got %T, want *dns.TXT", resp.Answer[0])
	}
	if len(txt.Txt) != 1 || txt.Txt[0] != "challenge-token" {
		t.Fatalf("txt: got %v, want [challenge-token]", txt.Txt)
	}
}

func TestIntegrationForwardingFehlerLiefertServfail(t *testing.T) {
	// Port 1 auf localhost ist unbelegt — die Weiterleitung scheitert sofort.
	udpAddr, _ := startTestServer(t, &handler{cfg: testZoneConfig(), forwardAddr: "127.0.0.1:1"})

	resp := exchange(t, "udp", udpAddr, newQuery(testInstallID+".auth.jotti.rocks.", dns.TypeTXT))

	if resp.Rcode != dns.RcodeServerFailure {
		t.Fatalf("rcode: got %d, want %d (SERVFAIL)", resp.Rcode, dns.RcodeServerFailure)
	}
}
