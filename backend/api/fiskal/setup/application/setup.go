package application

import (
	"context"
	"crypto/rand"
	"errors"
	"math/big"
	"strings"

	"github.com/google/uuid"
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
// wird die Neuanlage verweigert — ausser der Admin erzwingt sie in TEST per
// neuAnlegenTrotzVorhandener (F2): in der kostenlosen, steuerlich ungueltigen
// TEST-Umgebung darf so bewusst eine zweite, frische TSE entstehen (etwa wenn
// die vorhandene PIN-los und damit nicht uebernehmbar ist). In LIVE bleibt die
// Sperre hart.
func (c Command) RichteTSEEin(ctx context.Context, credentials tse.SetupCredentials, bestaetigteUmgebung tse.Umgebung, neuAnlegenTrotzVorhandener bool) (TSESetupErgebnis, error) {
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
	if err := c.ensureKeineOffeneKassensitzung(ctx); err != nil {
		return TSESetupErgebnis{}, err
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

	// In TEST darf der Admin die Sperre bewusst uebergehen (F2); in LIVE nie —
	// eine zweite LIVE-TSS verursacht laufende Kosten. umgebung ist hier bereits
	// gegen bestaetigteUmgebung abgeglichen und damit autoritativ.
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

	// Der fiskaly-Client wird unter einer eigenen, frischen UUIDv4 als
	// Ressourcen-ID (_id) angelegt — fiskaly-Konvention. Die Kassen-Seriennummer
	// ist die fachliche serial_number (DSFinV-K KASSE_SERIENNR). So bleibt der
	// technische Client-Identifikator von der fachlichen Seriennummer getrennt
	// und konsistent mit der Uebernahme einer bestehenden TSS (spaetere Phase),
	// bei der die vorgefundene Client-_id uebernommen wird.
	clientID := uuid.NewString()

	pin, err := generateAdminPIN()
	if err != nil {
		log.Error().Err(err).Msg("Failed to generate admin pin")
		return TSESetupErgebnis{}, ErrTSEEinrichtung
	}

	// Lebenszyklus CREATED -> UNINITIALIZED -> (PIN) -> INITIALIZED -> Client.
	// Eine frische TSS startet immer im Zustand CREATED; der PUK stammt direkt aus
	// der Anlage. Bricht ein Schritt ab, wird nichts gespeichert.
	erstellt, err := client.CreateTSS(ctx)
	if err != nil {
		// Das fiskaly-TSS-Limit (in TEST fuenf aktive TSS) ist kein technischer
		// Fehler, sondern ein verstaendlich zu meldender Zustand.
		if errors.Is(err, tse.ErrSetupTSSLimitErreicht) {
			return TSESetupErgebnis{}, ErrTSESetupTSSLimitErreicht
		}
		return TSESetupErgebnis{}, einrichtungsFehler(log, err, "tss anlegen", "")
	}
	if err := vollendeLebenszyklus(ctx, log, client, "CREATED", erstellt.ID, erstellt.PUK, pin, clientID, seriennummer, clientRegistrieren); err != nil {
		return TSESetupErgebnis{}, err
	}

	// Erst nach erfolgreichem Lebenszyklus wird die vollstaendige Konfiguration
	// atomar gespeichert (gemeinsamer Speicher-Schritt aller Einrichtungspfade).
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

// UebernimmTSE uebernimmt eine im Befund gewaehlte, bereits vorhandene TSS und
// setzt sie aus ihrem aktuellen Zustand bis zum registrierten Client fort. Das
// ersetzt fuer vorhandene TSS die Verweigerung aus RichteTSEEin und dient
// zugleich der Wiederaufnahme nach einem Abbruch:
//
//   - CREATED: der PUK wird idempotent erneut bezogen und eine frische Admin-PIN
//     erzeugt; beide werden dem Admin einmalig angezeigt. Keine Nutzereingabe.
//   - INITIALIZED mit passendem, bereits REGISTERED Client: die TSS ist faktisch
//     einsatzbereit. Es folgt keine privilegierte fiskaly-Operation, daher ist
//     keine Admin-PIN noetig (F8) — jotti speichert nur noch die Konfiguration.
//   - ab UNINITIALIZED (bzw. INITIALIZED ohne fertigen Client): der PUK ist nicht
//     mehr abrufbar; die vom Admin verwahrte Admin-PIN ist noetig (pin). Lehnt
//     fiskaly sie ab, endet der Flow als ErrTSESetupPINUnbekannt (Sackgasse mit
//     Auswegen), nicht als technischer Fehler. Es werden keine neuen Geheimnisse
//     angezeigt.
//   - ab UNINITIALIZED mit uebergebenem Admin-PUK (puk): die PIN ist verloren oder
//     nach fuenf Fehlversuchen gesperrt. jotti setzt mit dem PUK eine frische
//     Zufalls-PIN (zeigt sie einmalig an) und fuehrt die Uebernahme damit fort.
//     Der PUK bleibt unveraendert und wird nicht erneut angezeigt; ein falscher
//     PUK endet als ErrTSESetupPUKUnbekannt.
//
// Ein passender REGISTERED Client wird unveraendert uebernommen; ein passender,
// aber DEREGISTERED Client wird reaktiviert (state=REGISTERED) statt neu
// angelegt, da die serial_number je TSS eindeutig ist. Wie RichteTSEEin gilt der
// LIVE-Schutz (Abgleich bestaetigte vs. tatsaechliche Umgebung) und es wird erst
// nach erfolgreichem Abschluss atomar gespeichert.
func (c Command) UebernimmTSE(ctx context.Context, credentials tse.SetupCredentials, bestaetigteUmgebung tse.Umgebung, tssID, pin, puk string) (TSESetupErgebnis, error) {
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
	if err := c.ensureKeineOffeneKassensitzung(ctx); err != nil {
		return TSESetupErgebnis{}, err
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

	// Was mit dem Client zu tun ist: ein passender REGISTERED Client ist fertig,
	// ein passender DEREGISTERED Client wird reaktiviert (nicht neu angelegt — die
	// serial_number ist je TSS eindeutig), sonst wird ein neuer Client unter
	// eigener UUID registriert (wie bei der Neuanlage).
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

	// Eine INITIALIZED TSS mit fertigem (REGISTERED) Client ist faktisch
	// einsatzbereit: es folgt keine privilegierte fiskaly-Operation, daher ist
	// keine Admin-PIN noetig (F8). Jeder andere Pfad loest eine Admin-Operation
	// aus (Initialisieren, Client registrieren/reaktivieren) und braucht die PIN.
	einsatzbereit := state == "INITIALIZED" && aktion == clientFertig

	// PUK/PIN-Strategie nach Zustand (siehe Methodenkommentar). lebenszyklusPUK
	// treibt nur den CREATED-Schritt (Setzen der ersten PIN); ergebnisPUK/
	// ergebnisPIN sind die einmalig anzuzeigenden, neu entstandenen Geheimnisse.
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
			// Verlorene oder gesperrte PIN per PUK zuruecksetzen: eine frische
			// Zufalls-PIN mit dem PUK setzen und damit fortfahren. Die Zugangsdaten
			// sind durch ListTSS bereits bestaetigt, daher ist ein Fehler hier
			// praktisch immer ein falscher PUK.
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

// saveEinrichtung ist der gemeinsame Speicher-Schritt aller
// Einrichtungspfade (Neuanlage, Uebernahme, F8-Uebernahme und PUK-Reset): nach
// erfolgreichem fiskaly-Lebenszyklus wird die TSE-Konfiguration atomar
// gespeichert und — best effort — die fiskalischen TSS-Stammdaten fuer den
// DSFinV-K-Export nachgezogen. Schlaegt das Speichern der Konfiguration fehl, ist
// die Einrichtung nicht abgeschlossen: die TSS existiert bei fiskaly (per
// Uebernahme einsammelbar), tss_id/client_id werden geloggt (PUK/PIN niemals).
func (c Command) saveEinrichtung(ctx context.Context, log *zerolog.Logger, client tse.SetupClient, credentials tse.SetupCredentials, tssID, clientID string) error {
	konfiguration, err := tse.NewKonfiguration(credentials.ApiKey, credentials.ApiSecret, tssID, clientID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to build tse_konfiguration after setup")
		return ErrTSEEinrichtung
	}
	// SaveEinrichtung speichert die Konfiguration und markiert beim Uebergang
	// von nicht konfiguriert zu konfiguriert in derselben Transaktion die noch
	// offenen, vor-konfigurationellen Auftraege endgueltig (Einrichtungs-Sweep)
	// und schliesst den keine_konfiguration-Stoerungszeitraum.
	if err := c.TSERepo.SaveEinrichtung(ctx, konfiguration); err != nil {
		log.Error().Err(err).Str("tss_id", tssID).Str("client_id", clientID).
			Msg("Failed to save tse_konfiguration after setup; TSS exists at fiskaly, recoverable via takeover")
		return ErrDatabase
	}

	c.fetchTSEStammdaten(ctx, log, client, tssID)
	return nil
}

// fetchTSEStammdaten liest die fiskalischen TSS-Stammdaten von fiskaly und
// speichert sie fuer den DSFinV-K-Export. Best effort: jeder Fehler beim Abruf
// oder Speichern wird nur protokolliert, der Setup-Erfolg bleibt bestehen — die
// Stammdaten lassen sich beim naechsten Verbinden nachziehen.
func (c Command) fetchTSEStammdaten(ctx context.Context, log *zerolog.Logger, client tse.SetupClient, tssID string) {
	gelesen, err := client.RetrieveTSSStammdaten(ctx, tssID)
	if err != nil {
		log.Warn().Err(err).Str("tss_id", tssID).Msg("Failed to fetch TSE Stammdaten after setup; recoverable on next connect")
		return
	}
	stammdaten := tse.NewStammdaten(gelesen.SignaturAlgorithmus, gelesen.PublicKey, gelesen.Zertifikat, gelesen.LogTimeFormat)
	if err := c.TSERepo.UpsertTSEStammdaten(ctx, stammdaten); err != nil {
		log.Warn().Err(err).Str("tss_id", tssID).Msg("Failed to save TSE Stammdaten after setup; recoverable on next connect")
	}
}

// clientAktion beschreibt, was im Client-Schritt einer INITIALIZED TSS zu tun
// ist: einen neuen Client registrieren (kein passender vorhanden), einen
// vorhandenen DEREGISTERED Client reaktivieren, oder nichts (passender Client
// ist bereits REGISTERED — einsatzbereit).
type clientAktion int

const (
	clientRegistrieren clientAktion = iota
	clientReaktivieren
	clientFertig
)

// vollendeLebenszyklus treibt eine TSS von ihrem aktuellen Zustand bis zum
// registrierten Client. puk wird nur im Zustand CREATED gebraucht (Setzen der
// frischen Admin-PIN); ab UNINITIALIZED traegt pin die vorhandene Admin-PIN.
// aktion steuert den Client-Schritt (registrieren/reaktivieren/fertig). Ein
// bereits REGISTERED Client (clientFertig) ist einsatzbereit: dann faellt der
// privilegierte Client-Schritt samt Admin-Authentifizierung weg (F8). Schlaegt
// die Admin-Authentifizierung mit einer vom Nutzer eingegebenen PIN fehl
// (Ausgangszustand != CREATED), endet der Flow als ErrTSESetupPINUnbekannt statt
// als technischer Fehler.
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
		// Ein bereits REGISTERED Client ist fertig — keine fiskaly-Mutation und
		// damit keine Admin-Authentifizierung noetig (F8, „einsatzbereit ohne
		// Arbeit"). Aus CREATED/UNINITIALIZED faellt der Code nie hier mit
		// clientFertig ein, da es dann keinen vorhandenen Client gibt.
		if aktion == clientFertig {
			return nil
		}
		if err := client.AuthentifiziereAdmin(ctx, tssID, pin); err != nil {
			return authFehler(err, "admin-auth (client)")
		}
		// Ein DEREGISTERED Client wird per state=REGISTERED reaktiviert statt neu
		// angelegt — die serial_number ist je TSS eindeutig (F7).
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

// findTSS sucht eine TSS nach ihrer ID in der Konto-Liste.
func findTSS(tssListe []tse.TSSInfo, tssID string) (tse.TSSInfo, bool) {
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

// generateAdminPIN erzeugt eine kryptografisch zufaellige numerische Admin-PIN.
func generateAdminPIN() (string, error) {
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
