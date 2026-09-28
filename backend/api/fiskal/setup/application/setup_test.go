package application

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/nicograef/jotti/backend/domain/kasse"
	"github.com/nicograef/jotti/backend/domain/tse"
	"github.com/nicograef/jotti/backend/domain/tse/tsetest"
)

// stubCommandRepo records the saved configuration and returns a fixed Kassenidentitaet.
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

// A nil aktive (the default) means no Kassensitzung is active.
type stubKassensitzungReader struct {
	aktive *kasse.Kassensitzung
	err    error
}

func (s stubKassensitzungReader) GetAktiveKassensitzung(context.Context) (*kasse.Kassensitzung, error) {
	return s.aktive, s.err
}

func commandMit(repo *stubCommandRepo, client *tsetest.FakeSetupClient) Command {
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

// Full run: an empty account ends with an initialised TSS whose client carries the Kassen-Seriennummer.
func TestRichteTSEEin_LeeresKonto(t *testing.T) {
	seriennummer := uuid.New()
	repo := &stubCommandRepo{identitaet: tse.Kassenidentitaet{Seriennummer: seriennummer}}
	client := &tsetest.FakeSetupClient{
		UmgebungResponse:  tse.UmgebungTest,
		CreateTSSResponse: tse.TSSErstellt{ID: "tss-neu", PUK: "puk-123", State: "CREATED"},
	}

	ergebnis, err := commandMit(repo, client).RichteTSEEin(context.Background(), zugangsdaten(), tse.UmgebungTest, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ergebnis.TssID != "tss-neu" {
		t.Errorf("expected tss id tss-neu, got %q", ergebnis.TssID)
	}
	if ergebnis.PUK != "puk-123" {
		t.Errorf("expected puk to be returned, got %q", ergebnis.PUK)
	}
	if ergebnis.AdminPIN == "" {
		t.Error("expected an admin pin to be returned")
	}
	// The client _id is its own UUIDv4 (fiskaly convention), not the Kassen-Seriennummer.
	if ergebnis.ClientID == "" || ergebnis.ClientID == seriennummer.String() {
		t.Errorf("expected a distinct generated client id, got %q", ergebnis.ClientID)
	}

	if len(client.RegistrierteClients) != 1 {
		t.Fatalf("expected exactly one registered client, got %d", len(client.RegistrierteClients))
	}
	registriert := client.RegistrierteClients[0]
	if registriert.SerialNumber != seriennummer.String() {
		t.Errorf("expected client registered with kassen serial number as serial_number, got %+v", registriert)
	}
	if registriert.ClientID != ergebnis.ClientID {
		t.Errorf("expected client registered under the returned client id %q, got %q", ergebnis.ClientID, registriert.ClientID)
	}
	if client.GesetzteAdminPIN != ergebnis.AdminPIN {
		t.Errorf("expected the generated pin to be set on the TSS, got %q vs %q", client.GesetzteAdminPIN, ergebnis.AdminPIN)
	}

	if repo.gespeichert == nil {
		t.Fatal("expected the configuration to be saved")
	}
	if repo.gespeichert.TssID != "tss-neu" || repo.gespeichert.ClientID != ergebnis.ClientID {
		t.Errorf("expected full configuration to be saved, got %+v", repo.gespeichert)
	}
	if !repo.gespeichert.IstKonfiguriert() {
		t.Error("expected the saved configuration to be complete")
	}
}

// LIVE guard: confirmed TEST with LIVE credentials aborts before any TSS is created.
func TestRichteTSEEin_UmgebungAbweichung(t *testing.T) {
	repo := &stubCommandRepo{identitaet: tse.Kassenidentitaet{Seriennummer: uuid.New()}}
	client := &tsetest.FakeSetupClient{UmgebungResponse: tse.UmgebungLive}

	_, err := commandMit(repo, client).RichteTSEEin(context.Background(), zugangsdaten(), tse.UmgebungTest, false)
	if !errors.Is(err, ErrTSESetupUmgebungAbweichung) {
		t.Errorf("expected ErrTSESetupUmgebungAbweichung, got %v", err)
	}
	if len(client.ErstellteTSS) != 0 {
		t.Errorf("expected no TSS to be created on environment mismatch, got %+v", client.ErstellteTSS)
	}
	if repo.gespeichert != nil {
		t.Error("expected no configuration to be saved on environment mismatch")
	}
}

func TestRichteTSEEin_BestaetigteUmgebungUngueltig(t *testing.T) {
	repo := &stubCommandRepo{identitaet: tse.Kassenidentitaet{Seriennummer: uuid.New()}}
	client := &tsetest.FakeSetupClient{UmgebungResponse: tse.UmgebungTest}

	_, err := commandMit(repo, client).RichteTSEEin(context.Background(), zugangsdaten(), tse.Umgebung(""), false)
	if !errors.Is(err, ErrTSESetupUmgebungAbweichung) {
		t.Errorf("expected ErrTSESetupUmgebungAbweichung for an unconfirmed environment, got %v", err)
	}
	if len(client.ErstellteTSS) != 0 {
		t.Errorf("expected no TSS to be created, got %+v", client.ErstellteTSS)
	}
}

func TestRichteTSEEin_VorhandeneAktiveTSS(t *testing.T) {
	repo := &stubCommandRepo{identitaet: tse.Kassenidentitaet{Seriennummer: uuid.New()}}
	client := &tsetest.FakeSetupClient{
		UmgebungResponse: tse.UmgebungTest,
		TSSResponse:      []tse.TSSInfo{{ID: "tss-alt", State: "INITIALIZED"}},
	}

	_, err := commandMit(repo, client).RichteTSEEin(context.Background(), zugangsdaten(), tse.UmgebungTest, false)
	if !errors.Is(err, ErrTSEBereitsEingerichtet) {
		t.Errorf("expected ErrTSEBereitsEingerichtet, got %v", err)
	}
	if len(client.ErstellteTSS) != 0 {
		t.Errorf("expected no TSS to be created when an active TSS exists, got %+v", client.ErstellteTSS)
	}
	if repo.gespeichert != nil {
		t.Error("expected no configuration to be saved when an active TSS exists")
	}
}

func TestRichteTSEEin_DeaktivierteTSSBlocktNicht(t *testing.T) {
	seriennummer := uuid.New()
	repo := &stubCommandRepo{identitaet: tse.Kassenidentitaet{Seriennummer: seriennummer}}
	client := &tsetest.FakeSetupClient{
		UmgebungResponse:  tse.UmgebungTest,
		TSSResponse:       []tse.TSSInfo{{ID: "tss-tot", State: "DISABLED"}},
		CreateTSSResponse: tse.TSSErstellt{ID: "tss-neu", PUK: "puk", State: "CREATED"},
	}

	_, err := commandMit(repo, client).RichteTSEEin(context.Background(), zugangsdaten(), tse.UmgebungTest, false)
	if err != nil {
		t.Fatalf("unexpected error with only a disabled TSS present: %v", err)
	}
	if len(client.ErstellteTSS) != 1 {
		t.Errorf("expected a new TSS to be created, got %+v", client.ErstellteTSS)
	}
}

// A failed lifecycle step must leave no half configuration in the database.
func TestRichteTSEEin_AbbruchSpeichertNicht(t *testing.T) {
	repo := &stubCommandRepo{identitaet: tse.Kassenidentitaet{Seriennummer: uuid.New()}}
	client := &tsetest.FakeSetupClient{
		UmgebungResponse:  tse.UmgebungTest,
		CreateTSSResponse: tse.TSSErstellt{ID: "tss-neu", PUK: "puk", State: "CREATED"},
		RegistriereErr:    errors.New("client registration failed"),
	}

	_, err := commandMit(repo, client).RichteTSEEin(context.Background(), zugangsdaten(), tse.UmgebungTest, false)
	if !errors.Is(err, ErrTSEEinrichtung) {
		t.Errorf("expected ErrTSEEinrichtung on a failing step, got %v", err)
	}
	if repo.gespeichert != nil {
		t.Error("expected no configuration to be saved when a step fails")
	}
}

func TestRichteTSEEin_FalscheZugangsdaten(t *testing.T) {
	repo := &stubCommandRepo{identitaet: tse.Kassenidentitaet{Seriennummer: uuid.New()}}
	client := &tsetest.FakeSetupClient{TSSErr: tse.ErrSetupAuthFehlgeschlagen}

	_, err := commandMit(repo, client).RichteTSEEin(context.Background(), zugangsdaten(), tse.UmgebungTest, false)
	if !errors.Is(err, ErrTSESetupZugangsdaten) {
		t.Errorf("expected ErrTSESetupZugangsdaten, got %v", err)
	}
	if len(client.ErstellteTSS) != 0 {
		t.Errorf("expected no TSS to be created on auth failure, got %+v", client.ErstellteTSS)
	}
}

// In TEST the flag deliberately creates a second TSS next to an INITIALIZED one.
func TestRichteTSEEin_NeuAnlegenTrotzVorhandenerInTest(t *testing.T) {
	repo := &stubCommandRepo{identitaet: tse.Kassenidentitaet{Seriennummer: uuid.New()}}
	client := &tsetest.FakeSetupClient{
		UmgebungResponse:  tse.UmgebungTest,
		TSSResponse:       []tse.TSSInfo{{ID: "tss-alt", State: "INITIALIZED"}},
		CreateTSSResponse: tse.TSSErstellt{ID: "tss-neu", PUK: "puk", State: "CREATED"},
	}

	ergebnis, err := commandMit(repo, client).RichteTSEEin(context.Background(), zugangsdaten(), tse.UmgebungTest, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(client.ErstellteTSS) != 1 {
		t.Errorf("expected a new TSS to be created despite an existing one, got %+v", client.ErstellteTSS)
	}
	if ergebnis.TssID != "tss-neu" || repo.gespeichert == nil || repo.gespeichert.TssID != "tss-neu" {
		t.Errorf("expected the fresh TSS to be set up and saved, got result %q saved %+v", ergebnis.TssID, repo.gespeichert)
	}
}

// In LIVE the flag has no effect: a second TSS would incur ongoing cost.
func TestRichteTSEEin_NeuAnlegenTrotzVorhandenerInLiveVerweigert(t *testing.T) {
	repo := &stubCommandRepo{identitaet: tse.Kassenidentitaet{Seriennummer: uuid.New()}}
	client := &tsetest.FakeSetupClient{
		UmgebungResponse: tse.UmgebungLive,
		TSSResponse:      []tse.TSSInfo{{ID: "tss-live", State: "INITIALIZED"}},
	}

	_, err := commandMit(repo, client).RichteTSEEin(context.Background(), zugangsdaten(), tse.UmgebungLive, true)
	if !errors.Is(err, ErrTSEBereitsEingerichtet) {
		t.Errorf("expected ErrTSEBereitsEingerichtet in LIVE despite the flag, got %v", err)
	}
	if len(client.ErstellteTSS) != 0 {
		t.Errorf("expected no TSS to be created in LIVE, got %+v", client.ErstellteTSS)
	}
	if repo.gespeichert != nil {
		t.Error("expected no configuration to be saved in LIVE")
	}
}

// fiskaly's TEST limit of five active TSS surfaces as a user-facing state, not a technical error.
func TestRichteTSEEin_TSSLimitErreicht(t *testing.T) {
	repo := &stubCommandRepo{identitaet: tse.Kassenidentitaet{Seriennummer: uuid.New()}}
	client := &tsetest.FakeSetupClient{
		UmgebungResponse: tse.UmgebungTest,
		TSSResponse:      []tse.TSSInfo{{ID: "tss-alt", State: "INITIALIZED"}},
		CreateTSSErr:     tse.ErrSetupTSSLimitErreicht,
	}

	_, err := commandMit(repo, client).RichteTSEEin(context.Background(), zugangsdaten(), tse.UmgebungTest, true)
	if !errors.Is(err, ErrTSESetupTSSLimitErreicht) {
		t.Errorf("expected ErrTSESetupTSSLimitErreicht, got %v", err)
	}
	if repo.gespeichert != nil {
		t.Error("expected no configuration to be saved when the TSS limit is reached")
	}
}

// Resuming from CREATED refetches the PUK and sets a fresh PIN, without user input or a second TSS.
func TestUebernimmTSE_WiederaufnahmeCreated(t *testing.T) {
	seriennummer := uuid.New()
	repo := &stubCommandRepo{identitaet: tse.Kassenidentitaet{Seriennummer: seriennummer}}
	client := &tsetest.FakeSetupClient{
		UmgebungResponse:    tse.UmgebungTest,
		TSSResponse:         []tse.TSSInfo{{ID: "tss-halb", State: "CREATED"}},
		GetAdminPUKResponse: "puk-refetch",
	}

	ergebnis, err := commandMit(repo, client).UebernimmTSE(context.Background(), zugangsdaten(), tse.UmgebungTest, "tss-halb", "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(client.ErstellteTSS) != 0 {
		t.Errorf("expected no new TSS to be created on resume, got %+v", client.ErstellteTSS)
	}
	if client.GesetzterAdminPUK != "puk-refetch" {
		t.Errorf("expected the pin to be set with the refetched puk, got %q", client.GesetzterAdminPUK)
	}
	if ergebnis.PUK != "puk-refetch" || ergebnis.AdminPIN == "" {
		t.Errorf("expected refetched puk and a fresh pin, got %+v", ergebnis)
	}
	if client.GesetzteAdminPIN != ergebnis.AdminPIN {
		t.Errorf("expected the fresh pin to be set on the TSS, got %q vs %q", client.GesetzteAdminPIN, ergebnis.AdminPIN)
	}
	if len(client.RegistrierteClients) != 1 || client.RegistrierteClients[0].SerialNumber != seriennummer.String() {
		t.Errorf("expected exactly one client registered with the kassen serial, got %+v", client.RegistrierteClients)
	}
	if repo.gespeichert == nil || repo.gespeichert.TssID != "tss-halb" {
		t.Errorf("expected the configuration to be saved for the resumed TSS, got %+v", repo.gespeichert)
	}
}

// Resuming from UNINITIALIZED uses the stored PIN and creates no new secrets.
func TestUebernimmTSE_WiederaufnahmeUninitialized(t *testing.T) {
	seriennummer := uuid.New()
	repo := &stubCommandRepo{identitaet: tse.Kassenidentitaet{Seriennummer: seriennummer}}
	client := &tsetest.FakeSetupClient{
		UmgebungResponse: tse.UmgebungTest,
		TSSResponse:      []tse.TSSInfo{{ID: "tss-uninit", State: "UNINITIALIZED"}},
	}

	ergebnis, err := commandMit(repo, client).UebernimmTSE(context.Background(), zugangsdaten(), tse.UmgebungTest, "tss-uninit", "1234567890", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client.GesetzteAdminPIN != "" {
		t.Errorf("expected no new admin pin from UNINITIALIZED, got %q", client.GesetzteAdminPIN)
	}
	if client.AuthentifiziertePIN != "1234567890" {
		t.Errorf("expected the entered pin to be used for admin auth, got %q", client.AuthentifiziertePIN)
	}
	if ergebnis.PUK != "" || ergebnis.AdminPIN != "" {
		t.Errorf("expected no new secrets on resume from UNINITIALIZED, got %+v", ergebnis)
	}
	if len(client.RegistrierteClients) != 1 {
		t.Errorf("expected exactly one client registered, got %d", len(client.RegistrierteClients))
	}
	if repo.gespeichert == nil || repo.gespeichert.TssID != "tss-uninit" {
		t.Errorf("expected the configuration to be saved, got %+v", repo.gespeichert)
	}
}

func TestUebernimmTSE_InitialisiertOhneClient(t *testing.T) {
	seriennummer := uuid.New()
	repo := &stubCommandRepo{identitaet: tse.Kassenidentitaet{Seriennummer: seriennummer}}
	client := &tsetest.FakeSetupClient{
		UmgebungResponse: tse.UmgebungTest,
		TSSResponse:      []tse.TSSInfo{{ID: "tss-init", State: "INITIALIZED"}},
	}

	ergebnis, err := commandMit(repo, client).UebernimmTSE(context.Background(), zugangsdaten(), tse.UmgebungTest, "tss-init", "1234567890", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(client.RegistrierteClients) != 1 || client.RegistrierteClients[0].SerialNumber != seriennummer.String() {
		t.Errorf("expected the client to be registered with the kassen serial, got %+v", client.RegistrierteClients)
	}
	if repo.gespeichert == nil || repo.gespeichert.ClientID != ergebnis.ClientID {
		t.Errorf("expected the configuration to be saved with the registered client, got %+v", repo.gespeichert)
	}
}

func TestUebernimmTSE_VorhandenerPassenderClient(t *testing.T) {
	seriennummer := uuid.New()
	vorhandenerClient := uuid.NewString()
	repo := &stubCommandRepo{identitaet: tse.Kassenidentitaet{Seriennummer: seriennummer}}
	client := &tsetest.FakeSetupClient{
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
		t.Errorf("expected no new client registration when a matching client exists, got %+v", client.RegistrierteClients)
	}
	if repo.gespeichert == nil {
		t.Fatal("expected the configuration to be saved")
	}
	if ergebnis.ClientID != vorhandenerClient || repo.gespeichert.ClientID != vorhandenerClient {
		t.Errorf("expected the existing client to be adopted, got result %q saved %q", ergebnis.ClientID, repo.gespeichert.ClientID)
	}
}

// INITIALIZED with a REGISTERED client needs no PIN and no AuthentifiziereAdmin: no fiskaly
// mutation follows.
func TestUebernimmTSE_EinsatzbereitOhnePIN(t *testing.T) {
	seriennummer := uuid.New()
	vorhandenerClient := uuid.NewString()
	repo := &stubCommandRepo{identitaet: tse.Kassenidentitaet{Seriennummer: seriennummer}}
	client := &tsetest.FakeSetupClient{
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
	if client.AdminAuthentifiziert {
		t.Error("expected AuthentifiziereAdmin to be skipped for a ready TSS")
	}
	if len(client.RegistrierteClients) != 0 || len(client.ReaktivierteClients) != 0 {
		t.Errorf("expected no client mutation for a ready TSS, got registered %+v reactivated %+v", client.RegistrierteClients, client.ReaktivierteClients)
	}
	if ergebnis.PUK != "" || ergebnis.AdminPIN != "" {
		t.Errorf("expected no new secrets for a ready TSS, got %+v", ergebnis)
	}
	if ergebnis.ClientID != vorhandenerClient {
		t.Errorf("expected the existing client to be adopted, got %q", ergebnis.ClientID)
	}
	if repo.gespeichert == nil || repo.gespeichert.ClientID != vorhandenerClient {
		t.Errorf("expected the configuration to be saved with the existing client, got %+v", repo.gespeichert)
	}
}

// A matching DEREGISTERED client is reactivated under the same client_id, not treated as done.
func TestUebernimmTSE_DeregistrierterClientReaktiviert(t *testing.T) {
	seriennummer := uuid.New()
	vorhandenerClient := uuid.NewString()
	repo := &stubCommandRepo{identitaet: tse.Kassenidentitaet{Seriennummer: seriennummer}}
	client := &tsetest.FakeSetupClient{
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
	if !client.AdminAuthentifiziert {
		t.Error("expected AuthentifiziereAdmin to be called for a privileged reactivation")
	}
	if len(client.RegistrierteClients) != 0 {
		t.Errorf("expected no new client registration for a deregistered client, got %+v", client.RegistrierteClients)
	}
	if len(client.ReaktivierteClients) != 1 || client.ReaktivierteClients[0].ClientID != vorhandenerClient {
		t.Errorf("expected the same client to be reactivated, got %+v", client.ReaktivierteClients)
	}
	if repo.gespeichert == nil {
		t.Fatal("expected the configuration to be saved")
	}
	if ergebnis.ClientID != vorhandenerClient || repo.gespeichert.ClientID != vorhandenerClient {
		t.Errorf("expected the reactivated client to be saved, got result %q saved %+v", ergebnis.ClientID, repo.gespeichert)
	}
}

// Reactivation is privileged: without PIN it is refused before any write.
func TestUebernimmTSE_DeregistrierterClientBrauchtPIN(t *testing.T) {
	seriennummer := uuid.New()
	repo := &stubCommandRepo{identitaet: tse.Kassenidentitaet{Seriennummer: seriennummer}}
	client := &tsetest.FakeSetupClient{
		UmgebungResponse: tse.UmgebungTest,
		TSSResponse:      []tse.TSSInfo{{ID: "tss-init", State: "INITIALIZED"}},
		ClientsByTSS: map[string][]tse.ClientInfo{
			"tss-init": {{ID: uuid.NewString(), SerialNumber: seriennummer.String(), State: "DEREGISTERED"}},
		},
	}

	_, err := commandMit(repo, client).UebernimmTSE(context.Background(), zugangsdaten(), tse.UmgebungTest, "tss-init", "", "")
	if !errors.Is(err, ErrTSESetupPINErforderlich) {
		t.Errorf("expected ErrTSESetupPINErforderlich, got %v", err)
	}
	if len(client.ReaktivierteClients) != 0 || repo.gespeichert != nil {
		t.Error("expected no writes when the pin is missing")
	}
}

// Registration is privileged, so INITIALIZED without a matching client still needs the PIN.
func TestUebernimmTSE_InitialisiertOhneClientBrauchtPIN(t *testing.T) {
	repo := &stubCommandRepo{identitaet: tse.Kassenidentitaet{Seriennummer: uuid.New()}}
	client := &tsetest.FakeSetupClient{
		UmgebungResponse: tse.UmgebungTest,
		TSSResponse:      []tse.TSSInfo{{ID: "tss-init", State: "INITIALIZED"}},
	}

	_, err := commandMit(repo, client).UebernimmTSE(context.Background(), zugangsdaten(), tse.UmgebungTest, "tss-init", "", "")
	if !errors.Is(err, ErrTSESetupPINErforderlich) {
		t.Errorf("expected ErrTSESetupPINErforderlich, got %v", err)
	}
	if len(client.RegistrierteClients) != 0 || repo.gespeichert != nil {
		t.Error("expected no writes when the pin is missing")
	}
}

// A missing PIN from UNINITIALIZED is reported before any write.
func TestUebernimmTSE_PINErforderlich(t *testing.T) {
	repo := &stubCommandRepo{identitaet: tse.Kassenidentitaet{Seriennummer: uuid.New()}}
	client := &tsetest.FakeSetupClient{
		UmgebungResponse: tse.UmgebungTest,
		TSSResponse:      []tse.TSSInfo{{ID: "tss-uninit", State: "UNINITIALIZED"}},
	}

	_, err := commandMit(repo, client).UebernimmTSE(context.Background(), zugangsdaten(), tse.UmgebungTest, "tss-uninit", "", "")
	if !errors.Is(err, ErrTSESetupPINErforderlich) {
		t.Errorf("expected ErrTSESetupPINErforderlich, got %v", err)
	}
	if len(client.RegistrierteClients) != 0 || repo.gespeichert != nil {
		t.Error("expected no writes when the pin is missing")
	}
}

// A PIN rejected by fiskaly is a user-facing dead end, not a technical error.
func TestUebernimmTSE_UnbekanntePIN(t *testing.T) {
	repo := &stubCommandRepo{identitaet: tse.Kassenidentitaet{Seriennummer: uuid.New()}}
	client := &tsetest.FakeSetupClient{
		UmgebungResponse: tse.UmgebungTest,
		TSSResponse:      []tse.TSSInfo{{ID: "tss-init", State: "INITIALIZED"}},
		AuthAdminErr:     tse.ErrSetupAuthFehlgeschlagen,
	}

	_, err := commandMit(repo, client).UebernimmTSE(context.Background(), zugangsdaten(), tse.UmgebungTest, "tss-init", "0000000000", "")
	if !errors.Is(err, ErrTSESetupPINUnbekannt) {
		t.Errorf("expected ErrTSESetupPINUnbekannt, got %v", err)
	}
	if repo.gespeichert != nil {
		t.Error("expected no configuration to be saved on an unknown pin")
	}
}

// The stored PUK sets a fresh random PIN that finishes the takeover and is shown once.
// The PUK itself stays unchanged and is not returned again.
func TestUebernimmTSE_PINResetPerPUK(t *testing.T) {
	seriennummer := uuid.New()
	repo := &stubCommandRepo{identitaet: tse.Kassenidentitaet{Seriennummer: seriennummer}}
	client := &tsetest.FakeSetupClient{
		UmgebungResponse: tse.UmgebungTest,
		TSSResponse:      []tse.TSSInfo{{ID: "tss-init", State: "INITIALIZED"}},
	}

	ergebnis, err := commandMit(repo, client).UebernimmTSE(context.Background(), zugangsdaten(), tse.UmgebungTest, "tss-init", "", "puk-verwahrt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client.GesetzterAdminPUK != "puk-verwahrt" {
		t.Errorf("expected the supplied puk to be used for the reset, got %q", client.GesetzterAdminPUK)
	}
	if ergebnis.AdminPIN == "" || ergebnis.AdminPIN != client.GesetzteAdminPIN {
		t.Errorf("expected a fresh pin to be set and returned, got result %q set %q", ergebnis.AdminPIN, client.GesetzteAdminPIN)
	}
	if ergebnis.PUK != "" {
		t.Errorf("expected no puk to be returned on a reset, got %q", ergebnis.PUK)
	}
	// The fresh PIN drives the rest of the takeover (admin auth and client).
	if client.AuthentifiziertePIN != ergebnis.AdminPIN {
		t.Errorf("expected the fresh pin to be used for admin auth, got %q", client.AuthentifiziertePIN)
	}
	if len(client.RegistrierteClients) != 1 || client.RegistrierteClients[0].SerialNumber != seriennummer.String() {
		t.Errorf("expected the client to be registered with the kassen serial, got %+v", client.RegistrierteClients)
	}
	if repo.gespeichert == nil || repo.gespeichert.TssID != "tss-init" {
		t.Errorf("expected the configuration to be saved, got %+v", repo.gespeichert)
	}
}

// The PUK reset works in LIVE too, avoiding a new paid TSS; from UNINITIALIZED it also initialises.
func TestUebernimmTSE_PINResetPerPUKInLive(t *testing.T) {
	seriennummer := uuid.New()
	repo := &stubCommandRepo{identitaet: tse.Kassenidentitaet{Seriennummer: seriennummer}}
	client := &tsetest.FakeSetupClient{
		UmgebungResponse: tse.UmgebungLive,
		TSSResponse:      []tse.TSSInfo{{ID: "tss-uninit", State: "UNINITIALIZED"}},
	}

	ergebnis, err := commandMit(repo, client).UebernimmTSE(context.Background(), zugangsdaten(), tse.UmgebungLive, "tss-uninit", "", "puk-verwahrt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client.GesetzterAdminPUK != "puk-verwahrt" || ergebnis.AdminPIN == "" {
		t.Errorf("expected a puk-based pin reset in LIVE, got puk %q pin %q", client.GesetzterAdminPUK, ergebnis.AdminPIN)
	}
	if ergebnis.PUK != "" {
		t.Errorf("expected no puk to be returned on a reset, got %q", ergebnis.PUK)
	}
	if len(client.RegistrierteClients) != 1 || repo.gespeichert == nil {
		t.Errorf("expected the takeover to complete and save, got clients %+v saved %+v", client.RegistrierteClients, repo.gespeichert)
	}
}

// A rejected PUK ends before any further operation or save, as a user-facing error.
func TestUebernimmTSE_PINResetFalscherPUK(t *testing.T) {
	repo := &stubCommandRepo{identitaet: tse.Kassenidentitaet{Seriennummer: uuid.New()}}
	client := &tsetest.FakeSetupClient{
		UmgebungResponse: tse.UmgebungTest,
		TSSResponse:      []tse.TSSInfo{{ID: "tss-init", State: "INITIALIZED"}},
		SetAdminPINErr:   tse.ErrSetupAuthFehlgeschlagen,
	}

	_, err := commandMit(repo, client).UebernimmTSE(context.Background(), zugangsdaten(), tse.UmgebungTest, "tss-init", "", "puk-falsch")
	if !errors.Is(err, ErrTSESetupPUKUnbekannt) {
		t.Errorf("expected ErrTSESetupPUKUnbekannt, got %v", err)
	}
	if client.AdminAuthentifiziert || len(client.RegistrierteClients) != 0 || repo.gespeichert != nil {
		t.Error("expected no further operations or writes on a wrong puk")
	}
}

// The LIVE guard also holds for takeovers.
func TestUebernimmTSE_UmgebungAbweichung(t *testing.T) {
	repo := &stubCommandRepo{identitaet: tse.Kassenidentitaet{Seriennummer: uuid.New()}}
	client := &tsetest.FakeSetupClient{
		UmgebungResponse: tse.UmgebungLive,
		TSSResponse:      []tse.TSSInfo{{ID: "tss-x", State: "CREATED"}},
	}

	_, err := commandMit(repo, client).UebernimmTSE(context.Background(), zugangsdaten(), tse.UmgebungTest, "tss-x", "", "")
	if !errors.Is(err, ErrTSESetupUmgebungAbweichung) {
		t.Errorf("expected ErrTSESetupUmgebungAbweichung, got %v", err)
	}
	if client.GesetzteAdminPIN != "" || repo.gespeichert != nil {
		t.Error("expected no operations on environment mismatch")
	}
}

func TestUebernimmTSE_TSSNichtGefunden(t *testing.T) {
	repo := &stubCommandRepo{identitaet: tse.Kassenidentitaet{Seriennummer: uuid.New()}}
	client := &tsetest.FakeSetupClient{UmgebungResponse: tse.UmgebungTest}

	_, err := commandMit(repo, client).UebernimmTSE(context.Background(), zugangsdaten(), tse.UmgebungTest, "tss-fehlt", "", "")
	if !errors.Is(err, ErrTSESetupTSSNichtGefunden) {
		t.Errorf("expected ErrTSESetupTSSNichtGefunden, got %v", err)
	}
}

func TestUebernimmTSE_DeaktivierteTSS(t *testing.T) {
	repo := &stubCommandRepo{identitaet: tse.Kassenidentitaet{Seriennummer: uuid.New()}}
	client := &tsetest.FakeSetupClient{
		UmgebungResponse: tse.UmgebungTest,
		TSSResponse:      []tse.TSSInfo{{ID: "tss-tot", State: "DISABLED"}},
	}

	_, err := commandMit(repo, client).UebernimmTSE(context.Background(), zugangsdaten(), tse.UmgebungTest, "tss-tot", "1234567890", "")
	if !errors.Is(err, ErrTSESetupUebernahmeNichtMoeglich) {
		t.Errorf("expected ErrTSESetupUebernahmeNichtMoeglich, got %v", err)
	}
	if repo.gespeichert != nil {
		t.Error("expected no configuration to be saved for a disabled TSS")
	}
}

func stammdatenAntwort() tse.Stammdaten {
	return tse.Stammdaten{
		Seriennummer:        "abcdef1234567890abcdef1234567890",
		SignaturAlgorithmus: "ecdsa-plain-SHA256",
		PublicKey:           "public-key-b64",
		Zertifikat:          "certificate-b64",
		LogTimeFormat:       "unixTime",
	}
}

// checkStammdaten ignores the server-set timestamp.
func checkStammdaten(t *testing.T, gespeichert *tse.Stammdaten, erwartet tse.Stammdaten) {
	t.Helper()
	if gespeichert == nil {
		t.Fatal("expected the tse stammdaten to be persisted")
	}
	if gespeichert.Seriennummer != erwartet.Seriennummer ||
		gespeichert.SignaturAlgorithmus != erwartet.SignaturAlgorithmus ||
		gespeichert.PublicKey != erwartet.PublicKey ||
		gespeichert.Zertifikat != erwartet.Zertifikat ||
		gespeichert.LogTimeFormat != erwartet.LogTimeFormat {
		t.Errorf("persisted stammdaten do not match the fiskaly response, got %+v", gespeichert)
	}
}

// The DSFinV-K export needs algorithm, public key, certificate and log time format.
func TestRichteTSEEin_PersistiertStammdaten(t *testing.T) {
	repo := &stubCommandRepo{identitaet: tse.Kassenidentitaet{Seriennummer: uuid.New()}}
	client := &tsetest.FakeSetupClient{
		UmgebungResponse:   tse.UmgebungTest,
		CreateTSSResponse:  tse.TSSErstellt{ID: "tss-neu", PUK: "puk", State: "CREATED"},
		StammdatenResponse: stammdatenAntwort(),
	}

	_, err := commandMit(repo, client).RichteTSEEin(context.Background(), zugangsdaten(), tse.UmgebungTest, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client.StammdatenTssID != "tss-neu" {
		t.Errorf("expected stammdaten to be fetched for the new TSS, got %q", client.StammdatenTssID)
	}
	checkStammdaten(t, repo.gespeicherteStammdaten, stammdatenAntwort())
}

// Stammdaten hang on the shared save step, so even a ready TSS without any privileged call fetches them.
func TestUebernimmTSE_EinsatzbereitPersistiertStammdaten(t *testing.T) {
	seriennummer := uuid.New()
	vorhandenerClient := uuid.NewString()
	repo := &stubCommandRepo{identitaet: tse.Kassenidentitaet{Seriennummer: seriennummer}}
	client := &tsetest.FakeSetupClient{
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
	if client.AdminAuthentifiziert {
		t.Error("expected no lifecycle operations for a ready TSS")
	}
	if client.StammdatenTssID != "tss-init" {
		t.Errorf("expected stammdaten to be fetched for the adopted TSS, got %q", client.StammdatenTssID)
	}
	checkStammdaten(t, repo.gespeicherteStammdaten, stammdatenAntwort())
}

func TestUebernimmTSE_PINResetPersistiertStammdaten(t *testing.T) {
	repo := &stubCommandRepo{identitaet: tse.Kassenidentitaet{Seriennummer: uuid.New()}}
	client := &tsetest.FakeSetupClient{
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

// The export reads the TSE serial only from tse_stammdaten, so a fetch failure fails the setup.
func TestRichteTSEEin_StammdatenAbrufFehlerKipptSetup(t *testing.T) {
	repo := &stubCommandRepo{identitaet: tse.Kassenidentitaet{Seriennummer: uuid.New()}}
	client := &tsetest.FakeSetupClient{
		UmgebungResponse:  tse.UmgebungTest,
		CreateTSSResponse: tse.TSSErstellt{ID: "tss-neu", PUK: "puk", State: "CREATED"},
		StammdatenErr:     errors.New("fiskaly stammdaten read failed"),
	}

	_, err := commandMit(repo, client).RichteTSEEin(context.Background(), zugangsdaten(), tse.UmgebungTest, false)
	if !errors.Is(err, ErrTSEEinrichtung) {
		t.Errorf("expected ErrTSEEinrichtung when stammdaten fetch fails, got %v", err)
	}
}

// The setup client factory fails if called, proving the guard runs before any fiskaly work.
func commandMitAktiverKassensitzung(repo *stubCommandRepo, status kasse.KassensitzungStatus) Command {
	return Command{
		TSERepo:             repo,
		KassensitzungenRepo: stubKassensitzungReader{aktive: &kasse.Kassensitzung{ZNr: 1, Status: status}},
		NewTSESetupClient: func(tse.SetupCredentials) (tse.SetupClient, error) {
			return nil, errors.New("setup client must not be created while a Kassensitzung is active")
		},
	}
}

// The signing device must not change mid-Kassentag: all three change paths refuse and write nothing.
func TestUpdateTSEKonfiguration_MitOffenerKassensitzungAbgelehnt(t *testing.T) {
	repo := &stubCommandRepo{}
	conf, err := tse.NewKonfiguration("api-key", "api-secret", "tss-1", "client-1")
	if err != nil {
		t.Fatalf("unexpected error building konfiguration: %v", err)
	}

	err = commandMitAktiverKassensitzung(repo, kasse.KassensitzungOffen).UpdateTSEKonfiguration(context.Background(), conf)
	if !errors.Is(err, ErrTSEKonfigurationKassensitzungOffen) {
		t.Errorf("expected ErrTSEKonfigurationKassensitzungOffen, got %v", err)
	}
	if repo.gespeichert != nil {
		t.Errorf("expected no configuration to be saved, got %+v", repo.gespeichert)
	}
}

// wird_abgeschlossen counts as open: the TSS must not change while the closing runs.
func TestUpdateTSEKonfiguration_ImBarrierestatusAbgelehnt(t *testing.T) {
	repo := &stubCommandRepo{}
	conf, err := tse.NewKonfiguration("api-key", "api-secret", "tss-1", "client-1")
	if err != nil {
		t.Fatalf("unexpected error building konfiguration: %v", err)
	}

	err = commandMitAktiverKassensitzung(repo, kasse.KassensitzungWirdAbgeschlossen).
		UpdateTSEKonfiguration(context.Background(), conf)
	if !errors.Is(err, ErrTSEKonfigurationKassensitzungOffen) {
		t.Errorf("expected ErrTSEKonfigurationKassensitzungOffen, got %v", err)
	}
	if repo.gespeichert != nil {
		t.Errorf("expected no configuration to be saved, got %+v", repo.gespeichert)
	}
}

func TestRichteTSEEin_MitOffenerKassensitzungAbgelehnt(t *testing.T) {
	repo := &stubCommandRepo{identitaet: tse.Kassenidentitaet{Seriennummer: uuid.New()}}

	_, err := commandMitAktiverKassensitzung(repo, kasse.KassensitzungOffen).RichteTSEEin(context.Background(), zugangsdaten(), tse.UmgebungTest, false)
	if !errors.Is(err, ErrTSEKonfigurationKassensitzungOffen) {
		t.Errorf("expected ErrTSEKonfigurationKassensitzungOffen, got %v", err)
	}
	if repo.gespeichert != nil {
		t.Errorf("expected no configuration to be saved, got %+v", repo.gespeichert)
	}
}

func TestUebernimmTSE_MitOffenerKassensitzungAbgelehnt(t *testing.T) {
	repo := &stubCommandRepo{identitaet: tse.Kassenidentitaet{Seriennummer: uuid.New()}}

	_, err := commandMitAktiverKassensitzung(repo, kasse.KassensitzungOffen).UebernimmTSE(context.Background(), zugangsdaten(), tse.UmgebungTest, "tss-1", "", "")
	if !errors.Is(err, ErrTSEKonfigurationKassensitzungOffen) {
		t.Errorf("expected ErrTSEKonfigurationKassensitzungOffen, got %v", err)
	}
	if repo.gespeichert != nil {
		t.Errorf("expected no configuration to be saved, got %+v", repo.gespeichert)
	}
}
