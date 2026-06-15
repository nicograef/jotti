package main

import (
	"net/netip"
	"testing"

	"github.com/miekg/dns"
)

const testInstallID = "8e5700b1-3fa2-4b91-bc4a-1234567890ab"

func testZoneConfig() zoneConfig {
	return zoneConfig{
		zone:     "lokal.jotti.rocks.",
		authZone: "auth.jotti.rocks.",
		nsName:   "dns.jotti.rocks.",
		nsIP:     netip.MustParseAddr("203.0.113.10"),
	}
}

func TestResolve(t *testing.T) {
	tests := []struct {
		name  string
		qname string
		qtype uint16
		want  answer
	}{
		{
			name:  "private LAN-IP wird aus dem Namen berechnet",
			qname: "192-168-1-50." + testInstallID + ".lokal.jotti.rocks.",
			qtype: dns.TypeA,
			want:  answer{kind: kindA, ip: netip.MustParseAddr("192.168.1.50"), ttl: ttlA},
		},
		{
			name:  "oeffentliche IP funktioniert ebenso",
			qname: "203-0-113-7." + testInstallID + ".lokal.jotti.rocks.",
			qtype: dns.TypeA,
			want:  answer{kind: kindA, ip: netip.MustParseAddr("203.0.113.7"), ttl: ttlA},
		},
		{
			name:  "Oktett-Grenzen 0 und 255 sind gueltig",
			qname: "0-255-0-255." + testInstallID + ".lokal.jotti.rocks.",
			qtype: dns.TypeA,
			want:  answer{kind: kindA, ip: netip.MustParseAddr("0.255.0.255"), ttl: ttlA},
		},
		{
			name:  "Grossschreibung wird normalisiert",
			qname: "192-168-1-50." + testInstallID + ".LOKAL.JOTTI.ROCKS.",
			qtype: dns.TypeA,
			want:  answer{kind: kindA, ip: netip.MustParseAddr("192.168.1.50"), ttl: ttlA},
		},
		{
			name:  "Oktett ueber 255 ist NXDOMAIN",
			qname: "192-168-1-256." + testInstallID + ".lokal.jotti.rocks.",
			qtype: dns.TypeA,
			want:  answer{kind: kindNXDomain},
		},
		{
			name:  "nicht-numerisches Oktett ist NXDOMAIN",
			qname: "192-168-1-ab." + testInstallID + ".lokal.jotti.rocks.",
			qtype: dns.TypeA,
			want:  answer{kind: kindNXDomain},
		},
		{
			name:  "leeres Oktett ist NXDOMAIN",
			qname: "192--1-50." + testInstallID + ".lokal.jotti.rocks.",
			qtype: dns.TypeA,
			want:  answer{kind: kindNXDomain},
		},
		{
			name:  "fuenf Oktette sind NXDOMAIN",
			qname: "10-0-0-1-2." + testInstallID + ".lokal.jotti.rocks.",
			qtype: dns.TypeA,
			want:  answer{kind: kindNXDomain},
		},
		{
			name:  "AAAA auf gueltigen Namen ist NoData, nicht NXDOMAIN",
			qname: "192-168-1-50." + testInstallID + ".lokal.jotti.rocks.",
			qtype: dns.TypeAAAA,
			want:  answer{kind: kindNoData},
		},
		{
			name:  "acme-challenge liefert berechneten CNAME",
			qname: "_acme-challenge." + testInstallID + ".lokal.jotti.rocks.",
			qtype: dns.TypeTXT,
			want:  answer{kind: kindCNAME, cname: testInstallID + ".auth.jotti.rocks.", ttl: ttlDefault},
		},
		{
			name:  "acme-challenge-CNAME gilt fuer jeden Anfrage-Typ",
			qname: "_acme-challenge." + testInstallID + ".lokal.jotti.rocks.",
			qtype: dns.TypeA,
			want:  answer{kind: kindCNAME, cname: testInstallID + ".auth.jotti.rocks.", ttl: ttlDefault},
		},
		{
			name:  "nur eine Label-Ebene ist NoData (Empty Non-Terminal mit Kindern)",
			qname: testInstallID + ".lokal.jotti.rocks.",
			qtype: dns.TypeA,
			want:  answer{kind: kindNoData},
		},
		{
			name:  "fuehrende Null im Oktett ist NXDOMAIN",
			qname: "010-0-0-1." + testInstallID + ".lokal.jotti.rocks.",
			qtype: dns.TypeA,
			want:  answer{kind: kindNXDomain},
		},
		{
			name:  "drei Label-Ebenen sind NXDOMAIN",
			qname: "a.192-168-1-50." + testInstallID + ".lokal.jotti.rocks.",
			qtype: dns.TypeA,
			want:  answer{kind: kindNXDomain},
		},
		{
			name:  "Zone-Apex beantwortet SOA",
			qname: "lokal.jotti.rocks.",
			qtype: dns.TypeSOA,
			want:  answer{kind: kindSOA},
		},
		{
			name:  "Zone-Apex beantwortet NS",
			qname: "lokal.jotti.rocks.",
			qtype: dns.TypeNS,
			want:  answer{kind: kindNS},
		},
		{
			name:  "Zone-Apex ohne A-Record ist NoData",
			qname: "lokal.jotti.rocks.",
			qtype: dns.TypeA,
			want:  answer{kind: kindNoData},
		},
		{
			name:  "auth-Subdomain wird weitergeleitet",
			qname: testInstallID + ".auth.jotti.rocks.",
			qtype: dns.TypeTXT,
			want:  answer{kind: kindForward},
		},
		{
			name:  "auth-Apex wird weitergeleitet",
			qname: "auth.jotti.rocks.",
			qtype: dns.TypeSOA,
			want:  answer{kind: kindForward},
		},
		{
			name:  "eigener Nameserver beantwortet A mit der VPS-IP",
			qname: "dns.jotti.rocks.",
			qtype: dns.TypeA,
			want:  answer{kind: kindA, ip: netip.MustParseAddr("203.0.113.10"), ttl: ttlDefault},
		},
		{
			name:  "fremde Namen werden abgelehnt",
			qname: "example.com.",
			qtype: dns.TypeA,
			want:  answer{kind: kindRefused},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := resolve(testZoneConfig(), tt.qname, tt.qtype)
			if got != tt.want {
				t.Fatalf("resolve(%q, %d): got %+v, want %+v", tt.qname, tt.qtype, got, tt.want)
			}
		})
	}
}

func newQuery(qname string, qtype uint16) *dns.Msg {
	msg := new(dns.Msg)
	msg.SetQuestion(qname, qtype)
	return msg
}

func TestBuildResponseA(t *testing.T) {
	cfg := testZoneConfig()
	req := newQuery("192-168-1-50."+testInstallID+".lokal.jotti.rocks.", dns.TypeA)

	msg := buildResponse(cfg, req, resolve(cfg, req.Question[0].Name, dns.TypeA))

	if msg.Rcode != dns.RcodeSuccess {
		t.Fatalf("rcode: got %d, want %d", msg.Rcode, dns.RcodeSuccess)
	}
	if !msg.Authoritative {
		t.Fatalf("Antwort ist nicht autoritativ")
	}
	if len(msg.Answer) != 1 {
		t.Fatalf("answer count: got %d, want 1", len(msg.Answer))
	}
	a, ok := msg.Answer[0].(*dns.A)
	if !ok {
		t.Fatalf("answer type: got %T, want *dns.A", msg.Answer[0])
	}
	if got := a.A.String(); got != "192.168.1.50" {
		t.Fatalf("ip: got %s, want 192.168.1.50", got)
	}
	if a.Hdr.Ttl != 86400 {
		t.Fatalf("ttl: got %d, want 86400", a.Hdr.Ttl)
	}
}

func TestBuildResponseCNAME(t *testing.T) {
	cfg := testZoneConfig()
	req := newQuery("_acme-challenge."+testInstallID+".lokal.jotti.rocks.", dns.TypeTXT)

	msg := buildResponse(cfg, req, resolve(cfg, req.Question[0].Name, dns.TypeTXT))

	if len(msg.Answer) != 1 {
		t.Fatalf("answer count: got %d, want 1", len(msg.Answer))
	}
	cname, ok := msg.Answer[0].(*dns.CNAME)
	if !ok {
		t.Fatalf("answer type: got %T, want *dns.CNAME", msg.Answer[0])
	}
	if want := testInstallID + ".auth.jotti.rocks."; cname.Target != want {
		t.Fatalf("target: got %s, want %s", cname.Target, want)
	}
}

func TestBuildResponseNXDomainTraegtSOAImAuthorityAbschnitt(t *testing.T) {
	cfg := testZoneConfig()
	req := newQuery("unsinn."+testInstallID+".lokal.jotti.rocks.", dns.TypeA)

	msg := buildResponse(cfg, req, resolve(cfg, req.Question[0].Name, dns.TypeA))

	if msg.Rcode != dns.RcodeNameError {
		t.Fatalf("rcode: got %d, want %d (NXDOMAIN)", msg.Rcode, dns.RcodeNameError)
	}
	if len(msg.Answer) != 0 {
		t.Fatalf("answer count: got %d, want 0", len(msg.Answer))
	}
	if len(msg.Ns) != 1 {
		t.Fatalf("authority count: got %d, want 1", len(msg.Ns))
	}
	soa, ok := msg.Ns[0].(*dns.SOA)
	if !ok {
		t.Fatalf("authority type: got %T, want *dns.SOA", msg.Ns[0])
	}
	if soa.Hdr.Name != cfg.zone {
		t.Fatalf("soa owner: got %s, want %s", soa.Hdr.Name, cfg.zone)
	}
}

func TestBuildResponseNoData(t *testing.T) {
	cfg := testZoneConfig()
	req := newQuery("192-168-1-50."+testInstallID+".lokal.jotti.rocks.", dns.TypeAAAA)

	msg := buildResponse(cfg, req, resolve(cfg, req.Question[0].Name, dns.TypeAAAA))

	if msg.Rcode != dns.RcodeSuccess {
		t.Fatalf("rcode: got %d, want %d (NoData ist NOERROR)", msg.Rcode, dns.RcodeSuccess)
	}
	if len(msg.Answer) != 0 {
		t.Fatalf("answer count: got %d, want 0", len(msg.Answer))
	}
	if len(msg.Ns) != 1 {
		t.Fatalf("authority count: got %d, want 1 (SOA)", len(msg.Ns))
	}
}

func TestBuildResponseApex(t *testing.T) {
	cfg := testZoneConfig()

	soaMsg := buildResponse(cfg, newQuery(cfg.zone, dns.TypeSOA), resolve(cfg, cfg.zone, dns.TypeSOA))
	if len(soaMsg.Answer) != 1 {
		t.Fatalf("SOA answer count: got %d, want 1", len(soaMsg.Answer))
	}
	soa, ok := soaMsg.Answer[0].(*dns.SOA)
	if !ok {
		t.Fatalf("SOA answer type: got %T, want *dns.SOA", soaMsg.Answer[0])
	}
	if soa.Ns != cfg.nsName {
		t.Fatalf("SOA mname: got %s, want %s", soa.Ns, cfg.nsName)
	}

	nsMsg := buildResponse(cfg, newQuery(cfg.zone, dns.TypeNS), resolve(cfg, cfg.zone, dns.TypeNS))
	if len(nsMsg.Answer) != 1 {
		t.Fatalf("NS answer count: got %d, want 1", len(nsMsg.Answer))
	}
	ns, ok := nsMsg.Answer[0].(*dns.NS)
	if !ok {
		t.Fatalf("NS answer type: got %T, want *dns.NS", nsMsg.Answer[0])
	}
	if ns.Ns != cfg.nsName {
		t.Fatalf("NS target: got %s, want %s", ns.Ns, cfg.nsName)
	}
}

func TestBuildResponseRefused(t *testing.T) {
	cfg := testZoneConfig()
	req := newQuery("example.com.", dns.TypeA)

	msg := buildResponse(cfg, req, resolve(cfg, req.Question[0].Name, dns.TypeA))

	if msg.Rcode != dns.RcodeRefused {
		t.Fatalf("rcode: got %d, want %d (REFUSED)", msg.Rcode, dns.RcodeRefused)
	}
	if msg.Authoritative {
		t.Fatalf("REFUSED darf nicht autoritativ sein")
	}
}
