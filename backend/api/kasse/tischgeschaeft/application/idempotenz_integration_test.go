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
	"github.com/nicograef/jotti/backend/db/dbtest"
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
	db = dbtest.Open()
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
		t.Errorf("zweiter Aufruf (Duplikat) erwartet nil, bekam: %v", err)
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

// TestZahlungKassieren_MitPfand_DuplikatRueckgabeId_IdempotenterErfolg: a settled payment sent twice
// stores the payment and the Rückgabe once each, with one Signaturauftrag per Vorgang.
func TestZahlungKassieren_MitPfand_DuplikatRueckgabeId_IdempotenterErfolg(t *testing.T) {
	ctx, cmd, db, userID, ksNr, tischID, produktID, varianteID := setupBestellungIntegration(t)

	pfandProduktID, pfandVarianteID := insertPfandVariante(t, db)

	inputs := []enrichment.PositionInput{{ProduktID: produktID, VarianteID: varianteID, Menge: 2}}
	if err := cmd.BestellungAufnehmen(ctx, userID, "test", uuid.New().String(), tischID, inputs, ""); err != nil {
		t.Fatalf("Bestellung aufnehmen: %v", err)
	}
	session, err := kassenjournal_repo.NewRepository(db).ReadTischSession(ctx, kasse.TischSessionSubject(ksNr, tischID))
	if err != nil {
		t.Fatalf("Tisch-Session lesen: %v", err)
	}
	refs := []kasse.PositionRef{{PositionID: session.UnbezahltePositionen[0].PositionID, Menge: 2}}

	pfand := &enrichment.PfandEingabe{
		RueckgabeID: uuid.New().String(),
		Positionen:  []enrichment.PositionInput{{ProduktID: pfandProduktID, VarianteID: pfandVarianteID, Menge: 2}},
	}

	zahlungID := uuid.NewString()
	if err := cmd.ZahlungKassieren(ctx, userID, "test", zahlungID, tischID, refs, "", pfand); err != nil {
		t.Fatalf("erster Aufruf: %v", err)
	}
	if err := cmd.ZahlungKassieren(ctx, userID, "test", zahlungID, tischID, refs, "", pfand); err != nil {
		t.Errorf("zweiter Aufruf (Duplikat) erwartet nil, bekam: %v", err)
	}

	var zahlungen, rueckgaben int
	if err := db.QueryRow("SELECT COUNT(*) FILTER (WHERE type = 'zahlung-kassiert:v2'), COUNT(*) FILTER (WHERE type = 'pfand-zurueckgegeben:v1') FROM kassenjournal").Scan(&zahlungen, &rueckgaben); err != nil {
		t.Fatalf("kassenjournal zählen: %v", err)
	}
	if zahlungen != 1 || rueckgaben != 1 {
		t.Errorf("erwartet je 1 Zahlung und 1 Rückgabe, gespeichert: %d und %d", zahlungen, rueckgaben)
	}

	var rueckgabeTischID int
	if err := db.QueryRow("SELECT (data->>'tischId')::int FROM kassenjournal WHERE type = 'pfand-zurueckgegeben:v1' AND subject = $1 AND version = 1", kasse.PfandrueckgabeSubject(ksNr, pfand.RueckgabeID)).Scan(&rueckgabeTischID); err != nil {
		t.Fatalf("Rückgabe im eigenen Stream lesen: %v", err)
	}
	if rueckgabeTischID != tischID {
		t.Errorf("erwartet tischId %d an der Rückgabe, bekam %d", tischID, rueckgabeTischID)
	}

	var auftraegeCount int
	if err := db.QueryRow("SELECT COUNT(*) FROM tse_signaturauftraege a JOIN kassenjournal k ON k.id = a.event_id WHERE k.type IN ('zahlung-kassiert:v2', 'pfand-zurueckgegeben:v1')").Scan(&auftraegeCount); err != nil {
		t.Fatalf("tse_signaturauftraege zählen: %v", err)
	}
	if auftraegeCount != 2 {
		t.Errorf("erwartet 2 tse_signaturauftraege, vorhanden: %d", auftraegeCount)
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

// zahlungsFall carries what a case of TestZahlungKassieren_MitPfand_Abweichend_KeinEvent needs to
// book its first Vorgang.
type zahlungsFall struct {
	ctx                   context.Context
	cmd                   Command
	journal               kassenjournal_repo.Repository
	userID, ksNr, tischID int
	refs                  func(menge int) []kasse.PositionRef
	pfand                 func(menge int) *enrichment.PfandEingabe
}

// TestZahlungKassieren_MitPfand_Abweichend_KeinEvent: a settled payment whose rueckgabeId is booked
// with other content, or without the payment it was settled with, is refused with
// ErrPfandrueckgabeAbweichend and books nothing.
func TestZahlungKassieren_MitPfand_Abweichend_KeinEvent(t *testing.T) {
	// Each case books a first Vorgang; then follows the payment of one Bier with two Gläser Pfand.
	tests := []struct {
		name string
		erst func(t *testing.T, f zahlungsFall) error
	}{
		{"andere Pfand-Menge", func(_ *testing.T, f zahlungsFall) error {
			return f.cmd.ZahlungKassieren(f.ctx, f.userID, "test", uuid.NewString(), f.tischID, f.refs(1), "", f.pfand(3))
		}},
		{"andere Zahlung", func(_ *testing.T, f zahlungsFall) error {
			return f.cmd.ZahlungKassieren(f.ctx, f.userID, "test", uuid.NewString(), f.tischID, f.refs(2), "", f.pfand(2))
		}},
		{"Rückgabe ohne Zahlung", func(t *testing.T, f zahlungsFall) error {
			eingabe := f.pfand(2)
			positionen, err := enrichment.EnrichPfandPositionen(f.ctx, f.cmd.ProduktRepo, eingabe.Positionen)
			if err != nil {
				t.Fatalf("Pfand anreichern: %v", err)
			}
			rueckgabe, err := kasse.NewPfandZurueckgegebenEvent(kasse.PfandrueckgabeSubject(f.ksNr, eingabe.RueckgabeID), eingabe.RueckgabeID, f.tischID, f.userID, "test", positionen, "")
			if err != nil {
				t.Fatalf("Rückgabe-Event: %v", err)
			}
			rueckgabe.Version = 1
			_, err = f.journal.WriteEvent(f.ctx, rueckgabe, kasse.StreamTypePfandrueckgabe, f.ksNr)
			return err
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cmd, db, userID, ksNr, tischID, produktID, varianteID := setupBestellungIntegration(t)
			pfandProduktID, pfandVarianteID := insertPfandVariante(t, db)

			inputs := []enrichment.PositionInput{{ProduktID: produktID, VarianteID: varianteID, Menge: 2}}
			if err := cmd.BestellungAufnehmen(ctx, userID, "test", uuid.New().String(), tischID, inputs, ""); err != nil {
				t.Fatalf("Bestellung aufnehmen: %v", err)
			}
			session, err := kassenjournal_repo.NewRepository(db).ReadTischSession(ctx, kasse.TischSessionSubject(ksNr, tischID))
			if err != nil {
				t.Fatalf("Tisch-Session lesen: %v", err)
			}
			positionID := session.UnbezahltePositionen[0].PositionID
			rueckgabeID := uuid.New().String()
			f := zahlungsFall{
				ctx: ctx, cmd: cmd, journal: kassenjournal_repo.NewRepository(db), userID: userID, ksNr: ksNr, tischID: tischID,
				refs: func(menge int) []kasse.PositionRef {
					return []kasse.PositionRef{{PositionID: positionID, Menge: menge}}
				},
				pfand: func(menge int) *enrichment.PfandEingabe {
					return &enrichment.PfandEingabe{
						RueckgabeID: rueckgabeID,
						Positionen:  []enrichment.PositionInput{{ProduktID: pfandProduktID, VarianteID: pfandVarianteID, Menge: menge}},
					}
				},
			}

			if err := tt.erst(t, f); err != nil {
				t.Fatalf("erster Vorgang: %v", err)
			}
			var vorher int
			if err := db.QueryRow("SELECT COUNT(*) FROM kassenjournal").Scan(&vorher); err != nil {
				t.Fatalf("kassenjournal zählen: %v", err)
			}

			if err := cmd.ZahlungKassieren(ctx, userID, "test", uuid.NewString(), tischID, f.refs(1), "", f.pfand(2)); !errors.Is(err, ErrPfandrueckgabeAbweichend) {
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

// TestTischVorgang_Wiederholung: a Vorgang retried with its key after a lost response books nothing
// again, even for a partial quantity whose remainder still validates; the same key with other
// content is refused with ErrVorgangAbweichend. Each case starts from four unpaid Bier, two of them
// paid, so a Storno of three writes a Korrektur and a Warenrücknahme.
func TestTischVorgang_Wiederholung(t *testing.T) {
	tests := []struct {
		name    string
		vorgang func(f tischVorgangFall, schluessel uuid.UUID, menge int) error
	}{
		{"Zahlung", func(f tischVorgangFall, schluessel uuid.UUID, menge int) error {
			return f.cmd.ZahlungKassieren(f.ctx, f.userID, "test", schluessel.String(), f.tischID, f.refs(menge), "", nil)
		}},
		{"Stornierung", func(f tischVorgangFall, schluessel uuid.UUID, menge int) error {
			return f.cmd.StornierungErteilen(f.ctx, f.userID, "test", schluessel, f.tischID, f.refs(menge+2), "Reklamation")
		}},
		{"Umbuchung", func(f tischVorgangFall, schluessel uuid.UUID, menge int) error {
			return f.cmd.BestellungUmbuchen(f.ctx, f.userID, "test", schluessel.String(), f.tischID, f.zielTischID, f.refs(menge), "")
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := setupTischVorgangFall(t)
			schluessel := uuid.New()

			if err := tt.vorgang(f, schluessel, 1); err != nil {
				t.Fatalf("erster Versuch: %v", err)
			}
			vorher := f.anzahlEvents(t)

			if err := tt.vorgang(f, schluessel, 1); err != nil {
				t.Errorf("Wiederholung erwartet nil, bekam: %v", err)
			}
			if err := tt.vorgang(f, schluessel, 2); !errors.Is(err, ErrVorgangAbweichend) {
				t.Errorf("anderer Inhalt erwartet ErrVorgangAbweichend, bekam: %v", err)
			}
			if nachher := f.anzahlEvents(t); nachher != vorher {
				t.Errorf("erwartet kein neues Event, kassenjournal wuchs von %d auf %d", vorher, nachher)
			}
		})
	}
}

type tischVorgangFall struct {
	ctx             context.Context
	cmd             Command
	db              *sql.DB
	userID, tischID int
	zielTischID     int
	ksNr            int
	refs            func(menge int) []kasse.PositionRef
	bier            func(menge int) []enrichment.PositionInput
}

func (f tischVorgangFall) anzahlEvents(t *testing.T) int {
	t.Helper()
	var anzahl int
	if err := f.db.QueryRow("SELECT COUNT(*) FROM kassenjournal").Scan(&anzahl); err != nil {
		t.Fatalf("kassenjournal zählen: %v", err)
	}
	return anzahl
}

func setupTischVorgangFall(t *testing.T) tischVorgangFall {
	t.Helper()
	ctx, cmd, db, userID, ksNr, tischID, produktID, varianteID := setupBestellungIntegration(t)

	var zielTischID int
	if err := db.QueryRow(
		"INSERT INTO tische (name, status, created_at, updated_at) VALUES ('Tisch 2', 'active', now(), now()) RETURNING id",
	).Scan(&zielTischID); err != nil {
		t.Fatalf("create ziel-tisch: %v", err)
	}

	bier := func(menge int) []enrichment.PositionInput {
		return []enrichment.PositionInput{{ProduktID: produktID, VarianteID: varianteID, Menge: menge}}
	}
	if err := cmd.BestellungAufnehmen(ctx, userID, "test", uuid.NewString(), tischID, bier(4), ""); err != nil {
		t.Fatalf("Bestellung aufnehmen: %v", err)
	}
	session, err := kassenjournal_repo.NewRepository(db).ReadTischSession(ctx, kasse.TischSessionSubject(ksNr, tischID))
	if err != nil {
		t.Fatalf("Tisch-Session lesen: %v", err)
	}
	positionID := session.UnbezahltePositionen[0].PositionID
	refs := func(menge int) []kasse.PositionRef {
		return []kasse.PositionRef{{PositionID: positionID, Menge: menge}}
	}
	if err := cmd.ZahlungKassieren(ctx, userID, "test", uuid.NewString(), tischID, refs(2), "", nil); err != nil {
		t.Fatalf("zwei Bier kassieren: %v", err)
	}

	return tischVorgangFall{ctx: ctx, cmd: cmd, db: db, userID: userID, tischID: tischID, zielTischID: zielTischID, ksNr: ksNr, refs: refs, bier: bier}
}

// An umbuchungId is the Abgang's BON_ID, so the BON_ID index admits it once across all Vorgänge: the
// key of another Beleg, or of an Umbuchung from another table, books nothing.
func TestBestellungUmbuchen_FremdeBonID_KeinEvent(t *testing.T) {
	tests := []struct {
		name      string
		vorbuchen func(f tischVorgangFall, schluessel uuid.UUID)
	}{
		{"Zahlung", func(f tischVorgangFall, schluessel uuid.UUID) {
			if err := f.cmd.ZahlungKassieren(f.ctx, f.userID, "test", schluessel.String(), f.tischID, f.refs(1), "", nil); err != nil {
				t.Fatalf("Zahlung: %v", err)
			}
		}},
		{"Umbuchung eines anderen Tischs", func(f tischVorgangFall, schluessel uuid.UUID) {
			var andererTischID int
			if err := f.db.QueryRow(
				"INSERT INTO tische (name, status, created_at, updated_at) VALUES ('Tisch 3', 'active', now(), now()) RETURNING id",
			).Scan(&andererTischID); err != nil {
				t.Fatalf("create tisch 3: %v", err)
			}
			if err := f.cmd.BestellungAufnehmen(f.ctx, f.userID, "test", uuid.NewString(), andererTischID, f.bier(1), ""); err != nil {
				t.Fatalf("Bestellung Tisch 3: %v", err)
			}
			session, err := kassenjournal_repo.NewRepository(f.db).ReadTischSession(f.ctx, kasse.TischSessionSubject(f.ksNr, andererTischID))
			if err != nil {
				t.Fatalf("Tisch-Session lesen: %v", err)
			}
			refs := []kasse.PositionRef{{PositionID: session.UnbezahltePositionen[0].PositionID, Menge: 1}}
			if err := f.cmd.BestellungUmbuchen(f.ctx, f.userID, "test", schluessel.String(), andererTischID, f.zielTischID, refs, ""); err != nil {
				t.Fatalf("Umbuchung Tisch 3: %v", err)
			}
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := setupTischVorgangFall(t)
			schluessel := uuid.New()
			tt.vorbuchen(f, schluessel)
			vorher := f.anzahlEvents(t)

			err := f.cmd.BestellungUmbuchen(f.ctx, f.userID, "test", schluessel.String(), f.tischID, f.zielTischID, f.refs(1), "")
			if !errors.Is(err, ErrConflict) {
				t.Errorf("erwartet ErrConflict, bekam: %v", err)
			}
			if nachher := f.anzahlEvents(t); nachher != vorher {
				t.Errorf("erwartet kein neues Event, kassenjournal wuchs von %d auf %d", vorher, nachher)
			}
		})
	}
}
