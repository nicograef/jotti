package main

import "fmt"

// contentSecurityPolicy gilt für alle Caddy-Sites (über proxySnippet) und wörtlich
// auch für die demo-Site in nginx.rocks.conf; TestNginxRocksConfCarriesSameCSP hält
// beide Kopien zusammen.
const contentSecurityPolicy = "default-src 'none'; base-uri 'self'; frame-ancestors 'none'; form-action 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data: blob:; font-src 'self'; connect-src 'self'; manifest-src 'self'; worker-src 'self' blob:; media-src 'self'; frame-src 'none'; object-src 'none'; upgrade-insecure-requests"

// hstsLAN trägt nur max-age: der Zugriff läuft auch über die rohe LAN-IP
// (Fallback-Site), für die includeSubDomains/preload nicht zutreffen. hstsPublic ist
// der stärkere Wert für die öffentliche Domain (Parität zur prod-nginx).
const (
	hstsLAN    = "max-age=31536000"
	hstsPublic = "max-age=63072000; includeSubDomains; preload"
)

// leStagingCA: Entwicklung und Tests nutzen die LE-Staging-Umgebung
// (PROXY_LE_STAGING), um die Rate-Limits der echten Zone zu schonen.
const leStagingCA = "https://acme-staging-v02.api.letsencrypt.org/directory"

// challengeResolvers sind die Resolver für Caddys DNS-01-Propagation-Prüfung —
// bewusst öffentliche statt des LAN-Resolvers. Viele Heimrouter (empirisch Telekom
// Speedport → Telekom-Upstream) negativ-cachen die kurzlebigen acme-dns-TXT-Records
// (`<install-id>.auth.jotti.rocks`, TTL 1) aggressiv und liefern danach NXDOMAIN;
// die Prüfung liefe sonst über genau diesen Resolver in einen Timeout, obwohl Let's
// Encrypt den Record über eigene Resolver sieht.
const challengeResolvers = "1.1.1.1 8.8.8.8"

type caddyfileInput struct {
	state      InstallState // acme-dns-Credentials für die Wildcard-Site
	hasState   bool         // false ⇒ nur die Fallback-Site rendern
	zone       string       // z. B. "lokal.jotti.rocks"
	acmeDNSURL string       // acme-dns-API, z. B. "https://auth.jotti.rocks"
	leStaging  bool         // true ⇒ Zertifikate über die LE-Staging-CA holen
}

// renderCaddyfile erzeugt den Caddyfile des lokalen Stacks: die Wildcard-Site (nur
// mit acme-dns-Credentials) und immer die Fallback-Site mit Caddys interner CA.
func renderCaddyfile(in caddyfileInput) string {
	wildcard := ""
	if in.hasState {
		wildcard = wildcardSite(in) + "\n\n"
	}

	return fmt.Sprintf(`# Generiert vom jotti-reverse-proxy beim Start — nicht von Hand bearbeiten.
{
	admin off
}

# Gemeinsame Proxy- und Security-Header-Konfiguration für beide Sites.
%s

%s# Fallback-Site: Clients verbinden über die LAN-IP des Hosts. on_demand stellt
# beim ersten TLS-Handshake ein internes Zertifikat aus; sign_with_root + 365d
# halten die Browserwarnung einmalig.
https:// {
	tls {
		issuer internal {
			sign_with_root
			lifetime 365d
		}
		on_demand
	}
	import jotti_proxy
}

http:// {
	redir https://{host}{uri} permanent
}
`, proxySnippet(hstsLAN, false), wildcard)
}

// proxySnippet rendert das gemeinsame `(jotti_proxy)`-Snippet, damit Header und CSP
// über alle Modi identisch bleiben. Das Rate-Limit bildet die prod-nginx-Vorgabe
// (10r/s, burst 20) ab: caddy-ratelimit nutzt ein gleitendes Fenster ohne separaten
// Burst-Begriff, daher entspricht der nginx-Spitzenwert (durch `nodelay` sofort
// bedient) hier `events 30` pro `window 1s`.
func proxySnippet(hsts string, rateLimited bool) string {
	rateLimit := ""
	if rateLimited {
		rateLimit = `
		rate_limit {
			zone api {
				key {remote_host}
				events 30
				window 1s
			}
		}
`
	}

	return fmt.Sprintf(`(jotti_proxy) {
	header {
		Strict-Transport-Security %q
		X-Content-Type-Options "nosniff"
		X-Frame-Options "DENY"
		Referrer-Policy "no-referrer-when-downgrade"
		Permissions-Policy "geolocation=(), microphone=()"
		Content-Security-Policy "%s"
	}

	# Backend-API unter /api/ (Prefix beim Proxyen entfernen)
	handle_path /api/* {%s
		reverse_proxy backend:3000
	}

	# Frontend-SPA
	handle {
		reverse_proxy frontend:80
	}
}`, hsts, contentSecurityPolicy, rateLimit)
}

type publicInput struct {
	domain      string // öffentliche Domain, z. B. "jotti.meinverein.de"
	email       string // Kontakt-E-Mail für den ACME-Account
	wwwRedirect bool   // true ⇒ www.<domain> dauerhaft auf <domain> umleiten
	leStaging   bool   // true ⇒ Zertifikate über die LE-Staging-CA holen (Tests)
}

// renderPublicCaddyfile erzeugt den Caddyfile des Self-Hoster-prod-Stacks: dasselbe
// `(jotti_proxy)`-Snippet wie der LAN-Mode, aber stärkeres HSTS und das
// /api/-Rate-Limit.
func renderPublicCaddyfile(in publicInput) string {
	staging := ""
	if in.leStaging {
		staging = "\n\tacme_ca " + leStagingCA
	}

	www := ""
	if in.wwwRedirect {
		www = fmt.Sprintf("\n\nwww.%s {\n\tredir https://%s{uri} permanent\n}", in.domain, in.domain)
	}

	return fmt.Sprintf(`# Generiert vom jotti-reverse-proxy beim Start — nicht von Hand bearbeiten.
{
	admin off
	email %s%s
}

# Gemeinsame Proxy- und Security-Header-Konfiguration.
%s

%s {
	import jotti_proxy
}%s
`, in.email, staging, proxySnippet(hstsPublic, true), in.domain, www)
}

// renderHTTPOnlyCaddyfile bedient ausschließlich Klartext-HTTP auf :80 — nur für
// die E2E-Testumgebung (docker-compose.e2e.yml, PROXY_HTTP_ONLY=1), wo der Stack
// lokal bleibt. Das /api/-Rate-Limit bleibt aus, damit Test-Suiten nicht gedrosselt
// werden.
func renderHTTPOnlyCaddyfile() string {
	return fmt.Sprintf(`# Generiert vom jotti-reverse-proxy beim Start — nicht von Hand bearbeiten.
{
	admin off
	auto_https off
}

# Gemeinsame Proxy- und Security-Header-Konfiguration.
%s

http:// {
	import jotti_proxy
}
`, proxySnippet(hstsLAN, false))
}

// wildcardSite rendert `*.<install-id>.<zone>` mit DNS-01-Challenge über acme-dns;
// bis Caddy ausgestellt hat (oder offline) trägt die Fallback-Site.
func wildcardSite(in caddyfileInput) string {
	caBlock := ""
	if in.leStaging {
		caBlock = "\n\t\tca " + leStagingCA
	}

	return fmt.Sprintf(`# Vertrauenswürdige Wildcard-Site: echtes Let's-Encrypt-Zertifikat (grünes
# Schloss, keine Warnung) via DNS-01 über acme-dns.
*.%s.%s {
	tls {
		dns acmedns {
			username %q
			password %q
			subdomain %q
			server_url %q
		}
		resolvers %s%s
	}
	import jotti_proxy
}`, in.state.Subdomain, in.zone, in.state.Username, in.state.Password, in.state.Subdomain, in.acmeDNSURL, challengeResolvers, caBlock)
}
