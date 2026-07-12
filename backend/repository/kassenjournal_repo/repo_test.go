//go:build integration

package kassenjournal_repo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	dbpkg "github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/event"
	"github.com/nicograef/jotti/backend/domain/kasse"
	"github.com/nicograef/jotti/backend/repository/druckauftrag_repo"
)

func createUser(db *sql.DB) (int, error) {
	var userID int
	err := db.QueryRow("INSERT INTO users (name, username, role, status, password_hash, onetime_password_hash, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, now(), now()) RETURNING id", "nico", "nico", "admin", "active", "hashedpassword", "onetimesethash").Scan(&userID)
	if err != nil {
		return 0, err
	}
	return userID, nil
}

func createTisch(db *sql.DB, name string) (int, error) {
	var tischID int
	err := db.QueryRow("INSERT INTO tische (name, status, created_at, updated_at) VALUES ($1, 'active', now(), now()) RETURNING id", name).Scan(&tischID)
	if err != nil {
		return 0, err
	}
	return tischID, nil
}

func createKassensitzung(db *sql.DB) (int, error) {
	var zNr int
	err := db.QueryRow("INSERT INTO kassensitzungen (datum, bezeichnung, status, created_at, updated_at) VALUES ((NOW() AT TIME ZONE 'Europe/Berlin')::date, $1, $2, NOW(), NOW()) RETURNING z_nr", "Test-Sitzung", kasse.KassensitzungOffen).Scan(&zNr)
	if err != nil {
		return 0, err
	}
	return zNr, nil
}

// insertEventRaw inserts an event directly via SQL, bypassing WriteEvent and the projection.
// Use this for test setup where the projection is not relevant.
func insertEventRaw(db *sql.DB, e event.Event, kassensitzungNr int) (int, error) {
	var id int
	err := db.QueryRow(
		"INSERT INTO kassenjournal (user_id, user_name, type, subject, version, data, timestamp, kassensitzung_nr) VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id",
		e.UserID, e.UserName, e.Type, e.Subject, e.Version, e.Data, e.Time, kassensitzungNr,
	).Scan(&id)
	return id, err
}

// newTestEvent creates a test event with the given parameters and version.
func newTestEvent(userID int, eventType, subject string, version int, data any) event.Event {
	e, _ := event.New(userID, "nico", eventType, subject, data)
	e.Version = version
	return e
}

// validBestellungData returns valid bestellung-aufgenommen:v1 event data for testing.
func validBestellungData(bestellungID, positionID string, einzelpreis, menge int) map[string]any {
	return map[string]any{
		"bestellungId": bestellungID,
		"positionen": []map[string]any{
			{
				"positionId":       positionID,
				"varianteId":       1,
				"produktName":      "Bier",
				"varianteName":     "0.5L",
				"kategorie":        "getraenk",
				"einzelpreisCents": einzelpreis,
				"menge":            menge,
			},
		},
		"gesamtPreisCents": einzelpreis * menge,
		"kommentar":        "",
	}
}

// validZahlungData returns valid zahlung-kassiert:v1 event data for testing
// (fat positions — die fiskalische Projektion braucht Steuersatz und Preis).
func validZahlungData(positionID string, menge, gesamtCents int) map[string]any {
	return map[string]any{
		"zahlungId": "z0000000-0000-0000-0000-000000000001",
		"positionen": []map[string]any{
			{
				"positionId":       positionID,
				"varianteId":       1,
				"produktName":      "Bier",
				"varianteName":     "0.5L",
				"kategorie":        "getraenk",
				"steuersatz":       "regel",
				"einzelpreisCents": gesamtCents / menge,
				"menge":            menge,
			},
		},
		"gesamtZahlungCents": gesamtCents,
		"kommentar":          "",
	}
}

// validStornierungData returns valid stornierung-erteilt:v1 (kassenwirksame Warenrücknahme) event data.
func validStornierungData(betragCents int) map[string]any {
	return map[string]any{
		"stornierungId":          "11111111-0000-0000-0000-000000000001",
		"zahlungId":              "z0000000-0000-0000-0000-000000000001",
		"gesamtStornierungCents": betragCents,
		"kommentar":              "Rueckgabe",
		"positionen": []map[string]any{
			{
				"positionId":       "p0000000-0000-0000-0000-000000000001",
				"varianteId":       1,
				"produktName":      "Bier",
				"varianteName":     "0.5L",
				"kategorie":        "getraenk",
				"steuersatz":       "regel",
				"einzelpreisCents": betragCents,
				"menge":            1,
			},
		},
	}
}

// validKorrekturData returns valid bestellung-korrigiert:v1 (geldneutrale Korrektur) event data.
func validKorrekturData(betragCents int) map[string]any {
	return map[string]any{
		"korrekturId": "22222222-0000-0000-0000-000000000001",
		"gesamtCents": betragCents,
		"kommentar":   "",
		"positionen": []map[string]any{
			{
				"positionId":       "p0000000-0000-0000-0000-000000000002",
				"varianteId":       1,
				"produktName":      "Bier",
				"varianteName":     "0.5L",
				"kategorie":        "getraenk",
				"steuersatz":       "regel",
				"einzelpreisCents": betragCents,
				"menge":            1,
			},
		},
	}
}

func validDirektverkaufData(verkaufID string, gesamtbetragCents int) map[string]any {
	return map[string]any{
		"verkaufId":         verkaufID,
		"gesamtbetragCents": gesamtbetragCents,
		"positionen": []map[string]any{
			{
				"positionId":       "d0000000-0000-0000-0000-000000000001",
				"varianteId":       1,
				"produktName":      "Bier",
				"varianteName":     "0.5L",
				"kategorie":        "getraenk",
				"steuersatz":       "regel",
				"einzelpreisCents": gesamtbetragCents,
				"menge":            1,
			},
		},
		"kommentar": "",
	}
}

func validDirektverkaufStornoData(verkaufID string, gesamtStornierungCents int) map[string]any {
	return map[string]any{
		"stornierungId":          "s0000000-0000-0000-0000-000000000001",
		"verkaufId":              verkaufID,
		"gesamtStornierungCents": gesamtStornierungCents,
		"positionen": []map[string]any{
			{
				"positionId":       "d0000000-0000-0000-0000-000000000001",
				"varianteId":       1,
				"produktName":      "Bier",
				"varianteName":     "0.5L",
				"kategorie":        "getraenk",
				"steuersatz":       "regel",
				"einzelpreisCents": gesamtStornierungCents,
				"menge":            1,
			},
		},
		"kommentar": "Rueckgabe",
	}
}

func cleanDB(t *testing.T, db *sql.DB) {
	_, err := db.Exec("DELETE FROM tse_signaturauftraege")
	if err != nil {
		t.Fatalf("Failed to clean tse_signaturauftraege: %v", err)
	}
	_, err = db.Exec("DELETE FROM druckauftraege")
	if err != nil {
		t.Fatalf("Failed to clean druckauftraege: %v", err)
	}
	_, err = db.Exec("DELETE FROM tisch_sessions")
	if err != nil {
		t.Fatalf("Failed to clean tisch_sessions: %v", err)
	}
	_, err = db.Exec("ALTER TABLE kassenjournal DISABLE TRIGGER kassenjournal_no_delete")
	if err != nil {
		t.Fatalf("Failed to disable kassenjournal_no_delete trigger: %v", err)
	}
	_, err = db.Exec("DELETE FROM kassenjournal")
	if err != nil {
		t.Fatalf("Failed to clean kassenjournal table: %v", err)
	}
	_, err = db.Exec("ALTER TABLE kassenjournal ENABLE TRIGGER kassenjournal_no_delete")
	if err != nil {
		t.Fatalf("Failed to enable kassenjournal_no_delete trigger: %v", err)
	}
	_, err = db.Exec("DELETE FROM kassensitzungen")
	if err != nil {
		t.Fatalf("Failed to clean kassensitzungen table: %v", err)
	}
	_, err = db.Exec("DELETE FROM tische")
	if err != nil {
		t.Fatalf("Failed to clean tische table: %v", err)
	}
	_, err = db.Exec("DELETE FROM users")
	if err != nil {
		t.Fatalf("Failed to clean users table: %v", err)
	}
}

func setup(t *testing.T) (int, int, Repository, func(t *testing.T)) {
	db := dbpkg.OpenTestDatabase()

	cleanDB(t, db)

	userID, err := createUser(db)
	if err != nil {
		t.Fatalf("Failed to insert user: %v", err)
	}

	ksNr, err := createKassensitzung(db)
	if err != nil {
		t.Fatalf("Failed to create kassensitzung: %v", err)
	}

	return userID, ksNr, NewRepository(db), func(t *testing.T) {
		cleanDB(t, db)
		_ = db.Close()
	}
}

func TestWriteEvent_TischSession(t *testing.T) {
	userID, ksNr, repo, teardown := setup(t)
	defer teardown(t)

	tischID, err := createTisch(repo.db, "Tisch 1")
	if err != nil {
		t.Fatalf("Failed to create tisch: %v", err)
	}

	subject := kasse.TischSessionSubject(ksNr, tischID)
	data := validBestellungData("b0000000-0000-0000-0000-000000000001", "p0000000-0000-0000-0000-000000000001", 350, 2)
	e := newTestEvent(userID, "bestellung-aufgenommen:v1", subject, 1, data)

	eventID, err := repo.WriteEvent(context.Background(), e, kasse.StreamTypeTischSession, ksNr)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if eventID == 0 {
		t.Fatalf("Expected valid event ID, got %d", eventID)
	}
}

// ReadFavoritenTischStates liefert Name + Session je Favorit in einer Query:
// byte-identisch zu ReadTischSession für einen Tisch mit Projektion, eine
// Null-Session für einen Tisch ohne Events und keine Zeile für gelöschte/
// unbekannte Tische.
func TestReadFavoritenTischStates(t *testing.T) {
	userID, ksNr, repo, teardown := setup(t)
	defer teardown(t)

	tischA, err := createTisch(repo.db, "Tisch A")
	if err != nil {
		t.Fatalf("Failed to create tisch A: %v", err)
	}
	tischB, err := createTisch(repo.db, "Tisch B")
	if err != nil {
		t.Fatalf("Failed to create tisch B: %v", err)
	}

	// Tisch A bekommt eine Bestellung -> Session-Projektion entsteht; Tisch B bleibt leer.
	subjectA := kasse.TischSessionSubject(ksNr, tischA)
	data := validBestellungData("b0000000-0000-0000-0000-000000000001", "p0000000-0000-0000-0000-000000000001", 350, 2)
	e := newTestEvent(userID, "bestellung-aufgenommen:v1", subjectA, 1, data)
	if _, err := repo.WriteEvent(context.Background(), e, kasse.StreamTypeTischSession, ksNr); err != nil {
		t.Fatalf("Failed to write bestellung: %v", err)
	}

	// Gelöschter Tisch darf nicht auftauchen (WHERE status != 'deleted').
	tischDeleted, err := createTisch(repo.db, "Tisch Deleted")
	if err != nil {
		t.Fatalf("Failed to create tisch deleted: %v", err)
	}
	if _, err := repo.db.Exec("UPDATE tische SET status = 'deleted' WHERE id = $1", tischDeleted); err != nil {
		t.Fatalf("Failed to soft-delete tisch: %v", err)
	}

	states, err := repo.ReadFavoritenTischStates(context.Background(), []int{tischA, tischB, tischDeleted, 999999}, ksNr)
	if err != nil {
		t.Fatalf("ReadFavoritenTischStates: %v", err)
	}
	if len(states) != 2 {
		t.Fatalf("expected 2 states (A, B), got %d", len(states))
	}

	// Tisch A: Name + Session byte-identisch zu ReadTischSession.
	wantSessionA, err := repo.ReadTischSession(context.Background(), subjectA)
	if err != nil {
		t.Fatalf("ReadTischSession A: %v", err)
	}
	gotA, ok := states[tischA]
	if !ok {
		t.Fatalf("expected tisch A in states")
	}
	if gotA.Name != "Tisch A" {
		t.Errorf("expected name 'Tisch A', got %q", gotA.Name)
	}
	if !reflect.DeepEqual(gotA.Session, wantSessionA) {
		t.Errorf("session A mismatch:\n batch  = %+v\n single = %+v", gotA.Session, wantSessionA)
	}
	if gotA.Session.SaldoCents != 700 {
		t.Errorf("expected saldo 700 for tisch A, got %d", gotA.Session.SaldoCents)
	}

	// Tisch B: Name + Null-Session (wie ReadTischSession ohne Projektion).
	gotB, ok := states[tischB]
	if !ok {
		t.Fatalf("expected tisch B in states")
	}
	if gotB.Name != "Tisch B" {
		t.Errorf("expected name 'Tisch B', got %q", gotB.Name)
	}
	if !reflect.DeepEqual(gotB.Session, kasse.TischSession{}) {
		t.Errorf("expected zero-value session for tisch B, got %+v", gotB.Session)
	}

	if _, ok := states[tischDeleted]; ok {
		t.Errorf("deleted tisch must not appear in states")
	}
	if _, ok := states[999999]; ok {
		t.Errorf("unknown tisch must not appear in states")
	}
}

func TestWriteEventWithDruckauftraege_CommitsEventAndAuftrag(t *testing.T) {
	userID, ksNr, repo, teardown := setup(t)
	defer teardown(t)

	tischID, err := createTisch(repo.db, "Tisch 1")
	if err != nil {
		t.Fatalf("Failed to create tisch: %v", err)
	}

	subject := kasse.TischSessionSubject(ksNr, tischID)
	data := validBestellungData("b0000000-0000-0000-0000-000000000001", "p0000000-0000-0000-0000-000000000001", 350, 2)
	e := newTestEvent(userID, "bestellung-aufgenommen:v1", subject, 1, data)

	eventID, err := repo.WriteEventWithDruckauftraege(context.Background(), e, kasse.StreamTypeTischSession, ksNr,
		func(stored event.Event) []druckauftrag_repo.NeuerDruckauftrag {
			return []druckauftrag_repo.NeuerDruckauftrag{{
				ZielIP:   "192.168.1.50",
				Payload:  "AAA=",
				BonArt:   "arbeitsbon",
				Referenz: fmt.Sprintf("bestellung-aufgenommen:%d", stored.ID),
			}}
		})
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if eventID == 0 {
		t.Fatalf("Expected valid event ID, got %d", eventID)
	}

	events, err := repo.ReadEventsBySubject(context.Background(), subject)
	if err != nil {
		t.Fatalf("Expected no read error, got %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("Expected 1 persisted event, got %d", len(events))
	}

	// The druckauftrag references the generated event ID.
	var referenz string
	err = repo.db.QueryRow("SELECT referenz FROM druckauftraege").Scan(&referenz)
	if err != nil {
		t.Fatalf("Expected 1 persisted druckauftrag, got error %v", err)
	}
	if want := fmt.Sprintf("bestellung-aufgenommen:%d", eventID); referenz != want {
		t.Fatalf("Expected referenz %q, got %q", want, referenz)
	}
}

func TestWriteEventWithDruckauftraege_RollsBackEventOnAuftragError(t *testing.T) {
	userID, ksNr, repo, teardown := setup(t)
	defer teardown(t)

	tischID, err := createTisch(repo.db, "Tisch 1")
	if err != nil {
		t.Fatalf("Failed to create tisch: %v", err)
	}

	subject := kasse.TischSessionSubject(ksNr, tischID)
	data := validBestellungData("b0000000-0000-0000-0000-000000000001", "p0000000-0000-0000-0000-000000000001", 350, 2)
	e := newTestEvent(userID, "bestellung-aufgenommen:v1", subject, 1, data)

	// "ungueltig" violates the bon_art CHECK constraint, so the auftrag INSERT fails.
	_, err = repo.WriteEventWithDruckauftraege(context.Background(), e, kasse.StreamTypeTischSession, ksNr,
		func(_ event.Event) []druckauftrag_repo.NeuerDruckauftrag {
			return []druckauftrag_repo.NeuerDruckauftrag{{
				ZielIP:   "192.168.1.50",
				Payload:  "AAA=",
				BonArt:   "ungueltig",
				Referenz: "bestellung-aufgenommen:rollback",
			}}
		})
	if err == nil {
		t.Fatal("Expected error from invalid druckauftrag, got nil")
	}

	// The event must be rolled back together with the failed auftrag.
	events, err := repo.ReadEventsBySubject(context.Background(), subject)
	if err != nil {
		t.Fatalf("Expected no read error, got %v", err)
	}
	if len(events) != 0 {
		t.Fatalf("Expected event to be rolled back, found %d events", len(events))
	}

	// The projection must not have been updated either.
	session, err := repo.ReadTischSession(context.Background(), subject)
	if err != nil {
		t.Fatalf("Expected no read error, got %v", err)
	}
	if session.LastEventID != 0 {
		t.Fatalf("Expected no tisch session projection, got LastEventID %d", session.LastEventID)
	}
}

// Jeder signaturpflichtige Vorgang erzeugt im selben Commit genau einen offenen
// Signaturauftrag mit processType-/processData-Snapshot aus der fiskalischen
// Projektion — auch ohne TSE-Konfiguration (die Test-DB hat keine). Nicht
// signaturpflichtige Vorgänge erhalten keinen Auftrag.
func TestWriteEvent_SignaturpflichtigErzeugtOffenenAuftrag(t *testing.T) {
	userID, ksNr, repo, teardown := setup(t)
	defer teardown(t)

	tischID, err := createTisch(repo.db, "Tisch 1")
	if err != nil {
		t.Fatalf("Failed to create tisch: %v", err)
	}

	subject := kasse.TischSessionSubject(ksNr, tischID)

	bestellung := newTestEvent(userID, "bestellung-aufgenommen:v1", subject, 1, validBestellungData("b0000000-0000-0000-0000-000000000001", "10000000-0000-4000-8000-000000000001", 350, 1))
	bestellungID, err := repo.WriteEvent(context.Background(), bestellung, kasse.StreamTypeTischSession, ksNr)
	if err != nil {
		t.Fatalf("Failed to write bestellung event: %v", err)
	}

	zahlung := newTestEvent(userID, "zahlung-kassiert:v1", subject, 2, validZahlungData("10000000-0000-4000-8000-000000000001", 1, 350))
	zahlungID, err := repo.WriteEvent(context.Background(), zahlung, kasse.StreamTypeTischSession, ksNr)
	if err != nil {
		t.Fatalf("Failed to write zahlung event: %v", err)
	}

	assertOffenerAuftrag := func(eventID int, wantProcessType, wantProcessData string) {
		t.Helper()
		var status, processType, processData, txID string
		err := repo.db.QueryRow(
			"SELECT status, process_type, process_data, tx_id FROM tse_signaturauftraege WHERE event_id = $1",
			eventID,
		).Scan(&status, &processType, &processData, &txID)
		if err != nil {
			t.Fatalf("Expected auftrag for event %d, got error %v", eventID, err)
		}
		if status != "offen" {
			t.Fatalf("Expected status offen for event %d, got %q", eventID, status)
		}
		if processType != wantProcessType {
			t.Fatalf("Expected process_type %q for event %d, got %q", wantProcessType, eventID, processType)
		}
		if processData != wantProcessData {
			t.Fatalf("Expected process_data %q for event %d, got %q", wantProcessData, eventID, processData)
		}
		if txID == "" {
			t.Fatalf("Expected tx_id for event %d", eventID)
		}
	}

	assertOffenerAuftrag(bestellungID, "Bestellung-V1", `1;"Bier 0.5L";3.50`)
	assertOffenerAuftrag(zahlungID, "Kassenbeleg-V1", "Beleg^3.50_0.00_0.00_0.00_0.00^3.50:Bar")

	var gesamt int
	if err := repo.db.QueryRow("SELECT COUNT(*) FROM tse_signaturauftraege").Scan(&gesamt); err != nil {
		t.Fatalf("Failed to count auftraege: %v", err)
	}
	if gesamt != 2 {
		t.Fatalf("Expected exactly 2 auftraege (bestellung + zahlung), got %d", gesamt)
	}
}

func TestWriteUmbuchung_CommitsBothEventsAndProjections(t *testing.T) {
	userID, ksNr, repo, teardown := setup(t)
	defer teardown(t)

	quellTischID, err := createTisch(repo.db, "Tisch Quelle")
	if err != nil {
		t.Fatalf("Failed to create source tisch: %v", err)
	}
	zielTischID, err := createTisch(repo.db, "Tisch Ziel")
	if err != nil {
		t.Fatalf("Failed to create target tisch: %v", err)
	}

	quellSubject := kasse.TischSessionSubject(ksNr, quellTischID)
	zielSubject := kasse.TischSessionSubject(ksNr, zielTischID)
	quellPositionID := "10000000-0000-0000-0000-000000000001"

	quellBestellung := newTestEvent(userID, "bestellung-aufgenommen:v1", quellSubject, 1, validBestellungData("b0000000-0000-0000-0000-000000000001", quellPositionID, 350, 2))
	if _, err := repo.WriteEvent(context.Background(), quellBestellung, kasse.StreamTypeTischSession, ksNr); err != nil {
		t.Fatalf("Failed to write source bestellung: %v", err)
	}

	umbuchPosition := kasse.Position{
		PositionID:       quellPositionID,
		VarianteID:       1,
		ProduktName:      "Bier",
		VarianteName:     "0.5L",
		Kategorie:        "getraenk",
		Steuersatz:       "regel",
		EinzelpreisCents: 350,
		Menge:            2,
	}

	quellEvent, zielEvent, err := kasse.NewBestellungUmgebuchtEvents(ksNr, quellTischID, zielTischID, userID, "nico", []kasse.Position{umbuchPosition}, 700, "Umbuchung auf Tisch Ziel", "Umbuchung von Tisch Quelle", "")
	if err != nil {
		t.Fatalf("Failed to build umbuchung events: %v", err)
	}
	quellEvent.Version = 2
	zielEvent.Version = 1

	if err := repo.WriteUmbuchung(context.Background(), quellEvent, zielEvent, ksNr); err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	quellEvents, err := repo.ReadEventsBySubject(context.Background(), quellSubject)
	if err != nil {
		t.Fatalf("Expected no read error for source subject, got %v", err)
	}
	if len(quellEvents) != 2 {
		t.Fatalf("Expected 2 source events, got %d", len(quellEvents))
	}
	if quellEvents[1].Type != string(kasse.EventTypeBestellungUmgebuchtV1) {
		t.Fatalf("Expected source event type %q, got %q", kasse.EventTypeBestellungUmgebuchtV1, quellEvents[1].Type)
	}

	zielEvents, err := repo.ReadEventsBySubject(context.Background(), zielSubject)
	if err != nil {
		t.Fatalf("Expected no read error for target subject, got %v", err)
	}
	if len(zielEvents) != 1 {
		t.Fatalf("Expected 1 target event, got %d", len(zielEvents))
	}
	if zielEvents[0].Type != string(kasse.EventTypeBestellungUmgebuchtV1) {
		t.Fatalf("Expected target event type %q, got %q", kasse.EventTypeBestellungUmgebuchtV1, zielEvents[0].Type)
	}

	quellState, err := repo.ReadTischSession(context.Background(), quellSubject)
	if err != nil {
		t.Fatalf("Expected no source state read error, got %v", err)
	}
	if quellState.SaldoCents != 0 {
		t.Fatalf("Expected source saldo 0, got %d", quellState.SaldoCents)
	}
	if len(quellState.UnbezahltePositionen) != 0 {
		t.Fatalf("Expected no source unbezahlte positionen, got %d", len(quellState.UnbezahltePositionen))
	}

	zielState, err := repo.ReadTischSession(context.Background(), zielSubject)
	if err != nil {
		t.Fatalf("Expected no target state read error, got %v", err)
	}
	if zielState.SaldoCents != 700 {
		t.Fatalf("Expected target saldo 700, got %d", zielState.SaldoCents)
	}
	if len(zielState.UnbezahltePositionen) != 1 {
		t.Fatalf("Expected 1 target unbezahlte position, got %d", len(zielState.UnbezahltePositionen))
	}
	if zielState.UnbezahltePositionen[0].EinzelpreisCents != 350 {
		t.Fatalf("Expected target einzelpreis 350, got %d", zielState.UnbezahltePositionen[0].EinzelpreisCents)
	}
	if zielState.UnbezahltePositionen[0].Menge != 2 {
		t.Fatalf("Expected target menge 2, got %d", zielState.UnbezahltePositionen[0].Menge)
	}
}

func TestWriteUmbuchung_RollsBackWhenTargetWriteFails(t *testing.T) {
	userID, ksNr, repo, teardown := setup(t)
	defer teardown(t)

	quellTischID, err := createTisch(repo.db, "Tisch Quelle")
	if err != nil {
		t.Fatalf("Failed to create source tisch: %v", err)
	}
	zielTischID, err := createTisch(repo.db, "Tisch Ziel")
	if err != nil {
		t.Fatalf("Failed to create target tisch: %v", err)
	}

	quellSubject := kasse.TischSessionSubject(ksNr, quellTischID)
	zielSubject := kasse.TischSessionSubject(ksNr, zielTischID)
	quellPositionID := "10000000-0000-0000-0000-000000000002"

	quellBestellung := newTestEvent(userID, "bestellung-aufgenommen:v1", quellSubject, 1, validBestellungData("b0000000-0000-0000-0000-000000000001", quellPositionID, 350, 2))
	if _, err := repo.WriteEvent(context.Background(), quellBestellung, kasse.StreamTypeTischSession, ksNr); err != nil {
		t.Fatalf("Failed to write source bestellung: %v", err)
	}

	umbuchPosition := kasse.Position{
		PositionID:       quellPositionID,
		VarianteID:       1,
		ProduktName:      "Bier",
		VarianteName:     "0.5L",
		Kategorie:        "getraenk",
		Steuersatz:       "regel",
		EinzelpreisCents: 350,
		Menge:            2,
	}

	stornierungEvent, err := kasse.NewStornierungErteiltEvent(quellSubject, userID, "nico", "11111111-1111-1111-1111-111111111111", []kasse.Position{umbuchPosition}, 700, "Umbuchung")
	if err != nil {
		t.Fatalf("Failed to build stornierung event: %v", err)
	}
	stornierungEvent.Version = 2

	ungueltigesZielEvent := newTestEvent(userID, "unknown-event:v1", zielSubject, 1, map[string]any{"any": "value"})

	err = repo.WriteUmbuchung(context.Background(), stornierungEvent, ungueltigesZielEvent, ksNr)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	quellEvents, err := repo.ReadEventsBySubject(context.Background(), quellSubject)
	if err != nil {
		t.Fatalf("Expected no read error for source subject, got %v", err)
	}
	if len(quellEvents) != 1 {
		t.Fatalf("Expected source rollback (1 event), got %d", len(quellEvents))
	}

	zielEvents, err := repo.ReadEventsBySubject(context.Background(), zielSubject)
	if err != nil {
		t.Fatalf("Expected no read error for target subject, got %v", err)
	}
	if len(zielEvents) != 0 {
		t.Fatalf("Expected target rollback (0 events), got %d", len(zielEvents))
	}

	quellState, err := repo.ReadTischSession(context.Background(), quellSubject)
	if err != nil {
		t.Fatalf("Expected no source state read error, got %v", err)
	}
	if quellState.SaldoCents != 700 {
		t.Fatalf("Expected source saldo 700 after rollback, got %d", quellState.SaldoCents)
	}
	if len(quellState.UnbezahltePositionen) != 1 {
		t.Fatalf("Expected source unbezahlte positionen to remain 1, got %d", len(quellState.UnbezahltePositionen))
	}

	zielState, err := repo.ReadTischSession(context.Background(), zielSubject)
	if err != nil {
		t.Fatalf("Expected no target state read error, got %v", err)
	}
	if zielState.LastEventID != 0 {
		t.Fatalf("Expected no target projection update, got LastEventID %d", zielState.LastEventID)
	}
}

func TestWriteUmbuchung_OCCConflictRollsBackBothSides(t *testing.T) {
	userID, ksNr, repo, teardown := setup(t)
	defer teardown(t)

	quellTischID, err := createTisch(repo.db, "Tisch Quelle")
	if err != nil {
		t.Fatalf("Failed to create source tisch: %v", err)
	}
	zielTischID, err := createTisch(repo.db, "Tisch Ziel")
	if err != nil {
		t.Fatalf("Failed to create target tisch: %v", err)
	}

	quellSubject := kasse.TischSessionSubject(ksNr, quellTischID)
	zielSubject := kasse.TischSessionSubject(ksNr, zielTischID)
	quellPositionID := "10000000-0000-0000-0000-000000000003"
	zielPositionID := "10000000-0000-0000-0000-000000000004"

	quellBestellung := newTestEvent(userID, "bestellung-aufgenommen:v1", quellSubject, 1, validBestellungData("b0000000-0000-0000-0000-000000000001", quellPositionID, 350, 2))
	if _, err := repo.WriteEvent(context.Background(), quellBestellung, kasse.StreamTypeTischSession, ksNr); err != nil {
		t.Fatalf("Failed to write source bestellung: %v", err)
	}
	zielBestellung := newTestEvent(userID, "bestellung-aufgenommen:v1", zielSubject, 1, validBestellungData("b0000000-0000-0000-0000-000000000002", zielPositionID, 100, 1))
	if _, err := repo.WriteEvent(context.Background(), zielBestellung, kasse.StreamTypeTischSession, ksNr); err != nil {
		t.Fatalf("Failed to write target bestellung: %v", err)
	}

	// Zahlung für Quelltisch (Version 2), damit die Stornierung im WriteUmbuchung
	// nicht an gesamt_zahlungen_cents >= 0 scheitert (CHECK-Constraint v0.14.0).
	quellZahlung := newTestEvent(userID, "zahlung-kassiert:v1", quellSubject, 2, validZahlungData(quellPositionID, 2, 700))
	if _, err := repo.WriteEvent(context.Background(), quellZahlung, kasse.StreamTypeTischSession, ksNr); err != nil {
		t.Fatalf("Failed to write source zahlung: %v", err)
	}

	umbuchPosition := kasse.Position{
		PositionID:       quellPositionID,
		VarianteID:       1,
		ProduktName:      "Bier",
		VarianteName:     "0.5L",
		Kategorie:        "getraenk",
		Steuersatz:       "regel",
		EinzelpreisCents: 350,
		Menge:            2,
	}

	stornierungEvent, err := kasse.NewStornierungErteiltEvent(quellSubject, userID, "nico", "11111111-1111-1111-1111-111111111111", []kasse.Position{umbuchPosition}, 700, "Umbuchung")
	if err != nil {
		t.Fatalf("Failed to build stornierung event: %v", err)
	}
	stornierungEvent.Version = 3 // quell now has 2 events (bestellung v1 + zahlung v2)

	bestellungEvent, err := kasse.NewBestellungAufgenommenEvent(zielSubject, userID, "nico", "b0000000-0000-0000-0000-000000000003", []kasse.Position{umbuchPosition}, "Umbuchung")
	if err != nil {
		t.Fatalf("Failed to build bestellung event: %v", err)
	}
	bestellungEvent.Version = 1 // conflicts with existing target event version 1

	err = repo.WriteUmbuchung(context.Background(), stornierungEvent, bestellungEvent, ksNr)
	if !errors.Is(err, dbpkg.ErrAlreadyExists) {
		t.Fatalf("Expected ErrAlreadyExists conflict, got %v", err)
	}

	quellEvents, err := repo.ReadEventsBySubject(context.Background(), quellSubject)
	if err != nil {
		t.Fatalf("Expected no read error for source subject, got %v", err)
	}
	if len(quellEvents) != 2 {
		t.Fatalf("Expected source rollback (2 events: bestellung+zahlung), got %d", len(quellEvents))
	}

	zielEvents, err := repo.ReadEventsBySubject(context.Background(), zielSubject)
	if err != nil {
		t.Fatalf("Expected no read error for target subject, got %v", err)
	}
	if len(zielEvents) != 1 {
		t.Fatalf("Expected target unchanged (1 event), got %d", len(zielEvents))
	}

	quellState, err := repo.ReadTischSession(context.Background(), quellSubject)
	if err != nil {
		t.Fatalf("Expected no source state read error, got %v", err)
	}
	if quellState.SaldoCents != 0 {
		t.Fatalf("Expected source saldo 0 after zahlung rollback, got %d", quellState.SaldoCents)
	}

	zielState, err := repo.ReadTischSession(context.Background(), zielSubject)
	if err != nil {
		t.Fatalf("Expected no target state read error, got %v", err)
	}
	if zielState.SaldoCents != 100 {
		t.Fatalf("Expected target saldo 100 unchanged, got %d", zielState.SaldoCents)
	}
}

func TestReadEventsBySubject(t *testing.T) {
	userID, ksNr, repo, teardown := setup(t)
	defer teardown(t)

	subject1 := kasse.TischSessionSubject(ksNr, 1)
	subject2 := kasse.TischSessionSubject(ksNr, 42)

	event1 := newTestEvent(userID, "bestellung-aufgenommen:v1", subject1, 1, map[string]any{"k": "v"})
	event2 := newTestEvent(userID, "bestellung-aufgenommen:v1", subject2, 1, map[string]any{"k": "v"})
	_, _ = insertEventRaw(repo.db, event1, ksNr)
	_, _ = insertEventRaw(repo.db, event2, ksNr)

	events, err := repo.ReadEventsBySubject(context.Background(), subject2)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("Expected 1 event, got %d", len(events))
	}
	if events[0].Subject != subject2 {
		t.Fatalf("Expected subject %s, got %s", subject2, events[0].Subject)
	}
}

func TestGetMaxVersion(t *testing.T) {
	userID, ksNr, repo, teardown := setup(t)
	defer teardown(t)

	subject1 := kasse.TischSessionSubject(ksNr, 1)
	subject2 := kasse.TischSessionSubject(ksNr, 2)

	// No events yet
	version, err := repo.GetMaxVersion(context.Background(), subject1)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if version != 0 {
		t.Fatalf("Expected version 0 for empty subject, got %d", version)
	}

	// Add events
	e1 := newTestEvent(userID, "bestellung-aufgenommen:v1", subject1, 1, map[string]any{"order": 1})
	e2 := newTestEvent(userID, "bestellung-aufgenommen:v1", subject1, 2, map[string]any{"order": 2})
	e3 := newTestEvent(userID, "bestellung-aufgenommen:v1", subject2, 1, map[string]any{"order": 3})

	_, _ = insertEventRaw(repo.db, e1, ksNr)
	_, _ = insertEventRaw(repo.db, e2, ksNr)
	_, _ = insertEventRaw(repo.db, e3, ksNr)

	// Should return max version for subject1
	version, err = repo.GetMaxVersion(context.Background(), subject1)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if version != 2 {
		t.Fatalf("Expected version 2, got %d", version)
	}

	// Should return max version for subject2
	version, err = repo.GetMaxVersion(context.Background(), subject2)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if version != 1 {
		t.Fatalf("Expected version 1, got %d", version)
	}
}

func TestGetKassenbestand_DirektverkaufIncreasesThenStornoDecreases(t *testing.T) {
	userID, ksNr, repo, teardown := setup(t)
	defer teardown(t)

	subject := kasse.DirektverkaufSubject(ksNr, "verkauf-1")

	verkauf := newTestEvent(userID, "direktverkauf-getaetigt:v1", subject, 1, validDirektverkaufData("verkauf-1", 1200))
	if _, err := insertEventRaw(repo.db, verkauf, ksNr); err != nil {
		t.Fatalf("Failed to insert direktverkauf-getaetigt event: %v", err)
	}

	bestand, err := repo.GetKassenbestand(context.Background(), ksNr)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if bestand != 1200 {
		t.Fatalf("Expected kassenbestand 1200 after direktverkauf, got %d", bestand)
	}

	storno := newTestEvent(userID, "direktverkauf-storniert:v1", subject, 2, validDirektverkaufStornoData("verkauf-1", 500))
	if _, err := insertEventRaw(repo.db, storno, ksNr); err != nil {
		t.Fatalf("Failed to insert direktverkauf-storniert event: %v", err)
	}

	bestand, err = repo.GetKassenbestand(context.Background(), ksNr)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if bestand != 700 {
		t.Fatalf("Expected kassenbestand 700 after direktverkauf-storno, got %d", bestand)
	}
}

func TestGetKassenbestand_WarenruecknahmeDecreasesKorrekturDoesNot(t *testing.T) {
	userID, ksNr, repo, teardown := setup(t)
	defer teardown(t)

	tischSubject := kasse.TischSessionSubject(ksNr, 1)

	zahlung := newTestEvent(userID, "zahlung-kassiert:v1", tischSubject, 1, validZahlungData("p0000000-0000-0000-0000-000000000001", 1, 1000))
	if _, err := insertEventRaw(repo.db, zahlung, ksNr); err != nil {
		t.Fatalf("Failed to insert zahlung-kassiert event: %v", err)
	}

	bestand, err := repo.GetKassenbestand(context.Background(), ksNr)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if bestand != 1000 {
		t.Fatalf("Expected kassenbestand 1000 after zahlung, got %d", bestand)
	}

	// Geldneutrale Korrektur verändert den Kassenbestand nicht.
	korrektur := newTestEvent(userID, "bestellung-korrigiert:v1", tischSubject, 2, validKorrekturData(400))
	if _, err := insertEventRaw(repo.db, korrektur, ksNr); err != nil {
		t.Fatalf("Failed to insert bestellung-korrigiert event: %v", err)
	}

	bestand, err = repo.GetKassenbestand(context.Background(), ksNr)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if bestand != 1000 {
		t.Fatalf("Expected kassenbestand 1000 after geldneutrale Korrektur, got %d", bestand)
	}

	// Kassenwirksame Warenrücknahme gibt Bargeld zurück und mindert den Bestand.
	storno := newTestEvent(userID, "stornierung-erteilt:v1", tischSubject, 3, validStornierungData(300))
	if _, err := insertEventRaw(repo.db, storno, ksNr); err != nil {
		t.Fatalf("Failed to insert stornierung-erteilt event: %v", err)
	}

	bestand, err = repo.GetKassenbestand(context.Background(), ksNr)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if bestand != 700 {
		t.Fatalf("Expected kassenbestand 700 after Warenrücknahme, got %d", bestand)
	}
}

// Die Eröffnung ist atomar: Scheitert der Event-Write (hier: build-Fehler),
// bleibt keine offene Sitzung ohne Eröffnungs-Event zurück.
func TestEroeffneKassensitzung_RollbackBeiEventFehler(t *testing.T) {
	userID, _, repo, teardown := setup(t)
	defer teardown(t)

	var vorher int
	if err := repo.db.QueryRow("SELECT COUNT(*) FROM kassensitzungen").Scan(&vorher); err != nil {
		t.Fatalf("count kassensitzungen: %v", err)
	}

	_, err := repo.EroeffneKassensitzung(context.Background(), time.Now().UTC(), "Testsitzung", func(zNr int) (event.Event, error) {
		return event.Event{}, errors.New("event bauen fehlgeschlagen")
	})
	if err == nil {
		t.Fatal("expected error from build callback")
	}

	var nachher int
	if err := repo.db.QueryRow("SELECT COUNT(*) FROM kassensitzungen").Scan(&nachher); err != nil {
		t.Fatalf("count kassensitzungen: %v", err)
	}
	if nachher != vorher {
		t.Fatalf("expected rollback of kassensitzungen row: vorher %d, nachher %d", vorher, nachher)
	}
	_ = userID
}

func TestEroeffneKassensitzung_SchreibtEntitaetUndEventAtomar(t *testing.T) {
	userID, alteKsNr, repo, teardown := setup(t)
	defer teardown(t)

	// setup() legt bereits eine offene Sitzung an; die Einzigkeits-Invariante
	// (partieller Unique-Index) erlaubt nur eine offene Sitzung gleichzeitig.
	if _, err := repo.db.Exec("UPDATE kassensitzungen SET status = 'abgeschlossen' WHERE z_nr = $1", alteKsNr); err != nil {
		t.Fatalf("close previous kassensitzung: %v", err)
	}

	zNr, err := repo.EroeffneKassensitzung(context.Background(), time.Now().UTC(), "Testsitzung", func(zNr int) (event.Event, error) {
		evt := newTestEvent(userID, "kassensitzung-eroeffnet:v1", kasse.KassensitzungSubject(zNr), 1, validEroeffnungData(userID, 10000))
		return evt, nil
	})
	if err != nil {
		t.Fatalf("EroeffneKassensitzung: %v", err)
	}

	var status string
	if err := repo.db.QueryRow("SELECT status FROM kassensitzungen WHERE z_nr = $1", zNr).Scan(&status); err != nil {
		t.Fatalf("read kassensitzung: %v", err)
	}
	if status != "offen" {
		t.Fatalf("expected status offen, got %q", status)
	}

	events, err := repo.ReadEventsBySubject(context.Background(), kasse.KassensitzungSubject(zNr))
	if err != nil {
		t.Fatalf("read events: %v", err)
	}
	if len(events) != 1 || events[0].Type != "kassensitzung-eroeffnet:v1" {
		t.Fatalf("expected exactly the eroeffnet event, got %v", events)
	}

	// Der Signaturauftrag der Eröffnung (Anfangsbestand > 0) entsteht im selben
	// Commit aus der fiskalischen Projektion.
	var auftraege int
	if err := repo.db.QueryRow(
		"SELECT COUNT(*) FROM tse_signaturauftraege a JOIN kassenjournal e ON e.id = a.event_id WHERE e.subject = $1",
		kasse.KassensitzungSubject(zNr),
	).Scan(&auftraege); err != nil {
		t.Fatalf("count signaturauftraege: %v", err)
	}
	if auftraege != 1 {
		t.Fatalf("expected 1 signaturauftrag, got %d", auftraege)
	}
}

// Kein Event darf in eine abgeschlossene Kassensitzung geschrieben werden —
// der Status-Guard in writeEventInTx lehnt den Write mit Zeilensperre ab.
func TestWriteEvent_InGeschlosseneKassensitzungWirdAbgelehnt(t *testing.T) {
	userID, ksNr, repo, teardown := setup(t)
	defer teardown(t)

	tischID, err := createTisch(repo.db, "Tisch 1")
	if err != nil {
		t.Fatalf("Failed to create tisch: %v", err)
	}

	if _, err := repo.db.Exec("UPDATE kassensitzungen SET status = 'abgeschlossen' WHERE z_nr = $1", ksNr); err != nil {
		t.Fatalf("close kassensitzung: %v", err)
	}

	subject := kasse.TischSessionSubject(ksNr, tischID)
	e := newTestEvent(userID, "bestellung-aufgenommen:v1", subject, 1, validBestellungData("b0000000-0000-0000-0000-000000000001", "p0000000-0000-0000-0000-000000000001", 350, 1))

	_, err = repo.WriteEvent(context.Background(), e, kasse.StreamTypeTischSession, ksNr)
	if !errors.Is(err, ErrKassensitzungNichtOffen) {
		t.Fatalf("expected ErrKassensitzungNichtOffen, got %v", err)
	}

	events, err := repo.ReadEventsBySubject(context.Background(), subject)
	if err != nil {
		t.Fatalf("read events: %v", err)
	}
	if len(events) != 0 {
		t.Fatalf("expected no persisted events in closed session, got %d", len(events))
	}
}

// Im Zwischenstatus 'wird_abgeschlossen' lehnt der Status-Guard Buchungs-Events ab,
// lässt aber die Abschluss-Events (hier: Kassensturz) durch — genau das erlaubt dem
// Tagesabschluss, in die noch nicht endgültig geschlossene Sitzung zu schreiben.
func TestWriteEvent_ZwischenstatusWirdAbgeschlossen(t *testing.T) {
	userID, ksNr, repo, teardown := setup(t)
	defer teardown(t)

	tischID, err := createTisch(repo.db, "Tisch 1")
	if err != nil {
		t.Fatalf("Failed to create tisch: %v", err)
	}

	if _, err := repo.db.Exec("UPDATE kassensitzungen SET status = 'wird_abgeschlossen' WHERE z_nr = $1", ksNr); err != nil {
		t.Fatalf("set barrier status: %v", err)
	}

	// Buchungs-Event wird abgelehnt.
	tischSubject := kasse.TischSessionSubject(ksNr, tischID)
	bestellung := newTestEvent(userID, "bestellung-aufgenommen:v1", tischSubject, 1, validBestellungData("b0000000-0000-0000-0000-000000000001", "p0000000-0000-0000-0000-000000000001", 350, 1))
	if _, err := repo.WriteEvent(context.Background(), bestellung, kasse.StreamTypeTischSession, ksNr); !errors.Is(err, ErrKassensitzungNichtOffen) {
		t.Fatalf("expected booking to be rejected in wird_abgeschlossen, got %v", err)
	}

	// Abschluss-Event wird zugelassen.
	ksSubject := kasse.KassensitzungSubject(ksNr)
	kassensturz := newTestEvent(userID, "kassensturz-durchgefuehrt:v1", ksSubject, 1, map[string]any{
		"sollBestandCents": 0,
		"istBestandCents":  0,
		"differenzCents":   0,
		"durchgefuehrtVon": userID,
	})
	if _, err := repo.WriteEvent(context.Background(), kassensturz, kasse.StreamTypeKassensitzung, ksNr); err != nil {
		t.Fatalf("expected kassensturz to pass guard in wird_abgeschlossen, got %v", err)
	}

	events, err := repo.ReadEventsBySubject(context.Background(), tischSubject)
	if err != nil {
		t.Fatalf("read events: %v", err)
	}
	if len(events) != 0 {
		t.Fatalf("expected no persisted booking events, got %d", len(events))
	}
}

// validEroeffnungData returns valid kassensitzung-eroeffnet:v1 event data.
func validEroeffnungData(userID, betragCents int) map[string]any {
	return map[string]any{
		"datum":        "2026-07-03",
		"bezeichnung":  "Testsitzung",
		"betragCents":  betragCents,
		"eroeffnetVon": userID,
	}
}

// validDifferenzData returns valid differenz-soll-ist-gebucht:v1 event data.
// betragCents ist Soll − Ist (positiv = Fehlbetrag, negativ = Überschuss).
func validDifferenzData(userID, betragCents int) map[string]any {
	return map[string]any{
		"betragCents": betragCents,
		"gebuchtVon":  userID,
	}
}

// Nach der Differenzbuchung muss der Soll-Bestand dem gezählten Ist-Bestand
// entsprechen: Ein erneuter Kassensturz (z. B. beim Abschluss-Retry nach
// Teilfehler) berechnet dann Differenz 0 und bucht kein zweites Differenz-Event.
func TestGetKassenbestand_DifferenzbuchungGleichtSollAnIstAn(t *testing.T) {
	userID, ksNr, repo, teardown := setup(t)
	defer teardown(t)

	ksSubject := kasse.KassensitzungSubject(ksNr)
	tischSubject := kasse.TischSessionSubject(ksNr, 1)

	eroeffnung := newTestEvent(userID, "kassensitzung-eroeffnet:v1", ksSubject, 1, validEroeffnungData(userID, 10000))
	if _, err := insertEventRaw(repo.db, eroeffnung, ksNr); err != nil {
		t.Fatalf("Failed to insert kassensitzung-eroeffnet event: %v", err)
	}

	zahlung := newTestEvent(userID, "zahlung-kassiert:v1", tischSubject, 1, validZahlungData("p0000000-0000-0000-0000-000000000001", 1, 5000))
	if _, err := insertEventRaw(repo.db, zahlung, ksNr); err != nil {
		t.Fatalf("Failed to insert zahlung-kassiert event: %v", err)
	}

	bestand, err := repo.GetKassenbestand(context.Background(), ksNr)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if bestand != 15000 {
		t.Fatalf("Expected kassenbestand 15000 (Soll vor Kassensturz), got %d", bestand)
	}

	// Kassensturz zählt Ist = 13000 → Differenz = Soll − Ist = +2000 (Fehlbetrag).
	differenz := newTestEvent(userID, "differenz-soll-ist-gebucht:v1", ksSubject, 2, validDifferenzData(userID, 2000))
	if _, err := insertEventRaw(repo.db, differenz, ksNr); err != nil {
		t.Fatalf("Failed to insert differenz-soll-ist-gebucht event: %v", err)
	}

	bestand, err = repo.GetKassenbestand(context.Background(), ksNr)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if bestand != 13000 {
		t.Fatalf("Expected kassenbestand 13000 (= Ist) nach Fehlbetrag-Buchung, got %d", bestand)
	}
}

func TestGetKassenbestand_UeberschussErhoehtSollAufIst(t *testing.T) {
	userID, ksNr, repo, teardown := setup(t)
	defer teardown(t)

	ksSubject := kasse.KassensitzungSubject(ksNr)

	eroeffnung := newTestEvent(userID, "kassensitzung-eroeffnet:v1", ksSubject, 1, validEroeffnungData(userID, 10000))
	if _, err := insertEventRaw(repo.db, eroeffnung, ksNr); err != nil {
		t.Fatalf("Failed to insert kassensitzung-eroeffnet event: %v", err)
	}

	// Kassensturz zählt Ist = 10500 → Differenz = Soll − Ist = −500 (Überschuss).
	differenz := newTestEvent(userID, "differenz-soll-ist-gebucht:v1", ksSubject, 2, validDifferenzData(userID, -500))
	if _, err := insertEventRaw(repo.db, differenz, ksNr); err != nil {
		t.Fatalf("Failed to insert differenz-soll-ist-gebucht event: %v", err)
	}

	bestand, err := repo.GetKassenbestand(context.Background(), ksNr)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if bestand != 10500 {
		t.Fatalf("Expected kassenbestand 10500 (= Ist) nach Überschuss-Buchung, got %d", bestand)
	}
}

func TestGetReportingStats_IncludesDirektverkaufMetrics(t *testing.T) {
	userID, ksNr, repo, teardown := setup(t)
	defer teardown(t)

	tischSubject := kasse.TischSessionSubject(ksNr, 1)
	zahlung := newTestEvent(userID, "zahlung-kassiert:v1", tischSubject, 1, validZahlungData("p0000000-0000-0000-0000-000000000001", 1, 1000))
	if _, err := insertEventRaw(repo.db, zahlung, ksNr); err != nil {
		t.Fatalf("Failed to insert zahlung-kassiert event: %v", err)
	}

	storno := newTestEvent(userID, "stornierung-erteilt:v1", tischSubject, 2, validStornierungData(300))
	if _, err := insertEventRaw(repo.db, storno, ksNr); err != nil {
		t.Fatalf("Failed to insert stornierung-erteilt event: %v", err)
	}

	dvSubject := kasse.DirektverkaufSubject(ksNr, "verkauf-2")
	dv := newTestEvent(userID, "direktverkauf-getaetigt:v1", dvSubject, 1, validDirektverkaufData("verkauf-2", 700))
	if _, err := insertEventRaw(repo.db, dv, ksNr); err != nil {
		t.Fatalf("Failed to insert direktverkauf-getaetigt event: %v", err)
	}

	dvStorno := newTestEvent(userID, "direktverkauf-storniert:v1", dvSubject, 2, validDirektverkaufStornoData("verkauf-2", 200))
	if _, err := insertEventRaw(repo.db, dvStorno, ksNr); err != nil {
		t.Fatalf("Failed to insert direktverkauf-storniert event: %v", err)
	}

	stats, err := repo.q.GetReportingStats(context.Background(), ksNr)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if stats.GesamtUmsatzCents != 1200 {
		t.Fatalf("Expected gesamt_umsatz_cents 1200, got %d", stats.GesamtUmsatzCents)
	}
	if stats.AnzahlDirektverkaeufe != 1 {
		t.Fatalf("Expected anzahl_direktverkaeufe 1, got %d", stats.AnzahlDirektverkaeufe)
	}
	if stats.DirektverkaufUmsatzCents != 500 {
		t.Fatalf("Expected direktverkauf_umsatz_cents 500, got %d", stats.DirektverkaufUmsatzCents)
	}
}

// --- Projection integration tests ---

func TestWriteEvent_WithTischSessionProjection(t *testing.T) {
	userID, ksNr, repo, teardown := setup(t)
	defer teardown(t)

	tischID, err := createTisch(repo.db, "Tisch Proj")
	if err != nil {
		t.Fatalf("Failed to create tisch: %v", err)
	}

	subject := kasse.TischSessionSubject(ksNr, tischID)
	posID := "p0000000-0000-0000-0000-000000000001"
	data := validBestellungData("b0000000-0000-0000-0000-000000000001", posID, 350, 2)
	e := newTestEvent(userID, "bestellung-aufgenommen:v1", subject, 1, data)

	eventID, err := repo.WriteEvent(context.Background(), e, kasse.StreamTypeTischSession, ksNr)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	state, err := repo.ReadTischSession(context.Background(), subject)
	if err != nil {
		t.Fatalf("Expected no error reading tisch session, got %v", err)
	}

	if state.SaldoCents != 700 {
		t.Fatalf("Expected SaldoCents 700, got %d", state.SaldoCents)
	}
	if len(state.UnbezahltePositionen) != 1 {
		t.Fatalf("Expected 1 unbezahlte position, got %d", len(state.UnbezahltePositionen))
	}
	if state.UnbezahltePositionen[0].PositionID != posID {
		t.Fatalf("Expected position ID %s, got %s", posID, state.UnbezahltePositionen[0].PositionID)
	}
	if state.UnbezahltePositionen[0].Menge != 2 {
		t.Fatalf("Expected Menge 2, got %d", state.UnbezahltePositionen[0].Menge)
	}
	if state.TischID != tischID {
		t.Fatalf("Expected TischID %d, got %d", tischID, state.TischID)
	}
	if state.KassensitzungNr != ksNr {
		t.Fatalf("Expected KassensitzungNr %d, got %d", ksNr, state.KassensitzungNr)
	}
	if state.LastEventID != eventID {
		t.Fatalf("Expected LastEventID %d, got %d", eventID, state.LastEventID)
	}
	if state.LastEventVersion != 1 {
		t.Fatalf("Expected LastEventVersion 1, got %d", state.LastEventVersion)
	}
}

func TestReadTischSession_NotFound(t *testing.T) {
	_, _, repo, teardown := setup(t)
	defer teardown(t)

	state, err := repo.ReadTischSession(context.Background(), "kassensitzung-99/tisch-99999")
	if err != nil {
		t.Fatalf("Expected no error for non-existent subject, got %v", err)
	}

	if state.SaldoCents != 0 {
		t.Fatalf("Expected SaldoCents 0, got %d", state.SaldoCents)
	}
	if state.GesamtZahlungenCents != 0 {
		t.Fatalf("Expected GesamtZahlungenCents 0, got %d", state.GesamtZahlungenCents)
	}
	if len(state.UnbezahltePositionen) != 0 {
		t.Fatalf("Expected empty unbezahlte positionen, got %d", len(state.UnbezahltePositionen))
	}
}

func TestWriteEvent_MultipleEvents_ProjectionCorrect(t *testing.T) {
	userID, ksNr, repo, teardown := setup(t)
	defer teardown(t)

	tischID, err := createTisch(repo.db, "Tisch Multi")
	if err != nil {
		t.Fatalf("Failed to create tisch: %v", err)
	}

	subject := kasse.TischSessionSubject(ksNr, tischID)
	posID := "p0000000-0000-0000-0000-000000000002"

	// Write a Bestellung (2x Bier @ 350 = 700 cents)
	bestellungData := validBestellungData("b0000000-0000-0000-0000-000000000001", posID, 350, 2)
	e1 := newTestEvent(userID, "bestellung-aufgenommen:v1", subject, 1, bestellungData)
	_, err = repo.WriteEvent(context.Background(), e1, kasse.StreamTypeTischSession, ksNr)
	if err != nil {
		t.Fatalf("Expected no error writing bestellung, got %v", err)
	}

	// Write a Zahlung (pay for 1x Bier = 350 cents)
	zahlungData := validZahlungData(posID, 1, 350)
	e2 := newTestEvent(userID, "zahlung-kassiert:v1", subject, 2, zahlungData)
	_, err = repo.WriteEvent(context.Background(), e2, kasse.StreamTypeTischSession, ksNr)
	if err != nil {
		t.Fatalf("Expected no error writing zahlung, got %v", err)
	}

	state, err := repo.ReadTischSession(context.Background(), subject)
	if err != nil {
		t.Fatalf("Expected no error reading tisch session, got %v", err)
	}

	// Saldo ist 350 (700 abzueglich 350 gezahlt)
	if state.SaldoCents != 350 {
		t.Fatalf("Expected SaldoCents 350, got %d", state.SaldoCents)
	}
	// GesamtZahlungen: 350
	if state.GesamtZahlungenCents != 350 {
		t.Fatalf("Expected GesamtZahlungenCents 350, got %d", state.GesamtZahlungenCents)
	}
	// Unbezahlt: 1 position with Menge 1 (original 2, paid 1)
	if len(state.UnbezahltePositionen) != 1 {
		t.Fatalf("Expected 1 unbezahlte position, got %d", len(state.UnbezahltePositionen))
	}
	if state.UnbezahltePositionen[0].Menge != 1 {
		t.Fatalf("Expected remaining Menge 1, got %d", state.UnbezahltePositionen[0].Menge)
	}
	if state.LastEventVersion != 2 {
		t.Fatalf("Expected LastEventVersion 2, got %d", state.LastEventVersion)
	}
}

func TestWriteEvent_InvalidData_Rollback(t *testing.T) {
	userID, ksNr, repo, teardown := setup(t)
	defer teardown(t)

	tischID, err := createTisch(repo.db, "Tisch Rollback")
	if err != nil {
		t.Fatalf("Failed to create tisch: %v", err)
	}

	subject := kasse.TischSessionSubject(ksNr, tischID)

	// Use an unknown event type that ApplyEvent cannot handle → triggers rollback
	e := newTestEvent(userID, "unknown-event:v1", subject, 1, map[string]any{"k": "v"})

	_, err = repo.WriteEvent(context.Background(), e, kasse.StreamTypeTischSession, ksNr)
	if err == nil {
		t.Fatalf("Expected error for unknown event type, got nil")
	}

	// Verify no event was written (transaction rolled back)
	events, err := repo.ReadEventsBySubject(context.Background(), subject)
	if err != nil {
		t.Fatalf("Expected no error reading events, got %v", err)
	}
	if len(events) != 0 {
		t.Fatalf("Expected 0 events after rollback, got %d", len(events))
	}

	// Verify no tisch_session was written
	state, err := repo.ReadTischSession(context.Background(), subject)
	if err != nil {
		t.Fatalf("Expected no error reading tisch session, got %v", err)
	}
	if state.SaldoCents != 0 {
		t.Fatalf("Expected SaldoCents 0 after rollback, got %d", state.SaldoCents)
	}
}

// --- Kassensitzung integration tests ---

func TestWriteEvent_KassensitzungEroeffnet(t *testing.T) {
	userID, ksNr, repo, teardown := setup(t)
	defer teardown(t)

	// The kassensitzung is already created by setup (simulating application layer).
	// WriteEvent for kassensitzung-eroeffnet:v1 only inserts the event into kassenjournal;
	// the kassensitzungen CRUD entity is managed by the application layer.
	datum := "2026-03-22"
	bezeichnung := "Sommerfest Tag 1"
	data := map[string]any{
		"datum":        datum,
		"bezeichnung":  bezeichnung,
		"eroeffnetVon": userID,
	}
	subject := kasse.KassensitzungSubject(ksNr)
	e := newTestEvent(userID, "kassensitzung-eroeffnet:v1", subject, 1, data)

	eventID, err := repo.WriteEvent(context.Background(), e, kasse.StreamTypeKassensitzung, ksNr)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if eventID == 0 {
		t.Fatalf("Expected valid event ID, got %d", eventID)
	}

	// Verify the kassensitzung still exists and is offen
	var status string
	statErr := repo.db.QueryRow("SELECT status FROM kassensitzungen WHERE z_nr = $1", ksNr).Scan(&status)
	if statErr != nil {
		t.Fatalf("Expected no error reading kassensitzung status, got %v", statErr)
	}
	if status != string(kasse.KassensitzungOffen) {
		t.Fatalf("Expected status 'offen', got %s", status)
	}
}

func TestWriteEvent_TagesabschlussErstellt(t *testing.T) {
	userID, ksNr, repo, teardown := setup(t)
	defer teardown(t)

	subject := kasse.KassensitzungSubject(ksNr)
	data := map[string]any{
		"zNr":               ksNr,
		"zeitraumVon":       time.Now().Add(-8 * time.Hour).Format(time.RFC3339),
		"zeitraumBis":       time.Now().Format(time.RFC3339),
		"umsatzGesamtCents": 15000,
		"stornierungCents":  500,
		"geldtransitCents":  0,
		"erstelltVon":       userID,
	}
	e := newTestEvent(userID, "tagesabschluss-erstellt:v1", subject, 1, data)

	_, err := repo.WriteEvent(context.Background(), e, kasse.StreamTypeKassensitzung, ksNr)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify kassensitzung is now abgeschlossen
	var status string
	statErr := repo.db.QueryRow("SELECT status FROM kassensitzungen WHERE z_nr = $1", ksNr).Scan(&status)
	if statErr != nil {
		t.Fatalf("Expected no error reading kassensitzung status, got %v", statErr)
	}
	if status != string(kasse.KassensitzungAbgeschlossen) {
		t.Fatalf("Expected status 'abgeschlossen', got %s", status)
	}
}

func TestGetOffeneKassensitzung_NoneOpen(t *testing.T) {
	_, ksNr, repo, teardown := setup(t)
	defer teardown(t)

	// Close the kassensitzung created by setup
	_, err := repo.db.Exec("UPDATE kassensitzungen SET status = $1 WHERE z_nr = $2", kasse.KassensitzungAbgeschlossen, ksNr)
	if err != nil {
		t.Fatalf("Failed to close kassensitzung: %v", err)
	}
}

// --- Projection rebuild integration tests ---

func TestRebuildAllProjections_EmptyDB(t *testing.T) {
	_, _, repo, teardown := setup(t)
	defer teardown(t)

	count, err := repo.RebuildAllProjections(context.Background())
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if count != 0 {
		t.Fatalf("Expected 0 rebuilt subjects, got %d", count)
	}
}

func TestRebuildAllProjections_RebuildsFromEvents(t *testing.T) {
	userID, ksNr, repo, teardown := setup(t)
	defer teardown(t)

	tischID, err := createTisch(repo.db, "Tisch Rebuild")
	if err != nil {
		t.Fatalf("Failed to create tisch: %v", err)
	}

	subject := kasse.TischSessionSubject(ksNr, tischID)
	posID := "p0000000-0000-0000-0000-000000000099"

	// Write events through normal path (creates projection)
	bestellungData := validBestellungData("b0000000-0000-0000-0000-000000000001", posID, 500, 3) // 3x 500 = 1500
	e1 := newTestEvent(userID, "bestellung-aufgenommen:v1", subject, 1, bestellungData)
	_, err = repo.WriteEvent(context.Background(), e1, kasse.StreamTypeTischSession, ksNr)
	if err != nil {
		t.Fatalf("Expected no error writing bestellung, got %v", err)
	}

	zahlungData := validZahlungData(posID, 1, 500) // pay 1x 500
	e2 := newTestEvent(userID, "zahlung-kassiert:v1", subject, 2, zahlungData)
	_, err = repo.WriteEvent(context.Background(), e2, kasse.StreamTypeTischSession, ksNr)
	if err != nil {
		t.Fatalf("Expected no error writing zahlung, got %v", err)
	}

	// Read expected state before rebuild
	expectedState, err := repo.ReadTischSession(context.Background(), subject)
	if err != nil {
		t.Fatalf("Expected no error reading state, got %v", err)
	}

	// Delete projection manually to simulate seed scenario
	_, err = repo.db.Exec("DELETE FROM tisch_sessions")
	if err != nil {
		t.Fatalf("Failed to delete tisch_sessions: %v", err)
	}

	// Verify projection is gone
	emptyState, err := repo.ReadTischSession(context.Background(), subject)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if emptyState.SaldoCents != 0 {
		t.Fatalf("Expected SaldoCents 0 after delete, got %d", emptyState.SaldoCents)
	}

	// Rebuild
	count, err := repo.RebuildAllProjections(context.Background())
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if count != 1 {
		t.Fatalf("Expected 1 rebuilt subject, got %d", count)
	}

	// Read rebuilt state
	rebuiltState, err := repo.ReadTischSession(context.Background(), subject)
	if err != nil {
		t.Fatalf("Expected no error reading rebuilt state, got %v", err)
	}

	// Verify it matches the expected state
	if rebuiltState.SaldoCents != expectedState.SaldoCents {
		t.Fatalf("Expected SaldoCents %d, got %d", expectedState.SaldoCents, rebuiltState.SaldoCents)
	}
	if rebuiltState.GesamtZahlungenCents != expectedState.GesamtZahlungenCents {
		t.Fatalf("Expected GesamtZahlungenCents %d, got %d", expectedState.GesamtZahlungenCents, rebuiltState.GesamtZahlungenCents)
	}
	if len(rebuiltState.UnbezahltePositionen) != len(expectedState.UnbezahltePositionen) {
		t.Fatalf("Expected %d unbezahlte positionen, got %d", len(expectedState.UnbezahltePositionen), len(rebuiltState.UnbezahltePositionen))
	}
	if rebuiltState.LastEventID != expectedState.LastEventID {
		t.Fatalf("Expected LastEventID %d, got %d", expectedState.LastEventID, rebuiltState.LastEventID)
	}
	if rebuiltState.LastEventVersion != expectedState.LastEventVersion {
		t.Fatalf("Expected LastEventVersion %d, got %d", expectedState.LastEventVersion, rebuiltState.LastEventVersion)
	}
}

func TestRebuildAllProjections_MultipleSubjects(t *testing.T) {
	userID, ksNr, repo, teardown := setup(t)
	defer teardown(t)

	tisch1ID, err := createTisch(repo.db, "Tisch R1")
	if err != nil {
		t.Fatalf("Failed to create tisch 1: %v", err)
	}
	tisch2ID, err := createTisch(repo.db, "Tisch R2")
	if err != nil {
		t.Fatalf("Failed to create tisch 2: %v", err)
	}

	subject1 := kasse.TischSessionSubject(ksNr, tisch1ID)
	subject2 := kasse.TischSessionSubject(ksNr, tisch2ID)

	// Write events via raw insert (bypassing projection, simulating events without a projection)
	e1 := newTestEvent(userID, "bestellung-aufgenommen:v1", subject1, 1,
		validBestellungData("b0000000-0000-0000-0000-000000000001", "p1-1", 200, 2)) // 400
	_, err = insertEventRaw(repo.db, e1, ksNr)
	if err != nil {
		t.Fatalf("Failed to insert event: %v", err)
	}

	e2 := newTestEvent(userID, "bestellung-aufgenommen:v1", subject2, 1,
		validBestellungData("b0000000-0000-0000-0000-000000000002", "p2-1", 300, 1)) // 300
	_, err = insertEventRaw(repo.db, e2, ksNr)
	if err != nil {
		t.Fatalf("Failed to insert event: %v", err)
	}

	// No projections exist (raw insert bypasses them)

	// Rebuild
	count, err := repo.RebuildAllProjections(context.Background())
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if count != 2 {
		t.Fatalf("Expected 2 rebuilt subjects, got %d", count)
	}

	state1, err := repo.ReadTischSession(context.Background(), subject1)
	if err != nil {
		t.Fatalf("Expected no error reading state1, got %v", err)
	}
	if state1.SaldoCents != 400 {
		t.Fatalf("Expected SaldoCents 400 for tisch1, got %d", state1.SaldoCents)
	}

	state2, err := repo.ReadTischSession(context.Background(), subject2)
	if err != nil {
		t.Fatalf("Expected no error reading state2, got %v", err)
	}
	if state2.SaldoCents != 300 {
		t.Fatalf("Expected SaldoCents 300 for tisch2, got %d", state2.SaldoCents)
	}
}

func TestRebuildAllProjections_SkipsKassensitzungSubjects(t *testing.T) {
	userID, ksNr, repo, teardown := setup(t)
	defer teardown(t)

	tischID, err := createTisch(repo.db, "Tisch SkipKS")
	if err != nil {
		t.Fatalf("Failed to create tisch: %v", err)
	}

	// Insert a kassensitzung event (should be skipped during rebuild)
	ksSubject := kasse.KassensitzungSubject(ksNr)
	ksEvent := newTestEvent(userID, "kassensitzung-eroeffnet:v1", ksSubject, 1, map[string]any{
		"datum":        "2026-03-22",
		"bezeichnung":  "Test",
		"eroeffnetVon": userID,
	})
	_, err = insertEventRaw(repo.db, ksEvent, ksNr)
	if err != nil {
		t.Fatalf("Failed to insert kassensitzung event: %v", err)
	}

	// Insert a tisch-session event
	tischSubject := kasse.TischSessionSubject(ksNr, tischID)
	tischEvent := newTestEvent(userID, "bestellung-aufgenommen:v1", tischSubject, 1,
		validBestellungData("b0000000-0000-0000-0000-000000000001", "p1-1", 200, 1))
	_, err = insertEventRaw(repo.db, tischEvent, ksNr)
	if err != nil {
		t.Fatalf("Failed to insert tisch event: %v", err)
	}

	// Rebuild should only count tisch-session subjects
	count, err := repo.RebuildAllProjections(context.Background())
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if count != 1 {
		t.Fatalf("Expected 1 rebuilt subject (only tisch-session), got %d", count)
	}

	state, err := repo.ReadTischSession(context.Background(), tischSubject)
	if err != nil {
		t.Fatalf("Expected no error reading rebuilt state, got %v", err)
	}
	if state.SaldoCents != 200 {
		t.Fatalf("Expected SaldoCents 200, got %d", state.SaldoCents)
	}
}

func TestWriteEvent_KassensitzungOtherEvent_NoCRUDChange(t *testing.T) {
	userID, ksNr, repo, teardown := setup(t)
	defer teardown(t)

	subject := kasse.KassensitzungSubject(ksNr)

	// Write a kassenbewegung event — should NOT change kassensitzungen CRUD entity status
	data := map[string]any{
		"geldtransitId": "00000000-0000-0000-0000-000000000001",
		"richtung":      "einlage",
		"betragCents":   10000,
		"kommentar":     "Wechselgeld",
		"gebuchtVon":    userID,
	}
	e := newTestEvent(userID, "geldtransit-gebucht:v1", subject, 1, data)

	eventID, err := repo.WriteEvent(context.Background(), e, kasse.StreamTypeKassensitzung, ksNr)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if eventID == 0 {
		t.Fatalf("Expected valid event ID, got %d", eventID)
	}

	// Verify kassensitzung is still offen
	var status string
	statErr := repo.db.QueryRow("SELECT status FROM kassensitzungen WHERE z_nr = $1", ksNr).Scan(&status)
	if statErr != nil {
		t.Fatalf("Expected no error reading kassensitzung status, got %v", statErr)
	}
	if status != string(kasse.KassensitzungOffen) {
		t.Fatalf("Expected status 'offen', got %s", status)
	}
}

// TestMigration03_AusgabeEntfernen belegt die Datenbereinigung der Migration
// 03_ausgabe_entfernen: Ein Kassenjournal mit einem ausgabe-bestaetigt:v1-Alt-Event
// verhindert den Projektions-Rebuild (unbekannter Event-Typ → Fehler). Nach dem
// Löschen der Alt-Events — wie es die Migration innerhalb ihrer Transaktion mit
// deaktiviertem Delete-Trigger tut — läuft RebuildAllProjections fehlerfrei durch und
// die Tisch-Zustände (Saldo, unbezahlte Positionen) sind korrekt. Der
// Append-only-Schutz besteht danach unverändert (siehe
// docs/adrs/01_ausgabe-bestaetigen.md).
func TestMigration03_AusgabeEntfernen(t *testing.T) {
	userID, ksNr, repo, teardown := setup(t)
	defer teardown(t)

	tischID, err := createTisch(repo.db, "Tisch Migration")
	if err != nil {
		t.Fatalf("Failed to create tisch: %v", err)
	}

	subject := kasse.TischSessionSubject(ksNr, tischID)
	posID := "10000000-0000-4000-8000-000000000001"

	// Bestellung (2x @500 = 1000) und Zahlung (1x @500) über den normalen Schreibpfad.
	bestellung := newTestEvent(userID, "bestellung-aufgenommen:v1", subject, 1, validBestellungData("b0000000-0000-0000-0000-000000000001", posID, 500, 2))
	if _, err := repo.WriteEvent(context.Background(), bestellung, kasse.StreamTypeTischSession, ksNr); err != nil {
		t.Fatalf("Failed to write bestellung: %v", err)
	}
	zahlung := newTestEvent(userID, "zahlung-kassiert:v1", subject, 2, validZahlungData(posID, 1, 500))
	if _, err := repo.WriteEvent(context.Background(), zahlung, kasse.StreamTypeTischSession, ksNr); err != nil {
		t.Fatalf("Failed to write zahlung: %v", err)
	}

	// Ein ausgabe-bestaetigt:v1-Alt-Event, wie es Bestandsinstanzen vor der Migration
	// im Journal hatten, roh einfügen (INSERT ist erlaubt; nur DELETE/UPDATE sind gesperrt).
	ausgabe := newTestEvent(userID, "ausgabe-bestaetigt:v1", subject, 3, map[string]any{
		"ausgabeId": "a0000000-0000-4000-8000-000000000001",
		"positionen": []map[string]any{
			{"positionId": posID, "menge": 1},
		},
		"kommentar": "",
	})
	if _, err := insertEventRaw(repo.db, ausgabe, ksNr); err != nil {
		t.Fatalf("Failed to insert ausgabe old event: %v", err)
	}

	// Solange das Alt-Event im Journal liegt, scheitert der Rebuild — genau der Grund,
	// warum die Migration die Events löschen muss (exklusive Event-Switches).
	if _, err := repo.RebuildAllProjections(context.Background()); err == nil {
		t.Fatal("Expected RebuildAllProjections to fail while ausgabe-bestaetigt:v1 events remain")
	}

	// Datenbereinigung wie in der Migration: Delete-Trigger transaktional deaktivieren,
	// Alt-Events löschen, Trigger wieder aktivieren.
	if _, err := repo.db.Exec("ALTER TABLE kassenjournal DISABLE TRIGGER kassenjournal_no_delete"); err != nil {
		t.Fatalf("Failed to disable delete trigger: %v", err)
	}
	if _, err := repo.db.Exec("DELETE FROM kassenjournal WHERE type = 'ausgabe-bestaetigt:v1'"); err != nil {
		t.Fatalf("Failed to delete ausgabe events: %v", err)
	}
	if _, err := repo.db.Exec("ALTER TABLE kassenjournal ENABLE TRIGGER kassenjournal_no_delete"); err != nil {
		t.Fatalf("Failed to re-enable delete trigger: %v", err)
	}

	// Nach der Bereinigung läuft der Rebuild fehlerfrei und die Projektion ist korrekt.
	count, err := repo.RebuildAllProjections(context.Background())
	if err != nil {
		t.Fatalf("Expected RebuildAllProjections to succeed after cleanup, got %v", err)
	}
	if count != 1 {
		t.Fatalf("Expected 1 rebuilt subject, got %d", count)
	}

	state, err := repo.ReadTischSession(context.Background(), subject)
	if err != nil {
		t.Fatalf("Expected no error reading state, got %v", err)
	}
	if state.SaldoCents != 500 {
		t.Fatalf("Expected SaldoCents 500, got %d", state.SaldoCents)
	}
	if len(state.UnbezahltePositionen) != 1 || state.UnbezahltePositionen[0].Menge != 1 {
		t.Fatalf("Expected 1 unbezahlte position with menge 1, got %+v", state.UnbezahltePositionen)
	}

	// Der Append-only-Schutz besteht nach der Migration unverändert: DELETE schlägt fehl.
	if _, err := repo.db.Exec("DELETE FROM kassenjournal WHERE subject = $1", subject); err == nil {
		t.Fatal("Expected DELETE on kassenjournal to be rejected by the append-only trigger")
	}
}
