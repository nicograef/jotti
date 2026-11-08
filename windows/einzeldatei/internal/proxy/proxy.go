// Package proxy betreibt Caddy als Bibliothek im Prozess, mit denselben Plugins wie das Proxy-Image.
package proxy

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	_ "github.com/caddy-dns/acmedns"
	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig"
	_ "github.com/caddyserver/caddy/v2/caddyconfig/caddyfile" // Caddyfile-Adapter
	_ "github.com/caddyserver/caddy/v2/modules/standard"
	_ "github.com/mholt/caddy-ratelimit"
)

// Run adaptiert caddyfile, lädt die Konfiguration und hält Caddy, bis ctx endet.
func Run(ctx context.Context, caddyfile string) error {
	cfgJSON, err := adapt(caddyfile)
	if err != nil {
		return err
	}
	if err := caddy.Load(cfgJSON, true); err != nil {
		return fmt.Errorf("caddy starten: %w", err)
	}
	<-ctx.Done()
	if err := caddy.Stop(); err != nil {
		return fmt.Errorf("caddy stoppen: %w", err)
	}
	return nil
}

// adapt wandelt den Caddyfile in JSON. Die Einzeldatei speichert keine Autosave-Kopie der
// Konfiguration im Benutzerprofil; sie erzeugt sie bei jedem Start neu.
func adapt(caddyfile string) ([]byte, error) {
	adapter := caddyconfig.GetAdapter("caddyfile")
	if adapter == nil {
		return nil, fmt.Errorf("kein Caddyfile-Adapter registriert")
	}
	raw, warnings, err := adapter.Adapt([]byte(caddyfile), map[string]any{"filename": "Caddyfile"})
	if err != nil {
		return nil, fmt.Errorf("caddyfile adaptieren: %w", err)
	}
	for _, w := range warnings {
		slog.Warn("Caddyfile", "warnung", w.String())
	}

	var cfg caddy.Config
	if err := json.Unmarshal(raw, &cfg); err != nil {
		return nil, err
	}
	if cfg.Admin == nil {
		cfg.Admin = &caddy.AdminConfig{}
	}
	if cfg.Admin.Config == nil {
		cfg.Admin.Config = &caddy.ConfigSettings{}
	}
	persist := false
	cfg.Admin.Config.Persist = &persist
	return json.Marshal(cfg)
}
