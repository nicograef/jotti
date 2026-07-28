//go:build integration

package kassenjournal_repo

import (
	"bytes"
	"context"
	"errors"
	"testing"

	dbpkg "github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/event"
	"github.com/nicograef/jotti/backend/domain/kasse"
	"github.com/nicograef/jotti/backend/repository/druckauftrag_repo"
)

// Transaktionsverhalten der Idempotenz-Zeile (vorgang_idempotenz): Sie wird im
// selben Commit wie die Events des Vorgangs geschrieben — vor den Event-Inserts.
// Ein Primärschlüssel-Konflikt ist eindeutig eine Zweiteinreichung desselben
// Schlüssels; erst der gespeicherte payload_hash entscheidet, ob daraus eine
// stille Erfolgsantwort (ErrVorgangBereitsGebucht) oder ein Konflikt
// (ErrVorgangDatenAbweichend) wird. Ein UNIQUE(subject, version)-Konflikt bleibt
// eindeutig ein echter OCC-Konflikt und rollt die Idempotenz-Zeile mit zurück.

// testNutzdaten steht für die pro Kommando explizit deklarierte
// Nutzdaten-Struktur, aus der der payload_hash gebildet wird.
type testNutzdaten struct {
	TischID    int   `json:"tischId"`
	Positionen []int `json:"positionen"`
}

func payloadHash(t *testing.T, art string, nutzdaten any) []byte {
	t.Helper()
	hash, err := ComputePayloadHash(art, nutzdaten)
	if err != nil {
		t.Fatalf("ComputePayloadHash: %v", err)
	}
	return hash
}

func countRows(t *testing.T, repo Repository, query string, args ...any) int {
	t.Helper()
	var n int
	if err := repo.db.QueryRow(query, args...).Scan(&n); err != nil {
		t.Fatalf("count query %q: %v", query, err)
	}
	return n
}

// ComputePayloadHash ist deterministisch über die Reihenfolge: gleiche
// Nutzdaten ergeben denselben Hash, umsortierte Positionen einen anderen. Genau
// darauf beruht die Zusage, dass eine umsortierte Einreichung als abweichend
// gilt statt als Duplikat. Die Art geht mit in den Hash: Alle Arten teilen sich
// einen Schlüsselraum, und mehrere Kommandos reichen dieselbe Feldmenge ein.
func TestComputePayloadHash_DeterministischUndReihenfolgeabhaengig(t *testing.T) {
	basis := payloadHash(t, VorgangArtZahlung, testNutzdaten{TischID: 1, Positionen: []int{10, 20}})

	if gleich := payloadHash(t, VorgangArtZahlung, testNutzdaten{TischID: 1, Positionen: []int{10, 20}}); !bytes.Equal(basis, gleich) {
		t.Errorf("gleiche Nutzdaten ergeben unterschiedliche Hashes: %x vs %x", basis, gleich)
	}
	if umsortiert := payloadHash(t, VorgangArtZahlung, testNutzdaten{TischID: 1, Positionen: []int{20, 10}}); bytes.Equal(basis, umsortiert) {
		t.Error("umsortierte Positionen ergeben denselben Hash, erwartet: abweichend")
	}
	if geaendert := payloadHash(t, VorgangArtZahlung, testNutzdaten{TischID: 2, Positionen: []int{10, 20}}); bytes.Equal(basis, geaendert) {
		t.Error("geänderter Tisch ergibt denselben Hash, erwartet: abweichend")
	}
	if andereArt := payloadHash(t, VorgangArtStornierung, testNutzdaten{TischID: 1, Positionen: []int{10, 20}}); bytes.Equal(basis, andereArt) {
		t.Error("andere Art ergibt denselben Hash, erwartet: abweichend")
	}
}

// DetermineVorgangStatus bildet Schlüssel und Nutzdaten-Hash auf die drei Ausgänge ab —
// die Vorprüfung der Kommandos vor ihrer fachlichen Validierung.
func TestDetermineVorgangStatus_DreiAusgaenge(t *testing.T) {
	userID, ksNr, repo, teardown := setup(t)
	defer teardown(t)

	tischID, err := createTisch(repo.db, "Tisch 1")
	if err != nil {
		t.Fatalf("Failed to create tisch: %v", err)
	}

	vorgangID := "eeeeeeee-eeee-4eee-8eee-eeeeeeeeeeee"
	hash := payloadHash(t, VorgangArtBestellung, testNutzdaten{TischID: tischID, Positionen: []int{1}})

	status, err := repo.DetermineVorgangStatus(context.Background(), vorgangID, hash)
	if err != nil {
		t.Fatalf("DetermineVorgangStatus (unbekannter Schlüssel): %v", err)
	}
	if status != VorgangNeu {
		t.Errorf("erwartet VorgangNeu, bekam: %v", status)
	}

	subject := kasse.TischSessionSubject(ksNr, tischID)
	bestellung := newTestEvent(userID, "bestellung-aufgenommen:v1", subject, 1, validBestellungData("b0000000-0000-0000-0000-000000000001", "p0000000-0000-0000-0000-000000000001", 350, 2))
	vorgang := Vorgang{VorgangID: vorgangID, Art: VorgangArtBestellung, UserID: userID, PayloadHash: hash}
	if _, err := repo.WriteEventMitVorgang(context.Background(), vorgang, bestellung, kasse.StreamTypeTischSession, ksNr); err != nil {
		t.Fatalf("erster Write: %v", err)
	}

	status, err = repo.DetermineVorgangStatus(context.Background(), vorgangID, hash)
	if err != nil {
		t.Fatalf("DetermineVorgangStatus (gleicher Hash): %v", err)
	}
	if status != VorgangDuplikat {
		t.Errorf("erwartet VorgangDuplikat, bekam: %v", status)
	}

	andererHash := payloadHash(t, VorgangArtBestellung, testNutzdaten{TischID: tischID, Positionen: []int{1, 2}})
	status, err = repo.DetermineVorgangStatus(context.Background(), vorgangID, andererHash)
	if err != nil {
		t.Fatalf("DetermineVorgangStatus (anderer Hash): %v", err)
	}
	if status != VorgangDatenAbweichend {
		t.Errorf("erwartet VorgangDatenAbweichend, bekam: %v", status)
	}
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

	hash := payloadHash(t, VorgangArtZahlung, testNutzdaten{TischID: tischID, Positionen: []int{1}})
	vorgang := Vorgang{VorgangID: "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa", Art: VorgangArtZahlung, UserID: userID, PayloadHash: hash}
	zahlung := newTestEvent(userID, "zahlung-kassiert:v1", subject, 2, validZahlungData("p0000000-0000-0000-0000-000000000001", 2, 700))

	if _, err := repo.WriteEventMitVorgang(context.Background(), vorgang, zahlung, kasse.StreamTypeTischSession, ksNr); err != nil {
		t.Fatalf("erster Write: %v", err)
	}

	var art string
	var vorgangUserID int
	var gespeicherterHash []byte
	if err := repo.db.QueryRow("SELECT art, user_id, payload_hash FROM vorgang_idempotenz WHERE vorgang_id = $1", vorgang.VorgangID).Scan(&art, &vorgangUserID, &gespeicherterHash); err != nil {
		t.Fatalf("vorgang_idempotenz lesen: %v", err)
	}
	if art != VorgangArtZahlung {
		t.Errorf("erwartet art %q, gespeichert: %q", VorgangArtZahlung, art)
	}
	if vorgangUserID != userID {
		t.Errorf("erwartet user_id %d, gespeichert: %d", userID, vorgangUserID)
	}
	if !bytes.Equal(gespeicherterHash, hash) {
		t.Errorf("erwartet payload_hash %x, gespeichert: %x", hash, gespeicherterHash)
	}

	eventsVorher := countRows(t, repo, "SELECT COUNT(*) FROM kassenjournal")
	auftraegeVorher := countRows(t, repo, "SELECT COUNT(*) FROM tse_signaturauftraege")

	// Duplikat-Einreichung: gleiche vorgangId, gleiche Nutzdaten, nächste freie
	// Version — der PK-Konflikt greift VOR dem Event-Insert, und die Nachprüfung
	// nach dem gescheiterten Commit erkennt am gleichen Hash das Duplikat.
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

// Nachprüf-Pfad mit abweichenden Nutzdaten: Derselbe Schlüssel mit einem anderen
// payload_hash ist weder ein Duplikat (stiller Erfolg verschluckte die Änderung)
// noch ein neuer Vorgang (das buchte doppelt), sondern ein expliziter Konflikt.
func TestWriteEventMitVorgang_AbweichendeNutzdatenSchreibenNichts(t *testing.T) {
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

	vorgangID := "bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb"
	vorgang := Vorgang{VorgangID: vorgangID, Art: VorgangArtZahlung, UserID: userID, PayloadHash: payloadHash(t, VorgangArtZahlung, testNutzdaten{TischID: tischID, Positionen: []int{1}})}
	zahlung := newTestEvent(userID, "zahlung-kassiert:v1", subject, 2, validZahlungData("p0000000-0000-0000-0000-000000000001", 2, 700))

	if _, err := repo.WriteEventMitVorgang(context.Background(), vorgang, zahlung, kasse.StreamTypeTischSession, ksNr); err != nil {
		t.Fatalf("erster Write: %v", err)
	}

	eventsVorher := countRows(t, repo, "SELECT COUNT(*) FROM kassenjournal")

	abweichend := Vorgang{VorgangID: vorgangID, Art: VorgangArtZahlung, UserID: userID, PayloadHash: payloadHash(t, VorgangArtZahlung, testNutzdaten{TischID: tischID, Positionen: []int{1, 2}})}
	zweiteZahlung := newTestEvent(userID, "zahlung-kassiert:v1", subject, 3, validZahlungData("p0000000-0000-0000-0000-000000000001", 2, 700))

	_, err = repo.WriteEventMitVorgang(context.Background(), abweichend, zweiteZahlung, kasse.StreamTypeTischSession, ksNr)
	if !errors.Is(err, ErrVorgangDatenAbweichend) {
		t.Fatalf("erwartet ErrVorgangDatenAbweichend, bekam: %v", err)
	}

	if n := countRows(t, repo, "SELECT COUNT(*) FROM kassenjournal"); n != eventsVorher {
		t.Errorf("abweichende Einreichung hat Events geschrieben: %d -> %d", eventsVorher, n)
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
	vorgang := Vorgang{VorgangID: "cccccccc-cccc-4ccc-8ccc-cccccccccccc", Art: VorgangArtZahlung, UserID: userID, PayloadHash: payloadHash(t, VorgangArtZahlung, testNutzdaten{TischID: tischID, Positionen: []int{1}})}
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

// testDruckauftraege liefert genau einen gültigen Druckauftrag — die
// Bestellung und der Direktverkauf schreiben ihre Events zusammen mit ihren
// Arbeitsbons in einer Transaktion.
func testDruckauftraege(_ event.Event) []druckauftrag_repo.NeuerDruckauftrag {
	return []druckauftrag_repo.NeuerDruckauftrag{{
		ZielIP:   "192.168.1.50",
		Payload:  "AAA=",
		BonArt:   "arbeitsbon",
		Referenz: "bestellung-aufgenommen:test",
	}}
}

// Die Idempotenz-Zeile entsteht auch im Druckauftrags-Pfad vor dem Event: Eine
// Duplikat-Einreichung schreibt weder Event noch Arbeitsbon.
func TestWriteEventWithDruckauftraegeMitVorgang_DuplikatSchreibtNichts(t *testing.T) {
	userID, ksNr, repo, teardown := setup(t)
	defer teardown(t)

	tischID, err := createTisch(repo.db, "Tisch 1")
	if err != nil {
		t.Fatalf("Failed to create tisch: %v", err)
	}

	subject := kasse.TischSessionSubject(ksNr, tischID)
	hash := payloadHash(t, VorgangArtBestellung, testNutzdaten{TischID: tischID, Positionen: []int{1}})
	vorgang := Vorgang{VorgangID: "11111111-1111-4111-8111-111111111111", Art: VorgangArtBestellung, UserID: userID, PayloadHash: hash}
	bestellung := newTestEvent(userID, "bestellung-aufgenommen:v1", subject, 1, validBestellungData("b0000000-0000-0000-0000-000000000001", "p0000000-0000-0000-0000-000000000001", 350, 2))

	if _, err := repo.WriteEventWithDruckauftraegeMitVorgang(context.Background(), vorgang, bestellung, kasse.StreamTypeTischSession, ksNr, testDruckauftraege); err != nil {
		t.Fatalf("erster Write: %v", err)
	}

	eventsVorher := countRows(t, repo, "SELECT COUNT(*) FROM kassenjournal")
	auftraegeVorher := countRows(t, repo, "SELECT COUNT(*) FROM druckauftraege")

	// Duplikat: gleiche vorgangId, gleiche Nutzdaten, nächste freie Version.
	retry := newTestEvent(userID, "bestellung-aufgenommen:v1", subject, 2, validBestellungData("b0000000-0000-0000-0000-000000000001", "p0000000-0000-0000-0000-000000000001", 350, 2))
	_, err = repo.WriteEventWithDruckauftraegeMitVorgang(context.Background(), vorgang, retry, kasse.StreamTypeTischSession, ksNr, testDruckauftraege)
	if !errors.Is(err, ErrVorgangBereitsGebucht) {
		t.Fatalf("erwartet ErrVorgangBereitsGebucht, bekam: %v", err)
	}

	// Derselbe Schlüssel mit anderen Nutzdaten meldet den Konflikt.
	abweichend := Vorgang{VorgangID: vorgang.VorgangID, Art: VorgangArtBestellung, UserID: userID, PayloadHash: payloadHash(t, VorgangArtBestellung, testNutzdaten{TischID: tischID, Positionen: []int{1, 2}})}
	_, err = repo.WriteEventWithDruckauftraegeMitVorgang(context.Background(), abweichend, retry, kasse.StreamTypeTischSession, ksNr, testDruckauftraege)
	if !errors.Is(err, ErrVorgangDatenAbweichend) {
		t.Fatalf("erwartet ErrVorgangDatenAbweichend, bekam: %v", err)
	}

	if n := countRows(t, repo, "SELECT COUNT(*) FROM kassenjournal"); n != eventsVorher {
		t.Errorf("Zweiteinreichung hat Events geschrieben: %d -> %d", eventsVorher, n)
	}
	if n := countRows(t, repo, "SELECT COUNT(*) FROM druckauftraege"); n != auftraegeVorher {
		t.Errorf("Zweiteinreichung hat Druckaufträge geschrieben: %d -> %d", auftraegeVorher, n)
	}
}

// Ein echter OCC-Konflikt bleibt auch im Druckauftrags-Pfad von einer
// Duplikat-Einreichung unterscheidbar: Er kommt aus UNIQUE(subject, version) und
// liefert db.ErrAlreadyExists (die Application-Schicht bildet das auf ErrConflict
// ab), nicht einen der beiden Idempotenz-Sentinels.
func TestWriteEventWithDruckauftraegeMitVorgang_OCCKonfliktBleibtUnterscheidbar(t *testing.T) {
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

	// Neue vorgangId, aber veraltete Stream-Version (1 ist bereits belegt).
	vorgang := Vorgang{VorgangID: "22222222-2222-4222-8222-222222222222", Art: VorgangArtBestellung, UserID: userID, PayloadHash: payloadHash(t, VorgangArtBestellung, testNutzdaten{TischID: tischID, Positionen: []int{1}})}
	zweite := newTestEvent(userID, "bestellung-aufgenommen:v1", subject, 1, validBestellungData("b0000000-0000-0000-0000-000000000002", "p0000000-0000-0000-0000-000000000002", 350, 2))

	_, err = repo.WriteEventWithDruckauftraegeMitVorgang(context.Background(), vorgang, zweite, kasse.StreamTypeTischSession, ksNr, testDruckauftraege)
	if !errors.Is(err, dbpkg.ErrAlreadyExists) {
		t.Fatalf("erwartet db.ErrAlreadyExists (OCC-Konflikt), bekam: %v", err)
	}
	if errors.Is(err, ErrVorgangBereitsGebucht) || errors.Is(err, ErrVorgangDatenAbweichend) {
		t.Fatalf("OCC-Konflikt darf kein Idempotenz-Sentinel sein, bekam: %v", err)
	}

	if n := countRows(t, repo, "SELECT COUNT(*) FROM vorgang_idempotenz WHERE vorgang_id = $1", vorgang.VorgangID); n != 0 {
		t.Errorf("erwartet keine vorgang_idempotenz-Zeile nach Rollback, vorhanden: %d", n)
	}
	if n := countRows(t, repo, "SELECT COUNT(*) FROM druckauftraege"); n != 0 {
		t.Errorf("erwartet keine Druckaufträge nach Rollback, vorhanden: %d", n)
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

	hash := payloadHash(t, VorgangArtStornierung, testNutzdaten{TischID: tischID, Positionen: []int{1}})
	vorgang := Vorgang{VorgangID: "dddddddd-dddd-4ddd-8ddd-dddddddddddd", Art: VorgangArtStornierung, UserID: userID, PayloadHash: hash}

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

	// Derselbe Schlüssel mit anderen Nutzdaten meldet den Konflikt statt still
	// erfolgreich zu sein.
	abweichend := Vorgang{VorgangID: vorgang.VorgangID, Art: VorgangArtStornierung, UserID: userID, PayloadHash: payloadHash(t, VorgangArtStornierung, testNutzdaten{TischID: tischID, Positionen: []int{1, 2}})}
	err = repo.WriteTischSessionEventsAtomicMitVorgang(context.Background(), abweichend, retry, ksNr)
	if !errors.Is(err, ErrVorgangDatenAbweichend) {
		t.Fatalf("erwartet ErrVorgangDatenAbweichend, bekam: %v", err)
	}

	if n := countRows(t, repo, "SELECT COUNT(*) FROM kassenjournal"); n != eventsVorher {
		t.Errorf("abweichende Einreichung hat Events geschrieben: %d -> %d", eventsVorher, n)
	}
}
