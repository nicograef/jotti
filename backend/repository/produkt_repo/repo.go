package produkt_repo

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/produkt"
	"github.com/nicograef/jotti/backend/domain/steuer"
	"github.com/nicograef/jotti/backend/sqlc/dbgen"
)

func (r Repository) GetProdukt(ctx context.Context, id int) (produkt.Produkt, error) {
	row, err := r.q.GetProdukt(ctx, id)
	if err != nil {
		return produkt.Produkt{}, db.Error(err)
	}

	varianten, err := parseVariantenJSON(row.Varianten)
	if err != nil {
		return produkt.Produkt{}, fmt.Errorf("unmarshal varianten: %w", err)
	}

	return produkt.Produkt{
		ID:         row.ID,
		Name:       row.Name,
		Kategorie:  produkt.Kategorie(row.Kategorie),
		Steuersatz: steuer.Steuersatz(row.Steuersatz),
		Status:     produkt.Status(row.Status),
		Varianten:  varianten,
		CreatedAt:  row.CreatedAt,
		UpdatedAt:  row.UpdatedAt,
	}, nil
}

func (r Repository) GetAllProdukte(ctx context.Context) ([]produkt.Produkt, error) {
	rows, err := r.q.GetAlleProdukte(ctx)
	if err != nil {
		return nil, db.Error(err)
	}

	produkte := make([]produkt.Produkt, 0, len(rows))
	for i := range rows {
		varianten, err := parseVariantenJSON(rows[i].Varianten)
		if err != nil {
			return nil, fmt.Errorf("unmarshal varianten: %w", err)
		}

		produkte = append(produkte, produkt.Produkt{
			ID:         rows[i].ID,
			Name:       rows[i].Name,
			Kategorie:  produkt.Kategorie(rows[i].Kategorie),
			Steuersatz: steuer.Steuersatz(rows[i].Steuersatz),
			Status:     produkt.Status(rows[i].Status),
			Varianten:  varianten,
			CreatedAt:  rows[i].CreatedAt,
			UpdatedAt:  rows[i].UpdatedAt,
		})
	}

	return produkte, nil
}

func (r Repository) GetActiveProdukte(ctx context.Context) ([]produkt.Produkt, error) {
	rows, err := r.q.GetAktiveProdukte(ctx)
	if err != nil {
		return nil, db.Error(err)
	}

	produkte := make([]produkt.Produkt, 0, len(rows))
	for i := range rows {
		varianten, err := parseVariantenJSON(rows[i].Varianten)
		if err != nil {
			return nil, fmt.Errorf("unmarshal varianten: %w", err)
		}

		produkte = append(produkte, produkt.Produkt{
			ID:         rows[i].ID,
			Name:       rows[i].Name,
			Kategorie:  produkt.Kategorie(rows[i].Kategorie),
			Steuersatz: steuer.Steuersatz(rows[i].Steuersatz),
			Status:     produkt.Status(rows[i].Status),
			Varianten:  varianten,
			CreatedAt:  rows[i].CreatedAt,
			UpdatedAt:  rows[i].UpdatedAt,
		})
	}

	return produkte, nil
}

func (r Repository) CreateProdukt(ctx context.Context, p produkt.Produkt) (int, error) {
	id, err := r.q.CreateProdukt(ctx, dbgen.CreateProduktParams{
		Name:       p.Name,
		Kategorie:  dbgen.Produktkategorie(p.Kategorie),
		Steuersatz: dbgen.Steuersatz(p.Steuersatz),
		Status:     dbgen.Entitystatus(p.Status),
		CreatedAt:  p.CreatedAt,
		UpdatedAt:  p.UpdatedAt,
	})
	if err != nil {
		return 0, db.Error(err)
	}

	return id, nil
}

func (r Repository) UpdateProdukt(ctx context.Context, p produkt.Produkt) error {
	result, err := r.q.UpdateProdukt(ctx, dbgen.UpdateProduktParams{
		Name:       p.Name,
		Kategorie:  dbgen.Produktkategorie(p.Kategorie),
		Steuersatz: dbgen.Steuersatz(p.Steuersatz),
		Status:     dbgen.Entitystatus(p.Status),
		UpdatedAt:  p.UpdatedAt,
		ID:         p.ID,
	})
	if err != nil {
		return db.Error(err)
	}

	return db.ResultError(result)
}

// VerschiebeProdukt tauscht die Reihenfolge eines Produkts mit der seines
// unmittelbaren Nachbarn in derselben Kategorie; hoch bedeutet in Richtung
// Listenanfang. Beide Updates teilen sich eine Transaktion, damit nie nur die
// Hälfte des Tauschs persistiert wird und die Liste keinen Zwischenzustand
// zeigt. Steht das Produkt bereits am Rand seiner Kategorie, gibt es keinen
// Nachbarn und die Methode tut nichts — das Verschieben ist idempotent, nicht
// fehlerhaft.
//
// Nachbar ist immer die direkt angrenzende Zeile, auch wenn sie inaktiv und
// damit im Service unsichtbar ist: die Admin-Liste zeigt, was passiert.
func (r Repository) VerschiebeProdukt(ctx context.Context, produktID int, hoch bool) error {
	return db.WithTx(ctx, r.db, func(qtx *dbgen.Queries) error {
		aktuell, err := qtx.GetProduktReihenfolge(ctx, produktID)
		if err != nil {
			return db.Error(err)
		}

		nachbarID, nachbarReihenfolge, gefunden, err := produktNachbar(ctx, qtx, aktuell, hoch)
		if err != nil || !gefunden {
			return err
		}

		now := time.Now().UTC()
		if err := setProduktReihenfolge(ctx, qtx, produktID, nachbarReihenfolge, now); err != nil {
			return err
		}

		return setProduktReihenfolge(ctx, qtx, nachbarID, aktuell.Reihenfolge, now)
	})
}

// VerschiebeVariante tauscht die Reihenfolge einer Variante mit der ihres
// unmittelbaren Nachbarn im selben Produkt. Verhalten wie VerschiebeProdukt.
func (r Repository) VerschiebeVariante(ctx context.Context, varianteID int, hoch bool) error {
	return db.WithTx(ctx, r.db, func(qtx *dbgen.Queries) error {
		aktuell, err := qtx.GetVarianteReihenfolge(ctx, varianteID)
		if err != nil {
			return db.Error(err)
		}

		nachbarID, nachbarReihenfolge, gefunden, err := varianteNachbar(ctx, qtx, aktuell, hoch)
		if err != nil || !gefunden {
			return err
		}

		now := time.Now().UTC()
		if err := setVarianteReihenfolge(ctx, qtx, varianteID, nachbarReihenfolge, now); err != nil {
			return err
		}

		return setVarianteReihenfolge(ctx, qtx, nachbarID, aktuell.Reihenfolge, now)
	})
}

// produktNachbar liefert id und reihenfolge des angrenzenden Produkts. Das
// Ausbleiben eines Nachbarn ist ein regulärer Randfall (gefunden = false), kein
// Fehler.
func produktNachbar(ctx context.Context, qtx *dbgen.Queries, aktuell dbgen.GetProduktReihenfolgeRow, hoch bool) (int, int, bool, error) {
	if hoch {
		nachbar, err := qtx.GetProduktVorgaenger(ctx, dbgen.GetProduktVorgaengerParams{
			Kategorie:   aktuell.Kategorie,
			Reihenfolge: aktuell.Reihenfolge,
			ID:          aktuell.ID,
		})
		if err != nil {
			return 0, 0, false, ignoriereRandfall(err)
		}
		return nachbar.ID, nachbar.Reihenfolge, true, nil
	}

	nachbar, err := qtx.GetProduktNachfolger(ctx, dbgen.GetProduktNachfolgerParams{
		Kategorie:   aktuell.Kategorie,
		Reihenfolge: aktuell.Reihenfolge,
		ID:          aktuell.ID,
	})
	if err != nil {
		return 0, 0, false, ignoriereRandfall(err)
	}
	return nachbar.ID, nachbar.Reihenfolge, true, nil
}

func varianteNachbar(ctx context.Context, qtx *dbgen.Queries, aktuell dbgen.GetVarianteReihenfolgeRow, hoch bool) (int, int, bool, error) {
	if hoch {
		nachbar, err := qtx.GetVarianteVorgaenger(ctx, dbgen.GetVarianteVorgaengerParams{
			ProduktID:   aktuell.ProduktID,
			Reihenfolge: aktuell.Reihenfolge,
			ID:          aktuell.ID,
		})
		if err != nil {
			return 0, 0, false, ignoriereRandfall(err)
		}
		return nachbar.ID, nachbar.Reihenfolge, true, nil
	}

	nachbar, err := qtx.GetVarianteNachfolger(ctx, dbgen.GetVarianteNachfolgerParams{
		ProduktID:   aktuell.ProduktID,
		Reihenfolge: aktuell.Reihenfolge,
		ID:          aktuell.ID,
	})
	if err != nil {
		return 0, 0, false, ignoriereRandfall(err)
	}
	return nachbar.ID, nachbar.Reihenfolge, true, nil
}

// ignoriereRandfall macht aus "kein Nachbar vorhanden" ein nil-Ergebnis und
// lässt jeden anderen Datenbankfehler durch.
func ignoriereRandfall(err error) error {
	mapped := db.Error(err)
	if errors.Is(mapped, db.ErrNotFound) {
		return nil
	}
	return mapped
}

func setProduktReihenfolge(ctx context.Context, qtx *dbgen.Queries, id int, reihenfolge int, aktualisiert time.Time) error {
	result, err := qtx.SetProduktReihenfolge(ctx, dbgen.SetProduktReihenfolgeParams{
		Reihenfolge: reihenfolge,
		UpdatedAt:   aktualisiert,
		ID:          id,
	})
	if err != nil {
		return db.Error(err)
	}
	return db.ResultError(result)
}

func setVarianteReihenfolge(ctx context.Context, qtx *dbgen.Queries, id int, reihenfolge int, aktualisiert time.Time) error {
	result, err := qtx.SetVarianteReihenfolge(ctx, dbgen.SetVarianteReihenfolgeParams{
		Reihenfolge: reihenfolge,
		UpdatedAt:   aktualisiert,
		ID:          id,
	})
	if err != nil {
		return db.Error(err)
	}
	return db.ResultError(result)
}

// DeleteProduktMitVarianten persists the soft-delete of a produkt together with
// all its varianten in a single transaction. The caller passes the produkt with
// Delete() already applied to it and each variante; this method only writes the
// status transitions. Because all writes share one db.WithTx, a mid-operation
// failure rolls the whole delete back — the produkt and every variante stay in
// their pre-delete state, never a partial delete.
func (r Repository) DeleteProduktMitVarianten(ctx context.Context, p produkt.Produkt) error {
	return db.WithTx(ctx, r.db, func(qtx *dbgen.Queries) error {
		for i := range p.Varianten {
			v := p.Varianten[i]
			result, err := qtx.UpdateVariante(ctx, dbgen.UpdateVarianteParams{
				Name:       v.Name,
				PreisCents: v.PreisCents,
				Status:     dbgen.Entitystatus(v.Status),
				UpdatedAt:  v.UpdatedAt,
				ID:         v.ID,
			})
			if err != nil {
				return db.Error(err)
			}
			if err := db.ResultError(result); err != nil {
				return err
			}
		}

		result, err := qtx.UpdateProdukt(ctx, dbgen.UpdateProduktParams{
			Name:       p.Name,
			Kategorie:  dbgen.Produktkategorie(p.Kategorie),
			Steuersatz: dbgen.Steuersatz(p.Steuersatz),
			Status:     dbgen.Entitystatus(p.Status),
			UpdatedAt:  p.UpdatedAt,
			ID:         p.ID,
		})
		if err != nil {
			return db.Error(err)
		}
		return db.ResultError(result)
	})
}

// SortiereVariantenAlphabetisch vergibt die Reihenfolge aller Varianten eines
// Produkts alphabetisch neu. Eine einzelne UPDATE-Anweisung genuegt und ist von
// sich aus atomar, deshalb ohne explizite Transaktion. Ein Produkt ohne
// Varianten ist kein Fehler, sondern schlicht wirkungslos.
func (r Repository) SortiereVariantenAlphabetisch(ctx context.Context, produktID int) error {
	err := r.q.SortiereVariantenAlphabetisch(ctx, dbgen.SortiereVariantenAlphabetischParams{
		UpdatedAt: time.Now().UTC(),
		ProduktID: produktID,
	})
	if err != nil {
		return db.Error(err)
	}
	return nil
}
