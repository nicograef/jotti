package main

import (
	"encoding/base64"
	"html/template"
	"log"
	"net/http"
	"net/netip"
	"time"

	"rsc.io/qr"
)

const (
	// statusListenAddr ist die Adresse der Status-Seite im Container. Im Compose
	// wird sie nur an 127.0.0.1 des Hosts gemappt — sichtbar nur am Kassenrechner,
	// nicht aus dem WLAN.
	statusListenAddr = ":8484"
	// caddyHTTPSAddr ist der lokale HTTPS-Endpunkt des eigenen Caddy, gegen den
	// die Zertifikats-Probe ihren Handshake fährt.
	caddyHTTPSAddr = "127.0.0.1:443"
	// rebindGuideURL verweist auf die Router-Anleitung zum DNS-Rebind-Schutz in
	// der Projekt-Doku.
	rebindGuideURL = "https://github.com/nicograef/jotti/blob/main/docs/leitfaden.md#fehlersuche"
)

// statusConfig bündelt die Startwerte, aus denen die Status-Seite ihre festen
// Adressen ableitet (LAN-IP und Install-ID ändern sich über die Prozesslaufzeit
// nicht).
type statusConfig struct {
	zone      string
	state     InstallState
	hasState  bool
	lanIP     netip.Addr
	lanOK     bool
	leStaging bool
}

// statusServer serviert die lokale Status-Seite. greenURL/fallbackURL stehen beim
// Start fest; Zertifikat und Rebind werden bei jedem Seitenaufruf frisch geprüft,
// damit die Seite ohne Neustart von „Fallback" auf „grün" wechselt. probeCert und
// checkRebind sind Felder, damit Tests sie ohne echtes Netz ersetzen können.
type statusServer struct {
	greenURL    string
	fallbackURL string
	leStaging   bool
	probeCert   func() certState
	checkRebind func() bool
}

// newStatusServer leitet die Adressen aus der Config ab und verdrahtet die echten
// Proben gegen den eigenen Caddy bzw. den System-Resolver. Ohne LAN-IP gibt es
// keine Adresse, ohne State keinen grünen Hostnamen.
func newStatusServer(cfg statusConfig) *statusServer {
	var greenURL, fallbackURL, hostname string
	if cfg.lanOK {
		fallbackURL = "https://" + cfg.lanIP.String()
		if cfg.hasState {
			hostname = deriveHostname(cfg.lanIP, cfg.state.Subdomain, cfg.zone)
			greenURL = "https://" + hostname
		}
	}
	return &statusServer{
		greenURL:    greenURL,
		fallbackURL: fallbackURL,
		leStaging:   cfg.leStaging,
		probeCert:   func() certState { return probeCert(caddyHTTPSAddr, hostname) },
		checkRebind: func() bool { return checkRebind(hostname, cfg.lanIP, systemLookupIP) },
	}
}

// listenAndServe startet den HTTP-Server der Status-Seite. Blockiert bis zum
// Fehler; der Aufrufer betreibt ihn neben Caddy in einer eigenen Goroutine.
func (s *statusServer) listenAndServe() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handle)
	srv := &http.Server{
		Addr:              statusListenAddr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}
	return srv.ListenAndServe()
}

// currentView prüft Zertifikat und Rebind (nur wenn überhaupt eine grüne Adresse
// möglich ist) und entscheidet die Anzeige.
func (s *statusServer) currentView() statusView {
	in := statusInputs{greenURL: s.greenURL, fallbackURL: s.fallbackURL}
	if s.greenURL != "" {
		in.cert = s.probeCert()
		in.rebindOK = s.checkRebind()
	}
	return decideStatus(in)
}

func (s *statusServer) handle(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	data := s.pageData(s.currentView())
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := statusTemplate.Execute(w, data); err != nil {
		log.Printf("Status-Seite rendern: %v", err)
	}
}

// pageData ist das Template-Modell der Status-Seite.
type pageData struct {
	Headline    string
	Body        string
	PrimaryURL  string
	GreenURL    string
	FallbackURL string
	GreenActive bool
	ShowRebind  bool
	QR          template.URL
	Refresh     bool
	LEStaging   bool
	RebindGuide string
}

func (s *statusServer) pageData(view statusView) pageData {
	headline, body := noticeText(view.notice)
	d := pageData{
		Headline:    headline,
		Body:        body,
		PrimaryURL:  view.primaryURL,
		GreenURL:    s.greenURL,
		FallbackURL: s.fallbackURL,
		GreenActive: view.greenActive,
		ShowRebind:  view.notice == noticeRebind,
		Refresh:     view.refresh,
		LEStaging:   s.leStaging,
		RebindGuide: rebindGuideURL,
	}
	if view.showQR && s.greenURL != "" {
		d.QR = qrDataURI(s.greenURL)
	}
	return d
}

// noticeText liefert Überschrift und Fließtext zum Hinweis (deutschsprachig wie
// die übrige Vereins-Doku).
func noticeText(n notice) (headline, body string) {
	switch n {
	case noticeGreen:
		return "Vertrauenswürdige Adresse aktiv ✓",
			"Öffne die Adresse mit dem Smartphone oder scanne den QR-Code — grünes Schloss, keine Warnung."
	case noticeIssuing:
		return "Vertrauenswürdiges Zertifikat wird ausgestellt …",
			"Das dauert wenige Sekunden bis Minuten und braucht Internet. Bis dahin die Fallback-Adresse nutzen und die einmalige Zertifikatswarnung im Browser bestätigen. Diese Seite aktualisiert sich automatisch."
	case noticeRenewing:
		return "Zertifikat wird erneuert …",
			"Das Zertifikat ist abgelaufen (normal nach einer längeren Pause) und wird im Hintergrund erneuert. Solange die Fallback-Adresse nutzen. Diese Seite aktualisiert sich automatisch."
	case noticeRebind:
		return "DNS-Rebind-Schutz erkannt",
			"Der Router beantwortet den Namen nicht mit der lokalen IP, deshalb ist die vertrauenswürdige Adresse im WLAN nicht erreichbar. Trage lokal.jotti.rocks als Ausnahme im Rebind-Schutz des Routers ein oder nutze die Fallback-Adresse:"
	default: // noticeNoGreen
		return "Warte auf Registrierung und Netzwerk …",
			"Die einmalige Registrierung oder die LAN-IP fehlt noch. Bis dahin die Fallback-Adresse nutzen. Diese Seite aktualisiert sich automatisch."
	}
}

// qrDataURI kodiert text als QR-Code-PNG und liefert eine data:-URI zum direkten
// Einbetten in ein <img>-Tag. Ein Kodierfehler (zu langer Text) ⇒ leere URI; die
// Seite zeigt dann eben keinen Code.
func qrDataURI(text string) template.URL {
	code, err := qr.Encode(text, qr.M)
	if err != nil {
		return ""
	}
	return template.URL("data:image/png;base64," + base64.StdEncoding.EncodeToString(code.PNG()))
}

var statusTemplate = template.Must(template.New("status").Parse(statusHTML))

const statusHTML = `<!doctype html>
<html lang="de">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
{{if .Refresh}}<meta http-equiv="refresh" content="5">{{end}}
<title>jotti — lokaler Status</title>
<style>
  :root { color-scheme: light dark; }
  body { font-family: system-ui, -apple-system, sans-serif; margin: 0; padding: 2rem 1rem;
         line-height: 1.5; background: #f5f5f5; color: #1a1a1a; }
  main { max-width: 40rem; margin: 0 auto; }
  .card { background: #fff; border-radius: 12px; padding: 1.5rem; margin-bottom: 1.25rem;
          box-shadow: 0 1px 3px rgba(0,0,0,.1); }
  h1 { font-size: 1.4rem; margin: 0 0 1rem; }
  .headline { font-size: 1.15rem; font-weight: 600; margin: 0 0 .5rem; }
  .green .headline { color: #137333; }
  .addr { display: block; font-size: 1.1rem; font-weight: 600; word-break: break-all; margin: .5rem 0; }
  .qr { text-align: center; margin: 1rem 0; }
  .qr img { width: 240px; height: 240px; image-rendering: pixelated; background: #fff;
            padding: 8px; border-radius: 8px; }
  .label { text-transform: uppercase; letter-spacing: .04em; font-size: .72rem; color: #5f6368;
           margin-bottom: .25rem; }
  .muted { color: #5f6368; font-size: .9rem; }
  a { color: #1a73e8; }
  .staging { background: #fff4e5; border: 1px solid #f0c36d; }
</style>
</head>
<body>
<main>
  <h1>jotti — lokaler Status</h1>

  <section class="card{{if .GreenActive}} green{{end}}">
    <p class="headline">{{.Headline}}</p>
    <p>{{.Body}}{{if .ShowRebind}} <a href="{{.RebindGuide}}">Router-Anleitung öffnen</a>.{{end}}</p>
    {{if .QR}}
    <div class="qr">
      <img src="{{.QR}}" alt="QR-Code zur vertrauenswürdigen Adresse">
      <p class="muted">Mit der Smartphone-Kamera scannen</p>
    </div>
    {{end}}
    <p class="label">Diese Adresse öffnen</p>
    <a class="addr" href="{{.PrimaryURL}}">{{.PrimaryURL}}</a>
  </section>

  {{if and .GreenURL .FallbackURL}}
  <section class="card">
    <p class="label">Alle Adressen</p>
    <p><strong>Vertrauenswürdig (grünes Schloss):</strong><br>
       <a href="{{.GreenURL}}">{{.GreenURL}}</a></p>
    <p><strong>Fallback (einmalige Warnung):</strong><br>
       <a href="{{.FallbackURL}}">{{.FallbackURL}}</a></p>
  </section>
  {{end}}

  <section class="card">
    <p class="label">WLAN-Hinweis</p>
    <p class="muted">Das Smartphone muss im <strong>Vereins-WLAN</strong> sein — nicht im
       Mobilfunknetz und nicht im <strong>Gastnetz</strong> (Gastnetze blockieren den Zugriff
       auf das Kassengerät).</p>
  </section>

  {{if .LEStaging}}
  <section class="card staging">
    <p class="muted"><strong>Testmodus:</strong> Zertifikate stammen aus der
       Let's-Encrypt-Staging-CA und sind im Browser nicht vertrauenswürdig — die Anzeige bleibt
       deshalb beim Fallback.</p>
  </section>
  {{end}}
</main>
</body>
</html>
`
