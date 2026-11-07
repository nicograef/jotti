package application

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/nicograef/jotti/backend/domain/kasse"
	"github.com/nicograef/jotti/backend/domain/tisch"
	"github.com/nicograef/jotti/backend/repository/repotest"
)

func TestStornierungErteilen_AlreadyPaidPosition_Succeeds(t *testing.T) {
	ctx := context.Background()
	subject := kasse.TischSessionSubject(testKassensitzungNr, testActiveTisch.ID)

	orderEvent, _ := kasse.NewBestellungAufgenommenEvent(subject, 1, "Test User", "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa",
		[]kasse.Position{
			{VarianteID: 1, ProduktName: "Cola", VarianteName: "0,5l", Kategorie: "getraenk", Steuersatz: "regel", EinzelpreisCents: 350, Menge: 1},
		}, "")

	var orderData struct {
		Positionen []struct {
			PositionID string `json:"positionId"`
		} `json:"positionen"`
	}
	if err := json.Unmarshal(orderEvent.Data, &orderData); err != nil {
		t.Fatalf("expected no unmarshal error, got %v", err)
	}
	posID := orderData.Positionen[0].PositionID

	paymentEvent, _ := kasse.NewZahlungKassiertEvent(subject, 1, "Test User",
		[]kasse.Position{
			{PositionID: posID, VarianteID: 1, ProduktName: "Cola", VarianteName: "0,5l", Kategorie: "getraenk", Steuersatz: "regel", EinzelpreisCents: 350, Menge: 1},
		}, 350, "")

	eventMock := repotest.NewKassenjournalRepo(nil, nil)
	eventMock.SetTischSession(subject, kasse.TischSession{
		SaldoCents:           0,
		UnbezahltePositionen: []kasse.Position{},
		GesamtZahlungenCents: 350,
	})
	orderEvent.Subject = subject
	paymentEvent.Subject = subject
	eventMock.AddEvent(orderEvent)
	eventMock.AddEvent(paymentEvent)

	command := Command{
		TischRepo:           repotest.NewTischRepo([]tisch.Tisch{testActiveTisch}, nil),
		EventRepo:           eventMock,
		KassensitzungenRepo: repotest.NewKassensitzungenRepo(testOpenKS, nil),
	}

	refs := []kasse.PositionRef{{PositionID: posID, Menge: 1}}

	err := command.StornierungErteilen(ctx, 1, "Test User", testActiveTisch.ID, refs, "Reklamation")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestStornierungErteilen_AlreadyCancelledPosition_Fails(t *testing.T) {
	ctx := context.Background()
	subject := kasse.TischSessionSubject(testKassensitzungNr, testActiveTisch.ID)

	orderEvent, _ := kasse.NewBestellungAufgenommenEvent(subject, 1, "Test User", "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa",
		[]kasse.Position{
			{VarianteID: 1, ProduktName: "Cola", VarianteName: "0,5l", Kategorie: "getraenk", Steuersatz: "regel", EinzelpreisCents: 350, Menge: 1},
		}, "")

	var orderData struct {
		Positionen []struct {
			PositionID string `json:"positionId"`
		} `json:"positionen"`
	}
	if err := json.Unmarshal(orderEvent.Data, &orderData); err != nil {
		t.Fatalf("expected no unmarshal error, got %v", err)
	}
	posID := orderData.Positionen[0].PositionID

	// Die Position wurde bereits geldneutral korrigiert (unbezahlt) — ein erneuter
	// Storno darf nicht mehr greifen.
	cancelEvent, _ := kasse.NewBestellungKorrigiertEvent(subject, 1, "Test User",
		[]kasse.Position{
			{PositionID: posID, VarianteID: 1, ProduktName: "Cola", VarianteName: "0,5l", Kategorie: "getraenk", Steuersatz: "regel", EinzelpreisCents: 350, Menge: 1},
		}, 350, "Test")

	eventMock := repotest.NewKassenjournalRepo(nil, nil)
	eventMock.SetTischSession(subject, kasse.TischSession{
		SaldoCents:           0,
		UnbezahltePositionen: []kasse.Position{},
		GesamtZahlungenCents: 0,
	})
	orderEvent.Subject = subject
	cancelEvent.Subject = subject
	eventMock.AddEvent(orderEvent)
	eventMock.AddEvent(cancelEvent)

	command := Command{
		TischRepo:           repotest.NewTischRepo([]tisch.Tisch{testActiveTisch}, nil),
		EventRepo:           eventMock,
		KassensitzungenRepo: repotest.NewKassensitzungenRepo(testOpenKS, nil),
	}

	refs := []kasse.PositionRef{{PositionID: posID, Menge: 1}}

	err := command.StornierungErteilen(ctx, 1, "Test User", testActiveTisch.ID, refs, "")
	if !errors.Is(err, ErrPositionNichtStornierbar) {
		t.Errorf("expected ErrPositionNichtStornierbar, got %v", err)
	}
}

func TestStornierungErteilen_DuplikatPositionRefs(t *testing.T) {
	ctx := context.Background()
	subject := kasse.TischSessionSubject(testKassensitzungNr, testActiveTisch.ID)

	orderEvent, _ := kasse.NewBestellungAufgenommenEvent(subject, 1, "Test User", "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa",
		[]kasse.Position{
			{VarianteID: 1, ProduktName: "Cola", VarianteName: "0,5l", Kategorie: "getraenk", Steuersatz: "regel", EinzelpreisCents: 350, Menge: 3},
		}, "")

	var orderData struct {
		Positionen []struct {
			PositionID string `json:"positionId"`
		} `json:"positionen"`
	}
	if err := json.Unmarshal(orderEvent.Data, &orderData); err != nil {
		t.Fatalf("expected no unmarshal error, got %v", err)
	}
	posID := orderData.Positionen[0].PositionID

	eventMock := repotest.NewKassenjournalRepo(nil, nil)
	eventMock.SetTischSession(subject, duplikatTestSession())
	orderEvent.Subject = subject
	eventMock.AddEvent(orderEvent)

	command := Command{
		TischRepo:           repotest.NewTischRepo([]tisch.Tisch{testActiveTisch}, nil),
		EventRepo:           eventMock,
		KassensitzungenRepo: repotest.NewKassensitzungenRepo(testOpenKS, nil),
	}

	refs := []kasse.PositionRef{
		{PositionID: posID, Menge: 1},
		{PositionID: posID, Menge: 1},
	}

	err := command.StornierungErteilen(ctx, 1, "Test User", testActiveTisch.ID, refs, "Duplikat")
	if !errors.Is(err, ErrPositionNichtStornierbar) {
		t.Errorf("expected ErrPositionNichtStornierbar, got %v", err)
	}
}

func TestStornierungErteilen_GemischterStorno_AtomischKorrekturUndWarenruecknahme(t *testing.T) {
	ctx := context.Background()
	subject := kasse.TischSessionSubject(testKassensitzungNr, testActiveTisch.ID)

	// Bestellt 3 Cola, davon 1 bezahlt → ein Storno von 3 spaltet in 2 Korrektur
	// (unbezahlt) und 1 Warenrücknahme (bezahlt).
	orderEvent, _ := kasse.NewBestellungAufgenommenEvent(subject, 1, "Test User", "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa", []kasse.Position{{
		VarianteID: 1, ProduktName: "Cola", VarianteName: "0,5l", Kategorie: "getraenk", Steuersatz: "regel", EinzelpreisCents: 350, Menge: 3,
	}}, "")
	var orderData kasse.BestellungAufgenommenV1Data
	if err := json.Unmarshal(orderEvent.Data, &orderData); err != nil {
		t.Fatalf("expected no unmarshal error, got %v", err)
	}
	posID := orderData.Positionen[0].PositionID

	paymentEvent, _ := kasse.NewZahlungKassiertEvent(subject, 1, "Test User", []kasse.Position{{
		PositionID: posID, VarianteID: 1, ProduktName: "Cola", VarianteName: "0,5l", Kategorie: "getraenk", Steuersatz: "regel", EinzelpreisCents: 350, Menge: 1,
	}}, 350, "")

	eventMock := repotest.NewKassenjournalRepo(nil, nil)
	eventMock.SetTischSession(subject, kasse.TischSession{})
	eventMock.AddEvent(orderEvent)
	eventMock.AddEvent(paymentEvent)

	command := Command{
		TischRepo:           repotest.NewTischRepo([]tisch.Tisch{testActiveTisch}, nil),
		EventRepo:           eventMock,
		KassensitzungenRepo: repotest.NewKassensitzungenRepo(testOpenKS, nil),
	}

	err := command.StornierungErteilen(ctx, 2, "Leitung", testActiveTisch.ID, []kasse.PositionRef{{PositionID: posID, Menge: 3}}, "Reklamation")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	events, err := eventMock.ReadEventsBySubject(ctx, subject)
	if err != nil {
		t.Fatalf("expected no error reading events, got %v", err)
	}

	var korrektur, warenruecknahme int
	for _, evt := range events {
		switch evt.Type {
		case string(kasse.EventTypeBestellungKorrigiertV1):
			korrektur++
		case string(kasse.EventTypeStornierungErteiltV1):
			warenruecknahme++
		}
	}
	if korrektur != 1 {
		t.Errorf("expected exactly 1 bestellung-korrigiert, got %d", korrektur)
	}
	if warenruecknahme != 1 {
		t.Errorf("expected exactly 1 stornierung-erteilt, got %d", warenruecknahme)
	}
}
