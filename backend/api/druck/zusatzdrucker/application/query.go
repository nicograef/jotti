package application

import (
	"context"

	"github.com/nicograef/jotti/backend/domain/druckstation"
	"github.com/nicograef/jotti/backend/internal/logging"
)

type zusatzdruckerQueryRepo interface {
	GetAlleZusatzdrucker(ctx context.Context) ([]druckstation.Zusatzdrucker, error)
	GetProduktZuordnungen(ctx context.Context) ([]druckstation.ProduktZuordnung, error)
}

type Query struct {
	ZusatzdruckerRepo zusatzdruckerQueryRepo
}

// ZusatzdruckerMitProdukten carries the Produkte assigned to the Zusatzdrucker in the order of the Produktverwaltung.
type ZusatzdruckerMitProdukten struct {
	druckstation.Zusatzdrucker
	Produkte []druckstation.ProduktZuordnung
}

// Uebersicht is everything the Bondrucker page shows about Zusatzdrucker. ProduktAuswahl lists every
// non-deleted Produkt, assigned or not.
type Uebersicht struct {
	Zusatzdrucker  []ZusatzdruckerMitProdukten
	ProduktAuswahl []druckstation.ProduktZuordnung
}

func (q Query) GetAlle(ctx context.Context) (Uebersicht, error) {
	log := logging.FromContext(ctx)

	alle, err := q.ZusatzdruckerRepo.GetAlleZusatzdrucker(ctx)
	if err != nil {
		log.Error("Failed to retrieve zusatzdrucker", "error", err)
		return Uebersicht{}, ErrDatabase
	}

	zuordnungen, err := q.ZusatzdruckerRepo.GetProduktZuordnungen(ctx)
	if err != nil {
		log.Error("Failed to retrieve produkt zuordnungen", "error", err)
		return Uebersicht{}, ErrDatabase
	}

	produkteJeDrucker := make(map[int][]druckstation.ProduktZuordnung)
	for _, zuordnung := range zuordnungen {
		if zuordnung.ZusatzdruckerID != 0 {
			produkteJeDrucker[zuordnung.ZusatzdruckerID] = append(produkteJeDrucker[zuordnung.ZusatzdruckerID], zuordnung)
		}
	}

	result := Uebersicht{
		Zusatzdrucker:  make([]ZusatzdruckerMitProdukten, 0, len(alle)),
		ProduktAuswahl: zuordnungen,
	}
	for _, zd := range alle {
		result.Zusatzdrucker = append(result.Zusatzdrucker, ZusatzdruckerMitProdukten{
			Zusatzdrucker: zd,
			Produkte:      produkteJeDrucker[zd.ID],
		})
	}

	log.Info("Retrieved zusatzdrucker", "count", len(alle), "produkte", len(zuordnungen))
	return result, nil
}
