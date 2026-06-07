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

func (c Command) KassenbelegDrucken(ctx context.Context, tischID int, zahlungID string) error {
	log := zerolog.Ctx(ctx)

	if c.SettingsRepo == nil || c.DruckauftragRepo == nil {
		log.Error().Msg("KassenbelegDrucken called without required dependencies")
		return ErrDatabase
	}

	ks, err := c.getOffeneKassensitzungOderFehler(ctx)
	if err != nil {
		return err
	}

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
		Belegnummer:        fmt.Sprintf("%d", zahlungEvent.ID),
		Zeitpunkt:          zahlungEvent.Time,
		Positionen:         toKassePositionen(zahlungData.Positionen),
		GesamtbetragCents:  zahlungData.GesamtZahlungCents,
		Zahlungsart:        "bar",
	})

	auftrag := bondruckApp.Druckauftrag{
		ZielIP:   bondruckSettings.KassenbelegDruckerIP,
		Payload:  base64.StdEncoding.EncodeToString(payload),
		BonArt:   "kassenbeleg",
		Referenz: fmt.Sprintf("zahlung-kassiert:%d", zahlungEvent.ID),
	}

	if err := c.DruckauftragRepo.EnqueueDruckauftraege(ctx, []bondruckApp.Druckauftrag{auftrag}); err != nil {
		log.Error().Err(err).Int("tisch_id", tischID).Str("zahlung_id", zahlungID).Msg("Failed to enqueue kassenbeleg")
		return ErrDatabase
	}

	log.Info().Int("tisch_id", tischID).Int("event_id", zahlungEvent.ID).Msg("Kassenbeleg queued")
	return nil
}
