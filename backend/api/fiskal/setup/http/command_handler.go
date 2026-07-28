package http

import (
	"context"
	"errors"
	"net/http"
	"time"

	z "github.com/Oudwins/zog"
	"github.com/nicograef/jotti/backend/api/fiskal/setup/application"
	"github.com/nicograef/jotti/backend/api/helper"
	"github.com/nicograef/jotti/backend/domain/tse"
)

// tseSetupWriteTimeout ersetzt fuer die vier fiskaly-Endpunkte dieses Pakets
// die globale 10-Sekunden-Schreibfrist des Servers (backend/app/app.go). Sie
// alle sprechen synchron mit fiskaly: Die Neuanlage setzt im schlimmsten Fall
// zehn HTTP-Sequenzen nacheinander ab (Auth, ListTSS, CreateTSS,
// personalisieren, PIN setzen, zweimal Admin-Auth, initialisieren, Client
// registrieren, Stammdaten), jede mit 10 s Zeitlimit und bis zu vier Versuchen
// (defaultRetryAttempts = 3) mit Backoff (defaultHTTPTimeout,
// defaultRetryAttempts in backend/repository/tse_repo/fiskaly_client.go). Zwei
// Minuten decken den realistischen Verlauf ab; der Wert ist aus den Timeout-
// und Retry-Budgets abgeleitet, nicht gemessen.
//
// Die Frist begrenzt allein den Schreibvorgang der Antwort. Sie bricht keinen
// Handler ab: Laeuft sie ab, scheitert nur ein gerade laufender Schreibvorgang,
// die Arbeit im Handler laeuft davon unberuehrt weiter. Der einzige
// serverseitige Aufgabepunkt der beiden schreibenden Endpunkte ist der
// Leck-Waechter unten.
//
// Das Zeitlimit des Clients liegt darueber: TSE_TIMEOUT_MS = 150 s in
// frontend/src/admin/tse/TSEBackend.ts, also diese zwei Minuten plus 30 s
// Netzreserve. Der Grund ist die Antwort, nicht die Reihenfolge des Aufgebens:
// Ein spaet, aber erfolgreich geschriebener Antwort-Body soll den Client noch
// erreichen. Waere sein Budget nicht groesser als das Schreibbudget des
// Servers, haette er in genau dem Fenster schon aufgegeben, in dem der Server
// gerade noch schreibt — und PUK und Admin-PIN waeren verloren, obwohl sie
// unterwegs waren.
const tseSetupWriteTimeout = 2 * time.Minute

// tseSetupLebenszyklusTimeout ist der Leck-Waechter um den fiskaly-Lebenszyklus
// der beiden schreibenden Endpunkte — KEIN Reaktionszeit-Budget. Wie lange ein
// Admin auf eine Antwort wartet, entscheidet allein der Client
// (TSE_TIMEOUT_MS). Diese Frist verhindert ausschliesslich, dass eine haengende
// fiskaly-Verbindung den vom Request abgekoppelten Lebenszyklus dauerhaft
// offenhaelt.
//
// Der Wert liegt deshalb weit ueber dem Worst Case: Die Uebernahme setzt bis zu
// elf HTTP-Sequenzen nacheinander ab (Auth, ListTSS, ListClients, PUK beziehen
// bzw. PIN setzen, personalisieren, PIN setzen, zweimal Admin-Auth,
// initialisieren, Client registrieren, Stammdaten). Jede davon hat 10 s
// HTTP-Zeitlimit und bis zu vier Versuche (defaultRetryAttempts = 3) mit
// Backoff von 0,2 + 0,4 + 0,8 s — also rund 41 s, in Summe rund 7,5 Minuten.
// Zehn Minuten liegen darueber. Zu knapp gewaehlt waere dieser Waechter der
// Blocker in neuer Form: Er schnitte einen laufenden Lebenszyklus mittendrin
// ab.
//
// Bekannte Luecke: Ein von fiskaly geliefertes Retry-After uebernimmt
// parseRetryAfter ungedeckelt (retryDelay in
// backend/repository/tse_repo/fiskaly_client.go). Ein unbeschraenkter Wert
// laesst sich durch keinen festen Abstand decken — er sprengt die Rechnung
// oben. Was ihn begrenzt, ist dieser Waechter selbst: Das Warten haengt am
// Kontext (sleepWithContext) und endet mit ihm, dann allerdings mitten im
// Lebenszyklus.
const tseSetupLebenszyklusTimeout = 10 * time.Minute

// lebenszyklusKontext liefert den Kontext, unter dem die beiden schreibenden
// TSE-Endpunkte ihren fiskaly-Lebenszyklus fahren. Er ist bewusst vom
// Request-Kontext abgekoppelt: Schliesst der Client die Verbindung — Tab zu,
// Zeitlimit abgelaufen, WLAN weg —, storniert net/http r.Context(), und der
// Lebenszyklus braeche mitten in der fiskaly-Sequenz ab. Zurueck bliebe eine
// bezahlte, halbfertige TSS: hatAktiveTSS blockiert den zweiten
// Einrichtungsversuch mit tse_bereits_eingerichtet, und die Uebernahme
// scheitert an der Admin-PIN, die es nur in der verlorenen Antwort gab (PUK und
// Admin-PIN werden nirgends persistiert, siehe
// backend/api/fiskal/setup/application/setup.go). Der Lebenszyklus muss also
// auch ohne Zuhoerer zu Ende laufen und speichern — erst saveEinrichtung
// schreibt tssId, clientId und Zugangsdaten und macht die Instanz
// betriebsfaehig.
//
// context.WithoutCancel erhaelt die Kontext-Werte (Korrelations-ID aus
// CorrelationIDMiddleware, zerolog-Logger aus LoggingMiddleware) und nimmt nur
// die Stornierung weg; darueber liegt tseSetupLebenszyklusTimeout als
// Leck-Waechter.
//
// Abgekoppelt ist der Kontext, nicht der Ablauf: Der Handler startet keine
// Goroutine, sondern faehrt den Lebenszyklus synchron und kehrt erst mit ihm
// zurueck.
//
// Die Zusage gilt deshalb genau fuer den Client-Abbruch, nicht fuer ein
// Prozessende: Ein Deploy oder Neustart wartet ueber http.Server.Shutdown bis zu
// 30 s auf den noch laufenden Handler (backend/app/app.go); erst ein danach
// immer noch laufender Lebenszyklus wird mit dem Prozess mitgerissen, und der
// Endzustand ist wieder der Blocker — bezahlte TSS, hatAktiveTSS sperrt, die
// Uebernahme scheitert an der fehlenden PIN. Waehrend einer laufenden
// TSE-Einrichtung darf deshalb kein Deploy und kein Neustart erfolgen.
//
// Die beiden lesenden Endpunkte (TestTSEVerbindung, CheckTSESetup) behalten
// r.Context(): Sie sind idempotent und jederzeit wiederholbar, ein Abbruch
// hinterlaesst dort nichts.
func lebenszyklusKontext(r *http.Request) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.WithoutCancel(r.Context()), tseSetupLebenszyklusTimeout)
}

type settingsCommand interface {
	UpdateTSEKonfiguration(ctx context.Context, b tse.Konfiguration) error
	RichteTSEEin(ctx context.Context, credentials tse.SetupCredentials, bestaetigteUmgebung tse.Umgebung, neuAnlegenTrotzVorhandener bool) (application.TSESetupErgebnis, error)
	UebernimmTSE(ctx context.Context, credentials tse.SetupCredentials, bestaetigteUmgebung tse.Umgebung, tssID, pin, puk string) (application.TSESetupErgebnis, error)
}

type CommandHandler struct {
	Command settingsCommand
}

type updateTSEKonfigurationRequest struct {
	ApiKey    string `json:"apiKey"`
	ApiSecret string `json:"apiSecret"`
	TssID     string `json:"tssId"`
	ClientID  string `json:"clientId"`
}

var updateTSEKonfigurationSchema = z.Struct(z.Shape{
	"ApiKey":    z.String().Max(500, z.Message("API-Key darf höchstens 500 Zeichen lang sein")).Optional(),
	"ApiSecret": z.String().Max(500, z.Message("API-Secret darf höchstens 500 Zeichen lang sein")).Optional(),
	"TssID":     z.String().Max(255, z.Message("TSS-ID darf höchstens 255 Zeichen lang sein")).Optional(),
	"ClientID":  z.String().Max(255, z.Message("Client-ID darf höchstens 255 Zeichen lang sein")).Optional(),
})

type tseEinrichtenRequest struct {
	ApiKey                     string `json:"apiKey"`
	ApiSecret                  string `json:"apiSecret"`
	Umgebung                   string `json:"umgebung"`
	NeuAnlegenTrotzVorhandener bool   `json:"neuAnlegenTrotzVorhandener"`
}

// NeuAnlegenTrotzVorhandener ist optional (Default false). Nur in TEST und nur
// als bewusste Sekundaeraktion uebergeht das Backend damit die Sperre gegen eine
// zweite TSS (F2); LIVE bleibt hart gesperrt.
var tseEinrichtenSchema = z.Struct(z.Shape{
	"ApiKey":                     z.String().Min(1, z.Message("API-Key ist erforderlich")).Max(500, z.Message("API-Key darf höchstens 500 Zeichen lang sein")).Required(),
	"ApiSecret":                  z.String().Min(1, z.Message("API-Secret ist erforderlich")).Max(500, z.Message("API-Secret darf höchstens 500 Zeichen lang sein")).Required(),
	"Umgebung":                   z.String().OneOf([]string{string(tse.UmgebungTest), string(tse.UmgebungLive)}, z.Message("Ungültige Umgebung")).Required(),
	"NeuAnlegenTrotzVorhandener": z.Bool().Optional(),
})

type tseUebernehmenRequest struct {
	ApiKey    string `json:"apiKey"`
	ApiSecret string `json:"apiSecret"`
	Umgebung  string `json:"umgebung"`
	TssID     string `json:"tssId"`
	Pin       string `json:"pin"`
	Puk       string `json:"puk"`
}

// Pin ist optional: bei der Uebernahme einer TSS im Zustand CREATED nicht noetig,
// ab UNINITIALIZED traegt es die vom Admin verwahrte Admin-PIN. Puk ist ebenfalls
// optional und nur fuer den PIN-Reset gesetzt: ist die PIN verloren oder gesperrt,
// setzt jotti mit dem PUK eine frische PIN und uebernimmt damit weiter.
var tseUebernehmenSchema = z.Struct(z.Shape{
	"ApiKey":    z.String().Min(1, z.Message("API-Key ist erforderlich")).Max(500, z.Message("API-Key darf höchstens 500 Zeichen lang sein")).Required(),
	"ApiSecret": z.String().Min(1, z.Message("API-Secret ist erforderlich")).Max(500, z.Message("API-Secret darf höchstens 500 Zeichen lang sein")).Required(),
	"Umgebung":  z.String().OneOf([]string{string(tse.UmgebungTest), string(tse.UmgebungLive)}, z.Message("Ungültige Umgebung")).Required(),
	"TssID":     z.String().Min(1, z.Message("TSS-ID ist erforderlich")).Max(255, z.Message("TSS-ID darf höchstens 255 Zeichen lang sein")).Required(),
	"Pin":       z.String().Max(50, z.Message("Admin-PIN darf höchstens 50 Zeichen lang sein")).Optional(),
	"Puk":       z.String().Max(100, z.Message("Admin-PUK darf höchstens 100 Zeichen lang sein")).Optional(),
})

// tseEinrichtenResponse uebergibt PUK und Admin-PIN genau einmal an die UI. Sie
// werden nie persistiert und nie geloggt; der Admin verwahrt sie extern.
type tseEinrichtenResponse struct {
	TssID    string `json:"tssId"`
	ClientID string `json:"clientId"`
	Puk      string `json:"puk"`
	AdminPin string `json:"adminPin"`
	Umgebung string `json:"umgebung"`
}

func (h *CommandHandler) UpdateTSEKonfigurationHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body updateTSEKonfigurationRequest
		if !helper.ReadAndValidateBody(w, r, &body, updateTSEKonfigurationSchema) {
			return
		}

		conf, err := tse.NewKonfiguration(body.ApiKey, body.ApiSecret, body.TssID, body.ClientID)
		if err != nil {
			helper.SendClientError(w, "validation_error", nil)
			return
		}

		if err := h.Command.UpdateTSEKonfiguration(r.Context(), conf); err != nil {
			switch {
			// 409 wie bei Neuanlage und Uebernahme: Der Pfad teilt sich mit
			// ihnen das Schloss auf der TSE-Konfiguration.
			case errors.Is(err, application.ErrTSESetupLaeuftBereits):
				helper.SendConflict(w, "tse_setup_laeuft_bereits")
			case errors.Is(err, application.ErrTSEKonfigurationKassensitzungOffen):
				helper.SendClientError(w, "tse_konfiguration_kassensitzung_offen", nil)
			default:
				helper.SendServerError(w)
			}
			return
		}

		helper.SendEmptyResponse(w)
	}
}

// RichteTSEEinHandler legt eine neue TSS an und fuehrt sie bis zum
// registrierten Client. Der Lebenszyklus laeuft unter lebenszyklusKontext und
// damit unabhaengig davon, ob der Client noch zuhoert: Ein Abbruch mittendrin
// hinterliesse eine bezahlte, halbfertige TSS, deren PUK und Admin-PIN es nur
// in dieser einen Antwort gibt.
func (h *CommandHandler) RichteTSEEinHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		helper.ExtendWriteDeadline(w, r, tseSetupWriteTimeout)

		var body tseEinrichtenRequest
		if !helper.ReadAndValidateBody(w, r, &body, tseEinrichtenSchema) {
			return
		}

		ctx, cancel := lebenszyklusKontext(r)
		defer cancel()

		ergebnis, err := h.Command.RichteTSEEin(
			ctx,
			tse.SetupCredentials{ApiKey: body.ApiKey, ApiSecret: body.ApiSecret},
			tse.Umgebung(body.Umgebung),
			body.NeuAnlegenTrotzVorhandener,
		)

		// Zweites Setzen der Schreibfrist, jetzt fuer den Schreibvorgang selbst:
		// Die Frist vom Handler-Eingang ist eine absolute Zeit ab Request-Start
		// und nach einem langen Lebenszyklus abgelaufen. Diese eine Stelle deckt
		// den Fehler- wie den Erfolgszweig ab.
		helper.ExtendWriteDeadline(w, r, tseSetupWriteTimeout)

		if err != nil {
			switch {
			// 409 statt 400: Der Zustand ist voruebergehend, die Anfrage selbst
			// war in Ordnung — es schreibt nur gerade ein anderer Pfad auf der
			// TSE-Konfiguration.
			case errors.Is(err, application.ErrTSESetupLaeuftBereits):
				helper.SendConflict(w, "tse_setup_laeuft_bereits")
			case errors.Is(err, application.ErrTSESetupZugangsdaten):
				helper.SendClientError(w, "tse_setup_zugangsdaten_ungueltig", nil)
			case errors.Is(err, application.ErrTSESetupUmgebungAbweichung):
				helper.SendClientError(w, "tse_setup_umgebung_abweichung", nil)
			case errors.Is(err, application.ErrTSEKonfigurationKassensitzungOffen):
				helper.SendClientError(w, "tse_konfiguration_kassensitzung_offen", nil)
			case errors.Is(err, application.ErrTSEBereitsEingerichtet):
				helper.SendClientError(w, "tse_bereits_eingerichtet", nil)
			case errors.Is(err, application.ErrTSESetupTSSLimitErreicht):
				helper.SendClientError(w, "tse_setup_tss_limit_erreicht", nil)
			case errors.Is(err, application.ErrTSEEinrichtung),
				errors.Is(err, application.ErrTSEVerbindungFehlgeschlagen):
				helper.SendClientError(w, "tse_einrichtung_fehlgeschlagen", nil)
			default:
				helper.SendServerError(w)
			}
			return
		}

		helper.SendResponse(w, tseEinrichtenResponse{
			TssID:    ergebnis.TssID,
			ClientID: ergebnis.ClientID,
			Puk:      ergebnis.PUK,
			AdminPin: ergebnis.AdminPIN,
			Umgebung: ergebnis.Umgebung,
		})
	}
}

// UebernimmTSEHandler setzt eine vorhandene TSS aus ihrem aktuellen Zustand bis
// zum registrierten Client fort. Wie die Neuanlage laeuft der Lebenszyklus unter
// lebenszyklusKontext: Auch hier entstehen unterwegs PUK bzw. Admin-PIN, die es
// nur in dieser einen Antwort gibt, und ein Abbruch mittendrin liesse die TSS in
// einem Zustand zurueck, aus dem kein zweiter Versuch mehr herausfuehrt.
func (h *CommandHandler) UebernimmTSEHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		helper.ExtendWriteDeadline(w, r, tseSetupWriteTimeout)

		var body tseUebernehmenRequest
		if !helper.ReadAndValidateBody(w, r, &body, tseUebernehmenSchema) {
			return
		}

		ctx, cancel := lebenszyklusKontext(r)
		defer cancel()

		ergebnis, err := h.Command.UebernimmTSE(
			ctx,
			tse.SetupCredentials{ApiKey: body.ApiKey, ApiSecret: body.ApiSecret},
			tse.Umgebung(body.Umgebung),
			body.TssID,
			body.Pin,
			body.Puk,
		)

		// Zweites Setzen der Schreibfrist — siehe RichteTSEEinHandler.
		helper.ExtendWriteDeadline(w, r, tseSetupWriteTimeout)

		if err != nil {
			switch {
			// 409 wie bei der Neuanlage — beide teilen sich dasselbe Schloss.
			case errors.Is(err, application.ErrTSESetupLaeuftBereits):
				helper.SendConflict(w, "tse_setup_laeuft_bereits")
			case errors.Is(err, application.ErrTSESetupZugangsdaten):
				helper.SendClientError(w, "tse_setup_zugangsdaten_ungueltig", nil)
			case errors.Is(err, application.ErrTSESetupUmgebungAbweichung):
				helper.SendClientError(w, "tse_setup_umgebung_abweichung", nil)
			case errors.Is(err, application.ErrTSEKonfigurationKassensitzungOffen):
				helper.SendClientError(w, "tse_konfiguration_kassensitzung_offen", nil)
			case errors.Is(err, application.ErrTSESetupTSSNichtGefunden):
				helper.SendClientError(w, "tse_setup_tss_nicht_gefunden", nil)
			case errors.Is(err, application.ErrTSESetupPINErforderlich):
				helper.SendClientError(w, "tse_setup_pin_erforderlich", nil)
			case errors.Is(err, application.ErrTSESetupPINUnbekannt):
				helper.SendClientError(w, "tse_setup_pin_unbekannt", nil)
			case errors.Is(err, application.ErrTSESetupPUKUnbekannt):
				helper.SendClientError(w, "tse_setup_puk_unbekannt", nil)
			case errors.Is(err, application.ErrTSESetupUebernahmeNichtMoeglich):
				helper.SendClientError(w, "tse_setup_uebernahme_nicht_moeglich", nil)
			case errors.Is(err, application.ErrTSEEinrichtung),
				errors.Is(err, application.ErrTSEVerbindungFehlgeschlagen):
				helper.SendClientError(w, "tse_einrichtung_fehlgeschlagen", nil)
			default:
				helper.SendServerError(w)
			}
			return
		}

		helper.SendResponse(w, tseEinrichtenResponse{
			TssID:    ergebnis.TssID,
			ClientID: ergebnis.ClientID,
			Puk:      ergebnis.PUK,
			AdminPin: ergebnis.AdminPIN,
			Umgebung: ergebnis.Umgebung,
		})
	}
}
