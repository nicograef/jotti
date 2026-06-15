package main

import (
	"net/netip"
	"strings"

	"github.com/miekg/dns"
)

// TTLs in Sekunden. Das Mapping Name → IP ist unveränderlich (die IP steckt
// im Namen selbst), daher dürfen A-Antworten lange gecacht werden — Geräte,
// die den Namen einmal aufgelöst haben, überstehen so einen Internet-Ausfall
// während des Fests. Negative Antworten bleiben kurz.
const (
	ttlA        = 86400
	ttlDefault  = 3600
	ttlNegative = 300
)

// zoneConfig beschreibt die Zonen, für die der Resolver zuständig ist.
// Alle Namen sind kanonisch (kleingeschrieben, mit Punkt am Ende).
type zoneConfig struct {
	zone     string     // autoritative Zone, z. B. "lokal.jotti.rocks."
	authZone string     // acme-dns-Zone, z. B. "auth.jotti.rocks." (wird weitergeleitet)
	nsName   string     // Name des eigenen Nameservers, z. B. "dns.jotti.rocks."
	nsIP     netip.Addr // öffentliche IPv4 des Nameservers (VPS)
}

type answerKind int

const (
	kindRefused  answerKind = iota // Name außerhalb der eigenen Zonen
	kindForward                    // auth-Zone → Anfrage an acme-dns weiterleiten
	kindA                          // A-Record mit answer.ip
	kindCNAME                      // CNAME auf answer.cname
	kindSOA                        // SOA der Zone
	kindNS                         // NS der Zone
	kindNoData                     // Name existiert, hat aber keinen Record des gefragten Typs
	kindNXDomain                   // Name existiert nicht
)

// answer ist die berechnete Antwort-Entscheidung für eine DNS-Frage.
type answer struct {
	kind  answerKind
	ip    netip.Addr // gesetzt für kindA
	cname string     // gesetzt für kindCNAME (kanonisches Ziel)
	ttl   uint32     // gesetzt für kindA und kindCNAME
}

// resolve entscheidet rein rechnerisch (ohne I/O und ohne Zustand), wie eine
// DNS-Frage zu beantworten ist. Innerhalb der Zone gilt das Namensschema
// `<lan-ip-mit-bindestrichen>.<install-id>.<zone>`; die Install-ID ist die bei
// acme-dns registrierte Subdomain, deshalb ist auch der Challenge-CNAME
// `_acme-challenge.<install-id>.<zone>` → `<install-id>.<auth-zone>` berechenbar.
func resolve(cfg zoneConfig, qname string, qtype uint16) answer {
	name := dns.CanonicalName(qname)

	// Eigener Nameserver: A auf die VPS-IP (Diagnose; der reguläre A-Record
	// liegt beim DNS-Hoster der Elternzone).
	if name == cfg.nsName && qtype == dns.TypeA {
		return answer{kind: kindA, ip: cfg.nsIP, ttl: ttlDefault}
	}

	// Die auth-Zone gehört acme-dns; der Resolver leitet nur weiter.
	if name == cfg.authZone || strings.HasSuffix(name, "."+cfg.authZone) {
		return answer{kind: kindForward}
	}

	if name == cfg.zone {
		return apexAnswer(qtype)
	}
	if !strings.HasSuffix(name, "."+cfg.zone) {
		return answer{kind: kindRefused}
	}

	labels := strings.Split(strings.TrimSuffix(name, "."+cfg.zone), ".")
	// Eine einzelne Label-Ebene (z. B. `<install-id>.<zone>`) ist ein Empty
	// Non-Terminal: der Name selbst trägt keinen Record, hat aber berechenbare
	// Kinder (`<lan-ip>.<install-id>`, `_acme-challenge.<install-id>`). Dafür ist
	// NODATA korrekt, nicht NXDOMAIN — sonst dürfte ein Resolver mit
	// QNAME-Minimisation (RFC 7816/9156) das NXDOMAIN cachen und die Auflösung
	// der Kindnamen verweigern (RFC 8020).
	if len(labels) == 1 {
		return answer{kind: kindNoData}
	}
	if len(labels) != 2 {
		return answer{kind: kindNXDomain}
	}

	// `_acme-challenge.<install-id>` → CNAME `<install-id>.<auth-zone>`.
	// Ein CNAME gilt für jeden Anfrage-Typ, daher keine qtype-Prüfung.
	if labels[0] == "_acme-challenge" {
		return answer{kind: kindCNAME, cname: labels[1] + "." + cfg.authZone, ttl: ttlDefault}
	}

	ip, ok := parseDashedIPv4(labels[0])
	if !ok {
		return answer{kind: kindNXDomain}
	}
	if qtype != dns.TypeA {
		return answer{kind: kindNoData}
	}
	return answer{kind: kindA, ip: ip, ttl: ttlA}
}

// apexAnswer beantwortet Fragen an den Zone-Apex: SOA und NS, damit die
// Delegation funktioniert; für alle anderen Typen existiert der Name, hat
// aber keine Records.
func apexAnswer(qtype uint16) answer {
	switch qtype {
	case dns.TypeSOA:
		return answer{kind: kindSOA}
	case dns.TypeNS:
		return answer{kind: kindNS}
	default:
		return answer{kind: kindNoData}
	}
}

// parseDashedIPv4 liest eine IPv4-Adresse aus einem Label wie "192-168-1-50".
// Auch private Adressen sind gültig — genau dafür existiert der Resolver.
// netip.ParseAddr ist strikt und lehnt nicht-kanonische Labels (führende
// Nullen, Vorzeichen, falsche Oktett-Anzahl) ab; dashedIPv4 im reverse-proxy
// erzeugt nur kanonische Namen, der Round-Trip bleibt also unberührt.
func parseDashedIPv4(label string) (netip.Addr, bool) {
	ip, err := netip.ParseAddr(strings.ReplaceAll(label, "-", "."))
	if err != nil || !ip.Is4() {
		return netip.Addr{}, false
	}
	return ip, true
}

// buildResponse baut aus der Antwort-Entscheidung die vollständige
// DNS-Antwort. Ebenfalls rein — kindForward behandelt der Server selbst.
func buildResponse(cfg zoneConfig, req *dns.Msg, ans answer) *dns.Msg {
	msg := new(dns.Msg)
	msg.SetReply(req)
	msg.Authoritative = true

	q := req.Question[0]
	header := func(rrtype uint16, ttl uint32) dns.RR_Header {
		return dns.RR_Header{Name: q.Name, Rrtype: rrtype, Class: dns.ClassINET, Ttl: ttl}
	}

	switch ans.kind {
	case kindA:
		msg.Answer = []dns.RR{&dns.A{Hdr: header(dns.TypeA, ans.ttl), A: ans.ip.AsSlice()}}
	case kindCNAME:
		msg.Answer = []dns.RR{&dns.CNAME{Hdr: header(dns.TypeCNAME, ans.ttl), Target: ans.cname}}
	case kindSOA:
		msg.Answer = []dns.RR{soaRecord(cfg, ttlDefault)}
	case kindNS:
		msg.Answer = []dns.RR{&dns.NS{Hdr: header(dns.TypeNS, ttlDefault), Ns: cfg.nsName}}
	case kindNoData:
		msg.Ns = []dns.RR{soaRecord(cfg, ttlNegative)}
	case kindNXDomain:
		msg.Rcode = dns.RcodeNameError
		msg.Ns = []dns.RR{soaRecord(cfg, ttlNegative)}
	case kindRefused:
		msg.Rcode = dns.RcodeRefused
		msg.Authoritative = false
	default:
		// kindForward behandelt der Server vor buildResponse; jeder andere
		// unbehandelte Fall ist ein Programmierfehler.
		panic("buildResponse: unbehandelter answerKind")
	}
	return msg
}

// soaRecord liefert den SOA-Record der Zone. Die Serial ist konstant: Die
// Zone ist rein rechnerisch, ändert sich nie und kennt keine Zonentransfers.
func soaRecord(cfg zoneConfig, ttl uint32) *dns.SOA {
	return &dns.SOA{
		Hdr:     dns.RR_Header{Name: cfg.zone, Rrtype: dns.TypeSOA, Class: dns.ClassINET, Ttl: ttl},
		Ns:      cfg.nsName,
		Mbox:    "hostmaster." + cfg.zone,
		Serial:  1,
		Refresh: 7200,
		Retry:   3600,
		Expire:  1209600,
		Minttl:  ttlNegative,
	}
}
