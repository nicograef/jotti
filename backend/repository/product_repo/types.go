package product_repo

import (
	"database/sql"

	"github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/product"
)

type Repository struct {
	DB *sql.DB
}

type dbproduct struct {
	ID        int         `db:"id"`
	Name      string      `db:"name"`
	Category  string      `db:"category"`
	Variants  []dbvariant // Not stored directly in products table
	CreatedAt db.NullTime `db:"created_at"`
}

func (dp *dbproduct) toDomain() product.Product {
	return product.Product{
		ID:        dp.ID,
		Name:      dp.Name,
		Category:  product.Category(dp.Category),
		CreatedAt: dp.CreatedAt.Time,
	}
}

type dbvariant struct {
	ID         int         `db:"id" json:"id"`
	Name       string      `db:"name" json:"name"`
	PriceCents int         `db:"price_cents" json:"price_cents"`
	Status     string      `db:"status" json:"status"`
	CreatedAt  db.NullTime `db:"created_at" json:"created_at"`
}

func (dv *dbvariant) toDomain() product.Variant {
	return product.Variant{
		ID:         dv.ID,
		Name:       dv.Name,
		PriceCents: dv.PriceCents,
		Status:     product.Status(dv.Status),
		CreatedAt:  dv.CreatedAt.Time,
	}
}
