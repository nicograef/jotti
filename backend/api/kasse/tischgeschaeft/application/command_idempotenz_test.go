//go:build unit

package application

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/kasse"
	"github.com/nicograef/jotti/backend/domain/tisch"
	"github.com/nicograef/jotti/backend/repository/kassenjournal_repo"
	"github.com/nicograef/jotti/backend/repository/kassensitzungen_repo"
	"github.com/nicograef/jotti/backend/repository/tisch_repo"
)

// Idempotenz der buchenden Vorgänge (vorgangId): Erstbuchung, identische
// Wiederholung (genau ein Event-Satz, beide Male Erfolg) und echter
// OCC-Konflikt je Vorgang.

// kassierbareSession liefert eine Tisch-Session mit einer offenen Position
// (Menge 1, 350 Cent) samt zugehörigem PositionRef für die Idempotenz-Tests.
func kassierbareSession() (kasse.TischSession, []kasse.PositionRef) {
	pos := kasse.Position{PositionID: "22222222-2222-4222-8222-222222222222", VarianteID: 1, ProduktName: "Cola", VarianteName: "0,5l", Kategorie: "getraenk", Steuersatz: "regel", EinzelpreisCents: 350, Menge: 1}
	return kasse.TischSession{
		SaldoCents:           350,
		UnbezahltePositionen: []kasse.Position{pos},
	}, []kasse.PositionRef{{PositionID: pos.PositionID, Menge: 1}}
}

// Zwei identische Kassier-Aufrufe mit derselben vorgangId: genau ein Event,
// beide Male Erfolg — der zweite Aufruf bucht nicht erneut.
func TestZahlungKassieren_DuplikatVorgangId_IdempotenterErfolg(t *testing.T) {
	ctx := context.Background()
	subject := kasse.TischSessionSubject(testKassensitzungNr, testActiveTisch.ID)

	eventMock := kassenjournal_repo.NewMock(nil, nil)
	session, refs := kassierbareSession()
	eventMock.SetTischSession(subject, session)

	command := Command{
		TischRepo:           tisch_repo.NewMock([]tisch.Tisch{testActiveTisch}, nil),
		EventRepo:           eventMock,
		KassensitzungenRepo: kassensitzungen_repo.NewMock(testOpenKS, nil),
	}

	if err := command.ZahlungKassieren(ctx, 1, "Test User", testVorgangID, testActiveTisch.ID, refs, ""); err != nil {
		t.Fatalf("erster Aufruf: %v", err)
	}
	if err := command.ZahlungKassieren(ctx, 1, "Test User", testVorgangID, testActiveTisch.ID, refs, ""); err != nil {
		t.Fatalf("zweiter Aufruf (Duplikat) erwartet nil, bekam: %v", err)
	}

	events, err := eventMock.ReadEventsBySubject(ctx, subject)
	if err != nil {
		t.Fatalf("expected no read error, got %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("erwartet genau 1 Event, gespeichert: %d", len(events))
	}

	vorgang, ok := eventMock.GebuchterVorgang(testVorgangID)
	if !ok {
		t.Fatal("erwartet eine vorgang_idempotenz-Zeile für die vorgangId")
	}
	if vorgang.Art != kassenjournal_repo.VorgangArtZahlung {
		t.Errorf("erwartet art %q, gespeichert: %q", kassenjournal_repo.VorgangArtZahlung, vorgang.Art)
	}
	if vorgang.UserID != 1 {
		t.Errorf("erwartet user_id 1, gespeichert: %d", vorgang.UserID)
	}
}

// Ein echter OCC-Konflikt (neue vorgangId, veraltete Stream-Version) bleibt
// ErrConflict und hinterlässt weder Events noch eine Idempotenz-Zeile.
func TestZahlungKassieren_NeueVorgangIdVeralteteVersion_ErrConflict(t *testing.T) {
	ctx := context.Background()
	subject := kasse.TischSessionSubject(testKassensitzungNr, testActiveTisch.ID)

	eventMock := kassenjournal_repo.NewMock(nil, nil)
	session, refs := kassierbareSession()
	eventMock.SetTischSession(subject, session)

	command := Command{
		TischRepo:           tisch_repo.NewMock([]tisch.Tisch{testActiveTisch}, nil),
		EventRepo:           eventMock,
		KassensitzungenRepo: kassensitzungen_repo.NewMock(testOpenKS, nil),
	}

	if err := command.ZahlungKassieren(ctx, 1, "Test User", testVorgangID, testActiveTisch.ID, refs, ""); err != nil {
		t.Fatalf("erster Aufruf: %v", err)
	}

	// Neue vorgangId, aber die Projektion (und damit die erwartete Version) ist
	// veraltet — der Write kollidiert mit dem bereits geschriebenen Event.
	andereVorgangID := "cccccccc-cccc-4ccc-8ccc-cccccccccccc"
	if err := command.ZahlungKassieren(ctx, 1, "Test User", andereVorgangID, testActiveTisch.ID, refs, ""); !errors.Is(err, ErrConflict) {
		t.Fatalf("expected ErrConflict, got %v", err)
	}

	events, err := eventMock.ReadEventsBySubject(ctx, subject)
	if err != nil {
		t.Fatalf("expected no read error, got %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("erwartet weiterhin genau 1 Event, gespeichert: %d", len(events))
	}
	if _, ok := eventMock.GebuchterVorgang(andereVorgangID); ok {
		t.Error("erwartet keine vorgang_idempotenz-Zeile für den gescheiterten Vorgang (Rollback)")
	}
}

// stornoTestSetup füllt den Mock mit einem Tisch-Stream aus Bestellung (Menge 3)
// und einer Zahlung (Menge 1) — ein Storno über 3 spaltet in Korrektur +
// Warenrücknahme — und liefert Command, Subject und die Storno-Refs.
func stornoTestSetup(t *testing.T, eventMock *kassenjournal_repo.MockRepo) (Command, string, []kasse.PositionRef) {
	t.Helper()
	subject := kasse.TischSessionSubject(testKassensitzungNr, testActiveTisch.ID)

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

	eventMock.SetTischSession(subject, kasse.TischSession{})
	orderEvent.Version = 1
	paymentEvent.Version = 2
	eventMock.AddEvent(orderEvent)
	eventMock.AddEvent(paymentEvent)

	command := Command{
		TischRepo:           tisch_repo.NewMock([]tisch.Tisch{testActiveTisch}, nil),
		EventRepo:           eventMock,
		KassensitzungenRepo: kassensitzungen_repo.NewMock(testOpenKS, nil),
	}
	return command, subject, []kasse.PositionRef{{PositionID: posID, Menge: 3}}
}

// Zwei identische Storno-Aufrufe mit derselben vorgangId: die Event-Anzahl des
// Vorgangs (Korrektur + Warenrücknahme) bleibt unverändert, beide Male Erfolg.
func TestStornierungErteilen_DuplikatVorgangId_IdempotenterErfolg(t *testing.T) {
	ctx := context.Background()
	eventMock := kassenjournal_repo.NewMock(nil, nil)
	command, subject, refs := stornoTestSetup(t, eventMock)

	if err := command.StornierungErteilen(ctx, 2, "Leitung", testVorgangID, testActiveTisch.ID, refs, "Reklamation"); err != nil {
		t.Fatalf("erster Aufruf: %v", err)
	}
	eventsNachErstem, _ := eventMock.ReadEventsBySubject(ctx, subject)

	if err := command.StornierungErteilen(ctx, 2, "Leitung", testVorgangID, testActiveTisch.ID, refs, "Reklamation"); err != nil {
		t.Fatalf("zweiter Aufruf (Duplikat) erwartet nil, bekam: %v", err)
	}
	eventsNachZweitem, _ := eventMock.ReadEventsBySubject(ctx, subject)

	// 2 Setup-Events + 1 bestellung-korrigiert + 1 stornierung-erteilt
	if len(eventsNachErstem) != 4 {
		t.Fatalf("erwartet 4 Events nach dem ersten Aufruf, gespeichert: %d", len(eventsNachErstem))
	}
	if len(eventsNachZweitem) != len(eventsNachErstem) {
		t.Fatalf("Duplikat hat Events geschrieben: %d -> %d", len(eventsNachErstem), len(eventsNachZweitem))
	}

	vorgang, ok := eventMock.GebuchterVorgang(testVorgangID)
	if !ok {
		t.Fatal("erwartet eine vorgang_idempotenz-Zeile für die vorgangId")
	}
	if vorgang.Art != kassenjournal_repo.VorgangArtStornierung {
		t.Errorf("erwartet art %q, gespeichert: %q", kassenjournal_repo.VorgangArtStornierung, vorgang.Art)
	}
}

// Ein echter OCC-Konflikt beim Storno-Write (UNIQUE(subject, version), z. B.
// durch ein parallel committetes Event) bleibt ErrConflict und hinterlässt
// keine Idempotenz-Zeile.
func TestStornierungErteilen_NeueVorgangIdVeralteteVersion_ErrConflict(t *testing.T) {
	ctx := context.Background()
	eventMock := kassenjournal_repo.NewMockWithWriteErr(nil, db.ErrAlreadyExists)
	command, _, refs := stornoTestSetup(t, eventMock)

	err := command.StornierungErteilen(ctx, 2, "Leitung", testVorgangID, testActiveTisch.ID, refs, "Reklamation")
	if !errors.Is(err, ErrConflict) {
		t.Fatalf("expected ErrConflict, got %v", err)
	}
	if _, ok := eventMock.GebuchterVorgang(testVorgangID); ok {
		t.Error("erwartet keine vorgang_idempotenz-Zeile für den gescheiterten Vorgang (Rollback)")
	}
}

// Zwei identische Umbuchungs-Aufrufe mit derselben vorgangId: genau ein Quell-
// und ein Ziel-Event, beide Male Erfolg.
func TestBestellungUmbuchen_DuplikatVorgangId_IdempotenterErfolg(t *testing.T) {
	ctx := context.Background()
	quellTisch := tisch.Tisch{ID: 1, Name: "Tisch Quelle", Status: tisch.ActiveStatus}
	zielTisch := tisch.Tisch{ID: 2, Name: "Tisch Ziel", Status: tisch.ActiveStatus}

	eventMock := kassenjournal_repo.NewMock(nil, nil)
	quellSubject := kasse.TischSessionSubject(testKassensitzungNr, quellTisch.ID)
	zielSubject := kasse.TischSessionSubject(testKassensitzungNr, zielTisch.ID)
	quellPositionID := uuid.New().String()

	eventMock.SetTischSession(quellSubject, kasse.TischSession{
		SaldoCents: 350,
		UnbezahltePositionen: []kasse.Position{{
			PositionID: quellPositionID, VarianteID: 1, ProduktName: "Cola", VarianteName: "0,5l",
			Kategorie: "getraenk", Steuersatz: "regel", EinzelpreisCents: 350, Menge: 1,
		}},
	})

	command := Command{
		TischRepo:           tisch_repo.NewMock([]tisch.Tisch{quellTisch, zielTisch}, nil),
		EventRepo:           eventMock,
		KassensitzungenRepo: kassensitzungen_repo.NewMock(testOpenKS, nil),
	}

	refs := []kasse.PositionRef{{PositionID: quellPositionID, Menge: 1}}
	if err := command.BestellungUmbuchen(ctx, 1, "Test User", testVorgangID, quellTisch.ID, zielTisch.ID, refs, ""); err != nil {
		t.Fatalf("erster Aufruf: %v", err)
	}
	if err := command.BestellungUmbuchen(ctx, 1, "Test User", testVorgangID, quellTisch.ID, zielTisch.ID, refs, ""); err != nil {
		t.Fatalf("zweiter Aufruf (Duplikat) erwartet nil, bekam: %v", err)
	}

	quellEvents, _ := eventMock.ReadEventsBySubject(ctx, quellSubject)
	zielEvents, _ := eventMock.ReadEventsBySubject(ctx, zielSubject)
	if len(quellEvents) != 1 || len(zielEvents) != 1 {
		t.Fatalf("erwartet genau 1 Quell- und 1 Ziel-Event, gespeichert: quell=%d ziel=%d", len(quellEvents), len(zielEvents))
	}

	vorgang, ok := eventMock.GebuchterVorgang(testVorgangID)
	if !ok {
		t.Fatal("erwartet eine vorgang_idempotenz-Zeile für die vorgangId")
	}
	if vorgang.Art != kassenjournal_repo.VorgangArtUmbuchung {
		t.Errorf("erwartet art %q, gespeichert: %q", kassenjournal_repo.VorgangArtUmbuchung, vorgang.Art)
	}
}
