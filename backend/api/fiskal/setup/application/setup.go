package application

import (
	"context"
	"crypto/rand"
	"errors"
	"math/big"
	"strings"
	"sync/atomic"

	"github.com/google/uuid"
	"github.com/nicograef/jotti/backend/domain/tse"
	"github.com/rs/zerolog"
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
	log := zerolog.Ctx(ctx)

	freigeben, err := acquireEinrichtung()
	if err != nil {
		return TSESetupErgebnis{}, err
	}
	defer freigeben()

	if c.NewTSESetupClient == nil {
		log.Error().Msg("Missing TSE setup client factory")
		return TSESetupErgebnis{}, ErrDatabase
	}
	if err := credentials.Validate(); err != nil {
		return TSESetupErgebnis{}, ErrTSESetupZugangsdaten
	}
	if bestaetigteUmgebung != tse.UmgebungTest && bestaetigteUmgebung != tse.UmgebungLive {
		return TSESetupErgebnis{}, ErrTSESetupUmgebungAbweichung
	}
	if err := c.ensureKeineAktiveKassensitzung(ctx); err != nil {
		return TSESetupErgebnis{}, err
	}

	client, umgebung, tssListe, err := c.oeffneSetupClient(ctx, log, credentials, bestaetigteUmgebung, "setup")
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
		log.Error().Err(err).Msg("Failed to retrieve kassenidentitaet for setup")
		return TSESetupErgebnis{}, ErrDatabase
	}
	seriennummer := identitaet.Seriennummer.String()

	// The client _id is a fresh UUIDv4 (fiskaly convention), kept apart from the Kassen-Seriennummer.
	// The Kassen-Seriennummer is the client's serial_number (docs/compliance.md §3.7).
	clientID := uuid.NewString()

	pin, err := generateAdminPIN()
	if err != nil {
		log.Error().Err(err).Msg("Failed to generate admin pin")
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
	if err := vollendeLebenszyklus(ctx, log, client, "CREATED", erstellt.ID, erstellt.PUK, pin, clientID, seriennummer, clientRegistrieren); err != nil {
		return TSESetupErgebnis{}, err
	}

	if err := c.saveEinrichtung(ctx, log, client, credentials, erstellt.ID, clientID); err != nil {
		return TSESetupErgebnis{}, err
	}

	log.Info().Str("tss_id", erstellt.ID).Str("umgebung", string(umgebung)).Msg("TSE setup completed")

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
	log := zerolog.Ctx(ctx)

	freigeben, err := acquireEinrichtung()
	if err != nil {
		return TSESetupErgebnis{}, err
	}
	defer freigeben()

	if c.NewTSESetupClient == nil {
		log.Error().Msg("Missing TSE setup client factory")
		return TSESetupErgebnis{}, ErrDatabase
	}
	if err := credentials.Validate(); err != nil {
		return TSESetupErgebnis{}, ErrTSESetupZugangsdaten
	}
	if bestaetigteUmgebung != tse.UmgebungTest && bestaetigteUmgebung != tse.UmgebungLive {
		return TSESetupErgebnis{}, ErrTSESetupUmgebungAbweichung
	}
	tssID = strings.TrimSpace(tssID)
	if tssID == "" {
		return TSESetupErgebnis{}, ErrTSESetupTSSNichtGefunden
	}
	if err := c.ensureKeineAktiveKassensitzung(ctx); err != nil {
		return TSESetupErgebnis{}, err
	}

	client, umgebung, tssListe, err := c.oeffneSetupClient(ctx, log, credentials, bestaetigteUmgebung, "takeover")
	if err != nil {
		return TSESetupErgebnis{}, err
	}

	ziel, gefunden := findTSS(tssListe, tssID)
	if !gefunden {
		return TSESetupErgebnis{}, ErrTSESetupTSSNichtGefunden
	}
	state := strings.ToUpper(strings.TrimSpace(ziel.State))

	identitaet, err := c.TSERepo.GetKassenidentitaet(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to retrieve kassenidentitaet for takeover")
		return TSESetupErgebnis{}, ErrDatabase
	}
	seriennummer := identitaet.Seriennummer.String()

	clients, err := client.ListClients(ctx, tssID)
	if err != nil {
		log.Warn().Err(err).Str("tss_id", tssID).Msg("Failed to list clients during takeover")
		return TSESetupErgebnis{}, ErrTSEVerbindungFehlgeschlagen
	}

	clientID := uuid.NewString()
	aktion := clientRegistrieren
	if passender := passenderClient(clients, seriennummer); passender != nil {
		clientID = passender.ID
		if strings.EqualFold(strings.TrimSpace(passender.State), "REGISTERED") {
			aktion = clientFertig
		} else {
			aktion = clientReaktivieren
		}
	}

	// Only this case runs no privileged fiskaly operation and therefore needs no admin PIN.
	einsatzbereit := state == "INITIALIZED" && aktion == clientFertig

	// lebenszyklusPUK only drives the CREATED step (setting the first PIN).
	// ergebnisPUK and ergebnisPIN are the newly created secrets, shown once.
	var lebenszyklusPUK, ergebnisPUK, ergebnisPIN string
	pinReset := strings.TrimSpace(puk) != ""
	switch state {
	case "CREATED":
		lebenszyklusPUK, err = client.GetAdminPUK(ctx, tssID)
		if err != nil {
			return TSESetupErgebnis{}, einrichtungsFehler(log, err, "puk beziehen", tssID)
		}
		pin, err = generateAdminPIN()
		if err != nil {
			log.Error().Err(err).Msg("Failed to generate admin pin")
			return TSESetupErgebnis{}, ErrTSEEinrichtung
		}
		ergebnisPUK, ergebnisPIN = lebenszyklusPUK, pin
	case "UNINITIALIZED", "INITIALIZED":
		switch {
		case pinReset:
			// ListTSS already confirmed the credentials, so a failure while setting the PIN
			// via the PUK is practically always a wrong PUK.
			pin, err = generateAdminPIN()
			if err != nil {
				log.Error().Err(err).Msg("Failed to generate admin pin")
				return TSESetupErgebnis{}, ErrTSEEinrichtung
			}
			if err := client.SetAdminPIN(ctx, tssID, puk, pin); err != nil {
				log.Warn().Err(err).Str("tss_id", tssID).Msg("Admin PIN reset via PUK failed")
				return TSESetupErgebnis{}, ErrTSESetupPUKUnbekannt
			}
			ergebnisPIN = pin
		case !einsatzbereit && strings.TrimSpace(pin) == "":
			return TSESetupErgebnis{}, ErrTSESetupPINErforderlich
		}
	default:
		return TSESetupErgebnis{}, ErrTSESetupUebernahmeNichtMoeglich
	}

	if err := vollendeLebenszyklus(ctx, log, client, state, tssID, lebenszyklusPUK, pin, clientID, seriennummer, aktion); err != nil {
		return TSESetupErgebnis{}, err
	}

	if err := c.saveEinrichtung(ctx, log, client, credentials, tssID, clientID); err != nil {
		return TSESetupErgebnis{}, err
	}

	log.Info().Str("tss_id", tssID).Str("umgebung", string(umgebung)).Str("ausgangszustand", state).Msg("TSE takeover completed")

	return TSESetupErgebnis{
		TssID:    tssID,
		ClientID: clientID,
		PUK:      ergebnisPUK,
		AdminPIN: ergebnisPIN,
		Umgebung: string(umgebung),
	}, nil
}

// oeffneSetupClient holds the LIVE guard: an environment mismatch ends the setup before any write.
// zweck only labels the log message.
func (c Command) oeffneSetupClient(ctx context.Context, log *zerolog.Logger, credentials tse.SetupCredentials, bestaetigteUmgebung tse.Umgebung, zweck string) (tse.SetupClient, tse.Umgebung, []tse.TSSInfo, error) {
	client, err := c.NewTSESetupClient(credentials)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create TSE setup client")
		return nil, "", nil, ErrTSEVerbindungFehlgeschlagen
	}

	umgebung, tssListe, err := client.ListTSS(ctx)
	if err != nil {
		if errors.Is(err, tse.ErrSetupAuthFehlgeschlagen) {
			return nil, "", nil, ErrTSESetupZugangsdaten
		}
		log.Warn().Err(err).Msgf("Failed to list TSS during %s", zweck)
		return nil, "", nil, ErrTSEVerbindungFehlgeschlagen
	}

	if umgebung != bestaetigteUmgebung {
		log.Warn().
			Str("bestaetigt", string(bestaetigteUmgebung)).
			Str("tatsaechlich", string(umgebung)).
			Msg("Confirmed TSE environment does not match actual environment")
		return nil, "", nil, ErrTSESetupUmgebungAbweichung
	}

	return client, umgebung, tssListe, nil
}

// If saving fails, the TSS still exists at fiskaly and a takeover recovers it.
// tss_id and client_id are logged for that, PUK and PIN never.
func (c Command) saveEinrichtung(ctx context.Context, log *zerolog.Logger, client tse.SetupClient, credentials tse.SetupCredentials, tssID, clientID string) error {
	konfiguration, err := tse.NewKonfiguration(credentials.ApiKey, credentials.ApiSecret, tssID, clientID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to build tse_konfiguration after setup")
		return ErrTSEEinrichtung
	}
	// On the transition to configured, the same transaction finalises pre-configuration orders
	// and closes the keine_konfiguration outage (docs/handbuch.md §3.13).
	if err := c.TSERepo.SaveEinrichtung(ctx, konfiguration); err != nil {
		log.Error().Err(err).Str("tss_id", tssID).Str("client_id", clientID).
			Msg("Failed to save tse_konfiguration after setup; TSS exists at fiskaly, recoverable via takeover")
		return ErrDatabase
	}

	return c.fetchTSEStammdaten(ctx, log, client, tssID)
}

// The DSFinV-K export reads TSE serial, public key and certificate only from tse_stammdaten
// (docs/compliance.md §6.3), so a failure here fails the setup.
func (c Command) fetchTSEStammdaten(ctx context.Context, log *zerolog.Logger, client tse.SetupClient, tssID string) error {
	stammdaten, err := client.RetrieveTSSStammdaten(ctx, tssID)
	if err != nil {
		return einrichtungsFehler(log, err, "stammdaten abrufen", tssID)
	}
	if err := c.TSERepo.UpsertTSEStammdaten(ctx, stammdaten); err != nil {
		log.Error().Err(err).Str("tss_id", tssID).Msg("Failed to save TSE Stammdaten after setup")
		return ErrDatabase
	}
	return nil
}

// clientAktion is the client step of an INITIALIZED TSS: register a new client,
// reactivate a DEREGISTERED one, or nothing when the matching client is REGISTERED.
type clientAktion int

const (
	clientRegistrieren clientAktion = iota
	clientReaktivieren
	clientFertig
)

// puk is used only from CREATED to set the fresh admin PIN; from UNINITIALIZED, pin is the existing one.
// An auth failure with a user-entered PIN (start state != CREATED) ends as ErrTSESetupPINUnbekannt.
func vollendeLebenszyklus(ctx context.Context, log *zerolog.Logger, client tse.SetupClient, state, tssID, puk, pin, clientID, seriennummer string, aktion clientAktion) error {
	pinVomNutzer := state != "CREATED"
	authFehler := func(err error, schritt string) error {
		if pinVomNutzer && errors.Is(err, tse.ErrSetupAuthFehlgeschlagen) {
			return ErrTSESetupPINUnbekannt
		}
		return einrichtungsFehler(log, err, schritt, tssID)
	}

	switch state {
	case "CREATED":
		if err := client.PersonalisiereTSS(ctx, tssID); err != nil {
			return einrichtungsFehler(log, err, "personalisieren", tssID)
		}
		if err := client.SetAdminPIN(ctx, tssID, puk, pin); err != nil {
			return einrichtungsFehler(log, err, "admin-pin setzen", tssID)
		}
		fallthrough
	case "UNINITIALIZED":
		if err := client.AuthentifiziereAdmin(ctx, tssID, pin); err != nil {
			return authFehler(err, "admin-auth (init)")
		}
		if err := client.InitialisiereTSS(ctx, tssID); err != nil {
			return einrichtungsFehler(log, err, "initialisieren", tssID)
		}
		fallthrough
	case "INITIALIZED":
		// clientFertig never falls through from CREATED/UNINITIALIZED: those states have no client yet.
		if aktion == clientFertig {
			return nil
		}
		if err := client.AuthentifiziereAdmin(ctx, tssID, pin); err != nil {
			return authFehler(err, "admin-auth (client)")
		}
		// serial_number is unique per TSS, so a DEREGISTERED client is reactivated, not recreated.
		if aktion == clientReaktivieren {
			if err := client.ReaktiviereClient(ctx, tssID, clientID); err != nil {
				return einrichtungsFehler(log, err, "client reaktivieren", tssID)
			}
			return nil
		}
		if err := client.RegistriereClient(ctx, tssID, clientID, seriennummer); err != nil {
			return einrichtungsFehler(log, err, "client registrieren", tssID)
		}
	default:
		return ErrTSESetupUebernahmeNichtMoeglich
	}
	return nil
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
func einrichtungsFehler(log *zerolog.Logger, err error, schritt, tssID string) error {
	log.Warn().Err(err).Str("schritt", schritt).Str("tss_id", tssID).Msg("TSE setup step failed")
	return ErrTSEEinrichtung
}

// Only DISABLED TSS count as dead and do not block a new setup.
func hatAktiveTSS(tssListe []tse.TSSInfo) bool {
	for _, t := range tssListe {
		if !strings.EqualFold(strings.TrimSpace(t.State), "DISABLED") {
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
