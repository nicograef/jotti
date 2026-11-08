package caddyconf

import (
	"os"
	"strings"
	"testing"
)

func withState() LANInput {
	return LANInput{
		HasState:   true,
		Username:   "user-1",
		Password:   "pass-1",
		Subdomain:  "sub-1",
		Zone:       "lokal.jotti.rocks",
		ACMEDNSURL: "https://auth.jotti.rocks",
		Upstream:   ContainerUpstream,
		AppRoot:    ContainerAppRoot,
	}
}

func TestRenderCaddyfileWithState(t *testing.T) {
	out := RenderLAN(withState())

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
		appHandler(ContainerAppRoot),
		"https:// {",
		"on_demand",
		"skip_install_trust",
		"redir https://{host}{uri} permanent",
		`Strict-Transport-Security "` + hstsLAN + `"`,
		`X-Content-Type-Options "nosniff"`,
		`X-Frame-Options "DENY"`,
		`Content-Security-Policy "` + contentSecurityPolicy + `"`,
	}
	for _, want := range wants {
		if !strings.Contains(out, want) {
			t.Errorf("Caddyfile enthält %q nicht\n---\n%s", want, out)
		}
	}
	if strings.Contains(out, leStagingCA) {
		t.Errorf("ohne Staging-Schalter darf die Staging-CA nicht vorkommen")
	}
	if strings.Contains(out, "rate_limit") {
		t.Errorf("LAN-Mode darf kein Rate-Limit rendern\n---\n%s", out)
	}
	if strings.Contains(out, hstsPublic) {
		t.Errorf("LAN-Mode darf den Public-HSTS-Wert nicht setzen")
	}
}

func TestRenderCaddyfileWithoutStateOmitsWildcard(t *testing.T) {
	in := withState()
	in.HasState = false
	out := RenderLAN(in)

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
	in.LEStaging = true

	out := RenderLAN(in)
	if !strings.Contains(out, "ca "+leStagingCA) {
		t.Errorf("Staging-CA-Direktive fehlt\n---\n%s", out)
	}
}

func publicInputFixture() PublicInput {
	return PublicInput{
		Domain: "jotti.meinverein.de",
		Email:  "verein@example.de",
	}
}

func TestRenderPublicCaddyfile(t *testing.T) {
	out := RenderPublic(publicInputFixture())

	wants := []string{
		"email verein@example.de",
		"(jotti_proxy) {",
		"jotti.meinverein.de {",
		"import jotti_proxy",
		"handle_path /api/* {",
		"reverse_proxy backend:3000",
		appHandler(ContainerAppRoot),
		`Strict-Transport-Security "` + hstsPublic + `"`,
		`X-Content-Type-Options "nosniff"`,
		`X-Frame-Options "DENY"`,
		`Content-Security-Policy "` + contentSecurityPolicy + `"`,
		"rate_limit {",
		"zone api {",
		"key {remote_host}",
		"events 30",
		"window 1s",
	}
	for _, want := range wants {
		if !strings.Contains(out, want) {
			t.Errorf("Public-Caddyfile enthält %q nicht\n---\n%s", want, out)
		}
	}

	for _, unwanted := range []string{"acmedns", "on_demand", "https:// {", leStagingCA} {
		if strings.Contains(out, unwanted) {
			t.Errorf("Public-Caddyfile darf %q nicht enthalten\n---\n%s", unwanted, out)
		}
	}
}

func TestRenderPublicCaddyfileWWWRedirect(t *testing.T) {
	in := publicInputFixture()
	in.WWWRedirect = true

	out := RenderPublic(in)
	if !strings.Contains(out, "www.jotti.meinverein.de {") {
		t.Errorf("www-Redirect-Site fehlt\n---\n%s", out)
	}
	if !strings.Contains(out, "redir https://jotti.meinverein.de{uri} permanent") {
		t.Errorf("www→apex-Redirect fehlt\n---\n%s", out)
	}

	if strings.Contains(RenderPublic(publicInputFixture()), "www.") {
		t.Errorf("ohne wwwRedirect darf keine www-Site gerendert werden")
	}
}

func TestRenderPublicCaddyfileStagingAddsStagingCA(t *testing.T) {
	in := publicInputFixture()
	in.LEStaging = true

	out := RenderPublic(in)
	if !strings.Contains(out, "acme_ca "+leStagingCA) {
		t.Errorf("globale Staging-CA-Direktive fehlt\n---\n%s", out)
	}
}

func TestRenderHTTPOnlyCaddyfile(t *testing.T) {
	out := RenderHTTPOnly()

	wants := []string{
		"auto_https off",
		"(jotti_proxy) {",
		"http:// {",
		"import jotti_proxy",
		"handle_path /api/* {",
		"reverse_proxy backend:3000",
		appHandler(ContainerAppRoot),
		`Strict-Transport-Security "` + hstsLAN + `"`,
		`X-Content-Type-Options "nosniff"`,
		`X-Frame-Options "DENY"`,
		`Content-Security-Policy "` + contentSecurityPolicy + `"`,
	}
	for _, want := range wants {
		if !strings.Contains(out, want) {
			t.Errorf("HTTP-Only-Caddyfile enthält %q nicht\n---\n%s", want, out)
		}
	}

	for _, unwanted := range []string{"acmedns", "on_demand", "https:// {", "tls {", "rate_limit {"} {
		if strings.Contains(out, unwanted) {
			t.Errorf("HTTP-Only-Caddyfile darf %q nicht enthalten\n---\n%s", unwanted, out)
		}
	}
}

// TestRenderLAN_WindowsWerte renders the values of the Windows single file: loopback upstream,
// backslash paths with a space and an own storage root. Mapped back, the output equals the container's.
func TestRenderLAN_WindowsWerte(t *testing.T) {
	container := withState()
	container.HasState = false
	windows := container
	windows.Upstream = "127.0.0.1:13000"
	windows.AppRoot = `C:\Program Files\jotti\v0.0.1-ci\app`
	windows.StorageRoot = `C:\ProgramData\jotti\caddy`

	out := RenderLAN(windows)

	wants := []string{
		"\treverse_proxy 127.0.0.1:13000\n",
		"\troot * \"C:/Program Files/jotti/v0.0.1-ci/app\"\n",
		"\tskip_install_trust\n\tstorage file_system C:/ProgramData/jotti/caddy\n}",
	}
	for _, want := range wants {
		if !strings.Contains(out, want) {
			t.Errorf("Caddyfile enthält %q nicht\n---\n%s", want, out)
		}
	}

	back := strings.NewReplacer(
		"reverse_proxy 127.0.0.1:13000", "reverse_proxy backend:3000",
		`root * "C:/Program Files/jotti/v0.0.1-ci/app"`, "root * /srv/app",
		"\n\tstorage file_system C:/ProgramData/jotti/caddy", "",
	).Replace(out)
	if want := RenderLAN(container); back != want {
		t.Errorf("Ausgabe weicht außer Upstream, App-Wurzel und Speicherort vom Container ab\n--- Windows\n%s\n--- Container\n%s", out, want)
	}
}

// TestRocksCaddyfileCarriesSameCSP keeps the two CSP copies equal: contentSecurityPolicy
// and the demo site in Caddyfile.rocks.
func TestRocksCaddyfileCarriesSameCSP(t *testing.T) {
	conf, err := os.ReadFile("../Caddyfile.rocks")
	if err != nil {
		t.Fatalf("read Caddyfile.rocks: %v", err)
	}
	if !strings.Contains(string(conf), contentSecurityPolicy) {
		t.Errorf("Caddyfile.rocks does not carry the CSP verbatim; want:\n%s", contentSecurityPolicy)
	}
}

// TestRocksCaddyfileServesSameApp keeps the demo site in Caddyfile.rocks on the same
// app handler as every generated site.
func TestRocksCaddyfileServesSameApp(t *testing.T) {
	conf, err := os.ReadFile("../Caddyfile.rocks")
	if err != nil {
		t.Fatalf("read Caddyfile.rocks: %v", err)
	}
	if !strings.Contains(string(conf), appHandler(ContainerAppRoot)) {
		t.Errorf("Caddyfile.rocks does not carry appHandler verbatim; want:\n%s", appHandler(ContainerAppRoot))
	}
}
