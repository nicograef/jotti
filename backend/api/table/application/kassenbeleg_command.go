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
	"github.com/nicograef/jotti/backend/repository/druckauftrag_repo"
	"github.com/rs/zerolog"
)

func toKassePositionen(positionen []kasse.PositionEventData) []kasse.Position {
	out := make([]kasse.Position, 0, len(positionen))
	for _, pos := range positionen {
		out = append(out, kasse.Position(pos))
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

func toTSEAbschnitt(data *kasse.TSEData) (*escpos.TSEAbschnitt, error) {
	if data == nil {
		return nil, nil
	}

	start, err := time.Parse(time.RFC3339, data.LogTimeStart)
	if err != nil {
		return nil, err
	}
	end, err := time.Parse(time.RFC3339, data.LogTimeEnd)
	if err != nil {
		return nil, err
	}

	return &escpos.TSEAbschnitt{
		TransaktionNr:   data.TransactionNumber,
		Signaturzaehler: data.SignatureCounter,
		TSESeriennummer: data.SerialNumberTSE,
		ZeitpunktBeginn: start,
		ZeitpunktEnde:   end,
		Signatur:        data.Signature,
		QRCodeData:      data.QRCodeData,
	}, nil
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

// tseAbschnittFuerBeleg resolves the TSE section of a Kassenbeleg: signature data from the
// event itself, fallback to the nachsignierte Signatur aus der Seitentabelle (ueber die im
// Event persistierte tseTxId), otherwise the event's TSE-Ausfall flag as Ausfallvermerk
// (AEAO 1.14.2). Eine leere txID bedeutet: Es gab nie einen Signierversuch (TSE war nicht
// konfiguriert) — dann existiert auch keine nachsignierte Signatur.
func (c Command) tseAbschnittFuerBeleg(ctx context.Context, eventTSEData *kasse.TSEData, tseAusfall bool, txID string) (*escpos.TSEAbschnitt, bool, error) {
	abschnitt, err := toTSEAbschnitt(eventTSEData)
	if err != nil {
		return nil, false, err
	}
	if abschnitt != nil {
		return abschnitt, false, nil
	}
	if txID == "" {
		return nil, tseAusfall, nil
	}

	signaturData, err := c.EventRepo.GetTSESignaturByTxID(ctx, txID)
	switch {
	case err == nil:
		abschnitt, err = toTSEAbschnitt(&signaturData)
		if err != nil {
			return nil, false, err
		}
		return abschnitt, false, nil
	case errors.Is(err, db.ErrNotFound):
		return nil, tseAusfall, nil
	default:
		return nil, false, err
	}
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

func (c Command) KassenbelegDrucken(ctx context.Context, tischID int, zahlungID string, verkaufID string, stornierungID string) error {
	log := zerolog.Ctx(ctx)

	if c.DruckstationRepo == nil || c.SettingsRepo == nil || c.DruckauftragRepo == nil {
		log.Error().Msg("KassenbelegDrucken called without required dependencies")
		return ErrDatabase
	}

	ks, err := c.getOffeneKassensitzungOderFehler(ctx)
	if err != nil {
		return err
	}

	var quelleEvent event.Event
	var positionen []kasse.Position
	var gesamtbetragCents int
	var referenz string
	var tseAbschnitt *escpos.TSEAbschnitt
	var tseAusfallvermerk bool
	var ersteBestellungZeitpunkt *time.Time
	var stornobeleg bool
	var stornoZuBelegnummer string

	switch {
	case verkaufID != "" && stornierungID != "":
		subject := kasse.DirektverkaufSubject(ks.ZNr, verkaufID)
		events, err := c.EventRepo.ReadEventsBySubject(ctx, subject)
		if err != nil {
			log.Error().Err(err).Str("verkauf_id", verkaufID).Msg("Failed to read direktverkauf events for stornobeleg")
			return ErrDatabase
		}

		verkaufEvent, _, err := findDirektverkaufGetaetigtEvent(events, verkaufID)
		if err != nil {
			if errors.Is(err, ErrVerkaufNichtGefunden) {
				log.Warn().Str("verkauf_id", verkaufID).Msg("Direktverkauf not found for stornobeleg")
				return ErrVerkaufNichtGefunden
			}
			log.Error().Err(err).Str("verkauf_id", verkaufID).Msg("Failed to decode direktverkauf event data")
			return ErrDatabase
		}

		stornoEvent, stornoData, err := findDirektverkaufStorniertEvent(events, stornierungID)
		if err != nil {
			if errors.Is(err, ErrStornierungNichtGefunden) {
				log.Warn().Str("verkauf_id", verkaufID).Str("stornierung_id", stornierungID).Msg("Stornierung not found for stornobeleg")
				return ErrStornierungNichtGefunden
			}
			log.Error().Err(err).Str("verkauf_id", verkaufID).Str("stornierung_id", stornierungID).Msg("Failed to decode direktverkauf storno event data")
			return ErrDatabase
		}

		quelleEvent = stornoEvent
		positionen = toKassePositionen(stornoData.Positionen)
		gesamtbetragCents = -stornoData.GesamtStornierungCents
		referenz = fmt.Sprintf("direktverkauf-storniert:%d", stornoEvent.ID)
		stornobeleg = true
		stornoZuBelegnummer = fmt.Sprintf("%d", verkaufEvent.ID)

		tseAbschnitt, tseAusfallvermerk, err = c.tseAbschnittFuerBeleg(ctx, stornoData.TSEData, false, stornoData.TSETxID)
		if err != nil {
			log.Error().Err(err).Str("stornierung_id", stornierungID).Msg("Failed to resolve TSE section for stornobeleg")
			return ErrDatabase
		}

	case verkaufID != "":
		subject := kasse.DirektverkaufSubject(ks.ZNr, verkaufID)
		events, err := c.EventRepo.ReadEventsBySubject(ctx, subject)
		if err != nil {
			log.Error().Err(err).Str("verkauf_id", verkaufID).Msg("Failed to read direktverkauf events for kassenbeleg")
			return ErrDatabase
		}

		verkaufEvent, verkaufData, err := findDirektverkaufGetaetigtEvent(events, verkaufID)
		if err != nil {
			if errors.Is(err, ErrVerkaufNichtGefunden) {
				log.Warn().Str("verkauf_id", verkaufID).Msg("Direktverkauf not found for kassenbeleg")
				return ErrVerkaufNichtGefunden
			}
			log.Error().Err(err).Str("verkauf_id", verkaufID).Msg("Failed to decode direktverkauf event data")
			return ErrDatabase
		}

		quelleEvent = verkaufEvent
		positionen = toKassePositionen(verkaufData.Positionen)
		gesamtbetragCents = verkaufData.GesamtbetragCents
		referenz = fmt.Sprintf("direktverkauf-getaetigt:%d", verkaufEvent.ID)

		tseAbschnitt, tseAusfallvermerk, err = c.tseAbschnittFuerBeleg(ctx, verkaufData.TSEData, verkaufData.TSEAusfall, verkaufData.TSETxID)
		if err != nil {
			log.Error().Err(err).Str("verkauf_id", verkaufID).Msg("Failed to resolve TSE section for kassenbeleg")
			return ErrDatabase
		}

	default:
		subject := kasse.TischSessionSubject(ks.ZNr, tischID)
		state, err := c.EventRepo.ReadTischSession(ctx, subject)
		if err != nil {
			log.Error().Err(err).Int("tisch_id", tischID).Msg("Failed to read table projection for kassenbeleg")
			return ErrDatabase
		}
		ersteBestellungZeitpunkt = state.ErsteBestellungLogTime

		events, err := c.EventRepo.ReadEventsBySubject(ctx, subject)
		if err != nil {
			log.Error().Err(err).Int("tisch_id", tischID).Msg("Failed to read events for kassenbeleg")
			return ErrDatabase
		}

		zahlungEvent, zahlungData, err := findZahlungEvent(events, zahlungID)
		if err != nil {
			if errors.Is(err, ErrZahlungNichtGefunden) {
				log.Warn().Int("tisch_id", tischID).Str("zahlung_id", zahlungID).Msg("Zahlung not found for kassenbeleg")
				return ErrZahlungNichtGefunden
			}
			log.Error().Err(err).Int("tisch_id", tischID).Str("zahlung_id", zahlungID).Msg("Failed to decode zahlung event data")
			return ErrDatabase
		}

		quelleEvent = zahlungEvent
		positionen = toKassePositionen(zahlungData.Positionen)
		gesamtbetragCents = zahlungData.GesamtZahlungCents
		referenz = fmt.Sprintf("zahlung-kassiert:%d", zahlungEvent.ID)

		tseAbschnitt, tseAusfallvermerk, err = c.tseAbschnittFuerBeleg(ctx, zahlungData.TSEData, zahlungData.TSEAusfall, zahlungData.TSETxID)
		if err != nil {
			log.Error().Err(err).Int("tisch_id", tischID).Str("zahlung_id", zahlungID).Msg("Failed to resolve TSE section for kassenbeleg")
			return ErrDatabase
		}
	}

	stationen, err := c.DruckstationRepo.GetKonfigurierteDruckstationen(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to load druckstationen for kassenbeleg")
		return ErrDatabase
	}
	kassenbelegStation, ok := stationen[string(druckstation.KategorieKassenbeleg)]
	if !ok || kassenbelegStation.IP == "" {
		return ErrKassenbelegDruckerNichtKonfiguriert
	}

	betreiber, err := c.SettingsRepo.GetBetreiber(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to load betreiber for kassenbeleg")
		return ErrDatabase
	}

	kassenidentitaet, err := c.SettingsRepo.GetKassenidentitaet(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to load kassenidentitaet for kassenbeleg")
		return ErrDatabase
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
		TSEAusfallvermerk:        tseAusfallvermerk,
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
		return ErrDatabase
	}

	log.Info().Int("tisch_id", tischID).Str("verkauf_id", verkaufID).Int("event_id", quelleEvent.ID).Msg("Kassenbeleg queued")
	return nil
}
