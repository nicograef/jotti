//go:build unit

package application

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/nicograef/jotti/backend/api/kasse/enrichment"
	"github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/event"
	"github.com/nicograef/jotti/backend/domain/kasse"
	"github.com/nicograef/jotti/backend/domain/produkt"
	"github.com/nicograef/jotti/backend/domain/tisch"
	"github.com/nicograef/jotti/backend/repository/druckauftrag_repo"
	"github.com/nicograef/jotti/backend/repository/kassenjournal_repo"
	"github.com/nicograef/jotti/backend/repository/kassensitzungen_repo"
	"github.com/nicograef/jotti/backend/repository/produkt_repo"
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

// bestellCommand liefert ein Command mit aktivem Tisch und aktiver Variante samt
// der zugehörigen Bestell-Eingabe (Menge 1) für die Bestell-Idempotenz-Tests.
func bestellCommand(eventMock *kassenjournal_repo.MockRepo) (Command, []enrichment.PositionInput) {
	productMock := produkt_repo.NewMock([]produkt.Produkt{testProduct}, nil)
	productMock.AddVariante(testProduct.ID, testVariant)

	command := newTestCommandWithEventMock([]tisch.Tisch{testActiveTisch}, []produkt.Produkt{testProduct}, eventMock)
	command.ProduktRepo = productMock
	return command, []enrichment.PositionInput{{ProduktID: testProduct.ID, VarianteID: testVariant.ID, Menge: 1}}
}

// Zwei identische Bestell-Aufrufe mit derselben bestellungId: genau ein Event,
// beide Male Erfolg — der zweite Aufruf bucht nicht erneut.
func TestBestellungAufnehmen_DuplikatBestellungId_KeinZweitesEvent(t *testing.T) {
	ctx := context.Background()
	subject := kasse.TischSessionSubject(testKassensitzungNr, testActiveTisch.ID)

	eventMock := kassenjournal_repo.NewMock(nil, nil)
	command, inputs := bestellCommand(eventMock)

	if err := command.BestellungAufnehmen(ctx, 1, "Test User", testVorgangID, testActiveTisch.ID, inputs, ""); err != nil {
		t.Fatalf("erster Aufruf: %v", err)
	}
	if err := command.BestellungAufnehmen(ctx, 1, "Test User", testVorgangID, testActiveTisch.ID, inputs, ""); err != nil {
		t.Fatalf("zweiter Aufruf (Duplikat) erwartet nil, bekam: %v", err)
	}

	events, _ := eventMock.ReadEventsBySubject(ctx, subject)
	if len(events) != 1 {
		t.Fatalf("erwartet genau 1 Event, gespeichert: %d", len(events))
	}
	if len(eventMock.CapturedDruckauftraege()) != 0 {
		t.Errorf("ohne konfigurierte Druckstation dürfen keine Druckaufträge entstehen, erzeugt: %d", len(eventMock.CapturedDruckauftraege()))
	}

	vorgang, ok := eventMock.GebuchterVorgang(testVorgangID)
	if !ok {
		t.Fatal("erwartet eine vorgang_idempotenz-Zeile für die bestellungId")
	}
	if vorgang.Art != kassenjournal_repo.VorgangArtBestellung {
		t.Errorf("erwartet art %q, gespeichert: %q", kassenjournal_repo.VorgangArtBestellung, vorgang.Art)
	}
}

// Dieselbe bestellungId mit geänderter Menge ist weder ein Duplikat noch eine
// neue Buchung: Der Command meldet den Konflikt und schreibt kein zweites Event.
func TestBestellungAufnehmen_SelbeBestellungIdAndereNutzdaten_DatenAbweichend(t *testing.T) {
	ctx := context.Background()
	subject := kasse.TischSessionSubject(testKassensitzungNr, testActiveTisch.ID)

	eventMock := kassenjournal_repo.NewMock(nil, nil)
	command, inputs := bestellCommand(eventMock)

	if err := command.BestellungAufnehmen(ctx, 1, "Test User", testVorgangID, testActiveTisch.ID, inputs, ""); err != nil {
		t.Fatalf("erster Aufruf: %v", err)
	}

	geaendert := []enrichment.PositionInput{{ProduktID: testProduct.ID, VarianteID: testVariant.ID, Menge: 2}}
	err := command.BestellungAufnehmen(ctx, 1, "Test User", testVorgangID, testActiveTisch.ID, geaendert, "")
	if !errors.Is(err, ErrVorgangDatenAbweichend) {
		t.Fatalf("erwartet ErrVorgangDatenAbweichend, bekam: %v", err)
	}

	events, _ := eventMock.ReadEventsBySubject(ctx, subject)
	if len(events) != 1 {
		t.Fatalf("erwartet weiterhin genau 1 Event, gespeichert: %d", len(events))
	}
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

	// Zweiter Aufruf mit eigener vorgangId, der den Tischzustand noch VOR der
	// ersten Zahlung gelesen hat: Die Projektion wird auf den veralteten Stand
	// zurückgesetzt (der erste Write hat sie fortgeschrieben). Der Write des
	// zweiten Aufrufs kollidiert damit mit dem bereits geschriebenen Event.
	eventMock.SetTischSession(subject, session)
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

	orderEvent.Version = 1
	paymentEvent.Version = 2
	seedTischStream(t, eventMock, subject, orderEvent, paymentEvent)

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

// umbuchungTestSetup füllt den Mock mit einer unbezahlten Position auf dem
// Quell-Tisch und liefert Command, beide Tische und den Umbuchungs-Ref.
func umbuchungTestSetup(eventMock *kassenjournal_repo.MockRepo) (Command, tisch.Tisch, tisch.Tisch, []kasse.PositionRef) {
	quellTisch := tisch.Tisch{ID: 1, Name: "Tisch Quelle", Status: tisch.ActiveStatus}
	zielTisch := tisch.Tisch{ID: 2, Name: "Tisch Ziel", Status: tisch.ActiveStatus}
	quellPositionID := uuid.New().String()

	eventMock.SetTischSession(kasse.TischSessionSubject(testKassensitzungNr, quellTisch.ID), kasse.TischSession{
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
	return command, quellTisch, zielTisch, []kasse.PositionRef{{PositionID: quellPositionID, Menge: 1}}
}

// Zwei identische Umbuchungs-Aufrufe mit derselben vorgangId: genau ein Quell-
// und ein Ziel-Event, beide Male Erfolg.
func TestBestellungUmbuchen_DuplikatVorgangId_IdempotenterErfolg(t *testing.T) {
	ctx := context.Background()
	eventMock := kassenjournal_repo.NewMock(nil, nil)
	command, quellTisch, zielTisch, refs := umbuchungTestSetup(eventMock)
	quellSubject := kasse.TischSessionSubject(testKassensitzungNr, quellTisch.ID)
	zielSubject := kasse.TischSessionSubject(testKassensitzungNr, zielTisch.ID)

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

// Abweichende Nutzdaten unter bekanntem Schlüssel: Der Server meldet den
// Konflikt, statt die geänderte Einreichung zu verschlucken oder ein zweites Mal
// zu buchen. Die Vorprüfung greift dabei VOR der fachlichen Validierung — sonst
// endete der zweite Aufruf in „Position nicht bezahlbar" (die Positionen des
// ersten Aufrufs sind bereits kassiert).
func TestZahlungKassieren_SelbeVorgangIdAndereNutzdaten_DatenAbweichend(t *testing.T) {
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

	err := command.ZahlungKassieren(ctx, 1, "Test User", testVorgangID, testActiveTisch.ID, refs, "Trinkgeld")
	if !errors.Is(err, ErrVorgangDatenAbweichend) {
		t.Fatalf("erwartet ErrVorgangDatenAbweichend, bekam: %v", err)
	}

	events, _ := eventMock.ReadEventsBySubject(ctx, subject)
	if len(events) != 1 {
		t.Fatalf("erwartet weiterhin genau 1 Event, gespeichert: %d", len(events))
	}
}

func TestStornierungErteilen_SelbeVorgangIdAndereNutzdaten_DatenAbweichend(t *testing.T) {
	ctx := context.Background()
	eventMock := kassenjournal_repo.NewMock(nil, nil)
	command, subject, refs := stornoTestSetup(t, eventMock)

	if err := command.StornierungErteilen(ctx, 2, "Leitung", testVorgangID, testActiveTisch.ID, refs, "Reklamation"); err != nil {
		t.Fatalf("erster Aufruf: %v", err)
	}
	eventsNachErstem, _ := eventMock.ReadEventsBySubject(ctx, subject)

	err := command.StornierungErteilen(ctx, 2, "Leitung", testVorgangID, testActiveTisch.ID, refs, "Anderer Grund")
	if !errors.Is(err, ErrVorgangDatenAbweichend) {
		t.Fatalf("erwartet ErrVorgangDatenAbweichend, bekam: %v", err)
	}

	eventsNachZweitem, _ := eventMock.ReadEventsBySubject(ctx, subject)
	if len(eventsNachZweitem) != len(eventsNachErstem) {
		t.Fatalf("abweichende Einreichung hat Events geschrieben: %d -> %d", len(eventsNachErstem), len(eventsNachZweitem))
	}
}

func TestBestellungUmbuchen_SelbeVorgangIdAndereNutzdaten_DatenAbweichend(t *testing.T) {
	ctx := context.Background()
	eventMock := kassenjournal_repo.NewMock(nil, nil)
	command, quellTisch, zielTisch, refs := umbuchungTestSetup(eventMock)
	quellSubject := kasse.TischSessionSubject(testKassensitzungNr, quellTisch.ID)

	if err := command.BestellungUmbuchen(ctx, 1, "Test User", testVorgangID, quellTisch.ID, zielTisch.ID, refs, ""); err != nil {
		t.Fatalf("erster Aufruf: %v", err)
	}

	err := command.BestellungUmbuchen(ctx, 1, "Test User", testVorgangID, quellTisch.ID, zielTisch.ID, refs, "Anderer Kommentar")
	if !errors.Is(err, ErrVorgangDatenAbweichend) {
		t.Fatalf("erwartet ErrVorgangDatenAbweichend, bekam: %v", err)
	}

	quellEvents, _ := eventMock.ReadEventsBySubject(ctx, quellSubject)
	if len(quellEvents) != 1 {
		t.Fatalf("erwartet weiterhin genau 1 Quell-Event, gespeichert: %d", len(quellEvents))
	}
}

// Derselbe Schlüssel an zwei Endpunkten mit identischer Feldbelegung:
// zahlungNutzdaten und stornierungNutzdaten tragen dieselben Felder mit
// denselben json-Tags und serialisieren byteidentisch. Nur weil die Art des
// Vorgangs mit in den Hash geht, gilt die Stornierung nicht als Duplikat der
// Zahlung — sonst liefe sie in die stille Erfolgsantwort, ohne zu buchen.
func TestStornierungErteilen_VorgangIdDerZahlungGleicheFelder_KeinDuplikat(t *testing.T) {
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
		t.Fatalf("Zahlung: %v", err)
	}

	// Gleiche vorgangId, gleicher Tisch, gleiche Positionen, gleicher Kommentar —
	// allein die Art unterscheidet sich. Eine Neuanlage ist unmöglich (der
	// Primärschlüssel steht auf vorgang_id allein), also bleibt der Konflikt.
	err := command.StornierungErteilen(ctx, 1, "Test User", testVorgangID, testActiveTisch.ID, refs, "")
	if !errors.Is(err, ErrVorgangDatenAbweichend) {
		t.Fatalf("erwartet ErrVorgangDatenAbweichend, bekam: %v", err)
	}
}

// vorgangWriteSpy liefert den Schlüsselkonflikt aus der Schreibtransaktion,
// ohne dass die Vorprüfung anschlägt — genau das Rennen zweier gleichzeitiger
// Anfragen um denselben Schlüssel. Sequenziell ist dieser Zweig nicht
// erreichbar: Die Vorprüfung fängt jede Zweiteinreichung vorher ab.
type vorgangWriteSpy struct {
	*kassenjournal_repo.MockRepo
	writeVorgangErr error
}

func (s *vorgangWriteSpy) DetermineVorgangStatus(_ context.Context, _ string, _ []byte) (kassenjournal_repo.VorgangStatus, error) {
	return kassenjournal_repo.VorgangNeu, nil
}

func (s *vorgangWriteSpy) WriteEventMitVorgang(_ context.Context, _ kassenjournal_repo.Vorgang, _ event.Event, _ kasse.StreamType, _ int) (int, error) {
	return 0, s.writeVorgangErr
}

func (s *vorgangWriteSpy) WriteEventWithDruckauftraegeMitVorgang(_ context.Context, _ kassenjournal_repo.Vorgang, _ event.Event, _ kasse.StreamType, _ int, _ func(event.Event) []druckauftrag_repo.NeuerDruckauftrag) (int, error) {
	return 0, s.writeVorgangErr
}

func (s *vorgangWriteSpy) WriteTischSessionEventsAtomicMitVorgang(_ context.Context, _ kassenjournal_repo.Vorgang, _ []event.Event, _ int) error {
	return s.writeVorgangErr
}

func (s *vorgangWriteSpy) WriteUmbuchungMitVorgang(_ context.Context, _ kassenjournal_repo.Vorgang, _ event.Event, _ event.Event, _ int) error {
	return s.writeVorgangErr
}

// Verlieren zwei gleichzeitige Anfragen das Rennen um denselben Schlüssel,
// schlägt erst der Insert in der Schreibtransaktion fehl. Der Nutzdaten-Hash
// entscheidet auch dort: gleiche Nutzdaten ergeben die stille Erfolgsantwort,
// abweichende den Konflikt.
func TestBuchendeVorgaenge_SchluesselkonfliktImWrite(t *testing.T) {
	faelle := []struct {
		name      string
		writeErr  error
		erwartung error
	}{
		{name: "gleiche Nutzdaten", writeErr: kassenjournal_repo.ErrVorgangBereitsGebucht, erwartung: nil},
		{name: "abweichende Nutzdaten", writeErr: kassenjournal_repo.ErrVorgangDatenAbweichend, erwartung: ErrVorgangDatenAbweichend},
	}

	kommandos := []struct {
		name   string
		aufruf func(t *testing.T, spy *vorgangWriteSpy) error
	}{
		{
			name: "BestellungAufnehmen",
			aufruf: func(_ *testing.T, spy *vorgangWriteSpy) error {
				command, inputs := bestellCommand(spy.MockRepo)
				command.EventRepo = spy
				return command.BestellungAufnehmen(context.Background(), 1, "Test User", testVorgangID, testActiveTisch.ID, inputs, "")
			},
		},
		{
			name: "ZahlungKassieren",
			aufruf: func(_ *testing.T, spy *vorgangWriteSpy) error {
				session, refs := kassierbareSession()
				spy.SetTischSession(kasse.TischSessionSubject(testKassensitzungNr, testActiveTisch.ID), session)
				command := Command{
					TischRepo:           tisch_repo.NewMock([]tisch.Tisch{testActiveTisch}, nil),
					EventRepo:           spy,
					KassensitzungenRepo: kassensitzungen_repo.NewMock(testOpenKS, nil),
				}
				return command.ZahlungKassieren(context.Background(), 1, "Test User", testVorgangID, testActiveTisch.ID, refs, "")
			},
		},
		{
			name: "StornierungErteilen",
			aufruf: func(t *testing.T, spy *vorgangWriteSpy) error {
				command, _, refs := stornoTestSetup(t, spy.MockRepo)
				command.EventRepo = spy
				return command.StornierungErteilen(context.Background(), 2, "Leitung", testVorgangID, testActiveTisch.ID, refs, "Reklamation")
			},
		},
		{
			name: "BestellungUmbuchen",
			aufruf: func(_ *testing.T, spy *vorgangWriteSpy) error {
				command, quellTisch, zielTisch, refs := umbuchungTestSetup(spy.MockRepo)
				command.EventRepo = spy
				return command.BestellungUmbuchen(context.Background(), 1, "Test User", testVorgangID, quellTisch.ID, zielTisch.ID, refs, "")
			},
		},
	}

	for _, kommando := range kommandos {
		for _, fall := range faelle {
			t.Run(kommando.name+"/"+fall.name, func(t *testing.T) {
				spy := &vorgangWriteSpy{MockRepo: kassenjournal_repo.NewMock(nil, nil), writeVorgangErr: fall.writeErr}

				if err := kommando.aufruf(t, spy); !errors.Is(err, fall.erwartung) {
					t.Fatalf("erwartet %v, bekam: %v", fall.erwartung, err)
				}
			})
		}
	}
}
