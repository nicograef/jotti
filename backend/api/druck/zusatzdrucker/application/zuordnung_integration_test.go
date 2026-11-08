//go:build integration

package application_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/nicograef/jotti/backend/api/druck/zusatzdrucker/application"
	"github.com/nicograef/jotti/backend/domain/druckstation"
)

// createProdukt inserts a Produkt of Kategorie essen and deletes it after the test: all integration
// packages share one database.
func createProdukt(t *testing.T, db *sql.DB, name, status string) int {
	t.Helper()
	var id int
	if err := db.QueryRow(
		"INSERT INTO produkte (name, kategorie, steuersatz, status, created_at, updated_at) VALUES ($1, 'essen', 'regel', $2, now(), now()) RETURNING id",
		name, status,
	).Scan(&id); err != nil {
		t.Fatalf("Failed to insert produkt: %v", err)
	}
	t.Cleanup(func() {
		if _, err := db.Exec("DELETE FROM produkte WHERE id = $1", id); err != nil {
			t.Errorf("Failed to delete produkt: %v", err)
		}
	})
	return id
}

func anlegen(t *testing.T, cmd application.Command, name string) int {
	t.Helper()
	id, err := cmd.Anlegen(context.Background(), name)
	if err != nil {
		t.Fatalf("Anlegen %q: %v", name, err)
	}
	return id
}

// produkteVon returns the Produkt-IDs listed under the Zusatzdrucker.
func produkteVon(t *testing.T, uebersicht application.Uebersicht, zusatzdruckerID int) []int {
	t.Helper()
	for _, zd := range uebersicht.Zusatzdrucker {
		if zd.ID == zusatzdruckerID {
			ids := make([]int, 0, len(zd.Produkte))
			for _, p := range zd.Produkte {
				ids = append(ids, p.ProduktID)
			}
			return ids
		}
	}
	t.Fatalf("Zusatzdrucker %d missing from Uebersicht", zusatzdruckerID)
	return nil
}

// auswahlEintrag returns the entry of the Produkt in ProduktAuswahl and whether it is listed.
func auswahlEintrag(uebersicht application.Uebersicht, produktID int) (druckstation.ProduktZuordnung, bool) {
	for _, p := range uebersicht.ProduktAuswahl {
		if p.ProduktID == produktID {
			return p, true
		}
	}
	return druckstation.ProduktZuordnung{}, false
}

func getAlle(t *testing.T, query application.Query) application.Uebersicht {
	t.Helper()
	uebersicht, err := query.GetAlle(context.Background())
	if err != nil {
		t.Fatalf("GetAlle: %v", err)
	}
	return uebersicht
}

func TestZuordnung_ErscheintUnterDruckerUndInAuswahl(t *testing.T) {
	cmd, repo, db := setup(t)
	query := application.Query{ZusatzdruckerRepo: repo}
	ctx := context.Background()

	kaffee := anlegen(t, cmd, "Kaffee & Kuchen")
	leer := anlegen(t, cmd, "Grill")
	kuchen := createProdukt(t, db, "Zuordnung-Kuchen", "active")
	torte := createProdukt(t, db, "Zuordnung-Torte", "inactive")
	wasser := createProdukt(t, db, "Zuordnung-Wasser", "active")

	if err := cmd.ProduktZuordnen(ctx, kuchen, kaffee); err != nil {
		t.Fatalf("ProduktZuordnen Kuchen: %v", err)
	}
	if err := cmd.ProduktZuordnen(ctx, torte, kaffee); err != nil {
		t.Fatalf("ProduktZuordnen Torte: %v", err)
	}

	uebersicht := getAlle(t, query)

	if got := produkteVon(t, uebersicht, kaffee); len(got) != 2 || got[0] != kuchen || got[1] != torte {
		t.Errorf("Produkte unter Kaffee & Kuchen = %v, want [%d %d]", got, kuchen, torte)
	}
	if got := produkteVon(t, uebersicht, leer); len(got) != 0 {
		t.Errorf("Produkte unter Grill = %v, want none", got)
	}

	eintrag, ok := auswahlEintrag(uebersicht, kuchen)
	want := druckstation.ProduktZuordnung{
		ProduktID: kuchen, ProduktName: "Zuordnung-Kuchen", Kategorie: "essen",
		ZusatzdruckerID: kaffee, ZusatzdruckerName: "Kaffee & Kuchen",
	}
	if !ok || eintrag != want {
		t.Errorf("Auswahl Kuchen = %+v (listed %v), want %+v", eintrag, ok, want)
	}
	eintrag, ok = auswahlEintrag(uebersicht, wasser)
	if !ok || eintrag.ZusatzdruckerID != 0 || eintrag.ZusatzdruckerName != "" {
		t.Errorf("Auswahl Wasser = %+v (listed %v), want listed without Zusatzdrucker", eintrag, ok)
	}
}

func TestZuordnung_ZweiteZuordnungWandert(t *testing.T) {
	cmd, repo, db := setup(t)
	query := application.Query{ZusatzdruckerRepo: repo}
	ctx := context.Background()

	kaffee := anlegen(t, cmd, "Kaffee & Kuchen")
	bar := anlegen(t, cmd, "Bar")
	kuchen := createProdukt(t, db, "Zuordnung-Kuchen", "active")

	if err := cmd.ProduktZuordnen(ctx, kuchen, kaffee); err != nil {
		t.Fatalf("ProduktZuordnen Kaffee & Kuchen: %v", err)
	}
	if err := cmd.ProduktZuordnen(ctx, kuchen, bar); err != nil {
		t.Fatalf("ProduktZuordnen Bar: %v", err)
	}

	uebersicht := getAlle(t, query)
	if got := produkteVon(t, uebersicht, kaffee); len(got) != 0 {
		t.Errorf("Produkte unter Kaffee & Kuchen = %v, want none", got)
	}
	if got := produkteVon(t, uebersicht, bar); len(got) != 1 || got[0] != kuchen {
		t.Errorf("Produkte unter Bar = %v, want [%d]", got, kuchen)
	}
	if eintrag, _ := auswahlEintrag(uebersicht, kuchen); eintrag.ZusatzdruckerName != "Bar" {
		t.Errorf("Auswahl Kuchen zusatzdruckerName = %q, want %q", eintrag.ZusatzdruckerName, "Bar")
	}
}

func TestZuordnungEntfernen_MitUndOhneZuordnung(t *testing.T) {
	cmd, repo, db := setup(t)
	query := application.Query{ZusatzdruckerRepo: repo}
	ctx := context.Background()

	kaffee := anlegen(t, cmd, "Kaffee & Kuchen")
	kuchen := createProdukt(t, db, "Zuordnung-Kuchen", "active")
	if err := cmd.ProduktZuordnen(ctx, kuchen, kaffee); err != nil {
		t.Fatalf("ProduktZuordnen: %v", err)
	}

	if err := cmd.ZuordnungEntfernen(ctx, kuchen); err != nil {
		t.Fatalf("ZuordnungEntfernen: %v", err)
	}
	if err := cmd.ZuordnungEntfernen(ctx, kuchen); err != nil {
		t.Errorf("ZuordnungEntfernen without Zuordnung: %v, want success", err)
	}

	uebersicht := getAlle(t, query)
	if got := produkteVon(t, uebersicht, kaffee); len(got) != 0 {
		t.Errorf("Produkte unter Kaffee & Kuchen = %v, want none", got)
	}
	if eintrag, ok := auswahlEintrag(uebersicht, kuchen); !ok || eintrag.ZusatzdruckerName != "" {
		t.Errorf("Auswahl Kuchen = %+v (listed %v), want listed without Zusatzdrucker", eintrag, ok)
	}
}

func TestZuordnung_GeloeschtesProduktUndUnbekannterDrucker(t *testing.T) {
	cmd, repo, db := setup(t)
	query := application.Query{ZusatzdruckerRepo: repo}
	ctx := context.Background()

	kaffee := anlegen(t, cmd, "Kaffee & Kuchen")
	kuchen := createProdukt(t, db, "Zuordnung-Kuchen", "active")
	if err := cmd.ProduktZuordnen(ctx, kuchen, kaffee); err != nil {
		t.Fatalf("ProduktZuordnen: %v", err)
	}
	if _, err := db.Exec("UPDATE produkte SET status = 'deleted' WHERE id = $1", kuchen); err != nil {
		t.Fatalf("Failed to delete produkt: %v", err)
	}
	geloescht := createProdukt(t, db, "Zuordnung-Alt", "deleted")
	aktiv := createProdukt(t, db, "Zuordnung-Wasser", "active")

	uebersicht := getAlle(t, query)
	if got := produkteVon(t, uebersicht, kaffee); len(got) != 0 {
		t.Errorf("Produkte unter Kaffee & Kuchen = %v, want none after soft delete", got)
	}
	if _, ok := auswahlEintrag(uebersicht, kuchen); ok {
		t.Errorf("Deleted Produkt %d listed in ProduktAuswahl", kuchen)
	}

	if err := cmd.ProduktZuordnen(ctx, geloescht, kaffee); !errors.Is(err, application.ErrZuordnungNichtMoeglich) {
		t.Errorf("ProduktZuordnen deleted Produkt = %v, want ErrZuordnungNichtMoeglich", err)
	}
	if err := cmd.ProduktZuordnen(ctx, aktiv, kaffee+1000); !errors.Is(err, application.ErrZuordnungNichtMoeglich) {
		t.Errorf("ProduktZuordnen unknown Zusatzdrucker = %v, want ErrZuordnungNichtMoeglich", err)
	}
}
