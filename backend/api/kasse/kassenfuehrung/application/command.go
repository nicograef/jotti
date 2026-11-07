package application

import (
	"context"
	"errors"
	"time"

	"github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/betreiber"
	"github.com/nicograef/jotti/backend/domain/event"
	"github.com/nicograef/jotti/backend/domain/kasse"
	"github.com/nicograef/jotti/backend/domain/tse"
	"github.com/nicograef/jotti/backend/internal/logging"
	"github.com/nicograef/jotti/backend/internal/zeit"
	"github.com/nicograef/jotti/backend/repository/kassenjournal_repo"
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
		logging.FromContext(ctx).Error("Failed to load aktive Kassensitzung", "error", err)
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
	log := logging.FromContext(ctx)

	subject := kasse.KassensitzungSubject(kassensitzungNr)
	e.Version = expectedVersion + 1

	_, err := c.KassenjournalRepo.WriteEvent(ctx, e, kasse.StreamTypeKassensitzung, kassensitzungNr)
	if err != nil {
		if errors.Is(err, db.ErrAlreadyExists) {
			log.Warn("OCC Kassensitzung conflict", "version", e.Version, "subject", subject)
			return ErrConflict
		}
		if errors.Is(err, db.ErrConflict) {
			log.Warn("Deadlock on event write", "subject", subject)
			return ErrConflict
		}
		if errors.Is(err, kassenjournal_repo.ErrKassensitzungNichtOffen) {
			log.Warn("Kassensitzung nicht mehr offen", "subject", subject)
			return ErrKasseNichtGeoeffnet
		}
		log.Error("Failed to write Kassensitzung event", "error", err, "subject", subject, "version", e.Version)
		return ErrDatabase
	}

	return nil
}

func (c Command) KassensitzungEroeffnen(ctx context.Context, userID int, userName string, bezeichnung string, betragCents int) (int, error) {
	log := logging.FromContext(ctx)

	if err := c.pruefeBetreiber(ctx); err != nil {
		return 0, err
	}
	ohneTSE := c.ohneTSEKonfiguration(ctx)

	// Aktive Sitzung deckt 'offen' und 'wird_abgeschlossen' ab: Während ein Abschluss läuft, darf
	// keine zweite Sitzung eröffnet werden (idx_kassensitzungen_eine_aktiv ist die letzte Bremse).
	existing, err := c.KassensitzungenRepo.GetAktiveKassensitzung(ctx)
	if err != nil {
		log.Error("Failed to check for existing active Kassensitzung", "error", err)
		return 0, ErrDatabase
	}
	if existing != nil {
		log.Warn("Kassensitzung already active", "z_nr", existing.ZNr)
		return 0, ErrKasseAlreadyOpen
	}

	datum := zeit.Kalendertag(time.Now())

	// Entität und Eröffnungs-Event entstehen atomar in einer Transaktion; der Signaturauftrag (bei
	// Anfangsbestand > 0) im selben Commit über die fiskalische Projektion. Buchen blockiert nie
	// auf die TSE.
	zNr, err := c.KassenjournalRepo.EroeffneKassensitzung(ctx, datum, bezeichnung, func(zNr int) (event.Event, error) {
		evt, err := kasse.NewKassensitzungEroeffnetEvent(kasse.KassensitzungSubject(zNr), userID, userName, datum.Format("2006-01-02"), bezeichnung, betragCents)
		if err != nil {
			log.Error("Failed to create kassensitzung-eroeffnet event", "error", err, "z_nr", zNr)
			return event.Event{}, err
		}
		// Frischer Stream (neue z_nr): das Eröffnungs-Event ist version = 1.
		evt.Version = 1
		return evt, nil
	})
	if err != nil {
		// idx_kassensitzungen_eine_aktiv fängt eine parallele Eröffnung, die die Prüfung oben überholt hat.
		if errors.Is(err, db.ErrAlreadyExists) {
			log.Warn("Kassensitzung already active (concurrent open)", "error", err)
			return 0, ErrKasseAlreadyOpen
		}
		if errors.Is(err, ErrDatabase) {
			log.Error("Failed to open Kassensitzung", "error", err)
			return 0, ErrDatabase
		}
		return 0, err
	}

	log.Info("Kassensitzung eroeffnet", "z_nr", zNr)
	if ohneTSE {
		log.Warn("Kassensitzung ohne TSE-Konfiguration eroeffnet; Vorgaenge werden nicht signiert", "z_nr", zNr)
	}
	return zNr, nil
}

func (c Command) pruefeBetreiber(ctx context.Context) error {
	log := logging.FromContext(ctx)

	betreiber, err := c.BetreiberRepo.GetBetreiber(ctx)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			log.Warn("Kassensitzung blocked: betreiber not configured")
			return ErrBetreiberNichtKonfiguriert
		}
		log.Error("Failed to check betreiber configuration", "error", err)
		return ErrDatabase
	}
	if err = betreiber.Validate(); err != nil {
		log.Warn("Kassensitzung blocked: betreiber not configured", "error", err)
		return ErrBetreiberNichtKonfiguriert
	}
	return nil
}

// ohneTSEKonfiguration: ohne konfigurierte TSE wird nicht gesperrt (jotti funktioniert vollständig
// ohne); der unsignierte Betrieb soll nur im Log stehen. Best effort: ein Lesefehler unterdrückt die Warnung.
func (c Command) ohneTSEKonfiguration(ctx context.Context) bool {
	conf, err := c.TSERepo.GetTSEKonfiguration(ctx)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return true
		}
		logging.FromContext(ctx).Error("Failed to load TSE-Konfiguration", "error", err)
		return false
	}
	return !conf.IstKonfiguriert()
}

// geldtransitID is the client's idempotency key: on a UniqueViolation a stored event with that
// geldtransitId means idempotent success (same ID = same operation, payload not compared), none
// means a real OCC conflict (409).
func (c Command) GeldtransitBuchen(ctx context.Context, userID int, userName string, geldtransitID string, richtung string, betragCents int, kommentar string) error {
	log := logging.FromContext(ctx)

	ks, err := c.getOffeneKassensitzungOderFehler(ctx)
	if err != nil {
		return err
	}

	evt, err := kasse.NewGeldtransitGebuchtEvent(kasse.KassensitzungSubject(ks.ZNr), userID, userName, geldtransitID, richtung, betragCents, kommentar)
	if err != nil {
		log.Error("Failed to create geldtransit-gebucht event", "error", err, "z_nr", ks.ZNr)
		return err
	}

	// Geldtransit validiert keinen Stream-Zustand (reines Anhängen); die Version kommt erst unmittelbar vor dem Schreiben.
	maxVersion, err := c.KassenjournalRepo.GetMaxVersion(ctx, kasse.KassensitzungSubject(ks.ZNr))
	if err != nil {
		log.Error("Failed to load max version for geldtransit", "error", err, "z_nr", ks.ZNr)
		return ErrDatabase
	}

	if err := c.writeKassensitzungEvent(ctx, evt, ks.ZNr, maxVersion); err != nil {
		if errors.Is(err, ErrConflict) {
			exists, lookupErr := c.KassenjournalRepo.EventExistsByTypeAndVorgangsID(ctx, string(kasse.EventTypeGeldtransitGebuchtV1), geldtransitID, "geldtransitId")
			if lookupErr != nil {
				log.Error("Failed to lookup geldtransit idempotency", "error", lookupErr, "geldtransit_id", geldtransitID)
				return ErrDatabase
			}
			if exists {
				log.Info("Idempotenter Geldtransit: geldtransitId bereits vorhanden", "geldtransit_id", geldtransitID)
				return nil
			}
			return ErrConflict
		}
		return err
	}

	log.Info("Geldtransit gebucht", "z_nr", ks.ZNr, "richtung", richtung, "betrag_cents", betragCents)
	return nil
}

// KasseAbschliessen writes Kassensturz, Differenzbuchung (only if nonzero) and Tagesabschluss without
// an enclosing transaction: a retry resumes and skips a written Kassensturz (docs/handbuch.md §3.10).
// reporting_repo/summen_abschluss_test.go pins kasse.ComputeAbschlussSummen to the SQL reporting sums.
func (c Command) KasseAbschliessen(ctx context.Context, userID int, userName string, istBestandCents int) (ergebnis KassenabschlussErgebnis, err error) {
	log := logging.FromContext(ctx)

	// Aktive Sitzung akzeptiert 'offen' und 'wird_abgeschlossen' (Wiederanlauf im Zwischenstatus).
	ks, err := c.KassensitzungenRepo.GetAktiveKassensitzung(ctx)
	if err != nil {
		log.Error("Failed to load Kassensitzung for Kassenabschluss", "error", err)
		return KassenabschlussErgebnis{}, ErrDatabase
	}
	if ks == nil {
		return KassenabschlussErgebnis{}, ErrKasseNichtGeoeffnet
	}

	// Signatur-Gate vor der Barriere: prüft sofort, wartet nie. Ein ausstehender
	// Auftrag blockiert; nichts wurde bis hier verändert, kein Reset nötig.
	ergebnis, err = c.pruefeSignaturenVorAbschluss(ctx, ks.ZNr)
	if err != nil {
		return KassenabschlussErgebnis{}, err
	}

	// Barriere setzen: Der UPDATE wartet auf laufende Buchungen (FOR SHARE); danach lehnt der
	// Status-Guard jedes weitere Buchungs-Event ab. Idempotent für den Wiederholungs-Aufruf.
	rows, err := c.KassensitzungenRepo.SetKassensitzungWirdAbgeschlossen(ctx, ks.ZNr)
	if err != nil {
		log.Error("Failed to set Kassensitzung status wird_abgeschlossen", "error", err, "z_nr", ks.ZNr)
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
				log.Error("Failed to reset Kassensitzung status to offen after Abschluss error", "error", resetErr, "z_nr", ks.ZNr)
			}
		}
	}()

	subject := kasse.KassensitzungSubject(ks.ZNr)

	// OCC-Anker für die Abschluss-Events: Die Barriere verhindert bereits neue Buchungen; der
	// Anker erkennt zusätzlich einen konkurrierenden zweiten Abschluss (Versionskonflikt).
	expectedVersion, err := c.KassenjournalRepo.GetMaxVersion(ctx, subject)
	if err != nil {
		log.Error("Failed to load max version for Kassenabschluss", "error", err, "z_nr", ks.ZNr)
		return KassenabschlussErgebnis{}, ErrDatabase
	}

	bestand, err := c.ermittleAbschlussBestand(ctx, ks.ZNr, istBestandCents)
	if err != nil {
		return KassenabschlussErgebnis{}, err
	}

	if err := c.pruefeTischSalden(ctx, ks.ZNr); err != nil {
		return KassenabschlussErgebnis{}, err
	}

	// 1. Kassensturz (entfällt im Wiederanlauf: bereits im Journal)
	if !bestand.sturzVorhanden {
		kassensturzEvt, err := kasse.NewKassensturzDurchgefuehrtEvent(subject, userID, userName, bestand.sollCents, bestand.istCents, bestand.differenzCents())
		if err != nil {
			log.Error("Failed to create kassensturz-durchgefuehrt event", "error", err, "z_nr", ks.ZNr)
			return KassenabschlussErgebnis{}, err
		}
		if err := c.writeKassensitzungEvent(ctx, kassensturzEvt, ks.ZNr, expectedVersion); err != nil {
			return KassenabschlussErgebnis{}, err
		}
		expectedVersion++
	}

	// 2. Differenzbuchung nur bei Differenz ungleich Null
	if bestand.differenzCents() != 0 {
		diffEvt, err := kasse.NewDifferenzSollIstGebuchtEvent(subject, userID, userName, bestand.differenzCents())
		if err != nil {
			log.Error("Failed to create differenz-soll-ist-gebucht event", "error", err, "z_nr", ks.ZNr)
			return KassenabschlussErgebnis{}, err
		}
		if err := c.writeKassensitzungEvent(ctx, diffEvt, ks.ZNr, expectedVersion); err != nil {
			return KassenabschlussErgebnis{}, err
		}
		expectedVersion++
	}

	// 3. Tagesabschluss
	if err := c.schreibeTagesabschluss(ctx, ks, userID, userName, expectedVersion); err != nil {
		return KassenabschlussErgebnis{}, err
	}

	c.raeumeDruckOutboxAuf(ctx, ks.ZNr)

	log.Info("Kasse abgeschlossen", "z_nr", ks.ZNr, "soll_cents", bestand.sollCents, "ist_cents", bestand.istCents, "differenz_cents", bestand.differenzCents(), "ausfall_reste", ergebnis.AusfallResteAnzahl, "ohne_konfiguration", ergebnis.OhneKonfigurationAnzahl)
	return ergebnis, nil
}

// pruefeSignaturenVorAbschluss blocks on an ausstehend Signaturauftrag and reports the Ausfall rests.
func (c Command) pruefeSignaturenVorAbschluss(ctx context.Context, zNr int) (KassenabschlussErgebnis, error) {
	gate, err := c.checkSignaturGate(ctx, zNr)
	if err != nil {
		return KassenabschlussErgebnis{}, err
	}
	if gate.ausstehendAnzahl > 0 {
		logging.FromContext(ctx).Warn("Kassenabschluss blockiert: Signaturen ausstehend", "z_nr", zNr, "ausstehend", gate.ausstehendAnzahl)
		return KassenabschlussErgebnis{}, &SignaturenAusstehendError{
			Anzahl: gate.ausstehendAnzahl,
		}
	}
	return KassenabschlussErgebnis{
		AusfallResteAnzahl:      gate.ausfallResteAnzahl,
		OhneKonfigurationAnzahl: gate.ohneKonfigurationAnzahl,
	}, nil
}

// abschlussBestand: sturzVorhanden marks a Wiederanlauf whose logged Kassensturz fixes istCents.
type abschlussBestand struct {
	sollCents      int
	istCents       int
	sturzVorhanden bool
}

func (b abschlussBestand) differenzCents() int { return b.sollCents - b.istCents }

// ermittleAbschlussBestand: in a Wiederanlauf an earlier attempt's Kassensturz counts, with its
// Ist-Bestand. Bookings since then abort, or the old Ist-Bestand would book them as Differenz: a later
// event in this stream (Geldtransit) or a changed Soll-Bestand (Tisch and Direktverkauf sub-streams)
// reveals them.
func (c Command) ermittleAbschlussBestand(ctx context.Context, zNr int, istBestandCents int) (abschlussBestand, error) {
	log := logging.FromContext(ctx)

	kassenbestand, err := c.KassenjournalRepo.GetKassenbestand(ctx, zNr)
	if err != nil {
		log.Error("Failed to get Kassenbestand for Kassenabschluss", "error", err, "z_nr", zNr)
		return abschlussBestand{}, ErrDatabase
	}

	vorhandenerSturz, buchungenNachSturz, err := c.findeVorhandenenKassensturz(ctx, kasse.KassensitzungSubject(zNr))
	if err != nil {
		return abschlussBestand{}, err
	}
	if buchungenNachSturz {
		log.Warn("Kassenabschluss-Wiederanlauf abgebrochen: Buchungen nach dem protokollierten Kassensturz", "z_nr", zNr)
		return abschlussBestand{}, ErrBuchungenNachKassensturz
	}
	bestand := abschlussBestand{sollCents: kassenbestand.SollBestandCents, istCents: istBestandCents}
	if vorhandenerSturz == nil {
		return bestand, nil
	}

	// Verglichen wird ohne die abschluss-eigene Differenzbuchung: steht sie schon, entspricht
	// SollBestandCents dem gezählten Ist-Bestand und verdeckte jede Zwischenbuchung. Der
	// Kassensturz protokolliert seinen Soll-Bestand ebenfalls ohne Differenz.
	sollOhneDifferenzCents := kassenbestand.SollBestandOhneDifferenzCents()
	if sollOhneDifferenzCents != vorhandenerSturz.SollBestandCents {
		log.Warn("Kassenabschluss-Wiederanlauf abgebrochen: Soll-Bestand seit dem Kassensturz veraendert", "z_nr", zNr, "soll_kassensturz_cents", vorhandenerSturz.SollBestandCents, "soll_ohne_differenz_cents", sollOhneDifferenzCents)
		return abschlussBestand{}, ErrBuchungenNachKassensturz
	}
	log.Info("Kassenabschluss-Wiederanlauf: Kassensturz bereits vorhanden, Schritt 1 wird uebersprungen", "z_nr", zNr)
	bestand.istCents = vorhandenerSturz.IstBestandCents
	bestand.sturzVorhanden = true
	return bestand, nil
}

// pruefeTischSalden enforces the Tisch-Saldo-Sperre: every tisch session must have saldo_cents = 0.
func (c Command) pruefeTischSalden(ctx context.Context, zNr int) error {
	log := logging.FromContext(ctx)

	sessions, err := c.KassenjournalRepo.GetTischSessionsByKassensitzungNr(ctx, zNr)
	if err != nil {
		log.Error("Failed to get tisch sessions for Kassenabschluss", "error", err, "z_nr", zNr)
		return ErrDatabase
	}
	for _, s := range sessions {
		if s.SaldoCents != 0 {
			log.Warn("Kassenabschluss rejected: Tisch has non-zero saldo", "z_nr", zNr, "tisch_id", s.TischID, "saldo_cents", s.SaldoCents)
			return ErrTischeSaldoOffen
		}
	}
	return nil
}

// schreibeTagesabschluss: Kassensturz- und Differenz-Events sind bereits committed (kein
// In-Memory-Anhängen nötig; die Differenzbuchung ist summen-neutral).
func (c Command) schreibeTagesabschluss(ctx context.Context, ks *kasse.Kassensitzung, userID int, userName string, expectedVersion int) error {
	log := logging.FromContext(ctx)

	sitzungEvents, err := c.KassenjournalRepo.ReadKassensitzungEvents(ctx, ks.ZNr)
	if err != nil {
		log.Error("Failed to read events for Tagesabschluss", "error", err, "z_nr", ks.ZNr)
		return ErrDatabase
	}
	summen, err := kasse.ComputeAbschlussSummen(sitzungEvents)
	if err != nil {
		log.Error("Unparsebares Event verhindert Tagesabschluss", "error", err, "z_nr", ks.ZNr)
		return err
	}

	tagesabschlussEvt, err := kasse.NewTagesabschlussErstelltEvent(
		kasse.KassensitzungSubject(ks.ZNr), userID, userName,
		ks.ZNr,
		ks.CreatedAt, time.Now().UTC(),
		summen.UmsatzCents, summen.StornierungCents,
		summen.GeldtransitCents,
	)
	if err != nil {
		log.Error("Failed to create tagesabschluss-erstellt event", "error", err, "z_nr", ks.ZNr)
		return err
	}
	return c.writeKassensitzungEvent(ctx, tagesabschlussEvt, ks.ZNr, expectedVersion)
}

// raeumeDruckOutboxAuf runs best effort after the committed Tagesabschluss and returns no error, or a
// committed Abschluss would report failure.
func (c Command) raeumeDruckOutboxAuf(ctx context.Context, zNr int) {
	log := logging.FromContext(ctx)

	verworfen, err := c.DruckauftragRepo.DiscardAlleFehlgeschlagenen(ctx)
	if err != nil {
		log.Error("Failed to discard fehlgeschlagene Druckauftraege beim Tagesabschluss (Abschluss bleibt gueltig)", "error", err, "z_nr", zNr)
		return
	}
	if verworfen > 0 {
		log.Info("Fehlgeschlagene Druckauftraege beim Tagesabschluss verworfen", "z_nr", zNr, "verworfen", verworfen)
	}
}

// findeVorhandenenKassensturz returns the stream's kassensturz-durchgefuehrt event (or nil) and
// whether a non-Abschluss event (kasse.IsAbschlussEventType) follows it. In this stream only
// geldtransit-gebucht:v1 can be one; Tisch bookings live in their own sub-streams.
func (c Command) findeVorhandenenKassensturz(ctx context.Context, subject string) (sturz *kasse.KassensturzDurchgefuehrtV1Data, buchungenNachSturz bool, err error) {
	log := logging.FromContext(ctx)

	events, err := c.KassenjournalRepo.ReadEventsBySubject(ctx, subject)
	if err != nil {
		log.Error("Failed to read events for Kassensturz-Wiederanlauf check", "error", err, "subject", subject)
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
		data, err := event.Decode[kasse.KassensturzDurchgefuehrtV1Data](evt)
		if err != nil {
			log.Error("Unparsebares Kassensturz-Event verhindert Kassenabschluss", "error", err, "event_id", evt.ID)
			return nil, false, err
		}
		sturz = &data
	}
	return sturz, false, nil
}
