package application

import (
	"context"
	"errors"
	"unicode/utf8"

	"github.com/google/uuid"

	bondruckApp "github.com/nicograef/jotti/backend/api/druck/bondruck/application"
	"github.com/nicograef/jotti/backend/api/kasse/enrichment"
	"github.com/nicograef/jotti/backend/domain/druckstation"
	"github.com/nicograef/jotti/backend/domain/event"
	"github.com/nicograef/jotti/backend/domain/kasse"
	"github.com/nicograef/jotti/backend/domain/produkt"
	"github.com/nicograef/jotti/backend/domain/tisch"
	"github.com/nicograef/jotti/backend/internal/logging"
	"github.com/nicograef/jotti/backend/repository/kassenjournal_repo"
)

type tischRepo interface {
	GetTisch(ctx context.Context, id int) (tisch.Tisch, error)
}

type eventRepo interface {
	WriteEventsAtomic(ctx context.Context, kassensitzungNr int, events []kassenjournal_repo.NeuesEvent) error
	ReadTischSession(ctx context.Context, subject string) (kasse.TischSession, error)
	GetMaxVersion(ctx context.Context, subject string) (int, error)
	ReadEventsBySubject(ctx context.Context, subject string) ([]event.Event, error)
	EventExistsByTypeAndVorgangsID(ctx context.Context, eventType, vorgangsID, jsonKey string) (bool, error)
	GetPfandrueckgabe(ctx context.Context, rueckgabeID string) (event.Event, bool, error)
}

type kassensitzungenRepo interface {
	GetAktiveKassensitzung(ctx context.Context) (*kasse.Kassensitzung, error)
}

type produktRepo interface {
	GetVariantenByIDs(ctx context.Context, ids []int) (map[int]produkt.VarianteMitProdukt, error)
	GetProdukteByIDs(ctx context.Context, ids []int) (map[int]produkt.Produkt, error)
	GetPfandVarianten(ctx context.Context, ids []int) (map[int]bool, error)
}

type druckstationRepo interface {
	GetKonfigurierteDruckstationen(ctx context.Context) (map[string]druckstation.Druckstation, error)
}

type Command struct {
	TischRepo           tischRepo
	EventRepo           eventRepo
	ProduktRepo         produktRepo
	KassensitzungenRepo kassensitzungenRepo
	DruckstationRepo    druckstationRepo
}

func (c Command) getAktivenTisch(ctx context.Context, tischID int) (tisch.Tisch, error) {
	log := logging.FromContext(ctx)

	t, err := c.TischRepo.GetTisch(ctx, tischID)
	if err != nil {
		return tisch.Tisch{}, fromRepositoryError(err, log, tischID)
	}
	if t.Status != tisch.ActiveStatus {
		log.Warn("Tisch is not active", "tisch_id", tischID, "status", string(t.Status))
		return tisch.Tisch{}, ErrTischNotActive
	}
	return t, nil
}

// loadTischState returns subject, kassensitzungNr, tisch name and the projected TischSession.
func (c Command) loadTischState(ctx context.Context, tischID int) (string, int, string, kasse.TischSession, error) {
	log := logging.FromContext(ctx)

	ks, err := sentinels.OffeneKassensitzung(ctx, c.KassensitzungenRepo)
	if err != nil {
		return "", 0, "", kasse.TischSession{}, err
	}

	t, err := c.getAktivenTisch(ctx, tischID)
	if err != nil {
		return "", 0, "", kasse.TischSession{}, err
	}

	subject := kasse.TischSessionSubject(ks.ZNr, tischID)

	state, err := c.EventRepo.ReadTischSession(ctx, subject)
	if err != nil {
		log.Error("Failed to read tisch session", "error", err, "tisch_id", tischID)
		return "", 0, "", kasse.TischSession{}, ErrDatabase
	}

	return subject, ks.ZNr, t.Name, state, nil
}

// BestellungAufnehmen is idempotent via the client's bestellungID: on an OCC conflict a stored event
// with that bestellungId means success (same ID = same operation, payload not compared), none means
// a real conflict (409).
func (c Command) BestellungAufnehmen(ctx context.Context, userID int, userName string, bestellungID string, tischID int, inputs []enrichment.PositionInput, kommentar string) error {
	log := logging.FromContext(ctx)

	subject, kassensitzungNr, tischName, _, err := c.loadTischState(ctx, tischID)
	if err != nil {
		return err
	}

	positionen, err := enrichment.EnrichPositionen(ctx, c.ProduktRepo, inputs)
	if err != nil {
		return err
	}

	evt, err := kasse.NewBestellungAufgenommenEvent(subject, userID, userName, bestellungID, positionen, kommentar)
	if err != nil {
		log.Error("Failed to create bestellung aufgenommen event", "error", err, "tisch_id", tischID)
		return err
	}

	druckstationen, err := c.DruckstationRepo.GetKonfigurierteDruckstationen(ctx)
	if err != nil {
		log.Error("Failed to load druckstationen for arbeitsbon", "error", err, "tisch_id", tischID)
		return ErrDatabase
	}

	buildAuftraege := func(stored event.Event) []druckstation.NeuerDruckauftrag {
		return bondruckApp.CreateArbeitsbonAuftraegeFromEvent(stored, druckstationen, tischName)
	}

	// Bestellungen validieren keinen Stream-Zustand (reines Anhängen); die Version wird deshalb erst
	// unmittelbar vor dem Schreiben bestimmt. Bumpt sie dabei die Version eines parallel
	// validierenden Commands (Zahlung, Storno), läuft dieser korrekt in den OCC-Konflikt.
	maxVersion, err := c.EventRepo.GetMaxVersion(ctx, subject)
	if err != nil {
		log.Error("Failed to load max version for bestellung", "error", err, "tisch_id", tischID)
		return ErrDatabase
	}
	evt.Version = maxVersion + 1

	// Writes the event and the Druckaufträge built from it in one transaction
	// (transactional outbox); buildAuftraege sees the stored event with its generated ID.
	err = c.EventRepo.WriteEventsAtomic(ctx, kassensitzungNr, []kassenjournal_repo.NeuesEvent{
		{Event: evt, StreamType: kasse.StreamTypeTischSession, BuildAuftraege: buildAuftraege},
	})
	if err != nil {
		err = sentinels.Schreibfehler(ctx, err, "subject", subject, "version", evt.Version)
		if errors.Is(err, ErrConflict) {
			exists, lookupErr := c.EventRepo.EventExistsByTypeAndVorgangsID(ctx, string(kasse.EventTypeBestellungAufgenommenV1), bestellungID, "bestellungId")
			if lookupErr != nil {
				log.Error("Failed to lookup bestellung idempotency", "error", lookupErr, "bestellung_id", bestellungID)
				return ErrDatabase
			}
			if exists {
				log.Info("Idempotente Bestellung: bestellungId bereits vorhanden", "bestellung_id", bestellungID)
				return nil
			}
			return ErrConflict
		}
		return err
	}

	log.Info("Bestellung aufgenommen", "tisch_id", tischID)
	return nil
}

// maxUmbuchungKommentarBytes ist die Grenze des Kommentars im eingefrorenen
// Event-Schema (bestellungUmgebuchtV1DataSchema); zogs Max zählt Bytes.
const maxUmbuchungKommentarBytes = 100

// truncateBytes kürzt auf höchstens maxBytes und rückt bis zum Anfang der angeschnittenen UTF-8-Folge zurück, damit kein Umlaut zerfällt.
func truncateBytes(s string, maxBytes int) string {
	if len(s) <= maxBytes {
		return s
	}

	cut := maxBytes
	for cut > 0 && !utf8.RuneStart(s[cut]) {
		cut--
	}

	return s[:cut]
}

// buildUmbuchungKommentar: Tischname (bis 100 Bytes) plus Präfix reißt die Schemagrenze — gekürzt
// wird der Name, das Präfix bleibt ganz.
func buildUmbuchungKommentar(prefix string, tischName string) string {
	return prefix + truncateBytes(tischName, maxUmbuchungKommentarBytes-len(prefix))
}

// BestellungUmbuchen is idempotent via the client's umbuchungID: the source stream is searched before
// validating and again after a rejected write, see pruefeGebuchteUmbuchung. The BON_ID index admits
// the Abgang's umbuchungId once across all Vorgänge, so a reused key on another table is a conflict.
func (c Command) BestellungUmbuchen(ctx context.Context, userID int, userName string, umbuchungID string, quellTischID int, zielTischID int, positionen []kasse.PositionRef, benutzerKommentar string) error {
	log := logging.FromContext(ctx)

	if quellTischID == zielTischID {
		log.Warn("Umbuchung mit gleichem Quell- und Ziel-Tisch", "tisch_id", quellTischID)
		return ErrUmbuchungGleicherTisch
	}

	ks, err := sentinels.OffeneKassensitzung(ctx, c.KassensitzungenRepo)
	if err != nil {
		return err
	}

	quellTisch, err := c.getAktivenTisch(ctx, quellTischID)
	if err != nil {
		return err
	}
	zielTisch, err := c.getAktivenTisch(ctx, zielTischID)
	if err != nil {
		return err
	}

	quellSubject := kasse.TischSessionSubject(ks.ZNr, quellTischID)
	zielSubject := kasse.TischSessionSubject(ks.ZNr, zielTischID)

	// The state is read before the check: a commit the check misses then lies after the OCC base,
	// so the write conflicts and the re-check below answers it.
	quellState, err := c.EventRepo.ReadTischSession(ctx, quellSubject)
	if err != nil {
		log.Error("Failed to read source tisch session", "error", err, "tisch_id", quellTischID)
		return ErrDatabase
	}

	if gebucht, err := c.pruefeGebuchteUmbuchung(ctx, quellSubject, umbuchungID, quellTischID, zielTischID, positionen); gebucht || err != nil {
		return err
	}

	resolvedPositionen, gesamtCents, ok := kasse.ResolvePositionen(quellState.UnbezahltePositionen, positionen)
	if !ok {
		log.Warn("Umbuchungsinvariante verletzt: Positionen nicht umbuchbar", "quell_tisch_id", quellTischID)
		return ErrPositionNichtUmbuchbar
	}

	quellKommentar := buildUmbuchungKommentar("Umbuchung auf ", zielTisch.Name)
	zielKommentar := buildUmbuchungKommentar("Umbuchung von ", quellTisch.Name)

	quellEvent, zielEvent, err := kasse.NewBestellungUmgebuchtEvents(ks.ZNr, umbuchungID, quellTischID, zielTischID, userID, userName, resolvedPositionen, gesamtCents, quellKommentar, zielKommentar, benutzerKommentar)
	if err != nil {
		log.Error("Failed to create umbuchung events", "error", err, "quell_tisch_id", quellTischID, "ziel_tisch_id", zielTischID)
		return err
	}

	// The source uses OCC against its validated state; the target validates nothing (pure append), so
	// its version is read right before writing. Both sides get their Signaturauftrag in the same commit.
	zielMaxVersion, err := c.EventRepo.GetMaxVersion(ctx, zielSubject)
	if err != nil {
		log.Error("Failed to load max version for target subject", "error", err, "ziel_tisch_id", zielTischID)
		return ErrDatabase
	}

	quellEvent.Version = quellState.LastEventVersion + 1
	zielEvent.Version = zielMaxVersion + 1

	err = c.EventRepo.WriteEventsAtomic(ctx, ks.ZNr, []kassenjournal_repo.NeuesEvent{
		{Event: quellEvent, StreamType: kasse.StreamTypeTischSession},
		{Event: zielEvent, StreamType: kasse.StreamTypeTischSession},
	})
	if err != nil {
		err = sentinels.Schreibfehler(ctx, err, "quell_subject", quellSubject, "quell_version", quellEvent.Version, "ziel_subject", zielSubject, "ziel_version", zielEvent.Version)
		if errors.Is(err, ErrConflict) {
			if gebucht, err := c.pruefeGebuchteUmbuchung(ctx, quellSubject, umbuchungID, quellTischID, zielTischID, positionen); gebucht || err != nil {
				return err
			}
		}
		return err
	}

	log.Info("Bestellung umgebucht", "quell_tisch_id", quellTischID, "ziel_tisch_id", zielTischID)
	return nil
}

// pruefeGebuchteUmbuchung reports whether umbuchungID is booked in the source stream. A booked one
// answers success only if it moves the same quantities between the same tables.
func (c Command) pruefeGebuchteUmbuchung(ctx context.Context, quellSubject string, umbuchungID string, quellTischID, zielTischID int, positionen []kasse.PositionRef) (bool, error) {
	log := logging.FromContext(ctx)

	events, err := c.EventRepo.ReadEventsBySubject(ctx, quellSubject)
	if err != nil {
		log.Error("Failed to read source events for umbuchung", "error", err, "quell_tisch_id", quellTischID)
		return false, ErrDatabase
	}
	abgang, found, err := kasse.GebuchteUmbuchung(events, umbuchungID)
	if err != nil {
		log.Error("Failed to search booked umbuchung", "error", err, "umbuchung_id", umbuchungID)
		return false, ErrDatabase
	}
	if !found {
		return false, nil
	}
	gleich, err := kasse.GleicheUmbuchung(abgang, quellTischID, zielTischID, positionen)
	if err != nil {
		log.Error("Failed to compare booked umbuchung", "error", err, "umbuchung_id", umbuchungID)
		return true, ErrDatabase
	}
	if !gleich {
		log.Warn("Umbuchung mit anderem Inhalt bereits gebucht", "umbuchung_id", umbuchungID)
		return true, ErrVorgangAbweichend
	}
	log.Info("Idempotente Umbuchung: umbuchungId bereits vorhanden", "umbuchung_id", umbuchungID)
	return true, nil
}

// ZahlungKassieren writes the payment and, with pfand, the settled Rückgabe in one transaction. The
// client's zahlungID makes it idempotent: a retry finds its positions paid or hits the BON_ID index,
// and pruefeZahlungsWiederholung answers it. The rueckgabeId of a payment with Pfand is checked first,
// see pruefeVerrechneteZahlung.
func (c Command) ZahlungKassieren(ctx context.Context, userID int, userName string, zahlungID string, tischID int, positionen []kasse.PositionRef, kommentar string, pfand *enrichment.PfandEingabe) error {
	log := logging.FromContext(ctx)

	if pfand != nil {
		if err := c.pruefeGebuchteRueckgabe(ctx, tischID, positionen, pfand, nil); err != nil {
			return err
		}
	}

	subject, kassensitzungNr, _, state, err := c.loadTischState(ctx, tischID)
	if err != nil {
		return err
	}

	resolvedPositionen, gesamtZahlungCents, ok := kasse.ResolvePositionen(state.UnbezahltePositionen, positionen)
	if !ok {
		log.Warn("Bezahl-Invariante verletzt: angeforderte Positionen nicht verfügbar", "tisch_id", tischID)
		return c.pruefeZahlungsWiederholung(ctx, subject, zahlungID, tischID, positionen, pfand, ErrPositionNichtBezahlbar)
	}
	resolvedPositionen, err = c.mitPfandKennzeichen(ctx, resolvedPositionen, tischID)
	if err != nil {
		return err
	}

	evt, err := kasse.NewZahlungKassiertEvent(subject, zahlungID, userID, userName, resolvedPositionen, gesamtZahlungCents, kommentar)
	if err != nil {
		log.Error("Failed to create zahlung kassiert event", "error", err, "tisch_id", tischID)
		return err
	}
	// OCC against the validated projection: a stream change since reading hits UNIQUE(subject, version).
	evt.Version = state.LastEventVersion + 1

	events := []kassenjournal_repo.NeuesEvent{{Event: evt, StreamType: kasse.StreamTypeTischSession}}

	if pfand != nil {
		rueckgabe, err := enrichment.VerrechneteRueckgabe(ctx, c.ProduktRepo, kassensitzungNr, *pfand, tischID, evt)
		if err != nil {
			return err
		}
		events = append(events, kassenjournal_repo.NeuesEvent{Event: rueckgabe, StreamType: kasse.StreamTypePfandrueckgabe})
	}

	if err := c.EventRepo.WriteEventsAtomic(ctx, kassensitzungNr, events); err != nil {
		err = sentinels.Schreibfehler(ctx, err, "subject", subject, "tisch_id", tischID)
		if errors.Is(err, ErrConflict) {
			return c.pruefeZahlungsWiederholung(ctx, subject, zahlungID, tischID, positionen, pfand, ErrConflict)
		}
		return err
	}

	log.Info("Zahlung kassiert", "tisch_id", tischID, "mit_pfand", pfand != nil)
	return nil
}

// pruefeZahlungsWiederholung answers a payment whose positions are paid or whose write was rejected.
// A booked zahlungID answers success only if it pays the same quantities; with pfand its Rückgabe
// must match too. Without one, a parallel attempt of the same rueckgabeId may have booked since the
// first check; otherwise fallback stands.
func (c Command) pruefeZahlungsWiederholung(ctx context.Context, subject string, zahlungID string, tischID int, positionen []kasse.PositionRef, pfand *enrichment.PfandEingabe, fallback error) error {
	log := logging.FromContext(ctx)

	events, err := c.EventRepo.ReadEventsBySubject(ctx, subject)
	if err != nil {
		log.Error("Failed to read events for zahlung", "error", err, "tisch_id", tischID)
		return ErrDatabase
	}
	zahlung, found, err := kasse.GebuchteZahlung(events, zahlungID)
	if err != nil {
		log.Error("Failed to search booked zahlung", "error", err, "zahlung_id", zahlungID)
		return ErrDatabase
	}
	if !found {
		if pfand != nil {
			return c.pruefeGebuchteRueckgabe(ctx, tischID, positionen, pfand, fallback)
		}
		return fallback
	}
	if pfand != nil {
		return c.pruefeGebuchteRueckgabe(ctx, tischID, positionen, pfand, ErrVorgangAbweichend)
	}
	gleich, err := kasse.GleicheZahlung(zahlung, positionen)
	if err != nil {
		log.Error("Failed to compare booked zahlung", "error", err, "zahlung_id", zahlungID)
		return ErrDatabase
	}
	if !gleich {
		log.Warn("Zahlung mit anderem Inhalt bereits gebucht", "zahlung_id", zahlungID, "tisch_id", tischID)
		return ErrVorgangAbweichend
	}
	log.Info("Idempotente Zahlung: zahlungId bereits vorhanden", "zahlung_id", zahlungID)
	return nil
}

// mitPfandKennzeichen: the Zahlung is the Beleg, so it freezes the Pfand flag the product carries now.
func (c Command) mitPfandKennzeichen(ctx context.Context, positionen []kasse.Position, tischID int) ([]kasse.Position, error) {
	varianteIDs := make([]int, len(positionen))
	for i, p := range positionen {
		varianteIDs[i] = p.VarianteID
	}
	pfandVarianten, err := c.ProduktRepo.GetPfandVarianten(ctx, varianteIDs)
	if err != nil {
		logging.FromContext(ctx).Error("Failed to read the pfand flag of the paid positions", "error", err, "tisch_id", tischID)
		return nil, ErrDatabase
	}
	return kasse.MitPfandKennzeichen(positionen, pfandVarianten), nil
}

// pruefeGebuchteRueckgabe answers with pruefeVerrechneteZahlung when pfand.RueckgabeID is already
// booked, else with fallback.
func (c Command) pruefeGebuchteRueckgabe(ctx context.Context, tischID int, positionen []kasse.PositionRef, pfand *enrichment.PfandEingabe, fallback error) error {
	rueckgabe, found, err := c.getPfandrueckgabe(ctx, pfand.RueckgabeID)
	if err != nil {
		return err
	}
	if found {
		return c.pruefeVerrechneteZahlung(ctx, rueckgabe, tischID, positionen, pfand)
	}
	return fallback
}

func (c Command) getPfandrueckgabe(ctx context.Context, rueckgabeID string) (event.Event, bool, error) {
	rueckgabe, found, err := c.EventRepo.GetPfandrueckgabe(ctx, rueckgabeID)
	if err != nil {
		logging.FromContext(ctx).Error("Failed to read pfandrueckgabe", "error", err, "rueckgabe_id", rueckgabeID)
		return event.Event{}, false, ErrDatabase
	}
	return rueckgabe, found, nil
}

// pruefeVerrechneteZahlung answers a payment whose rueckgabeId is already booked: success only if
// the Rückgabe books this table and these quantities and the payment written with it pays the
// requested positions. A Rückgabe without that payment, say a pure return, is a different Vorgang.
func (c Command) pruefeVerrechneteZahlung(ctx context.Context, rueckgabe event.Event, tischID int, positionen []kasse.PositionRef, pfand *enrichment.PfandEingabe) error {
	log := logging.FromContext(ctx)

	gleich, err := kasse.GleicheRueckgabe(rueckgabe, tischID, enrichment.MengenJeVariante(pfand.Positionen))
	if err != nil {
		log.Error("Failed to compare stored pfandrueckgabe", "error", err, "rueckgabe_id", pfand.RueckgabeID)
		return ErrDatabase
	}
	if gleich {
		gleich, err = c.gleicheVerrechneteZahlung(ctx, rueckgabe, tischID, positionen)
		if err != nil {
			log.Error("Failed to compare stored zahlung", "error", err, "rueckgabe_id", pfand.RueckgabeID)
			return ErrDatabase
		}
	}
	if !gleich {
		log.Warn("Verrechnete Zahlung weicht von der gebuchten Pfandrückgabe ab", "tisch_id", tischID, "rueckgabe_id", pfand.RueckgabeID)
		return ErrPfandrueckgabeAbweichend
	}

	log.Info("Idempotente Zahlung: rueckgabeId bereits vorhanden", "rueckgabe_id", pfand.RueckgabeID)
	return nil
}

func (c Command) gleicheVerrechneteZahlung(ctx context.Context, rueckgabe event.Event, tischID int, positionen []kasse.PositionRef) (bool, error) {
	zNr, err := kasse.ParseZNrFromSubject(rueckgabe.Subject)
	if err != nil {
		return false, err
	}
	events, err := c.EventRepo.ReadEventsBySubject(ctx, kasse.TischSessionSubject(zNr, tischID))
	if err != nil {
		return false, err
	}
	zahlung, found := kasse.VerrechnungsPartner(rueckgabe, events, kasse.EventTypeZahlungKassiertV1, kasse.EventTypeZahlungKassiertV2)
	if !found {
		return false, nil
	}
	return kasse.GleicheZahlung(zahlung, positionen)
}

// StornierungErteilen splits by payment status: unpaid quantities become one bestellung-korrigiert,
// paid ones one stornierung-erteilt per Zahlung (FIFO). Each event gets its own TSE transaction, all
// are written atomically (docs/handbuch.md §3.7). The client's stornoID makes it idempotent: every
// event ID derives from it, see pruefeGebuchteStornierung.
func (c Command) StornierungErteilen(ctx context.Context, userID int, userName string, stornoID uuid.UUID, tischID int, positionen []kasse.PositionRef, kommentar string) error {
	log := logging.FromContext(ctx)

	subject, kassensitzungNr, _, _, err := c.loadTischState(ctx, tischID)
	if err != nil {
		return err
	}

	events, err := c.EventRepo.ReadEventsBySubject(ctx, subject)
	if err != nil {
		log.Error("Failed to read events for stornierung", "error", err, "tisch_id", tischID)
		return ErrDatabase
	}

	if gebucht, err := pruefeGebuchteStornierung(ctx, events, stornoID, positionen); gebucht || err != nil {
		return err
	}

	// false = angeforderte Positionen sind nicht stornierbar: die Menge übersteigt die noch
	// stornierbare Menge, oder eine PositionID ist doppelt referenziert.
	aufteilung, ok := kasse.ComputeStornoAufteilung(events, positionen)
	if !ok {
		log.Warn("Stornierungsinvariante verletzt: angeforderte Positionen nicht stornierbar", "tisch_id", tischID)
		return ErrPositionNichtStornierbar
	}

	stornoEvents, err := buildStornoEvents(ctx, subject, stornoID, userID, userName, aufteilung, kommentar)
	if err != nil {
		return err
	}

	expectedVersion := 0
	if len(events) > 0 {
		expectedVersion = events[len(events)-1].Version
	}

	err = c.persistStornoEvents(ctx, stornoEvents, subject, expectedVersion, kassensitzungNr, tischID)
	if errors.Is(err, ErrConflict) {
		events, readErr := c.EventRepo.ReadEventsBySubject(ctx, subject)
		if readErr != nil {
			log.Error("Failed to read events after stornierung conflict", "error", readErr, "tisch_id", tischID)
			return ErrDatabase
		}
		if gebucht, err := pruefeGebuchteStornierung(ctx, events, stornoID, positionen); gebucht || err != nil {
			return err
		}
	}
	return err
}

// pruefeGebuchteStornierung reports whether stornoID is booked in the stream. A booked one answers
// success only if its events cancel the same quantities.
func pruefeGebuchteStornierung(ctx context.Context, events []event.Event, stornoID uuid.UUID, positionen []kasse.PositionRef) (bool, error) {
	log := logging.FromContext(ctx)

	gebucht, err := kasse.GebuchteTischStornierung(events, stornoID)
	if err != nil {
		log.Error("Failed to search booked stornierung", "error", err, "storno_id", stornoID)
		return false, ErrDatabase
	}
	if len(gebucht) == 0 {
		return false, nil
	}
	gleich, err := kasse.GleicheTischStornierung(gebucht, positionen)
	if err != nil {
		log.Error("Failed to compare booked stornierung", "error", err, "storno_id", stornoID)
		return true, ErrDatabase
	}
	if !gleich {
		log.Warn("Stornierung mit anderem Inhalt bereits gebucht", "storno_id", stornoID)
		return true, ErrVorgangAbweichend
	}
	log.Info("Idempotente Stornierung: stornoId bereits vorhanden", "storno_id", stornoID)
	return true, nil
}

// buildStornoEvents numbers the events in write order; the n-th gets kasse.AbgeleiteteID(stornoID, n).
func buildStornoEvents(ctx context.Context, subject string, stornoID uuid.UUID, userID int, userName string, aufteilung kasse.StornoAufteilung, kommentar string) ([]event.Event, error) {
	log := logging.FromContext(ctx)

	var events []event.Event

	if len(aufteilung.Korrektur) > 0 {
		evt, err := kasse.NewBestellungKorrigiertEvent(subject, kasse.AbgeleiteteID(stornoID, len(events)), userID, userName, aufteilung.Korrektur, aufteilung.KorrekturCents, kommentar)
		if err != nil {
			log.Error("Failed to create bestellung korrigiert event", "error", err, "subject", subject)
			return nil, err
		}
		events = append(events, evt)
	}

	for _, wr := range aufteilung.Warenruecknahmen {
		evt, err := kasse.NewStornierungErteiltEvent(subject, kasse.AbgeleiteteID(stornoID, len(events)), userID, userName, wr.ZahlungID, wr.Positionen, wr.GesamtCents, kommentar)
		if err != nil {
			log.Error("Failed to create stornierung erteilt event", "error", err, "subject", subject)
			return nil, err
		}
		events = append(events, evt)
	}

	return events, nil
}

func (c Command) persistStornoEvents(ctx context.Context, stornoEvents []event.Event, subject string, expectedVersion int, kassensitzungNr int, tischID int) error {
	log := logging.FromContext(ctx)

	neueEvents := make([]kassenjournal_repo.NeuesEvent, 0, len(stornoEvents))
	for i, evt := range stornoEvents {
		evt.Version = expectedVersion + 1 + i
		neueEvents = append(neueEvents, kassenjournal_repo.NeuesEvent{Event: evt, StreamType: kasse.StreamTypeTischSession})
	}

	if err := c.EventRepo.WriteEventsAtomic(ctx, kassensitzungNr, neueEvents); err != nil {
		return sentinels.Schreibfehler(ctx, err, "subject", subject)
	}

	log.Info("Stornierung erteilt", "tisch_id", tischID, "anzahl_events", len(neueEvents))
	return nil
}
