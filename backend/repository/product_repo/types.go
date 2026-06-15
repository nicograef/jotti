package product_repo

import (
	"database/sql"
	"encoding/json"

	"github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/product"
	"github.com/nicograef/jotti/backend/sqlc/dbgen"
)

type Repository struct {
	db *sql.DB
	q  *dbgen.Queries
}

func NewRepository(db *sql.DB) Repository {
	return Repository{db: db, q: dbgen.New(db)}
}

type jsonVariant struct {
	ID         int         `json:"id"`
	Name       string      `json:"name"`
	PreisCents int         `json:"preisCents"`
	Status     string      `json:"status"`
	CreatedAt  db.NullTime `json:"createdAt"`
	UpdatedAt  db.NullTime `json:"updatedAt"`
}

func (jv *jsonVariant) toDomain() product.Variante {
	return product.Variante{
		ID:         jv.ID,
		Name:       jv.Name,
		PreisCents: jv.PreisCents,
		Status:     product.Status(jv.Status),
		CreatedAt:  jv.CreatedAt.Time,
		UpdatedAt:  jv.UpdatedAt.Time,
	}
}

func parseVariantsJSON(data json.RawMessage) ([]product.Variante, error) {
	var variants []jsonVariant
	if err := json.Unmarshal(data, &variants); err != nil {
		return nil, err
	}

	result := make([]product.Variante, 0, len(variants))
	for _, v := range variants {
		result = append(result, v.toDomain())
	}

	return result, nil
}

func variantRowToDomain(row dbgen.GetVarianteRow) product.Variante {
	return product.Variante{
		ID:         row.ID,
		Name:       row.Name,
		PreisCents: row.PreisCents,
		Status:     product.Status(row.Status),
		CreatedAt:  row.CreatedAt,
		UpdatedAt:  row.UpdatedAt,
	}
}
