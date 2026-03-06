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
	PriceCents int         `json:"price_cents"`
	Status     string      `json:"status"`
	CreatedAt  db.NullTime `json:"created_at"`
}

func (jv *jsonVariant) toDomain() product.Variant {
	return product.Variant{
		ID:         jv.ID,
		Name:       jv.Name,
		PriceCents: jv.PriceCents,
		Status:     product.Status(jv.Status),
		CreatedAt:  jv.CreatedAt.Time,
	}
}

func parseVariantsJSON(data json.RawMessage) ([]product.Variant, error) {
	var variants []jsonVariant
	if err := json.Unmarshal(data, &variants); err != nil {
		return nil, err
	}

	result := make([]product.Variant, 0, len(variants))
	for _, v := range variants {
		result = append(result, v.toDomain())
	}

	return result, nil
}

func variantRowToDomain(row dbgen.GetVariantRow) product.Variant {
	return product.Variant{
		ID:         row.ID,
		Name:       row.Name,
		PriceCents: row.PriceCents,
		Status:     product.Status(row.Status),
		CreatedAt:  row.CreatedAt,
	}
}
