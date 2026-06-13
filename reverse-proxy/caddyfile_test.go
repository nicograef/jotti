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
