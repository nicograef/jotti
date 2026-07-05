//go:build unit

package application

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/nicograef/jotti/backend/domain/kasse"
	"github.com/nicograef/jotti/backend/domain/tse"
)

// stubCommandRepo erfasst die gespeicherte Konfiguration und liefert eine feste
// Kassenidentitaet, damit die Orchestrator-Tests Seriennummer und Speicherung
// pruefen koennen.
type stubCommandRepo struct {
	identitaet             tse.Kassenidentitaet
	gespeichert            *tse.Konfiguration
	gespeicherteStammdaten *tse.Stammdaten
	upsertErr              error
	stammdatenUpsertErr    error
}

func (s *stubCommandRepo) GetKassenidentitaet(context.Context) (tse.Kassenidentitaet, error) {
	return s.identitaet, nil
}

func (s *stubCommandRepo) SaveEinrichtung(_ context.Context, c tse.Konfiguration) error {
	if s.upsertErr != nil {
		return s.upsertErr
	}
	s.gespeichert = &c
	return nil
}

func (s *stubCommandRepo) UpsertTSEStammdaten(_ context.Context, st tse.Stammdaten) error {
	if s.stammdatenUpsertErr != nil {
		return s.stammdatenUpsertErr
	}
	s.gespeicherteStammdaten = &st
	return nil
}

// stubKassensitzungReader liefert die offene Kassensitzung fuer den
// Konfigurations-Guard; nil (Default) heisst: keine offen.
type stubKassensitzungReader struct {
	offene *kasse.Kassensitzung
	err    error
}

func (s stubKassensitzungReader) GetOffeneKassensitzung(context.Context) (*kasse.Kassensitzung, error) {
	return s.offene, s.err
}

func commandMit(repo *stubCommandRepo, client *tse.FakeSetupClient) Command {
	return Command{
		TSERepo:             repo,
		KassensitzungenRepo: stubKassensitzungReader{},
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
	repo := &stubCommandRepo{identitaet: tse.Kassenidentitaet{Seriennummer: seriennummer}}
	client := &tse.FakeSetupClient{
		UmgebungResponse:  tse.UmgebungTest,
		CreateTSSResponse: tse.TSSErstellt{ID: "tss-neu", PUK: "puk-123", State: "CREATED"},
	}

	ergebnis, err := commandMit(repo, client).RichteTSEEin(context.Background(), zugangsdaten(), tse.UmgebungTest, false)
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
	repo := &stubCommandRepo{identitaet: tse.Kassenidentitaet{Seriennummer: uuid.New()}}
	client := &tse.FakeSetupClient{UmgebungResponse: tse.UmgebungLive}

	_, err := commandMit(repo, client).RichteTSEEin(context.Background(), zugangsdaten(), tse.UmgebungTest, false)
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
	repo := &stubCommandRepo{identitaet: tse.Kassenidentitaet{Seriennummer: uuid.New()}}
	client := &tse.FakeSetupClient{UmgebungResponse: tse.UmgebungTest}

	_, err := commandMit(repo, client).RichteTSEEin(context.Background(), zugangsdaten(), tse.Umgebung(""), false)
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
	repo := &stubCommandRepo{identitaet: tse.Kassenidentitaet{Seriennummer: uuid.New()}}
	client := &tse.FakeSetupClient{
		UmgebungResponse: tse.UmgebungTest,
		TSSResponse:      []tse.TSSInfo{{ID: "tss-alt", State: "INITIALIZED"}},
	}

	_, err := commandMit(repo, client).RichteTSEEin(context.Background(), zugangsdaten(), tse.UmgebungTest, false)
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
	repo := &stubCommandRepo{identitaet: tse.Kassenidentitaet{Seriennummer: seriennummer}}
	client := &tse.FakeSetupClient{
		UmgebungResponse:  tse.UmgebungTest,
		TSSResponse:       []tse.TSSInfo{{ID: "tss-tot", State: "DISABLED"}},
		CreateTSSResponse: tse.TSSErstellt{ID: "tss-neu", PUK: "puk", State: "CREATED"},
	}

	_, err := commandMit(repo, client).RichteTSEEin(context.Background(), zugangsdaten(), tse.UmgebungTest, false)
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
	repo := &stubCommandRepo{identitaet: tse.Kassenidentitaet{Seriennummer: uuid.New()}}
	client := &tse.FakeSetupClient{
		UmgebungResponse:  tse.UmgebungTest,
		CreateTSSResponse: tse.TSSErstellt{ID: "tss-neu", PUK: "puk", State: "CREATED"},
		RegistriereErr:    errors.New("client registration failed"),
	}

	_, err := commandMit(repo, client).RichteTSEEin(context.Background(), zugangsdaten(), tse.UmgebungTest, false)
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
	repo := &stubCommandRepo{identitaet: tse.Kassenidentitaet{Seriennummer: uuid.New()}}
	client := &tse.FakeSetupClient{TSSErr: tse.ErrSetupAuthFehlgeschlagen}

	_, err := commandMit(repo, client).RichteTSEEin(context.Background(), zugangsdaten(), tse.UmgebungTest, false)
	if !errors.Is(err, ErrTSESetupZugangsdaten) {
		t.Fatalf("expected ErrTSESetupZugangsdaten, got %v", err)
	}
	if client.CreateTSSCalls != 0 {
		t.Fatalf("expected no TSS to be created on auth failure, got %d calls", client.CreateTSSCalls)
	}
}

// TestRichteTSEEin_NeuAnlegenTrotzVorhandenerInTest sichert F2: in TEST darf der
// Admin trotz vorhandener (hier INITIALIZED) TSS bewusst eine zweite, frische
// TSE anlegen, wenn er die Sperre per Flag uebergeht.
func TestRichteTSEEin_NeuAnlegenTrotzVorhandenerInTest(t *testing.T) {
	repo := &stubCommandRepo{identitaet: tse.Kassenidentitaet{Seriennummer: uuid.New()}}
	client := &tse.FakeSetupClient{
		UmgebungResponse:  tse.UmgebungTest,
		TSSResponse:       []tse.TSSInfo{{ID: "tss-alt", State: "INITIALIZED"}},
		CreateTSSResponse: tse.TSSErstellt{ID: "tss-neu", PUK: "puk", State: "CREATED"},
	}

	ergebnis, err := commandMit(repo, client).RichteTSEEin(context.Background(), zugangsdaten(), tse.UmgebungTest, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client.CreateTSSCalls != 1 {
		t.Fatalf("expected a new TSS to be created despite an existing one, got %d calls", client.CreateTSSCalls)
	}
	if ergebnis.TssID != "tss-neu" || repo.gespeichert == nil || repo.gespeichert.TssID != "tss-neu" {
		t.Fatalf("expected the fresh TSS to be set up and saved, got result %q saved %+v", ergebnis.TssID, repo.gespeichert)
	}
}

// TestRichteTSEEin_NeuAnlegenTrotzVorhandenerInLiveVerweigert sichert, dass das
// Flag in LIVE wirkungslos bleibt: die Sperre gegen eine zweite TSS greift hart.
func TestRichteTSEEin_NeuAnlegenTrotzVorhandenerInLiveVerweigert(t *testing.T) {
	repo := &stubCommandRepo{identitaet: tse.Kassenidentitaet{Seriennummer: uuid.New()}}
	client := &tse.FakeSetupClient{
		UmgebungResponse: tse.UmgebungLive,
		TSSResponse:      []tse.TSSInfo{{ID: "tss-live", State: "INITIALIZED"}},
	}

	_, err := commandMit(repo, client).RichteTSEEin(context.Background(), zugangsdaten(), tse.UmgebungLive, true)
	if !errors.Is(err, ErrTSEBereitsEingerichtet) {
		t.Fatalf("expected ErrTSEBereitsEingerichtet in LIVE despite the flag, got %v", err)
	}
	if client.CreateTSSCalls != 0 {
		t.Fatalf("expected no TSS to be created in LIVE, got %d calls", client.CreateTSSCalls)
	}
	if repo.gespeichert != nil {
		t.Fatal("expected no configuration to be saved in LIVE")
	}
}

// TestRichteTSEEin_TSSLimitErreicht sichert, dass das fiskaly-TSS-Limit (in TEST
// fuenf aktive TSS) als verstaendliche Meldung statt als technischer Fehler
// durchgereicht wird.
func TestRichteTSEEin_TSSLimitErreicht(t *testing.T) {
	repo := &stubCommandRepo{identitaet: tse.Kassenidentitaet{Seriennummer: uuid.New()}}
	client := &tse.FakeSetupClient{
		UmgebungResponse: tse.UmgebungTest,
		TSSResponse:      []tse.TSSInfo{{ID: "tss-alt", State: "INITIALIZED"}},
		CreateTSSErr:     tse.ErrSetupTSSLimitErreicht,
	}

	_, err := commandMit(repo, client).RichteTSEEin(context.Background(), zugangsdaten(), tse.UmgebungTest, true)
	if !errors.Is(err, ErrTSESetupTSSLimitErreicht) {
		t.Fatalf("expected ErrTSESetupTSSLimitErreicht, got %v", err)
	}
	if repo.gespeichert != nil {
		t.Fatal("expected no configuration to be saved when the TSS limit is reached")
	}
}

// TestUebernimmTSE_WiederaufnahmeCreated sichert die Wiederaufnahme nach einem
// Abbruch im Zustand CREATED: der PUK wird idempotent erneut bezogen, eine
// frische PIN erzeugt und der Lebenszyklus vollendet — ohne zweite TSS und ohne
// Nutzereingabe.
func TestUebernimmTSE_WiederaufnahmeCreated(t *testing.T) {
	seriennummer := uuid.New()
	repo := &stubCommandRepo{identitaet: tse.Kassenidentitaet{Seriennummer: seriennummer}}
	client := &tse.FakeSetupClient{
		UmgebungResponse:    tse.UmgebungTest,
		TSSResponse:         []tse.TSSInfo{{ID: "tss-halb", State: "CREATED"}},
		GetAdminPUKResponse: "puk-refetch",
	}

	ergebnis, err := commandMit(repo, client).UebernimmTSE(context.Background(), zugangsdaten(), tse.UmgebungTest, "tss-halb", "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client.CreateTSSCalls != 0 {
		t.Fatalf("expected no new TSS to be created on resume, got %d calls", client.CreateTSSCalls)
	}
	if client.GetAdminPUKCalls != 1 {
		t.Fatalf("expected the puk to be refetched once, got %d calls", client.GetAdminPUKCalls)
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
	repo := &stubCommandRepo{identitaet: tse.Kassenidentitaet{Seriennummer: seriennummer}}
	client := &tse.FakeSetupClient{
		UmgebungResponse: tse.UmgebungTest,
		TSSResponse:      []tse.TSSInfo{{ID: "tss-uninit", State: "UNINITIALIZED"}},
	}

	ergebnis, err := commandMit(repo, client).UebernimmTSE(context.Background(), zugangsdaten(), tse.UmgebungTest, "tss-uninit", "1234567890", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client.GetAdminPUKCalls != 0 {
		t.Fatalf("expected no puk refetch from UNINITIALIZED, got %d calls", client.GetAdminPUKCalls)
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
	repo := &stubCommandRepo{identitaet: tse.Kassenidentitaet{Seriennummer: seriennummer}}
	client := &tse.FakeSetupClient{
		UmgebungResponse: tse.UmgebungTest,
		TSSResponse:      []tse.TSSInfo{{ID: "tss-init", State: "INITIALIZED"}},
	}

	ergebnis, err := commandMit(repo, client).UebernimmTSE(context.Background(), zugangsdaten(), tse.UmgebungTest, "tss-init", "1234567890", "")
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
	repo := &stubCommandRepo{identitaet: tse.Kassenidentitaet{Seriennummer: seriennummer}}
	client := &tse.FakeSetupClient{
		UmgebungResponse: tse.UmgebungTest,
		TSSResponse:      []tse.TSSInfo{{ID: "tss-init", State: "INITIALIZED"}},
		ClientsByTSS: map[string][]tse.ClientInfo{
			"tss-init": {{ID: vorhandenerClient, SerialNumber: seriennummer.String(), State: "REGISTERED"}},
		},
	}

	ergebnis, err := commandMit(repo, client).UebernimmTSE(context.Background(), zugangsdaten(), tse.UmgebungTest, "tss-init", "1234567890", "")
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

// TestUebernimmTSE_EinsatzbereitOhnePIN sichert F8: eine INITIALIZED TSS mit
// bereits REGISTERED Client ist einsatzbereit. Die Uebernahme gelingt mit leerer
// PIN, ohne dass AuthentifiziereAdmin aufgerufen wird (keine fiskaly-Mutation);
// es wird nur die Konfiguration gespeichert.
func TestUebernimmTSE_EinsatzbereitOhnePIN(t *testing.T) {
	seriennummer := uuid.New()
	vorhandenerClient := uuid.NewString()
	repo := &stubCommandRepo{identitaet: tse.Kassenidentitaet{Seriennummer: seriennummer}}
	client := &tse.FakeSetupClient{
		UmgebungResponse: tse.UmgebungTest,
		TSSResponse:      []tse.TSSInfo{{ID: "tss-init", State: "INITIALIZED"}},
		ClientsByTSS: map[string][]tse.ClientInfo{
			"tss-init": {{ID: vorhandenerClient, SerialNumber: seriennummer.String(), State: "REGISTERED"}},
		},
	}

	ergebnis, err := commandMit(repo, client).UebernimmTSE(context.Background(), zugangsdaten(), tse.UmgebungTest, "tss-init", "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client.AuthAdminCalls != 0 {
		t.Fatalf("expected AuthentifiziereAdmin to be skipped for a ready TSS, got %d calls", client.AuthAdminCalls)
	}
	if len(client.RegistrierteClients) != 0 || len(client.ReaktivierteClients) != 0 {
		t.Fatalf("expected no client mutation for a ready TSS, got registered %+v reactivated %+v", client.RegistrierteClients, client.ReaktivierteClients)
	}
	if ergebnis.PUK != "" || ergebnis.AdminPIN != "" {
		t.Fatalf("expected no new secrets for a ready TSS, got %+v", ergebnis)
	}
	if ergebnis.ClientID != vorhandenerClient {
		t.Fatalf("expected the existing client to be adopted, got %q", ergebnis.ClientID)
	}
	if repo.gespeichert == nil || repo.gespeichert.ClientID != vorhandenerClient {
		t.Fatalf("expected the configuration to be saved with the existing client, got %+v", repo.gespeichert)
	}
}

// TestUebernimmTSE_DeregistrierterClientReaktiviert sichert die F7-Heilung: ein
// passender, aber DEREGISTERED Client wird mit der PIN reaktiviert (derselbe
// client_id, kein neuer Client) statt still als fertig gewertet zu werden.
func TestUebernimmTSE_DeregistrierterClientReaktiviert(t *testing.T) {
	seriennummer := uuid.New()
	vorhandenerClient := uuid.NewString()
	repo := &stubCommandRepo{identitaet: tse.Kassenidentitaet{Seriennummer: seriennummer}}
	client := &tse.FakeSetupClient{
		UmgebungResponse: tse.UmgebungTest,
		TSSResponse:      []tse.TSSInfo{{ID: "tss-init", State: "INITIALIZED"}},
		ClientsByTSS: map[string][]tse.ClientInfo{
			"tss-init": {{ID: vorhandenerClient, SerialNumber: seriennummer.String(), State: "DEREGISTERED"}},
		},
	}

	ergebnis, err := commandMit(repo, client).UebernimmTSE(context.Background(), zugangsdaten(), tse.UmgebungTest, "tss-init", "1234567890", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client.AuthAdminCalls == 0 {
		t.Fatal("expected AuthentifiziereAdmin to be called for a privileged reactivation")
	}
	if len(client.RegistrierteClients) != 0 {
		t.Fatalf("expected no new client registration for a deregistered client, got %+v", client.RegistrierteClients)
	}
	if len(client.ReaktivierteClients) != 1 || client.ReaktivierteClients[0].ClientID != vorhandenerClient {
		t.Fatalf("expected the same client to be reactivated, got %+v", client.ReaktivierteClients)
	}
	if ergebnis.ClientID != vorhandenerClient || repo.gespeichert.ClientID != vorhandenerClient {
		t.Fatalf("expected the reactivated client to be saved, got result %q saved %+v", ergebnis.ClientID, repo.gespeichert)
	}
}

// TestUebernimmTSE_DeregistrierterClientBrauchtPIN sichert, dass die
// Reaktivierung eine privilegierte Operation bleibt: ohne PIN wird sie
// abgewiesen, bevor irgendetwas geschrieben wird.
func TestUebernimmTSE_DeregistrierterClientBrauchtPIN(t *testing.T) {
	seriennummer := uuid.New()
	repo := &stubCommandRepo{identitaet: tse.Kassenidentitaet{Seriennummer: seriennummer}}
	client := &tse.FakeSetupClient{
		UmgebungResponse: tse.UmgebungTest,
		TSSResponse:      []tse.TSSInfo{{ID: "tss-init", State: "INITIALIZED"}},
		ClientsByTSS: map[string][]tse.ClientInfo{
			"tss-init": {{ID: uuid.NewString(), SerialNumber: seriennummer.String(), State: "DEREGISTERED"}},
		},
	}

	_, err := commandMit(repo, client).UebernimmTSE(context.Background(), zugangsdaten(), tse.UmgebungTest, "tss-init", "", "")
	if !errors.Is(err, ErrTSESetupPINErforderlich) {
		t.Fatalf("expected ErrTSESetupPINErforderlich, got %v", err)
	}
	if len(client.ReaktivierteClients) != 0 || repo.gespeichert != nil {
		t.Fatal("expected no writes when the pin is missing")
	}
}

// TestUebernimmTSE_InitialisiertOhneClientBrauchtPIN sichert, dass eine
// INITIALIZED TSS ohne passenden Client weiterhin die PIN verlangt (Registrierung
// ist privilegiert) — die F8-Lockerung greift nur bei fertigem Client.
func TestUebernimmTSE_InitialisiertOhneClientBrauchtPIN(t *testing.T) {
	repo := &stubCommandRepo{identitaet: tse.Kassenidentitaet{Seriennummer: uuid.New()}}
	client := &tse.FakeSetupClient{
		UmgebungResponse: tse.UmgebungTest,
		TSSResponse:      []tse.TSSInfo{{ID: "tss-init", State: "INITIALIZED"}},
	}

	_, err := commandMit(repo, client).UebernimmTSE(context.Background(), zugangsdaten(), tse.UmgebungTest, "tss-init", "", "")
	if !errors.Is(err, ErrTSESetupPINErforderlich) {
		t.Fatalf("expected ErrTSESetupPINErforderlich, got %v", err)
	}
	if len(client.RegistrierteClients) != 0 || repo.gespeichert != nil {
		t.Fatal("expected no writes when the pin is missing")
	}
}

// TestUebernimmTSE_PINErforderlich sichert, dass die Uebernahme ab UNINITIALIZED
// ohne PIN klar als fehlende PIN gemeldet wird — vor jeder Schreiboperation.
func TestUebernimmTSE_PINErforderlich(t *testing.T) {
	repo := &stubCommandRepo{identitaet: tse.Kassenidentitaet{Seriennummer: uuid.New()}}
	client := &tse.FakeSetupClient{
		UmgebungResponse: tse.UmgebungTest,
		TSSResponse:      []tse.TSSInfo{{ID: "tss-uninit", State: "UNINITIALIZED"}},
	}

	_, err := commandMit(repo, client).UebernimmTSE(context.Background(), zugangsdaten(), tse.UmgebungTest, "tss-uninit", "", "")
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
	repo := &stubCommandRepo{identitaet: tse.Kassenidentitaet{Seriennummer: uuid.New()}}
	client := &tse.FakeSetupClient{
		UmgebungResponse: tse.UmgebungTest,
		TSSResponse:      []tse.TSSInfo{{ID: "tss-init", State: "INITIALIZED"}},
		AuthAdminErr:     tse.ErrSetupAuthFehlgeschlagen,
	}

	_, err := commandMit(repo, client).UebernimmTSE(context.Background(), zugangsdaten(), tse.UmgebungTest, "tss-init", "0000000000", "")
	if !errors.Is(err, ErrTSESetupPINUnbekannt) {
		t.Fatalf("expected ErrTSESetupPINUnbekannt, got %v", err)
	}
	if repo.gespeichert != nil {
		t.Fatal("expected no configuration to be saved on an unknown pin")
	}
}

// TestUebernimmTSE_PINResetPerPUK sichert den PUK-Reset: mit dem verwahrten
// Admin-PUK setzt jotti eine frische Zufalls-PIN, schliesst damit die Uebernahme
// ab und zeigt die neue PIN einmalig an. Der PUK wird dabei verwendet, bleibt
// aber unveraendert und wird nicht erneut zurueckgegeben.
func TestUebernimmTSE_PINResetPerPUK(t *testing.T) {
	seriennummer := uuid.New()
	repo := &stubCommandRepo{identitaet: tse.Kassenidentitaet{Seriennummer: seriennummer}}
	client := &tse.FakeSetupClient{
		UmgebungResponse: tse.UmgebungTest,
		TSSResponse:      []tse.TSSInfo{{ID: "tss-init", State: "INITIALIZED"}},
	}

	ergebnis, err := commandMit(repo, client).UebernimmTSE(context.Background(), zugangsdaten(), tse.UmgebungTest, "tss-init", "", "puk-verwahrt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client.GesetzterAdminPUK != "puk-verwahrt" {
		t.Fatalf("expected the supplied puk to be used for the reset, got %q", client.GesetzterAdminPUK)
	}
	if ergebnis.AdminPIN == "" || ergebnis.AdminPIN != client.GesetzteAdminPIN {
		t.Fatalf("expected a fresh pin to be set and returned, got result %q set %q", ergebnis.AdminPIN, client.GesetzteAdminPIN)
	}
	// Der PUK aendert sich beim Reset nicht und wird nicht erneut angezeigt.
	if ergebnis.PUK != "" {
		t.Fatalf("expected no puk to be returned on a reset, got %q", ergebnis.PUK)
	}
	// Die frische PIN treibt den Rest der Uebernahme (Admin-Auth + Client).
	if client.AuthentifiziertePIN != ergebnis.AdminPIN {
		t.Fatalf("expected the fresh pin to be used for admin auth, got %q", client.AuthentifiziertePIN)
	}
	if len(client.RegistrierteClients) != 1 || client.RegistrierteClients[0].SerialNumber != seriennummer.String() {
		t.Fatalf("expected the client to be registered with the kassen serial, got %+v", client.RegistrierteClients)
	}
	if repo.gespeichert == nil || repo.gespeichert.TssID != "tss-init" {
		t.Fatalf("expected the configuration to be saved, got %+v", repo.gespeichert)
	}
}

// TestUebernimmTSE_PINResetPerPUKInLive sichert, dass der PUK-Reset auch in LIVE
// funktioniert (keine neue, kostenpflichtige TSS) — hier aus UNINITIALIZED, sodass
// zusaetzlich initialisiert wird.
func TestUebernimmTSE_PINResetPerPUKInLive(t *testing.T) {
	seriennummer := uuid.New()
	repo := &stubCommandRepo{identitaet: tse.Kassenidentitaet{Seriennummer: seriennummer}}
	client := &tse.FakeSetupClient{
		UmgebungResponse: tse.UmgebungLive,
		TSSResponse:      []tse.TSSInfo{{ID: "tss-uninit", State: "UNINITIALIZED"}},
	}

	ergebnis, err := commandMit(repo, client).UebernimmTSE(context.Background(), zugangsdaten(), tse.UmgebungLive, "tss-uninit", "", "puk-verwahrt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client.GesetzterAdminPUK != "puk-verwahrt" || ergebnis.AdminPIN == "" {
		t.Fatalf("expected a puk-based pin reset in LIVE, got puk %q pin %q", client.GesetzterAdminPUK, ergebnis.AdminPIN)
	}
	if ergebnis.PUK != "" {
		t.Fatalf("expected no puk to be returned on a reset, got %q", ergebnis.PUK)
	}
	if len(client.RegistrierteClients) != 1 || repo.gespeichert == nil {
		t.Fatalf("expected the takeover to complete and save, got clients %+v saved %+v", client.RegistrierteClients, repo.gespeichert)
	}
}

// TestUebernimmTSE_PINResetFalscherPUK sichert, dass ein von fiskaly abgelehnter
// PUK als ErrTSESetupPUKUnbekannt endet — vor jeder weiteren Operation und ohne
// Speicherung, nicht als technischer Fehler.
func TestUebernimmTSE_PINResetFalscherPUK(t *testing.T) {
	repo := &stubCommandRepo{identitaet: tse.Kassenidentitaet{Seriennummer: uuid.New()}}
	client := &tse.FakeSetupClient{
		UmgebungResponse: tse.UmgebungTest,
		TSSResponse:      []tse.TSSInfo{{ID: "tss-init", State: "INITIALIZED"}},
		SetAdminPINErr:   tse.ErrSetupAuthFehlgeschlagen,
	}

	_, err := commandMit(repo, client).UebernimmTSE(context.Background(), zugangsdaten(), tse.UmgebungTest, "tss-init", "", "puk-falsch")
	if !errors.Is(err, ErrTSESetupPUKUnbekannt) {
		t.Fatalf("expected ErrTSESetupPUKUnbekannt, got %v", err)
	}
	if client.AuthAdminCalls != 0 || len(client.RegistrierteClients) != 0 || repo.gespeichert != nil {
		t.Fatal("expected no further operations or writes on a wrong puk")
	}
}

// TestUebernimmTSE_UmgebungAbweichung sichert, dass der LIVE-Schutz auch bei der
// Uebernahme greift: weicht die tatsaechliche Umgebung von der bestaetigten ab,
// bricht der Flow vor jeder Operation ab.
func TestUebernimmTSE_UmgebungAbweichung(t *testing.T) {
	repo := &stubCommandRepo{identitaet: tse.Kassenidentitaet{Seriennummer: uuid.New()}}
	client := &tse.FakeSetupClient{
		UmgebungResponse: tse.UmgebungLive,
		TSSResponse:      []tse.TSSInfo{{ID: "tss-x", State: "CREATED"}},
	}

	_, err := commandMit(repo, client).UebernimmTSE(context.Background(), zugangsdaten(), tse.UmgebungTest, "tss-x", "", "")
	if !errors.Is(err, ErrTSESetupUmgebungAbweichung) {
		t.Fatalf("expected ErrTSESetupUmgebungAbweichung, got %v", err)
	}
	if client.GetAdminPUKCalls != 0 || repo.gespeichert != nil {
		t.Fatal("expected no operations on environment mismatch")
	}
}

// TestUebernimmTSE_TSSNichtGefunden sichert, dass eine im Konto fehlende TSS klar
// gemeldet wird.
func TestUebernimmTSE_TSSNichtGefunden(t *testing.T) {
	repo := &stubCommandRepo{identitaet: tse.Kassenidentitaet{Seriennummer: uuid.New()}}
	client := &tse.FakeSetupClient{UmgebungResponse: tse.UmgebungTest}

	_, err := commandMit(repo, client).UebernimmTSE(context.Background(), zugangsdaten(), tse.UmgebungTest, "tss-fehlt", "", "")
	if !errors.Is(err, ErrTSESetupTSSNichtGefunden) {
		t.Fatalf("expected ErrTSESetupTSSNichtGefunden, got %v", err)
	}
}

// TestUebernimmTSE_DeaktivierteTSS sichert, dass eine deaktivierte TSS nicht
// uebernommen werden kann.
func TestUebernimmTSE_DeaktivierteTSS(t *testing.T) {
	repo := &stubCommandRepo{identitaet: tse.Kassenidentitaet{Seriennummer: uuid.New()}}
	client := &tse.FakeSetupClient{
		UmgebungResponse: tse.UmgebungTest,
		TSSResponse:      []tse.TSSInfo{{ID: "tss-tot", State: "DISABLED"}},
	}

	_, err := commandMit(repo, client).UebernimmTSE(context.Background(), zugangsdaten(), tse.UmgebungTest, "tss-tot", "1234567890", "")
	if !errors.Is(err, ErrTSESetupUebernahmeNichtMoeglich) {
		t.Fatalf("expected ErrTSESetupUebernahmeNichtMoeglich, got %v", err)
	}
	if repo.gespeichert != nil {
		t.Fatal("expected no configuration to be saved for a disabled TSS")
	}
}

// stammdatenAntwort ist die fiskaly-Stammdaten-Antwort fuer die
// Persistenz-Tests des DSFinV-K-Exports.
func stammdatenAntwort() tse.TSSStammdaten {
	return tse.TSSStammdaten{
		SignaturAlgorithmus: "ecdsa-plain-SHA256",
		PublicKey:           "public-key-b64",
		Zertifikat:          "certificate-b64",
		LogTimeFormat:       "unixTime",
	}
}

// checkStammdaten vergleicht die gespeicherten Stammdaten mit der erwarteten
// fiskaly-Antwort (ohne den serverseitig gesetzten Zeitstempel).
func checkStammdaten(t *testing.T, gespeichert *tse.Stammdaten, erwartet tse.TSSStammdaten) {
	t.Helper()
	if gespeichert == nil {
		t.Fatal("expected the tse stammdaten to be persisted")
	}
	if gespeichert.SignaturAlgorithmus != erwartet.SignaturAlgorithmus ||
		gespeichert.PublicKey != erwartet.PublicKey ||
		gespeichert.Zertifikat != erwartet.Zertifikat ||
		gespeichert.LogTimeFormat != erwartet.LogTimeFormat {
		t.Fatalf("persisted stammdaten do not match the fiskaly response, got %+v", gespeichert)
	}
}

// TestRichteTSEEin_PersistiertStammdaten sichert, dass nach erfolgreicher
// Neuanlage die fiskalischen TSS-Stammdaten (Algorithmus, Public Key, Zertifikat,
// Log-Time-Format) fuer den DSFinV-K-Export gespeichert werden.
func TestRichteTSEEin_PersistiertStammdaten(t *testing.T) {
	repo := &stubCommandRepo{identitaet: tse.Kassenidentitaet{Seriennummer: uuid.New()}}
	client := &tse.FakeSetupClient{
		UmgebungResponse:   tse.UmgebungTest,
		CreateTSSResponse:  tse.TSSErstellt{ID: "tss-neu", PUK: "puk", State: "CREATED"},
		StammdatenResponse: stammdatenAntwort(),
	}

	_, err := commandMit(repo, client).RichteTSEEin(context.Background(), zugangsdaten(), tse.UmgebungTest, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client.StammdatenCalls != 1 || client.StammdatenTssID != "tss-neu" {
		t.Fatalf("expected stammdaten to be fetched once for the new TSS, got %d calls for %q", client.StammdatenCalls, client.StammdatenTssID)
	}
	checkStammdaten(t, repo.gespeicherteStammdaten, stammdatenAntwort())
}

// TestUebernimmTSE_EinsatzbereitPersistiertStammdaten sichert, dass die
// Stammdaten-Persistenz am gemeinsamen Speicher-Schritt haengt, nicht am
// Anlage-Lebenszyklus: selbst die F8-Uebernahme einer einsatzbereiten TSS (ohne
// jede privilegierte fiskaly-Operation) zieht die Stammdaten nach.
func TestUebernimmTSE_EinsatzbereitPersistiertStammdaten(t *testing.T) {
	seriennummer := uuid.New()
	vorhandenerClient := uuid.NewString()
	repo := &stubCommandRepo{identitaet: tse.Kassenidentitaet{Seriennummer: seriennummer}}
	client := &tse.FakeSetupClient{
		UmgebungResponse: tse.UmgebungTest,
		TSSResponse:      []tse.TSSInfo{{ID: "tss-init", State: "INITIALIZED"}},
		ClientsByTSS: map[string][]tse.ClientInfo{
			"tss-init": {{ID: vorhandenerClient, SerialNumber: seriennummer.String(), State: "REGISTERED"}},
		},
		StammdatenResponse: stammdatenAntwort(),
	}

	_, err := commandMit(repo, client).UebernimmTSE(context.Background(), zugangsdaten(), tse.UmgebungTest, "tss-init", "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client.AuthAdminCalls != 0 {
		t.Fatalf("expected no lifecycle operations for a ready TSS, got %d admin auth calls", client.AuthAdminCalls)
	}
	if client.StammdatenCalls != 1 || client.StammdatenTssID != "tss-init" {
		t.Fatalf("expected stammdaten to be fetched once for the adopted TSS, got %d calls for %q", client.StammdatenCalls, client.StammdatenTssID)
	}
	checkStammdaten(t, repo.gespeicherteStammdaten, stammdatenAntwort())
}

// TestUebernimmTSE_PINResetPersistiertStammdaten sichert, dass auch der
// PUK-Reset-Pfad die Stammdaten nachzieht.
func TestUebernimmTSE_PINResetPersistiertStammdaten(t *testing.T) {
	repo := &stubCommandRepo{identitaet: tse.Kassenidentitaet{Seriennummer: uuid.New()}}
	client := &tse.FakeSetupClient{
		UmgebungResponse:   tse.UmgebungTest,
		TSSResponse:        []tse.TSSInfo{{ID: "tss-init", State: "INITIALIZED"}},
		StammdatenResponse: stammdatenAntwort(),
	}

	_, err := commandMit(repo, client).UebernimmTSE(context.Background(), zugangsdaten(), tse.UmgebungTest, "tss-init", "", "puk-verwahrt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	checkStammdaten(t, repo.gespeicherteStammdaten, stammdatenAntwort())
}

// TestRichteTSEEin_StammdatenAbrufFehlerKipptSetupNicht sichert die
// Best-Effort-Semantik: schlaegt der Stammdaten-Abruf fehl, bleibt die
// Einrichtung erfolgreich und die Konfiguration gespeichert — nur die Stammdaten
// fehlen (beim naechsten Verbinden nachziehbar).
func TestRichteTSEEin_StammdatenAbrufFehlerKipptSetupNicht(t *testing.T) {
	repo := &stubCommandRepo{identitaet: tse.Kassenidentitaet{Seriennummer: uuid.New()}}
	client := &tse.FakeSetupClient{
		UmgebungResponse:  tse.UmgebungTest,
		CreateTSSResponse: tse.TSSErstellt{ID: "tss-neu", PUK: "puk", State: "CREATED"},
		StammdatenErr:     errors.New("fiskaly stammdaten read failed"),
	}

	ergebnis, err := commandMit(repo, client).RichteTSEEin(context.Background(), zugangsdaten(), tse.UmgebungTest, false)
	if err != nil {
		t.Fatalf("expected setup to succeed despite a stammdaten fetch failure, got %v", err)
	}
	if ergebnis.TssID != "tss-neu" || repo.gespeichert == nil {
		t.Fatalf("expected the configuration to be saved, got result %q saved %+v", ergebnis.TssID, repo.gespeichert)
	}
	if repo.gespeicherteStammdaten != nil {
		t.Fatal("expected no stammdaten to be saved when the fetch fails")
	}
}

// commandMitOffenerKassensitzung baut ein Command, dessen Konfigurations-Guard
// eine offene Kassensitzung sieht. Der Setup-Client wuerde beim Aufruf failen —
// so belegt der Test, dass der Guard vor jeder fiskaly-Arbeit greift.
func commandMitOffenerKassensitzung(repo *stubCommandRepo) Command {
	return Command{
		TSERepo:             repo,
		KassensitzungenRepo: stubKassensitzungReader{offene: &kasse.Kassensitzung{ZNr: 1, Status: kasse.KassensitzungOffen}},
		NewTSESetupClient: func(tse.SetupCredentials) (tse.SetupClient, error) {
			return nil, errors.New("setup client must not be created while a Kassensitzung is open")
		},
	}
}

// TSE-Konfigurationsaenderungen sind bei offener Kassensitzung nicht erlaubt:
// Das Signaturgeraet darf nicht mitten in einem laufenden Kassentag wechseln.
// Alle drei Aenderungspfade lehnen mit ErrTSEKonfigurationKassensitzungOffen ab
// und schreiben nichts.
func TestUpdateTSEKonfiguration_MitOffenerKassensitzungAbgelehnt(t *testing.T) {
	repo := &stubCommandRepo{}
	conf, err := tse.NewKonfiguration("api-key", "api-secret", "tss-1", "client-1")
	if err != nil {
		t.Fatalf("unexpected error building konfiguration: %v", err)
	}

	err = commandMitOffenerKassensitzung(repo).UpdateTSEKonfiguration(context.Background(), conf)
	if !errors.Is(err, ErrTSEKonfigurationKassensitzungOffen) {
		t.Fatalf("expected ErrTSEKonfigurationKassensitzungOffen, got %v", err)
	}
	if repo.gespeichert != nil {
		t.Fatalf("expected no configuration to be saved, got %+v", repo.gespeichert)
	}
}

func TestRichteTSEEin_MitOffenerKassensitzungAbgelehnt(t *testing.T) {
	repo := &stubCommandRepo{identitaet: tse.Kassenidentitaet{Seriennummer: uuid.New()}}

	_, err := commandMitOffenerKassensitzung(repo).RichteTSEEin(context.Background(), zugangsdaten(), tse.UmgebungTest, false)
	if !errors.Is(err, ErrTSEKonfigurationKassensitzungOffen) {
		t.Fatalf("expected ErrTSEKonfigurationKassensitzungOffen, got %v", err)
	}
	if repo.gespeichert != nil {
		t.Fatalf("expected no configuration to be saved, got %+v", repo.gespeichert)
	}
}

func TestUebernimmTSE_MitOffenerKassensitzungAbgelehnt(t *testing.T) {
	repo := &stubCommandRepo{identitaet: tse.Kassenidentitaet{Seriennummer: uuid.New()}}

	_, err := commandMitOffenerKassensitzung(repo).UebernimmTSE(context.Background(), zugangsdaten(), tse.UmgebungTest, "tss-1", "", "")
	if !errors.Is(err, ErrTSEKonfigurationKassensitzungOffen) {
		t.Fatalf("expected ErrTSEKonfigurationKassensitzungOffen, got %v", err)
	}
	if repo.gespeichert != nil {
		t.Fatalf("expected no configuration to be saved, got %+v", repo.gespeichert)
	}
}
