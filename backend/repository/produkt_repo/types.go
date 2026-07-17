package produkt_repo

import (
	"database/sql"
	"encoding/json"

	"github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/produkt"
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
