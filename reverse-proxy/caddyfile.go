package main

import "fmt"

// contentSecurityPolicy ist die CSP des lokalen Stacks — unverändert aus dem
// früheren nginx-/Caddy-Setup übernommen (Parität). Muss identisch zur CSP in
// nginx.conf und nginx.rocks.conf (demo) bleiben.
const contentSecurityPolicy = "default-src 'none'; base-uri 'self'; frame-ancestors 'none'; form-action 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data: blob:; font-src 'self'; connect-src 'self'; manifest-src 'self'; worker-src 'self' blob:; media-src 'self'; frame-src 'none'; object-src 'none'; upgrade-insecure-requests"

// hstsLAN ist der HSTS-Wert des LAN-Stacks (lokal/release): nur max-age, da der
// Zugriff auch über die rohe LAN-IP (Fallback-Site) läuft, für die
// includeSubDomains/preload nicht zutreffen.
//
// hstsPublic ist der stärkere Public-Wert (Parität zur prod-nginx): zwei Jahre
// inklusive Subdomains und Preload — eine bewusste Zusage für die öffentliche
// Domain des Self-Hosters.
const (
	hstsLAN    = "max-age=31536000"
	hstsPublic = "max-age=63072000; includeSubDomains; preload"
)

// leStagingCA ist das ACME-Verzeichnis der Let's-Encrypt-Staging-Umgebung.
// Entwicklung und Tests nutzen es, um die Rate-Limits der echten Zone zu
// schonen (Schalter PROXY_LE_STAGING).
const leStagingCA = "https://acme-staging-v02.api.letsencrypt.org/directory"

// challengeResolvers sind die rekursiven DNS-Resolver, die Caddy für die
// DNS-01-Propagation-Prüfung verwendet — bewusst öffentliche Resolver statt des
// LAN-Resolvers. Viele Heimrouter (empirisch z. B. Telekom Speedport →
// Telekom-Upstream) negativ-cachen die kurzlebigen acme-dns-TXT-Records
// (`<install-id>.auth.jotti.rocks`, TTL 1) aggressiv und liefern danach
// NXDOMAIN. Caddys Propagation-Prüfung läuft sonst über genau diesen
// LAN-Resolver und läuft in einen Timeout, obwohl der Record real ausgestellt
// ist und Let's Encrypt ihn über die eigenen Resolver sieht. Öffentliche
// Resolver lösen den CNAME→TXT-Pfad (`_acme-challenge` → `…auth.jotti.rocks`)
// zuverlässig auf und entkoppeln die Ausstellung vom Router des Vereins-WLANs.
const challengeResolvers = "1.1.1.1 8.8.8.8"

// caddyfileInput beschreibt, woraus der Caddyfile gerendert wird.
type caddyfileInput struct {
	state      InstallState // acme-dns-Credentials für die Wildcard-Site
	hasState   bool         // false ⇒ nur die Fallback-Site rendern
	zone       string       // z. B. "lokal.jotti.rocks"
	acmeDNSURL string       // acme-dns-API, z. B. "https://auth.jotti.rocks"
	leStaging  bool         // true ⇒ Zertifikate über die LE-Staging-CA holen
}

// renderCaddyfile erzeugt den vollständigen Caddyfile des lokalen Stacks: die
// vertrauenswürdige Wildcard-Site (Let's Encrypt via DNS-01 über acme-dns, nur
// wenn Credentials vorliegen) und immer die Fallback-Site mit Caddys interner CA
// (eingebauter Option-2-Ersatz). Beide Sites proxyen identisch über ein
// gemeinsames Snippet. Reine Funktion ohne I/O.
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
# halten die Browserwarnung einmalig (wie das frühere selbstsignierte Zertifikat).
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

// proxySnippet rendert das gemeinsame `(jotti_proxy)`-Snippet (Security-Header,
// CSP, API-Proxy, SPA-Proxy), das LAN- und Public-Mode teilen, damit Header und
// CSP über beide Modi identisch bleiben. hsts ist der Strict-Transport-Security-
// Wert (im Public-Mode stärker); rateLimited schaltet das /api/-Rate-Limit zu.
//
// Das Rate-Limit bildet die prod-nginx-Vorgabe (10r/s, burst 20) ab: Caddys
// caddy-ratelimit-Modul nutzt ein gleitendes Fenster ohne separaten Burst-
// Begriff, daher entspricht der Spitzenwert von nginx (rate + burst, durch
// `nodelay` sofort bedient) hier `events 30` pro `window 1s`.
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

// publicInput beschreibt, woraus der Public-Mode-Caddyfile gerendert wird: eine
// einzige öffentliche Site mit automatischem Let's-Encrypt-Zertifikat
// (HTTP-01/TLS-ALPN). Den Modus nutzt der Self-Hoster-prod-Stack.
type publicInput struct {
	domain      string // öffentliche Domain, z. B. "jotti.meinverein.de"
	email       string // Kontakt-E-Mail für den ACME-Account
	wwwRedirect bool   // true ⇒ www.<domain> dauerhaft auf <domain> umleiten
	leStaging   bool   // true ⇒ Zertifikate über die LE-Staging-CA holen (Tests)
}

// renderPublicCaddyfile erzeugt den Caddyfile des Self-Hoster-prod-Stacks: eine
// öffentliche Site für die Domain mit automatischem Let's-Encrypt-Zertifikat.
// Sie proxyt über dasselbe `(jotti_proxy)`-Snippet wie der LAN-Mode (gleiche
// Security-Header und CSP), setzt aber den stärkeren HSTS-Wert und das
// /api/-Rate-Limit. Caddy leitet HTTP automatisch auf HTTPS um. Reine Funktion
// ohne I/O.
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

// wildcardSite rendert die vertrauenswürdige Site `*.<install-id>.<zone>` mit
// DNS-01-Challenge über acme-dns. Caddy holt und erneuert das Zertifikat
// automatisch im Hintergrund; bis dahin (oder offline) trägt die Fallback-Site.
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
