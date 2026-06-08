package application

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"

	bondruckApp "github.com/nicograef/jotti/backend/api/bondruck/application"
	"github.com/nicograef/jotti/backend/api/bondruck/application/escpos"
	"github.com/nicograef/jotti/backend/domain/event"
	"github.com/nicograef/jotti/backend/domain/kasse"
	"github.com/rs/zerolog"
)

func toKassePositionen(positionen []zahlungPositionData) []kasse.Position {
	out := make([]kasse.Position, 0, len(positionen))
	for _, pos := range positionen {
		out = append(out, kasse.Position{
			PositionID:   pos.PositionID,
			VarianteID:   pos.VarianteID,
			ProduktName:  pos.ProduktName,
			VarianteName: pos.VarianteName,
			Kategorie:    pos.Kategorie,
			Einzelpreis:  pos.Einzelpreis,
			Menge:        pos.Menge,
		})
	}
	return out
}

func findZahlungEvent(events []event.Event, zahlungID string) (event.Event, zahlungKassiertV1Data, error) {
	for _, evt := range events {
		if evt.Type != string(kasse.EventTypeZahlungKassiertV1) {
			continue
		}

		var data zahlungKassiertV1Data
		if err := json.Unmarshal(evt.Data, &data); err != nil {
			return event.Event{}, zahlungKassiertV1Data{}, err
		}
		if data.ZahlungID == zahlungID {
			return evt, data, nil
		}
	}

	return event.Event{}, zahlungKassiertV1Data{}, ErrZahlungNichtGefunden
}

type direktverkaufGetaetigtV1Data struct {
	VerkaufID         string                `json:"verkaufId"`
	Positionen        []zahlungPositionData `json:"positionen"`
	GesamtbetragCents int                   `json:"gesamtbetragCents"`
	Kommentar         string                `json:"kommentar"`
}

func findDirektverkaufGetaetigtEvent(events []event.Event, verkaufID string) (event.Event, direktverkaufGetaetigtV1Data, error) {
	for _, evt := range events {
		if evt.Type != string(kasse.EventTypeDirektverkaufGetaetigtV1) {
			continue
		}

		var data direktverkaufGetaetigtV1Data
		if err := json.Unmarshal(evt.Data, &data); err != nil {
			return event.Event{}, direktverkaufGetaetigtV1Data{}, err
		}
		if data.VerkaufID == verkaufID {
			return evt, data, nil
		}
	}

	return event.Event{}, direktverkaufGetaetigtV1Data{}, ErrVerkaufNichtGefunden
}

func (c Command) KassenbelegDrucken(ctx context.Context, tischID int, zahlungID string, verkaufID string) error {
	log := zerolog.Ctx(ctx)

	if c.SettingsRepo == nil || c.DruckauftragRepo == nil {
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

	if verkaufID != "" {
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
	} else {
		subject := kasse.TischSessionSubject(ks.ZNr, tischID)
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
	}

	bondruckSettings, err := c.SettingsRepo.GetBondruckEinstellungen(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to load bondruck settings")
		return ErrDatabase
	}
	if bondruckSettings.KassenbelegDruckerIP == "" {
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

	payload := escpos.FormatKassenbeleg(escpos.KassenbelegData{
		Vereinsname:        betreiber.Vereinsname,
		Strasse:            betreiber.Strasse,
		Plz:                betreiber.Plz,
		Ort:                betreiber.Ort,
		KassenSeriennummer: kassenidentitaet.Seriennummer.String(),
		Belegnummer:        fmt.Sprintf("%d", quelleEvent.ID),
		Zeitpunkt:          quelleEvent.Time,
		Positionen:         positionen,
		GesamtbetragCents:  gesamtbetragCents,
		Zahlungsart:        "bar",
	})

	auftrag := bondruckApp.Druckauftrag{
		ZielIP:   bondruckSettings.KassenbelegDruckerIP,
		Payload:  base64.StdEncoding.EncodeToString(payload),
		BonArt:   "kassenbeleg",
		Referenz: referenz,
	}

	if err := c.DruckauftragRepo.EnqueueDruckauftraege(ctx, []bondruckApp.Druckauftrag{auftrag}); err != nil {
		log.Error().Err(err).Int("tisch_id", tischID).Str("zahlung_id", zahlungID).Str("verkauf_id", verkaufID).Msg("Failed to enqueue kassenbeleg")
		return ErrDatabase
	}

	log.Info().Int("tisch_id", tischID).Str("verkauf_id", verkaufID).Int("event_id", quelleEvent.ID).Msg("Kassenbeleg queued")
	return nil
}
