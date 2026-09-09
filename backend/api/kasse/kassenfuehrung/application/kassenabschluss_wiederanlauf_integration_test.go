//go:build integration

package application

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/nicograef/jotti/backend/domain/event"
	"github.com/nicograef/jotti/backend/domain/kasse"
	"github.com/nicograef/jotti/backend/repository/kassenjournal_repo"
	"github.com/nicograef/jotti/backend/repository/kassensitzungen_repo"
	"github.com/nicograef/jotti/backend/repository/tse_repo"
)

// teilfehlerJournalRepo umhüllt das echte Repository und lässt den ersten
// Schreibversuch des konfigurierten Event-Typs fehlschlagen — simuliert einen
// Teilfehler des Kassenabschlusses nach Schritt 1.
type teilfehlerJournalRepo struct {
	kassenjournalRepo
	failType   string
	failedOnce bool
}

func (f *teilfehlerJournalRepo) WriteEvent(ctx context.Context, e event.Event, streamType kasse.StreamType, kassensitzungNr int) (int, error) {
	if !f.failedOnce && e.Type == f.failType {
		f.failedOnce = true
		return 0, errors.New("injizierter Teilfehler")
	}
	return f.kassenjournalRepo.WriteEvent(ctx, e, streamType, kassensitzungNr)
}

func countJournalEvents(t *testing.T, db *sql.DB, eventType string) int {
	t.Helper()
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM kassenjournal WHERE type = $1", eventType).Scan(&count); err != nil {
		t.Fatalf("kassenjournal zaehlen (%s): %v", eventType, err)
	}
	return count
}

// TestKasseAbschliessen_RetryNachTeilfehler_KeinZweiterKassensturz: Der erste
// Abschluss-Versuch schreibt den Kassensturz und scheitert an der
// Differenzbuchung (Teilfehler). Der Wiederanlauf erkennt den vorhandenen
// Kassensturz, überspringt Schritt 1 und schließt ab — im Journal steht
// genau ein kassensturz-durchgefuehrt:v1.
func TestKasseAbschliessen_RetryNachTeilfehler_KeinZweiterKassensturz(t *testing.T) {
	ctx, _, db, userID := setupKassenfuehrungIntegration(t)

	failing := &teilfehlerJournalRepo{
		kassenjournalRepo: kassenjournal_repo.NewRepository(db),
		failType:          string(kasse.EventTypeDifferenzSollIstGebuchtV1),
	}
	cmd := Command{
		KassenjournalRepo:   failing,
		KassensitzungenRepo: kassensitzungen_repo.NewRepository(db),
		TSERepo:             tse_repo.NewRepository(db),
	}

	// Soll-Bestand ist 0 (keine Buchungen); Ist-Bestand 500 erzwingt eine
	// Differenzbuchung — genau dort schlägt der erste Versuch fehl.
	if _, err := cmd.KasseAbschliessen(ctx, userID, "test", 500); err == nil {
		t.Fatal("erster Versuch: Teilfehler erwartet, bekam nil")
	}

	if count := countJournalEvents(t, db, string(kasse.EventTypeKassensturzDurchgefuehrtV1)); count != 1 {
		t.Fatalf("nach Teilfehler: erwartet 1 kassensturz-Event, gespeichert: %d", count)
	}
	if count := countJournalEvents(t, db, string(kasse.EventTypeTagesabschlussErstelltV1)); count != 0 {
		t.Fatalf("nach Teilfehler: erwartet 0 tagesabschluss-Events, gespeichert: %d", count)
	}

	// Wiederanlauf: erkennt den vorhandenen Kassensturz und schließt ab.
	if _, err := cmd.KasseAbschliessen(ctx, userID, "test", 500); err != nil {
		t.Fatalf("Wiederanlauf erwartet Erfolg, bekam: %v", err)
	}

	if count := countJournalEvents(t, db, string(kasse.EventTypeKassensturzDurchgefuehrtV1)); count != 1 {
		t.Errorf("nach Wiederanlauf: erwartet genau 1 kassensturz-Event, gespeichert: %d", count)
	}
	if count := countJournalEvents(t, db, string(kasse.EventTypeDifferenzSollIstGebuchtV1)); count != 1 {
		t.Errorf("nach Wiederanlauf: erwartet genau 1 differenz-Event, gespeichert: %d", count)
	}
	if count := countJournalEvents(t, db, string(kasse.EventTypeTagesabschlussErstelltV1)); count != 1 {
		t.Errorf("nach Wiederanlauf: erwartet genau 1 tagesabschluss-Event, gespeichert: %d", count)
	}

	// Die Differenz rechnet gegen den im Kassensturz dokumentierten Ist-Bestand
	// (Soll 0 − Ist 500 = −500, Überschuss).
	var differenzCents int
	if err := db.QueryRow(
		"SELECT (data->>'betragCents')::int FROM kassenjournal WHERE type = $1",
		string(kasse.EventTypeDifferenzSollIstGebuchtV1),
	).Scan(&differenzCents); err != nil {
		t.Fatalf("differenz-Betrag lesen: %v", err)
	}
	if differenzCents != -500 {
		t.Errorf("erwartet Differenz -500, gespeichert: %d", differenzCents)
	}

	var status string
	if err := db.QueryRow("SELECT status FROM kassensitzungen").Scan(&status); err != nil {
		t.Fatalf("kassensitzung status lesen: %v", err)
	}
	if status != string(kasse.KassensitzungAbgeschlossen) {
		t.Errorf("erwartet Status %q, gespeichert: %q", kasse.KassensitzungAbgeschlossen, status)
	}
}

// TestKasseAbschliessen_RetryNachZwischenbuchung_BrichtAb: Der erste Abschluss-Versuch
// schreibt den Kassensturz und scheitert an der Differenzbuchung (Teilfehler). Der defer
// setzt die Sitzung zurück auf 'offen'; danach entsteht eine echte Zwischenbuchung
// (Geldtransit). Der Wiederanlauf erkennt die Buchung nach dem protokollierten Kassensturz
// und bricht mit ErrBuchungenNachKassensturz ab, ohne ein Abschluss-Event zu schreiben —
// der veraltete Ist-Bestand wird nicht wiederverwendet.
func TestKasseAbschliessen_RetryNachZwischenbuchung_BrichtAb(t *testing.T) {
	ctx, _, db, userID := setupKassenfuehrungIntegration(t)

	failing := &teilfehlerJournalRepo{
		kassenjournalRepo: kassenjournal_repo.NewRepository(db),
		failType:          string(kasse.EventTypeDifferenzSollIstGebuchtV1),
	}
	cmd := Command{
		KassenjournalRepo:   failing,
		KassensitzungenRepo: kassensitzungen_repo.NewRepository(db),
		TSERepo:             tse_repo.NewRepository(db),
	}

	// Soll-Bestand ist 0 (keine Buchungen); Ist-Bestand 500 erzwingt eine
	// Differenzbuchung — genau dort schlägt der erste Versuch fehl.
	if _, err := cmd.KasseAbschliessen(ctx, userID, "test", 500); err == nil {
		t.Fatal("erster Versuch: Teilfehler erwartet, bekam nil")
	}

	// Der defer hat die Sitzung nach dem Teilfehler wieder auf 'offen' gesetzt.
	var status string
	if err := db.QueryRow("SELECT status FROM kassensitzungen").Scan(&status); err != nil {
		t.Fatalf("kassensitzung status lesen: %v", err)
	}
	if status != string(kasse.KassensitzungOffen) {
		t.Fatalf("nach Teilfehler erwartet Status 'offen', gespeichert: %q", status)
	}

	// Zwischenbuchung: eine legitime Einlage nach dem protokollierten Kassensturz.
	if err := cmd.GeldtransitBuchen(ctx, userID, "test", uuid.NewString(), "einlage", 1000, "Wechselgeld nachgelegt"); err != nil {
		t.Fatalf("Zwischenbuchung fehlgeschlagen: %v", err)
	}

	signaturauftraegeErledigen(t, db)

	// Wiederanlauf muss abbrechen: Der alte Ist-Bestand ist durch die Buchung veraltet.
	if _, err := cmd.KasseAbschliessen(ctx, userID, "test", 500); !errors.Is(err, ErrBuchungenNachKassensturz) {
		t.Fatalf("Wiederanlauf erwartet ErrBuchungenNachKassensturz, bekam: %v", err)
	}

	// Kein Abschluss-Event darf geschrieben worden sein.
	if count := countJournalEvents(t, db, string(kasse.EventTypeDifferenzSollIstGebuchtV1)); count != 0 {
		t.Errorf("erwartet 0 differenz-Events nach Abbruch, gespeichert: %d", count)
	}
	if count := countJournalEvents(t, db, string(kasse.EventTypeTagesabschlussErstelltV1)); count != 0 {
		t.Errorf("erwartet 0 tagesabschluss-Events nach Abbruch, gespeichert: %d", count)
	}

	// Die Sitzung bleibt nach dem Abbruch wieder 'offen' (defer-Reset greift auch hier).
	if err := db.QueryRow("SELECT status FROM kassensitzungen").Scan(&status); err != nil {
		t.Fatalf("kassensitzung status nach Abbruch lesen: %v", err)
	}
	if status != string(kasse.KassensitzungOffen) {
		t.Errorf("nach Abbruch erwartet Status 'offen', gespeichert: %q", status)
	}
}

// signaturauftraegeErledigen markiert alle offenen Signaturaufträge als erledigt.
// In Produktion arbeitet sie der Outbox-Worker vor dem nächsten Abschluss ab; hier
// muss das Signatur-Gate durchlassen, damit der Wiederanlauf die Prüfung auf
// Zwischenbuchungen überhaupt erreicht.
func signaturauftraegeErledigen(t *testing.T, db *sql.DB) {
	t.Helper()
	if _, err := db.Exec("UPDATE tse_signaturauftraege SET status = 'erledigt', erledigt_am = now() WHERE status = 'offen'"); err != nil {
		t.Fatalf("Signaturauftraege als erledigt markieren: %v", err)
	}
}

// bezahlteBestellungAmTisch bucht an einem frischen Tisch eine Bestellung und ihre
// Zahlung über dieselben Positionen; die Tisch-Session bleibt mit Saldo 0 zurück.
// Beide Events liegen im Tisch-Sub-Stream, den die Stream-Prüfung des Wiederanlaufs
// nicht sieht — die Zahlung zeigt sich allein am Soll-Kassenbestand.
func bezahlteBestellungAmTisch(ctx context.Context, t *testing.T, db *sql.DB, userID, betragCents int) {
	t.Helper()

	var zNr int
	if err := db.QueryRow("SELECT z_nr FROM kassensitzungen").Scan(&zNr); err != nil {
		t.Fatalf("z_nr lesen: %v", err)
	}
	var tischID int
	if err := db.QueryRow(
		"INSERT INTO tische (name, status, created_at, updated_at) VALUES ('Tisch 1', 'active', now(), now()) RETURNING id",
	).Scan(&tischID); err != nil {
		t.Fatalf("Tisch anlegen: %v", err)
	}

	repo := kassenjournal_repo.NewRepository(db)
	subject := kasse.TischSessionSubject(zNr, tischID)
	bestellt := []kasse.Position{{
		VarianteID:       1,
		ProduktName:      "Bier",
		VarianteName:     "0,5 l",
		Kategorie:        "getraenk",
		Steuersatz:       "regel",
		EinzelpreisCents: betragCents,
		Menge:            1,
	}}

	bestellung, err := kasse.NewBestellungAufgenommenEvent(subject, userID, "test", uuid.NewString(), bestellt, "")
	if err != nil {
		t.Fatalf("bestellung-aufgenommen bauen: %v", err)
	}
	bestellung.Version = 1
	if _, err := repo.WriteEvent(ctx, bestellung, kasse.StreamTypeTischSession, zNr); err != nil {
		t.Fatalf("bestellung-aufgenommen schreiben: %v", err)
	}

	// Die Zahlung muss die Positionen samt der beim Bestellen erzeugten PositionIDs
	// nennen, sonst weist die Projektion sie ab.
	var bestellungData kasse.BestellungAufgenommenV1Data
	if err := json.Unmarshal(bestellung.Data, &bestellungData); err != nil {
		t.Fatalf("bestellung-aufgenommen lesen: %v", err)
	}
	bezahlt := make([]kasse.Position, 0, len(bestellungData.Positionen))
	for _, pos := range bestellungData.Positionen {
		bezahlt = append(bezahlt, kasse.PositionFromEventData(pos))
	}

	zahlung, err := kasse.NewZahlungKassiertEvent(subject, userID, "test", bezahlt, betragCents, "")
	if err != nil {
		t.Fatalf("zahlung-kassiert bauen: %v", err)
	}
	zahlung.Version = 2
	if _, err := repo.WriteEvent(ctx, zahlung, kasse.StreamTypeTischSession, zNr); err != nil {
		t.Fatalf("zahlung-kassiert schreiben: %v", err)
	}
}

// TestKasseAbschliessen_RetryNachTischzahlung_BrichtAb: Der erste Versuch schreibt
// den Kassensturz und scheitert an der Differenzbuchung; der defer setzt die Sitzung
// zurück auf 'offen'. Danach bezahlt ein Tisch seine Bestellung. Der Wiederanlauf
// erkennt die Zwischenbuchung am veränderten Soll-Bestand — obwohl sie in einem
// Sub-Stream liegt — und bricht mit ErrBuchungenNachKassensturz ab.
func TestKasseAbschliessen_RetryNachTischzahlung_BrichtAb(t *testing.T) {
	ctx, _, db, userID := setupKassenfuehrungIntegration(t)

	failing := &teilfehlerJournalRepo{
		kassenjournalRepo: kassenjournal_repo.NewRepository(db),
		failType:          string(kasse.EventTypeDifferenzSollIstGebuchtV1),
	}
	cmd := Command{
		KassenjournalRepo:   failing,
		KassensitzungenRepo: kassensitzungen_repo.NewRepository(db),
		TSERepo:             tse_repo.NewRepository(db),
	}

	// Soll-Bestand ist 0 (keine Buchungen); Ist-Bestand 500 erzwingt eine
	// Differenzbuchung — genau dort schlägt der erste Versuch fehl.
	if _, err := cmd.KasseAbschliessen(ctx, userID, "test", 500); err == nil {
		t.Fatal("erster Versuch: Teilfehler erwartet, bekam nil")
	}

	bezahlteBestellungAmTisch(ctx, t, db, userID, 350)
	signaturauftraegeErledigen(t, db)

	if _, err := cmd.KasseAbschliessen(ctx, userID, "test", 500); !errors.Is(err, ErrBuchungenNachKassensturz) {
		t.Fatalf("Wiederanlauf erwartet ErrBuchungenNachKassensturz, bekam: %v", err)
	}

	if count := countJournalEvents(t, db, string(kasse.EventTypeDifferenzSollIstGebuchtV1)); count != 0 {
		t.Errorf("erwartet 0 differenz-Events nach Abbruch, gespeichert: %d", count)
	}
	if count := countJournalEvents(t, db, string(kasse.EventTypeTagesabschlussErstelltV1)); count != 0 {
		t.Errorf("erwartet 0 tagesabschluss-Events nach Abbruch, gespeichert: %d", count)
	}
}

// TestKasseAbschliessen_RetryNachDifferenzbuchung_LaeuftDurch: Der erste Versuch
// schreibt Kassensturz und Differenzbuchung und scheitert am Tagesabschluss. Beim
// Wiederanlauf hat die gebuchte Differenz den Soll-Bestand an den gezählten
// Ist-Bestand angeglichen; der Vergleich ohne sie sieht deshalb keine
// Zwischenbuchung und der Abschluss läuft durch.
func TestKasseAbschliessen_RetryNachDifferenzbuchung_LaeuftDurch(t *testing.T) {
	ctx, _, db, userID := setupKassenfuehrungIntegration(t)

	failing := &teilfehlerJournalRepo{
		kassenjournalRepo: kassenjournal_repo.NewRepository(db),
		failType:          string(kasse.EventTypeTagesabschlussErstelltV1),
	}
	cmd := Command{
		KassenjournalRepo:   failing,
		KassensitzungenRepo: kassensitzungen_repo.NewRepository(db),
		TSERepo:             tse_repo.NewRepository(db),
	}

	if _, err := cmd.KasseAbschliessen(ctx, userID, "test", 500); err == nil {
		t.Fatal("erster Versuch: Teilfehler erwartet, bekam nil")
	}
	if count := countJournalEvents(t, db, string(kasse.EventTypeDifferenzSollIstGebuchtV1)); count != 1 {
		t.Fatalf("nach Teilfehler: erwartet 1 differenz-Event, gespeichert: %d", count)
	}

	signaturauftraegeErledigen(t, db)

	if _, err := cmd.KasseAbschliessen(ctx, userID, "test", 500); err != nil {
		t.Fatalf("Wiederanlauf erwartet Erfolg, bekam: %v", err)
	}

	if count := countJournalEvents(t, db, string(kasse.EventTypeKassensturzDurchgefuehrtV1)); count != 1 {
		t.Errorf("nach Wiederanlauf: erwartet genau 1 kassensturz-Event, gespeichert: %d", count)
	}
	if count := countJournalEvents(t, db, string(kasse.EventTypeDifferenzSollIstGebuchtV1)); count != 1 {
		t.Errorf("nach Wiederanlauf: erwartet genau 1 differenz-Event, gespeichert: %d", count)
	}
	if count := countJournalEvents(t, db, string(kasse.EventTypeTagesabschlussErstelltV1)); count != 1 {
		t.Errorf("nach Wiederanlauf: erwartet genau 1 tagesabschluss-Event, gespeichert: %d", count)
	}

	var status string
	if err := db.QueryRow("SELECT status FROM kassensitzungen").Scan(&status); err != nil {
		t.Fatalf("kassensitzung status lesen: %v", err)
	}
	if status != string(kasse.KassensitzungAbgeschlossen) {
		t.Errorf("erwartet Status %q, gespeichert: %q", kasse.KassensitzungAbgeschlossen, status)
	}
}
