//go:build integration

package kassenjournal_repo

import (
	"context"
	"errors"
	"testing"

	dbpkg "github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/event"
	"github.com/nicograef/jotti/backend/domain/kasse"
)

// Transaktionsverhalten der Idempotenz-Zeile (vorgang_idempotenz): Sie wird im
// selben Commit wie die Events des Vorgangs geschrieben — vor den Event-Inserts.
// Ein Primärschlüssel-Konflikt ist eindeutig eine Duplikat-Einreichung
// (ErrVorgangBereitsGebucht, nichts geschrieben); ein UNIQUE(subject, version)-
// Konflikt bleibt eindeutig ein echter OCC-Konflikt und rollt die
// Idempotenz-Zeile mit zurück.

func countRows(t *testing.T, repo Repository, query string, args ...any) int {
	t.Helper()
	var n int
	if err := repo.db.QueryRow(query, args...).Scan(&n); err != nil {
		t.Fatalf("count query %q: %v", query, err)
	}
	return n
}

func TestWriteEventMitVorgang_DuplikatSchreibtNichtsUndLiefertSentinel(t *testing.T) {
	userID, ksNr, repo, teardown := setup(t)
	defer teardown(t)

	tischID, err := createTisch(repo.db, "Tisch 1")
	if err != nil {
		t.Fatalf("Failed to create tisch: %v", err)
	}

	subject := kasse.TischSessionSubject(ksNr, tischID)
	bestellung := newTestEvent(userID, "bestellung-aufgenommen:v1", subject, 1, validBestellungData("b0000000-0000-0000-0000-000000000001", "p0000000-0000-0000-0000-000000000001", 350, 2))
	if _, err := repo.WriteEvent(context.Background(), bestellung, kasse.StreamTypeTischSession, ksNr); err != nil {
		t.Fatalf("Failed to write bestellung: %v", err)
	}

	vorgang := Vorgang{VorgangID: "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa", Art: VorgangArtZahlung, UserID: userID}
	zahlung := newTestEvent(userID, "zahlung-kassiert:v1", subject, 2, validZahlungData("p0000000-0000-0000-0000-000000000001", 2, 700))

	if _, err := repo.WriteEventMitVorgang(context.Background(), vorgang, zahlung, kasse.StreamTypeTischSession, ksNr); err != nil {
		t.Fatalf("erster Write: %v", err)
	}

	var art string
	var vorgangUserID int
	if err := repo.db.QueryRow("SELECT art, user_id FROM vorgang_idempotenz WHERE vorgang_id = $1", vorgang.VorgangID).Scan(&art, &vorgangUserID); err != nil {
		t.Fatalf("vorgang_idempotenz lesen: %v", err)
	}
	if art != VorgangArtZahlung {
		t.Errorf("erwartet art %q, gespeichert: %q", VorgangArtZahlung, art)
	}
	if vorgangUserID != userID {
		t.Errorf("erwartet user_id %d, gespeichert: %d", userID, vorgangUserID)
	}

	eventsVorher := countRows(t, repo, "SELECT COUNT(*) FROM kassenjournal")
	auftraegeVorher := countRows(t, repo, "SELECT COUNT(*) FROM tse_signaturauftraege")

	// Duplikat-Einreichung: gleiche vorgangId, nächste freie Version — der
	// PK-Konflikt greift VOR dem Event-Insert.
	zahlungRetry := newTestEvent(userID, "zahlung-kassiert:v1", subject, 3, validZahlungData("p0000000-0000-0000-0000-000000000001", 2, 700))
	_, err = repo.WriteEventMitVorgang(context.Background(), vorgang, zahlungRetry, kasse.StreamTypeTischSession, ksNr)
	if !errors.Is(err, ErrVorgangBereitsGebucht) {
		t.Fatalf("erwartet ErrVorgangBereitsGebucht, bekam: %v", err)
	}

	if n := countRows(t, repo, "SELECT COUNT(*) FROM kassenjournal"); n != eventsVorher {
		t.Errorf("Duplikat hat Events geschrieben: %d -> %d", eventsVorher, n)
	}
	if n := countRows(t, repo, "SELECT COUNT(*) FROM tse_signaturauftraege"); n != auftraegeVorher {
		t.Errorf("Duplikat hat Signaturaufträge erzeugt: %d -> %d", auftraegeVorher, n)
	}
}

func TestWriteEventMitVorgang_OCCKonfliktRolltIdempotenzZeileZurueck(t *testing.T) {
	userID, ksNr, repo, teardown := setup(t)
	defer teardown(t)

	tischID, err := createTisch(repo.db, "Tisch 1")
	if err != nil {
		t.Fatalf("Failed to create tisch: %v", err)
	}

	subject := kasse.TischSessionSubject(ksNr, tischID)
	bestellung := newTestEvent(userID, "bestellung-aufgenommen:v1", subject, 1, validBestellungData("b0000000-0000-0000-0000-000000000001", "p0000000-0000-0000-0000-000000000001", 350, 2))
	if _, err := repo.WriteEvent(context.Background(), bestellung, kasse.StreamTypeTischSession, ksNr); err != nil {
		t.Fatalf("Failed to write bestellung: %v", err)
	}

	// Neue vorgangId, aber veraltete Stream-Version (1 ist bereits belegt):
	// echter OCC-Konflikt — der Event-Insert scheitert, die zuvor geschriebene
	// Idempotenz-Zeile rollt mit zurück.
	vorgang := Vorgang{VorgangID: "cccccccc-cccc-4ccc-8ccc-cccccccccccc", Art: VorgangArtZahlung, UserID: userID}
	zahlung := newTestEvent(userID, "zahlung-kassiert:v1", subject, 1, validZahlungData("p0000000-0000-0000-0000-000000000001", 2, 700))

	_, err = repo.WriteEventMitVorgang(context.Background(), vorgang, zahlung, kasse.StreamTypeTischSession, ksNr)
	if !errors.Is(err, dbpkg.ErrAlreadyExists) {
		t.Fatalf("erwartet db.ErrAlreadyExists (OCC-Konflikt), bekam: %v", err)
	}

	if n := countRows(t, repo, "SELECT COUNT(*) FROM vorgang_idempotenz WHERE vorgang_id = $1", vorgang.VorgangID); n != 0 {
		t.Errorf("erwartet keine vorgang_idempotenz-Zeile nach Rollback, vorhanden: %d", n)
	}
	if n := countRows(t, repo, "SELECT COUNT(*) FROM kassenjournal"); n != 1 {
		t.Errorf("erwartet weiterhin genau 1 Event, gespeichert: %d", n)
	}
}

func TestWriteTischSessionEventsAtomicMitVorgang_DuplikatSchreibtKeinesDerEvents(t *testing.T) {
	userID, ksNr, repo, teardown := setup(t)
	defer teardown(t)

	tischID, err := createTisch(repo.db, "Tisch 1")
	if err != nil {
		t.Fatalf("Failed to create tisch: %v", err)
	}

	subject := kasse.TischSessionSubject(ksNr, tischID)
	bestellung := newTestEvent(userID, "bestellung-aufgenommen:v1", subject, 1, validBestellungData("b0000000-0000-0000-0000-000000000001", "p0000000-0000-0000-0000-000000000001", 350, 2))
	zahlung := newTestEvent(userID, "zahlung-kassiert:v1", subject, 2, validZahlungData("p0000000-0000-0000-0000-000000000001", 2, 700))

	vorgang := Vorgang{VorgangID: "dddddddd-dddd-4ddd-8ddd-dddddddddddd", Art: VorgangArtStornierung, UserID: userID}

	if err := repo.WriteTischSessionEventsAtomicMitVorgang(context.Background(), vorgang, []event.Event{bestellung, zahlung}, ksNr); err != nil {
		t.Fatalf("erster Write: %v", err)
	}

	eventsVorher := countRows(t, repo, "SELECT COUNT(*) FROM kassenjournal")

	retry := []event.Event{
		newTestEvent(userID, "bestellung-aufgenommen:v1", subject, 3, validBestellungData("b0000000-0000-0000-0000-000000000002", "p0000000-0000-0000-0000-000000000002", 350, 2)),
	}
	err = repo.WriteTischSessionEventsAtomicMitVorgang(context.Background(), vorgang, retry, ksNr)
	if !errors.Is(err, ErrVorgangBereitsGebucht) {
		t.Fatalf("erwartet ErrVorgangBereitsGebucht, bekam: %v", err)
	}

	if n := countRows(t, repo, "SELECT COUNT(*) FROM kassenjournal"); n != eventsVorher {
		t.Errorf("Duplikat hat Events geschrieben: %d -> %d", eventsVorher, n)
	}
}
