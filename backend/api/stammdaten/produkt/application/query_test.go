package application

import (
	"context"
	"slices"
	"testing"

	"github.com/nicograef/jotti/backend/domain/produkt"
	"github.com/nicograef/jotti/backend/repository/repotest"
)

func TestGetActiveProdukte_PfandProdukteNurMitKennzeichen(t *testing.T) {
	aktiv := []produkt.Variante{{ID: 1, Name: "0,5l", PreisCents: 350, Status: produkt.ActiveStatus}}
	bier := produkt.Produkt{ID: 1, Name: "Bier", Kategorie: produkt.GetraenkKategorie, Status: produkt.ActiveStatus, Varianten: aktiv}
	pfand := produkt.Produkt{ID: 2, Name: "Pfand", Kategorie: produkt.SonstigesKategorie, Status: produkt.ActiveStatus, Pfand: true, Varianten: aktiv}
	query := Query{ProduktRepo: repotest.NewProduktRepo([]produkt.Produkt{bier, pfand}, nil)}

	got, err := query.GetActiveProdukte(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if ids := produktIDs(got.Produkte); !slices.Equal(ids, []int{1, 2}) {
		t.Errorf("expected Produkte [1 2], got %v", ids)
	}
	if ids := produktIDs(got.PfandProdukte); !slices.Equal(ids, []int{2}) {
		t.Errorf("expected PfandProdukte [2], got %v", ids)
	}
}

func produktIDs(produkte []produkt.Produkt) []int {
	ids := make([]int, 0, len(produkte))
	for _, p := range produkte {
		ids = append(ids, p.ID)
	}
	slices.Sort(ids)
	return ids
}

// The service drops a link whose target it does not offer; the Admin sees the stored link.
func TestGetActiveProdukte_PfandVerweis(t *testing.T) {
	glasID := 20
	bier := produkt.Produkt{ID: 1, Name: "Bier", Status: produkt.ActiveStatus, Varianten: []produkt.Variante{
		{ID: 10, Name: "0,5 l", PreisCents: 400, Status: produkt.ActiveStatus, PfandVarianteID: &glasID},
	}}
	pfand := produkt.Produkt{ID: 2, Name: "Glaspfand", Status: produkt.ActiveStatus, Pfand: true, Varianten: []produkt.Variante{
		{ID: 20, Name: "Glas", PreisCents: 200, Status: produkt.InactiveStatus},
		{ID: 21, Name: "Becher", PreisCents: 100, Status: produkt.ActiveStatus},
	}}
	repo := repotest.NewProduktRepo([]produkt.Produkt{bier, pfand}, nil)
	ctx := context.Background()

	aktive, err := Query{ProduktRepo: repo}.GetActiveProdukte(ctx)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if got := pfandVerweisVon(t, aktive.Produkte, 10); got != nil {
		t.Errorf("service: expected no Pfand for an inactive Pfand-Variante, got %d", *got)
	}

	alle, err := Query{ProduktRepo: repo}.GetAllProdukte(ctx)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if got := pfandVerweisVon(t, alle, 10); got == nil || *got != glasID {
		t.Errorf("admin: expected Pfand %d, got %v", glasID, got)
	}
}

func pfandVerweisVon(t *testing.T, produkte []produkt.Produkt, varianteID int) *int {
	t.Helper()
	for _, p := range produkte {
		for _, v := range p.Varianten {
			if v.ID == varianteID {
				return v.PfandVarianteID
			}
		}
	}
	t.Fatalf("variante %d missing", varianteID)
	return nil
}
