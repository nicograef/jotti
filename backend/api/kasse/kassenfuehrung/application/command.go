package application

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/betreiber"
	"github.com/nicograef/jotti/backend/domain/event"
	"github.com/nicograef/jotti/backend/domain/kasse"
	"github.com/nicograef/jotti/backend/domain/tse"
	"github.com/nicograef/jotti/backend/repository/kassenjournal_repo"
	"github.com/rs/zerolog"
)

type kassenjournalRepo interface {
	EroeffneKassensitzung(ctx context.Context, datum time.Time, bezeichnung string, build func(zNr int) (event.Event, error)) (int, error)
	WriteEvent(ctx context.Context, e event.Event, streamType kasse.StreamType, kassensitzungNr int) (int, error)
	GetMaxVersion(ctx context.Context, subject string) (int, error)
	ReadEventsBySubject(ctx context.Context, subject string) ([]event.Event, error)
	GetKassenbestand(ctx context.Context, kassensitzungNr int) (kasse.Kassenbestand, error)
	GetGeldtransitListe(ctx context.Context, kassensitzungNr int) ([]kasse.Geldtransit, error)
	GetTischSessionsByKassensitzungNr(ctx context.Context, kassensitzungNr int) ([]kasse.TischSession, error)
	ReadKassensitzungEvents(ctx context.Context, kassensitzungNr int) ([]event.Event, error)
	EventExistsByTypeAndVorgangsID(ctx context.Context, eventType, vorgangsID, jsonKey string) (bool, error)
}

type kassensitzungenRepo interface {
	GetAktiveKassensitzung(ctx context.Context) (*kasse.Kassensitzung, error)
	SetKassensitzungWirdAbgeschlossen(ctx context.Context, zNr int) (int64, error)
	SetKassensitzungOffen(ctx context.Context, zNr int) (int64, error)
}

type betreiberRepo interface {
	GetBetreiber(ctx context.Context) (betreiber.Betreiber, error)
}

// tseGateRepo: Signatur-Stände und Störungszeitraum füttern das Kassenabschluss-Gate, die
// Konfiguration die Eröffnungs-Warnung ohne TSE.
type tseGateRepo interface {
	GetOffeneSignaturauftragStaendeFuerKassensitzung(ctx context.Context, kassensitzungNr int) ([]tse.SignaturauftragStand, error)
	GetAktiveTSEStoerung(ctx context.Context) (*tse.Stoerung, error)
	GetTSEKonfiguration(ctx context.Context) (tse.Konfiguration, error)
}

// druckauftragCleaner verwirft beim Tagesabschluss die verbliebenen fehlgeschlagenen
// Druckaufträge, damit die Outbox zum nächsten Einsatz leer startet.
type druckauftragCleaner interface {
	DiscardAlleFehlgeschlagenen(ctx context.Context) (int64, error)
}

type Command struct {
	KassenjournalRepo   kassenjournalRepo
	KassensitzungenRepo kassensitzungenRepo
	BetreiberRepo       betreiberRepo
	TSERepo             tseGateRepo
	DruckauftragRepo    druckauftragCleaner
}

// getOffeneKassensitzungOderFehler rejects the booking before any TSE roundtrip when no Kassensitzung is open or the barrier is active.
func (c Command) getOffeneKassensitzungOderFehler(ctx context.Context) (*kasse.Kassensitzung, error) {
	ks, err := c.KassensitzungenRepo.GetAktiveKassensitzung(ctx)
	if err != nil {
		return nil, ErrDatabase
	}
	if ks == nil {
		return nil, ErrKasseNichtGeoeffnet
	}
	if ks.Status == kasse.KassensitzungWirdAbgeschlossen {
		return nil, ErrKasseWirdAbgeschlossen
	}
	return ks, nil
}

// expectedVersion ist die Version des Zustands, gegen den der Command validiert hat
// (frischer Stream: 0). Ein UNIQUE(subject, version)-Konflikt wird zu ErrConflict.
func (c Command) writeKassensitzungEvent(ctx context.Context, e event.Event, kassensitzungNr int, expectedVersion int) error {
	log := zerolog.Ctx(ctx)

	subject := kasse.KassensitzungSubject(kassensitzungNr)
	e.Version = expectedVersion + 1

	_, err := c.KassenjournalRepo.WriteEvent(ctx, e, kasse.StreamTypeKassensitzung, kassensitzungNr)
	if err != nil {
		if errors.Is(err, db.ErrAlreadyExists) {
			log.Warn().Int("version", e.Version).Str("subject", subject).Msg("OCC Kassensitzung conflict")
			return ErrConflict
		}
		if errors.Is(err, db.ErrConflict) {
			log.Warn().Str("subject", subject).Msg("Deadlock on event write")
			return ErrConflict
		}
		if errors.Is(err, kassenjournal_repo.ErrKassensitzungNichtOffen) {
			log.Warn().Str("subject", subject).Msg("Kassensitzung nicht mehr offen")
			return ErrKasseNichtGeoeffnet
		}
		return ErrDatabase
	}

	return nil
}

// betriebstag liefert das Wandkalenderdatum in ort als Mitternacht UTC (passend zur DATE-Spalte).
// Ein Truncate(24h) schnitte auf UTC-Mitternacht und gäbe einer zwischen 00:00 und 02:00 Ortszeit
// eröffneten Sitzung das Datum des Vortags.
func betriebstag(now time.Time, ort *time.Location) time.Time {
	jahr, monat, tag := now.In(ort).Date()
	return time.Date(jahr, monat, tag, 0, 0, 0, 0, time.UTC)
}

func (c Command) KassensitzungEroeffnen(ctx context.Context, userID int, userName string, bezeichnung string, betragCents int) (int, error) {
	log := zerolog.Ctx(ctx)

	betreiber, err := c.BetreiberRepo.GetBetreiber(ctx)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			log.Warn().Msg("Kassensitzung blocked: betreiber not configured")
			return 0, ErrBetreiberNichtKonfiguriert
		}
		log.Error().Err(err).Msg("Failed to check betreiber configuration")
		return 0, ErrDatabase
	}
	if err = betreiber.Validate(); err != nil {
		log.Warn().Err(err).Msg("Kassensitzung blocked: betreiber not configured")
		return 0, ErrBetreiberNichtKonfiguriert
	}

	// Ohne konfigurierte TSE wird nicht gesperrt (jotti funktioniert vollständig ohne); der
	// unsignierte Betrieb soll nur im Log stehen. Best effort: ein Lesefehler unterdrückt die Warnung.
	ohneTSE := false
	if conf, err := c.TSERepo.GetTSEKonfiguration(ctx); err != nil {
		if errors.Is(err, db.ErrNotFound) {
			ohneTSE = true
		} else {
			log.Error().Err(err).Msg("Failed to load TSE-Konfiguration")
		}
	} else if !conf.IstKonfiguriert() {
		ohneTSE = true
	}

	// Aktive Sitzung deckt 'offen' und 'wird_abgeschlossen' ab: Während ein Abschluss läuft, darf
	// keine zweite Sitzung eröffnet werden (idx_kassensitzungen_eine_aktiv wäre die letzte Bremse).
	existing, err := c.KassensitzungenRepo.GetAktiveKassensitzung(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to check for existing active Kassensitzung")
		return 0, ErrDatabase
	}
	if existing != nil {
		log.Warn().Int("z_nr", existing.ZNr).Msg("Kassensitzung already active")
		return 0, ErrKasseAlreadyOpen
	}

	berliner, err := time.LoadLocation("Europe/Berlin")
	if err != nil {
		log.Error().Err(err).Msg("Failed to load Europe/Berlin timezone")
		return 0, ErrDatabase
	}
	datum := betriebstag(time.Now(), berliner)

	// Entität und Eröffnungs-Event entstehen atomar in einer Transaktion; der Signaturauftrag (bei
	// Anfangsbestand > 0) im selben Commit über die fiskalische Projektion. Buchen blockiert nie
	// auf die TSE.
	zNr, err := c.KassenjournalRepo.EroeffneKassensitzung(ctx, datum, bezeichnung, func(zNr int) (event.Event, error) {
		evt, err := kasse.NewKassensitzungEroeffnetEvent(kasse.KassensitzungSubject(zNr), userID, userName, datum.Format("2006-01-02"), bezeichnung, betragCents)
		if err != nil {
			log.Error().Err(err).Int("z_nr", zNr).Msg("Failed to create kassensitzung-eroeffnet event")
			return event.Event{}, err
		}
		// Frischer Stream (neue z_nr): das Eröffnungs-Event ist version = 1.
		evt.Version = 1
		return evt, nil
	})
	if err != nil {
		if errors.Is(err, ErrDatabase) || errors.Is(err, db.ErrDatabase) {
			log.Error().Err(err).Msg("Failed to open Kassensitzung")
			return 0, ErrDatabase
		}
		return 0, err
	}

	log.Info().Int("z_nr", zNr).Msg("Kassensitzung eroeffnet")
	if ohneTSE {
		log.Warn().Int("z_nr", zNr).Msg("Kassensitzung ohne TSE-Konfiguration eroeffnet; Vorgaenge werden nicht signiert")
	}
	return zNr, nil
}

// geldtransitID ist eine client-seitig erzeugte UUID (Idempotenz-Schlüssel). Bei
// UniqueViolation (Duplikat-Einreichung) wird per geldtransitId nachgeschlagen:
// Treffer = idempotente Erfolgsantwort; kein Treffer = echter OCC-Konflikt (409).
// Gleiche ID bedeutet denselben Vorgang — der Payload wird nicht verglichen.
func (c Command) GeldtransitBuchen(ctx context.Context, userID int, userName string, geldtransitID string, richtung string, betragCents int, kommentar string) error {
	log := zerolog.Ctx(ctx)

	ks, err := c.getOffeneKassensitzungOderFehler(ctx)
	if err != nil {
		return err
	}

	evt, err := kasse.NewGeldtransitGebuchtEvent(kasse.KassensitzungSubject(ks.ZNr), userID, userName, geldtransitID, richtung, betragCents, kommentar)
	if err != nil {
		log.Error().Err(err).Int("z_nr", ks.ZNr).Msg("Failed to create geldtransit-gebucht event")
		return err
	}

	// Geldtransit validiert keinen Stream-Zustand (reines Anhängen); die Version kommt erst unmittelbar vor dem Schreiben.
	maxVersion, err := c.KassenjournalRepo.GetMaxVersion(ctx, kasse.KassensitzungSubject(ks.ZNr))
	if err != nil {
		log.Error().Err(err).Int("z_nr", ks.ZNr).Msg("Failed to load max version for geldtransit")
		return ErrDatabase
	}

	if err := c.writeKassensitzungEvent(ctx, evt, ks.ZNr, maxVersion); err != nil {
		if errors.Is(err, ErrConflict) {
			exists, lookupErr := c.KassenjournalRepo.EventExistsByTypeAndVorgangsID(ctx, string(kasse.EventTypeGeldtransitGebuchtV1), geldtransitID, "geldtransitId")
			if lookupErr != nil {
				log.Error().Err(lookupErr).Str("geldtransit_id", geldtransitID).Msg("Failed to lookup geldtransit idempotency")
				return ErrDatabase
			}
			if exists {
				log.Info().Str("geldtransit_id", geldtransitID).Msg("Idempotenter Geldtransit: geldtransitId bereits vorhanden")
				return nil
			}
			return ErrConflict
		}
		return err
	}

	log.Info().Int("z_nr", ks.ZNr).Str("richtung", richtung).Int("betrag_cents", betragCents).Msg("Geldtransit gebucht")
	return nil
}

// KasseAbschliessen schreibt Kassensturz, Differenzbuchung (nur bei Differenz ungleich Null) und
// Tagesabschluss in dieser Reihenfolge. Es gibt bewusst keine umschließende Transaktion über die
// drei: ein Teilfehler wird durch einen erneuten Aufruf fortgesetzt, der einen bereits
// geschriebenen Kassensturz idempotent überspringt. Die Tagessummen berechnet
// kasse.ComputeAbschlussSummen; dass sie nicht von der SQL-Auswertung abweichen, sichert
// backend/repository/reporting_repo/summen_abschluss_test.go zu.
func (c Command) KasseAbschliessen(ctx context.Context, userID int, userName string, istBestandCents int) (ergebnis KassenabschlussErgebnis, err error) {
	log := zerolog.Ctx(ctx)

	// Aktive Sitzung akzeptiert 'offen' und 'wird_abgeschlossen' (Wiederanlauf im Zwischenstatus).
	ks, err := c.KassensitzungenRepo.GetAktiveKassensitzung(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to load Kassensitzung for Kassenabschluss")
		return KassenabschlussErgebnis{}, ErrDatabase
	}
	if ks == nil {
		return KassenabschlussErgebnis{}, ErrKasseNichtGeoeffnet
	}

	// Signatur-Gate vor der Barriere: prüft sofort, wartet nie. Ein ausstehender
	// Auftrag blockiert; nichts wurde bis hier verändert, kein Reset nötig.
	gate, err := c.checkSignaturGate(ctx, ks.ZNr)
	if err != nil {
		return KassenabschlussErgebnis{}, err
	}
	if gate.ausstehendAnzahl > 0 {
		log.Warn().Int("z_nr", ks.ZNr).Int("ausstehend", gate.ausstehendAnzahl).Msg("Kassenabschluss blockiert: Signaturen ausstehend")
		return KassenabschlussErgebnis{}, &SignaturenAusstehendError{
			Anzahl: gate.ausstehendAnzahl,
		}
	}
	ergebnis = KassenabschlussErgebnis{
		AusfallResteAnzahl:      gate.ausfallResteAnzahl,
		OhneKonfigurationAnzahl: gate.ohneKonfigurationAnzahl,
	}

	// Barriere setzen: Der UPDATE wartet auf laufende Buchungen (FOR SHARE); danach lehnt der
	// Status-Guard jedes weitere Buchungs-Event ab. Idempotent für den Wiederholungs-Aufruf.
	rows, err := c.KassensitzungenRepo.SetKassensitzungWirdAbgeschlossen(ctx, ks.ZNr)
	if err != nil {
		log.Error().Err(err).Int("z_nr", ks.ZNr).Msg("Failed to set Kassensitzung status wird_abgeschlossen")
		return KassenabschlussErgebnis{}, ErrDatabase
	}
	if rows == 0 {
		return KassenabschlussErgebnis{}, ErrKasseNichtGeoeffnet
	}

	// Fehler nach dem Statuswechsel setzen die Sitzung best effort zurück auf 'offen'. Ausnahme: Ein
	// Versionskonflikt bedeutet einen konkurrierenden zweiten Abschluss — dann darf diese Instanz die
	// Barriere nicht unter dem gewinnenden Abschluss wegräumen.
	defer func() {
		if err != nil && !errors.Is(err, ErrConflict) {
			if _, resetErr := c.KassensitzungenRepo.SetKassensitzungOffen(ctx, ks.ZNr); resetErr != nil {
				log.Error().Err(resetErr).Int("z_nr", ks.ZNr).Msg("Failed to reset Kassensitzung status to offen after Abschluss error")
			}
		}
	}()

	subject := kasse.KassensitzungSubject(ks.ZNr)

	// OCC-Anker für die Abschluss-Events: Die Barriere verhindert bereits neue Buchungen; der
	// Anker erkennt zusätzlich einen konkurrierenden zweiten Abschluss (Versionskonflikt).
	expectedVersion, err := c.KassenjournalRepo.GetMaxVersion(ctx, subject)
	if err != nil {
		log.Error().Err(err).Int("z_nr", ks.ZNr).Msg("Failed to load max version for Kassenabschluss")
		return KassenabschlussErgebnis{}, ErrDatabase
	}

	kassenbestand, err := c.KassenjournalRepo.GetKassenbestand(ctx, ks.ZNr)
	if err != nil {
		log.Error().Err(err).Int("z_nr", ks.ZNr).Msg("Failed to get Kassenbestand for Kassenabschluss")
		return KassenabschlussErgebnis{}, ErrDatabase
	}
	sollBestandCents := kassenbestand.SollBestandCents

	// Wiederanlauf: Ein vorheriger Versuch kann den Kassensturz schon geschrieben haben. Der
	// dokumentierte Kassensturz zählt — Schritt 1 entfällt, sein Ist-Bestand bleibt maßgeblich.
	//
	// Zwischenbuchungen brechen ab: nach einem defer-Reset auf 'offen' können neue Buchungen
	// entstehen, die der alte Ist-Bestand als Soll-Ist-Differenz verbuchen würde. Zwei Signale
	// brechen ab: eine Buchung nach dem Kassensturz im Kassensitzungs-Stream und ein seither
	// veränderter Soll-Bestand (Tischzahlung, Warenrücknahme, Direktverkauf liegen in eigenen
	// Sub-Streams).
	vorhandenerSturz, buchungenNachSturz, err := c.findeVorhandenenKassensturz(ctx, subject)
	if err != nil {
		return KassenabschlussErgebnis{}, err
	}
	if buchungenNachSturz {
		log.Warn().Int("z_nr", ks.ZNr).Msg("Kassenabschluss-Wiederanlauf abgebrochen: Buchungen nach dem protokollierten Kassensturz")
		return KassenabschlussErgebnis{}, ErrBuchungenNachKassensturz
	}
	if vorhandenerSturz != nil {
		// Verglichen wird ohne die abschluss-eigene Differenzbuchung: steht sie schon, entspricht
		// sollBestandCents dem gezählten Ist-Bestand und verdeckte jede Zwischenbuchung. Der
		// Kassensturz protokolliert seinen Soll-Bestand ebenfalls ohne Differenz.
		sollOhneDifferenzCents := kassenbestand.SollBestandOhneDifferenzCents()
		if sollOhneDifferenzCents != vorhandenerSturz.SollBestandCents {
			log.Warn().Int("z_nr", ks.ZNr).
				Int("soll_kassensturz_cents", vorhandenerSturz.SollBestandCents).
				Int("soll_ohne_differenz_cents", sollOhneDifferenzCents).
				Msg("Kassenabschluss-Wiederanlauf abgebrochen: Soll-Bestand seit dem Kassensturz veraendert")
			return KassenabschlussErgebnis{}, ErrBuchungenNachKassensturz
		}
		log.Info().Int("z_nr", ks.ZNr).Msg("Kassenabschluss-Wiederanlauf: Kassensturz bereits vorhanden, Schritt 1 wird uebersprungen")
		istBestandCents = vorhandenerSturz.IstBestandCents
	}
	differenzCents := sollBestandCents - istBestandCents

	// Invariant: Tisch-Saldo-Sperre — all tisch sessions must have saldo_cents = 0
	sessions, err := c.KassenjournalRepo.GetTischSessionsByKassensitzungNr(ctx, ks.ZNr)
	if err != nil {
		log.Error().Err(err).Int("z_nr", ks.ZNr).Msg("Failed to get tisch sessions for Kassenabschluss")
		return KassenabschlussErgebnis{}, ErrDatabase
	}
	for _, s := range sessions {
		if s.SaldoCents != 0 {
			log.Warn().Int("z_nr", ks.ZNr).Int("tisch_id", s.TischID).Int("saldo_cents", s.SaldoCents).
				Msg("Kassenabschluss rejected: Tisch has non-zero saldo")
			return KassenabschlussErgebnis{}, ErrTischeSaldoOffen
		}
	}

	// 1. Kassensturz (entfällt im Wiederanlauf: bereits im Journal)
	if vorhandenerSturz == nil {
		kassensturzEvt, err := kasse.NewKassensturzDurchgefuehrtEvent(subject, userID, userName, sollBestandCents, istBestandCents, differenzCents)
		if err != nil {
			log.Error().Err(err).Int("z_nr", ks.ZNr).Msg("Failed to create kassensturz-durchgefuehrt event")
			return KassenabschlussErgebnis{}, err
		}
		if err := c.writeKassensitzungEvent(ctx, kassensturzEvt, ks.ZNr, expectedVersion); err != nil {
			return KassenabschlussErgebnis{}, err
		}
		expectedVersion++
	}

	// 2. Differenzbuchung nur bei Differenz ungleich Null
	if differenzCents != 0 {
		diffEvt, err := kasse.NewDifferenzSollIstGebuchtEvent(subject, userID, userName, differenzCents)
		if err != nil {
			log.Error().Err(err).Int("z_nr", ks.ZNr).Msg("Failed to create differenz-soll-ist-gebucht event")
			return KassenabschlussErgebnis{}, err
		}
		if err := c.writeKassensitzungEvent(ctx, diffEvt, ks.ZNr, expectedVersion); err != nil {
			return KassenabschlussErgebnis{}, err
		}
		expectedVersion++
	}

	// 3. Tagesabschluss: Kassensturz- und Differenz-Events sind bereits committed (kein
	// In-Memory-Anhängen nötig; die Differenzbuchung ist summen-neutral).
	sitzungEvents, err := c.KassenjournalRepo.ReadKassensitzungEvents(ctx, ks.ZNr)
	if err != nil {
		log.Error().Err(err).Int("z_nr", ks.ZNr).Msg("Failed to read events for Tagesabschluss")
		return KassenabschlussErgebnis{}, ErrDatabase
	}
	summen, err := kasse.ComputeAbschlussSummen(sitzungEvents)
	if err != nil {
		log.Error().Err(err).Int("z_nr", ks.ZNr).Msg("Unparsebares Event verhindert Tagesabschluss")
		return KassenabschlussErgebnis{}, err
	}

	now := time.Now().UTC()
	tagesabschlussEvt, err := kasse.NewTagesabschlussErstelltEvent(
		subject, userID, userName,
		ks.ZNr,
		ks.CreatedAt, now,
		summen.UmsatzCents, summen.StornierungCents,
		summen.GeldtransitCents,
	)
	if err != nil {
		log.Error().Err(err).Int("z_nr", ks.ZNr).Msg("Failed to create tagesabschluss-erstellt event")
		return KassenabschlussErgebnis{}, err
	}
	if err := c.writeKassensitzungEvent(ctx, tagesabschlussEvt, ks.ZNr, expectedVersion); err != nil {
		return KassenabschlussErgebnis{}, err
	}

	// Druck-Outbox aufräumen: mit dem committeten Tagesabschluss ist die Sitzung fiskalisch
	// geschlossen. Best effort — der Fehler wird NICHT in den benannten Return err geschrieben,
	// sonst meldete der Abschluss einen Fehler, obwohl der Tagesabschluss committed ist; der
	// defer-Reset selbst bliebe folgenlos, weil SetKassensitzungOffen nur in
	// 'wird_abgeschlossen' greift. Der Cleaner ist optional (nil-guard).
	if c.DruckauftragRepo != nil {
		if verworfen, cleanupErr := c.DruckauftragRepo.DiscardAlleFehlgeschlagenen(ctx); cleanupErr != nil {
			log.Error().Err(cleanupErr).Int("z_nr", ks.ZNr).Msg("Failed to discard fehlgeschlagene Druckauftraege beim Tagesabschluss (Abschluss bleibt gueltig)")
		} else if verworfen > 0 {
			log.Info().Int("z_nr", ks.ZNr).Int64("verworfen", verworfen).Msg("Fehlgeschlagene Druckauftraege beim Tagesabschluss verworfen")
		}
	}

	log.Info().Int("z_nr", ks.ZNr).
		Int("soll_cents", sollBestandCents).
		Int("ist_cents", istBestandCents).
		Int("differenz_cents", differenzCents).
		Int("ausfall_reste", ergebnis.AusfallResteAnzahl).
		Int("ohne_konfiguration", ergebnis.OhneKonfigurationAnzahl).
		Msg("Kasse abgeschlossen")
	return ergebnis, nil
}

// findeVorhandenenKassensturz liefert das bereits im Stream stehende
// kassensturz-durchgefuehrt-Event (oder nil) und ob danach eine Zwischenbuchung liegt.
// buchungenNachSturz ist true, sobald nach dem Kassensturz ein Event liegt, das nicht zum Abschluss
// selbst gehört (kasse.IsAbschlussEventType nimmt die aus). Der Kassensitzungs-Stream kennt als
// solche Zwischenbuchung nur geldtransit-gebucht:v1; Tisch-Buchungen laufen über eigene Sub-Streams.
func (c Command) findeVorhandenenKassensturz(ctx context.Context, subject string) (sturz *kasse.KassensturzDurchgefuehrtV1Data, buchungenNachSturz bool, err error) {
	log := zerolog.Ctx(ctx)

	events, err := c.KassenjournalRepo.ReadEventsBySubject(ctx, subject)
	if err != nil {
		log.Error().Err(err).Str("subject", subject).Msg("Failed to read events for Kassensturz-Wiederanlauf check")
		return nil, false, ErrDatabase
	}
	for _, evt := range events {
		if sturz != nil {
			if !kasse.IsAbschlussEventType(evt.Type) {
				return sturz, true, nil
			}
			continue
		}
		if kasse.EventType(evt.Type) != kasse.EventTypeKassensturzDurchgefuehrtV1 {
			continue
		}
		var data kasse.KassensturzDurchgefuehrtV1Data
		if err := json.Unmarshal(evt.Data, &data); err != nil {
			log.Error().Err(err).Int("event_id", evt.ID).Msg("Unparsebares Kassensturz-Event verhindert Kassenabschluss")
			return nil, false, fmt.Errorf("event %d (%s): %w", evt.ID, evt.Type, err)
		}
		sturz = &data
	}
	return sturz, false, nil
}
