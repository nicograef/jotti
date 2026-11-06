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
	"github.com/nicograef/jotti/backend/db/dbtest"
	"github.com/nicograef/jotti/backend/domain/kasse"
	"github.com/nicograef/jotti/backend/repository/kassenjournal_repo"
	"github.com/nicograef/jotti/backend/repository/kassensitzungen_repo"
	"github.com/nicograef/jotti/backend/repository/produkt_repo"
	"github.com/nicograef/jotti/backend/repository/tisch_repo"
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
		"DELETE FROM tische",
		"DELETE FROM produkt_varianten",
		"DELETE FROM produkte",
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
	db = dbtest.Open()
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
		TischRepo: tisch_repo.NewRepository(db),
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

	if err := cmd.DirektverkaufTaetigen(ctx, userID, "test", verkaufID, inputs, "", nil); err != nil {
		t.Fatalf("erster Aufruf: %v", err)
	}

	if err := cmd.DirektverkaufTaetigen(ctx, userID, "test", verkaufID, inputs, "", nil); err != nil {
		t.Errorf("zweiter Aufruf (Duplikat) erwartet nil, bekam: %v", err)
	}

	var eventCount int
	if err := db.QueryRow("SELECT COUNT(*) FROM kassenjournal WHERE type = 'direktverkauf-getaetigt:v2'").Scan(&eventCount); err != nil {
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

// TestDirektverkaufTaetigen_VersionskonfliktAndereVerkaufId_ErrConflict: Existiert für den
// Direktverkauf-Stream bereits version 1 mit einer anderen verkaufId im Datenfeld,
// antwortet der Command mit ErrConflict (echter OCC-Konflikt, keine Duplikat-Einreichung).
func TestDirektverkaufTaetigen_VersionskonfliktAndereVerkaufId_ErrConflict(t *testing.T) {
	ctx, cmd, db, userID, ksNr, produktID, varianteID := setupDVIntegration(t)

	verkaufID := uuid.New().String()
	subject := kasse.DirektverkaufSubject(ksNr, verkaufID)

	// Pre-insert version 1 in the same stream with another verkaufId: the command write hits a taken
	// (subject, version), a real OCC conflict, and the idempotency check finds no match.
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
	if err := cmd.DirektverkaufTaetigen(ctx, userID, "test", verkaufID, inputs, "", nil); !errors.Is(err, ErrConflict) {
		t.Errorf("erwartet ErrConflict, bekam: %v", err)
	}
}

// insertPfandVariante legt ein aktives Pfand-Produkt „Pfand" mit der Variante „Glas" zu 200 Cent an.
func insertPfandVariante(t *testing.T, db *sql.DB) (produktID, varianteID int) {
	t.Helper()
	if err := db.QueryRow(
		"INSERT INTO produkte (name, kategorie, steuersatz, status, pfand, created_at, updated_at) VALUES ('Pfand', 'sonstiges', 'regel', 'active', true, now(), now()) RETURNING id",
	).Scan(&produktID); err != nil {
		t.Fatalf("create pfand produkt: %v", err)
	}
	if err := db.QueryRow(
		"INSERT INTO produkt_varianten (produkt_id, name, preis_cents, status, created_at, updated_at) VALUES ($1, 'Glas', 200, 'active', now(), now()) RETURNING id",
		produktID,
	).Scan(&varianteID); err != nil {
		t.Fatalf("create pfand variante: %v", err)
	}
	return produktID, varianteID
}

// TestDirektverkaufTaetigen_MitPfand_DuplikatRueckgabeId_IdempotenterErfolg: a settled Direktverkauf
// sent twice stores the sale and the Rückgabe once each, with one Signaturauftrag per Vorgang.
func TestDirektverkaufTaetigen_MitPfand_DuplikatRueckgabeId_IdempotenterErfolg(t *testing.T) {
	ctx, cmd, db, userID, ksNr, produktID, varianteID := setupDVIntegration(t)
	pfandProduktID, pfandVarianteID := insertPfandVariante(t, db)

	verkaufID := uuid.New().String()
	inputs := []enrichment.PositionInput{{ProduktID: produktID, VarianteID: varianteID, Menge: 2}}
	pfand := &enrichment.PfandEingabe{
		RueckgabeID: uuid.New().String(),
		Positionen:  []enrichment.PositionInput{{ProduktID: pfandProduktID, VarianteID: pfandVarianteID, Menge: 2}},
	}

	if err := cmd.DirektverkaufTaetigen(ctx, userID, "test", verkaufID, inputs, "", pfand); err != nil {
		t.Fatalf("erster Aufruf: %v", err)
	}
	if err := cmd.DirektverkaufTaetigen(ctx, userID, "test", verkaufID, inputs, "", pfand); err != nil {
		t.Errorf("zweiter Aufruf (Duplikat) erwartet nil, bekam: %v", err)
	}

	var verkaeufe, rueckgaben int
	if err := db.QueryRow("SELECT COUNT(*) FILTER (WHERE type = 'direktverkauf-getaetigt:v2'), COUNT(*) FILTER (WHERE type = 'pfand-zurueckgegeben:v1') FROM kassenjournal").Scan(&verkaeufe, &rueckgaben); err != nil {
		t.Fatalf("kassenjournal zählen: %v", err)
	}
	if verkaeufe != 1 || rueckgaben != 1 {
		t.Errorf("erwartet je 1 Verkauf und 1 Rückgabe, gespeichert: %d und %d", verkaeufe, rueckgaben)
	}

	var rueckgabeData string
	if err := db.QueryRow("SELECT data::text FROM kassenjournal WHERE type = 'pfand-zurueckgegeben:v1' AND subject = $1 AND version = 1", kasse.PfandrueckgabeSubject(ksNr, pfand.RueckgabeID)).Scan(&rueckgabeData); err != nil {
		t.Fatalf("Rückgabe im eigenen Stream lesen: %v", err)
	}
	var rueckgabe kasse.PfandZurueckgegebenV1Data
	if err := json.Unmarshal([]byte(rueckgabeData), &rueckgabe); err != nil {
		t.Fatalf("Rückgabe dekodieren: %v", err)
	}
	if rueckgabe.TischID != 0 || rueckgabe.GesamtbetragCents != 400 {
		t.Errorf("erwartet tischId 0 und 400 Cent, bekam tischId %d und %d Cent", rueckgabe.TischID, rueckgabe.GesamtbetragCents)
	}

	var auftraegeCount int
	if err := db.QueryRow("SELECT COUNT(*) FROM tse_signaturauftraege").Scan(&auftraegeCount); err != nil {
		t.Fatalf("tse_signaturauftraege zählen: %v", err)
	}
	if auftraegeCount != 2 {
		t.Errorf("erwartet 2 tse_signaturauftraege, vorhanden: %d", auftraegeCount)
	}
}

// verrechnungsFall carries what a case of the Abweichend tests needs to book its first Vorgang and
// its retry.
type verrechnungsFall struct {
	ctx                    context.Context
	cmd                    Command
	userID                 int
	verkaufID, rueckgabeID string
	verkauf, pfand         func(menge int) []enrichment.PositionInput
}

// TestDirektverkaufTaetigen_MitPfand_Abweichend_KeinEvent: a Direktverkauf whose verkaufId or
// rueckgabeId is booked with other content, or with or without a partner the retry lacks or has, is
// refused with ErrPfandrueckgabeAbweichend and books nothing.
func TestDirektverkaufTaetigen_MitPfand_Abweichend_KeinEvent(t *testing.T) {
	mitPfand := func(f verrechnungsFall, pfandMenge int) error {
		return f.cmd.DirektverkaufTaetigen(f.ctx, f.userID, "test", f.verkaufID, f.verkauf(2), "", &enrichment.PfandEingabe{RueckgabeID: f.rueckgabeID, Positionen: f.pfand(pfandMenge)})
	}
	// Each case books a first Vorgang, then retries under the same IDs; the sale has Menge 2.
	tests := []struct {
		name        string
		erst, retry func(f verrechnungsFall) error
	}{
		{"andere Pfand-Menge",
			func(f verrechnungsFall) error { return mitPfand(f, 3) },
			func(f verrechnungsFall) error { return mitPfand(f, 2) }},
		{"anderer Verkauf",
			func(f verrechnungsFall) error {
				return f.cmd.DirektverkaufTaetigen(f.ctx, f.userID, "test", f.verkaufID, f.verkauf(1), "", &enrichment.PfandEingabe{RueckgabeID: f.rueckgabeID, Positionen: f.pfand(2)})
			},
			func(f verrechnungsFall) error { return mitPfand(f, 2) }},
		{"Verkauf ohne Rückgabe",
			func(f verrechnungsFall) error {
				return f.cmd.DirektverkaufTaetigen(f.ctx, f.userID, "test", f.verkaufID, f.verkauf(2), "", nil)
			},
			func(f verrechnungsFall) error { return mitPfand(f, 2) }},
		{"Rückgabe ohne Verkauf",
			func(f verrechnungsFall) error {
				return f.cmd.PfandZurueckgeben(f.ctx, f.userID, "test", f.rueckgabeID, 0, f.pfand(2), "")
			},
			func(f verrechnungsFall) error { return mitPfand(f, 2) }},
		{"Verkauf mit Rückgabe, Wiederholung ohne Pfand",
			func(f verrechnungsFall) error { return mitPfand(f, 2) },
			func(f verrechnungsFall) error {
				return f.cmd.DirektverkaufTaetigen(f.ctx, f.userID, "test", f.verkaufID, f.verkauf(2), "", nil)
			}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cmd, db, userID, _, produktID, varianteID := setupDVIntegration(t)
			pfandProduktID, pfandVarianteID := insertPfandVariante(t, db)
			f := verrechnungsFall{
				ctx: ctx, cmd: cmd, userID: userID,
				verkaufID: uuid.New().String(), rueckgabeID: uuid.New().String(),
				verkauf: func(menge int) []enrichment.PositionInput {
					return []enrichment.PositionInput{{ProduktID: produktID, VarianteID: varianteID, Menge: menge}}
				},
				pfand: func(menge int) []enrichment.PositionInput {
					return []enrichment.PositionInput{{ProduktID: pfandProduktID, VarianteID: pfandVarianteID, Menge: menge}}
				},
			}

			if err := tt.erst(f); err != nil {
				t.Fatalf("erster Vorgang: %v", err)
			}
			var vorher int
			if err := db.QueryRow("SELECT COUNT(*) FROM kassenjournal").Scan(&vorher); err != nil {
				t.Fatalf("kassenjournal zählen: %v", err)
			}

			if err := tt.retry(f); !errors.Is(err, ErrPfandrueckgabeAbweichend) {
				t.Errorf("erwartet ErrPfandrueckgabeAbweichend, bekam: %v", err)
			}

			var nachher int
			if err := db.QueryRow("SELECT COUNT(*) FROM kassenjournal").Scan(&nachher); err != nil {
				t.Fatalf("kassenjournal zählen: %v", err)
			}
			if nachher != vorher {
				t.Errorf("erwartet kein neues Event, kassenjournal wuchs von %d auf %d", vorher, nachher)
			}
		})
	}
}
