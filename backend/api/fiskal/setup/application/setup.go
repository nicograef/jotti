package application

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"log/slog"
	"math/big"
	"strings"
	"sync/atomic"

	"github.com/google/uuid"
	"github.com/nicograef/jotti/backend/domain/tse"
	"github.com/nicograef/jotti/backend/internal/logging"
)

// Ten digits lie safely within the admin PIN length fiskaly accepts.
const adminPINStellen = 10

// einrichtungLaeuft admits at most one writer on the TSE configuration (see docs/handbuch.md §3.13).
// Package-level because Command has value receivers, which would copy a value field per call.
var einrichtungLaeuft atomic.Bool

// acquireEinrichtung fails fast with ErrTSESetupLaeuftBereits instead of waiting.
// Callers defer the release so every error path and a panic unlock too.
func acquireEinrichtung() (func(), error) {
	if !einrichtungLaeuft.CompareAndSwap(false, true) {
		return nil, ErrTSESetupLaeuftBereits
	}
	return func() { einrichtungLaeuft.Store(false) }, nil
}

// TSESetupErgebnis is the only place PUK and AdminPIN appear: never persisted or logged,
// handed to the UI once for the admin to keep outside jotti.
type TSESetupErgebnis struct {
	TssID    string
	ClientID string
	PUK      string
	AdminPIN string
	Umgebung string
}

// RichteTSEEin runs the fiskaly lifecycle on an empty account and saves only after it completes.
// LIVE guard and the existing-TSS lock with its TEST-only override: docs/handbuch.md §3.13.
func (c Command) RichteTSEEin(ctx context.Context, credentials tse.SetupCredentials, bestaetigteUmgebung tse.Umgebung, neuAnlegenTrotzVorhandener bool) (TSESetupErgebnis, error) {
	log := logging.FromContext(ctx)

	freigeben, err := acquireEinrichtung()
	if err != nil {
		return TSESetupErgebnis{}, err
	}
	defer freigeben()

	client, umgebung, tssListe, err := c.vorbereiteEinrichtung(ctx, log, credentials, bestaetigteUmgebung, "setup")
	if err != nil {
		return TSESetupErgebnis{}, err
	}

	// Only TEST may bypass the lock: a second LIVE TSS incurs ongoing cost.
	// umgebung is authoritative here because oeffneSetupClient matched it against bestaetigteUmgebung.
	neuanlageErzwungen := neuAnlegenTrotzVorhandener && umgebung == tse.UmgebungTest
	if hatAktiveTSS(tssListe) && !neuanlageErzwungen {
		return TSESetupErgebnis{}, ErrTSEBereitsEingerichtet
	}

	identitaet, err := c.TSERepo.GetKassenidentitaet(ctx)
	if err != nil {
		log.Error("Failed to retrieve kassenidentitaet for setup", "error", err)
		return TSESetupErgebnis{}, ErrDatabase
	}
	seriennummer := identitaet.Seriennummer.String()

	// The client _id is a fresh UUIDv4 (fiskaly convention), kept apart from the Kassen-Seriennummer.
	// The Kassen-Seriennummer is the client's serial_number (docs/compliance.md §3.7).
	clientID := uuid.NewString()

	pin, err := generateAdminPIN()
	if err != nil {
		log.Error("Failed to generate admin pin", "error", err)
		return TSESetupErgebnis{}, ErrTSEEinrichtung
	}

	// Lifecycle CREATED -> UNINITIALIZED -> (PIN) -> INITIALIZED -> client.
	// A fresh TSS always starts CREATED, and its PUK comes straight from the creation response.
	erstellt, err := client.CreateTSS(ctx)
	if err != nil {
		// The fiskaly TSS limit (five active TSS in TEST) is a user-facing state, not a technical error.
		if errors.Is(err, tse.ErrSetupTSSLimitErreicht) {
			return TSESetupErgebnis{}, ErrTSESetupTSSLimitErreicht
		}
		return TSESetupErgebnis{}, einrichtungsFehler(log, err, "tss anlegen", "")
	}
	if err := vollendeLebenszyklus(ctx, log, client, lebenszyklus{
		start:        tse.TSSCreated,
		tssID:        erstellt.ID,
		puk:          erstellt.PUK,
		pin:          pin,
		clientID:     clientID,
		seriennummer: seriennummer,
		aktion:       clientRegistrieren,
	}); err != nil {
		return TSESetupErgebnis{}, err
	}

	if err := c.saveEinrichtung(ctx, log, client, credentials, erstellt.ID, clientID); err != nil {
		return TSESetupErgebnis{}, err
	}

	log.Info("TSE setup completed", "tss_id", erstellt.ID, "umgebung", string(umgebung))

	return TSESetupErgebnis{
		TssID:    erstellt.ID,
		ClientID: clientID,
		PUK:      erstellt.PUK,
		AdminPIN: pin,
		Umgebung: string(umgebung),
	}, nil
}

// UebernimmTSE drives an existing TSS from its current state to a registered client,
// which also resumes an aborted setup. PIN/PUK handling per start state: docs/handbuch.md §3.13.
func (c Command) UebernimmTSE(ctx context.Context, credentials tse.SetupCredentials, bestaetigteUmgebung tse.Umgebung, tssID, pin, puk string) (TSESetupErgebnis, error) {
	log := logging.FromContext(ctx)

	freigeben, err := acquireEinrichtung()
	if err != nil {
		return TSESetupErgebnis{}, err
	}
	defer freigeben()

	tssID = strings.TrimSpace(tssID)
	if tssID == "" {
		return TSESetupErgebnis{}, ErrTSESetupTSSNichtGefunden
	}

	client, umgebung, tssListe, err := c.vorbereiteEinrichtung(ctx, log, credentials, bestaetigteUmgebung, "takeover")
	if err != nil {
		return TSESetupErgebnis{}, err
	}

	ziel, gefunden := findTSS(tssListe, tssID)
	if !gefunden {
		return TSESetupErgebnis{}, ErrTSESetupTSSNichtGefunden
	}
	state := ziel.State

	identitaet, err := c.TSERepo.GetKassenidentitaet(ctx)
	if err != nil {
		log.Error("Failed to retrieve kassenidentitaet for takeover", "error", err)
		return TSESetupErgebnis{}, ErrDatabase
	}
	seriennummer := identitaet.Seriennummer.String()

	clients, err := client.ListClients(ctx, tssID)
	if err != nil {
		log.Warn("Failed to list clients during takeover", "error", err, "tss_id", tssID)
		return TSESetupErgebnis{}, ErrTSEVerbindungFehlgeschlagen
	}

	bewertung := bewerteUebernahme(state, passenderClient(clients, seriennummer))
	if !bewertung.Moeglich {
		return TSESetupErgebnis{}, ErrTSESetupUebernahmeNichtMoeglich
	}
	clientID, aktion := passendeClientAktion(clients, seriennummer)

	zugang, err := bereiteAdminPINVor(ctx, log, client, state, tssID, pin, puk, bewertung.PINErforderlich)
	if err != nil {
		return TSESetupErgebnis{}, err
	}

	if err := vollendeLebenszyklus(ctx, log, client, lebenszyklus{
		start:        state,
		tssID:        tssID,
		puk:          zugang.lebenszyklusPUK,
		pin:          zugang.pin,
		clientID:     clientID,
		seriennummer: seriennummer,
		aktion:       aktion,
	}); err != nil {
		return TSESetupErgebnis{}, err
	}

	if err := c.saveEinrichtung(ctx, log, client, credentials, tssID, clientID); err != nil {
		return TSESetupErgebnis{}, err
	}

	log.Info("TSE takeover completed", "tss_id", tssID, "umgebung", string(umgebung), "ausgangszustand", string(state))

	return TSESetupErgebnis{
		TssID:    tssID,
		ClientID: clientID,
		PUK:      zugang.neuerPUK,
		AdminPIN: zugang.neuePIN,
		Umgebung: string(umgebung),
	}, nil
}

// vorbereiteEinrichtung runs the checks RichteTSEEin and UebernimmTSE share before any write at fiskaly.
func (c Command) vorbereiteEinrichtung(ctx context.Context, log *slog.Logger, credentials tse.SetupCredentials, bestaetigteUmgebung tse.Umgebung, zweck string) (tse.SetupClient, tse.Umgebung, []tse.TSSInfo, error) {
	if err := credentials.Validate(); err != nil {
		return nil, "", nil, ErrTSESetupZugangsdaten
	}
	if bestaetigteUmgebung != tse.UmgebungTest && bestaetigteUmgebung != tse.UmgebungLive {
		return nil, "", nil, ErrTSESetupUmgebungAbweichung
	}
	if err := c.ensureKeineAktiveKassensitzung(ctx); err != nil {
		return nil, "", nil, err
	}
	return oeffneSetupClient(ctx, log, c.NewTSESetupClient, credentials, &bestaetigteUmgebung, zweck)
}

// oeffneSetupClient lists the account's TSS; a nil bestaetigteUmgebung accepts any environment.
// zweck only labels the log message.
func oeffneSetupClient(ctx context.Context, log *slog.Logger, newClient NewTSESetupClient, credentials tse.SetupCredentials, bestaetigteUmgebung *tse.Umgebung, zweck string) (tse.SetupClient, tse.Umgebung, []tse.TSSInfo, error) {
	client, err := newClient(credentials)
	if err != nil {
		log.Error("Failed to create TSE setup client", "error", err)
		return nil, "", nil, ErrTSEVerbindungFehlgeschlagen
	}

	umgebung, tssListe, err := client.ListTSS(ctx)
	if err != nil {
		if errors.Is(err, tse.ErrSetupAuthFehlgeschlagen) {
			return nil, "", nil, ErrTSESetupZugangsdaten
		}
		log.Warn(fmt.Sprintf("Failed to list TSS during %s", zweck), "error", err)
		return nil, "", nil, ErrTSEVerbindungFehlgeschlagen
	}

	if bestaetigteUmgebung != nil && umgebung != *bestaetigteUmgebung {
		log.Warn("Confirmed TSE environment does not match actual environment", "bestaetigt", string(*bestaetigteUmgebung), "tatsaechlich", string(umgebung))
		return nil, "", nil, ErrTSESetupUmgebungAbweichung
	}

	return client, umgebung, tssListe, nil
}

// If saving fails, the TSS still exists at fiskaly and a takeover recovers it.
// tss_id and client_id are logged for that, PUK and PIN never.
func (c Command) saveEinrichtung(ctx context.Context, log *slog.Logger, client tse.SetupClient, credentials tse.SetupCredentials, tssID, clientID string) error {
	konfiguration, err := tse.NewKonfiguration(credentials.ApiKey, credentials.ApiSecret, tssID, clientID)
	if err != nil {
		log.Error("Failed to build tse_konfiguration after setup", "error", err)
		return ErrTSEEinrichtung
	}
	// On the transition to configured, the same transaction finalises pre-configuration orders
	// and closes the keine_konfiguration outage (docs/handbuch.md §3.13).
	if err := c.TSERepo.SaveEinrichtung(ctx, konfiguration); err != nil {
		log.Error("Failed to save tse_konfiguration after setup; TSS exists at fiskaly, recoverable via takeover", "error", err, "tss_id", tssID, "client_id", clientID)
		return ErrDatabase
	}

	return c.fetchTSEStammdaten(ctx, log, client, tssID)
}

// The DSFinV-K export fills tse.csv (serial, public key, certificate; docs/compliance.md §6.3) only
// from the stored TSE master data, so a failure here fails the setup.
func (c Command) fetchTSEStammdaten(ctx context.Context, log *slog.Logger, client tse.SetupClient, tssID string) error {
	stammdaten, err := client.RetrieveTSSStammdaten(ctx, tssID)
	if err != nil {
		return einrichtungsFehler(log, err, "stammdaten abrufen", tssID)
	}
	if err := c.TSERepo.UpsertTSEStammdaten(ctx, stammdaten); err != nil {
		log.Error("Failed to save TSE Stammdaten after setup", "error", err, "tss_id", tssID)
		return ErrDatabase
	}
	return nil
}

// Uebernahme is the one decision on taking over an existing TSS; CheckTSESetup shows it, UebernimmTSE enforces it.
// Einsatzbereit needs no privileged fiskaly operation, so only it and CREATED (jotti fetches the PUK) need no admin PIN.
type Uebernahme struct {
	Moeglich        bool
	Einsatzbereit   bool
	PINErforderlich bool
}

func bewerteUebernahme(state tse.TSSZustand, passender *ClientBefund) Uebernahme {
	switch state {
	case tse.TSSCreated:
		return Uebernahme{Moeglich: true}
	case tse.TSSUninitialized:
		return Uebernahme{Moeglich: true, PINErforderlich: true}
	case tse.TSSInitialized:
		bereit := passender != nil && passender.State == tse.ClientRegistered
		return Uebernahme{Moeglich: true, Einsatzbereit: bereit, PINErforderlich: !bereit}
	case tse.TSSDisabled, tse.TSSDeleted, tse.TSSDefective, tse.TSSEvicted:
		return Uebernahme{}
	}
	return Uebernahme{}
}

// clientAktion is the client step of an INITIALIZED TSS: register a new client,
// reactivate a DEREGISTERED one, or nothing when the matching client is REGISTERED.
type clientAktion int

const (
	clientRegistrieren clientAktion = iota
	clientReaktivieren
	clientFertig
)

// passendeClientAktion finds this Kasse's client among the TSS's clients: a REGISTERED one is ready,
// another is reactivated, and without one a new client gets a fresh UUID.
func passendeClientAktion(clients []tse.ClientInfo, seriennummer string) (string, clientAktion) {
	passender := passenderClient(clients, seriennummer)
	if passender == nil {
		return uuid.NewString(), clientRegistrieren
	}
	if passender.State == tse.ClientRegistered {
		return passender.ID, clientFertig
	}
	return passender.ID, clientReaktivieren
}

// adminZugang holds the PIN the lifecycle authenticates with and the secrets created on the way,
// which jotti shows exactly once.
type adminZugang struct {
	lebenszyklusPUK string // only from CREATED, to set the first PIN
	pin             string
	neuerPUK        string
	neuePIN         string
}

// bereiteAdminPINVor provides the admin PIN per start state (docs/handbuch.md §3.13): CREATED gets a fresh
// one via the re-read PUK, later states use the stored PIN or reset it with the user's PUK.
func bereiteAdminPINVor(ctx context.Context, log *slog.Logger, client tse.SetupClient, start tse.TSSZustand, tssID, pin, puk string, pinErforderlich bool) (adminZugang, error) {
	switch start {
	case tse.TSSCreated:
		neuerPUK, err := client.GetAdminPUK(ctx, tssID)
		if err != nil {
			return adminZugang{}, einrichtungsFehler(log, err, "puk beziehen", tssID)
		}
		neuePIN, err := generateAdminPIN()
		if err != nil {
			log.Error("Failed to generate admin pin", "error", err)
			return adminZugang{}, ErrTSEEinrichtung
		}
		return adminZugang{lebenszyklusPUK: neuerPUK, pin: neuePIN, neuerPUK: neuerPUK, neuePIN: neuePIN}, nil
	case tse.TSSUninitialized, tse.TSSInitialized:
		if strings.TrimSpace(puk) != "" {
			neuePIN, err := generateAdminPIN()
			if err != nil {
				log.Error("Failed to generate admin pin", "error", err)
				return adminZugang{}, ErrTSEEinrichtung
			}
			// ListTSS already confirmed the credentials, so a failure while setting the PIN
			// via the PUK is practically always a wrong PUK.
			if err := client.SetAdminPIN(ctx, tssID, puk, neuePIN); err != nil {
				log.Warn("Admin PIN reset via PUK failed", "error", err, "tss_id", tssID)
				return adminZugang{}, ErrTSESetupPUKUnbekannt
			}
			return adminZugang{pin: neuePIN, neuePIN: neuePIN}, nil
		}
		if pinErforderlich && strings.TrimSpace(pin) == "" {
			return adminZugang{}, ErrTSESetupPINErforderlich
		}
		return adminZugang{pin: pin}, nil
	case tse.TSSDisabled, tse.TSSDeleted, tse.TSSDefective, tse.TSSEvicted:
		return adminZugang{}, ErrTSESetupUebernahmeNichtMoeglich
	}
	return adminZugang{}, ErrTSESetupUebernahmeNichtMoeglich
}

// lebenszyklus drives a TSS from its start state to a registered client. puk is used only from
// CREATED to set the fresh admin PIN; from UNINITIALIZED, pin is the existing one.
type lebenszyklus struct {
	start        tse.TSSZustand
	tssID        string
	puk          string
	pin          string
	clientID     string
	seriennummer string
	aktion       clientAktion
}

// vollendeLebenszyklus ends an auth failure with a user-entered PIN (start state != CREATED) as ErrTSESetupPINUnbekannt.
func vollendeLebenszyklus(ctx context.Context, log *slog.Logger, client tse.SetupClient, l lebenszyklus) error {
	pinVomNutzer := l.start != tse.TSSCreated
	authFehler := func(err error, schritt string) error {
		if pinVomNutzer && errors.Is(err, tse.ErrSetupAuthFehlgeschlagen) {
			return ErrTSESetupPINUnbekannt
		}
		return einrichtungsFehler(log, err, schritt, l.tssID)
	}

	switch l.start {
	case tse.TSSCreated:
		if err := client.PersonalisiereTSS(ctx, l.tssID); err != nil {
			return einrichtungsFehler(log, err, "personalisieren", l.tssID)
		}
		if err := client.SetAdminPIN(ctx, l.tssID, l.puk, l.pin); err != nil {
			return einrichtungsFehler(log, err, "admin-pin setzen", l.tssID)
		}
		fallthrough
	case tse.TSSUninitialized:
		if err := client.AuthentifiziereAdmin(ctx, l.tssID, l.pin); err != nil {
			return authFehler(err, "admin-auth (init)")
		}
		if err := client.InitialisiereTSS(ctx, l.tssID); err != nil {
			return einrichtungsFehler(log, err, "initialisieren", l.tssID)
		}
		fallthrough
	case tse.TSSInitialized:
		// clientFertig never falls through from CREATED/UNINITIALIZED: those states have no client yet.
		if l.aktion == clientFertig {
			return nil
		}
		if err := client.AuthentifiziereAdmin(ctx, l.tssID, l.pin); err != nil {
			return authFehler(err, "admin-auth (client)")
		}
		// serial_number is unique per TSS, so a DEREGISTERED client is reactivated, not recreated.
		if l.aktion == clientReaktivieren {
			if err := client.ReaktiviereClient(ctx, l.tssID, l.clientID); err != nil {
				return einrichtungsFehler(log, err, "client reaktivieren", l.tssID)
			}
			return nil
		}
		if err := client.RegistriereClient(ctx, l.tssID, l.clientID, l.seriennummer); err != nil {
			return einrichtungsFehler(log, err, "client registrieren", l.tssID)
		}
		return nil
	case tse.TSSDisabled, tse.TSSDeleted, tse.TSSDefective, tse.TSSEvicted:
		return ErrTSESetupUebernahmeNichtMoeglich
	}
	return ErrTSESetupUebernahmeNichtMoeglich
}

func findTSS(tssListe []tse.TSSInfo, tssID string) (tse.TSSInfo, bool) {
	for _, t := range tssListe {
		if t.ID == tssID {
			return t, true
		}
	}
	return tse.TSSInfo{}, false
}

// einrichtungsFehler logs a failed lifecycle step without PUK or PIN.
func einrichtungsFehler(log *slog.Logger, err error, schritt, tssID string) error {
	log.Warn("TSE setup step failed", "error", err, "schritt", schritt, "tss_id", tssID)
	return ErrTSEEinrichtung
}

// Only DISABLED TSS count as dead and do not block a new setup.
func hatAktiveTSS(tssListe []tse.TSSInfo) bool {
	for _, t := range tssListe {
		if t.State != tse.TSSDisabled {
			return true
		}
	}
	return false
}

func generateAdminPIN() (string, error) {
	var sb strings.Builder
	for range adminPINStellen {
		ziffer, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			return "", err
		}
		sb.WriteByte(byte('0' + ziffer.Int64()))
	}
	return sb.String(), nil
}
