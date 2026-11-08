// Command jotti ist die Windows-Einzeldatei: Postgres, Migrationen, Backend und Caddy mit
// der App laufen in diesem einen Prozess. Stufe 1 startet alles in der Konsole; Enter beendet
// in der Reihenfolge Caddy, Backend, Postgres.
package main

import (
	"bufio"
	"context"
	"crypto/tls"
	"crypto/x509"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/nicograef/jotti/backend/app"
	"github.com/nicograef/jotti/backend/config"
	dbpkg "github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/reverse-proxy/caddyconf"
	"github.com/nicograef/jotti/windows/einzeldatei/assets"
	"github.com/nicograef/jotti/windows/einzeldatei/internal/layout"
	"github.com/nicograef/jotti/windows/einzeldatei/internal/postgres"
	"github.com/nicograef/jotti/windows/einzeldatei/internal/proxy"
	"github.com/nicograef/jotti/windows/einzeldatei/internal/schema"
	"github.com/nicograef/jotti/windows/starter/core"
	"github.com/nicograef/jotti/windows/starter/host"
)

// version wird beim Build per -ldflags "-X main.version=vX.Y.Z" gesetzt.
var version = "dev"

// Feste Loopback-Ports unterhalb des dynamischen Portbereichs von Windows (ab 49152).
const (
	backendPort  = 13000
	postgresPort = 15432
	loopback     = "127.0.0.1"
	databaseName = "jotti"
	healthURL    = "https://127.0.0.1/api/health"
)

// healthTimeout deckt den ersten Start ab: Entpacken, initdb, Migrationen und das erste Zertifikat.
const healthTimeout = 3 * time.Minute

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	enter := enterPressed(os.Stdin)
	code := run(ctx, enter)
	stop()
	if code != 0 {
		// Ein per Doppelklick geöffnetes Fenster schließt beim Exit; die Meldung bleibt bis Enter lesbar.
		fmt.Println("\nEnter schliesst dieses Fenster.")
		select {
		case <-enter:
		case <-ctx.Done():
		}
	}
	os.Exit(code)
}

// enterPressed meldet jede Zeile auf stdin; ein geschlossenes stdin zählt wie Enter.
func enterPressed(r io.Reader) <-chan struct{} {
	ch := make(chan struct{}, 1)
	go func() {
		scanner := bufio.NewScanner(r)
		for scanner.Scan() {
			ch <- struct{}{}
		}
		close(ch)
	}()
	return ch
}

func run(ctx context.Context, enter <-chan struct{}) int {
	fmt.Printf("jotti %s\n\n", version)

	programFiles, programData := os.Getenv("ProgramFiles"), os.Getenv("PROGRAMDATA")
	if programFiles == "" || programData == "" {
		fmt.Println("Diese Datei laeuft nur unter Windows (ProgramFiles und PROGRAMDATA fehlen).")
		return 1
	}
	paths := layout.NewPaths(programFiles, programData, version)

	if err := host.PrepareStateDir(paths.Data); err != nil {
		fmt.Println(err)
		return 1
	}
	docker, err := layout.DockerInstallation(paths.Data)
	if err != nil {
		fmt.Printf("Datenverzeichnis %s nicht lesbar: %v\n", paths.Data, err)
		return 1
	}
	if docker {
		fmt.Printf("Start verweigert: In %s liegt eine jotti-Installation mit Docker.\n", paths.Data)
		fmt.Println("  Ihre Daten liegen im Docker-Volume; diese Datei kann sie noch nicht uebernehmen.")
		fmt.Println("  Bitte weiter jotti-start.exe aus dem ZIP nutzen: https://github.com/nicograef/jotti/releases/latest")
		return 1
	}

	created, err := core.MaterializeEnv(paths.Env, fileExists, writePrivate)
	if err != nil {
		fmt.Printf("Konfiguration (.env) konnte nicht erstellt werden: %v\n", err)
		return 1
	}
	if created {
		fmt.Println("Konfiguration (.env) mit frischen Zugangsdaten erstellt.")
	}
	env, err := readEnv(paths.Env)
	if err != nil {
		fmt.Printf("Konfiguration (.env) nicht lesbar: %v\n", err)
		return 1
	}

	if lv := readLastVersion(paths.LastVersion); core.IsDowngrade(version, lv) {
		fmt.Printf("Start verweigert: Diese Version (%s) ist aelter als die zuletzt gestartete (%s).\n", version, lv)
		fmt.Println("  Neuere Daten lassen sich nicht auf aeltere Versionen zurueckrollen.")
		fmt.Println("  Neue Version herunterladen: https://github.com/nicograef/jotti/releases/latest")
		return 1
	}

	files, err := fs.Sub(assets.FS, "files")
	if err != nil {
		fmt.Println(err)
		return 1
	}
	fmt.Printf("Programmdateien: %s\n", paths.Program)
	if err := layout.Extract(files, paths.Program); err != nil {
		fmt.Printf("Programmdateien konnten nicht entpackt werden: %v\n", err)
		return 1
	}

	cfg := config.Config{
		ListenHost:     loopback,
		Port:           backendPort,
		JWTSecret:      env["JWT_SECRET"],
		RelayToken:     env["RELAY_AUTH_TOKEN"],
		FiskalyBaseURL: config.DefaultFiskalyBaseURL,
	}
	cfg.Postgres.Host = loopback
	cfg.Postgres.Port = postgresPort
	cfg.Postgres.User = env["POSTGRES_USER"]
	cfg.Postgres.Password = env["POSTGRES_PASSWORD"]
	cfg.Postgres.DBName = databaseName
	if err := config.ValidateSecrets(cfg); err != nil {
		fmt.Printf("Konfiguration (%s) ungueltig: %v\n", paths.Env, err)
		return 1
	}

	cluster := postgres.Cluster{Bin: paths.PostgresBin, Data: paths.PGData, Logs: paths.Logs, Port: postgresPort}
	if !cluster.Initialized() {
		fmt.Println("Datenbank wird angelegt ...")
		if err := cluster.Init(cfg.Postgres.User, cfg.Postgres.Password); err != nil {
			fmt.Printf("Datenbank konnte nicht angelegt werden: %v\n", err)
			return 1
		}
	}
	if err := cluster.Start(); err != nil {
		fmt.Printf("Datenbank startet nicht: %v\n", err)
		return 1
	}
	defer func() {
		if err := cluster.Stop(); err != nil {
			fmt.Printf("Datenbank liess sich nicht sauber stoppen: %v\n", err)
		}
	}()

	db, err := prepareDatabase(ctx, cfg, files)
	if err != nil {
		fmt.Printf("Datenbank nicht bereit: %v\n", err)
		return 1
	}
	defer db.Close() //nolint:errcheck // Postgres stoppt direkt danach

	return serve(ctx, enter, cfg, db, paths)
}

// prepareDatabase legt die Datenbank an, migriert sie und öffnet den Pool des Backends.
func prepareDatabase(ctx context.Context, cfg config.Config, files fs.FS) (*sql.DB, error) {
	pg := cfg.Postgres
	port := strconv.Itoa(pg.Port)
	if err := schema.EnsureDatabase(ctx, dbpkg.ConnString(pg.Host, port, pg.User, pg.Password, "postgres"), pg.DBName); err != nil {
		return nil, err
	}
	migrations, err := fs.Sub(files, "migrations")
	if err != nil {
		return nil, err
	}
	dsn := url.URL{
		Scheme:   "pgx5",
		User:     url.UserPassword(pg.User, pg.Password),
		Host:     net.JoinHostPort(pg.Host, port),
		Path:     "/" + pg.DBName,
		RawQuery: "sslmode=disable",
	}
	if err := schema.Migrate(ctx, dsn.String(), migrations); err != nil {
		return nil, fmt.Errorf("migrationen: %w", err)
	}
	return dbpkg.Open(ctx, dbpkg.ConnString(pg.Host, port, pg.User, pg.Password, pg.DBName))
}

// serve startet Backend und Caddy, wartet auf Enter und beendet Caddy vor dem Backend.
func serve(ctx context.Context, enter <-chan struct{}, cfg config.Config, db *sql.DB, paths layout.Paths) int {
	backendCtx, stopBackend := context.WithCancel(context.Background())
	defer stopBackend()
	backendDone := make(chan error, 1)
	go func() { backendDone <- app.Serve(backendCtx, cfg, db, version) }()

	caddyfile := caddyconf.RenderLAN(caddyconf.LANInput{
		Upstream:    net.JoinHostPort(loopback, strconv.Itoa(backendPort)),
		AppRoot:     paths.App,
		StorageRoot: paths.Caddy,
	})
	proxyCtx, stopProxy := context.WithCancel(context.Background())
	defer stopProxy()
	proxyDone := make(chan error, 1)
	go func() { proxyDone <- proxy.Run(proxyCtx, caddyfile) }()

	code := 0
	healthCtx, stopHealth := context.WithCancel(ctx)
	defer stopHealth()
	healthy := make(chan error, 1)
	go func() { healthy <- waitHealthy(healthCtx, paths.Caddy) }()
	select {
	case err := <-healthy:
		if err != nil {
			fmt.Printf("jotti antwortet nicht: %v\n", err)
			code = 1
			break
		}
		if err := writePrivate(paths.LastVersion, []byte(version+"\n")); err != nil {
			fmt.Printf("Hinweis: Versionsmarker konnte nicht geschrieben werden (%v).\n", err)
		}
		fmt.Println("\njotti laeuft: https://127.0.0.1")
		fmt.Println("Enter beendet jotti.")
		code = waitForEnd(ctx, enter, backendDone, proxyDone)
	case err := <-backendDone:
		fmt.Printf("Backend beendet: %v\n", err)
		backendDone <- err
		code = 1
	case err := <-proxyDone:
		fmt.Printf("Caddy beendet: %v\n", err)
		proxyDone <- err
		code = 1
	}

	fmt.Println("jotti wird beendet ...")
	stopProxy()
	if err := <-proxyDone; err != nil {
		fmt.Printf("Caddy: %v\n", err)
		code = 1
	}
	stopBackend()
	if err := <-backendDone; err != nil {
		fmt.Printf("Backend: %v\n", err)
		code = 1
	}
	return code
}

// waitForEnd wartet auf Enter oder Strg+C; endet vorher Backend oder Caddy, ist das ein Fehler.
func waitForEnd(ctx context.Context, enter <-chan struct{}, backendDone, proxyDone chan error) int {
	select {
	case <-enter:
		return 0
	case <-ctx.Done():
		return 0
	case err := <-backendDone:
		fmt.Printf("Backend unerwartet beendet: %v\n", err)
		backendDone <- err
		return 1
	case err := <-proxyDone:
		fmt.Printf("Caddy unerwartet beendet: %v\n", err)
		proxyDone <- err
		return 1
	}
}

// waitHealthy fragt /api/health über Caddy ab und vertraut dabei nur Caddys lokaler CA.
func waitHealthy(ctx context.Context, caddyStorage string) error {
	ctx, cancel := context.WithTimeout(ctx, healthTimeout)
	defer cancel()
	rootCert := filepath.Join(caddyStorage, "pki", "authorities", "local", "root.crt")
	var lastErr error
	for {
		if lastErr = probeHealth(ctx, rootCert); lastErr == nil {
			return nil
		}
		select {
		case <-ctx.Done():
			return fmt.Errorf("%s nach %s: %w", healthURL, healthTimeout, lastErr)
		case <-time.After(time.Second):
		}
	}
}

func probeHealth(ctx context.Context, rootCert string) error {
	pem, err := os.ReadFile(rootCert) //nolint:gosec // G703: Pfad aus %PROGRAMDATA%, nicht aus Eingaben
	if err != nil {
		return err
	}
	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(pem) {
		return fmt.Errorf("%s enthaelt kein Zertifikat", rootCert)
	}
	client := &http.Client{
		Timeout:   5 * time.Second,
		Transport: &http.Transport{TLSClientConfig: &tls.Config{RootCAs: pool, MinVersion: tls.VersionTLS12}},
	}
	defer client.CloseIdleConnections()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, healthURL, nil)
	if err != nil {
		return err
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close() //nolint:errcheck // nur der Status zählt
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("status %d", resp.StatusCode)
	}
	return nil
}

// readEnv liest KEY=VALUE-Zeilen; Kommentare, Leerzeilen und ein UTF-8-BOM zählen nicht.
func readEnv(path string) (map[string]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	env := map[string]string{}
	for line := range strings.Lines(strings.TrimPrefix(string(data), "\xef\xbb\xbf")) {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		env[strings.TrimSpace(key)] = strings.TrimSpace(value)
	}
	return env, nil
}

func readLastVersion(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

func writePrivate(path string, data []byte) error {
	return os.WriteFile(path, data, 0o600)
}

func fileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	return false, err
}
