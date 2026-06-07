//go:build unit

package application

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/event"
	"github.com/nicograef/jotti/backend/domain/kasse"
	"github.com/nicograef/jotti/backend/domain/product"
	"github.com/nicograef/jotti/backend/repository/kassensitzungen_repo"
	"github.com/nicograef/jotti/backend/repository/product_repo"
)

const testKassensitzungNr = 1

var testOpenKS = &kasse.Kassensitzung{
	ZNr:    testKassensitzungNr,
	Status: kasse.KassensitzungOffen,
}

var testProduct = product.Produkt{
	ID:        1,
	Name:      "Cola",
	Kategorie: product.GetraenkKategorie,
	Status:    product.ActiveStatus,
}

var testVariant = product.Variante{
	ID:         1,
	Name:       "Cola 0,5l",
	PreisCents: 350,
	Status:     product.ActiveStatus,
}

// spyEventRepo records WriteEvent calls so tests can assert the exact write contract
// (single event, stream type, version, subject).
type spyEventRepo struct {
	written    []writtenEvent
	maxVersion int
	writeErr   error
}

type writtenEvent struct {
	event           event.Event
	streamType      kasse.StreamType
	kassensitzungNr int
}

func (s *spyEventRepo) GetMaxVersion(_ context.Context, _ string) (int, error) {
	return s.maxVersion, nil
}

func (s *spyEventRepo) WriteEvent(_ context.Context, e event.Event, streamType kasse.StreamType, kassensitzungNr int) (int, error) {
	if s.writeErr != nil {
		return 0, s.writeErr
	}
	s.written = append(s.written, writtenEvent{event: e, streamType: streamType, kassensitzungNr: kassensitzungNr})
	return len(s.written), nil
}

func newProductMock() productRepo {
	productMock := product_repo.NewMock([]product.Produkt{testProduct}, nil)
	productMock.AddVariant(testProduct.ID, testVariant)
	return productMock
}

func newCommand(eventRepo eventRepo, ks *kasse.Kassensitzung) Command {
	return Command{
		EventRepo:           eventRepo,
		ProductRepo:         newProductMock(),
		KassensitzungenRepo: kassensitzungen_repo.NewMock(ks, nil),
	}
}

var testInputs = []VerkaufPositionInput{
	{ProduktID: testProduct.ID, VarianteID: testVariant.ID, Menge: 2},
}

func TestDirektverkaufTaetigen_KasseNichtGeoeffnet(t *testing.T) {
	command := newCommand(&spyEventRepo{}, nil)

	err := command.DirektverkaufTaetigen(context.Background(), 1, "Test User", testInputs, "")
	if err != ErrKasseNichtGeoeffnet {
		t.Fatalf("expected ErrKasseNichtGeoeffnet, got %v", err)
	}
}

func TestDirektverkaufTaetigen_WritesSingleEvent(t *testing.T) {
	spy := &spyEventRepo{}
	command := newCommand(spy, testOpenKS)

	err := command.DirektverkaufTaetigen(context.Background(), 1, "Test User", testInputs, "Direktverkauf")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(spy.written) != 1 {
		t.Fatalf("expected exactly 1 written event, got %d", len(spy.written))
	}

	got := spy.written[0]
	if got.streamType != kasse.StreamTypeDirektverkauf {
		t.Errorf("expected stream type %s, got %s", kasse.StreamTypeDirektverkauf, got.streamType)
	}
	if got.event.Version != 1 {
		t.Errorf("expected version 1, got %d", got.event.Version)
	}
	if got.kassensitzungNr != testKassensitzungNr {
		t.Errorf("expected kassensitzungNr %d, got %d", testKassensitzungNr, got.kassensitzungNr)
	}

	verkaufID, err := kasse.ParseVerkaufIDFromSubject(got.event.Subject)
	if err != nil {
		t.Fatalf("expected a direktverkauf subject, got %q (%v)", got.event.Subject, err)
	}
	if verkaufID == "" {
		t.Error("expected a non-empty verkaufId in the subject")
	}

	var data struct {
		GesamtbetragCents int `json:"gesamtbetragCents"`
	}
	if err := json.Unmarshal(got.event.Data, &data); err != nil {
		t.Fatalf("failed to unmarshal event data: %v", err)
	}
	if data.GesamtbetragCents != testVariant.PreisCents*2 {
		t.Errorf("expected gesamtbetragCents %d, got %d", testVariant.PreisCents*2, data.GesamtbetragCents)
	}
}

func TestDirektverkaufTaetigen_ProduktNotFound(t *testing.T) {
	spy := &spyEventRepo{}
	command := Command{
		EventRepo:           spy,
		ProductRepo:         product_repo.NewMock([]product.Produkt{}, nil),
		KassensitzungenRepo: kassensitzungen_repo.NewMock(testOpenKS, nil),
	}

	err := command.DirektverkaufTaetigen(context.Background(), 1, "Test User", testInputs, "")
	if err != ErrProduktNotFound {
		t.Fatalf("expected ErrProduktNotFound, got %v", err)
	}
	if len(spy.written) != 0 {
		t.Fatalf("expected no event written, got %d", len(spy.written))
	}
}

func TestDirektverkaufTaetigen_Conflict(t *testing.T) {
	spy := &spyEventRepo{writeErr: db.ErrAlreadyExists}
	command := newCommand(spy, testOpenKS)

	err := command.DirektverkaufTaetigen(context.Background(), 1, "Test User", testInputs, "")
	if err != ErrConflict {
		t.Fatalf("expected ErrConflict, got %v", err)
	}
}
