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
	"github.com/nicograef/jotti/backend/api/kasse/enrichment"
	dbpkg "github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/kasse"
	"github.com/nicograef/jotti/backend/repository/kassenjournal_repo"
	"github.com/nicograef/jotti/backend/repository/kassensitzungen_repo"
	"github.com/nicograef/jotti/backend/repository/produkt_repo"
)

func cleanDVDB(t *testing.T, db *sql.DB) {
	t.Helper()
	stmts := []string{
		"DELETE FROM tse_signaturauftraege",
		"DELETE FROM druckauftraege",
		"DELETE FROM tisch_sessions",
		"ALTER TABLE kassenjournal DISABLE TRIGGER kassenjournal_no_delete",
		"DELETE FROM kassenjournal",
		"ALTER TABLE kassenjournal ENABLE TRIGGER kassenjournal_no_delete",
		"DELETE FROM kassensitzungen",
		"DELETE FROM produkt_varianten",
		"DELETE FROM produkte",
		"DELETE FROM vorgang_idempotenz",
		"DELETE FROM users",
	}
	for _, stmt := range stmts {
		if _, err := db.Exec(stmt); err != nil {
			t.Fatalf("cleanDVDB %q: %v", stmt, err)
		}
	}
}

func setupDVIntegration(t *testing.T) (ctx context.Context, cmd Command, db *sql.DB, userID, ksNr, produktID, varianteID int) {
	t.Helper()
	db = dbpkg.OpenTestDatabase()
	cleanDVDB(t, db)
	t.Cleanup(func() {
		cleanDVDB(t, db)
		_ = db.Close()
	})

	ctx = context.Background()

	if err := db.QueryRow(
		"INSERT INTO users (name, username, role, status, password_hash, onetime_password_hash, created_at, updated_at) VALUES ('Test', 'test', 'admin', 'active', 'hash', 'hash', now(), now()) RETURNING id",
	).Scan(&userID); err != nil {
		t.Fatalf("create user: %v", err)
	}

	if err := db.QueryRow(
		"INSERT INTO kassensitzungen (datum, bezeichnung, status, created_at, updated_at) VALUES ((NOW() AT TIME ZONE 'Europe/Berlin')::date, 'Test', 'offen', now(), now()) RETURNING z_nr",
	).Scan(&ksNr); err != nil {
		t.Fatalf("create kassensitzung: %v", err)
	}

	if err := db.QueryRow(
		"INSERT INTO produkte (name, kategorie, steuersatz, status, created_at, updated_at) VALUES ('Bier', 'getraenk', 'regel', 'active', now(), now()) RETURNING id",
	).Scan(&produktID); err != nil {
		t.Fatalf("create produkt: %v", err)
	}

	if err := db.QueryRow(
		"INSERT INTO produkt_varianten (produkt_id, name, preis_cents, status, created_at, updated_at) VALUES ($1, '0.5L', 350, 'active', now(), now()) RETURNING id",
		produktID,
	).Scan(&varianteID); err != nil {
		t.Fatalf("create variante: %v", err)
	}

	cmd = Command{
		EventRepo:           kassenjournal_repo.NewRepository(db),
		ProduktRepo:         produkt_repo.NewRepository(db),
		KassensitzungenRepo: kassensitzungen_repo.NewRepository(db),
		// DruckstationRepo: nil — nil-bewacht, keine Druckaufträge im Test
	}
	return
}

// TestDirektverkaufTaetigen_DuplikatVerkaufId_IdempotenterErfolg: Zwei identische Aufrufe mit
// derselben verkaufId erzeugen genau ein kassenjournal-Event und einen Signaturauftrag; der
// zweite Aufruf gibt nil zurück (idempotenter Erfolg).
func TestDirektverkaufTaetigen_DuplikatVerkaufId_IdempotenterErfolg(t *testing.T) {
	ctx, cmd, db, userID, _, produktID, varianteID := setupDVIntegration(t)

	verkaufID := uuid.New().String()
	inputs := []enrichment.PositionInput{{ProduktID: produktID, VarianteID: varianteID, Menge: 1}}

	if err := cmd.DirektverkaufTaetigen(ctx, userID, "test", verkaufID, inputs, ""); err != nil {
		t.Fatalf("erster Aufruf: %v", err)
	}

	if err := cmd.DirektverkaufTaetigen(ctx, userID, "test", verkaufID, inputs, ""); err != nil {
		t.Fatalf("zweiter Aufruf (Duplikat) erwartet nil, bekam: %v", err)
	}

	var eventCount int
	if err := db.QueryRow("SELECT COUNT(*) FROM kassenjournal WHERE type = 'direktverkauf-getaetigt:v1'").Scan(&eventCount); err != nil {
		t.Fatalf("kassenjournal zählen: %v", err)
	}
	if eventCount != 1 {
		t.Errorf("erwartet 1 Event im kassenjournal, gespeichert: %d", eventCount)
	}

	var auftraegeCount int
	if err := db.QueryRow("SELECT COUNT(*) FROM tse_signaturauftraege").Scan(&auftraegeCount); err != nil {
		t.Fatalf("tse_signaturauftraege zählen: %v", err)
	}
	if auftraegeCount != 1 {
		t.Errorf("erwartet 1 tse_signaturauftrag, vorhanden: %d", auftraegeCount)
	}
}

// TestDirektverkaufTaetigen_SelbeVerkaufIdAndereNutzdaten_KeinZweitesEvent: Der im
// Review beschriebene Fehlbetrags-Pfad — zwei Bier gehen durch, die Antwort geht
// verloren, der Gast bestellt eine Cola nach, und die Servicekraft schließt den
// erweiterten Korb unter derselben verkaufId erneut ab. Weder eine zweite Buchung
// (Doppelbuchung, Fehlbetrag im Kassensturz) noch ein stiller Erfolg (verschluckte
// Cola) wäre richtig: Der Command meldet ErrVorgangDatenAbweichend und schreibt
// kein zweites Event.
func TestDirektverkaufTaetigen_SelbeVerkaufIdAndereNutzdaten_KeinZweitesEvent(t *testing.T) {
	ctx, cmd, db, userID, _, produktID, varianteID := setupDVIntegration(t)

	verkaufID := uuid.New().String()
	inputs := []enrichment.PositionInput{{ProduktID: produktID, VarianteID: varianteID, Menge: 2}}
	if err := cmd.DirektverkaufTaetigen(ctx, userID, "test", verkaufID, inputs, ""); err != nil {
		t.Fatalf("erster Aufruf: %v", err)
	}

	var auftraegeVorher int
	if err := db.QueryRow("SELECT COUNT(*) FROM tse_signaturauftraege").Scan(&auftraegeVorher); err != nil {
		t.Fatalf("tse_signaturauftraege zählen: %v", err)
	}

	colaProduktID, colaVarianteID := createProduktMitVariante(t, db, "Cola", "0.33L", 300)
	geaendert := []enrichment.PositionInput{
		{ProduktID: produktID, VarianteID: varianteID, Menge: 2},
		{ProduktID: colaProduktID, VarianteID: colaVarianteID, Menge: 1},
	}
	if err := cmd.DirektverkaufTaetigen(ctx, userID, "test", verkaufID, geaendert, ""); !errors.Is(err, ErrVorgangDatenAbweichend) {
		t.Fatalf("erwartet ErrVorgangDatenAbweichend, bekam: %v", err)
	}

	var eventCount int
	if err := db.QueryRow("SELECT COUNT(*) FROM kassenjournal WHERE type = 'direktverkauf-getaetigt:v1'").Scan(&eventCount); err != nil {
		t.Fatalf("kassenjournal zählen: %v", err)
	}
	if eventCount != 1 {
		t.Errorf("erwartet weiterhin 1 direktverkauf-getaetigt:v1-Event, gespeichert: %d", eventCount)
	}

	var auftraegeNachher int
	if err := db.QueryRow("SELECT COUNT(*) FROM tse_signaturauftraege").Scan(&auftraegeNachher); err != nil {
		t.Fatalf("tse_signaturauftraege zählen: %v", err)
	}
	if auftraegeNachher != auftraegeVorher {
		t.Errorf("abweichende Einreichung hat Signaturaufträge erzeugt: %d -> %d", auftraegeVorher, auftraegeNachher)
	}

	var vorgaenge int
	if err := db.QueryRow("SELECT COUNT(*) FROM vorgang_idempotenz WHERE vorgang_id = $1", verkaufID).Scan(&vorgaenge); err != nil {
		t.Fatalf("vorgang_idempotenz zählen: %v", err)
	}
	if vorgaenge != 1 {
		t.Errorf("erwartet genau 1 vorgang_idempotenz-Zeile, vorhanden: %d", vorgaenge)
	}
}

// createProduktMitVariante legt ein Produkt mit genau einer Variante an und
// liefert beide IDs.
func createProduktMitVariante(t *testing.T, db *sql.DB, produktName, variantenName string, preisCents int) (produktID, varianteID int) {
	t.Helper()
	if err := db.QueryRow(
		"INSERT INTO produkte (name, kategorie, steuersatz, status, created_at, updated_at) VALUES ($1, 'getraenk', 'regel', 'active', now(), now()) RETURNING id",
		produktName,
	).Scan(&produktID); err != nil {
		t.Fatalf("create produkt %q: %v", produktName, err)
	}
	if err := db.QueryRow(
		"INSERT INTO produkt_varianten (produkt_id, name, preis_cents, status, created_at, updated_at) VALUES ($1, $2, $3, 'active', now(), now()) RETURNING id",
		produktID, variantenName, preisCents,
	).Scan(&varianteID); err != nil {
		t.Fatalf("create variante %q: %v", variantenName, err)
	}
	return
}

// TestDirektverkaufTaetigen_VersionskonfliktAndereVerkaufId_ErrConflict: Existiert für den
// Direktverkauf-Stream bereits version 1 mit einer anderen verkaufId im Datenfeld,
// antwortet der Command mit ErrConflict (echter OCC-Konflikt, keine Duplikat-Einreichung).
func TestDirektverkaufTaetigen_VersionskonfliktAndereVerkaufId_ErrConflict(t *testing.T) {
	ctx, cmd, db, userID, ksNr, produktID, varianteID := setupDVIntegration(t)

	verkaufID := uuid.New().String()
	subject := kasse.DirektverkaufSubject(ksNr, verkaufID)

	// Vorab-Insert: version 1 im selben Stream, aber mit anderer verkaufId im Datenfeld.
	// Simuliert einen echten OCC-Konflikt: der Command-Write trifft auf eine bereits belegte
	// (subject, version). Die verkaufId ist unbekannt, also greift die Idempotenz nicht — der
	// Konflikt bleibt als UNIQUE(subject, version)-Verletzung ein OCC-Konflikt.
	andereVerkaufID := uuid.New().String()
	data, err := json.Marshal(map[string]any{
		"verkaufId":         andereVerkaufID,
		"positionen":        []any{},
		"gesamtbetragCents": 0,
		"kommentar":         "",
	})
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}
	if _, err := db.Exec(
		"INSERT INTO kassenjournal (user_id, user_name, type, subject, version, data, timestamp, kassensitzung_nr) VALUES ($1, 'test', 'direktverkauf-getaetigt:v1', $2, 1, $3::jsonb, now(), $4)",
		userID, subject, string(data), ksNr,
	); err != nil {
		t.Fatalf("Vorab-Insert: %v", err)
	}

	inputs := []enrichment.PositionInput{{ProduktID: produktID, VarianteID: varianteID, Menge: 1}}
	if err := cmd.DirektverkaufTaetigen(ctx, userID, "test", verkaufID, inputs, ""); !errors.Is(err, ErrConflict) {
		t.Errorf("erwartet ErrConflict, bekam: %v", err)
	}
}

// TestDirektverkaufStornieren_DuplikatVorgangId_GenauEinEventUndSignaturauftrag: Zwei
// identische Storno-Anfragen mit derselben vorgangId erzeugen genau ein
// direktverkauf-storniert:v1-Event und genau einen zusätzlichen Signaturauftrag;
// beide Aufrufe geben nil zurück (idempotenter Erfolg).
func TestDirektverkaufStornieren_DuplikatVorgangId_GenauEinEventUndSignaturauftrag(t *testing.T) {
	ctx, cmd, db, userID, ksNr, produktID, varianteID := setupDVIntegration(t)

	verkaufID := uuid.New().String()
	inputs := []enrichment.PositionInput{{ProduktID: produktID, VarianteID: varianteID, Menge: 1}}
	if err := cmd.DirektverkaufTaetigen(ctx, userID, "test", verkaufID, inputs, ""); err != nil {
		t.Fatalf("direktverkauf: %v", err)
	}

	// Server-erzeugte positionId aus dem Verkaufs-Stream lesen.
	subject := kasse.DirektverkaufSubject(ksNr, verkaufID)
	events, err := cmd.EventRepo.ReadEventsBySubject(ctx, subject)
	if err != nil || len(events) != 1 {
		t.Fatalf("verkauf events lesen: %v (%d)", err, len(events))
	}
	var data struct {
		Positionen []struct {
			PositionID string `json:"positionId"`
		} `json:"positionen"`
	}
	if err := json.Unmarshal(events[0].Data, &data); err != nil {
		t.Fatalf("verkauf data unmarshal: %v", err)
	}
	refs := []kasse.PositionRef{{PositionID: data.Positionen[0].PositionID, Menge: 1}}

	var auftraegeVorher int
	if err := db.QueryRow("SELECT COUNT(*) FROM tse_signaturauftraege").Scan(&auftraegeVorher); err != nil {
		t.Fatalf("tse_signaturauftraege zählen: %v", err)
	}

	vorgangID := uuid.New().String()
	if err := cmd.DirektverkaufStornieren(ctx, userID, "test", vorgangID, verkaufID, refs, "Rueckgabe"); err != nil {
		t.Fatalf("erster Aufruf: %v", err)
	}
	if err := cmd.DirektverkaufStornieren(ctx, userID, "test", vorgangID, verkaufID, refs, "Rueckgabe"); err != nil {
		t.Fatalf("zweiter Aufruf (Duplikat) erwartet nil, bekam: %v", err)
	}

	var eventCount int
	if err := db.QueryRow("SELECT COUNT(*) FROM kassenjournal WHERE type = 'direktverkauf-storniert:v1'").Scan(&eventCount); err != nil {
		t.Fatalf("kassenjournal zählen: %v", err)
	}
	if eventCount != 1 {
		t.Errorf("erwartet 1 direktverkauf-storniert:v1-Event, gespeichert: %d", eventCount)
	}

	var auftraegeNachher int
	if err := db.QueryRow("SELECT COUNT(*) FROM tse_signaturauftraege").Scan(&auftraegeNachher); err != nil {
		t.Fatalf("tse_signaturauftraege zählen: %v", err)
	}
	if auftraegeNachher != auftraegeVorher+1 {
		t.Errorf("erwartet genau 1 zusätzlichen Signaturauftrag, vorher %d, nachher %d", auftraegeVorher, auftraegeNachher)
	}
}

// TestDirektverkaufStornieren_SelbeVorgangIdAndereNutzdaten_KeinZweitesEvent:
// Dieselbe vorgangId mit geänderten Nutzdaten ist weder ein Duplikat noch eine
// neue Buchung — der Command meldet ErrVorgangDatenAbweichend und schreibt kein
// zweites Event. Die Vorprüfung greift dabei vor der fachlichen Validierung: Die
// geänderte Menge wäre hier noch stornierbar.
func TestDirektverkaufStornieren_SelbeVorgangIdAndereNutzdaten_KeinZweitesEvent(t *testing.T) {
	ctx, cmd, db, userID, ksNr, produktID, varianteID := setupDVIntegration(t)

	verkaufID := uuid.New().String()
	inputs := []enrichment.PositionInput{{ProduktID: produktID, VarianteID: varianteID, Menge: 3}}
	if err := cmd.DirektverkaufTaetigen(ctx, userID, "test", verkaufID, inputs, ""); err != nil {
		t.Fatalf("direktverkauf: %v", err)
	}

	// Server-erzeugte positionId aus dem Verkaufs-Stream lesen.
	subject := kasse.DirektverkaufSubject(ksNr, verkaufID)
	events, err := cmd.EventRepo.ReadEventsBySubject(ctx, subject)
	if err != nil || len(events) != 1 {
		t.Fatalf("verkauf events lesen: %v (%d)", err, len(events))
	}
	var data struct {
		Positionen []struct {
			PositionID string `json:"positionId"`
		} `json:"positionen"`
	}
	if err := json.Unmarshal(events[0].Data, &data); err != nil {
		t.Fatalf("verkauf data unmarshal: %v", err)
	}
	positionID := data.Positionen[0].PositionID

	vorgangID := uuid.New().String()
	if err := cmd.DirektverkaufStornieren(ctx, userID, "test", vorgangID, verkaufID, []kasse.PositionRef{{PositionID: positionID, Menge: 1}}, "Rueckgabe"); err != nil {
		t.Fatalf("erster Aufruf: %v", err)
	}

	var auftraegeVorher int
	if err := db.QueryRow("SELECT COUNT(*) FROM tse_signaturauftraege").Scan(&auftraegeVorher); err != nil {
		t.Fatalf("tse_signaturauftraege zählen: %v", err)
	}

	geaenderteRefs := []kasse.PositionRef{{PositionID: positionID, Menge: 2}}
	if err := cmd.DirektverkaufStornieren(ctx, userID, "test", vorgangID, verkaufID, geaenderteRefs, "Rueckgabe"); !errors.Is(err, ErrVorgangDatenAbweichend) {
		t.Fatalf("erwartet ErrVorgangDatenAbweichend, bekam: %v", err)
	}

	var eventCount int
	if err := db.QueryRow("SELECT COUNT(*) FROM kassenjournal WHERE type = 'direktverkauf-storniert:v1'").Scan(&eventCount); err != nil {
		t.Fatalf("kassenjournal zählen: %v", err)
	}
	if eventCount != 1 {
		t.Errorf("erwartet weiterhin 1 direktverkauf-storniert:v1-Event, gespeichert: %d", eventCount)
	}

	var auftraegeNachher int
	if err := db.QueryRow("SELECT COUNT(*) FROM tse_signaturauftraege").Scan(&auftraegeNachher); err != nil {
		t.Fatalf("tse_signaturauftraege zählen: %v", err)
	}
	if auftraegeNachher != auftraegeVorher {
		t.Errorf("abweichende Einreichung hat Signaturaufträge erzeugt: %d -> %d", auftraegeVorher, auftraegeNachher)
	}

	var vorgaenge int
	if err := db.QueryRow("SELECT COUNT(*) FROM vorgang_idempotenz WHERE vorgang_id = $1", vorgangID).Scan(&vorgaenge); err != nil {
		t.Fatalf("vorgang_idempotenz zählen: %v", err)
	}
	if vorgaenge != 1 {
		t.Errorf("erwartet genau 1 vorgang_idempotenz-Zeile, vorhanden: %d", vorgaenge)
	}
}
