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

const tseZahlungsartBar = "Bar"

type tseNachsignierAuftrag struct {
	TxID        string
	ProcessType string
	ProcessData string
}

type eventSignierungErgebnis struct {
	Event              event.Event
	NachsignierAuftrag *tseNachsignierAuftrag
}

type direktverkaufGetaetigtV1Data struct {
	VerkaufID         string           `json:"verkaufId"`
	GesamtbetragCents int              `json:"gesamtbetragCents"`
	Positionen        []kasse.Position `json:"positionen"`
	Kommentar         string           `json:"kommentar"`
	TSEData           *kasse.TSEData   `json:"tseData,omitempty"`
}

type direktverkaufStorniertV1Data struct {
	StornierungID          string           `json:"stornierungId"`
	VerkaufID              string           `json:"verkaufId"`
	Positionen             []kasse.Position `json:"positionen"`
	GesamtStornierungCents int              `json:"gesamtStornierungCents"`
	Kommentar              string           `json:"kommentar"`
	TSEData                *kasse.TSEData   `json:"tseData,omitempty"`
}

func (c Command) signDirektverkaufGetaetigtEvent(ctx context.Context, evt event.Event, positionen []kasse.Position) (eventSignierungErgebnis, error) {
	processData, err := buildKassenbelegProcessData(positionen, sumPositionenCents(positionen))
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("Failed to build TSE process_data for direktverkauf")
		return eventSignierungErgebnis{}, ErrDatabase
	}

	return c.signEventWithTSE(
		ctx,
		evt,
		tse.ProcessTypeKassenbelegV1,
		processData,
		tseTransactionIDForDirektverkaufGetaetigtEvent,
		withDirektverkaufGetaetigtEventTSEData,
	)
}

func (c Command) signDirektverkaufStorniertEvent(ctx context.Context, evt event.Event, positionen []kasse.Position, stornoBetragCents int) (eventSignierungErgebnis, error) {
	processData, err := buildKassenbelegProcessDataWithFaktor(positionen, -stornoBetragCents, -1)
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("Failed to build TSE process_data for direktverkauf storno")
		return eventSignierungErgebnis{}, ErrDatabase
	}

	return c.signEventWithTSE(
		ctx,
		evt,
		tse.ProcessTypeKassenbelegV1,
		processData,
		tseTransactionIDForDirektverkaufStorniertEvent,
		withDirektverkaufStorniertEventTSEData,
	)
}

func (c Command) signEventWithTSE(
	ctx context.Context,
	evt event.Event,
	processType string,
	processData string,
	transactionID func(event.Event) (string, error),
	withTSEData func(event.Event, kasse.TSEData) (event.Event, error),
) (eventSignierungErgebnis, error) {
	if c.SettingsRepo == nil || c.NewTSEClient == nil {
		return eventSignierungErgebnis{Event: evt}, nil
	}

	log := zerolog.Ctx(ctx)

	conf, err := c.SettingsRepo.GetTSEKonfiguration(ctx)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return eventSignierungErgebnis{Event: evt}, nil
		}
		log.Error().Err(err).Msg("Failed to load TSE-Konfiguration for signierung")
		return eventSignierungErgebnis{}, ErrDatabase
	}
	if !conf.IstKonfiguriert() {
		return eventSignierungErgebnis{Event: evt}, nil
	}

	txID, err := transactionID(evt)
	if err != nil {
		log.Error().Err(err).Msg("Failed to derive deterministic tx_id")
		return eventSignierungErgebnis{}, ErrDatabase
	}

	client, err := c.NewTSEClient(tse.Credentials{
		ApiKey:    conf.ApiKey,
		ApiSecret: conf.ApiSecret,
		TssID:     conf.TssID,
		ClientID:  conf.ClientID,
	})
	if err != nil {
		return handleSignierAusfall(log, evt, txID, processType, processData, err)
	}

	startResult, err := client.StartTransaction(ctx, txID, processType, processData)
	if err != nil {
		return handleSignierAusfall(log, evt, txID, processType, processData, err)
	}

	finishResult, err := client.FinishTransaction(ctx, txID, startResult.TransactionNumber, processType, processData)
	if err != nil {
		return handleSignierAusfall(log, evt, txID, processType, processData, err)
	}

	tseData := kasse.TSEData{
		TransactionNumber: finishResult.TransactionNumber,
		SignatureCounter:  finishResult.SignatureCounter,
		SerialNumberTSE:   strings.TrimSpace(finishResult.SerialNumberTSE),
		LogTimeStart:      tseTimeString(nonZeroTime(startResult.LogTime, finishResult.LogTimeStart)),
		LogTimeEnd:        tseTimeString(nonZeroTime(finishResult.LogTime, finishResult.LogTimeEnd)),
		Signature:         strings.TrimSpace(finishResult.Signature),
		ProcessType:       processType,
		QRCodeData:        strings.TrimSpace(finishResult.QRCodeData),
	}

	signedEvent, err := withTSEData(evt, tseData)
	if err != nil {
		log.Error().Err(err).Msg("Failed to embed TSE data into event")
		return eventSignierungErgebnis{}, ErrDatabase
	}

	return eventSignierungErgebnis{Event: signedEvent}, nil
}

func handleSignierAusfall(log *zerolog.Logger, evt event.Event, txID string, processType string, processData string, cause error) (eventSignierungErgebnis, error) {
	log.Warn().Err(cause).Str("tx_id", txID).Msg("TSE-Signierung fehlgeschlagen, Vorgang wird unsigniert persistiert und zur Nachsignierung vorgemerkt")

	return eventSignierungErgebnis{
		Event: evt,
		NachsignierAuftrag: &tseNachsignierAuftrag{
			TxID:        txID,
			ProcessType: processType,
			ProcessData: processData,
		},
	}, nil
}

func withDirektverkaufGetaetigtEventTSEData(evt event.Event, tseData kasse.TSEData) (event.Event, error) {
	if evt.Type != string(kasse.EventTypeDirektverkaufGetaetigtV1) {
		return event.Event{}, fmt.Errorf("unsupported event type for TSE data: %s", evt.Type)
	}
	if err := tseData.Validate(); err != nil {
		return event.Event{}, err
	}

	data := direktverkaufGetaetigtV1Data{}
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

func withDirektverkaufStorniertEventTSEData(evt event.Event, tseData kasse.TSEData) (event.Event, error) {
	if evt.Type != string(kasse.EventTypeDirektverkaufStorniertV1) {
		return event.Event{}, fmt.Errorf("unsupported event type for TSE data: %s", evt.Type)
	}
	if err := tseData.Validate(); err != nil {
		return event.Event{}, err
	}

	data := direktverkaufStorniertV1Data{}
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
	return buildKassenbelegProcessDataWithFaktor(positionen, zahlbetragCents, 1)
}

func buildKassenbelegProcessDataWithFaktor(positionen []kasse.Position, zahlbetragCents int, faktor int) (string, error) {
	if faktor != 1 && faktor != -1 {
		return "", fmt.Errorf("invalid faktor %d", faktor)
	}

	var betragNormalCents int
	var betragErmaessigtCents int
	var betragBefreitCents int

	for _, pos := range positionen {
		basisBrutto := pos.Einzelpreis * pos.Menge
		aufteilungen := steuer.Aufteilen(basisBrutto, steuer.Steuersatz(pos.Steuersatz))
		if len(aufteilungen) == 0 {
			return "", fmt.Errorf("unsupported steuersatz %q", pos.Steuersatz)
		}
		for _, aufteilung := range aufteilungen {
			brutto := aufteilung.Brutto * faktor
			switch aufteilung.Satz {
			case steuer.RegelSteuersatz:
				betragNormalCents += brutto
			case steuer.ErmaessigtSteuersatz:
				betragErmaessigtCents += brutto
			case steuer.BefreitSteuersatz:
				betragBefreitCents += brutto
			default:
				return "", fmt.Errorf("unsupported steuersatz in aufteilung %q", aufteilung.Satz)
			}
		}
	}

	// DSFinV-K Anhang I: Zahlungen von 0.00 müssen entfallen.
	zahlungen := ""
	if zahlbetragCents != 0 {
		zahlungen = tseBetragString(zahlbetragCents) + ":" + tseZahlungsartBar
	}

	return fmt.Sprintf(
		"Beleg^%s_%s_%s_%s_%s^%s",
		tseBetragString(betragNormalCents),
		tseBetragString(betragErmaessigtCents),
		tseBetragString(0),
		tseBetragString(0),
		tseBetragString(betragBefreitCents),
		zahlungen,
	), nil
}

func sumPositionenCents(positionen []kasse.Position) int {
	sum := 0
	for _, pos := range positionen {
		sum += pos.Einzelpreis * pos.Menge
	}
	return sum
}

func tseBetragString(cents int) string {
	sign := ""
	if cents < 0 {
		sign = "-"
		cents = -cents
	}
	return fmt.Sprintf("%s%d.%02d", sign, cents/100, cents%100)
}

func tseTransactionIDForDirektverkaufGetaetigtEvent(evt event.Event) (string, error) {
	if evt.Type != string(kasse.EventTypeDirektverkaufGetaetigtV1) {
		return "", fmt.Errorf("unsupported event type for tx_id: %s", evt.Type)
	}

	data := direktverkaufGetaetigtV1Data{}
	if err := json.Unmarshal(evt.Data, &data); err != nil {
		return "", err
	}
	if strings.TrimSpace(data.VerkaufID) == "" {
		return "", fmt.Errorf("verkaufId missing in event data")
	}

	seed := fmt.Sprintf("%s|%s|%s", evt.Type, evt.Subject, data.VerkaufID)
	return uuid.NewSHA1(uuid.NameSpaceOID, []byte(seed)).String(), nil
}

func tseTransactionIDForDirektverkaufStorniertEvent(evt event.Event) (string, error) {
	if evt.Type != string(kasse.EventTypeDirektverkaufStorniertV1) {
		return "", fmt.Errorf("unsupported event type for tx_id: %s", evt.Type)
	}

	data := direktverkaufStorniertV1Data{}
	if err := json.Unmarshal(evt.Data, &data); err != nil {
		return "", err
	}
	if strings.TrimSpace(data.StornierungID) == "" {
		return "", fmt.Errorf("stornierungId missing in event data")
	}

	seed := fmt.Sprintf("%s|%s|%s", evt.Type, evt.Subject, data.StornierungID)
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
