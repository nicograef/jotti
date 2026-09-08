//go:build integration

package application_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/nicograef/jotti/backend/api/druck/station/application"
	dbpkg "github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/druckstation"
	"github.com/nicograef/jotti/backend/repository/druckauftrag_repo"
	"github.com/nicograef/jotti/backend/repository/druckstation_repo"
)

// setup gibt ein Command mit echten Repos, das offene sql.DB (für direkte
// Assertions gegen druckauftraege) und ein Teardown zurück. Die Druckstationen
// werden auf den Default (leere IP) zurückgesetzt und die Outbox geleert.
func setup(t *testing.T) (application.Command, *sql.DB, func()) {
	t.Helper()
	db := dbpkg.OpenTestDatabase()

	resetStationen := func() {
		if _, err := db.Exec("UPDATE druckstationen SET drucker_ip = '', bonmodus = 'pro_position' WHERE kategorie IN ('essen', 'getraenk', 'sonstiges')"); err != nil {
			t.Fatalf("Failed to reset produktstationen: %v", err)
		}
		if _, err := db.Exec("UPDATE druckstationen SET drucker_ip = '', bonmodus = 'pro_bestellung' WHERE kategorie = 'abholbon'"); err != nil {
			t.Fatalf("Failed to reset abholbon: %v", err)
		}
		if _, err := db.Exec("UPDATE druckstationen SET drucker_ip = '', bonmodus = NULL WHERE kategorie = 'kassenbeleg'"); err != nil {
			t.Fatalf("Failed to reset kassenbeleg: %v", err)
		}
	}
	resetAuftraege := func() {
		if _, err := db.Exec("DELETE FROM druckauftraege"); err != nil {
			t.Fatalf("Failed to reset druckauftraege: %v", err)
		}
	}

	resetStationen()
	resetAuftraege()

	cmd := application.Command{
		DruckstationRepo: druckstation_repo.NewRepository(db),
		DruckauftragRepo: druckauftrag_repo.NewRepository(db),
	}

	return cmd, db, func() {
		resetStationen()
		resetAuftraege()
		_ = db.Close()
	}
}

func TestTestbonDrucken_ReihtAuftragEin(t *testing.T) {
	cmd, db, teardown := setup(t)
	defer teardown()

	ctx := context.Background()

	// Station „Essen" mit Drucker-IP konfigurieren.
	if err := druckstation_repo.NewRepository(db).UpsertDruckstation(ctx, druckstation.Druckstation{
		Kategorie: druckstation.KategorieEssen,
		DruckerIP: "192.168.1.50",
		Bonmodus:  druckstation.BonmodusProPosition,
	}); err != nil {
		t.Fatalf("Failed to configure station: %v", err)
	}

	if err := cmd.TestbonDrucken(ctx, "essen"); err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	var count int
	var bonArt, referenz, zielIP, status string
	row := db.QueryRow(
		"SELECT count(*), max(bon_art), max(referenz), max(ziel_ip), max(status) FROM druckauftraege WHERE bon_art = 'testbon'",
	)
	if err := row.Scan(&count, &bonArt, &referenz, &zielIP, &status); err != nil {
		t.Fatalf("Failed to read druckauftraege: %v", err)
	}
	if count != 1 {
		t.Fatalf("Expected exactly 1 testbon auftrag, got %d", count)
	}
	if bonArt != "testbon" {
		t.Errorf("Expected bon_art 'testbon', got %q", bonArt)
	}
	if referenz != "testdruck:essen" {
		t.Errorf("Expected referenz 'testdruck:essen', got %q", referenz)
	}
	if zielIP != "192.168.1.50" {
		t.Errorf("Expected ziel_ip '192.168.1.50', got %q", zielIP)
	}
	if status != "offen" {
		t.Errorf("Expected status 'offen', got %q", status)
	}
}

func TestTestbonDrucken_OhneIPWirdAbgelehnt(t *testing.T) {
	cmd, db, teardown := setup(t)
	defer teardown()

	ctx := context.Background()

	// Keine Station konfiguriert -> Fehler, kein Auftrag.
	err := cmd.TestbonDrucken(ctx, "essen")
	if !errors.Is(err, application.ErrDruckstationNichtKonfiguriert) {
		t.Fatalf("Expected ErrDruckstationNichtKonfiguriert, got %v", err)
	}

	var count int
	if err := db.QueryRow("SELECT count(*) FROM druckauftraege").Scan(&count); err != nil {
		t.Fatalf("Failed to count druckauftraege: %v", err)
	}
	if count != 0 {
		t.Errorf("Expected no druckauftraege after rejected testbon, got %d", count)
	}
}
