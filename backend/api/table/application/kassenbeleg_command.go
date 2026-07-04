package application

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/nicograef/jotti/backend/api/bondruck/application/escpos"
	"github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/druckstation"
	"github.com/nicograef/jotti/backend/domain/event"
	"github.com/nicograef/jotti/backend/domain/kasse"
	"github.com/nicograef/jotti/backend/domain/steuer"
	"github.com/nicograef/jotti/backend/domain/tse"
	"github.com/nicograef/jotti/backend/repository/druckauftrag_repo"
	"github.com/rs/zerolog"
)

func toKassePositionen(positionen []kasse.PositionEventData) []kasse.Position {
	out := make([]kasse.Position, 0, len(positionen))
	for _, pos := range positionen {
		out = append(out, kasse.PositionFromEventData(pos))
	}
	return out
}

func toSteuermatrixPositionen(positionen []kasse.Position) []steuer.SteuermatrixPosition {
	matrixPositionen := make([]steuer.SteuermatrixPosition, 0, len(positionen))
	for _, position := range positionen {
		matrixPositionen = append(matrixPositionen, steuer.SteuermatrixPosition{
			Brutto:     position.Einzelpreis * position.Menge,
			Steuersatz: steuer.Steuersatz(position.Steuersatz),
		})
	}

	return matrixPositionen
}

func findZahlungEvent(events []event.Event, zahlungID string) (event.Event, kasse.ZahlungKassiertV1Data, error) {
	for _, evt := range events {
		if evt.Type != string(kasse.EventTypeZahlungKassiertV1) {
			continue
		}

		var data kasse.ZahlungKassiertV1Data
		if err := json.Unmarshal(evt.Data, &data); err != nil {
			return event.Event{}, kasse.ZahlungKassiertV1Data{}, err
		}
		if data.ZahlungID == zahlungID {
			return evt, data, nil
		}
	}

	return event.Event{}, kasse.ZahlungKassiertV1Data{}, ErrZahlungNichtGefunden
}

func tseAbschnittAusSignatur(signatur *tse.Signatur) *escpos.TSEAbschnitt {
	return &escpos.TSEAbschnitt{
		TransaktionNr:   signatur.TransaktionNummer,
		Signaturzaehler: signatur.SignaturZaehler,
		TSESeriennummer: signatur.TSESeriennummer,
		ZeitpunktBeginn: signatur.LogTimeStart,
		ZeitpunktEnde:   signatur.LogTimeEnd,
		Signatur:        signatur.Signatur,
		QRCodeData:      signatur.QRCodeData,
	}
}

func findStornierungEvent(events []event.Event, stornierungID string) (event.Event, kasse.StornierungErteiltV1Data, error) {
	for _, evt := range events {
		if evt.Type != string(kasse.EventTypeStornierungErteiltV1) {
			continue
		}

		var data kasse.StornierungErteiltV1Data
		if err := json.Unmarshal(evt.Data, &data); err != nil {
			return event.Event{}, kasse.StornierungErteiltV1Data{}, err
		}
		if data.StornierungID == stornierungID {
			return evt, data, nil
		}
	}

	return event.Event{}, kasse.StornierungErteiltV1Data{}, ErrStornierungNichtGefunden
}

func findDirektverkaufGetaetigtEvent(events []event.Event, verkaufID string) (event.Event, kasse.DirektverkaufGetaetigtV1Data, error) {
	for _, evt := range events {
		if evt.Type != string(kasse.EventTypeDirektverkaufGetaetigtV1) {
			continue
		}

		var data kasse.DirektverkaufGetaetigtV1Data
		if err := json.Unmarshal(evt.Data, &data); err != nil {
			return event.Event{}, kasse.DirektverkaufGetaetigtV1Data{}, err
		}
		if data.VerkaufID == verkaufID {
			return evt, data, nil
		}
	}

	return event.Event{}, kasse.DirektverkaufGetaetigtV1Data{}, ErrVerkaufNichtGefunden
}

func findDirektverkaufStorniertEvent(events []event.Event, stornierungID string) (event.Event, kasse.DirektverkaufStorniertV1Data, error) {
	for _, evt := range events {
		if evt.Type != string(kasse.EventTypeDirektverkaufStorniertV1) {
			continue
		}

		var data kasse.DirektverkaufStorniertV1Data
		if err := json.Unmarshal(evt.Data, &data); err != nil {
			return event.Event{}, kasse.DirektverkaufStorniertV1Data{}, err
		}
		if data.StornierungID == stornierungID {
			return evt, data, nil
		}
	}

	return event.Event{}, kasse.DirektverkaufStorniertV1Data{}, ErrStornierungNichtGefunden
}

// BelegStatus ist die Beleg-Abruf-Antwort: eingereiht (Druckauftrag angelegt)
// oder ausstehend (TSE-Signatur liegt noch nicht vor, kein Druckauftrag — die
// UI ruft denselben Endpunkt erneut auf).
type BelegStatus string

const (
	BelegStatusEingereiht BelegStatus = "eingereiht"
	BelegStatusAusstehend BelegStatus = "ausstehend"
)

// tseAbschnittFuerBeleg loest den TSE-Abschnitt eines Belegs ueber die
// Signaturstatus-Funktion auf — die einzige Implementierung des
// Ausfallbegriffs. Vier Ergebnisarten: Signatur vorhanden (Abschnitt aus den
// Signaturspalten des Auftrags), vorhanden mit Nachsigniert-Kennzeichen
// (verspaetete Signatur), Ausfall mit belegbarem Grund (Beleg ohne TSE-Daten,
// aber mit Ausfallvermerk) oder ausstehend (kein Druckauftrag, die UI fasst
// nach). Kein Auftrag heisst: nicht signaturpflichtig, Beleg ohne
// TSE-Abschnitt und ohne Vermerk.
func (c Command) tseAbschnittFuerBeleg(ctx context.Context, eventID int) (abschnitt *escpos.TSEAbschnitt, vermerk escpos.TSEBelegvermerk, ausstehend bool, err error) {
	stand, err := c.TSERepo.GetSignaturauftragZuEvent(ctx, eventID)
	if errors.Is(err, db.ErrNotFound) {
		return nil, escpos.KeinTSEVermerk, false, nil
	}
	if err != nil {
		return nil, escpos.KeinTSEVermerk, false, err
	}

	aktiveStoerung, err := c.TSERepo.GetAktiveTSEStoerung(ctx)
	if err != nil {
		return nil, escpos.KeinTSEVermerk, false, err
	}

	ergebnis := tse.BestimmeSignaturstatus(stand, aktiveStoerung)
	switch ergebnis.Status {
	case tse.SignaturstatusVorhanden:
		return tseAbschnittAusSignatur(ergebnis.Signatur), escpos.KeinTSEVermerk, false, nil
	case tse.SignaturstatusNachsigniert:
		abschnitt := tseAbschnittAusSignatur(ergebnis.Signatur)
		abschnitt.Nachsigniert = true
		return abschnitt, escpos.KeinTSEVermerk, false, nil
	case tse.SignaturstatusAusfall:
		zerolog.Ctx(ctx).Warn().Int("event_id", eventID).Str("ausfall_grund", ergebnis.AusfallGrund).Msg("Kassenbeleg ohne TSE-Signatur: dokumentierter Ausfall")
		return nil, vermerkFuerAusfall(ergebnis.AusfallGrund), false, nil
	default: // ausstehend
		return nil, escpos.KeinTSEVermerk, true, nil
	}
}

// vermerkFuerAusfall waehlt den Beleg-Hinweis nach dem Ausfallgrund: fehlende
// TSE-Konfiguration (endgueltiger Auftragsstatus oder keine_konfiguration-
// Stoerung) traegt „keine TSE konfiguriert" und wird nicht nachsigniert; jeder
// andere Ausfall (voruebergehende Nichterreichbarkeit) wird nachsigniert.
func vermerkFuerAusfall(ausfallGrund string) escpos.TSEBelegvermerk {
	if ausfallGrund == tse.StatusTSENichtKonfiguriert || ausfallGrund == tse.StoerungGrundKeineKonfiguration {
		return escpos.TSEVermerkKeineKonfiguration
	}
	return escpos.TSEVermerkVoruebergehend
}

// negierePositionen flips the Einzelpreis sign so a Stornobeleg shows negative amounts.
func negierePositionen(positionen []kasse.Position) []kasse.Position {
	out := make([]kasse.Position, 0, len(positionen))
	for _, pos := range positionen {
		pos.Einzelpreis = -pos.Einzelpreis
		out = append(out, pos)
	}
	return out
}

// negiereAufteilungen flips the sign of all Steuermatrix amounts for a Stornobeleg.
// steuer.Aufteilen ignores negative Brutto, so the matrix is computed from the positive
// amounts and negated afterwards (same approach as the faktor in the TSE-processData).
func negiereAufteilungen(aufteilungen []steuer.Aufteilung) []steuer.Aufteilung {
	out := make([]steuer.Aufteilung, 0, len(aufteilungen))
	for _, aufteilung := range aufteilungen {
		aufteilung.Brutto = -aufteilung.Brutto
		aufteilung.Netto = -aufteilung.Netto
		aufteilung.Steuer = -aufteilung.Steuer
		out = append(out, aufteilung)
	}
	return out
}

func (c Command) KassenbelegDrucken(ctx context.Context, tischID int, zahlungID string, verkaufID string, stornierungID string) (BelegStatus, error) {
	log := zerolog.Ctx(ctx)

	if c.DruckstationRepo == nil || c.SettingsRepo == nil || c.DruckauftragRepo == nil || c.TSERepo == nil {
		log.Error().Msg("KassenbelegDrucken called without required dependencies")
		return "", ErrDatabase
	}

	ks, err := c.getOffeneKassensitzungOderFehler(ctx)
	if err != nil {
		return "", err
	}

	var quelleEvent event.Event
	var positionen []kasse.Position
	var gesamtbetragCents int
	var referenz string
	var ersteBestellungZeitpunkt *time.Time
	var stornobeleg bool
	var stornoZuBelegnummer string

	switch {
	case verkaufID != "" && stornierungID != "":
		subject := kasse.DirektverkaufSubject(ks.ZNr, verkaufID)
		events, err := c.EventRepo.ReadEventsBySubject(ctx, subject)
		if err != nil {
			log.Error().Err(err).Str("verkauf_id", verkaufID).Msg("Failed to read direktverkauf events for stornobeleg")
			return "", ErrDatabase
		}

		verkaufEvent, _, err := findDirektverkaufGetaetigtEvent(events, verkaufID)
		if err != nil {
			if errors.Is(err, ErrVerkaufNichtGefunden) {
				log.Warn().Str("verkauf_id", verkaufID).Msg("Direktverkauf not found for stornobeleg")
				return "", ErrVerkaufNichtGefunden
			}
			log.Error().Err(err).Str("verkauf_id", verkaufID).Msg("Failed to decode direktverkauf event data")
			return "", ErrDatabase
		}

		stornoEvent, stornoData, err := findDirektverkaufStorniertEvent(events, stornierungID)
		if err != nil {
			if errors.Is(err, ErrStornierungNichtGefunden) {
				log.Warn().Str("verkauf_id", verkaufID).Str("stornierung_id", stornierungID).Msg("Stornierung not found for stornobeleg")
				return "", ErrStornierungNichtGefunden
			}
			log.Error().Err(err).Str("verkauf_id", verkaufID).Str("stornierung_id", stornierungID).Msg("Failed to decode direktverkauf storno event data")
			return "", ErrDatabase
		}

		quelleEvent = stornoEvent
		positionen = toKassePositionen(stornoData.Positionen)
		gesamtbetragCents = -stornoData.GesamtStornierungCents
		referenz = fmt.Sprintf("direktverkauf-storniert:%d", stornoEvent.ID)
		stornobeleg = true
		stornoZuBelegnummer = fmt.Sprintf("%d", verkaufEvent.ID)

	case stornierungID != "":
		// Tisch-Storno-Beleg (Warenrücknahme): negativer Betrag, Referenz auf den
		// ursprünglichen Zahlungsbeleg — analog zum Direktverkauf-Storno-Beleg.
		subject := kasse.TischSessionSubject(ks.ZNr, tischID)
		events, err := c.EventRepo.ReadEventsBySubject(ctx, subject)
		if err != nil {
			log.Error().Err(err).Int("tisch_id", tischID).Msg("Failed to read events for stornobeleg")
			return "", ErrDatabase
		}

		stornoEvent, stornoData, err := findStornierungEvent(events, stornierungID)
		if err != nil {
			if errors.Is(err, ErrStornierungNichtGefunden) {
				log.Warn().Int("tisch_id", tischID).Str("stornierung_id", stornierungID).Msg("Stornierung not found for stornobeleg")
				return "", ErrStornierungNichtGefunden
			}
			log.Error().Err(err).Int("tisch_id", tischID).Str("stornierung_id", stornierungID).Msg("Failed to decode stornierung event data")
			return "", ErrDatabase
		}

		zahlungEvent, _, err := findZahlungEvent(events, stornoData.ZahlungID)
		if err != nil {
			if errors.Is(err, ErrZahlungNichtGefunden) {
				log.Warn().Int("tisch_id", tischID).Str("zahlung_id", stornoData.ZahlungID).Msg("Referenzierte Zahlung not found for stornobeleg")
				return "", ErrZahlungNichtGefunden
			}
			log.Error().Err(err).Int("tisch_id", tischID).Str("zahlung_id", stornoData.ZahlungID).Msg("Failed to decode referenzierte zahlung event data")
			return "", ErrDatabase
		}

		quelleEvent = stornoEvent
		positionen = toKassePositionen(stornoData.Positionen)
		gesamtbetragCents = -stornoData.GesamtStornierungCents
		referenz = fmt.Sprintf("stornierung-erteilt:%d", stornoEvent.ID)
		stornobeleg = true
		stornoZuBelegnummer = fmt.Sprintf("%d", zahlungEvent.ID)

	case verkaufID != "":
		subject := kasse.DirektverkaufSubject(ks.ZNr, verkaufID)
		events, err := c.EventRepo.ReadEventsBySubject(ctx, subject)
		if err != nil {
			log.Error().Err(err).Str("verkauf_id", verkaufID).Msg("Failed to read direktverkauf events for kassenbeleg")
			return "", ErrDatabase
		}

		verkaufEvent, verkaufData, err := findDirektverkaufGetaetigtEvent(events, verkaufID)
		if err != nil {
			if errors.Is(err, ErrVerkaufNichtGefunden) {
				log.Warn().Str("verkauf_id", verkaufID).Msg("Direktverkauf not found for kassenbeleg")
				return "", ErrVerkaufNichtGefunden
			}
			log.Error().Err(err).Str("verkauf_id", verkaufID).Msg("Failed to decode direktverkauf event data")
			return "", ErrDatabase
		}

		quelleEvent = verkaufEvent
		positionen = toKassePositionen(verkaufData.Positionen)
		gesamtbetragCents = verkaufData.GesamtbetragCents
		referenz = fmt.Sprintf("direktverkauf-getaetigt:%d", verkaufEvent.ID)

	default:
		subject := kasse.TischSessionSubject(ks.ZNr, tischID)
		state, err := c.EventRepo.ReadTischSession(ctx, subject)
		if err != nil {
			log.Error().Err(err).Int("tisch_id", tischID).Msg("Failed to read table projection for kassenbeleg")
			return "", ErrDatabase
		}
		ersteBestellungZeitpunkt = state.ErsteBestellungLogTime

		events, err := c.EventRepo.ReadEventsBySubject(ctx, subject)
		if err != nil {
			log.Error().Err(err).Int("tisch_id", tischID).Msg("Failed to read events for kassenbeleg")
			return "", ErrDatabase
		}

		zahlungEvent, zahlungData, err := findZahlungEvent(events, zahlungID)
		if err != nil {
			if errors.Is(err, ErrZahlungNichtGefunden) {
				log.Warn().Int("tisch_id", tischID).Str("zahlung_id", zahlungID).Msg("Zahlung not found for kassenbeleg")
				return "", ErrZahlungNichtGefunden
			}
			log.Error().Err(err).Int("tisch_id", tischID).Str("zahlung_id", zahlungID).Msg("Failed to decode zahlung event data")
			return "", ErrDatabase
		}

		quelleEvent = zahlungEvent
		positionen = toKassePositionen(zahlungData.Positionen)
		gesamtbetragCents = zahlungData.GesamtZahlungCents
		referenz = fmt.Sprintf("zahlung-kassiert:%d", zahlungEvent.ID)
	}

	// Sofortantwort statt Warten: Liegt die Signatur des Vorgangs noch nicht am
	// Auftrag und ist kein Ausfall dokumentiert, entsteht kein Druckauftrag —
	// die UI fasst über denselben Endpunkt nach, bis der Signatur-Worker
	// quittiert hat. Bei dokumentiertem Ausfall entsteht der Beleg ohne
	// TSE-Daten, weist den Ausfall aber aus.
	tseAbschnitt, tseVermerk, ausstehend, err := c.tseAbschnittFuerBeleg(ctx, quelleEvent.ID)
	if err != nil {
		log.Error().Err(err).Int("event_id", quelleEvent.ID).Msg("Failed to resolve TSE section for kassenbeleg")
		return "", ErrDatabase
	}
	if ausstehend {
		log.Info().Int("event_id", quelleEvent.ID).Msg("Kassenbeleg zurückgestellt: TSE-Signatur ausstehend")
		return BelegStatusAusstehend, nil
	}

	stationen, err := c.DruckstationRepo.GetKonfigurierteDruckstationen(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to load druckstationen for kassenbeleg")
		return "", ErrDatabase
	}
	kassenbelegStation, ok := stationen[string(druckstation.KategorieKassenbeleg)]
	if !ok || kassenbelegStation.IP == "" {
		return "", ErrKassenbelegDruckerNichtKonfiguriert
	}

	betreiber, err := c.SettingsRepo.GetBetreiber(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to load betreiber for kassenbeleg")
		return "", ErrDatabase
	}

	kassenidentitaet, err := c.SettingsRepo.GetKassenidentitaet(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to load kassenidentitaet for kassenbeleg")
		return "", ErrDatabase
	}

	steuermatrix := steuer.Steuermatrix(toSteuermatrixPositionen(positionen))
	if stornobeleg {
		steuermatrix = negiereAufteilungen(steuermatrix)
		positionen = negierePositionen(positionen)
	}

	payload := escpos.FormatKassenbeleg(escpos.KassenbelegData{
		Vereinsname:              betreiber.Vereinsname,
		Strasse:                  betreiber.Strasse,
		Plz:                      betreiber.Plz,
		Ort:                      betreiber.Ort,
		KassenSeriennummer:       kassenidentitaet.Seriennummer.String(),
		Belegnummer:              fmt.Sprintf("%d", quelleEvent.ID),
		Zeitpunkt:                quelleEvent.Time,
		ErsteBestellungZeitpunkt: ersteBestellungZeitpunkt,
		Positionen:               positionen,
		Steuermatrix:             steuermatrix,
		TSE:                      tseAbschnitt,
		TSEVermerk:               tseVermerk,
		GesamtbetragCents:        gesamtbetragCents,
		Zahlungsart:              "bar",
		Stornobeleg:              stornobeleg,
		StornoZuBelegnummer:      stornoZuBelegnummer,
	})

	auftrag := druckauftrag_repo.NeuerDruckauftrag{
		ZielIP:   kassenbelegStation.IP,
		Payload:  base64.StdEncoding.EncodeToString(payload),
		BonArt:   "kassenbeleg",
		Referenz: referenz,
	}

	if err := c.DruckauftragRepo.EnqueueDruckauftraege(ctx, []druckauftrag_repo.NeuerDruckauftrag{auftrag}); err != nil {
		log.Error().Err(err).Int("tisch_id", tischID).Str("zahlung_id", zahlungID).Str("verkauf_id", verkaufID).Msg("Failed to enqueue kassenbeleg")
		return "", ErrDatabase
	}

	log.Info().Int("tisch_id", tischID).Str("verkauf_id", verkaufID).Int("event_id", quelleEvent.ID).Msg("Kassenbeleg queued")
	return BelegStatusEingereiht, nil
}
