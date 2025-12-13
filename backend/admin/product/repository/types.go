package repository

import (
	"database/sql"

	"github.com/nicograef/jotti/backend/admin/product/domain"
)

// Repository implements product persistence layer using a SQL database.
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

func (dp *dbproduct) toDomain() domain.Product {
	return domain.Product{
		ID:            dp.ID,
		Name:          dp.Name,
		Description:   dp.Description,
		NetPriceCents: dp.NetPriceCents,
		Status:        domain.Status(dp.Status),
		Category:      domain.Category(dp.Category),
		CreatedAt:     dp.CreatedAt.Time,
	}
}
