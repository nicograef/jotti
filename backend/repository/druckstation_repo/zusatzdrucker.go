package druckstation_repo

import (
	"context"
	"database/sql"

	"github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/druckstation"
	"github.com/nicograef/jotti/backend/domain/produkt"
	"github.com/nicograef/jotti/backend/sqlc/dbgen"
)

func (r Repository) GetAlleZusatzdrucker(ctx context.Context) ([]druckstation.Zusatzdrucker, error) {
	rows, err := r.q.GetAlleZusatzdrucker(ctx)
	if err != nil {
		return nil, db.Error(err)
	}
	result := make([]druckstation.Zusatzdrucker, 0, len(rows))
	for _, row := range rows {
		result = append(result, zusatzdruckerFromRow(dbgen.GetZusatzdruckerRow(row)))
	}
	return result, nil
}

func (r Repository) GetZusatzdrucker(ctx context.Context, id int) (druckstation.Zusatzdrucker, error) {
	row, err := r.q.GetZusatzdrucker(ctx, id)
	if err != nil {
		return druckstation.Zusatzdrucker{}, db.Error(err)
	}
	return zusatzdruckerFromRow(row), nil
}

func (r Repository) CreateZusatzdrucker(ctx context.Context, zd druckstation.Zusatzdrucker) (int, error) {
	id, err := r.q.CreateZusatzdrucker(ctx, dbgen.CreateZusatzdruckerParams{
		Name:      zd.Name,
		DruckerIp: zd.DruckerIP,
		Bonmodus:  string(zd.Bonmodus),
	})
	if err != nil {
		return 0, db.Error(err)
	}
	return id, nil
}

func (r Repository) UpdateZusatzdrucker(ctx context.Context, zd druckstation.Zusatzdrucker) error {
	result, err := r.q.UpdateZusatzdrucker(ctx, dbgen.UpdateZusatzdruckerParams{
		ID:        zd.ID,
		Name:      zd.Name,
		DruckerIp: zd.DruckerIP,
		Bonmodus:  string(zd.Bonmodus),
	})
	return einerBetroffen(result, err)
}

// DeleteZusatzdrucker removes its Produkt-Zuordnungen in the same statement (ON DELETE CASCADE).
func (r Repository) DeleteZusatzdrucker(ctx context.Context, id int) error {
	result, err := r.q.DeleteZusatzdrucker(ctx, id)
	return einerBetroffen(result, err)
}

// AssignProdukt moves an existing Zuordnung; a deleted Produkt or a missing Zusatzdrucker gives db.ErrNotFound.
func (r Repository) AssignProdukt(ctx context.Context, produktID, zusatzdruckerID int) error {
	result, err := r.q.AssignProduktZusatzdrucker(ctx, dbgen.AssignProduktZusatzdruckerParams{
		ProduktID:       produktID,
		ZusatzdruckerID: zusatzdruckerID,
	})
	return einerBetroffen(result, err)
}

// UnassignProdukt succeeds without an existing Zuordnung.
func (r Repository) UnassignProdukt(ctx context.Context, produktID int) error {
	return db.Error(r.q.UnassignProduktZusatzdrucker(ctx, produktID))
}

// GetProduktZuordnungen lists every non-deleted Produkt in the order of the Produktverwaltung.
func (r Repository) GetProduktZuordnungen(ctx context.Context) ([]druckstation.ProduktZuordnung, error) {
	rows, err := r.q.GetProduktZuordnungen(ctx)
	if err != nil {
		return nil, db.Error(err)
	}
	result := make([]druckstation.ProduktZuordnung, 0, len(rows))
	for _, row := range rows {
		result = append(result, druckstation.ProduktZuordnung{
			ProduktID:         row.ProduktID,
			ProduktName:       row.ProduktName,
			Kategorie:         produkt.Kategorie(row.Kategorie),
			ZusatzdruckerID:   row.ZusatzdruckerID,
			ZusatzdruckerName: row.ZusatzdruckerName,
		})
	}
	return result, nil
}

// GetZusatzdruckerJeVariante keys the Zusatzdrucker of each Variante's Produkt by Variante-ID; a Variante
// without one has no entry. A Zusatzdrucker without IP is included: it prints nothing and does not fall back.
func (r Repository) GetZusatzdruckerJeVariante(ctx context.Context, varianteIDs []int) (map[int]druckstation.Zusatzdrucker, error) {
	result := make(map[int]druckstation.Zusatzdrucker)
	if len(varianteIDs) == 0 {
		return result, nil
	}

	const query = `SELECT v.id, z.id, z.name, z.drucker_ip, z.bonmodus
		FROM produkt_varianten v
		JOIN zusatzdrucker_produkte zp ON zp.produkt_id = v.produkt_id
		JOIN zusatzdrucker z ON z.id = zp.zusatzdrucker_id
		WHERE v.id = ANY($1)`

	rows, err := r.db.QueryContext(ctx, query, db.Int4Array(varianteIDs))
	if err != nil {
		return nil, db.Error(err)
	}
	defer rows.Close() //nolint:errcheck // explicit Close with error check below

	for rows.Next() {
		var (
			varianteID int
			zd         druckstation.Zusatzdrucker
			bonmodus   string
		)
		if err := rows.Scan(&varianteID, &zd.ID, &zd.Name, &zd.DruckerIP, &bonmodus); err != nil {
			return nil, db.Error(err)
		}
		zd.Bonmodus = druckstation.Bonmodus(bonmodus)
		result[varianteID] = zd
	}
	if err := rows.Close(); err != nil {
		return nil, db.Error(err)
	}
	if err := rows.Err(); err != nil {
		return nil, db.Error(err)
	}

	return result, nil
}

func zusatzdruckerFromRow(row dbgen.GetZusatzdruckerRow) druckstation.Zusatzdrucker {
	return druckstation.Zusatzdrucker{
		ID:        row.ID,
		Name:      row.Name,
		DruckerIP: row.DruckerIp,
		Bonmodus:  druckstation.Bonmodus(row.Bonmodus),
	}
}

// einerBetroffen turns an UPDATE, DELETE or INSERT … SELECT that touched no row into db.ErrNotFound.
func einerBetroffen(result sql.Result, err error) error {
	if err != nil {
		return db.Error(err)
	}
	n, err := result.RowsAffected()
	if err != nil {
		return db.Error(err)
	}
	if n == 0 {
		return db.ErrNotFound
	}
	return nil
}
