package produkt_repo

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/produkt"
	"github.com/nicograef/jotti/backend/domain/steuer"
	"github.com/nicograef/jotti/backend/sqlc/dbgen"
)

type Repository struct {
	db *sql.DB
	q  *dbgen.Queries
}

func NewRepository(db *sql.DB) Repository {
	return Repository{db: db, q: dbgen.New(db)}
}

type jsonVariante struct {
	ID         int         `json:"id"`
	Name       string      `json:"name"`
	PreisCents int         `json:"preisCents"`
	Status     string      `json:"status"`
	CreatedAt  db.NullTime `json:"createdAt"`
	UpdatedAt  db.NullTime `json:"updatedAt"`
}

func (jv *jsonVariante) toDomain() produkt.Variante {
	return produkt.Variante{
		ID:         jv.ID,
		Name:       jv.Name,
		PreisCents: jv.PreisCents,
		Status:     produkt.Status(jv.Status),
		CreatedAt:  jv.CreatedAt.Time,
		UpdatedAt:  jv.UpdatedAt.Time,
	}
}

func parseVariantenJSON(data json.RawMessage) ([]produkt.Variante, error) {
	var varianten []jsonVariante
	if err := json.Unmarshal(data, &varianten); err != nil {
		return nil, err
	}

	result := make([]produkt.Variante, 0, len(varianten))
	for _, v := range varianten {
		result = append(result, v.toDomain())
	}

	return result, nil
}

// produktRowToDomain baut ein Produkt aus einer Produkt-Zeile samt ihrer
// Varianten-JSON-Spalte. GetProduktRow, GetAlleProdukteRow und
// GetAktiveProdukteRow sind feldgleich (dieselben sqlc-Query-Spalten), deshalb
// konvertiert jeder Aufrufer seine Zeile per Typkonvertierung auf GetProduktRow.
func produktRowToDomain(row dbgen.GetProduktRow) (produkt.Produkt, error) {
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

func varianteRowToDomain(row dbgen.GetVarianteRow) produkt.Variante {
	return produkt.Variante{
		ID:         row.ID,
		Name:       row.Name,
		PreisCents: row.PreisCents,
		Status:     produkt.Status(row.Status),
		CreatedAt:  row.CreatedAt,
		UpdatedAt:  row.UpdatedAt,
	}
}
