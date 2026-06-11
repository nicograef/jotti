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

type geldtransitGebuchtV1Data struct {
	BewegungID  string         `json:"bewegungId"`
	Richtung    string         `json:"richtung"`
	BetragCents int            `json:"betragCents"`
	Kommentar   string         `json:"kommentar"`
	GebuchtVon  int            `json:"gebuchtVon"`
	TSEData     *kasse.TSEData `json:"tseData,omitempty"`
}

type differenzSollIstGebuchtV1Data struct {
	BetragCents int            `json:"betragCents"`
	GebuchtVon  int            `json:"gebuchtVon"`
	TSEData     *kasse.TSEData `json:"tseData,omitempty"`
}

type tagesabschlussErstelltV1Data struct {
	ZNr               int            `json:"zNr"`
	ZeitraumVon       time.Time      `json:"zeitraumVon"`
	ZeitraumBis       time.Time      `json:"zeitraumBis"`
	UmsatzGesamtCents int            `json:"umsatzGesamtCents"`
	StornierungCents  int            `json:"stornierungCents"`
	AuszahlungenCents int            `json:"auszahlungenCents"`
	GeldtransitCents  int            `json:"geldtransitCents"`
	ErstelltVon       int            `json:"erstelltVon"`
	TSEData           *kasse.TSEData `json:"tseData,omitempty"`
}

func (c Command) signGeldtransitGebuchtEvent(ctx context.Context, evt event.Event, richtung string, betragCents int) (eventSignierungErgebnis, error) {
	processData, err := buildGeldtransitProcessData(richtung, betragCents)
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("Failed to build TSE process_data for geldtransit")
		return eventSignierungErgebnis{}, ErrDatabase
	}

	return c.signEventWithTSE(
		ctx,
		evt,
		tse.ProcessTypeKassenbelegV1,
		processData,
		tseTransactionIDForGeldtransitEvent,
		withGeldtransitEventTSEData,
	)
}

func (c Command) signDifferenzSollIstGebuchtEvent(ctx context.Context, evt event.Event, differenzCents int) (eventSignierungErgebnis, error) {
	processData := buildEigenbelegProcessData(differenzCents)
	return c.signEventWithTSE(
		ctx,
		evt,
		tse.ProcessTypeKassenbelegV1,
		processData,
		tseTransactionIDForDifferenzEvent,
		withDifferenzEventTSEData,
	)
}

func (c Command) signTagesabschlussErstelltEvent(ctx context.Context, evt event.Event, zNr int, zeitraumVon time.Time, zeitraumBis time.Time) (eventSignierungErgebnis, error) {
	processData := buildTagesabschlussProcessData(zNr, zeitraumVon, zeitraumBis)
	return c.signEventWithTSE(
		ctx,
		evt,
		tse.ProcessTypeSonstigerVorgang,
		processData,
		tseTransactionIDForTagesabschlussEvent,
		withTagesabschlussEventTSEData,
	)
}

// tseSignierDeadline liefert die Gesamt-Deadline fuer den synchronen
// Signierversuch; 0 bedeutet tse.SignierDeadline (Override nur fuer Tests).
func (c Command) tseSignierDeadline() time.Duration {
	if c.TSESignierDeadline > 0 {
		return c.TSESignierDeadline
	}
	return tse.SignierDeadline
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

	// Gesamt-Deadline fuer den synchronen Signierversuch: Bei fiskaly-Stoerung
	// wartet der Kassier-Request hoechstens diese Zeitspanne, danach greift der
	// Ausfallpfad (Nachsignier-Auftrag fuer den Worker).
	ctx, cancel := context.WithTimeout(ctx, c.tseSignierDeadline())
	defer cancel()

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

func withGeldtransitEventTSEData(evt event.Event, tseData kasse.TSEData) (event.Event, error) {
	if evt.Type != string(kasse.EventTypeGeldtransitGebuchtV1) {
		return event.Event{}, fmt.Errorf("unsupported event type for TSE data: %s", evt.Type)
	}
	if err := tseData.Validate(); err != nil {
		return event.Event{}, err
	}

	data := geldtransitGebuchtV1Data{}
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

func withDifferenzEventTSEData(evt event.Event, tseData kasse.TSEData) (event.Event, error) {
	if evt.Type != string(kasse.EventTypeDifferenzSollIstGebuchtV1) {
		return event.Event{}, fmt.Errorf("unsupported event type for TSE data: %s", evt.Type)
	}
	if err := tseData.Validate(); err != nil {
		return event.Event{}, err
	}

	data := differenzSollIstGebuchtV1Data{}
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

func withTagesabschlussEventTSEData(evt event.Event, tseData kasse.TSEData) (event.Event, error) {
	if evt.Type != string(kasse.EventTypeTagesabschlussErstelltV1) {
		return event.Event{}, fmt.Errorf("unsupported event type for TSE data: %s", evt.Type)
	}
	if err := tseData.Validate(); err != nil {
		return event.Event{}, err
	}

	data := tagesabschlussErstelltV1Data{}
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

func buildGeldtransitProcessData(richtung string, betragCents int) (string, error) {
	switch richtung {
	case "einlage":
		return buildEigenbelegProcessData(betragCents), nil
	case "entnahme":
		return buildEigenbelegProcessData(-betragCents), nil
	default:
		return "", fmt.Errorf("unsupported richtung %q", richtung)
	}
}

// buildEigenbelegProcessData erzeugt Kassenbeleg-V1-processData für Geschäftsvorfälle
// ohne Umsatz (Eigenbelege nach AEAO 2.2.3.6.1): alle Steuerbeträge 0.00, nur der
// Zahlbetrag ist gefüllt. DSFinV-K Anhang I: Zahlungen von 0.00 müssen entfallen.
func buildEigenbelegProcessData(zahlbetragCents int) string {
	zahlungen := ""
	if zahlbetragCents != 0 {
		zahlungen = tseBetragString(zahlbetragCents) + ":" + tseZahlungsartBar
	}
	return fmt.Sprintf("Beleg^0.00_0.00_0.00_0.00_0.00^%s", zahlungen)
}

func buildTagesabschlussProcessData(zNr int, zeitraumVon time.Time, zeitraumBis time.Time) string {
	return fmt.Sprintf(
		"Tagesabschluss^ZNr:%d^Von:%s^Bis:%s",
		zNr,
		zeitraumVon.UTC().Format(time.RFC3339),
		zeitraumBis.UTC().Format(time.RFC3339),
	)
}

func tseBetragString(cents int) string {
	sign := ""
	if cents < 0 {
		sign = "-"
		cents = -cents
	}
	return fmt.Sprintf("%s%d.%02d", sign, cents/100, cents%100)
}

func tseTransactionIDForGeldtransitEvent(evt event.Event) (string, error) {
	if evt.Type != string(kasse.EventTypeGeldtransitGebuchtV1) {
		return "", fmt.Errorf("unsupported event type for tx_id: %s", evt.Type)
	}

	data := geldtransitGebuchtV1Data{}
	if err := json.Unmarshal(evt.Data, &data); err != nil {
		return "", err
	}
	if strings.TrimSpace(data.BewegungID) == "" {
		return "", fmt.Errorf("bewegungId missing in event data")
	}

	seed := fmt.Sprintf("%s|%s|%s", evt.Type, evt.Subject, data.BewegungID)
	return uuid.NewSHA1(uuid.NameSpaceOID, []byte(seed)).String(), nil
}

func tseTransactionIDForDifferenzEvent(evt event.Event) (string, error) {
	if evt.Type != string(kasse.EventTypeDifferenzSollIstGebuchtV1) {
		return "", fmt.Errorf("unsupported event type for tx_id: %s", evt.Type)
	}

	data := differenzSollIstGebuchtV1Data{}
	if err := json.Unmarshal(evt.Data, &data); err != nil {
		return "", err
	}

	seed := fmt.Sprintf("%s|%s|%d|%d|%s", evt.Type, evt.Subject, data.BetragCents, data.GebuchtVon, evt.Time.UTC().Format(time.RFC3339Nano))
	return uuid.NewSHA1(uuid.NameSpaceOID, []byte(seed)).String(), nil
}

func tseTransactionIDForTagesabschlussEvent(evt event.Event) (string, error) {
	if evt.Type != string(kasse.EventTypeTagesabschlussErstelltV1) {
		return "", fmt.Errorf("unsupported event type for tx_id: %s", evt.Type)
	}

	data := tagesabschlussErstelltV1Data{}
	if err := json.Unmarshal(evt.Data, &data); err != nil {
		return "", err
	}

	seed := fmt.Sprintf("%s|%s|%d|%s|%s", evt.Type, evt.Subject, data.ZNr, data.ZeitraumVon.UTC().Format(time.RFC3339Nano), data.ZeitraumBis.UTC().Format(time.RFC3339Nano))
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
