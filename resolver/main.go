package main

import (
	"fmt"
	"log"
	"net"
	"net/netip"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/miekg/dns"
)

// Die Zonen müssen zum Hostnamen des reverse-proxy (lokal.jotti.rocks) und zu acme-dns
// (auth.jotti.rocks) passen, darum sind sie nicht konfigurierbar.
const (
	zone       = "lokal.jotti.rocks."
	authZone   = "auth.jotti.rocks."
	nsName     = "dns.jotti.rocks."
	listenAddr = ":53"
)

// forwardTimeout begrenzt die Weiterleitung an acme-dns; die Instanz läuft
// Docker-intern auf demselben Host, längere Wartezeiten helfen nicht.
const forwardTimeout = 5 * time.Second

type resolverConfig struct {
	zones       zoneConfig
	forwardAddr string // acme-dns (Docker-intern), host:port
}

func loadConfigFromEnv(getenv func(string) string) (resolverConfig, error) {
	nsIPRaw := strings.TrimSpace(getenv("RESOLVER_NS_IP"))
	if nsIPRaw == "" {
		return resolverConfig{}, fmt.Errorf("RESOLVER_NS_IP ist erforderlich (öffentliche IPv4 des Nameservers)")
	}
	nsIP, err := netip.ParseAddr(nsIPRaw)
	if err != nil || !nsIP.Is4() {
		return resolverConfig{}, fmt.Errorf("RESOLVER_NS_IP muss eine IPv4-Adresse sein")
	}

	forwardAddr := strings.TrimSpace(getenv("RESOLVER_AUTH_FORWARD"))
	if forwardAddr == "" {
		return resolverConfig{}, fmt.Errorf("RESOLVER_AUTH_FORWARD ist erforderlich (Adresse von acme-dns, z. B. acme-dns:53)")
	}
	if _, _, err := net.SplitHostPort(forwardAddr); err != nil {
		forwardAddr = net.JoinHostPort(forwardAddr, "53")
	}

	return resolverConfig{
		zones: zoneConfig{
			zone:     zone,
			authZone: authZone,
			nsName:   nsName,
			nsIP:     nsIP,
		},
		forwardAddr: forwardAddr,
	}, nil
}

type handler struct {
	cfg         zoneConfig
	forwardAddr string
}

func (h *handler) ServeDNS(w dns.ResponseWriter, req *dns.Msg) {
	if len(req.Question) != 1 {
		formErr := new(dns.Msg).SetRcode(req, dns.RcodeFormatError)
		_ = w.WriteMsg(formErr)
		return
	}

	q := req.Question[0]
	ans := resolve(h.cfg, q.Name, q.Qtype)
	if ans.kind == kindForward {
		h.forward(w, req)
		return
	}
	_ = w.WriteMsg(buildResponse(h.cfg, req, ans))
}

func (h *handler) forward(w dns.ResponseWriter, req *dns.Msg) {
	client := &dns.Client{Net: w.RemoteAddr().Network(), Timeout: forwardTimeout}
	resp, _, err := client.Exchange(req, h.forwardAddr)
	if err != nil {
		log.Printf("Weiterleitung an %s fehlgeschlagen: %v", h.forwardAddr, err)
		servfail := new(dns.Msg).SetRcode(req, dns.RcodeServerFailure)
		_ = w.WriteMsg(servfail)
		return
	}
	_ = w.WriteMsg(resp)
}

func main() {
	config, err := loadConfigFromEnv(os.Getenv)
	if err != nil {
		log.Fatal(err)
	}

	h := &handler{cfg: config.zones, forwardAddr: config.forwardAddr}
	udpServer := &dns.Server{Addr: listenAddr, Net: "udp", Handler: h}
	tcpServer := &dns.Server{Addr: listenAddr, Net: "tcp", Handler: h}

	log.Printf("jotti DNS-Resolver gestartet | Zone: %s | NS: %s (%s) | auth-Forward: %s → %s | Listen: %s",
		config.zones.zone, config.zones.nsName, config.zones.nsIP,
		config.zones.authZone, config.forwardAddr, listenAddr)

	serverErrors := make(chan error, 2)
	go func() { serverErrors <- udpServer.ListenAndServe() }()
	go func() { serverErrors <- tcpServer.ListenAndServe() }()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		log.Fatalf("DNS-Server beendet: %v", err)
	case sig := <-quit:
		log.Printf("Signal %s empfangen. Beende.", sig)
		_ = udpServer.Shutdown()
		_ = tcpServer.Shutdown()
	}
}
