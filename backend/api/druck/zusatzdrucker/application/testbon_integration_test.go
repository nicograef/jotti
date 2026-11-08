//go:build integration

package application_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/nicograef/jotti/backend/api/druck/zusatzdrucker/application"
	"github.com/nicograef/jotti/backend/db/dbtest"
	"github.com/nicograef/jotti/backend/domain/druckstation"
	"github.com/nicograef/jotti/backend/repository/druckauftrag_repo"
	"github.com/nicograef/jotti/backend/repository/druckstation_repo"
)

// setup returns a Command over real repositories, the repository for reading rows back and the open
// sql.DB for assertions against druckauftraege. Zusatzdrucker and the outbox start and end empty.
func setup(t *testing.T) (application.Command, druckstation_repo.Repository, *sql.DB) {
	t.Helper()
	db := dbtest.Open()

	reset := func() {
		if _, err := db.Exec("DELETE FROM zusatzdrucker"); err != nil {
			t.Fatalf("Failed to reset zusatzdrucker: %v", err)
		}
		if _, err := db.Exec("DELETE FROM druckauftraege"); err != nil {
			t.Fatalf("Failed to reset druckauftraege: %v", err)
		}
	}

	reset()
	t.Cleanup(func() {
		reset()
		_ = db.Close()
	})

	repo := druckstation_repo.NewRepository(db)
	return application.Command{
		ZusatzdruckerRepo: repo,
		DruckauftragRepo:  druckauftrag_repo.NewRepository(db),
	}, repo, db
}

func TestTestbonDrucken_ReihtAuftragAnZusatzdruckerEin(t *testing.T) {
	cmd, _, db := setup(t)
	ctx := context.Background()

	id, err := cmd.Anlegen(ctx, "Kaffee & Kuchen")
	if err != nil {
		t.Fatalf("Anlegen: %v", err)
	}
	if err := cmd.Aendern(ctx, id, "Kaffee & Kuchen", "192.168.1.70", "pro_position"); err != nil {
		t.Fatalf("Aendern: %v", err)
	}

	if err := cmd.TestbonDrucken(ctx, id); err != nil {
		t.Fatalf("TestbonDrucken: %v", err)
	}

	var count int
	var bonArt, referenz, zielIP, status string
	row := db.QueryRow(
		"SELECT count(*), max(bon_art), max(referenz), max(ziel_ip), max(status) FROM druckauftraege",
	)
	if err := row.Scan(&count, &bonArt, &referenz, &zielIP, &status); err != nil {
		t.Fatalf("Failed to read druckauftraege: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected exactly 1 auftrag, got %d", count)
	}
	if bonArt != "testbon" {
		t.Errorf("Expected bon_art 'testbon', got %q", bonArt)
	}
	if referenz != "testdruck:Kaffee & Kuchen" {
		t.Errorf("Expected referenz 'testdruck:Kaffee & Kuchen', got %q", referenz)
	}
	if zielIP != "192.168.1.70" {
		t.Errorf("Expected ziel_ip '192.168.1.70', got %q", zielIP)
	}
	if status != "offen" {
		t.Errorf("Expected status 'offen', got %q", status)
	}
}

func TestTestbonDrucken_ZusatzdruckerOhneIPWirdAbgelehnt(t *testing.T) {
	cmd, repo, db := setup(t)
	ctx := context.Background()

	id, err := cmd.Anlegen(ctx, "Kaffee & Kuchen")
	if err != nil {
		t.Fatalf("Anlegen: %v", err)
	}
	got, err := repo.GetZusatzdrucker(ctx, id)
	if err != nil {
		t.Fatalf("GetZusatzdrucker: %v", err)
	}
	want := druckstation.Zusatzdrucker{ID: id, Name: "Kaffee & Kuchen", DruckerIP: "", Bonmodus: druckstation.BonmodusProPosition}
	if got != want {
		t.Errorf("Anlegen stored %+v, want %+v", got, want)
	}

	err = cmd.TestbonDrucken(ctx, id)
	if !errors.Is(err, application.ErrZusatzdruckerNichtKonfiguriert) {
		t.Errorf("Expected ErrZusatzdruckerNichtKonfiguriert, got %v", err)
	}

	var count int
	if err := db.QueryRow("SELECT count(*) FROM druckauftraege").Scan(&count); err != nil {
		t.Fatalf("Failed to count druckauftraege: %v", err)
	}
	if count != 0 {
		t.Errorf("Expected no druckauftraege after rejected testbon, got %d", count)
	}
}
