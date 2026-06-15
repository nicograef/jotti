package main

import (
	"strings"
	"testing"
)

func withState() caddyfileInput {
	return caddyfileInput{
		state:      InstallState{Username: "user-1", Password: "pass-1", Subdomain: "sub-1"},
		hasState:   true,
		zone:       "lokal.jotti.rocks",
		acmeDNSURL: "https://auth.jotti.rocks",
	}
}

func TestRenderCaddyfileWithState(t *testing.T) {
	out := renderCaddyfile(withState())

	wants := []string{
		"(jotti_proxy) {",
		"*.sub-1.lokal.jotti.rocks {",
		"dns acmedns {",
		`username "user-1"`,
		`password "pass-1"`,
		`subdomain "sub-1"`,
		`server_url "https://auth.jotti.rocks"`,
		"resolvers 1.1.1.1 8.8.8.8",
		"import jotti_proxy",
		"handle_path /api/* {",
		"reverse_proxy backend:3000",
		"reverse_proxy frontend:80",
		"https:// {",
		"on_demand",
		"redir https://{host}{uri} permanent",
		`Strict-Transport-Security "` + hstsLAN + `"`,
		contentSecurityPolicy,
	}
	for _, want := range wants {
		if !strings.Contains(out, want) {
			t.Errorf("Caddyfile enthält %q nicht\n---\n%s", want, out)
		}
	}
	if strings.Contains(out, leStagingCA) {
		t.Errorf("ohne Staging-Schalter darf die Staging-CA nicht vorkommen")
	}
	// Rate-Limit und der stärkere Public-HSTS gehören nur in den Public-Mode.
	if strings.Contains(out, "rate_limit") {
		t.Errorf("LAN-Mode darf kein Rate-Limit rendern\n---\n%s", out)
	}
	if strings.Contains(out, hstsPublic) {
		t.Errorf("LAN-Mode darf den Public-HSTS-Wert nicht setzen")
	}
}

func TestRenderCaddyfileWithoutStateOmitsWildcard(t *testing.T) {
	out := renderCaddyfile(caddyfileInput{
		hasState:   false,
		zone:       "lokal.jotti.rocks",
		acmeDNSURL: "https://auth.jotti.rocks",
	})

	if strings.Contains(out, "acmedns") {
		t.Errorf("ohne State darf keine acmedns-Site gerendert werden\n---\n%s", out)
	}
	if strings.Contains(out, "lokal.jotti.rocks {") {
		t.Errorf("ohne State darf keine Wildcard-Site gerendert werden")
	}
	if !strings.Contains(out, "https:// {") {
		t.Errorf("Fallback-Site fehlt")
	}
	if !strings.Contains(out, "redir https://{host}{uri} permanent") {
		t.Errorf("HTTP→HTTPS-Redirect fehlt")
	}
}

func TestRenderCaddyfileStagingAddsStagingCA(t *testing.T) {
	in := withState()
	in.leStaging = true

	out := renderCaddyfile(in)
	if !strings.Contains(out, "ca "+leStagingCA) {
		t.Errorf("Staging-CA-Direktive fehlt\n---\n%s", out)
	}
}

func publicInputFixture() publicInput {
	return publicInput{
		domain: "jotti.meinverein.de",
		email:  "verein@example.de",
	}
}

func TestRenderPublicCaddyfile(t *testing.T) {
	out := renderPublicCaddyfile(publicInputFixture())

	wants := []string{
		"email verein@example.de",
		"(jotti_proxy) {",
		"jotti.meinverein.de {",
		"import jotti_proxy",
		"handle_path /api/* {",
		"reverse_proxy backend:3000",
		"reverse_proxy frontend:80",
		`Strict-Transport-Security "` + hstsPublic + `"`,
		"rate_limit {",
		"zone api {",
		"key {remote_host}",
		"events 30",
		"window 1s",
		contentSecurityPolicy,
	}
	for _, want := range wants {
		if !strings.Contains(out, want) {
			t.Errorf("Public-Caddyfile enthält %q nicht\n---\n%s", want, out)
		}
	}

	// Public-Mode kennt weder die LAN-Fallback-Site noch acme-dns/Staging.
	for _, unwanted := range []string{"acmedns", "on_demand", "https:// {", leStagingCA} {
		if strings.Contains(out, unwanted) {
			t.Errorf("Public-Caddyfile darf %q nicht enthalten\n---\n%s", unwanted, out)
		}
	}
}

func TestRenderPublicCaddyfileWWWRedirect(t *testing.T) {
	in := publicInputFixture()
	in.wwwRedirect = true

	out := renderPublicCaddyfile(in)
	if !strings.Contains(out, "www.jotti.meinverein.de {") {
		t.Errorf("www-Redirect-Site fehlt\n---\n%s", out)
	}
	if !strings.Contains(out, "redir https://jotti.meinverein.de{uri} permanent") {
		t.Errorf("www→apex-Redirect fehlt\n---\n%s", out)
	}

	if strings.Contains(renderPublicCaddyfile(publicInputFixture()), "www.") {
		t.Errorf("ohne wwwRedirect darf keine www-Site gerendert werden")
	}
}

func TestRenderPublicCaddyfileStagingAddsStagingCA(t *testing.T) {
	in := publicInputFixture()
	in.leStaging = true

	out := renderPublicCaddyfile(in)
	if !strings.Contains(out, "acme_ca "+leStagingCA) {
		t.Errorf("globale Staging-CA-Direktive fehlt\n---\n%s", out)
	}
}
