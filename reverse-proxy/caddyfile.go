package main

import "fmt"

// contentSecurityPolicy ist die CSP des lokalen Stacks — unverändert aus dem
// früheren nginx-/Caddy-Setup übernommen (Parität).
const contentSecurityPolicy = "default-src 'none'; base-uri 'self'; frame-ancestors 'none'; form-action 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data: blob:; font-src 'self'; connect-src 'self'; manifest-src 'self'; worker-src 'self' blob:; media-src 'self'; frame-src 'none'; object-src 'none'; upgrade-insecure-requests"

// leStagingCA ist das ACME-Verzeichnis der Let's-Encrypt-Staging-Umgebung.
// Entwicklung und Tests nutzen es, um die Rate-Limits der echten Zone zu
// schonen (Schalter PROXY_LE_STAGING).
const leStagingCA = "https://acme-staging-v02.api.letsencrypt.org/directory"

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

	return fmt.Sprintf(`# Generiert vom jotti-local-proxy beim Start — nicht von Hand bearbeiten.
{
	admin off
}

# Gemeinsame Proxy- und Security-Header-Konfiguration für beide Sites.
(jotti_proxy) {
	header {
		Strict-Transport-Security "max-age=31536000"
		X-Content-Type-Options "nosniff"
		X-Frame-Options "DENY"
		Referrer-Policy "no-referrer-when-downgrade"
		Permissions-Policy "geolocation=(), microphone=()"
		Content-Security-Policy "%s"
	}

	# Backend-API unter /api/ (Prefix beim Proxyen entfernen)
	handle_path /api/* {
		reverse_proxy backend:3000
	}

	# Frontend-SPA
	handle {
		reverse_proxy frontend:80
	}
}

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
`, contentSecurityPolicy, wildcard)
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
		}%s
	}
	import jotti_proxy
}`, in.state.Subdomain, in.zone, in.state.Username, in.state.Password, in.state.Subdomain, in.acmeDNSURL, caBlock)
}
