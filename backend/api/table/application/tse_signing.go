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

type tseNachsignierAuftrag struct {
	TxID        string
	ProcessType string
	ProcessData string
}

type zahlungSignierungErgebnis struct {
	Event              event.Event
	NachsignierAuftrag *tseNachsignierAuftrag
}

func (c Command) signZahlungKassiertEvent(ctx context.Context, evt event.Event, positionen []kasse.Position, zahlbetragCents int) (zahlungSignierungErgebnis, error) {
	if c.SettingsRepo == nil || c.NewTSEClient == nil {
		return zahlungSignierungErgebnis{Event: evt}, nil
	}

	log := zerolog.Ctx(ctx)

	conf, err := c.SettingsRepo.GetTSEKonfiguration(ctx)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return zahlungSignierungErgebnis{Event: evt}, nil
		}
		log.Error().Err(err).Msg("Failed to load TSE-Konfiguration for zahlung signierung")
		return zahlungSignierungErgebnis{}, ErrDatabase
	}
	if !conf.IstKonfiguriert() {
		return zahlungSignierungErgebnis{Event: evt}, nil
	}

	processData, err := buildKassenbelegProcessData(positionen, zahlbetragCents)
	if err != nil {
		log.Error().Err(err).Msg("Failed to build TSE process_data for zahlung")
		return zahlungSignierungErgebnis{}, ErrDatabase
	}

	txID, err := tseTransactionIDForZahlungEvent(evt)
	if err != nil {
		log.Error().Err(err).Msg("Failed to derive deterministic tx_id for zahlung")
		return zahlungSignierungErgebnis{}, ErrDatabase
	}

	client, err := c.NewTSEClient(tse.Credentials{
		ApiKey:    conf.ApiKey,
		ApiSecret: conf.ApiSecret,
		TssID:     conf.TssID,
		ClientID:  conf.ClientID,
	})
	if err != nil {
		return handleZahlungSignierAusfall(log, evt, txID, processData, err)
	}

	startResult, err := client.StartTransaction(ctx, txID, tseProcessTypeKassenbelegV1, processData)
	if err != nil {
		return handleZahlungSignierAusfall(log, evt, txID, processData, err)
	}

	finishResult, err := client.FinishTransaction(ctx, txID, startResult.TransactionNumber, tseProcessTypeKassenbelegV1, processData)
	if err != nil {
		return handleZahlungSignierAusfall(log, evt, txID, processData, err)
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
		return zahlungSignierungErgebnis{}, ErrDatabase
	}

	return zahlungSignierungErgebnis{Event: signedEvent}, nil
}

func handleZahlungSignierAusfall(log *zerolog.Logger, evt event.Event, txID string, processData string, cause error) (zahlungSignierungErgebnis, error) {
	log.Warn().Err(cause).Str("tx_id", txID).Msg("TSE-Signierung fehlgeschlagen, Vorgang wird unsigniert persistiert und zur Nachsignierung vorgemerkt")

	unsignedEvent, err := withZahlungEventTSEAusfall(evt)
	if err != nil {
		log.Error().Err(err).Msg("Failed to mark zahlung event with TSE-Ausfall")
		return zahlungSignierungErgebnis{}, ErrDatabase
	}

	return zahlungSignierungErgebnis{
		Event: unsignedEvent,
		NachsignierAuftrag: &tseNachsignierAuftrag{
			TxID:        txID,
			ProcessType: tseProcessTypeKassenbelegV1,
			ProcessData: processData,
		},
	}, nil
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
	data.TSEAusfall = false

	encoded, err := json.Marshal(data)
	if err != nil {
		return event.Event{}, err
	}

	evt.Data = encoded
	return evt, nil
}

func withZahlungEventTSEAusfall(evt event.Event) (event.Event, error) {
	if evt.Type != string(kasse.EventTypeZahlungKassiertV1) {
		return event.Event{}, fmt.Errorf("unsupported event type for TSE-Ausfall: %s", evt.Type)
	}

	data := zahlungKassiertV1Data{}
	if err := json.Unmarshal(evt.Data, &data); err != nil {
		return event.Event{}, err
	}

	data.TSEData = nil
	data.TSEAusfall = true

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
