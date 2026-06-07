//go:build unit

package application

import (
	"context"
	"errors"
	"testing"
)

type mockDruckauftragRepo struct {
	offene []DruckAuftrag
	err    error
}

func (m *mockDruckauftragRepo) GetOffeneDruckauftraege(_ context.Context) ([]DruckAuftrag, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.offene, nil
}

func TestGetOffeneDruckauftraege_Empty(t *testing.T) {
	q := Query{
		DruckauftragRepo: &mockDruckauftragRepo{offene: nil},
	}

	auftraege, err := q.GetOffeneDruckauftraege(context.Background())

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if auftraege != nil {
		t.Errorf("expected nil, got %v", auftraege)
	}
}

func TestGetOffeneDruckauftraege_RepoError(t *testing.T) {
	q := Query{
		DruckauftragRepo: &mockDruckauftragRepo{err: errors.New("db error")},
	}

	_, err := q.GetOffeneDruckauftraege(context.Background())

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGetOffeneDruckauftraege_Success(t *testing.T) {
	offene := []DruckAuftrag{
		{ID: 12, ZielIP: "192.168.1.51", Payload: "AAA="},
		{ID: 13, ZielIP: "192.168.1.52", Payload: "BBB="},
	}
	q := Query{
		DruckauftragRepo: &mockDruckauftragRepo{offene: offene},
	}

	auftraege, err := q.GetOffeneDruckauftraege(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(auftraege) != 2 {
		t.Fatalf("expected 2 auftraege, got %d", len(auftraege))
	}
	if auftraege[0].ID != 12 || auftraege[1].ID != 13 {
		t.Fatalf("unexpected IDs: %+v", auftraege)
	}
	if auftraege[0].ZielIP != "192.168.1.51" || auftraege[1].ZielIP != "192.168.1.52" {
		t.Fatalf("unexpected zielIp values: %+v", auftraege)
	}
}
