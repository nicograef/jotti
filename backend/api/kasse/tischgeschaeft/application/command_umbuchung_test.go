package application

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/kasse"
	"github.com/nicograef/jotti/backend/domain/tisch"
	"github.com/nicograef/jotti/backend/repository/repotest"
)

type umbuchungTischRepoMock struct {
	tables map[int]tisch.Tisch
}

func (m *umbuchungTischRepoMock) GetTisch(_ context.Context, id int) (tisch.Tisch, error) {
	entry, ok := m.tables[id]
	if !ok {
		return tisch.Tisch{}, db.ErrNotFound
	}
	return entry, nil
}

func (m *umbuchungTischRepoMock) CreateTisch(_ context.Context, _ tisch.Tisch) (int, error) {
	return 0, nil
}

func (m *umbuchungTischRepoMock) UpdateTisch(_ context.Context, _ tisch.Tisch) error {
	return nil
}

func (m *umbuchungTischRepoMock) GetAlleTische(_ context.Context) ([]tisch.Tisch, error) {
	return nil, nil
}

func (m *umbuchungTischRepoMock) GetAktiveTischeMitFavoriten(_ context.Context, _, _ int) ([]tisch.AktiverTischMitFavorit, error) {
	return nil, nil
}

func TestBestellungUmbuchen_HappyPath(t *testing.T) {
	ctx := context.Background()
	quellTisch := tisch.Tisch{ID: 1, Name: "Tisch Quelle", Status: tisch.ActiveStatus}
	zielTisch := tisch.Tisch{ID: 2, Name: "Tisch Ziel", Status: tisch.ActiveStatus}

	eventMock := repotest.NewKassenjournalRepo(nil, nil)
	quellSubject := kasse.TischSessionSubject(testKassensitzungNr, quellTisch.ID)
	zielSubject := kasse.TischSessionSubject(testKassensitzungNr, zielTisch.ID)
	quellPositionID := uuid.New().String()

	eventMock.SetTischSession(quellSubject, kasse.TischSession{
		SaldoCents: 700,
		UnbezahltePositionen: []kasse.Position{
			{
				PositionID:   quellPositionID,
				VarianteID:   1,
				ProduktName:  "Cola",
				VarianteName: "0,5l",
				Kategorie:    "getraenk", Steuersatz: "regel",
				EinzelpreisCents: 350,
				Menge:            2,
			},
		},
	})

	command := Command{
		TischRepo:           repotest.NewTischRepo([]tisch.Tisch{quellTisch, zielTisch}, nil),
		EventRepo:           eventMock,
		KassensitzungenRepo: repotest.NewKassensitzungenRepo(testOpenKS, nil),
	}

	err := command.BestellungUmbuchen(ctx, 1, "Test User", uuid.NewString(), quellTisch.ID, zielTisch.ID, []kasse.PositionRef{{PositionID: quellPositionID, Menge: 1}}, "Gast gewechselt")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	quellEvents, err := eventMock.ReadEventsBySubject(ctx, quellSubject)
	if err != nil {
		t.Fatalf("expected no error reading source events, got %v", err)
	}
	if len(quellEvents) != 1 {
		t.Fatalf("expected 1 source event, got %d", len(quellEvents))
	}
	if quellEvents[0].Type != string(kasse.EventTypeBestellungUmgebuchtV1) {
		t.Errorf("expected source event type %s, got %s", kasse.EventTypeBestellungUmgebuchtV1, quellEvents[0].Type)
	}

	zielEvents, err := eventMock.ReadEventsBySubject(ctx, zielSubject)
	if err != nil {
		t.Fatalf("expected no error reading target events, got %v", err)
	}
	if len(zielEvents) != 1 {
		t.Fatalf("expected 1 target event, got %d", len(zielEvents))
	}
	if zielEvents[0].Type != string(kasse.EventTypeBestellungUmgebuchtV1) {
		t.Errorf("expected target event type %s, got %s", kasse.EventTypeBestellungUmgebuchtV1, zielEvents[0].Type)
	}

	var quellData kasse.BestellungUmgebuchtV1Data
	if err := json.Unmarshal(quellEvents[0].Data, &quellData); err != nil {
		t.Fatalf("expected no unmarshal error for source umbuchung data, got %v", err)
	}

	if quellData.GesamtCents != 350 {
		t.Errorf("expected source amount 350, got %d", quellData.GesamtCents)
	}
	if quellData.QuellTischID != quellTisch.ID || quellData.ZielTischID != zielTisch.ID {
		t.Errorf("unexpected source tisch refs: quell=%d ziel=%d", quellData.QuellTischID, quellData.ZielTischID)
	}
	if quellData.Kommentar != "Umbuchung auf Tisch Ziel" {
		t.Errorf("unexpected source comment: %q", quellData.Kommentar)
	}
	if quellData.BenutzerKommentar != "Gast gewechselt" {
		t.Errorf("unexpected source benutzerKommentar: %q", quellData.BenutzerKommentar)
	}
	if len(quellData.Positionen) != 1 {
		t.Fatalf("expected 1 source position, got %d", len(quellData.Positionen))
	}
	if quellData.Positionen[0].PositionID != quellPositionID {
		t.Errorf("expected source position ID %q, got %q", quellPositionID, quellData.Positionen[0].PositionID)
	}
	if quellData.Positionen[0].EinzelpreisCents != 350 {
		t.Errorf("expected source einzelpreis 350, got %d", quellData.Positionen[0].EinzelpreisCents)
	}

	var zielData kasse.BestellungUmgebuchtV1Data
	if err := json.Unmarshal(zielEvents[0].Data, &zielData); err != nil {
		t.Fatalf("expected no unmarshal error for target umbuchung data, got %v", err)
	}

	if zielData.GesamtCents != 350 {
		t.Errorf("expected target amount 350, got %d", zielData.GesamtCents)
	}
	// Beide Seiten teilen sich dieselbe UmbuchungID.
	if zielData.UmbuchungID != quellData.UmbuchungID {
		t.Errorf("expected shared UmbuchungID, got quell=%q ziel=%q", quellData.UmbuchungID, zielData.UmbuchungID)
	}
	if zielData.Kommentar != "Umbuchung von Tisch Quelle" {
		t.Errorf("unexpected target comment: %q", zielData.Kommentar)
	}
	if zielData.BenutzerKommentar != "Gast gewechselt" {
		t.Errorf("unexpected target benutzerKommentar: %q", zielData.BenutzerKommentar)
	}
	if len(zielData.Positionen) != 1 {
		t.Fatalf("expected 1 target position, got %d", len(zielData.Positionen))
	}
	if zielData.Positionen[0].PositionID == quellPositionID {
		t.Errorf("expected target position ID to be regenerated, but remained %q", zielData.Positionen[0].PositionID)
	}
	if zielData.Positionen[0].EinzelpreisCents != 350 {
		t.Errorf("expected target einzelpreis 350, got %d", zielData.Positionen[0].EinzelpreisCents)
	}
}

// The Kommentar is cut to the schema limit of 100 bytes (zog's Max counts bytes). The target name ends
// on a rune boundary, the source name forces a one-byte step back so no UTF-8 sequence splits.
func TestBestellungUmbuchen_KommentarWirdGekuerzt(t *testing.T) {
	ctx := context.Background()
	quellTisch := tisch.Tisch{ID: 1, Name: "T" + strings.Repeat("ä", 49), Status: tisch.ActiveStatus}
	zielTisch := tisch.Tisch{ID: 2, Name: strings.Repeat("Ä", 50), Status: tisch.ActiveStatus}

	eventMock := repotest.NewKassenjournalRepo(nil, nil)
	quellSubject := kasse.TischSessionSubject(testKassensitzungNr, quellTisch.ID)
	zielSubject := kasse.TischSessionSubject(testKassensitzungNr, zielTisch.ID)
	quellPositionID := uuid.New().String()

	eventMock.SetTischSession(quellSubject, kasse.TischSession{
		UnbezahltePositionen: []kasse.Position{{
			PositionID:   quellPositionID,
			VarianteID:   1,
			ProduktName:  "Cola",
			VarianteName: "0,5l",
			Kategorie:    "getraenk", Steuersatz: "regel",
			EinzelpreisCents: 350,
			Menge:            1,
		}},
	})

	command := Command{
		TischRepo:           repotest.NewTischRepo([]tisch.Tisch{quellTisch, zielTisch}, nil),
		EventRepo:           eventMock,
		KassensitzungenRepo: repotest.NewKassensitzungenRepo(testOpenKS, nil),
	}

	err := command.BestellungUmbuchen(ctx, 1, "Test User", uuid.NewString(), quellTisch.ID, zielTisch.ID, []kasse.PositionRef{{PositionID: quellPositionID, Menge: 1}}, "")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	quellEvents, _ := eventMock.ReadEventsBySubject(ctx, quellSubject)
	zielEvents, _ := eventMock.ReadEventsBySubject(ctx, zielSubject)

	var quellData kasse.BestellungUmgebuchtV1Data
	if err := json.Unmarshal(quellEvents[0].Data, &quellData); err != nil {
		t.Fatalf("expected no unmarshal error for source umbuchung data, got %v", err)
	}
	var zielData kasse.BestellungUmgebuchtV1Data
	if err := json.Unmarshal(zielEvents[0].Data, &zielData); err != nil {
		t.Fatalf("expected no unmarshal error for target umbuchung data, got %v", err)
	}

	// 14 Byte Präfix + 43 mal "Ä" (2 Byte) füllen die Grenze genau aus.
	wantQuell := "Umbuchung auf " + strings.Repeat("Ä", 43)
	if quellData.Kommentar != wantQuell {
		t.Errorf("expected source comment %q, got %q", wantQuell, quellData.Kommentar)
	}
	// 14 Byte Präfix + "T" + 42 mal "ä" sind 99 Bytes; das 100. Byte gehört zur
	// nächsten UTF-8-Folge und fällt mit ihr weg.
	wantZiel := "Umbuchung von T" + strings.Repeat("ä", 42)
	if zielData.Kommentar != wantZiel {
		t.Errorf("expected target comment %q, got %q", wantZiel, zielData.Kommentar)
	}
}

func TestBestellungUmbuchen_PositionNichtUmbuchbar(t *testing.T) {
	ctx := context.Background()
	quellTisch := tisch.Tisch{ID: 1, Name: "Tisch Quelle", Status: tisch.ActiveStatus}
	zielTisch := tisch.Tisch{ID: 2, Name: "Tisch Ziel", Status: tisch.ActiveStatus}

	eventMock := repotest.NewKassenjournalRepo(nil, nil)
	quellSubject := kasse.TischSessionSubject(testKassensitzungNr, quellTisch.ID)
	eventMock.SetTischSession(quellSubject, kasse.TischSession{
		UnbezahltePositionen: []kasse.Position{{
			PositionID:   uuid.New().String(),
			VarianteID:   1,
			ProduktName:  "Cola",
			VarianteName: "0,5l",
			Kategorie:    "getraenk", Steuersatz: "regel",
			EinzelpreisCents: 350,
			Menge:            1,
		}},
	})

	command := Command{
		TischRepo:           repotest.NewTischRepo([]tisch.Tisch{quellTisch, zielTisch}, nil),
		EventRepo:           eventMock,
		KassensitzungenRepo: repotest.NewKassensitzungenRepo(testOpenKS, nil),
	}

	err := command.BestellungUmbuchen(ctx, 1, "Test User", uuid.NewString(), quellTisch.ID, zielTisch.ID, []kasse.PositionRef{{PositionID: uuid.New().String(), Menge: 1}}, "")
	if !errors.Is(err, ErrPositionNichtUmbuchbar) {
		t.Errorf("expected ErrPositionNichtUmbuchbar, got %v", err)
	}
}

func TestBestellungUmbuchen_GleicherTisch(t *testing.T) {
	err := Command{}.BestellungUmbuchen(context.Background(), 1, "Test User", uuid.NewString(), 3, 3, []kasse.PositionRef{{PositionID: uuid.New().String(), Menge: 1}}, "")
	if !errors.Is(err, ErrUmbuchungGleicherTisch) {
		t.Errorf("expected ErrUmbuchungGleicherTisch, got %v", err)
	}
}

func TestBestellungUmbuchen_ZielTischNotActive(t *testing.T) {
	ctx := context.Background()
	quellTisch := tisch.Tisch{ID: 1, Name: "Tisch Quelle", Status: tisch.ActiveStatus}
	zielTisch := tisch.Tisch{ID: 2, Name: "Tisch Ziel", Status: tisch.InactiveStatus}

	command := Command{
		TischRepo:           repotest.NewTischRepo([]tisch.Tisch{quellTisch, zielTisch}, nil),
		EventRepo:           repotest.NewKassenjournalRepo(nil, nil),
		KassensitzungenRepo: repotest.NewKassensitzungenRepo(testOpenKS, nil),
	}

	err := command.BestellungUmbuchen(ctx, 1, "Test User", uuid.NewString(), quellTisch.ID, zielTisch.ID, []kasse.PositionRef{{PositionID: uuid.New().String(), Menge: 1}}, "")
	if !errors.Is(err, ErrTischNotActive) {
		t.Errorf("expected ErrTischNotActive, got %v", err)
	}
}

func TestBestellungUmbuchen_ZielTischNotFound(t *testing.T) {
	ctx := context.Background()
	quellTisch := tisch.Tisch{ID: 1, Name: "Tisch Quelle", Status: tisch.ActiveStatus}

	command := Command{
		TischRepo: &umbuchungTischRepoMock{tables: map[int]tisch.Tisch{
			quellTisch.ID: quellTisch,
		}},
		EventRepo:           repotest.NewKassenjournalRepo(nil, nil),
		KassensitzungenRepo: repotest.NewKassensitzungenRepo(testOpenKS, nil),
	}

	err := command.BestellungUmbuchen(ctx, 1, "Test User", uuid.NewString(), quellTisch.ID, 99, []kasse.PositionRef{{PositionID: uuid.New().String(), Menge: 1}}, "")
	if !errors.Is(err, ErrTischNotFound) {
		t.Errorf("expected ErrTischNotFound, got %v", err)
	}
}

func TestBestellungUmbuchen_KasseNichtGeoeffnet(t *testing.T) {
	ctx := context.Background()
	command := Command{
		TischRepo:           repotest.NewTischRepo([]tisch.Tisch{testActiveTisch}, nil),
		EventRepo:           repotest.NewKassenjournalRepo(nil, nil),
		KassensitzungenRepo: repotest.NewKassensitzungenRepo(nil, nil),
	}

	err := command.BestellungUmbuchen(ctx, 1, "Test User", uuid.NewString(), 1, 2, []kasse.PositionRef{{PositionID: uuid.New().String(), Menge: 1}}, "")
	if !errors.Is(err, ErrKasseNichtGeoeffnet) {
		t.Errorf("expected ErrKasseNichtGeoeffnet, got %v", err)
	}
}

// parallelCommitRepo runs vorLesen once, right before the first projection read: a parallel commit
// that lands after any earlier read of the stream.
type parallelCommitRepo struct {
	*repotest.KassenjournalRepo
	vorLesen func()
}

func (r *parallelCommitRepo) ReadTischSession(ctx context.Context, subject string) (kasse.TischSession, error) {
	if r.vorLesen != nil {
		r.vorLesen()
		r.vorLesen = nil
	}
	return r.KassenjournalRepo.ReadTischSession(ctx, subject)
}

// A parallel attempt with the same umbuchungId commits just before the projection read. The retry
// must see it and book nothing, not move the quantity a second time from the remaining stock.
func TestBestellungUmbuchen_ParallelerVersuch_BuchtEinmal(t *testing.T) {
	ctx := context.Background()
	quellTisch := tisch.Tisch{ID: 1, Name: "Tisch Quelle", Status: tisch.ActiveStatus}
	zielTisch := tisch.Tisch{ID: 2, Name: "Tisch Ziel", Status: tisch.ActiveStatus}
	quellSubject := kasse.TischSessionSubject(testKassensitzungNr, quellTisch.ID)
	positionID := uuid.NewString()
	bier := func(menge, version int) kasse.TischSession {
		return kasse.TischSession{
			UnbezahltePositionen: []kasse.Position{{
				PositionID: positionID, VarianteID: 1, ProduktName: "Bier", VarianteName: "0,5l",
				Kategorie: "getraenk", Steuersatz: "regel", EinzelpreisCents: 400, Menge: menge,
			}},
			LastEventVersion: version,
		}
	}

	inner := repotest.NewKassenjournalRepo(nil, nil)
	inner.SetTischSession(quellSubject, bier(5, 0))
	repo := &parallelCommitRepo{KassenjournalRepo: inner}
	command := Command{
		TischRepo:           repotest.NewTischRepo([]tisch.Tisch{quellTisch, zielTisch}, nil),
		EventRepo:           repo,
		KassensitzungenRepo: repotest.NewKassensitzungenRepo(testOpenKS, nil),
	}
	umbuchungID := uuid.NewString()
	refs := []kasse.PositionRef{{PositionID: positionID, Menge: 2}}

	repo.vorLesen = func() {
		parallel := command
		parallel.EventRepo = inner
		if err := parallel.BestellungUmbuchen(ctx, 1, "Test User", umbuchungID, quellTisch.ID, zielTisch.ID, refs, ""); err != nil {
			t.Fatalf("parallel attempt: %v", err)
		}
		inner.SetTischSession(quellSubject, bier(3, 1))
	}

	if err := command.BestellungUmbuchen(ctx, 1, "Test User", umbuchungID, quellTisch.ID, zielTisch.ID, refs, ""); err != nil {
		t.Fatalf("expected the retry to succeed, got %v", err)
	}

	quellEvents, err := inner.ReadEventsBySubject(ctx, quellSubject)
	if err != nil {
		t.Fatalf("read source events: %v", err)
	}
	if len(quellEvents) != 1 {
		t.Errorf("expected 1 Abgang, got %d", len(quellEvents))
	}
}
