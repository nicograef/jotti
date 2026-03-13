package application

import (
	"context"
	"errors"
	"strconv"

	"github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/event"
	"github.com/nicograef/jotti/backend/domain/product"
	"github.com/nicograef/jotti/backend/domain/table"
	"github.com/rs/zerolog"
)

type tableRepoCommand interface {
	GetTable(ctx context.Context, id int) (table.Tisch, error)
	CreateTable(ctx context.Context, t table.Tisch) (int, error)
	UpdateTable(ctx context.Context, t table.Tisch) error
}

type eventRepoCommand interface {
	WriteEvent(ctx context.Context, event event.Event) (int, error)
	ReadEventsWithSnapshot(ctx context.Context, subject string, snapshotEventType string) ([]event.Event, error)
	GetMaxVersion(ctx context.Context, subject string) (int, error)
}

type productRepoCommand interface {
	GetProduct(ctx context.Context, productID int) (product.Produkt, error)
	GetVariant(ctx context.Context, variantID int) (product.Variante, error)
}

// BestellPositionInput represents the input for a single position in an order.
// The application layer enriches this with product/variant details (fat events).
type BestellPositionInput struct {
	ProduktID  int `json:"produktId"`
	VarianteID int `json:"varianteId"`
	Menge      int `json:"menge"`
}

type Command struct {
	TableRepo   tableRepoCommand
	EventRepo   eventRepoCommand
	ProductRepo productRepoCommand
}

// writeEvent writes an event with optimistic concurrency control.
// It reads the current max version for the subject, sets event.Version = maxVersion + 1,
// and writes the event. Returns ErrConflict on UNIQUE constraint violation (version conflict).
func writeEvent(ctx context.Context, eventRepo eventRepoCommand, e event.Event, subject string) error {
	maxVersion, err := eventRepo.GetMaxVersion(ctx, subject)
	if err != nil {
		return err
	}

	e.Version = maxVersion + 1

	_, err = eventRepo.WriteEvent(ctx, e)
	if err != nil {
		if errors.Is(err, db.ErrAlreadyExists) {
			zerolog.Ctx(ctx).Warn().
				Int("version", e.Version).
				Str("subject", subject).
				Msg("OCC conflict")
			return ErrConflict
		}
		return err
	}

	return nil
}

// loadTischState loads and validates the tisch, then reads its events.
// Returns ErrTischNotFound if the tisch doesn't exist, ErrTischNotActive if not active.
func (c Command) loadTischState(ctx context.Context, tischID int) ([]event.Event, error) {
	log := zerolog.Ctx(ctx)

	tisch, err := c.TableRepo.GetTable(ctx, tischID)
	if err != nil {
		return nil, fromRepositoryError(err, log, tischID)
	}

	if tisch.Status != table.ActiveStatus {
		log.Warn().Int("tisch_id", tischID).Str("status", string(tisch.Status)).Msg("Tisch is not active")
		return nil, ErrTischNotActive
	}

	subject := "tisch:" + strconv.Itoa(tischID)
	events, err := c.EventRepo.ReadEventsWithSnapshot(ctx, subject, string(table.EventTypeSnapshotV1))
	if err != nil {
		log.Error().Err(err).Int("tisch_id", tischID).Msg("Failed to read events for tisch state")
		return nil, ErrDatabase
	}

	return events, nil
}

// validatePositionRefs checks that every requested PositionRef exists in the available positions
// and that the requested Menge does not exceed the available Menge.
func validatePositionRefs(available []table.Position, requested []table.PositionRef) bool {
	for _, ref := range requested {
		found := false
		for _, pos := range available {
			if pos.PositionID == ref.PositionID {
				if ref.Menge > pos.Menge {
					return false
				}
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func (c Command) TischErstellen(ctx context.Context, name string) (int, error) {
	log := zerolog.Ctx(ctx)

	tisch, err := table.NewTisch(name)
	if err != nil {
		log.Warn().Err(err).Str("tisch_name", name).Msg("Invalid tisch data")
		return 0, ErrInvalidTischData
	}

	id, err := c.TableRepo.CreateTable(ctx, tisch)
	if err != nil {
		return 0, fromRepositoryError(err, log, 0)
	}

	log.Info().Int("tisch_id", id).Msg("Tisch created")
	return id, nil
}

func (c Command) TischAktualisieren(ctx context.Context, id int, name string) error {
	log := zerolog.Ctx(ctx)

	tisch, err := c.TableRepo.GetTable(ctx, id)
	if err != nil {
		return fromRepositoryError(err, log, id)
	}

	err = tisch.Rename(name)
	if err != nil {
		log.Warn().Err(err).Int("tisch_id", id).Msg("Invalid tisch data for update")
		return ErrInvalidTischData
	}

	err = c.TableRepo.UpdateTable(ctx, tisch)
	if err != nil {
		return fromRepositoryError(err, log, id)
	}

	log.Info().Int("tisch_id", id).Msg("Tisch updated")
	return nil
}

func (c Command) TischAktivieren(ctx context.Context, id int) error {
	log := zerolog.Ctx(ctx)

	tisch, err := c.TableRepo.GetTable(ctx, id)
	if err != nil {
		return fromRepositoryError(err, log, id)
	}

	tisch.Activate()

	err = c.TableRepo.UpdateTable(ctx, tisch)
	if err != nil {
		return fromRepositoryError(err, log, id)
	}

	log.Info().Int("tisch_id", id).Msg("Tisch activated")
	return nil
}

func (c Command) TischDeaktivieren(ctx context.Context, id int) error {
	log := zerolog.Ctx(ctx)
	tisch, err := c.TableRepo.GetTable(ctx, id)
	if err != nil {
		return fromRepositoryError(err, log, id)
	}

	tisch.Deactivate()

	err = c.TableRepo.UpdateTable(ctx, tisch)
	if err != nil {
		return fromRepositoryError(err, log, id)
	}

	log.Info().Int("tisch_id", id).Msg("Tisch deactivated")
	return nil
}

func (c Command) TischLoeschen(ctx context.Context, id int) error {
	log := zerolog.Ctx(ctx)
	tisch, err := c.TableRepo.GetTable(ctx, id)
	if err != nil {
		return fromRepositoryError(err, log, id)
	}

	tisch.Delete()

	err = c.TableRepo.UpdateTable(ctx, tisch)
	if err != nil {
		return fromRepositoryError(err, log, id)
	}

	log.Info().Int("tisch_id", id).Msg("Tisch deleted")
	return nil
}

func (c Command) BestellungAufgeben(ctx context.Context, userID int, userName string, tischID int, inputs []BestellPositionInput, kommentar string) error {
	log := zerolog.Ctx(ctx)

	// Tisch-Existenz und Status prüfen (Bestellungen brauchen keinen Invarianten-Check)
	if _, err := c.loadTischState(ctx, tischID); err != nil {
		return err
	}

	// Enrich positions with product/variant data (fat events)
	positionen := make([]table.Position, 0, len(inputs))
	for _, input := range inputs {
		variant, err := c.ProductRepo.GetVariant(ctx, input.VarianteID)
		if err != nil {
			log.Error().Err(err).Int("variante_id", input.VarianteID).Msg("Failed to get variant for position enrichment")
			return ErrProduktNotFound
		}

		prod, err := c.ProductRepo.GetProduct(ctx, input.ProduktID)
		if err != nil {
			log.Error().Err(err).Int("produkt_id", input.ProduktID).Msg("Failed to get product for position enrichment")
			return ErrProduktNotFound
		}

		positionen = append(positionen, table.Position{
			VarianteID:   input.VarianteID,
			ProduktName:  prod.Name,
			VarianteName: variant.Name,
			Kategorie:    string(prod.Kategorie),
			Einzelpreis:  variant.PreisCents,
			Menge:        input.Menge,
		})
	}

	event, err := table.NewBestellungAufgegebenEvent(userID, userName, tischID, positionen, kommentar)
	if err != nil {
		log.Error().Err(err).Int("tisch_id", tischID).Msg("Failed to create bestellung aufgegeben event")
		return err
	}

	subject := "tisch:" + strconv.Itoa(tischID)
	if err := writeEvent(ctx, c.EventRepo, event, subject); err != nil {
		if errors.Is(err, ErrConflict) {
			return ErrConflict
		}
		log.Error().Err(err).Int("tisch_id", tischID).Msg("Failed to write bestellung aufgegeben event to database")
		return ErrDatabase
	}

	log.Info().Int("tisch_id", tischID).Msg("Bestellung aufgegeben")
	return nil
}

func (c Command) ZahlungRegistrieren(ctx context.Context, userID int, userName string, tischID int, positionen []table.PositionRef, gesamtZahlungCents int, kommentar string) error {
	log := zerolog.Ctx(ctx)

	// Tisch-Existenz, Status und Events laden
	events, err := c.loadTischState(ctx, tischID)
	if err != nil {
		return err
	}

	// Bezahl-Invariante: nur unbezahlte Positionen können bezahlt werden
	unbezahlt, err := table.GetUnbezahltePositionenFromEvents(events)
	if err != nil {
		log.Error().Err(err).Int("tisch_id", tischID).Msg("Failed to compute unbezahlte positionen")
		return ErrDatabase
	}

	if !validatePositionRefs(unbezahlt, positionen) {
		log.Warn().Int("tisch_id", tischID).Msg("Bezahl-Invariante verletzt: angeforderte Positionen nicht verfügbar")
		return ErrPositionNichtBezahlbar
	}

	event, err := table.NewZahlungRegistriertEvent(userID, userName, tischID, positionen, gesamtZahlungCents, kommentar)
	if err != nil {
		log.Error().Err(err).Int("tisch_id", tischID).Msg("Failed to create zahlung registriert event")
		return err
	}

	subject := "tisch:" + strconv.Itoa(tischID)
	if err := writeEvent(ctx, c.EventRepo, event, subject); err != nil {
		if errors.Is(err, ErrConflict) {
			return ErrConflict
		}
		log.Error().Err(err).Int("tisch_id", tischID).Msg("Failed to write zahlung registriert event to database")
		return ErrDatabase
	}

	log.Info().Int("tisch_id", tischID).Msg("Zahlung registriert")
	return nil
}

func (c Command) ProdukteStornieren(ctx context.Context, userID int, userName string, tischID int, positionen []table.PositionRef, gesamtStornierungCents int, kommentar string) error {
	log := zerolog.Ctx(ctx)

	// Tisch-Existenz, Status und Events laden
	events, err := c.loadTischState(ctx, tischID)
	if err != nil {
		return err
	}

	// Stornierungsinvariante: nur unbezahlte, nicht-stornierte Positionen können storniert werden
	unbezahlt, err := table.GetUnbezahltePositionenFromEvents(events)
	if err != nil {
		log.Error().Err(err).Int("tisch_id", tischID).Msg("Failed to compute unbezahlte positionen for stornierung")
		return ErrDatabase
	}

	if !validatePositionRefs(unbezahlt, positionen) {
		log.Warn().Int("tisch_id", tischID).Msg("Stornierungsinvariante verletzt: angeforderte Positionen nicht stornierbar")
		return ErrPositionNichtStornierbar
	}

	event, err := table.NewProdukteStorniertEvent(userID, userName, tischID, positionen, gesamtStornierungCents, kommentar)
	if err != nil {
		log.Error().Err(err).Int("tisch_id", tischID).Msg("Failed to create produkte storniert event")
		return err
	}

	subject := "tisch:" + strconv.Itoa(tischID)
	if err := writeEvent(ctx, c.EventRepo, event, subject); err != nil {
		if errors.Is(err, ErrConflict) {
			return ErrConflict
		}
		log.Error().Err(err).Int("tisch_id", tischID).Msg("Failed to write produkte storniert event to database")
		return ErrDatabase
	}

	log.Info().Int("tisch_id", tischID).Msg("Produkte storniert")
	return nil
}

func (c Command) ProdukteLiefern(ctx context.Context, userID int, userName string, tischID int, positionen []table.PositionRef, kommentar string) error {
	log := zerolog.Ctx(ctx)

	// Tisch-Existenz, Status und Events laden
	events, err := c.loadTischState(ctx, tischID)
	if err != nil {
		return err
	}

	// Liefer-Invariante: nur ungelieferte Positionen können geliefert werden
	ungeliefert, err := table.GetUngeliefertePositionenFromEvents(events)
	if err != nil {
		log.Error().Err(err).Int("tisch_id", tischID).Msg("Failed to compute ungelieferte positionen")
		return ErrDatabase
	}

	if !validatePositionRefs(ungeliefert, positionen) {
		log.Warn().Int("tisch_id", tischID).Msg("Liefer-Invariante verletzt: angeforderte Positionen nicht lieferbar")
		return ErrPositionNichtLieferbar
	}

	event, err := table.NewProdukteGeliefertEvent(userID, userName, tischID, positionen, kommentar)
	if err != nil {
		log.Error().Err(err).Int("tisch_id", tischID).Msg("Failed to create produkte geliefert event")
		return err
	}

	subject := "tisch:" + strconv.Itoa(tischID)
	if err := writeEvent(ctx, c.EventRepo, event, subject); err != nil {
		if errors.Is(err, ErrConflict) {
			return ErrConflict
		}
		log.Error().Err(err).Int("tisch_id", tischID).Msg("Failed to write produkte geliefert event to database")
		return ErrDatabase
	}

	log.Info().Int("tisch_id", tischID).Msg("Produkte geliefert")
	return nil
}

func (c Command) TischSnapshotErstellen(ctx context.Context, userID int, userName string, tischID int) error {
	log := zerolog.Ctx(ctx)

	subject := "tisch:" + strconv.Itoa(tischID)
	events, err := c.EventRepo.ReadEventsWithSnapshot(ctx, subject, string(table.EventTypeSnapshotV1))
	if err != nil {
		log.Error().Err(err).Int("tisch_id", tischID).Msg("Failed to read events for snapshot")
		return ErrDatabase
	}

	saldo, err := table.GetSaldoFromEvents(events)
	if err != nil {
		return err
	}
	unbezahlt, err := table.GetUnbezahltePositionenFromEvents(events)
	if err != nil {
		return err
	}
	ungeliefert, err := table.GetUngeliefertePositionenFromEvents(events)
	if err != nil {
		return err
	}
	gesamtZahlungen, err := table.GetGesamtZahlungenFromEvents(events)
	if err != nil {
		return err
	}

	log.Debug().
		Int("tisch_id", tischID).
		Int("saldo", saldo).
		Int("unbezahlt_count", len(unbezahlt)).
		Int("ungeliefert_count", len(ungeliefert)).
		Int("gesamt_zahlungen", gesamtZahlungen).
		Msg("Creating snapshot with computed state")

	snapshotEvent, err := table.NewSnapshotEvent(userID, userName, tischID, saldo, unbezahlt, ungeliefert, gesamtZahlungen)
	if err != nil {
		log.Error().Err(err).Int("tisch_id", tischID).Msg("Failed to create snapshot event")
		return err
	}

	if err := writeEvent(ctx, c.EventRepo, snapshotEvent, subject); err != nil {
		if errors.Is(err, ErrConflict) {
			return ErrConflict
		}
		log.Error().Err(err).Int("tisch_id", tischID).Msg("Failed to write snapshot event")
		return ErrDatabase
	}

	log.Info().Int("tisch_id", tischID).Int("saldo", saldo).Msg("Snapshot created")
	return nil
}
