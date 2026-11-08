// Package caddyconf renders the Caddyfiles of the jotti proxy: LAN, public and HTTP-only.
// The container passes the Container* values; the Windows single file passes its own upstream and paths.
package caddyconf

import (
	"fmt"
	"strings"
)

// Container values: the backend service in Compose and the app build in the proxy image.
const (
	ContainerUpstream = "backend:3000"
	ContainerAppRoot  = "/srv/app"
)

// contentSecurityPolicy applies to every generated site via proxySnippet and, verbatim,
// to the demo site in Caddyfile.rocks. TestRocksCaddyfileCarriesSameCSP keeps both copies equal.
const contentSecurityPolicy = "default-src 'none'; base-uri 'self'; frame-ancestors 'none'; form-action 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data: blob:; font-src 'self'; connect-src 'self'; manifest-src 'self'; worker-src 'self' blob:; media-src 'self'; frame-src 'none'; object-src 'none'; upgrade-insecure-requests"

// hstsLAN carries max-age only, because clients also reach the fallback site by raw LAN IP.
// hstsPublic is the stronger value for a public domain, matching Caddyfile.rocks.
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

// appHandler serves the app's static build from root, the last handler of every generated
// site and, with ContainerAppRoot, verbatim of the demo site in Caddyfile.rocks;
// TestRocksCaddyfileServesSameApp keeps both copies equal. Response headers are set only for
// files that exist, so a 404 carries none.
func appHandler(root string) string {
	return fmt.Sprintf(appHandlerTemplate, caddyPath(root))
}

const appHandlerTemplate = `	# App: statischer Vite-Build aus dem Image
	handle {
		root * %s
		encode zstd gzip

		# Jeder Dateiname unter /assets/ trägt einen Inhalts-Hash: für immer cachen.
		handle /assets/* {
			@asset_file file
			header @asset_file Cache-Control "public, max-age=31536000, immutable"
			file_server
		}

		# Caddy kennt den Medientyp des PWA-Manifests nicht.
		handle /manifest.webmanifest {
			@manifest_file file
			header @manifest_file {
				Cache-Control "public, max-age=3600"
				Content-Type application/manifest+json
			}
			file_server
		}

		# Ungehashte Dateien aus public/ (Favicons): kurz cachen.
		@public_static path_regexp (?i)\.(ico|png|svg|webp|woff2?)$
		handle @public_static {
			@public_file file
			header @public_file Cache-Control "public, max-age=3600"
			file_server
		}

		# SPA-Routing: jede andere Route bekommt index.html, nie gecacht, damit
		# Clients nach einem Update sofort die neuen Bundles laden.
		handle {
			header Cache-Control "no-cache"
			try_files {path} /index.html
			file_server {
				disable_canonical_uris
			}
		}
	}`

// LANInput beschreibt den lokalen Stack. Username, Password und Subdomain sind die
// acme-dns-Zugangsdaten der Wildcard-Site und zählen nur mit HasState.
type LANInput struct {
	HasState    bool   // false ⇒ nur die Fallback-Site rendern
	Username    string // acme-dns-Benutzer
	Password    string // acme-dns-Passwort
	Subdomain   string // acme-dns-Subdomain, zugleich die Install-ID
	Zone        string // z. B. "lokal.jotti.rocks"
	ACMEDNSURL  string // acme-dns-API, z. B. "https://auth.jotti.rocks"
	LEStaging   bool   // true ⇒ Zertifikate über die LE-Staging-CA holen
	Upstream    string // host:port des Backends, im Container ContainerUpstream
	AppRoot     string // Verzeichnis des App-Builds, im Container ContainerAppRoot
	StorageRoot string // Caddys Speicherort für Zertifikate und CA; leer ⇒ Caddys Vorgabe
}

// RenderLAN erzeugt den Caddyfile des lokalen Stacks: die Wildcard-Site (nur
// mit acme-dns-Credentials) und immer die Fallback-Site mit Caddys interner CA.
func RenderLAN(in LANInput) string {
	wildcard := ""
	if in.HasState {
		wildcard = wildcardSite(in) + "\n\n"
	}

	storage := ""
	if in.StorageRoot != "" {
		storage = "\n\tstorage file_system " + caddyPath(in.StorageRoot)
	}

	return fmt.Sprintf(`# Generiert vom jotti-reverse-proxy beim Start — nicht von Hand bearbeiten.
{
	admin off
	# Installing the local root into the container's trust store is pointless and fails on the read-only root filesystem.
	skip_install_trust%s
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
`, storage, proxySnippet(hstsLAN, false, in.Upstream, in.AppRoot), wildcard)
}

// proxySnippet renders the shared `(jotti_proxy)` snippet, so headers and CSP match in every mode.
// The rate limit models 10 r/s with burst 20: caddy-ratelimit has a sliding window without a
// separate burst, so the peak served at once is `events 30` per `window 1s`.
func proxySnippet(hsts string, rateLimited bool, upstream, appRoot string) string {
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
		reverse_proxy %s
	}

%s
}`, hsts, contentSecurityPolicy, rateLimit, upstream, appHandler(appRoot))
}

// PublicInput beschreibt den Self-Hoster-prod-Stack.
type PublicInput struct {
	Domain      string // öffentliche Domain, z. B. "jotti.meinverein.de"
	Email       string // Kontakt-E-Mail für den ACME-Account
	WWWRedirect bool   // true ⇒ www.<domain> dauerhaft auf <domain> umleiten
	LEStaging   bool   // true ⇒ Zertifikate über die LE-Staging-CA holen (Tests)
}

// RenderPublic erzeugt den Caddyfile des Self-Hoster-prod-Stacks: dasselbe
// `(jotti_proxy)`-Snippet wie der LAN-Mode, aber stärkeres HSTS und das
// /api/-Rate-Limit.
func RenderPublic(in PublicInput) string {
	staging := ""
	if in.LEStaging {
		staging = "\n\tacme_ca " + leStagingCA
	}

	www := ""
	if in.WWWRedirect {
		www = fmt.Sprintf("\n\nwww.%s {\n\tredir https://%s{uri} permanent\n}", in.Domain, in.Domain)
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
`, in.Email, staging, proxySnippet(hstsPublic, true, ContainerUpstream, ContainerAppRoot), in.Domain, www)
}

// RenderHTTPOnly bedient ausschließlich Klartext-HTTP auf :80 — nur für
// die E2E-Testumgebung (docker-compose.e2e.yml, PROXY_HTTP_ONLY=1), wo der Stack
// lokal bleibt. Das /api/-Rate-Limit bleibt aus, damit Test-Suiten nicht gedrosselt
// werden.
func RenderHTTPOnly() string {
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
`, proxySnippet(hstsLAN, false, ContainerUpstream, ContainerAppRoot))
}

// wildcardSite rendert `*.<install-id>.<zone>` mit DNS-01-Challenge über acme-dns;
// bis Caddy ausgestellt hat (oder offline) trägt die Fallback-Site.
func wildcardSite(in LANInput) string {
	caBlock := ""
	if in.LEStaging {
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
}`, in.Subdomain, in.Zone, in.Username, in.Password, in.Subdomain, in.ACMEDNSURL, challengeResolvers, caBlock)
}

// caddyPath renders a filesystem path as one Caddyfile token: forward slashes, which Caddy
// accepts on Windows too, and quotes around a path with whitespace (C:/Program Files).
func caddyPath(p string) string {
	p = strings.ReplaceAll(p, `\`, "/")
	if strings.ContainsAny(p, " \t") {
		return `"` + p + `"`
	}
	return p
}
