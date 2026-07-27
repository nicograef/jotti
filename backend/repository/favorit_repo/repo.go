package favorit_repo

import (
	"context"

	"github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/sqlc/dbgen"
)

func (r Repository) Add(ctx context.Context, userID, tischID int) error {
	err := r.q.AddFavorit(ctx, dbgen.AddFavoritParams{
		UserID:  userID,
		TischID: tischID,
	})
	if err != nil {
		return db.Error(err)
	}
	return nil
}

func (r Repository) Remove(ctx context.Context, userID, tischID int) error {
	err := r.q.RemoveFavorit(ctx, dbgen.RemoveFavoritParams{
		UserID:  userID,
		TischID: tischID,
	})
	if err != nil {
		return db.Error(err)
	}
	return nil
}

// RemoveByTisch entfernt die Markierungen aller Servicekräfte für einen Tisch.
// Wird beim Löschen eines Tisches ausgeführt; eine zurückbleibende Markierung
// wäre nicht mehr abwählbar, weil ein gelöschter Tisch in der Tischauswahl nicht
// mehr erscheint.
func (r Repository) RemoveByTisch(ctx context.Context, tischID int) error {
	err := r.q.RemoveFavoritenByTisch(ctx, tischID)
	if err != nil {
		return db.Error(err)
	}
	return nil
}

func (r Repository) GetByUser(ctx context.Context, userID int) ([]int, error) {
	ids, err := r.q.GetFavoritenByUser(ctx, userID)
	if err != nil {
		return nil, db.Error(err)
	}
	return ids, nil
}
