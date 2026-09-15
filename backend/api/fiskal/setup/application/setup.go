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

// adminPINStellen ist die Länge der zufällig erzeugten Admin-PIN. Zehn Ziffern
// liegen sicher innerhalb der von fiskaly akzeptierten Länge.
const adminPINStellen = 10

// einrichtungLaeuft hält fest, ob gerade jemand an der TSE-Konfiguration
// schreibt, und trägt damit die fachliche Invariante "es schreibt höchstens
// einer auf der TSE-Konfiguration". Alle drei Schreibpfade nehmen es:
// RichteTSEEin, UebernimmTSE und UpdateTSEKonfiguration (command.go) — sie
// enden alle in SaveEinrichtung.
//
// Nötig, seit der Lebenszyklus vom Client-Abbruch entkoppelt ist
// (lebenszyklusKontext in backend/api/fiskal/setup/http/command_handler.go): Er
// läuft nach einem Abbruch im Hintergrund weiter, während der Admin bereits
// eine Fehlermeldung sieht und sofort erneut starten kann. Ohne diese Sperre
// sähe der zweite Aufruf in ListTSS noch das leere Konto, hatAktiveTSS meldete
// false, und er legte eine ZWEITE bezahlte LIVE-TSS an. Beide Läufe endeten in
// saveEinrichtung, der zweite überschriebe den ersten — die dem Admin
// angezeigten PUK und Admin-PIN gehörten dann zur nicht konfigurierten TSS.
//
// Derselbe Ausgang droht ohne den fiskaly-Umweg: Der manuelle
// Zugangsdaten-Wechsel liegt in der Oberfläche direkt unter dem Wizard
// (frontend/src/admin/tse/TSEEinrichtungPage.tsx). Speichert der Admin dort von
// Hand, während die Einrichtung im Hintergrund noch läuft, gewinnt der letzte
// Schreiber, und die Instanz signiert anschließend gegen eine TSS/Client-
// Kombination, die nicht die eingerichtete ist.
//
// Ein prozessinternes Schloss genügt: jotti läuft je Verein als eine einzige
// Backend-Instanz (Docker Compose), es gibt keine zweite Instanz, gegen die zu
// koordinieren wäre. Ein atomarer Schalter statt eines Mutex, weil der zweite
// Aufruf nicht warten, sondern sofort mit ErrTSESetupLaeuftBereits abbrechen
// soll. Er liegt auf Paketebene und nicht als Feld in Command: Command hat
// Wert-Empfänger, ein Wert-Feld wäre pro Methodenaufruf eine eigene Kopie und
// damit wirkungslos. Ein Zeiger-Feld (*atomic.Bool, einmal in
// backend/api/admin.go befüllt) wäre prozessweit dasselbe Schloss und damit
// korrekt — aber unnötige Verdrahtung mit einer Nil-Falle für jeden, der ein
// Command ohne dieses Feld baut.
var einrichtungLaeuft atomic.Bool

// acquireEinrichtung reserviert das Schreibrecht auf der TSE-Konfiguration und
// liefert die Freigabe dazu; schreibt bereits jemand, endet der Aufruf sofort
// mit ErrTSESetupLaeuftBereits. Die Freigabe gehört in ein defer, damit auch
// jeder Fehlerpfad und eine Panik das Schloss wieder lösen.
func acquireEinrichtung() (func(), error) {
	if !einrichtungLaeuft.CompareAndSwap(false, true) {
		return nil, ErrTSESetupLaeuftBereits
	}
	return func() { einrichtungLaeuft.Store(false) }, nil
}

// TSESetupErgebnis ist das Ergebnis der geführten Einrichtung. PUK und AdminPIN
// erscheinen genau hier — sie werden weder persistiert noch geloggt und nur
// einmalig an die UI übergeben, damit der Admin sie extern verwahren kann.
type TSESetupErgebnis struct {
	TssID    string
	ClientID string
	PUK      string
	AdminPIN string
	Umgebung string
}

// RichteTSEEin führt den vollständigen fiskaly-Lebenszyklus für ein leeres Konto
// durch: TSS anlegen, personalisieren, Admin-PIN setzen, initialisieren, Client
// mit der Kassen-Seriennummer registrieren. Gespeichert wird erst nach
// erfolgreichem Abschluss — ein Abbruch hinterlässt keine halbe Konfiguration.
//
// Weicht bestaetigteUmgebung von der tatsächlichen ab, bricht die Einrichtung vor
// jeder Schreiboperation ab (Schutz vor versehentlicher LIVE-Anlage). Existiert
// bereits eine aktive TSS, wird die Neuanlage verweigert — außer der Admin
// erzwingt sie in TEST per neuAnlegenTrotzVorhandener: dort darf bewusst eine
// zweite, frische TSE entstehen. In LIVE bleibt die Sperre hart.
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

	// In TEST darf der Admin die Sperre bewusst übergehen; in LIVE nie — eine zweite
	// LIVE-TSS verursacht laufende Kosten. umgebung ist hier bereits gegen
	// bestaetigteUmgebung abgeglichen und damit autoritativ.
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
	// und konsistent mit der Übernahme einer bestehenden TSS, bei der die
	// vorgefundene Client-_id übernommen wird.
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
		// Das fiskaly-TSS-Limit (in TEST fünf aktive TSS) ist kein technischer
		// Fehler, sondern ein verständlich zu meldender Zustand.
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

// UebernimmTSE übernimmt eine vorhandene TSS und setzt sie aus ihrem aktuellen
// Zustand bis zum registrierten Client fort — auch als Wiederaufnahme nach einem
// Abbruch:
//
//   - CREATED: der PUK wird idempotent erneut bezogen und eine frische Admin-PIN
//     erzeugt; beide werden dem Admin einmalig angezeigt. Keine Nutzereingabe.
//   - INITIALIZED mit passendem, bereits REGISTERED Client: einsatzbereit. Es folgt
//     keine privilegierte fiskaly-Operation, also keine Admin-PIN nötig — jotti
//     speichert nur noch die Konfiguration.
//   - ab UNINITIALIZED (bzw. INITIALIZED ohne fertigen Client): der PUK ist nicht
//     mehr abrufbar, die verwahrte Admin-PIN ist nötig (pin). Lehnt fiskaly sie ab,
//     endet der Flow als ErrTSESetupPINUnbekannt. Es werden keine neuen Geheimnisse
//     angezeigt.
//   - ab UNINITIALIZED mit Admin-PUK (puk): die PIN ist verloren oder nach fünf
//     Fehlversuchen gesperrt. jotti setzt mit dem PUK eine frische Zufalls-PIN
//     (einmalig angezeigt) und fährt fort; der PUK bleibt unverändert und wird nicht
//     erneut angezeigt, ein falscher endet als ErrTSESetupPUKUnbekannt.
//
// Ein passender REGISTERED Client wird unverändert übernommen, ein DEREGISTERED
// reaktiviert statt neu angelegt (serial_number ist je TSS eindeutig). Wie bei der
// Neuanlage gilt der LIVE-Schutz, und gespeichert wird erst nach Erfolg.
func (c Command) UebernimmTSE(ctx context.Context, credentials tse.SetupCredentials, bestaetigteUmgebung tse.Umgebung, tssID, pin, puk string) (TSESetupErgebnis, error) {
	log := zerolog.Ctx(ctx)

	// Dasselbe Schloss wie die Neuanlage: Beide Pfade führen denselben
	// fiskaly-Lebenszyklus und dürfen sich nicht überlappen.
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

	// Eine INITIALIZED TSS mit fertigem (REGISTERED) Client ist einsatzbereit: keine
	// privilegierte fiskaly-Operation, daher keine Admin-PIN nötig. Jeder andere Pfad
	// löst eine Admin-Operation aus und braucht die PIN.
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
			// Verlorene oder gesperrte PIN per PUK zurücksetzen: eine frische
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

// oeffneSetupClient baut den fiskaly-Setup-Client, liest die TSS-Liste und
// hält den LIVE-Schutz: Weicht die tatsächliche Umgebung von der bestätigten
// ab, endet der Aufruf vor jeder Schreiboperation. zweck geht allein in die
// Log-Meldung ("setup" / "takeover").
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

// saveEinrichtung ist der gemeinsame Speicher-Schritt aller Einrichtungspfade:
// nach erfolgreichem fiskaly-Lebenszyklus wird die TSE-Konfiguration atomar
// gespeichert und die fiskalischen TSS-Stammdaten für den DSFinV-K-Export
// nachgezogen. Schlägt das Speichern fehl, ist die Einrichtung nicht
// abgeschlossen: die TSS existiert bei fiskaly (per Übernahme einsammelbar),
// tss_id/client_id werden geloggt (PUK/PIN niemals).
func (c Command) saveEinrichtung(ctx context.Context, log *zerolog.Logger, client tse.SetupClient, credentials tse.SetupCredentials, tssID, clientID string) error {
	konfiguration, err := tse.NewKonfiguration(credentials.ApiKey, credentials.ApiSecret, tssID, clientID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to build tse_konfiguration after setup")
		return ErrTSEEinrichtung
	}
	// SaveEinrichtung speichert die Konfiguration und markiert beim Übergang
	// von nicht konfiguriert zu konfiguriert in derselben Transaktion die noch
	// offenen, vor-konfigurationellen Aufträge endgültig (Einrichtungs-Sweep)
	// und schließt den keine_konfiguration-Störungszeitraum.
	if err := c.TSERepo.SaveEinrichtung(ctx, konfiguration); err != nil {
		log.Error().Err(err).Str("tss_id", tssID).Str("client_id", clientID).
			Msg("Failed to save tse_konfiguration after setup; TSS exists at fiskaly, recoverable via takeover")
		return ErrDatabase
	}

	return c.fetchTSEStammdaten(ctx, log, client, tssID)
}

// fetchTSEStammdaten liest die fiskalischen TSS-Stammdaten von fiskaly und
// speichert sie für den DSFinV-K-Export. Die Stammdaten enthalten die
// TSS-Seriennummer (TSE_SERIAL in der DSFinV-K) sowie Public Key und Zertifikat,
// die der Export allein aus tse_stammdaten liest; daher ist ein Fehler hier ein
// harter Einrichtungsfehler.
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
// frischen Admin-PIN); ab UNINITIALIZED trägt pin die vorhandene Admin-PIN. Bei
// clientFertig entfällt der privilegierte Client-Schritt samt
// Admin-Authentifizierung. Schlägt die Authentifizierung mit einer vom Nutzer
// eingegebenen PIN fehl (Ausgangszustand != CREATED), endet der Flow als
// ErrTSESetupPINUnbekannt.
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
		// Ein bereits REGISTERED Client ist fertig — keine fiskaly-Mutation, keine
		// Admin-Authentifizierung. Aus CREATED/UNINITIALIZED fällt der Code nie mit
		// clientFertig hier ein, da es dann keinen vorhandenen Client gibt.
		if aktion == clientFertig {
			return nil
		}
		if err := client.AuthentifiziereAdmin(ctx, tssID, pin); err != nil {
			return authFehler(err, "admin-auth (client)")
		}
		// Ein DEREGISTERED Client wird per state=REGISTERED reaktiviert statt neu
		// angelegt — die serial_number ist je TSS eindeutig.
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

// einrichtungsFehler protokolliert einen fehlgeschlagenen Lebenszyklus-Schritt
// (ohne PUK/PIN) und liefert das einheitliche Einrichtungs-Sentinel.
func einrichtungsFehler(log *zerolog.Logger, err error, schritt, tssID string) error {
	log.Warn().Err(err).Str("schritt", schritt).Str("tss_id", tssID).Msg("TSE setup step failed")
	return ErrTSEEinrichtung
}

// hatAktiveTSS meldet, ob das Konto eine nicht deaktivierte TSS enthält. Nur
// deaktivierte (DISABLED) TSS gelten als tot und blockieren die Neuanlage nicht.
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
	for i := 0; i < adminPINStellen; i++ {
		ziffer, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			return "", err
		}
		sb.WriteByte(byte('0' + ziffer.Int64()))
	}
	return sb.String(), nil
}
