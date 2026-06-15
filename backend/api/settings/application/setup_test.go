//go:build unit

package application

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/nicograef/jotti/backend/domain/settings"
	"github.com/nicograef/jotti/backend/domain/tse"
)

// stubCommandRepo erfasst die gespeicherte Konfiguration und liefert eine feste
// Kassenidentitaet, damit die Orchestrator-Tests Seriennummer und Speicherung
// pruefen koennen.
type stubCommandRepo struct {
	identitaet  settings.Kassenidentitaet
	gespeichert *settings.TSEKonfiguration
	upsertErr   error
}

func (s *stubCommandRepo) UpsertBetreiber(context.Context, settings.Betreiber) error {
	return errors.New("not implemented")
}

func (s *stubCommandRepo) GetKassenidentitaet(context.Context) (settings.Kassenidentitaet, error) {
	return s.identitaet, nil
}

func (s *stubCommandRepo) UpsertTSEKonfiguration(_ context.Context, c settings.TSEKonfiguration) error {
	if s.upsertErr != nil {
		return s.upsertErr
	}
	s.gespeichert = &c
	return nil
}

func commandMit(repo *stubCommandRepo, client *tse.FakeSetupClient) Command {
	return Command{
		SettingsRepo: repo,
		NewTSESetupClient: func(tse.SetupCredentials) (tse.SetupClient, error) {
			return client, nil
		},
	}
}

func zugangsdaten() tse.SetupCredentials {
	return tse.SetupCredentials{ApiKey: "api-key", ApiSecret: "api-secret"}
}

// TestRichteTSEEin_LeeresKonto sichert den Voll-Durchlauf: aus einem leeren
// Konto entsteht eine initialisierte TSS mit registriertem Client, dessen
// serial_number die Kassen-Seriennummer ist; PUK und PIN werden zurueckgegeben
// und die vollstaendige Konfiguration gespeichert.
func TestRichteTSEEin_LeeresKonto(t *testing.T) {
	seriennummer := uuid.New()
	repo := &stubCommandRepo{identitaet: settings.Kassenidentitaet{Seriennummer: seriennummer}}
	client := &tse.FakeSetupClient{
		UmgebungResponse:  tse.UmgebungTest,
		CreateTSSResponse: tse.TSSErstellt{ID: "tss-neu", PUK: "puk-123", State: "CREATED"},
	}

	ergebnis, err := commandMit(repo, client).RichteTSEEin(context.Background(), zugangsdaten(), tse.UmgebungTest)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ergebnis.TssID != "tss-neu" {
		t.Fatalf("expected tss id tss-neu, got %q", ergebnis.TssID)
	}
	if ergebnis.PUK != "puk-123" {
		t.Fatalf("expected puk to be returned, got %q", ergebnis.PUK)
	}
	if ergebnis.AdminPIN == "" {
		t.Fatal("expected an admin pin to be returned")
	}
	// Der Client wird unter einer eigenen UUIDv4 angelegt (fiskaly-Konvention),
	// nicht unter der Kassen-Seriennummer.
	if ergebnis.ClientID == "" || ergebnis.ClientID == seriennummer.String() {
		t.Fatalf("expected a distinct generated client id, got %q", ergebnis.ClientID)
	}

	if len(client.RegistrierteClients) != 1 {
		t.Fatalf("expected exactly one registered client, got %d", len(client.RegistrierteClients))
	}
	registriert := client.RegistrierteClients[0]
	if registriert.SerialNumber != seriennummer.String() {
		t.Fatalf("expected client registered with kassen serial number as serial_number, got %+v", registriert)
	}
	if registriert.ClientID != ergebnis.ClientID {
		t.Fatalf("expected client registered under the returned client id %q, got %q", ergebnis.ClientID, registriert.ClientID)
	}
	if client.GesetzteAdminPIN != ergebnis.AdminPIN {
		t.Fatalf("expected the generated pin to be set on the TSS, got %q vs %q", client.GesetzteAdminPIN, ergebnis.AdminPIN)
	}

	if repo.gespeichert == nil {
		t.Fatal("expected the configuration to be saved")
	}
	if repo.gespeichert.TssID != "tss-neu" || repo.gespeichert.ClientID != ergebnis.ClientID {
		t.Fatalf("expected full configuration to be saved, got %+v", repo.gespeichert)
	}
	if !repo.gespeichert.IstKonfiguriert() {
		t.Fatal("expected the saved configuration to be complete")
	}
}

// TestRichteTSEEin_UmgebungAbweichung sichert den LIVE-Schutz: bestaetigt der
// Admin TEST, zeigen die Zugangsdaten aber auf LIVE, bricht die Einrichtung ab,
// bevor irgendeine TSS angelegt wird.
func TestRichteTSEEin_UmgebungAbweichung(t *testing.T) {
	repo := &stubCommandRepo{identitaet: settings.Kassenidentitaet{Seriennummer: uuid.New()}}
	client := &tse.FakeSetupClient{UmgebungResponse: tse.UmgebungLive}

	_, err := commandMit(repo, client).RichteTSEEin(context.Background(), zugangsdaten(), tse.UmgebungTest)
	if !errors.Is(err, ErrTSESetupUmgebungAbweichung) {
		t.Fatalf("expected ErrTSESetupUmgebungAbweichung, got %v", err)
	}
	if client.CreateTSSCalls != 0 {
		t.Fatalf("expected no TSS to be created on environment mismatch, got %d calls", client.CreateTSSCalls)
	}
	if repo.gespeichert != nil {
		t.Fatal("expected no configuration to be saved on environment mismatch")
	}
}

// TestRichteTSEEin_BestaetigteUmgebungUngueltig sichert ab, dass eine nicht
// bestaetigte Umgebung (weder TEST noch LIVE) abgewiesen wird.
func TestRichteTSEEin_BestaetigteUmgebungUngueltig(t *testing.T) {
	repo := &stubCommandRepo{identitaet: settings.Kassenidentitaet{Seriennummer: uuid.New()}}
	client := &tse.FakeSetupClient{UmgebungResponse: tse.UmgebungTest}

	_, err := commandMit(repo, client).RichteTSEEin(context.Background(), zugangsdaten(), tse.Umgebung(""))
	if !errors.Is(err, ErrTSESetupUmgebungAbweichung) {
		t.Fatalf("expected ErrTSESetupUmgebungAbweichung for an unconfirmed environment, got %v", err)
	}
	if client.CreateTSSCalls != 0 {
		t.Fatalf("expected no TSS to be created, got %d calls", client.CreateTSSCalls)
	}
}

// TestRichteTSEEin_VorhandeneAktiveTSS sichert: existiert bereits eine aktive
// TSS, wird keine neue angelegt.
func TestRichteTSEEin_VorhandeneAktiveTSS(t *testing.T) {
	repo := &stubCommandRepo{identitaet: settings.Kassenidentitaet{Seriennummer: uuid.New()}}
	client := &tse.FakeSetupClient{
		UmgebungResponse: tse.UmgebungTest,
		TSSResponse:      []tse.TSSInfo{{ID: "tss-alt", State: "INITIALIZED"}},
	}

	_, err := commandMit(repo, client).RichteTSEEin(context.Background(), zugangsdaten(), tse.UmgebungTest)
	if !errors.Is(err, ErrTSEBereitsEingerichtet) {
		t.Fatalf("expected ErrTSEBereitsEingerichtet, got %v", err)
	}
	if client.CreateTSSCalls != 0 {
		t.Fatalf("expected no TSS to be created when an active TSS exists, got %d calls", client.CreateTSSCalls)
	}
	if repo.gespeichert != nil {
		t.Fatal("expected no configuration to be saved when an active TSS exists")
	}
}

// TestRichteTSEEin_DeaktivierteTSSBlocktNicht sichert, dass eine ausschliesslich
// deaktivierte (DISABLED) TSS die Neuanlage nicht blockiert.
func TestRichteTSEEin_DeaktivierteTSSBlocktNicht(t *testing.T) {
	seriennummer := uuid.New()
	repo := &stubCommandRepo{identitaet: settings.Kassenidentitaet{Seriennummer: seriennummer}}
	client := &tse.FakeSetupClient{
		UmgebungResponse:  tse.UmgebungTest,
		TSSResponse:       []tse.TSSInfo{{ID: "tss-tot", State: "DISABLED"}},
		CreateTSSResponse: tse.TSSErstellt{ID: "tss-neu", PUK: "puk", State: "CREATED"},
	}

	_, err := commandMit(repo, client).RichteTSEEin(context.Background(), zugangsdaten(), tse.UmgebungTest)
	if err != nil {
		t.Fatalf("unexpected error with only a disabled TSS present: %v", err)
	}
	if client.CreateTSSCalls != 1 {
		t.Fatalf("expected a new TSS to be created, got %d calls", client.CreateTSSCalls)
	}
}

// TestRichteTSEEin_AbbruchSpeichertNicht sichert die Atomaritaet: bricht ein
// Lebenszyklus-Schritt ab, bleibt keine halbe Konfiguration in der DB.
func TestRichteTSEEin_AbbruchSpeichertNicht(t *testing.T) {
	repo := &stubCommandRepo{identitaet: settings.Kassenidentitaet{Seriennummer: uuid.New()}}
	client := &tse.FakeSetupClient{
		UmgebungResponse:  tse.UmgebungTest,
		CreateTSSResponse: tse.TSSErstellt{ID: "tss-neu", PUK: "puk", State: "CREATED"},
		RegistriereErr:    errors.New("client registration failed"),
	}

	_, err := commandMit(repo, client).RichteTSEEin(context.Background(), zugangsdaten(), tse.UmgebungTest)
	if !errors.Is(err, ErrTSEEinrichtung) {
		t.Fatalf("expected ErrTSEEinrichtung on a failing step, got %v", err)
	}
	if repo.gespeichert != nil {
		t.Fatal("expected no configuration to be saved when a step fails")
	}
}

// TestRichteTSEEin_FalscheZugangsdaten sichert, dass ein Auth-Fehler in eine
// verstaendliche Zugangsdaten-Meldung uebersetzt wird.
func TestRichteTSEEin_FalscheZugangsdaten(t *testing.T) {
	repo := &stubCommandRepo{identitaet: settings.Kassenidentitaet{Seriennummer: uuid.New()}}
	client := &tse.FakeSetupClient{TSSErr: tse.ErrSetupAuthFehlgeschlagen}

	_, err := commandMit(repo, client).RichteTSEEin(context.Background(), zugangsdaten(), tse.UmgebungTest)
	if !errors.Is(err, ErrTSESetupZugangsdaten) {
		t.Fatalf("expected ErrTSESetupZugangsdaten, got %v", err)
	}
	if client.CreateTSSCalls != 0 {
		t.Fatalf("expected no TSS to be created on auth failure, got %d calls", client.CreateTSSCalls)
	}
}

// TestUebernimmTSE_WiederaufnahmeCreated sichert die Wiederaufnahme nach einem
// Abbruch im Zustand CREATED: der PUK wird idempotent erneut bezogen, eine
// frische PIN erzeugt und der Lebenszyklus vollendet — ohne zweite TSS und ohne
// Nutzereingabe.
func TestUebernimmTSE_WiederaufnahmeCreated(t *testing.T) {
	seriennummer := uuid.New()
	repo := &stubCommandRepo{identitaet: settings.Kassenidentitaet{Seriennummer: seriennummer}}
	client := &tse.FakeSetupClient{
		UmgebungResponse:     tse.UmgebungTest,
		TSSResponse:          []tse.TSSInfo{{ID: "tss-halb", State: "CREATED"}},
		HoleAdminPUKResponse: "puk-refetch",
	}

	ergebnis, err := commandMit(repo, client).UebernimmTSE(context.Background(), zugangsdaten(), tse.UmgebungTest, "tss-halb", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client.CreateTSSCalls != 0 {
		t.Fatalf("expected no new TSS to be created on resume, got %d calls", client.CreateTSSCalls)
	}
	if client.HoleAdminPUKCalls != 1 {
		t.Fatalf("expected the puk to be refetched once, got %d calls", client.HoleAdminPUKCalls)
	}
	if ergebnis.PUK != "puk-refetch" || ergebnis.AdminPIN == "" {
		t.Fatalf("expected refetched puk and a fresh pin, got %+v", ergebnis)
	}
	if client.GesetzteAdminPIN != ergebnis.AdminPIN {
		t.Fatalf("expected the fresh pin to be set on the TSS, got %q vs %q", client.GesetzteAdminPIN, ergebnis.AdminPIN)
	}
	if len(client.RegistrierteClients) != 1 || client.RegistrierteClients[0].SerialNumber != seriennummer.String() {
		t.Fatalf("expected exactly one client registered with the kassen serial, got %+v", client.RegistrierteClients)
	}
	if repo.gespeichert == nil || repo.gespeichert.TssID != "tss-halb" {
		t.Fatalf("expected the configuration to be saved for the resumed TSS, got %+v", repo.gespeichert)
	}
}

// TestUebernimmTSE_WiederaufnahmeUninitialized sichert die Wiederaufnahme ab
// UNINITIALIZED: mit der vom Admin verwahrten PIN wird initialisiert und der
// Client registriert; es entstehen keine neuen Geheimnisse.
func TestUebernimmTSE_WiederaufnahmeUninitialized(t *testing.T) {
	seriennummer := uuid.New()
	repo := &stubCommandRepo{identitaet: settings.Kassenidentitaet{Seriennummer: seriennummer}}
	client := &tse.FakeSetupClient{
		UmgebungResponse: tse.UmgebungTest,
		TSSResponse:      []tse.TSSInfo{{ID: "tss-uninit", State: "UNINITIALIZED"}},
	}

	ergebnis, err := commandMit(repo, client).UebernimmTSE(context.Background(), zugangsdaten(), tse.UmgebungTest, "tss-uninit", "1234567890")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client.HoleAdminPUKCalls != 0 {
		t.Fatalf("expected no puk refetch from UNINITIALIZED, got %d calls", client.HoleAdminPUKCalls)
	}
	if client.AuthentifiziertePIN != "1234567890" {
		t.Fatalf("expected the entered pin to be used for admin auth, got %q", client.AuthentifiziertePIN)
	}
	if ergebnis.PUK != "" || ergebnis.AdminPIN != "" {
		t.Fatalf("expected no new secrets on resume from UNINITIALIZED, got %+v", ergebnis)
	}
	if len(client.RegistrierteClients) != 1 {
		t.Fatalf("expected exactly one client registered, got %d", len(client.RegistrierteClients))
	}
	if repo.gespeichert == nil || repo.gespeichert.TssID != "tss-uninit" {
		t.Fatalf("expected the configuration to be saved, got %+v", repo.gespeichert)
	}
}

// TestUebernimmTSE_InitialisiertOhneClient sichert die Uebernahme einer bereits
// initialisierten TSS, die noch keinen Client hat: mit PIN wird nur noch der
// Client registriert.
func TestUebernimmTSE_InitialisiertOhneClient(t *testing.T) {
	seriennummer := uuid.New()
	repo := &stubCommandRepo{identitaet: settings.Kassenidentitaet{Seriennummer: seriennummer}}
	client := &tse.FakeSetupClient{
		UmgebungResponse: tse.UmgebungTest,
		TSSResponse:      []tse.TSSInfo{{ID: "tss-init", State: "INITIALIZED"}},
	}

	ergebnis, err := commandMit(repo, client).UebernimmTSE(context.Background(), zugangsdaten(), tse.UmgebungTest, "tss-init", "1234567890")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(client.RegistrierteClients) != 1 || client.RegistrierteClients[0].SerialNumber != seriennummer.String() {
		t.Fatalf("expected the client to be registered with the kassen serial, got %+v", client.RegistrierteClients)
	}
	if repo.gespeichert == nil || repo.gespeichert.ClientID != ergebnis.ClientID {
		t.Fatalf("expected the configuration to be saved with the registered client, got %+v", repo.gespeichert)
	}
}

// TestUebernimmTSE_VorhandenerPassenderClient sichert, dass ein bereits passender
// Client uebernommen und kein neuer registriert wird.
func TestUebernimmTSE_VorhandenerPassenderClient(t *testing.T) {
	seriennummer := uuid.New()
	vorhandenerClient := uuid.NewString()
	repo := &stubCommandRepo{identitaet: settings.Kassenidentitaet{Seriennummer: seriennummer}}
	client := &tse.FakeSetupClient{
		UmgebungResponse: tse.UmgebungTest,
		TSSResponse:      []tse.TSSInfo{{ID: "tss-init", State: "INITIALIZED"}},
		ClientsByTSS: map[string][]tse.ClientInfo{
			"tss-init": {{ID: vorhandenerClient, SerialNumber: seriennummer.String(), State: "REGISTERED"}},
		},
	}

	ergebnis, err := commandMit(repo, client).UebernimmTSE(context.Background(), zugangsdaten(), tse.UmgebungTest, "tss-init", "1234567890")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(client.RegistrierteClients) != 0 {
		t.Fatalf("expected no new client registration when a matching client exists, got %+v", client.RegistrierteClients)
	}
	if ergebnis.ClientID != vorhandenerClient || repo.gespeichert.ClientID != vorhandenerClient {
		t.Fatalf("expected the existing client to be adopted, got result %q saved %q", ergebnis.ClientID, repo.gespeichert.ClientID)
	}
}

// TestUebernimmTSE_PINErforderlich sichert, dass die Uebernahme ab UNINITIALIZED
// ohne PIN klar als fehlende PIN gemeldet wird — vor jeder Schreiboperation.
func TestUebernimmTSE_PINErforderlich(t *testing.T) {
	repo := &stubCommandRepo{identitaet: settings.Kassenidentitaet{Seriennummer: uuid.New()}}
	client := &tse.FakeSetupClient{
		UmgebungResponse: tse.UmgebungTest,
		TSSResponse:      []tse.TSSInfo{{ID: "tss-uninit", State: "UNINITIALIZED"}},
	}

	_, err := commandMit(repo, client).UebernimmTSE(context.Background(), zugangsdaten(), tse.UmgebungTest, "tss-uninit", "")
	if !errors.Is(err, ErrTSESetupPINErforderlich) {
		t.Fatalf("expected ErrTSESetupPINErforderlich, got %v", err)
	}
	if len(client.RegistrierteClients) != 0 || repo.gespeichert != nil {
		t.Fatal("expected no writes when the pin is missing")
	}
}

// TestUebernimmTSE_UnbekanntePIN sichert, dass eine von fiskaly abgelehnte PIN in
// eine verstaendliche Sackgassen-Meldung uebersetzt wird, nicht in einen
// technischen Fehler.
func TestUebernimmTSE_UnbekanntePIN(t *testing.T) {
	repo := &stubCommandRepo{identitaet: settings.Kassenidentitaet{Seriennummer: uuid.New()}}
	client := &tse.FakeSetupClient{
		UmgebungResponse: tse.UmgebungTest,
		TSSResponse:      []tse.TSSInfo{{ID: "tss-init", State: "INITIALIZED"}},
		AuthAdminErr:     tse.ErrSetupAuthFehlgeschlagen,
	}

	_, err := commandMit(repo, client).UebernimmTSE(context.Background(), zugangsdaten(), tse.UmgebungTest, "tss-init", "0000000000")
	if !errors.Is(err, ErrTSESetupPINUnbekannt) {
		t.Fatalf("expected ErrTSESetupPINUnbekannt, got %v", err)
	}
	if repo.gespeichert != nil {
		t.Fatal("expected no configuration to be saved on an unknown pin")
	}
}

// TestUebernimmTSE_UmgebungAbweichung sichert, dass der LIVE-Schutz auch bei der
// Uebernahme greift: weicht die tatsaechliche Umgebung von der bestaetigten ab,
// bricht der Flow vor jeder Operation ab.
func TestUebernimmTSE_UmgebungAbweichung(t *testing.T) {
	repo := &stubCommandRepo{identitaet: settings.Kassenidentitaet{Seriennummer: uuid.New()}}
	client := &tse.FakeSetupClient{
		UmgebungResponse: tse.UmgebungLive,
		TSSResponse:      []tse.TSSInfo{{ID: "tss-x", State: "CREATED"}},
	}

	_, err := commandMit(repo, client).UebernimmTSE(context.Background(), zugangsdaten(), tse.UmgebungTest, "tss-x", "")
	if !errors.Is(err, ErrTSESetupUmgebungAbweichung) {
		t.Fatalf("expected ErrTSESetupUmgebungAbweichung, got %v", err)
	}
	if client.HoleAdminPUKCalls != 0 || repo.gespeichert != nil {
		t.Fatal("expected no operations on environment mismatch")
	}
}

// TestUebernimmTSE_TSSNichtGefunden sichert, dass eine im Konto fehlende TSS klar
// gemeldet wird.
func TestUebernimmTSE_TSSNichtGefunden(t *testing.T) {
	repo := &stubCommandRepo{identitaet: settings.Kassenidentitaet{Seriennummer: uuid.New()}}
	client := &tse.FakeSetupClient{UmgebungResponse: tse.UmgebungTest}

	_, err := commandMit(repo, client).UebernimmTSE(context.Background(), zugangsdaten(), tse.UmgebungTest, "tss-fehlt", "")
	if !errors.Is(err, ErrTSESetupTSSNichtGefunden) {
		t.Fatalf("expected ErrTSESetupTSSNichtGefunden, got %v", err)
	}
}

// TestUebernimmTSE_DeaktivierteTSS sichert, dass eine deaktivierte TSS nicht
// uebernommen werden kann.
func TestUebernimmTSE_DeaktivierteTSS(t *testing.T) {
	repo := &stubCommandRepo{identitaet: settings.Kassenidentitaet{Seriennummer: uuid.New()}}
	client := &tse.FakeSetupClient{
		UmgebungResponse: tse.UmgebungTest,
		TSSResponse:      []tse.TSSInfo{{ID: "tss-tot", State: "DISABLED"}},
	}

	_, err := commandMit(repo, client).UebernimmTSE(context.Background(), zugangsdaten(), tse.UmgebungTest, "tss-tot", "1234567890")
	if !errors.Is(err, ErrTSESetupUebernahmeNichtMoeglich) {
		t.Fatalf("expected ErrTSESetupUebernahmeNichtMoeglich, got %v", err)
	}
	if repo.gespeichert != nil {
		t.Fatal("expected no configuration to be saved for a disabled TSS")
	}
}
