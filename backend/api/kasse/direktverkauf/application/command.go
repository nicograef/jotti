package application

import (
	"context"
	"errors"

	bondruckApp "github.com/nicograef/jotti/backend/api/druck/bondruck/application"
	"github.com/nicograef/jotti/backend/api/kasse/enrichment"
	"github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/druckstation"
	"github.com/nicograef/jotti/backend/domain/event"
	"github.com/nicograef/jotti/backend/domain/kasse"
	"github.com/nicograef/jotti/backend/domain/produkt"
	"github.com/nicograef/jotti/backend/repository/druckauftrag_repo"
	"github.com/nicograef/jotti/backend/repository/kassenjournal_repo"
	"github.com/rs/zerolog"
)

type eventRepo interface {
	WriteEventMitVorgang(ctx context.Context, vorgang kassenjournal_repo.Vorgang, e event.Event, streamType kasse.StreamType, kassensitzungNr int) (int, error)
	WriteEventWithDruckauftraegeMitVorgang(ctx context.Context, vorgang kassenjournal_repo.Vorgang, e event.Event, streamType kasse.StreamType, kassensitzungNr int, buildAuftraege func(event.Event) []druckauftrag_repo.NeuerDruckauftrag) (int, error)
	GetMaxVersion(ctx context.Context, subject string) (int, error)
	ReadEventsBySubject(ctx context.Context, subject string) ([]event.Event, error)
	DetermineVorgangStatus(ctx context.Context, vorgangID string, payloadHash []byte) (kassenjournal_repo.VorgangStatus, error)
}

type kassensitzungenRepo interface {
	GetOffeneKassensitzung(ctx context.Context) (*kasse.Kassensitzung, error)
	GetAktiveKassensitzung(ctx context.Context) (*kasse.Kassensitzung, error)
}

type produktRepo interface {
	GetVariantenByIDs(ctx context.Context, ids []int) (map[int]produkt.Variante, error)
	GetProdukteByIDs(ctx context.Context, ids []int) (map[int]produkt.Produkt, error)
}

type druckstationRepo interface {
	GetKonfigurierteDruckstationen(ctx context.Context) (map[string]druckstation.Druckstation, error)
}

type Command struct {
	EventRepo           eventRepo
	ProduktRepo         produktRepo
	KassensitzungenRepo kassensitzungenRepo
	DruckstationRepo    druckstationRepo
}

// getOffeneKassensitzungOderFehler retrieves the open Kassensitzung for a Direktverkauf. It returns
// ErrKasseNichtGeoeffnet when none is active and ErrKasseWirdAbgeschlossen while the Kassensitzung is
// being closed (barrier active), rejecting the Direktverkauf before any TSE roundtrip.
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

// checkVorgang bindet den client-gelieferten Idempotenz-Schlüssel an Art und
// Nutzdaten des Vorgangs und wertet ihn VOR der fachlichen Validierung aus. Die
// Art geht mit in den Hash, weil sich alle Arten einen Schlüsselraum teilen und
// mehrere Kommandos dieselbe Feldmenge einreichen.
//
// Ohne die Bindung entschiede allein der Client, was „derselbe Vorgang" ist. Mit
// ihr hat die Prüfung drei Ausgänge:
//   - VorgangNeu — regulär buchen.
//   - VorgangDuplikat (gleicher Schlüssel, gleiche Nutzdaten) — stille
//     Erfolgsantwort: Ein Wiederholversuch nach erfolgreicher Buchung darf nicht
//     an inzwischen geänderten Invarianten scheitern (z. B. „Position nicht mehr
//     stornierbar"), sondern wiederholt die Erfolgsantwort.
//   - VorgangDatenAbweichend (gleicher Schlüssel, andere Nutzdaten) —
//     ErrVorgangDatenAbweichend: Eine zweite Buchung bucht doppelt, eine stille
//     Erfolgsantwort verschluckt die geänderte Einreichung.
//
// Das Rennen zweier gleichzeitiger Anfragen um denselben Schlüssel fängt
// zusätzlich der Insert der Idempotenz-Zeile in der Schreibtransaktion ab.
func (c Command) checkVorgang(ctx context.Context, vorgangID string, art string, userID int, nutzdaten any) (kassenjournal_repo.Vorgang, kassenjournal_repo.VorgangStatus, error) {
	log := zerolog.Ctx(ctx)

	payloadHash, err := kassenjournal_repo.ComputePayloadHash(art, nutzdaten)
	if err != nil {
		log.Error().Err(err).Str("vorgang_id", vorgangID).Msg("Failed to hash vorgang payload")
		return kassenjournal_repo.Vorgang{}, kassenjournal_repo.VorgangNeu, err
	}

	status, err := c.EventRepo.DetermineVorgangStatus(ctx, vorgangID, payloadHash)
	if err != nil {
		log.Error().Err(err).Str("vorgang_id", vorgangID).Msg("Failed to check vorgang idempotency")
		return kassenjournal_repo.Vorgang{}, kassenjournal_repo.VorgangNeu, ErrDatabase
	}

	vorgang := kassenjournal_repo.Vorgang{VorgangID: vorgangID, Art: art, UserID: userID, PayloadHash: payloadHash}
	return vorgang, status, nil
}

// verkaufPositionNutzdaten ist eine angeforderte Verkaufsposition in der
// Nutzdaten-Sicht: welches Produkt in welcher Variante und Menge. Genau das
// schickt der Client; die serverseitige Anreicherung (Produktname, Preis,
// erzeugte positionId) gehört nicht dazu — sie kann sich zwischen zwei
// Einreichungen ändern und ließe eine echte Wiederholung fälschlich als
// abweichend gelten.
type verkaufPositionNutzdaten struct {
	ProduktID  int `json:"produktId"`
	VarianteID int `json:"varianteId"`
	Menge      int `json:"menge"`
}

func toVerkaufPositionNutzdaten(inputs []enrichment.PositionInput) []verkaufPositionNutzdaten {
	out := make([]verkaufPositionNutzdaten, len(inputs))
	for i, input := range inputs {
		out[i] = verkaufPositionNutzdaten{ProduktID: input.ProduktID, VarianteID: input.VarianteID, Menge: input.Menge}
	}
	return out
}

// verkaufNutzdaten sind die Nutzdaten, an die der Idempotenz-Schlüssel eines
// Direktverkaufs gebunden wird: was in welcher Menge mit welchem Kommentar
// verkauft wird. Einen Tisch gibt es hier nicht — der Verkauf hat seinen eigenen
// Stream.
type verkaufNutzdaten struct {
	Positionen []verkaufPositionNutzdaten `json:"positionen"`
	Kommentar  string                     `json:"kommentar"`
}

// DirektverkaufTaetigen records a Direktverkauf as a single immutable event in its own stream
// (kassensitzung-{nr}/direktverkauf-{uuid}). It requires an open Kassensitzung and writes nothing
// to any projection. Returns ErrKasseNichtGeoeffnet (HTTP 409) when no Kassensitzung is open.
// verkaufID ist eine client-seitig erzeugte UUID (Idempotenz-Schlüssel),
// serverseitig an die Nutzdaten des Verkaufs gebunden: Dieselbe verkaufId mit
// denselben Nutzdaten wird zur stillen Erfolgsantwort, ohne ein zweites Mal zu
// buchen; dieselbe verkaufId mit abweichenden Nutzdaten ergibt
// ErrVorgangDatenAbweichend. Ein echter OCC-Konflikt bleibt davon unberührt und
// ergibt weiterhin ErrConflict.
func (c Command) DirektverkaufTaetigen(ctx context.Context, userID int, userName string, verkaufID string, inputs []enrichment.PositionInput, kommentar string) error {
	log := zerolog.Ctx(ctx)

	nutzdaten := verkaufNutzdaten{Positionen: toVerkaufPositionNutzdaten(inputs), Kommentar: kommentar}
	vorgang, status, err := c.checkVorgang(ctx, verkaufID, kassenjournal_repo.VorgangArtDirektverkauf, userID, nutzdaten)
	if err != nil {
		return err
	}
	if status == kassenjournal_repo.VorgangDuplikat {
		log.Info().Str("verkauf_id", verkaufID).Msg("Idempotenter Direktverkauf: verkaufId bereits gebucht")
		return nil
	}
	if status == kassenjournal_repo.VorgangDatenAbweichend {
		log.Warn().Str("verkauf_id", verkaufID).Msg("Direktverkauf mit abweichenden Nutzdaten unter bekannter verkaufId")
		return ErrVorgangDatenAbweichend
	}

	ks, err := c.getOffeneKassensitzungOderFehler(ctx)
	if err != nil {
		return err
	}

	positionen, err := enrichment.EnrichPositionen(ctx, c.ProduktRepo, inputs)
	if err != nil {
		return err
	}

	subject := kasse.DirektverkaufSubject(ks.ZNr, verkaufID)

	evt, err := kasse.NewDirektverkaufGetaetigtEvent(subject, verkaufID, userID, userName, positionen, kommentar)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create direktverkauf getaetigt event")
		return err
	}

	druckstationen, err := c.konfigurierteDruckstationen(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to load druckstationen for direktverkauf")
		return ErrDatabase
	}

	buildAuftraege := func(stored event.Event) []druckauftrag_repo.NeuerDruckauftrag {
		return bondruckApp.CreateArbeitsbonAuftraegeFromEvent(stored, druckstationen, "")
	}

	// Frischer Stream (client-UUID): erwartete Version 0, das Event ist immer version = 1.
	// Die Druckaufträge entstehen im selben Commit wie das Event, der Idempotenz-Schlüssel
	// des Vorgangs davor. Verliert der Verkauf das Rennen um den Schlüssel, entscheidet der
	// Hash: gleiche Nutzdaten ergeben die stille Erfolgsantwort ohne zweite Buchung,
	// abweichende ErrVorgangDatenAbweichend. Ein Versionskonflikt bleibt davon
	// unterscheidbar — er kommt aus UNIQUE(subject, version) und ergibt ErrConflict.
	err = writeVersionedEvent(ctx, evt, subject, 0, func(versioned event.Event) (int, error) {
		return c.EventRepo.WriteEventWithDruckauftraegeMitVorgang(ctx, vorgang, versioned, kasse.StreamTypeDirektverkauf, ks.ZNr, buildAuftraege)
	})
	if err != nil {
		if errors.Is(err, kassenjournal_repo.ErrVorgangBereitsGebucht) {
			log.Info().Str("verkauf_id", verkaufID).Msg("Idempotenter Direktverkauf: verkaufId bereits gebucht")
			return nil
		}
		if errors.Is(err, kassenjournal_repo.ErrVorgangDatenAbweichend) {
			log.Warn().Str("verkauf_id", verkaufID).Msg("Direktverkauf mit abweichenden Nutzdaten unter bekannter verkaufId")
			return ErrVorgangDatenAbweichend
		}
		if errors.Is(err, ErrConflict) {
			return ErrConflict
		}
		log.Error().Err(err).Msg("Failed to write direktverkauf getaetigt event")
		return ErrDatabase
	}

	log.Info().Str("verkauf_id", verkaufID).Msg("Direktverkauf getaetigt")
	return nil
}

// konfigurierteDruckstationen returns the configured Druckstationen, or an empty
// map when no DruckstationRepo is wired (e.g. in tests). Without configured stations
// the policy derives no print jobs.
func (c Command) konfigurierteDruckstationen(ctx context.Context) (map[string]druckstation.Druckstation, error) {
	if c.DruckstationRepo == nil {
		return nil, nil
	}
	return c.DruckstationRepo.GetKonfigurierteDruckstationen(ctx)
}

// positionNutzdaten ist eine angeforderte Position in der Nutzdaten-Sicht eines
// Vorgangs: welche Position in welcher Menge. Bewusst nicht kasse.PositionRef —
// Domain-Modelle tragen keine json-Tags, und der Hash braucht eine hier
// deklarierte, stabile Feldreihenfolge.
type positionNutzdaten struct {
	PositionID string `json:"positionId"`
	Menge      int    `json:"menge"`
}

func toPositionNutzdaten(refs []kasse.PositionRef) []positionNutzdaten {
	out := make([]positionNutzdaten, len(refs))
	for i, ref := range refs {
		out[i] = positionNutzdaten{PositionID: ref.PositionID, Menge: ref.Menge}
	}
	return out
}

// stornierungNutzdaten sind die Nutzdaten, an die der Idempotenz-Schlüssel einer
// Direktverkauf-Stornierung gebunden wird: welcher Verkauf mit welchen Positionen
// in welcher Menge und mit welchem Grund storniert wird. Genau diese Angaben
// stellt die Servicekraft zusammen; ändert sich eine davon, ist es ein anderer
// Vorgang. Serverseitig Angereichertes (Produktnamen, Preise, Stornobetrag)
// gehört nicht dazu: Es kann sich zwischen zwei Einreichungen ändern und ließe
// eine echte Wiederholung fälschlich als abweichend gelten.
type stornierungNutzdaten struct {
	VerkaufID  string              `json:"verkaufId"`
	Positionen []positionNutzdaten `json:"positionen"`
	Kommentar  string              `json:"kommentar"`
}

// DirektverkaufStornieren records a position-precise cancellation of a Direktverkauf as an immutable
// event appended to that verkauf's own stream (version = maxVersion + 1, OCC). The returned cash
// reduces the Soll-Kassenbestand directly — there is no separate Auszahlung, because a Direktverkauf
// has no open Saldo. Requires an open Kassensitzung (ErrKasseNichtGeoeffnet otherwise). Returns
// ErrVerkaufNichtGefunden when the verkauf does not exist and ErrPositionNichtStornierbar when a
// requested position is not (or no longer) cancellable.
// vorgangID ist eine client-seitig erzeugte UUID (Idempotenz-Schlüssel),
// serverseitig an die Nutzdaten der Stornierung gebunden: Dieselbe vorgangId mit
// denselben Nutzdaten wird zur stillen Erfolgsantwort, ohne ein zweites Mal zu
// buchen; dieselbe vorgangId mit abweichenden Nutzdaten ergibt
// ErrVorgangDatenAbweichend.
func (c Command) DirektverkaufStornieren(ctx context.Context, userID int, userName string, vorgangID string, verkaufID string, positionen []kasse.PositionRef, kommentar string) error {
	log := zerolog.Ctx(ctx)

	nutzdaten := stornierungNutzdaten{VerkaufID: verkaufID, Positionen: toPositionNutzdaten(positionen), Kommentar: kommentar}
	vorgang, status, err := c.checkVorgang(ctx, vorgangID, kassenjournal_repo.VorgangArtDirektverkaufStornierung, userID, nutzdaten)
	if err != nil {
		return err
	}
	if status == kassenjournal_repo.VorgangDuplikat {
		log.Info().Str("vorgang_id", vorgangID).Str("verkauf_id", verkaufID).Msg("Idempotente Direktverkauf-Stornierung: vorgangId bereits gebucht")
		return nil
	}
	if status == kassenjournal_repo.VorgangDatenAbweichend {
		log.Warn().Str("vorgang_id", vorgangID).Str("verkauf_id", verkaufID).Msg("Direktverkauf-Stornierung mit abweichenden Nutzdaten unter bekannter vorgangId")
		return ErrVorgangDatenAbweichend
	}

	ks, err := c.getOffeneKassensitzungOderFehler(ctx)
	if err != nil {
		return err
	}

	subject := kasse.DirektverkaufSubject(ks.ZNr, verkaufID)

	events, err := c.EventRepo.ReadEventsBySubject(ctx, subject)
	if err != nil {
		log.Error().Err(err).Str("verkauf_id", verkaufID).Msg("Failed to read direktverkauf events")
		return ErrDatabase
	}
	if len(events) == 0 {
		return ErrVerkaufNichtGefunden
	}

	nichtStorniert, err := kasse.ComputeNichtStornierteVerkaufPositionen(events)
	if err != nil {
		log.Error().Err(err).Str("verkauf_id", verkaufID).Msg("Failed to compute nicht-stornierte Positionen")
		return ErrDatabase
	}

	if !kasse.ValidatePositionRefs(nichtStorniert, positionen) {
		log.Warn().Str("verkauf_id", verkaufID).Msg("Storno-Invariante verletzt: angeforderte Positionen nicht stornierbar")
		return ErrPositionNichtStornierbar
	}

	resolvedPositionen, gesamtStornierungCents := kasse.ResolvePositionen(nichtStorniert, positionen)

	evt, err := kasse.NewDirektverkaufStorniertEvent(subject, verkaufID, userID, userName, resolvedPositionen, gesamtStornierungCents, kommentar)
	if err != nil {
		log.Error().Err(err).Str("verkauf_id", verkaufID).Msg("Failed to create direktverkauf storniert event")
		return err
	}

	// OCC gegen den validierten Zustand: Basis ist die höchste Version des Replays,
	// gegen den die Storno-Invariante geprüft wurde. Der Idempotenz-Schlüssel des
	// Vorgangs wird samt Nutzdaten-Hash im selben Commit festgehalten (vor dem
	// Event-Insert); verliert der Vorgang das Rennen um den Schlüssel, entscheidet
	// der Hash zwischen stiller Erfolgsantwort und Konflikt.
	err = writeVersionedEvent(ctx, evt, subject, events[len(events)-1].Version, func(versioned event.Event) (int, error) {
		return c.EventRepo.WriteEventMitVorgang(ctx, vorgang, versioned, kasse.StreamTypeDirektverkauf, ks.ZNr)
	})
	if err != nil {
		if errors.Is(err, kassenjournal_repo.ErrVorgangBereitsGebucht) {
			log.Info().Str("vorgang_id", vorgangID).Str("verkauf_id", verkaufID).Msg("Idempotente Direktverkauf-Stornierung: vorgangId bereits gebucht")
			return nil
		}
		if errors.Is(err, kassenjournal_repo.ErrVorgangDatenAbweichend) {
			log.Warn().Str("vorgang_id", vorgangID).Str("verkauf_id", verkaufID).Msg("Direktverkauf-Stornierung mit abweichenden Nutzdaten unter bekannter vorgangId")
			return ErrVorgangDatenAbweichend
		}
		if errors.Is(err, ErrConflict) {
			return ErrConflict
		}
		log.Error().Err(err).Str("verkauf_id", verkaufID).Msg("Failed to write direktverkauf storniert event")
		return ErrDatabase
	}

	log.Info().Str("verkauf_id", verkaufID).Int("gesamt_stornierung_cents", gesamtStornierungCents).Msg("Direktverkauf storniert")
	return nil
}

// writeVersionedEvent writes the event with version expectedVersion+1 via write.
// expectedVersion ist die Version des Zustands, gegen den der Command validiert hat
// (1. Event eines frischen Streams: 0; Storno: höchste Version des Replays). Ein
// UNIQUE(subject, version)-Konflikt — der Stream hat sich seit dem Lesen geändert —
// wird zu ErrConflict.
func writeVersionedEvent(ctx context.Context, e event.Event, subject string, expectedVersion int, write func(event.Event) (int, error)) error {
	e.Version = expectedVersion + 1

	if _, err := write(e); err != nil {
		if errors.Is(err, db.ErrAlreadyExists) {
			// Neutral formuliert, weil hier zwei Quellen zusammenlaufen: der
			// OCC-Versionskonflikt aus UNIQUE(subject, version) — der Regelfall —
			// und, nur für Vorgänge aus der Zeit vor Migration 07, einer der alten
			// partiellen Indexe auf dem Event-JSON (01_initial.up.sql).
			zerolog.Ctx(ctx).Warn().Int("version", e.Version).Str("subject", subject).Msg("Unique violation on event write")
			return ErrConflict
		}
		if errors.Is(err, db.ErrConflict) {
			zerolog.Ctx(ctx).Warn().Str("subject", subject).Msg("Deadlock on event write")
			return ErrConflict
		}
		if errors.Is(err, kassenjournal_repo.ErrKassensitzungNichtOffen) {
			zerolog.Ctx(ctx).Warn().Str("subject", subject).Msg("Kassensitzung nicht mehr offen")
			return ErrKasseNichtGeoeffnet
		}
		return err
	}

	return nil
}
