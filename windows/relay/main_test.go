package main

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestPruefeRelayStatus(t *testing.T) {
	tests := []struct {
		name        string
		statusCode  int
		body        string
		wantErr     bool
		errContains string
	}{
		{
			name:       "200 ist kein Fehler",
			statusCode: http.StatusOK,
			body:       `{"auftraege":[]}`,
		},
		{
			name:        "400 mit code unauthorized ergibt Token-Hinweis",
			statusCode:  http.StatusBadRequest,
			body:        `{"code":"unauthorized"}`,
			wantErr:     true,
			errContains: "Token",
		},
		{
			name:        "anderer Fehlercode meldet nur den HTTP-Status",
			statusCode:  http.StatusBadRequest,
			body:        `{"code":"invalid_json"}`,
			wantErr:     true,
			errContains: "400",
		},
		{
			name:        "Serverfehler ohne Body meldet nur den HTTP-Status",
			statusCode:  http.StatusInternalServerError,
			body:        "",
			wantErr:     true,
			errContains: "500",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := &http.Response{
				StatusCode: tt.statusCode,
				Body:       io.NopCloser(strings.NewReader(tt.body)),
			}

			err := pruefeRelayStatus(resp)

			if !tt.wantErr {
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !strings.Contains(err.Error(), tt.errContains) {
				t.Fatalf("expected error to contain %q, got %q", tt.errContains, err.Error())
			}
		})
	}
}

func TestLoadConfigFromEnv(t *testing.T) {
	tests := []struct {
		name        string
		env         map[string]string
		wantConfig  RelayConfig
		wantErr     bool
		errContains string
	}{
		{
			name: "uses defaults when optional values are missing",
			env: map[string]string{
				"RELAY_AUTH_TOKEN": "devrelay",
			},
			wantConfig: RelayConfig{
				Token:         "devrelay",
				BackendURL:    defaultBackendURL,
				PollSeconds:   defaultPollSeconds,
				TLSSkipVerify: true,
			},
		},
		{
			name: "uses overridden backend url and poll interval",
			env: map[string]string{
				"RELAY_AUTH_TOKEN":   "devrelay",
				"RELAY_BACKEND_URL":  "https://example.org/api",
				"RELAY_POLL_SECONDS": "5",
			},
			wantConfig: RelayConfig{
				Token:         "devrelay",
				BackendURL:    "https://example.org/api",
				PollSeconds:   5,
				TLSSkipVerify: false,
			},
		},
		{
			name: "trims trailing slash from backend url",
			env: map[string]string{
				"RELAY_AUTH_TOKEN":  "devrelay",
				"RELAY_BACKEND_URL": "https://example.org/api/",
			},
			wantConfig: RelayConfig{
				Token:         "devrelay",
				BackendURL:    "https://example.org/api",
				PollSeconds:   defaultPollSeconds,
				TLSSkipVerify: false,
			},
		},
		{
			name: "accepts explicit TLS skip true",
			env: map[string]string{
				"RELAY_AUTH_TOKEN":      "devrelay",
				"RELAY_BACKEND_URL":     "https://example.org/api",
				"RELAY_TLS_SKIP_VERIFY": "true",
			},
			wantConfig: RelayConfig{
				Token:         "devrelay",
				BackendURL:    "https://example.org/api",
				PollSeconds:   defaultPollSeconds,
				TLSSkipVerify: true,
			},
		},
		{
			name: "accepts explicit TLS skip false on localhost",
			env: map[string]string{
				"RELAY_AUTH_TOKEN":      "devrelay",
				"RELAY_TLS_SKIP_VERIFY": "0",
			},
			wantConfig: RelayConfig{
				Token:         "devrelay",
				BackendURL:    defaultBackendURL,
				PollSeconds:   defaultPollSeconds,
				TLSSkipVerify: false,
			},
		},
		{
			name: "fails on invalid TLS skip value",
			env: map[string]string{
				"RELAY_AUTH_TOKEN":      "devrelay",
				"RELAY_TLS_SKIP_VERIFY": "ja",
			},
			wantErr:     true,
			errContains: "RELAY_TLS_SKIP_VERIFY",
		},
		{
			name: "fails when token is missing",
			env: map[string]string{
				"RELAY_BACKEND_URL": "https://example.org/api",
			},
			wantErr:     true,
			errContains: "RELAY_AUTH_TOKEN",
		},
		{
			name: "fails when poll is not a positive integer",
			env: map[string]string{
				"RELAY_AUTH_TOKEN":   "devrelay",
				"RELAY_POLL_SECONDS": "0",
			},
			wantErr:     true,
			errContains: "RELAY_POLL_SECONDS",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := loadConfigFromEnv(func(key string) string {
				return tt.env[key]
			})

			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Fatalf("expected error to contain %q, got %q", tt.errContains, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
			if config.Token != tt.wantConfig.Token {
				t.Fatalf("token mismatch: got %q, want %q", config.Token, tt.wantConfig.Token)
			}
			if config.BackendURL != tt.wantConfig.BackendURL {
				t.Fatalf("backend url mismatch: got %q, want %q", config.BackendURL, tt.wantConfig.BackendURL)
			}
			if config.PollSeconds != tt.wantConfig.PollSeconds {
				t.Fatalf("poll seconds mismatch: got %d, want %d", config.PollSeconds, tt.wantConfig.PollSeconds)
			}
			if config.TLSSkipVerify != tt.wantConfig.TLSSkipVerify {
				t.Fatalf("tls skip verify mismatch: got %t, want %t", config.TLSSkipVerify, tt.wantConfig.TLSSkipVerify)
			}
		})
	}
}

// fakeGruppenDrucker ist eine injizierbare Gruppen-Zustellung: sie protokolliert
// je Ziel-IP die zugestellten Gruppen und lässt für vorgegebene IPs die gesamte
// Gruppe scheitern (ein Fehlversuch je Auftrag, nichts zugestellt).
type fakeGruppenDrucker struct {
	mu          sync.Mutex
	gruppenByIP map[string][][]int
	fehlerByIP  map[string]string
}

func newFakeGruppenDrucker(fehlerByIP map[string]string) *fakeGruppenDrucker {
	return &fakeGruppenDrucker{
		gruppenByIP: make(map[string][][]int),
		fehlerByIP:  fehlerByIP,
	}
}

func (f *fakeGruppenDrucker) zustelle(zielIP string, auftraege []DruckAuftrag) ([]int, []fehlversuch) {
	f.mu.Lock()
	defer f.mu.Unlock()

	ids := auftragsIDs(auftraege)
	f.gruppenByIP[zielIP] = append(f.gruppenByIP[zielIP], ids)
	if meldung, ok := f.fehlerByIP[zielIP]; ok {
		return nil, gruppenFehlversuche(auftraege, errors.New(meldung))
	}
	return ids, nil
}

func TestVerarbeiteZyklusGruppiertUndHaeltReihenfolge(t *testing.T) {
	auftraege := []DruckAuftrag{
		{ID: 1, ZielIP: "10.0.0.1"},
		{ID: 2, ZielIP: "10.0.0.2"},
		{ID: 3, ZielIP: "10.0.0.1"},
		{ID: 4, ZielIP: "10.0.0.2"},
		{ID: 5, ZielIP: "10.0.0.1"},
	}
	drucker := newFakeGruppenDrucker(nil)

	ergebnis := verarbeiteZyklus(auftraege, drucker.zustelle)

	if want := []int{1, 2, 3, 4, 5}; !reflect.DeepEqual(ergebnis.gedruckteIDs, want) {
		t.Fatalf("gedruckteIDs: got %v, want %v", ergebnis.gedruckteIDs, want)
	}
	if len(ergebnis.fehlversuche) != 0 {
		t.Fatalf("fehlversuche: got %v, want none", ergebnis.fehlversuche)
	}
	// Je Ziel-IP genau eine Zustellung, mit allen Aufträgen in ID-Reihenfolge.
	if want := [][]int{{1, 3, 5}}; !reflect.DeepEqual(drucker.gruppenByIP["10.0.0.1"], want) {
		t.Fatalf("Gruppen IP1: got %v, want %v", drucker.gruppenByIP["10.0.0.1"], want)
	}
	if want := [][]int{{2, 4}}; !reflect.DeepEqual(drucker.gruppenByIP["10.0.0.2"], want) {
		t.Fatalf("Gruppen IP2: got %v, want %v", drucker.gruppenByIP["10.0.0.2"], want)
	}
}

// barriereDrucker blockiert jede Gruppen-Zustellung, bis alle erwarteten Gruppen
// gestartet sind. Werden die Gruppen sequenziell verarbeitet, erreicht keine die
// Barriere rechtzeitig und der Test schlägt fehl, statt dauerhaft zu blockieren.
type barriereDrucker struct {
	mu         sync.Mutex
	ausstehend int
	verpasst   int
	alleDa     chan struct{}
}

func neuerBarriereDrucker(gruppen int) *barriereDrucker {
	return &barriereDrucker{ausstehend: gruppen, alleDa: make(chan struct{})}
}

func (b *barriereDrucker) zustelle(_ string, auftraege []DruckAuftrag) ([]int, []fehlversuch) {
	b.mu.Lock()
	b.ausstehend--
	if b.ausstehend == 0 {
		close(b.alleDa)
	}
	b.mu.Unlock()

	select {
	case <-b.alleDa:
	case <-time.After(2 * time.Second):
		b.mu.Lock()
		b.verpasst++
		b.mu.Unlock()
	}
	return auftragsIDs(auftraege), nil
}

func (b *barriereDrucker) verpassteBarrieren() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.verpasst
}

func TestVerarbeiteZyklusStelltGruppenParallelZu(t *testing.T) {
	// Eine Gruppe wartet bis zu quittungBasis + n*quittungProBon auf ihre
	// Quittung; liefen die Gruppen nacheinander, blockierte ein stummer Drucker
	// alle anderen Stationen für diese Spanne.
	auftraege := []DruckAuftrag{
		{ID: 1, ZielIP: "10.0.0.1"},
		{ID: 2, ZielIP: "10.0.0.2"},
		{ID: 3, ZielIP: "10.0.0.3"},
	}
	drucker := neuerBarriereDrucker(3)

	ergebnis := verarbeiteZyklus(auftraege, drucker.zustelle)

	if got := drucker.verpassteBarrieren(); got != 0 {
		t.Fatalf("%d Gruppen liefen nicht parallel", got)
	}
	if want := []int{1, 2, 3}; !reflect.DeepEqual(ergebnis.gedruckteIDs, want) {
		t.Fatalf("gedruckteIDs: got %v, want %v", ergebnis.gedruckteIDs, want)
	}
}

func TestVerarbeiteZyklusFehlerEinerGruppeBetrifftDieAnderenNicht(t *testing.T) {
	auftraege := []DruckAuftrag{
		{ID: 1, ZielIP: "10.0.0.1"},
		{ID: 2, ZielIP: "10.0.0.1"},
		{ID: 3, ZielIP: "10.0.0.1"},
		{ID: 4, ZielIP: "10.0.0.2"},
		{ID: 5, ZielIP: "10.0.0.2"},
	}
	drucker := newFakeGruppenDrucker(map[string]string{"10.0.0.1": "nicht erreichbar"})

	ergebnis := verarbeiteZyklus(auftraege, drucker.zustelle)

	// Der tote Drucker blockiert IP2 nicht.
	if want := []int{4, 5}; !reflect.DeepEqual(ergebnis.gedruckteIDs, want) {
		t.Fatalf("gedruckteIDs: got %v, want %v", ergebnis.gedruckteIDs, want)
	}
	// Ein Fehlversuch je Auftrag der gescheiterten Gruppe von IP1.
	wantFehlversuche := []fehlversuch{
		{ID: 1, Fehler: "nicht erreichbar"},
		{ID: 2, Fehler: "nicht erreichbar"},
		{ID: 3, Fehler: "nicht erreichbar"},
	}
	if !reflect.DeepEqual(ergebnis.fehlversuche, wantFehlversuche) {
		t.Fatalf("fehlversuche: got %v, want %v", ergebnis.fehlversuche, wantFehlversuche)
	}
	if want := [][]int{{1, 2, 3}}; !reflect.DeepEqual(drucker.gruppenByIP["10.0.0.1"], want) {
		t.Fatalf("Gruppen IP1: got %v, want %v", drucker.gruppenByIP["10.0.0.1"], want)
	}
	if want := [][]int{{4, 5}}; !reflect.DeepEqual(drucker.gruppenByIP["10.0.0.2"], want) {
		t.Fatalf("Gruppen IP2: got %v, want %v", drucker.gruppenByIP["10.0.0.2"], want)
	}
}

func TestVerarbeiteZyklusBegrenztDieBonsJeGruppe(t *testing.T) {
	// Ohne Obergrenze wächst das Quittungsfenster einer Gruppe mit der Länge der
	// Warteschlange — ein Drucker mit vielen offenen Bons ließe alle anderen
	// Stationen minutenlang warten.
	drucker := newFakeGruppenDrucker(nil)

	ergebnis := verarbeiteZyklus(testAuftraege(maxBonsProZyklus+2), drucker.zustelle)

	// Die ältesten maxBonsProZyklus Aufträge — die Eingabe kommt nach ID sortiert.
	want := auftragsIDs(testAuftraege(maxBonsProZyklus))
	if !reflect.DeepEqual(ergebnis.gedruckteIDs, want) {
		t.Fatalf("gedruckteIDs: got %v, want %v", ergebnis.gedruckteIDs, want)
	}
	// Die abgeschnittenen Aufträge bleiben offen: weder gedruckt noch gescheitert.
	if len(ergebnis.fehlversuche) != 0 {
		t.Fatalf("fehlversuche: got %v, want none", ergebnis.fehlversuche)
	}
	if got := drucker.gruppenByIP[testZielIP]; !reflect.DeepEqual(got, [][]int{want}) {
		t.Fatalf("zugestellte Gruppe: got %v, want %v", got, [][]int{want})
	}
}

func TestFuehreZyklusAusMeldetErfolgeUndFehlversuche(t *testing.T) {
	auftraege := []DruckAuftrag{
		{ID: 1, ZielIP: "10.0.0.1"},
		{ID: 2, ZielIP: "10.0.0.1"},
		{ID: 3, ZielIP: "10.0.0.2"},
		{ID: 4, ZielIP: "10.0.0.2"},
	}
	drucker := newFakeGruppenDrucker(map[string]string{"10.0.0.2": "kaputt"})

	var gemeldet []zyklusErgebnis
	melde := func(ergebnis zyklusErgebnis) error {
		gemeldet = append(gemeldet, ergebnis)
		return nil
	}

	ergebnis, err := fuehreZyklusAus(auftraege, drucker.zustelle, melde)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(gemeldet) != 1 {
		t.Fatalf("melde calls: got %d, want 1", len(gemeldet))
	}
	if !reflect.DeepEqual(gemeldet[0], ergebnis) {
		t.Fatalf("gemeldetes Ergebnis weicht ab: got %+v, want %+v", gemeldet[0], ergebnis)
	}
	if want := []int{1, 2}; !reflect.DeepEqual(ergebnis.gedruckteIDs, want) {
		t.Fatalf("gedruckteIDs: got %v, want %v", ergebnis.gedruckteIDs, want)
	}
	if want := []fehlversuch{{ID: 3, Fehler: "kaputt"}, {ID: 4, Fehler: "kaputt"}}; !reflect.DeepEqual(ergebnis.fehlversuche, want) {
		t.Fatalf("fehlversuche: got %v, want %v", ergebnis.fehlversuche, want)
	}
}

func TestFuehreZyklusAusOhneAuftraegeMeldetNicht(t *testing.T) {
	drucker := newFakeGruppenDrucker(nil)

	called := false
	melde := func(zyklusErgebnis) error {
		called = true
		return nil
	}

	ergebnis, err := fuehreZyklusAus(nil, drucker.zustelle, melde)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if called {
		t.Fatalf("melde wurde ohne Aufträge aufgerufen")
	}
	if len(ergebnis.gedruckteIDs) != 0 || len(ergebnis.fehlversuche) != 0 {
		t.Fatalf("Ergebnis nicht leer: %+v", ergebnis)
	}
}

func TestFuehreZyklusAusGibtMeldefehlerWeiter(t *testing.T) {
	auftraege := []DruckAuftrag{{ID: 1, ZielIP: "10.0.0.1"}}
	drucker := newFakeGruppenDrucker(nil)

	melde := func(zyklusErgebnis) error {
		return errors.New("backend nicht erreichbar")
	}

	_, err := fuehreZyklusAus(auftraege, drucker.zustelle, melde)
	if err == nil || !strings.Contains(err.Error(), "backend nicht erreichbar") {
		t.Fatalf("expected melde error, got %v", err)
	}
}

// Gekürzte Zustell-Timeouts: die Tests sollen nicht auf echte Drucker- und
// Netzzeiten warten. In Produktion gilt produktionsTimeouts.
var testTimeouts = zustellTimeouts{
	papier:         20 * time.Millisecond,
	spuelen:        40 * time.Millisecond,
	quittungBasis:  50 * time.Millisecond,
	quittungProBon: 10 * time.Millisecond,
}

const testZielIP = "10.0.0.1"

// papierOK ist die Papierstatus-Antwort eines Druckers mit Papier.
var papierOK = []byte{0x00}

// druckerOptionen konfiguriert das Verhalten des Test-Druckers.
type druckerOptionen struct {
	papierstatusAntwort       []byte        // Antwort auf DLE EOT n=4 (leer = keine Antwort)
	zweitePapierstatusAntwort []byte        // Antwort ab der zweiten DLE-EOT-Abfrage (leer = wie die erste)
	papierstatusVerzoegerung  time.Duration // Wartezeit vor der Papierstatus-Antwort
	antwortetAufQuittung      bool          // beantwortet GS r 1
	quittungVerzoegerung      time.Duration // Wartezeit vor der Quittungsantwort
	schliesstVorQuittung      bool          // schließt die Verbindung, sobald GS r 1 eintrifft
}

// testDrucker ist ein echter lokaler TCP-Listener als Drucker-Ersatz: er zählt
// Verbindungen und hält je Verbindung die empfangenen Bytes fest.
type testDrucker struct {
	opt      druckerOptionen
	listener net.Listener

	// abgeschlossen meldet je bediente Verbindung, dass alle Bytes gelesen sind.
	abgeschlossen chan struct{}

	mu           sync.Mutex
	verbindungen int
	empfangen    [][]byte
}

func starteTestDrucker(t *testing.T, opt druckerOptionen) *testDrucker {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Listener konnte nicht starten: %v", err)
	}
	t.Cleanup(func() { _ = listener.Close() })

	drucker := &testDrucker{
		opt:           opt,
		listener:      listener,
		abgeschlossen: make(chan struct{}, 8),
	}
	go drucker.akzeptiere()
	return drucker
}

// verbinde ist der Injektionspunkt für zustelleGruppe: die Ziel-IP des Auftrags
// wird ignoriert, verbunden wird mit dem lokalen Test-Listener.
func (d *testDrucker) verbinde(string) (net.Conn, error) {
	return net.DialTimeout("tcp", d.listener.Addr().String(), dialTimeout)
}

func (d *testDrucker) akzeptiere() {
	for {
		conn, err := d.listener.Accept()
		if err != nil {
			return
		}

		d.mu.Lock()
		d.verbindungen++
		d.mu.Unlock()

		go d.bediene(conn)
	}
}

func (d *testDrucker) bediene(conn net.Conn) {
	defer func() {
		_ = conn.Close()
		d.abgeschlossen <- struct{}{}
	}()

	spur := d.neueSpur()
	papierAbfragen := 0
	byteBuf := make([]byte, 1)
	for {
		n, err := conn.Read(byteBuf)
		if n > 0 {
			empfangen := d.merke(spur, byteBuf[0])
			switch {
			case bytes.HasSuffix(empfangen, dlePapierstatus):
				papierAbfragen++
				if antwort := d.papierstatusAntwort(papierAbfragen); len(antwort) > 0 {
					time.Sleep(d.opt.papierstatusVerzoegerung)
					_, _ = conn.Write(antwort)
				}
			case bytes.HasSuffix(empfangen, gsStatusabfrage):
				if d.opt.schliesstVorQuittung {
					return
				}
				if d.opt.antwortetAufQuittung {
					time.Sleep(d.opt.quittungVerzoegerung)
					_, _ = conn.Write([]byte{0x00})
				}
			}
		}
		if err != nil {
			return
		}
	}
}

// papierstatusAntwort liefert die Antwort auf die n-te DLE-EOT-Abfrage. Ab der
// zweiten kann sie abweichen — so lässt sich Papier simulieren, das mitten in der
// Gruppe ausgeht.
func (d *testDrucker) papierstatusAntwort(abfrage int) []byte {
	if abfrage > 1 && len(d.opt.zweitePapierstatusAntwort) > 0 {
		return d.opt.zweitePapierstatusAntwort
	}
	return d.opt.papierstatusAntwort
}

func (d *testDrucker) neueSpur() int {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.empfangen = append(d.empfangen, nil)
	return len(d.empfangen) - 1
}

func (d *testDrucker) merke(spur int, b byte) []byte {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.empfangen[spur] = append(d.empfangen[spur], b)
	return d.empfangen[spur]
}

// warteAufAbschluss wartet, bis der Drucker eine Verbindung fertig abgearbeitet
// hat — erst danach sind die empfangenen Bytes vollständig.
func (d *testDrucker) warteAufAbschluss(t *testing.T) {
	t.Helper()
	select {
	case <-d.abgeschlossen:
	case <-time.After(5 * time.Second):
		t.Fatal("Drucker hat die Verbindung nicht abgeschlossen")
	}
}

func (d *testDrucker) anzahlVerbindungen() int {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.verbindungen
}

func (d *testDrucker) empfangeneBytes(spur int) []byte {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.empfangen[spur]
}

func bonDaten(id int) []byte {
	return []byte(fmt.Sprintf("BON-%d\n", id))
}

func testAuftraege(anzahl int) []DruckAuftrag {
	auftraege := make([]DruckAuftrag, 0, anzahl)
	for id := 1; id <= anzahl; id++ {
		auftraege = append(auftraege, DruckAuftrag{
			ID:      id,
			ZielIP:  testZielIP,
			Payload: base64.StdEncoding.EncodeToString(bonDaten(id)),
		})
	}
	return auftraege
}

// erwarteterDatenstrom ist die Bytefolge, die der Drucker über die eine
// Verbindung sehen muss: Papierstatus-Abfrage, alle Bons in ID-Reihenfolge,
// Quittungskommando.
func erwarteterDatenstrom(anzahlBons int) []byte {
	strom := append([]byte(nil), dlePapierstatus...)
	for id := 1; id <= anzahlBons; id++ {
		strom = append(strom, bonDaten(id)...)
	}
	return append(strom, gsStatusabfrage...)
}

func TestZustelleGruppeNutztEineVerbindungFuerAlleBons(t *testing.T) {
	// Regression auf das gemeldete Fehlerbild: Bondrucker mit Ethernet-Modul
	// nehmen typischerweise genau eine Verbindung gleichzeitig an und weisen
	// jede weitere ab; früher gingen die Bons 2..6 dabei verloren.
	drucker := starteTestDrucker(t, druckerOptionen{
		papierstatusAntwort:  papierOK,
		antwortetAufQuittung: true,
	})
	auftraege := testAuftraege(6)

	gedruckte, fehler := zustelleGruppe(drucker.verbinde, testTimeouts)(testZielIP, auftraege)
	drucker.warteAufAbschluss(t)

	if want := []int{1, 2, 3, 4, 5, 6}; !reflect.DeepEqual(gedruckte, want) {
		t.Fatalf("gedruckte: got %v, want %v", gedruckte, want)
	}
	if len(fehler) != 0 {
		t.Fatalf("fehler: got %+v, want keine", fehler)
	}
	if got := drucker.anzahlVerbindungen(); got != 1 {
		t.Fatalf("TCP-Verbindungen: got %d, want 1", got)
	}
	if got, want := drucker.empfangeneBytes(0), erwarteterDatenstrom(6); !bytes.Equal(got, want) {
		t.Fatalf("Datenstrom: got %q, want %q", got, want)
	}
}

func TestStelleGruppeZuMeldetDenQuittungsAusgang(t *testing.T) {
	// Der Quittungs-Ausgang ist das einzige Diagnosemittel im Log, um nach einem
	// Einsatz zu erkennen, ob der Drucker GS r beantwortet.
	tests := []struct {
		name           string
		opt            druckerOptionen
		wantAusgang    string
		wantGedruckte  []int
		wantFehlerIDs  []int  // leer = kein Fehlversuch erwartet
		wantFehlerText string // Textbaustein, den jeder Fehlversuch enthalten muss
	}{
		{
			name:          "Antwort auf GS r bestaetigt die ganze Gruppe",
			opt:           druckerOptionen{papierstatusAntwort: papierOK, antwortetAufQuittung: true},
			wantAusgang:   ausgangBestaetigt,
			wantGedruckte: []int{1, 2, 3},
		},
		{
			// Drucker ohne GS-r-Unterstützung: keine Antwort ist ein Lese-Timeout,
			// die Papierprüfung danach meldet aber Papier — das darf die Gruppe
			// nicht dauerhaft unbenutzbar machen.
			name:          "ohne Antwort auf GS r gilt die Gruppe trotzdem als zugestellt",
			opt:           druckerOptionen{papierstatusAntwort: papierOK},
			wantAusgang:   ausgangUnbeantwortet,
			wantGedruckte: []int{1, 2, 3},
		},
		{
			// Drucker ohne jede Statusunterstützung: weder GS r noch DLE EOT werden
			// beantwortet. Auch er muss benutzbar bleiben.
			name:          "Drucker ohne jede Statusantwort gilt als zugestellt",
			opt:           druckerOptionen{},
			wantAusgang:   ausgangUnbeantwortet,
			wantGedruckte: []int{1, 2, 3},
		},
		{
			// Geht das Papier während der Gruppe aus, geht der Drucker laut
			// ESC/POS-Referenz offline und führt GS r nicht mehr aus. Nur die
			// Papierprüfung nach dem Timeout deckt auf, dass die Bons fehlen.
			name: "Papierende waehrend der Gruppe bestaetigt nichts",
			opt: druckerOptionen{
				papierstatusAntwort:       papierOK,
				zweitePapierstatusAntwort: []byte{0x60},
			},
			wantAusgang:    ausgangAbgebrochen,
			wantFehlerIDs:  []int{1, 2, 3},
			wantFehlerText: "papier leer",
		},
		{
			name:           "Verbindungsabbruch vor der Quittung bestaetigt nichts",
			opt:            druckerOptionen{papierstatusAntwort: papierOK, schliesstVorQuittung: true},
			wantAusgang:    ausgangAbgebrochen,
			wantFehlerIDs:  []int{1, 2, 3},
			wantFehlerText: "quittung",
		},
		{
			// Mehrbytige Papierstatus-Antwort (z. B. bei aktivem Automatic Status
			// Back): das überzählige Byte darf nicht als Quittung durchgehen.
			name:          "zweites Byte der Papierstatus-Antwort quittiert nicht",
			opt:           druckerOptionen{papierstatusAntwort: []byte{0x00, 0x00}},
			wantAusgang:   ausgangUnbeantwortet,
			wantGedruckte: []int{1, 2, 3},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			drucker := starteTestDrucker(t, tt.opt)

			ergebnis := stelleGruppeZu(drucker.verbinde, testTimeouts, testZielIP, testAuftraege(3))
			drucker.warteAufAbschluss(t)

			if ergebnis.ausgang != tt.wantAusgang {
				t.Fatalf("ausgang: got %q, want %q", ergebnis.ausgang, tt.wantAusgang)
			}
			if !reflect.DeepEqual(ergebnis.gedruckteIDs, tt.wantGedruckte) {
				t.Fatalf("gedruckteIDs: got %v, want %v", ergebnis.gedruckteIDs, tt.wantGedruckte)
			}
			if got := fehlversuchsIDs(ergebnis.fehler); !reflect.DeepEqual(got, tt.wantFehlerIDs) {
				t.Fatalf("Fehlversuche: got %v, want %v", got, tt.wantFehlerIDs)
			}
			for _, f := range ergebnis.fehler {
				if !strings.Contains(f.Fehler, tt.wantFehlerText) {
					t.Fatalf("Fehlermeldung: got %q, want Hinweis auf %q", f.Fehler, tt.wantFehlerText)
				}
			}
		})
	}
}

// fehlversuchsIDs reduziert Fehlversuche auf ihre Auftrags-IDs; nil bei keinem
// Fehlversuch, damit der Vergleich mit dem leeren Erwartungswert aufgeht.
func fehlversuchsIDs(fehlversuche []fehlversuch) []int {
	var ids []int
	for _, f := range fehlversuche {
		ids = append(ids, f.ID)
	}
	return ids
}

func TestStelleGruppeZuZaehltGesendeteBons(t *testing.T) {
	// gesendet speist die Logzeile und ist damit das einzige Mittel, nach einem
	// Einsatz zu erkennen, wie weit eine abgebrochene Gruppe gekommen ist. Ohne
	// diesen Test dürfte der Zähler dauerhaft 0 melden, ohne dass es auffällt.
	tests := []struct {
		name         string
		opt          druckerOptionen
		schreibunfug []byte // Inhalt, an dem das Schreiben scheitert (nil = kein Fehler)
		wantGesendet int
	}{
		{
			name:         "alle Bons gesendet",
			opt:          druckerOptionen{papierstatusAntwort: papierOK, antwortetAufQuittung: true},
			wantGesendet: 3,
		},
		{
			name:         "Schreibfehler bei Bon 3 zaehlt nur die beiden davor",
			opt:          druckerOptionen{papierstatusAntwort: papierOK},
			schreibunfug: bonDaten(3),
			wantGesendet: 2,
		},
		{
			name:         "Papier leer sendet keinen Bon",
			opt:          druckerOptionen{papierstatusAntwort: []byte{0x60}},
			wantGesendet: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			drucker := starteTestDrucker(t, tt.opt)

			verbinde := drucker.verbinde
			if tt.schreibunfug != nil {
				verbinde = func(zielIP string) (net.Conn, error) {
					conn, err := drucker.verbinde(zielIP)
					if err != nil {
						return nil, err
					}
					return &schreibfehlerConn{Conn: conn, fehlerBeiInhalt: tt.schreibunfug}, nil
				}
			}

			ergebnis := stelleGruppeZu(verbinde, testTimeouts, testZielIP, testAuftraege(3))
			drucker.warteAufAbschluss(t)

			if ergebnis.gesendet != tt.wantGesendet {
				t.Fatalf("gesendet: got %d, want %d", ergebnis.gesendet, tt.wantGesendet)
			}
		})
	}
}

func TestStelleGruppeZuVerwirftVerspaetetePapierstatusAntwort(t *testing.T) {
	// Antwortet ein langsamer Drucker erst nach dem Papierstatus-Timeout, liegt
	// sein Byte noch im Empfangspuffer, wenn die Quittung eingeholt wird. Es darf
	// die Gruppe nicht bestätigen — der Drucker hat nichts verarbeitet.
	drucker := starteTestDrucker(t, druckerOptionen{
		papierstatusAntwort:      papierOK,
		papierstatusVerzoegerung: 3 * testTimeouts.papier,
	})
	timeouts := testTimeouts
	timeouts.spuelen = 300 * time.Millisecond

	ergebnis := stelleGruppeZu(drucker.verbinde, timeouts, testZielIP, testAuftraege(3))
	drucker.warteAufAbschluss(t)

	if ergebnis.ausgang != ausgangUnbeantwortet {
		t.Fatalf("ausgang: got %q, want %q", ergebnis.ausgang, ausgangUnbeantwortet)
	}
	if want := []int{1, 2, 3}; !reflect.DeepEqual(ergebnis.gedruckteIDs, want) {
		t.Fatalf("gedruckteIDs: got %v, want %v", ergebnis.gedruckteIDs, want)
	}
}

func TestStelleGruppeZuWartetLaengerJeBonAufDieQuittung(t *testing.T) {
	// GS r wird erst beantwortet, wenn alle Bons gedruckt und geschnitten sind —
	// das Quittungsfenster muss deshalb mit der Bonzahl wachsen. Die Verzögerung
	// hier überschreitet die Basis allein um das Vierfache und passt nur in die
	// Summe aus Basis und Zuschlag für drei Bons (25 ms + 3 * 200 ms).
	drucker := starteTestDrucker(t, druckerOptionen{
		papierstatusAntwort:  papierOK,
		antwortetAufQuittung: true,
		quittungVerzoegerung: 100 * time.Millisecond,
	})
	timeouts := testTimeouts
	timeouts.quittungBasis = 25 * time.Millisecond
	timeouts.quittungProBon = 200 * time.Millisecond

	ergebnis := stelleGruppeZu(drucker.verbinde, timeouts, testZielIP, testAuftraege(3))
	drucker.warteAufAbschluss(t)

	if ergebnis.ausgang != ausgangBestaetigt {
		t.Fatalf("ausgang: got %q, want %q", ergebnis.ausgang, ausgangBestaetigt)
	}
	if want := []int{1, 2, 3}; !reflect.DeepEqual(ergebnis.gedruckteIDs, want) {
		t.Fatalf("gedruckteIDs: got %v, want %v", ergebnis.gedruckteIDs, want)
	}
}

func TestZustelleGruppeUngueltigesBase64BrichtDieGruppeAb(t *testing.T) {
	drucker := starteTestDrucker(t, druckerOptionen{
		papierstatusAntwort:  papierOK,
		antwortetAufQuittung: true,
	})
	auftraege := testAuftraege(3)
	auftraege[1].Payload = "kein base64!"

	gedruckte, fehler := zustelleGruppe(drucker.verbinde, testTimeouts)(testZielIP, auftraege)
	drucker.warteAufAbschluss(t)

	if len(gedruckte) != 0 {
		t.Fatalf("gedruckte: got %v, want none", gedruckte)
	}
	// Nur der Auftrag mit der kaputten Payload verbraucht einen Versuch — die
	// gesunden Aufträge der Gruppe dürfen daran nicht verbrennen.
	if len(fehler) != 1 || fehler[0].ID != 2 {
		t.Fatalf("fehler: got %+v, want genau einen auf Auftrag 2", fehler)
	}
	if !strings.Contains(fehler[0].Fehler, "Base64") {
		t.Fatalf("Fehlermeldung: got %q, want Hinweis auf Base64", fehler[0].Fehler)
	}
	// Bon 1 ging noch raus, danach nichts mehr — auch keine Quittungsabfrage.
	want := append([]byte(nil), dlePapierstatus...)
	want = append(want, bonDaten(1)...)
	if got := drucker.empfangeneBytes(0); !bytes.Equal(got, want) {
		t.Fatalf("Datenstrom: got %q, want %q", got, want)
	}
}

// schreibfehlerConn lässt genau den Schreibvorgang scheitern, dessen Inhalt
// fehlerBeiInhalt entspricht — so lässt sich ein Drucker simulieren, der mitten
// in der Gruppe keine Daten mehr annimmt, ohne von der Zahl der Schreibvorgänge
// abzuhängen.
type schreibfehlerConn struct {
	net.Conn
	fehlerBeiInhalt []byte
}

func (c *schreibfehlerConn) Write(b []byte) (int, error) {
	if bytes.Equal(b, c.fehlerBeiInhalt) {
		return 0, errors.New("verbindung abgebrochen")
	}
	return c.Conn.Write(b)
}

func TestZustelleGruppeSchreibfehlerMeldetNurDenBetroffenenBon(t *testing.T) {
	drucker := starteTestDrucker(t, druckerOptionen{papierstatusAntwort: papierOK})
	auftraege := testAuftraege(6)

	// Bon 3 lässt sich nicht mehr schreiben.
	verbinde := func(zielIP string) (net.Conn, error) {
		conn, err := drucker.verbinde(zielIP)
		if err != nil {
			return nil, err
		}
		return &schreibfehlerConn{Conn: conn, fehlerBeiInhalt: bonDaten(3)}, nil
	}

	gedruckte, fehler := zustelleGruppe(verbinde, testTimeouts)(testZielIP, auftraege)
	drucker.warteAufAbschluss(t)

	if len(gedruckte) != 0 {
		t.Fatalf("gedruckte: got %v, want none", gedruckte)
	}
	if len(fehler) != 1 || fehler[0].ID != 3 {
		t.Fatalf("fehler: got %+v, want genau einen auf Auftrag 3", fehler)
	}
	if !strings.Contains(fehler[0].Fehler, "senden fehlgeschlagen") {
		t.Fatalf("Fehlermeldung: got %q, want Hinweis auf das Senden", fehler[0].Fehler)
	}
	// Nach dem Abbruch dürfen die Bons 4..6 nicht mehr gesendet worden sein.
	want := append([]byte(nil), dlePapierstatus...)
	want = append(want, bonDaten(1)...)
	want = append(want, bonDaten(2)...)
	if got := drucker.empfangeneBytes(0); !bytes.Equal(got, want) {
		t.Fatalf("Datenstrom: got %q, want %q", got, want)
	}
}

func TestZustelleGruppePapierLeerSendetKeinenBon(t *testing.T) {
	drucker := starteTestDrucker(t, druckerOptionen{papierstatusAntwort: []byte{0x60}})
	auftraege := testAuftraege(6)

	gedruckte, fehler := zustelleGruppe(drucker.verbinde, testTimeouts)(testZielIP, auftraege)
	drucker.warteAufAbschluss(t)

	if len(gedruckte) != 0 {
		t.Fatalf("gedruckte: got %v, want none", gedruckte)
	}
	if want := []int{1, 2, 3, 4, 5, 6}; !reflect.DeepEqual(fehlversuchsIDs(fehler), want) {
		t.Fatalf("Fehlversuche: got %v, want %v", fehlversuchsIDs(fehler), want)
	}
	if !strings.Contains(fehler[0].Fehler, "papier leer") {
		t.Fatalf("Fehlermeldung: got %q, want Hinweis auf leeres Papier", fehler[0].Fehler)
	}
	if got := drucker.empfangeneBytes(0); !bytes.Equal(got, dlePapierstatus) {
		t.Fatalf("Datenstrom: got %q, want nur die Papierstatus-Abfrage", got)
	}
}

func TestZustelleGruppeOhnePapierstatusAntwortDrucktTrotzdem(t *testing.T) {
	// Nicht jeder Drucker beantwortet DLE EOT; das darf die Zustellung nicht
	// verhindern.
	drucker := starteTestDrucker(t, druckerOptionen{antwortetAufQuittung: true})
	auftraege := testAuftraege(2)

	gedruckte, fehler := zustelleGruppe(drucker.verbinde, testTimeouts)(testZielIP, auftraege)
	drucker.warteAufAbschluss(t)

	if want := []int{1, 2}; !reflect.DeepEqual(gedruckte, want) {
		t.Fatalf("gedruckte: got %v, want %v", gedruckte, want)
	}
	if len(fehler) != 0 {
		t.Fatalf("fehler: got %+v, want keine", fehler)
	}
}

func TestZustelleGruppeNichtErreichbarerDrucker(t *testing.T) {
	// Jeder Auftrag der Gruppe muss einen Fehlversuch bekommen: sonst erreicht nur
	// der erste die Höchstzahl an Versuchen und die übrigen bleiben unbemerkt
	// offen, statt im Admin als fehlgeschlagen aufzutauchen.
	verbinde := func(string) (net.Conn, error) {
		return nil, errors.New("connection refused")
	}
	auftraege := testAuftraege(3)

	gedruckte, fehler := zustelleGruppe(verbinde, testTimeouts)(testZielIP, auftraege)

	if len(gedruckte) != 0 {
		t.Fatalf("gedruckte: got %v, want none", gedruckte)
	}
	if want := []int{1, 2, 3}; !reflect.DeepEqual(fehlversuchsIDs(fehler), want) {
		t.Fatalf("Fehlversuche: got %v, want %v", fehlversuchsIDs(fehler), want)
	}
	for _, f := range fehler {
		if !strings.Contains(f.Fehler, "nicht erreichbar") {
			t.Fatalf("Fehlermeldung: got %q, want Hinweis auf die Erreichbarkeit", f.Fehler)
		}
	}
}

func TestZustelleGruppeLeereGruppeVerbindetNicht(t *testing.T) {
	verbunden := false
	verbinde := func(string) (net.Conn, error) {
		verbunden = true
		return nil, errors.New("darf nicht passieren")
	}

	gedruckte, fehler := zustelleGruppe(verbinde, testTimeouts)(testZielIP, nil)

	if verbunden {
		t.Fatalf("leere Gruppe hat eine Verbindung geöffnet")
	}
	if len(gedruckte) != 0 || len(fehler) != 0 {
		t.Fatalf("Ergebnis: got %v / %+v, want leer", gedruckte, fehler)
	}
}
