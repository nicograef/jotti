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
	"github.com/nicograef/jotti/backend/domain/settings"
	"github.com/nicograef/jotti/backend/domain/tse"
	"github.com/rs/zerolog"
)

type tseKonfigurationRepo interface {
	GetTSEKonfiguration(ctx context.Context) (settings.TSEKonfiguration, error)
}

// NewTSEClient erzeugt einen TSE-Client fuer die uebergebenen Zugangsdaten.
type NewTSEClient func(credentials tse.Credentials) (tse.TSEClient, error)

// NachsignierAuftrag beschreibt eine fehlgeschlagene Signierung, die der
// Nachsignier-Worker nachholt.
type NachsignierAuftrag struct {
	TxID        string
	ProcessType string
	ProcessData string
}

// Signierung ist das Ergebnis eines Signierversuchs: das Event (immer um die
// tx-ID, bei Erfolg zusaetzlich um die TSE-Daten erweitert) plus — bei
// Ausfall — der Nachsignier-Auftrag fuer den Worker.
type Signierung struct {
	Event              event.Event
	NachsignierAuftrag *NachsignierAuftrag
}

// EmbedTSE persistiert das Signier-Ergebnis in den Daten eines konkreten
// Event-Typs: die tx-ID (`tseTxId`) immer, die TSE-Daten nur bei Erfolg.
// data == nil bedeutet Ausfall; Event-Typen mit Ausfallvermerk setzen dann
// ihr TSEAusfall-Flag.
type EmbedTSE func(evt event.Event, txID string, data *kasse.TSEData) (event.Event, error)

// EmbedTSEInData baut eine EmbedTSE-Funktion fuer einen Event-Data-Typ T:
// Typ-Check, JSON-Roundtrip und Zurueckschreiben sind fuer alle Event-Typen
// identisch, nur das Setzen der TSE-Felder (apply) ist typspezifisch.
func EmbedTSEInData[T any](eventType kasse.EventType, apply func(data *T, txID string, tseData *kasse.TSEData)) EmbedTSE {
	return func(evt event.Event, txID string, tseData *kasse.TSEData) (event.Event, error) {
		if evt.Type != string(eventType) {
			return event.Event{}, fmt.Errorf("unsupported event type for TSE data: %s", evt.Type)
		}

		var data T
		if err := json.Unmarshal(evt.Data, &data); err != nil {
			return event.Event{}, err
		}

		apply(&data, txID, tseData)

		encoded, err := json.Marshal(data)
		if err != nil {
			return event.Event{}, err
		}

		evt.Data = encoded
		return evt, nil
	}
}

// Signierer orchestriert die synchrone TSE-Signierung von Events fuer alle
// Kassen-Kontexte (Tisch, Direktverkauf, Kassensitzung). Ohne SettingsRepo/
// NewTSEClient oder ohne TSE-Konfiguration bleiben Events unsigniert.
type Signierer struct {
	SettingsRepo tseKonfigurationRepo
	NewTSEClient NewTSEClient
	// SignierDeadline ueberschreibt die Gesamt-Deadline des synchronen
	// Signierversuchs; 0 bedeutet tse.SignierDeadline. Nur fuer Tests gedacht.
	SignierDeadline time.Duration
}

func (s Signierer) deadline() time.Duration {
	if s.SignierDeadline > 0 {
		return s.SignierDeadline
	}
	return tse.SignierDeadline
}

// SignEvent signiert ein Event bei der TSE (Start + Finish, atomares Muster).
// Die tx-ID wird einmalig als UUIDv4 erzeugt und ueber embedTSE als `tseTxId`
// in den Event-Daten persistiert. Schlaegt die Signierung fehl, wird das Event
// unsigniert zurueckgegeben und ein NachsignierAuftrag fuer den Worker erstellt.
func (s Signierer) SignEvent(
	ctx context.Context,
	evt event.Event,
	processType string,
	processData string,
	embedTSE EmbedTSE,
) (Signierung, error) {
	if s.SettingsRepo == nil || s.NewTSEClient == nil {
		return Signierung{Event: evt}, nil
	}

	log := zerolog.Ctx(ctx)

	conf, err := s.SettingsRepo.GetTSEKonfiguration(ctx)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return Signierung{Event: evt}, nil
		}
		log.Error().Err(err).Msg("Failed to load TSE-Konfiguration for signierung")
		return Signierung{}, ErrDatabase
	}
	if !conf.IstKonfiguriert() {
		return Signierung{Event: evt}, nil
	}

	// Gesamt-Deadline fuer den synchronen Signierversuch: Bei fiskaly-Stoerung
	// wartet der Kassier-Request hoechstens diese Zeitspanne, danach greift der
	// Ausfallpfad (Nachsignier-Auftrag fuer den Worker).
	ctx, cancel := context.WithTimeout(ctx, s.deadline())
	defer cancel()

	txID := uuid.New().String()

	client, err := s.NewTSEClient(tse.Credentials{
		ApiKey:    conf.ApiKey,
		ApiSecret: conf.ApiSecret,
		TssID:     conf.TssID,
		ClientID:  conf.ClientID,
	})
	if err != nil {
		return signierAusfall(log, evt, txID, processType, processData, err, embedTSE)
	}

	startResult, err := client.StartTransaction(ctx, txID)
	if err != nil {
		return signierAusfall(log, evt, txID, processType, processData, err, embedTSE)
	}

	finishResult, err := client.FinishTransaction(ctx, txID, processType, processData)
	if err != nil {
		return signierAusfall(log, evt, txID, processType, processData, err, embedTSE)
	}

	tseData := kasse.TSEData{
		TransactionNumber: finishResult.TransactionNumber,
		SignatureCounter:  finishResult.SignatureCounter,
		SerialNumberTSE:   strings.TrimSpace(finishResult.SerialNumberTSE),
		LogTimeStart:      timeString(nonZeroTime(startResult.LogTime, finishResult.LogTimeStart)),
		LogTimeEnd:        timeString(nonZeroTime(finishResult.LogTime, finishResult.LogTimeEnd)),
		Signature:         strings.TrimSpace(finishResult.Signature),
		ProcessType:       processType,
		QRCodeData:        strings.TrimSpace(finishResult.QRCodeData),
	}
	if err := tseData.Validate(); err != nil {
		log.Error().Err(err).Str("tx_id", txID).Msg("TSE returned invalid signature data")
		return Signierung{}, ErrDatabase
	}

	signedEvent, err := embedTSE(evt, txID, &tseData)
	if err != nil {
		log.Error().Err(err).Msg("Failed to embed TSE data into event")
		return Signierung{}, ErrDatabase
	}

	return Signierung{Event: signedEvent}, nil
}

func signierAusfall(log *zerolog.Logger, evt event.Event, txID string, processType string, processData string, cause error, embedTSE EmbedTSE) (Signierung, error) {
	log.Warn().Err(cause).Str("tx_id", txID).Msg("TSE-Signierung fehlgeschlagen, Vorgang wird unsigniert persistiert und zur Nachsignierung vorgemerkt")

	unsignedEvent, err := embedTSE(evt, txID, nil)
	if err != nil {
		log.Error().Err(err).Msg("Failed to mark event with TSE-Ausfall")
		return Signierung{}, ErrDatabase
	}

	return Signierung{
		Event: unsignedEvent,
		NachsignierAuftrag: &NachsignierAuftrag{
			TxID:        txID,
			ProcessType: processType,
			ProcessData: processData,
		},
	}, nil
}

func timeString(value time.Time) string {
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
