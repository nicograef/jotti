package application

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/event"
	"github.com/nicograef/jotti/backend/domain/kasse"
	"github.com/nicograef/jotti/backend/domain/steuer"
	"github.com/nicograef/jotti/backend/domain/tse"
	"github.com/rs/zerolog"
)

const (
	tseProcessTypeKassenbelegV1 = "Kassenbeleg-V1"
	tseZahlungsartBar           = "Bar"
)

func (c Command) signZahlungKassiertEvent(ctx context.Context, evt event.Event, positionen []kasse.Position, zahlbetragCents int) (event.Event, error) {
	if c.SettingsRepo == nil || c.NewTSEClient == nil {
		return evt, nil
	}

	log := zerolog.Ctx(ctx)

	conf, err := c.SettingsRepo.GetTSEKonfiguration(ctx)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return evt, nil
		}
		log.Error().Err(err).Msg("Failed to load TSE-Konfiguration for zahlung signierung")
		return event.Event{}, ErrDatabase
	}
	if !conf.IstKonfiguriert() {
		return evt, nil
	}

	processData, err := buildKassenbelegProcessData(positionen, zahlbetragCents)
	if err != nil {
		log.Error().Err(err).Msg("Failed to build TSE process_data for zahlung")
		return event.Event{}, ErrDatabase
	}

	txID, err := tseTransactionIDForZahlungEvent(evt)
	if err != nil {
		log.Error().Err(err).Msg("Failed to derive deterministic tx_id for zahlung")
		return event.Event{}, ErrDatabase
	}

	client, err := c.NewTSEClient(tse.Credentials{
		ApiKey:    conf.ApiKey,
		ApiSecret: conf.ApiSecret,
		TssID:     conf.TssID,
		ClientID:  conf.ClientID,
	})
	if err != nil {
		log.Error().Err(err).Msg("Failed to create TSE client for zahlung signierung")
		return event.Event{}, ErrDatabase
	}

	startResult, err := client.StartTransaction(ctx, txID, tseProcessTypeKassenbelegV1, processData)
	if err != nil {
		log.Error().Err(err).Str("tx_id", txID).Msg("Failed to start TSE transaction for zahlung")
		return event.Event{}, ErrDatabase
	}

	finishResult, err := client.FinishTransaction(ctx, txID, startResult.TransactionNumber, tseProcessTypeKassenbelegV1, processData)
	if err != nil {
		log.Error().Err(err).Str("tx_id", txID).Int("tx_number", startResult.TransactionNumber).Msg("Failed to finish TSE transaction for zahlung")
		return event.Event{}, ErrDatabase
	}

	tseData := kasse.TSEData{
		TransactionNumber: finishResult.TransactionNumber,
		SignatureCounter:  finishResult.SignatureCounter,
		SerialNumberTSE:   strings.TrimSpace(finishResult.SerialNumberTSE),
		LogTimeStart:      tseTimeString(nonZeroTime(startResult.LogTime, finishResult.LogTimeStart)),
		LogTimeEnd:        tseTimeString(nonZeroTime(finishResult.LogTime, finishResult.LogTimeEnd)),
		Signature:         strings.TrimSpace(finishResult.Signature),
		ProcessType:       tseProcessTypeKassenbelegV1,
		QRCodeData:        strings.TrimSpace(finishResult.QRCodeData),
	}

	signedEvent, err := withZahlungEventTSEData(evt, tseData)
	if err != nil {
		log.Error().Err(err).Msg("Failed to embed TSE data into zahlung event")
		return event.Event{}, ErrDatabase
	}

	return signedEvent, nil
}

func withZahlungEventTSEData(evt event.Event, tseData kasse.TSEData) (event.Event, error) {
	if evt.Type != string(kasse.EventTypeZahlungKassiertV1) {
		return event.Event{}, fmt.Errorf("unsupported event type for TSE data: %s", evt.Type)
	}
	if err := tseData.Validate(); err != nil {
		return event.Event{}, err
	}

	data := zahlungKassiertV1Data{}
	if err := json.Unmarshal(evt.Data, &data); err != nil {
		return event.Event{}, err
	}

	data.TSEData = &tseData

	encoded, err := json.Marshal(data)
	if err != nil {
		return event.Event{}, err
	}

	evt.Data = encoded
	return evt, nil
}

func buildKassenbelegProcessData(positionen []kasse.Position, zahlbetragCents int) (string, error) {
	var betragNormalCents int
	var betragErmaessigtCents int
	var betragBefreitCents int

	for _, pos := range positionen {
		brutto := pos.Einzelpreis * pos.Menge
		aufteilungen := steuer.Aufteilen(brutto, steuer.Steuersatz(pos.Steuersatz))
		if len(aufteilungen) == 0 {
			return "", fmt.Errorf("unsupported steuersatz %q", pos.Steuersatz)
		}
		for _, aufteilung := range aufteilungen {
			switch aufteilung.Satz {
			case steuer.RegelSteuersatz:
				betragNormalCents += aufteilung.Brutto
			case steuer.ErmaessigtSteuersatz:
				betragErmaessigtCents += aufteilung.Brutto
			case steuer.BefreitSteuersatz:
				betragBefreitCents += aufteilung.Brutto
			default:
				return "", fmt.Errorf("unsupported steuersatz in aufteilung %q", aufteilung.Satz)
			}
		}
	}

	return fmt.Sprintf(
		"Beleg^%s_%s_%s_%s_%s^%s:%s",
		tseBetragString(betragNormalCents),
		tseBetragString(betragErmaessigtCents),
		tseBetragString(0),
		tseBetragString(0),
		tseBetragString(betragBefreitCents),
		tseBetragString(zahlbetragCents),
		tseZahlungsartBar,
	), nil
}

func tseBetragString(cents int) string {
	sign := ""
	if cents < 0 {
		sign = "-"
		cents = -cents
	}
	return fmt.Sprintf("%s%d.%02d", sign, cents/100, cents%100)
}

func tseTransactionIDForZahlungEvent(evt event.Event) (string, error) {
	if evt.Type != string(kasse.EventTypeZahlungKassiertV1) {
		return "", fmt.Errorf("unsupported event type for tx_id: %s", evt.Type)
	}

	data := zahlungKassiertV1Data{}
	if err := json.Unmarshal(evt.Data, &data); err != nil {
		return "", err
	}
	if strings.TrimSpace(data.ZahlungID) == "" {
		return "", fmt.Errorf("zahlungId missing in event data")
	}

	seed := fmt.Sprintf("%s|%s|%s", evt.Type, evt.Subject, data.ZahlungID)
	return uuid.NewSHA1(uuid.NameSpaceOID, []byte(seed)).String(), nil
}

func tseTimeString(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.UTC().Format(time.RFC3339)
}

func nonZeroTime(primary time.Time, fallback time.Time) time.Time {
	if !primary.IsZero() {
		return primary
	}
	return fallback
}
