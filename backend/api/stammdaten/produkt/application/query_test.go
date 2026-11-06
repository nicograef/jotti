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
