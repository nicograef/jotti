package product_repo

import (
	"database/sql"

	"github.com/nicograef/jotti/backend/domain/product"
)

type Repository struct {
	DB *sql.DB
}

type dbproduct struct {
	ID            int          `db:"id"`
	Name          string       `db:"name"`
	Description   string       `db:"description"`
	NetPriceCents int          `db:"net_price_cents"`
	Status        string       `db:"status"`
	Category      string       `db:"category"`
	CreatedAt     sql.NullTime `db:"created_at"`
}

func (dp *dbproduct) toDomain() product.Product {
	return product.Product{
		ID:            dp.ID,
		Name:          dp.Name,
		Description:   dp.Description,
		NetPriceCents: dp.NetPriceCents,
		Status:        product.Status(dp.Status),
		Category:      product.Category(dp.Category),
		CreatedAt:     dp.CreatedAt.Time,
	}
}
