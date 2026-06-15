package application

import (
	"context"
	"crypto/rand"
	"errors"
	"math/big"
	"strings"

	"github.com/google/uuid"
	"github.com/nicograef/jotti/backend/domain/settings"
	"github.com/nicograef/jotti/backend/domain/tse"
	"github.com/rs/zerolog"
)

// adminPINStellen ist die Laenge der zufaellig erzeugten Admin-PIN. Zehn Ziffern
// liegen sicher innerhalb der von fiskaly akzeptierten Laenge.
const adminPINStellen = 10

// TSESetupErgebnis ist das Ergebnis der gefuehrten Einrichtung. PUK und AdminPIN
// erscheinen genau hier — sie werden weder persistiert noch geloggt und nur
// einmalig an die UI uebergeben, damit der Admin sie extern verwahren kann.
type TSESetupErgebnis struct {
	TssID    string
	ClientID string
	PUK      string
	AdminPIN string
	Umgebung string
}

// RichteTSEEin fuehrt den vollstaendigen fiskaly-Lebenszyklus fuer ein leeres
// Konto durch: TSS anlegen, personalisieren, Admin-PIN setzen, initialisieren
// und einen Client mit der Kassen-Seriennummer registrieren. Die Konfiguration
// wird erst nach erfolgreichem Abschluss atomar gespeichert — ein Abbruch
// hinterlaesst keine halbe Konfiguration.
//
// bestaetigteUmgebung ist die vom Admin bestaetigte Umgebung; weicht sie von der
// tatsaechlichen ab, bricht die Einrichtung vor jeder Schreiboperation ab
// (Schutz vor versehentlicher LIVE-Anlage). Existiert bereits eine aktive TSS,
// wird die Neuanlage verweigert.
func (c Command) RichteTSEEin(ctx context.Context, credentials tse.SetupCredentials, bestaetigteUmgebung tse.Umgebung) (TSESetupErgebnis, error) {
	log := zerolog.Ctx(ctx)

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

	client, err := c.NewTSESetupClient(credentials)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create TSE setup client")
		return TSESetupErgebnis{}, ErrTSEVerbindungFehlgeschlagen
	}

	umgebung, tssListe, err := client.ListTSS(ctx)
	if err != nil {
		if errors.Is(err, tse.ErrSetupAuthFehlgeschlagen) {
			return TSESetupErgebnis{}, ErrTSESetupZugangsdaten
		}
		log.Warn().Err(err).Msg("Failed to list TSS during setup")
		return TSESetupErgebnis{}, ErrTSEVerbindungFehlgeschlagen
	}

	if umgebung != bestaetigteUmgebung {
		log.Warn().
			Str("bestaetigt", string(bestaetigteUmgebung)).
			Str("tatsaechlich", string(umgebung)).
			Msg("Confirmed TSE environment does not match actual environment")
		return TSESetupErgebnis{}, ErrTSESetupUmgebungAbweichung
	}

	if hatAktiveTSS(tssListe) {
		return TSESetupErgebnis{}, ErrTSEBereitsEingerichtet
	}

	identitaet, err := c.SettingsRepo.GetKassenidentitaet(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to retrieve kassenidentitaet for setup")
		return TSESetupErgebnis{}, ErrDatabase
	}
	seriennummer := identitaet.Seriennummer.String()

	// Der fiskaly-Client wird unter einer eigenen, frischen UUIDv4 als
	// Ressourcen-ID (_id) angelegt — fiskaly-Konvention. Die Kassen-Seriennummer
	// ist die fachliche serial_number (DSFinV-K KASSE_SERIENNR). So bleibt der
	// technische Client-Identifikator von der fachlichen Seriennummer getrennt
	// und konsistent mit der Uebernahme einer bestehenden TSS (spaetere Phase),
	// bei der die vorgefundene Client-_id uebernommen wird.
	clientID := uuid.NewString()

	pin, err := erzeugeAdminPIN()
	if err != nil {
		log.Error().Err(err).Msg("Failed to generate admin pin")
		return TSESetupErgebnis{}, ErrTSEEinrichtung
	}

	// Lebenszyklus CREATED -> UNINITIALIZED -> (PIN) -> INITIALIZED -> Client.
	// Eine frische TSS startet immer im Zustand CREATED; der PUK stammt direkt aus
	// der Anlage. Bricht ein Schritt ab, wird nichts gespeichert.
	erstellt, err := client.CreateTSS(ctx)
	if err != nil {
		return TSESetupErgebnis{}, einrichtungsFehler(log, err, "tss anlegen", "")
	}
	if err := vollendeLebenszyklus(ctx, log, client, "CREATED", erstellt.ID, erstellt.PUK, pin, clientID, seriennummer, false); err != nil {
		return TSESetupErgebnis{}, err
	}

	// Erst nach erfolgreichem Lebenszyklus wird die vollstaendige Konfiguration
	// atomar gespeichert.
	konfiguration, err := settings.NewTSEKonfiguration(credentials.ApiKey, credentials.ApiSecret, erstellt.ID, clientID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to build tse_konfiguration after setup")
		return TSESetupErgebnis{}, ErrTSEEinrichtung
	}
	if err := c.SettingsRepo.UpsertTSEKonfiguration(ctx, konfiguration); err != nil {
		// Der fiskaly-Lebenszyklus war erfolgreich, nur das Speichern schlug fehl:
		// Die TSS existiert bei fiskaly, ohne dass jotti sie kennt. tss_id/client_id
		// werden geloggt (PUK/PIN niemals), damit sich die TSS spaeter ueber die
		// Uebernahme (UebernimmTSE) wieder einsammeln laesst.
		log.Error().Err(err).Str("tss_id", erstellt.ID).Str("client_id", clientID).
			Msg("Failed to save tse_konfiguration after setup; TSS exists at fiskaly, recoverable via takeover")
		return TSESetupErgebnis{}, ErrDatabase
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

// UebernimmTSE uebernimmt eine im Befund gewaehlte, bereits vorhandene TSS und
// setzt sie aus ihrem aktuellen Zustand bis zum registrierten Client fort. Das
// ersetzt fuer vorhandene TSS die Verweigerung aus RichteTSEEin und dient
// zugleich der Wiederaufnahme nach einem Abbruch:
//
//   - CREATED: der PUK wird idempotent erneut bezogen und eine frische Admin-PIN
//     erzeugt; beide werden dem Admin einmalig angezeigt. Keine Nutzereingabe.
//   - ab UNINITIALIZED: der PUK ist nicht mehr abrufbar; die vom Admin verwahrte
//     Admin-PIN ist noetig (pin). Lehnt fiskaly sie ab, endet der Flow als
//     ErrTSESetupPINUnbekannt (Sackgasse mit Auswegen), nicht als technischer
//     Fehler. Es werden keine neuen Geheimnisse angezeigt.
//
// Ein bereits vorhandener Client mit passender Kassen-Seriennummer wird
// uebernommen statt neu registriert. Wie RichteTSEEin gilt der LIVE-Schutz
// (Abgleich bestaetigte vs. tatsaechliche Umgebung) und es wird erst nach
// erfolgreichem Abschluss atomar gespeichert.
func (c Command) UebernimmTSE(ctx context.Context, credentials tse.SetupCredentials, bestaetigteUmgebung tse.Umgebung, tssID, pin string) (TSESetupErgebnis, error) {
	log := zerolog.Ctx(ctx)

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

	client, err := c.NewTSESetupClient(credentials)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create TSE setup client")
		return TSESetupErgebnis{}, ErrTSEVerbindungFehlgeschlagen
	}

	umgebung, tssListe, err := client.ListTSS(ctx)
	if err != nil {
		if errors.Is(err, tse.ErrSetupAuthFehlgeschlagen) {
			return TSESetupErgebnis{}, ErrTSESetupZugangsdaten
		}
		log.Warn().Err(err).Msg("Failed to list TSS during takeover")
		return TSESetupErgebnis{}, ErrTSEVerbindungFehlgeschlagen
	}
	if umgebung != bestaetigteUmgebung {
		log.Warn().
			Str("bestaetigt", string(bestaetigteUmgebung)).
			Str("tatsaechlich", string(umgebung)).
			Msg("Confirmed TSE environment does not match actual environment")
		return TSESetupErgebnis{}, ErrTSESetupUmgebungAbweichung
	}

	ziel, gefunden := findeTSS(tssListe, tssID)
	if !gefunden {
		return TSESetupErgebnis{}, ErrTSESetupTSSNichtGefunden
	}
	state := strings.ToUpper(strings.TrimSpace(ziel.State))

	identitaet, err := c.SettingsRepo.GetKassenidentitaet(ctx)
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

	// Einen passenden Client uebernehmen, sonst einen neuen unter eigener UUID
	// registrieren (wie bei der Neuanlage).
	clientID := uuid.NewString()
	hatClient := false
	if passender := passenderClient(clients, seriennummer); passender != nil {
		clientID = passender.ID
		hatClient = true
	}

	// PUK/PIN-Strategie nach Zustand (siehe Methodenkommentar).
	var puk, ergebnisPUK, ergebnisPIN string
	switch state {
	case "CREATED":
		puk, err = client.HoleAdminPUK(ctx, tssID)
		if err != nil {
			return TSESetupErgebnis{}, einrichtungsFehler(log, err, "puk beziehen", tssID)
		}
		pin, err = erzeugeAdminPIN()
		if err != nil {
			log.Error().Err(err).Msg("Failed to generate admin pin")
			return TSESetupErgebnis{}, ErrTSEEinrichtung
		}
		ergebnisPUK, ergebnisPIN = puk, pin
	case "UNINITIALIZED", "INITIALIZED":
		if strings.TrimSpace(pin) == "" {
			return TSESetupErgebnis{}, ErrTSESetupPINErforderlich
		}
	default:
		return TSESetupErgebnis{}, ErrTSESetupUebernahmeNichtMoeglich
	}

	if err := vollendeLebenszyklus(ctx, log, client, state, tssID, puk, pin, clientID, seriennummer, hatClient); err != nil {
		return TSESetupErgebnis{}, err
	}

	konfiguration, err := settings.NewTSEKonfiguration(credentials.ApiKey, credentials.ApiSecret, tssID, clientID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to build tse_konfiguration after takeover")
		return TSESetupErgebnis{}, ErrTSEEinrichtung
	}
	if err := c.SettingsRepo.UpsertTSEKonfiguration(ctx, konfiguration); err != nil {
		// Wie bei der Neuanlage: Lebenszyklus erfolgreich, nur das Speichern schlug
		// fehl. Die TSS bleibt bei fiskaly und laesst sich erneut uebernehmen.
		log.Error().Err(err).Str("tss_id", tssID).Str("client_id", clientID).
			Msg("Failed to save tse_konfiguration after takeover; TSS exists at fiskaly, recoverable via takeover")
		return TSESetupErgebnis{}, ErrDatabase
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

// vollendeLebenszyklus treibt eine TSS von ihrem aktuellen Zustand bis zum
// registrierten Client. puk wird nur im Zustand CREATED gebraucht (Setzen der
// frischen Admin-PIN); ab UNINITIALIZED traegt pin die vorhandene Admin-PIN.
// hatClient ueberspringt die Registrierung, wenn bereits ein passender Client
// existiert. Schlaegt die Admin-Authentifizierung mit einer vom Nutzer
// eingegebenen PIN fehl (Ausgangszustand != CREATED), endet der Flow als
// ErrTSESetupPINUnbekannt statt als technischer Fehler.
func vollendeLebenszyklus(ctx context.Context, log *zerolog.Logger, client tse.SetupClient, state, tssID, puk, pin, clientID, seriennummer string, hatClient bool) error {
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
		if err := client.SetzeAdminPIN(ctx, tssID, puk, pin); err != nil {
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
		if err := client.AuthentifiziereAdmin(ctx, tssID, pin); err != nil {
			return authFehler(err, "admin-auth (client)")
		}
		if !hatClient {
			if err := client.RegistriereClient(ctx, tssID, clientID, seriennummer); err != nil {
				return einrichtungsFehler(log, err, "client registrieren", tssID)
			}
		}
	default:
		return ErrTSESetupUebernahmeNichtMoeglich
	}
	return nil
}

// findeTSS sucht eine TSS nach ihrer ID in der Konto-Liste.
func findeTSS(tssListe []tse.TSSInfo, tssID string) (tse.TSSInfo, bool) {
	for _, t := range tssListe {
		if t.ID == tssID {
			return t, true
		}
	}
	return tse.TSSInfo{}, false
}

// einrichtungsFehler protokolliert einen fehlgeschlagenen Lebenszyklus-Schritt
// (ohne PUK/PIN) und liefert das einheitliche Einrichtungs-Sentinel.
func einrichtungsFehler(log *zerolog.Logger, err error, schritt, tssID string) error {
	log.Warn().Err(err).Str("schritt", schritt).Str("tss_id", tssID).Msg("TSE setup step failed")
	return ErrTSEEinrichtung
}

// hatAktiveTSS meldet, ob das Konto eine noch nutzbare TSS enthaelt. Deaktivierte
// (DISABLED) TSS gelten als tot und blockieren die Neuanlage nicht.
func hatAktiveTSS(tssListe []tse.TSSInfo) bool {
	for _, t := range tssListe {
		if !strings.EqualFold(strings.TrimSpace(t.State), "DISABLED") {
			return true
		}
	}
	return false
}

// erzeugeAdminPIN erzeugt eine kryptografisch zufaellige numerische Admin-PIN.
func erzeugeAdminPIN() (string, error) {
	var sb strings.Builder
	for i := 0; i < adminPINStellen; i++ {
		ziffer, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			return "", err
		}
		sb.WriteByte(byte('0' + ziffer.Int64()))
	}
	return sb.String(), nil
}
