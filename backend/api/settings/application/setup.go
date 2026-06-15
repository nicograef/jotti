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
	// Vor den Admin-Operationen (Initialisieren, Client registrieren) wird das
	// Token jeweils auf Admin-Rechte angehoben. Bricht ein Schritt ab, wird
	// nichts gespeichert.
	erstellt, err := client.CreateTSS(ctx)
	if err != nil {
		return TSESetupErgebnis{}, einrichtungsFehler(log, err, "tss anlegen", "")
	}
	if err := client.PersonalisiereTSS(ctx, erstellt.ID); err != nil {
		return TSESetupErgebnis{}, einrichtungsFehler(log, err, "personalisieren", erstellt.ID)
	}
	if err := client.SetzeAdminPIN(ctx, erstellt.ID, erstellt.PUK, pin); err != nil {
		return TSESetupErgebnis{}, einrichtungsFehler(log, err, "admin-pin setzen", erstellt.ID)
	}
	if err := client.AuthentifiziereAdmin(ctx, erstellt.ID, pin); err != nil {
		return TSESetupErgebnis{}, einrichtungsFehler(log, err, "admin-auth (init)", erstellt.ID)
	}
	if err := client.InitialisiereTSS(ctx, erstellt.ID); err != nil {
		return TSESetupErgebnis{}, einrichtungsFehler(log, err, "initialisieren", erstellt.ID)
	}
	if err := client.AuthentifiziereAdmin(ctx, erstellt.ID, pin); err != nil {
		return TSESetupErgebnis{}, einrichtungsFehler(log, err, "admin-auth (client)", erstellt.ID)
	}
	if err := client.RegistriereClient(ctx, erstellt.ID, clientID, seriennummer); err != nil {
		return TSESetupErgebnis{}, einrichtungsFehler(log, err, "client registrieren", erstellt.ID)
	}

	// Erst nach erfolgreichem Lebenszyklus wird die vollstaendige Konfiguration
	// atomar gespeichert.
	konfiguration, err := settings.NewTSEKonfiguration(credentials.ApiKey, credentials.ApiSecret, erstellt.ID, clientID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to build tse_konfiguration after setup")
		return TSESetupErgebnis{}, ErrTSEEinrichtung
	}
	if err := c.SettingsRepo.UpsertTSEKonfiguration(ctx, konfiguration); err != nil {
		log.Error().Err(err).Msg("Failed to save tse_konfiguration after setup")
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
