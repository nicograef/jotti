package application

import (
	"context"
	"errors"
	"time"

	"github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/event"
	"github.com/nicograef/jotti/backend/domain/kasse"
	"github.com/nicograef/jotti/backend/domain/settings"
	"github.com/nicograef/jotti/backend/domain/tse"
	"github.com/nicograef/jotti/backend/repository/kassenjournal_repo"
	"github.com/rs/zerolog"
)

type kassenjournalRepo interface {
	EroeffneKassensitzung(ctx context.Context, datum time.Time, bezeichnung string, build func(zNr int) (event.Event, error)) (int, error)
	WriteEvent(ctx context.Context, e event.Event, streamType kasse.StreamType, kassensitzungNr int) (int, error)
	GetMaxVersion(ctx context.Context, subject string) (int, error)
	GetKassenbestand(ctx context.Context, kassensitzungNr int) (int, error)
	GetTischSessionsByKassensitzungNr(ctx context.Context, kassensitzungNr int) ([]kasse.TischSession, error)
	ReadKassensitzungEvents(ctx context.Context, kassensitzungNr int) ([]event.Event, error)
}

type kassensitzungenRepo interface {
	GetOffeneKassensitzung(ctx context.Context) (*kasse.Kassensitzung, error)
	GetAktiveKassensitzung(ctx context.Context) (*kasse.Kassensitzung, error)
	SetKassensitzungWirdAbgeschlossen(ctx context.Context, zNr int) (int64, error)
	SetKassensitzungOffen(ctx context.Context, zNr int) (int64, error)
}

type settingsRepo interface {
	GetBetreiber(ctx context.Context) (settings.Betreiber, error)
	GetTSEKonfiguration(ctx context.Context) (settings.TSEKonfiguration, error)
}

// tseGateRepo liefert dem Kassenabschluss-Gate die Signatur-Staende der
// Kassensitzung und den aktiven Stoerungszeitraum. Beide fuettern
// tse.BestimmeSignaturstatus — dieselbe Zurechnung wie beim Beleg-Abruf.
type tseGateRepo interface {
	GetOffeneSignaturauftragStaendeFuerKassensitzung(ctx context.Context, kassensitzungNr int) ([]tse.SignaturauftragStand, error)
	GetAktiveTSEStoerung(ctx context.Context) (*tse.Stoerung, error)
}

type Command struct {
	KassenjournalRepo   kassenjournalRepo
	KassensitzungenRepo kassensitzungenRepo
	SettingsRepo        settingsRepo
	TSERepo             tseGateRepo
}

// getOffeneKassensitzungOderFehler returns the open Kassensitzung for a booking. It returns
// ErrKasseNichtGeoeffnet when none is active and ErrKasseWirdAbgeschlossen while the Kassensitzung
// is being closed (barrier active), so bookings are rejected before any TSE roundtrip.
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

// writeKassensitzungEvent writes a Kassensitzung event with OCC against expectedVersion.
// expectedVersion ist die Version des Zustands, gegen den der Command validiert hat
// (frischer Stream: 0). Ein UNIQUE(subject, version)-Konflikt wird zu ErrKonflikt.
func (c Command) writeKassensitzungEvent(ctx context.Context, e event.Event, kassensitzungNr int, expectedVersion int) error {
	log := zerolog.Ctx(ctx)

	subject := kasse.KassensitzungSubject(kassensitzungNr)
	e.Version = expectedVersion + 1

	_, err := c.KassenjournalRepo.WriteEvent(ctx, e, kasse.StreamTypeKassensitzung, kassensitzungNr)
	if err != nil {
		if errors.Is(err, db.ErrAlreadyExists) {
			log.Warn().Int("version", e.Version).Str("subject", subject).Msg("OCC Kassensitzung conflict")
			return ErrKonflikt
		}
		if errors.Is(err, db.ErrConflict) {
			log.Warn().Str("subject", subject).Msg("Deadlock on event write")
			return ErrKonflikt
		}
		if errors.Is(err, kassenjournal_repo.ErrKassensitzungNichtOffen) {
			log.Warn().Str("subject", subject).Msg("Kassensitzung nicht mehr offen")
			return ErrKasseNichtGeoeffnet
		}
		return ErrDatabase
	}

	return nil
}

// betriebstag liefert das Wandkalenderdatum des Zeitpunkts in der übergebenen
// Zeitzone (als Mitternacht UTC, passend zur DATE-Spalte). Ein Truncate(24h)
// würde auf UTC-Mitternacht schneiden und einer nach Mitternacht (00:00–02:00
// Ortszeit) eröffneten Sitzung das Datum des Vortags geben.
func betriebstag(now time.Time, ort *time.Location) time.Time {
	jahr, monat, tag := now.In(ort).Date()
	return time.Date(jahr, monat, tag, 0, 0, 0, 0, time.UTC)
}

// KassensitzungEroeffnen opens a new Kassensitzung. Returns ErrKasseAlreadyOpen if one is already open.
func (c Command) KassensitzungEroeffnen(ctx context.Context, userID int, userName string, bezeichnung string, betragCents int) (int, error) {
	log := zerolog.Ctx(ctx)

	betreiber, err := c.SettingsRepo.GetBetreiber(ctx)
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

	// Ohne konfigurierte TSE wird nicht gesperrt (jotti funktioniert vollständig ohne),
	// aber der unsignierte Betrieb soll im Log nachvollziehbar sein. Der Check ist best
	// effort: Ein Lesefehler verhindert die Eröffnung nicht und unterdrückt nur die Warnung.
	ohneTSE := false
	if conf, err := c.SettingsRepo.GetTSEKonfiguration(ctx); err != nil {
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

	// Entität und Eröffnungs-Event entstehen atomar in einer Transaktion: Schlägt der
	// Event-Write fehl, bleibt keine offene Sitzung ohne Eröffnungs-Event zurück.
	// Der Signaturauftrag (bei Anfangsbestand > 0) entsteht im selben Commit über
	// die fiskalische Projektion; Buchen blockiert nie auf die TSE.
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

// GeldtransitBuchen books a Geldtransit (einlage or entnahme).
func (c Command) GeldtransitBuchen(ctx context.Context, userID int, userName string, richtung string, betragCents int, kommentar string) error {
	log := zerolog.Ctx(ctx)

	ks, err := c.getOffeneKassensitzungOderFehler(ctx)
	if err != nil {
		return err
	}

	evt, err := kasse.NewGeldtransitGebuchtEvent(kasse.KassensitzungSubject(ks.ZNr), userID, userName, richtung, betragCents, kommentar)
	if err != nil {
		log.Error().Err(err).Int("z_nr", ks.ZNr).Msg("Failed to create geldtransit-gebucht event")
		return err
	}

	// Geldtransit validiert keinen Stream-Zustand (reines Anhängen); die Version wird
	// erst unmittelbar vor dem Schreiben bestimmt.
	maxVersion, err := c.KassenjournalRepo.GetMaxVersion(ctx, kasse.KassensitzungSubject(ks.ZNr))
	if err != nil {
		log.Error().Err(err).Int("z_nr", ks.ZNr).Msg("Failed to load max version for geldtransit")
		return ErrDatabase
	}

	if err := c.writeKassensitzungEvent(ctx, evt, ks.ZNr, maxVersion); err != nil {
		return err
	}

	log.Info().Int("z_nr", ks.ZNr).Str("richtung", richtung).Int("betrag_cents", betragCents).Msg("Geldtransit gebucht")
	return nil
}

// KasseAbschliessen schließt die Kasse in einem Schritt ab: Kassensturz,
// Differenzbuchung (bei Differenz ungleich Null) und Tagesabschluss.
//
// Feste Schreibreihenfolge:
//  1. kassensturz-durchgefuehrt:v1 (immer)
//  2. differenz-soll-ist-gebucht:v1 (nur bei Differenz ungleich Null, signiert)
//  3. tagesabschluss-erstellt:v1 (signiert, schließt die Kassensitzung)
//
// Invariante: Tisch-Saldo-Sperre — alle Tisch-Sessions müssen saldo_cents = 0
// haben. Die Tagessummen des Z-Bons kommen aus GetReporting.
//
// Zweiphasig über die Barriere: Als erste Handlung setzt die Sitzung auf
// 'wird_abgeschlossen'. Ab diesem Commit lehnt der Status-Guard alle Buchungs-Events ab;
// erst danach laufen Saldo-Prüfung, Reporting und TSE-Signierungen auf einem eingefrorenen
// Datenstand. Schlägt danach etwas fehl, wird die Sitzung best effort auf 'offen'
// zurückgesetzt; unabhängig davon setzt ein erneuter Aufruf im Zwischenstatus fort.
//
// Teilfehler: Schlägt ein Schreibvorgang nach dem ersten Event fehl, kann der Abschluss
// wiederholt werden. Es gibt bewusst keine umschließende Transaktion über alle Events.
//
// Gate: Als allererste Handlung — noch vor der Barriere — klassifiziert das
// Signatur-Gate jeden noch nicht erledigten Signaturauftrag der Sitzung über die
// Signaturstatus-Funktion. Ein frischer offener Auftrag ohne Störung (Ergebnis
// ausstehend) blockiert mit *SignaturenAusstehendError; Ausfall-Reste lassen den
// Abschluss zu und werden über KassenabschlussErgebnis in der Abschlussmeldung
// ausgewiesen. Die signaturpflichtigen Abschluss-Events entstehen danach und
// laufen regulär über die Queue.
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
	gate, err := c.pruefeSignaturGate(ctx, ks.ZNr)
	if err != nil {
		return KassenabschlussErgebnis{}, err
	}
	if gate.ausstehendAnzahl > 0 {
		log.Warn().Int("z_nr", ks.ZNr).Int("ausstehend", gate.ausstehendAnzahl).Msg("Kassenabschluss blockiert: Signaturen ausstehend")
		return KassenabschlussErgebnis{}, &SignaturenAusstehendError{
			Anzahl:              gate.ausstehendAnzahl,
			AeltesterErstelltAm: gate.aeltesterAusstehend,
		}
	}
	ergebnis = KassenabschlussErgebnis{
		AusfallResteAnzahl:      gate.ausfallResteAnzahl,
		OhneKonfigurationAnzahl: gate.ohneKonfigurationAnzahl,
	}

	// Phase 1: Barriere setzen. Der UPDATE wartet auf noch laufende Buchungen (FOR SHARE);
	// danach lehnt der Status-Guard alle weiteren Buchungs-Events ab. Idempotent, damit ein
	// Wiederholungs-Aufruf im Zwischenstatus fortsetzt.
	rows, err := c.KassensitzungenRepo.SetKassensitzungWirdAbgeschlossen(ctx, ks.ZNr)
	if err != nil {
		log.Error().Err(err).Int("z_nr", ks.ZNr).Msg("Failed to set Kassensitzung status wird_abgeschlossen")
		return KassenabschlussErgebnis{}, ErrDatabase
	}
	if rows == 0 {
		return KassenabschlussErgebnis{}, ErrKasseNichtGeoeffnet
	}

	// Fehler nach dem Statuswechsel setzen die Sitzung best effort zurück auf 'offen', damit sie
	// nicht im Zwischenstatus hängen bleibt und Buchungen wieder möglich werden. Ausnahme: Ein
	// Versionskonflikt bedeutet einen konkurrierenden zweiten Abschluss — dann darf diese Instanz
	// die Barriere nicht unter dem gewinnenden Abschluss wegräumen (SetKassensitzungOffen greift
	// ohnehin nur, solange noch nicht 'abgeschlossen').
	defer func() {
		if err != nil && !errors.Is(err, ErrKonflikt) {
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

	sollBestandCents, err := c.KassenjournalRepo.GetKassenbestand(ctx, ks.ZNr)
	if err != nil {
		log.Error().Err(err).Int("z_nr", ks.ZNr).Msg("Failed to get Kassenbestand for Kassenabschluss")
		return KassenabschlussErgebnis{}, ErrDatabase
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

	// 1. Kassensturz
	kassensturzEvt, err := kasse.NewKassensturzDurchgefuehrtEvent(subject, userID, userName, sollBestandCents, istBestandCents, differenzCents)
	if err != nil {
		log.Error().Err(err).Int("z_nr", ks.ZNr).Msg("Failed to create kassensturz-durchgefuehrt event")
		return KassenabschlussErgebnis{}, err
	}
	if err := c.writeKassensitzungEvent(ctx, kassensturzEvt, ks.ZNr, expectedVersion); err != nil {
		return KassenabschlussErgebnis{}, err
	}
	expectedVersion++

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

	// 3. Tagesabschluss: Kassensitzungs-Events lesen und Summen per Domänenfunktion berechnen.
	// Kassensturz- und Differenz-Events sind zu diesem Zeitpunkt bereits committed (kein
	// In-Memory-Anhängen nötig; Differenzbuchung ist summen-neutral, siehe Resolved decisions).
	sitzungEvents, err := c.KassenjournalRepo.ReadKassensitzungEvents(ctx, ks.ZNr)
	if err != nil {
		log.Error().Err(err).Int("z_nr", ks.ZNr).Msg("Failed to read events for Tagesabschluss")
		return KassenabschlussErgebnis{}, ErrDatabase
	}
	summen := kasse.BerechneAbschlussSummen(sitzungEvents)

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

	log.Info().Int("z_nr", ks.ZNr).
		Int("soll_cents", sollBestandCents).
		Int("ist_cents", istBestandCents).
		Int("differenz_cents", differenzCents).
		Int("ausfall_reste", ergebnis.AusfallResteAnzahl).
		Int("ohne_konfiguration", ergebnis.OhneKonfigurationAnzahl).
		Msg("Kasse abgeschlossen")
	return ergebnis, nil
}
