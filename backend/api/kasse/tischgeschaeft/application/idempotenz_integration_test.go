//go:build integration

package application

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/nicograef/jotti/backend/api/kasse/enrichment"
	dbpkg "github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/kasse"
	"github.com/nicograef/jotti/backend/repository/druckstation_repo"
	"github.com/nicograef/jotti/backend/repository/kassenjournal_repo"
	"github.com/nicograef/jotti/backend/repository/kassensitzungen_repo"
	"github.com/nicograef/jotti/backend/repository/produkt_repo"
	"github.com/nicograef/jotti/backend/repository/tisch_repo"
)

func cleanTischDB(t *testing.T, db *sql.DB) {
	t.Helper()
	stmts := []string{
		"DELETE FROM tse_signaturauftraege",
		"DELETE FROM druckauftraege",
		"DELETE FROM tisch_sessions",
		"ALTER TABLE kassenjournal DISABLE TRIGGER kassenjournal_no_delete",
		"DELETE FROM kassenjournal",
		"ALTER TABLE kassenjournal ENABLE TRIGGER kassenjournal_no_delete",
		"DELETE FROM kassensitzungen",
		"DELETE FROM tische",
		"DELETE FROM produkt_varianten",
		"DELETE FROM produkte",
		"DELETE FROM vorgang_idempotenz",
		"DELETE FROM users",
	}
	for _, stmt := range stmts {
		if _, err := db.Exec(stmt); err != nil {
			t.Fatalf("cleanTischDB %q: %v", stmt, err)
		}
	}
}

func setupBestellungIntegration(t *testing.T) (ctx context.Context, cmd Command, db *sql.DB, userID, ksNr, tischID, produktID, varianteID int) {
	t.Helper()
	db = dbpkg.OpenTestDatabase()
	cleanTischDB(t, db)
	t.Cleanup(func() {
		cleanTischDB(t, db)
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
		"INSERT INTO tische (name, status, created_at, updated_at) VALUES ('Tisch 1', 'active', now(), now()) RETURNING id",
	).Scan(&tischID); err != nil {
		t.Fatalf("create tisch: %v", err)
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
		TischRepo:           tisch_repo.NewRepository(db),
		EventRepo:           kassenjournal_repo.NewRepository(db),
		ProduktRepo:         produkt_repo.NewRepository(db),
		KassensitzungenRepo: kassensitzungen_repo.NewRepository(db),
		DruckstationRepo:    druckstation_repo.NewRepository(db),
	}
	return
}

// TestBestellungAufnehmen_DuplikatBestellungId_IdempotenterErfolg: Zwei identische Aufrufe mit
// derselben bestellungId erzeugen genau ein kassenjournal-Event und einen Signaturauftrag; der
// zweite Aufruf gibt nil zurück (idempotenter Erfolg).
func TestBestellungAufnehmen_DuplikatBestellungId_IdempotenterErfolg(t *testing.T) {
	ctx, cmd, db, userID, _, tischID, produktID, varianteID := setupBestellungIntegration(t)

	bestellungID := uuid.New().String()
	inputs := []enrichment.PositionInput{{ProduktID: produktID, VarianteID: varianteID, Menge: 1}}

	if err := cmd.BestellungAufnehmen(ctx, userID, "test", bestellungID, tischID, inputs, ""); err != nil {
		t.Fatalf("erster Aufruf: %v", err)
	}

	if err := cmd.BestellungAufnehmen(ctx, userID, "test", bestellungID, tischID, inputs, ""); err != nil {
		t.Fatalf("zweiter Aufruf (Duplikat) erwartet nil, bekam: %v", err)
	}

	var eventCount int
	if err := db.QueryRow("SELECT COUNT(*) FROM kassenjournal WHERE type = 'bestellung-aufgenommen:v1'").Scan(&eventCount); err != nil {
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

// Der im Review beschriebene Fehlbetrags-Pfad: Zwei Bier gehen durch, die Antwort
// geht verloren, der Gast bestellt eine Cola nach, und die Servicekraft schickt
// die erweiterte Bestellung unter derselben bestellungId erneut ab. Weder eine
// zweite Buchung (Doppelbuchung, Fehlbetrag im Kassensturz) noch ein stiller
// Erfolg (verschluckte Cola) wäre richtig — der Server meldet
// ErrVorgangDatenAbweichend und schreibt kein zweites Event.
func TestBestellungAufnehmen_SelbeBestellungIdAndereNutzdaten_KeinZweitesEvent(t *testing.T) {
	ctx, cmd, db, userID, _, tischID, produktID, varianteID := setupBestellungIntegration(t)

	bestellungID := uuid.New().String()
	inputs := []enrichment.PositionInput{{ProduktID: produktID, VarianteID: varianteID, Menge: 2}}
	if err := cmd.BestellungAufnehmen(ctx, userID, "test", bestellungID, tischID, inputs, ""); err != nil {
		t.Fatalf("erster Aufruf: %v", err)
	}
	auftraegeVorher := countSignaturauftraege(t, db)

	colaProduktID, colaVarianteID := createProduktMitVariante(t, db, "Cola", "0.33L", 300)
	geaendert := []enrichment.PositionInput{
		{ProduktID: produktID, VarianteID: varianteID, Menge: 2},
		{ProduktID: colaProduktID, VarianteID: colaVarianteID, Menge: 1},
	}
	if err := cmd.BestellungAufnehmen(ctx, userID, "test", bestellungID, tischID, geaendert, ""); !errors.Is(err, ErrVorgangDatenAbweichend) {
		t.Fatalf("erwartet ErrVorgangDatenAbweichend, bekam: %v", err)
	}

	if n := countByType(t, db, "bestellung-aufgenommen:v1"); n != 1 {
		t.Errorf("erwartet weiterhin 1 bestellung-aufgenommen:v1-Event, gespeichert: %d", n)
	}
	if n := countSignaturauftraege(t, db); n != auftraegeVorher {
		t.Errorf("abweichende Einreichung hat Signaturaufträge erzeugt: %d -> %d", auftraegeVorher, n)
	}
	if n := countVorgaenge(t, db, bestellungID); n != 1 {
		t.Errorf("erwartet genau 1 vorgang_idempotenz-Zeile, vorhanden: %d", n)
	}
}

// veralteteVersionRepo liefert eine um eins zu niedrige Stream-Version und
// erzwingt damit im realen Schreibpfad einen echten OCC-Konflikt: Das Event
// trifft auf eine bereits belegte (subject, version).
type veralteteVersionRepo struct{ eventRepo }

func (r veralteteVersionRepo) GetMaxVersion(ctx context.Context, subject string) (int, error) {
	version, err := r.eventRepo.GetMaxVersion(ctx, subject)
	if err != nil {
		return 0, err
	}
	return version - 1, nil
}

// Ein echter Versionskonflikt ist keine Duplikat-Einreichung: Eine neue
// bestellungId gegen eine veraltete Stream-Version scheitert an
// UNIQUE(subject, version) und ergibt weiterhin ErrConflict — weder die stille
// Erfolgsantwort noch ErrVorgangDatenAbweichend. Die Idempotenz-Zeile des
// gescheiterten Vorgangs rollt mit zurück.
func TestBestellungAufnehmen_NeueBestellungIdVeralteteVersion_ErrConflict(t *testing.T) {
	ctx, cmd, db, userID, _, tischID, produktID, varianteID := setupBestellungIntegration(t)

	inputs := []enrichment.PositionInput{{ProduktID: produktID, VarianteID: varianteID, Menge: 1}}
	if err := cmd.BestellungAufnehmen(ctx, userID, "test", uuid.New().String(), tischID, inputs, ""); err != nil {
		t.Fatalf("erste Bestellung: %v", err)
	}

	cmd.EventRepo = veralteteVersionRepo{cmd.EventRepo}

	zweiteID := uuid.New().String()
	if err := cmd.BestellungAufnehmen(ctx, userID, "test", zweiteID, tischID, inputs, ""); !errors.Is(err, ErrConflict) {
		t.Fatalf("erwartet ErrConflict, bekam: %v", err)
	}

	if n := countByType(t, db, "bestellung-aufgenommen:v1"); n != 1 {
		t.Errorf("erwartet weiterhin 1 bestellung-aufgenommen:v1-Event, gespeichert: %d", n)
	}
	if n := countVorgaenge(t, db, zweiteID); n != 0 {
		t.Errorf("erwartet keine vorgang_idempotenz-Zeile nach Rollback, vorhanden: %d", n)
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

// countByType zählt kassenjournal-Events des gegebenen Typs.
func countByType(t *testing.T, db *sql.DB, eventType string) int {
	t.Helper()
	var n int
	if err := db.QueryRow("SELECT COUNT(*) FROM kassenjournal WHERE type = $1", eventType).Scan(&n); err != nil {
		t.Fatalf("kassenjournal zählen (%s): %v", eventType, err)
	}
	return n
}

// countSignaturauftraege zählt alle Signaturaufträge.
func countSignaturauftraege(t *testing.T, db *sql.DB) int {
	t.Helper()
	var n int
	if err := db.QueryRow("SELECT COUNT(*) FROM tse_signaturauftraege").Scan(&n); err != nil {
		t.Fatalf("tse_signaturauftraege zählen: %v", err)
	}
	return n
}

// countVorgaenge zählt die vorgang_idempotenz-Zeilen einer vorgangId.
func countVorgaenge(t *testing.T, db *sql.DB, vorgangID string) int {
	t.Helper()
	var n int
	if err := db.QueryRow("SELECT COUNT(*) FROM vorgang_idempotenz WHERE vorgang_id = $1", vorgangID).Scan(&n); err != nil {
		t.Fatalf("vorgang_idempotenz zählen: %v", err)
	}
	return n
}

// offeneRefs liest die unbezahlten Positionen der Tisch-Session als PositionRefs.
func offeneRefs(t *testing.T, ctx context.Context, cmd Command, subject string) []kasse.PositionRef {
	t.Helper()
	state, err := cmd.EventRepo.ReadTischSession(ctx, subject)
	if err != nil {
		t.Fatalf("tisch session lesen: %v", err)
	}
	refs := make([]kasse.PositionRef, 0, len(state.UnbezahltePositionen))
	for _, p := range state.UnbezahltePositionen {
		refs = append(refs, kasse.PositionRef{PositionID: p.PositionID, Menge: p.Menge})
	}
	return refs
}

// Zwei identische Kassier-Anfragen mit derselben vorgangId erzeugen genau ein
// zahlung-kassiert:v1-Event und genau einen zusätzlichen Signaturauftrag;
// beide Aufrufe geben nil zurück (idempotenter Erfolg).
func TestZahlungKassieren_DuplikatVorgangId_GenauEinEventUndSignaturauftrag(t *testing.T) {
	ctx, cmd, db, userID, ksNr, tischID, produktID, varianteID := setupBestellungIntegration(t)

	inputs := []enrichment.PositionInput{{ProduktID: produktID, VarianteID: varianteID, Menge: 1}}
	if err := cmd.BestellungAufnehmen(ctx, userID, "test", uuid.New().String(), tischID, inputs, ""); err != nil {
		t.Fatalf("bestellung: %v", err)
	}

	subject := kasse.TischSessionSubject(ksNr, tischID)
	refs := offeneRefs(t, ctx, cmd, subject)
	auftraegeVorher := countSignaturauftraege(t, db)

	vorgangID := uuid.New().String()
	if err := cmd.ZahlungKassieren(ctx, userID, "test", vorgangID, tischID, refs, ""); err != nil {
		t.Fatalf("erster Aufruf: %v", err)
	}
	if err := cmd.ZahlungKassieren(ctx, userID, "test", vorgangID, tischID, refs, ""); err != nil {
		t.Fatalf("zweiter Aufruf (Duplikat) erwartet nil, bekam: %v", err)
	}

	if n := countByType(t, db, "zahlung-kassiert:v1"); n != 1 {
		t.Errorf("erwartet 1 zahlung-kassiert:v1-Event, gespeichert: %d", n)
	}
	if n := countSignaturauftraege(t, db); n != auftraegeVorher+1 {
		t.Errorf("erwartet genau 1 zusätzlichen Signaturauftrag, vorher %d, nachher %d", auftraegeVorher, n)
	}
}

// Dieselbe vorgangId mit geänderten Nutzdaten ist weder ein Duplikat noch eine
// neue Buchung: Der Server meldet ErrVorgangDatenAbweichend und schreibt kein
// zweites Event. Die Vorprüfung greift dabei vor der fachlichen Validierung —
// die Positionen des ersten Aufrufs sind bereits kassiert.
func TestZahlungKassieren_SelbeVorgangIdAndereNutzdaten_KeinZweitesEvent(t *testing.T) {
	ctx, cmd, db, userID, ksNr, tischID, produktID, varianteID := setupBestellungIntegration(t)

	inputs := []enrichment.PositionInput{{ProduktID: produktID, VarianteID: varianteID, Menge: 1}}
	if err := cmd.BestellungAufnehmen(ctx, userID, "test", uuid.New().String(), tischID, inputs, ""); err != nil {
		t.Fatalf("bestellung: %v", err)
	}

	subject := kasse.TischSessionSubject(ksNr, tischID)
	refs := offeneRefs(t, ctx, cmd, subject)

	vorgangID := uuid.New().String()
	if err := cmd.ZahlungKassieren(ctx, userID, "test", vorgangID, tischID, refs, ""); err != nil {
		t.Fatalf("erster Aufruf: %v", err)
	}
	auftraegeVorher := countSignaturauftraege(t, db)

	if err := cmd.ZahlungKassieren(ctx, userID, "test", vorgangID, tischID, refs, "Trinkgeld"); !errors.Is(err, ErrVorgangDatenAbweichend) {
		t.Fatalf("erwartet ErrVorgangDatenAbweichend, bekam: %v", err)
	}

	if n := countByType(t, db, "zahlung-kassiert:v1"); n != 1 {
		t.Errorf("erwartet weiterhin 1 zahlung-kassiert:v1-Event, gespeichert: %d", n)
	}
	if n := countSignaturauftraege(t, db); n != auftraegeVorher {
		t.Errorf("abweichende Einreichung hat Signaturaufträge erzeugt: %d -> %d", auftraegeVorher, n)
	}
	if n := countVorgaenge(t, db, vorgangID); n != 1 {
		t.Errorf("erwartet genau 1 vorgang_idempotenz-Zeile, vorhanden: %d", n)
	}
}

// Zwei identische Storno-Anfragen mit derselben vorgangId lassen die
// Event-Anzahl des Vorgangs unverändert, auch wenn er mehrere Events umfasst
// (bestellung-korrigiert für unbezahlte plus stornierung-erteilt je Zahlung).
func TestStornierungErteilen_DuplikatVorgangId_EventAnzahlUnveraendert(t *testing.T) {
	ctx, cmd, db, userID, ksNr, tischID, produktID, varianteID := setupBestellungIntegration(t)

	// Bestellt 3, davon 1 bezahlt → Storno über 3 = 1 Korrektur + 1 Warenrücknahme.
	inputs := []enrichment.PositionInput{{ProduktID: produktID, VarianteID: varianteID, Menge: 3}}
	if err := cmd.BestellungAufnehmen(ctx, userID, "test", uuid.New().String(), tischID, inputs, ""); err != nil {
		t.Fatalf("bestellung: %v", err)
	}

	subject := kasse.TischSessionSubject(ksNr, tischID)
	alleRefs := offeneRefs(t, ctx, cmd, subject)
	if len(alleRefs) != 1 {
		t.Fatalf("erwartet 1 offene Position, vorhanden: %d", len(alleRefs))
	}
	teilRef := []kasse.PositionRef{{PositionID: alleRefs[0].PositionID, Menge: 1}}
	if err := cmd.ZahlungKassieren(ctx, userID, "test", uuid.New().String(), tischID, teilRef, ""); err != nil {
		t.Fatalf("zahlung: %v", err)
	}

	stornoRefs := []kasse.PositionRef{{PositionID: alleRefs[0].PositionID, Menge: 3}}
	vorgangID := uuid.New().String()
	if err := cmd.StornierungErteilen(ctx, userID, "test", vorgangID, tischID, stornoRefs, "Reklamation"); err != nil {
		t.Fatalf("erster Aufruf: %v", err)
	}

	korrekturVorher := countByType(t, db, "bestellung-korrigiert:v1")
	warenruecknahmeVorher := countByType(t, db, "stornierung-erteilt:v1")
	auftraegeVorher := countSignaturauftraege(t, db)
	if korrekturVorher != 1 || warenruecknahmeVorher != 1 {
		t.Fatalf("erwartet 1 Korrektur + 1 Warenrücknahme, gespeichert: %d + %d", korrekturVorher, warenruecknahmeVorher)
	}

	if err := cmd.StornierungErteilen(ctx, userID, "test", vorgangID, tischID, stornoRefs, "Reklamation"); err != nil {
		t.Fatalf("zweiter Aufruf (Duplikat) erwartet nil, bekam: %v", err)
	}

	if n := countByType(t, db, "bestellung-korrigiert:v1"); n != korrekturVorher {
		t.Errorf("Duplikat hat Korrektur-Events geschrieben: %d -> %d", korrekturVorher, n)
	}
	if n := countByType(t, db, "stornierung-erteilt:v1"); n != warenruecknahmeVorher {
		t.Errorf("Duplikat hat Warenrücknahme-Events geschrieben: %d -> %d", warenruecknahmeVorher, n)
	}
	if n := countSignaturauftraege(t, db); n != auftraegeVorher {
		t.Errorf("Duplikat hat Signaturaufträge erzeugt: %d -> %d", auftraegeVorher, n)
	}
}

// Ein Storno unter bekannter vorgangId mit geändertem Grund ist eine
// abweichende Einreichung: ErrVorgangDatenAbweichend, keine weiteren Events.
func TestStornierungErteilen_SelbeVorgangIdAndereNutzdaten_KeineWeiterenEvents(t *testing.T) {
	ctx, cmd, db, userID, ksNr, tischID, produktID, varianteID := setupBestellungIntegration(t)

	// Bestellt 3, davon 1 bezahlt → Storno über 3 = 1 Korrektur + 1 Warenrücknahme.
	inputs := []enrichment.PositionInput{{ProduktID: produktID, VarianteID: varianteID, Menge: 3}}
	if err := cmd.BestellungAufnehmen(ctx, userID, "test", uuid.New().String(), tischID, inputs, ""); err != nil {
		t.Fatalf("bestellung: %v", err)
	}

	subject := kasse.TischSessionSubject(ksNr, tischID)
	alleRefs := offeneRefs(t, ctx, cmd, subject)
	if len(alleRefs) != 1 {
		t.Fatalf("erwartet 1 offene Position, vorhanden: %d", len(alleRefs))
	}
	teilRef := []kasse.PositionRef{{PositionID: alleRefs[0].PositionID, Menge: 1}}
	if err := cmd.ZahlungKassieren(ctx, userID, "test", uuid.New().String(), tischID, teilRef, ""); err != nil {
		t.Fatalf("zahlung: %v", err)
	}

	stornoRefs := []kasse.PositionRef{{PositionID: alleRefs[0].PositionID, Menge: 3}}
	vorgangID := uuid.New().String()
	if err := cmd.StornierungErteilen(ctx, userID, "test", vorgangID, tischID, stornoRefs, "Reklamation"); err != nil {
		t.Fatalf("erster Aufruf: %v", err)
	}

	korrekturVorher := countByType(t, db, "bestellung-korrigiert:v1")
	warenruecknahmeVorher := countByType(t, db, "stornierung-erteilt:v1")
	auftraegeVorher := countSignaturauftraege(t, db)

	if err := cmd.StornierungErteilen(ctx, userID, "test", vorgangID, tischID, stornoRefs, "Anderer Grund"); !errors.Is(err, ErrVorgangDatenAbweichend) {
		t.Fatalf("erwartet ErrVorgangDatenAbweichend, bekam: %v", err)
	}

	if n := countByType(t, db, "bestellung-korrigiert:v1"); n != korrekturVorher {
		t.Errorf("abweichende Einreichung hat Korrektur-Events geschrieben: %d -> %d", korrekturVorher, n)
	}
	if n := countByType(t, db, "stornierung-erteilt:v1"); n != warenruecknahmeVorher {
		t.Errorf("abweichende Einreichung hat Warenrücknahme-Events geschrieben: %d -> %d", warenruecknahmeVorher, n)
	}
	if n := countSignaturauftraege(t, db); n != auftraegeVorher {
		t.Errorf("abweichende Einreichung hat Signaturaufträge erzeugt: %d -> %d", auftraegeVorher, n)
	}
	if n := countVorgaenge(t, db, vorgangID); n != 1 {
		t.Errorf("erwartet genau 1 vorgang_idempotenz-Zeile, vorhanden: %d", n)
	}
}

// Zwei identische Umbuchungs-Anfragen mit derselben vorgangId erzeugen genau
// ein Quell- und ein Ziel-Event (zwei bestellung-umgebucht:v1 insgesamt).
func TestBestellungUmbuchen_DuplikatVorgangId_GenauZweiEvents(t *testing.T) {
	ctx, cmd, db, userID, ksNr, tischID, produktID, varianteID := setupBestellungIntegration(t)

	var zielTischID int
	if err := db.QueryRow(
		"INSERT INTO tische (name, status, created_at, updated_at) VALUES ('Tisch 2', 'active', now(), now()) RETURNING id",
	).Scan(&zielTischID); err != nil {
		t.Fatalf("create ziel-tisch: %v", err)
	}

	inputs := []enrichment.PositionInput{{ProduktID: produktID, VarianteID: varianteID, Menge: 1}}
	if err := cmd.BestellungAufnehmen(ctx, userID, "test", uuid.New().String(), tischID, inputs, ""); err != nil {
		t.Fatalf("bestellung: %v", err)
	}

	subject := kasse.TischSessionSubject(ksNr, tischID)
	refs := offeneRefs(t, ctx, cmd, subject)
	auftraegeVorher := countSignaturauftraege(t, db)

	vorgangID := uuid.New().String()
	if err := cmd.BestellungUmbuchen(ctx, userID, "test", vorgangID, tischID, zielTischID, refs, ""); err != nil {
		t.Fatalf("erster Aufruf: %v", err)
	}
	if err := cmd.BestellungUmbuchen(ctx, userID, "test", vorgangID, tischID, zielTischID, refs, ""); err != nil {
		t.Fatalf("zweiter Aufruf (Duplikat) erwartet nil, bekam: %v", err)
	}

	if n := countByType(t, db, "bestellung-umgebucht:v1"); n != 2 {
		t.Errorf("erwartet genau 2 bestellung-umgebucht:v1-Events (Quelle + Ziel), gespeichert: %d", n)
	}
	if n := countSignaturauftraege(t, db); n != auftraegeVorher+2 {
		t.Errorf("erwartet genau 2 zusätzliche Signaturaufträge, vorher %d, nachher %d", auftraegeVorher, n)
	}
}

// Eine Umbuchung unter bekannter vorgangId mit geändertem Benutzer-Kommentar ist
// eine abweichende Einreichung: ErrVorgangDatenAbweichend, keine weiteren Events.
func TestBestellungUmbuchen_SelbeVorgangIdAndereNutzdaten_KeineWeiterenEvents(t *testing.T) {
	ctx, cmd, db, userID, ksNr, tischID, produktID, varianteID := setupBestellungIntegration(t)

	var zielTischID int
	if err := db.QueryRow(
		"INSERT INTO tische (name, status, created_at, updated_at) VALUES ('Tisch 2', 'active', now(), now()) RETURNING id",
	).Scan(&zielTischID); err != nil {
		t.Fatalf("create ziel-tisch: %v", err)
	}

	inputs := []enrichment.PositionInput{{ProduktID: produktID, VarianteID: varianteID, Menge: 1}}
	if err := cmd.BestellungAufnehmen(ctx, userID, "test", uuid.New().String(), tischID, inputs, ""); err != nil {
		t.Fatalf("bestellung: %v", err)
	}

	subject := kasse.TischSessionSubject(ksNr, tischID)
	refs := offeneRefs(t, ctx, cmd, subject)

	vorgangID := uuid.New().String()
	if err := cmd.BestellungUmbuchen(ctx, userID, "test", vorgangID, tischID, zielTischID, refs, ""); err != nil {
		t.Fatalf("erster Aufruf: %v", err)
	}
	auftraegeVorher := countSignaturauftraege(t, db)

	if err := cmd.BestellungUmbuchen(ctx, userID, "test", vorgangID, tischID, zielTischID, refs, "Falscher Tisch"); !errors.Is(err, ErrVorgangDatenAbweichend) {
		t.Fatalf("erwartet ErrVorgangDatenAbweichend, bekam: %v", err)
	}

	if n := countByType(t, db, "bestellung-umgebucht:v1"); n != 2 {
		t.Errorf("erwartet weiterhin genau 2 bestellung-umgebucht:v1-Events, gespeichert: %d", n)
	}
	if n := countSignaturauftraege(t, db); n != auftraegeVorher {
		t.Errorf("abweichende Einreichung hat Signaturaufträge erzeugt: %d -> %d", auftraegeVorher, n)
	}
	if n := countVorgaenge(t, db, vorgangID); n != 1 {
		t.Errorf("erwartet genau 1 vorgang_idempotenz-Zeile, vorhanden: %d", n)
	}
}
