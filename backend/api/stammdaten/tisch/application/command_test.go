//go:build unit

package application

import (
	"context"
	"testing"
	"time"

	"github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/produkt"
	"github.com/nicograef/jotti/backend/domain/tisch"
	"github.com/nicograef/jotti/backend/repository/favorit_repo"
	"github.com/nicograef/jotti/backend/repository/tisch_repo"
)

func newTestCommand(tables []tisch.Tisch, _ []produkt.Produkt) Command {
	return Command{
		TischRepo:   tisch_repo.NewMock(tables, nil),
		FavoritRepo: favorit_repo.NewMock(nil, nil),
	}
}

func TestTischErstellen(t *testing.T) {
	ctx := context.Background()
	command := newTestCommand(nil, nil)

	tischId, err := command.TischErstellen(ctx, "Tisch 1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if tischId != 1 {
		t.Errorf("expected tisch ID 1, got %d", tischId)
	}

	tisch, err := command.TischRepo.GetTable(ctx, tischId)
	if err != nil {
		t.Fatalf("expected no error retrieving tisch, got %v", err)
	}
	if tisch.Name != "Tisch 1" {
		t.Errorf("expected tisch name 'Tisch 1', got %s", tisch.Name)
	}
}

func TestTischErstellen_Error(t *testing.T) {
	repo := tisch_repo.NewMock([]tisch.Tisch{}, db.ErrAlreadyExists)
	command := Command{TischRepo: repo}

	_, err := command.TischErstellen(context.Background(), "Tisch 1")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func TestTischAktualisieren(t *testing.T) {
	repo := tisch_repo.NewMock([]tisch.Tisch{{ID: 1, Name: "Old Name", Status: tisch.ActiveStatus, UpdatedAt: time.Now().UTC()}}, nil)
	command := Command{TischRepo: repo}

	err := command.TischAktualisieren(context.Background(), 1, "New Name")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	tisch, err := command.TischRepo.GetTable(context.Background(), 1)
	if err != nil {
		t.Fatalf("expected no error retrieving tisch, got %v", err)
	}
	if tisch.Name != "New Name" {
		t.Errorf("expected tisch name to be 'New Name', got %s", tisch.Name)
	}
}

func TestTischAktualisieren_NotFound(t *testing.T) {
	repo := tisch_repo.NewMock([]tisch.Tisch{}, db.ErrNotFound)
	command := Command{TischRepo: repo}

	err := command.TischAktualisieren(context.Background(), 999, "New Name")
	if err != ErrTischNotFound {
		t.Fatalf("expected ErrTischNotFound, got %v", err)
	}
}

func TestTischAktivieren(t *testing.T) {
	repo := tisch_repo.NewMock([]tisch.Tisch{{ID: 1, Name: "Tisch 1", Status: tisch.InactiveStatus, UpdatedAt: time.Now().UTC()}}, nil)
	command := Command{TischRepo: repo}

	err := command.TischAktivieren(context.Background(), 1)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	tbl, err := repo.GetTable(context.Background(), 1)
	if err != nil {
		t.Fatalf("expected no error retrieving tisch, got %v", err)
	}
	if tbl.Status != tisch.ActiveStatus {
		t.Errorf("expected tisch status to be Active, got %v", tbl.Status)
	}
}

func TestTischAktivieren_NotFound(t *testing.T) {
	repo := tisch_repo.NewMock([]tisch.Tisch{}, db.ErrNotFound)
	command := Command{TischRepo: repo}

	err := command.TischAktivieren(context.Background(), 999)
	if err != ErrTischNotFound {
		t.Fatalf("expected ErrTischNotFound, got %v", err)
	}
}

func TestTischDeaktivieren(t *testing.T) {
	repo := tisch_repo.NewMock([]tisch.Tisch{{ID: 1, Name: "Tisch 1", Status: tisch.ActiveStatus, UpdatedAt: time.Now().UTC()}}, nil)
	command := Command{TischRepo: repo}

	err := command.TischDeaktivieren(context.Background(), 1)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	tbl, err := repo.GetTable(context.Background(), 1)
	if err != nil {
		t.Fatalf("expected no error retrieving tisch, got %v", err)
	}
	if tbl.Status != tisch.InactiveStatus {
		t.Errorf("expected tisch status to be Inactive, got %v", tbl.Status)
	}
}

func TestTischDeaktivieren_NotFound(t *testing.T) {
	repo := tisch_repo.NewMock([]tisch.Tisch{}, db.ErrNotFound)
	command := Command{TischRepo: repo}

	err := command.TischDeaktivieren(context.Background(), 999)
	if err != ErrTischNotFound {
		t.Fatalf("expected ErrTischNotFound, got %v", err)
	}
}

func TestTischDeaktivieren_SaldoOffen(t *testing.T) {
	repo := tisch_repo.NewMock([]tisch.Tisch{{ID: 1, Name: "Tisch 1", Status: tisch.ActiveStatus, UpdatedAt: time.Now().UTC()}}, nil)
	repo.SetOffenerSaldo(1, 9850)
	command := Command{TischRepo: repo}

	err := command.TischDeaktivieren(context.Background(), 1)
	if err != ErrTischSaldoOffen {
		t.Fatalf("expected ErrTischSaldoOffen, got %v", err)
	}

	tbl, err := repo.GetTable(context.Background(), 1)
	if err != nil {
		t.Fatalf("expected no error retrieving tisch, got %v", err)
	}
	if tbl.Status != tisch.ActiveStatus {
		t.Errorf("tisch must stay active when it has an open saldo, got status %v", tbl.Status)
	}
}

func TestTischLoeschen_OhneSaldo(t *testing.T) {
	repo := tisch_repo.NewMock([]tisch.Tisch{{ID: 1, Name: "Tisch 1", Status: tisch.ActiveStatus, UpdatedAt: time.Now().UTC()}}, nil)
	command := Command{TischRepo: repo, FavoritRepo: favorit_repo.NewMock(nil, nil)}

	err := command.TischLoeschen(context.Background(), 1)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	tbl, err := repo.GetTable(context.Background(), 1)
	if err != nil {
		t.Fatalf("expected no error retrieving tisch, got %v", err)
	}
	if tbl.Status != tisch.DeletedStatus {
		t.Errorf("expected tisch status to be Deleted, got %v", tbl.Status)
	}
}

// Ein gelöschter Tisch verschwindet aus der Tischauswahl; seine Markierungen
// müssen mit ihm gehen, sonst hängen sie unabwählbar in der Tischübersicht der
// betroffenen Servicekräfte. Markierungen anderer Tische bleiben unberührt.
func TestTischLoeschen_EntferntFavoriten(t *testing.T) {
	repo := tisch_repo.NewMock([]tisch.Tisch{{ID: 1, Name: "Tisch 1", Status: tisch.ActiveStatus, UpdatedAt: time.Now().UTC()}}, nil)
	favoriten := favorit_repo.NewMock(map[int][]int{5: {1, 2}, 6: {1}}, nil)
	repo.SetFavoritenCleanup(favoriten.RemoveByTisch)
	command := Command{TischRepo: repo, FavoritRepo: favoriten}

	if err := command.TischLoeschen(context.Background(), 1); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	fuenf, err := favoriten.GetByUser(context.Background(), 5)
	if err != nil {
		t.Fatalf("expected no error retrieving favoriten, got %v", err)
	}
	if len(fuenf) != 1 || fuenf[0] != 2 {
		t.Errorf("expected only tisch 2 to remain for user 5, got %v", fuenf)
	}

	sechs, err := favoriten.GetByUser(context.Background(), 6)
	if err != nil {
		t.Fatalf("expected no error retrieving favoriten, got %v", err)
	}
	if len(sechs) != 0 {
		t.Errorf("expected no favoriten for user 6, got %v", sechs)
	}
}

// Deaktivieren ist kein Löschen: der Tisch kommt wieder, die Markierung bleibt.
func TestTischDeaktivieren_BehaeltFavoriten(t *testing.T) {
	repo := tisch_repo.NewMock([]tisch.Tisch{{ID: 1, Name: "Tisch 1", Status: tisch.ActiveStatus, UpdatedAt: time.Now().UTC()}}, nil)
	favoriten := favorit_repo.NewMock(map[int][]int{5: {1}}, nil)
	repo.SetFavoritenCleanup(favoriten.RemoveByTisch)
	command := Command{TischRepo: repo, FavoritRepo: favoriten}

	if err := command.TischDeaktivieren(context.Background(), 1); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	fuenf, err := favoriten.GetByUser(context.Background(), 5)
	if err != nil {
		t.Fatalf("expected no error retrieving favoriten, got %v", err)
	}
	if len(fuenf) != 1 || fuenf[0] != 1 {
		t.Errorf("expected tisch 1 to stay marked for user 5, got %v", fuenf)
	}
}

// Statuswechsel und Favoriten-Cleanup teilen sich eine Transaktion: Scheitert
// der Cleanup, bleibt der Tisch aktiv. Es entsteht nie ein gelöschter Tisch mit
// zurückgebliebenen — unsichtbaren und unabwählbaren — Markierungen.
func TestTischLoeschen_FavoritenCleanupFehlschlag(t *testing.T) {
	repo := tisch_repo.NewMock([]tisch.Tisch{{ID: 1, Name: "Tisch 1", Status: tisch.ActiveStatus, UpdatedAt: time.Now().UTC()}}, nil)
	favoriten := favorit_repo.NewMock(map[int][]int{5: {1}}, db.ErrDatabase)
	repo.SetFavoritenCleanup(favoriten.RemoveByTisch)
	command := Command{TischRepo: repo, FavoritRepo: favoriten}

	if err := command.TischLoeschen(context.Background(), 1); err != ErrDatabase {
		t.Fatalf("expected ErrDatabase, got %v", err)
	}

	tbl, err := repo.GetTable(context.Background(), 1)
	if err != nil {
		t.Fatalf("expected no error retrieving tisch, got %v", err)
	}
	if tbl.Status != tisch.ActiveStatus {
		t.Errorf("tisch must stay active when the favoriten cleanup fails, got status %v", tbl.Status)
	}
}

func TestTischLoeschen_SaldoOffen(t *testing.T) {
	repo := tisch_repo.NewMock([]tisch.Tisch{{ID: 1, Name: "Tisch 1", Status: tisch.ActiveStatus, UpdatedAt: time.Now().UTC()}}, nil)
	repo.SetOffenerSaldo(1, 9850)
	favoriten := favorit_repo.NewMock(map[int][]int{5: {1}}, nil)
	command := Command{TischRepo: repo, FavoritRepo: favoriten}

	err := command.TischLoeschen(context.Background(), 1)
	if err != ErrTischSaldoOffen {
		t.Fatalf("expected ErrTischSaldoOffen, got %v", err)
	}

	tbl, err := repo.GetTable(context.Background(), 1)
	if err != nil {
		t.Fatalf("expected no error retrieving tisch, got %v", err)
	}
	if tbl.Status != tisch.ActiveStatus {
		t.Errorf("tisch must not be deleted when it has an open saldo, got status %v", tbl.Status)
	}

	fuenf, err := favoriten.GetByUser(context.Background(), 5)
	if err != nil {
		t.Fatalf("expected no error retrieving favoriten, got %v", err)
	}
	if len(fuenf) != 1 {
		t.Errorf("favoriten must stay untouched when the saldo guard rejects the delete, got %v", fuenf)
	}
}

// TestTischDeaktivieren_OhneOffeneSitzung bestätigt, dass ohne offene
// Kassensitzung (leere Saldo-Map) kein Tisch geschützt ist.
func TestTischDeaktivieren_OhneOffeneSitzung(t *testing.T) {
	repo := tisch_repo.NewMock([]tisch.Tisch{{ID: 1, Name: "Tisch 1", Status: tisch.ActiveStatus, UpdatedAt: time.Now().UTC()}}, nil)
	command := Command{TischRepo: repo}

	err := command.TischDeaktivieren(context.Background(), 1)
	if err != nil {
		t.Fatalf("expected no error without open session, got %v", err)
	}

	tbl, _ := repo.GetTable(context.Background(), 1)
	if tbl.Status != tisch.InactiveStatus {
		t.Errorf("expected tisch status to be Inactive, got %v", tbl.Status)
	}
}
