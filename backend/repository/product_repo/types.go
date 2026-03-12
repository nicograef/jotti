package product_repo

import (
	"database/sql"
	"encoding/json"

	"github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/product"
	"github.com/nicograef/jotti/backend/sqlc/dbgen"
)

type Repository struct {
	DB *sql.DB
	q  *dbgen.Queries
}

func NewRepository(db *sql.DB) Repository {
	return Repository{DB: db, q: dbgen.New(db)}
}

type jsonVariant struct {
	ID         int         `json:"id"`
	Name       string      `json:"name"`
	PreisCents int         `json:"price_cents"`
	Status     string      `json:"status"`
	CreatedAt  db.NullTime `json:"created_at"`
	UpdatedAt  db.NullTime `json:"updated_at"`
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

func variantRowToDomain(row dbgen.GetVariantRow) product.Variante {
	return product.Variante{
		ID:         row.ID,
		Name:       row.Name,
		PreisCents: row.PriceCents,
		Status:     product.Status(row.Status),
		CreatedAt:  row.CreatedAt,
		UpdatedAt:  row.UpdatedAt,
	}
}
