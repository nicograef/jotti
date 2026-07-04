//go:build unit

package application

import (
	"context"
	"testing"
	"time"

	"github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/product"
	"github.com/nicograef/jotti/backend/domain/table"
	"github.com/nicograef/jotti/backend/repository/table_repo"
)

func newTestCommand(tables []table.Tisch, _ []product.Produkt) Command {
	return Command{
		TableRepo: table_repo.NewMock(tables, nil),
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

	tisch, err := command.TableRepo.GetTable(ctx, tischId)
	if err != nil {
		t.Fatalf("expected no error retrieving tisch, got %v", err)
	}
	if tisch.Name != "Tisch 1" {
		t.Errorf("expected tisch name 'Tisch 1', got %s", tisch.Name)
	}
}

func TestTischErstellen_Error(t *testing.T) {
	repo := table_repo.NewMock([]table.Tisch{}, db.ErrAlreadyExists)
	command := Command{TableRepo: repo}

	_, err := command.TischErstellen(context.Background(), "Tisch 1")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func TestTischAktualisieren(t *testing.T) {
	repo := table_repo.NewMock([]table.Tisch{{ID: 1, Name: "Old Name", Status: table.ActiveStatus, UpdatedAt: time.Now().UTC()}}, nil)
	command := Command{TableRepo: repo}

	err := command.TischAktualisieren(context.Background(), 1, "New Name")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	tisch, err := command.TableRepo.GetTable(context.Background(), 1)
	if err != nil {
		t.Fatalf("expected no error retrieving tisch, got %v", err)
	}
	if tisch.Name != "New Name" {
		t.Errorf("expected tisch name to be 'New Name', got %s", tisch.Name)
	}
}

func TestTischAktualisieren_NotFound(t *testing.T) {
	repo := table_repo.NewMock([]table.Tisch{}, db.ErrNotFound)
	command := Command{TableRepo: repo}

	err := command.TischAktualisieren(context.Background(), 999, "New Name")
	if err != ErrTischNotFound {
		t.Fatalf("expected ErrTischNotFound, got %v", err)
	}
}

func TestTischAktivieren(t *testing.T) {
	repo := table_repo.NewMock([]table.Tisch{{ID: 1, Name: "Tisch 1", Status: table.InactiveStatus, UpdatedAt: time.Now().UTC()}}, nil)
	command := Command{TableRepo: repo}

	err := command.TischAktivieren(context.Background(), 1)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	tbl, err := repo.GetTable(context.Background(), 1)
	if err != nil {
		t.Fatalf("expected no error retrieving tisch, got %v", err)
	}
	if tbl.Status != table.ActiveStatus {
		t.Errorf("expected tisch status to be Active, got %v", tbl.Status)
	}
}

func TestTischAktivieren_NotFound(t *testing.T) {
	repo := table_repo.NewMock([]table.Tisch{}, db.ErrNotFound)
	command := Command{TableRepo: repo}

	err := command.TischAktivieren(context.Background(), 999)
	if err != ErrTischNotFound {
		t.Fatalf("expected ErrTischNotFound, got %v", err)
	}
}

func TestTischDeaktivieren(t *testing.T) {
	repo := table_repo.NewMock([]table.Tisch{{ID: 1, Name: "Tisch 1", Status: table.ActiveStatus, UpdatedAt: time.Now().UTC()}}, nil)
	command := Command{TableRepo: repo}

	err := command.TischDeaktivieren(context.Background(), 1)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	tbl, err := repo.GetTable(context.Background(), 1)
	if err != nil {
		t.Fatalf("expected no error retrieving tisch, got %v", err)
	}
	if tbl.Status != table.InactiveStatus {
		t.Errorf("expected tisch status to be Inactive, got %v", tbl.Status)
	}
}

func TestTischDeaktivieren_NotFound(t *testing.T) {
	repo := table_repo.NewMock([]table.Tisch{}, db.ErrNotFound)
	command := Command{TableRepo: repo}

	err := command.TischDeaktivieren(context.Background(), 999)
	if err != ErrTischNotFound {
		t.Fatalf("expected ErrTischNotFound, got %v", err)
	}
}
